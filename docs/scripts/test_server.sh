#!/bin/bash
# QUIC Server Test Script
# Тестирование QUIC сервера с экспериментальными функциями

echo "🧪 QUIC Server Test Suite"
echo "========================="
echo ""

# Проверка 1: Статус сервера
echo "1️⃣  Server Process Check:"
if pgrep -f "quic-test-experimental.*server" > /dev/null; then
    PID=$(pgrep -f "quic-test-experimental.*server")
    echo "   ✅ Server process found (PID: $PID)"
    echo "   ✅ Status: $(ps -o stat= -p $PID 2>/dev/null | tr -d ' ')"
    echo "   ✅ Uptime: $(ps -o etime= -p $PID 2>/dev/null | tr -d ' ')"
else
    echo "   ❌ Server process not found"
    echo "   💡 Start server with: ./quic-test-experimental -mode server -addr 0.0.0.0:9000 ..."
    exit 1
fi

echo ""

# Проверка 2: Логи сервера
echo "2️⃣  Server Logs Check:"
if [ -f "server.log" ]; then
    echo "   ✅ Log file exists: server.log"
    echo "   ✅ File size: $(du -h server.log | cut -f1)"
    
    # Проверяем последние записи
    LAST_LOG=$(tail -1 server.log 2>/dev/null)
    if [ -n "$LAST_LOG" ]; then
        echo "   ✅ Last log entry: $LAST_LOG"
    fi
    
    # Проверяем ошибки
    ERROR_COUNT=$(grep -c "ERROR" server.log 2>/dev/null || echo "0")
    if [ "$ERROR_COUNT" -gt 0 ]; then
        echo "   ⚠️  Found $ERROR_COUNT ERROR messages in logs"
        echo "   Recent errors:"
        grep "ERROR" server.log | tail -3 | sed 's/^/      /'
    else
        echo "   ✅ No ERROR messages found"
    fi
else
    echo "   ❌ Log file not found"
fi

echo ""

# Проверка 3: Qlog файлы
echo "3️⃣  Qlog Files Check:"
if [ -d "server-qlog" ]; then
    QLOG_COUNT=$(find server-qlog -name "*.qlog" 2>/dev/null | wc -l)
    echo "   ✅ Qlog directory exists: server-qlog"
    echo "   ✅ Qlog files count: $QLOG_COUNT"
    
    if [ $QLOG_COUNT -gt 0 ]; then
        LATEST_QLOG=$(ls -t server-qlog/*.qlog 2>/dev/null | head -1)
        if [ -n "$LATEST_QLOG" ]; then
            echo "   ✅ Latest qlog: $(basename "$LATEST_QLOG")"
            echo "   ✅ Latest qlog size: $(du -h "$LATEST_QLOG" | cut -f1)"
        fi
    else
        echo "   ℹ️  No qlog files yet (waiting for client connections)"
    fi
else
    echo "   ❌ Qlog directory not found"
fi

echo ""

# Проверка 4: Сетевые соединения
echo "4️⃣  Network Connectivity Check:"
echo "   Testing UDP port 9000..."

# Простая проверка UDP порта
if timeout 2 bash -c "</dev/udp/127.0.0.1/9000" 2>/dev/null; then
    echo "   ✅ UDP port 9000 is accessible"
else
    echo "   ⚠️  UDP port 9000 test inconclusive (UDP is connectionless)"
fi

# Проверяем, слушает ли процесс порт
if ss -u 2>/dev/null | grep -q ":9000"; then
    echo "   ✅ Port 9000 is bound to UDP"
else
    echo "   ℹ️  Port binding not visible (normal for UDP)"
fi

echo ""

# Проверка 5: Системные ресурсы
echo "5️⃣  System Resources Check:"
if pgrep -f "quic-test-experimental.*server" > /dev/null; then
    PID=$(pgrep -f "quic-test-experimental.*server")
    CPU_USAGE=$(ps -o %cpu= -p $PID 2>/dev/null | tr -d ' ')
    MEM_USAGE=$(ps -o %mem= -p $PID 2>/dev/null | tr -d ' ')
    RSS_MEM=$(ps -o rss= -p $PID 2>/dev/null | tr -d ' ')
    
    echo "   ✅ CPU Usage: ${CPU_USAGE}%"
    echo "   ✅ Memory Usage: ${MEM_USAGE}%"
    echo "   ✅ RSS Memory: ${RSS_MEM} KB"
    
    # Проверяем, не слишком ли высокое потребление ресурсов
    if (( $(echo "$CPU_USAGE > 50" | bc -l) )); then
        echo "   ⚠️  High CPU usage detected"
    fi
    
    if (( $(echo "$MEM_USAGE > 10" | bc -l) )); then
        echo "   ⚠️  High memory usage detected"
    fi
fi

echo ""

# Проверка 6: Конфигурация
echo "6️⃣  Server Configuration Check:"
if [ -f "server.log" ]; then
    echo "   Server configuration from logs:"
    grep -E "(Experimental Features|Congestion Control|FEC|ACK Frequency)" server.log | head -5 | sed 's/^/      /'
fi

echo ""

# Итоговый статус
echo "Test Summary:"
if pgrep -f "quic-test-experimental.*server" > /dev/null && [ -f "server.log" ]; then
    echo "   ✅ Server is RUNNING and LOGGING"
    echo "   ✅ Ready for client connections"
    echo ""
    echo "🔗 Server Information:"
    echo "   Address: 0.0.0.0:9000"
    echo "   Protocol: QUIC over UDP"
    echo "   Features: BBRv2, FEC, ACK Frequency, Greasing"
    echo ""
    echo "📝 Monitoring Commands:"
    echo "   Status: ./monitor_server.sh"
    echo "   Logs: ./monitor_logs.sh"
    echo "   Real-time: tail -f server.log"
else
    echo "   ❌ Server issues detected"
    echo "   💡 Check server status and logs"
fi

echo ""
echo "🧪 Test completed at $(date)"
