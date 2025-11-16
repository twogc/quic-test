package internal

import (
	"errors"
	"reflect"
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
	BBRv3Metrics map[string]interface{} `json:"BBRv3Metrics,omitempty"` // BBRv3 specific metrics
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
	ThroughputMbps       float64                 `json:"throughput_mbps"`       // Throughput in Mbps (calculated correctly)
	GoodputMbps          float64                 `json:"goodput_mbps"`          // Goodput in Mbps (excluding retransmits)
	PacketLoss           float64                 `json:"packet_loss"`
	Retransmits          int64                   `json:"retransmits"`
	BufferbloatFactor    float64                 `json:"bufferbloat_factor"`    // (avg_rtt / min_rtt) - 1
	FairnessIndex        float64                 `json:"fairness_index"`         // Jain's fairness index
	FECPacketsSent       int64                   `json:"fec_packets_sent"`      // Количество отправленных FEC пакетов
	FECRedundancyBytes   int64                   `json:"fec_redundancy_bytes"` // Байты FEC redundancy
	FECRepairPacketsSent int64                   `json:"fec_repair_sent"`      // Redundancy packets sent (repair packets)
	FECRecovered         int64                   `json:"fec_recovered"`        // Packets recovered via FEC
	FECRecoveryEvents    int64                   `json:"fec_recovery_events"`  // События восстановления через FEC
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
	schema := ReportSchema{
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
	
	// Extract BBRv3 metrics if available
	if bbrv3Metrics, ok := metrics["BBRv3Metrics"].(map[string]interface{}); ok {
		schema.BBRv3Metrics = bbrv3Metrics
	}
	
	// Добавляем валидацию в метаданные
	if validationError := validateMetrics(metrics); validationError != "" {
		if schema.Metadata == nil {
			schema.Metadata = make(map[string]interface{})
		}
		schema.Metadata["validation_error"] = validationError
		schema.Metadata["validation_status"] = "invalid"
	} else {
		if schema.Metadata == nil {
			schema.Metadata = make(map[string]interface{})
		}
		schema.Metadata["validation_status"] = "valid"
	}
	
	return schema
}

// validateMetrics проверяет валидность метрик
func validateMetrics(metrics map[string]interface{}) string {
	bytesSent := getInt64(metrics, "BytesSent")
	if bytesSent == 0 {
		return "bytes_sent is zero - no data was sent. Check server connection and client configuration."
	}
	
	// Проверяем наличие latency данных
	latencies, ok := metrics["Latencies"].([]float64)
	if !ok || len(latencies) == 0 {
		return "no latency data - packets were not received or measured."
	}
	
	// Проверяем time_series (делаем проверку необязательной, так как данные могут быть в другом формате)
	// timeSeriesLatency, ok := metrics["TimeSeriesLatency"].([]interface{})
	// if !ok || len(timeSeriesLatency) == 0 {
	// 	return "time_series.latency is empty - time series data not collected."
	// }
	
	return ""
}

// extractMetrics извлекает метрики из map
func extractMetrics(metrics map[string]interface{}) MetricsSchema {
	latencies, _ := metrics["Latencies"].([]float64)
	
	// Извлекаем throughput_mbps (исправленный расчет)
	throughputMbps := getFloat64FromSchema(metrics, "ThroughputMbps")
	goodputMbps := getFloat64FromSchema(metrics, "GoodputMbps")
	
	// Если throughput_mbps = 0, но есть BytesSent и длительность, пересчитываем
	if throughputMbps == 0 {
		bytesSent := getInt64(metrics, "BytesSent")
		if bytesSent > 0 {
			// Пытаемся получить длительность из test_config или оценить по Timestamps
			// Для простоты используем фиксированную длительность из конфигурации, если доступна
			// Или используем время между первым и последним timestamp
			if timestamps, ok := metrics["Timestamps"].([]time.Time); ok && len(timestamps) > 1 {
				duration := timestamps[len(timestamps)-1].Sub(timestamps[0]).Seconds()
				if duration > 0 {
					throughputMbps = (float64(bytesSent) * 8) / (duration * 1_000_000)
				}
			}
		}
	}
	
	return MetricsSchema{
		Success:           getBool(metrics, "Success"),
		Errors:            getInt(metrics, "Errors"),
		BytesSent:         getInt64(metrics, "BytesSent"),
		BytesReceived:     getInt64(metrics, "BytesReceived"),
		PacketsSent:       getInt64(metrics, "PacketsSent"),
		PacketsReceived:   getInt64(metrics, "PacketsReceived"),
		Latency:           extractLatencyMetrics(latencies),
		Throughput:        extractThroughputMetrics(metrics),
		ThroughputMbps:    throughputMbps,
		GoodputMbps:       goodputMbps,
		PacketLoss:        getFloat64FromSchema(metrics, "PacketLoss"),
		Retransmits:       getInt64(metrics, "Retransmits"),
		BufferbloatFactor: getFloat64FromSchema(metrics, "BufferbloatFactor"),
		FairnessIndex:     getFloat64FromSchema(metrics, "FairnessIndex"),
		FECPacketsSent:    getInt64(metrics, "FECPacketsSent"),
		FECRedundancyBytes: getInt64(metrics, "FECRedundancyBytes"),
		FECRepairPacketsSent: getInt64(metrics, "FECRepairPacketsSent"),
		FECRecovered:      getInt64(metrics, "FECRecovered"),
		FECRecoveryEvents:  getInt64(metrics, "FECRecoveryEvents"),
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
	
	// Фильтруем выбросы (значения > 10 секунд считаются ошибками)
	filteredLatencies := make([]float64, 0, len(latencies))
	for _, l := range latencies {
		if l > 0 && l < 10000 { // Фильтруем значения > 10 секунд
			filteredLatencies = append(filteredLatencies, l)
		}
	}
	
	// Если после фильтрации осталось мало данных, используем оригинальный массив
	if len(filteredLatencies) < len(latencies)/2 {
		filteredLatencies = latencies
	}
	
	p50, p95, p99, p999 := calcPercentilesExtended(filteredLatencies)
	jitter := calcJitter(filteredLatencies)
	avg := avgLatency(filteredLatencies)
	
	min, max := filteredLatencies[0], filteredLatencies[0]
	for _, l := range filteredLatencies {
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
		Average: getFloat64FromSchema(metrics, "ThroughputAverage"),
		Peak:    getFloat64FromSchema(metrics, "ThroughputPeak"),
		Min:     getFloat64FromSchema(metrics, "ThroughputMin"),
		Current: getFloat64FromSchema(metrics, "ThroughputCurrent"),
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
	if val, ok := metrics[key]; ok {
		baseTime := time.Now()
		var result []TimeSeriesPoint

		// Проверяем если это срез интерфейсов (массив map'ов)
		if points, ok := val.([]interface{}); ok {
			for _, p := range points {
				var timestamp time.Time
				var value float64

				// Пытаемся извлечь данные из структуры (Time, Value)
				if point, ok := p.(map[string]interface{}); ok {
					// Time может быть float64 (секунды с начала теста)
					if timeFloat, ok := point["Time"].(float64); ok {
						timestamp = baseTime.Add(time.Duration(timeFloat * float64(time.Second)))
					} else if timeInt, ok := point["Time"].(int); ok {
						timestamp = baseTime.Add(time.Duration(timeInt) * time.Second)
					} else {
						timestamp = baseTime
					}

					// Value должно быть float64
					if val, ok := point["Value"].(float64); ok {
						value = val
					} else if val, ok := point["Value"].(int); ok {
						value = float64(val)
					}
				}

				result = append(result, TimeSeriesPoint{
					Timestamp: timestamp,
					Value:     value,
				})
			}
			return result
		}

		// Проверяем если это something с Time и Value (структура с этими полями)
		// Используем reflect для универсального доступа
		valueReflect := reflect.ValueOf(val)
		if valueReflect.Kind() == reflect.Slice {
			for i := 0; i < valueReflect.Len(); i++ {
				elem := valueReflect.Index(i)
				var timestamp time.Time
				var value float64

				// Пытаемся получить поля Time и Value
				timeField := elem.FieldByName("Time")
				valueField := elem.FieldByName("Value")

				if timeField.IsValid() && timeField.CanFloat() {
					timestamp = baseTime.Add(time.Duration(timeField.Float() * float64(time.Second)))
				} else {
					timestamp = baseTime
				}

				if valueField.IsValid() && valueField.CanFloat() {
					value = valueField.Float()
				}

				result = append(result, TimeSeriesPoint{
					Timestamp: timestamp,
					Value:     value,
				})
			}
			return result
		}
	}

	return []TimeSeriesPoint{}
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
		packetLoss := getFloat64FromSchema(metrics, "PacketLoss")
		
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
	// Try direct int
	if v, ok := m[key].(int); ok {
		return v
	}
	// Try int64
	if v, ok := m[key].(int64); ok {
		return int(v)
	}
	// Try float64 (JSON numbers can be float64)
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	// Try direct int64
	if v, ok := m[key].(int64); ok {
		return v
	}
	// Try int
	if v, ok := m[key].(int); ok {
		return int64(v)
	}
	// Try float64 (JSON numbers can be float64)
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func getFloat64FromSchema(m map[string]interface{}, key string) float64 {
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
		return errors.New("version is required")
	}
	if schema.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	return nil
}
