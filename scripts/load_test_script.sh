#!/bin/bash

# Load Test Script for QUIC Performance Testing
# =============================================

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
LOAD_LEVELS=(100 300 600 1000 2000)    # Packets per second
CONNECTION_COUNTS=(1 2 4 8 16)          # Number of connections
ALGORITHMS=("cubic" "bbrv2")            # Алгоритмы congestion control
TEST_DURATION=60                         # Длительность теста в секундах
OUTPUT_DIR="./load-test-results"         # Директория для результатов

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

# Функция для запуска нагрузочного теста
run_load_test() {
    local load_pps=$1
    local connections=$2
    local algorithm=$3
    local test_id="load_${load_pps}pps_${connections}conn_${algorithm}"
    local test_dir="${OUTPUT_DIR}/${test_id}"
    
    print_color $YELLOW "🔄 Running load test: ${load_pps}pps, ${connections}conn, ${algorithm}"
    
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
    
    # Запускаем клиент с нагрузкой
    print_color $BLUE "🔗 Starting client with load..."
    timeout ${TEST_DURATION}s go run main.go \
        --mode client \
        --addr 127.0.0.1:9000 \
        --cc $algorithm \
        --duration ${TEST_DURATION}s \
        --connections $connections \
        --streams 1 \
        --rate $load_pps \
        --packet-size 1200 \
        > "${test_dir}/client.log" 2>&1 &
    
    local client_pid=$!
    
    # Ждем завершения клиента
    wait $client_pid 2>/dev/null || true
    
    # Останавливаем сервер
    kill $server_pid 2>/dev/null || true
    wait $server_pid 2>/dev/null || true
    
    # Пауза между тестами
    sleep 3
    
    print_color $GREEN "✅ Load test completed: ${test_id}"
}

# Функция для анализа результатов
analyze_results() {
    print_color $BLUE "Analyzing load test results..."
    
    local analysis_file="${OUTPUT_DIR}/load_test_analysis.txt"
    
    cat > "$analysis_file" << EOF
Load Test Performance Analysis
=============================

Generated: $(date)
Test Duration: ${TEST_DURATION}s per test
Total Tests: $((${#LOAD_LEVELS[@]} * ${#CONNECTION_COUNTS[@]} * ${#ALGORITHMS[@]}))

Test Results:
EOF
    
    # Анализируем каждый тест
    for load_pps in "${LOAD_LEVELS[@]}"; do
        for connections in "${CONNECTION_COUNTS[@]}"; do
            for algorithm in "${ALGORITHMS[@]}"; do
                local test_id="load_${load_pps}pps_${connections}conn_${algorithm}"
                local test_dir="${OUTPUT_DIR}/${test_id}"
                
                if [ -d "$test_dir" ]; then
                    echo "" >> "$analysis_file"
                    echo "Test: ${test_id}" >> "$analysis_file"
                    echo "Load: ${load_pps} pps" >> "$analysis_file"
                    echo "Connections: ${connections}" >> "$analysis_file"
                    echo "Algorithm: ${algorithm}" >> "$analysis_file"
                    echo "Files:" >> "$analysis_file"
                    echo "  - Server log: ${test_dir}/server.log" >> "$analysis_file"
                    echo "  - Client log: ${test_dir}/client.log" >> "$analysis_file"
                    echo "  - Server qlog: ${test_dir}/server-qlog/" >> "$analysis_file"
                    echo "  - Client qlog: ${test_dir}/client-qlog/" >> "$analysis_file"
                fi
            done
        done
    done
    
    print_color $GREEN "✅ Analysis saved to: $analysis_file"
}

# Функция для генерации отчета
generate_report() {
    print_color $BLUE "Generating load test performance report..."
    
    local report_file="${OUTPUT_DIR}/load_test_performance_report.md"
    
    cat > "$report_file" << EOF
# Load Test Performance Report

**Generated:** $(date)
**Test Duration:** ${TEST_DURATION}s per test
**Total Tests:** $((${#LOAD_LEVELS[@]} * ${#CONNECTION_COUNTS[@]} * ${#ALGORITHMS[@]}))

## Test Configuration

### Load Levels (Packets Per Second)
$(printf '%s\n' "${LOAD_LEVELS[@]}" | sed 's/^/- /')

### Connection Counts
$(printf '%s\n' "${CONNECTION_COUNTS[@]}" | sed 's/^/- /')

### Algorithms
$(printf '%s\n' "${ALGORITHMS[@]}" | sed 's/^/- /')

## Test Results

| Load (pps) | Connections | Algorithm | Status | Files |
|------------|-------------|-----------|--------|-------|
EOF
    
    # Добавляем результаты в таблицу
    for load_pps in "${LOAD_LEVELS[@]}"; do
        for connections in "${CONNECTION_COUNTS[@]}"; do
            for algorithm in "${ALGORITHMS[@]}"; do
                local test_id="load_${load_pps}pps_${connections}conn_${algorithm}"
                local test_dir="${OUTPUT_DIR}/${test_id}"
                
                if [ -d "$test_dir" ]; then
                    echo "| ${load_pps} | ${connections} | ${algorithm} | ✅ Completed | \`${test_dir}/\` |" >> "$report_file"
                else
                    echo "| ${load_pps} | ${connections} | ${algorithm} | ❌ Failed | - |" >> "$report_file"
                fi
            done
        done
    done
    
    cat >> "$report_file" << EOF

## Analysis

### Key Metrics to Compare
1. **Throughput Scaling**
   - How does throughput scale with load and connections?
   - Which algorithm handles high load better?

2. **Latency Under Load**
   - How does latency change with increasing load?
   - Which algorithm maintains low latency under load?

3. **Connection Stability**
   - How many connections can each algorithm handle?
   - Which algorithm is more stable under high load?

4. **Resource Utilization**
   - How does CPU and memory usage scale with load?
   - Which algorithm is more efficient?

### Expected Results
- **CUBIC**: Conservative, stable under moderate load
- **BBRv2**: Aggressive, better under high load
- **Scaling**: Performance should degrade gracefully

### Performance Thresholds
- **Good Performance**: < 5% packet loss, < 100ms latency
- **Acceptable Performance**: < 10% packet loss, < 200ms latency
- **Poor Performance**: > 10% packet loss, > 200ms latency

### Next Steps
1. Analyze qlog files with qvis
2. Extract performance metrics from logs
3. Create load vs performance charts
4. Identify optimal configurations for different scenarios

## Files Structure

\`\`\`
${OUTPUT_DIR}/
├── load_100pps_1conn_cubic/
├── load_100pps_1conn_bbrv2/
├── load_300pps_2conn_cubic/
├── load_300pps_2conn_bbrv2/
├── ...
├── load_test_analysis.txt
└── load_test_performance_report.md
\`\`\`

EOF
    
    print_color $GREEN "✅ Report generated: $report_file"
}

# Функция для показа справки
show_help() {
    cat << EOF
Load Test Script for QUIC Performance Testing
=============================================

Usage: $0 [OPTIONS]

OPTIONS:
  --profile PROFILE    - Test profile to run:
                         * standard (default): 100-2000 pps with 1-16 connections
                         * real-world: Multi-scenario realistic load testing
  --load VALUES        - Load levels to test (comma-separated, default: 100,300,600,1000,2000)
  --connections VALUES - Connection counts to test (comma-separated, default: 1,2,4,8,16)
  --algorithms VALUES  - Algorithms to test (comma-separated, default: cubic,bbrv2)
  --duration SECONDS   - Test duration per test (default: 60)
  --output DIR         - Output directory (default: ./load-test-results or ./real-world-results)
  --cleanup            - Clean up previous results before running
  --analysis-only      - Only analyze existing results
  --help               - Show this help

PROFILES:
  standard             - Standard load testing: 100/300/600/1000/2000 pps, 1-16 connections
  real-world           - Real-world scenarios:
                         * low_latency: 5ms RTT, 1 conn, 100 pps
                         * medium: 25ms RTT, 2 conn, 300 pps
                         * high_latency: 100ms RTT, 4 conn, 600 pps
                         * high_load: 50ms RTT, 8 conn, 1000 pps
                         * stress: 200ms RTT, 16 conn, 2000 pps

EXAMPLES:
  $0                                    # Run standard tests with default settings
  $0 --profile real-world               # Run real-world scenario tests
  $0 --load 300,600 --connections 2,4  # Test specific load and connection combinations
  $0 --duration 120 --cleanup           # Run 120-second tests, clean first
  $0 --analysis-only                    # Analyze existing results

EOF
}

# Функция для применения профилей
apply_profile() {
    local profile=$1

    case $profile in
        standard)
            # Default settings already set
            OUTPUT_DIR="./load-test-results"
            print_color $BLUE "Profile: Standard load testing"
            ;;
        real-world)
            # Real-world scenario settings
            LOAD_LEVELS=(100 300 600 1000 2000)
            CONNECTION_COUNTS=(1 2 4 8 16)
            ALGORITHMS=("cubic" "bbrv2")
            TEST_DURATION=120  # Longer for realistic scenarios
            OUTPUT_DIR="./real-world-results"
            print_color $BLUE "Profile: Real-world scenario testing"
            ;;
        *)
            print_color $RED "❌ Unknown profile: $profile"
            exit 1
            ;;
    esac
}

# Основная логика
main() {
    local cleanup_flag=false
    local analysis_only=false
    local profile="standard"

    # Парсим аргументы
    while [[ $# -gt 0 ]]; do
        case $1 in
            --profile)
                profile=$2
                shift 2
                ;;
            --load)
                IFS=',' read -ra LOAD_LEVELS <<< "$2"
                shift 2
                ;;
            --connections)
                IFS=',' read -ra CONNECTION_COUNTS <<< "$2"
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

    # Применяем профиль
    apply_profile "$profile"
    
    print_color $GREEN "🧪 Load Test Performance Suite"
    print_color $GREEN "=============================="
    print_color $BLUE "Load Levels: ${LOAD_LEVELS[*]} pps"
    print_color $BLUE "Connections: ${CONNECTION_COUNTS[*]}"
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
    local total_tests=$((${#LOAD_LEVELS[@]} * ${#CONNECTION_COUNTS[@]} * ${#ALGORITHMS[@]}))
    
    for load_pps in "${LOAD_LEVELS[@]}"; do
        for connections in "${CONNECTION_COUNTS[@]}"; do
            for algorithm in "${ALGORITHMS[@]}"; do
                test_count=$((test_count + 1))
                print_color $YELLOW "🔄 Test ${test_count}/${total_tests}: Load=${load_pps}pps, Connections=${connections}, Algorithm=${algorithm}"
                
                run_load_test $load_pps $connections $algorithm
            done
        done
    done
    
    # Анализируем результаты
    analyze_results
    generate_report
    
    print_color $GREEN "🎉 Load test performance testing completed!"
    print_color $BLUE "Results available in: ${OUTPUT_DIR}/"
}

# Запускаем основную функцию
main "$@"

