#!/bin/bash

# Скрипт для демонстрации TUI дашборда

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - TUI Demo${NC}"
echo "Демонстрация TUI дашборда"
echo ""

# Проверяем, собран ли TUI
if [ ! -f "./build/tui" ]; then
    echo -e "${YELLOW}Собираем TUI компонент...${NC}"
    cd cmd/tui
    go build -o ../../build/tui .
    cd ../..
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Ошибка сборки TUI компонента${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ TUI компонент готов${NC}"
echo ""

# Показываем варианты запуска
echo -e "${YELLOW}Выберите режим демонстрации:${NC}"
echo "1. Демо режим (автоматические данные)"
echo "2. Реальные данные (JSON из stdin)"
echo "3. Интеграция с QUIC тестером"
echo ""

read -p "Введите номер (1-3): " choice

case $choice in
    1)
        echo -e "${YELLOW}Запуск демо режима...${NC}"
        echo "TUI будет генерировать тестовые данные"
        echo "Нажмите Ctrl+C для выхода"
        echo ""
        ./build/tui --demo --fps 10
        ;;
    2)
        echo -e "${YELLOW}Режим реальных данных...${NC}"
        echo "Введите JSON данные в формате:"
        echo '{"latency_ms":12.3,"code":200,"cpu":37.5,"rtt_ms":15.2}'
        echo "Нажмите Ctrl+C для выхода"
        echo ""
        ./build/tui --fps 5
        ;;
    3)
        echo -e "${YELLOW}Интеграция с QUIC тестером...${NC}"
        echo "Запускаем QUIC сервер и клиент с TUI"
        echo ""
        
        # Запускаем сервер в фоне
        echo -e "${YELLOW}Запуск QUIC сервера...${NC}"
        timeout 30s docker run --rm --name 2gc-network-server --network 2gc-network-suite \
            -p 9000:9000/udp -p 2113:2113 -p 6060:6060 \
            2gc-network-suite:server &
        
        sleep 3
        
        # Запускаем клиент с выводом в TUI
        echo -e "${YELLOW}Запуск QUIC клиента с TUI...${NC}"
        docker run --rm --network host \
            -e QUIC_CLIENT_ADDR=localhost:9000 \
            -e QUIC_CONNECTIONS=1 \
            -e QUIC_STREAMS=1 \
            -e QUIC_RATE=10 \
            -e QUIC_DURATION=20s \
            -e QUIC_NO_TLS=true \
            2gc-network-suite:client | ./build/tui --fps 5
        
        # Останавливаем сервер
        docker stop 2gc-network-server 2>/dev/null
        ;;
    *)
        echo -e "${RED}Неверный выбор${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}Демонстрация завершена!${NC}"
echo ""
echo -e "${BLUE}Полезные команды:${NC}"
echo "  Сборка TUI: cd cmd/tui && go build -o ../../build/tui ."
echo "  Демо режим: ./build/tui --demo --fps 10"
echo "  Реальные данные: echo '{\"latency_ms\":12.3,\"code\":200,\"cpu\":37.5,\"rtt_ms\":15.2}' | ./build/tui"
echo "  Docker TUI: ./scripts/docker-tui.sh"

