package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// IntegrationManager управляет всеми интеграциями с quic-go
type IntegrationManager struct {
	logger *zap.Logger
	
	// Интеграции
	ccIntegration  *CongestionControlIntegration
	afIntegration  *ACKFrequencyIntegration
	fecIntegration *FECIntegration
	
	// Состояние
	mu       sync.RWMutex
	isActive bool
	
	// Конфигурация
	config *IntegrationConfig
}

// IntegrationConfig конфигурация интеграций
type IntegrationConfig struct {
	// Congestion Control
	CCAlgorithm string        `json:"cc_algorithm"`
	CCEnabled   bool          `json:"cc_enabled"`
	
	// ACK Frequency
	AFEnabled   bool          `json:"af_enabled"`
	AFFrequency int           `json:"af_frequency"`
	AFMaxDelay  time.Duration `json:"af_max_delay"`
	
	// FEC
	FECEnabled    bool    `json:"fec_enabled"`
	FECRedundancy float64 `json:"fec_redundancy"`
	
	// Общие настройки
	EnableMetrics bool          `json:"enable_metrics"`
	MetricsInterval time.Duration `json:"metrics_interval"`
}

// NewIntegrationManager создает новый менеджер интеграций
func NewIntegrationManager(logger *zap.Logger, config *IntegrationConfig) *IntegrationManager {
	return &IntegrationManager{
		logger: logger,
		config: config,
	}
}

// Initialize инициализирует все интеграции
func (im *IntegrationManager) Initialize() error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	im.logger.Info("Initializing integration manager")
	
	// Инициализируем congestion control
	if im.config.CCEnabled {
		im.ccIntegration = NewCongestionControlIntegration(im.logger, im.config.CCAlgorithm)
		if err := im.ccIntegration.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize congestion control: %w", err)
		}
		im.logger.Info("Congestion control integration initialized",
			zap.String("algorithm", im.config.CCAlgorithm))
	}
	
	// Инициализируем ACK frequency
	if im.config.AFEnabled {
		im.afIntegration = NewACKFrequencyIntegration(im.logger, im.config.AFFrequency, im.config.AFMaxDelay)
		if err := im.afIntegration.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize ACK frequency: %w", err)
		}
		im.logger.Info("ACK frequency integration initialized",
			zap.Int("frequency", im.config.AFFrequency),
			zap.Duration("max_delay", im.config.AFMaxDelay))
	}
	
	// Инициализируем FEC
	if im.config.FECEnabled {
		im.fecIntegration = NewFECIntegration(im.logger, im.config.FECRedundancy)
		if err := im.fecIntegration.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize FEC: %w", err)
		}
		im.logger.Info("FEC integration initialized",
			zap.Float64("redundancy", im.config.FECRedundancy))
	}
	
	im.logger.Info("Integration manager initialized successfully")
	return nil
}

// Start запускает все интеграции
func (im *IntegrationManager) Start(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if im.isActive {
		return fmt.Errorf("integration manager is already active")
	}
	
	im.logger.Info("Starting integration manager")
	
	// Запускаем congestion control
	if im.ccIntegration != nil {
		if err := im.ccIntegration.Start(ctx); err != nil {
			return fmt.Errorf("failed to start congestion control: %w", err)
		}
	}
	
	// Запускаем ACK frequency
	if im.afIntegration != nil {
		if err := im.afIntegration.Start(ctx); err != nil {
			return fmt.Errorf("failed to start ACK frequency: %w", err)
		}
	}
	
	// Запускаем FEC
	if im.fecIntegration != nil {
		if err := im.fecIntegration.Start(ctx); err != nil {
			return fmt.Errorf("failed to start FEC: %w", err)
		}
	}
	
	im.isActive = true
	im.logger.Info("Integration manager started successfully")
	
	return nil
}

// Stop останавливает все интеграции
func (im *IntegrationManager) Stop() error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if !im.isActive {
		return fmt.Errorf("integration manager is not active")
	}
	
	im.logger.Info("Stopping integration manager")
	
	// Останавливаем все интеграции
	if im.ccIntegration != nil {
		if err := im.ccIntegration.Stop(); err != nil {
			im.logger.Error("Failed to stop congestion control", zap.Error(err))
		}
	}
	
	if im.afIntegration != nil {
		if err := im.afIntegration.Stop(); err != nil {
			im.logger.Error("Failed to stop ACK frequency", zap.Error(err))
		}
	}
	
	if im.fecIntegration != nil {
		if err := im.fecIntegration.Stop(); err != nil {
			im.logger.Error("Failed to stop FEC", zap.Error(err))
		}
	}
	
	im.isActive = false
	im.logger.Info("Integration manager stopped")
	
	return nil
}

// GetCongestionControlIntegration возвращает интеграцию congestion control
func (im *IntegrationManager) GetCongestionControlIntegration() *CongestionControlIntegration {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	return im.ccIntegration
}

// GetACKFrequencyIntegration возвращает интеграцию ACK frequency
func (im *IntegrationManager) GetACKFrequencyIntegration() *ACKFrequencyIntegration {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	return im.afIntegration
}

// GetFECIntegration возвращает интеграцию FEC
func (im *IntegrationManager) GetFECIntegration() *FECIntegration {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	return im.fecIntegration
}

// IsActive проверяет, активен ли менеджер
func (im *IntegrationManager) IsActive() bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	return im.isActive
}

// GetConfig возвращает текущую конфигурацию
func (im *IntegrationManager) GetConfig() *IntegrationConfig {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	return im.config
}

// UpdateConfig обновляет конфигурацию
func (im *IntegrationManager) UpdateConfig(newConfig *IntegrationConfig) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if im.isActive {
		return fmt.Errorf("cannot update config while integration manager is active")
	}
	
	im.config = newConfig
	im.logger.Info("Integration manager config updated")
	
	return nil
}

// GetStatus возвращает статус всех интеграций
func (im *IntegrationManager) GetStatus() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	status := map[string]interface{}{
		"active": im.isActive,
		"config": im.config,
	}
	
	// Статус congestion control
	if im.ccIntegration != nil {
		status["congestion_control"] = map[string]interface{}{
			"active": im.ccIntegration.IsActive(),
		}
	}
	
	// Статус ACK frequency
	if im.afIntegration != nil {
		status["ack_frequency"] = map[string]interface{}{
			"active":    im.afIntegration.IsActive(),
			"frequency": im.afIntegration.GetFrequency(),
			"max_delay": im.afIntegration.GetMaxDelay(),
		}
	}
	
	// Статус FEC
	if im.fecIntegration != nil {
		status["fec"] = map[string]interface{}{
			"active":     im.fecIntegration.IsActive(),
			"redundancy": im.fecIntegration.GetRedundancy(),
		}
	}
	
	return status
}

