package quic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// QUICManager управляет QUIC сервером и клиентом
type QUICManager struct {
	logger *zap.Logger
	server *QUICServer
	client *QUICClient
	mu     sync.RWMutex
}

// QUICManagerConfig конфигурация менеджера
type QUICManagerConfig struct {
	ServerAddr     string        `json:"server_addr"`
	MaxConnections int           `json:"max_connections"`
	MaxStreams     int           `json:"max_streams"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
}

// NewQUICManager создает новый QUIC менеджер
func NewQUICManager(logger *zap.Logger, config *QUICManagerConfig) *QUICManager {
	// Конфигурация сервера
	serverConfig := &QUICServerConfig{
		Addr:           config.ServerAddr,
		MaxConnections: config.MaxConnections,
		IdleTimeout:    config.IdleTimeout,
		KeepAlive:      10 * time.Second,
	}

	// Конфигурация клиента
	clientConfig := &QUICClientConfig{
		ServerAddr:     "localhost" + config.ServerAddr,
		MaxStreams:     config.MaxStreams,
		ConnectTimeout: config.ConnectTimeout,
		IdleTimeout:    config.IdleTimeout,
	}

	return &QUICManager{
		logger: logger,
		server: NewQUICServer(logger, serverConfig),
		client: NewQUICClient(logger, clientConfig),
	}
}

// StartServer запускает QUIC сервер
func (qm *QUICManager) StartServer() error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if qm.server.IsRunning() {
		return fmt.Errorf("QUIC server is already running")
	}

	qm.logger.Info("Starting QUIC server")
	return qm.server.Start()
}

// StopServer останавливает QUIC сервер
func (qm *QUICManager) StopServer() error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if !qm.server.IsRunning() {
		return fmt.Errorf("QUIC server is not running")
	}

	qm.logger.Info("Stopping QUIC server")
	return qm.server.Stop()
}

// StartClient запускает QUIC клиент
func (qm *QUICManager) StartClient() error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if qm.client.IsConnected() {
		return fmt.Errorf("QUIC client is already connected")
	}

	if !qm.server.IsRunning() {
		return fmt.Errorf("QUIC server is not running")
	}

	qm.logger.Info("Starting QUIC client")
	return qm.client.Connect()
}

// StopClient останавливает QUIC клиент
func (qm *QUICManager) StopClient() error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if !qm.client.IsConnected() {
		return fmt.Errorf("QUIC client is not connected")
	}

	qm.logger.Info("Stopping QUIC client")
	return qm.client.Disconnect()
}

// SendTestData отправляет тестовые данные
func (qm *QUICManager) SendTestData(packetSize int, count int) error {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	if !qm.client.IsConnected() {
		return fmt.Errorf("QUIC client is not connected")
	}

	return qm.client.SendTestData(packetSize, count)
}

// GetStatus возвращает статус сервера и клиента
func (qm *QUICManager) GetStatus() map[string]interface{} {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	return map[string]interface{}{
		"server": map[string]interface{}{
			"running":     qm.server.IsRunning(),
			"connections": qm.server.GetConnections(),
		},
		"client": map[string]interface{}{
			"connected": qm.client.IsConnected(),
			"streams":   qm.client.GetStreams(),
		},
	}
}

// RunTest выполняет тест QUIC соединения
func (qm *QUICManager) RunTest(ctx context.Context, testConfig *TestConfig) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if !qm.server.IsRunning() {
		return fmt.Errorf("QUIC server is not running")
	}

	if !qm.client.IsConnected() {
		return fmt.Errorf("QUIC client is not connected")
	}

	qm.logger.Info("Running QUIC test",
		zap.Int("packet_size", testConfig.PacketSize),
		zap.Int("packet_count", testConfig.PacketCount),
		zap.Duration("duration", testConfig.Duration))

	// Запускаем тест в отдельной горутине
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		startTime := time.Now()
		packetsSent := 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if time.Since(startTime) >= testConfig.Duration {
					qm.logger.Info("QUIC test completed",
						zap.Int("packets_sent", packetsSent),
						zap.Duration("duration", time.Since(startTime)))
					return
				}

				// Отправляем пакеты
				if err := qm.client.SendTestData(testConfig.PacketSize, testConfig.PacketCount); err != nil {
					qm.logger.Error("Failed to send test data", zap.Error(err))
					return
				}

				packetsSent += testConfig.PacketCount
			}
		}
	}()

	return nil
}

// TestConfig конфигурация теста
type TestConfig struct {
	PacketSize  int           `json:"packet_size"`
	PacketCount int           `json:"packet_count"`
	Duration    time.Duration `json:"duration"`
}

// GetServer возвращает сервер
func (qm *QUICManager) GetServer() *QUICServer {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.server
}

// GetClient возвращает клиент
func (qm *QUICManager) GetClient() *QUICClient {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.client
}

