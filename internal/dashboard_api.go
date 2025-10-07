package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DashboardState хранит состояние дашборда
type DashboardState struct {
	mu           sync.RWMutex
	ServerRunning bool                 `json:"server_running"`
	ClientRunning bool                 `json:"client_running"`
	TestConfig   TestConfig           `json:"test_config"`
	Metrics      map[string]interface{} `json:"metrics"`
	LastUpdate   time.Time            `json:"last_update"`
}

// DashboardAPI предоставляет REST API для дашборда
type DashboardAPI struct {
	state *DashboardState
}

// NewDashboardAPI создает новый экземпляр API
func NewDashboardAPI() *DashboardAPI {
	return &DashboardAPI{
		state: &DashboardState{
			ServerRunning: false,
			ClientRunning: false,
			Metrics:      make(map[string]interface{}),
			LastUpdate:   time.Now(),
		},
	}
}

// StatusHandler возвращает статус дашборда
func (api *DashboardAPI) StatusHandler(w http.ResponseWriter, r *http.Request) {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server": map[string]interface{}{
			"running": api.state.ServerRunning,
		},
		"client": map[string]interface{}{
			"running": api.state.ClientRunning,
		},
		"last_update": api.state.LastUpdate,
	})
}

// RunTestHandler запускает тест
func (api *DashboardAPI) RunTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var config TestConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	api.state.mu.Lock()
	api.state.TestConfig = config
	api.state.ClientRunning = true
	api.state.LastUpdate = time.Now()
	api.state.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"message": "Test started",
		"config":  config,
	})
}

// StopTestHandler останавливает тест
func (api *DashboardAPI) StopTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	api.state.mu.Lock()
	api.state.ClientRunning = false
	api.state.ServerRunning = false
	api.state.LastUpdate = time.Now()
	api.state.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "stopped",
		"message": "Test stopped",
	})
}

// PresetHandler управляет пресетами
func (api *DashboardAPI) PresetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Возвращаем список пресетов
		presets := map[string]interface{}{
			"scenarios": ListScenarios(),
			"profiles":  ListNetworkProfiles(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(presets)
		
	case http.MethodPost:
		// Применяем пресет
		var request struct {
			Type string `json:"type"`
			Name string `json:"name"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		var config TestConfig
		
		switch request.Type {
		case "scenario":
			scenario, err := GetScenario(request.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			config = scenario.Config
		case "profile":
			profile, err := GetNetworkProfile(request.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			config = TestConfig{
				Mode: "test",
				Addr: ":9000",
			}
			ApplyNetworkProfile(&config, profile)
		default:
			http.Error(w, "Invalid preset type", http.StatusBadRequest)
			return
		}
		
		api.state.mu.Lock()
		api.state.TestConfig = config
		api.state.LastUpdate = time.Now()
		api.state.mu.Unlock()
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "applied",
			"message": fmt.Sprintf("Preset %s applied", request.Name),
			"config":  config,
		})
	}
}

// ReportHandler возвращает отчет
func (api *DashboardAPI) ReportHandler(w http.ResponseWriter, r *http.Request) {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()
	
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"config":  api.state.TestConfig,
			"metrics": api.state.Metrics,
			"timestamp": api.state.LastUpdate,
		})
	case "csv":
		// TODO: реализовать CSV формат
		http.Error(w, "CSV format not implemented", http.StatusNotImplemented)
	case "md":
		// TODO: реализовать Markdown формат
		http.Error(w, "Markdown format not implemented", http.StatusNotImplemented)
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}

// MetricsHandler возвращает метрики
func (api *DashboardAPI) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.state.Metrics)
}

// UpdateMetrics обновляет метрики
func (api *DashboardAPI) UpdateMetrics(metrics map[string]interface{}) {
	api.state.mu.Lock()
	defer api.state.mu.Unlock()
	
	api.state.Metrics = metrics
	api.state.LastUpdate = time.Now()
}

// GetState возвращает текущее состояние
func (api *DashboardAPI) GetState() DashboardState {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()
	
	return *api.state
}

// SetServerState устанавливает состояние сервера
func (api *DashboardAPI) SetServerState(running bool) {
	api.state.mu.Lock()
	defer api.state.mu.Unlock()
	
	api.state.ServerRunning = running
	api.state.LastUpdate = time.Now()
}

// SetClientState устанавливает состояние клиента
func (api *DashboardAPI) SetClientState(running bool) {
	api.state.mu.Lock()
	defer api.state.mu.Unlock()
	
	api.state.ClientRunning = running
	api.state.LastUpdate = time.Now()
}
