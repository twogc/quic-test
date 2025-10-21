package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MetricsBridge обеспечивает передачу метрик в QUIC Bottom
type MetricsBridge struct {
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
	mu         sync.RWMutex
	metrics    QUICMetrics
}

// QUICMetrics структура метрик для передачи в QUIC Bottom
type QUICMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	Latency       float64  `json:"latency"`        // в миллисекундах
	Throughput    float64  `json:"throughput"`     // в Mbps
	Connections   int      `json:"connections"`
	Errors        int      `json:"errors"`
	PacketLoss    float64  `json:"packet_loss"`    // в процентах
	Retransmits   int      `json:"retransmits"`
	Jitter        float64  `json:"jitter"`         // в миллисекундах
	CongestionWindow int   `json:"congestion_window"`
	RTT           float64  `json:"rtt"`            // в миллисекундах
	BytesReceived int64    `json:"bytes_received"`
	BytesSent     int64    `json:"bytes_sent"`
	Streams       int      `json:"streams"`
	HandshakeTime float64  `json:"handshake_time"` // в миллисекундах
}

// NewMetricsBridge создает новый мост для передачи метрик
func NewMetricsBridge(logger *zap.Logger, baseURL string) *MetricsBridge {
	return &MetricsBridge{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: baseURL,
	}
}

// UpdateMetrics обновляет метрики и отправляет их в QUIC Bottom
func (mb *MetricsBridge) UpdateMetrics(metrics QUICMetrics) error {
	mb.mu.Lock()
	mb.metrics = metrics
	mb.mu.Unlock()

	// Отправляем метрики в QUIC Bottom
	return mb.sendMetricsToBottom(metrics)
}

// GetCurrentMetrics возвращает текущие метрики
func (mb *MetricsBridge) GetCurrentMetrics() QUICMetrics {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.metrics
}

// sendMetricsToBottom отправляет метрики в QUIC Bottom через HTTP API
func (mb *MetricsBridge) sendMetricsToBottom(metrics QUICMetrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %v", err)
	}

	url := fmt.Sprintf("%s/api/metrics", mb.baseURL)
	resp, err := mb.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Не критично, если QUIC Bottom не запущен
		mb.logger.Debug("Failed to send metrics to QUIC Bottom", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		mb.logger.Debug("QUIC Bottom returned non-OK status", zap.Int("status", resp.StatusCode))
	}

	return nil
}

// StartMetricsServer запускает HTTP сервер для получения метрик от QUIC Bottom
func (mb *MetricsBridge) StartMetricsServer(port int) error {
	mux := http.NewServeMux()
	
	// Эндпоинт для получения метрик
	mux.HandleFunc("/api/metrics", mb.handleMetrics)
	
	// Эндпоинт для проверки здоровья
	mux.HandleFunc("/health", mb.handleHealth)
	
	// Эндпоинт для получения текущих метрик
	mux.HandleFunc("/api/current", mb.handleCurrentMetrics)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	mb.logger.Info("Starting metrics bridge server", zap.Int("port", port))
	return server.ListenAndServe()
}

// handleMetrics обрабатывает POST запросы с метриками
func (mb *MetricsBridge) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var metrics QUICMetrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mb.mu.Lock()
	mb.metrics = metrics
	mb.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleHealth обрабатывает запросы проверки здоровья
func (mb *MetricsBridge) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// handleCurrentMetrics возвращает текущие метрики
func (mb *MetricsBridge) handleCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := mb.GetCurrentMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// MetricsCollector собирает метрики из QUIC соединений
type MetricsCollector struct {
	bridge *MetricsBridge
	logger *zap.Logger
}

// NewMetricsCollector создает новый сборщик метрик
func NewMetricsCollector(logger *zap.Logger, bridge *MetricsBridge) *MetricsCollector {
	return &MetricsCollector{
		bridge: bridge,
		logger: logger,
	}
}

// CollectMetrics собирает метрики из QUIC соединения
func (mc *MetricsCollector) CollectMetrics(conn *quic.Connection) error {
	// Получаем статистику соединения
	stats := conn.GetStats()
	
	metrics := QUICMetrics{
		Timestamp:       time.Now(),
		Latency:         stats.RTT.Milliseconds(),
		Throughput:      mc.calculateThroughput(stats),
		Connections:     1, // Для одного соединения
		Errors:          int(stats.PacketsLost),
		PacketLoss:      mc.calculatePacketLoss(stats),
		Retransmits:     int(stats.PacketsRetransmitted),
		Jitter:          mc.calculateJitter(stats),
		CongestionWindow: int(stats.CongestionWindow),
		RTT:             stats.RTT.Milliseconds(),
		BytesReceived:   stats.BytesReceived,
		BytesSent:       stats.BytesSent,
		Streams:         int(stats.StreamsOpened),
		HandshakeTime:   stats.HandshakeTime.Milliseconds(),
	}

	return mc.bridge.UpdateMetrics(metrics)
}

// calculateThroughput вычисляет пропускную способность
func (mc *MetricsCollector) calculateThroughput(stats quic.ConnStats) float64 {
	// Простое вычисление на основе переданных байт за время
	if stats.Duration > 0 {
		bytesPerSecond := float64(stats.BytesReceived) / stats.Duration.Seconds()
		return bytesPerSecond / (1024 * 1024) // Конвертируем в Mbps
	}
	return 0
}

// calculatePacketLoss вычисляет потерю пакетов
func (mc *MetricsCollector) calculatePacketLoss(stats quic.ConnStats) float64 {
	totalPackets := stats.PacketsSent + stats.PacketsLost
	if totalPackets > 0 {
		return float64(stats.PacketsLost) / float64(totalPackets) * 100
	}
	return 0
}

// calculateJitter вычисляет джиттер
func (mc *MetricsCollector) calculateJitter(stats quic.ConnStats) float64 {
	// Упрощенное вычисление джиттера
	// В реальной реализации нужно анализировать временные метки пакетов
	return stats.RTT.Milliseconds() * 0.1 // Примерное значение
}
