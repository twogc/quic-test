# CloudBridge Relay Integration

## 🎯 Интеграция с CloudBridge Relay

QUIC Testing Tool теперь настроен для работы с реальным CloudBridge Relay, развернутым в Minikube.

## 🔌 Используемые порты

### ✅ Безопасные порты (не 80/443)
- **Dashboard**: 9990 (QUIC Testing Tool)
- **MASQUE**: 32095 (CloudBridge Relay)
- **STUN**: 32302 (CloudBridge Relay)
- **QUIC**: 32094 (CloudBridge Relay)
- **WireGuard**: 31820 (CloudBridge Relay)

## 🚀 Конфигурация

### MASQUE Testing
- **Сервер**: 192.168.58.2:32095
- **Протокол**: HTTPS/TLS
- **Функции**: CONNECT-UDP, CONNECT-IP, HTTP Datagrams

### ICE/STUN/TURN Testing
- **STUN**: 192.168.58.2:32302
- **Резервные**: stun.l.google.com:19302
- **Функции**: NAT traversal, candidate gathering

## 📊 Dashboard

### Доступные вкладки
1. **Protocol Analysis** - анализ QUIC протокола
2. **Network Simulation** - симуляция сетевых условий
3. **Deep Analysis** - глубокий анализ протокола
4. **Protocol Comparison** - сравнение протоколов
5. **MASQUE Testing** - тестирование MASQUE (RFC 9298, RFC 9484)
6. **ICE/STUN/TURN** - тестирование NAT traversal

### API Endpoints
- `/api/metrics` - динамические метрики
- `/api/config` - конфигурация сервера
- `/api/masque/start` - запуск MASQUE тестирования
- `/api/ice/start` - запуск ICE тестирования
- `/api/history` - история метрик

## 🧪 Тестирование

### Запуск тестов
```bash
# Dashboard
./build/quck-test dashboard

# MASQUE тестирование
./build/quck-test masque

# ICE тестирование
./build/quck-test ice

# Расширенное тестирование
./build/quck-test enhanced
```

### Веб-интерфейс
- **URL**: http://localhost:9990
- **MASQUE**: Настроен на 192.168.58.2:32095
- **ICE**: Настроен на 192.168.58.2:32302

## 🔍 Мониторинг

### CloudBridge Relay
- **Статус**: 3/3 подов работают
- **Namespace**: cloudbridge
- **Метрики**: Prometheus на порту 30651

### QUIC Testing Tool
- **Статус**: Dashboard на порту 9990
- **Метрики**: Динамические обновления
- **Логи**: Реальное время

## 📈 Результаты

### Динамические метрики
- **Latency**: 10-50ms (в зависимости от режима)
- **Throughput**: 10-300 Mbps
- **Connections**: 0-10 активных
- **Packet Loss**: 0-3%

### Режимы тестирования
- **Неактивное**: Низкие значения
- **MASQUE**: Средние значения (15-40ms latency)
- **ICE**: Высокие значения (30-70ms latency)
- **Активное**: Максимальные значения

## 🎉 Готово к использованию!

Все компоненты настроены и готовы к тестированию:
- ✅ CloudBridge Relay работает в Minikube
- ✅ QUIC Testing Tool настроен на правильные порты
- ✅ Dashboard показывает динамические метрики
- ✅ MASQUE и ICE тестирование интегрированы
- ✅ Безопасные порты (не 80/443)

Откройте http://localhost:9990 и начинайте тестирование! 🚀

