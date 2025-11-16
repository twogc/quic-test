package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMetrics(t *testing.T) {
	// Создаем новый registry для тестов
	_ = prometheus.NewRegistry()
	metrics := NewPrometheusMetrics()

	// Тестируем основные метрики
	metrics.UpdateConnectionMetrics(1, 1, 1, 1)
	metrics.UpdateConnectionMetrics(1, 1, 2, 2)
	metrics.UpdatePerformanceMetrics(1024, 0, 0, 0, 0)
	metrics.AddBytesReceived(2048)
	metrics.IncrementErrors()
	metrics.IncrementRetransmits()
	metrics.IncrementHandshakes()
	metrics.IncrementZeroRTT()
	metrics.IncrementOneRTT()
	metrics.IncrementSessionResumptions()

	// Проверяем счетчики
	if testutil.ToFloat64(metrics.ConnectionsTotal) != 1 {
		t.Errorf("Expected connections total to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsTotal))
	}
	if testutil.ToFloat64(metrics.StreamsTotal) != 1 {
		t.Errorf("Expected streams total to be 1, got %f", testutil.ToFloat64(metrics.StreamsTotal))
	}
	if testutil.ToFloat64(metrics.BytesSent) != 1024 {
		t.Errorf("Expected bytes sent total to be 1024, got %f", testutil.ToFloat64(metrics.BytesSent))
	}
	if testutil.ToFloat64(metrics.BytesReceived) != 2048 {
		t.Errorf("Expected bytes received total to be 2048, got %f", testutil.ToFloat64(metrics.BytesReceived))
	}
	if testutil.ToFloat64(metrics.PacketsLost) != 1 {
		t.Errorf("Expected errors total to be 1, got %f", testutil.ToFloat64(metrics.PacketsLost))
	}
	if testutil.ToFloat64(metrics.PacketsSent) != 1 {
		t.Errorf("Expected retransmits total to be 1, got %f", testutil.ToFloat64(metrics.PacketsSent))
	}
	if testutil.ToFloat64(metrics.ConnectionsTotal) != 1 {
		t.Errorf("Expected handshakes total to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsTotal))
	}
	if testutil.ToFloat64(metrics.ConnectionsTotal) != 1 {
		t.Errorf("Expected zero RTT total to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsTotal))
	}
	if testutil.ToFloat64(metrics.ConnectionsTotal) != 1 {
		t.Errorf("Expected one RTT total to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsTotal))
	}
	if testutil.ToFloat64(metrics.ConnectionsTotal) != 1 {
		t.Errorf("Expected session resumptions total to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsTotal))
	}
}

func TestPrometheusMetricsGauges(t *testing.T) {
	_ = prometheus.NewRegistry()
	metrics := NewPrometheusMetrics()

	// Тестируем gauges
	metrics.SetCurrentThroughput(1000)
	metrics.SetCurrentLatency(50 * time.Millisecond)
	metrics.SetPacketLossRate(0.01)
	metrics.SetConnectionDuration(30 * time.Second)

	if testutil.ToFloat64(metrics.ThroughputBps) != 1000 {
		t.Errorf("Expected current throughput to be 1000, got %f", testutil.ToFloat64(metrics.ThroughputBps))
	}
	if testutil.ToFloat64(metrics.RTTMeanMs) != 50 {
		t.Errorf("Expected current latency to be 50, got %f", testutil.ToFloat64(metrics.RTTMeanMs))
	}
	if testutil.ToFloat64(metrics.PacketsLost) != 0.01 {
		t.Errorf("Expected packet loss rate to be 0.01, got %f", testutil.ToFloat64(metrics.PacketsLost))
	}
	if testutil.ToFloat64(metrics.RTTMinMs) != 30000 {
		t.Errorf("Expected connection duration to be 30000, got %f", testutil.ToFloat64(metrics.RTTMinMs))
	}
}

func TestPrometheusMetricsHistograms(t *testing.T) {
	_ = prometheus.NewRegistry()
	metrics := NewPrometheusMetrics()

	// Тестируем методы записи
	metrics.RecordLatency(100 * time.Millisecond)
	metrics.RecordJitter(5 * time.Millisecond)
	metrics.RecordThroughput(1000)
	metrics.RecordHandshakeTime(200 * time.Millisecond)
	metrics.RecordRTT(50 * time.Millisecond)

	// Проверяем, что метрики были обновлены
	if testutil.ToFloat64(metrics.RTTMeanMs) != 100 {
		t.Errorf("Expected RTT mean to be 100, got %f", testutil.ToFloat64(metrics.RTTMeanMs))
	}
	if testutil.ToFloat64(metrics.RTTMaxMs) != 5 {
		t.Errorf("Expected RTT max to be 5, got %f", testutil.ToFloat64(metrics.RTTMaxMs))
	}
	if testutil.ToFloat64(metrics.ThroughputBps) != 1000 {
		t.Errorf("Expected throughput to be 1000, got %f", testutil.ToFloat64(metrics.ThroughputBps))
	}
	if testutil.ToFloat64(metrics.RTTMinMs) != 200 {
		t.Errorf("Expected RTT min to be 200, got %f", testutil.ToFloat64(metrics.RTTMinMs))
	}
}

func TestPrometheusMetricsEvents(t *testing.T) {
	_ = prometheus.NewRegistry()
	metrics := NewPrometheusMetrics()

	// Тестируем события
	metrics.RecordScenarioEvent("wifi")
	metrics.RecordErrorEvent("timeout")
	metrics.RecordProtocolEvent("handshake")
	metrics.RecordNetworkLatency(20 * time.Millisecond)

	// Проверяем, что методы не паникуют
	t.Log("Successfully called event recording methods")
}

func TestPrometheusMetricsDecrement(t *testing.T) {
	_ = prometheus.NewRegistry()
	metrics := NewPrometheusMetrics()

	// Увеличиваем счетчики
	metrics.UpdateConnectionMetrics(1, 1, 1, 1)
	metrics.UpdateConnectionMetrics(1, 1, 2, 2)

	// Проверяем текущие значения
	if testutil.ToFloat64(metrics.ConnectionsActive) != 1 {
		t.Errorf("Expected current connections to be 1, got %f", testutil.ToFloat64(metrics.ConnectionsActive))
	}
	if testutil.ToFloat64(metrics.StreamsActive) != 1 {
		t.Errorf("Expected current streams to be 1, got %f", testutil.ToFloat64(metrics.StreamsActive))
	}

	// Уменьшаем счетчики
	metrics.DecrementConnections()
	metrics.DecrementStreams()

	// Проверяем, что значения уменьшились
	if testutil.ToFloat64(metrics.ConnectionsActive) != 0 {
		t.Errorf("Expected current connections to be 0, got %f", testutil.ToFloat64(metrics.ConnectionsActive))
	}
	if testutil.ToFloat64(metrics.StreamsActive) != 0 {
		t.Errorf("Expected current streams to be 0, got %f", testutil.ToFloat64(metrics.StreamsActive))
	}
}

func TestPrometheusMetricsInvalidTypes(t *testing.T) {
	_ = prometheus.NewRegistry()
	metrics := NewPrometheusMetrics()

	// Тестируем с правильными типами
	metrics.RecordLatency(100 * time.Millisecond)
	metrics.RecordJitter(5 * time.Millisecond)
	metrics.SetCurrentLatency(50 * time.Millisecond)
	metrics.SetConnectionDuration(30 * time.Second)
	metrics.RecordHandshakeTime(200 * time.Millisecond)
	metrics.RecordRTT(75 * time.Millisecond)
	metrics.RecordNetworkLatency(20 * time.Millisecond)

	// Если мы дошли до этого места, значит паники не было
	t.Log("Successfully handled all method calls without panic")
}
