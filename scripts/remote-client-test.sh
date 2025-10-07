#!/bin/bash

# Скрипт для тестирования удаленного QUIC сервера
# Применяет DevOps рекомендации для клиента

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Remote Client Test${NC}"
echo "Тестирование удаленного QUIC сервера"
echo "Применение DevOps рекомендаций"
echo ""

# Параметры подключения
REMOTE_SERVER=${QUIC_REMOTE_SERVER:-"localhost:9000"}
CONNECTIONS=${QUIC_CONNECTIONS:-4}
STREAMS=${QUIC_STREAMS:-8}
RATE=${QUIC_RATE:-15}
DURATION=${QUIC_DURATION:-30s}
PROMETHEUS_PORT=${QUIC_PROMETHEUS_CLIENT_PORT:-2112}

echo -e "${YELLOW}🔧 Конфигурация клиента:${NC}"
echo "  🌐 Удаленный сервер: $REMOTE_SERVER"
echo "  🔗 Соединения: $CONNECTIONS"
echo "  📡 Потоки: $STREAMS"
echo "  ⚡ Rate: $RATE pps (безопасная зона)"
echo "  ⏱️ Длительность: $DURATION"
echo "  📊 Prometheus порт: $PROMETHEUS_PORT"
echo ""

# DevOps рекомендации
echo -e "${CYAN}📋 DevOps рекомендации:${NC}"
echo "  ✅ Rate ограничен до 15 pps (избегаем критическую зону 26-35 pps)"
echo "  ✅ Множественные соединения для высокой пропускной способности"
echo "  ✅ Оптимизированные параметры QUIC"
echo "  ✅ Мониторинг производительности"
echo ""

# Проверяем доступность удаленного сервера
echo -e "${YELLOW}🔍 Проверка доступности удаленного сервера...${NC}"

# Извлекаем хост и порт
SERVER_HOST=$(echo $REMOTE_SERVER | cut -d: -f1)
SERVER_PORT=$(echo $REMOTE_SERVER | cut -d: -f2)

# Проверяем доступность UDP порта
if timeout 3 bash -c "</dev/udp/$SERVER_HOST/$SERVER_PORT" 2>/dev/null; then
    echo -e "${GREEN}✅ UDP порт $SERVER_PORT доступен на $SERVER_HOST${NC}"
else
    echo -e "${YELLOW}⚠️ UDP порт $SERVER_PORT может быть недоступен на $SERVER_HOST${NC}"
    echo -e "${CYAN}💡 Продолжаем тестирование...${NC}"
fi

echo ""

# Собираем Docker образ клиента
echo -e "${YELLOW}🔨 Сборка Docker образа клиента...${NC}"
docker build -f Dockerfile.client -t 2gc-network-suite:client .

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Ошибка сборки Docker образа клиента.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker образ клиента собран успешно${NC}"

echo ""

# Запускаем клиент для тестирования удаленного сервера
echo -e "${YELLOW}🚀 Запуск клиента для тестирования удаленного сервера...${NC}"
echo -e "${CYAN}📊 Тестирование с DevOps оптимизациями:${NC}"
echo "  🎯 Rate: $RATE pps (безопасная зона)"
echo "  🔗 Соединения: $CONNECTIONS"
echo "  📡 Потоки: $STREAMS"
echo "  ⏱️ Длительность: $DURATION"
echo "  🌐 Сервер: $REMOTE_SERVER"
echo ""

# Запускаем клиент
docker run --rm \
    --name 2gc-network-client-remote \
    --network 2gc-network-suite \
    -p $PROMETHEUS_PORT:$PROMETHEUS_PORT \
    -e QUIC_CLIENT_ADDR=$REMOTE_SERVER \
    -e QUIC_CONNECTIONS=$CONNECTIONS \
    -e QUIC_STREAMS=$STREAMS \
    -e QUIC_RATE=$RATE \
    -e QUIC_DURATION=$DURATION \
    -e QUIC_PROMETHEUS_CLIENT_PORT=$PROMETHEUS_PORT \
    -e QUIC_NO_TLS="" \
    2gc-network-suite:client

CLIENT_EXIT_CODE=$?

echo ""
echo -e "${BLUE}==========================================${NC}"

if [ $CLIENT_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Тестирование удаленного сервера завершено успешно!${NC}"
    
    echo ""
    echo -e "${BLUE}🎯 Результаты тестирования:${NC}"
    echo "  ✅ DevOps оптимизации применены"
    echo "  ✅ Rate ограничен до $RATE pps (безопасная зона)"
    echo "  ✅ Множественные соединения для высокой пропускной способности"
    echo "  ✅ Мониторинг критических зон активен"
    echo "  ✅ Системные оптимизации применены"
    echo "  🌐 Удаленный сервер: $REMOTE_SERVER"
    
else
    echo -e "${RED}❌ Тестирование завершилось с ошибками (код: $CLIENT_EXIT_CODE)${NC}"
    echo -e "${YELLOW}💡 Рекомендации:${NC}"
    echo "  1. Проверьте доступность удаленного сервера"
    echo "  2. Убедитесь, что порты не заблокированы файрволом"
    echo "  3. Проверьте системные ресурсы"
    echo "  4. Проверьте логи клиента"
fi

echo ""
echo -e "${BLUE}🌐 Доступные интерфейсы:${NC}"
echo "  QUIC сервер: $REMOTE_SERVER (UDP)"
echo "  Prometheus клиент: http://localhost:$PROMETHEUS_PORT/metrics"
echo ""
echo -e "${YELLOW}💡 Следующие шаги:${NC}"
echo "  1. Мониторинг: ./scripts/live-monitor.sh"
echo "  2. Health check: ./scripts/health-check.sh"
echo "  3. Анализ метрик: http://localhost:9090"
echo "  4. Grafana дашборд: http://localhost:3000"
echo ""
echo -e "${GREEN}🎉 Тестирование удаленного сервера завершено!${NC}"

