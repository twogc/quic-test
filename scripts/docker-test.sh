#!/bin/bash

# Скрипт для запуска тестирования 2GC Network Protocol Suite в Docker

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Test${NC}"
echo -e "${BLUE}==========================================${NC}"
echo "Запуск тестирования в Docker контейнерах"

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Ошибка: Docker не установлен${NC}"
    exit 1
fi

# Параметры по умолчанию
SERVER_ADDR=${SERVER_ADDR:-":9000"}
CONNECTIONS=${CONNECTIONS:-2}
STREAMS=${STREAMS:-4}
RATE=${RATE:-100}
DURATION=${DURATION:-60s}
PACKET_SIZE=${PACKET_SIZE:-1200}
PATTERN=${PATTERN:-"random"}

echo -e "${YELLOW}Параметры тестирования:${NC}"
echo "  Сервер: $SERVER_ADDR"
echo "  Соединения: $CONNECTIONS"
echo "  Потоки: $STREAMS"
echo "  Скорость: $RATE пакетов/сек"
echo "  Длительность: $DURATION"
echo "  Размер пакета: $PACKET_SIZE байт"
echo "  Паттерн данных: $PATTERN"

# Функция для очистки
cleanup() {
    echo -e "${YELLOW}Останавливаем контейнеры...${NC}"
    docker stop 2gc-network-server 2gc-network-client 2>/dev/null || true
    docker rm 2gc-network-server 2gc-network-client 2>/dev/null || true
}

# Устанавливаем обработчик сигналов
trap cleanup EXIT

# Очищаем существующие контейнеры
cleanup

# Собираем образы
echo -e "${YELLOW}Собираем Docker образы...${NC}"
docker build -f Dockerfile.server -t 2gc-network-suite:server .
docker build -f Dockerfile.client -t 2gc-network-suite:client .

# Запускаем сервер
echo -e "${YELLOW}Запускаем сервер...${NC}"
docker run -d \
    --name 2gc-network-server \
    --network host \
    -p 9000:9000 \
    -p 2113:2113 \
    -p 6060:6060 \
    -e QUIC_SERVER_ADDR=$SERVER_ADDR \
    -e QUIC_PROMETHEUS_SERVER_PORT=2113 \
    -e QUIC_PPROF_ADDR=:6060 \
    2gc-network-suite:server

# Ждем запуска сервера
echo -e "${YELLOW}Ждем запуска сервера...${NC}"
sleep 5

# Проверяем, что сервер запустился
if ! docker ps -f name=2gc-network-server | grep -q 2gc-network-server; then
    echo -e "${RED}❌ Ошибка запуска сервера${NC}"
    docker logs 2gc-network-server
    exit 1
fi

echo -e "${GREEN}✅ Сервер запущен${NC}"

# Запускаем клиент
echo -e "${YELLOW}Запускаем клиент...${NC}"
docker run -d \
    --name 2gc-network-client \
    --network host \
    -p 2112:2112 \
    -e QUIC_CLIENT_ADDR=localhost:9000 \
    -e QUIC_CONNECTIONS=$CONNECTIONS \
    -e QUIC_STREAMS=$STREAMS \
    -e QUIC_RATE=$RATE \
    -e QUIC_DURATION=$DURATION \
    -e QUIC_PROMETHEUS_CLIENT_PORT=2112 \
    2gc-network-suite:client

# Ждем завершения теста
echo -e "${YELLOW}Тест запущен, ждем завершения ($DURATION)...${NC}"
echo -e "${BLUE}Мониторинг:${NC}"
echo "  Сервер метрики: http://localhost:2113/metrics"
echo "  Клиент метрики: http://localhost:2112/metrics"
echo "  pprof: http://localhost:6060/debug/pprof/"
echo ""

# Мониторим логи клиента
docker logs -f 2gc-network-client &
CLIENT_PID=$!

# Ждем завершения клиента
wait $CLIENT_PID 2>/dev/null || true

# Показываем результаты
echo -e "${GREEN}✅ Тест завершен!${NC}"
echo ""
echo -e "${BLUE}Результаты:${NC}"

# Показываем логи сервера
echo -e "${YELLOW}Логи сервера:${NC}"
docker logs 2gc-network-server --tail 20

echo ""
echo -e "${YELLOW}Логи клиента:${NC}"
docker logs 2gc-network-client --tail 20

echo ""
echo -e "${BLUE}Метрики доступны по адресам:${NC}"
echo "  Сервер: http://localhost:2113/metrics"
echo "  Клиент: http://localhost:2112/metrics"
echo "  pprof: http://localhost:6060/debug/pprof/"

echo ""
echo -e "${GREEN}Тестирование завершено успешно!${NC}"
