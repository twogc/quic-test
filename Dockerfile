# 2GC CloudBridge QUICK testing - Multi-stage Dockerfile
# Сборка всех компонентов в одном образе

# Этап 1: Сборка
FROM golang:1.21-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git make

# Установка рабочей директории
WORKDIR /app

# Копирование go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка всех компонентов
RUN make build

# Этап 2: Runtime
FROM alpine:3.20

# Установка зависимостей runtime
RUN apk add --no-cache ca-certificates tzdata

# Создание пользователя для безопасности
RUN addgroup -g 1001 -S quck && \
    adduser -u 1001 -S quck -G quck

# Установка рабочей директории
WORKDIR /app

# Копирование собранных бинарников
COPY --from=builder /app/build/ ./

# Копирование статических файлов
COPY --from=builder /app/static/ ./static/
COPY --from=builder /app/index.html ./

# Копирование документации
COPY --from=builder /app/README.md ./
COPY --from=builder /app/LICENSE ./

# Установка прав доступа
RUN chown -R quck:quck /app
USER quck

# Открытие портов
EXPOSE 9000 9990 2112 2113 6060

# Переменные окружения
ENV QUICK_SERVER_ADDR=:9000
ENV QUICK_DASHBOARD_ADDR=:9990
ENV QUICK_PROMETHEUS_CLIENT_PORT=2112
ENV QUICK_PROMETHEUS_SERVER_PORT=2113
ENV QUICK_PPROF_ADDR=:6060

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9990/ || exit 1

# Метки для контейнера
LABEL org.opencontainers.image.title="2GC CloudBridge QUICK testing"
LABEL org.opencontainers.image.description="QUIC performance testing tool with dashboard"
LABEL org.opencontainers.image.url="https://github.com/cloudbridge-relay-installer/quck-test"
LABEL org.opencontainers.image.source="https://github.com/cloudbridge-relay-installer/quck-test"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="2GC CloudBridge"

# Команда по умолчанию
CMD ["./quck-test", "--mode=test", "--connections=2", "--streams=4", "--rate=100", "--prometheus"]
