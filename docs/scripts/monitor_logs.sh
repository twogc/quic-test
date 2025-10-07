#!/bin/bash
# QUIC Server Log Monitoring Script
# Мониторинг логов QUIC сервера в реальном времени

echo "📝 QUIC Server Log Monitor"
echo "========================="
echo ""

# Проверка существования лог файла
if [ ! -f "server.log" ]; then
    echo "❌ Log file 'server.log' not found!"
    echo "   Make sure the server is running and logging to server.log"
    exit 1
fi

echo "✅ Monitoring server.log"
echo "   File size: $(du -h server.log | cut -f1)"
echo "   Last modified: $(stat -c %y server.log | cut -d'.' -f1)"
echo ""

# Функция для цветного вывода логов
colorize_logs() {
    while IFS= read -r line; do
        if echo "$line" | grep -q "ERROR"; then
            echo -e "\033[31m$line\033[0m"  # Red for ERROR
        elif echo "$line" | grep -q "WARN"; then
            echo -e "\033[33m$line\033[0m"  # Yellow for WARN
        elif echo "$line" | grep -q "INFO"; then
            echo -e "\033[32m$line\033[0m"  # Green for INFO
        elif echo "$line" | grep -q "DEBUG"; then
            echo -e "\033[36m$line\033[0m"  # Cyan for DEBUG
        else
            echo "$line"  # Default color
        fi
    done
}

# Функция для фильтрации логов
filter_logs() {
    case "$1" in
        "error")
            grep --color=always "ERROR"
            ;;
        "warn")
            grep --color=always "WARN\|ERROR"
            ;;
        "info")
            grep --color=always "INFO\|WARN\|ERROR"
            ;;
        "all")
            cat
            ;;
        *)
            echo "Usage: $0 [all|info|warn|error]"
            echo "  all  - Show all logs"
            echo "  info - Show INFO, WARN, ERROR logs"
            echo "  warn - Show WARN, ERROR logs"
            echo "  error - Show only ERROR logs"
            exit 1
            ;;
    esac
}

# Определяем уровень фильтрации
FILTER_LEVEL=${1:-"all"}

echo "🔍 Filter level: $FILTER_LEVEL"
echo "📊 Starting log monitoring..."
echo "   Press Ctrl+C to stop"
echo ""

# Запускаем мониторинг
tail -f server.log | filter_logs "$FILTER_LEVEL" | colorize_logs
