package internal

import (
	"testing"
	"time"
)

func TestGetScenario(t *testing.T) {
	// Тест существующего сценария
	scenario, err := GetScenario("wifi")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if scenario.Name != "WiFi Network" {
		t.Errorf("Expected name 'WiFi Network', got '%s'", scenario.Name)
	}
	
	if scenario.Config.Connections != 2 {
		t.Errorf("Expected connections 2, got %d", scenario.Config.Connections)
	}
	
	if scenario.Config.Streams != 4 {
		t.Errorf("Expected streams 4, got %d", scenario.Config.Streams)
	}
	
	if scenario.Config.EmulateLoss != 0.02 {
		t.Errorf("Expected loss 0.02, got %f", scenario.Config.EmulateLoss)
	}
	
	if scenario.Config.EmulateLatency != 10*time.Millisecond {
		t.Errorf("Expected latency 10ms, got %v", scenario.Config.EmulateLatency)
	}
}

func TestGetScenarioNotFound(t *testing.T) {
	_, err := GetScenario("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent scenario")
	}
}

func TestListScenarios(t *testing.T) {
	scenarios := ListScenarios()
	
	expected := []string{
		"wifi",
		"lte", 
		"sat",
		"dc-eu",
		"ru-eu",
		"loss-burst",
		"reorder",
	}
	
	if len(scenarios) != len(expected) {
		t.Errorf("Expected %d scenarios, got %d", len(expected), len(scenarios))
	}
	
	for _, expectedScenario := range expected {
		found := false
		for _, scenario := range scenarios {
			if scenario == expectedScenario {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected scenario '%s' not found", expectedScenario)
		}
	}
}

func TestValidateScenario(t *testing.T) {
	scenario := &TestScenario{
		Name: "Test Scenario",
		Expected: ExpectedMetrics{
			MinThroughput: 50.0,
			MaxRTT:        100 * time.Millisecond,
			MaxLoss:       0.05,
			MaxErrors:     10,
		},
	}
	
	// Тест с проходящими метриками
	metrics := map[string]interface{}{
		"ThroughputAverage": 75.0,
		"Latencies":         []float64{50.0, 60.0, 70.0, 80.0, 90.0}, // p95 = 90ms < 100ms
		"PacketLoss":        0.03, // 3% < 5%
		"Errors":            int64(5), // 5 < 10
	}
	
	passed, violations := ValidateScenario(scenario, metrics)
	
	if !passed {
		t.Error("Expected scenario to pass")
	}
	
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}
}

func TestValidateScenarioFailures(t *testing.T) {
	scenario := &TestScenario{
		Name: "Test Scenario",
		Expected: ExpectedMetrics{
			MinThroughput: 100.0,
			MaxRTT:        50 * time.Millisecond,
			MaxLoss:       0.01,
			MaxErrors:     5,
		},
	}
	
	// Тест с нарушающими метриками
	metrics := map[string]interface{}{
		"ThroughputAverage": 50.0, // 50 < 100
		"Latencies":         []float64{50.0, 60.0, 70.0, 80.0, 90.0, 100.0, 120.0}, // p95 = 120ms > 50ms
		"PacketLoss":        0.05, // 5% > 1%
		"Errors":            int64(10), // 10 > 5
	}
	
	passed, violations := ValidateScenario(scenario, metrics)
	
	if passed {
		t.Error("Expected scenario to fail")
	}
	
	if len(violations) == 0 {
		t.Error("Expected violations")
	}
	
	// Проверяем, что все нарушения найдены
	expectedViolations := 4
	if len(violations) != expectedViolations {
		t.Errorf("Expected %d violations, got %d", expectedViolations, len(violations))
	}
}

func TestScenarioConfigurations(t *testing.T) {
	// Тест WiFi сценария
	wifi, err := GetScenario("wifi")
	if err != nil {
		t.Fatalf("Failed to get WiFi scenario: %v", err)
	}
	
	if wifi.Config.EmulateLoss != 0.02 {
		t.Errorf("WiFi scenario: expected loss 0.02, got %f", wifi.Config.EmulateLoss)
	}
	
	if wifi.Config.EmulateLatency != 10*time.Millisecond {
		t.Errorf("WiFi scenario: expected latency 10ms, got %v", wifi.Config.EmulateLatency)
	}
	
	// Тест LTE сценария
	lte, err := GetScenario("lte")
	if err != nil {
		t.Fatalf("Failed to get LTE scenario: %v", err)
	}
	
	if lte.Config.EmulateLoss != 0.05 {
		t.Errorf("LTE scenario: expected loss 0.05, got %f", lte.Config.EmulateLoss)
	}
	
	if lte.Config.EmulateLatency != 30*time.Millisecond {
		t.Errorf("LTE scenario: expected latency 30ms, got %v", lte.Config.EmulateLatency)
	}
	
	// Тест спутникового сценария
	sat, err := GetScenario("sat")
	if err != nil {
		t.Fatalf("Failed to get satellite scenario: %v", err)
	}
	
	if sat.Config.EmulateLoss != 0.01 {
		t.Errorf("Satellite scenario: expected loss 0.01, got %f", sat.Config.EmulateLoss)
	}
	
	if sat.Config.EmulateLatency != 500*time.Millisecond {
		t.Errorf("Satellite scenario: expected latency 500ms, got %v", sat.Config.EmulateLatency)
	}
	
	// Тест дата-центра
	dc, err := GetScenario("dc-eu")
	if err != nil {
		t.Fatalf("Failed to get datacenter scenario: %v", err)
	}
	
	if dc.Config.EmulateLoss != 0.001 {
		t.Errorf("Datacenter scenario: expected loss 0.001, got %f", dc.Config.EmulateLoss)
	}
	
	if dc.Config.EmulateLatency != 1*time.Millisecond {
		t.Errorf("Datacenter scenario: expected latency 1ms, got %v", dc.Config.EmulateLatency)
	}
}
