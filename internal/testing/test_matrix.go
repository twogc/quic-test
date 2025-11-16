package testing

import (
	"fmt"
	"time"
)

// TestMatrix определяет матрицу тестов для QUIC
type TestMatrix struct {
	// Параметры нагрузки
	PacketRates []int // pps: 100, 300, 600, 1000
	
	// Параметры соединений
	Connections []int // 1, 2, 4, 8
	
	// Параметры потерь
	LossRates []float64 // 0%, 1%, 3%, 5%
	
	// Параметры RTT
	RTTs []time.Duration // 5ms, 30ms, 100ms
	
	// Дополнительные параметры
	PacketSize    int           // Размер пакета
	TestDuration  time.Duration // Длительность теста
	Streams       int           // Количество стримов на соединение
	
	// Конфигурация тестирования
	WarmupDuration time.Duration // Время разогрева
	CooldownDuration time.Duration // Время остывания
	Iterations     int           // Количество итераций
}

// TestScenario представляет один сценарий тестирования
type TestScenario struct {
	ID           string        `json:"id"`
	PacketRate   int           `json:"packet_rate"`
	Connections  int           `json:"connections"`
	LossRate     float64       `json:"loss_rate"`
	RTT          time.Duration `json:"rtt"`
	PacketSize   int           `json:"packet_size"`
	TestDuration time.Duration `json:"test_duration"`
	Streams      int           `json:"streams"`
	
	// Ожидаемые результаты
	ExpectedGoodputMbps    float64 `json:"expected_goodput_mbps"`
	ExpectedLatencyMs      float64 `json:"expected_latency_ms"`
	ExpectedLossRate       float64 `json:"expected_loss_rate"`
	
	// Описание
	Description string `json:"description"`
}

// TestResult содержит результаты теста
type TestResult struct {
	ScenarioID    string        `json:"scenario_id"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	
	// Метрики производительности
	GoodputMbps      float64 `json:"goodput_mbps"`
	ThroughputMbps   float64 `json:"throughput_mbps"`
	LatencyMinMs     float64 `json:"latency_min_ms"`
	LatencyMaxMs     float64 `json:"latency_max_ms"`
	LatencyMeanMs    float64 `json:"latency_mean_ms"`
	LatencyP95Ms     float64 `json:"latency_p95_ms"`
	LatencyP99Ms     float64 `json:"latency_p99_ms"`
	
	// Метрики потерь
	LossRatePercent  float64 `json:"loss_rate_percent"`
	PacketsSent      int64   `json:"packets_sent"`
	PacketsLost      int64   `json:"packets_lost"`
	PacketsReceived  int64   `json:"packets_received"`
	
	// Метрики congestion control
	BandwidthBps     float64 `json:"bandwidth_bps"`
	CWNDBytes        int64   `json:"cwnd_bytes"`
	PacingRateBps    int64   `json:"pacing_rate_bps"`
	CCState          string  `json:"cc_state"`
	
	// Метрики ACK
	ACKDelayMs       float64 `json:"ack_delay_ms"`
	ACKFrequency     int     `json:"ack_frequency"`
	
	// Метрики FEC
	FECRedundancy    float64 `json:"fec_redundancy"`
	FECRecoveryRate  float64 `json:"fec_recovery_rate"`
	
	// Статус теста
	Passed           bool     `json:"passed"`
	Errors           []string `json:"errors"`
	Warnings         []string `json:"warnings"`
	
	// SLA результаты
	SLAViolations    []string `json:"sla_violations"`
	SLAScore         float64  `json:"sla_score"`
}

// NewTestMatrix создает новую матрицу тестов с разумными значениями по умолчанию
func NewTestMatrix() *TestMatrix {
	return &TestMatrix{
		PacketRates:      []int{100, 300, 600, 1000},
		Connections:      []int{1, 2, 4, 8},
		LossRates:        []float64{0.0, 1.0, 3.0, 5.0},
		RTTs:             []time.Duration{5 * time.Millisecond, 30 * time.Millisecond, 100 * time.Millisecond},
		PacketSize:       1200,
		TestDuration:     30 * time.Second,
		Streams:          1,
		WarmupDuration:   5 * time.Second,
		CooldownDuration: 2 * time.Second,
		Iterations:       3,
	}
}

// NewTestMatrixLight создает легкую матрицу тестов для быстрого тестирования
func NewTestMatrixLight() *TestMatrix {
	return &TestMatrix{
		PacketRates:      []int{100, 300},
		Connections:      []int{1, 2},
		LossRates:        []float64{0.0, 1.0},
		RTTs:             []time.Duration{5 * time.Millisecond, 30 * time.Millisecond},
		PacketSize:       1200,
		TestDuration:     10 * time.Second,
		Streams:          1,
		WarmupDuration:   2 * time.Second,
		CooldownDuration: 1 * time.Second,
		Iterations:       1,
	}
}

// NewTestMatrixHeavy создает тяжелую матрицу тестов для полного тестирования
func NewTestMatrixHeavy() *TestMatrix {
	return &TestMatrix{
		PacketRates:      []int{100, 300, 600, 1000, 2000},
		Connections:      []int{1, 2, 4, 8, 16},
		LossRates:        []float64{0.0, 0.5, 1.0, 2.0, 3.0, 5.0, 10.0},
		RTTs:             []time.Duration{1 * time.Millisecond, 5 * time.Millisecond, 30 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond},
		PacketSize:       1200,
		TestDuration:     60 * time.Second,
		Streams:          4,
		WarmupDuration:   10 * time.Second,
		CooldownDuration: 5 * time.Second,
		Iterations:       5,
	}
}

// GenerateScenarios генерирует все сценарии тестирования
func (tm *TestMatrix) GenerateScenarios() []TestScenario {
	var scenarios []TestScenario
	scenarioID := 1
	
	for _, packetRate := range tm.PacketRates {
		for _, connections := range tm.Connections {
			for _, lossRate := range tm.LossRates {
				for _, rtt := range tm.RTTs {
					scenario := TestScenario{
						ID:           fmt.Sprintf("test_%03d", scenarioID),
						PacketRate:   packetRate,
						Connections:  connections,
						LossRate:     lossRate,
						RTT:          rtt,
						PacketSize:   tm.PacketSize,
						TestDuration: tm.TestDuration,
						Streams:      tm.Streams,
						Description:  fmt.Sprintf("PPS=%d, Conns=%d, Loss=%.1f%%, RTT=%v", 
							packetRate, connections, lossRate, rtt),
					}
					
					// Вычисляем ожидаемые результаты
					scenario.ExpectedGoodputMbps = tm.calculateExpectedGoodput(packetRate, connections, lossRate, rtt)
					scenario.ExpectedLatencyMs = tm.calculateExpectedLatency(rtt)
					scenario.ExpectedLossRate = lossRate
					
					scenarios = append(scenarios, scenario)
					scenarioID++
				}
			}
		}
	}
	
	return scenarios
}

// calculateExpectedGoodput вычисляет ожидаемую goodput
func (tm *TestMatrix) calculateExpectedGoodput(packetRate, connections int, lossRate float64, rtt time.Duration) float64 {
	// Базовая пропускная способность
	baseThroughput := float64(packetRate * connections * tm.PacketSize * 8) / 1e6 // Mbps
	
	// Учитываем потери
	lossFactor := 1.0 - (lossRate / 100.0)
	
	// Учитываем RTT (больший RTT = меньшая эффективность)
	rttFactor := 1.0
	if rtt > 50*time.Millisecond {
		rttFactor = 0.8
	} else if rtt > 100*time.Millisecond {
		rttFactor = 0.6
	}
	
	return baseThroughput * lossFactor * rttFactor
}

// calculateExpectedLatency вычисляет ожидаемую задержку
func (tm *TestMatrix) calculateExpectedLatency(rtt time.Duration) float64 {
	// Базовая задержка = RTT + обработка
	baseLatency := float64(rtt.Milliseconds())
	
	// Добавляем задержку обработки (зависит от нагрузки)
	processingLatency := 1.0 // 1ms базовая обработка
	
	return baseLatency + processingLatency
}

// GetTotalScenarios возвращает общее количество сценариев
func (tm *TestMatrix) GetTotalScenarios() int {
	return len(tm.PacketRates) * len(tm.Connections) * len(tm.LossRates) * len(tm.RTTs)
}

// GetEstimatedDuration возвращает примерное время выполнения всех тестов
func (tm *TestMatrix) GetEstimatedDuration() time.Duration {
	totalScenarios := tm.GetTotalScenarios()
	scenarioDuration := tm.TestDuration + tm.WarmupDuration + tm.CooldownDuration
	totalDuration := time.Duration(totalScenarios) * scenarioDuration * time.Duration(tm.Iterations)
	return totalDuration
}

