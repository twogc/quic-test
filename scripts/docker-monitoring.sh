#!/bin/bash

# Скрипт для запуска мониторинга 2GC Network Protocol Suite
# Prometheus + Grafana + Jaeger

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Monitoring${NC}"
echo -e "${BLUE}==========================================${NC}"
echo "Запуск стека мониторинга: Prometheus + Grafana + Jaeger"

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Ошибка: Docker не установлен${NC}"
    exit 1
fi

# Функция для отображения помощи
show_help() {
    echo "Использование: $0 [команда] [опции]"
    echo ""
    echo "Команды:"
    echo "  up          - Запустить мониторинг"
    echo "  down        - Остановить мониторинг"
    echo "  restart     - Перезапустить мониторинг"
    echo "  logs        - Показать логи"
    echo "  status      - Показать статус"
    echo "  clean       - Очистить все"
    echo ""
    echo "Опции:"
    echo "  --build     - Пересобрать образы"
    echo "  --detach    - Запустить в фоновом режиме"
    echo "  --follow    - Следовать за логами"
    echo ""
    echo "Примеры:"
    echo "  $0 up --detach"
    echo "  $0 logs --follow"
    echo "  $0 down"
}

# Параметры по умолчанию
BUILD_FLAG=""
DETACH_FLAG=""
FOLLOW_FLAG=""

# Парсим аргументы
while [[ $# -gt 0 ]]; do
    case $1 in
        --build)
            BUILD_FLAG="--build"
            shift
            ;;
        --detach)
            DETACH_FLAG="-d"
            shift
            ;;
        --follow)
            FOLLOW_FLAG="-f"
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            COMMAND=$1
            shift
            ;;
    esac
done

# Выполняем команду
case $COMMAND in
    up)
        echo -e "${YELLOW}Запускаем мониторинг...${NC}"
        docker compose -f docker compose.monitoring.yml up $BUILD_FLAG $DETACH_FLAG
        ;;
    down)
        echo -e "${YELLOW}Останавливаем мониторинг...${NC}"
        docker compose -f docker compose.monitoring.yml down
        ;;
    restart)
        echo -e "${YELLOW}Перезапускаем мониторинг...${NC}"
        docker compose -f docker compose.monitoring.yml restart
        ;;
    logs)
        echo -e "${YELLOW}Показываем логи...${NC}"
        docker compose -f docker compose.monitoring.yml logs $FOLLOW_FLAG
        ;;
    status)
        echo -e "${YELLOW}Статус сервисов:${NC}"
        docker compose -f docker compose.monitoring.yml ps
        ;;
    clean)
        echo -e "${YELLOW}Очищаем все...${NC}"
        docker compose -f docker compose.monitoring.yml down --rmi all --volumes --remove-orphans
        ;;
    *)
        echo -e "${RED}Неизвестная команда: $COMMAND${NC}"
        show_help
        exit 1
        ;;
esac

if [ "$COMMAND" = "up" ] && [ -z "$DETACH_FLAG" ]; then
    echo ""
    echo -e "${GREEN}✅ Мониторинг запущен!${NC}"
    echo ""
    echo -e "${BLUE}Доступные сервисы:${NC}"
    echo "  Prometheus: http://localhost:9090"
    echo "  Grafana: http://localhost:3000 (admin/admin)"
    echo "  Jaeger: http://localhost:16686"
    echo ""
    echo -e "${BLUE}Полезные команды:${NC}"
    echo "  Остановка: $0 down"
    echo "  Логи: $0 logs"
    echo "  Статус: $0 status"
fi
