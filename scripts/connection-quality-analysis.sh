#!/bin/bash

# Скрипт для анализа качества подключения клиента к серверу

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Quality Analysis${NC}"
echo "Анализ качества подключения клиента к серверу"
echo ""

# Функция для проверки сервера
check_server() {
    echo -e "${YELLOW}🔍 Проверка сервера...${NC}"
    
    if docker ps | grep -q "2gc-network-server"; then
        echo -e "${GREEN}✅ QUIC сервер запущен${NC}"
        
        # Получаем статистику сервера
        echo -e "${CYAN}📊 Статистика сервера:${NC}"
        docker logs --tail 3 2gc-network-server 2>/dev/null | grep -E "(Connections|Streams|Bytes|Errors|Uptime)" | tail -1
        
        # Проверяем порты
        echo -e "${CYAN}🌐 Сетевые порты:${NC}"
        if sudo ss -ulpn | grep -q ":9000"; then
            echo -e "${GREEN}  ✅ UDP порт 9000 слушает${NC}"
        else
            echo -e "${RED}  ❌ UDP порт 9000 не слушает${NC}"
        fi
        
        if curl -s http://localhost:2113/metrics >/dev/null 2>&1; then
            echo -e "${GREEN}  ✅ Prometheus метрики доступны${NC}"
        else
            echo -e "${RED}  ❌ Prometheus метрики недоступны${NC}"
        fi
        
        return 0
    else
        echo -e "${RED}❌ QUIC сервер не запущен${NC}"
        echo "Запустите сервер: ./scripts/docker-server.sh"
        return 1
    fi
}

# Функция для анализа сетевого подключения
analyze_network() {
    echo -e "${YELLOW}🌐 Анализ сетевого подключения...${NC}"
    
    # Проверяем UFW
    echo -e "${CYAN}🔥 Файрвол (UFW):${NC}"
    if sudo ufw status | grep -q "9000/udp"; then
        echo -e "${GREEN}  ✅ Порт 9000/udp открыт${NC}"
    else
        echo -e "${RED}  ❌ Порт 9000/udp закрыт${NC}"
    fi
    
    # Проверяем внешний IP
    local external_ip=$(curl -s ifconfig.me 2>/dev/null || echo "Недоступен")
    echo -e "${CYAN}🌍 Внешний IP: ${external_ip}${NC}"
    
    # Проверяем локальное подключение
    echo -e "${CYAN}🔗 Локальное подключение:${NC}"
    if timeout 3s nc -u -z localhost 9000 2>/dev/null; then
        echo -e "${GREEN}  ✅ Локальное UDP подключение работает${NC}"
    else
        echo -e "${RED}  ❌ Локальное UDP подключение не работает${NC}"
    fi
}

# Функция для тестирования клиента
test_client_connection() {
    echo -e "${YELLOW}🧪 Тестирование подключения клиента...${NC}"
    
    # Параметры теста
    local connections=${QUIC_CONNECTIONS:-1}
    local streams=${QUIC_STREAMS:-1}
    local rate=${QUIC_RATE:-10}
    local duration=${QUIC_DURATION:-10s}
    
    echo -e "${CYAN}📋 Параметры теста:${NC}"
    echo "  Соединения: $connections"
    echo "  Потоки: $streams"
    echo "  Скорость: $rate пакетов/сек"
    echo "  Длительность: $duration"
    echo ""
    
    # Запускаем клиент с детальным выводом
    echo -e "${YELLOW}🚀 Запуск клиента...${NC}"
    timeout 15s docker run --rm --network host \
        -e QUIC_CLIENT_ADDR=localhost:9000 \
        -e QUIC_CONNECTIONS=$connections \
        -e QUIC_STREAMS=$streams \
        -e QUIC_RATE=$rate \
        -e QUIC_DURATION=$duration \
        -e QUIC_NO_TLS=true \
        2gc-network-suite:client
    
    local client_exit_code=$?
    
    if [ $client_exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ Клиент успешно завершил работу${NC}"
    else
        echo -e "${RED}❌ Клиент завершился с ошибкой (код: $client_exit_code)${NC}"
    fi
    
    return $client_exit_code
}

# Функция для анализа метрик
analyze_metrics() {
    echo -e "${YELLOW}📊 Анализ метрик...${NC}"
    
    # Проверяем метрики сервера
    if curl -s http://localhost:2113/metrics >/dev/null 2>&1; then
        echo -e "${CYAN}📈 Метрики сервера:${NC}"
        curl -s http://localhost:2113/metrics 2>/dev/null | grep -E "quic_server_(connections|streams|bytes|errors)_total" | while read line; do
            echo "  $line"
        done
        
        echo -e "${CYAN}⏱️ Время работы сервера:${NC}"
        curl -s http://localhost:2113/metrics 2>/dev/null | grep "quic_server_uptime_seconds" | while read line; do
            echo "  $line"
        done
    else
        echo -e "${RED}❌ Метрики сервера недоступны${NC}"
    fi
}

# Функция для анализа отчетов
analyze_reports() {
    echo -e "${YELLOW}📄 Анализ отчетов...${NC}"
    
    # Ищем последний отчет
    local latest_report=$(ls -t *.md *.json 2>/dev/null | grep -E "(report|test)" | head -1)
    
    if [ -n "$latest_report" ]; then
        echo -e "${CYAN}📋 Последний отчет: $latest_report${NC}"
        
        if [[ "$latest_report" == *.json ]]; then
            if command -v jq &> /dev/null; then
                echo -e "${CYAN}📊 Статистика из отчета:${NC}"
                echo "  Успешные соединения: $(jq -r '.metrics.Success' "$latest_report")"
                echo "  Ошибки: $(jq -r '.metrics.Errors' "$latest_report")"
                echo "  Отправлено байт: $(jq -r '.metrics.BytesSent' "$latest_report")"
                echo "  Потеря пакетов: $(jq -r '.metrics.PacketLoss' "$latest_report")"
                echo "  Повторные передачи: $(jq -r '.metrics.Retransmits' "$latest_report")"
                
                # Анализ времени handshake
                local handshake_times=$(jq -r '.metrics.HandshakeTimes | join(", ")' "$latest_report")
                if [ "$handshake_times" != "null" ] && [ -n "$handshake_times" ]; then
                    echo "  Время handshake: $handshake_times мс"
                fi
            else
                echo "  (Установите jq для детального анализа: sudo apt install jq)"
            fi
        else
            echo "  (Markdown отчет - полный текст)"
        fi
    else
        echo -e "${RED}❌ Отчеты не найдены${NC}"
    fi
}

# Функция для генерации рекомендаций
generate_recommendations() {
    echo -e "${YELLOW}💡 Рекомендации по улучшению качества подключения:${NC}"
    echo ""
    
    echo -e "${CYAN}🔧 Сетевые настройки:${NC}"
    echo "  1. Проверьте MTU размер: ip link show"
    echo "  2. Оптимизируйте UDP буферы: sysctl net.core.rmem_max"
    echo "  3. Настройте congestion control: sysctl net.ipv4.tcp_congestion_control"
    echo ""
    
    echo -e "${CYAN}📊 Мониторинг:${NC}"
    echo "  1. Используйте Prometheus: http://localhost:9090"
    echo "  2. Настройте Grafana: http://localhost:3000"
    echo "  3. Анализируйте трейсы в Jaeger: http://localhost:16686"
    echo ""
    
    echo -e "${CYAN}🧪 Тестирование:${NC}"
    echo "  1. Увеличьте количество соединений: QUIC_CONNECTIONS=5"
    echo "  2. Тестируйте разные скорости: QUIC_RATE=50,100,200"
    echo "  3. Проверьте с TLS: убрать QUIC_NO_TLS=true"
    echo ""
    
    echo -e "${CYAN}📈 Анализ производительности:${NC}"
    echo "  1. Мониторинг в реальном времени: ./scripts/live-monitor.sh"
    echo "  2. Просмотр логов: ./scripts/live-logs.sh"
    echo "  3. Анализ метрик: ./scripts/live-metrics.sh"
}

# Основная функция
main() {
    echo -e "${GREEN}🚀 Начинаем анализ качества подключения...${NC}"
    echo ""
    
    # Проверяем сервер
    if ! check_server; then
        echo -e "${RED}❌ Анализ прерван: сервер не запущен${NC}"
        exit 1
    fi
    
    echo ""
    
    # Анализируем сеть
    analyze_network
    
    echo ""
    
    # Тестируем клиента
    test_client_connection
    local test_result=$?
    
    echo ""
    
    # Анализируем метрики
    analyze_metrics
    
    echo ""
    
    # Анализируем отчеты
    analyze_reports
    
    echo ""
    
    # Генерируем рекомендации
    generate_recommendations
    
    echo ""
    echo -e "${BLUE}==========================================${NC}"
    
    if [ $test_result -eq 0 ]; then
        echo -e "${GREEN}✅ Анализ завершен успешно${NC}"
        echo -e "${GREEN}🎯 Качество подключения: ХОРОШЕЕ${NC}"
    else
        echo -e "${RED}❌ Анализ выявил проблемы${NC}"
        echo -e "${RED}⚠️ Качество подключения: ТРЕБУЕТ УЛУЧШЕНИЯ${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Полезные команды:${NC}"
    echo "  Повторный анализ: $0"
    echo "  Мониторинг в реальном времени: ./scripts/live-monitor.sh"
    echo "  Просмотр логов: ./scripts/live-logs.sh"
    echo "  Анализ метрик: ./scripts/live-metrics.sh"
}

# Запуск анализа
main "$@"

