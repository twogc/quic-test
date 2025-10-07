package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// HDRMetrics предоставляет HDR-гистограммы для точных метрик
type HDRMetrics struct {
	mu sync.RWMutex
	
	// Гистограммы для различных метрик
	latencyHist    *hdrhistogram.Histogram
	jitterHist     *hdrhistogram.Histogram
	handshakeHist  *hdrhistogram.Histogram
	throughputHist *hdrhistogram.Histogram
	
	// Счетчики
	packetsSent     int64
	packetsReceived int64
	bytesSent       int64
	bytesReceived   int64
	errors          int64
	retransmits     int64
	
	// Временные ряды
	timeSeries []TimeSeriesPoint
}

// TimeSeriesPoint представляет точку временного ряда
type TimeSeriesPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// NewHDRMetrics создает новый экземпляр HDR метрик
func NewHDRMetrics() *HDRMetrics {
	// Настройка гистограмм для различных диапазонов
	// Латенсия: от 1 микросекунды до 10 секунд
	latencyHist := hdrhistogram.New(1, 10000000, 3) // 1μs to 10s, 3 significant digits
	
	// Джиттер: от 1 микросекунды до 1 секунды
	jitterHist := hdrhistogram.New(1, 1000000, 3) // 1μs to 1s, 3 significant digits
	
	// Handshake: от 1 миллисекунды до 30 секунд
	handshakeHist := hdrhistogram.New(1000, 30000000, 3) // 1ms to 30s, 3 significant digits
	
	// Throughput: от 1 байта до 1GB в секунду
	throughputHist := hdrhistogram.New(1, 1000000000, 3) // 1B to 1GB/s, 3 significant digits
	
	return &HDRMetrics{
		latencyHist:    latencyHist,
		jitterHist:     jitterHist,
		handshakeHist:  handshakeHist,
		throughputHist: throughputHist,
		timeSeries:     make([]TimeSeriesPoint, 0),
	}
}

// RecordLatency записывает задержку в микросекундах
func (h *HDRMetrics) RecordLatency(latency time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	latencyMicros := latency.Microseconds()
	if latencyMicros > 0 {
		h.latencyHist.RecordValue(latencyMicros)
	}
}

// RecordJitter записывает джиттер в микросекундах
func (h *HDRMetrics) RecordJitter(jitter time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	jitterMicros := jitter.Microseconds()
	if jitterMicros > 0 {
		h.jitterHist.RecordValue(jitterMicros)
	}
}

// RecordHandshakeTime записывает время handshake в микросекундах
func (h *HDRMetrics) RecordHandshakeTime(handshakeTime time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	handshakeMicros := handshakeTime.Microseconds()
	if handshakeMicros > 0 {
		h.handshakeHist.RecordValue(handshakeMicros)
	}
}

// RecordThroughput записывает пропускную способность в байтах в секунду
func (h *HDRMetrics) RecordThroughput(throughput float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if throughput > 0 {
		h.throughputHist.RecordValue(int64(throughput))
	}
}

// IncrementPacketsSent увеличивает счетчик отправленных пакетов
func (h *HDRMetrics) IncrementPacketsSent() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.packetsSent++
}

// IncrementPacketsReceived увеличивает счетчик полученных пакетов
func (h *HDRMetrics) IncrementPacketsReceived() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.packetsReceived++
}

// AddBytesSent добавляет количество отправленных байт
func (h *HDRMetrics) AddBytesSent(bytes int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.bytesSent += bytes
}

// AddBytesReceived добавляет количество полученных байт
func (h *HDRMetrics) AddBytesReceived(bytes int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.bytesReceived += bytes
}

// IncrementErrors увеличивает счетчик ошибок
func (h *HDRMetrics) IncrementErrors() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.errors++
}

// IncrementRetransmits увеличивает счетчик ретрансмиссий
func (h *HDRMetrics) IncrementRetransmits() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.retransmits++
}

// AddTimeSeriesPoint добавляет точку временного ряда
func (h *HDRMetrics) AddTimeSeriesPoint(timestamp time.Time, metrics map[string]interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.timeSeries = append(h.timeSeries, TimeSeriesPoint{
		Timestamp: timestamp,
		Metrics:   metrics,
	})
}

// GetLatencyStats возвращает статистику латенсии
func (h *HDRMetrics) GetLatencyStats() LatencyStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if h.latencyHist.TotalCount() == 0 {
		return LatencyStats{}
	}
	
	return LatencyStats{
		P50:   float64(h.latencyHist.ValueAtQuantile(50.0)) / 1000.0, // Конвертируем в миллисекунды
		P90:   float64(h.latencyHist.ValueAtQuantile(90.0)) / 1000.0,
		P95:   float64(h.latencyHist.ValueAtQuantile(95.0)) / 1000.0,
		P99:   float64(h.latencyHist.ValueAtQuantile(99.0)) / 1000.0,
		P999:  float64(h.latencyHist.ValueAtQuantile(99.9)) / 1000.0,
		Min:   float64(h.latencyHist.Min()) / 1000.0,
		Max:   float64(h.latencyHist.Max()) / 1000.0,
		Mean:  h.latencyHist.Mean() / 1000.0,
		Count: h.latencyHist.TotalCount(),
	}
}

// GetJitterStats возвращает статистику джиттера
func (h *HDRMetrics) GetJitterStats() JitterStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if h.jitterHist.TotalCount() == 0 {
		return JitterStats{}
	}
	
	return JitterStats{
		P50:   float64(h.jitterHist.ValueAtQuantile(50.0)) / 1000.0,
		P90:   float64(h.jitterHist.ValueAtQuantile(90.0)) / 1000.0,
		P95:   float64(h.jitterHist.ValueAtQuantile(95.0)) / 1000.0,
		P99:   float64(h.jitterHist.ValueAtQuantile(99.0)) / 1000.0,
		P999:  float64(h.jitterHist.ValueAtQuantile(99.9)) / 1000.0,
		Min:   float64(h.jitterHist.Min()) / 1000.0,
		Max:   float64(h.jitterHist.Max()) / 1000.0,
		Mean:  h.jitterHist.Mean() / 1000.0,
		Count: h.jitterHist.TotalCount(),
	}
}

// GetHandshakeStats возвращает статистику handshake
func (h *HDRMetrics) GetHandshakeStats() HandshakeStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if h.handshakeHist.TotalCount() == 0 {
		return HandshakeStats{}
	}
	
	return HandshakeStats{
		P50:   float64(h.handshakeHist.ValueAtQuantile(50.0)) / 1000.0,
		P90:   float64(h.handshakeHist.ValueAtQuantile(90.0)) / 1000.0,
		P95:   float64(h.handshakeHist.ValueAtQuantile(95.0)) / 1000.0,
		P99:   float64(h.handshakeHist.ValueAtQuantile(99.0)) / 1000.0,
		P999:  float64(h.handshakeHist.ValueAtQuantile(99.9)) / 1000.0,
		Min:   float64(h.handshakeHist.Min()) / 1000.0,
		Max:   float64(h.handshakeHist.Max()) / 1000.0,
		Mean:  h.handshakeHist.Mean() / 1000.0,
		Count: h.handshakeHist.TotalCount(),
	}
}

// GetThroughputStats возвращает статистику пропускной способности
func (h *HDRMetrics) GetThroughputStats() ThroughputStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if h.throughputHist.TotalCount() == 0 {
		return ThroughputStats{}
	}
	
	return ThroughputStats{
		P50:   float64(h.throughputHist.ValueAtQuantile(50.0)) / (1024 * 1024), // Конвертируем в MB/s
		P90:   float64(h.throughputHist.ValueAtQuantile(90.0)) / (1024 * 1024),
		P95:   float64(h.throughputHist.ValueAtQuantile(95.0)) / (1024 * 1024),
		P99:   float64(h.throughputHist.ValueAtQuantile(99.0)) / (1024 * 1024),
		P999:  float64(h.throughputHist.ValueAtQuantile(99.9)) / (1024 * 1024),
		Min:   float64(h.throughputHist.Min()) / (1024 * 1024),
		Max:   float64(h.throughputHist.Max()) / (1024 * 1024),
		Mean:  h.throughputHist.Mean() / (1024 * 1024),
		Count: h.throughputHist.TotalCount(),
	}
}

// GetNetworkStats возвращает сетевую статистику
func (h *HDRMetrics) GetNetworkStats() NetworkStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	lossPercent := 0.0
	if h.packetsSent > 0 {
		lossPercent = float64(h.packetsSent-h.packetsReceived) / float64(h.packetsSent) * 100.0
	}
	
	return NetworkStats{
		PacketsSent:     h.packetsSent,
		PacketsReceived: h.packetsReceived,
		PacketsLost:     h.packetsSent - h.packetsReceived,
		LossPercent:     lossPercent,
		BytesSent:       h.bytesSent,
		BytesReceived:   h.bytesReceived,
		Retransmits:     h.retransmits,
		Errors:          h.errors,
	}
}

// GetTimeSeries возвращает временные ряды
func (h *HDRMetrics) GetTimeSeries() []TimeSeriesPoint {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Возвращаем копию, чтобы избежать race conditions
	result := make([]TimeSeriesPoint, len(h.timeSeries))
	copy(result, h.timeSeries)
	return result
}

// ExportToPrometheus экспортирует метрики в формате Prometheus
func (h *HDRMetrics) ExportToPrometheus() map[string]string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	latencyStats := h.GetLatencyStats()
	jitterStats := h.GetJitterStats()
	handshakeStats := h.GetHandshakeStats()
	throughputStats := h.GetThroughputStats()
	networkStats := h.GetNetworkStats()
	
	metrics := make(map[string]string)
	
	// Latency metrics
	metrics["quic_latency_p50_ms"] = fmt.Sprintf("%.3f", latencyStats.P50)
	metrics["quic_latency_p90_ms"] = fmt.Sprintf("%.3f", latencyStats.P90)
	metrics["quic_latency_p95_ms"] = fmt.Sprintf("%.3f", latencyStats.P95)
	metrics["quic_latency_p99_ms"] = fmt.Sprintf("%.3f", latencyStats.P99)
	metrics["quic_latency_p999_ms"] = fmt.Sprintf("%.3f", latencyStats.P999)
	metrics["quic_latency_min_ms"] = fmt.Sprintf("%.3f", latencyStats.Min)
	metrics["quic_latency_max_ms"] = fmt.Sprintf("%.3f", latencyStats.Max)
	metrics["quic_latency_mean_ms"] = fmt.Sprintf("%.3f", latencyStats.Mean)
	metrics["quic_latency_count"] = fmt.Sprintf("%d", latencyStats.Count)
	
	// Jitter metrics
	metrics["quic_jitter_p50_ms"] = fmt.Sprintf("%.3f", jitterStats.P50)
	metrics["quic_jitter_p90_ms"] = fmt.Sprintf("%.3f", jitterStats.P90)
	metrics["quic_jitter_p95_ms"] = fmt.Sprintf("%.3f", jitterStats.P95)
	metrics["quic_jitter_p99_ms"] = fmt.Sprintf("%.3f", jitterStats.P99)
	metrics["quic_jitter_p999_ms"] = fmt.Sprintf("%.3f", jitterStats.P999)
	metrics["quic_jitter_min_ms"] = fmt.Sprintf("%.3f", jitterStats.Min)
	metrics["quic_jitter_max_ms"] = fmt.Sprintf("%.3f", jitterStats.Max)
	metrics["quic_jitter_mean_ms"] = fmt.Sprintf("%.3f", jitterStats.Mean)
	metrics["quic_jitter_count"] = fmt.Sprintf("%d", jitterStats.Count)
	
	// Handshake metrics
	metrics["quic_handshake_p50_ms"] = fmt.Sprintf("%.3f", handshakeStats.P50)
	metrics["quic_handshake_p90_ms"] = fmt.Sprintf("%.3f", handshakeStats.P90)
	metrics["quic_handshake_p95_ms"] = fmt.Sprintf("%.3f", handshakeStats.P95)
	metrics["quic_handshake_p99_ms"] = fmt.Sprintf("%.3f", handshakeStats.P99)
	metrics["quic_handshake_p999_ms"] = fmt.Sprintf("%.3f", handshakeStats.P999)
	metrics["quic_handshake_min_ms"] = fmt.Sprintf("%.3f", handshakeStats.Min)
	metrics["quic_handshake_max_ms"] = fmt.Sprintf("%.3f", handshakeStats.Max)
	metrics["quic_handshake_mean_ms"] = fmt.Sprintf("%.3f", handshakeStats.Mean)
	metrics["quic_handshake_count"] = fmt.Sprintf("%d", handshakeStats.Count)
	
	// Throughput metrics
	metrics["quic_throughput_p50_mbps"] = fmt.Sprintf("%.3f", throughputStats.P50)
	metrics["quic_throughput_p90_mbps"] = fmt.Sprintf("%.3f", throughputStats.P90)
	metrics["quic_throughput_p95_mbps"] = fmt.Sprintf("%.3f", throughputStats.P95)
	metrics["quic_throughput_p99_mbps"] = fmt.Sprintf("%.3f", throughputStats.P99)
	metrics["quic_throughput_p999_mbps"] = fmt.Sprintf("%.3f", throughputStats.P999)
	metrics["quic_throughput_min_mbps"] = fmt.Sprintf("%.3f", throughputStats.Min)
	metrics["quic_throughput_max_mbps"] = fmt.Sprintf("%.3f", throughputStats.Max)
	metrics["quic_throughput_mean_mbps"] = fmt.Sprintf("%.3f", throughputStats.Mean)
	metrics["quic_throughput_count"] = fmt.Sprintf("%d", throughputStats.Count)
	
	// Network metrics
	metrics["quic_packets_sent_total"] = fmt.Sprintf("%d", networkStats.PacketsSent)
	metrics["quic_packets_received_total"] = fmt.Sprintf("%d", networkStats.PacketsReceived)
	metrics["quic_packets_lost_total"] = fmt.Sprintf("%d", networkStats.PacketsLost)
	metrics["quic_packet_loss_percent"] = fmt.Sprintf("%.3f", networkStats.LossPercent)
	metrics["quic_bytes_sent_total"] = fmt.Sprintf("%d", networkStats.BytesSent)
	metrics["quic_bytes_received_total"] = fmt.Sprintf("%d", networkStats.BytesReceived)
	metrics["quic_retransmits_total"] = fmt.Sprintf("%d", networkStats.Retransmits)
	metrics["quic_errors_total"] = fmt.Sprintf("%d", networkStats.Errors)
	
	return metrics
}

// LatencyStats содержит статистику латенсии
type LatencyStats struct {
	P50   float64 `json:"p50_ms"`
	P90   float64 `json:"p90_ms"`
	P95   float64 `json:"p95_ms"`
	P99   float64 `json:"p99_ms"`
	P999  float64 `json:"p999_ms"`
	Min   float64 `json:"min_ms"`
	Max   float64 `json:"max_ms"`
	Mean  float64 `json:"mean_ms"`
	Count int64   `json:"count"`
}

// JitterStats содержит статистику джиттера
type JitterStats struct {
	P50   float64 `json:"p50_ms"`
	P90   float64 `json:"p90_ms"`
	P95   float64 `json:"p95_ms"`
	P99   float64 `json:"p99_ms"`
	P999  float64 `json:"p999_ms"`
	Min   float64 `json:"min_ms"`
	Max   float64 `json:"max_ms"`
	Mean  float64 `json:"mean_ms"`
	Count int64   `json:"count"`
}

// HandshakeStats содержит статистику handshake
type HandshakeStats struct {
	P50   float64 `json:"p50_ms"`
	P90   float64 `json:"p90_ms"`
	P95   float64 `json:"p95_ms"`
	P99   float64 `json:"p99_ms"`
	P999  float64 `json:"p999_ms"`
	Min   float64 `json:"min_ms"`
	Max   float64 `json:"max_ms"`
	Mean  float64 `json:"mean_ms"`
	Count int64   `json:"count"`
}

// ThroughputStats содержит статистику пропускной способности
type ThroughputStats struct {
	P50   float64 `json:"p50_mbps"`
	P90   float64 `json:"p90_mbps"`
	P95   float64 `json:"p95_mbps"`
	P99   float64 `json:"p99_mbps"`
	P999  float64 `json:"p999_mbps"`
	Min   float64 `json:"min_mbps"`
	Max   float64 `json:"max_mbps"`
	Mean  float64 `json:"mean_mbps"`
	Count int64   `json:"count"`
}

// NetworkStats содержит сетевую статистику
type NetworkStats struct {
	PacketsSent     int64   `json:"packets_sent"`
	PacketsReceived int64   `json:"packets_received"`
	PacketsLost     int64   `json:"packets_lost"`
	LossPercent     float64 `json:"loss_percent"`
	BytesSent       int64   `json:"bytes_sent"`
	BytesReceived   int64   `json:"bytes_received"`
	Retransmits     int64   `json:"retransmits"`
	Errors          int64   `json:"errors"`
}
