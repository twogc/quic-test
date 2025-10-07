#!/bin/bash

# RTT Test Script for QUIC Performance Testing
# ===========================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода с цветом
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Конфигурация тестов
RTT_VALUES=(5 10 25 50 100 200 500)  # RTT в миллисекундах
ALGORITHMS=("cubic" "bbrv2")          # Алгоритмы congestion control
TEST_DURATION=30                      # Длительность теста в секундах
OUTPUT_DIR="./performance-results"    # Директория для результатов

# Функция для проверки зависимостей
check_dependencies() {
    print_color $BLUE "🔍 Checking dependencies..."
    
    # Проверяем наличие quic-test-experimental
    if [ ! -f "./quic-test-experimental" ]; then
        print_color $YELLOW "⚠️  quic-test-experimental not found, building..."
        go build -o quic-test-experimental ./cmd/experimental/
    fi
    
    # Проверяем наличие tc (traffic control)
    if ! command -v tc &> /dev/null; then
        print_color $RED "❌ tc (traffic control) is not installed"
        print_color $YELLOW "Install with: sudo apt-get install iproute2"
        exit 1
    fi
    
    print_color $GREEN "✅ Dependencies OK"
}

# Функция для настройки RTT с помощью tc
setup_rtt() {
    local rtt_ms=$1
    local interface="lo"  # Используем loopback интерфейс
    
    print_color $BLUE "🔧 Setting up RTT: ${rtt_ms}ms"
    
    # Очищаем предыдущие правила
    sudo tc qdisc del dev $interface root 2>/dev/null || true
    
    # Добавляем задержку
    sudo tc qdisc add dev $interface root netem delay ${rtt_ms}ms
    
    print_color $GREEN "✅ RTT set to ${rtt_ms}ms"
}

# Функция для очистки RTT настроек
cleanup_rtt() {
    local interface="lo"
    
    print_color $BLUE "🧹 Cleaning up RTT settings..."
    
    # Очищаем правила tc
    sudo tc qdisc del dev $interface root 2>/dev/null || true
    
    print_color $GREEN "✅ RTT settings cleaned up"
}

# Функция для запуска теста с заданным RTT и алгоритмом
run_rtt_test() {
    local rtt_ms=$1
    local algorithm=$2
    local test_id="rtt_${rtt_ms}ms_${algorithm}"
    local test_dir="${OUTPUT_DIR}/${test_id}"
    
    print_color $YELLOW "🔄 Running test: RTT=${rtt_ms}ms, Algorithm=${algorithm}"
    
    # Создаем директорию для теста
    mkdir -p "$test_dir"
    
    # Настраиваем RTT
    setup_rtt $rtt_ms
    
    # Запускаем сервер
    print_color $BLUE "🚀 Starting server with ${algorithm}..."
    nohup ./quic-test-experimental \
        --mode server \
        --cc $algorithm \
        --qlog "${test_dir}/server-qlog" \
        --verbose \
        --metrics-interval 1s \
        > "${test_dir}/server.log" 2>&1 &
    
    local server_pid=$!
    print_color $GREEN "✅ Server started (PID: $server_pid)"
    
    # Ждем запуска сервера
    sleep 3
    
    # Запускаем клиент
    print_color $BLUE "🔗 Starting client..."
    timeout ${TEST_DURATION}s ./quic-test-experimental \
        --mode client \
        --addr 127.0.0.1:9000 \
        --cc $algorithm \
        --qlog "${test_dir}/client-qlog" \
        --duration ${TEST_DURATION}s \
        --connections 1 \
        --streams 1 \
        --rate 100 \
        --packet-size 1200 \
        --verbose \
        > "${test_dir}/client.log" 2>&1 &
    
    local client_pid=$!
    
    # Ждем завершения клиента
    wait $client_pid 2>/dev/null || true
    
    # Останавливаем сервер
    kill $server_pid 2>/dev/null || true
    wait $server_pid 2>/dev/null || true
    
    # Очищаем RTT настройки
    cleanup_rtt
    
    # Пауза между тестами
    sleep 2
    
    print_color $GREEN "✅ Test completed: ${test_id}"
}

# Функция для анализа результатов
analyze_results() {
    print_color $BLUE "📊 Analyzing test results..."
    
    local analysis_file="${OUTPUT_DIR}/rtt_analysis.txt"
    
    cat > "$analysis_file" << EOF
RTT Performance Test Analysis
============================

Generated: $(date)
Test Duration: ${TEST_DURATION}s per test
Total Tests: $((${#RTT_VALUES[@]} * ${#ALGORITHMS[@]}))

Test Results:
EOF
    
    # Анализируем каждый тест
    for rtt in "${RTT_VALUES[@]}"; do
        for algorithm in "${ALGORITHMS[@]}"; do
            local test_id="rtt_${rtt}ms_${algorithm}"
            local test_dir="${OUTPUT_DIR}/${test_id}"
            
            if [ -d "$test_dir" ]; then
                echo "" >> "$analysis_file"
                echo "Test: ${test_id}" >> "$analysis_file"
                echo "RTT: ${rtt}ms" >> "$analysis_file"
                echo "Algorithm: ${algorithm}" >> "$analysis_file"
                echo "Files:" >> "$analysis_file"
                echo "  - Server log: ${test_dir}/server.log" >> "$analysis_file"
                echo "  - Client log: ${test_dir}/client.log" >> "$analysis_file"
                echo "  - Server qlog: ${test_dir}/server-qlog/" >> "$analysis_file"
                echo "  - Client qlog: ${test_dir}/client-qlog/" >> "$analysis_file"
            fi
        done
    done
    
    print_color $GREEN "✅ Analysis saved to: $analysis_file"
}

# Функция для генерации отчета
generate_report() {
    print_color $BLUE "📋 Generating performance report..."
    
    local report_file="${OUTPUT_DIR}/rtt_performance_report.md"
    
    cat > "$report_file" << EOF
# RTT Performance Test Report

**Generated:** $(date)
**Test Duration:** ${TEST_DURATION}s per test
**Total Tests:** $((${#RTT_VALUES[@]} * ${#ALGORITHMS[@]}))

## Test Configuration

### RTT Values
$(printf '%s\n' "${RTT_VALUES[@]}" | sed 's/^/- /')

### Algorithms
$(printf '%s\n' "${ALGORITHMS[@]}" | sed 's/^/- /')

## Test Results

| RTT (ms) | Algorithm | Status | Files |
|----------|-----------|--------|-------|
EOF
    
    # Добавляем результаты в таблицу
    for rtt in "${RTT_VALUES[@]}"; do
        for algorithm in "${ALGORITHMS[@]}"; do
            local test_id="rtt_${rtt}ms_${algorithm}"
            local test_dir="${OUTPUT_DIR}/${test_id}"
            
            if [ -d "$test_dir" ]; then
                echo "| ${rtt} | ${algorithm} | ✅ Completed | \`${test_dir}/\` |" >> "$report_file"
            else
                echo "| ${rtt} | ${algorithm} | ❌ Failed | - |" >> "$report_file"
            fi
        done
    done
    
    cat >> "$report_file" << EOF

## Analysis

### Key Metrics to Compare
1. **RTT Impact on Throughput**
   - How does RTT affect goodput for each algorithm?
   - Which algorithm performs better at high RTT?

2. **Congestion Control Behavior**
   - How does each algorithm adapt to different RTT values?
   - Which algorithm is more stable across RTT ranges?

3. **Connection Establishment**
   - How does RTT affect handshake time?
   - Which algorithm establishes connections faster?

### Next Steps
1. Analyze qlog files with qvis
2. Extract numerical metrics from logs
3. Create performance comparison charts
4. Identify optimal algorithm for different RTT ranges

## Files Structure

\`\`\`
${OUTPUT_DIR}/
├── rtt_5ms_cubic/
├── rtt_5ms_bbrv2/
├── rtt_10ms_cubic/
├── rtt_10ms_bbrv2/
├── ...
├── rtt_analysis.txt
└── rtt_performance_report.md
\`\`\`

EOF
    
    print_color $GREEN "✅ Report generated: $report_file"
}

# Функция для показа справки
show_help() {
    cat << EOF
RTT Test Script for QUIC Performance Testing
===========================================

Usage: $0 [OPTIONS]

OPTIONS:
  --rtt VALUES        - RTT values to test (comma-separated, default: 5,10,25,50,100,200,500)
  --algorithms VALUES  - Algorithms to test (comma-separated, default: cubic,bbrv2)
  --duration SECONDS  - Test duration per test (default: 30)
  --output DIR        - Output directory (default: ./performance-results)
  --cleanup           - Clean up previous results before running
  --analysis-only     - Only analyze existing results
  --help              - Show this help

EXAMPLES:
  $0                                    # Run all tests with default settings
  $0 --rtt 10,50,100 --algorithms bbrv2 # Test specific RTT values with BBRv2 only
  $0 --duration 60 --cleanup            # Run 60-second tests, clean first
  $0 --analysis-only                    # Analyze existing results

EOF
}

# Основная логика
main() {
    local cleanup_flag=false
    local analysis_only=false
    
    # Парсим аргументы
    while [[ $# -gt 0 ]]; do
        case $1 in
            --rtt)
                IFS=',' read -ra RTT_VALUES <<< "$2"
                shift 2
                ;;
            --algorithms)
                IFS=',' read -ra ALGORITHMS <<< "$2"
                shift 2
                ;;
            --duration)
                TEST_DURATION=$2
                shift 2
                ;;
            --output)
                OUTPUT_DIR=$2
                shift 2
                ;;
            --cleanup)
                cleanup_flag=true
                shift
                ;;
            --analysis-only)
                analysis_only=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_color $RED "❌ Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    print_color $GREEN "🧪 RTT Performance Test Suite"
    print_color $GREEN "============================="
    print_color $BLUE "RTT Values: ${RTT_VALUES[*]}"
    print_color $BLUE "Algorithms: ${ALGORITHMS[*]}"
    print_color $BLUE "Duration: ${TEST_DURATION}s"
    print_color $BLUE "Output: ${OUTPUT_DIR}"
    
    # Проверяем зависимости
    check_dependencies
    
    # Очистка при необходимости
    if [ "$cleanup_flag" = true ]; then
        print_color $BLUE "🧹 Cleaning up previous results..."
        rm -rf "$OUTPUT_DIR" 2>/dev/null || true
    fi
    
    # Создаем директорию для результатов
    mkdir -p "$OUTPUT_DIR"
    
    if [ "$analysis_only" = true ]; then
        analyze_results
        generate_report
        return
    fi
    
    # Запускаем тесты
    local test_count=0
    local total_tests=$((${#RTT_VALUES[@]} * ${#ALGORITHMS[@]}))
    
    for rtt in "${RTT_VALUES[@]}"; do
        for algorithm in "${ALGORITHMS[@]}"; do
            test_count=$((test_count + 1))
            print_color $YELLOW "🔄 Test ${test_count}/${total_tests}: RTT=${rtt}ms, Algorithm=${algorithm}"
            
            run_rtt_test $rtt $algorithm
        done
    done
    
    # Анализируем результаты
    analyze_results
    generate_report
    
    print_color $GREEN "🎉 RTT performance testing completed!"
    print_color $BLUE "📁 Results available in: ${OUTPUT_DIR}/"
}

# Запускаем основную функцию
main "$@"

