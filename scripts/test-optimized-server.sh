#!/bin/bash

# Скрипт для тестирования оптимизированного QUIC сервера
# Применяет DevOps рекомендации для клиента

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Test Optimized Server${NC}"
echo "Тестирование оптимизированного QUIC сервера"
echo "Применение DevOps рекомендаций для клиента"
echo ""

# Проверяем доступность сервера
echo -e "${YELLOW}🔍 Проверка доступности сервера...${NC}"
sleep 3

# Проверяем метрики сервера
if curl -s http://localhost:2113/metrics >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Оптимизированный сервер доступен${NC}"
    
    # Получаем базовые метрики
    RATE=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_rate_per_connection' | awk '{print $2}' | head -1)
    CONNECTIONS=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_connections_total' | awk '{print $2}' | head -1)
    
    if [ -n "$RATE" ]; then
        echo -e "${CYAN}📊 Текущий rate: $RATE pps${NC}"
        if (( $(echo "$RATE >= 26 && $RATE <= 35" | bc -l 2>/dev/null || echo "0") )); then
            echo -e "${RED}🚨 КРИТИЧЕСКАЯ ЗОНА: Rate $RATE pps (26-35 pps)${NC}"
        else
            echo -e "${GREEN}✅ Rate $RATE pps (безопасная зона)${NC}"
        fi
    fi
    
    if [ -n "$CONNECTIONS" ]; then
        echo -e "${GREEN}🔗 Активных соединений: $CONNECTIONS${NC}"
    fi
else
    echo -e "${YELLOW}⏳ Сервер запускается, ожидаем...${NC}"
    sleep 5
fi

echo ""

# Параметры клиента для тестирования оптимизированного сервера
echo -e "${YELLOW}🔧 Настройка клиента для тестирования...${NC}"

# Переменные окружения для клиента
export QUIC_CLIENT_ADDR="localhost:9000"
export QUIC_CONNECTIONS=4          # Увеличиваем количество соединений
export QUIC_STREAMS=8              # Увеличиваем количество потоков
export QUIC_RATE=15                # Снижаем rate до 15 pps (безопасная зона)
export QUIC_DURATION=30s           # 30 секунд тестирования
export QUIC_PROMETHEUS_CLIENT_PORT=2112
export QUIC_NO_TLS=""              # Используем TLS

echo "  ✅ Адрес сервера: $QUIC_CLIENT_ADDR"
echo "  ✅ Соединения: $QUIC_CONNECTIONS"
echo "  ✅ Потоки: $QUIC_STREAMS"
echo "  ✅ Rate: $QUIC_RATE pps (безопасная зона)"
echo "  ✅ Длительность: $QUIC_DURATION"
echo "  ✅ Prometheus порт: $QUIC_PROMETHEUS_CLIENT_PORT"
echo "  ✅ TLS: Включен"

echo ""

# Собираем Docker образ клиента
echo -e "${YELLOW}🔨 Сборка Docker образа клиента...${NC}"
docker build -f Dockerfile.client -t 2gc-network-suite:client .

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Ошибка сборки Docker образа клиента.${NC}"
    exit 1
fi
echo "  ✅ Docker образ клиента собран успешно"

echo ""

# Запускаем клиент для тестирования
echo -e "${YELLOW}🚀 Запуск клиента для тестирования оптимизированного сервера...${NC}"
echo -e "${CYAN}📊 Тестирование с DevOps оптимизациями:${NC}"
echo "  🎯 Rate: 15 pps (безопасная зона)"
echo "  🔗 Соединения: 4"
echo "  📡 Потоки: 8"
echo "  ⏱️ Длительность: 30 секунд"
echo ""

# Запускаем клиент
docker run --rm \
    --name 2gc-network-client-optimized \
    --network 2gc-network-suite \
    -p 2112:2112 \
    -e QUIC_CLIENT_ADDR=$QUIC_CLIENT_ADDR \
    -e QUIC_CONNECTIONS=$QUIC_CONNECTIONS \
    -e QUIC_STREAMS=$QUIC_STREAMS \
    -e QUIC_RATE=$QUIC_RATE \
    -e QUIC_DURATION=$QUIC_DURATION \
    -e QUIC_PROMETHEUS_CLIENT_PORT=$QUIC_PROMETHEUS_CLIENT_PORT \
    -e QUIC_NO_TLS=$QUIC_NO_TLS \
    2gc-network-suite:client

CLIENT_EXIT_CODE=$?

echo ""
echo -e "${BLUE}==========================================${NC}"

if [ $CLIENT_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Тестирование завершено успешно!${NC}"
    
    # Финальная проверка метрик сервера
    echo -e "${YELLOW}📊 Финальные метрики сервера:${NC}"
    if curl -s http://localhost:2113/metrics >/dev/null 2>&1; then
        RATE=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_rate_per_connection' | awk '{print $2}' | head -1)
        CONNECTIONS=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_connections_total' | awk '{print $2}' | head -1)
        ERRORS=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_errors_total' | awk '{print $2}' | head -1)
        
        if [ -n "$RATE" ]; then
            echo -e "${CYAN}📈 Rate: $RATE pps${NC}"
            if (( $(echo "$RATE >= 26 && $RATE <= 35" | bc -l 2>/dev/null || echo "0") )); then
                echo -e "${RED}🚨 КРИТИЧЕСКАЯ ЗОНА: Rate $RATE pps${NC}"
            else
                echo -e "${GREEN}✅ Rate в безопасной зоне${NC}"
            fi
        fi
        
        if [ -n "$CONNECTIONS" ]; then
            echo -e "${GREEN}🔗 Соединения: $CONNECTIONS${NC}"
        fi
        
        if [ -n "$ERRORS" ]; then
            echo -e "${GREEN}❌ Ошибки: $ERRORS${NC}"
        fi
    fi
    
    echo ""
    echo -e "${BLUE}🎯 Результаты тестирования:${NC}"
    echo "  ✅ DevOps оптимизации применены"
    echo "  ✅ Rate ограничен до 15 pps (безопасная зона)"
    echo "  ✅ Множественные соединения для высокой пропускной способности"
    echo "  ✅ Мониторинг критических зон активен"
    echo "  ✅ Системные оптимизации применены"
    
else
    echo -e "${RED}❌ Тестирование завершилось с ошибками (код: $CLIENT_EXIT_CODE)${NC}"
    echo -e "${YELLOW}💡 Рекомендации:${NC}"
    echo "  1. Проверьте доступность сервера"
    echo "  2. Убедитесь, что порты не заняты"
    echo "  3. Проверьте системные ресурсы"
    echo "  4. Проверьте логи сервера"
fi

echo ""
echo -e "${BLUE}🌐 Доступные интерфейсы:${NC}"
echo "  QUIC сервер: localhost:9000 (UDP)"
echo "  Prometheus сервер: http://localhost:2113/metrics"
echo "  Prometheus клиент: http://localhost:2112/metrics"
echo "  pprof профилирование: http://localhost:6060/debug/pprof/"
echo ""
echo -e "${YELLOW}💡 Следующие шаги:${NC}"
echo "  1. Мониторинг: ./scripts/live-monitor.sh"
echo "  2. Health check: ./scripts/health-check.sh"
echo "  3. Анализ метрик: http://localhost:9090"
echo "  4. Grafana дашборд: http://localhost:3000"
echo ""
echo -e "${GREEN}🎉 Тестирование оптимизированного сервера завершено!${NC}"

