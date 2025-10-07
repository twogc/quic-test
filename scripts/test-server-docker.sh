#!/bin/bash

# Тест QUIC сервера с правильным адресом в Docker сети
# Использует имя контейнера сервера

# Цвета
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Docker Test${NC}"
echo "Тестирование QUIC сервера в Docker сети"
echo ""

# Параметры теста - используем имя контейнера сервера
SERVER="2gc-network-server-optimized:9000"
CONNECTIONS=4
STREAMS=8
RATE=15
DURATION=30s

echo -e "${CYAN}📋 Параметры теста:${NC}"
echo "  🌐 Сервер: $SERVER (Docker контейнер)"
echo "  🔗 Соединения: $CONNECTIONS"
echo "  📡 Потоки: $STREAMS"
echo "  ⚡ Rate: $RATE pps (безопасная зона)"
echo "  ⏱️ Длительность: $DURATION"
echo "  🔒 TLS: Отключен"
echo ""

# Запускаем клиент в той же Docker сети
echo -e "${YELLOW}🚀 Запуск клиента в Docker сети...${NC}"
docker run --rm \
    --name 2gc-client-docker \
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
    echo ""
    echo -e "${BLUE}📊 Результаты:${NC}"
    echo "  ✅ DevOps оптимизации применены"
    echo "  ✅ Rate: $RATE pps (безопасная зона)"
    echo "  ✅ Соединения: $CONNECTIONS"
    echo "  ✅ Потоки: $STREAMS"
    echo "  ✅ Docker сеть: 2gc-network-suite"
    echo ""
    echo -e "${BLUE}🌐 Доступные интерфейсы:${NC}"
    echo "  QUIC сервер: localhost:9000"
    echo "  Prometheus сервер: http://localhost:2113/metrics"
    echo "  Prometheus клиент: http://localhost:2112/metrics"
    echo "  Grafana: http://localhost:3000"
    echo "  Prometheus UI: http://localhost:9090"
else
    echo -e "${RED}❌ Тест завершился с ошибками${NC}"
    echo -e "${YELLOW}💡 Рекомендации:${NC}"
    echo "  1. Проверьте, что сервер запущен"
    echo "  2. Проверьте Docker сеть"
    echo "  3. Проверьте логи сервера"
fi

echo ""
echo -e "${GREEN}🎉 Готово!${NC}"
