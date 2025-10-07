package experimental

import (
	"sync"

	"go.uber.org/zap"
)

// MultipathManager управляет multipath QUIC соединениями
type MultipathManager struct {
	logger   *zap.Logger
	paths    []string
	strategy string
	mu       sync.RWMutex
	isActive bool
	metrics  *MultipathMetrics
}

// MultipathMetrics метрики multipath
type MultipathMetrics struct {
	ActivePaths     int     `json:"active_paths"`
	TotalPaths      int     `json:"total_paths"`
	BytesPerPath    map[string]int64 `json:"bytes_per_path"`
	SwitchEvents    int64   `json:"switch_events"`
	FailedPaths     int64   `json:"failed_paths"`
	RecoveryEvents  int64   `json:"recovery_events"`
}

// NewMultipathManager создает новый multipath менеджер
func NewMultipathManager(logger *zap.Logger, paths []string, strategy string) *MultipathManager {
	return &MultipathManager{
		logger:   logger,
		paths:    paths,
		strategy: strategy,
		metrics:  &MultipathMetrics{
			TotalPaths: len(paths),
			BytesPerPath: make(map[string]int64),
		},
		isActive: true,
	}
}

// GetMetrics возвращает метрики multipath
func (mm *MultipathManager) GetMetrics() *MultipathMetrics {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.metrics
}

// Stop останавливает multipath менеджер
func (mm *MultipathManager) Stop() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.isActive = false
	mm.logger.Info("Multipath manager stopped")
}
