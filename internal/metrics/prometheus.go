package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics содержит все метрики Prometheus для QUIC тестирования
type PrometheusMetrics struct {
	// Счетчики
	ConnectionsTotal     prometheus.Counter
	StreamsTotal         prometheus.Counter
	BytesSentTotal       prometheus.Counter
	BytesReceivedTotal   prometheus.Counter
	ErrorsTotal          prometheus.Counter
	RetransmitsTotal     prometheus.Counter
	HandshakesTotal      prometheus.Counter
	ZeroRTTTotal         prometheus.Counter
	OneRTTTotal          prometheus.Counter
	SessionResumptionsTotal prometheus.Counter

	// Гистограммы с адекватными бакетами
	LatencyHistogram     prometheus.Histogram
	JitterHistogram      prometheus.Histogram
	ThroughputHistogram  prometheus.Histogram
	HandshakeTimeHistogram prometheus.Histogram
	RTTHistogram         prometheus.Histogram

	// Gauges
	ActiveConnections    prometheus.Gauge
	ActiveStreams        prometheus.Gauge
	CurrentThroughput    prometheus.Gauge
	CurrentLatency       prometheus.Gauge
	PacketLossRate       prometheus.Gauge
	ConnectionDuration   prometheus.Gauge

	// Счетчики по сценариям
	ScenarioCounters     *prometheus.CounterVec
	ErrorCounters        *prometheus.CounterVec
	ProtocolCounters     *prometheus.CounterVec

	// Гистограммы по сценариям
	ScenarioHistograms   *prometheus.HistogramVec
	NetworkHistograms    *prometheus.HistogramVec

	mu sync.RWMutex
}

// NewPrometheusMetrics создает новый экземпляр метрик Prometheus
func NewPrometheusMetrics() *PrometheusMetrics {
	return NewPrometheusMetricsWithRegistry(prometheus.DefaultRegisterer)
}

// NewPrometheusMetricsWithRegistry создает новый экземпляр метрик с указанным registry
func NewPrometheusMetricsWithRegistry(registry prometheus.Registerer) *PrometheusMetrics {
	factory := promauto.With(registry)
	
	return &PrometheusMetrics{
		// Счетчики
		ConnectionsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_connections_total",
			Help: "Total number of QUIC connections established",
		}),
		StreamsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_streams_total",
			Help: "Total number of QUIC streams created",
		}),
		BytesSentTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_bytes_sent_total",
			Help: "Total bytes sent over QUIC connections",
		}),
		BytesReceivedTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_bytes_received_total",
			Help: "Total bytes received over QUIC connections",
		}),
		ErrorsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_errors_total",
			Help: "Total number of QUIC errors",
		}),
		RetransmitsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_retransmits_total",
			Help: "Total number of QUIC packet retransmissions",
		}),
		HandshakesTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_handshakes_total",
			Help: "Total number of QUIC handshakes completed",
		}),
		ZeroRTTTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_zero_rtt_total",
			Help: "Total number of 0-RTT connections",
		}),
		OneRTTTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_one_rtt_total",
			Help: "Total number of 1-RTT connections",
		}),
		SessionResumptionsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "quic_session_resumptions_total",
			Help: "Total number of QUIC session resumptions",
		}),

		// Гистограммы с адекватными бакетами для QUIC
		LatencyHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "quic_latency_seconds",
			Help:    "QUIC latency distribution",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}, // 1ms to 10s
		}),
		JitterHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "quic_jitter_seconds",
			Help:    "QUIC jitter distribution",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5}, // 0.1ms to 500ms
		}),
		ThroughputHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "quic_throughput_bytes_per_second",
			Help:    "QUIC throughput distribution",
			Buckets: []float64{1024, 10240, 102400, 1048576, 10485760, 104857600, 1073741824, 10737418240}, // 1KB to 10GB/s
		}),
		HandshakeTimeHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "quic_handshake_duration_seconds",
			Help:    "QUIC handshake duration distribution",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 25.0, 50.0}, // 10ms to 50s
		}),
		RTTHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "quic_rtt_seconds",
			Help:    "QUIC RTT distribution",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}, // 1ms to 10s
		}),

		// Gauges
		ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_active_connections",
			Help: "Number of active QUIC connections",
		}),
		ActiveStreams: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_active_streams",
			Help: "Number of active QUIC streams",
		}),
		CurrentThroughput: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_current_throughput_bytes_per_second",
			Help: "Current QUIC throughput in bytes per second",
		}),
		CurrentLatency: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_current_latency_seconds",
			Help: "Current QUIC latency in seconds",
		}),
		PacketLossRate: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_packet_loss_rate",
			Help: "Current QUIC packet loss rate (0.0 to 1.0)",
		}),
		ConnectionDuration: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_connection_duration_seconds",
			Help: "Current QUIC connection duration in seconds",
		}),

		// Векторные счетчики с фиксированными лейблами
		ScenarioCounters: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_scenario_total",
			Help: "Total operations by scenario",
		}, []string{"scenario", "connection_id", "stream_id", "result"}),
		ErrorCounters: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_error_total",
			Help: "Total errors by type",
		}, []string{"error_type", "connection_id", "stream_id", "severity"}),
		ProtocolCounters: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_protocol_total",
			Help: "Total protocol events",
		}, []string{"event_type", "connection_id", "tls_version", "cipher_suite"}),

		// Векторные гистограммы
		ScenarioHistograms: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "quic_scenario_duration_seconds",
			Help:    "Scenario execution duration",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0},
		}, []string{"scenario", "connection_id", "result"}),
		NetworkHistograms: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "quic_network_latency_seconds",
			Help:    "Network latency by profile",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}, []string{"network_profile", "connection_id", "region"}),
	}
}

// RecordLatency записывает задержку в гистограмму
func (pm *PrometheusMetrics) RecordLatency(latency time.Duration) {
	pm.LatencyHistogram.Observe(latency.Seconds())
}

// RecordJitter записывает джиттер в гистограмму
func (pm *PrometheusMetrics) RecordJitter(jitter time.Duration) {
	pm.JitterHistogram.Observe(jitter.Seconds())
}

// RecordThroughput записывает пропускную способность в гистограмму
func (pm *PrometheusMetrics) RecordThroughput(throughput float64) {
	pm.ThroughputHistogram.Observe(throughput)
}

// RecordHandshakeTime записывает время handshake в гистограмму
func (pm *PrometheusMetrics) RecordHandshakeTime(duration time.Duration) {
	pm.HandshakeTimeHistogram.Observe(duration.Seconds())
}

// RecordRTT записывает RTT в гистограмму
func (pm *PrometheusMetrics) RecordRTT(rtt time.Duration) {
	pm.RTTHistogram.Observe(rtt.Seconds())
}

// IncrementConnections увеличивает счетчик соединений
func (pm *PrometheusMetrics) IncrementConnections() {
	pm.ConnectionsTotal.Inc()
	pm.ActiveConnections.Inc()
}

// DecrementConnections уменьшает счетчик активных соединений
func (pm *PrometheusMetrics) DecrementConnections() {
	pm.ActiveConnections.Dec()
}

// IncrementStreams увеличивает счетчик потоков
func (pm *PrometheusMetrics) IncrementStreams() {
	pm.StreamsTotal.Inc()
	pm.ActiveStreams.Inc()
}

// DecrementStreams уменьшает счетчик активных потоков
func (pm *PrometheusMetrics) DecrementStreams() {
	pm.ActiveStreams.Dec()
}

// AddBytesSent добавляет отправленные байты
func (pm *PrometheusMetrics) AddBytesSent(bytes int64) {
	pm.BytesSentTotal.Add(float64(bytes))
}

// AddBytesReceived добавляет полученные байты
func (pm *PrometheusMetrics) AddBytesReceived(bytes int64) {
	pm.BytesReceivedTotal.Add(float64(bytes))
}

// IncrementErrors увеличивает счетчик ошибок
func (pm *PrometheusMetrics) IncrementErrors() {
	pm.ErrorsTotal.Inc()
}

// IncrementRetransmits увеличивает счетчик ретрансмиссий
func (pm *PrometheusMetrics) IncrementRetransmits() {
	pm.RetransmitsTotal.Inc()
}

// IncrementHandshakes увеличивает счетчик handshake
func (pm *PrometheusMetrics) IncrementHandshakes() {
	pm.HandshakesTotal.Inc()
}

// IncrementZeroRTT увеличивает счетчик 0-RTT
func (pm *PrometheusMetrics) IncrementZeroRTT() {
	pm.ZeroRTTTotal.Inc()
}

// IncrementOneRTT увеличивает счетчик 1-RTT
func (pm *PrometheusMetrics) IncrementOneRTT() {
	pm.OneRTTTotal.Inc()
}

// IncrementSessionResumptions увеличивает счетчик возобновлений сессии
func (pm *PrometheusMetrics) IncrementSessionResumptions() {
	pm.SessionResumptionsTotal.Inc()
}

// SetCurrentThroughput устанавливает текущую пропускную способность
func (pm *PrometheusMetrics) SetCurrentThroughput(throughput float64) {
	pm.CurrentThroughput.Set(throughput)
}

// SetCurrentLatency устанавливает текущую задержку
func (pm *PrometheusMetrics) SetCurrentLatency(latency time.Duration) {
	pm.CurrentLatency.Set(latency.Seconds())
}

// SetPacketLossRate устанавливает коэффициент потерь пакетов
func (pm *PrometheusMetrics) SetPacketLossRate(rate float64) {
	pm.PacketLossRate.Set(rate)
}

// SetConnectionDuration устанавливает длительность соединения
func (pm *PrometheusMetrics) SetConnectionDuration(duration time.Duration) {
	pm.ConnectionDuration.Set(duration.Seconds())
}

// RecordScenarioEvent записывает событие сценария
func (pm *PrometheusMetrics) RecordScenarioEvent(scenario, connectionID, streamID, result string) {
	pm.ScenarioCounters.WithLabelValues(scenario, connectionID, streamID, result).Inc()
}

// RecordErrorEvent записывает событие ошибки
func (pm *PrometheusMetrics) RecordErrorEvent(errorType, connectionID, streamID, severity string) {
	pm.ErrorCounters.WithLabelValues(errorType, connectionID, streamID, severity).Inc()
}

// RecordProtocolEvent записывает событие протокола
func (pm *PrometheusMetrics) RecordProtocolEvent(eventType, connectionID, tlsVersion, cipherSuite string) {
	pm.ProtocolCounters.WithLabelValues(eventType, connectionID, tlsVersion, cipherSuite).Inc()
}

// RecordScenarioDuration записывает длительность сценария
func (pm *PrometheusMetrics) RecordScenarioDuration(scenario, connectionID, result string, duration time.Duration) {
	pm.ScenarioHistograms.WithLabelValues(scenario, connectionID, result).Observe(duration.Seconds())
}

// RecordNetworkLatency записывает сетевую задержку по профилю
func (pm *PrometheusMetrics) RecordNetworkLatency(networkProfile, connectionID, region string, latency time.Duration) {
	pm.NetworkHistograms.WithLabelValues(networkProfile, connectionID, region).Observe(latency.Seconds())
}
