#!/bin/bash

# Тест QUIC сервера без TLS
# Для локального тестирования

# Цвета
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - No TLS Test${NC}"
echo "Тестирование QUIC сервера без TLS"
echo ""

# Параметры теста
SERVER="localhost:9000"
CONNECTIONS=4
STREAMS=8
RATE=15
DURATION=30s

echo -e "${CYAN}📋 Параметры теста:${NC}"
echo "  🌐 Сервер: $SERVER"
echo "  🔗 Соединения: $CONNECTIONS"
echo "  📡 Потоки: $STREAMS"
echo "  ⚡ Rate: $RATE pps"
echo "  ⏱️ Длительность: $DURATION"
echo "  🔒 TLS: Отключен"
echo ""

# Запускаем клиент без TLS
echo -e "${YELLOW}🚀 Запуск клиента без TLS...${NC}"
docker run --rm \
    --name 2gc-client-no-tls \
    --network 2gc-network-suite \
    -p 2112:2112 \
    -e QUIC_CLIENT_ADDR=$SERVER \
    -e QUIC_CONNECTIONS=$CONNECTIONS \
    -e QUIC_STREAMS=$STREAMS \
    -e QUIC_RATE=$RATE \
    -e QUIC_DURATION=$DURATION \
    -e QUIC_PROMETHEUS_CLIENT_PORT=2112 \
    -e QUIC_NO_TLS=true \
    2gc-network-suite:client

TEST_EXIT_CODE=$?

echo ""
echo -e "${BLUE}==========================================${NC}"

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Тест завершен успешно!${NC}"
else
    echo -e "${RED}❌ Тест завершился с ошибками${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Готово!${NC}"
