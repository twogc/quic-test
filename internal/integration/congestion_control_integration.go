package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"quic-test/internal/congestion"
	"quic-test/internal/experimental"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// CongestionControlIntegration интегрирует наши компоненты с quic-go
type CongestionControlIntegration struct {
	logger         *zap.Logger
	ccManager      *experimental.CongestionControlManager
	sendController *congestion.SendController
	mu             sync.RWMutex
	isActive       bool
}

// NewCongestionControlIntegration создает новую интеграцию
func NewCongestionControlIntegration(logger *zap.Logger, algorithm string) *CongestionControlIntegration {
	ccManager := experimental.NewCongestionControlManager(logger, algorithm)
	
	return &CongestionControlIntegration{
		logger:    logger,
		ccManager: ccManager,
		isActive:  true,
	}
}

// Initialize инициализирует интеграцию
func (cci *CongestionControlIntegration) Initialize() error {
	cci.mu.Lock()
	defer cci.mu.Unlock()
	
	if cci.ccManager == nil {
		return fmt.Errorf("congestion control manager not initialized")
	}
	
	// Получаем send controller из менеджера
	cci.sendController = cci.ccManager.GetSendController()
	
	cci.logger.Info("Congestion control integration initialized",
		zap.String("algorithm", cci.ccManager.GetAlgorithm()))
	
	return nil
}

// OnConnectionEstablished вызывается при установке соединения
func (cci *CongestionControlIntegration) OnConnectionEstablished(conn quic.Connection) {
	cci.mu.Lock()
	defer cci.mu.Unlock()
	
	if !cci.isActive {
		return
	}
	
	cci.logger.Info("Connection established, initializing congestion control",
		zap.String("remote_addr", conn.RemoteAddr().String()))
	
	// Инициализируем congestion control для нового соединения
	cci.ccManager.OnConnectionEstablished(conn)
}

// OnPacketSent вызывается при отправке пакета
func (cci *CongestionControlIntegration) OnPacketSent(conn quic.Connection, size int, isAppLimited bool) {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return
	}
	
	// Уведомляем send controller о отправленном пакете
	cci.sendController.OnPacketSent(time.Now(), size, isAppLimited)
	
	// Обновляем метрики
	cci.ccManager.OnPacketSent(conn, size, isAppLimited)
}

// OnAckReceived вызывается при получении ACK
func (cci *CongestionControlIntegration) OnAckReceived(conn quic.Connection, ackedBytes int, rtt time.Duration) {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return
	}
	
	// Уведомляем send controller о полученном ACK
	cci.sendController.OnAck(time.Now(), ackedBytes, rtt)
	
	// Обновляем метрики
	cci.ccManager.OnAckReceived(conn, ackedBytes, rtt)
}

// OnLossDetected вызывается при обнаружении потерь
func (cci *CongestionControlIntegration) OnLossDetected(conn quic.Connection, bytesLost int) {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return
	}
	
	// Уведомляем send controller о потерях
	cci.sendController.OnLoss(bytesLost)
	
	// Обновляем метрики
	cci.ccManager.OnLossDetected(conn, bytesLost)
}

// CanSend проверяет, можно ли отправить пакет
func (cci *CongestionControlIntegration) CanSend(conn quic.Connection, size int) bool {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return true // Разрешаем отправку по умолчанию
	}
	
	// Проверяем pacing и congestion window
	return cci.sendController.CanSend(time.Now(), size)
}

// GetCongestionWindow возвращает текущий congestion window
func (cci *CongestionControlIntegration) GetCongestionWindow(conn quic.Connection) int {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return 1460 * 10 // Значение по умолчанию
	}
	
	return cci.sendController.GetCWND()
}

// GetPacingRate возвращает текущую pacing rate
func (cci *CongestionControlIntegration) GetPacingRate(conn quic.Connection) int64 {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return 1000000 // 1 Mbps по умолчанию
	}
	
	return cci.sendController.GetPacingRate()
}

// GetBandwidth возвращает текущую оценку пропускной способности
func (cci *CongestionControlIntegration) GetBandwidth(conn quic.Connection) float64 {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return 1000000.0 // 1 Mbps по умолчанию
	}
	
	return cci.sendController.GetBandwidth()
}

// GetMinRTT возвращает минимальный RTT
func (cci *CongestionControlIntegration) GetMinRTT(conn quic.Connection) time.Duration {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive || cci.sendController == nil {
		return 10 * time.Millisecond // Значение по умолчанию
	}
	
	return cci.sendController.GetMinRTT()
}

// GetMetrics возвращает текущие метрики
func (cci *CongestionControlIntegration) GetMetrics(conn quic.Connection) *experimental.CCMetrics {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	if !cci.isActive {
		return nil
	}
	
	return cci.ccManager.GetMetrics()
}

// SetAlgorithm устанавливает новый алгоритм congestion control
func (cci *CongestionControlIntegration) SetAlgorithm(algorithm string) error {
	cci.mu.Lock()
	defer cci.mu.Unlock()
	
	if !cci.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	// Создаем новый менеджер с новым алгоритмом
	newCCManager := experimental.NewCongestionControlManager(cci.logger, algorithm)
	
	// Обновляем компоненты
	cci.ccManager = newCCManager
	cci.sendController = cci.ccManager.GetSendController()
	
	cci.logger.Info("Congestion control algorithm changed",
		zap.String("new_algorithm", algorithm))
	
	return nil
}

// Start запускает интеграцию
func (cci *CongestionControlIntegration) Start(ctx context.Context) error {
	cci.mu.Lock()
	defer cci.mu.Unlock()
	
	if cci.isActive {
		return fmt.Errorf("integration is already active")
	}
	
	cci.isActive = true
	
	cci.logger.Info("Congestion control integration started")
	
	return nil
}

// Stop останавливает интеграцию
func (cci *CongestionControlIntegration) Stop() error {
	cci.mu.Lock()
	defer cci.mu.Unlock()
	
	if !cci.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	cci.isActive = false
	
	cci.logger.Info("Congestion control integration stopped")
	
	return nil
}

// IsActive проверяет, активна ли интеграция
func (cci *CongestionControlIntegration) IsActive() bool {
	cci.mu.RLock()
	defer cci.mu.RUnlock()
	
	return cci.isActive
}
