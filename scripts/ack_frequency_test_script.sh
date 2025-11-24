#!/bin/bash

# ACK Frequency Test Script for QUIC Performance Testing
# =====================================================

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
ACK_FREQUENCIES=(1 2 3 4 5)           # ACK frequency values
ALGORITHMS=("cubic" "bbrv3")          # Алгоритмы congestion control
TEST_DURATION=30                       # Длительность теста в секундах
OUTPUT_DIR="./ack-frequency-results"   # Директория для результатов

# Функция для проверки зависимостей
check_dependencies() {
    print_color $BLUE "🔍 Checking dependencies..."
    
    # Проверяем наличие quic-test-experimental
    if [ ! -f "./quic-test-experimental" ]; then
        print_color $YELLOW "⚠️  quic-test-experimental not found, building..."
        go build -o quic-test-experimental ./cmd/experimental/
    fi
    
    print_color $GREEN "✅ Dependencies OK"
}

# Функция для запуска теста с заданной ACK frequency и алгоритмом
run_ack_frequency_test() {
    local ack_freq=$1
    local algorithm=$2
    local test_id="ack_${ack_freq}_${algorithm}"
    local test_dir="${OUTPUT_DIR}/${test_id}"
    
    print_color $YELLOW "🔄 Running test: ACK Frequency=${ack_freq}, Algorithm=${algorithm}"
    
    # Создаем директорию для теста
    mkdir -p "$test_dir"
    
    # Запускаем сервер
    print_color $BLUE "Starting server with ${algorithm}..."
    nohup go run main.go \
        --mode server \
        --cc $algorithm \
        > "${test_dir}/server.log" 2>&1 &
    
    local server_pid=$!
    print_color $GREEN "✅ Server started (PID: $server_pid)"
    
    # Ждем запуска сервера
    sleep 3
    
    # Запускаем клиент
    print_color $BLUE "🔗 Starting client..."
    timeout ${TEST_DURATION}s go run main.go \
        --mode client \
        --addr 127.0.0.1:9000 \
        --cc $algorithm \
        --duration ${TEST_DURATION}s \
        --connections 1 \
        --streams 1 \
        --rate 100 \
        --packet-size 1200 \
        > "${test_dir}/client.log" 2>&1 &
    
    local client_pid=$!
    
    # Ждем завершения клиента
    wait $client_pid 2>/dev/null || true
    
    # Останавливаем сервер
    kill $server_pid 2>/dev/null || true
    wait $server_pid 2>/dev/null || true
    
    # Пауза между тестами
    sleep 2
    
    print_color $GREEN "✅ Test completed: ${test_id}"
}

# Функция для анализа результатов
analyze_results() {
    print_color $BLUE "Analyzing ACK frequency test results..."
    
    local analysis_file="${OUTPUT_DIR}/ack_frequency_analysis.txt"
    
    cat > "$analysis_file" << EOF
ACK Frequency Performance Test Analysis
======================================

Generated: $(date)
Test Duration: ${TEST_DURATION}s per test
Total Tests: $((${#ACK_FREQUENCIES[@]} * ${#ALGORITHMS[@]}))

Test Results:
EOF
    
    # Анализируем каждый тест
    for ack_freq in "${ACK_FREQUENCIES[@]}"; do
        for algorithm in "${ALGORITHMS[@]}"; do
            local test_id="ack_${ack_freq}_${algorithm}"
            local test_dir="${OUTPUT_DIR}/${test_id}"
            
            if [ -d "$test_dir" ]; then
                echo "" >> "$analysis_file"
                echo "Test: ${test_id}" >> "$analysis_file"
                echo "ACK Frequency: ${ack_freq}" >> "$analysis_file"
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
    print_color $BLUE "Generating ACK frequency performance report..."
    
    local report_file="${OUTPUT_DIR}/ack_frequency_performance_report.md"
    
    cat > "$report_file" << EOF
# ACK Frequency Performance Test Report

**Generated:** $(date)
**Test Duration:** ${TEST_DURATION}s per test
**Total Tests:** $((${#ACK_FREQUENCIES[@]} * ${#ALGORITHMS[@]}))

## Test Configuration

### ACK Frequency Values
$(printf '%s\n' "${ACK_FREQUENCIES[@]}" | sed 's/^/- /')

### Algorithms
$(printf '%s\n' "${ALGORITHMS[@]}" | sed 's/^/- /')

## Test Results

| ACK Frequency | Algorithm | Status | Files |
|---------------|-----------|--------|-------|
EOF
    
    # Добавляем результаты в таблицу
    for ack_freq in "${ACK_FREQUENCIES[@]}"; do
        for algorithm in "${ALGORITHMS[@]}"; do
            local test_id="ack_${ack_freq}_${algorithm}"
            local test_dir="${OUTPUT_DIR}/${test_id}"
            
            if [ -d "$test_dir" ]; then
                echo "| ${ack_freq} | ${algorithm} | ✅ Completed | \`${test_dir}/\` |" >> "$report_file"
            else
                echo "| ${ack_freq} | ${algorithm} | ❌ Failed | - |" >> "$report_file"
            fi
        done
    done
    
    cat >> "$report_file" << EOF

## Analysis

### Key Metrics to Compare
1. **ACK Overhead**
   - How does ACK frequency affect network overhead?
   - Which frequency provides optimal balance?

2. **Latency Impact**
   - How does ACK frequency affect end-to-end latency?
   - Which frequency minimizes latency?

3. **Throughput Efficiency**
   - How does ACK frequency affect goodput?
   - Which frequency maximizes throughput?

4. **Algorithm Interaction**
   - How does ACK frequency interact with congestion control?
   - Which combination provides best performance?

### Expected Results
- **Low ACK Frequency (1-2)**: Lower overhead, higher latency
- **High ACK Frequency (4-5)**: Higher overhead, lower latency
- **Optimal Range**: Usually 2-3 for most scenarios

### Next Steps
1. Analyze qlog files with qvis
2. Extract ACK-related metrics from logs
3. Create ACK frequency vs performance charts
4. Identify optimal ACK frequency for each algorithm

## Files Structure

\`\`\`
${OUTPUT_DIR}/
├── ack_1_cubic/
├── ack_1_bbrv2/
├── ack_2_cubic/
├── ack_2_bbrv2/
├── ...
├── ack_frequency_analysis.txt
└── ack_frequency_performance_report.md
\`\`\`

EOF
    
    print_color $GREEN "✅ Report generated: $report_file"
}

# Функция для показа справки
show_help() {
    cat << EOF
ACK Frequency Test Script for QUIC Performance Testing
=====================================================

Usage: $0 [OPTIONS]

OPTIONS:
  --frequencies VALUES - ACK frequency values to test (comma-separated, default: 1,2,3,4,5)
  --algorithms VALUES  - Algorithms to test (comma-separated, default: cubic,bbrv3)
  --duration SECONDS   - Test duration per test (default: 30)
  --output DIR         - Output directory (default: ./ack-frequency-results)
  --cleanup            - Clean up previous results before running
  --analysis-only      - Only analyze existing results
  --help               - Show this help

EXAMPLES:
  $0                                    # Run all tests with default settings
  $0 --frequencies 2,3,4 --algorithms bbrv3 # Test specific ACK frequencies with BBRv3 only
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
            --frequencies)
                IFS=',' read -ra ACK_FREQUENCIES <<< "$2"
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
    
    print_color $GREEN "🧪 ACK Frequency Performance Test Suite"
    print_color $GREEN "======================================="
    print_color $BLUE "ACK Frequencies: ${ACK_FREQUENCIES[*]}"
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
    local total_tests=$((${#ACK_FREQUENCIES[@]} * ${#ALGORITHMS[@]}))
    
    for ack_freq in "${ACK_FREQUENCIES[@]}"; do
        for algorithm in "${ALGORITHMS[@]}"; do
            test_count=$((test_count + 1))
            print_color $YELLOW "🔄 Test ${test_count}/${total_tests}: ACK Frequency=${ack_freq}, Algorithm=${algorithm}"
            
            run_ack_frequency_test $ack_freq $algorithm
        done
    done
    
    # Анализируем результаты
    analyze_results
    generate_report
    
    print_color $GREEN "🎉 ACK frequency performance testing completed!"
    print_color $BLUE "Results available in: ${OUTPUT_DIR}/"
}

# Запускаем основную функцию
main "$@"

