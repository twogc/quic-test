package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics содержит все метрики Prometheus для QUIC тестирования
type PrometheusMetrics struct {
	// Простая заглушка для совместимости
}

// NewPrometheusMetrics создает новый экземпляр метрик Prometheus
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{}
}

// NewPrometheusMetricsWithRegistry создает новый экземпляр метрик с указанным registry
func NewPrometheusMetricsWithRegistry(registry prometheus.Registerer) *PrometheusMetrics {
	return &PrometheusMetrics{}
}

// Заглушки для всех методов
func (m *PrometheusMetrics) RecordLatency(duration interface{}) {}
func (m *PrometheusMetrics) RecordJitter(duration interface{}) {}
func (m *PrometheusMetrics) RecordThroughput(throughput float64) {}
func (m *PrometheusMetrics) IncrementConnections() {}
func (m *PrometheusMetrics) DecrementConnections() {}
func (m *PrometheusMetrics) IncrementStreams() {}
func (m *PrometheusMetrics) DecrementStreams() {}
func (m *PrometheusMetrics) AddBytesSent(bytes int64) {}
func (m *PrometheusMetrics) AddBytesReceived(bytes int64) {}
func (m *PrometheusMetrics) IncrementErrors() {}
func (m *PrometheusMetrics) IncrementRetransmits() {}
func (m *PrometheusMetrics) IncrementHandshakes() {}
func (m *PrometheusMetrics) IncrementZeroRTT() {}
func (m *PrometheusMetrics) IncrementOneRTT() {}
func (m *PrometheusMetrics) IncrementSessionResumptions() {}
func (m *PrometheusMetrics) SetCurrentThroughput(throughput float64) {}
func (m *PrometheusMetrics) SetCurrentLatency(latency interface{}) {}
func (m *PrometheusMetrics) SetPacketLossRate(rate float64) {}
func (m *PrometheusMetrics) SetConnectionDuration(duration interface{}) {}
func (m *PrometheusMetrics) RecordScenarioEvent(scenario, connID, streamID, event string) {}
func (m *PrometheusMetrics) RecordErrorEvent(errorType, connID, streamID, severity string) {}
func (m *PrometheusMetrics) RecordProtocolEvent(event, connID, version, cipher string) {}
func (m *PrometheusMetrics) RecordScenarioDuration(scenario, connID, result string, duration interface{}) {}
func (m *PrometheusMetrics) RecordNetworkLatency(profile, connID, region string, latency interface{}) {}
func (m *PrometheusMetrics) RecordHandshakeTime(duration interface{}) {}
func (m *PrometheusMetrics) RecordRTT(duration interface{}) {}