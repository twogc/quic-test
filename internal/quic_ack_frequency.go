package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// ACKFrequencyManager управляет частотой отправки ACK пакетов
// Реализует draft-ietf-quic-ack-frequency
type ACKFrequencyManager struct {
	logger        *zap.Logger
	config        *ACKFrequencyConfig
	mu            sync.RWMutex
	connections   map[string]*ACKFrequencyConnection
	lastUpdate    time.Time
	metrics       *ACKFrequencyMetrics
}

// ACKFrequencyConfig конфигурация ACK Frequency
type ACKFrequencyConfig struct {
	// Базовые настройки
	MaxACKDelay        time.Duration // Максимальная задержка ACK
	MinACKDelay        time.Duration // Минимальная задержка ACK
	DefaultACKDelay    time.Duration // Задержка по умолчанию
	
	// Адаптивные настройки
	EnableAdaptive      bool          // Адаптивная частота ACK
	HighLoadThreshold   int           // Порог высокой нагрузки (пакетов/сек)
	LowLoadThreshold    int           // Порог низкой нагрузки (пакетов/сек)
	
	// Настройки для разных типов трафика
	BulkTransferDelay   time.Duration // Задержка для bulk transfer
	InteractiveDelay    time.Duration // Задержка для interactive трафика
	RealTimeDelay       time.Duration // Задержка для real-time трафика
}

// ACKFrequencyConnection управляет ACK частотой для одного соединения
type ACKFrequencyConnection struct {
	connectionID    string
	lastACKTime    time.Time
	packetCount     int64
	bytesReceived   int64
	ackDelay        time.Duration
	adaptiveMode    bool
	lastUpdate      time.Time
	metrics         *ConnectionACKMetrics
}

// ACKFrequencyMetrics метрики ACK Frequency
type ACKFrequencyMetrics struct {
	TotalConnections     int64         `json:"total_connections"`
	ActiveConnections    int64         `json:"active_connections"`
	TotalACKs            int64         `json:"total_acks"`
	DelayedACKs          int64         `json:"delayed_acks"`
	AdaptiveACKs         int64         `json:"adaptive_acks"`
	AverageACKDelay      time.Duration `json:"average_ack_delay"`
	BandwidthSaved       int64         `json:"bandwidth_saved_bytes"`
}

// ConnectionACKMetrics метрики ACK для соединения
type ConnectionACKMetrics struct {
	ACKsSent         int64         `json:"acks_sent"`
	PacketsReceived  int64         `json:"packets_received"`
	BytesReceived    int64         `json:"bytes_received"`
	AverageDelay     time.Duration `json:"average_delay"`
	LastACKTime      time.Time     `json:"last_ack_time"`
}

// NewACKFrequencyManager создает новый менеджер ACK Frequency
func NewACKFrequencyManager(logger *zap.Logger, config *ACKFrequencyConfig) *ACKFrequencyManager {
	if config == nil {
		config = DefaultACKFrequencyConfig()
	}
	
	return &ACKFrequencyManager{
		logger:      logger,
		config:      config,
		connections: make(map[string]*ACKFrequencyConnection),
		metrics:     &ACKFrequencyMetrics{},
		lastUpdate:  time.Now(),
	}
}

// DefaultACKFrequencyConfig возвращает конфигурацию по умолчанию
func DefaultACKFrequencyConfig() *ACKFrequencyConfig {
	return &ACKFrequencyConfig{
		MaxACKDelay:        25 * time.Millisecond,
		MinACKDelay:        1 * time.Millisecond,
		DefaultACKDelay:    5 * time.Millisecond,
		EnableAdaptive:     true,
		HighLoadThreshold:  1000, // 1000 пакетов/сек
		LowLoadThreshold:   100,  // 100 пакетов/сек
		BulkTransferDelay: 10 * time.Millisecond,
		InteractiveDelay:   2 * time.Millisecond,
		RealTimeDelay:      1 * time.Millisecond,
	}
}

// RegisterConnection регистрирует новое соединение
func (afm *ACKFrequencyManager) RegisterConnection(conn *quic.Conn) string {
	afm.mu.Lock()
	defer afm.mu.Unlock()
	
	connID := fmt.Sprintf("%p", conn)
	afm.connections[connID] = &ACKFrequencyConnection{
		connectionID: connID,
		lastACKTime: time.Now(),
		ackDelay:    afm.config.DefaultACKDelay,
		adaptiveMode: afm.config.EnableAdaptive,
		lastUpdate:  time.Now(),
		metrics:     &ConnectionACKMetrics{},
	}
	
	afm.metrics.TotalConnections++
	afm.metrics.ActiveConnections++
	
	afm.logger.Info("ACK Frequency connection registered",
		zap.String("connection_id", connID),
		zap.Duration("initial_delay", afm.config.DefaultACKDelay))
	
	return connID
}

// UnregisterConnection удаляет соединение
func (afm *ACKFrequencyManager) UnregisterConnection(connID string) {
	afm.mu.Lock()
	defer afm.mu.Unlock()
	
	if conn, exists := afm.connections[connID]; exists {
		afm.logger.Info("ACK Frequency connection unregistered",
			zap.String("connection_id", connID),
			zap.Int64("acks_sent", conn.metrics.ACKsSent),
			zap.Int64("packets_received", conn.metrics.PacketsReceived))
		
		delete(afm.connections, connID)
		afm.metrics.ActiveConnections--
	}
}

// ShouldSendACK определяет, нужно ли отправить ACK
func (afm *ACKFrequencyManager) ShouldSendACK(connID string, packetSize int) bool {
	afm.mu.RLock()
	defer afm.mu.RUnlock()
	
	conn, exists := afm.connections[connID]
	if !exists {
		return true // Если соединение не зарегистрировано, отправляем ACK
	}
	
	now := time.Now()
	
	// Обновляем счетчики
	conn.packetCount++
	conn.bytesReceived += int64(packetSize)
	conn.metrics.PacketsReceived++
	conn.metrics.BytesReceived += int64(packetSize)
	
	// Проверяем, прошло ли достаточно времени
	timeSinceLastACK := now.Sub(conn.lastACKTime)
	
	// Адаптивная настройка задержки
	if conn.adaptiveMode {
		afm.adjustACKDelay(conn, now)
	}
	
	shouldSend := timeSinceLastACK >= conn.ackDelay
	
	if shouldSend {
		conn.lastACKTime = now
		conn.metrics.ACKsSent++
		conn.metrics.LastACKTime = now
		afm.metrics.TotalACKs++
		
		// Обновляем среднюю задержку
		if conn.metrics.ACKsSent > 0 {
			conn.metrics.AverageDelay = time.Duration(
				int64(conn.metrics.AverageDelay)*int64(conn.metrics.ACKsSent-1) + 
				int64(timeSinceLastACK)) / time.Duration(conn.metrics.ACKsSent)
		}
	}
	
	return shouldSend
}

// adjustACKDelay адаптивно настраивает задержку ACK
func (afm *ACKFrequencyManager) adjustACKDelay(conn *ACKFrequencyConnection, now time.Time) {
	// Вычисляем скорость получения пакетов
	timeSinceUpdate := now.Sub(conn.lastUpdate)
	if timeSinceUpdate < time.Second {
		return // Не обновляем слишком часто
	}
	
	packetsPerSecond := float64(conn.packetCount) / timeSinceUpdate.Seconds()
	
	// Адаптивная настройка на основе нагрузки
	if packetsPerSecond > float64(afm.config.HighLoadThreshold) {
		// Высокая нагрузка - увеличиваем задержку
		conn.ackDelay = time.Duration(float64(conn.ackDelay) * 1.2)
		if conn.ackDelay > afm.config.MaxACKDelay {
			conn.ackDelay = afm.config.MaxACKDelay
		}
		afm.metrics.AdaptiveACKs++
	} else if packetsPerSecond < float64(afm.config.LowLoadThreshold) {
		// Низкая нагрузка - уменьшаем задержку
		conn.ackDelay = time.Duration(float64(conn.ackDelay) * 0.8)
		if conn.ackDelay < afm.config.MinACKDelay {
			conn.ackDelay = afm.config.MinACKDelay
		}
		afm.metrics.AdaptiveACKs++
	}
	
	// Сбрасываем счетчики
	conn.packetCount = 0
	conn.lastUpdate = now
}

// SetTrafficType устанавливает тип трафика для оптимизации ACK
func (afm *ACKFrequencyManager) SetTrafficType(connID string, trafficType string) {
	afm.mu.Lock()
	defer afm.mu.Unlock()
	
	conn, exists := afm.connections[connID]
	if !exists {
		return
	}
	
	switch trafficType {
	case "bulk":
		conn.ackDelay = afm.config.BulkTransferDelay
	case "interactive":
		conn.ackDelay = afm.config.InteractiveDelay
	case "realtime":
		conn.ackDelay = afm.config.RealTimeDelay
	default:
		conn.ackDelay = afm.config.DefaultACKDelay
	}
	
	afm.logger.Debug("ACK delay adjusted for traffic type",
		zap.String("connection_id", connID),
		zap.String("traffic_type", trafficType),
		zap.Duration("ack_delay", conn.ackDelay))
}

// GetMetrics возвращает метрики ACK Frequency
func (afm *ACKFrequencyManager) GetMetrics() *ACKFrequencyMetrics {
	afm.mu.RLock()
	defer afm.mu.RUnlock()
	
	// Обновляем среднюю задержку
	totalDelay := time.Duration(0)
	connectionCount := 0
	
	for _, conn := range afm.connections {
		if conn.metrics.ACKsSent > 0 {
			totalDelay += conn.metrics.AverageDelay
			connectionCount++
		}
	}
	
	if connectionCount > 0 {
		afm.metrics.AverageACKDelay = totalDelay / time.Duration(connectionCount)
	}
	
	// Вычисляем сэкономленную полосу пропускания
	// (примерная оценка - каждый отложенный ACK экономит ~64 байта)
	afm.metrics.BandwidthSaved = afm.metrics.DelayedACKs * 64
	
	return afm.metrics
}

// GetConnectionMetrics возвращает метрики для конкретного соединения
func (afm *ACKFrequencyManager) GetConnectionMetrics(connID string) *ConnectionACKMetrics {
	afm.mu.RLock()
	defer afm.mu.RUnlock()
	
	conn, exists := afm.connections[connID]
	if !exists {
		return nil
	}
	
	return conn.metrics
}

// StartMonitoring запускает мониторинг ACK Frequency
func (afm *ACKFrequencyManager) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			afm.logMonitoringStats()
		}
	}
}

// logMonitoringStats выводит статистику мониторинга
func (afm *ACKFrequencyManager) logMonitoringStats() {
	afm.mu.RLock()
	defer afm.mu.RUnlock()
	
	afm.logger.Info("ACK Frequency monitoring stats",
		zap.Int64("active_connections", afm.metrics.ActiveConnections),
		zap.Int64("total_acks", afm.metrics.TotalACKs),
		zap.Int64("adaptive_acks", afm.metrics.AdaptiveACKs),
		zap.Duration("average_delay", afm.metrics.AverageACKDelay),
		zap.Int64("bandwidth_saved_bytes", afm.metrics.BandwidthSaved))
}

