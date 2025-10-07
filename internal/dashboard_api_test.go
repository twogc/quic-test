package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDashboardAPIStatus(t *testing.T) {
	api := NewDashboardAPI()
	
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.StatusHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	// Проверяем структуру ответа
	if _, ok := response["server"]; !ok {
		t.Error("Response missing 'server' field")
	}
	if _, ok := response["client"]; !ok {
		t.Error("Response missing 'client' field")
	}
	if _, ok := response["last_update"]; !ok {
		t.Error("Response missing 'last_update' field")
	}
}

func TestDashboardAPIRunTest(t *testing.T) {
	api := NewDashboardAPI()
	
	config := TestConfig{
		Mode:        "test",
		Addr:        ":9000",
		Connections: 2,
		Streams:     4,
		Duration:    30 * time.Second,
	}
	
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/run-test", bytes.NewBuffer(configJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	api.RunTestHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Run test handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if response["status"] != "started" {
		t.Errorf("Expected status 'started', got %v", response["status"])
	}
	
	// Проверяем, что состояние обновилось
	state := api.GetState()
	if !state.ClientRunning {
		t.Error("Expected client to be running")
	}
}

func TestDashboardAPIStopTest(t *testing.T) {
	api := NewDashboardAPI()
	
	// Сначала запускаем тест
	api.SetClientState(true)
	api.SetServerState(true)
	
	req, err := http.NewRequest("POST", "/stop-test", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.StopTestHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Stop test handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if response["status"] != "stopped" {
		t.Errorf("Expected status 'stopped', got %v", response["status"])
	}
	
	// Проверяем, что состояние обновилось
	state := api.GetState()
	if state.ClientRunning {
		t.Error("Expected client to be stopped")
	}
	if state.ServerRunning {
		t.Error("Expected server to be stopped")
	}
}

func TestDashboardAPIPresetHandler(t *testing.T) {
	api := NewDashboardAPI()
	
	// Тест GET запроса для получения списка пресетов
	req, err := http.NewRequest("GET", "/presets", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.PresetHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Preset handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if _, ok := response["scenarios"]; !ok {
		t.Error("Response missing 'scenarios' field")
	}
	if _, ok := response["profiles"]; !ok {
		t.Error("Response missing 'profiles' field")
	}
}

func TestDashboardAPIPresetHandlerScenario(t *testing.T) {
	api := NewDashboardAPI()
	
	// Тест POST запроса для применения сценария
	request := map[string]string{
		"type": "scenario",
		"name": "wifi",
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/presets", bytes.NewBuffer(requestJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	api.PresetHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Preset handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if response["status"] != "applied" {
		t.Errorf("Expected status 'applied', got %v", response["status"])
	}
}

func TestDashboardAPIPresetHandlerProfile(t *testing.T) {
	api := NewDashboardAPI()
	
	// Тест POST запроса для применения профиля
	request := map[string]string{
		"type": "profile",
		"name": "wifi",
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/presets", bytes.NewBuffer(requestJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	api.PresetHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Preset handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if response["status"] != "applied" {
		t.Errorf("Expected status 'applied', got %v", response["status"])
	}
}

func TestDashboardAPIPresetHandlerInvalidType(t *testing.T) {
	api := NewDashboardAPI()
	
	// Тест с неверным типом пресета
	request := map[string]string{
		"type": "invalid",
		"name": "test",
	}
	
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/presets", bytes.NewBuffer(requestJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	api.PresetHandler(rr, req)
	
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Preset handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestDashboardAPIMetrics(t *testing.T) {
	api := NewDashboardAPI()
	
	// Устанавливаем тестовые метрики
	testMetrics := map[string]interface{}{
		"Success":           100,
		"Errors":            5,
		"BytesSent":         int64(1024000),
		"BytesReceived":     int64(1024000),
		"LatencyAverage":    25.5,
		"ThroughputAverage": 1000.0,
		"PacketLoss":        0.01,
	}
	
	api.UpdateMetrics(testMetrics)
	
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.MetricsHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Metrics handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	// Проверяем, что метрики соответствуют установленным
	if response["Success"] != float64(100) {
		t.Errorf("Expected Success to be 100, got %v", response["Success"])
	}
	if response["Errors"] != float64(5) {
		t.Errorf("Expected Errors to be 5, got %v", response["Errors"])
	}
}

func TestDashboardAPIReportJSON(t *testing.T) {
	api := NewDashboardAPI()
	
	// Устанавливаем тестовые данные
	config := TestConfig{
		Mode:        "test",
		Addr:        ":9000",
		Connections: 2,
		Streams:     4,
	}
	
	api.state.TestConfig = config
	api.UpdateMetrics(map[string]interface{}{
		"Success": 100,
		"Errors":  5,
	})
	
	req, err := http.NewRequest("GET", "/report?format=json", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.ReportHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Report handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if _, ok := response["config"]; !ok {
		t.Error("Response missing 'config' field")
	}
	if _, ok := response["metrics"]; !ok {
		t.Error("Response missing 'metrics' field")
	}
	if _, ok := response["timestamp"]; !ok {
		t.Error("Response missing 'timestamp' field")
	}
}

func TestDashboardAPIReportCSV(t *testing.T) {
	api := NewDashboardAPI()
	
	// Устанавливаем тестовые данные
	config := TestConfig{
		Mode:        "test",
		Addr:        ":9000",
		Connections: 2,
		Streams:     4,
		PacketSize:  1200,
		Rate:        100,
	}
	
	api.state.TestConfig = config
	api.UpdateMetrics(map[string]interface{}{
		"Success":           100,
		"Errors":            5,
		"BytesSent":         int64(1024000),
		"BytesReceived":     int64(1024000),
		"LatencyAverage":    25.5,
		"ThroughputAverage": 1000.0,
		"PacketLoss":        0.01,
	})
	
	req, err := http.NewRequest("GET", "/report?format=csv", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.ReportHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Report handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Проверяем, что это CSV
	if rr.Header().Get("Content-Type") != "text/csv" {
		t.Errorf("Expected Content-Type 'text/csv', got %v", rr.Header().Get("Content-Type"))
	}
	
	// Проверяем содержимое CSV
	body := rr.Body.String()
	if !contains(body, "Parameter,Value") {
		t.Error("CSV missing header")
	}
	if !contains(body, "Mode,test") {
		t.Error("CSV missing mode")
	}
	if !contains(body, "Connections,2") {
		t.Error("CSV missing connections")
	}
}

func TestDashboardAPIReportMarkdown(t *testing.T) {
	api := NewDashboardAPI()
	
	// Устанавливаем тестовые данные
	config := TestConfig{
		Mode:        "test",
		Addr:        ":9000",
		Connections: 2,
		Streams:     4,
		PacketSize:  1200,
		Rate:        100,
	}
	
	api.state.TestConfig = config
	api.UpdateMetrics(map[string]interface{}{
		"Success":           100,
		"Errors":            5,
		"BytesSent":         int64(1024000),
		"BytesReceived":     int64(1024000),
		"LatencyAverage":    25.5,
		"ThroughputAverage": 1000.0,
		"PacketLoss":        0.01,
		"Latencies":         []float64{20.0, 25.0, 30.0, 35.0, 40.0},
	})
	
	req, err := http.NewRequest("GET", "/report?format=md", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.ReportHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Report handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Проверяем, что это Markdown
	if rr.Header().Get("Content-Type") != "text/markdown" {
		t.Errorf("Expected Content-Type 'text/markdown', got %v", rr.Header().Get("Content-Type"))
	}
	
	// Проверяем содержимое Markdown
	body := rr.Body.String()
	if !contains(body, "# QUIC Test Report") {
		t.Error("Markdown missing title")
	}
	if !contains(body, "## Test Configuration") {
		t.Error("Markdown missing configuration section")
	}
	if !contains(body, "## Test Results") {
		t.Error("Markdown missing results section")
	}
}

func TestDashboardAPIReportUnsupportedFormat(t *testing.T) {
	api := NewDashboardAPI()
	
	req, err := http.NewRequest("GET", "/report?format=unsupported", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	api.ReportHandler(rr, req)
	
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Report handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// Вспомогательная функция для проверки содержимого строки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}