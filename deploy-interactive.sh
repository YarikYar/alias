#!/bin/bash

# Интерактивный скрипт развертывания Elias (Alias game)
# Автор: Claude
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Функции для красивого вывода
print_header() {
    echo -e "\n${BOLD}${MAGENTA}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${MAGENTA}║${NC}  ${CYAN}$1${NC}"
    echo -e "${BOLD}${MAGENTA}╚════════════════════════════════════════════════════════════╝${NC}\n"
}

print_step() {
    echo -e "${BOLD}${BLUE}▶${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

# Функция для запроса ввода с значением по умолчанию
ask() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"
    local secret="$4"

    if [ -n "$default" ]; then
        echo -ne "${BOLD}${prompt}${NC} ${YELLOW}[${default}]${NC}: "
    else
        echo -ne "${BOLD}${prompt}${NC}: "
    fi

    if [ "$secret" = "true" ]; then
        read -s value
        echo
    else
        read value
    fi

    if [ -z "$value" ]; then
        value="$default"
    fi

    eval "$var_name='$value'"
}

ask_yes_no() {
    local prompt="$1"
    local default="$2"

    if [ "$default" = "y" ]; then
        echo -ne "${BOLD}${prompt}${NC} ${YELLOW}[Y/n]${NC}: "
    else
        echo -ne "${BOLD}${prompt}${NC} ${YELLOW}[y/N]${NC}: "
    fi

    read answer

    if [ -z "$answer" ]; then
        answer="$default"
    fi

    case "$answer" in
        [Yy]* ) return 0;;
        * ) return 1;;
    esac
}

# Проверка зависимостей
check_dependencies() {
    print_step "Проверка зависимостей..."

    local missing_deps=()

    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi

    if ! docker compose version &> /dev/null; then
        missing_deps+=("docker-compose")
    fi

    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Не установлены необходимые зависимости: ${missing_deps[*]}"
        echo -e "\nУстановите их и повторите попытку."
        exit 1
    fi

    print_success "Все зависимости установлены"
}

# Заголовок
clear
echo -e "${BOLD}${CYAN}"
cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║   ███████╗██╗     ██╗ █████╗ ███████╗                        ║
║   ██╔════╝██║     ██║██╔══██╗██╔════╝                        ║
║   █████╗  ██║     ██║███████║███████╗                        ║
║   ██╔══╝  ██║     ██║██╔══██║╚════██║                        ║
║   ███████╗███████╗██║██║  ██║███████║                        ║
║   ╚══════╝╚══════╝╚═╝╚═╝  ╚═╝╚══════╝                        ║
║                                                               ║
║        Интерактивный скрипт развертывания игры                ║
║                 Telegram Mini App                             ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}\n"

print_info "Этот скрипт поможет развернуть игру Elias (Alias)"
print_info "Вам потребуется ответить на несколько вопросов\n"

if ! ask_yes_no "Продолжить?" "y"; then
    echo -e "\n${YELLOW}Развертывание отменено${NC}"
    exit 0
fi

# Проверка зависимостей
check_dependencies

# Определение режима развертывания
print_header "Режим развертывания"

echo -e "Выберите режим развертывания:\n"
echo -e "  ${BOLD}1)${NC} Production (с доменом и SSL)"
echo -e "  ${BOLD}2)${NC} Development (локальная разработка)"
echo ""

while true; do
    echo -ne "${BOLD}Выберите режим${NC} ${YELLOW}[1/2]${NC}: "
    read mode_choice

    case "$mode_choice" in
        1)
            MODE="production"
            break
            ;;
        2)
            MODE="development"
            break
            ;;
        *)
            print_error "Выберите 1 или 2"
            ;;
    esac
done

# Настройка Production
if [ "$MODE" = "production" ]; then
    print_header "Настройка Production"

    print_info "Для production развертывания требуется:"
    print_info "  • Доменное имя"
    print_info "  • Telegram Bot Token"
    print_info "  • Email для Let's Encrypt SSL сертификата"
    echo ""

    ask "Введите домен (например, alias.example.com)" "" DOMAIN
    ask "Введите Telegram Bot Token" "" BOT_TOKEN "true"
    ask "Введите email для SSL сертификата" "" SSL_EMAIL
    ask "Введите пароль для PostgreSQL" "elias_secret_password" POSTGRES_PASSWORD "true"
    ask "Имя бота (для frontend)" "elias_bot" BOT_USERNAME

    APP_URL="https://${DOMAIN}"
    API_URL="https://${DOMAIN}"
    WS_URL="wss://${DOMAIN}"

    # Создание .env
    print_step "Создание конфигурационных файлов..."

    cat > .env << EOF
# Production Configuration
TELEGRAM_BOT_TOKEN=${BOT_TOKEN}
APP_URL=${APP_URL}

# Database
POSTGRES_USER=elias
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
POSTGRES_DB=elias

# SSL
SSL_EMAIL=${SSL_EMAIL}
DOMAIN=${DOMAIN}
EOF

    cat > frontend/.env << EOF
VITE_API_URL=${API_URL}
VITE_WS_URL=${WS_URL}
VITE_BOT_USERNAME=${BOT_USERNAME}
EOF

    # Обновление docker-compose.yml для production
    COMPOSE_FILE="docker-compose.yml"

    print_success "Конфигурация создана"

else
    # Настройка Development
    print_header "Настройка Development"

    print_info "Для локальной разработки используются упрощенные настройки"
    echo ""

    ask "Введите Telegram Bot Token" "" BOT_TOKEN "true"
    ask "Введите пароль для PostgreSQL" "elias_secret_password" POSTGRES_PASSWORD "true"
    ask "Имя бота (для frontend)" "elias_local_bot" BOT_USERNAME
    ask "Хост для доступа" "desktop.lan" DEV_HOST

    APP_URL="http://${DEV_HOST}:8082"
    API_URL="http://${DEV_HOST}:8082"
    WS_URL="ws://${DEV_HOST}:8082"

    # Создание .env
    print_step "Создание конфигурационных файлов..."

    cat > .env << EOF
# Development Configuration
TELEGRAM_BOT_TOKEN=${BOT_TOKEN}
APP_URL=${APP_URL}

# Database
POSTGRES_USER=elias
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
POSTGRES_DB=elias
EOF

    cat > frontend/.env << EOF
VITE_API_URL=${API_URL}
VITE_WS_URL=${WS_URL}
VITE_BOT_USERNAME=${BOT_USERNAME}
EOF

    COMPOSE_FILE="docker-compose.yml"

    print_success "Конфигурация создана"
fi

# Проверка портов
print_header "Проверка портов"

if [ "$MODE" = "production" ]; then
    PORTS_TO_CHECK="80 443"
else
    PORTS_TO_CHECK="8082 3001 5433 6380"
fi

print_step "Проверка доступности портов..."

for port in $PORTS_TO_CHECK; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        print_warning "Порт $port занят"
        if ask_yes_no "Остановить процесс на порту $port?" "n"; then
            sudo kill -9 $(lsof -t -i:$port) 2>/dev/null || true
            print_success "Процесс остановлен"
        fi
    else
        print_success "Порт $port свободен"
    fi
done

# Docker Compose
print_header "Запуск контейнеров"

print_step "Остановка старых контейнеров..."
docker compose down 2>/dev/null || true
print_success "Старые контейнеры остановлены"

print_step "Сборка и запуск контейнеров..."
if docker compose up -d --build; then
    print_success "Контейнеры запущены"
else
    print_error "Ошибка при запуске контейнеров"
    exit 1
fi

# Ожидание запуска БД
print_step "Ожидание запуска PostgreSQL..."
sleep 5

max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if docker compose exec -T postgres pg_isready -U elias >/dev/null 2>&1; then
        print_success "PostgreSQL готов"
        break
    fi
    attempt=$((attempt + 1))
    echo -ne "\r  Попытка $attempt/$max_attempts..."
    sleep 1
done

if [ $attempt -eq $max_attempts ]; then
    print_error "Не удалось подключиться к PostgreSQL"
    exit 1
fi

# Применение миграций
print_header "Применение миграций"

print_step "Проверка применения миграций..."

# Миграции применяются автоматически через init скрипты PostgreSQL
# Но на всякий случай проверим и применим вручную если нужно

MIGRATION_DIR="./backend/migrations"
if [ -d "$MIGRATION_DIR" ]; then
    for migration in "$MIGRATION_DIR"/*.sql; do
        if [ -f "$migration" ]; then
            filename=$(basename "$migration")
            print_info "Применение миграции: $filename"
            docker compose exec -T postgres psql -U elias -d elias -f - < "$migration" 2>&1 | grep -v "already exists" | grep -v "duplicate" || true
        fi
    done
    print_success "Миграции применены"
else
    print_warning "Директория с миграциями не найдена"
fi

# Проверка статуса
print_header "Проверка статуса"

print_step "Статус контейнеров:"
docker compose ps

echo ""
print_step "Проверка логов backend..."
sleep 2
if docker compose logs backend | grep -q "Server started"; then
    print_success "Backend запущен успешно"
else
    print_warning "Backend еще запускается, проверьте логи: docker compose logs backend"
fi

# Итоговая информация
print_header "Развертывание завершено!"

if [ "$MODE" = "production" ]; then
    echo -e "${BOLD}🌐 URL приложения:${NC} ${GREEN}${APP_URL}${NC}"
    echo -e "${BOLD}🔒 SSL:${NC} ${GREEN}Автоматически через Let's Encrypt${NC}"
    echo ""
    print_warning "Убедитесь что DNS запись для ${DOMAIN} указывает на этот сервер!"
    echo ""
    print_info "Настройте Mini App в BotFather:"
    echo -e "  1. Отправьте ${CYAN}/mybots${NC} в @BotFather"
    echo -e "  2. Выберите своего бота"
    echo -e "  3. Bot Settings → Menu Button → Configure menu button"
    echo -e "  4. URL: ${GREEN}${APP_URL}${NC}"
else
    echo -e "${BOLD}🌐 Frontend:${NC} ${GREEN}http://${DEV_HOST}:3001${NC}"
    echo -e "${BOLD}🔌 Backend API:${NC} ${GREEN}http://${DEV_HOST}:8082${NC}"
    echo -e "${BOLD}💾 PostgreSQL:${NC} ${GREEN}localhost:5433${NC}"
    echo -e "${BOLD}📦 Redis:${NC} ${GREEN}localhost:6380${NC}"
    echo ""
    print_info "Настройте Mini App в BotFather:"
    echo -e "  1. Отправьте ${CYAN}/mybots${NC} в @BotFather"
    echo -e "  2. Выберите своего бота"
    echo -e "  3. Bot Settings → Menu Button → Configure menu button"
    echo -e "  4. URL: ${GREEN}http://${DEV_HOST}:3001${NC}"
fi

echo ""
print_info "Полезные команды:"
echo -e "  ${CYAN}docker compose logs -f${NC}          - Смотреть все логи"
echo -e "  ${CYAN}docker compose logs -f backend${NC}  - Смотреть логи backend"
echo -e "  ${CYAN}docker compose ps${NC}                - Статус контейнеров"
echo -e "  ${CYAN}docker compose restart${NC}           - Перезапустить все"
echo -e "  ${CYAN}docker compose down${NC}              - Остановить все"

if [ "$MODE" = "production" ]; then
    echo -e "  ${CYAN}docker compose exec postgres psql -U elias -d elias${NC} - Подключиться к БД"
fi

echo ""
echo -e "${BOLD}${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${GREEN}║                                                           ║${NC}"
echo -e "${BOLD}${GREEN}║            🎮 Приложение успешно развернуто! 🎮            ║${NC}"
echo -e "${BOLD}${GREEN}║                                                           ║${NC}"
echo -e "${BOLD}${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
