package internal

import (
	"testing"
	"time"
)

func TestCheckSLA(t *testing.T) {
	cfg := TestConfig{
		SlaRttP95:     100 * time.Millisecond,
		SlaLoss:       0.01,
		SlaThroughput: 50.0,
		SlaErrors:     10,
	}

	// Тест с проходящими SLA
	metrics := map[string]interface{}{
		"Latencies":        []float64{50.0, 60.0, 70.0, 80.0, 90.0}, // p95 = 90ms < 100ms
		"PacketLoss":       0.005, // 0.5% < 1%
		"ThroughputAverage": 75.0, // 75 KB/s > 50 KB/s
		"Errors":           int64(5), // 5 < 10
	}

	passed, violations, exitCode := CheckSLA(cfg, metrics)
	
	if !passed {
		t.Error("Expected SLA to pass")
	}
	
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}
	
	if exitCode != ExitCodeSuccess {
		t.Errorf("Expected exit code %d, got %d", ExitCodeSuccess, exitCode)
	}
}

func TestCheckSLAFailures(t *testing.T) {
	cfg := TestConfig{
		SlaRttP95:     50 * time.Millisecond,
		SlaLoss:       0.01,
		SlaThroughput: 100.0,
		SlaErrors:     5,
	}

	// Тест с нарушающими SLA
	metrics := map[string]interface{}{
		"Latencies":        []float64{50.0, 60.0, 70.0, 80.0, 90.0, 100.0, 120.0}, // p95 = 120ms > 50ms
		"PacketLoss":       0.02, // 2% > 1%
		"ThroughputAverage": 50.0, // 50 KB/s < 100 KB/s
		"Errors":           int64(10), // 10 > 5
	}

	passed, violations, exitCode := CheckSLA(cfg, metrics)
	
	if passed {
		t.Error("Expected SLA to fail")
	}
	
	if len(violations) == 0 {
		t.Error("Expected violations")
	}
	
	if exitCode != ExitCodeCriticalFailure {
		t.Errorf("Expected exit code %d, got %d", ExitCodeCriticalFailure, exitCode)
	}
}

func TestCheckSLANoConfig(t *testing.T) {
	cfg := TestConfig{} // Нет SLA настроек

	metrics := map[string]interface{}{
		"Latencies":        []float64{50.0, 60.0, 70.0, 80.0, 90.0},
		"PacketLoss":       0.05,
		"ThroughputAverage": 25.0,
		"Errors":           int64(50),
	}

	passed, violations, exitCode := CheckSLA(cfg, metrics)
	
	if !passed {
		t.Error("Expected SLA to pass when no config")
	}
	
	if len(violations) != 0 {
		t.Errorf("Expected no violations when no config, got %d", len(violations))
	}
	
	if exitCode != ExitCodeSuccess {
		t.Errorf("Expected exit code %d, got %d", ExitCodeSuccess, exitCode)
	}
}

func TestSLAViolationTypes(t *testing.T) {
	// Тест RTT нарушения
	cfg := TestConfig{
		SlaRttP95: 50 * time.Millisecond,
	}
	
	metrics := map[string]interface{}{
		"Latencies": []float64{100.0, 120.0, 150.0}, // p95 = 150ms > 50ms
	}
	
	_, violations, _ := CheckSLA(cfg, metrics)
	
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}
	
	if violations[0].Type != ViolationRTT {
		t.Errorf("Expected violation type %s, got %s", ViolationRTT, violations[0].Type)
	}
	
	// Тест потери пакетов
	cfg = TestConfig{
		SlaLoss: 0.01,
	}
	
	metrics = map[string]interface{}{
		"PacketLoss": 0.05, // 5% > 1%
	}
	
	_, violations, _ = CheckSLA(cfg, metrics)
	
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}
	
	if violations[0].Type != ViolationLoss {
		t.Errorf("Expected violation type %s, got %s", ViolationLoss, violations[0].Type)
	}
	
	// Тест пропускной способности
	cfg = TestConfig{
		SlaThroughput: 100.0,
	}
	
	metrics = map[string]interface{}{
		"ThroughputAverage": 50.0, // 50 KB/s < 100 KB/s
	}
	
	_, violations, _ = CheckSLA(cfg, metrics)
	
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}
	
	if violations[0].Type != ViolationThroughput {
		t.Errorf("Expected violation type %s, got %s", ViolationThroughput, violations[0].Type)
	}
	
	// Тест ошибок
	cfg = TestConfig{
		SlaErrors: 5,
	}
	
	metrics = map[string]interface{}{
		"Errors": int64(10), // 10 > 5
	}
	
	_, violations, _ = CheckSLA(cfg, metrics)
	
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}
	
	if violations[0].Type != ViolationErrors {
		t.Errorf("Expected violation type %s, got %s", ViolationErrors, violations[0].Type)
	}
}
