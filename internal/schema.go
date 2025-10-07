package internal

import (
	"encoding/json"
	"time"
)

// ReportSchema определяет JSON-схему для отчетов
type ReportSchema struct {
	Version     string                 `json:"version"`
	Timestamp   time.Time             `json:"timestamp"`
	TestConfig  TestConfigSchema      `json:"test_config"`
	Metrics     MetricsSchema         `json:"metrics"`
	TimeSeries  TimeSeriesSchema      `json:"time_series"`
	SLA         SLASchema             `json:"sla,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TestConfigSchema описывает конфигурацию теста
type TestConfigSchema struct {
	Mode         string        `json:"mode"`
	Address      string        `json:"address"`
	Connections  int           `json:"connections"`
	Streams      int           `json:"streams"`
	Duration     time.Duration `json:"duration"`
	PacketSize   int           `json:"packet_size"`
	Rate         int           `json:"rate"`
	Pattern      string        `json:"pattern"`
	NoTLS        bool          `json:"no_tls"`
	Prometheus   bool          `json:"prometheus"`
	EmulateLoss  float64       `json:"emulate_loss"`
	EmulateLatency time.Duration `json:"emulate_latency"`
	EmulateDup   float64       `json:"emulate_dup"`
	PprofAddr    string        `json:"pprof_addr,omitempty"`
}

// MetricsSchema описывает основные метрики
type MetricsSchema struct {
	Success              bool                    `json:"success"`
	Errors               int                     `json:"errors"`
	BytesSent            int64                   `json:"bytes_sent"`
	BytesReceived        int64                   `json:"bytes_received"`
	PacketsSent          int64                   `json:"packets_sent"`
	PacketsReceived      int64                   `json:"packets_received"`
	Latency              LatencyMetrics         `json:"latency"`
	Throughput           ThroughputMetrics      `json:"throughput"`
	PacketLoss           float64                 `json:"packet_loss"`
	Retransmits          int64                   `json:"retransmits"`
	TLSVersion           string                  `json:"tls_version"`
	CipherSuite          string                  `json:"cipher_suite"`
	SessionResumption    int64                   `json:"session_resumption_count"`
	ZeroRTT              int64                   `json:"zero_rtt_count"`
	OneRTT               int64                   `json:"one_rtt_count"`
	OutOfOrder           int64                   `json:"out_of_order_count"`
	FlowControlEvents    int64                   `json:"flow_control_events"`
	KeyUpdateEvents      int64                   `json:"key_update_events"`
	ErrorTypeCounts      map[string]int64        `json:"error_type_counts"`
	ConnectionMetrics    []ConnectionMetrics     `json:"connection_metrics,omitempty"`
	StreamMetrics        []StreamMetrics         `json:"stream_metrics,omitempty"`
}

// LatencyMetrics описывает метрики задержки
type LatencyMetrics struct {
	Average float64   `json:"average"`
	P50     float64   `json:"p50"`
	P95     float64   `json:"p95"`
	P99     float64   `json:"p99"`
	P999    float64   `json:"p999"`
	Jitter  float64   `json:"jitter"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	Values  []float64 `json:"values,omitempty"`
}

// ThroughputMetrics описывает метрики пропускной способности
type ThroughputMetrics struct {
	Average float64 `json:"average"`
	Peak    float64 `json:"peak"`
	Min     float64 `json:"min"`
	Current float64 `json:"current"`
}

// ConnectionMetrics описывает метрики соединения
type ConnectionMetrics struct {
	ConnectionID    int           `json:"connection_id"`
	HandshakeTime   time.Duration `json:"handshake_time"`
	BytesSent       int64         `json:"bytes_sent"`
	BytesReceived   int64         `json:"bytes_received"`
	PacketsSent     int64         `json:"packets_sent"`
	PacketsReceived int64         `json:"packets_received"`
	Retransmits     int64         `json:"retransmits"`
	Errors          int64         `json:"errors"`
	Latency         LatencyMetrics `json:"latency"`
}

// StreamMetrics описывает метрики потока
type StreamMetrics struct {
	ConnectionID int   `json:"connection_id"`
	StreamID     int   `json:"stream_id"`
	BytesSent    int64 `json:"bytes_sent"`
	BytesReceived int64 `json:"bytes_received"`
	Retransmits  int64 `json:"retransmits"`
	Errors       int64 `json:"errors"`
}

// TimeSeriesSchema описывает временные ряды
type TimeSeriesSchema struct {
	Latency      []TimeSeriesPoint `json:"latency"`
	Throughput   []TimeSeriesPoint `json:"throughput"`
	PacketLoss   []TimeSeriesPoint `json:"packet_loss"`
	Retransmits  []TimeSeriesPoint `json:"retransmits"`
	HandshakeTime []TimeSeriesPoint `json:"handshake_time"`
	Errors       []TimeSeriesPoint `json:"errors"`
}

// TimeSeriesPoint представляет точку временного ряда
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// SLASchema описывает SLA проверки
type SLASchema struct {
	Enabled     bool          `json:"enabled"`
	RttP95      time.Duration `json:"rtt_p95,omitempty"`
	Loss        float64       `json:"loss,omitempty"`
	Throughput  float64       `json:"throughput,omitempty"`
	Errors      int64         `json:"errors,omitempty"`
	Passed      bool          `json:"passed"`
	Violations  []SLAViolation `json:"violations,omitempty"`
}

// SLAViolation описывает нарушение SLA
type SLAViolation struct {
	Metric     string        `json:"metric"`
	Expected   interface{}   `json:"expected"`
	Actual     interface{}   `json:"actual"`
	Timestamp  time.Time     `json:"timestamp"`
	Severity   string        `json:"severity"` // "warning", "critical"
}

// CreateReportSchema создает схему отчета из конфигурации и метрик
func CreateReportSchema(cfg TestConfig, metrics map[string]interface{}) ReportSchema {
	return ReportSchema{
		Version:   "1.0.0",
		Timestamp: time.Now(),
		TestConfig: TestConfigSchema{
			Mode:          cfg.Mode,
			Address:       cfg.Addr,
			Connections:   cfg.Connections,
			Streams:       cfg.Streams,
			Duration:      cfg.Duration,
			PacketSize:    cfg.PacketSize,
			Rate:          cfg.Rate,
			Pattern:       cfg.Pattern,
			NoTLS:         cfg.NoTLS,
			Prometheus:    cfg.Prometheus,
			EmulateLoss:   cfg.EmulateLoss,
			EmulateLatency: cfg.EmulateLatency,
			EmulateDup:    cfg.EmulateDup,
			PprofAddr:     cfg.PprofAddr,
		},
		Metrics:    extractMetrics(metrics),
		TimeSeries: extractTimeSeries(metrics),
		SLA:        extractSLA(cfg, metrics),
		Metadata: map[string]interface{}{
			"go_version": "1.21",
			"quic_version": "0.40.0",
			"build_time": time.Now().Format(time.RFC3339),
		},
	}
}

// extractMetrics извлекает метрики из map
func extractMetrics(metrics map[string]interface{}) MetricsSchema {
	latencies, _ := metrics["Latencies"].([]float64)
	
	return MetricsSchema{
		Success:           getBool(metrics, "Success"),
		Errors:            getInt(metrics, "Errors"),
		BytesSent:         getInt64(metrics, "BytesSent"),
		BytesReceived:     getInt64(metrics, "BytesReceived"),
		PacketsSent:       getInt64(metrics, "PacketsSent"),
		PacketsReceived:   getInt64(metrics, "PacketsReceived"),
		Latency:           extractLatencyMetrics(latencies),
		Throughput:        extractThroughputMetrics(metrics),
		PacketLoss:        getFloat64(metrics, "PacketLoss"),
		Retransmits:       getInt64(metrics, "Retransmits"),
		TLSVersion:        getString(metrics, "TLSVersion"),
		CipherSuite:       getString(metrics, "CipherSuite"),
		SessionResumption: getInt64(metrics, "SessionResumptionCount"),
		ZeroRTT:           getInt64(metrics, "ZeroRTTCount"),
		OneRTT:            getInt64(metrics, "OneRTTCount"),
		OutOfOrder:        getInt64(metrics, "OutOfOrderCount"),
		FlowControlEvents: getInt64(metrics, "FlowControlEvents"),
		KeyUpdateEvents:   getInt64(metrics, "KeyUpdateEvents"),
		ErrorTypeCounts:   getStringInt64Map(metrics, "ErrorTypeCounts"),
	}
}

// extractLatencyMetrics извлекает метрики задержки
func extractLatencyMetrics(latencies []float64) LatencyMetrics {
	if len(latencies) == 0 {
		return LatencyMetrics{}
	}
	
	p50, p95, p99, p999 := calcPercentilesExtended(latencies)
	jitter := calcJitter(latencies)
	avg := avgLatency(latencies)
	
	min, max := latencies[0], latencies[0]
	for _, l := range latencies {
		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
	}
	
	return LatencyMetrics{
		Average: avg,
		P50:     p50,
		P95:     p95,
		P99:     p99,
		P999:    p999,
		Jitter:  jitter,
		Min:     min,
		Max:     max,
	}
}

// extractThroughputMetrics извлекает метрики пропускной способности
func extractThroughputMetrics(metrics map[string]interface{}) ThroughputMetrics {
	return ThroughputMetrics{
		Average: getFloat64(metrics, "ThroughputAverage"),
		Peak:    getFloat64(metrics, "ThroughputPeak"),
		Min:     getFloat64(metrics, "ThroughputMin"),
		Current: getFloat64(metrics, "ThroughputCurrent"),
	}
}

// extractTimeSeries извлекает временные ряды
func extractTimeSeries(metrics map[string]interface{}) TimeSeriesSchema {
	return TimeSeriesSchema{
		Latency:       extractTimeSeriesPoints(metrics, "TimeSeriesLatency"),
		Throughput:    extractTimeSeriesPoints(metrics, "TimeSeriesThroughput"),
		PacketLoss:    extractTimeSeriesPoints(metrics, "TimeSeriesPacketLoss"),
		Retransmits:   extractTimeSeriesPoints(metrics, "TimeSeriesRetransmits"),
		HandshakeTime: extractTimeSeriesPoints(metrics, "TimeSeriesHandshakeTime"),
		Errors:        extractTimeSeriesPoints(metrics, "TimeSeriesErrors"),
	}
}

// extractTimeSeriesPoints извлекает точки временного ряда
func extractTimeSeriesPoints(metrics map[string]interface{}, key string) []TimeSeriesPoint {
	points, ok := metrics[key].([]interface{})
	if !ok {
		return []TimeSeriesPoint{}
	}
	
	var result []TimeSeriesPoint
	for _, p := range points {
		if point, ok := p.(map[string]interface{}); ok {
			timestamp := time.Now() // В реальной реализации нужно извлекать из point
			value := getFloat64FromMap(point, "Value")
			result = append(result, TimeSeriesPoint{
				Timestamp: timestamp,
				Value:     value,
			})
		}
	}
	return result
}

// extractSLA извлекает SLA информацию
func extractSLA(cfg TestConfig, metrics map[string]interface{}) SLASchema {
	sla := SLASchema{
		Enabled: cfg.SlaRttP95 > 0 || cfg.SlaLoss > 0,
		RttP95:  cfg.SlaRttP95,
		Loss:    cfg.SlaLoss,
	}
	
	if sla.Enabled {
		// Проверяем SLA
		latencies, _ := metrics["Latencies"].([]float64)
		_, p95, _ := calcPercentiles(latencies)
		packetLoss := getFloat64(metrics, "PacketLoss")
		
		sla.Passed = true
		
		if cfg.SlaRttP95 > 0 && time.Duration(p95*float64(time.Millisecond)) > cfg.SlaRttP95 {
			sla.Passed = false
			sla.Violations = append(sla.Violations, SLAViolation{
				Metric:    "rtt_p95",
				Expected:  cfg.SlaRttP95,
				Actual:    time.Duration(p95 * float64(time.Millisecond)),
				Timestamp: time.Now(),
				Severity:  "critical",
			})
		}
		
		if cfg.SlaLoss > 0 && packetLoss > cfg.SlaLoss {
			sla.Passed = false
			sla.Violations = append(sla.Violations, SLAViolation{
				Metric:    "packet_loss",
				Expected:  cfg.SlaLoss,
				Actual:    packetLoss,
				Timestamp: time.Now(),
				Severity:  "critical",
			})
		}
	}
	
	return sla
}

// Вспомогательные функции для извлечения значений
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(int64); ok {
		return v
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringInt64Map(m map[string]interface{}, key string) map[string]int64 {
	if v, ok := m[key].(map[string]int64); ok {
		return v
	}
	return make(map[string]int64)
}

func getFloat64FromMap(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// ValidateReportSchema проверяет корректность схемы отчета
func ValidateReportSchema(schema ReportSchema) error {
	if schema.Version == "" {
		return json.NewEncoder(nil).Encode("version is required")
	}
	if schema.Timestamp.IsZero() {
		return json.NewEncoder(nil).Encode("timestamp is required")
	}
	return nil
}
