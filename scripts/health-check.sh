#!/bin/bash

# Health check для QUIC сервера
# Проверяет критические зоны и производительность

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

SERVER_URL="http://localhost:2113/metrics"
CRITICAL_ZONE_ALERT=false
WARNINGS=0

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  QUIC Server Health Check${NC}"
echo "Проверка состояния сервера"
echo ""

# Проверяем доступность сервера
if ! curl -s $SERVER_URL >/dev/null 2>&1; then
    echo -e "${RED}❌ Сервер недоступен${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Сервер доступен${NC}"

# Проверяем критическую зону (26-35 pps)
echo -e "${YELLOW}🔍 Проверка критической зоны...${NC}"
RATE=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_rate_per_connection' | awk '{print $2}' | head -1)

if [ -n "$RATE" ] && (( $(echo "$RATE >= 26 && $RATE <= 35" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "${RED}🚨 КРИТИЧЕСКАЯ ЗОНА: Rate $RATE pps (26-35 pps)${NC}"
    CRITICAL_ZONE_ALERT=true
    WARNINGS=$((WARNINGS + 1))
elif [ -n "$RATE" ]; then
    echo -e "${GREEN}✅ Rate $RATE pps (безопасная зона)${NC}"
fi

# Проверяем jitter
echo -e "${YELLOW}Проверка jitter...${NC}"
JITTER=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_jitter_seconds' | awk '{print $2}' | head -1)

if [ -n "$JITTER" ] && (( $(echo "$JITTER > 0.1" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "${RED}⚠️ Высокий jitter: $JITTER секунд${NC}"
    WARNINGS=$((WARNINGS + 1))
elif [ -n "$JITTER" ]; then
    echo -e "${GREEN}✅ Jitter: $JITTER секунд (норма)${NC}"
fi

# Проверяем ошибки
echo -e "${YELLOW}❌ Проверка ошибок...${NC}"
ERRORS=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_errors_total' | awk '{print $2}' | head -1)

if [ -n "$ERRORS" ] && (( $(echo "$ERRORS > 10" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "${RED}⚠️ Высокий уровень ошибок: $ERRORS${NC}"
    WARNINGS=$((WARNINGS + 1))
elif [ -n "$ERRORS" ]; then
    echo -e "${GREEN}✅ Ошибки: $ERRORS (норма)${NC}"
fi

# Проверяем соединения
echo -e "${YELLOW}🔗 Проверка соединений...${NC}"
CONNECTIONS=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_connections_total' | awk '{print $2}' | head -1)

if [ -n "$CONNECTIONS" ]; then
    echo -e "${GREEN}✅ Активных соединений: $CONNECTIONS${NC}"
fi

# Итоговый статус
echo ""
echo -e "${BLUE}==========================================${NC}"

if [ "$CRITICAL_ZONE_ALERT" = true ]; then
    echo -e "${RED}🚨 КРИТИЧЕСКОЕ СОСТОЯНИЕ: Сервер в критической зоне${NC}"
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠️ ПРЕДУПРЕЖДЕНИЯ: $WARNINGS проблем обнаружено${NC}"
    exit 2
else
    echo -e "${GREEN}✅ СЕРВЕР В НОРМЕ: Все проверки пройдены${NC}"
    exit 0
fi
