package integration

import (
	"context"
	"fmt"
	"sync"

	"quic-test/internal/experimental"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// FECIntegration интегрирует FEC с quic-go
type FECIntegration struct {
	logger        *zap.Logger
	fecManager    *experimental.FECManager
	mu            sync.RWMutex
	isActive      bool
	connections   map[string]quic.Connection
}

// NewFECIntegration создает новую интеграцию FEC
func NewFECIntegration(logger *zap.Logger, redundancy float64) *FECIntegration {
	fecManager := experimental.NewFECManager(logger, redundancy)
	
	return &FECIntegration{
		logger:      logger,
		fecManager:  fecManager,
		connections: make(map[string]quic.Connection),
		isActive:    true,
	}
}

// Initialize инициализирует интеграцию
func (fi *FECIntegration) Initialize() error {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	
	if fi.fecManager == nil {
		return fmt.Errorf("FEC manager not initialized")
	}
	
	fi.logger.Info("FEC integration initialized",
		zap.Float64("redundancy", fi.fecManager.GetRedundancy()))
	
	return nil
}

// OnConnectionEstablished вызывается при установке соединения
func (fi *FECIntegration) OnConnectionEstablished(conn quic.Connection) {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	
	if !fi.isActive {
		return
	}
	
	connID := conn.RemoteAddr().String()
	fi.connections[connID] = conn
	
	// Регистрируем соединение в менеджере
	fi.fecManager.RegisterConnection(conn)
	
	fi.logger.Info("FEC enabled for connection",
		zap.String("conn_id", connID))
}

// OnConnectionClosed вызывается при закрытии соединения
func (fi *FECIntegration) OnConnectionClosed(conn quic.Connection) {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	
	if !fi.isActive {
		return
	}
	
	connID := conn.RemoteAddr().String()
	delete(fi.connections, connID)
	
	// Удаляем соединение из менеджера
	fi.fecManager.UnregisterConnection(conn)
	
	fi.logger.Info("FEC disabled for connection",
		zap.String("conn_id", connID))
}

// OnDatagramSent вызывается при отправке datagram
func (fi *FECIntegration) OnDatagramSent(conn quic.Connection, data []byte) error {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	if !fi.isActive {
		return nil // Пропускаем FEC обработку
	}
	
	// Обрабатываем datagram через FEC менеджер
	return fi.fecManager.OnDatagramSent(conn, data)
}

// OnDatagramReceived вызывается при получении datagram
func (fi *FECIntegration) OnDatagramReceived(conn quic.Connection, data []byte) ([]byte, error) {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	if !fi.isActive {
		return data, nil // Возвращаем данные как есть
	}
	
	// Обрабатываем datagram через FEC менеджер
	return fi.fecManager.OnDatagramReceived(conn, data)
}

// OnPacketLoss вызывается при обнаружении потерь пакетов
func (fi *FECIntegration) OnPacketLoss(conn quic.Connection, lostPackets []uint64) {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	if !fi.isActive {
		return
	}
	
	// Уведомляем FEC менеджер о потерях
	fi.fecManager.OnPacketLoss(conn, lostPackets)
}

// OnFECPacketReceived вызывается при получении FEC пакета
func (fi *FECIntegration) OnFECPacketReceived(conn quic.Connection, fecData []byte) error {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	if !fi.isActive {
		return nil
	}
	
	// Обрабатываем FEC пакет
	return fi.fecManager.OnFECPacketReceived(conn, fecData)
}

// SetRedundancy устанавливает новый уровень избыточности
func (fi *FECIntegration) SetRedundancy(redundancy float64) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	
	if !fi.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	// Обновляем избыточность в менеджере
	fi.fecManager.SetRedundancy(redundancy)
	
	fi.logger.Info("FEC redundancy updated",
		zap.Float64("new_redundancy", redundancy))
	
	return nil
}

// GetRedundancy возвращает текущий уровень избыточности
func (fi *FECIntegration) GetRedundancy() float64 {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	if !fi.isActive {
		return 0.0 // Нет избыточности по умолчанию
	}
	
	return fi.fecManager.GetRedundancy()
}

// GetRecoveryRate возвращает текущую скорость восстановления
func (fi *FECIntegration) GetRecoveryRate(conn quic.Connection) float64 {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	if !fi.isActive {
		return 0.0
	}
	
	return fi.fecManager.GetRecoveryRate(conn)
}

// GetMetrics возвращает текущие метрики FEC
func (fi *FECIntegration) GetMetrics(conn quic.Connection) *experimental.FECMetrics {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	if !fi.isActive {
		return nil
	}
	
	return fi.fecManager.GetMetrics(conn)
}

// Start запускает интеграцию
func (fi *FECIntegration) Start(ctx context.Context) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	
	if fi.isActive {
		return fmt.Errorf("integration is already active")
	}
	
	fi.isActive = true
	
	fi.logger.Info("FEC integration started")
	
	return nil
}

// Stop останавливает интеграцию
func (fi *FECIntegration) Stop() error {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	
	if !fi.isActive {
		return fmt.Errorf("integration is not active")
	}
	
	fi.isActive = false
	
	fi.logger.Info("FEC integration stopped")
	
	return nil
}

// IsActive проверяет, активна ли интеграция
func (fi *FECIntegration) IsActive() bool {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	
	return fi.isActive
}

