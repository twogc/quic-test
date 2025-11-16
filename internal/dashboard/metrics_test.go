package dashboard

import (
	"testing"
)

func TestMetricsManager_UpdateMetrics(t *testing.T) {
	mm := NewMetricsManager()

	// Проверяем начальное состояние
	mm.mu.RLock()
	if mm.ServerRunning || mm.ClientRunning {
		t.Error("Initial state should be inactive")
	}
	mm.mu.RUnlock()

	// Активируем сервер и клиент
	mm.SetServerRunning(true)
	mm.SetClientRunning(true)

	// Обновляем метрики
	mm.UpdateMetrics()

	// Проверяем, что метрики обновились
	mm.mu.RLock()
	if !mm.ServerRunning || !mm.ClientRunning {
		t.Error("Server and client should be running")
	}
	if mm.Latency <= 0 {
		t.Error("Latency should be positive")
	}
	if mm.Throughput <= 0 {
		t.Error("Throughput should be positive")
	}
	mm.mu.RUnlock()
}

func TestMetricsManager_SetServerRunning(t *testing.T) {
	mm := NewMetricsManager()

	mm.SetServerRunning(true)
	mm.mu.RLock()
	if !mm.ServerRunning {
		t.Error("Server should be running")
	}
	mm.mu.RUnlock()

	mm.SetServerRunning(false)
	mm.mu.RLock()
	if mm.ServerRunning {
		t.Error("Server should not be running")
	}
	mm.mu.RUnlock()
}

func TestMetricsManager_SetClientRunning(t *testing.T) {
	mm := NewMetricsManager()

	mm.SetClientRunning(true)
	mm.mu.RLock()
	if !mm.ClientRunning {
		t.Error("Client should be running")
	}
	mm.mu.RUnlock()

	mm.SetClientRunning(false)
	mm.mu.RLock()
	if mm.ClientRunning {
		t.Error("Client should not be running")
	}
	mm.mu.RUnlock()
}

func TestMetricsManager_SetMASQUEActive(t *testing.T) {
	mm := NewMetricsManager()

	mm.SetMASQUEActive(true)
	mm.mu.RLock()
	if !mm.MASQUEActive {
		t.Error("MASQUE should be active")
	}
	if mm.MASQUETests != 1 {
		t.Errorf("MASQUETests = %d, want 1", mm.MASQUETests)
	}
	mm.mu.RUnlock()

	mm.SetMASQUEActive(false)
	mm.mu.RLock()
	if mm.MASQUEActive {
		t.Error("MASQUE should not be active")
	}
	mm.mu.RUnlock()
}

func TestMetricsManager_SetICEActive(t *testing.T) {
	mm := NewMetricsManager()

	mm.SetICEActive(true)
	mm.mu.RLock()
	if !mm.ICEActive {
		t.Error("ICE should be active")
	}
	if mm.ICETests != 1 {
		t.Errorf("ICETests = %d, want 1", mm.ICETests)
	}
	mm.mu.RUnlock()

	mm.SetICEActive(false)
	mm.mu.RLock()
	if mm.ICEActive {
		t.Error("ICE should not be active")
	}
	mm.mu.RUnlock()
}

func TestMetricsManager_GetMetrics(t *testing.T) {
	mm := NewMetricsManager()
	mm.SetServerRunning(true)
	mm.SetClientRunning(true)
	mm.UpdateMetrics()

	metrics := mm.GetMetrics()

	if metrics["server_running"] != true {
		t.Error("server_running should be true")
	}
	if metrics["client_running"] != true {
		t.Error("client_running should be true")
	}

	// Проверяем, что метрики существуют
	if _, exists := metrics["latency"]; !exists {
		t.Error("latency metric should exist")
	}
	if _, exists := metrics["throughput"]; !exists {
		t.Error("throughput metric should exist")
	}
}

func TestMetricsManager_GetHistory(t *testing.T) {
	mm := NewMetricsManager()

	// Добавляем несколько точек в историю
	mm.mu.Lock()
	mm.LatencyHistory = []float64{10, 20, 30}
	mm.ThroughputHistory = []float64{100, 200, 300}
	mm.mu.Unlock()

	history := mm.GetHistory()

	if len(history) != 3 {
		t.Errorf("History length = %d, want 3", len(history))
	}

	// Проверяем, что история не пустая
	if len(history) == 0 {
		t.Error("History should not be empty")
	}
}

func TestSecureFloat64(t *testing.T) {
	// Тестируем, что функция возвращает значения в диапазоне [0, 1)
	for i := 0; i < 100; i++ {
		val := secureFloat64()
		if val < 0 || val >= 1 {
			t.Errorf("secureFloat64() = %v, want value in range [0, 1)", val)
		}
	}
}

func TestSecureInt63n(t *testing.T) {
	// Тестируем с разными значениями n
	tests := []int64{1, 5, 10, 100}

	for _, n := range tests {
		for i := 0; i < 100; i++ {
			val := secureInt63n(n)
			if val < 0 || val >= n {
				t.Errorf("secureInt63n(%d) = %d, want value in range [0, %d)", n, val, n)
			}
		}
	}

	// Тестируем граничные случаи
	if secureInt63n(0) != 0 {
		t.Error("secureInt63n(0) should return 0")
	}
	if secureInt63n(-1) != 0 {
		t.Error("secureInt63n(-1) should return 0")
	}
}
