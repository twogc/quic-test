#!/bin/bash

# Скрипт для тестирования внешнего доступа к QUIC серверу

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - External Test${NC}"
echo "Тестирование внешнего доступа к QUIC серверу"
echo ""

# Получаем внешний IP
EXTERNAL_IP=$(curl -s ifconfig.me)
echo -e "${YELLOW}Внешний IP сервера: ${EXTERNAL_IP}${NC}"
echo -e "${YELLOW}QUIC сервер доступен на: ${EXTERNAL_IP}:9000${NC}"
echo ""

# Проверяем, что сервер запущен
if ! docker ps | grep -q "2gc-network-server"; then
    echo -e "${RED}Ошибка: QUIC сервер не запущен${NC}"
    echo "Запустите сервер: ./scripts/docker-server.sh"
    exit 1
fi

echo -e "${YELLOW}Тестируем локальное подключение...${NC}"
timeout 5s docker run --rm --network 2gc-network-suite \
    -e QUIC_CLIENT_ADDR=2gc-network-server:9000 \
    -e QUIC_CONNECTIONS=1 \
    -e QUIC_STREAMS=1 \
    -e QUIC_RATE=10 \
    -e QUIC_DURATION=3s \
    -e QUIC_NO_TLS=true \
    2gc-network-suite:client

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Локальное подключение работает!${NC}"
else
    echo -e "${RED}❌ Локальное подключение не работает${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Тестируем внешнее подключение...${NC}"
timeout 10s docker run --rm --network host \
    -e QUIC_CLIENT_ADDR=${EXTERNAL_IP}:9000 \
    -e QUIC_CONNECTIONS=1 \
    -e QUIC_STREAMS=1 \
    -e QUIC_RATE=10 \
    -e QUIC_DURATION=5s \
    -e QUIC_NO_TLS=true \
    2gc-network-suite:client

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Внешнее подключение работает!${NC}"
    echo ""
    echo -e "${BLUE}🌐 Сервер доступен извне:${NC}"
    echo "  QUIC: ${EXTERNAL_IP}:9000 (UDP)"
    echo "  Prometheus: http://${EXTERNAL_IP}:2113/metrics"
    echo "  pprof: http://${EXTERNAL_IP}:6060/debug/pprof/"
    echo ""
    echo -e "${BLUE}📊 Мониторинг:${NC}"
    echo "  Grafana: http://${EXTERNAL_IP}:3000 (admin/admin)"
    echo "  Prometheus: http://${EXTERNAL_IP}:9090"
    echo "  Jaeger: http://${EXTERNAL_IP}:16686"
else
    echo -e "${RED}❌ Внешнее подключение не работает${NC}"
    echo ""
    echo -e "${YELLOW}Возможные причины:${NC}"
    echo "  1. Файрвол блокирует UDP порт 9000"
    echo "  2. NAT не пробрасывает UDP трафик"
    echo "  3. Провайдер блокирует входящие UDP соединения"
    echo ""
    echo -e "${YELLOW}Проверьте:${NC}"
    echo "  - Открыт ли порт 9000/udp в файрволе"
    echo "  - Настроен ли NAT для UDP трафика"
    echo "  - Разрешены ли входящие UDP соединения"
fi

echo ""
echo -e "${BLUE}Полезные команды:${NC}"
echo "  Просмотр логов сервера: docker logs -f 2gc-network-server"
echo "  Проверка портов: sudo ss -ulpn | grep :9000"
echo "  Тест UDP: nc -u ${EXTERNAL_IP} 9000"

