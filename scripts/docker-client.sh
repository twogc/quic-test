#!/bin/bash

# Скрипт для запуска 2GC Network Protocol Suite клиента в Docker

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Client${NC}"
echo -e "${BLUE}==========================================${NC}"
echo "Запуск QUIC клиента в Docker контейнере"

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Ошибка: Docker не установлен${NC}"
    exit 1
fi

# Параметры по умолчанию
SERVER_ADDR=${SERVER_ADDR:-"localhost:9000"}
CONNECTIONS=${CONNECTIONS:-2}
STREAMS=${STREAMS:-4}
RATE=${RATE:-100}
DURATION=${DURATION:-60s}
PROMETHEUS_PORT=${PROMETHEUS_PORT:-2112}
CONTAINER_NAME=${CONTAINER_NAME:-"2gc-network-client"}

echo -e "${YELLOW}Параметры запуска:${NC}"
echo "  Сервер: $SERVER_ADDR"
echo "  Соединения: $CONNECTIONS"
echo "  Потоки: $STREAMS"
echo "  Скорость: $RATE пакетов/сек"
echo "  Длительность: $DURATION"
echo "  Prometheus: $PROMETHEUS_PORT"
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
docker build -f Dockerfile.client -t 2gc-network-suite:client .

# Запускаем контейнер
echo -e "${YELLOW}Запускаем клиент...${NC}"
docker run -d \
    --name $CONTAINER_NAME \
    --network host \
    -p $PROMETHEUS_PORT:$PROMETHEUS_PORT \
    -e QUIC_CLIENT_ADDR=$SERVER_ADDR \
    -e QUIC_CONNECTIONS=$CONNECTIONS \
    -e QUIC_STREAMS=$STREAMS \
    -e QUIC_RATE=$RATE \
    -e QUIC_DURATION=$DURATION \
    -e QUIC_PROMETHEUS_CLIENT_PORT=$PROMETHEUS_PORT \
    2gc-network-suite:client

# Проверяем статус
echo -e "${YELLOW}Проверяем статус контейнера...${NC}"
sleep 2

if docker ps -f name=$CONTAINER_NAME | grep -q $CONTAINER_NAME; then
    echo -e "${GREEN}✅ Клиент успешно запущен!${NC}"
    echo ""
    echo -e "${BLUE}Доступные порты:${NC}"
    echo "  Prometheus метрики: http://localhost:$PROMETHEUS_PORT/metrics"
    echo ""
    echo -e "${BLUE}Полезные команды:${NC}"
    echo "  Просмотр логов: docker logs -f $CONTAINER_NAME"
    echo "  Остановка: docker stop $CONTAINER_NAME"
    echo "  Удаление: docker rm $CONTAINER_NAME"
    echo "  Вход в контейнер: docker exec -it $CONTAINER_NAME sh"
    echo ""
    echo -e "${YELLOW}Клиент будет работать $DURATION, затем автоматически остановится${NC}"
else
    echo -e "${RED}❌ Ошибка запуска клиента${NC}"
    echo -e "${YELLOW}Логи контейнера:${NC}"
    docker logs $CONTAINER_NAME
    exit 1
fi
