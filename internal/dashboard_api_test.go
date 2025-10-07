package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDashboardAPI(t *testing.T) {
	api := NewDashboardAPI()
	
	if api == nil {
		t.Fatal("Expected non-nil API")
	}
	
	if api.state == nil {
		t.Fatal("Expected non-nil state")
	}
	
	if api.state.ServerRunning {
		t.Error("Expected server to not be running initially")
	}
	
	if api.state.ClientRunning {
		t.Error("Expected client to not be running initially")
	}
	
	if api.state.Metrics == nil {
		t.Error("Expected metrics map to be initialized")
	}
}

func TestStatusHandler(t *testing.T) {
	api := NewDashboardAPI()
	
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	
	api.StatusHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Проверяем структуру ответа
	if _, ok := response["server"]; !ok {
		t.Error("Expected 'server' field in response")
	}
	
	if _, ok := response["client"]; !ok {
		t.Error("Expected 'client' field in response")
	}
	
	if _, ok := response["last_update"]; !ok {
		t.Error("Expected 'last_update' field in response")
	}
}

func TestRunTestHandler(t *testing.T) {
	api := NewDashboardAPI()
	
	config := TestConfig{
		Mode:        "test",
		Addr:        ":9000",
		Connections: 2,
		Streams:     4,
		Duration:    30 * time.Second,
		PacketSize:  1200,
		Rate:        100,
	}
	
	configJSON, _ := json.Marshal(config)
	req := httptest.NewRequest("POST", "/api/run", bytes.NewReader(configJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	api.RunTestHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "started" {
		t.Errorf("Expected status 'started', got %s", response["status"])
	}
	
	// Проверяем, что состояние обновилось
	state := api.GetState()
	if !state.ClientRunning {
		t.Error("Expected client to be running")
	}
	
	if state.TestConfig.Mode != config.Mode {
		t.Errorf("Expected mode %s, got %s", config.Mode, state.TestConfig.Mode)
	}
}

func TestRunTestHandlerInvalidJSON(t *testing.T) {
	api := NewDashboardAPI()
	
	req := httptest.NewRequest("POST", "/api/run", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	api.RunTestHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRunTestHandlerWrongMethod(t *testing.T) {
	api := NewDashboardAPI()
	
	req := httptest.NewRequest("GET", "/api/run", nil)
	w := httptest.NewRecorder()
	
	api.RunTestHandler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestStopTestHandler(t *testing.T) {
	api := NewDashboardAPI()
	
	// Сначала запускаем тест
	api.SetClientState(true)
	api.SetServerState(true)
	
	req := httptest.NewRequest("POST", "/api/stop", nil)
	w := httptest.NewRecorder()
	
	api.StopTestHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "stopped" {
		t.Errorf("Expected status 'stopped', got %s", response["status"])
	}
	
	// Проверяем, что состояние обновилось
	state := api.GetState()
	if state.ClientRunning {
		t.Error("Expected client to not be running")
	}
	
	if state.ServerRunning {
		t.Error("Expected server to not be running")
	}
}

func TestPresetHandlerGet(t *testing.T) {
	api := NewDashboardAPI()
	
	req := httptest.NewRequest("GET", "/api/preset", nil)
	w := httptest.NewRecorder()
	
	api.PresetHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if _, ok := response["scenarios"]; !ok {
		t.Error("Expected 'scenarios' field in response")
	}
	
	if _, ok := response["profiles"]; !ok {
		t.Error("Expected 'profiles' field in response")
	}
}

func TestPresetHandlerPostScenario(t *testing.T) {
	api := NewDashboardAPI()
	
	request := map[string]string{
		"type": "scenario",
		"name": "wifi",
	}
	
	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/preset", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	api.PresetHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "applied" {
		t.Errorf("Expected status 'applied', got %s", response["status"])
	}
	
	// Проверяем, что конфигурация обновилась
	state := api.GetState()
	if state.TestConfig.Mode != "test" {
		t.Errorf("Expected mode 'test', got %s", state.TestConfig.Mode)
	}
}

func TestPresetHandlerPostProfile(t *testing.T) {
	api := NewDashboardAPI()
	
	request := map[string]string{
		"type": "profile",
		"name": "lte",
	}
	
	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/preset", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	api.PresetHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "applied" {
		t.Errorf("Expected status 'applied', got %s", response["status"])
	}
}

func TestPresetHandlerPostInvalidType(t *testing.T) {
	api := NewDashboardAPI()
	
	request := map[string]string{
		"type": "invalid",
		"name": "test",
	}
	
	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/preset", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	api.PresetHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestPresetHandlerPostInvalidJSON(t *testing.T) {
	api := NewDashboardAPI()
	
	req := httptest.NewRequest("POST", "/api/preset", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	api.PresetHandler(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestReportHandler(t *testing.T) {
	api := NewDashboardAPI()
	
	// Устанавливаем тестовые метрики
	metrics := map[string]interface{}{
		"Success": true,
		"Errors":  5,
		"BytesSent": int64(1024000),
	}
	api.UpdateMetrics(metrics)
	
	req := httptest.NewRequest("GET", "/api/report?format=json", nil)
	w := httptest.NewRecorder()
	
	api.ReportHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if _, ok := response["config"]; !ok {
		t.Error("Expected 'config' field in response")
	}
	
	if _, ok := response["metrics"]; !ok {
		t.Error("Expected 'metrics' field in response")
	}
	
	if _, ok := response["timestamp"]; !ok {
		t.Error("Expected 'timestamp' field in response")
	}
}

func TestMetricsHandler(t *testing.T) {
	api := NewDashboardAPI()
	
	// Устанавливаем тестовые метрики
	metrics := map[string]interface{}{
		"Success": true,
		"Errors":  5,
		"BytesSent": int64(1024000),
	}
	api.UpdateMetrics(metrics)
	
	req := httptest.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()
	
	api.MetricsHandler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["Success"] != true {
		t.Error("Expected Success to be true")
	}
	
	if response["Errors"] != float64(5) {
		t.Errorf("Expected Errors 5, got %v", response["Errors"])
	}
}

func TestSetServerState(t *testing.T) {
	api := NewDashboardAPI()
	
	api.SetServerState(true)
	
	state := api.GetState()
	if !state.ServerRunning {
		t.Error("Expected server to be running")
	}
	
	api.SetServerState(false)
	
	state = api.GetState()
	if state.ServerRunning {
		t.Error("Expected server to not be running")
	}
}

func TestSetClientState(t *testing.T) {
	api := NewDashboardAPI()
	
	api.SetClientState(true)
	
	state := api.GetState()
	if !state.ClientRunning {
		t.Error("Expected client to be running")
	}
	
	api.SetClientState(false)
	
	state = api.GetState()
	if state.ClientRunning {
		t.Error("Expected client to not be running")
	}
}

func TestUpdateMetrics(t *testing.T) {
	api := NewDashboardAPI()
	
	metrics := map[string]interface{}{
		"Success": true,
		"Errors":  5,
		"BytesSent": int64(1024000),
	}
	
	api.UpdateMetrics(metrics)
	
	state := api.GetState()
	if state.Metrics["Success"] != true {
		t.Error("Expected Success to be true")
	}
	
	if state.Metrics["Errors"] != 5 {
		t.Errorf("Expected Errors 5, got %v", state.Metrics["Errors"])
	}
	
	if state.Metrics["BytesSent"] != int64(1024000) {
		t.Errorf("Expected BytesSent 1024000, got %v", state.Metrics["BytesSent"])
	}
}
