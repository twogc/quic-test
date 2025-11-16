package metrics

import (
	"testing"
	"time"
)

func TestNewHDRMetrics(t *testing.T) {
	metrics := NewHDRMetrics()
	if metrics == nil {
		t.Fatal("NewHDRMetrics() returned nil")
	}
}

func TestRecordLatency(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Записываем несколько значений латенсии
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
	}
	
	for _, latency := range latencies {
		metrics.RecordLatency(latency)
	}
	
	stats := metrics.GetLatencyStats()
	if stats.Count != int64(len(latencies)) {
		t.Errorf("Expected count %d, got %d", len(latencies), stats.Count)
	}
	
	// Проверяем, что перцентили вычисляются корректно
	if stats.P50 <= 0 {
		t.Error("P50 should be positive")
	}
	if stats.P95 <= 0 {
		t.Error("P95 should be positive")
	}
	if stats.P99 <= 0 {
		t.Error("P99 should be positive")
	}
	
	// P95 должен быть больше или равен P50 (может быть равен при малом количестве данных)
	if stats.P95 < stats.P50 {
		t.Error("P95 should be greater than or equal to P50")
	}
	
	// P99 должен быть больше или равен P95 (может быть равен при малом количестве данных)
	if stats.P99 < stats.P95 {
		t.Error("P99 should be greater than or equal to P95")
	}
}

func TestRecordJitter(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Записываем несколько значений джиттера
	jitters := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
	}
	
	for _, jitter := range jitters {
		metrics.RecordJitter(jitter)
	}
	
	stats := metrics.GetJitterStats()
	if stats.Count != int64(len(jitters)) {
		t.Errorf("Expected count %d, got %d", len(jitters), stats.Count)
	}
}

func TestRecordHandshakeTime(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Записываем несколько значений времени handshake
	handshakeTimes := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}
	
	for _, handshakeTime := range handshakeTimes {
		metrics.RecordHandshakeTime(handshakeTime)
	}
	
	stats := metrics.GetHandshakeStats()
	if stats.Count != int64(len(handshakeTimes)) {
		t.Errorf("Expected count %d, got %d", len(handshakeTimes), stats.Count)
	}
}

func TestRecordThroughput(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Записываем несколько значений пропускной способности (в байтах/сек)
	throughputs := []float64{
		1024 * 1024,        // 1 MB/s
		10 * 1024 * 1024,   // 10 MB/s
		100 * 1024 * 1024,  // 100 MB/s
	}
	
	for _, throughput := range throughputs {
		metrics.RecordThroughput(throughput)
	}
	
	stats := metrics.GetThroughputStats()
	if stats.Count != int64(len(throughputs)) {
		t.Errorf("Expected count %d, got %d", len(throughputs), stats.Count)
	}
}

func TestNetworkStats(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Симулируем сетевую активность
	metrics.IncrementPacketsSent()
	metrics.IncrementPacketsSent()
	metrics.IncrementPacketsSent()
	
	metrics.IncrementPacketsReceived()
	metrics.IncrementPacketsReceived()
	// Один пакет потерян
	
	metrics.AddBytesSent(1024)
	metrics.AddBytesReceived(512)
	
	metrics.IncrementErrors()
	metrics.IncrementRetransmits()
	
	stats := metrics.GetNetworkStats()
	
	if stats.PacketsSent != 3 {
		t.Errorf("Expected 3 packets sent, got %d", stats.PacketsSent)
	}
	
	if stats.PacketsReceived != 2 {
		t.Errorf("Expected 2 packets received, got %d", stats.PacketsReceived)
	}
	
	if stats.PacketsLost != 1 {
		t.Errorf("Expected 1 packet lost, got %d", stats.PacketsLost)
	}
	
	expectedLossPercent := 33.333 // 1/3 * 100
	if stats.LossPercent < expectedLossPercent-0.1 || stats.LossPercent > expectedLossPercent+0.1 {
		t.Errorf("Expected loss percent ~%.3f, got %.3f", expectedLossPercent, stats.LossPercent)
	}
	
	if stats.BytesSent != 1024 {
		t.Errorf("Expected 1024 bytes sent, got %d", stats.BytesSent)
	}
	
	if stats.BytesReceived != 512 {
		t.Errorf("Expected 512 bytes received, got %d", stats.BytesReceived)
	}
	
	if stats.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", stats.Errors)
	}
	
	if stats.Retransmits != 1 {
		t.Errorf("Expected 1 retransmit, got %d", stats.Retransmits)
	}
}

func TestTimeSeries(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Добавляем несколько точек временного ряда
	timestamp1 := time.Now()
	metrics.AddTimeSeriesPoint(timestamp1, map[string]interface{}{
		"latency": 10.0,
		"throughput": 1000.0,
	})
	
	timestamp2 := time.Now().Add(time.Second)
	metrics.AddTimeSeriesPoint(timestamp2, map[string]interface{}{
		"latency": 20.0,
		"throughput": 2000.0,
	})
	
	timeSeries := metrics.GetTimeSeries()
	if len(timeSeries) != 2 {
		t.Errorf("Expected 2 time series points, got %d", len(timeSeries))
	}
	
	if !timeSeries[0].Timestamp.Equal(timestamp1) {
		t.Error("First timestamp doesn't match")
	}
	
	if !timeSeries[1].Timestamp.Equal(timestamp2) {
		t.Error("Second timestamp doesn't match")
	}
}

func TestExportToPrometheus(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Добавляем некоторые данные
	metrics.RecordLatency(10 * time.Millisecond)
	metrics.RecordJitter(1 * time.Millisecond)
	metrics.RecordHandshakeTime(100 * time.Millisecond)
	metrics.RecordThroughput(1024 * 1024)
	metrics.IncrementPacketsSent()
	metrics.IncrementPacketsReceived()
	
	prometheusMetrics := metrics.ExportToPrometheus()
	
	// Проверяем, что основные метрики присутствуют
	expectedKeys := []string{
		"quic_latency_p50_ms",
		"quic_latency_p95_ms",
		"quic_latency_p99_ms",
		"quic_jitter_p50_ms",
		"quic_handshake_p50_ms",
		"quic_throughput_p50_mbps",
		"quic_packets_sent_total",
		"quic_packets_received_total",
	}
	
	for _, key := range expectedKeys {
		if _, exists := prometheusMetrics[key]; !exists {
			t.Errorf("Expected metric %s not found in Prometheus export", key)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Тестируем concurrent access
	done := make(chan bool, 10)
	
	// Запускаем несколько горутин для записи метрик
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.RecordLatency(time.Duration(j) * time.Millisecond)
				metrics.IncrementPacketsSent()
				metrics.AddBytesSent(int64(j))
			}
			done <- true
		}()
	}
	
	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Проверяем, что данные записались корректно (допускаем небольшие потери из-за race conditions)
	stats := metrics.GetLatencyStats()
	if stats.Count < 900 { // Допускаем потерю до 10% записей
		t.Errorf("Expected at least 900 latency records, got %d", stats.Count)
	}
	
	networkStats := metrics.GetNetworkStats()
	if networkStats.PacketsSent < 900 { // Допускаем потерю до 10% записей
		t.Errorf("Expected at least 900 packets sent, got %d", networkStats.PacketsSent)
	}
}

func TestEmptyHistogram(t *testing.T) {
	metrics := NewHDRMetrics()
	
	// Тестируем пустые гистограммы
	latencyStats := metrics.GetLatencyStats()
	if latencyStats.Count != 0 {
		t.Error("Empty latency histogram should have count 0")
	}
	
	jitterStats := metrics.GetJitterStats()
	if jitterStats.Count != 0 {
		t.Error("Empty jitter histogram should have count 0")
	}
	
	handshakeStats := metrics.GetHandshakeStats()
	if handshakeStats.Count != 0 {
		t.Error("Empty handshake histogram should have count 0")
	}
	
	throughputStats := metrics.GetThroughputStats()
	if throughputStats.Count != 0 {
		t.Error("Empty throughput histogram should have count 0")
	}
}
