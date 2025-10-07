#!/bin/bash

# Быстрый тест оптимизированного QUIC сервера
# Применяет DevOps рекомендации

# Цвета
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Quick Test${NC}"
echo "Тестирование оптимизированного QUIC сервера"
echo ""

# Проверяем доступность сервера
echo -e "${YELLOW}🔍 Проверка сервера...${NC}"
if curl -s http://localhost:2113/metrics >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Сервер доступен${NC}"
else
    echo -e "${RED}❌ Сервер недоступен${NC}"
    exit 1
fi

# Параметры теста (DevOps рекомендации)
SERVER="localhost:9000"
CONNECTIONS=4
STREAMS=8
RATE=15  # Безопасная зона
DURATION=30s

echo -e "${CYAN}📋 Параметры теста:${NC}"
echo "  🌐 Сервер: $SERVER"
echo "  🔗 Соединения: $CONNECTIONS"
echo "  📡 Потоки: $STREAMS"
echo "  ⚡ Rate: $RATE pps (безопасная зона)"
echo "  ⏱️ Длительность: $DURATION"
echo ""

# Собираем клиент
echo -e "${YELLOW}🔨 Сборка клиента...${NC}"
docker build -f Dockerfile.client -t 2gc-network-suite:client .

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Ошибка сборки клиента${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Клиент собран успешно${NC}"

echo ""

# Запускаем тест
echo -e "${YELLOW}🚀 Запуск теста...${NC}"
docker run --rm \
    --name 2gc-client-test \
    --network 2gc-network-suite \
    -p 2112:2112 \
    -e QUIC_CLIENT_ADDR=$SERVER \
    -e QUIC_CONNECTIONS=$CONNECTIONS \
    -e QUIC_STREAMS=$STREAMS \
    -e QUIC_RATE=$RATE \
    -e QUIC_DURATION=$DURATION \
    -e QUIC_PROMETHEUS_CLIENT_PORT=2112 \
    2gc-network-suite:client

TEST_EXIT_CODE=$?

echo ""
echo -e "${BLUE}==========================================${NC}"

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Тест завершен успешно!${NC}"
    echo ""
    echo -e "${BLUE}📊 Результаты:${NC}"
    echo "  ✅ DevOps оптимизации применены"
    echo "  ✅ Rate: $RATE pps (безопасная зона)"
    echo "  ✅ Соединения: $CONNECTIONS"
    echo "  ✅ Потоки: $STREAMS"
    echo ""
    echo -e "${BLUE}🌐 Доступные интерфейсы:${NC}"
    echo "  QUIC сервер: localhost:9000"
    echo "  Prometheus сервер: http://localhost:2113/metrics"
    echo "  Prometheus клиент: http://localhost:2112/metrics"
    echo "  Grafana: http://localhost:3000"
    echo "  Prometheus UI: http://localhost:9090"
else
    echo -e "${RED}❌ Тест завершился с ошибками${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Готово!${NC}"
