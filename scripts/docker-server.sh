#!/bin/bash

# Скрипт для запуска 2GC Network Protocol Suite сервера в Docker

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Server${NC}"
echo -e "${BLUE}==========================================${NC}"
echo "Запуск QUIC сервера в Docker контейнере"

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Ошибка: Docker не установлен${NC}"
    exit 1
fi

# Параметры по умолчанию
SERVER_ADDR=${SERVER_ADDR:-":9000"}
PROMETHEUS_PORT=${PROMETHEUS_PORT:-2113}
PPROF_ADDR=${PPROF_ADDR:-":6060"}
CONTAINER_NAME=${CONTAINER_NAME:-"2gc-network-server"}

echo -e "${YELLOW}Параметры запуска:${NC}"
echo "  Сервер: $SERVER_ADDR"
echo "  Prometheus: $PROMETHEUS_PORT"
echo "  pprof: $PPROF_ADDR"
echo "  Контейнер: $CONTAINER_NAME"

# Останавливаем существующий контейнер, если есть
if docker ps -q -f name=$CONTAINER_NAME | grep -q .; then
    echo -e "${YELLOW}Останавливаем существующий контейнер...${NC}"
    docker stop $CONTAINER_NAME
fi

if docker ps -aq -f name=$CONTAINER_NAME | grep -q .; then
    echo -e "${YELLOW}Удаляем существующий контейнер...${NC}"
    docker rm $CONTAINER_NAME
fi

# Собираем образ
echo -e "${YELLOW}Собираем Docker образ...${NC}"
docker build -f Dockerfile.server -t 2gc-network-suite:server .

# Запускаем контейнер
echo -e "${YELLOW}Запускаем сервер...${NC}"
docker run -d \
    --name $CONTAINER_NAME \
    --network host \
    -p 9000:9000 \
    -p $PROMETHEUS_PORT:$PROMETHEUS_PORT \
    -p 6060:6060 \
    -e QUIC_SERVER_ADDR=$SERVER_ADDR \
    -e QUIC_PROMETHEUS_SERVER_PORT=$PROMETHEUS_PORT \
    -e QUIC_PPROF_ADDR=$PPROF_ADDR \
    2gc-network-suite:server

# Проверяем статус
echo -e "${YELLOW}Проверяем статус контейнера...${NC}"
sleep 2

if docker ps -f name=$CONTAINER_NAME | grep -q $CONTAINER_NAME; then
    echo -e "${GREEN}✅ Сервер успешно запущен!${NC}"
    echo ""
    echo -e "${BLUE}Доступные порты:${NC}"
    echo "  QUIC сервер: localhost:9000"
    echo "  Prometheus метрики: http://localhost:$PROMETHEUS_PORT/metrics"
    echo "  pprof профилирование: http://localhost:6060/debug/pprof/"
    echo ""
    echo -e "${BLUE}Полезные команды:${NC}"
    echo "  Просмотр логов: docker logs -f $CONTAINER_NAME"
    echo "  Остановка: docker stop $CONTAINER_NAME"
    echo "  Удаление: docker rm $CONTAINER_NAME"
    echo "  Вход в контейнер: docker exec -it $CONTAINER_NAME sh"
else
    echo -e "${RED}❌ Ошибка запуска сервера${NC}"
    echo -e "${YELLOW}Логи контейнера:${NC}"
    docker logs $CONTAINER_NAME
    exit 1
fi
