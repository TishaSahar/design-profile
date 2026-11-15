# Развёртывание проекта

## Содержание

1. [Требования](#требования)
2. [Локальная разработка](#локальная-разработка)
3. [Настройка конфигурации](#настройка-конфигурации)
4. [Swagger и тесты](#swagger-и-тесты)
5. [Продакшн: как устроен деплой](#продакшн-как-устроен-деплой)
6. [Первичная настройка сервера](#первичная-настройка-сервера)
7. [Настройка GitHub](#настройка-github)
8. [Проверки здоровья](#проверки-здоровья)
9. [Повседневные операции](#повседневные-операции)
10. [Частые проблемы](#частые-проблемы)
11. [Docker (необязательно)](#docker-необязательно)
12. [Структура проекта](#структура-проекта)

---

## Требования

| Компонент  | Версия | Где нужен |
|------------|--------|-----------|
| Go         | 1.23+  | локально и в CI (на сервере не нужен — туда попадает готовый бинарник) |
| PostgreSQL | 15+    | локально и на сервере |
| nginx      | 1.24+  | на сервере |
| rsync, curl | любые | на сервере (их использует деплой) |
| Python 3   | любой  | локально, для раздачи `webui/` (`start.sh`) |

---

## Локальная разработка

### 1. Конфигурация

```bash
cp backend/config/config.yaml.example backend/config/config.yaml
```

Заполните `backend/config/config.yaml` (см. [Настройка конфигурации](#настройка-конфигурации)).
**Этот файл содержит пароли и не должен попадать в Git** — он указан в `.gitignore`.

Базу данных для разработки можно создать так:

```bash
sudo -u postgres psql <<'SQL'
CREATE USER design_user WITH PASSWORD 'your_strong_password';
CREATE DATABASE design_portfolio OWNER design_user;
SQL
```

Таблицы создавать не нужно: SQL-миграции из `backend/migrations/` встроены в бинарник и применяются при каждом старте бэкенда (они идемпотентны).

### 2. Запуск

```bash
./start.sh
```

Скрипт собирает и запускает бэкенд (`http://localhost:8080`) и раздаёт `webui/` (`http://localhost:3000`). Другие команды: `./start.sh stop`, `./start.sh status`, `./start.sh logs back|web`.

| Адрес | Что |
|-------|-----|
| http://localhost:3000 | публичный сайт |
| http://localhost:3000/admin.html | панель администратора |
| http://localhost:8080/swagger/index.html | Swagger UI |
| http://localhost:8080/health | проверка бэкенда |

Локальный сервер фронтенда не запрещает кэширование: если после правок виден старый дизайн, обновите страницу с очисткой кэша (**Cmd+Shift+R**).

---

## Настройка конфигурации

### `backend/config/config.yaml`

```yaml
server:
  host: 127.0.0.1          # на сервере — только локально, наружу API отдаёт nginx
  port: 8080

database:
  host: localhost
  port: 5432
  name: design_portfolio
  user: design_user
  password: your_password
  ssl_mode: disable        # "require", если база на другом сервере

email:
  smtp_host: smtp.gmail.com
  smtp_port: 587
  username: olya-portfolio@gmail.com   # Gmail-адрес приложения
  password: xxxx xxxx xxxx xxxx        # App Password из настроек Google
  sender: olya-portfolio@gmail.com
  admin_email: olga@example.com        # email администратора, которому приходят OTP-коды

jwt:
  secret: замените-на-случайную-строку-не-менее-32-символов   # openssl rand -base64 32
  expiration_hours: 24
```

#### Gmail App Password

1. Перейдите в аккаунт Google → **Безопасность**.
2. Включите **Двухэтапную верификацию** (если ещё не включена).
3. В разделе «Как вы входите в Google» найдите **Пароли приложений**.
4. Создайте пароль для приложения «Почта» и скопируйте 16-символьный пароль.
5. Вставьте его в `email.password`.

### `backend/config/about/` — контакты и страница «Обо мне»

Папка лежит **рядом с `config.yaml`**: бэкенд ищет её относительно пути к конфигу.

| Файл | Что задаёт | В Git |
|------|------------|-------|
| `contacts.yaml` | e-mail, Instagram, Telegram и имя файла фото | да, при деплое копируется на сервер |
| `bio.txt` | текст страницы «Обо мне» (пустая строка — новый абзац) | да, при деплое копируется на сервер |
| `photo.jpg` | фото для первого экрана и страницы «Обо мне» | нет, кладётся на сервер вручную |

### `webui/config/config.js`

Для локальной разработки — `BASE_URL: "http://localhost:8080/api/v1"`. При деплое workflow сам заменяет его на `"/api/v1"`: в продакшне сайт и API обслуживает один nginx на одном домене.

---

## Swagger и тесты

Swagger-документация генерируется из аннотаций в обработчиках и коммитится в `backend/docs/`:

```bash
cd backend && make swagger
```

Юнит-тесты не требуют базы данных:

```bash
cd backend && make test
```

На каждый пуш и pull request GitHub Actions запускает `.github/workflows/ci.yml`: сборку, тесты бэкенда и проверку синтаксиса JavaScript.

---

## Продакшн: как устроен деплой

Деплой полностью автоматический: каждый пуш в `main` запускает `.github/workflows/deploy.yml`. Повторить деплой без нового коммита можно кнопкой **Run workflow** на вкладке **Actions**.

1. **Тесты** — `go test ./...`; если они падают, деплой не начинается.
2. **Сборка** — статически слинкованный бинарник `server` (миграции и Swagger встроены в него); копия `webui/` с продакшн-`config.js` и сгенерированным `health.json` (`deploy/build-webui-manifest.sh`).
3. **Загрузка** по SSH (rsync) в `/opt/design-profile/releases/<коммит>/`; содержимое `backend/config/about/` — в `shared/config/about/` (без удаления `photo.jpg`).
4. **Активация** — на сервере выполняется `deploy/activate.sh`:
   - ссылка `current` переключается на новый релиз, `design-backend` перезапускается;
   - проверяется бэкенд (`/health`, включая доступ к PostgreSQL) и сайт через nginx (`check-webui.sh`, см. [Проверки здоровья](#проверки-здоровья));
   - если проверки прошли — хранятся 5 последних релизов, остальные удаляются;
   - если нет — `current` возвращается на предыдущий релиз, бэкенд перезапускается, деплой помечается как упавший.

Два деплоя одновременно не выполняются: следующий ждёт окончания текущего.

### Структура на сервере

```
/opt/design-profile/                    владелец deploy
├── current -> releases/<коммит>        активный релиз (бэкенд и сайт переключаются вместе)
├── releases/                           владелец deploy
│   └── <коммит>/
│       ├── server                      бинарник бэкенда
│       ├── check-webui.sh              проверка сайта
│       └── webui/                      файлы сайта + health.json
└── shared/                             владелец root
    └── config/                         владелец root
        ├── config.yaml                 root:www-data, 640 — секреты, deploy их не видит
        └── about/                      владелец deploy — contacts.yaml, bio.txt, photo.jpg
```

| Кто | Что делает |
|-----|------------|
| `deploy` | пользователь, под которым заходит GitHub Actions; вход только по ключу; может лишь загружать релизы и перезапускать сервис |
| `www-data` | от его имени работают бэкенд (`design-backend`) и nginx |

---

## Первичная настройка сервера

Выполняется один раз под администратором с `sudo` (Ubuntu / Debian).

### 1. Пакеты

```bash
sudo apt update && sudo apt install -y nginx postgresql rsync curl
```

### 2. База данных

```bash
sudo -u postgres psql <<'SQL'
CREATE USER design_user WITH PASSWORD 'новый_надёжный_пароль';
CREATE DATABASE design_portfolio OWNER design_user;
SQL
```

### 3. Пользователь `deploy` и ключ для GitHub

Ключ генерируется **на вашем компьютере**, а не на сервере, без пароля (GitHub Actions не сможет его ввести):

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/design_profile_deploy -N ""
```

На сервере создайте пользователя без пароля и добавьте ему **публичный** ключ (содержимое `design_profile_deploy.pub`):

```bash
sudo adduser --disabled-password --gecos "" deploy
sudo install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
echo 'ssh-ed25519 AAAA… github-actions-deploy' | sudo tee -a /home/deploy/.ssh/authorized_keys
sudo chown deploy:deploy /home/deploy/.ssh/authorized_keys && sudo chmod 600 /home/deploy/.ssh/authorized_keys
```

**Приватный** ключ кладётся только в секрет GitHub (см. [Настройка GitHub](#настройка-github)), не в настройки аккаунта и не в Deploy keys.

### 4. Папки и права

```bash
sudo install -d -o deploy -g deploy -m 755 /opt/design-profile /opt/design-profile/releases
sudo install -d -o root -g root -m 755 /opt/design-profile/shared /opt/design-profile/shared/config
sudo install -d -o deploy -g deploy -m 755 /opt/design-profile/shared/config/about
```

### 5. Конфиг и фото

Скопируйте подготовленный `config.yaml` в домашнюю папку администратора (`scp`), затем:

```bash
sudo install -o root -g www-data -m 640 ~/config.yaml /opt/design-profile/shared/config/config.yaml && rm ~/config.yaml
```

Используйте для сервера **новые** пароли и новый `jwt.secret`, а `server.host` установите в `127.0.0.1`.
Фото положите в `/opt/design-profile/shared/config/about/photo.jpg`.

### 6. Сервис бэкенда

```bash
sudo tee /etc/systemd/system/design-backend.service > /dev/null <<'EOF'
[Unit]
Description=Design Profile Backend
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/design-profile
ExecStart=/opt/design-profile/current/server -config /opt/design-profile/shared/config/config.yaml
Restart=on-failure
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload && sudo systemctl enable design-backend
```

Сервис запустится при первом деплое, когда появится `current/server`.

### 7. Право `deploy` перезапускать сервис

Через `sudo visudo -f /etc/sudoers.d/deploy`:

```
deploy ALL=(root) NOPASSWD: /usr/bin/systemctl restart design-backend
```

### 8. nginx

```bash
sudo tee /etc/nginx/sites-available/design-profile > /dev/null <<'EOF'
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;   # замените на ваш-домен.ru www.ваш-домен.ru

    root /opt/design-profile/current/webui;
    index index.html;

    # Сайт. no-cache: браузер перепроверяет файлы и сразу видит новый релиз.
    location / {
        add_header Cache-Control "no-cache";
        try_files $uri $uri/ /index.html;
    }

    # Проверка сайта (health.json создаётся при деплое).
    location = /health {
        default_type application/json;
        add_header Cache-Control "no-store";
        try_files /health.json =404;
    }

    location /api/ {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        client_max_body_size 210M;   # до 10 фото по 20 МБ в заявке
    }

    location /swagger/ {
        proxy_pass http://127.0.0.1:8080;
    }
}
EOF
sudo ln -sf /etc/nginx/sites-available/design-profile /etc/nginx/sites-enabled/design-profile
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl enable --now nginx && sudo systemctl reload nginx
```

> **Важно:** `listen` не должен быть привязан к конкретному IP (`listen 203.0.113.10:80;`). Проверка при деплое обращается к nginx на `127.0.0.1` и с такой настройкой не сможет подключиться.

### 9. HTTPS (если есть домен)

A-запись домена должна указывать на сервер, а в `server_name` — стоять домен. certbot сам добавит HTTPS и перенаправление с HTTP:

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d ваш-домен.ru -d www.ваш-домен.ru
```

### 10. Файрвол

Если включён `ufw`, откройте HTTP/HTTPS. **Не включайте ufw, не разрешив сначала SSH**, иначе потеряете доступ к серверу:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
```

---

## Настройка GitHub

Репозиторий → **Settings → Environments → `production`** (имя должно совпадать с `environment: production` в workflow). Значения секретов после сохранения не показываются — их можно только заменить.

| Имя | Тип | Значение |
|-----|-----|----------|
| `SSH_HOST` | Secret | IP или домен сервера |
| `SSH_USER` | Secret | `deploy` |
| `SSH_PRIVATE_KEY` | Secret | содержимое `~/.ssh/design_profile_deploy` целиком, со строками `BEGIN`/`END` |
| `SSH_KNOWN_HOSTS` | Secret | вывод `ssh-keyscan -t ed25519 <адрес>`; сверьте отпечаток с панелью хостинга |
| `SITE_URL` | Variable | адрес сайта: `https://ваш-домен.ru` или `http://<IP>`; без неё сайт проверяется по `http://127.0.0.1` |
| `SSH_PORT` | Variable, необязательно | порт SSH, если не 22 |
| `DEPLOY_GOARCH` | Variable, необязательно | `arm64` для сервера на ARM (по умолчанию `amd64`) |

Если в приватном репозитории на бесплатном тарифе раздела Environments нет, те же значения можно завести в **Settings → Secrets and variables → Actions**.

В логах GitHub скрывает (`***`) любой текст, совпадающий со значением секрета. Поскольку `SSH_USER` = `deploy`, путь `deploy/activate.sh` отображается как `***/activate.sh` — это не ошибка.

---

## Проверки здоровья

| Адрес | Что проверяет | Ответ |
|-------|---------------|-------|
| `http://127.0.0.1:8080/health` (на сервере) | бэкенд и доступ к PostgreSQL (ping, таймаут 2 с) | `200 {"status":"ok","checks":{"postgres":"ok"}}` или `503` с `"unavailable"`; причина ошибки — только в журнале сервиса |
| `<сайт>/api/v1/health` | то же, но через nginx | как выше |
| `<сайт>/health` (или `/health.json`) | какой релиз сайта отдаёт nginx | `{"status":"ok","release":"<коммит>","files":{…}}` |

При деплое `check-webui.sh` проверяет сайт через nginx на самом сервере (даже если `SITE_URL` — публичный домен, запросы идут на `127.0.0.1` с правильным именем и сертификатом):

- `health.json` указывает на только что выложенный релиз — то есть `root` в nginx смотрит в `current/webui`;
- каждый файл сайта отдаётся и побайтно совпадает с собранным (SHA-256), `/` отдаёт `index.html`;
- через тот же адрес отвечает `/api/v1/health` — nginx доходит до бэкенда.

Выполнение JavaScript в браузере эта проверка не покрывает.

Проверить сайт можно и вручную, с любого компьютера:

```bash
RESOLVE_TO= bash deploy/check-webui.sh https://ваш-домен.ru
```

---

## Повседневные операции

### Логи

| Что | Где |
|-----|-----|
| Бэкенд | `sudo journalctl -u design-backend -f` (в реальном времени), `-n 100 --no-pager` (последние строки), `-p err` (только ошибки) |
| nginx | `/var/log/nginx/access.log`, `/var/log/nginx/error.log` (502 — nginx не достучался до бэкенда) |
| PostgreSQL | `/var/log/postgresql/postgresql-<версия>-main.log` |
| Деплой | вкладка **Actions** в GitHub |

### Повторить деплой

**Actions** → нужный запуск → **Re-run failed jobs** (тот же коммит) или **Deploy → Run workflow** (последний коммит `main`).

### Откатиться вручную

```bash
ls -1t /opt/design-profile/releases/          # релизы, новые сверху
sudo -u deploy ln -sfn releases/<коммит> /opt/design-profile/current
sudo systemctl restart design-backend
bash /opt/design-profile/current/check-webui.sh "<SITE_URL>" <коммит>
```

Ссылка задаётся относительной (`releases/<коммит>`), как это делает деплой. Следующий пуш в `main` снова выложит последнюю версию.

---

## Частые проблемы

| Ошибка в логе деплоя | Причина | Что сделать |
|---|---|---|
| `rsync: mkdir "/opt/design-profile/releases/…" failed: Permission denied` | у `deploy` нет права записи | владельцем `/opt/design-profile`, `releases` и `shared/config/about` должен быть `deploy` ([шаг 4](#4-папки-и-права)); `SSH_USER` должен быть `deploy` |
| `Unit design-backend.service not found` | сервис не создан | [шаг 6](#6-сервис-бэкенда) |
| `sudo: a password is required` | нет правила sudoers | [шаг 7](#7-право-deploy-перезапускать-сервис) |
| `Backend check failed: … last response: {"…postgres":"unavailable"}` | бэкенд не подключается к базе | `sudo journalctl -u design-backend -n 50`; проверьте раздел `database` в `config.yaml` и `systemctl status postgresql` |
| `Backend check failed: … no response` | бэкенд не стартует | журнал сервиса: ошибка чтения конфига, занятый порт и т. п. |
| `Failed to connect to 127.0.0.1 port 80` / `cannot load …/health.json` | nginx не запущен или `listen` привязан к IP | [шаг 8](#8-nginx); `sudo ss -ltnp \| grep -E ':(80\|443)'` |
| `the site serves release '…', expected '…'` | nginx раздаёт сайт не из `current` | `root /opt/design-profile/current/webui;` |
| `/… differs from release …` | отдаётся другая версия файла (кэширующий прокси, ручная правка на сервере) | уберите кэширование/правку и повторите деплой |
| `Host key verification failed` | `SSH_KNOWN_HOSTS` не задан или ключ сервера сменился | обновите секрет выводом `ssh-keyscan`, сверив отпечаток |

---

## Docker (необязательно)

В продакшн-деплое Docker не используется; `backend/docker/Dockerfile` пригоден для локальных экспериментов. Его сборочный образ (`golang:1.22`) старше версии Go в `go.mod` (1.23) — перед использованием его стоит обновить.

```bash
cd backend
docker build -t design-profile-backend -f docker/Dockerfile .
docker run -d --name design-backend -p 8080:8080 \
  -v "$(pwd)/config:/app/config" design-profile-backend
```

Конфиг и папка `about/` монтируются снаружи и в образ не попадают.

---

## Структура проекта

```
design-profile/
├── .github/workflows/
│   ├── ci.yml                    # сборка и тесты на каждый пуш
│   └── deploy.yml                # деплой на сервер при пуше в main
├── deploy/
│   ├── activate.sh               # переключение релиза, проверки, откат (выполняется на сервере)
│   ├── build-webui-manifest.sh   # генерация health.json при сборке
│   └── check-webui.sh            # проверка сайта через nginx
├── backend/
│   ├── cmd/main/main.go          # точка входа
│   ├── config/
│   │   ├── config.go             # загрузка конфигурации
│   │   ├── config.yaml.example   # шаблон конфига
│   │   ├── config.yaml           # локальный конфиг с секретами (не коммитится)
│   │   └── about/                # contacts.yaml, bio.txt (+ photo.jpg на сервере)
│   ├── docker/Dockerfile
│   ├── docs/                     # сгенерированный Swagger (make swagger)
│   ├── internal/
│   │   ├── auth/                 # OTP, JWT
│   │   ├── email/                # SMTP-клиент
│   │   ├── handler/              # HTTP-обработчики, в т. ч. /health
│   │   ├── middleware/           # JWT middleware
│   │   ├── model/                # доменные модели
│   │   ├── repository/           # работа с БД
│   │   └── service/              # бизнес-логика
│   ├── migrations/               # SQL-миграции (встроены в бинарник, применяются при старте)
│   ├── go.mod
│   └── Makefile
├── webui/
│   ├── config/config.js          # URL бэкенда (в продакшне заменяется при деплое)
│   ├── index.html                # публичный сайт
│   ├── admin.html                # панель администратора
│   └── src/
│       ├── api.js                # HTTP-клиент
│       ├── main.js               # логика публичного сайта
│       ├── admin.js              # логика админ-панели
│       ├── locales.js            # переводы RU/EN
│       └── styles/main.css
└── start.sh                      # локальный запуск бэкенда и фронтенда
```
