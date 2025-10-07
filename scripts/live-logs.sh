#!/bin/bash

# Скрипт для просмотра логов 2GC Network Protocol Suite в реальном времени

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Live Logs${NC}"
echo -e "${BLUE}==========================================${NC}"
echo ""

# Проверяем, запущен ли сервер
if ! docker ps | grep -q "2gc-network-server"; then
    echo -e "${RED}❌ QUIC сервер не запущен${NC}"
    echo "Запустите сервер: ./scripts/docker-server.sh"
    exit 1
fi

echo -e "${GREEN}✅ QUIC сервер запущен${NC}"
echo -e "${YELLOW}📊 Просмотр логов в реальном времени...${NC}"
echo -e "${CYAN}Нажмите Ctrl+C для выхода${NC}"
echo ""

# Показываем логи сервера в реальном времени
docker logs -f 2gc-network-server

