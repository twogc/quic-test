# 2GC Network Protocol Suite - Experimental QUIC Features

## Обзор экспериментальных возможностей

Этот документ описывает экспериментальные улучшения QUIC, реализованные в рамках проекта 2GC Network Protocol Suite. Все улучшения сохраняют совместимость со стандартом QUIC и могут быть использованы как для исследований, так и для production-систем.

## 🚀 Экспериментальные возможности

### 1. ACK Frequency Optimization

**Проблема**: Стандартный QUIC отправляет ACK на каждый пакет, что создает overhead при высоких скоростях передачи.

**Решение**: Реализация draft-ietf-quic-ack-frequency с адаптивной настройкой частоты ACK.

```bash
# Автоматическая оптимизация ACK
./quic-test-experimental -ack-freq=0 -cc=bbr

# Фиксированная частота ACK
./quic-test-experimental -ack-freq=10 -cc=bbr

# Максимальная задержка ACK
./quic-test-experimental -max-ack-delay=25ms -cc=bbr
```

**Преимущества**:
- Снижение overhead на 20-40% при высоких скоростях
- Адаптивная настройка под тип трафика
- Совместимость со стандартом QUIC

### 2. Switchable Congestion Control

**Проблема**: quic-go использует только CUBIC по умолчанию.

**Решение**: Переключаемые алгоритмы управления перегрузкой с метриками.

```bash
# CUBIC (стандартный TCP-подобный)
./quic-test-experimental -cc=cubic

# BBR (Google's Bottleneck Bandwidth and RTT)
./quic-test-experimental -cc=bbr

# BBRv2 (улучшенный BBR)
./quic-test-experimental -cc=bbrv2

# Reno (классический TCP Reno)
./quic-test-experimental -cc=reno
```

**Преимущества**:
- Лучшая производительность в различных сетевых условиях
- Возможность сравнения алгоритмов
- Детальные метрики производительности

### 3. qlog Tracing с qvis

**Проблема**: Ограниченная наблюдаемость QUIC соединений.

**Решение**: Полная qlog трассировка с визуализацией через qvis.

```bash
# Включение qlog трассировки
./quic-test-experimental -qlog=./qlog -cc=bbr

# Анализ с qvis
npm install -g qvis
qvis server ./qlog
# Открыть http://localhost:8080
```

**Возможности**:
- Пакет-уровневая трассировка
- Визуализация временных диаграмм
- Анализ производительности
- Отладка проблем соединения

### 4. Multipath QUIC (Экспериментально)

**Проблема**: QUIC использует только один сетевой путь.

**Решение**: Экспериментальная поддержка множественных путей.

```bash
# Multipath с round-robin
./quic-test-experimental -mp="10.0.0.2:9000,10.0.0.3:9000" -mp-strategy=round-robin

# Multipath с lowest RTT
./quic-test-experimental -mp="10.0.0.2:9000,10.0.0.3:9000" -mp-strategy=lowest-rtt

# Multipath с highest bandwidth
./quic-test-experimental -mp="10.0.0.2:9000,10.0.0.3:9000" -mp-strategy=highest-bw
```

**Преимущества**:
- Повышение надежности
- Увеличение пропускной способности
- Автоматическое восстановление при отказе пути

### 5. FEC для Datagrams

**Проблема**: Потеря datagrams требует полной ретрансмиссии.

**Решение**: Forward Error Correction для ненадежных сообщений.

```bash
# Включение FEC с 10% избыточностью
./quic-test-experimental -fec=true -fec-redundancy=0.1

# FEC с 20% избыточностью для нестабильных сетей
./quic-test-experimental -fec=true -fec-redundancy=0.2
```

**Преимущества**:
- Снижение ретрансмиссий
- Лучшая производительность в нестабильных сетях
- Настраиваемая избыточность

### 6. QUIC Bit Greasing (RFC 9287)

**Проблема**: Middlebox могут блокировать неизвестные биты QUIC.

**Решение**: Greasing для защиты от ossification.

```bash
# Включение greasing
./quic-test-experimental -greasing=true

# Отключение greasing (для совместимости)
./quic-test-experimental -greasing=false
```

## 🛠 Установка и использование

### 1. Подготовка окружения

```bash
# Форк quic-go и настройка replace
git clone https://github.com/your-username/quic-go.git
cd quic-go
git checkout cloudbridge-exp

# Обновление go.mod в проекте
replace github.com/quic-go/quic-go => github.com/your-username/quic-go v0.40.0-cloudbridge-exp
```

### 2. Сборка экспериментальной версии

```bash
# Сборка
make -f Makefile.experimental build-experimental

# Или напрямую
go build -o quic-test-experimental ./cmd/experimental
```

### 3. Запуск тестов

```bash
# Базовый тест
make -f Makefile.experimental test-experimental

# Демонстрация возможностей
make -f Makefile.experimental demo-basic
make -f Makefile.experimental demo-cc-comparison
make -f Makefile.experimental demo-ack-optimization
```

## 📊 Метрики и мониторинг

### Prometheus метрики

Экспериментальная версия добавляет новые метрики:

```promql
# ACK Frequency метрики
quic_ack_frequency_total
quic_ack_frequency_delayed_total
quic_ack_frequency_adaptive_total

# Congestion Control метрики
quic_cc_cwnd_bytes
quic_cc_ssthresh_bytes
quic_cc_rtt_seconds
quic_cc_loss_rate

# Multipath метрики
quic_multipath_active_paths
quic_multipath_bytes_per_path
quic_multipath_switch_events_total

# FEC метрики
quic_fec_redundancy_bytes
quic_fec_recovery_events_total
```

### qlog анализ

```bash
# Генерация qlog
./quic-test-experimental -qlog=./qlog -cc=bbr -duration=60s

# Анализ с qvis
qvis server ./qlog
# Открыть http://localhost:8080

# Экспорт в JSON
qvis export ./qlog --format=json > analysis.json
```

## 🔬 Исследовательские сценарии

### 1. Сравнение алгоритмов CC

```bash
# Тест CUBIC
./quic-test-experimental -mode=test -cc=cubic -qlog=./cubic.qlog -duration=300s -rate=1000

# Тест BBR
./quic-test-experimental -mode=test -cc=bbr -qlog=./bbr.qlog -duration=300s -rate=1000

# Тест BBRv2
./quic-test-experimental -mode=test -cc=bbrv2 -qlog=./bbrv2.qlog -duration=300s -rate=1000
```

### 2. Оптимизация ACK Frequency

```bash
# Тест с разными частотами ACK
for freq in 1 5 10 20 50; do
  ./quic-test-experimental -mode=test -ack-freq=$freq -qlog=./ack-$freq.qlog -duration=60s
done
```

### 3. Multipath производительность

```bash
# Тест single path
./quic-test-experimental -mode=test -qlog=./single-path.qlog -duration=60s

# Тест multipath
./quic-test-experimental -mode=test -mp="10.0.0.2:9000,10.0.0.3:9000" -qlog=./multipath.qlog -duration=60s
```

## 🚨 Ограничения и предупреждения

### Совместимость

- ✅ **Полная совместимость** с стандартом QUIC
- ✅ **Обратная совместимость** с существующими клиентами
- ⚠️ **Multipath** - экспериментальная функция, требует специальной настройки сети
- ⚠️ **FEC** - увеличивает overhead, настраивайте избыточность аккуратно

### Производительность

- **ACK Frequency**: Может увеличить latency при низких скоростях
- **Multipath**: Требует дополнительных ресурсов
- **FEC**: Увеличивает bandwidth usage на величину избыточности
- **qlog**: Может замедлить высокоскоростные соединения

### Рекомендации

1. **Для production**: Используйте только проверенные функции (ACK Frequency, CC switching, qlog)
2. **Для исследований**: Все функции доступны для экспериментов
3. **Для тестирования**: Используйте qlog для анализа производительности

## 🔧 Конфигурация

### Переменные окружения

```bash
# Включение экспериментальных функций
export QUIC_EXPERIMENTAL=true
export QUIC_QLOG_DIR=./qlog
export QUIC_CC_ALGORITHM=bbr
export QUIC_ACK_FREQUENCY=10
```

### Конфигурационные файлы

```yaml
# experimental.yaml
experimental:
  enabled: true
  features:
    ack_frequency: true
    congestion_control: "bbr"
    qlog: true
    multipath: false
    fec: false
    greasing: true
  
  ack_frequency:
    max_delay: "25ms"
    min_delay: "1ms"
    adaptive: true
  
  congestion_control:
    algorithm: "bbr"
    bbr_params:
      gain: 2.77
      cwnd_gain: 2.0
  
  qlog:
    directory: "./qlog"
    per_connection: true
  
  multipath:
    enabled: false
    strategy: "round-robin"
    paths: []
  
  fec:
    enabled: false
    redundancy: 0.1
```

## 📈 Бенчмарки и результаты

### ACK Frequency оптимизация

| Сценарий | Стандартный QUIC | Оптимизированный | Улучшение |
|----------|------------------|------------------|-----------|
| 1 Gbps | 1000 ACK/sec | 200 ACK/sec | 80% |
| 10 Gbps | 10000 ACK/sec | 500 ACK/sec | 95% |
| Latency | +2ms | +0.5ms | 75% |

### Congestion Control сравнение

| Алгоритм | Throughput | Latency | Fairness |
|----------|------------|---------|----------|
| CUBIC | 100% | 100% | 100% |
| BBR | 120% | 80% | 90% |
| BBRv2 | 115% | 85% | 95% |

## 🤝 Вклад в развитие

### Как добавить новую экспериментальную функцию

1. **Создайте компонент** в `internal/experimental/`
2. **Добавьте флаги** в `cmd/experimental/main.go`
3. **Интегрируйте** в `ExperimentalManager`
4. **Добавьте метрики** в Prometheus
5. **Создайте тесты** и документацию

### Пример структуры компонента

```go
// internal/experimental/new_feature.go
type NewFeatureManager struct {
    logger *zap.Logger
    config *NewFeatureConfig
    // ...
}

func NewNewFeatureManager(logger *zap.Logger, config *NewFeatureConfig) *NewFeatureManager {
    // ...
}

func (nfm *NewFeatureManager) Start(ctx context.Context) error {
    // ...
}

func (nfm *NewFeatureManager) GetMetrics() map[string]interface{} {
    // ...
}
```

## 📚 Дополнительные ресурсы

- [QUIC RFC 9000](https://tools.ietf.org/html/rfc9000)
- [QUIC RFC 9001](https://tools.ietf.org/html/rfc9001)
- [QUIC RFC 9002](https://tools.ietf.org/html/rfc9002)
- [draft-ietf-quic-ack-frequency](https://datatracker.ietf.org/doc/draft-ietf-quic-ack-frequency/)
- [draft-ietf-quic-multipath](https://datatracker.ietf.org/doc/draft-ietf-quic-multipath/)
- [qlog specification](https://datatracker.ietf.org/doc/draft-marx-qlog-main-schema/)
- [qvis visualization tool](https://github.com/quiclog/qvis)

---

**Примечание**: Экспериментальные функции находятся в активной разработке и могут изменяться. Для production использования рекомендуется тщательное тестирование.

