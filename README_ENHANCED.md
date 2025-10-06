# 🚀 **CloudBridge QUIC Testing Suite - Enhanced Edition**

## **📋 Обзор**

Расширенная версия тестирования QUIC протокола для CloudBridge 2GC с поддержкой **MASQUE** (RFC 9298, RFC 9484) и **ICE/STUN/TURN** тестирования.

### **🔥 Новые возможности:**

- **MASQUE Protocol Testing** - CONNECT-UDP, CONNECT-IP, HTTP Datagrams
- **ICE/STUN/TURN Testing** - NAT traversal, candidate gathering, connectivity
- **Enhanced Integration** - Комплексное тестирование всех протоколов
- **Go 1.25.1** - Поддержка последней версии Go
- **Real-time Metrics** - Детальная аналитика производительности

## **🛠️ Установка и настройка**

### **Требования:**
- Go 1.25.1+
- Linux/macOS/Windows
- Доступ к интернету для STUN/TURN серверов

### **Установка зависимостей:**
```bash
go mod tidy
```

### **Сборка:**
```bash
go build -o quck-test
```

## **🚀 Использование**

### **1. Базовое QUIC тестирование**
```bash
# Сервер
./quck-test -mode=server -addr=:9000

# Клиент
./quck-test -mode=client -addr=localhost:9000 -connections=5 -streams=10
```

### **2. MASQUE тестирование**
```bash
# Тестирование CONNECT-UDP
./quck-test -mode=masque \
  -masque-server=localhost:8443 \
  -masque-targets="8.8.8.8:53,1.1.1.1:53,cloudflare.com:443"

# Тестирование с кастомными параметрами
./quck-test -mode=masque \
  -masque-server=masque.example.com:8443 \
  -masque-targets="target1.example.com:80,target2.example.com:443"
```

### **3. ICE/STUN/TURN тестирование**
```bash
# Базовое STUN тестирование
./quck-test -mode=ice \
  -ice-stun="stun.l.google.com:19302,stun1.l.google.com:19302"

# Полное ICE тестирование с TURN
./quck-test -mode=ice \
  -ice-stun="stun.l.google.com:19302" \
  -ice-turn="turn.example.com:3478" \
  -ice-turn-user="username" \
  -ice-turn-pass="password"
```

### **4. Расширенное тестирование (MASQUE + ICE + QUIC)**
```bash
# Комплексное тестирование всех протоколов
./quck-test -mode=enhanced \
  -masque-server=localhost:8443 \
  -masque-targets="8.8.8.8:53,1.1.1.1:53" \
  -ice-stun="stun.l.google.com:19302" \
  -ice-turn="turn.example.com:3478" \
  -ice-turn-user="username" \
  -ice-turn-pass="password"
```

### **5. Веб-интерфейс**
```bash
# Запуск dashboard
./quck-test -mode=dashboard
```

## **📊 Режимы тестирования**

### **MASQUE Testing (`-mode=masque`)**

Тестирует MASQUE протокол (RFC 9298, RFC 9484):

- **CONNECT-UDP** - UDP туннелирование через HTTP/3
- **CONNECT-IP** - IP туннелирование с Capsule поддержкой
- **HTTP Datagrams** - Передача данных через QUIC datagrams
- **Capsule Fallback** - Fallback механизм при недоступности datagrams

**Параметры:**
- `-masque-server` - MASQUE сервер (по умолчанию: localhost:8443)
- `-masque-targets` - Целевые хосты для CONNECT-UDP (через запятую)

### **ICE Testing (`-mode=ice`)**

Тестирует ICE/STUN/TURN функциональность:

- **STUN Testing** - Проверка NAT discovery
- **TURN Testing** - Тестирование relay серверов
- **Candidate Gathering** - Сбор ICE кандидатов
- **NAT Traversal** - Тестирование различных типов NAT

**Параметры:**
- `-ice-stun` - STUN серверы (через запятую)
- `-ice-turn` - TURN серверы (через запятую)
- `-ice-turn-user` - TURN username
- `-ice-turn-pass` - TURN password

### **Enhanced Testing (`-mode=enhanced`)**

Комплексное тестирование всех компонентов:

- **MASQUE + ICE + QUIC** - Полное тестирование стека
- **Performance Analysis** - Анализ производительности
- **Connectivity Testing** - Тестирование соединений
- **Protocol Comparison** - Сравнение протоколов

## **📈 Метрики и аналитика**

### **MASQUE Metrics:**
- `connect_udp_successes` - Успешные CONNECT-UDP соединения
- `connect_ip_successes` - Успешные CONNECT-IP соединения
- `datagram_loss_rate` - Потеря datagrams (%)
- `throughput_mbps` - Пропускная способность (MB/s)
- `average_latency` - Средняя задержка

### **ICE Metrics:**
- `stun_requests/responses` - STUN запросы/ответы
- `turn_allocations` - TURN аллокации
- `candidates_gathered` - Собранные кандидаты
- `connections_successful` - Успешные соединения

### **Enhanced Metrics:**
- `total_tests` - Общее количество тестов
- `successful_tests` - Успешные тесты
- `success_rate` - Процент успеха (%)
- `test_duration` - Длительность тестирования

## **🔧 Конфигурация**

### **MASQUE Configuration:**
```go
type MASQUEConfig struct {
    ServerURL      string        `json:"server_url"`
    UDPTargets     []string      `json:"udp_targets"`
    IPTargets      []string      `json:"ip_targets"`
    TLSConfig      *tls.Config   `json:"-"`
    ConnectTimeout time.Duration `json:"connect_timeout"`
    TestTimeout    time.Duration `json:"test_timeout"`
    ConcurrentTests int          `json:"concurrent_tests"`
    TestDuration   time.Duration `json:"test_duration"`
}
```

### **ICE Configuration:**
```go
type ICEConfig struct {
    StunServers      []string      `json:"stun_servers"`
    TurnServers      []string      `json:"turn_servers"`
    TurnUsername     string        `json:"turn_username"`
    TurnPassword     string        `json:"turn_password"`
    GatheringTimeout time.Duration `json:"gathering_timeout"`
    ConnectionTimeout time.Duration `json:"connection_timeout"`
    TestDuration     time.Duration `json:"test_duration"`
    ConcurrentTests  int           `json:"concurrent_tests"`
}
```

## **🌐 Интеграция с CloudBridge Relay**

Проект полностью совместим с CloudBridge Relay сервером:

- **MASQUE сервер** - Порт 8443 (HTTP/3)
- **STUN сервер** - Порт 19302
- **TURN сервер** - Порт 3478
- **DERP сервер** - Порт 3479

### **Пример интеграции:**
```bash
# Тестирование с CloudBridge Relay
./quck-test -mode=enhanced \
  -masque-server=relay.cloudbridge.example.com:8443 \
  -masque-targets="internal.service:80,db.internal:5432" \
  -ice-stun="relay.cloudbridge.example.com:19302" \
  -ice-turn="relay.cloudbridge.example.com:3478" \
  -ice-turn-user="tenant1" \
  -ice-turn-pass="secure_password"
```

## **📝 Примеры использования**

### **1. Тестирование корпоративной сети**
```bash
# MASQUE для обхода корпоративного firewall
./quck-test -mode=masque \
  -masque-server=corporate-proxy.company.com:8443 \
  -masque-targets="github.com:443,api.external.com:443"
```

### **2. Тестирование NAT traversal**
```bash
# ICE для соединения через NAT
./quck-test -mode=ice \
  -ice-stun="stun.l.google.com:19302,stun1.l.google.com:19302" \
  -ice-turn="turn.company.com:3478" \
  -ice-turn-user="user123" \
  -ice-turn-pass="pass456"
```

### **3. Комплексное тестирование**
```bash
# Полное тестирование CloudBridge стека
./quck-test -mode=enhanced \
  -masque-server=cloudbridge.example.com:8443 \
  -masque-targets="service1.internal:80,service2.internal:443" \
  -ice-stun="cloudbridge.example.com:19302" \
  -ice-turn="cloudbridge.example.com:3478" \
  -ice-turn-user="tenant_123" \
  -ice-turn-pass="secure_key"
```

## **🛡️ Безопасность**

- **TLS 1.3** - Все соединения зашифрованы
- **JWT Authentication** - Аутентификация через JWT токены
- **Certificate Validation** - Проверка сертификатов
- **Secure Defaults** - Безопасные настройки по умолчанию

## **📚 Дополнительные ресурсы**

- [RFC 9298 - CONNECT-UDP](https://tools.ietf.org/html/rfc9298)
- [RFC 9484 - CONNECT-IP](https://tools.ietf.org/html/rfc9484)
- [RFC 9297 - HTTP Datagrams](https://tools.ietf.org/html/rfc9297)
- [RFC 5389 - STUN](https://tools.ietf.org/html/rfc5389)
- [RFC 5766 - TURN](https://tools.ietf.org/html/rfc5766)

## **🤝 Поддержка**

Для вопросов и поддержки обращайтесь к команде CloudBridge 2GC.

---

**© 2024 CloudBridge 2GC. Все права защищены.**

