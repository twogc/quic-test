package experimental

import (
	"context"
	"fmt"
	"sync"
	"time"

	"quic-test/internal/quic"

	"go.uber.org/zap"
)

// ExperimentalClient экспериментальный QUIC клиент
type ExperimentalClient struct {
	logger      *zap.Logger
	config      *ExperimentalConfig
	ackManager  interface{}
	ccManager   *CongestionControlManager
	qlogTracer  *QlogTracer
	multipathMgr *MultipathManager
	fecManager  *FECManager
	mu          sync.RWMutex
	isActive    bool
}

// NewExperimentalClient создает новый экспериментальный клиент
func NewExperimentalClient(logger *zap.Logger, config *ExperimentalConfig, ackManager interface{}, ccManager *CongestionControlManager, qlogTracer *QlogTracer, multipathMgr *MultipathManager, fecManager *FECManager) *ExperimentalClient {
	return &ExperimentalClient{
		logger:      logger,
		config:      config,
		ackManager:  ackManager,
		ccManager:   ccManager,
		qlogTracer:  qlogTracer,
		multipathMgr: multipathMgr,
		fecManager:  fecManager,
	}
}

// Start запускает экспериментальный клиент
func (ec *ExperimentalClient) Start(ctx context.Context) error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	if ec.isActive {
		return fmt.Errorf("experimental client is already active")
	}
	
	ec.logger.Info("Starting experimental QUIC client",
		zap.String("addr", ec.config.Addr),
		zap.String("cc", ec.config.CongestionControl))
	
	ec.isActive = true
	
	// Здесь должна быть реальная реализация QUIC клиента
	// Пока просто симулируем работу
	go ec.runClient(ctx)
	
	return nil
}

// Stop останавливает экспериментальный клиент
func (ec *ExperimentalClient) Stop() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	if !ec.isActive {
		return
	}
	
	ec.isActive = false
	ec.logger.Info("Experimental QUIC client stopped")
}

// runClient запускает клиент в горутине
func (ec *ExperimentalClient) runClient(ctx context.Context) {
	ec.logger.Info("Experimental client running")
	
	// Создаем реальный QUIC клиент
	quicConfig := &quic.QUICClientConfig{
		ServerAddr:     ec.config.Addr,
		MaxStreams:     10,
		ConnectTimeout: 10 * time.Second,
		IdleTimeout:    30 * time.Second,
	}
	
	quicClient := quic.NewQUICClient(ec.logger, quicConfig)
	
	// Подключаемся к серверу
	if err := quicClient.Connect(); err != nil {
		ec.logger.Error("Failed to connect to QUIC server", zap.Error(err))
		return
	}
	
	ec.logger.Info("Experimental QUIC client connected successfully")
	
	// Отправляем тестовые данные
	testData := []byte("Hello from experimental QUIC client!")
	if err := quicClient.SendData(testData); err != nil {
		ec.logger.Error("Failed to send data", zap.Error(err))
	}
	
	// Ждем завершения контекста
	<-ctx.Done()
	
	// Отключаемся
	if err := quicClient.Disconnect(); err != nil {
		ec.logger.Error("Failed to disconnect", zap.Error(err))
	}
	
	ec.logger.Info("Experimental client stopped by context")
}
