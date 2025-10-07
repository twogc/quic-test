#!/bin/bash

echo "🧪 Тестирование экспериментальных функций QUIC"
echo "=============================================="

# Остановим все предыдущие процессы
pkill -f "experimental" 2>/dev/null || true
sleep 2

echo "1. Тестирование BBRv2 Congestion Control..."
go run cmd/experimental/main.go -mode client -addr localhost:9000 -cc bbrv2 -duration 5s -verbose &
sleep 6

echo "2. Тестирование ACK Frequency..."
go run cmd/experimental/main.go -mode client -addr localhost:9000 -ack-freq 2 -duration 5s -verbose &
sleep 6

echo "3. Тестирование QUIC Datagrams..."
go run cmd/experimental/main.go -mode client -addr localhost:9000 -enable-datagrams -duration 5s -verbose &
sleep 6

echo "4. Тестирование с qlog трассировкой..."
go run cmd/experimental/main.go -mode client -addr localhost:9000 -qlog /tmp/experimental.qlog -duration 5s -verbose &
sleep 6

echo "5. Проверка qlog файла..."
if [ -f "/tmp/experimental.qlog" ]; then
    echo "✅ qlog файл создан:"
    ls -la /tmp/experimental.qlog
    echo "Содержимое qlog:"
    head -20 /tmp/experimental.qlog
else
    echo "❌ qlog файл не найден"
fi

echo "6. Тестирование завершено!"
echo "Проверим статус сервера:"
ps aux | grep quic-server | grep -v grep
