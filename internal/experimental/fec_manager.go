package experimental

import (
	"sync"

	"go.uber.org/zap"
)

// FECManager управляет Forward Error Correction
type FECManager struct {
	logger      *zap.Logger
	redundancy  float64
	mu          sync.RWMutex
	isActive    bool
	metrics     *FECMetrics
}

// FECMetrics метрики FEC
type FECMetrics struct {
	RedundancyBytes   int64   `json:"redundancy_bytes"`
	RecoveryEvents    int64   `json:"recovery_events"`
	FailedRecoveries  int64   `json:"failed_recoveries"`
	Efficiency        float64 `json:"efficiency"`
}

// NewFECManager создает новый FEC менеджер
func NewFECManager(logger *zap.Logger, redundancy float64) *FECManager {
	return &FECManager{
		logger:     logger,
		redundancy: redundancy,
		metrics:    &FECMetrics{},
		isActive:   true,
	}
}

// GetMetrics возвращает метрики FEC
func (fm *FECManager) GetMetrics() *FECMetrics {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.metrics
}

// Stop останавливает FEC менеджер
func (fm *FECManager) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.isActive = false
	fm.logger.Info("FEC manager stopped")
}
