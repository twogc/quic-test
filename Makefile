# 2GC CloudBridge QUICK testing - Makefile
# Автоматизация сборки, тестирования и развертывания

.PHONY: help build clean test run-dashboard run-server run-client run-test docker-build docker-run lint fmt vet

# Переменные
BINARY_NAME=quck-test
DASHBOARD_BINARY=dashboard
CLIENT_BINARY=quic-client
SERVER_BINARY=quic-server
BUILD_DIR=build
DOCKER_IMAGE=quck-test
DOCKER_TAG=latest

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)2GC CloudBridge QUICK testing$(NC)"
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: clean ## Собрать все бинарные файлы
	@echo "$(GREEN)Сборка основных компонентов...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@echo "$(YELLOW)Сборка основного бинарника...$(NC)"
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(YELLOW)Сборка dashboard...$(NC)"
	@go build -o $(BUILD_DIR)/$(DASHBOARD_BINARY) ./cmd/dashboard
	@echo "$(YELLOW)Сборка QUIC клиента...$(NC)"
	@go build -o $(BUILD_DIR)/$(CLIENT_BINARY) ./cmd/quic-client
	@echo "$(YELLOW)Сборка QUIC сервера...$(NC)"
	@go build -o $(BUILD_DIR)/$(SERVER_BINARY) ./cmd/quic-server
	@echo "$(GREEN)Сборка завершена! Файлы в $(BUILD_DIR)/$(NC)"

clean: ## Очистить собранные файлы
	@echo "$(YELLOW)Очистка...$(NC)"
	@rm -rf $(BUILD_DIR)
	@go clean

test: ## Запустить тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии: coverage.html$(NC)"

test-integration: ## Запустить интеграционные тесты
	@echo "$(GREEN)Запуск интеграционных тестов...$(NC)"
	@go test -v -tags=integration ./...

run-dashboard: build ## Запустить веб-дашборд
	@echo "$(GREEN)Запуск веб-дашборда...$(NC)"
	@$(BUILD_DIR)/$(DASHBOARD_BINARY) --addr=:9990

run-server: build ## Запустить QUIC сервер
	@echo "$(GREEN)Запуск QUIC сервера...$(NC)"
	@$(BUILD_DIR)/$(SERVER_BINARY) --addr=:9000 --prometheus

run-client: build ## Запустить QUIC клиент
	@echo "$(GREEN)Запуск QUIC клиента...$(NC)"
	@$(BUILD_DIR)/$(CLIENT_BINARY) --addr=127.0.0.1:9000 --connections=2 --streams=4 --rate=100

run-test: build ## Запустить полный тест (сервер+клиент)
	@echo "$(GREEN)Запуск полного теста...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --connections=2 --streams=4 --rate=100 --duration=30s

# Docker команды
docker-build: ## Собрать Docker образы
	@echo "$(GREEN)Сборка Docker образов...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@docker build -t $(DOCKER_IMAGE):dashboard -f Dockerfile.dashboard .

docker-run: ## Запустить через Docker Compose
	@echo "$(GREEN)Запуск через Docker Compose...$(NC)"
	@docker-compose up --build

docker-stop: ## Остановить Docker Compose
	@echo "$(YELLOW)Остановка Docker Compose...$(NC)"
	@docker-compose down

# Качество кода
lint: ## Запустить линтеры
	@echo "$(GREEN)Запуск линтеров...$(NC)"
	@golangci-lint run
	@gosec ./...

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	@go fmt ./...

vet: ## Запустить go vet
	@echo "$(GREEN)Запуск go vet...$(NC)"
	@go vet ./...

# Разработка
dev-setup: ## Настройка окружения для разработки
	@echo "$(GREEN)Настройка окружения для разработки...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Установка инструментов разработки...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Профилирование
profile-cpu: build ## Запустить с профилированием CPU
	@echo "$(GREEN)Запуск с профилированием CPU...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --pprof-addr=:6060 --duration=60s

profile-mem: build ## Запустить с профилированием памяти
	@echo "$(GREEN)Запуск с профилированием памяти...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --pprof-addr=:6060 --duration=60s

# Сценарии тестирования
test-scenarios: build ## Запустить все сценарии тестирования
	@echo "$(GREEN)Запуск сценариев тестирования...$(NC)"
	@echo "$(YELLOW)Сценарий: many-streams$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --streams=100 --connections=1 --duration=30s --report=report-many-streams.md
	@echo "$(YELLOW)Сценарий: loss-burst$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --emulate-loss=0.1 --emulate-latency=50ms --duration=30s --report=report-loss-burst.md
	@echo "$(YELLOW)Сценарий: reorder$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --emulate-dup=0.05 --duration=30s --report=report-reorder.md

# Сетевые профили
test-wifi: build ## Тест с профилем WiFi
	@echo "$(GREEN)Тест с профилем WiFi...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --emulate-loss=0.02 --emulate-latency=10ms --duration=30s --report=report-wifi.md

test-lte: build ## Тест с профилем LTE
	@echo "$(GREEN)Тест с профилем LTE...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --emulate-loss=0.05 --emulate-latency=30ms --duration=30s --report=report-lte.md

test-sat: build ## Тест с профилем спутниковой связи
	@echo "$(GREEN)Тест с профилем спутниковой связи...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --emulate-loss=0.01 --emulate-latency=500ms --duration=30s --report=report-sat.md

# SLA тестирование
test-sla: build ## Запустить SLA тестирование
	@echo "$(GREEN)Запуск SLA тестирования...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --mode=test --sla-rtt-p95=100ms --sla-loss=0.01 --duration=60s --report=report-sla.json --report-format=json

# Установка зависимостей
install-deps: ## Установить зависимости
	@echo "$(GREEN)Установка зависимостей...$(NC)"
	@go mod download
	@go mod tidy

# Генерация документации
docs: ## Генерация документации
	@echo "$(GREEN)Генерация документации...$(NC)"
	@godoc -http=:6060 &
	@echo "$(GREEN)Документация доступна на http://localhost:6060$(NC)"

# Полная сборка и тестирование
all: clean install-deps fmt vet lint test build ## Полная сборка и тестирование
	@echo "$(GREEN)Все этапы выполнены успешно!$(NC)"

# По умолчанию показываем справку
.DEFAULT_GOAL := help