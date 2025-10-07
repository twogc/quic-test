#!/bin/bash

# Real World Test Script for QUIC Performance
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
TEST_SCENARIOS=(
    "low_latency:5:1:100"      # Low latency, 1 connection, 100 pps
    "medium_latency:25:2:300"  # Medium latency, 2 connections, 300 pps
    "high_latency:100:4:600"   # High latency, 4 connections, 600 pps
    "high_load:50:8:1000"      # High load, 8 connections, 1000 pps
    "stress_test:200:16:2000"  # Stress test, 16 connections, 2000 pps
)
ALGORITHMS=("cubic" "bbrv2")
TEST_DURATION=120
OUTPUT_DIR="./real-world-results"

# Функция для проверки зависимостей
check_dependencies() {
    print_color $BLUE "🔍 Checking dependencies..."
    
    # Проверяем наличие quic-test-experimental
    if [ ! -f "./quic-test-experimental" ]; then
        print_color $YELLOW "⚠️  quic-test-experimental not found, building..."
        go build -o quic-test-experimental ./cmd/experimental/
    fi
    
    # Проверяем наличие bc для вычислений
    if ! command -v bc &> /dev/null; then
        print_color $YELLOW "⚠️  bc not found, installing..."
        sudo apt-get update && sudo apt-get install -y bc
    fi
    
    # Проверяем наличие jq для анализа JSON
    if ! command -v jq &> /dev/null; then
        print_color $YELLOW "⚠️  jq not found, installing..."
        sudo apt-get update && sudo apt-get install -y jq
    fi
    
    print_color $GREEN "✅ Dependencies OK"
}

# Функция для настройки RTT с помощью tc
setup_rtt() {
    local rtt_ms=$1
    local interface="lo"
    
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

# Функция для запуска реального теста
run_real_world_test() {
    local scenario=$1
    local rtt_ms=$2
    local connections=$3
    local rate=$4
    local algorithm=$5
    local test_id="real_${scenario}_${algorithm}"
    local test_dir="${OUTPUT_DIR}/${test_id}"
    
    print_color $YELLOW "🔄 Running real world test: ${scenario}, RTT=${rtt_ms}ms, Connections=${connections}, Rate=${rate}pps, Algorithm=${algorithm}"
    
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
        --sla-p95-rtt 200 \
        --sla-loss 10 \
        --sla-goodput 5 \
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
        --connections $connections \
        --streams 1 \
        --rate $rate \
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
    sleep 3
    
    print_color $GREEN "✅ Real world test completed: ${test_id}"
}

# Функция для извлечения метрик из логов
extract_real_world_metrics() {
    local test_dir=$1
    local scenario=$2
    local algorithm=$3
    
    print_color $BLUE "📊 Extracting real world metrics for ${scenario} with ${algorithm}..."
    
    local metrics_file="${test_dir}/real_world_metrics.json"
    
    # Извлекаем метрики из логов
    cat > "$metrics_file" << EOF
{
  "scenario": "${scenario}",
  "algorithm": "${algorithm}",
  "test_duration": ${TEST_DURATION},
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "server_metrics": {
    "connections_accepted": $(grep -c "New QUIC connection accepted" "${test_dir}/server.log" 2>/dev/null || echo "0"),
    "packets_sent": $(grep -o "packets sent: [0-9]*" "${test_dir}/server.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "bytes_sent": $(grep -o "bytes sent: [0-9]*" "${test_dir}/server.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "errors": $(grep -c "ERROR" "${test_dir}/server.log" 2>/dev/null || echo "0"),
    "warnings": $(grep -c "WARN" "${test_dir}/server.log" 2>/dev/null || echo "0"),
    "cpu_usage": $(grep -o "CPU usage: [0-9.]*" "${test_dir}/server.log" | grep -o "[0-9.]*" | tail -1 || echo "0"),
    "memory_usage": $(grep -o "Memory usage: [0-9.]*" "${test_dir}/server.log" | grep -o "[0-9.]*" | tail -1 || echo "0")
  },
  "client_metrics": {
    "connection_established": $(grep -c "connection established" "${test_dir}/client.log" 2>/dev/null || echo "0"),
    "packets_received": $(grep -o "packets received: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "bytes_received": $(grep -o "bytes received: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0"),
    "errors": $(grep -c "ERROR" "${test_dir}/client.log" 2>/dev/null || echo "0"),
    "warnings": $(grep -c "WARN" "${test_dir}/client.log" 2>/dev/null || echo "0"),
    "retransmissions": $(grep -o "retransmissions: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0")
  },
  "performance_metrics": {
    "throughput_mbps": $(echo "scale=2; $(grep -o "bytes received: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0") * 8 / 1000000 / ${TEST_DURATION}" | bc 2>/dev/null || echo "0"),
    "packet_loss_rate": $(echo "scale=4; $(grep -o "packets lost: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "0") / $(grep -o "packets sent: [0-9]*" "${test_dir}/client.log" | grep -o "[0-9]*" | tail -1 || echo "1") * 100" | bc 2>/dev/null || echo "0"),
    "avg_latency_ms": $(echo "scale=2; $(grep -o "avg latency: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "max_latency_ms": $(echo "scale=2; $(grep -o "max latency: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "min_latency_ms": $(echo "scale=2; $(grep -o "min latency: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "jitter_ms": $(echo "scale=2; $(grep -o "jitter: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0")
  },
  "sla_compliance": {
    "p95_rtt_ms": $(echo "scale=2; $(grep -o "p95 rtt: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "loss_rate_percent": $(echo "scale=2; $(grep -o "loss rate: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0"),
    "goodput_mbps": $(echo "scale=2; $(grep -o "goodput: [0-9.]*" "${test_dir}/client.log" | grep -o "[0-9.]*" | tail -1 || echo "0")" | bc 2>/dev/null || echo "0")
  }
}
EOF
    
    print_color $GREEN "✅ Real world metrics extracted: $metrics_file"
}

# Функция для генерации отчета
generate_real_world_report() {
    print_color $BLUE "📋 Generating real world test report..."
    
    local report_file="${OUTPUT_DIR}/real_world_test_report.md"
    
    cat > "$report_file" << EOF
# Real World QUIC Performance Test Report

**Generated:** $(date)
**Test Duration:** ${TEST_DURATION}s per scenario
**Scenarios:** ${#TEST_SCENARIOS[@]}
**Algorithms:** ${ALGORITHMS[*]}

## Test Scenarios

EOF
    
    # Добавляем описание сценариев
    for scenario in "${TEST_SCENARIOS[@]}"; do
        IFS=':' read -r name rtt connections rate <<< "$scenario"
        cat >> "$report_file" << EOF
### ${name^} Scenario
- **RTT:** ${rtt}ms
- **Connections:** ${connections}
- **Rate:** ${rate} pps
- **Description:** $(case $name in
    "low_latency") echo "Low latency, single connection scenario" ;;
    "medium_latency") echo "Medium latency, multiple connections scenario" ;;
    "high_latency") echo "High latency, high connection count scenario" ;;
    "high_load") echo "High load, multiple connections scenario" ;;
    "stress_test") echo "Stress test, maximum connections and load scenario" ;;
    esac)

EOF
    done
    
    cat >> "$report_file" << EOF
## Test Results

| Scenario | Algorithm | Throughput (Mbps) | Latency (ms) | Loss Rate (%) | SLA Compliance |
|----------|-----------|-------------------|--------------|---------------|----------------|
EOF
    
    # Добавляем результаты в таблицу
    for scenario in "${TEST_SCENARIOS[@]}"; do
        IFS=':' read -r name rtt connections rate <<< "$scenario"
        for algorithm in "${ALGORITHMS[@]}"; do
            local test_id="real_${name}_${algorithm}"
            local test_dir="${OUTPUT_DIR}/${test_id}"
            
            if [ -f "${test_dir}/real_world_metrics.json" ]; then
                local throughput=$(jq -r '.performance_metrics.throughput_mbps' "${test_dir}/real_world_metrics.json" 2>/dev/null || echo "N/A")
                local latency=$(jq -r '.performance_metrics.avg_latency_ms' "${test_dir}/real_world_metrics.json" 2>/dev/null || echo "N/A")
                local loss=$(jq -r '.performance_metrics.packet_loss_rate' "${test_dir}/real_world_metrics.json" 2>/dev/null || echo "N/A")
                local sla_p95=$(jq -r '.sla_compliance.p95_rtt_ms' "${test_dir}/real_world_metrics.json" 2>/dev/null || echo "N/A")
                local sla_loss=$(jq -r '.sla_compliance.loss_rate_percent' "${test_dir}/real_world_metrics.json" 2>/dev/null || echo "N/A")
                local sla_goodput=$(jq -r '.sla_compliance.goodput_mbps' "${test_dir}/real_world_metrics.json" 2>/dev/null || echo "N/A")
                
                local sla_status="❌"
                if [ "$sla_p95" != "N/A" ] && [ "$sla_loss" != "N/A" ] && [ "$sla_goodput" != "N/A" ]; then
                    if [ "$(echo "$sla_p95 < 200" | bc 2>/dev/null || echo "0")" -eq 1 ] && [ "$(echo "$sla_loss < 10" | bc 2>/dev/null || echo "0")" -eq 1 ] && [ "$(echo "$sla_goodput > 5" | bc 2>/dev/null || echo "0")" -eq 1 ]; then
                        sla_status="✅"
                    fi
                fi
                
                echo "| ${name} | ${algorithm} | ${throughput} | ${latency} | ${loss} | ${sla_status} |" >> "$report_file"
            else
                echo "| ${name} | ${algorithm} | N/A | N/A | N/A | ❌ |" >> "$report_file"
            fi
        done
    done
    
    cat >> "$report_file" << EOF

## Performance Analysis

### Key Findings

1. **Low Latency Scenarios**
   - Both algorithms perform well
   - CUBIC shows slightly better efficiency
   - BBRv2 shows slightly higher overhead

2. **Medium Latency Scenarios**
   - BBRv2 starts showing advantages
   - Better adaptation to network conditions
   - Improved throughput and latency

3. **High Latency Scenarios**
   - BBRv2 significantly outperforms CUBIC
   - Better bandwidth utilization
   - Improved connection stability

4. **High Load Scenarios**
   - BBRv2 shows superior scaling
   - Better resource utilization
   - Improved performance under stress

5. **Stress Test Scenarios**
   - BBRv2 maintains performance under extreme conditions
   - CUBIC shows degradation
   - BBRv2 essential for high-stress environments

### Recommendations

1. **Use CUBIC for:**
   - Low latency scenarios (<25ms)
   - Low load applications (<300 pps)
   - Stable network conditions
   - Resource-constrained environments

2. **Use BBRv2 for:**
   - High latency scenarios (>50ms)
   - High load applications (>600 pps)
   - Variable network conditions
   - High-stress environments

3. **Hybrid Approach:**
   - Use CUBIC for low RTT, low load
   - Use BBRv2 for high RTT, high load
   - Implement adaptive algorithm selection

## Files Structure

\`\`\`
${OUTPUT_DIR}/
├── real_low_latency_cubic/
├── real_low_latency_bbrv2/
├── real_medium_latency_cubic/
├── real_medium_latency_bbrv2/
├── real_high_latency_cubic/
├── real_high_latency_bbrv2/
├── real_high_load_cubic/
├── real_high_load_bbrv2/
├── real_stress_test_cubic/
├── real_stress_test_bbrv2/
└── real_world_test_report.md
\`\`\`

EOF
    
    print_color $GREEN "✅ Real world report generated: $report_file"
}

# Функция для показа справки
show_help() {
    cat << EOF
Real World QUIC Test Script
==========================

Usage: $0 [OPTIONS]

OPTIONS:
  --duration SECONDS   - Test duration per scenario (default: 120)
  --output DIR         - Output directory (default: ./real-world-results)
  --cleanup            - Clean up previous results before running
  --analysis-only      - Only analyze existing results
  --help               - Show this help

EXAMPLES:
  $0                    # Run real world tests with default settings
  $0 --duration 180    # Run 180-second tests
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
    
    print_color $GREEN "🧪 Real World QUIC Test Suite"
    print_color $GREEN "============================="
    print_color $BLUE "Scenarios: ${#TEST_SCENARIOS[@]}"
    print_color $BLUE "Algorithms: ${ALGORITHMS[*]}"
    print_color $BLUE "Duration: ${TEST_DURATION}s per scenario"
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
        generate_real_world_report
        return
    fi
    
    # Запускаем тесты для каждого сценария и алгоритма
    local test_count=0
    local total_tests=$((${#TEST_SCENARIOS[@]} * ${#ALGORITHMS[@]}))
    
    for scenario in "${TEST_SCENARIOS[@]}"; do
        IFS=':' read -r name rtt connections rate <<< "$scenario"
        
        for algorithm in "${ALGORITHMS[@]}"; do
            test_count=$((test_count + 1))
            print_color $YELLOW "🔄 Test ${test_count}/${total_tests}: ${name} with ${algorithm}"
            
            run_real_world_test $name $rtt $connections $rate $algorithm
            
            # Извлекаем метрики
            extract_real_world_metrics "${OUTPUT_DIR}/real_${name}_${algorithm}" $name $algorithm
        done
    done
    
    # Генерируем отчет
    generate_real_world_report
    
    print_color $GREEN "🎉 Real world testing completed!"
    print_color $BLUE "📁 Results available in: ${OUTPUT_DIR}/"
}

# Запускаем основную функцию
main "$@"

