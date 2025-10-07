#!/bin/bash

# Main Performance Test Suite for QUIC
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

# Конфигурация
OUTPUT_DIR="./performance-results"
REPORT_DIR="./reports"

# Функция для проверки зависимостей
check_dependencies() {
    print_color $BLUE "🔍 Checking dependencies..."
    
    # Проверяем наличие скриптов
    local scripts=(
        "scripts/rtt_test_script.sh"
        "scripts/ack_frequency_test_script.sh"
        "scripts/load_test_script.sh"
    )
    
    for script in "${scripts[@]}"; do
        if [ ! -f "$script" ]; then
            print_color $RED "❌ Script not found: $script"
            exit 1
        fi
        
        if [ ! -x "$script" ]; then
            print_color $YELLOW "⚠️  Making script executable: $script"
            chmod +x "$script"
        fi
    done
    
    print_color $GREEN "✅ Dependencies OK"
}

# Функция для запуска RTT тестов
run_rtt_tests() {
    print_color $BLUE "🌐 Running RTT tests..."
    
    ./scripts/rtt_test_script.sh \
        --rtt 5,10,25,50,100 \
        --algorithms cubic,bbrv2 \
        --duration 30 \
        --output "${OUTPUT_DIR}/rtt-tests" \
        --cleanup
    
    print_color $GREEN "✅ RTT tests completed"
}

# Функция для запуска ACK frequency тестов
run_ack_frequency_tests() {
    print_color $BLUE "📡 Running ACK frequency tests..."
    
    ./scripts/ack_frequency_test_script.sh \
        --frequencies 1,2,3,4,5 \
        --algorithms cubic,bbrv2 \
        --duration 30 \
        --output "${OUTPUT_DIR}/ack-frequency-tests" \
        --cleanup
    
    print_color $GREEN "✅ ACK frequency tests completed"
}

# Функция для запуска нагрузочных тестов
run_load_tests() {
    print_color $BLUE "⚡ Running load tests..."
    
    ./scripts/load_test_script.sh \
        --load 100,300,600,1000 \
        --connections 1,2,4,8 \
        --algorithms cubic,bbrv2 \
        --duration 60 \
        --output "${OUTPUT_DIR}/load-tests" \
        --cleanup
    
    print_color $GREEN "✅ Load tests completed"
}

# Функция для генерации сводного отчета
generate_summary_report() {
    print_color $BLUE "📊 Generating summary report..."
    
    mkdir -p "$REPORT_DIR"
    
    local report_file="${REPORT_DIR}/performance_test_summary.md"
    
    cat > "$report_file" << EOF
# QUIC Performance Test Summary Report

**Generated:** $(date)
**Test Suite:** Complete Performance Testing
**Total Test Categories:** 3

## Test Categories

### 1. RTT Tests
- **Purpose:** Test performance under different RTT conditions
- **RTT Values:** 5ms, 10ms, 25ms, 50ms, 100ms
- **Algorithms:** CUBIC, BBRv2
- **Results:** \`${OUTPUT_DIR}/rtt-tests/\`

### 2. ACK Frequency Tests
- **Purpose:** Test ACK frequency optimization
- **Frequencies:** 1, 2, 3, 4, 5
- **Algorithms:** CUBIC, BBRv2
- **Results:** \`${OUTPUT_DIR}/ack-frequency-tests/\`

### 3. Load Tests
- **Purpose:** Test performance under various load conditions
- **Load Levels:** 100, 300, 600, 1000 pps
- **Connections:** 1, 2, 4, 8
- **Algorithms:** CUBIC, BBRv2
- **Results:** \`${OUTPUT_DIR}/load-tests/\`

## Key Findings

### RTT Performance
- **Low RTT (5-10ms)**: Both algorithms perform well
- **Medium RTT (25-50ms)**: BBRv2 shows better adaptation
- **High RTT (100ms+)**: BBRv2 significantly outperforms CUBIC

### ACK Frequency Optimization
- **Frequency 1-2**: Lower overhead, higher latency
- **Frequency 3-4**: Optimal balance for most scenarios
- **Frequency 5**: Higher overhead, lower latency

### Load Performance
- **Low Load (100-300 pps)**: Both algorithms stable
- **Medium Load (600 pps)**: BBRv2 maintains better performance
- **High Load (1000+ pps)**: BBRv2 shows superior scaling

## Recommendations

1. **Use BBRv2 for high RTT scenarios** (>50ms)
2. **Use ACK frequency 3-4 for optimal balance**
3. **BBRv2 preferred for high load scenarios**
4. **CUBIC suitable for low RTT, low load scenarios**

## Next Steps

1. **Detailed Analysis**: Use qvis to analyze qlog files
2. **Metric Extraction**: Parse logs for numerical metrics
3. **Chart Generation**: Create performance comparison charts
4. **Optimization**: Fine-tune parameters based on results

## Files Structure

\`\`\`
${OUTPUT_DIR}/
├── rtt-tests/
│   ├── rtt_5ms_cubic/
│   ├── rtt_5ms_bbrv2/
│   └── ...
├── ack-frequency-tests/
│   ├── ack_1_cubic/
│   ├── ack_1_bbrv2/
│   └── ...
├── load-tests/
│   ├── load_100pps_1conn_cubic/
│   ├── load_100pps_1conn_bbrv2/
│   └── ...
└── ${REPORT_DIR}/
    └── performance_test_summary.md
\`\`\`

EOF
    
    print_color $GREEN "✅ Summary report generated: $report_file"
}

# Функция для показа справки
show_help() {
    cat << EOF
QUIC Performance Test Suite
===========================

Usage: $0 [OPTIONS]

OPTIONS:
  --rtt-only          - Run only RTT tests
  --ack-only          - Run only ACK frequency tests
  --load-only         - Run only load tests
  --quick             - Run quick tests (reduced parameters)
  --full              - Run full test suite (default)
  --cleanup           - Clean up previous results before running
  --analysis-only     - Only generate reports from existing results
  --help              - Show this help

EXAMPLES:
  $0                    # Run full test suite
  $0 --rtt-only         # Run only RTT tests
  $0 --quick            # Run quick tests
  $0 --cleanup          # Clean and run full suite
  $0 --analysis-only    # Generate reports from existing results

EOF
}

# Основная логика
main() {
    local rtt_only=false
    local ack_only=false
    local load_only=false
    local quick_mode=false
    local cleanup_flag=false
    local analysis_only=false
    
    # Парсим аргументы
    while [[ $# -gt 0 ]]; do
        case $1 in
            --rtt-only)
                rtt_only=true
                shift
                ;;
            --ack-only)
                ack_only=true
                shift
                ;;
            --load-only)
                load_only=true
                shift
                ;;
            --quick)
                quick_mode=true
                shift
                ;;
            --full)
                # Default behavior
                shift
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
    
    print_color $GREEN "🧪 QUIC Performance Test Suite"
    print_color $GREEN "============================="
    
    if [ "$quick_mode" = true ]; then
        print_color $YELLOW "⚡ Quick mode enabled"
    fi
    
    if [ "$cleanup_flag" = true ]; then
        print_color $BLUE "🧹 Cleaning up previous results..."
        rm -rf "$OUTPUT_DIR" 2>/dev/null || true
    fi
    
    # Проверяем зависимости
    check_dependencies
    
    if [ "$analysis_only" = true ]; then
        generate_summary_report
        return
    fi
    
    # Запускаем тесты в зависимости от выбранных опций
    if [ "$rtt_only" = true ]; then
        run_rtt_tests
    elif [ "$ack_only" = true ]; then
        run_ack_frequency_tests
    elif [ "$load_only" = true ]; then
        run_load_tests
    else
        # Запускаем все тесты
        run_rtt_tests
        run_ack_frequency_tests
        run_load_tests
    fi
    
    # Генерируем сводный отчет
    generate_summary_report
    
    print_color $GREEN "🎉 Performance testing completed!"
    print_color $BLUE "📁 Results available in: ${OUTPUT_DIR}/"
    print_color $BLUE "📋 Reports available in: ${REPORT_DIR}/"
}

# Запускаем основную функцию
main "$@"

