# quck-test

Утилита для тестирования производительности и устойчивости QUIC-протокола (QUIC + TLS 1.3).

## Запуск

### Сервер
```
go run main.go --mode=server --addr=:9000
```

### Клиент
```
go run main.go --mode=client --addr=127.0.0.1:9000 --connections=2 --streams=4 --packet-size=1200 --rate=100 --report=report.md --report-format=md --pattern=random
```

### Тест (сервер+клиент)
```
go run main.go --mode=test
```

## Описание флагов
- `--mode` — режим работы: `server`, `client`, `test` (по умолчанию `test`)
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

## Расширенные возможности
- **Расширенные метрики:**
  - Percentile latency (p50, p95, p99), jitter, packet loss, retransmits, handshake time, session resumption, 0-RTT/1-RTT, flow control, key update, out-of-order, error breakdown.
- **Временные ряды:**
  - Для latency, throughput, packet loss, retransmits, handshake time и др.
- **ASCII-графики:**
  - В отчёте Markdown для всех ключевых метрик (asciigraph).
- **Ramp-up/ramp-down:**
  - Скорость отправки пакетов динамически увеличивается и уменьшается для стресс-тестирования.
- **Эмуляция плохих сетей:**
  - Задержки, потери, дублирование пакетов (см. параметры выше).
- **Интеграция с CI/CD:**
  - JSON-отчёты, exit code по SLA.
- **Prometheus:**
  - Экспорт live-метрик для мониторинга.

## Пример запуска с эмуляцией плохой сети
```
go run main.go --mode=client --addr=127.0.0.1:9000 --connections=2 --streams=4 --packet-size=1200 --rate=200 --emulate-loss=0.05 --emulate-latency=20ms --emulate-dup=0.01 --report=report.md
```

## Поведение по умолчанию
- Если не указан `--duration`, тест продолжается до ручного завершения (Ctrl+C).
- После завершения теста автоматически формируется и сохраняется отчёт в выбранном формате.

## Примеры отчётов
- Markdown, CSV, JSON — содержат параметры теста, агрегированные метрики, временные ряды, ASCII-графики, ошибки.

## Зависимости
- [quic-go](https://github.com/lucas-clemente/quic-go)
- [tablewriter](https://github.com/olekukonko/tablewriter)
- [asciigraph](https://github.com/guptarohit/asciigraph)
- [prometheus/client_golang](https://github.com/prometheus/client_golang)