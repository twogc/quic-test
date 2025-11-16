package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
		// Реализуем CSV формат
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=report.csv")
		generateCSVReport(w, api.state.TestConfig, api.state.Metrics)
	case "md":
		// Реализуем Markdown формат
		w.Header().Set("Content-Type", "text/markdown")
		w.Header().Set("Content-Disposition", "attachment; filename=report.md")
		generateMarkdownReport(w, api.state.TestConfig, api.state.Metrics)
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
func (api *DashboardAPI) GetState() *DashboardState {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()
	
	// Создаем копию состояния без мьютекса
	state := &DashboardState{
		ServerRunning: api.state.ServerRunning,
		ClientRunning: api.state.ClientRunning,
		TestConfig:   api.state.TestConfig,
		Metrics:      make(map[string]interface{}),
		LastUpdate:   api.state.LastUpdate,
	}
	
	// Копируем метрики
	for k, v := range api.state.Metrics {
		state.Metrics[k] = v
	}
	
	return state
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

// generateCSVReport генерирует CSV отчет
func generateCSVReport(w http.ResponseWriter, config TestConfig, metrics map[string]interface{}) {
	// Заголовки CSV
	fmt.Fprintf(w, "Parameter,Value\n")
	fmt.Fprintf(w, "Mode,%s\n", config.Mode)
	fmt.Fprintf(w, "Address,%s\n", config.Addr)
	fmt.Fprintf(w, "Connections,%d\n", config.Connections)
	fmt.Fprintf(w, "Streams,%d\n", config.Streams)
	fmt.Fprintf(w, "Packet Size,%d\n", config.PacketSize)
	fmt.Fprintf(w, "Rate,%d\n", config.Rate)
	fmt.Fprintf(w, "Duration,%v\n", config.Duration)
	
	// Метрики
	if success, ok := metrics["Success"].(int); ok {
		fmt.Fprintf(w, "Success,%d\n", success)
	}
	if errors, ok := metrics["Errors"].(int); ok {
		fmt.Fprintf(w, "Errors,%d\n", errors)
	}
	if bytesSent, ok := metrics["BytesSent"].(int64); ok {
		fmt.Fprintf(w, "Bytes Sent,%d\n", bytesSent)
	}
	if bytesReceived, ok := metrics["BytesReceived"].(int64); ok {
		fmt.Fprintf(w, "Bytes Received,%d\n", bytesReceived)
	}
	if latency, ok := metrics["LatencyAverage"].(float64); ok {
		fmt.Fprintf(w, "Average Latency (ms),%.2f\n", latency)
	}
	if throughput, ok := metrics["ThroughputAverage"].(float64); ok {
		fmt.Fprintf(w, "Average Throughput (KB/s),%.2f\n", throughput)
	}
	if packetLoss, ok := metrics["PacketLoss"].(float64); ok {
		fmt.Fprintf(w, "Packet Loss (%%),%.2f\n", packetLoss*100)
	}
}

// generateMarkdownReport генерирует Markdown отчет
func generateMarkdownReport(w http.ResponseWriter, config TestConfig, metrics map[string]interface{}) {
	fmt.Fprintf(w, "# QUIC Test Report\n\n")
	fmt.Fprintf(w, "**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	
	// Конфигурация теста
	fmt.Fprintf(w, "## Test Configuration\n\n")
	fmt.Fprintf(w, "| Parameter | Value |\n")
	fmt.Fprintf(w, "|-----------|-------|\n")
	fmt.Fprintf(w, "| Mode | %s |\n", config.Mode)
	fmt.Fprintf(w, "| Address | %s |\n", config.Addr)
	fmt.Fprintf(w, "| Connections | %d |\n", config.Connections)
	fmt.Fprintf(w, "| Streams | %d |\n", config.Streams)
	fmt.Fprintf(w, "| Packet Size | %d bytes |\n", config.PacketSize)
	fmt.Fprintf(w, "| Rate | %d packets/s |\n", config.Rate)
	fmt.Fprintf(w, "| Duration | %v |\n", config.Duration)
	fmt.Fprintf(w, "\n")
	
	// Основные метрики
	fmt.Fprintf(w, "## Test Results\n\n")
	fmt.Fprintf(w, "| Metric | Value |\n")
	fmt.Fprintf(w, "|--------|-------|\n")
	
	if success, ok := metrics["Success"].(int); ok {
		fmt.Fprintf(w, "| Successful Requests | %d |\n", success)
	}
	if errors, ok := metrics["Errors"].(int); ok {
		fmt.Fprintf(w, "| Errors | %d |\n", errors)
	}
	if bytesSent, ok := metrics["BytesSent"].(int64); ok {
		fmt.Fprintf(w, "| Bytes Sent | %d |\n", bytesSent)
	}
	if bytesReceived, ok := metrics["BytesReceived"].(int64); ok {
		fmt.Fprintf(w, "| Bytes Received | %d |\n", bytesReceived)
	}
	if latency, ok := metrics["LatencyAverage"].(float64); ok {
		fmt.Fprintf(w, "| Average Latency | %.2f ms |\n", latency)
	}
	if throughput, ok := metrics["ThroughputAverage"].(float64); ok {
		fmt.Fprintf(w, "| Average Throughput | %.2f KB/s |\n", throughput)
	}
	if packetLoss, ok := metrics["PacketLoss"].(float64); ok {
		fmt.Fprintf(w, "| Packet Loss | %.2f%% |\n", packetLoss*100)
	}
	
	// Детальные метрики задержки
	if latencies, ok := metrics["Latencies"].([]float64); ok && len(latencies) > 0 {
		fmt.Fprintf(w, "\n## Latency Statistics\n\n")
		fmt.Fprintf(w, "| Percentile | Value (ms) |\n")
		fmt.Fprintf(w, "|------------|------------|\n")
		
		// Сортируем для расчета перцентилей
		sort.Float64s(latencies)
		n := len(latencies)
		
		if n > 0 {
			fmt.Fprintf(w, "| Min | %.2f |\n", latencies[0])
			fmt.Fprintf(w, "| Max | %.2f |\n", latencies[n-1])
			fmt.Fprintf(w, "| P50 | %.2f |\n", latencies[int(float64(n)*0.5)])
			fmt.Fprintf(w, "| P95 | %.2f |\n", latencies[int(float64(n)*0.95)])
			fmt.Fprintf(w, "| P99 | %.2f |\n", latencies[int(float64(n)*0.99)])
		}
	}
	
	fmt.Fprintf(w, "\n## Summary\n\n")
	fmt.Fprintf(w, "Test completed successfully with the above metrics.\n")
}
