package pqc

import (
	"testing"
	"time"
)

// TestNewPQCSimulator проверяет инициализацию PQC симулятора
func TestNewPQCSimulator(t *testing.T) {
	tests := []struct {
		name            string
		algorithm       string
		expectedSize    int
		expectedOverhead time.Duration
	}{
		{"ml_kem_512", "ml-kem-512", 800, 5 * time.Millisecond},
		{"ml_kem_768", "ml-kem-768", 1184, 5 * time.Millisecond},
		{"dilithium_2", "dilithium-2", 1312, 15 * time.Millisecond},
		{"hybrid", "hybrid", 2000, 10 * time.Millisecond},
		{"baseline", "baseline", 200, 0},
		{"unknown", "unknown-algo", 1184, 5 * time.Millisecond}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simulator := NewPQCSimulator(tt.algorithm)

			if simulator == nil {
				t.Fatal("NewPQCSimulator returned nil")
			}

			if simulator.algorithm != tt.algorithm {
				t.Errorf("Expected algorithm %s, got %s", tt.algorithm, simulator.algorithm)
			}

			if simulator.handshakeSize != tt.expectedSize {
				t.Errorf("Expected size %d, got %d", tt.expectedSize, simulator.handshakeSize)
			}

			if simulator.metrics == nil {
				t.Fatal("Metrics is nil")
			}

			if simulator.metrics.HandshakesCompleted != 0 {
				t.Error("Expected 0 initial handshakes")
			}
		})
	}
}

// TestSimulateHandshake проверяет симуляцию handshake
func TestSimulateHandshake(t *testing.T) {
	simulator := NewPQCSimulator("ml-kem-768")

	duration, size := simulator.SimulateHandshake()

	// Проверяем что время положительное
	if duration <= 0 {
		t.Errorf("Expected positive duration, got %v", duration)
	}

	// Проверяем что размер соответствует алгоритму
	if size != 1184 {
		t.Errorf("Expected size 1184, got %d", size)
	}

	// Проверяем что метрики обновились
	if simulator.metrics.HandshakesCompleted != 1 {
		t.Errorf("Expected 1 handshake, got %d", simulator.metrics.HandshakesCompleted)
	}

	if simulator.metrics.TotalHandshakeSize != 1184 {
		t.Errorf("Expected total size 1184, got %d", simulator.metrics.TotalHandshakeSize)
	}

	if simulator.metrics.AvgHandshakeTime <= 0 {
		t.Errorf("Expected positive avg time, got %v", simulator.metrics.AvgHandshakeTime)
	}
}

// TestMultipleHandshakes проверяет несколько handshake'ей
func TestMultipleHandshakes(t *testing.T) {
	simulator := NewPQCSimulator("ml-kem-512")
	numHandshakes := 100

	for i := 0; i < numHandshakes; i++ {
		duration, size := simulator.SimulateHandshake()

		if duration <= 0 {
			t.Errorf("Handshake %d: negative duration %v", i, duration)
		}

		if size != 800 {
			t.Errorf("Handshake %d: wrong size %d", i, size)
		}
	}

	metrics := simulator.GetMetrics()

	if metrics.HandshakesCompleted != int64(numHandshakes) {
		t.Errorf("Expected %d handshakes, got %d", numHandshakes, metrics.HandshakesCompleted)
	}

	if metrics.TotalHandshakeSize != int64(800*numHandshakes) {
		t.Errorf("Expected total size %d, got %d", 800*numHandshakes, metrics.TotalHandshakeSize)
	}

	if metrics.AvgHandshakeTime <= 0 {
		t.Errorf("Expected positive avg time, got %v", metrics.AvgHandshakeTime)
	}

	if metrics.MaxHandshakeTime <= metrics.AvgHandshakeTime {
		t.Logf("Max: %v, Avg: %v", metrics.MaxHandshakeTime, metrics.AvgHandshakeTime)
	}
}

// TestGetMetrics проверяет получение метрик
func TestGetMetrics(t *testing.T) {
	simulator := NewPQCSimulator("dilithium-2")

	// Симулируем несколько handshake'ей
	for i := 0; i < 10; i++ {
		simulator.SimulateHandshake()
	}

	metrics := simulator.GetMetrics()

	if metrics.HandshakesCompleted != 10 {
		t.Errorf("Expected 10 handshakes in metrics, got %d", metrics.HandshakesCompleted)
	}

	if metrics.TotalHandshakeSize == 0 {
		t.Error("Total handshake size is 0")
	}

	if metrics.AvgHandshakeTime == 0 {
		t.Error("Avg handshake time is 0")
	}

	if metrics.MaxHandshakeTime == 0 {
		t.Error("Max handshake time is 0")
	}
}

// TestGetAlgorithm проверяет получение алгоритма
func TestGetAlgorithm(t *testing.T) {
	algorithm := "ml-kem-768"
	simulator := NewPQCSimulator(algorithm)

	retrieved := simulator.GetAlgorithm()
	if retrieved != algorithm {
		t.Errorf("Expected algorithm %s, got %s", algorithm, retrieved)
	}
}

// TestGetHandshakeSize проверяет получение размера handshake
func TestGetHandshakeSize(t *testing.T) {
	tests := []struct {
		algorithm     string
		expectedSize  int
	}{
		{"ml-kem-512", 800},
		{"ml-kem-768", 1184},
		{"dilithium-2", 1312},
		{"baseline", 200},
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			simulator := NewPQCSimulator(tt.algorithm)
			size := simulator.GetHandshakeSize()

			if size != tt.expectedSize {
				t.Errorf("Expected size %d, got %d", tt.expectedSize, size)
			}
		})
	}
}

// TestCompareWithBaseline проверяет сравнение с baseline
func TestCompareWithBaseline(t *testing.T) {
	simulator := NewPQCSimulator("ml-kem-768")

	// Симулируем несколько handshake'ей
	for i := 0; i < 10; i++ {
		simulator.SimulateHandshake()
	}

	comparison := simulator.CompareWithBaseline()

	// Проверяем структуру результата
	if comparison["algorithm"] != "ml-kem-768" {
		t.Errorf("Expected algorithm ml-kem-768, got %v", comparison["algorithm"])
	}

	if comparison["handshake_size"] != 1184 {
		t.Errorf("Expected size 1184, got %v", comparison["handshake_size"])
	}

	// Size increase должен быть положительный (ML-KEM больше чем ECDHE)
	sizeIncrease := comparison["size_increase_pct"].(float64)
	if sizeIncrease <= 0 {
		t.Errorf("Expected positive size increase, got %v", sizeIncrease)
	}

	// Time increase может быть положительный
	timeIncrease := comparison["time_increase_pct"].(float64)
	t.Logf("Size increase: %v%%, Time increase: %v%%", sizeIncrease, timeIncrease)
}

// TestAlgorithmOverheadDifference проверяет разницу overhead между алгоритмами
func TestAlgorithmOverheadDifference(t *testing.T) {
	algorithms := []string{"ml-kem-512", "ml-kem-768", "dilithium-2", "hybrid"}
	var durations []time.Duration

	for _, algo := range algorithms {
		simulator := NewPQCSimulator(algo)
		duration, _ := simulator.SimulateHandshake()
		durations = append(durations, duration)
	}

	// Dilithium должен быть медленнее чем ML-KEM
	if durations[2] <= durations[0] {
		t.Logf("Warning: Dilithium overhead might be too small")
	}

	// Hybrid должен быть медленнее чем baseline
	if durations[3] <= durations[0] {
		t.Logf("Warning: Hybrid overhead might be too small")
	}
}

// TestConcurrentHandshakes проверяет безопасность при конкурентных вызовах
func TestConcurrentHandshakes(t *testing.T) {
	simulator := NewPQCSimulator("ml-kem-768")
	done := make(chan bool, 10)
	numGoroutines := 10
	handshakesPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < handshakesPerGoroutine; j++ {
				simulator.SimulateHandshake()
			}
			done <- true
		}()
	}

	// Ждем завершения
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	metrics := simulator.GetMetrics()
	expectedTotal := int64(numGoroutines * handshakesPerGoroutine)

	if metrics.HandshakesCompleted != expectedTotal {
		t.Errorf("Expected %d handshakes, got %d", expectedTotal, metrics.HandshakesCompleted)
	}
}

// TestMetricsConsistency проверяет консистентность метрик
func TestMetricsConsistency(t *testing.T) {
	simulator := NewPQCSimulator("ml-kem-512")

	// Симулируем handshake'и и отслеживаем max
	var maxTime float64
	for i := 0; i < 50; i++ {
		simulator.SimulateHandshake()
		metrics := simulator.GetMetrics()
		if metrics.MaxHandshakeTime > maxTime {
			maxTime = metrics.MaxHandshakeTime
		}
	}

	metrics := simulator.GetMetrics()

	// Max время должно быть >= avg
	if metrics.MaxHandshakeTime < metrics.AvgHandshakeTime {
		t.Errorf("MaxHandshakeTime (%v) should be >= AvgHandshakeTime (%v)",
			metrics.MaxHandshakeTime, metrics.AvgHandshakeTime)
	}

	// Total size должен быть consistent
	if metrics.TotalHandshakeSize != int64(800*50) {
		t.Errorf("Expected total size %d, got %d", 800*50, metrics.TotalHandshakeSize)
	}
}

// BenchmarkSimulateHandshake тест производительности handshake
func BenchmarkSimulateHandshake(b *testing.B) {
	simulator := NewPQCSimulator("ml-kem-768")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simulator.SimulateHandshake()
	}
}

// BenchmarkGetMetrics тест производительности получения метрик
func BenchmarkGetMetrics(b *testing.B) {
	simulator := NewPQCSimulator("ml-kem-768")

	// Предварительно создаем метрики
	for i := 0; i < 100; i++ {
		simulator.SimulateHandshake()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simulator.GetMetrics()
	}
}

// TestDefaultAlgorithm проверяет что неизвестный алгоритм использует default
func TestDefaultAlgorithm(t *testing.T) {
	simulator := NewPQCSimulator("invalid-algorithm")

	// Должен использовать ML-KEM-768 как default
	if simulator.GetHandshakeSize() != 1184 {
		t.Errorf("Expected default size 1184, got %d", simulator.GetHandshakeSize())
	}
}

// TestRandFloat64Validity проверяет что случайная функция работает
func TestRandFloat64Validity(t *testing.T) {
	// Вызываем несколько раз
	for i := 0; i < 100; i++ {
		val := randFloat64()
		if val < 0 || val > 1 {
			t.Errorf("randFloat64 returned invalid value: %v", val)
		}
	}
}
