#!/bin/bash

# Оптимизированный скрипт для удаленного QUIC клиента
# Решает проблемы с высоким jitter и ошибками

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Optimized Remote Client${NC}"
echo "Оптимизированное подключение для удаленного сервера"
echo ""

# Параметры для удаленного подключения
SERVER_IP="212.233.79.160"
SERVER_PORT="9000"
CONNECTIONS=${QUIC_CONNECTIONS:-1}
STREAMS=${QUIC_STREAMS:-1}
RATE=${QUIC_RATE:-20}  # Снижена частота для стабильности
DURATION=${QUIC_DURATION:-30s}
PACKET_SIZE=${QUIC_PACKET_SIZE:-1200}

echo -e "${YELLOW}🔧 Оптимизированные параметры для удаленного подключения:${NC}"
echo "  Сервер: ${SERVER_IP}:${SERVER_PORT}"
echo "  Соединения: ${CONNECTIONS}"
echo "  Потоки: ${STREAMS}"
echo "  Частота: ${RATE} пакетов/сек (снижена для стабильности)"
echo "  Размер пакета: ${PACKET_SIZE} байт"
echo "  Длительность: ${DURATION}"
echo ""

# Функция для проверки доступности сервера
check_server_availability() {
    echo -e "${YELLOW}🔍 Проверка доступности сервера...${NC}"
    
    # Проверяем UDP порт (ping может быть отключен)
    if timeout 5s nc -u -z ${SERVER_IP} ${SERVER_PORT} 2>/dev/null; then
        echo -e "${GREEN}✅ UDP порт ${SERVER_IP}:${SERVER_PORT} доступен${NC}"
        return 0
    else
        echo -e "${RED}❌ UDP порт ${SERVER_IP}:${SERVER_PORT} недоступен${NC}"
        return 1
    fi
}

# Функция для оптимизации сетевых параметров
optimize_network_settings() {
    echo -e "${YELLOW}🔧 Оптимизация сетевых параметров...${NC}"
    
    # Увеличиваем UDP буферы для лучшей производительности
    echo -e "${CYAN}📡 Настройка UDP буферов:${NC}"
    sudo sysctl -w net.core.rmem_max=4194304 >/dev/null 2>&1
    sudo sysctl -w net.core.rmem_default=4194304 >/dev/null 2>&1
    sudo sysctl -w net.core.wmem_max=4194304 >/dev/null 2>&1
    sudo sysctl -w net.core.wmem_default=4194304 >/dev/null 2>&1
    
    echo "  ✅ UDP буферы увеличены до 4MB"
    
    # Настройка TCP параметров для стабильности
    echo -e "${CYAN}🌐 Настройка TCP параметров:${NC}"
    sudo sysctl -w net.ipv4.tcp_congestion_control=bbr >/dev/null 2>&1
    sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 4194304" >/dev/null 2>&1
    sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 4194304" >/dev/null 2>&1
    
    echo "  ✅ TCP congestion control: BBR"
    echo "  ✅ TCP буферы оптимизированы"
}

# Функция для запуска оптимизированного клиента
run_optimized_client() {
    echo -e "${YELLOW}🚀 Запуск оптимизированного клиента...${NC}"
    
    # Переменные окружения для оптимизации
    export QUIC_CLIENT_ADDR="${SERVER_IP}:${SERVER_PORT}"
    export QUIC_CONNECTIONS="${CONNECTIONS}"
    export QUIC_STREAMS="${STREAMS}"
    export QUIC_RATE="${RATE}"
    export QUIC_DURATION="${DURATION}"
    export QUIC_PACKET_SIZE="${PACKET_SIZE}"
    export QUIC_NO_TLS="true"
    
    # Дополнительные параметры для стабильности
    export QUIC_HANDSHAKE_TIMEOUT="30s"
    export QUIC_MAX_IDLE_TIMEOUT="60s"
    export QUIC_KEEP_ALIVE="30s"
    
    echo -e "${CYAN}📋 Параметры клиента:${NC}"
    echo "  Адрес: ${QUIC_CLIENT_ADDR}"
    echo "  Соединения: ${QUIC_CONNECTIONS}"
    echo "  Потоки: ${QUIC_STREAMS}"
    echo "  Частота: ${QUIC_RATE} пакетов/сек"
    echo "  Длительность: ${QUIC_DURATION}"
    echo "  Размер пакета: ${QUIC_PACKET_SIZE} байт"
    echo "  TLS: отключен (для стабильности)"
    echo ""
    
    # Запускаем клиент с оптимизированными параметрами
    timeout 60s docker run --rm --network host \
        -e QUIC_CLIENT_ADDR="${QUIC_CLIENT_ADDR}" \
        -e QUIC_CONNECTIONS="${QUIC_CONNECTIONS}" \
        -e QUIC_STREAMS="${QUIC_STREAMS}" \
        -e QUIC_RATE="${QUIC_RATE}" \
        -e QUIC_DURATION="${QUIC_DURATION}" \
        -e QUIC_PACKET_SIZE="${QUIC_PACKET_SIZE}" \
        -e QUIC_NO_TLS="${QUIC_NO_TLS}" \
        -e QUIC_HANDSHAKE_TIMEOUT="${QUIC_HANDSHAKE_TIMEOUT}" \
        -e QUIC_MAX_IDLE_TIMEOUT="${QUIC_MAX_IDLE_TIMEOUT}" \
        -e QUIC_KEEP_ALIVE="${QUIC_KEEP_ALIVE}" \
        2gc-network-suite:client
    
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ Клиент успешно завершил работу${NC}"
    elif [ $exit_code -eq 124 ]; then
        echo -e "${YELLOW}⏰ Клиент завершен по таймауту (60s)${NC}"
    else
        echo -e "${RED}❌ Клиент завершился с ошибкой (код: $exit_code)${NC}"
    fi
    
    return $exit_code
}

# Функция для анализа результатов
analyze_results() {
    echo -e "${YELLOW}📊 Анализ результатов...${NC}"
    
    # Ищем последний отчет
    local latest_report=$(ls -t *.md *.json 2>/dev/null | grep -E "(report|test)" | head -1)
    
    if [ -n "$latest_report" ]; then
        echo -e "${CYAN}📋 Анализ отчета: $latest_report${NC}"
        
        if [[ "$latest_report" == *.json ]] && command -v jq &> /dev/null; then
            local success=$(jq -r '.metrics.Success' "$latest_report")
            local errors=$(jq -r '.metrics.Errors' "$latest_report")
            local bytes=$(jq -r '.metrics.BytesSent' "$latest_report")
            local total=$((success + errors))
            
            if [ $total -gt 0 ]; then
                local error_rate=$((errors * 100 / total))
                echo "  ✅ Успешных пакетов: $success"
                echo "  ❌ Ошибок: $errors"
                echo "  📦 Отправлено данных: $bytes KB"
                echo "  📊 Процент ошибок: ${error_rate}%"
                
                if [ $error_rate -lt 5 ]; then
                    echo -e "${GREEN}  🎯 Качество подключения: ОТЛИЧНОЕ (< 5% ошибок)${NC}"
                elif [ $error_rate -lt 15 ]; then
                    echo -e "${YELLOW}  ⚠️ Качество подключения: ХОРОШЕЕ (5-15% ошибок)${NC}"
                else
                    echo -e "${RED}  🚨 Качество подключения: ТРЕБУЕТ УЛУЧШЕНИЯ (> 15% ошибок)${NC}"
                fi
            fi
        fi
    fi
}

# Функция для генерации рекомендаций
generate_recommendations() {
    echo -e "${YELLOW}💡 Рекомендации для дальнейшего улучшения:${NC}"
    echo ""
    
    echo -e "${CYAN}🔧 Дополнительная оптимизация:${NC}"
    echo "  1. Увеличить частоту постепенно: QUIC_RATE=30,50,100"
    echo "  2. Тестировать с TLS: убрать QUIC_NO_TLS=true"
    echo "  3. Добавить больше соединений: QUIC_CONNECTIONS=2,3,5"
    echo ""
    
    echo -e "${CYAN}📊 Мониторинг:${NC}"
    echo "  1. Использовать Prometheus метрики: --prometheus"
    echo "  2. Анализировать логи в реальном времени"
    echo "  3. Настроить алерты на высокий уровень ошибок"
    echo ""
    
    echo -e "${CYAN}🌐 Сетевые улучшения:${NC}"
    echo "  1. Проверить MTU размер: ip link show"
    echo "  2. Оптимизировать congestion control"
    echo "  3. Настроить QoS для QUIC трафика"
}

# Основная функция
main() {
    echo -e "${GREEN}🚀 Запуск оптимизированного удаленного клиента...${NC}"
    echo ""
    
    # Проверяем доступность сервера
    if ! check_server_availability; then
        echo -e "${RED}❌ Сервер недоступен. Проверьте настройки сети.${NC}"
        exit 1
    fi
    
    echo ""
    
    # Оптимизируем сетевые настройки
    optimize_network_settings
    
    echo ""
    
    # Запускаем оптимизированного клиента
    run_optimized_client
    local client_result=$?
    
    echo ""
    
    # Анализируем результаты
    analyze_results
    
    echo ""
    
    # Генерируем рекомендации
    generate_recommendations
    
    echo ""
    echo -e "${BLUE}==========================================${NC}"
    
    if [ $client_result -eq 0 ]; then
        echo -e "${GREEN}✅ Оптимизированное подключение успешно${NC}"
        echo -e "${GREEN}🎯 Рекомендуется использовать эти параметры для стабильной работы${NC}"
    else
        echo -e "${YELLOW}⚠️ Подключение завершено с предупреждениями${NC}"
        echo -e "${YELLOW}💡 Попробуйте снизить частоту или увеличить таймауты${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Полезные команды:${NC}"
    echo "  Повторный тест: $0"
    echo "  С увеличенной частотой: QUIC_RATE=50 $0"
    echo "  С TLS: QUIC_NO_TLS=false $0"
    echo "  Мониторинг: ./scripts/live-monitor.sh"
}

# Запуск
main "$@"

