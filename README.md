# Elias - Telegram Mini App Game

Игра для объяснения слов в Telegram Mini App. Играй с друзьями, объясняй слова и побеждай!

## Технологии

- **Backend**: Go + Fiber + WebSocket
- **Frontend**: React + TypeScript + Vite
- **Database**: PostgreSQL + Redis
- **Deployment**: Docker + Docker Compose + Traefik

## Структура проекта

```
.
├── backend/           # Go backend
│   ├── cmd/          # Entry points
│   ├── internal/     # Internal packages
│   └── seeds/        # Database seeds
├── frontend/         # React frontend
│   ├── src/          # Source files
│   └── public/       # Static files
├── .github/          # GitHub Actions CI/CD
└── docker-compose.*.yml
```

## Локальная разработка

### Требования

- Go 1.21+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 15
- Redis 7

### Установка

1. Клонируй репозиторий:
```bash
git clone git@github.com:YarikYar/alias.git
cd alias
```

2. Скопируй .env файл:
```bash
cp .env.example .env
```

3. Настрой переменные окружения в `.env`:
```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=elias
DB_PASSWORD=your_password
DB_NAME=elias

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Telegram Bot
BOT_TOKEN=your_bot_token

# API
API_URL=http://localhost:8080
```

4. Запусти в режиме разработки:
```bash
docker compose up -d
```

5. Frontend доступен на `http://localhost:5173`
6. Backend API на `http://localhost:8080`

### Разработка без Docker

#### Backend
```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

## Production деплой

### Требования на сервере

- Docker & Docker Compose
- Traefik (опционально, для reverse proxy)
- SSH доступ

### Настройка CI/CD

1. Добавь секреты в GitHub:
   - `SSH_PRIVATE_KEY` - SSH ключ для доступа к серверу
   - `SERVER_HOST` - IP или домен сервера
   - `SERVER_USER` - пользователь SSH
   - `DEPLOY_PATH` - путь к проекту на сервере

2. На сервере создай `.env` файл в `DEPLOY_PATH`:
```bash
mkdir -p /path/to/deploy
cd /path/to/deploy
nano .env
```

3. Настрой production переменные:
```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=elias
DB_PASSWORD=strong_password_here
DB_NAME=elias

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Telegram Bot
BOT_TOKEN=your_production_bot_token

# API
API_URL=https://api.yourdomain.com
VITE_API_URL=https://api.yourdomain.com

# Traefik (если используется)
TRAEFIK_NETWORK=traefik_default
```

4. Сделай deploy.sh исполняемым:
```bash
chmod +x deploy.sh
```

### Автоматический деплой

Push в ветку `main` автоматически запустит деплой через GitHub Actions:
```bash
git push origin main
```

### Ручной деплой

На сервере:
```bash
cd /path/to/deploy
git pull origin main
./deploy.sh
```

## GitHub Secrets

Добавь следующие секреты в настройках репозитория:

| Секрет | Описание | Пример |
|--------|----------|--------|
| `SSH_PRIVATE_KEY` | Приватный SSH ключ | `-----BEGIN RSA PRIVATE KEY-----...` |
| `SERVER_HOST` | IP или домен сервера | `123.45.67.89` или `server.example.com` |
| `SERVER_USER` | SSH пользователь | `ubuntu` или `root` |
| `DEPLOY_PATH` | Путь к проекту на сервере | `/home/ubuntu/alias` |

### Генерация SSH ключа

```bash
# На локальной машине
ssh-keygen -t rsa -b 4096 -C "github-actions" -f ~/.ssh/alias_deploy

# Скопируй публичный ключ на сервер
ssh-copy-id -i ~/.ssh/alias_deploy.pub user@server

# Добавь приватный ключ в GitHub Secrets
cat ~/.ssh/alias_deploy
```

## Архитектура

### Backend (Go)

```
backend/
├── cmd/server/           # Main application
├── internal/
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic
│   ├── models/          # Data models
│   ├── ws/             # WebSocket hub
│   └── telegram/        # Telegram bot
```

### Frontend (React)

```
frontend/
├── src/
│   ├── pages/          # Page components
│   ├── components/     # Reusable components
│   ├── stores/         # Zustand stores
│   ├── hooks/          # Custom hooks
│   ├── lib/           # Utilities
│   └── types/         # TypeScript types
```

## API Endpoints

### REST API

- `POST /api/rooms` - Создать комнату
- `GET /api/rooms/:id` - Получить комнату
- `POST /api/rooms/:id/join` - Присоединиться к комнате
- `POST /api/rooms/:id/team` - Сменить команду
- `POST /api/rooms/:id/start` - Начать игру
- `GET /api/rooms/:id/stats` - Статистика игры

### WebSocket

- `/ws/:roomId` - WebSocket соединение для игры

#### WebSocket события

**От сервера:**
- `player_joined` - Игрок присоединился
- `player_left` - Игрок вышел
- `team_changed` - Игрок сменил команду
- `game_started` - Игра началась
- `new_word` - Новое слово
- `word_result` - Результат слова
- `timer` - Обновление таймера
- `round_end` - Конец раунда
- `game_end` - Конец игры
- `score_update` - Обновление счета

**От клиента:**
- `swipe` - Свайп (up/down)

## База данных

### Миграции

Миграции выполняются автоматически при старте backend.

### Таблицы

- `users` - Пользователи Telegram
- `rooms` - Игровые комнаты
- `players` - Игроки в комнатах
- `words` - Слова для игры
- `game_states` - Состояние игр
- `word_attempts` - Попытки отгадывания

## Тематики слов

- 🎯 Общие
- 🦁 Животные
- 🍕 Еда
- 🌍 Страны и города
- 👨‍💼 Профессии

## Мониторинг

### Логи

```bash
# Все сервисы
docker compose -f docker-compose.prod.yml logs -f

# Только backend
docker compose -f docker-compose.prod.yml logs -f backend

# Только frontend
docker compose -f docker-compose.prod.yml logs -f frontend
```

### Статус контейнеров

```bash
docker compose -f docker-compose.prod.yml ps
```

### Проверка здоровья

```bash
# Backend health
curl http://localhost:8080/health

# Database
docker exec -it alias-postgres-1 psql -U elias -d elias -c "SELECT COUNT(*) FROM words;"

# Redis
docker exec -it alias-redis-1 redis-cli PING
```

## Troubleshooting

### Backend не стартует

1. Проверь логи: `docker logs alias-backend-1`
2. Проверь подключение к БД
3. Проверь переменные окружения

### Frontend не загружается

1. Проверь `VITE_API_URL` в `.env`
2. Проверь Nginx конфигурацию
3. Rebuild: `docker compose build frontend`

### WebSocket не подключается

1. Проверь CORS настройки в backend
2. Проверь Traefik конфигурацию (если используется)
3. Проверь firewall на сервере

### База данных

```bash
# Подключиться к БД
docker exec -it alias-postgres-1 psql -U elias -d elias

# Бэкап
docker exec alias-postgres-1 pg_dump -U elias elias > backup.sql

# Восстановление
docker exec -i alias-postgres-1 psql -U elias elias < backup.sql
```

## Разработка

### Добавление новых слов

```bash
docker exec -i alias-postgres-1 psql -U elias -d elias << 'EOF'
INSERT INTO words (word, lang, category) VALUES
('новое слово', 'ru', 'general');
EOF
```

### Hot reload

- Backend: используй `air` для hot reload
- Frontend: Vite автоматически перезагружает

## Production checklist

- [ ] Обновлены все пароли в `.env`
- [ ] Настроен SSL (Traefik + Let's Encrypt)
- [ ] Настроен firewall
- [ ] Настроены бэкапы БД
- [ ] Добавлен мониторинг
- [ ] Настроены GitHub Secrets
- [ ] Протестирован CI/CD пайплайн

## Лицензия

MIT

## Поддержка

Для вопросов и предложений: [GitHub Issues](https://github.com/YarikYar/alias/issues)
