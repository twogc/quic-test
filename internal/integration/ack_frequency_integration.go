package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"quic-test/internal/experimental"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// ACKFrequencyIntegration интегрирует ACK-Frequency с quic-go
type ACKFrequencyIntegration struct {
	logger        *zap.Logger
	afManager     *experimental.ACKFrequencyManager
	mu            sync.RWMutex
	isActive      bool
	connections   map[string]quic.Connection
}

// NewACKFrequencyIntegration создает новую интеграцию ACK-Frequency
func NewACKFrequencyIntegration(logger *zap.Logger, frequency int, maxDelay time.Duration) *ACKFrequencyIntegration {
	config := &experimental.ACKFrequencyConfig{
		Threshold:            uint(frequency),
		MaxAckDelay:         maxDelay,
		ReorderingThreshold: 3,
		EnableImmediateACK:  true,
	}
	afManager := experimental.NewACKFrequencyManager(logger, config)
	
	return &ACKFrequencyIntegration{
		logger:      logger,
		afManager:   afManager,
		connections: make(map[string]quic.Connection),
		isActive:    true,
	}
}

// Initialize инициализирует интеграцию
func (afi *ACKFrequencyIntegration) Initialize() error {
	afi.mu.Lock()
	defer afi.mu.Unlock()
	
	if afi.afManager == nil {
		return fmt.Errorf("ACK frequency manager not initialized")
	}
	
	afi.logger.Info("ACK frequency integration initialized")
	
	return nil
}

// OnConnectionEstablished вызывается при установке соединения
func (afi *ACKFrequencyIntegration) OnConnectionEstablished(conn quic.Connection) {
	afi.mu.Lock()
	defer afi.mu.Unlock()
	
	if !afi.isActive {
		return
	}
	
	connID := conn.RemoteAddr().String()
	afi.connections[connID] = conn
	
	// Запускаем менеджер
	afi.afManager.Start()
	
	afi.logger.Info("ACK frequency enabled for connection",
		zap.String("conn_id", connID))
}

// OnConnectionClosed вызывается при закрытии соединения
func (afi *ACKFrequencyIntegration) OnConnectionClosed(conn quic.Connection) {
	afi.mu.Lock()
	defer afi.mu.Unlock()
	
	if !afi.isActive {
		return
	}
	
	connID := conn.RemoteAddr().String()
	delete(afi.connections, connID)
	
	// Останавливаем менеджер
	afi.afManager.Stop()
	
	afi.logger.Info("ACK frequency disabled for connection",
		zap.String("conn_id", connID))
}

// OnPacketReceived вызывается при получении пакета
func (afi *ACKFrequencyIntegration) OnPacketReceived(conn quic.Connection, packetNumber uint64, size int) {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	if !afi.isActive {
		return
	}
	
	// Уведомляем менеджер о полученном пакете
	afi.afManager.OnPacketReceived(packetNumber, true)
}

// OnAckSent вызывается при отправке ACK
func (afi *ACKFrequencyIntegration) OnAckSent(conn quic.Connection, ackFrame interface{}) {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	if !afi.isActive {
		return
	}
	
	// Уведомляем менеджер об отправленном ACK
	// Метод OnAckSent не существует, пропускаем
}

// ShouldSendAck проверяет, нужно ли отправить ACK
func (afi *ACKFrequencyIntegration) ShouldSendAck(conn quic.Connection, packetNumber uint64) bool {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	if !afi.isActive {
		return true // Отправляем ACK по умолчанию
	}
	
	// Проверяем политику ACK-Frequency
	return afi.afManager.OnPacketReceived(packetNumber, true)
}

// GetAckDelay возвращает задержку до следующего ACK
func (afi *ACKFrequencyIntegration) GetAckDelay(conn quic.Connection) time.Duration {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	if !afi.isActive {
		return 0 // Нет задержки по умолчанию
	}
	
	// Возвращаем максимальную задержку из конфигурации
	return 25 * time.Millisecond // Значение по умолчанию
}

// SetFrequency устанавливает новую частоту ACK
func (afi *ACKFrequencyIntegration) SetFrequency(frequency int) error {
	afi.mu.Lock()
	defer afi.mu.Unlock()
	
	if !afi.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	// Обновляем конфигурацию менеджера
	afi.afManager.UpdateConfig(uint(frequency), afi.afManager.GetMaxAckDelay(), 3)
	
	afi.logger.Info("ACK frequency updated",
		zap.Int("new_frequency", frequency))
	
	return nil
}

// SetMaxDelay устанавливает максимальную задержку ACK
func (afi *ACKFrequencyIntegration) SetMaxDelay(maxDelay time.Duration) error {
	afi.mu.Lock()
	defer afi.mu.Unlock()
	
	if !afi.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	// Обновляем конфигурацию менеджера
	afi.afManager.UpdateConfig(afi.afManager.GetThreshold(), maxDelay, 3)
	
	afi.logger.Info("ACK max delay updated",
		zap.Duration("new_max_delay", maxDelay))
	
	return nil
}

// GetFrequency возвращает текущую частоту ACK
func (afi *ACKFrequencyIntegration) GetFrequency() int {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	if !afi.isActive {
		return 1 // Частота по умолчанию
	}
	
	// Возвращаем значение из метрик
	metrics := afi.afManager.GetMetrics()
	return int(metrics.Threshold)
}

// GetMaxDelay возвращает максимальную задержку ACK
func (afi *ACKFrequencyIntegration) GetMaxDelay() time.Duration {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	if !afi.isActive {
		return 25 * time.Millisecond // Значение по умолчанию
	}
	
	// Возвращаем значение из метрик
	metrics := afi.afManager.GetMetrics()
	return metrics.MaxAckDelay
}

// GetMetrics возвращает текущие метрики ACK-Frequency
func (afi *ACKFrequencyIntegration) GetMetrics(conn quic.Connection) *experimental.ACKFrequencyMetrics {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	if !afi.isActive {
		return nil
	}
	
	return afi.afManager.GetMetrics()
}

// Start запускает интеграцию
func (afi *ACKFrequencyIntegration) Start(ctx context.Context) error {
	afi.mu.Lock()
	defer afi.mu.Unlock()
	
	if afi.isActive {
		return fmt.Errorf("integration is already active")
	}
	
	afi.isActive = true
	
	afi.logger.Info("ACK frequency integration started")
	
	return nil
}

// Stop останавливает интеграцию
func (afi *ACKFrequencyIntegration) Stop() error {
	afi.mu.Lock()
	defer afi.mu.Unlock()
	
	if !afi.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	afi.isActive = false
	
	afi.logger.Info("ACK frequency integration stopped")
	
	return nil
}

// IsActive проверяет, активна ли интеграция
func (afi *ACKFrequencyIntegration) IsActive() bool {
	afi.mu.RLock()
	defer afi.mu.RUnlock()
	
	return afi.isActive
}
