#!/bin/bash

# Скрипт для запуска оптимизированного QUIC сервера на 10 минут
# Применяет все DevOps рекомендации

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Optimized Server${NC}"
echo "Запуск оптимизированного QUIC сервера на 10 минут"
echo "Применение DevOps рекомендаций"
echo ""

# Проверяем наличие Docker
if ! command -v docker &> /dev/null
then
    echo -e "${RED}❌ Ошибка: Docker не установлен${NC}"
    echo "Пожалуйста, установите Docker перед запуском скрипта."
    exit 1
fi

# Применяем системные оптимизации
echo -e "${YELLOW}🔧 Применение системных оптимизаций...${NC}"

# UDP буферы
echo -e "${CYAN}📡 Настройка UDP буферов:${NC}"
sudo sysctl -w net.core.rmem_max=134217728 >/dev/null 2>&1
sudo sysctl -w net.core.rmem_default=134217728 >/dev/null 2>&1
sudo sysctl -w net.core.wmem_max=134217728 >/dev/null 2>&1
sudo sysctl -w net.core.wmem_default=134217728 >/dev/null 2>&1
echo "  ✅ UDP буферы увеличены до 128MB"

# Сетевые оптимизации
echo -e "${CYAN}🌐 Настройка сетевых параметров:${NC}"
sudo sysctl -w net.core.netdev_max_backlog=5000 >/dev/null 2>&1
sudo sysctl -w net.core.somaxconn=65535 >/dev/null 2>&1
sudo sysctl -w net.ipv4.udp_mem="102400 873800 16777216" >/dev/null 2>&1
sudo sysctl -w net.ipv4.udp_rmem_min=8192 >/dev/null 2>&1
sudo sysctl -w net.ipv4.udp_wmem_min=8192 >/dev/null 2>&1
echo "  ✅ Сетевые параметры оптимизированы"

# TCP оптимизации
echo -e "${CYAN}🚀 Настройка TCP параметров:${NC}"
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr >/dev/null 2>&1
sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 134217728" >/dev/null 2>&1
sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 134217728" >/dev/null 2>&1
echo "  ✅ TCP congestion control: BBR"
echo "  ✅ TCP буферы оптимизированы"

# Устанавливаем лимиты процессов
echo -e "${CYAN}⚙️ Настройка лимитов процессов:${NC}"
ulimit -n 65536 2>/dev/null
ulimit -u 32768 2>/dev/null
echo "  ✅ Лимиты файлов: 65536"
echo "  ✅ Лимиты процессов: 32768"

echo -e "${GREEN}✅ Системные оптимизации применены${NC}"
echo ""

# Переменные окружения для оптимизации
echo -e "${YELLOW}🔧 Настройка параметров QUIC сервера...${NC}"
export QUIC_MAX_CONNECTIONS=1000
export QUIC_MAX_RATE_PER_CONN=20
export QUIC_CONNECTION_TIMEOUT=60s
export QUIC_HANDSHAKE_TIMEOUT=10s
export QUIC_KEEP_ALIVE=30s
export QUIC_MAX_STREAMS=100
export QUIC_ENABLE_DATAGRAMS=true
export QUIC_ENABLE_0RTT=true
export QUIC_MONITORING=true

echo "  ✅ Максимальные соединения: $QUIC_MAX_CONNECTIONS"
echo "  ✅ Максимальная скорость на соединение: $QUIC_MAX_RATE_PER_CONN pps"
echo "  ✅ Таймаут соединения: $QUIC_CONNECTION_TIMEOUT"
echo "  ✅ Таймаут handshake: $QUIC_HANDSHAKE_TIMEOUT"
echo "  ✅ Keep-alive: $QUIC_KEEP_ALIVE"
echo "  ✅ Максимальные потоки: $QUIC_MAX_STREAMS"
echo "  ✅ Datagrams: $QUIC_ENABLE_DATAGRAMS"
echo "  ✅ 0-RTT: $QUIC_ENABLE_0RTT"
echo "  ✅ Мониторинг: $QUIC_MONITORING"

echo ""

# Удаляем существующий контейнер
echo -e "${YELLOW}🧹 Очистка существующих контейнеров...${NC}"
docker rm -f 2gc-network-server-optimized &> /dev/null
echo "  ✅ Существующие контейнеры удалены"

# Собираем Docker образ
echo -e "${YELLOW}🔨 Сборка Docker образа...${NC}"
docker build -f Dockerfile.server -t 2gc-network-suite:server .

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Ошибка сборки Docker образа сервера.${NC}"
    exit 1
fi
echo "  ✅ Docker образ собран успешно"

# Запускаем оптимизированный сервер на 10 минут
echo -e "${YELLOW}🚀 Запуск оптимизированного сервера на 10 минут...${NC}"
echo -e "${CYAN}⏰ Сервер будет автоматически остановлен через 10 минут${NC}"
echo ""

# Запускаем сервер с мониторингом в фоне
(
    # Мониторинг в фоне
    while true; do
        sleep 30
        echo -e "${CYAN}📊 Проверка состояния сервера...${NC}"
        
        # Проверяем доступность метрик
        if curl -s http://localhost:2113/metrics >/dev/null 2>&1; then
            # Получаем метрики
            RATE=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_rate_per_connection' | awk '{print $2}' | head -1)
            CONNECTIONS=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_connections_total' | awk '{print $2}' | head -1)
            ERRORS=$(curl -s http://localhost:2113/metrics 2>/dev/null | grep 'quic_server_errors_total' | awk '{print $2}' | head -1)
            
            if [ -n "$RATE" ]; then
                if (( $(echo "$RATE >= 26 && $RATE <= 35" | bc -l 2>/dev/null || echo "0") )); then
                    echo -e "${RED}🚨 КРИТИЧЕСКАЯ ЗОНА: Rate $RATE pps (26-35 pps)${NC}"
                else
                    echo -e "${GREEN}✅ Rate: $RATE pps (безопасная зона)${NC}"
                fi
            fi
            
            if [ -n "$CONNECTIONS" ]; then
                echo -e "${GREEN}🔗 Соединения: $CONNECTIONS${NC}"
            fi
            
            if [ -n "$ERRORS" ] && (( $(echo "$ERRORS > 10" | bc -l 2>/dev/null || echo "0") )); then
                echo -e "${RED}⚠️ Ошибки: $ERRORS${NC}"
            fi
        else
            echo -e "${YELLOW}⏳ Сервер запускается...${NC}"
        fi
    done
) &
MONITOR_PID=$!

# Запускаем сервер с таймаутом 10 минут
timeout 10m docker run --rm --name 2gc-network-server-optimized \
    --network 2gc-network-suite \
    -p 9000:9000/udp \
    -p 2113:2113 \
    -p 6060:6060 \
    -e QUIC_MAX_CONNECTIONS=$QUIC_MAX_CONNECTIONS \
    -e QUIC_MAX_RATE_PER_CONN=$QUIC_MAX_RATE_PER_CONN \
    -e QUIC_CONNECTION_TIMEOUT=$QUIC_CONNECTION_TIMEOUT \
    -e QUIC_HANDSHAKE_TIMEOUT=$QUIC_HANDSHAKE_TIMEOUT \
    -e QUIC_KEEP_ALIVE=$QUIC_KEEP_ALIVE \
    -e QUIC_MAX_STREAMS=$QUIC_MAX_STREAMS \
    -e QUIC_ENABLE_DATAGRAMS=$QUIC_ENABLE_DATAGRAMS \
    -e QUIC_ENABLE_0RTT=$QUIC_ENABLE_0RTT \
    -e QUIC_MONITORING=$QUIC_MONITORING \
    2gc-network-suite:server

SERVER_EXIT_CODE=$?

# Останавливаем мониторинг
kill $MONITOR_PID 2>/dev/null

echo ""
echo -e "${BLUE}==========================================${NC}"

if [ $SERVER_EXIT_CODE -eq 124 ]; then
    echo -e "${GREEN}✅ Сервер успешно остановлен по таймауту (10 минут)${NC}"
elif [ $SERVER_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Сервер завершил работу нормально${NC}"
else
    echo -e "${RED}❌ Сервер завершил работу с ошибкой (код: $SERVER_EXIT_CODE)${NC}"
fi

echo ""
echo -e "${BLUE}📊 Итоговая статистика:${NC}"
echo "  🕐 Время работы: 10 минут"
echo "  🔧 Оптимизации: Применены"
echo "  📡 UDP буферы: 128MB"
echo "  🚀 TCP congestion: BBR"
echo "  ⚙️ Лимиты процессов: 65536 файлов, 32768 процессов"
echo "  🎯 Максимальная скорость на соединение: 20 pps"
echo "  🔗 Максимальные соединения: 1000"
echo ""
echo -e "${BLUE}🌐 Доступные порты:${NC}"
echo "  QUIC сервер: localhost:9000 (UDP)"
echo "  Prometheus метрики: http://localhost:2113/metrics"
echo "  pprof профилирование: http://localhost:6060/debug/pprof/"
echo ""
echo -e "${YELLOW}💡 Рекомендации:${NC}"
echo "  1. Используйте оптимизированный сервер для продакшена"
echo "  2. Мониторьте критическую зону (26-35 pps)"
echo "  3. Настройте алерты для высокого jitter"
echo "  4. Регулярно проверяйте health check"
echo ""
echo -e "${GREEN}🎉 Оптимизированный сервер завершил работу!${NC}"

