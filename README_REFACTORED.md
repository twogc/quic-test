# QUIC Testing Tool - Рефакторированная версия

## 🚀 Обзор

Это рефакторированная версия QUIC Testing Tool с улучшенной архитектурой и модульной структурой. Проект теперь организован по принципу разделения ответственности (Separation of Concerns) с четким разделением на команды, модули и компоненты.

## 📁 Структура проекта

```
quck-test/
├── cmd/                    # Команды CLI
│   ├── dashboard/          # Dashboard команда
│   ├── masque/            # MASQUE тестирование
│   ├── ice/               # ICE/STUN/TURN тестирование
│   └── enhanced/          # Расширенное тестирование
├── internal/              # Внутренние модули
│   ├── cli/               # CLI логика
│   ├── masque/            # MASQUE протокол
│   ├── ice/               # ICE/STUN/TURN
│   └── integration/       # Интеграция компонентов
├── static/                # Статические файлы
│   └── js/                # JavaScript для UI
├── main.go                # Главный файл (упрощенный)
├── Makefile               # Автоматизация сборки
└── README_REFACTORED.md   # Этот файл
```

## 🛠️ Команды

### Основные команды

```bash
# Сборка проекта
make build

# Запуск dashboard
make run-dashboard

# Запуск MASQUE тестирования
make run-masque

# Запуск ICE тестирования
make run-ice

# Запуск расширенного тестирования
make run-enhanced
```

### CLI команды

```bash
# Запуск dashboard
./build/quck-test dashboard

# Запуск MASQUE тестирования
./build/quck-test masque

# Запуск ICE тестирования
./build/quck-test ice

# Запуск расширенного тестирования
./build/quck-test enhanced
```

## 🔧 Разработка

### Настройка окружения

```bash
# Установка зависимостей
make deps

# Форматирование кода
make fmt

# Проверка линтером
make lint

# Запуск тестов
make test

# Полная настройка для разработки
make dev-setup
```

### Сборка

```bash
# Обычная сборка
make build

# CI сборка (с тестами и проверками)
make ci-build

# Очистка
make clean
```

## 📊 Поддерживаемые протоколы

- **QUIC** (RFC 9000) - Основной протокол
- **MASQUE** (RFC 9298, RFC 9484) - Туннелирование UDP/IP
- **ICE/STUN/TURN** - NAT traversal
- **HTTP/3** - HTTP поверх QUIC
- **WebSocket** - WebSocket туннелирование

## 🎯 Режимы тестирования

1. **server** - QUIC сервер
2. **client** - QUIC клиент
3. **test** - Комбинированное тестирование
4. **dashboard** - Веб-интерфейс
5. **masque** - MASQUE тестирование
6. **ice** - ICE/STUN/TURN тестирование
7. **enhanced** - Расширенное тестирование

## 🔍 Архитектурные улучшения

### 1. Модульность
- Каждая команда в отдельном пакете
- Четкое разделение ответственности
- Легкое добавление новых команд

### 2. CLI структура
- Упрощенный main.go
- Централизованная обработка команд
- Гибкая система флагов

### 3. Сборка и развертывание
- Makefile для автоматизации
- Поддержка CI/CD
- Простая установка

### 4. Тестирование
- Модульные тесты
- Интеграционные тесты
- Автоматизированная проверка

## 🚀 Быстрый старт

```bash
# Клонирование и настройка
git clone <repository>
cd quck-test

# Установка зависимостей
make deps

# Сборка
make build

# Запуск dashboard
make run-dashboard
```

## 📈 Мониторинг и метрики

- **Prometheus** метрики
- **Grafana** дашборды
- **Real-time** мониторинг
- **Performance** анализ

## 🔒 Безопасность

- **TLS 1.3** поддержка
- **Certificate** управление
- **Secure** коммуникация
- **Authentication** системы

## 📚 Документация

- [Архитектура](ARCHITECTURE.md)
- [API Reference](API.md)
- [Конфигурация](CONFIG.md)
- [Развертывание](DEPLOYMENT.md)

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Внесите изменения
4. Добавьте тесты
5. Создайте Pull Request

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE) файл

## 🆘 Поддержка

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Documentation**: Wiki
- **Community**: Discord/Telegram

---

**Примечание**: Это рефакторированная версия с улучшенной архитектурой. Для обратной совместимости сохранены все существующие функции, но с лучшей организацией кода.

