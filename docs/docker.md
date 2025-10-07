# Docker Deployment Guide

## Обзор

2GC Network Protocol Suite поддерживает развертывание в Docker контейнерах для изоляции сервера и клиента, что идеально подходит для тестирования сетевых протоколов.

## Быстрый старт

### 1. Запуск полного стека

```bash
# Запуск всех сервисов (сервер + клиент + дашборд + мониторинг)
./scripts/docker-compose.sh up --build

# Запуск в фоновом режиме
./scripts/docker-compose.sh up --build --detach
```

### 2. Запуск только сервера

```bash
# Запуск QUIC сервера
./scripts/docker-server.sh

# С настройкой параметров
SERVER_ADDR=":8443" ./scripts/docker-server.sh
```

### 3. Запуск только клиента

```bash
# Запуск QUIC клиента
./scripts/docker-client.sh

# С настройкой параметров
SERVER_ADDR="localhost:8443" CONNECTIONS=4 STREAMS=8 RATE=200 DURATION=120s ./scripts/docker-client.sh
```

### 4. Запуск тестирования

```bash
# Запуск полного теста (сервер + клиент)
./scripts/docker-test.sh

# С настройкой параметров
CONNECTIONS=4 STREAMS=8 RATE=200 DURATION=120s ./scripts/docker-test.sh
```

## Docker образы

### Основной образ
- **Файл**: `Dockerfile`
- **Назначение**: Универсальный образ для всех режимов
- **Команда**: `docker build -t 2gc-network-suite .`

### Сервер
- **Файл**: `Dockerfile.server`
- **Назначение**: Оптимизированный образ для QUIC сервера
- **Команда**: `docker build -f Dockerfile.server -t 2gc-network-suite:server .`

### Клиент
- **Файл**: `Dockerfile.client`
- **Назначение**: Оптимизированный образ для QUIC клиента
- **Команда**: `docker build -f Dockerfile.client -t 2gc-network-suite:client .`

## Docker Compose

### Основные сервисы

```yaml
services:
  quic-server:     # QUIC сервер
  quic-client:     # QUIC клиент
  dashboard:        # Веб-дашборд
  prometheus:       # Сбор метрик
  grafana:          # Визуализация
  jaeger:          # Трейсинг
```

### Дополнительные сервисы

```yaml
services:
  nginx:           # Балансировка нагрузки
  redis:           # Кэширование
```

### Команды Docker Compose

```bash
# Запуск всех сервисов
./scripts/docker-compose.sh up

# Запуск с пересборкой
./scripts/docker-compose.sh up --build

# Запуск в фоновом режиме
./scripts/docker-compose.sh up --detach

# Остановка всех сервисов
./scripts/docker-compose.sh down

# Просмотр логов
./scripts/docker-compose.sh logs

# Следование за логами
./scripts/docker-compose.sh logs --follow

# Статус сервисов
./scripts/docker-compose.sh status

# Очистка
./scripts/docker-compose.sh clean
```

## Переменные окружения

### Сервер

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `QUIC_SERVER_ADDR` | `:9000` | Адрес сервера |
| `QUIC_PROMETHEUS_SERVER_PORT` | `2113` | Порт Prometheus метрик |
| `QUIC_PPROF_ADDR` | `:6060` | Адрес pprof профилирования |

### Клиент

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `QUIC_CLIENT_ADDR` | `localhost:9000` | Адрес сервера |
| `QUIC_CONNECTIONS` | `2` | Количество соединений |
| `QUIC_STREAMS` | `4` | Количество потоков |
| `QUIC_RATE` | `100` | Скорость отправки (пакетов/сек) |
| `QUIC_DURATION` | `60s` | Длительность теста |
| `QUIC_PROMETHEUS_CLIENT_PORT` | `2112` | Порт Prometheus метрик |

## Порты

| Сервис | Порт | Описание |
|--------|------|----------|
| QUIC сервер | 9000 | Основной QUIC порт |
| Веб-дашборд | 9990 | Веб-интерфейс |
| Prometheus | 9090 | Метрики |
| Grafana | 3000 | Визуализация |
| Jaeger | 16686 | Трейсинг |
| Сервер метрики | 2113 | Prometheus метрики сервера |
| Клиент метрики | 2112 | Prometheus метрики клиента |
| pprof | 6060 | Профилирование |

## Мониторинг

### Prometheus метрики
- Сервер: http://localhost:2113/metrics
- Клиент: http://localhost:2112/metrics
- Prometheus: http://localhost:9090

### Профилирование
- pprof: http://localhost:6060/debug/pprof/

### Визуализация
- Grafana: http://localhost:3000 (admin/admin)
- Jaeger: http://localhost:16686

## Примеры использования

### 1. Базовое тестирование

```bash
# Запуск сервера
./scripts/docker-server.sh

# В другом терминале - запуск клиента
./scripts/docker-client.sh
```

### 2. Нагрузочное тестирование

```bash
# Высокая нагрузка
CONNECTIONS=10 STREAMS=20 RATE=500 DURATION=300s ./scripts/docker-test.sh
```

### 3. Длительное тестирование

```bash
# 24-часовой тест
DURATION=24h ./scripts/docker-test.sh
```

### 4. Мониторинг в реальном времени

```bash
# Запуск с мониторингом
./scripts/docker-compose.sh up --build

# Просмотр метрик
curl http://localhost:2113/metrics
curl http://localhost:2112/metrics
```

## Troubleshooting

### Проблемы с сетью

```bash
# Проверка сетевых подключений
docker network ls
docker network inspect 2gc-network-suite

# Проверка портов
netstat -tlnp | grep -E "(9000|9990|9090|3000)"
```

### Проблемы с контейнерами

```bash
# Просмотр логов
docker logs 2gc-network-server
docker logs 2gc-network-client

# Вход в контейнер
docker exec -it 2gc-network-server sh
docker exec -it 2gc-network-client sh
```

### Очистка

```bash
# Остановка всех контейнеров
./scripts/docker-compose.sh down

# Полная очистка
./scripts/docker-compose.sh clean

# Очистка Docker системы
docker system prune -a
```

## Безопасность

- Все контейнеры запускаются от непривилегированного пользователя
- Используются минимальные Alpine образы
- Ограниченные права доступа к файловой системе
- Изолированная сетевая среда

## Производительность

- Оптимизированные образы для сервера и клиента
- Минимальный overhead контейнеризации
- Эффективное использование ресурсов
- Поддержка масштабирования
