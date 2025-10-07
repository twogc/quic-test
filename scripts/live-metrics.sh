#!/bin/bash

# Скрипт для мониторинга метрик Prometheus в реальном времени

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
    echo -e "${BLUE}  2GC Network Protocol Suite - Live Metrics${NC}"
    echo -e "${BLUE}==========================================${NC}"
    echo ""
}

# Функция для получения метрик
get_metrics() {
    local url="$1"
    if curl -s "$url" >/dev/null 2>&1; then
        curl -s "$url" 2>/dev/null
    else
        echo ""
    fi
}

# Функция для показа ключевых метрик
show_key_metrics() {
    local metrics=$(get_metrics "http://localhost:2113/metrics")
    
    if [ -n "$metrics" ]; then
        echo -e "${YELLOW}📊 Ключевые метрики QUIC сервера:${NC}"
        
        # Показываем общие метрики
        echo -e "${CYAN}  🔢 Счетчики:${NC}"
        echo "$metrics" | grep -E "quic_server_(connections|streams|bytes|errors)_total" | while read line; do
            echo "    $line"
        done
        
        echo -e "${CYAN}  ⏱️  Время работы:${NC}"
        echo "$metrics" | grep -E "quic_server_uptime_seconds" | while read line; do
            echo "    $line"
        done
        
        echo -e "${CYAN}  📈 Гистограммы:${NC}"
        echo "$metrics" | grep -E "quic_server_(latency|handshake_time)_" | head -3 | while read line; do
            echo "    $line"
        done
        
        echo -e "${CYAN}  🎯 Gauge метрики:${NC}"
        echo "$metrics" | grep -E "quic_server_(active_connections|active_streams)" | while read line; do
            echo "    $line"
        done
    else
        echo -e "${RED}❌ Метрики недоступны${NC}"
        echo "Проверьте, что сервер запущен и Prometheus endpoint работает"
    fi
    echo ""
}

# Функция для показа метрик клиента
show_client_metrics() {
    local metrics=$(get_metrics "http://localhost:2112/metrics")
    
    if [ -n "$metrics" ]; then
        echo -e "${YELLOW}📱 Метрики QUIC клиента:${NC}"
        
        echo -e "${CYAN}  🔢 Счетчики клиента:${NC}"
        echo "$metrics" | grep -E "quic_client_(test_type|data_pattern)_" | head -5 | while read line; do
            echo "    $line"
        done
        
        echo -e "${CYAN}  📊 Гистограммы клиента:${NC}"
        echo "$metrics" | grep -E "quic_client_data_pattern_duration_seconds" | head -3 | while read line; do
            echo "    $line"
        done
    else
        echo -e "${RED}❌ Метрики клиента недоступны${NC}"
    fi
    echo ""
}

# Функция для показа системных метрик
show_system_metrics() {
    echo -e "${YELLOW}💻 Системные метрики:${NC}"
    
    echo -e "${CYAN}  🐳 Docker контейнеры:${NC}"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(2gc|quic)" | while read line; do
        echo "    $line"
    done
    
    echo -e "${CYAN}  📊 Использование памяти:${NC}"
    free -h | grep -E "Mem:" | awk '{print "    Память: " $3 "/" $2 " (" $5 ")"}'
    
    echo -e "${CYAN}  💾 Использование диска:${NC}"
    df -h / | tail -1 | awk '{print "    Диск: " $3 "/" $2 " (" $5 ")"}'
    
    echo -e "${CYAN}  🌐 Сетевые соединения:${NC}"
    ss -tuln | grep -E ":9000|:2113|:6060" | while read line; do
        echo "    $line"
    done
    echo ""
}

# Основной цикл мониторинга
main() {
    local refresh_interval=${1:-5}  # Интервал обновления в секундах
    
    echo -e "${GREEN}🚀 Запуск мониторинга метрик в реальном времени (обновление каждые ${refresh_interval}с)${NC}"
    echo -e "${YELLOW}Нажмите Ctrl+C для выхода${NC}"
    echo ""
    
    while true; do
        clear_screen
        
        # Показываем время
        echo -e "${CYAN}🕐 Время: $(date)${NC}"
        echo ""
        
        # Показываем все секции
        show_key_metrics
        show_client_metrics
        show_system_metrics
        
        # Показываем команды для управления
        echo -e "${BLUE}🔧 Полезные команды:${NC}"
        echo "  Просмотр всех метрик: curl http://localhost:2113/metrics"
        echo "  Просмотр метрик клиента: curl http://localhost:2112/metrics"
        echo "  Prometheus UI: http://localhost:9090"
        echo "  Grafana UI: http://localhost:3000"
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

