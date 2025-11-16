package experimental

import (
	"fmt"
	"time"
)

// PredefinedErrorScenarios предустановленные сценарии тестирования ошибок
var PredefinedErrorScenarios = map[string]*ErrorScenario{
	"network-stress": {
		Name:        "Network Stress Test",
		Description: "Тестирование при высоких сетевых нагрузках и потерях пакетов",
		Config: &ErrorTestingConfig{
			Duration:           5 * time.Minute,
			ConcurrentTests:    10,
			NetworkErrors:      true,
			PacketLoss:         0.05, // 5% потерь
			PacketDuplication:  0.02, // 2% дублирования
			PacketReordering:   true,
			PacketCorruption:   0.01, // 1% повреждений
			LatencyVariation:   50 * time.Millisecond,
			JitterVariation:    10 * time.Millisecond,
			ConnectionDrops:    true,
		},
		Duration: 5 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.10, // 10% максимальная ошибка
			MinRecoveryTime:  2 * time.Second,
			MaxLatency:       200 * time.Millisecond,
			MinThroughput:    50.0, // 50 Mbps минимум
			MaxPacketLoss:    0.15, // 15% максимум потерь
		},
	},
	
	"quic-protocol-errors": {
		Name:        "QUIC Protocol Error Test",
		Description: "Тестирование ошибок QUIC протокола",
		Config: &ErrorTestingConfig{
			Duration:         3 * time.Minute,
			ConcurrentTests:  5,
			QUICErrors:       true,
			StreamErrors:     true,
			HandshakeErrors:  true,
			VersionErrors:    true,
			NetworkErrors:    true,
			PacketLoss:       0.02,
			PacketCorruption: 0.005,
		},
		Duration: 3 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.05, // 5% максимальная ошибка
			MinRecoveryTime:  5 * time.Second,
			MaxLatency:       500 * time.Millisecond,
			MinThroughput:    80.0, // 80 Mbps минимум
			MaxPacketLoss:    0.05, // 5% максимум потерь
		},
	},
	
	"experimental-features": {
		Name:        "Experimental Features Error Test",
		Description: "Тестирование ошибок экспериментальных функций",
		Config: &ErrorTestingConfig{
			Duration:           4 * time.Minute,
			ConcurrentTests:    8,
			ACKFrequencyErrors: true,
			CCErrors:           true,
			MultipathErrors:    true,
			FECErrors:          true,
			NetworkErrors:      true,
			PacketLoss:         0.03,
			PacketReordering:   true,
		},
		Duration: 4 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.08, // 8% максимальная ошибка
			MinRecoveryTime:  3 * time.Second,
			MaxLatency:       300 * time.Millisecond,
			MinThroughput:    70.0, // 70 Mbps минимум
			MaxPacketLoss:    0.08, // 8% максимум потерь
		},
	},
	
	"high-latency-network": {
		Name:        "High Latency Network Test",
		Description: "Тестирование в условиях высоких задержек",
		Config: &ErrorTestingConfig{
			Duration:           6 * time.Minute,
			ConcurrentTests:    3,
			NetworkErrors:      true,
			PacketLoss:         0.01,
			LatencyVariation:   200 * time.Millisecond,
			JitterVariation:    50 * time.Millisecond,
			ConnectionDrops:    true,
			QUICErrors:         true,
			HandshakeErrors:    true,
		},
		Duration: 6 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.15, // 15% максимальная ошибка
			MinRecoveryTime:  10 * time.Second,
			MaxLatency:       1000 * time.Millisecond,
			MinThroughput:    30.0, // 30 Mbps минимум
			MaxPacketLoss:    0.10, // 10% максимум потерь
		},
	},
	
	"unstable-connection": {
		Name:        "Unstable Connection Test",
		Description: "Тестирование при нестабильных соединениях",
		Config: &ErrorTestingConfig{
			Duration:           4 * time.Minute,
			ConcurrentTests:    6,
			NetworkErrors:      true,
			PacketLoss:         0.08,
			PacketDuplication:  0.05,
			PacketReordering:   true,
			PacketCorruption:   0.03,
			ConnectionDrops:    true,
			QUICErrors:         true,
			StreamErrors:       true,
			HandshakeErrors:    true,
		},
		Duration: 4 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.20, // 20% максимальная ошибка
			MinRecoveryTime:  8 * time.Second,
			MaxLatency:       800 * time.Millisecond,
			MinThroughput:    40.0, // 40 Mbps минимум
			MaxPacketLoss:    0.20, // 20% максимум потерь
		},
	},
	
	"congestion-control-stress": {
		Name:        "Congestion Control Stress Test",
		Description: "Тестирование алгоритмов управления перегрузкой при стрессовых условиях",
		Config: &ErrorTestingConfig{
			Duration:           7 * time.Minute,
			ConcurrentTests:    12,
			NetworkErrors:      true,
			PacketLoss:         0.10,
			PacketReordering:   true,
			LatencyVariation:   100 * time.Millisecond,
			JitterVariation:    30 * time.Millisecond,
			CCErrors:           true,
			ACKFrequencyErrors: true,
		},
		Duration: 7 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.25, // 25% максимальная ошибка
			MinRecoveryTime:  5 * time.Second,
			MaxLatency:       500 * time.Millisecond,
			MinThroughput:    60.0, // 60 Mbps минимум
			MaxPacketLoss:    0.25, // 25% максимум потерь
		},
	},
	
	"multipath-failure": {
		Name:        "Multipath Failure Test",
		Description: "Тестирование отказов путей в multipath QUIC",
		Config: &ErrorTestingConfig{
			Duration:           5 * time.Minute,
			ConcurrentTests:    4,
			MultipathErrors:    true,
			NetworkErrors:      true,
			PacketLoss:         0.15,
			ConnectionDrops:    true,
			QUICErrors:         true,
			StreamErrors:       true,
		},
		Duration: 5 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.30, // 30% максимальная ошибка
			MinRecoveryTime:  15 * time.Second,
			MaxLatency:       1000 * time.Millisecond,
			MinThroughput:    50.0, // 50 Mbps минимум
			MaxPacketLoss:    0.30, // 30% максимум потерь
		},
	},
	
	"fec-recovery": {
		Name:        "FEC Recovery Test",
		Description: "Тестирование восстановления с помощью FEC",
		Config: &ErrorTestingConfig{
			Duration:           3 * time.Minute,
			ConcurrentTests:    8,
			FECErrors:          true,
			NetworkErrors:      true,
			PacketLoss:         0.20,
			PacketCorruption:   0.05,
			PacketDuplication:  0.03,
		},
		Duration: 3 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.12, // 12% максимальная ошибка
			MinRecoveryTime:  2 * time.Second,
			MaxLatency:       400 * time.Millisecond,
			MinThroughput:    80.0, // 80 Mbps минимум
			MaxPacketLoss:    0.12, // 12% максимум потерь
		},
	},
	
	"extreme-conditions": {
		Name:        "Extreme Conditions Test",
		Description: "Тестирование в экстремальных условиях",
		Config: &ErrorTestingConfig{
			Duration:           10 * time.Minute,
			ConcurrentTests:    15,
			NetworkErrors:      true,
			PacketLoss:         0.30,
			PacketDuplication:  0.10,
			PacketReordering:   true,
			PacketCorruption:   0.08,
			LatencyVariation:   500 * time.Millisecond,
			JitterVariation:    100 * time.Millisecond,
			ConnectionDrops:    true,
			QUICErrors:         true,
			StreamErrors:       true,
			HandshakeErrors:    true,
			VersionErrors:      true,
			ACKFrequencyErrors: true,
			CCErrors:           true,
			MultipathErrors:    true,
			FECErrors:          true,
		},
		Duration: 10 * time.Minute,
		Expected: &ExpectedErrorResults{
			MaxErrorRate:     0.50, // 50% максимальная ошибка
			MinRecoveryTime:  30 * time.Second,
			MaxLatency:       2000 * time.Millisecond,
			MinThroughput:    20.0, // 20 Mbps минимум
			MaxPacketLoss:    0.50, // 50% максимум потерь
		},
	},
}

// GetErrorScenario возвращает сценарий по имени
func GetErrorScenario(name string) (*ErrorScenario, error) {
	if scenario, exists := PredefinedErrorScenarios[name]; exists {
		return scenario, nil
	}
	return nil, fmt.Errorf("error scenario '%s' not found", name)
}

// ListErrorScenarios возвращает список доступных сценариев
func ListErrorScenarios() []string {
	scenarios := make([]string, 0, len(PredefinedErrorScenarios))
	for name := range PredefinedErrorScenarios {
		scenarios = append(scenarios, name)
	}
	return scenarios
}

// CreateCustomErrorScenario создает пользовательский сценарий
func CreateCustomErrorScenario(name, description string, config *ErrorTestingConfig, expected *ExpectedErrorResults) *ErrorScenario {
	return &ErrorScenario{
		Name:        name,
		Description: description,
		Config:      config,
		Duration:    config.Duration,
		Expected:    expected,
	}
}

// ValidateErrorScenario проверяет корректность сценария
func ValidateErrorScenario(scenario *ErrorScenario) error {
	if scenario.Name == "" {
		return fmt.Errorf("scenario name cannot be empty")
	}
	
	if scenario.Duration <= 0 {
		return fmt.Errorf("scenario duration must be positive")
	}
	
	if scenario.Config == nil {
		return fmt.Errorf("scenario config cannot be nil")
	}
	
	if scenario.Config.ConcurrentTests <= 0 {
		return fmt.Errorf("concurrent tests must be positive")
	}
	
	// Проверяем разумность параметров
	if scenario.Config.PacketLoss < 0 || scenario.Config.PacketLoss > 1 {
		return fmt.Errorf("packet loss must be between 0 and 1")
	}
	
	if scenario.Config.PacketDuplication < 0 || scenario.Config.PacketDuplication > 1 {
		return fmt.Errorf("packet duplication must be between 0 and 1")
	}
	
	if scenario.Config.PacketCorruption < 0 || scenario.Config.PacketCorruption > 1 {
		return fmt.Errorf("packet corruption must be between 0 and 1")
	}
	
	return nil
}
