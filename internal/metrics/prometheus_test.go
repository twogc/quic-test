package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMetrics(t *testing.T) {
	// Создаем новый registry для тестов
	registry := prometheus.NewRegistry()
	metrics := NewPrometheusMetricsWithRegistry(registry)

	// Тестируем основные метрики
	metrics.IncrementConnections()
	metrics.IncrementStreams()
	metrics.AddBytesSent(1024)
	metrics.AddBytesReceived(2048)
	metrics.IncrementErrors()
	metrics.IncrementRetransmits()
	metrics.IncrementHandshakes()
	metrics.IncrementZeroRTT()
	metrics.IncrementOneRTT()
	metrics.IncrementSessionResumptions()

	// Проверяем счетчики
	if testutil.ToFloat64(metrics.connectionsTotal) != 1 {
		t.Errorf("Expected connections total to be 1, got %f", testutil.ToFloat64(metrics.connectionsTotal))
	}
	if testutil.ToFloat64(metrics.streamsTotal) != 1 {
		t.Errorf("Expected streams total to be 1, got %f", testutil.ToFloat64(metrics.streamsTotal))
	}
	if testutil.ToFloat64(metrics.bytesSentTotal) != 1024 {
		t.Errorf("Expected bytes sent total to be 1024, got %f", testutil.ToFloat64(metrics.bytesSentTotal))
	}
	if testutil.ToFloat64(metrics.bytesReceivedTotal) != 2048 {
		t.Errorf("Expected bytes received total to be 2048, got %f", testutil.ToFloat64(metrics.bytesReceivedTotal))
	}
	if testutil.ToFloat64(metrics.errorsTotal) != 1 {
		t.Errorf("Expected errors total to be 1, got %f", testutil.ToFloat64(metrics.errorsTotal))
	}
	if testutil.ToFloat64(metrics.retransmitsTotal) != 1 {
		t.Errorf("Expected retransmits total to be 1, got %f", testutil.ToFloat64(metrics.retransmitsTotal))
	}
	if testutil.ToFloat64(metrics.handshakesTotal) != 1 {
		t.Errorf("Expected handshakes total to be 1, got %f", testutil.ToFloat64(metrics.handshakesTotal))
	}
	if testutil.ToFloat64(metrics.zeroRTTTotal) != 1 {
		t.Errorf("Expected zero RTT total to be 1, got %f", testutil.ToFloat64(metrics.zeroRTTTotal))
	}
	if testutil.ToFloat64(metrics.oneRTTTotal) != 1 {
		t.Errorf("Expected one RTT total to be 1, got %f", testutil.ToFloat64(metrics.oneRTTTotal))
	}
	if testutil.ToFloat64(metrics.sessionResumptionsTotal) != 1 {
		t.Errorf("Expected session resumptions total to be 1, got %f", testutil.ToFloat64(metrics.sessionResumptionsTotal))
	}
}

func TestPrometheusMetricsGauges(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewPrometheusMetricsWithRegistry(registry)

	// Тестируем gauges
	metrics.SetCurrentThroughput(1000.5)
	metrics.SetCurrentLatency(50 * time.Millisecond)
	metrics.SetPacketLossRate(0.01)
	metrics.SetConnectionDuration(30 * time.Second)

	if testutil.ToFloat64(metrics.currentThroughput) != 1000.5 {
		t.Errorf("Expected current throughput to be 1000.5, got %f", testutil.ToFloat64(metrics.currentThroughput))
	}
	if testutil.ToFloat64(metrics.currentLatency) != 0.05 {
		t.Errorf("Expected current latency to be 0.05, got %f", testutil.ToFloat64(metrics.currentLatency))
	}
	if testutil.ToFloat64(metrics.packetLossRate) != 0.01 {
		t.Errorf("Expected packet loss rate to be 0.01, got %f", testutil.ToFloat64(metrics.packetLossRate))
	}
	if testutil.ToFloat64(metrics.connectionDuration) != 30 {
		t.Errorf("Expected connection duration to be 30, got %f", testutil.ToFloat64(metrics.connectionDuration))
	}
}

func TestPrometheusMetricsHistograms(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewPrometheusMetricsWithRegistry(registry)

	// Тестируем гистограммы
	metrics.RecordLatency(100 * time.Millisecond)
	metrics.RecordJitter(5 * time.Millisecond)
	metrics.RecordThroughput(1000.0)
	metrics.RecordHandshakeTime(200 * time.Millisecond)
	metrics.RecordRTT(50 * time.Millisecond)

	// Проверяем, что гистограммы были обновлены
	latencyCount := testutil.CollectAndCount(metrics.latencyHistogram)
	if latencyCount == 0 {
		t.Error("Expected latency histogram to have observations")
	}

	jitterCount := testutil.CollectAndCount(metrics.jitterHistogram)
	if jitterCount == 0 {
		t.Error("Expected jitter histogram to have observations")
	}

	throughputCount := testutil.CollectAndCount(metrics.throughputHistogram)
	if throughputCount == 0 {
		t.Error("Expected throughput histogram to have observations")
	}

	handshakeCount := testutil.CollectAndCount(metrics.handshakeHistogram)
	if handshakeCount == 0 {
		t.Error("Expected handshake histogram to have observations")
	}
}

func TestPrometheusMetricsEvents(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewPrometheusMetricsWithRegistry(registry)

	// Тестируем события
	metrics.RecordScenarioEvent("wifi", "conn1", "stream1", "start")
	metrics.RecordErrorEvent("timeout", "conn1", "stream1", "warning")
	metrics.RecordProtocolEvent("handshake", "conn1", "v1", "TLS_AES_256_GCM_SHA384")
	metrics.RecordNetworkLatency("wifi", "conn1", "us-east", 20*time.Millisecond)

	// Проверяем, что события были зарегистрированы
	scenarioCount := testutil.CollectAndCount(metrics.scenarioEvents)
	if scenarioCount == 0 {
		t.Error("Expected scenario events to have observations")
	}

	errorCount := testutil.CollectAndCount(metrics.errorEvents)
	if errorCount == 0 {
		t.Error("Expected error events to have observations")
	}

	protocolCount := testutil.CollectAndCount(metrics.protocolEvents)
	if protocolCount == 0 {
		t.Error("Expected protocol events to have observations")
	}

	networkCount := testutil.CollectAndCount(metrics.networkLatency)
	if networkCount == 0 {
		t.Error("Expected network latency to have observations")
	}
}

func TestPrometheusMetricsDecrement(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewPrometheusMetricsWithRegistry(registry)

	// Увеличиваем счетчики
	metrics.IncrementConnections()
	metrics.IncrementStreams()

	// Проверяем текущие значения
	if testutil.ToFloat64(metrics.currentConnections) != 1 {
		t.Errorf("Expected current connections to be 1, got %f", testutil.ToFloat64(metrics.currentConnections))
	}
	if testutil.ToFloat64(metrics.currentStreams) != 1 {
		t.Errorf("Expected current streams to be 1, got %f", testutil.ToFloat64(metrics.currentStreams))
	}

	// Уменьшаем счетчики
	metrics.DecrementConnections()
	metrics.DecrementStreams()

	// Проверяем, что значения уменьшились
	if testutil.ToFloat64(metrics.currentConnections) != 0 {
		t.Errorf("Expected current connections to be 0, got %f", testutil.ToFloat64(metrics.currentConnections))
	}
	if testutil.ToFloat64(metrics.currentStreams) != 0 {
		t.Errorf("Expected current streams to be 0, got %f", testutil.ToFloat64(metrics.currentStreams))
	}
}

func TestPrometheusMetricsInvalidTypes(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewPrometheusMetricsWithRegistry(registry)

	// Тестируем с неверными типами - не должно паниковать
	metrics.RecordLatency("invalid")
	metrics.RecordJitter(123) // int вместо Duration
	metrics.SetCurrentLatency("invalid")
	metrics.SetConnectionDuration(123) // int вместо Duration
	metrics.RecordHandshakeTime("invalid")
	metrics.RecordRTT(123) // int вместо Duration
	metrics.RecordNetworkLatency("profile", "conn", "region", "invalid")

	// Если мы дошли до этого места, значит паники не было
	t.Log("Successfully handled invalid types without panic")
}
