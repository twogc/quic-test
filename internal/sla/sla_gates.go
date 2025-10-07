package sla

import (
	"time"
)

// SLAGates определяет пороги для автоматических проверок
type SLAGates struct {
	// RTT пороги
	P95RTTMs    float64 // 95-й процентиль RTT в миллисекундах
	MaxRTTMs    float64 // Максимальный RTT в миллисекундах
	MeanRTTMs   float64 // Средний RTT в миллисекундах
	
	// Loss пороги
	LossRatePercent float64 // Процент потерь пакетов
	MaxLossRate     float64 // Максимально допустимый процент потерь
	
	// Throughput пороги
	MinGoodputMbps  float64 // Минимальная goodput в Mbps
	MinThroughputMbps float64 // Минимальная throughput в Mbps
	
	// Congestion Control пороги
	MinBandwidthBps   float64 // Минимальная пропускная способность
	MaxCWNDBytes      int64   // Максимальный congestion window
	MinCWNDBytes      int64   // Минимальный congestion window
	
	// ACK Frequency пороги
	MaxACKDelayMs     float64 // Максимальная задержка ACK
	MinACKFrequency   int     // Минимальная частота ACK
	
	// FEC пороги
	MaxFECRedundancy   float64 // Максимальная избыточность FEC
	MinFECRecoveryRate float64 // Минимальная скорость восстановления FEC
}

// SLAMetrics содержит метрики для проверки SLA
type SLAMetrics struct {
	// RTT метрики
	RTTMinMs      float64
	RTTMaxMs      float64
	RTTMeanMs    float64
	RTTPercentile95Ms float64
	
	// Loss метрики
	LossRatePercent float64
	PacketsLost     int64
	PacketsSent     int64
	
	// Throughput метрики
	GoodputMbps     float64
	ThroughputMbps  float64
	BytesReceived   int64
	BytesSent       int64
	
	// Congestion Control метрики
	BandwidthBps    float64
	CWNDBytes       int64
	PacingRateBps   int64
	CCState         string
	
	// ACK метрики
	ACKDelayMs      float64
	ACKFrequency    int
	
	// FEC метрики
	FECRedundancy   float64
	FECRecoveryRate float64
	
	// Временные метрики
	TestDuration    time.Duration
	StartTime       time.Time
	EndTime         time.Time
}

// SLAResult результат проверки SLA
type SLAResult struct {
	Passed     bool                   `json:"passed"`
	Score      float64                `json:"score"`      // 0.0 - 1.0
	Violations []SLAViolation         `json:"violations"`
	Metrics    SLAMetrics             `json:"metrics"`
	Summary    string                 `json:"summary"`
}

// SLAViolation нарушение SLA
type SLAViolation struct {
	Metric    string  `json:"metric"`
	Expected  float64 `json:"expected"`
	Actual    float64 `json:"actual"`
	Severity  string  `json:"severity"` // "critical", "warning", "info"
	Message   string  `json:"message"`
}

// NewSLAGates создает новые SLA-гейты с разумными значениями по умолчанию
func NewSLAGates() *SLAGates {
	return &SLAGates{
		// RTT пороги (в миллисекундах)
		P95RTTMs:    100.0,  // 95-й процентиль RTT не должен превышать 100ms
		MaxRTTMs:    200.0,  // Максимальный RTT не должен превышать 200ms
		MeanRTTMs:   50.0,   // Средний RTT не должен превышать 50ms
		
		// Loss пороги
		LossRatePercent: 1.0,   // Потери не должны превышать 1%
		MaxLossRate:     5.0,   // Максимально допустимые потери 5%
		
		// Throughput пороги (в Mbps)
		MinGoodputMbps:   10.0,  // Минимальная goodput 10 Mbps
		MinThroughputMbps: 15.0, // Минимальная throughput 15 Mbps
		
		// Congestion Control пороги
		MinBandwidthBps: 1000000, // Минимальная пропускная способность 1 Mbps
		MaxCWNDBytes:    1000000, // Максимальный CWND 1 MB
		MinCWNDBytes:    10000,   // Минимальный CWND 10 KB
		
		// ACK Frequency пороги
		MaxACKDelayMs:   25.0,  // Максимальная задержка ACK 25ms
		MinACKFrequency: 1,     // Минимальная частота ACK 1
		
		// FEC пороги
		MaxFECRedundancy:   0.2,  // Максимальная избыточность FEC 20%
		MinFECRecoveryRate: 0.8,  // Минимальная скорость восстановления FEC 80%
	}
}

// NewSLAGatesStrict создает строгие SLA-гейты для высокопроизводительных систем
func NewSLAGatesStrict() *SLAGates {
	return &SLAGates{
		// Более строгие пороги
		P95RTTMs:    50.0,   // 95-й процентиль RTT не должен превышать 50ms
		MaxRTTMs:    100.0,  // Максимальный RTT не должен превышать 100ms
		MeanRTTMs:   25.0,   // Средний RTT не должен превышать 25ms
		
		LossRatePercent: 0.1,   // Потери не должны превышать 0.1%
		MaxLossRate:     1.0,   // Максимально допустимые потери 1%
		
		MinGoodputMbps:   50.0,  // Минимальная goodput 50 Mbps
		MinThroughputMbps: 60.0, // Минимальная throughput 60 Mbps
		
		MinBandwidthBps: 5000000, // Минимальная пропускная способность 5 Mbps
		MaxCWNDBytes:    2000000, // Максимальный CWND 2 MB
		MinCWNDBytes:    50000,   // Минимальный CWND 50 KB
		
		MaxACKDelayMs:   10.0,  // Максимальная задержка ACK 10ms
		MinACKFrequency: 2,     // Минимальная частота ACK 2
		
		MaxFECRedundancy:   0.1,  // Максимальная избыточность FEC 10%
		MinFECRecoveryRate: 0.9,  // Минимальная скорость восстановления FEC 90%
	}
}

// NewSLAGatesLenient создает мягкие SLA-гейты для тестирования
func NewSLAGatesLenient() *SLAGates {
	return &SLAGates{
		// Более мягкие пороги для тестирования
		P95RTTMs:    500.0,  // 95-й процентиль RTT не должен превышать 500ms
		MaxRTTMs:    1000.0, // Максимальный RTT не должен превышать 1000ms
		MeanRTTMs:   200.0,  // Средний RTT не должен превышать 200ms
		
		LossRatePercent: 5.0,   // Потери не должны превышать 5%
		MaxLossRate:     10.0, // Максимально допустимые потери 10%
		
		MinGoodputMbps:   1.0,   // Минимальная goodput 1 Mbps
		MinThroughputMbps: 2.0,  // Минимальная throughput 2 Mbps
		
		MinBandwidthBps: 100000,  // Минимальная пропускная способность 100 Kbps
		MaxCWNDBytes:    10000000, // Максимальный CWND 10 MB
		MinCWNDBytes:    1000,     // Минимальный CWND 1 KB
		
		MaxACKDelayMs:   100.0, // Максимальная задержка ACK 100ms
		MinACKFrequency: 1,     // Минимальная частота ACK 1
		
		MaxFECRedundancy:   0.5,  // Максимальная избыточность FEC 50%
		MinFECRecoveryRate: 0.5,  // Минимальная скорость восстановления FEC 50%
	}
}
