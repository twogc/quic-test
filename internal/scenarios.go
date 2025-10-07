package internal

import (
	"fmt"
	"time"
)

// TestScenario описывает тестовый сценарий
type TestScenario struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Config      TestConfig    `json:"config"`
	Expected    ExpectedMetrics `json:"expected"`
}

// ExpectedMetrics описывает ожидаемые метрики для сценария
type ExpectedMetrics struct {
	MinThroughput float64       `json:"min_throughput"` // KB/s
	MaxRTT        time.Duration `json:"max_rtt"`
	MaxLoss       float64       `json:"max_loss"`
	MaxErrors     int64         `json:"max_errors"`
}

// GetScenario возвращает предустановленный сценарий по имени
func GetScenario(name string) (*TestScenario, error) {
	scenarios := map[string]TestScenario{
		"wifi": {
			Name:        "WiFi Network",
			Description: "Стандартная WiFi сеть с умеренными задержками и потерями",
			Config: TestConfig{
				Mode:          "test",
				Addr:          ":9000",
				Connections:   2,
				Streams:       4,
				Duration:      30 * time.Second,
				PacketSize:    1200,
				Rate:          100,
				EmulateLoss:   0.02, // 2%
				EmulateLatency: 10 * time.Millisecond,
				EmulateDup:    0.01, // 1%
			},
			Expected: ExpectedMetrics{
				MinThroughput: 50.0,  // KB/s
				MaxRTT:        50 * time.Millisecond,
				MaxLoss:       0.05,  // 5%
				MaxErrors:     10,
			},
		},
		"lte": {
			Name:        "LTE Network",
			Description: "Мобильная LTE сеть с переменными задержками",
			Config: TestConfig{
				Mode:          "test",
				Addr:          ":9000",
				Connections:   2,
				Streams:       4,
				Duration:      30 * time.Second,
				PacketSize:    1200,
				Rate:          100,
				EmulateLoss:   0.05, // 5%
				EmulateLatency: 30 * time.Millisecond,
				EmulateDup:    0.02, // 2%
			},
			Expected: ExpectedMetrics{
				MinThroughput: 30.0,  // KB/s
				MaxRTT:        100 * time.Millisecond,
				MaxLoss:       0.08,  // 8%
				MaxErrors:     20,
			},
		},
		"sat": {
			Name:        "Satellite Network",
			Description: "Спутниковая связь с высокими задержками",
			Config: TestConfig{
				Mode:          "test",
				Addr:          ":9000",
				Connections:   1,
				Streams:       2,
				Duration:      60 * time.Second,
				PacketSize:    1200,
				Rate:          50,
				EmulateLoss:   0.01, // 1%
				EmulateLatency: 500 * time.Millisecond,
				EmulateDup:    0.005, // 0.5%
			},
			Expected: ExpectedMetrics{
				MinThroughput: 10.0,  // KB/s
				MaxRTT:        1000 * time.Millisecond,
				MaxLoss:       0.02,  // 2%
				MaxErrors:     5,
			},
		},
		"dc-eu": {
			Name:        "Data Center EU",
			Description: "Европейский дата-центр с низкими задержками",
			Config: TestConfig{
				Mode:          "test",
				Addr:          ":9000",
				Connections:   4,
				Streams:       8,
				Duration:      30 * time.Second,
				PacketSize:    1200,
				Rate:          200,
				EmulateLoss:   0.001, // 0.1%
				EmulateLatency: 1 * time.Millisecond,
				EmulateDup:    0.001, // 0.1%
			},
			Expected: ExpectedMetrics{
				MinThroughput: 200.0, // KB/s
				MaxRTT:        10 * time.Millisecond,
				MaxLoss:       0.005, // 0.5%
				MaxErrors:     2,
			},
		},
		"ru-eu": {
			Name:        "Russia to EU",
			Description: "Международное соединение Россия-Европа",
			Config: TestConfig{
				Mode:          "test",
				Addr:          ":9000",
				Connections:   2,
				Streams:       4,
				Duration:      45 * time.Second,
				PacketSize:    1200,
				Rate:          100,
				EmulateLoss:   0.03, // 3%
				EmulateLatency: 80 * time.Millisecond,
				EmulateDup:    0.01, // 1%
			},
			Expected: ExpectedMetrics{
				MinThroughput: 40.0,  // KB/s
				MaxRTT:        150 * time.Millisecond,
				MaxLoss:       0.05,  // 5%
				MaxErrors:     15,
			},
		},
		"loss-burst": {
			Name:        "Loss Burst",
			Description: "Сценарий с периодическими всплесками потерь",
			Config: TestConfig{
				Mode:          "test",
				Addr:          ":9000",
				Connections:   2,
				Streams:       4,
				Duration:      60 * time.Second,
				PacketSize:    1200,
				Rate:          100,
				EmulateLoss:   0.1, // 10% - высокие потери
				EmulateLatency: 20 * time.Millisecond,
				EmulateDup:    0.05, // 5%
			},
			Expected: ExpectedMetrics{
				MinThroughput: 20.0,  // KB/s
				MaxRTT:        200 * time.Millisecond,
				MaxLoss:       0.15, // 15%
				MaxErrors:     50,
			},
		},
		"reorder": {
			Name:        "Packet Reordering",
			Description: "Сценарий с переупорядочиванием пакетов",
			Config: TestConfig{
				Mode:          "test",
				Addr:          ":9000",
				Connections:   2,
				Streams:       4,
				Duration:      30 * time.Second,
				PacketSize:    1200,
				Rate:          100,
				EmulateLoss:   0.02, // 2%
				EmulateLatency: 15 * time.Millisecond,
				EmulateDup:    0.1, // 10% - высокое дублирование
			},
			Expected: ExpectedMetrics{
				MinThroughput: 30.0,  // KB/s
				MaxRTT:        100 * time.Millisecond,
				MaxLoss:       0.05,  // 5%
				MaxErrors:     25,
			},
		},
	}
	
	scenario, exists := scenarios[name]
	if !exists {
		return nil, fmt.Errorf("сценарий '%s' не найден", name)
	}
	
	return &scenario, nil
}

// ListScenarios возвращает список доступных сценариев
func ListScenarios() []string {
	return []string{
		"wifi",
		"lte", 
		"sat",
		"dc-eu",
		"ru-eu",
		"loss-burst",
		"reorder",
	}
}

// PrintScenarioInfo выводит информацию о сценарии
func PrintScenarioInfo(scenario *TestScenario) {
	fmt.Printf("🎯 Test Scenario: %s\n", scenario.Name)
	fmt.Printf("📝 Description: %s\n", scenario.Description)
	fmt.Printf("⚙️  Configuration:\n")
	fmt.Printf("  - Connections: %d\n", scenario.Config.Connections)
	fmt.Printf("  - Streams: %d\n", scenario.Config.Streams)
	fmt.Printf("  - Duration: %v\n", scenario.Config.Duration)
	fmt.Printf("  - Packet Size: %d bytes\n", scenario.Config.PacketSize)
	fmt.Printf("  - Rate: %d packets/s\n", scenario.Config.Rate)
	fmt.Printf("  - Loss: %.2f%%\n", scenario.Config.EmulateLoss*100)
	fmt.Printf("  - Latency: %v\n", scenario.Config.EmulateLatency)
	fmt.Printf("  - Duplication: %.2f%%\n", scenario.Config.EmulateDup*100)
	fmt.Printf("📊 Expected Metrics:\n")
	fmt.Printf("  - Min Throughput: %.1f KB/s\n", scenario.Expected.MinThroughput)
	fmt.Printf("  - Max RTT: %v\n", scenario.Expected.MaxRTT)
	fmt.Printf("  - Max Loss: %.2f%%\n", scenario.Expected.MaxLoss*100)
	fmt.Printf("  - Max Errors: %d\n", scenario.Expected.MaxErrors)
	fmt.Println()
}

// RunScenario запускает тестовый сценарий
func RunScenario(scenarioName string) error {
	scenario, err := GetScenario(scenarioName)
	if err != nil {
		return err
	}
	
	PrintScenarioInfo(scenario)
	
	// Здесь можно добавить логику запуска сценария
	// Например, вызов функции тестирования с конфигурацией сценария
	
	return nil
}

// ValidateScenario проверяет соответствие метрик ожидаемым значениям
func ValidateScenario(scenario *TestScenario, metrics map[string]interface{}) (bool, []string) {
	var violations []string
	
	// Проверяем пропускную способность
	throughput := getFloat64(metrics, "ThroughputAverage")
	if throughput < scenario.Expected.MinThroughput {
		violations = append(violations, fmt.Sprintf("Throughput %.2f KB/s below expected %.2f KB/s", 
			throughput, scenario.Expected.MinThroughput))
	}
	
	// Проверяем RTT
	latencies, _ := metrics["Latencies"].([]float64)
	if len(latencies) > 0 {
		_, p95, _ := calcPercentiles(latencies)
		actualRTT := time.Duration(p95 * float64(time.Millisecond))
		if actualRTT > scenario.Expected.MaxRTT {
			violations = append(violations, fmt.Sprintf("RTT p95 %v exceeds expected %v", 
				actualRTT, scenario.Expected.MaxRTT))
		}
	}
	
	// Проверяем потерю пакетов
	packetLoss := getFloat64(metrics, "PacketLoss")
	if packetLoss > scenario.Expected.MaxLoss {
		violations = append(violations, fmt.Sprintf("Packet loss %.2f%% exceeds expected %.2f%%", 
			packetLoss*100, scenario.Expected.MaxLoss*100))
	}
	
	// Проверяем ошибки
	errors := getInt64(metrics, "Errors")
	if errors > scenario.Expected.MaxErrors {
		violations = append(violations, fmt.Sprintf("Error count %d exceeds expected %d", 
			errors, scenario.Expected.MaxErrors))
	}
	
	return len(violations) == 0, violations
}
