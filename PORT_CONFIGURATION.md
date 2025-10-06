# QUIC Testing Tool - Конфигурация портов

## 🎯 Обновленная конфигурация портов

Все порты настроены для избежания конфликтов с системными портами 80 и 443.

## 🔌 Используемые порты

### QUIC Testing Tool
| Компонент | Порт | Назначение | Статус |
|-----------|------|------------|--------|
| **Dashboard** | 9990 | Веб-интерфейс | ✅ Активен |
| **QUIC Server** | 9000 | QUIC сервер | ✅ Настроен |
| **QUIC Client** | 9000 | QUIC клиент | ✅ Настроен |

### CloudBridge Relay (Minikube)
| Компонент | Порт | Назначение | Статус |
|-----------|------|------------|--------|
| **MASQUE** | 32095 | MASQUE Proxy | ✅ Доступен |
| **STUN** | 32302 | STUN Server | ✅ Доступен |
| **QUIC** | 32094 | QUIC Transport | ✅ Доступен |
| **WireGuard** | 31820 | VPN туннелирование | ✅ Доступен |

## 📊 Конфигурация в Dashboard

### Server Control
- **Address**: `:9000` (QUIC сервер)
- **Certificate**: `server.crt`
- **Private Key**: `server.key`
- **Prometheus**: Включен

### Client Control
- **Server Address**: `localhost:9000` (QUIC клиент)
- **Connections**: 1
- **Streams**: 1
- **Packet Size**: 1200
- **Rate**: 100 packets/sec
- **Pattern**: Random

### MASQUE Testing
- **MASQUE Server**: `192.168.58.2:32095` (CloudBridge Relay)
- **Target Hosts**: `8.8.8.8:53,1.1.1.1:53`

### ICE/STUN/TURN Testing
- **STUN Servers**: `192.168.58.2:32302,stun.l.google.com:19302`
- **TURN Servers**: (опционально)

## 🚀 Запуск компонентов

### Dashboard
```bash
./build/quck-test dashboard
# Доступен на http://localhost:9990
```

### QUIC Server
```bash
./build/quck-test server
# Запускается на порту 9000
```

### QUIC Client
```bash
./build/quck-test client
# Подключается к localhost:9000
```

### MASQUE Testing
```bash
./build/quck-test masque
# Тестирует CloudBridge Relay MASQUE
```

### ICE Testing
```bash
./build/quck-test ice
# Тестирует CloudBridge Relay STUN
```

## 🔍 Проверка портов

### Локальные порты
```bash
# Dashboard
curl http://localhost:9990/api/config

# QUIC Server (после запуска)
curl https://localhost:9000/health
```

### CloudBridge Relay
```bash
# MASQUE
curl -k https://192.168.58.2:32095/health

# STUN
nc -u 192.168.58.2 32302
```

## 📈 Мониторинг

### Метрики
- **Dashboard**: http://localhost:9990
- **Prometheus**: http://localhost:9990/api/metrics
- **CloudBridge**: http://192.168.58.2:32091

### Логи
```bash
# QUIC Testing Tool
./build/quck-test dashboard

# CloudBridge Relay
kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay
```

## ✅ Безопасность портов

### Избегаем системные порты
- ❌ **80** - HTTP (системный)
- ❌ **443** - HTTPS (системный)
- ✅ **9000** - QUIC Testing Tool
- ✅ **9990** - Dashboard
- ✅ **32095** - CloudBridge MASQUE
- ✅ **32302** - CloudBridge STUN

### Доступные порты
- **9000-9999**: QUIC Testing Tool
- **30000-32767**: CloudBridge Relay (NodePort)
- **10000-19999**: Резервные порты

## 🎉 Готово к использованию!

Все порты настроены безопасно:
- ✅ Нет конфликтов с системными портами
- ✅ QUIC Testing Tool на порту 9000
- ✅ Dashboard на порту 9990
- ✅ CloudBridge Relay интеграция
- ✅ MASQUE и ICE тестирование

Откройте http://localhost:9990 и начинайте тестирование! 🚀

