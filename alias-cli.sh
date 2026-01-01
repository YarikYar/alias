#!/bin/bash

# Alias CLI - Быстрые команды для управления приложением
# Использование: ./alias-cli.sh [команда]

set -e

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

show_help() {
    cat << EOF
${CYAN}Alias CLI - Управление приложением Elias${NC}

${YELLOW}Использование:${NC}
  ./alias-cli.sh [команда]

${YELLOW}Команды:${NC}

  ${GREEN}start${NC}           - Запустить приложение
  ${GREEN}stop${NC}            - Остановить приложение
  ${GREEN}restart${NC}         - Перезапустить приложение
  ${GREEN}rebuild${NC}         - Пересобрать и запустить
  ${GREEN}logs${NC}            - Показать логи (все)
  ${GREEN}logs-backend${NC}    - Показать логи backend
  ${GREEN}logs-frontend${NC}   - Показать логи frontend
  ${GREEN}status${NC}          - Статус контейнеров
  ${GREEN}db${NC}              - Подключиться к PostgreSQL
  ${GREEN}migrate${NC}         - Применить миграции
  ${GREEN}clean${NC}           - Очистить все (включая volumes)
  ${GREEN}test${NC}            - Запустить тесты
  ${GREEN}help${NC}            - Показать эту справку

${YELLOW}Примеры:${NC}
  ./alias-cli.sh start
  ./alias-cli.sh logs-backend
  ./alias-cli.sh db

EOF
}

cmd_start() {
    echo -e "${CYAN}▶ Запуск приложения...${NC}"
    docker compose up -d
    echo -e "${GREEN}✓ Приложение запущено${NC}"
    cmd_status
}

cmd_stop() {
    echo -e "${CYAN}▶ Остановка приложения...${NC}"
    docker compose down
    echo -e "${GREEN}✓ Приложение остановлено${NC}"
}

cmd_restart() {
    echo -e "${CYAN}▶ Перезапуск приложения...${NC}"
    docker compose restart
    echo -e "${GREEN}✓ Приложение перезапущено${NC}"
    cmd_status
}

cmd_rebuild() {
    echo -e "${CYAN}▶ Пересборка и запуск...${NC}"
    docker compose down
    docker compose up -d --build
    echo -e "${GREEN}✓ Приложение пересобрано и запущено${NC}"
    sleep 3
    cmd_status
}

cmd_logs() {
    echo -e "${CYAN}▶ Логи (Ctrl+C для выхода)${NC}"
    docker compose logs -f
}

cmd_logs_backend() {
    echo -e "${CYAN}▶ Логи backend (Ctrl+C для выхода)${NC}"
    docker compose logs -f backend
}

cmd_logs_frontend() {
    echo -e "${CYAN}▶ Логи frontend (Ctrl+C для выхода)${NC}"
    docker compose logs -f frontend
}

cmd_status() {
    echo -e "${CYAN}▶ Статус контейнеров:${NC}"
    docker compose ps
}

cmd_db() {
    echo -e "${CYAN}▶ Подключение к PostgreSQL...${NC}"
    docker compose exec postgres psql -U elias -d elias
}

cmd_migrate() {
    echo -e "${CYAN}▶ Применение миграций...${NC}"

    MIGRATION_DIR="./backend/migrations"
    if [ ! -d "$MIGRATION_DIR" ]; then
        echo -e "${RED}✗ Директория migrations не найдена${NC}"
        exit 1
    fi

    for migration in "$MIGRATION_DIR"/*.sql; do
        if [ -f "$migration" ]; then
            filename=$(basename "$migration")
            echo -e "  Применение: ${YELLOW}$filename${NC}"
            docker compose exec -T postgres psql -U elias -d elias -f - < "$migration" 2>&1 | grep -v "already exists" | grep -v "duplicate" || true
        fi
    done

    echo -e "${GREEN}✓ Миграции применены${NC}"
}

cmd_clean() {
    echo -e "${YELLOW}⚠ ВНИМАНИЕ: Это удалит все данные!${NC}"
    echo -ne "Продолжить? [y/N]: "
    read answer

    if [[ "$answer" =~ ^[Yy]$ ]]; then
        echo -e "${CYAN}▶ Очистка...${NC}"
        docker compose down -v
        echo -e "${GREEN}✓ Все очищено${NC}"
    else
        echo -e "${YELLOW}Отменено${NC}"
    fi
}

cmd_test() {
    echo -e "${CYAN}▶ Запуск тестов...${NC}"

    echo -e "${YELLOW}Backend тесты:${NC}"
    docker compose exec backend go test ./internal/services/... -v

    echo -e "\n${YELLOW}Frontend сборка:${NC}"
    docker compose exec frontend npm run build

    echo -e "${GREEN}✓ Тесты завершены${NC}"
}

# Главная логика
case "${1:-help}" in
    start)
        cmd_start
        ;;
    stop)
        cmd_stop
        ;;
    restart)
        cmd_restart
        ;;
    rebuild)
        cmd_rebuild
        ;;
    logs)
        cmd_logs
        ;;
    logs-backend)
        cmd_logs_backend
        ;;
    logs-frontend)
        cmd_logs_frontend
        ;;
    status)
        cmd_status
        ;;
    db)
        cmd_db
        ;;
    migrate)
        cmd_migrate
        ;;
    clean)
        cmd_clean
        ;;
    test)
        cmd_test
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}✗ Неизвестная команда: $1${NC}\n"
        show_help
        exit 1
        ;;
esac
