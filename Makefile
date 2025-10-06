# QUIC Testing Tool Makefile

.PHONY: build clean test run-dashboard run-masque run-ice run-enhanced help

# Переменные
BINARY_NAME=quck-test
BUILD_DIR=build
GO_VERSION=1.25.1

# Цвета для вывода
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(CYAN)QUIC Testing Tool - Расширенное тестирование QUIC протокола$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Собрать бинарный файл
	@echo "$(BLUE)Сборка $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)Сборка завершена: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

clean: ## Очистить build директорию
	@echo "$(YELLOW)Очистка build директории...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Очистка завершена$(NC)"

test: ## Запустить тесты
	@echo "$(BLUE)Запуск тестов...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)Тесты завершены$(NC)"

run-dashboard: build ## Запустить dashboard
	@echo "$(PURPLE)Запуск QUIC Testing Dashboard...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) dashboard

run-masque: build ## Запустить MASQUE тестирование
	@echo "$(PURPLE)Запуск MASQUE тестирования...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) masque

run-ice: build ## Запустить ICE тестирование
	@echo "$(PURPLE)Запуск ICE/STUN/TURN тестирования...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) ice

run-enhanced: build ## Запустить расширенное тестирование
	@echo "$(PURPLE)Запуск расширенного тестирования...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) enhanced

install: build ## Установить в систему
	@echo "$(BLUE)Установка $(BINARY_NAME) в систему...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)Установка завершена$(NC)"

uninstall: ## Удалить из системы
	@echo "$(YELLOW)Удаление $(BINARY_NAME) из системы...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)Удаление завершено$(NC)"

deps: ## Установить зависимости
	@echo "$(BLUE)Установка зависимостей...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Зависимости установлены$(NC)"

fmt: ## Форматировать код
	@echo "$(BLUE)Форматирование кода...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Форматирование завершено$(NC)"

lint: ## Проверить код линтером
	@echo "$(BLUE)Проверка кода линтером...$(NC)"
	@go vet ./...
	@echo "$(GREEN)Проверка завершена$(NC)"

# Специальные команды для разработки
dev-setup: deps fmt lint ## Настроить окружение для разработки
	@echo "$(GREEN)Окружение для разработки настроено$(NC)"

# Команды для CI/CD
ci-build: deps fmt lint test build ## Полная сборка для CI
	@echo "$(GREEN)CI сборка завершена$(NC)"

# Показать информацию о проекте
info: ## Показать информацию о проекте
	@echo "$(CYAN)==============================$(NC)"
	@echo "$(CYAN)  QUIC Testing Tool Info$(NC)"
	@echo "$(CYAN)==============================$(NC)"
	@echo "$(YELLOW)Go версия:$(NC) $(shell go version)"
	@echo "$(YELLOW)Проект:$(NC) $(BINARY_NAME)"
	@echo "$(YELLOW)Build директория:$(NC) $(BUILD_DIR)"
	@echo "$(YELLOW)Поддерживаемые протоколы:$(NC)"
	@echo "  - QUIC (RFC 9000)"
	@echo "  - MASQUE (RFC 9298, RFC 9484)"
	@echo "  - ICE/STUN/TURN"
	@echo "  - HTTP/3"
	@echo "  - WebSocket"
	@echo "$(YELLOW)Режимы тестирования:$(NC)"
	@echo "  - server: QUIC сервер"
	@echo "  - client: QUIC клиент"
	@echo "  - test: Комбинированное тестирование"
	@echo "  - dashboard: Веб-интерфейс"
	@echo "  - masque: MASQUE тестирование"
	@echo "  - ice: ICE/STUN/TURN тестирование"
	@echo "  - enhanced: Расширенное тестирование"