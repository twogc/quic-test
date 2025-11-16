package experimental

import (
	"fmt"
	"sync"
	"time"

	"quic-test/internal/congestion"

	"go.uber.org/zap"
)

// CongestionControlManager управляет алгоритмами управления перегрузкой
type CongestionControlManager struct {
	logger         *zap.Logger
	algorithm      string
	mu             sync.RWMutex
	metrics        *CCMetrics
	isActive       bool
	sendController *congestion.SendController
}

// CCMetrics метрики управления перегрузкой
type CCMetrics struct {
	Algorithm        string        `json:"algorithm"`
	CongestionWindow int64         `json:"congestion_window"`
	SlowStartThreshold int64       `json:"slow_start_threshold"`
	RTT              time.Duration `json:"rtt"`
	RTTVariance      time.Duration `json:"rtt_variance"`
	PacketsInFlight  int64         `json:"packets_in_flight"`
	LossRate         float64       `json:"loss_rate"`
	Throughput       float64       `json:"throughput_mbps"`
	LastUpdate       time.Time     `json:"last_update"`
}

// NewCongestionControlManager создает новый менеджер управления перегрузкой
func NewCongestionControlManager(logger *zap.Logger, algorithm string) *CongestionControlManager {
	// Создаем send controller в зависимости от алгоритма
	var sendController *congestion.SendController
	if algorithm == "bbrv2" || algorithm == "bbrv3" || algorithm == "bbr" {
		sendController = congestion.NewSendController(1460, 32000, algorithm) // MTU и начальный CWND
	}
	
	return &CongestionControlManager{
		logger:         logger,
		algorithm:      algorithm,
		sendController: sendController,
		metrics:        &CCMetrics{
			Algorithm: algorithm,
			LastUpdate: time.Now(),
		},
		isActive: true,
	}
}

// GetAlgorithm возвращает текущий алгоритм
func (ccm *CongestionControlManager) GetAlgorithm() string {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()
	return ccm.algorithm
}

// SetAlgorithm устанавливает новый алгоритм
func (ccm *CongestionControlManager) SetAlgorithm(algorithm string) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()
	
	validAlgorithms := map[string]bool{
		"cubic":  true,
		"bbr":    true,
		"bbrv2":  true,
		"bbrv3":  true,
		"reno":   true,
	}
	
	if !validAlgorithms[algorithm] {
		return fmt.Errorf("invalid congestion control algorithm: %s", algorithm)
	}
	
	ccm.algorithm = algorithm
	ccm.metrics.Algorithm = algorithm
	ccm.metrics.LastUpdate = time.Now()
	
	ccm.logger.Info("Congestion control algorithm changed",
		zap.String("algorithm", algorithm))
	
	return nil
}

// UpdateMetrics обновляет метрики управления перегрузкой
func (ccm *CongestionControlManager) UpdateMetrics(cwnd, ssthresh int64, rtt, rttVar time.Duration, packetsInFlight int64, lossRate, throughput float64) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()
	
	ccm.metrics.CongestionWindow = cwnd
	ccm.metrics.SlowStartThreshold = ssthresh
	ccm.metrics.RTT = rtt
	ccm.metrics.RTTVariance = rttVar
	ccm.metrics.PacketsInFlight = packetsInFlight
	ccm.metrics.LossRate = lossRate
	ccm.metrics.Throughput = throughput
	ccm.metrics.LastUpdate = time.Now()
}

// GetMetrics возвращает текущие метрики
func (ccm *CongestionControlManager) GetMetrics() *CCMetrics {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()
	return ccm.metrics
}

// GetSendController returns the send controller (if available)
func (ccm *CongestionControlManager) GetSendController() *congestion.SendController {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()
	return ccm.sendController
}

// Stop останавливает менеджер
func (ccm *CongestionControlManager) Stop() {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()
	
	ccm.isActive = false
	ccm.logger.Info("Congestion control manager stopped")
}

// IsActive возвращает статус активности
func (ccm *CongestionControlManager) IsActive() bool {
	ccm.mu.RLock()
	defer ccm.mu.RUnlock()
	return ccm.isActive
}
