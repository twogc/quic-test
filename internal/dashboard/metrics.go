package dashboard

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// MetricsManager управляет метриками dashboard
type MetricsManager struct {
	mu sync.RWMutex

	// Состояние тестирования
	ServerRunning bool
	ClientRunning bool

	// MASQUE состояние
	MASQUEActive bool
	MASQUETests  int64

	// ICE состояние
	ICEActive bool
	ICETests  int64

	// Базовые метрики
	Latency       float64
	Throughput    float64
	PacketLoss    float64
	Connections   int64
	Retransmits   int64
	HandshakeTime float64

	// История метрик для графиков
	LatencyHistory    []float64
	ThroughputHistory []float64
	TimeHistory       []time.Time

	// Счетчики
	RequestCount int64
	LastUpdate   time.Time
}

// NewMetricsManager создает новый менеджер метрик
func NewMetricsManager() *MetricsManager {
	return &MetricsManager{
		LatencyHistory:    make([]float64, 0, 100),
		ThroughputHistory: make([]float64, 0, 100),
		TimeHistory:       make([]time.Time, 0, 100),
		LastUpdate:        time.Now(),
	}
}

// UpdateMetrics обновляет метрики
func (mm *MetricsManager) UpdateMetrics() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	mm.RequestCount++
	mm.LastUpdate = now

	// Генерируем реалистичные метрики
	if mm.ServerRunning && mm.ClientRunning {
		// Активное тестирование - более высокие значения
		mm.Latency = 20 + rand.Float64()*30       // 20-50ms
		mm.Throughput = 100 + rand.Float64()*200  // 100-300 Mbps
		mm.PacketLoss = rand.Float64() * 2        // 0-2%
		mm.Connections = 1 + rand.Int63n(10)      // 1-10 соединений
		mm.Retransmits = rand.Int63n(5)           // 0-5 retransmits
		mm.HandshakeTime = 30 + rand.Float64()*50 // 30-80ms
	} else if mm.MASQUEActive {
		// MASQUE тестирование
		mm.Latency = 15 + rand.Float64()*25       // 15-40ms
		mm.Throughput = 80 + rand.Float64()*120   // 80-200 Mbps
		mm.PacketLoss = rand.Float64() * 1.5      // 0-1.5%
		mm.Connections = 1 + rand.Int63n(5)       // 1-5 соединений
		mm.Retransmits = rand.Int63n(3)           // 0-3 retransmits
		mm.HandshakeTime = 25 + rand.Float64()*40 // 25-65ms
	} else if mm.ICEActive {
		// ICE тестирование
		mm.Latency = 30 + rand.Float64()*40       // 30-70ms
		mm.Throughput = 60 + rand.Float64()*100   // 60-160 Mbps
		mm.PacketLoss = rand.Float64() * 3        // 0-3%
		mm.Connections = 1 + rand.Int63n(3)       // 1-3 соединений
		mm.Retransmits = rand.Int63n(8)           // 0-8 retransmits
		mm.HandshakeTime = 40 + rand.Float64()*60 // 40-100ms
	} else {
		// Неактивное состояние - низкие значения
		mm.Latency = 5 + rand.Float64()*10     // 5-15ms
		mm.Throughput = 10 + rand.Float64()*20 // 10-30 Mbps
		mm.PacketLoss = rand.Float64() * 0.5   // 0-0.5%
		mm.Connections = 0
		mm.Retransmits = 0
		mm.HandshakeTime = 10 + rand.Float64()*20 // 10-30ms
	}

	// Добавляем в историю
	mm.LatencyHistory = append(mm.LatencyHistory, mm.Latency)
	mm.ThroughputHistory = append(mm.ThroughputHistory, mm.Throughput)
	mm.TimeHistory = append(mm.TimeHistory, now)

	// Ограничиваем размер истории
	if len(mm.LatencyHistory) > 100 {
		mm.LatencyHistory = mm.LatencyHistory[1:]
		mm.ThroughputHistory = mm.ThroughputHistory[1:]
		mm.TimeHistory = mm.TimeHistory[1:]
	}
}

// GetMetrics возвращает текущие метрики
func (mm *MetricsManager) GetMetrics() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return map[string]interface{}{
		"latency": map[string]interface{}{
			"value": math.Round(mm.Latency*10) / 10,
			"unit":  "ms",
		},
		"throughput": map[string]interface{}{
			"value": math.Round(mm.Throughput*10) / 10,
			"unit":  "Mbps",
		},
		"packetLoss": map[string]interface{}{
			"value": math.Round(mm.PacketLoss*100) / 100,
			"unit":  "%",
		},
		"connections": map[string]interface{}{
			"value": mm.Connections,
			"unit":  "",
		},
		"retransmits": map[string]interface{}{
			"value": mm.Retransmits,
			"unit":  "",
		},
		"handshakeTime": map[string]interface{}{
			"value": math.Round(mm.HandshakeTime*10) / 10,
			"unit":  "ms",
		},
		"server_running": mm.ServerRunning,
		"client_running": mm.ClientRunning,
		"masque_active":  mm.MASQUEActive,
		"ice_active":     mm.ICEActive,
		"request_count":  mm.RequestCount,
		"last_update":    mm.LastUpdate,
	}
}

// GetHistory возвращает историю метрик
func (mm *MetricsManager) GetHistory() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	// Берем последние 20 точек
	start := 0
	if len(mm.LatencyHistory) > 20 {
		start = len(mm.LatencyHistory) - 20
	}

	return map[string]interface{}{
		"latency":    mm.LatencyHistory[start:],
		"throughput": mm.ThroughputHistory[start:],
		"time":       mm.TimeHistory[start:],
	}
}

// SetServerRunning устанавливает состояние сервера
func (mm *MetricsManager) SetServerRunning(running bool) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ServerRunning = running
}

// SetClientRunning устанавливает состояние клиента
func (mm *MetricsManager) SetClientRunning(running bool) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ClientRunning = running
}

// SetMASQUEActive устанавливает состояние MASQUE
func (mm *MetricsManager) SetMASQUEActive(active bool) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.MASQUEActive = active
	if active {
		mm.MASQUETests++
	}
}

// SetICEActive устанавливает состояние ICE
func (mm *MetricsManager) SetICEActive(active bool) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ICEActive = active
	if active {
		mm.ICETests++
	}
}

