
# 2GC Network Protocol Suite

A comprehensive platform for testing and analyzing network protocols: QUIC, MASQUE, ICE/STUN/TURN and others

## 🚀 Features

- **QUIC Protocol Testing** - Advanced QUIC implementation with experimental features
- **MASQUE Protocol Support** - Tunneling and proxying capabilities  
- **ICE/STUN/TURN Testing** - NAT traversal and P2P connection testing
- **TLS 1.3 Security** - Modern cryptography for secure connections
- **HTTP/3 Support** - HTTP over QUIC implementation
- **Experimental Features** - BBRv2, ACK-Frequency, FEC, Bit Greasing
- **Real-time Monitoring** - Prometheus metrics and Grafana dashboards
- **Comprehensive Testing** - Automated test matrix and regression testing

## Поддерживаемые протоколы

- **QUIC** - Быстрый и надежный транспортный протокол
- **MASQUE** - Протокол для туннелирования и проксирования
- **ICE/STUN/TURN** - Протоколы для NAT traversal и P2P соединений
- **TLS 1.3** - Современная криптография для безопасных соединений
- **HTTP/3** - HTTP поверх QUIC

[![Смотреть демо-видео](https://customer-aedqzjrbponeadcg.cloudflarestream.com/d31af3803090bcb58597de9fe685a746/thumbnails/thumbnail.jpg)](https://customer-aedqzjrbponeadcg.cloudflarestream.com/d31af3803090bcb58597de9fe685a746/watch)

[![Build](https://github.com/twogc/quic-test/workflows/CI/badge.svg)](https://github.com/twogc/quic-test/actions)
[![Lint](https://github.com/twogc/quic-test/workflows/Lint/badge.svg)](https://github.com/twogc/quic-test/actions)
[![Security](https://github.com/twogc/quic-test/workflows/Security/badge.svg)](https://github.com/twogc/quic-test/security)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

## 🛠️ Usage

### QUIC Testing
```bash
# Server
go run main.go --mode=server --addr=:9000

# Client
go run main.go --mode=client --addr=127.0.0.1:9000 --connections=2 --streams=4 --packet-size=1200 --rate=100 --report=report.md --report-format=md --pattern=random

# Full test (server+client)
go run main.go --mode=test
```

### Experimental QUIC Features
```bash
# BBRv2 Congestion Control
go run main.go --mode=experimental --cc=bbrv2 --ackfreq=3 --fec=0.1

# ACK Frequency Optimization
go run main.go --mode=experimental --ackfreq=5 --qlog=out.qlog

# FEC with Bit Greasing
go run main.go --mode=experimental --fec=0.2 --greasing=true
```

### MASQUE Testing
```bash
go run main.go --mode=masque --masque-server=localhost:8443 --masque-targets=8.8.8.8:53,1.1.1.1:53
```

### ICE/STUN/TURN Testing
```bash
go run main.go --mode=ice --ice-stun=stun.l.google.com:19302 --ice-turn=turn.example.com:3478
```

### Web Dashboard
```bash
go run main.go --mode=dashboard
```

### Расширенное тестирование
```bash
go run main.go --mode=enhanced
```

## Описание флагов
- `--mode` — режим работы: `server`, `client`, `test`, `dashboard`, `masque`, `ice`, `enhanced` (по умолчанию `test`)
- `--addr` — адрес для подключения или прослушивания (по умолчанию `:9000`)
- `--connections` — количество QUIC-соединений (по умолчанию 1)
- `--streams` — количество потоков на соединение (по умолчанию 1)
- `--duration` — длительность теста (0 — до ручного завершения, по умолчанию 0)
- `--packet-size` — размер пакета в байтах (по умолчанию 1200)
- `--rate` — частота отправки пакетов в секунду (по умолчанию 100, поддерживает ramp-up/ramp-down)
- `--report` — путь к файлу для отчёта (опционально)
- `--report-format` — формат отчёта: `csv`, `md`, `json` (по умолчанию `md`)
- `--cert` — путь к TLS-сертификату (опционально)
- `--key` — путь к TLS-ключу (опционально)
- `--pattern` — шаблон данных: `random`, `zeroes`, `increment` (по умолчанию `random`)
- `--no-tls` — отключить TLS (для тестов)
- `--prometheus` — экспортировать метрики Prometheus на `/metrics`
- `--emulate-loss` — вероятность потери пакета (0..1, например 0.05 для 5%)
- `--emulate-latency` — дополнительная задержка перед отправкой пакета (например, 20ms)
- `--emulate-dup` — вероятность дублирования пакета (0..1)

## SLA проверки
- `--sla-rtt-p95` — максимальный RTT p95 (например, 100ms)
- `--sla-loss` — максимальная потеря пакетов (0..1, например, 0.01 для 1%)
- `--sla-throughput` — минимальная пропускная способность (KB/s)
- `--sla-errors` — максимальное количество ошибок

## QUIC тюнинг
- `--cc` — алгоритм управления перегрузкой: cubic, bbr, reno
- `--max-idle-timeout` — максимальное время простоя соединения
- `--handshake-timeout` — таймаут handshake
- `--keep-alive` — интервал keep-alive
- `--max-streams` — максимальное количество потоков
- `--max-stream-data` — максимальный размер данных потока
- `--enable-0rtt` — включить 0-RTT
- `--enable-key-update` — включить key update
- `--enable-datagrams` — включить datagrams
- `--max-incoming-streams` — максимальное количество входящих потоков
- `--max-incoming-uni-streams` — максимальное количество входящих unidirectional потоков

## Тестовые сценарии
- `--scenario` — предустановленный сценарий: wifi, lte, sat, dc-eu, ru-eu, loss-burst, reorder
- `--list-scenarios` — показать список доступных сценариев

## Сетевые профили
- `--network-profile` — сетевой профиль: wifi, lte, 5g, satellite, ethernet, fiber, datacenter
- `--list-profiles` — показать список доступных сетевых профилей

## Расширенные возможности
- **Расширенные метрики:**
  - Percentile latency (p50, p95, p99, p999), jitter, packet loss, retransmits, handshake time, session resumption, 0-RTT/1-RTT, flow control, key update, out-of-order, error breakdown.
- **Временные ряды:**
  - Для latency, throughput, packet loss, retransmits, handshake time и др.
- **ASCII-графики:**
  - В отчёте Markdown для всех ключевых метрик (asciigraph).
- **Ramp-up/ramp-down:**
  - Скорость отправки пакетов динамически увеличивается и уменьшается для стресс-тестирования.
- **Эмуляция плохих сетей:**
  - Задержки, потери, дублирование пакетов (см. параметры выше).
- **Интеграция с CI/CD:**
  - JSON-отчёты с версионированной схемой, exit code по SLA.
- **Prometheus:**
  - Экспорт live-метрик для мониторинга.
- **SLA проверки:**
  - Автоматическая проверка соответствия метрик SLA требованиям с exit code.
- **QUIC тюнинг:**
  - Настройка алгоритмов управления перегрузкой, таймаутов, потоков, 0-RTT, key update, datagrams.
- **Тестовые сценарии:**
  - Предустановленные сценарии для различных типов сетей (WiFi, LTE, спутниковая связь, дата-центры).
- **Сетевые профили:**
  - Реалистичные профили сетей с конкретными значениями RTT, jitter, loss, bandwidth.
- **Веб-дашборд:**
  - REST API, Server-Sent Events для real-time обновлений, встроенные статические файлы.

## Примеры использования

### Базовый тест с SLA проверками
```
go run main.go --mode=test --sla-rtt-p95=100ms --sla-loss=0.01 --sla-throughput=50 --report=report.json --report-format=json
```

### Тест с QUIC тюнингом
```
go run main.go --mode=test --cc=bbr --enable-0rtt --enable-datagrams --max-streams=100 --keep-alive=30s
```

### Тест с предустановленным сценарием
```
go run main.go --scenario=wifi --report=wifi-test.md
```

### Тест с сетевым профилем
```
go run main.go --network-profile=lte --report=lte-test.json --report-format=json
```

### Запуск веб-дашборда
```
go run cmd/dashboard/dashboard.go --addr=:9990
```

### Список доступных сценариев
```
go run main.go --list-scenarios
```

### Список сетевых профилей
```
go run main.go --list-profiles
```

## Сетевые пресеты

| Пресет | RTT | Jitter | Loss | Bandwidth | Ожидаемый P95 | Описание |
|--------|-----|--------|------|-----------|---------------|----------|
| `wifi` | 20ms | 5ms | 0.1% | 100 Mbps | 25-30ms | Домашний WiFi |
| `lte` | 50ms | 15ms | 0.5% | 50 Mbps | 70-80ms | Мобильный LTE |
| `satellite` | 600ms | 50ms | 1% | 10 Mbps | 650-700ms | Спутниковый интернет |
| `datacenter` | 1ms | 0.1ms | 0% | 10 Gbps | 2-3ms | Локальная сеть ЦОД |
| `eu-ru` | 80ms | 10ms | 0.2% | 1 Gbps | 90-100ms | Между континентами |

## Поведение по умолчанию
- Если не указан `--duration`, тест продолжается до ручного завершения (Ctrl+C).
- После завершения теста автоматически формируется и сохраняется отчёт в выбранном формате.

## Примеры отчётов
- Markdown, CSV, JSON — содержат параметры теста, агрегированные метрики, временные ряды, ASCII-графики, ошибки.

## 🚀 Автоматические релизы

QUIC Test использует автоматическую систему релизов через GitHub Actions.

### Быстрое обновление версии
```bash
# Обновить версию до v1.2.3
./scripts/update-version.sh v1.2.3

# Зафиксировать и отправить
git add tag.txt && git commit -m "chore: bump version to v1.2.3"
git push origin main
```

GitHub Actions автоматически:
- ✅ Создаст Git тег
- ✅ Соберет бинарники для всех платформ (Linux, Windows, macOS)
- ✅ Создаст GitHub Release
- ✅ Опубликует Docker образы

📋 **Подробнее**: [RELEASES.md](RELEASES.md)

## 📚 Documentation

- [Deployment Guide](docs/deployment.md)
- [API Documentation](docs/api.md)
- [Usage Guide](docs/usage.md)
- [Docker Configuration](docs/docker.md)
- [Versioning](docs/versioning.md)

### Research Reports
- [Experimental QUIC Laboratory Research Report](docs/reports/Experimental_QUIC_Laboratory_Research_Report.md)
- [QUIC Performance Comparison Report](docs/reports/QUIC_Performance_Comparison_Report.md)
- [Implementation Complete Report](docs/reports/IMPLEMENTATION_COMPLETE.md)
- [Release Notes](docs/reports/RELEASE_NOTES.md)

## Dependencies
- [quic-go](https://github.com/lucas-clemente/quic-go)
- [tablewriter](https://github.com/olekukonko/tablewriter)
- [asciigraph](https://github.com/guptarohit/asciigraph)
- [prometheus/client_golang](https://github.com/prometheus/client_golang)