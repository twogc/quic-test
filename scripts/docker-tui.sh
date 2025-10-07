#!/bin/bash

# Скрипт для запуска TUI дашборда в Docker контейнере

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - TUI${NC}"
echo "Запуск TUI дашборда в Docker контейнере"
echo ""

# Проверяем наличие Docker
if ! command -v docker &> /dev/null
then
    echo -e "${RED}Ошибка: Docker не установлен${NC}"
    echo "Пожалуйста, установите Docker перед запуском скрипта."
    exit 1
fi

# Переменные окружения для конфигурации TUI
FPS=${TUI_FPS:-10}
DEMO=${TUI_DEMO:-true}
CONTAINER_NAME="2gc-network-tui"
IMAGE_NAME="2gc-network-suite:tui"

echo -e "${YELLOW}Собираем Docker образ для TUI...${NC}"
docker build -f Dockerfile.tui -t $IMAGE_NAME .

if [ $? -ne 0 ]; then
    echo -e "${RED}Ошибка сборки Docker образа TUI.${NC}"
    exit 1
fi

echo -e "${YELLOW}Запускаем TUI дашборд...${NC}"
echo -e "${YELLOW}Для выхода нажмите Ctrl+C${NC}"
echo ""

# Запускаем TUI в интерактивном режиме
docker run --rm -it \
    --name $CONTAINER_NAME \
    -e TUI_FPS=$FPS \
    -e TUI_DEMO=$DEMO \
    $IMAGE_NAME \
    --demo --fps=$FPS

if [ $? -ne 0 ]; then
    echo -e "${RED}Ошибка запуска TUI дашборда.${NC}"
    exit 1
fi

echo -e "${YELLOW}TUI дашборд завершен.${NC}"
