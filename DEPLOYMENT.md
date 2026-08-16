# Развёртывание проекта

## Содержание

1. [Требования](#требования)
2. [Зависимости](#зависимости)
3. [Быстрый старт (разработка)](#быстрый-старт-разработка)
4. [Настройка конфигурации](#настройка-конфигурации)
5. [Генерация Swagger-документации](#генерация-swagger-документации)
6. [Запуск тестов](#запуск-тестов)
7. [Сборка и запуск через Docker](#сборка-и-запуск-через-docker)
8. [Продакшн-развёртывание (nginx + systemd)](#продакшн-развёртывание)

---

## Требования

| Компонент  | Версия     |
|------------|------------|
| Go         | 1.22+      |
| PostgreSQL | 15+        |
| Docker     | 24+ (опционально) |
| nginx      | 1.24+ (для продакшна) |

---

## Зависимости

### 1. Установка Go

```bash
# Ubuntu / Debian
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 2. Установка PostgreSQL

```bash
sudo apt update
sudo apt install -y postgresql postgresql-contrib
sudo systemctl enable --now postgresql

# Создать базу данных и пользователя
sudo -u postgres psql <<'SQL'
CREATE USER design_user WITH PASSWORD 'your_strong_password';
CREATE DATABASE design_portfolio OWNER design_user;
GRANT ALL PRIVILEGES ON DATABASE design_portfolio TO design_user;
SQL
```

### 3. Установка swag (генератор Swagger)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
# Убедитесь, что $GOPATH/bin в PATH:
export PATH=$PATH:$(go env GOPATH)/bin
```

---

## Быстрый старт (разработка)

```bash
# 1. Клонировать репозиторий
git clone <repo-url> design-profile
cd design-profile/backend

# 2. Установить зависимости Go
go mod tidy

# 3. Скопировать и заполнить конфигурацию
cp config/config.yaml config/config.local.yaml
# Отредактируйте config/config.local.yaml (см. раздел ниже)

# 4. Сгенерировать Swagger-документацию
make swagger

# 5. Запустить сервер
go run ./cmd/main/main.go -config ./config/config.local.yaml
```

Сервер будет доступен по адресу `http://localhost:8080`.
Swagger UI: `http://localhost:8080/swagger/index.html`.

### Открыть frontend

Откройте `webui/index.html` напрямую в браузере **или** раздайте папку `webui/` через любой статический HTTP-сервер:

```bash
# Быстрый вариант — встроенный сервер Python
cd webui
python3 -m http.server 3000
# Откройте http://localhost:3000
```

---

## Настройка конфигурации

Файл: `backend/config/config.yaml`

```yaml
server:
  host: 0.0.0.0
  port: 8080               # порт, на котором слушает API

database:
  host: localhost
  port: 5432
  name: design_portfolio   # имя базы данных
  user: design_user        # пользователь PostgreSQL
  password: your_password  # пароль пользователя
  ssl_mode: disable        # для продакшна рекомендуется "require"

email:
  smtp_host: smtp.gmail.com
  smtp_port: 587
  username: olya-portfolio@gmail.com   # Gmail-адрес приложения
  password: xxxx xxxx xxxx xxxx        # App Password из настроек Google
  sender: olya-portfolio@gmail.com
  admin_email: olga@example.com        # email администратора, которому приходят OTP-коды

jwt:
  secret: замените-на-случайную-строку-не-менее-32-символов
  expiration_hours: 24
```

### Настройка Gmail App Password

1. Перейдите в аккаунт Google → **Безопасность**.
2. Включите **Двухэтапную верификацию** (если ещё не включена).
3. В разделе «Как вы входите в Google» найдите **Пароли приложений**.
4. Создайте пароль для приложения «Почта» → скопируйте 16-символьный пароль.
5. Вставьте его в `email.password` в конфиге.

### Настройка frontend

Файл: `webui/config/config.js`

```js
const CONFIG = {
  BASE_URL: "http://your-server-ip:8080/api/v1",
};
```

---

## Генерация Swagger-документации

```bash
cd backend
make swagger
```

После этого Swagger UI доступен на `<host>:<port>/swagger/index.html`.

---

## Запуск тестов

```bash
cd backend
make test
```

Покрытие по пакетам выводится в консоль. Тесты не требуют запущенной базы данных — бизнес-логика и утилиты покрыты юнит-тестами без зависимости от инфраструктуры.

---

## Сборка и запуск через Docker

### Только бэкенд

```bash
cd backend

# Сборка образа (включает swag init)
docker build -t design-profile-backend -f docker/Dockerfile .

# Запуск (конфиг монтируется снаружи)
docker run -d \
  --name design-backend \
  -p 8080:8080 \
  -v /path/to/backend/config:/app/config \
  design-profile-backend
```

### Docker Compose (рекомендуется)

Создайте файл `docker-compose.yml` в корне проекта:

```yaml
version: "3.9"

services:
  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: design_portfolio
      POSTGRES_USER: design_user
      POSTGRES_PASSWORD: your_password
    volumes:
      - pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U design_user -d design_portfolio"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: ./backend
      dockerfile: docker/Dockerfile
    restart: unless-stopped
    depends_on:
      db:
        condition: service_healthy
    ports:
      - "8080:8080"
    volumes:
      - ./backend/config:/app/config:ro

volumes:
  pg_data:
```

```bash
# Запуск всего стека
docker compose up -d

# Остановка
docker compose down

# Просмотр логов
docker compose logs -f backend
```

---

## Продакшн-развёртывание

### nginx (reverse proxy + статика)

Пример конфигурации `/etc/nginx/sites-available/design-profile`:

```nginx
server {
    listen 80;
    server_name your-domain.ru;

    # Redirect HTTP → HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.ru;

    ssl_certificate     /etc/letsencrypt/live/your-domain.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.ru/privkey.pem;

    # Frontend (статические файлы)
    root /var/www/design-profile/webui;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # Панель администратора
    location /admin {
        try_files /admin.html =404;
        # Опционально: ограничьте доступ по IP
        # allow 1.2.3.4;
        # deny all;
    }

    # API (проксируется на бэкенд)
    location /api/ {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;

        # Ограничение на размер загружаемых файлов (10 фото × 20 МБ)
        client_max_body_size 210M;
    }

    # Swagger UI
    location /swagger/ {
        proxy_pass http://127.0.0.1:8080;
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/design-profile /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### Получение TLS-сертификата (Let's Encrypt)

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.ru
```

### systemd-сервис для бэкенда

Если вы не используете Docker, можно запустить бинарник как systemd-сервис.

Создайте `/etc/systemd/system/design-backend.service`:

```ini
[Unit]
Description=Design Profile Backend
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/design-profile/backend
ExecStart=/opt/design-profile/backend/server -config /opt/design-profile/backend/config/config.yaml
Restart=on-failure
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now design-backend
sudo systemctl status design-backend
```

### Развёртывание frontend

```bash
# Скопировать файлы на сервер
rsync -avz webui/ user@your-server:/var/www/design-profile/webui/

# Не забудьте обновить BASE_URL в webui/config/config.js
# перед копированием!
```

---

## Структура проекта

```
design-profile/
├── backend/
│   ├── cmd/main/main.go          # точка входа
│   ├── config/
│   │   ├── config.go             # загрузка конфигурации
│   │   └── config.yaml           # конфиг (не коммитить с секретами)
│   ├── docker/Dockerfile
│   ├── docs/                     # сгенерированный Swagger (make swagger)
│   ├── internal/
│   │   ├── auth/                 # OTP, JWT
│   │   ├── email/                # SMTP-клиент
│   │   ├── handler/              # HTTP-обработчики
│   │   ├── middleware/           # JWT middleware
│   │   ├── model/                # доменные модели
│   │   ├── repository/           # работа с БД
│   │   └── service/              # бизнес-логика
│   ├── migrations/               # SQL-миграции (применяются при старте)
│   ├── go.mod
│   └── Makefile
└── webui/
    ├── config/config.js          # URL бэкенда
    ├── index.html                # публичный сайт
    ├── admin.html                # панель администратора
    └── src/
        ├── api.js                # HTTP-клиент
        ├── main.js               # логика публичного сайта
        ├── admin.js              # логика админ-панели
        └── styles/main.css
```
