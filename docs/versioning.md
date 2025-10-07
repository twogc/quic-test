# Управление версиями

## Обзор

Программа QUIC Testing Tool использует файл `tag.txt` для хранения версии. Версия автоматически читается из этого файла при запуске программы.

## Файл tag.txt

Версия хранится в файле `tag.txt` в корне проекта. Файл должен содержать только версию в формате `vX.Y.Z` (например, `v1.2.3`).

## Использование

### Просмотр версии

```bash
# Показать версию программы
./quic-test --version

# Версия также отображается в справке
./quic-test --help
```

### Обновление версии

#### Автоматическое обновление (рекомендуется)

Используйте скрипт `scripts/update-version.sh`:

```bash
# Обновить до версии v1.2.3
./scripts/update-version.sh v1.2.3
```

#### Ручное обновление

```bash
# Редактировать файл tag.txt
echo "v1.2.3" > tag.txt

# Проверить версию
./quic-test --version
```

## Формат версии

Версия должна следовать семантическому версионированию (Semantic Versioning):

- **MAJOR** (X): Несовместимые изменения в API
- **MINOR** (Y): Новая функциональность с обратной совместимостью
- **PATCH** (Z): Исправления ошибок с обратной совместимостью

Примеры:
- `v1.0.0` - Первая стабильная версия
- `v1.1.0` - Новая функциональность
- `v1.1.1` - Исправление ошибок
- `v2.0.0` - Несовместимые изменения

## Интеграция в CI/CD

### GitHub Actions

```yaml
- name: Update version
  run: |
    echo "v${{ github.ref_name }}" > tag.txt
    git add tag.txt
    git commit -m "Update version to v${{ github.ref_name }}"
    git push
```

### GitLab CI

```yaml
update_version:
  script:
    - echo "v$CI_COMMIT_TAG" > tag.txt
    - git add tag.txt
    - git commit -m "Update version to v$CI_COMMIT_TAG"
    - git push
```

## Программный доступ к версии

В коде Go можно получить версию программно:

```go
import "quic-test/internal"

// Получить версию
version, err := internal.GetVersion()
if err != nil {
    log.Printf("Ошибка получения версии: %v", err)
} else {
    log.Printf("Версия: %s", version)
}

// Получить полную информацию о версии
versionInfo := internal.GetVersionInfo()
fmt.Println(versionInfo) // "QUIC Testing Tool v1.2.3"

// Вывести версию
internal.PrintVersion()
```

## Обработка ошибок

- Если файл `tag.txt` не найден, версия будет `"unknown"`
- Если файл пустой, будет возвращена ошибка
- Если файл содержит неверный формат, версия будет использована как есть

## Тестирование

Тесты для функциональности версий находятся в `internal/version_test.go`:

```bash
# Запустить тесты версий
go test ./internal/version_test.go ./internal/version.go -v
```
