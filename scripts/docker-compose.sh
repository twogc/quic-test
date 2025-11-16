#!/bin/bash

# Скрипт для запуска полного стека 2GC Network Protocol Suite через Docker Compose

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite${NC}"
echo -e "${BLUE}==========================================${NC}"
echo "Запуск полного стека через Docker Compose"

# Проверяем наличие Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Ошибка: Docker Compose не установлен${NC}"
    exit 1
fi

# Функция для отображения помощи
show_help() {
    echo "Использование: $0 [команда] [опции]"
    echo ""
    echo "Команды:"
    echo "  up          - Запустить все сервисы"
    echo "  down        - Остановить все сервисы"
    echo "  restart     - Перезапустить все сервисы"
    echo "  logs        - Показать логи всех сервисов"
    echo "  status      - Показать статус сервисов"
    echo "  clean       - Очистить все контейнеры и образы"
    echo "  server-only - Запустить только сервер"
    echo "  client-only - Запустить только клиент"
    echo "  dashboard   - Запустить только дашборд"
    echo ""
    echo "Опции:"
    echo "  --build     - Пересобрать образы"
    echo "  --detach    - Запустить в фоновом режиме"
    echo "  --follow    - Следовать за логами"
    echo ""
    echo "Примеры:"
    echo "  $0 up --build"
    echo "  $0 server-only"
    echo "  $0 logs --follow"
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
        echo -e "${YELLOW}Запускаем все сервисы...${NC}"
        docker-compose up $BUILD_FLAG $DETACH_FLAG
        ;;
    down)
        echo -e "${YELLOW}Останавливаем все сервисы...${NC}"
        docker-compose down
        ;;
    restart)
        echo -e "${YELLOW}Перезапускаем все сервисы...${NC}"
        docker-compose restart
        ;;
    logs)
        echo -e "${YELLOW}Показываем логи...${NC}"
        docker-compose logs $FOLLOW_FLAG
        ;;
    status)
        echo -e "${YELLOW}Статус сервисов:${NC}"
        docker-compose ps
        ;;
    clean)
        echo -e "${YELLOW}Очищаем все контейнеры и образы...${NC}"
        docker-compose down --rmi all --volumes --remove-orphans
        ;;
    server-only)
        echo -e "${YELLOW}Запускаем только сервер...${NC}"
        docker-compose up $BUILD_FLAG $DETACH_FLAG quic-server
        ;;
    client-only)
        echo -e "${YELLOW}Запускаем только клиент...${NC}"
        docker-compose up $BUILD_FLAG $DETACH_FLAG quic-client
        ;;
    dashboard)
        echo -e "${YELLOW}Запускаем только дашборд...${NC}"
        docker-compose up $BUILD_FLAG $DETACH_FLAG dashboard
        ;;
    *)
        echo -e "${RED}Неизвестная команда: $COMMAND${NC}"
        show_help
        exit 1
        ;;
esac

if [ "$COMMAND" = "up" ] && [ -z "$DETACH_FLAG" ]; then
    echo ""
    echo -e "${GREEN}✅ Все сервисы запущены!${NC}"
    echo ""
    echo -e "${BLUE}Доступные сервисы:${NC}"
    echo "  QUIC сервер: localhost:9000"
    echo "  Веб-дашборд: http://localhost:9990"
    echo "  Prometheus: http://localhost:9090"
    echo "  Grafana: http://localhost:3000 (admin/admin)"
    echo "  Jaeger: http://localhost:16686"
    echo ""
    echo -e "${BLUE}Полезные команды:${NC}"
    echo "  Остановка: $0 down"
    echo "  Логи: $0 logs"
    echo "  Статус: $0 status"
fi
