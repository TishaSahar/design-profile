#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# start.sh — управление окружением Design-Portfolio
#
# Использование:
#   ./start.sh           — остановить всё и запустить заново
#   ./start.sh stop      — остановить все сервисы
#   ./start.sh status    — состояние сервисов
#   ./start.sh logs      — вывести последние строки обоих логов
#   ./start.sh logs back — следить за логом бэкенда (tail -f)
#   ./start.sh logs web  — следить за логом фронтенда (tail -f)
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

# ── Пути ──────────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend"
WEBUI_DIR="$SCRIPT_DIR/webui"
PID_DIR="$SCRIPT_DIR/.pids"
LOG_DIR="$SCRIPT_DIR/.logs"

BACKEND_PORT=8080
FRONTEND_PORT=3000
BACKEND_CONFIG="./config/config.yaml"

# ── Цвета ──────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

info()  { echo -e "${GREEN}✔${NC}  $*"; }
warn()  { echo -e "${YELLOW}⚠${NC}  $*"; }
error() { echo -e "${RED}✖${NC}  $*" >&2; }
step()  { echo -e "\n${BOLD}${BLUE}▶ $*${NC}"; }
url()   { echo -e "   ${CYAN}$*${NC}"; }

# ── Создать рабочие папки ──────────────────────────────────────────────────────
mkdir -p "$PID_DIR" "$LOG_DIR"

# ── Вспомогательные функции ────────────────────────────────────────────────────

# Запустить команду в bash login-shell (нужно для GVM / Homebrew Go)
go_run() {
  bash --login -c "$*"
}

# Сохранить PID в файл
save_pid() {
  echo "$1" > "$PID_DIR/$2.pid"
}

# Прочитать PID из файла
read_pid() {
  local f="$PID_DIR/$1.pid"
  [[ -f "$f" ]] && cat "$f" || echo ""
}

# Проверить, живёт ли процесс
is_alive() {
  local pid="$1"
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

# Остановить процесс по PID-файлу
stop_by_pid() {
  local name="$1"
  local pid
  pid=$(read_pid "$name")
  if is_alive "$pid"; then
    kill "$pid" 2>/dev/null && info "Остановлен $name (PID $pid)"
    # Дождаться завершения (до 5 с)
    local i=0
    while is_alive "$pid" && (( i < 50 )); do sleep 0.1; (( i++ )) || true; done
    is_alive "$pid" && kill -9 "$pid" 2>/dev/null || true
  fi
  rm -f "$PID_DIR/$name.pid"
}

# Освободить порт (на случай «зомби»-процессов)
free_port() {
  local port="$1"
  local pids
  pids=$(lsof -ti:"$port" 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    # shellcheck disable=SC2086
    kill -9 $pids 2>/dev/null || true
    info "Освобождён порт $port"
  fi
}

# Проверить доступность HTTP-адреса (до N попыток)
wait_http() {
  local url="$1"
  local attempts="${2:-30}"
  local i=0
  while (( i < attempts )); do
    if curl -sf "$url" -o /dev/null 2>/dev/null; then
      return 0
    fi
    sleep 0.5
    (( i++ )) || true
  done
  return 1
}

# ── Команда: stop ──────────────────────────────────────────────────────────────
cmd_stop() {
  step "Останавливаю сервисы…"
  stop_by_pid "backend"
  stop_by_pid "frontend"
  free_port "$BACKEND_PORT"
  free_port "$FRONTEND_PORT"
  info "Все сервисы остановлены"
}

# ── Команда: status ────────────────────────────────────────────────────────────
cmd_status() {
  step "Состояние сервисов"

  local backend_pid frontend_pid
  backend_pid=$(read_pid "backend")
  frontend_pid=$(read_pid "frontend")

  if is_alive "$backend_pid"; then
    info "Backend   запущен  (PID $backend_pid) → http://localhost:$BACKEND_PORT"
  else
    warn "Backend   остановлен"
  fi

  if is_alive "$frontend_pid"; then
    info "Frontend  запущен  (PID $frontend_pid) → http://localhost:$FRONTEND_PORT"
  else
    warn "Frontend  остановлен"
  fi

  # PostgreSQL
  if pg_isready -h localhost -q 2>/dev/null; then
    info "PostgreSQL доступен"
  else
    warn "PostgreSQL недоступен"
  fi
}

# ── Команда: logs ──────────────────────────────────────────────────────────────
cmd_logs() {
  local target="${1:-}"
  case "$target" in
    back|backend)
      exec tail -f "$LOG_DIR/backend.log"
      ;;
    web|frontend)
      exec tail -f "$LOG_DIR/frontend.log"
      ;;
    *)
      step "Последние 30 строк каждого лога"
      echo -e "\n${BOLD}── Backend ($LOG_DIR/backend.log) ──${NC}"
      tail -n 30 "$LOG_DIR/backend.log" 2>/dev/null || warn "Лог бэкенда пуст"
      echo -e "\n${BOLD}── Frontend ($LOG_DIR/frontend.log) ──${NC}"
      tail -n 30 "$LOG_DIR/frontend.log" 2>/dev/null || warn "Лог фронтенда пуст"
      ;;
  esac
}

# ── Команда: start (по умолчанию) ─────────────────────────────────────────────
cmd_start() {

  # 1. Остановить текущие сервисы
  cmd_stop

  # 2. PostgreSQL
  step "Проверяю PostgreSQL…"
  if ! pg_isready -h localhost -q 2>/dev/null; then
    warn "PostgreSQL не запущен — пробую запустить через brew…"
    brew services start postgresql@14 2>/dev/null \
      || brew services start postgresql 2>/dev/null \
      || { error "Не удалось запустить PostgreSQL"; exit 1; }
    sleep 2
    pg_isready -h localhost -q 2>/dev/null \
      || { error "PostgreSQL всё равно недоступен. Проверьте установку."; exit 1; }
  fi
  info "PostgreSQL доступен"

  # 3. Зависимости Go
  step "Проверяю зависимости Go…"
  if ! go_run "cd '$BACKEND_DIR' && go build ./... 2>&1" > /dev/null; then
    warn "Зависимости устарели — запускаю go mod tidy…"
    go_run "cd '$BACKEND_DIR' && go mod tidy" \
      || { error "go mod tidy завершился с ошибкой"; exit 1; }
  fi
  info "Зависимости в порядке"

  # 4. Backend
  step "Запускаю backend (порт $BACKEND_PORT)…"
  bash --login -c \
    "cd '$BACKEND_DIR' && go run ./cmd/main/main.go -config '$BACKEND_CONFIG'" \
    > "$LOG_DIR/backend.log" 2>&1 &
  local backend_pid=$!
  save_pid "$backend_pid" "backend"

  info "Backend запущен (PID $backend_pid) — жду готовности…"
  if wait_http "http://localhost:$BACKEND_PORT/api/v1/contacts" 40; then
    info "Backend готов"
  else
    error "Backend не ответил за 20 с. Проверьте лог:"
    tail -n 20 "$LOG_DIR/backend.log" >&2
    exit 1
  fi

  # 5. Frontend
  step "Запускаю frontend (порт $FRONTEND_PORT)…"
  python3 -m http.server "$FRONTEND_PORT" \
    --directory "$WEBUI_DIR" \
    > "$LOG_DIR/frontend.log" 2>&1 &
  local frontend_pid=$!
  save_pid "$frontend_pid" "frontend"

  if wait_http "http://localhost:$FRONTEND_PORT/" 10; then
    info "Frontend готов (PID $frontend_pid)"
  else
    error "Frontend не ответил. Проверьте лог: $LOG_DIR/frontend.log"
    exit 1
  fi

  # 6. Итог
  echo -e "\n${BOLD}${GREEN}═══════════════════════════════════════${NC}"
  echo -e "${BOLD}${GREEN}  Платформа запущена  ${NC}"
  echo -e "${BOLD}${GREEN}═══════════════════════════════════════${NC}"
  echo ""
  echo -e "  ${BOLD}Публичный сайт${NC}"
  url "http://localhost:$FRONTEND_PORT"
  echo ""
  echo -e "  ${BOLD}Панель администратора${NC}"
  url "http://localhost:$FRONTEND_PORT/admin.html"
  echo ""
  echo -e "  ${BOLD}Swagger UI${NC}"
  url "http://localhost:$BACKEND_PORT/swagger/index.html"
  echo ""
  echo -e "  ${BOLD}Логи${NC}"
  echo -e "   backend:  $LOG_DIR/backend.log"
  echo -e "   frontend: $LOG_DIR/frontend.log"
  echo ""
  echo -e "  ${BOLD}Остановить:${NC}  ./start.sh stop"
  echo -e "  ${BOLD}Статус:${NC}      ./start.sh status"
  echo -e "  ${BOLD}Логи:${NC}        ./start.sh logs [back|web]"
  echo ""
}

# ── Диспетчер команд ───────────────────────────────────────────────────────────
case "${1:-start}" in
  stop)             cmd_stop ;;
  status)           cmd_status ;;
  restart|start|"") cmd_start ;;
  logs)             cmd_logs "${2:-}" ;;
  *)
    error "Неизвестная команда: $1"
    echo "Использование: $0 [start|stop|restart|status|logs [back|web]]"
    exit 1
    ;;
esac
