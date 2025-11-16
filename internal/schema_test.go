package internal

import (
	"testing"
	"time"
)

func TestCreateReportSchema(t *testing.T) {
	cfg := TestConfig{
		Mode:          "test",
		Addr:          ":9000",
		Connections:   2,
		Streams:       4,
		Duration:      30 * time.Second,
		PacketSize:    1200,
		Rate:          100,
		EmulateLoss:   0.02,
		EmulateLatency: 10 * time.Millisecond,
		EmulateDup:    0.01,
		SlaRttP95:     100 * time.Millisecond,
		SlaLoss:       0.01,
		SlaThroughput: 50.0,
		SlaErrors:     10,
	}

	metrics := map[string]interface{}{
		"Success":              true,
		"Errors":               5,
		"BytesSent":             int64(1024000),
		"BytesReceived":        int64(1020000),
		"PacketsSent":          int64(1000),
		"PacketsReceived":      int64(980),
		"Latencies":            []float64{10.5, 12.3, 8.7, 15.2, 9.8},
		"PacketLoss":           0.02,
		"Retransmits":          int64(20),
		"TLSVersion":           "TLS 1.3",
		"CipherSuite":          "TLS_AES_256_GCM_SHA384",
		"SessionResumptionCount": int64(1),
		"ZeroRTTCount":         int64(0),
		"OneRTTCount":          int64(1),
		"OutOfOrderCount":      int64(5),
		"FlowControlEvents":    int64(2),
		"KeyUpdateEvents":      int64(0),
		"ErrorTypeCounts":     map[string]int64{"timeout": 3, "connection": 2},
		"ThroughputAverage":    50.5,
		"ThroughputPeak":       75.2,
		"ThroughputMin":        25.1,
		"ThroughputCurrent":    48.3,
		"TimeSeriesLatency":    []interface{}{},
		"TimeSeriesThroughput": []interface{}{},
		"TimeSeriesPacketLoss": []interface{}{},
		"TimeSeriesRetransmits": []interface{}{},
		"TimeSeriesHandshakeTime": []interface{}{},
		"TimeSeriesErrors":     []interface{}{},
	}

	schema := CreateReportSchema(cfg, metrics)

	// Проверяем основные поля
	if schema.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", schema.Version)
	}

	if schema.TestConfig.Mode != cfg.Mode {
		t.Errorf("Expected mode %s, got %s", cfg.Mode, schema.TestConfig.Mode)
	}

	if schema.TestConfig.Connections != cfg.Connections {
		t.Errorf("Expected connections %d, got %d", cfg.Connections, schema.TestConfig.Connections)
	}

	// Проверяем метрики
	if !schema.Metrics.Success {
		t.Error("Expected success to be true")
	}

	if schema.Metrics.Errors != 5 {
		t.Errorf("Expected errors 5, got %d", schema.Metrics.Errors)
	}

	if schema.Metrics.BytesSent != 1024000 {
		t.Errorf("Expected bytes sent 1024000, got %d", schema.Metrics.BytesSent)
	}

	// Проверяем SLA
	if !schema.SLA.Enabled {
		t.Error("Expected SLA to be enabled")
	}

	if schema.SLA.RttP95 != cfg.SlaRttP95 {
		t.Errorf("Expected SLA RTT p95 %v, got %v", cfg.SlaRttP95, schema.SLA.RttP95)
	}

	if schema.SLA.Loss != cfg.SlaLoss {
		t.Errorf("Expected SLA loss %f, got %f", cfg.SlaLoss, schema.SLA.Loss)
	}
}

func TestValidateReportSchema(t *testing.T) {
	schema := ReportSchema{
		Version:   "1.0.0",
		Timestamp: time.Now(),
	}

	err := ValidateReportSchema(schema)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Тест с пустой версией
	schema.Version = ""
	err = ValidateReportSchema(schema)
	if err == nil {
		t.Error("Expected error for empty version")
	}

	// Тест с нулевым timestamp
	schema.Version = "1.0.0"
	schema.Timestamp = time.Time{}
	err = ValidateReportSchema(schema)
	if err == nil {
		t.Error("Expected error for zero timestamp")
	}
}

func TestExtractLatencyMetrics(t *testing.T) {
	latencies := []float64{10.5, 12.3, 8.7, 15.2, 9.8, 11.1, 13.4, 7.9, 14.6, 10.2}
	
	metrics := extractLatencyMetrics(latencies)
	
	if metrics.Average == 0 {
		t.Error("Expected non-zero average")
	}
	
	if metrics.P50 == 0 {
		t.Error("Expected non-zero P50")
	}
	
	if metrics.P95 == 0 {
		t.Error("Expected non-zero P95")
	}
	
	if metrics.P99 == 0 {
		t.Error("Expected non-zero P99")
	}
	
	if metrics.P999 == 0 {
		t.Error("Expected non-zero P999")
	}
	
	if metrics.Jitter == 0 {
		t.Error("Expected non-zero jitter")
	}
	
	if metrics.Min == 0 {
		t.Error("Expected non-zero min")
	}
	
	if metrics.Max == 0 {
		t.Error("Expected non-zero max")
	}
}

func TestExtractLatencyMetricsEmpty(t *testing.T) {
	latencies := []float64{}
	
	metrics := extractLatencyMetrics(latencies)
	
	if metrics.Average != 0 {
		t.Errorf("Expected zero average for empty latencies, got %f", metrics.Average)
	}
}
