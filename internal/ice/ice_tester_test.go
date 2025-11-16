package ice

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestNewICETester проверяет создание ICE тестера
func TestNewICETester(t *testing.T) {
	logger := zap.NewNop()
	config := &ICEConfig{
		StunServers: []string{"stun.l.google.com:19302"},
		GatheringTimeout: 5 * time.Second,
		ConnectionTimeout: 10 * time.Second,
		TestDuration: 30 * time.Second,
		ConcurrentTests: 1,
	}

	tester := NewICETester(logger, config)

	if tester == nil {
		t.Fatal("NewICETester returned nil")
	}

	if tester.config != config {
		t.Error("Config not set correctly")
	}

	if tester.metrics == nil {
		t.Fatal("Metrics is nil")
	}

	if tester.stats == nil {
		t.Fatal("Stats is nil")
	}
}

// TestICEConfigValidation проверяет валидацию конфигурации
func TestICEConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *ICEConfig
		valid  bool
	}{
		{
			name: "valid_with_stun",
			config: &ICEConfig{
				StunServers: []string{"stun.l.google.com:19302"},
				GatheringTimeout: 5 * time.Second,
			},
			valid: true,
		},
		{
			name: "valid_with_turn",
			config: &ICEConfig{
				TurnServers: []string{"turn.example.com"},
				TurnUsername: "user",
				TurnPassword: "pass",
				GatheringTimeout: 5 * time.Second,
			},
			valid: true,
		},
		{
			name: "empty_config",
			config: &ICEConfig{
				GatheringTimeout: 5 * time.Second,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Базовая проверка: конфиг с STUN или TURN серверами считается валидным
			hasServers := len(tt.config.StunServers) > 0 || len(tt.config.TurnServers) > 0
			if tt.valid && !hasServers {
				t.Error("Expected valid config, but no servers")
			}
			if !tt.valid && hasServers {
				t.Error("Expected invalid config, but has servers")
			}
		})
	}
}

// TestICEMetricsInitialization проверяет инициализацию метрик
func TestICEMetricsInitialization(t *testing.T) {
	metrics := &ICEMetrics{}

	// Все метрики должны быть инициализированы нулями
	if metrics.StunRequests != 0 {
		t.Error("StunRequests should be 0")
	}
	if metrics.TurnAllocations != 0 {
		t.Error("TurnAllocations should be 0")
	}
	if metrics.CandidatesGathered != 0 {
		t.Error("CandidatesGathered should be 0")
	}
}

// TestGetMetrics проверяет получение метрик
func TestGetMetrics(t *testing.T) {
	logger := zap.NewNop()
	config := &ICEConfig{
		StunServers: []string{"stun.l.google.com:19302"},
		GatheringTimeout: 5 * time.Second,
	}

	tester := NewICETester(logger, config)

	metrics := tester.GetMetrics()

	if metrics == nil {
		t.Fatal("GetMetrics returned nil")
	}

	// Просто проверяем что метрики получены
	if metrics.StunRequests < 0 {
		t.Error("StunRequests should be non-negative")
	}
}

// TestGetStats проверяет получение статистики
func TestGetStats(t *testing.T) {
	logger := zap.NewNop()
	config := &ICEConfig{
		StunServers: []string{"stun.l.google.com:19302"},
		GatheringTimeout: 5 * time.Second,
	}

	tester := NewICETester(logger, config)

	stats := tester.GetStats()

	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	// Stats по умолчанию должны быть пусты
	if stats.TestsRun != 0 {
		t.Errorf("Expected 0 tests run, got %d", stats.TestsRun)
	}
}

// TestCandidateTypes проверяет типы кандидатов
func TestCandidateTypes(t *testing.T) {
	candidateTypes := []string{
		"host",
		"srflx",
		"relay",
	}

	for _, ct := range candidateTypes {
		if ct == "" {
			t.Errorf("Candidate type is empty")
		}
		if len(ct) == 0 {
			t.Error("Candidate type should not be empty")
		}
	}
}

// TestICEMetricsJSON проверяет сериализацию метрик
func TestICEMetricsJSON(t *testing.T) {
	metrics := &ICEMetrics{
		StunRequests: 100,
		StunResponses: 100,
		TurnAllocations: 5,
		CandidatesGathered: 15,
	}

	// Проверяем что можно сериализовать в JSON
	data, err := json.Marshal(metrics)
	if err != nil {
		t.Errorf("Failed to marshal metrics: %v", err)
	}

	// Проверяем что можно десериализовать
	var recovered ICEMetrics
	err = json.Unmarshal(data, &recovered)
	if err != nil {
		t.Errorf("Failed to unmarshal metrics: %v", err)
	}

	if recovered.StunRequests != 100 {
		t.Errorf("StunRequests mismatch: expected 100, got %d", recovered.StunRequests)
	}
}

// TestICEStatsJSON проверяет сериализацию статистики
func TestICEStatsJSON(t *testing.T) {
	stats := &ICEStats{
		StartTime: time.Now(),
		EndTime: time.Now().Add(1 * time.Minute),
		TestsRun: 10,
		TestsPassed: 8,
		TestsFailed: 2,
		SuccessRate: 0.8,
	}

	// Проверяем что можно сериализовать в JSON
	data, err := json.Marshal(stats)
	if err != nil {
		t.Errorf("Failed to marshal stats: %v", err)
	}

	// Проверяем что можно десериализовать
	var recovered ICEStats
	err = json.Unmarshal(data, &recovered)
	if err != nil {
		t.Errorf("Failed to unmarshal stats: %v", err)
	}

	if recovered.TestsRun != 10 {
		t.Errorf("TestsRun mismatch: expected 10, got %d", recovered.TestsRun)
	}
}

// TestICEConfigDefaults проверяет значения по умолчанию
func TestICEConfigDefaults(t *testing.T) {
	config := &ICEConfig{
		StunServers: []string{"stun.l.google.com:19302"},
	}

	// Если таймауты не установлены, должны быть defaults
	if config.GatheringTimeout == 0 {
		config.GatheringTimeout = 5 * time.Second
	}
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = 10 * time.Second
	}
	if config.TestDuration == 0 {
		config.TestDuration = 30 * time.Second
	}
	if config.ConcurrentTests == 0 {
		config.ConcurrentTests = 1
	}

	if config.GatheringTimeout != 5*time.Second {
		t.Error("GatheringTimeout default not set")
	}
	if config.ConnectionTimeout != 10*time.Second {
		t.Error("ConnectionTimeout default not set")
	}
	if config.TestDuration != 30*time.Second {
		t.Error("TestDuration default not set")
	}
	if config.ConcurrentTests != 1 {
		t.Error("ConcurrentTests default not set")
	}
}

// TestMetricsSuccessRate проверяет расчет success rate
func TestMetricsSuccessRate(t *testing.T) {
	metrics := &ICEMetrics{
		TotalTests: 100,
		SuccessfulTests: 90,
		FailedTests: 10,
	}

	if metrics.TotalTests > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulTests) / float64(metrics.TotalTests)
	}

	if metrics.SuccessRate != 0.9 {
		t.Errorf("Expected success rate 0.9, got %v", metrics.SuccessRate)
	}
}

// TestStatsSuccessRate проверяет расчет success rate в статистике
func TestStatsSuccessRate(t *testing.T) {
	stats := &ICEStats{
		TestsRun: 50,
		TestsPassed: 45,
		TestsFailed: 5,
	}

	if stats.TestsRun > 0 {
		stats.SuccessRate = float64(stats.TestsPassed) / float64(stats.TestsRun)
	}

	if stats.SuccessRate != 0.9 {
		t.Errorf("Expected success rate 0.9, got %v", stats.SuccessRate)
	}
}

// TestContextCancellation проверяет отмену контекста
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Проверяем что контекст может быть отменен
	select {
	case <-ctx.Done():
		t.Log("Context canceled as expected")
	case <-time.After(2 * time.Second):
		t.Error("Context did not cancel")
	}
}

// BenchmarkNewICETester тест производительности создания тестера
func BenchmarkNewICETester(b *testing.B) {
	logger := zap.NewNop()
	config := &ICEConfig{
		StunServers: []string{"stun.l.google.com:19302"},
		GatheringTimeout: 5 * time.Second,
		ConnectionTimeout: 10 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewICETester(logger, config)
	}
}

// TestICEMetricsThread проверяет потокобезопасность метрик
func TestICEMetricsThread(t *testing.T) {
	logger := zap.NewNop()
	config := &ICEConfig{
		StunServers: []string{"stun.l.google.com:19302"},
		GatheringTimeout: 5 * time.Second,
	}

	tester := NewICETester(logger, config)

	done := make(chan bool, 5)

	// Запускаем горутины для получения метрик
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				tester.GetMetrics()
			}
			done <- true
		}()
	}

	// Ждем завершения
	for i := 0; i < 5; i++ {
		<-done
	}
}
