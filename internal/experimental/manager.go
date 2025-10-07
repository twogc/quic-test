package experimental

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ExperimentalManager управляет экспериментальными QUIC возможностями
type ExperimentalManager struct {
	logger *zap.Logger
	config *ExperimentalConfig
	
	// Компоненты
	ackManager    interface{} // Заглушка для ACKFrequencyManager
	ccManager     *CongestionControlManager
	qlogTracer    *QlogTracer
	multipathMgr  *MultipathManager
	fecManager    *FECManager
	
	// Состояние
	mu       sync.RWMutex
	isActive bool
	server   *ExperimentalServer
	client   *ExperimentalClient
}

// ExperimentalConfig конфигурация экспериментальных возможностей
type ExperimentalConfig struct {
	// Базовые настройки
	Addr        string
	Mode        string
	Connections int
	Streams     int
	Duration    time.Duration
	PacketSize  int
	Rate        int
	
	// Экспериментальные настройки QUIC
	CongestionControl string
	QlogDir          string
	ACKFrequency     int
	MaxACKDelay      time.Duration
	
	// Multipath
	Multipath        []string
	MultipathStrategy string
	
	// FEC
	EnableFEC     bool
	FECRedundancy float64
	
	// Greasing
	EnableGreasing bool
	
	// Производительность
	EnableGSO        bool
	EnableGRO        bool
	SocketBufferSize int
	
	// Наблюдаемость
	EnableTracing    bool
	MetricsInterval  time.Duration
}

// NewExperimentalManager создает новый экспериментальный менеджер
func NewExperimentalManager(logger *zap.Logger, config *ExperimentalConfig) *ExperimentalManager {
	return &ExperimentalManager{
		logger: logger,
		config: config,
	}
}

// Validate проверяет корректность конфигурации
func (cfg *ExperimentalConfig) Validate() error {
	// Проверяем алгоритм управления перегрузкой
	validCC := map[string]bool{
		"cubic":  true,
		"bbr":    true,
		"bbrv2":  true,
		"reno":   true,
	}
	if !validCC[cfg.CongestionControl] {
		return fmt.Errorf("invalid congestion control: %s", cfg.CongestionControl)
	}
	
	// Проверяем стратегию multipath
	validMP := map[string]bool{
		"round-robin": true,
		"lowest-rtt":  true,
		"highest-bw":  true,
	}
	if cfg.MultipathStrategy != "" && !validMP[cfg.MultipathStrategy] {
		return fmt.Errorf("invalid multipath strategy: %s", cfg.MultipathStrategy)
	}
	
	// Проверяем FEC redundancy
	if cfg.FECRedundancy < 0 || cfg.FECRedundancy > 1 {
		return fmt.Errorf("FEC redundancy must be between 0 and 1, got %f", cfg.FECRedundancy)
	}
	
	return nil
}

// Print выводит конфигурацию
func (cfg *ExperimentalConfig) Print() {
	fmt.Printf("🔬 Experimental QUIC Configuration:\n")
	fmt.Printf("  - Mode: %s\n", cfg.Mode)
	fmt.Printf("  - Address: %s\n", cfg.Addr)
	fmt.Printf("  - Connections: %d\n", cfg.Connections)
	fmt.Printf("  - Streams: %d\n", cfg.Streams)
	fmt.Printf("  - Duration: %v\n", cfg.Duration)
	fmt.Printf("  - Packet Size: %d bytes\n", cfg.PacketSize)
	fmt.Printf("  - Rate: %d pps\n", cfg.Rate)
	
	fmt.Printf("\n🚀 Experimental Features:\n")
	fmt.Printf("  - Congestion Control: %s\n", cfg.CongestionControl)
	if cfg.QlogDir != "" {
		fmt.Printf("  - qlog Directory: %s\n", cfg.QlogDir)
	}
	if cfg.ACKFrequency > 0 {
		fmt.Printf("  - ACK Frequency: %d\n", cfg.ACKFrequency)
	}
	fmt.Printf("  - Max ACK Delay: %v\n", cfg.MaxACKDelay)
	
	if len(cfg.Multipath) > 0 {
		fmt.Printf("  - Multipath: %v\n", cfg.Multipath)
		fmt.Printf("  - Multipath Strategy: %s\n", cfg.MultipathStrategy)
	}
	
	if cfg.EnableFEC {
		fmt.Printf("  - FEC: enabled (redundancy: %.1f%%)\n", cfg.FECRedundancy*100)
	}
	
	if cfg.EnableGreasing {
		fmt.Printf("  - QUIC Bit Greasing: enabled\n")
	}
	
	if cfg.EnableGSO {
		fmt.Printf("  - UDP GSO: enabled\n")
	}
	if cfg.EnableGRO {
		fmt.Printf("  - UDP GRO: enabled\n")
	}
	fmt.Printf("  - Socket Buffer: %d bytes\n", cfg.SocketBufferSize)
	
	if cfg.EnableTracing {
		fmt.Printf("  - OpenTelemetry Tracing: enabled\n")
	}
	fmt.Printf("  - Metrics Interval: %v\n", cfg.MetricsInterval)
	fmt.Println()
}

// StartServer запускает экспериментальный сервер
func (em *ExperimentalManager) StartServer(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	if em.isActive {
		return fmt.Errorf("experimental manager is already active")
	}
	
	em.logger.Info("Initializing experimental QUIC server")
	
	// Инициализируем компоненты
	if err := em.initializeComponents(ctx); err != nil {
		return fmt.Errorf("failed to initialize components: %v", err)
	}
	
	// Создаем экспериментальный сервер
	em.server = NewExperimentalServer(em.logger, em.config, em.ackManager, em.ccManager, em.qlogTracer, em.multipathMgr, em.fecManager)
	
	// Запускаем сервер
	if err := em.server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start experimental server: %v", err)
	}
	
	em.isActive = true
	em.logger.Info("Experimental QUIC server started successfully")
	
	return nil
}

// StartClient запускает экспериментальный клиент
func (em *ExperimentalManager) StartClient(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	if em.isActive {
		return fmt.Errorf("experimental manager is already active")
	}
	
	em.logger.Info("Initializing experimental QUIC client")
	
	// Инициализируем компоненты
	if err := em.initializeComponents(ctx); err != nil {
		return fmt.Errorf("failed to initialize components: %v", err)
	}
	
	// Создаем экспериментальный клиент
	em.client = NewExperimentalClient(em.logger, em.config, em.ackManager, em.ccManager, em.qlogTracer, em.multipathMgr, em.fecManager)
	
	// Запускаем клиент
	if err := em.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start experimental client: %v", err)
	}
	
	em.isActive = true
	em.logger.Info("Experimental QUIC client started successfully")
	
	return nil
}

// RunTest запускает экспериментальный тест
func (em *ExperimentalManager) RunTest(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	if em.isActive {
		return fmt.Errorf("experimental manager is already active")
	}
	
	em.logger.Info("Starting experimental QUIC test")
	
	// Инициализируем компоненты
	if err := em.initializeComponents(ctx); err != nil {
		return fmt.Errorf("failed to initialize components: %v", err)
	}
	
	// Создаем сервер и клиент
	em.server = NewExperimentalServer(em.logger, em.config, em.ackManager, em.ccManager, em.qlogTracer, em.multipathMgr, em.fecManager)
	em.client = NewExperimentalClient(em.logger, em.config, em.ackManager, em.ccManager, em.qlogTracer, em.multipathMgr, em.fecManager)
	
	// Запускаем сервер в горутине
	serverCtx, serverCancel := context.WithCancel(ctx)
	defer serverCancel()
	
	go func() {
		if err := em.server.Start(serverCtx); err != nil {
			em.logger.Error("Server error", zap.Error(err))
		}
	}()
	
	// Ждем запуска сервера
	time.Sleep(2 * time.Second)
	
	// Запускаем клиент
	if err := em.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start experimental client: %v", err)
	}
	
	em.isActive = true
	em.logger.Info("Experimental QUIC test started successfully")
	
	// Ждем завершения теста
	select {
	case <-ctx.Done():
		em.logger.Info("Test completed by context cancellation")
	case <-time.After(em.config.Duration):
		em.logger.Info("Test completed by timeout")
	}
	
	// Останавливаем компоненты
	em.stop()
	
	return nil
}

// initializeComponents инициализирует экспериментальные компоненты
func (em *ExperimentalManager) initializeComponents(ctx context.Context) error {
	// ACK Frequency Manager (заглушка)
	em.ackManager = nil
	
	// Congestion Control Manager
	em.ccManager = NewCongestionControlManager(em.logger, em.config.CongestionControl)
	
	// qlog Tracer
	if em.config.QlogDir != "" {
		em.qlogTracer = NewQlogTracer(em.logger, em.config.QlogDir)
	}
	
	// Multipath Manager
	if len(em.config.Multipath) > 0 {
		em.multipathMgr = NewMultipathManager(em.logger, em.config.Multipath, em.config.MultipathStrategy)
	}
	
	// FEC Manager
	if em.config.EnableFEC {
		em.fecManager = NewFECManager(em.logger, em.config.FECRedundancy)
	}
	
	em.logger.Info("Experimental components initialized",
		zap.String("cc", em.config.CongestionControl),
		zap.Bool("qlog", em.config.QlogDir != ""),
		zap.Bool("multipath", len(em.config.Multipath) > 0),
		zap.Bool("fec", em.config.EnableFEC))
	
	return nil
}

// stop останавливает все компоненты
func (em *ExperimentalManager) stop() {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	if !em.isActive {
		return
	}
	
	em.logger.Info("Stopping experimental components")
	
	// Останавливаем сервер
	if em.server != nil {
		em.server.Stop()
	}
	
	// Останавливаем клиент
	if em.client != nil {
		em.client.Stop()
	}
	
	// Останавливаем менеджеры
	if em.ackManager != nil {
		// ACK Manager не требует явной остановки
	}
	
	if em.ccManager != nil {
		em.ccManager.Stop()
	}
	
	if em.qlogTracer != nil {
		em.qlogTracer.Close()
	}
	
	if em.multipathMgr != nil {
		em.multipathMgr.Stop()
	}
	
	if em.fecManager != nil {
		em.fecManager.Stop()
	}
	
	em.isActive = false
	em.logger.Info("Experimental components stopped")
}

// GetMetrics возвращает метрики экспериментальных компонентов
func (em *ExperimentalManager) GetMetrics() map[string]interface{} {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	metrics := make(map[string]interface{})
	
	if em.ackManager != nil {
		// metrics["ack_frequency"] = em.ackManager.GetMetrics()
		metrics["ack_frequency"] = "not implemented"
	}
	
	if em.ccManager != nil {
		metrics["congestion_control"] = em.ccManager.GetMetrics()
	}
	
	if em.multipathMgr != nil {
		metrics["multipath"] = em.multipathMgr.GetMetrics()
	}
	
	if em.fecManager != nil {
		metrics["fec"] = em.fecManager.GetMetrics()
	}
	
	return metrics
}
