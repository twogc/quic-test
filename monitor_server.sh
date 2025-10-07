#!/bin/bash
# QUIC Server Monitoring Script
# Мониторинг QUIC сервера с экспериментальными функциями

echo "🔍 QUIC Server Monitoring Dashboard"
echo "=================================="
echo ""

# Проверка статуса сервера
echo "📊 Server Status:"
if pgrep -f "quic-test-experimental.*server" > /dev/null; then
    echo "✅ Server is RUNNING"
    PID=$(pgrep -f "quic-test-experimental.*server")
    echo "   PID: $PID"
    echo "   Uptime: $(ps -o etime= -p $PID 2>/dev/null | tr -d ' ')"
else
    echo "❌ Server is NOT RUNNING"
fi

echo ""

# Проверка порта
echo "🌐 Network Status:"
if netstat -tuln 2>/dev/null | grep -q ":9000 "; then
    echo "✅ Port 9000 is LISTENING"
else
    echo "⚠️  Port 9000 not detected (UDP ports may not show in netstat)"
fi

echo ""

# Проверка логов
echo "📝 Log Status:"
if [ -f "server.log" ]; then
    echo "✅ Log file exists: server.log"
    echo "   Size: $(du -h server.log | cut -f1)"
    echo "   Last modified: $(stat -c %y server.log 2>/dev/null | cut -d'.' -f1)"
    echo "   Last 3 lines:"
    tail -3 server.log | sed 's/^/     /'
else
    echo "❌ Log file not found"
fi

echo ""

# Проверка qlog
echo "📊 Qlog Status:"
if [ -d "server-qlog" ]; then
    QLOG_COUNT=$(find server-qlog -name "*.qlog" 2>/dev/null | wc -l)
    echo "✅ Qlog directory exists: server-qlog"
    echo "   Qlog files: $QLOG_COUNT"
    if [ $QLOG_COUNT -gt 0 ]; then
        echo "   Latest qlog: $(ls -t server-qlog/*.qlog 2>/dev/null | head -1 | xargs basename 2>/dev/null)"
    fi
else
    echo "❌ Qlog directory not found"
fi

echo ""

# Системные ресурсы
echo "💻 System Resources:"
if pgrep -f "quic-test-experimental.*server" > /dev/null; then
    PID=$(pgrep -f "quic-test-experimental.*server")
    echo "   CPU Usage: $(ps -o %cpu= -p $PID 2>/dev/null | tr -d ' ')%"
    echo "   Memory Usage: $(ps -o %mem= -p $PID 2>/dev/null | tr -d ' ')%"
    echo "   RSS Memory: $(ps -o rss= -p $PID 2>/dev/null | tr -d ' ') KB"
fi

echo ""

# Сетевые соединения
echo "🔗 Network Connections:"
CONNECTIONS=$(ss -u 2>/dev/null | grep -c ":9000" || echo "0")
echo "   UDP connections on port 9000: $CONNECTIONS"

echo ""

# Статистика логов
echo "📈 Log Statistics:"
if [ -f "server.log" ]; then
    echo "   Total lines: $(wc -l < server.log)"
    echo "   INFO messages: $(grep -c "INFO" server.log 2>/dev/null || echo "0")"
    echo "   ERROR messages: $(grep -c "ERROR" server.log 2>/dev/null || echo "0")"
    echo "   WARN messages: $(grep -c "WARN" server.log 2>/dev/null || echo "0")"
fi

echo ""
echo "🔄 Run this script again to refresh status"
echo "📊 For real-time monitoring: tail -f server.log"
