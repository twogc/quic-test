#!/bin/bash

# Быстрый тест удаленного QUIC сервера с DevOps настройками

# Цвета
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}🚀 Быстрый тест удаленного QUIC сервера${NC}"
echo ""

# Параметры по умолчанию
SERVER=${1:-"localhost:9000"}
CONNECTIONS=${2:-4}
RATE=${3:-15}

echo -e "${YELLOW}📋 Параметры теста:${NC}"
echo "  🌐 Сервер: $SERVER"
echo "  🔗 Соединения: $CONNECTIONS"
echo "  ⚡ Rate: $RATE pps (безопасная зона)"
echo ""

# Запускаем клиент
echo -e "${GREEN}▶️ Запуск клиента...${NC}"
docker run --rm \
    --name 2gc-client-quick \
    --network 2gc-network-suite \
    -p 2112:2112 \
    -e QUIC_CLIENT_ADDR=$SERVER \
    -e QUIC_CONNECTIONS=$CONNECTIONS \
    -e QUIC_STREAMS=8 \
    -e QUIC_RATE=$RATE \
    -e QUIC_DURATION=30s \
    -e QUIC_PROMETHEUS_CLIENT_PORT=2112 \
    2gc-network-suite:client

echo -e "${GREEN}✅ Тест завершен!${NC}"

