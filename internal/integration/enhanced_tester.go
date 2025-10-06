package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"quic-test/internal/ice"
	"quic-test/internal/masque"
)

// EnhancedTester интегрирует MASQUE и ICE тестирование с существующим QUIC тестированием
type EnhancedTester struct {
	logger *zap.Logger

	// Компоненты тестирования
	masqueTester *masque.MASQUETester
	iceTester    *ice.ICETester

	// Конфигурация
	config *EnhancedConfig

	// Состояние
	mu        sync.RWMutex
	isActive  bool
	startTime time.Time
}

// EnhancedConfig конфигурация для расширенного тестирования
type EnhancedConfig struct {
	// MASQUE конфигурация
	MASQUE masque.MASQUEConfig `json:"masque"`

	// ICE конфигурация
	ICE ice.ICEConfig `json:"ice"`

	// Общие параметры
	TestDuration    time.Duration `json:"test_duration"`
	ConcurrentTests int           `json:"concurrent_tests"`
	EnableMASQUE    bool          `json:"enable_masque"`
	EnableICE       bool          `json:"enable_ice"`
}

// EnhancedMetrics объединенные метрики всех компонентов
type EnhancedMetrics struct {
	// MASQUE метрики
	MASQUE *masque.MASQUEMetrics `json:"masque"`

	// ICE метрики
	ICE *ice.ICEMetrics `json:"ice"`

	// Общие метрики
	TotalTests      int64         `json:"total_tests"`
	SuccessfulTests int64         `json:"successful_tests"`
	FailedTests     int64         `json:"failed_tests"`
	SuccessRate     float64       `json:"success_rate"`
	TestDuration    time.Duration `json:"test_duration"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
}

// NewEnhancedTester создает новый расширенный тестер
func NewEnhancedTester(logger *zap.Logger, config *EnhancedConfig) *EnhancedTester {
	return &EnhancedTester{
		logger: logger,
		config: config,
	}
}

// Start запускает расширенное тестирование
func (et *EnhancedTester) Start(ctx context.Context) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if et.isActive {
		return fmt.Errorf("enhanced tester is already active")
	}

	et.logger.Info("Starting enhanced testing",
		zap.Bool("masque_enabled", et.config.EnableMASQUE),
		zap.Bool("ice_enabled", et.config.EnableICE),
		zap.Duration("test_duration", et.config.TestDuration))

	et.isActive = true
	et.startTime = time.Now()

	// Инициализируем MASQUE тестирование
	if et.config.EnableMASQUE {
		et.logger.Info("Initializing MASQUE testing")
		et.masqueTester = masque.NewMASQUETester(et.logger, &et.config.MASQUE)

		if err := et.masqueTester.Start(ctx); err != nil {
			et.logger.Error("Failed to start MASQUE testing", zap.Error(err))
			return fmt.Errorf("failed to start MASQUE testing: %v", err)
		}
	}

	// Инициализируем ICE тестирование
	if et.config.EnableICE {
		et.logger.Info("Initializing ICE testing")
		et.iceTester = ice.NewICETester(et.logger, &et.config.ICE)

		if err := et.iceTester.Start(ctx); err != nil {
			et.logger.Error("Failed to start ICE testing", zap.Error(err))
			return fmt.Errorf("failed to start ICE testing: %v", err)
		}
	}

	// Запускаем мониторинг
	go et.monitorTests(ctx)

	et.logger.Info("Enhanced testing started successfully")
	return nil
}

// Stop останавливает расширенное тестирование
func (et *EnhancedTester) Stop() error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if !et.isActive {
		return nil
	}

	et.logger.Info("Stopping enhanced testing")

	// Останавливаем MASQUE тестирование
	if et.masqueTester != nil {
		if err := et.masqueTester.Stop(); err != nil {
			et.logger.Error("Failed to stop MASQUE testing", zap.Error(err))
		}
	}

	// Останавливаем ICE тестирование
	if et.iceTester != nil {
		if err := et.iceTester.Stop(); err != nil {
			et.logger.Error("Failed to stop ICE testing", zap.Error(err))
		}
	}

	et.isActive = false
	et.logger.Info("Enhanced testing stopped")
	return nil
}

// monitorTests мониторит выполнение тестов
func (et *EnhancedTester) monitorTests(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			et.logTestStatus()
		}
	}
}

// logTestStatus логирует статус тестирования
func (et *EnhancedTester) logTestStatus() {
	et.mu.RLock()
	defer et.mu.RUnlock()

	et.logger.Info("Test status",
		zap.Bool("masque_active", et.masqueTester != nil && et.masqueTester.IsActive()),
		zap.Bool("ice_active", et.iceTester != nil && et.iceTester.IsActive()),
		zap.Duration("elapsed", time.Since(et.startTime)))
}

// GetMetrics возвращает объединенные метрики
func (et *EnhancedTester) GetMetrics() *EnhancedMetrics {
	et.mu.RLock()
	defer et.mu.RUnlock()

	metrics := &EnhancedMetrics{
		StartTime:    et.startTime,
		EndTime:      time.Now(),
		TestDuration: time.Since(et.startTime),
	}

	// Получаем MASQUE метрики
	if et.masqueTester != nil {
		metrics.MASQUE = et.masqueTester.GetMetrics()
	}

	// Получаем ICE метрики
	if et.iceTester != nil {
		metrics.ICE = et.iceTester.GetMetrics()
	}

	// Вычисляем общие метрики
	if metrics.MASQUE != nil {
		metrics.TotalTests += metrics.MASQUE.ConnectUDPRequests + metrics.MASQUE.ConnectIPRequests
		metrics.SuccessfulTests += metrics.MASQUE.ConnectUDPSuccesses + metrics.MASQUE.ConnectIPSuccesses
		metrics.FailedTests += metrics.MASQUE.ConnectUDPFailures + metrics.MASQUE.ConnectIPFailures
	}

	if metrics.ICE != nil {
		metrics.TotalTests += metrics.ICE.TotalTests
		metrics.SuccessfulTests += metrics.ICE.SuccessfulTests
		metrics.FailedTests += metrics.ICE.FailedTests
	}

	// Вычисляем процент успеха
	if metrics.TotalTests > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulTests) / float64(metrics.TotalTests) * 100
	}

	return metrics
}

// IsActive возвращает статус активности тестера
func (et *EnhancedTester) IsActive() bool {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.isActive
}

// GetMASQUETester возвращает MASQUE тестер
func (et *EnhancedTester) GetMASQUETester() *masque.MASQUETester {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.masqueTester
}

// GetICETester возвращает ICE тестер
func (et *EnhancedTester) GetICETester() *ice.ICETester {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.iceTester
}
