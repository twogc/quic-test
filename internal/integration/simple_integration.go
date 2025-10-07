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

// SimpleIntegration простая интеграция компонентов с quic-go
type SimpleIntegration struct {
	logger         *zap.Logger
	sendController *congestion.SendController
	ccManager      *experimental.CongestionControlManager
	mu             sync.RWMutex
	isActive       bool
}

// NewSimpleIntegration создает новую простую интеграцию
func NewSimpleIntegration(logger *zap.Logger, algorithm string) *SimpleIntegration {
	ccManager := experimental.NewCongestionControlManager(logger, algorithm)
	
	return &SimpleIntegration{
		logger:    logger,
		ccManager: ccManager,
		isActive:  true,
	}
}

// Initialize инициализирует интеграцию
func (si *SimpleIntegration) Initialize() error {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if si.ccManager == nil {
		return fmt.Errorf("congestion control manager not initialized")
	}
	
	// Получаем send controller из менеджера
	// Пока что создаем напрямую
	si.sendController = congestion.NewSendController(1460, 32000)
	
	si.logger.Info("Simple integration initialized",
		zap.String("algorithm", si.ccManager.GetAlgorithm()))
	
	return nil
}

// OnConnectionEstablished вызывается при установке соединения
func (si *SimpleIntegration) OnConnectionEstablished(conn quic.Connection) {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if !si.isActive {
		return
	}
	
	si.logger.Info("Connection established",
		zap.String("remote_addr", conn.RemoteAddr().String()))
}

// OnPacketSent вызывается при отправке пакета
func (si *SimpleIntegration) OnPacketSent(conn quic.Connection, size int, isAppLimited bool) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return
	}
	
	// Уведомляем send controller о отправленном пакете
	si.sendController.OnPacketSent(time.Now(), size, isAppLimited)
}

// OnAckReceived вызывается при получении ACK
func (si *SimpleIntegration) OnAckReceived(conn quic.Connection, ackedBytes int, rtt time.Duration) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return
	}
	
	// Уведомляем send controller о полученном ACK
	si.sendController.OnAck(time.Now(), ackedBytes, rtt)
}

// OnLossDetected вызывается при обнаружении потерь
func (si *SimpleIntegration) OnLossDetected(conn quic.Connection, bytesLost int) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return
	}
	
	// Уведомляем send controller о потерях
	si.sendController.OnLoss(bytesLost)
}

// CanSend проверяет, можно ли отправить пакет
func (si *SimpleIntegration) CanSend(conn quic.Connection, size int) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return true // Разрешаем отправку по умолчанию
	}
	
	// Проверяем pacing и congestion window
	return si.sendController.CanSend(time.Now(), size)
}

// GetCongestionWindow возвращает текущий congestion window
func (si *SimpleIntegration) GetCongestionWindow(conn quic.Connection) int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return 1460 * 10 // Значение по умолчанию
	}
	
	return si.sendController.GetCWND()
}

// GetPacingRate возвращает текущую pacing rate
func (si *SimpleIntegration) GetPacingRate(conn quic.Connection) int64 {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return 1000000 // 1 Mbps по умолчанию
	}
	
	return si.sendController.GetPacingRate()
}

// GetBandwidth возвращает текущую оценку пропускной способности
func (si *SimpleIntegration) GetBandwidth(conn quic.Connection) float64 {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return 1000000.0 // 1 Mbps по умолчанию
	}
	
	return si.sendController.GetBandwidth()
}

// GetMinRTT возвращает минимальный RTT
func (si *SimpleIntegration) GetMinRTT(conn quic.Connection) time.Duration {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive || si.sendController == nil {
		return 10 * time.Millisecond // Значение по умолчанию
	}
	
	return si.sendController.GetMinRTT()
}

// GetMetrics возвращает текущие метрики
func (si *SimpleIntegration) GetMetrics(conn quic.Connection) *experimental.CCMetrics {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.isActive {
		return nil
	}
	
	return si.ccManager.GetMetrics()
}

// Start запускает интеграцию
func (si *SimpleIntegration) Start(ctx context.Context) error {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if si.isActive {
		return fmt.Errorf("integration is already active")
	}
	
	si.isActive = true
	
	si.logger.Info("Simple integration started")
	
	return nil
}

// Stop останавливает интеграцию
func (si *SimpleIntegration) Stop() error {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if !si.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	si.isActive = false
	
	si.logger.Info("Simple integration stopped")
	
	return nil
}

// IsActive проверяет, активна ли интеграция
func (si *SimpleIntegration) IsActive() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	return si.isActive
}

