package internal

import (
	"testing"
	"time"
)

func TestGetNetworkProfile(t *testing.T) {
	// Тест WiFi профиля
	profile, err := GetNetworkProfile("wifi")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if profile.Name != "WiFi 802.11n" {
		t.Errorf("Expected name 'WiFi 802.11n', got '%s'", profile.Name)
	}
	
	if profile.RTT != 20*time.Millisecond {
		t.Errorf("Expected RTT 20ms, got %v", profile.RTT)
	}
	
	if profile.Loss != 0.02 {
		t.Errorf("Expected loss 0.02, got %f", profile.Loss)
	}
	
	if profile.Bandwidth != 1000 {
		t.Errorf("Expected bandwidth 1000 KB/s, got %f", profile.Bandwidth)
	}
}

func TestGetNetworkProfileNotFound(t *testing.T) {
	_, err := GetNetworkProfile("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent profile")
	}
}

func TestListNetworkProfiles(t *testing.T) {
	profiles := ListNetworkProfiles()
	
	expected := []string{
		"wifi",
		"wifi-5g",
		"lte",
		"lte-advanced",
		"5g",
		"satellite",
		"satellite-leo",
		"ethernet",
		"ethernet-10g",
		"dsl",
		"cable",
		"fiber",
		"mobile-3g",
		"edge",
		"international",
		"datacenter",
	}
	
	if len(profiles) != len(expected) {
		t.Errorf("Expected %d profiles, got %d", len(expected), len(profiles))
	}
	
	for _, expectedProfile := range expected {
		found := false
		for _, profile := range profiles {
			if profile == expectedProfile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected profile '%s' not found", expectedProfile)
		}
	}
}

func TestApplyNetworkProfile(t *testing.T) {
	profile := &NetworkProfile{
		Name:        "Test Profile",
		Description: "Test network profile",
		RTT:         50 * time.Millisecond,
		Jitter:      10 * time.Millisecond,
		Loss:        0.05,
		Bandwidth:   2000, // 2 MB/s
		Duplication:  0.02,
		Latency:     25 * time.Millisecond,
	}
	
	cfg := &TestConfig{
		Mode: "test",
		Addr: ":9000",
	}
	
	ApplyNetworkProfile(cfg, profile)
	
	if cfg.EmulateLoss != profile.Loss {
		t.Errorf("Expected loss %f, got %f", profile.Loss, cfg.EmulateLoss)
	}
	
	if cfg.EmulateLatency != profile.Latency {
		t.Errorf("Expected latency %v, got %v", profile.Latency, cfg.EmulateLatency)
	}
	
	if cfg.EmulateDup != profile.Duplication {
		t.Errorf("Expected duplication %f, got %f", profile.Duplication, cfg.EmulateDup)
	}
	
	// Проверяем адаптацию параметров под профиль
	if cfg.Rate != 100 {
		t.Errorf("Expected rate 100 for medium bandwidth, got %d", cfg.Rate)
	}
	
	if cfg.Connections != 2 {
		t.Errorf("Expected connections 2 for medium bandwidth, got %d", cfg.Connections)
	}
	
	if cfg.Streams != 4 {
		t.Errorf("Expected streams 4 for medium bandwidth, got %d", cfg.Streams)
	}
}

func TestApplyNetworkProfileSlowNetwork(t *testing.T) {
	profile := &NetworkProfile{
		Bandwidth: 500, // Медленная сеть
	}
	
	cfg := &TestConfig{
		Mode: "test",
		Addr: ":9000",
	}
	
	ApplyNetworkProfile(cfg, profile)
	
	if cfg.Rate != 50 {
		t.Errorf("Expected rate 50 for slow network, got %d", cfg.Rate)
	}
	
	if cfg.Connections != 1 {
		t.Errorf("Expected connections 1 for slow network, got %d", cfg.Connections)
	}
	
	if cfg.Streams != 2 {
		t.Errorf("Expected streams 2 for slow network, got %d", cfg.Streams)
	}
}

func TestApplyNetworkProfileFastNetwork(t *testing.T) {
	profile := &NetworkProfile{
		Bandwidth: 50000, // Быстрая сеть
	}
	
	cfg := &TestConfig{
		Mode: "test",
		Addr: ":9000",
	}
	
	ApplyNetworkProfile(cfg, profile)
	
	if cfg.Rate != 200 {
		t.Errorf("Expected rate 200 for fast network, got %d", cfg.Rate)
	}
	
	if cfg.Connections != 4 {
		t.Errorf("Expected connections 4 for fast network, got %d", cfg.Connections)
	}
	
	if cfg.Streams != 8 {
		t.Errorf("Expected streams 8 for fast network, got %d", cfg.Streams)
	}
}

func TestApplyNetworkProfilePacketSize(t *testing.T) {
	// Тест высоких задержек
	profile := &NetworkProfile{
		RTT: 200 * time.Millisecond,
	}
	
	cfg := &TestConfig{
		Mode: "test",
		Addr: ":9000",
	}
	
	ApplyNetworkProfile(cfg, profile)
	
	if cfg.PacketSize != 800 {
		t.Errorf("Expected packet size 800 for high latency, got %d", cfg.PacketSize)
	}
	
	// Тест низких задержек
	profile.RTT = 5 * time.Millisecond
	cfg.PacketSize = 1200 // Сброс
	
	ApplyNetworkProfile(cfg, profile)
	
	if cfg.PacketSize != 1400 {
		t.Errorf("Expected packet size 1400 for low latency, got %d", cfg.PacketSize)
	}
}

func TestGetProfileRecommendations(t *testing.T) {
	// Тест высоких задержек
	profile := &NetworkProfile{
		RTT:  200 * time.Millisecond,
		Loss: 0.1,
	}
	
	recommendations := GetProfileRecommendations(profile)
	
	if len(recommendations) == 0 {
		t.Error("Expected recommendations for high latency network")
	}
	
	// Проверяем наличие рекомендаций по BBR
	hasBBR := false
	for _, rec := range recommendations {
		if contains(rec, "BBR") {
			hasBBR = true
			break
		}
	}
	if !hasBBR {
		t.Error("Expected BBR recommendation for high latency")
	}
	
	// Тест быстрой сети
	profile = &NetworkProfile{
		Bandwidth: 100000, // Очень быстрая сеть
	}
	
	recommendations = GetProfileRecommendations(profile)
	
	// Проверяем наличие рекомендаций по 0-RTT
	has0RTT := false
	for _, rec := range recommendations {
		if contains(rec, "0-RTT") {
			has0RTT = true
			break
		}
	}
	if !has0RTT {
		t.Error("Expected 0-RTT recommendation for fast network")
	}
}

func TestNetworkProfileCharacteristics(t *testing.T) {
	// Тест WiFi профиля
	wifi, err := GetNetworkProfile("wifi")
	if err != nil {
		t.Fatalf("Failed to get WiFi profile: %v", err)
	}
	
	if wifi.RTT != 20*time.Millisecond {
		t.Errorf("WiFi: expected RTT 20ms, got %v", wifi.RTT)
	}
	
	if wifi.Loss != 0.02 {
		t.Errorf("WiFi: expected loss 0.02, got %f", wifi.Loss)
	}
	
	if wifi.Bandwidth != 1000 {
		t.Errorf("WiFi: expected bandwidth 1000 KB/s, got %f", wifi.Bandwidth)
	}
	
	// Тест 5G профиля
	fiveg, err := GetNetworkProfile("5g")
	if err != nil {
		t.Fatalf("Failed to get 5G profile: %v", err)
	}
	
	if fiveg.RTT != 5*time.Millisecond {
		t.Errorf("5G: expected RTT 5ms, got %v", fiveg.RTT)
	}
	
	if fiveg.Loss != 0.001 {
		t.Errorf("5G: expected loss 0.001, got %f", fiveg.Loss)
	}
	
	if fiveg.Bandwidth != 50000 {
		t.Errorf("5G: expected bandwidth 50000 KB/s, got %f", fiveg.Bandwidth)
	}
	
	// Тест спутникового профиля
	sat, err := GetNetworkProfile("satellite")
	if err != nil {
		t.Fatalf("Failed to get satellite profile: %v", err)
	}
	
	if sat.RTT != 500*time.Millisecond {
		t.Errorf("Satellite: expected RTT 500ms, got %v", sat.RTT)
	}
	
	if sat.Loss != 0.01 {
		t.Errorf("Satellite: expected loss 0.01, got %f", sat.Loss)
	}
	
	if sat.Bandwidth != 500 {
		t.Errorf("Satellite: expected bandwidth 500 KB/s, got %f", sat.Bandwidth)
	}
}

// Вспомогательная функция для проверки содержания строки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
