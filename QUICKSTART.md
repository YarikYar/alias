# 🚀 Быстрый старт Elias (Alias Game)

Telegram Mini App игра в слова с поддержкой 2-5 команд и смешными случайными названиями команд.

## Требования

- Docker и Docker Compose
- Telegram Bot Token (получить у @BotFather)
- Для production: домен с настроенным DNS

## Установка

### 1. Интерактивное развертывание (рекомендуется)

```bash
./deploy-interactive.sh
```

Скрипт проведет вас через все этапы:
- Выбор режима (Production/Development)
- Ввод конфигурации (домен, токен бота, пароли)
- Автоматическая настройка и запуск

### 2. Ручное развертывание

#### Production (с доменом и SSL)

```bash
# 1. Создать .env
cat > .env << EOF
TELEGRAM_BOT_TOKEN=your_bot_token_here
APP_URL=https://alias.zaruchevskiy.ru
POSTGRES_USER=elias
POSTGRES_PASSWORD=secure_password_here
POSTGRES_DB=elias
EOF

# 2. Создать frontend/.env
cat > frontend/.env << EOF
VITE_API_URL=https://alias.zaruchevskiy.ru
VITE_WS_URL=wss://alias.zaruchevskiy.ru
VITE_BOT_USERNAME=your_bot_name
EOF

# 3. Запустить
docker compose up -d --build

# 4. Проверить статус
docker compose ps
docker compose logs -f
```

#### Development (локальная разработка)

```bash
# 1. Создать .env
cat > .env << EOF
TELEGRAM_BOT_TOKEN=your_bot_token_here
APP_URL=http://desktop.lan:8082
POSTGRES_USER=elias
POSTGRES_PASSWORD=elias_secret_password
POSTGRES_DB=elias
EOF

# 2. Создать frontend/.env
cat > frontend/.env << EOF
VITE_API_URL=http://desktop.lan:8082
VITE_WS_URL=ws://desktop.lan:8082
VITE_BOT_USERNAME=elias_local_bot
EOF

# 3. Изменить docker-compose.yml на development режим
# (убрать Traefik, добавить прямые порты)

# 4. Запустить
docker compose up -d --build
```

## Настройка Telegram бота

После развертывания настройте Mini App в BotFather:

1. Отправьте `/mybots` в @BotFather
2. Выберите своего бота
3. **Bot Settings** → **Menu Button** → **Configure menu button**
4. Введите URL:
   - Production: `https://your-domain.com`
   - Development: `http://desktop.lan:3001`

## Архитектура

```
┌─────────────────┐
│  Telegram Bot   │
└────────┬────────┘
         │
┌────────▼─────────────────────────────────┐
│          Traefik (Production)            │
│     HTTP → HTTPS redirect + SSL          │
└────────┬─────────────────────────────────┘
         │
    ┌────┴────┐
    │         │
┌───▼──┐  ┌──▼──────┐
│ React│  │ Go Fiber│
│ UI   │  │ Backend │
└──────┘  └────┬────┘
               │
          ┌────┴────┐
          │         │
      ┌───▼───┐ ┌──▼───┐
      │Postgre│ │Redis │
      │  SQL  │ │      │
      └───────┘ └──────┘
```

## Особенности

✨ **Смешные названия команд** - автоматически генерируются русские названия с согласованием по родам (например, "Быстрый Бегемот", "Хитрая Панда")

🎮 **2-5 команд** - поддержка от 2 до 5 команд в одной игре

🔄 **WebSocket** - реальное время для всех действий

🎨 **Telegram Mini App** - нативная интеграция с Telegram

🔒 **SSL** - автоматические сертификаты через Let's Encrypt

## Полезные команды

```bash
# Логи
docker compose logs -f              # Все логи
docker compose logs -f backend      # Только backend
docker compose logs -f frontend     # Только frontend

# Управление
docker compose ps                    # Статус
docker compose restart               # Перезапуск
docker compose restart backend       # Перезапуск только backend
docker compose down                  # Остановка
docker compose up -d --build        # Пересборка и запуск

# База данных
docker compose exec postgres psql -U elias -d elias

# Очистка
docker compose down -v              # Удалить с volumes
docker system prune -a              # Очистить Docker
```

## Миграции

Миграции находятся в `backend/migrations/` и применяются автоматически при первом запуске PostgreSQL.

Ручное применение миграции:
```bash
docker compose exec postgres psql -U elias -d elias -f /docker-entrypoint-initdb.d/001_initial.sql
```

## Тестирование

```bash
# Backend тесты
cd backend
go test ./internal/services/... -v

# Frontend сборка
cd frontend
npm run build
```

## Структура проекта

```
.
├── backend/
│   ├── cmd/server/          # Точка входа
│   ├── internal/
│   │   ├── handlers/        # HTTP/WebSocket handlers
│   │   ├── middleware/      # Auth middleware
│   │   ├── models/          # Data models
│   │   └── services/        # Business logic
│   └── migrations/          # SQL миграции
├── frontend/
│   ├── src/
│   │   ├── components/      # React компоненты
│   │   ├── pages/           # Страницы
│   │   ├── stores/          # Zustand state
│   │   └── lib/             # Utilities
│   └── public/
├── docker-compose.yml       # Docker конфигурация
└── deploy-interactive.sh    # Интерактивное развертывание
```

## Troubleshooting

### Порт уже используется
```bash
# Найти процесс
lsof -i :8082

# Остановить
kill -9 <PID>
```

### PostgreSQL не стартует
```bash
# Проверить логи
docker compose logs postgres

# Удалить volume и пересоздать
docker compose down -v
docker compose up -d
```

### Frontend не собирается
```bash
# Очистить node_modules
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### SSL не работает
- Проверьте что DNS указывает на сервер: `dig your-domain.com`
- Проверьте порты 80 и 443 открыты
- Проверьте логи Traefik: `docker compose logs traefik`

## Поддержка

Вопросы и issue: https://github.com/YarikYar/alias/issues

---

**Приятной игры! 🎮**
