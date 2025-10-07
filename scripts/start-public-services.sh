#!/bin/bash

# Скрипт для запуска всех сервисов 2GC Network Protocol Suite с внешним доступом

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Public Services${NC}"
echo "Запуск всех сервисов с внешним доступом"
echo ""

# Получаем внешний IP
EXTERNAL_IP=$(curl -s ifconfig.me)
echo -e "${YELLOW}Внешний IP: ${EXTERNAL_IP}${NC}"
echo ""

# Проверяем UFW
echo -e "${YELLOW}Проверяем UFW правила...${NC}"
if sudo ufw status | grep -q "9000/udp"; then
    echo -e "${GREEN}✅ UDP порт 9000 открыт${NC}"
else
    echo -e "${RED}❌ UDP порт 9000 не открыт${NC}"
    echo "Открываем порт: sudo ufw allow 9000/udp"
    sudo ufw allow 9000/udp
fi

# Создаем Docker сеть если не существует
if ! docker network ls | grep -q "2gc-network-suite"; then
    echo -e "${YELLOW}Создаем Docker сеть...${NC}"
    docker network create 2gc-network-suite
fi

# Запускаем QUIC сервер
echo -e "${YELLOW}Запускаем QUIC сервер...${NC}"
docker stop 2gc-network-server 2>/dev/null
docker rm 2gc-network-server 2>/dev/null
docker run -d --name 2gc-network-server --network 2gc-network-suite \
    -p 9000:9000/udp -p 2113:2113 -p 6060:6060 \
    2gc-network-suite:server

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ QUIC сервер запущен${NC}"
else
    echo -e "${RED}❌ Ошибка запуска QUIC сервера${NC}"
    exit 1
fi

# Запускаем мониторинг стек
echo -e "${YELLOW}Запускаем мониторинг стек...${NC}"
docker compose -f docker-compose.monitoring.yml up -d

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Мониторинг стек запущен${NC}"
else
    echo -e "${RED}❌ Ошибка запуска мониторинга${NC}"
    exit 1
fi

# Ждем запуска сервисов
echo -e "${YELLOW}Ожидаем запуска сервисов...${NC}"
sleep 5

# Проверяем статус
echo -e "${YELLOW}Проверяем статус сервисов...${NC}"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo -e "${GREEN}🎉 Все сервисы запущены!${NC}"
echo ""
echo -e "${BLUE}🌐 Внешний доступ:${NC}"
echo "  QUIC Server: ${EXTERNAL_IP}:9000 (UDP)"
echo "  Prometheus: http://${EXTERNAL_IP}:9090"
echo "  Grafana: http://${EXTERNAL_IP}:3000 (admin/admin)"
echo "  Jaeger: http://${EXTERNAL_IP}:16686"
echo ""
echo -e "${BLUE}📊 Метрики:${NC}"
echo "  QUIC Server Metrics: http://${EXTERNAL_IP}:2113/metrics"
echo "  pprof Profiling: http://${EXTERNAL_IP}:6060/debug/pprof/"
echo ""
echo -e "${BLUE}🔧 Управление:${NC}"
echo "  Остановить все: docker compose -f docker-compose.monitoring.yml down && docker stop 2gc-network-server"
echo "  Логи сервера: docker logs -f 2gc-network-server"
echo "  Логи мониторинга: docker compose -f docker-compose.monitoring.yml logs -f"
echo ""
echo -e "${BLUE}🧪 Тестирование:${NC}"
echo "  ./scripts/test-external-access.sh"
echo "  ./scripts/docker-client.sh"

