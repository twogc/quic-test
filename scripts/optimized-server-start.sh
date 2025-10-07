#!/bin/bash

# Оптимизированный запуск QUIC сервера
# Применяет все DevOps рекомендации

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Optimized Server${NC}"
echo "Запуск оптимизированного QUIC сервера"
echo ""

# Применяем системные оптимизации
echo -e "${YELLOW}🔧 Применение системных оптимизаций...${NC}"
sudo sysctl -w net.core.rmem_max=134217728 >/dev/null 2>&1
sudo sysctl -w net.core.wmem_max=134217728 >/dev/null 2>&1
sudo sysctl -w net.core.netdev_max_backlog=5000 >/dev/null 2>&1
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr >/dev/null 2>&1

# Устанавливаем лимиты процессов
ulimit -n 65536 2>/dev/null
ulimit -u 32768 2>/dev/null

echo -e "${GREEN}✅ Системные оптимизации применены${NC}"

# Запускаем сервер с оптимизированными параметрами
echo -e "${YELLOW}🚀 Запуск оптимизированного сервера...${NC}"

# Переменные окружения для оптимизации
export QUIC_MAX_CONNECTIONS=1000
export QUIC_MAX_RATE_PER_CONN=20
export QUIC_CONNECTION_TIMEOUT=60s
export QUIC_HANDSHAKE_TIMEOUT=10s
export QUIC_KEEP_ALIVE=30s
export QUIC_MAX_STREAMS=100
export QUIC_ENABLE_DATAGRAMS=true
export QUIC_ENABLE_0RTT=true
export QUIC_MONITORING=true

# Запускаем Docker контейнер с оптимизированными параметрами
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

echo -e "${GREEN}✅ Оптимизированный сервер завершил работу${NC}"
