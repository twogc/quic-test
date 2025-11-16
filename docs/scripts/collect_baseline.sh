#!/bin/bash

# QUIC Baseline Data Collection Script
# ====================================

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

# Функция для проверки зависимостей
check_dependencies() {
    print_color $BLUE "🔍 Checking dependencies..."
    
    # Проверяем Go
    if ! command -v go &> /dev/null; then
        print_color $RED "❌ Go is not installed"
        exit 1
    fi
    
    # Проверяем наличие quic-test-experimental
    if [ ! -f "./quic-test-experimental" ]; then
        print_color $YELLOW "⚠️  quic-test-experimental not found, building..."
        go build -o quic-test-experimental ./cmd/experimental/
    fi
    
    print_color $GREEN "✅ Dependencies OK"
}

# Функция для очистки предыдущих результатов
cleanup() {
    print_color $BLUE "🧹 Cleaning up previous baseline data..."
    
    # Убиваем запущенные процессы
    pkill -f quic-test-experimental 2>/dev/null || true
    
    # Очищаем старые данные
    rm -rf ./baseline-data/* 2>/dev/null || true
    rm -rf ./server-qlog 2>/dev/null || true
    rm -rf ./client-qlog 2>/dev/null || true
    
    print_color $GREEN "✅ Cleanup completed"
}

# Функция для создания структуры директорий
setup_directories() {
    print_color $BLUE "Setting up directory structure..."
    
    mkdir -p ./baseline-data/{cubic,bbrv2}/{server,client}/{qlog,metrics,logs}
    mkdir -p ./baseline-data/reports
    mkdir -p ./baseline-data/analysis
    
    print_color $GREEN "✅ Directory structure created"
}

# Функция для сбора данных с CUBIC
collect_cubic_baseline() {
    print_color $BLUE "Collecting CUBIC baseline data..."
    
    local test_duration=${1:-30}
    local output_dir="./baseline-data/cubic"
    
    print_color $YELLOW "🔄 Running CUBIC server test (${test_duration}s)..."
    
    # Запускаем сервер с CUBIC
    nohup ./quic-test-experimental \
        --mode server \
        --cc cubic \
        --qlog "${output_dir}/server/qlog" \
        --verbose \
        --metrics-interval 1s \
        > "${output_dir}/server/logs/server.log" 2>&1 &
    
    local server_pid=$!
    print_color $GREEN "✅ CUBIC server started (PID: $server_pid)"
    
    # Ждем запуска сервера
    sleep 3
    
    # Запускаем клиент в фоне с таймаутом
    print_color $YELLOW "🔄 Running CUBIC client test..."
    timeout ${test_duration}s ./quic-test-experimental \
        --mode client \
        --addr 127.0.0.1:9000 \
        --cc cubic \
        --qlog "${output_dir}/client/qlog" \
        --duration ${test_duration}s \
        --connections 1 \
        --streams 1 \
        --rate 100 \
        --packet-size 1200 \
        --verbose \
        > "${output_dir}/client/logs/client.log" 2>&1 &
    
    local client_pid=$!
    
    # Ждем завершения клиента
    wait $client_pid 2>/dev/null || true
    
    # Останавливаем сервер
    kill $server_pid 2>/dev/null || true
    wait $server_pid 2>/dev/null || true
    
    # Дополнительная пауза для завершения
    sleep 2
    
    print_color $GREEN "✅ CUBIC baseline data collected"
}

# Функция для сбора данных с BBRv2
collect_bbrv2_baseline() {
    print_color $BLUE "Collecting BBRv2 baseline data..."
    
    local test_duration=${1:-30}
    local output_dir="./baseline-data/bbrv2"
    
    print_color $YELLOW "🔄 Running BBRv2 server test (${test_duration}s)..."
    
    # Запускаем сервер с BBRv2
    nohup ./quic-test-experimental \
        --mode server \
        --cc bbrv2 \
        --ack-freq 2 \
        --fec \
        --fec-redundancy 0.1 \
        --greasing \
        --qlog "${output_dir}/server/qlog" \
        --verbose \
        --metrics-interval 1s \
        > "${output_dir}/server/logs/server.log" 2>&1 &
    
    local server_pid=$!
    print_color $GREEN "✅ BBRv2 server started (PID: $server_pid)"
    
    # Ждем запуска сервера
    sleep 3
    
    # Запускаем клиент в фоне с таймаутом
    print_color $YELLOW "🔄 Running BBRv2 client test..."
    timeout ${test_duration}s ./quic-test-experimental \
        --mode client \
        --addr 127.0.0.1:9000 \
        --cc bbrv2 \
        --qlog "${output_dir}/client/qlog" \
        --duration ${test_duration}s \
        --connections 1 \
        --streams 1 \
        --rate 100 \
        --packet-size 1200 \
        --verbose \
        > "${output_dir}/client/logs/client.log" 2>&1 &
    
    local client_pid=$!
    
    # Ждем завершения клиента
    wait $client_pid 2>/dev/null || true
    
    # Останавливаем сервер
    kill $server_pid 2>/dev/null || true
    wait $server_pid 2>/dev/null || true
    
    # Дополнительная пауза для завершения
    sleep 2
    
    print_color $GREEN "✅ BBRv2 baseline data collected"
}

# Функция для анализа собранных данных
analyze_baseline_data() {
    print_color $BLUE "📈 Analyzing baseline data..."
    
    local analysis_dir="./baseline-data/analysis"
    
    # Создаем простой анализ
    cat > "${analysis_dir}/baseline_analysis.txt" << EOF
QUIC Baseline Data Analysis
===========================

Generated: $(date)

Data Collection Summary:
- CUBIC baseline: ./baseline-data/cubic/
- BBRv2 baseline: ./baseline-data/bbrv2/

Files Collected:
- Server qlog: server/qlog/
- Client qlog: client/qlog/
- Server logs: server/logs/server.log
- Client logs: client/logs/client.log

Next Steps:
1. Analyze qlog files with qvis
2. Compare CUBIC vs BBRv2 performance
3. Extract key metrics (RTT, throughput, loss)
4. Generate performance comparison charts
5. Create regression test baselines

Analysis Commands:
- qvis: Open qlog files in browser
- Compare metrics: Check logs for key performance indicators
- Generate charts: Use collected data for visualization

EOF
    
    print_color $GREEN "✅ Baseline analysis created"
}

# Функция для генерации отчета
generate_report() {
    print_color $BLUE "Generating baseline report..."
    
    local report_dir="./baseline-data/reports"
    
    # Создаем отчет
    cat > "${report_dir}/baseline_report.md" << EOF
# QUIC Baseline Data Collection Report

**Generated:** $(date)
**Test Duration:** ${1:-30} seconds

## Overview

This report summarizes the baseline data collection for QUIC performance testing.

## Data Collection

### CUBIC Baseline
- **Algorithm:** CUBIC congestion control
- **Features:** Standard QUIC implementation
- **Data Location:** \`./baseline-data/cubic/\`

### BBRv2 Baseline  
- **Algorithm:** BBRv2 congestion control
- **Features:** BBRv2 + ACK-Frequency + FEC + Greasing
- **Data Location:** \`./baseline-data/bbrv2/\`

## Collected Data

### Server Data
- **qlog files:** Detailed QUIC protocol events
- **Server logs:** Application-level metrics
- **Metrics:** Prometheus-style metrics (if enabled)

### Client Data
- **qlog files:** Client-side QUIC events
- **Client logs:** Performance measurements
- **Connection data:** RTT, throughput, loss rates

## Analysis

### Key Metrics to Compare
1. **RTT (Round Trip Time)**
   - Min, Max, Mean, P95, P99
2. **Throughput**
   - Goodput (application data)
   - Total throughput (including overhead)
3. **Loss Rate**
   - Packet loss percentage
   - Retransmission rate
4. **Congestion Control**
   - CWND evolution
   - Bandwidth utilization
   - State transitions

### Expected Differences
- **CUBIC:** Conservative, loss-based
- **BBRv2:** Aggressive, delay-based
- **Performance:** BBRv2 should show better performance in high-RTT scenarios

## Next Steps

1. **qvis Analysis:** Open qlog files in qvis for detailed protocol analysis
2. **Metric Extraction:** Parse logs to extract numerical metrics
3. **Comparison Charts:** Create side-by-side performance comparisons
4. **Regression Tests:** Use this data as baseline for future tests

## Files Structure

\`\`\`
baseline-data/
├── cubic/
│   ├── server/
│   │   ├── qlog/          # Server qlog files
│   │   └── logs/          # Server application logs
│   └── client/
│       ├── qlog/          # Client qlog files
│       └── logs/          # Client application logs
├── bbrv2/
│   ├── server/
│   │   ├── qlog/          # Server qlog files
│   │   └── logs/          # Server application logs
│   └── client/
│       ├── qlog/          # Client qlog files
│       └── logs/          # Client application logs
├── analysis/              # Analysis results
└── reports/               # Generated reports
\`\`\`

EOF
    
    print_color $GREEN "✅ Baseline report generated: ${report_dir}/baseline_report.md"
}

# Функция для показа справки
show_help() {
    cat << EOF
QUIC Baseline Data Collection
=============================

Usage: $0 [OPTIONS]

OPTIONS:
  --duration SECONDS    - Test duration (default: 30)
  --cubic-only         - Collect only CUBIC baseline
  --bbrv2-only         - Collect only BBRv2 baseline
  --cleanup            - Clean up previous data before collection
  --analysis           - Run analysis after collection
  --report             - Generate report after collection
  --help               - Show this help

EXAMPLES:
  $0                           # Collect both CUBIC and BBRv2 baselines
  $0 --duration 60             # Collect with 60-second tests
  $0 --cubic-only --cleanup    # Collect only CUBIC, clean first
  $0 --analysis --report       # Collect data, analyze, and generate report

EOF
}

# Основная логика
main() {
    local duration=30
    local cubic_only=false
    local bbrv2_only=false
    local cleanup_flag=false
    local analysis_flag=false
    local report_flag=false
    
    # Парсим аргументы
    while [[ $# -gt 0 ]]; do
        case $1 in
            --duration)
                duration=$2
                shift 2
                ;;
            --cubic-only)
                cubic_only=true
                shift
                ;;
            --bbrv2-only)
                bbrv2_only=true
                shift
                ;;
            --cleanup)
                cleanup_flag=true
                shift
                ;;
            --analysis)
                analysis_flag=true
                shift
                ;;
            --report)
                report_flag=true
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
    
    print_color $GREEN "QUIC Baseline Data Collection"
    print_color $GREEN "================================="
    print_color $BLUE "Duration: ${duration}s"
    
    # Проверяем зависимости
    check_dependencies
    
    # Очистка при необходимости
    if [ "$cleanup_flag" = true ]; then
        cleanup
    fi
    
    # Создаем структуру директорий
    setup_directories
    
    # Собираем данные
    if [ "$cubic_only" = true ]; then
        collect_cubic_baseline $duration
    elif [ "$bbrv2_only" = true ]; then
        collect_bbrv2_baseline $duration
    else
        collect_cubic_baseline $duration
        sleep 2  # Пауза между тестами
        collect_bbrv2_baseline $duration
    fi
    
    # Анализ при необходимости
    if [ "$analysis_flag" = true ]; then
        analyze_baseline_data
    fi
    
    # Генерируем отчет при необходимости
    if [ "$report_flag" = true ]; then
        generate_report $duration
    fi
    
    print_color $GREEN "🎉 Baseline data collection completed!"
    print_color $BLUE "Data available in: ./baseline-data/"
}

# Запускаем основную функцию
main "$@"
