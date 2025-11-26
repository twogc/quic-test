# quic-test

Professional QUIC protocol testing platform for network engineers, researchers, and educators.

[![CI](https://github.com/twogc/quic-test/actions/workflows/pipeline.yml/badge.svg)](https://github.com/twogc/quic-test/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/twogc/quic-test)](https://goreportcard.com/report/github.com/twogc/quic-test)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](readme_en.md) | **Русский**

## Что это?

`quic-test` — это инструмент для тестирования и анализа производительности протокола QUIC. Разработан для образовательных и исследовательских целей, с акцентом на воспроизводимость результатов и детальную аналитику.

**Основные возможности:**
- Измерение latency, jitter, throughput для QUIC и TCP
- Эмуляция различных сетевых условий (потери, задержки, bandwidth)
- Real-time TUI визуализация (`quic-bottom`)
- Экспорт метрик в Prometheus
- Интеграция с машинным обучением (AI Routing Lab)

## Quick Start

### Docker (рекомендуется)

```bash
# Запуск клиента (тест производительности)
docker run mlanies/quic-test:latest --mode=client --server=demo.quic.tech:4433

# Запуск сервера
docker run -p 4433:4433/udp mlanies/quic-test:latest --mode=server
```

### Из исходников

```bash
# Требования: Go 1.21+, clang (для FEC)
git clone https://github.com/twogc/quic-test
cd quic-test

# Сборка FEC библиотеки
cd internal/fec && make && cd ../..

# Сборка
go build -o quic-test cmd/quic-test/main.go

# Запуск
./quic-test --mode=client --server=demo.quic.tech:4433
```

## Основные режимы

```bash
# Простой тест latency/throughput
./quic-test --mode=client --server=localhost:4433 --duration=30s

# Сравнение QUIC vs TCP
./quic-test --mode=client --compare-tcp --duration=60s

# Эмуляция мобильной сети
./quic-test --profile=mobile --duration=30s

# TUI мониторинг
quic-bottom --server=localhost:4433
```

## Архитектура

```
quic-test/
├── cmd/
│   ├── quic-test/      # Основной CLI
│   └── quic-bottom/    # TUI мониторинг
├── client/             # QUIC клиент
├── server/             # QUIC сервер
├── internal/
│   ├── quic/           # QUIC логика
│   ├── fec/            # Forward Error Correction (C++/AVX2)
│   ├── metrics/        # Prometheus метрики
│   └── congestion/     # BBRv2/BBRv3
└── docs/               # Документация
```

**Подробнее:** [docs/architecture.md](docs/architecture.md)

## Возможности

### ✅ Работает и протестировано

- QUIC client/server (на базе quic-go)
- Измерение RTT, jitter, throughput
- Эмуляция сетевых профилей (mobile, satellite, fiber)
- TUI визуализация (`quic-bottom`)
- Prometheus экспорт
- BBRv2 congestion control

### ⚗️ Экспериментально

- BBRv3 congestion control
- Forward Error Correction (FEC) с AVX2
- MASQUE VPN тестирование
- TCP-over-QUIC туннелирование
- ICE/STUN/TURN тесты

### 🛠 В планах (Roadmap)

- HTTP/3 load testing
- Автоматическое обнаружение аномалий
- Multi-cloud deployment
- WebTransport support

**Полный roadmap:** [docs/roadmap.md](docs/roadmap.md)

## Документация

- **[CLI Reference](docs/cli.md)** — полная справка по командам
- **[Architecture](docs/architecture.md)** — детальная архитектура
- **[Education](docs/education.md)** — лабораторные работы для университетов
- **[AI Integration](docs/ai-routing-integration.md)** — интеграция с AI Routing Lab
- **[Case Studies](docs/case-studies.md)** — результаты тестов с методикой

## Для университетов

Проект разработан с акцентом на образование. Включает готовые лабораторные работы:

- **ЛР #1:** Основы QUIC — handshake, 0-RTT, миграция соединений
- **ЛР #2:** Congestion Control — сравнение BBRv2 vs BBRv3
- **ЛР #3:** Производительность — QUIC vs TCP в различных условиях

**Подробнее:** [docs/education.md](docs/education.md)

## Интеграция с AI Routing Lab

`quic-test` экспортирует метрики в Prometheus, которые используются в [AI Routing Lab](https://github.com/twogc/ai-routing-lab) для обучения моделей предсказания оптимальных маршрутов.

**Пример:**
```bash
# Запуск с Prometheus экспортом
./quic-test --mode=server --prometheus-port=9090

# AI Routing Lab собирает метрики
curl http://localhost:9090/metrics
```

**Подробнее:** [docs/ai-routing-integration.md](docs/ai-routing-integration.md)

## Разработка

```bash
# Запуск тестов
go test ./...

# Линтинг
golangci-lint run

# Сборка Docker образа
docker build -t quic-test .
```

## Лицензия

MIT License. См. [LICENSE](LICENSE).

## Контакты

- **GitHub:** [twogc/quic-test](https://github.com/twogc/quic-test)
- **Блог:** [cloudbridge-research.ru](https://cloudbridge-research.ru)
- **Email:** research@cloudbridge-research.ru

---

**Примечание:** Проект находится в активной разработке. Для production использования рекомендуется дождаться релиза v1.0.0.
