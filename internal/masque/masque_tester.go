package masque

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MASQUETester тестирует MASQUE протокол (RFC 9298, RFC 9484)
type MASQUETester struct {
	logger *zap.Logger
	config *MASQUEConfig

	// Тестируемые компоненты
	connectUDPTester *ConnectUDPTester
	connectIPTester  *ConnectIPTester
	capsuleTester    *CapsuleTester

	// Метрики
	metrics *MASQUEMetrics
	stats   *MASQUEStats

	// Состояние
	mu       sync.RWMutex
	isActive bool
}

// MASQUEConfig конфигурация для MASQUE тестирования
type MASQUEConfig struct {
	// MASQUE сервер для тестирования
	ServerURL string `json:"server_url"`

	// Целевые хосты для CONNECT-UDP
	UDPTargets []string `json:"udp_targets"`

	// Целевые IP для CONNECT-IP
	IPTargets []string `json:"ip_targets"`

	// TLS конфигурация (упрощено для тестирования)
	TLSConfig interface{} `json:"-"`

	// Таймауты
	ConnectTimeout time.Duration `json:"connect_timeout"`
	TestTimeout    time.Duration `json:"test_timeout"`

	// Параметры тестирования
	ConcurrentTests int `json:"concurrent_tests"`
	TestDuration    time.Duration `json:"test_duration"`
}

// MASQUEMetrics метрики MASQUE тестирования
type MASQUEMetrics struct {
	// CONNECT-UDP метрики
	ConnectUDPRequests    int64 `json:"connect_udp_requests"`
	ConnectUDPSuccesses   int64 `json:"connect_udp_successes"`
	ConnectUDPFailures    int64 `json:"connect_udp_failures"`
	ConnectUDPLatency     time.Duration `json:"connect_udp_latency"`

	// CONNECT-IP метрики
	ConnectIPRequests     int64 `json:"connect_ip_requests"`
	ConnectIPSuccesses    int64 `json:"connect_ip_successes"`
	ConnectIPFailures     int64 `json:"connect_ip_failures"`
	ConnectIPLatency      time.Duration `json:"connect_ip_latency"`

	// HTTP Datagrams метрики
	DatagramsSent         int64 `json:"datagrams_sent"`
	DatagramsReceived     int64 `json:"datagrams_received"`
	DatagramLossRate      float64 `json:"datagram_loss_rate"`

	// Capsule метрики
	CapsulesSent          int64 `json:"capsules_sent"`
	CapsulesReceived      int64 `json:"capsules_received"`
	CapsuleFallbackCount  int64 `json:"capsule_fallback_count"`

	// Общие метрики
	TotalConnections      int64 `json:"total_connections"`
	ActiveConnections     int64 `json:"active_connections"`
	FailedConnections     int64 `json:"failed_connections"`
	AverageLatency        time.Duration `json:"average_latency"`
	Throughput            float64 `json:"throughput_mbps"`
}

// MASQUEStats статистика тестирования
type MASQUEStats struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	TestsRun     int `json:"tests_run"`
	TestsPassed  int `json:"tests_passed"`
	TestsFailed  int `json:"tests_failed"`
	SuccessRate  float64 `json:"success_rate"`
}

// NewMASQUETester создает новый MASQUE тестер
func NewMASQUETester(logger *zap.Logger, config *MASQUEConfig) *MASQUETester {
	return &MASQUETester{
		logger: logger,
		config: config,
		metrics: &MASQUEMetrics{},
		stats:   &MASQUEStats{},
	}
}

// Start запускает MASQUE тестирование
func (mt *MASQUETester) Start(ctx context.Context) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.isActive {
		return fmt.Errorf("MASQUE tester is already active")
	}

	mt.logger.Info("Starting MASQUE testing",
		zap.String("server_url", mt.config.ServerURL),
		zap.Int("udp_targets", len(mt.config.UDPTargets)),
		zap.Int("ip_targets", len(mt.config.IPTargets)))

	mt.isActive = true
	mt.stats.StartTime = time.Now()

	// Инициализируем компоненты тестирования
	mt.connectUDPTester = NewConnectUDPTester(mt.logger, mt.config)
	mt.connectIPTester = NewConnectIPTester(mt.logger, mt.config)
	mt.capsuleTester = NewCapsuleTester(mt.logger, mt.config)

	// Запускаем тестирование
	go mt.runTests(ctx)

	return nil
}

// Stop останавливает MASQUE тестирование
func (mt *MASQUETester) Stop() error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if !mt.isActive {
		return nil
	}

	mt.logger.Info("Stopping MASQUE testing")
	mt.isActive = false
	mt.stats.EndTime = time.Now()
	mt.stats.Duration = mt.stats.EndTime.Sub(mt.stats.StartTime)

	// Останавливаем компоненты
	if mt.connectUDPTester != nil {
		mt.connectUDPTester.Stop()
	}
	if mt.connectIPTester != nil {
		mt.connectIPTester.Stop()
	}
	if mt.capsuleTester != nil {
		mt.capsuleTester.Stop()
	}

	return nil
}

// runTests запускает все тесты MASQUE
func (mt *MASQUETester) runTests(ctx context.Context) {
	mt.logger.Info("Running MASQUE tests")

	// CONNECT-UDP тестирование
	if len(mt.config.UDPTargets) > 0 {
		mt.logger.Info("Testing CONNECT-UDP")
		if err := mt.testConnectUDP(ctx); err != nil {
			mt.logger.Error("CONNECT-UDP testing failed", zap.Error(err))
		}
	}

	// CONNECT-IP тестирование
	if len(mt.config.IPTargets) > 0 {
		mt.logger.Info("Testing CONNECT-IP")
		if err := mt.testConnectIP(ctx); err != nil {
			mt.logger.Error("CONNECT-IP testing failed", zap.Error(err))
		}
	}

	// HTTP Datagrams тестирование
	mt.logger.Info("Testing HTTP Datagrams")
	if err := mt.testHTTPDatagrams(ctx); err != nil {
		mt.logger.Error("HTTP Datagrams testing failed", zap.Error(err))
	}

	// Capsule fallback тестирование
	mt.logger.Info("Testing Capsule fallback")
	if err := mt.testCapsuleFallback(ctx); err != nil {
		mt.logger.Error("Capsule fallback testing failed", zap.Error(err))
	}

	// Производительность тестирование
	mt.logger.Info("Testing performance")
	if err := mt.testPerformance(ctx); err != nil {
		mt.logger.Error("Performance testing failed", zap.Error(err))
	}

	mt.logger.Info("MASQUE testing completed")
}

// testConnectUDP тестирует CONNECT-UDP функциональность
func (mt *MASQUETester) testConnectUDP(ctx context.Context) error {
	mt.logger.Info("Starting CONNECT-UDP tests")

	for _, target := range mt.config.UDPTargets {
		mt.logger.Info("Testing CONNECT-UDP to target", zap.String("target", target))
		
		// Создаем CONNECT-UDP соединение
		conn, err := mt.connectUDPTester.Connect(ctx, target)
		if err != nil {
			mt.logger.Error("Failed to connect", zap.String("target", target), zap.Error(err))
			mt.metrics.ConnectUDPFailures++
			continue
		}

		mt.metrics.ConnectUDPSuccesses++

		// Тестируем передачу данных
		if err := mt.testDataTransfer(conn, "CONNECT-UDP", target); err != nil {
			mt.logger.Error("Data transfer test failed", zap.String("target", target), zap.Error(err))
		}

		// Закрываем соединение
		conn.Close()
	}

	return nil
}

// testConnectIP тестирует CONNECT-IP функциональность
func (mt *MASQUETester) testConnectIP(ctx context.Context) error {
	mt.logger.Info("Starting CONNECT-IP tests")

	for _, target := range mt.config.IPTargets {
		mt.logger.Info("Testing CONNECT-IP to target", zap.String("target", target))
		
		// Создаем CONNECT-IP соединение
		conn, err := mt.connectIPTester.Connect(ctx, target)
		if err != nil {
			mt.logger.Error("Failed to connect", zap.String("target", target), zap.Error(err))
			mt.metrics.ConnectIPFailures++
			continue
		}

		mt.metrics.ConnectIPSuccesses++

		// Тестируем передачу данных
		if err := mt.testDataTransfer(conn, "CONNECT-IP", target); err != nil {
			mt.logger.Error("Data transfer test failed", zap.String("target", target), zap.Error(err))
		}

		// Закрываем соединение
		conn.Close()
	}

	return nil
}

// testHTTPDatagrams тестирует HTTP Datagrams
func (mt *MASQUETester) testHTTPDatagrams(ctx context.Context) error {
	mt.logger.Info("Starting HTTP Datagrams tests")

	// Создаем тестовые данные
	testData := []byte("Hello, MASQUE HTTP Datagrams!")
	
	// Тестируем отправку и получение datagrams
	sent, received, err := mt.capsuleTester.TestDatagrams(ctx, testData)
	if err != nil {
		return fmt.Errorf("HTTP Datagrams test failed: %v", err)
	}

	mt.metrics.DatagramsSent += sent
	mt.metrics.DatagramsReceived += received

	// Вычисляем потери
	if sent > 0 {
		mt.metrics.DatagramLossRate = float64(sent-received) / float64(sent) * 100
	}

	mt.logger.Info("HTTP Datagrams test completed",
		zap.Int64("sent", sent),
		zap.Int64("received", received),
		zap.Float64("loss_rate", mt.metrics.DatagramLossRate))

	return nil
}

// testCapsuleFallback тестирует Capsule fallback механизм
func (mt *MASQUETester) testCapsuleFallback(ctx context.Context) error {
	mt.logger.Info("Starting Capsule fallback tests")

	// Тестируем отправку через Capsules
	sent, received, err := mt.capsuleTester.TestCapsules(ctx)
	if err != nil {
		return fmt.Errorf("Capsule fallback test failed: %v", err)
	}

	mt.metrics.CapsulesSent += sent
	mt.metrics.CapsulesReceived += received

	if sent > received {
		mt.metrics.CapsuleFallbackCount++
	}

	mt.logger.Info("Capsule fallback test completed",
		zap.Int64("sent", sent),
		zap.Int64("received", received))

	return nil
}

// testPerformance тестирует производительность MASQUE
func (mt *MASQUETester) testPerformance(ctx context.Context) error {
	mt.logger.Info("Starting performance tests")

	// Тестируем пропускную способность
	throughput, err := mt.capsuleTester.TestThroughput(ctx, mt.config.TestDuration)
	if err != nil {
		return fmt.Errorf("Performance test failed: %v", err)
	}

	mt.metrics.Throughput = throughput

	// Тестируем задержку
	latency, err := mt.capsuleTester.TestLatency(ctx)
	if err != nil {
		return fmt.Errorf("Latency test failed: %v", err)
	}

	mt.metrics.AverageLatency = latency

	mt.logger.Info("Performance tests completed",
		zap.Float64("throughput_mbps", throughput),
		zap.Duration("average_latency", latency))

	return nil
}

// testDataTransfer тестирует передачу данных через соединение
func (mt *MASQUETester) testDataTransfer(conn net.Conn, protocol, target string) error {
	mt.logger.Info("Testing data transfer",
		zap.String("protocol", protocol),
		zap.String("target", target))

	// Тестовые данные
	testData := []byte("Hello, MASQUE!")
	
	// Отправляем данные
	start := time.Now()
	_, err := conn.Write(testData)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	// Читаем ответ
	buffer := make([]byte, len(testData))
	_, err = conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read data: %v", err)
	}

	latency := time.Since(start)
	mt.metrics.AverageLatency = latency

	mt.logger.Info("Data transfer completed",
		zap.String("protocol", protocol),
		zap.String("target", target),
		zap.Duration("latency", latency))

	return nil
}

// GetMetrics возвращает метрики тестирования
func (mt *MASQUETester) GetMetrics() *MASQUEMetrics {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.metrics
}

// GetStats возвращает статистику тестирования
func (mt *MASQUETester) GetStats() *MASQUEStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.stats
}

// IsActive возвращает статус активности тестера
func (mt *MASQUETester) IsActive() bool {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.isActive
}
