#!/bin/bash

# Скрипт для мониторинга 2GC Network Protocol Suite в реальном времени

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функция для очистки экрана
clear_screen() {
    clear
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${BLUE}  2GC Network Protocol Suite - Live Monitor${NC}"
    echo -e "${BLUE}==========================================${NC}"
    echo ""
}

# Функция для показа статуса сервера
show_server_status() {
    echo -e "${YELLOW}🖥️  QUIC Сервер:${NC}"
    if docker ps | grep -q "2gc-network-server"; then
        echo -e "${GREEN}  ✅ Запущен${NC}"
        # Показываем последние логи сервера
        echo -e "${CYAN}  📊 Последние метрики:${NC}"
        docker logs --tail 3 2gc-network-server 2>/dev/null | grep -E "(Connections|Streams|Bytes|Errors|Uptime)" | tail -1
    else
        echo -e "${RED}  ❌ Не запущен${NC}"
    fi
    echo ""
}

# Функция для показа метрик Prometheus
show_prometheus_metrics() {
    echo -e "${YELLOW}📊 Prometheus метрики:${NC}"
    if curl -s http://localhost:2113/metrics >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Доступны${NC}"
        # Показываем ключевые метрики
        echo -e "${CYAN}  📈 Ключевые метрики:${NC}"
        curl -s http://localhost:2113/metrics 2>/dev/null | grep -E "(quic_server_|quic_client_)" | head -5 | while read line; do
            echo "    $line"
        done
    else
        echo -e "${RED}  ❌ Недоступны${NC}"
    fi
    echo ""
}

# Функция для показа последних отчетов
show_recent_reports() {
    echo -e "${YELLOW}📄 Последние отчеты:${NC}"
    local reports=$(ls -t *.md *.json 2>/dev/null | grep -E "(report|test)" | head -3)
    if [ -n "$reports" ]; then
        for report in $reports; do
            local size=$(ls -lh "$report" 2>/dev/null | awk '{print $5}')
            local time=$(ls -l "$report" 2>/dev/null | awk '{print $6, $7, $8}')
            echo -e "${CYAN}  📄 $report ($size, $time)${NC}"
            
            # Показываем краткую статистику для JSON отчетов
            if [[ "$report" == *.json ]]; then
                if command -v jq &> /dev/null; then
                    local success=$(jq -r '.metrics.Success' "$report" 2>/dev/null)
                    local errors=$(jq -r '.metrics.Errors' "$report" 2>/dev/null)
                    local bytes=$(jq -r '.metrics.BytesSent' "$report" 2>/dev/null)
                    echo "    ✅ Успешно: $success, ❌ Ошибки: $errors, 📦 Байт: $bytes"
                fi
            fi
        done
    else
        echo -e "${RED}  ❌ Отчеты не найдены${NC}"
    fi
    echo ""
}

# Функция для показа сетевой статистики
show_network_stats() {
    echo -e "${YELLOW}🌐 Сетевая статистика:${NC}"
    echo -e "${CYAN}  📡 UDP порт 9000:${NC}"
    if sudo ss -ulpn | grep -q ":9000"; then
        echo -e "${GREEN}    ✅ Слушает${NC}"
    else
        echo -e "${RED}    ❌ Не слушает${NC}"
    fi
    
    echo -e "${CYAN}  🔥 UFW статус:${NC}"
    if sudo ufw status | grep -q "9000/udp"; then
        echo -e "${GREEN}    ✅ Порт 9000/udp открыт${NC}"
    else
        echo -e "${RED}    ❌ Порт 9000/udp закрыт${NC}"
    fi
    
    echo -e "${CYAN}  🌍 Внешний IP:${NC}"
    local external_ip=$(curl -s ifconfig.me 2>/dev/null || echo "Недоступен")
    echo "    $external_ip:9000"
    echo ""
}

# Функция для показа использования ресурсов
show_resource_usage() {
    echo -e "${YELLOW}💻 Использование ресурсов:${NC}"
    echo -e "${CYAN}  🐳 Docker контейнеры:${NC}"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(2gc|quic)" | while read line; do
        echo "    $line"
    done
    
    echo -e "${CYAN}  📊 Использование памяти:${NC}"
    free -h | grep -E "(Mem|Swap)" | while read line; do
        echo "    $line"
    done
    
    echo -e "${CYAN}  💾 Использование диска:${NC}"
    df -h / | tail -1 | awk '{print "    Диск: " $3 "/" $2 " (" $5 ")"}'
    echo ""
}

# Основной цикл мониторинга
main() {
    local refresh_interval=${1:-5}  # Интервал обновления в секундах (по умолчанию 5)
    
    echo -e "${GREEN}🚀 Запуск мониторинга в реальном времени (обновление каждые ${refresh_interval}с)${NC}"
    echo -e "${YELLOW}Нажмите Ctrl+C для выхода${NC}"
    echo ""
    
    while true; do
        clear_screen
        
        # Показываем время
        echo -e "${CYAN}🕐 Время: $(date)${NC}"
        echo ""
        
        # Показываем все секции
        show_server_status
        show_prometheus_metrics
        show_recent_reports
        show_network_stats
        show_resource_usage
        
        # Показываем команды для управления
        echo -e "${BLUE}🔧 Полезные команды:${NC}"
        echo "  Просмотр логов сервера: docker logs -f 2gc-network-server"
        echo "  Остановка сервера: docker stop 2gc-network-server"
        echo "  Запуск клиента: ./scripts/docker-client.sh"
        echo "  Просмотр отчетов: ./scripts/view-reports.sh"
        echo ""
        
        # Ждем перед следующим обновлением
        sleep "$refresh_interval"
    done
}

# Обработка аргументов
case "${1:-}" in
    -h|--help)
        echo "Использование: $0 [интервал_в_секундах]"
        echo "Примеры:"
        echo "  $0        # Обновление каждые 5 секунд"
        echo "  $0 2      # Обновление каждые 2 секунды"
        echo "  $0 10     # Обновление каждые 10 секунд"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac

