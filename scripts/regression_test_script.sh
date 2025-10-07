#!/bin/bash

# Regression Test Script for QUIC Performance Comparison
# ====================================================

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
ALGORITHMS=("cubic" "bbrv2")            # Алгоритмы для сравнения
TEST_DURATION=60                         # Длительность теста в секундах
OUTPUT_DIR="./regression-results"        # Директория для результатов
METRICS_INTERVAL=1                       # Интервал сбора метрик в секундах

# Функция для проверки зависимостей
check_dependencies() {
    print_color $BLUE "🔍 Checking dependencies..."
    
    # Проверяем наличие quic-test-experimental
    if [ ! -f "./quic-test-experimental" ]; then
        print_color $YELLOW "⚠️  quic-test-experimental not found, building..."
        go build -o quic-test-experimental ./cmd/experimental/
    fi
    
    # Проверяем наличие jq для анализа JSON
    if ! command -v jq &> /dev/null; then
        print_color $YELLOW "⚠️  jq not found, installing..."
        sudo apt-get update && sudo apt-get install -y jq
    fi
    
    print_color $GREEN "✅ Dependencies OK"
}

# Функция для запуска теста с заданным алгоритмом
run_regression_test() {
    local algorithm=$1
    local test_id="regression_${algorithm}"
    local test_dir="${OUTPUT_DIR}/${test_id}"
    
    print_color $YELLOW "🔄 Running regression test: Algorithm=${algorithm}"
    
    # Создаем директорию для теста
    mkdir -p "$test_dir"
    
    # Запускаем сервер
    print_color $BLUE "🚀 Starting server with ${algorithm}..."
    nohup ./quic-test-experimental \
        --mode server \
        --cc $algorithm \
        --qlog "${test_dir}/server-qlog" \
        --verbose \
        --metrics-interval ${METRICS_INTERVAL}s \
        --sla-p95-rtt 100 \
        --sla-loss 5 \
        --sla-goodput 10 \
        > "${test_dir}/server.log" 2>&1 &
    
    local server_pid=$!
    print_color $GREEN "✅ Server started (PID: $server_pid)"
    
    # Ждем запуска сервера
    sleep 5
    
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
    
    # Пауза между тестами
    sleep 3
    
    print_color $GREEN "✅ Regression test completed: ${test_id}"
}

# Функция для извлечения метрик из логов
extract_metrics() {
    local test_dir=$1
    local algorithm=$2
    
    print_color $BLUE "📊 Extracting metrics for ${algorithm}..."
    
    local metrics_file="${test_dir}/metrics.json"
    
    # Извлекаем метрики из логов
    cat > "$metrics_file" << EOF
{
  "algorithm": "${algorithm}",
  "test_duration": ${TEST_DURATION},
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "server_metrics": {
    "connections_accepted": $(grep -c "New QUIC connection accepted" "${test_dir}/server.log" 2>/dev/null || echo "0"),
    "packets_sent": $(grep -o "packets sent: [0-9]*" "${test_dir}/server.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "bytes_sent": $(grep -o "bytes sent: [0-9]*" "${test_dir}/server.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "errors": $(grep -c "ERROR" "${test_dir}/server.log" 2>/dev/null || echo "0"),
    "warnings": $(grep -c "WARN" "${test_dir}/server.log" 2>/dev/null || echo "0")
  },
  "client_metrics": {
    "connection_established": $(grep -c "connection established" "${test_dir}/client.log" 2>/dev/null || echo "0"),
    "packets_received": $(grep -o "packets received: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "bytes_received": $(grep -o "bytes received: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "errors": $(grep -c "ERROR" "${test_dir}/client.log" 2>/dev/null || echo "0"),
    "warnings": $(grep -c "WARN" "${test_dir}/client.log" 2>/dev/null || echo "0")
  },
  "performance_metrics": {
    "throughput_mbps": $(echo "scale=2; $(grep -o "bytes received: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0") * 8 / 1000000 / ${TEST_DURATION}" | bc 2>/dev/null || echo "0"),
    "packet_loss_rate": $(echo "scale=4; $(grep -o "packets lost: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0") / $(grep -o "packets sent: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "1") * 100" | bc 2>/dev/null || echo "0"),
    "avg_latency_ms": $(echo "scale=2; $(grep -o "avg latency: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "max_latency_ms": $(echo "scale=2; $(grep -o "max latency: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "min_latency_ms": $(echo "scale=2; $(grep -o "min latency: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0")
  },
  "sla_compliance": {
    "p95_rtt_ms": $(echo "scale=2; $(grep -o "p95 rtt: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "loss_rate_percent": $(echo "scale=2; $(grep -o "loss rate: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "goodput_mbps": $(echo "scale=2; $(grep -o "goodput: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0")
  }
}
EOF
    
    print_color $GREEN "✅ Metrics extracted: $metrics_file"
}

# Функция для сравнения результатов
compare_results() {
    print_color $BLUE "📊 Comparing regression test results..."
    
    local comparison_file="${OUTPUT_DIR}/regression_comparison.json"
    
    # Создаем файл сравнения
    cat > "$comparison_file" << EOF
{
  "comparison_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "test_duration": ${TEST_DURATION},
  "algorithms_compared": ["cubic", "bbrv2"],
  "comparison_results": {
    "cubic": $(cat "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "{}"),
    "bbrv2": $(cat "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "{}")
  }
}
EOF
    
    # Вычисляем улучшения
    local cubic_throughput=$(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0")
    local bbrv2_throughput=$(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")
    
    local throughput_improvement=$(echo "scale=2; ($bbrv2_throughput - $cubic_throughput) / $cubic_throughput * 100" | bc 2>/dev/null || echo "0")
    
    # Добавляем анализ улучшений
    cat >> "$comparison_file" << EOF
,
  "performance_improvements": {
    "throughput_improvement_percent": ${throughput_improvement},
    "latency_improvement_percent": $(echo "scale=2; ($(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0") - $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")) / $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "1") * 100" | bc 2>/dev/null || echo "0"),
    "loss_rate_improvement_percent": $(echo "scale=2; ($(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0") - $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")) / $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "1") * 100" | bc 2>/dev/null || echo "0")
  }
}
EOF
    
    print_color $GREEN "✅ Comparison completed: $comparison_file"
}

# Функция для генерации отчета
generate_report() {
    print_color $BLUE "📋 Generating regression test report..."
    
    local report_file="${OUTPUT_DIR}/regression_test_report.md"
    
    cat > "$report_file" << EOF
# QUIC Regression Test Report

**Generated:** $(date)
**Test Duration:** ${TEST_DURATION}s per algorithm
**Algorithms Compared:** CUBIC vs BBRv2

## Test Configuration

- **Test Duration:** ${TEST_DURATION} seconds per algorithm
- **Connections:** 1
- **Streams:** 1
- **Packet Rate:** 100 pps
- **Packet Size:** 1200 bytes
- **SLA Gates:** P95 RTT < 100ms, Loss < 5%, Goodput > 10 Mbps

## Test Results

### CUBIC Algorithm
EOF
    
    # Добавляем результаты CUBIC
    if [ -f "${OUTPUT_DIR}/regression_cubic/metrics.json" ]; then
        cat >> "$report_file" << EOF
- **Throughput:** $(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json") Mbps
- **Average Latency:** $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json") ms
- **Packet Loss Rate:** $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_cubic/metrics.json")%
- **P95 RTT:** $(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json") ms
- **Goodput:** $(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json") Mbps
- **Errors:** $(jq -r '.server_metrics.errors + .client_metrics.errors' "${OUTPUT_DIR}/regression_cubic/metrics.json")
EOF
    else
        echo "- **Status:** Test failed or not completed" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

### BBRv2 Algorithm
EOF
    
    # Добавляем результаты BBRv2
    if [ -f "${OUTPUT_DIR}/regression_bbrv2/metrics.json" ]; then
        cat >> "$report_file" << EOF
- **Throughput:** $(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json") Mbps
- **Average Latency:** $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json") ms
- **Packet Loss Rate:** $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_bbrv2/metrics.json")%
- **P95 RTT:** $(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json") ms
- **Goodput:** $(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json") Mbps
- **Errors:** $(jq -r '.server_metrics.errors + .client_metrics.errors' "${OUTPUT_DIR}/regression_bbrv2/metrics.json")
EOF
    else
        echo "- **Status:** Test failed or not completed" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Performance Comparison

| Metric | CUBIC | BBRv2 | Improvement |
|--------|-------|-------|-------------|
| Throughput (Mbps) | $(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "N/A") | $(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "N/A") | $(echo "scale=1; ($(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0") - $(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0")) / $(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "1") * 100" | bc 2>/dev/null || echo "N/A")% |
| Avg Latency (ms) | $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "N/A") | $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "N/A") | $(echo "scale=1; ($(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0") - $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")) / $(jq -r '.performance_metrics.avg_latency_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "1") * 100" | bc 2>/dev/null || echo "N/A")% |
| Loss Rate (%) | $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "N/A") | $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "N/A") | $(echo "scale=1; ($(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0") - $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")) / $(jq -r '.performance_metrics.packet_loss_rate' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "1") * 100" | bc 2>/dev/null || echo "N/A")% |
| P95 RTT (ms) | $(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "N/A") | $(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "N/A") | $(echo "scale=1; ($(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0") - $(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")) / $(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "1") * 100" | bc 2>/dev/null || echo "N/A")% |
| Goodput (Mbps) | $(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "N/A") | $(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "N/A") | $(echo "scale=1; ($(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0") - $(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0")) / $(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "1") * 100" | bc 2>/dev/null || echo "N/A")% |

## SLA Compliance

### CUBIC SLA Compliance
- **P95 RTT < 100ms:** $(if [ "$(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0")" -lt 100 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)
- **Loss Rate < 5%:** $(if [ "$(jq -r '.sla_compliance.loss_rate_percent' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0")" -lt 5 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)
- **Goodput > 10 Mbps:** $(if [ "$(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0")" -gt 10 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)

### BBRv2 SLA Compliance
- **P95 RTT < 100ms:** $(if [ "$(jq -r '.sla_compliance.p95_rtt_ms' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")" -lt 100 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)
- **Loss Rate < 5%:** $(if [ "$(jq -r '.sla_compliance.loss_rate_percent' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")" -lt 5 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)
- **Goodput > 10 Mbps:** $(if [ "$(jq -r '.sla_compliance.goodput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")" -gt 10 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)

## Conclusion

$(if [ "$(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_bbrv2/metrics.json" 2>/dev/null || echo "0")" -gt "$(jq -r '.performance_metrics.throughput_mbps' "${OUTPUT_DIR}/regression_cubic/metrics.json" 2>/dev/null || echo "0")" ]; then echo "BBRv2 shows improved performance over CUBIC in this regression test."; else echo "CUBIC shows better performance than BBRv2 in this regression test."; fi)

## Files Structure

\`\`\`
${OUTPUT_DIR}/
├── regression_cubic/
│   ├── server.log
│   ├── client.log
│   ├── server-qlog/
│   ├── client-qlog/
│   └── metrics.json
├── regression_bbrv2/
│   ├── server.log
│   ├── client.log
│   ├── server-qlog/
│   ├── client-qlog/
│   └── metrics.json
├── regression_comparison.json
└── regression_test_report.md
\`\`\`

EOF
    
    print_color $GREEN "✅ Report generated: $report_file"
}

# Функция для показа справки
show_help() {
    cat << EOF
QUIC Regression Test Script
==========================

Usage: $0 [OPTIONS]

OPTIONS:
  --duration SECONDS   - Test duration per algorithm (default: 60)
  --output DIR         - Output directory (default: ./regression-results)
  --cleanup            - Clean up previous results before running
  --analysis-only      - Only analyze existing results
  --help               - Show this help

EXAMPLES:
  $0                    # Run regression tests with default settings
  $0 --duration 120     # Run 120-second tests
  $0 --cleanup          # Clean and run tests
  $0 --analysis-only    # Analyze existing results

EOF
}

# Основная логика
main() {
    local cleanup_flag=false
    local analysis_only=false
    
    # Парсим аргументы
    while [[ $# -gt 0 ]]; do
        case $1 in
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
    
    print_color $GREEN "🧪 QUIC Regression Test Suite"
    print_color $GREEN "============================="
    print_color $BLUE "Algorithms: ${ALGORITHMS[*]}"
    print_color $BLUE "Duration: ${TEST_DURATION}s per algorithm"
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
        compare_results
        generate_report
        return
    fi
    
    # Запускаем тесты для каждого алгоритма
    for algorithm in "${ALGORITHMS[@]}"; do
        print_color $YELLOW "🔄 Testing algorithm: ${algorithm}"
        
        run_regression_test $algorithm
        
        # Извлекаем метрики
        extract_metrics "${OUTPUT_DIR}/regression_${algorithm}" $algorithm
    done
    
    # Сравниваем результаты
    compare_results
    
    # Генерируем отчет
    generate_report
    
    print_color $GREEN "🎉 Regression testing completed!"
    print_color $BLUE "📁 Results available in: ${OUTPUT_DIR}/"
}

# Запускаем основную функцию
main "$@"

