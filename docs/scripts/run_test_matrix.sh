#!/bin/bash

# QUIC Test Matrix Runner Script
# ===============================

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
    print_color $BLUE "🧹 Cleaning up previous results..."
    
    # Убиваем запущенные процессы
    pkill -f quic-test-experimental 2>/dev/null || true
    
    # Очищаем временные файлы
    rm -rf ./test-results/* 2>/dev/null || true
    rm -rf ./server-qlog 2>/dev/null || true
    rm -rf ./client-qlog 2>/dev/null || true
    
    print_color $GREEN "✅ Cleanup completed"
}

# Функция для сборки тестового раннера
build_test_runner() {
    print_color $BLUE "🔨 Building test matrix runner..."
    
    go build -o test-matrix-runner ./cmd/test-matrix/
    
    if [ $? -eq 0 ]; then
        print_color $GREEN "✅ Test matrix runner built successfully"
    else
        print_color $RED "❌ Failed to build test matrix runner"
        exit 1
    fi
}

# Функция для запуска легких тестов
run_light_tests() {
    print_color $BLUE "🚀 Running LIGHT test matrix..."
    
    ./test-matrix-runner \
        --profile light \
        --output ./test-results/light \
        --verbose
    
    print_color $GREEN "✅ Light tests completed"
}

# Функция для запуска обычных тестов
run_normal_tests() {
    print_color $BLUE "🚀 Running NORMAL test matrix..."
    
    ./test-matrix-runner \
        --profile normal \
        --output ./test-results/normal \
        --verbose
    
    print_color $GREEN "✅ Normal tests completed"
}

# Функция для запуска тяжелых тестов
run_heavy_tests() {
    print_color $BLUE "🚀 Running HEAVY test matrix..."
    
    ./test-matrix-runner \
        --profile heavy \
        --output ./test-results/heavy \
        --verbose
    
    print_color $GREEN "✅ Heavy tests completed"
}

# Функция для генерации отчета
generate_report() {
    print_color $BLUE "📊 Generating test report..."
    
    # Создаем директорию для отчета
    mkdir -p ./test-results/reports
    
    # Генерируем простой отчет
    cat > ./test-results/reports/summary.txt << EOF
QUIC Test Matrix Results
========================

Generated: $(date)
Profile: $1
Output Directory: ./test-results/$1

Test Results:
- Check individual scenario directories for detailed results
- Server logs: ./test-results/$1/*/server-qlog/
- Client logs: ./test-results/$1/*/client-qlog/

Next Steps:
1. Analyze qlog files with qvis
2. Check SLA violations
3. Compare performance metrics
4. Generate performance graphs

EOF
    
    print_color $GREEN "✅ Report generated: ./test-results/reports/summary.txt"
}

# Функция для показа справки
show_help() {
    cat << EOF
QUIC Test Matrix Runner
======================

Usage: $0 [PROFILE] [OPTIONS]

PROFILES:
  light    - Quick tests (2x2x2x2 = 16 scenarios)
  normal   - Standard tests (4x4x4x3 = 192 scenarios)  
  heavy    - Comprehensive tests (5x5x7x5 = 875 scenarios)

OPTIONS:
  --cleanup    - Clean up previous results before running
  --build      - Build test runner before running
  --report     - Generate report after tests
  --help       - Show this help

EXAMPLES:
  $0 light                    # Run light tests
  $0 normal --cleanup         # Clean and run normal tests
  $0 heavy --build --report  # Build, run heavy tests, generate report

EOF
}

# Основная логика
main() {
    local profile="normal"
    local cleanup_flag=false
    local build_flag=false
    local report_flag=false
    
    # Парсим аргументы
    while [[ $# -gt 0 ]]; do
        case $1 in
            light|normal|heavy)
                profile=$1
                shift
                ;;
            --cleanup)
                cleanup_flag=true
                shift
                ;;
            --build)
                build_flag=true
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
    
    print_color $GREEN "🧪 QUIC Test Matrix Runner"
    print_color $GREEN "========================="
    print_color $BLUE "Profile: $profile"
    
    # Проверяем зависимости
    check_dependencies
    
    # Очистка при необходимости
    if [ "$cleanup_flag" = true ]; then
        cleanup
    fi
    
    # Сборка при необходимости
    if [ "$build_flag" = true ]; then
        build_test_runner
    fi
    
    # Запускаем тесты в зависимости от профиля
    case $profile in
        light)
            run_light_tests
            ;;
        normal)
            run_normal_tests
            ;;
        heavy)
            run_heavy_tests
            ;;
    esac
    
    # Генерируем отчет при необходимости
    if [ "$report_flag" = true ]; then
        generate_report $profile
    fi
    
    print_color $GREEN "🎉 Test matrix execution completed!"
    print_color $BLUE "📁 Results available in: ./test-results/$profile/"
}

# Запускаем основную функцию
main "$@"

