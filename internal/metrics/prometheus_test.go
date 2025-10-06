package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// newTestMetrics создает метрики для тестов с отдельным registry
func newTestMetrics() *PrometheusMetrics {
	registry := prometheus.NewRegistry()
	return NewPrometheusMetricsWithRegistry(registry)
}

func TestNewPrometheusMetrics(t *testing.T) {
	metrics := newTestMetrics()

	if metrics == nil {
		t.Fatal("NewPrometheusMetrics returned nil")
	}

	// Проверяем, что все метрики инициализированы
	if metrics.ConnectionsTotal == nil {
		t.Error("ConnectionsTotal is nil")
	}
	if metrics.StreamsTotal == nil {
		t.Error("StreamsTotal is nil")
	}
	if metrics.LatencyHistogram == nil {
		t.Error("LatencyHistogram is nil")
	}
	if metrics.ScenarioCounters == nil {
		t.Error("ScenarioCounters is nil")
	}
}

func TestRecordLatency(t *testing.T) {
	metrics := newTestMetrics()
	latency := 50 * time.Millisecond

	// Не должно паниковать
	metrics.RecordLatency(latency)
}

func TestRecordJitter(t *testing.T) {
	metrics := newTestMetrics()
	jitter := 5 * time.Millisecond

	// Не должно паниковать
	metrics.RecordJitter(jitter)
}

func TestRecordThroughput(t *testing.T) {
	metrics := newTestMetrics()
	throughput := 1024.0

	// Не должно паниковать
	metrics.RecordThroughput(throughput)
}

func TestIncrementConnections(t *testing.T) {
	metrics := newTestMetrics()

	// Не должно паниковать
	metrics.IncrementConnections()
	metrics.DecrementConnections()
}

func TestIncrementStreams(t *testing.T) {
	metrics := newTestMetrics()

	// Не должно паниковать
	metrics.IncrementStreams()
	metrics.DecrementStreams()
}

func TestAddBytes(t *testing.T) {
	metrics := newTestMetrics()

	// Не должно паниковать
	metrics.AddBytesSent(1024)
	metrics.AddBytesReceived(2048)
}

func TestIncrementCounters(t *testing.T) {
	metrics := newTestMetrics()

	// Не должно паниковать
	metrics.IncrementErrors()
	metrics.IncrementRetransmits()
	metrics.IncrementHandshakes()
	metrics.IncrementZeroRTT()
	metrics.IncrementOneRTT()
	metrics.IncrementSessionResumptions()
}

func TestSetGauges(t *testing.T) {
	metrics := newTestMetrics()

	// Не должно паниковать
	metrics.SetCurrentThroughput(1024.0)
	metrics.SetCurrentLatency(50 * time.Millisecond)
	metrics.SetPacketLossRate(0.01)
	metrics.SetConnectionDuration(30 * time.Second)
}

func TestRecordEvents(t *testing.T) {
	metrics := newTestMetrics()

	// Не должно паниковать
	metrics.RecordScenarioEvent("test", "conn1", "stream1", "success")
	metrics.RecordErrorEvent("timeout", "conn1", "stream1", "warning")
	metrics.RecordProtocolEvent("handshake", "conn1", "TLS1.3", "AES256-GCM")
	metrics.RecordScenarioDuration("test", "conn1", "success", 100*time.Millisecond)
	metrics.RecordNetworkLatency("wifi", "conn1", "us-east", 50*time.Millisecond)
}

func TestRecordHandshakeTime(t *testing.T) {
	metrics := newTestMetrics()
	handshakeTime := 200 * time.Millisecond

	// Не должно паниковать
	metrics.RecordHandshakeTime(handshakeTime)
}

func TestRecordRTT(t *testing.T) {
	metrics := newTestMetrics()
	rtt := 30 * time.Millisecond

	// Не должно паниковать
	metrics.RecordRTT(rtt)
}
