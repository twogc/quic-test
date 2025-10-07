package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics содержит все метрики Prometheus для QUIC тестирования
type PrometheusMetrics struct {
	// Гистограммы
	latencyHistogram     *prometheus.HistogramVec
	jitterHistogram      *prometheus.HistogramVec
	handshakeHistogram   *prometheus.HistogramVec
	throughputHistogram  *prometheus.HistogramVec
	
	// Счетчики
	connectionsTotal     prometheus.Counter
	streamsTotal         prometheus.Counter
	bytesSentTotal       prometheus.Counter
	bytesReceivedTotal   prometheus.Counter
	errorsTotal          prometheus.Counter
	retransmitsTotal      prometheus.Counter
	handshakesTotal      prometheus.Counter
	zeroRTTTotal         prometheus.Counter
	oneRTTTotal          prometheus.Counter
	sessionResumptionsTotal prometheus.Counter
	
	// Gauges
	currentConnections   prometheus.Gauge
	currentStreams       prometheus.Gauge
	currentThroughput    prometheus.Gauge
	currentLatency       prometheus.Gauge
	packetLossRate       prometheus.Gauge
	connectionDuration   prometheus.Gauge
	
	// События
	scenarioEvents       *prometheus.CounterVec
	errorEvents          *prometheus.CounterVec
	protocolEvents       *prometheus.CounterVec
	networkLatency       *prometheus.HistogramVec
}

// NewPrometheusMetrics создает новый экземпляр метрик Prometheus
func NewPrometheusMetrics() *PrometheusMetrics {
	return NewPrometheusMetricsWithRegistry(prometheus.DefaultRegisterer)
}

// NewPrometheusMetricsWithRegistry создает новый экземпляр метрик с указанным registry
func NewPrometheusMetricsWithRegistry(registry prometheus.Registerer) *PrometheusMetrics {
	m := &PrometheusMetrics{
		// Гистограммы
		latencyHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "quic_latency_seconds",
				Help:    "QUIC request latency in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to 32s
			},
			[]string{"connection_id", "stream_id"},
		),
		jitterHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "quic_jitter_seconds",
				Help:    "QUIC jitter in seconds",
				Buckets: prometheus.ExponentialBuckets(0.0001, 2, 12), // 0.1ms to 400ms
			},
			[]string{"connection_id"},
		),
		handshakeHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "quic_handshake_seconds",
				Help:    "QUIC handshake duration in seconds",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms to 40s
			},
			[]string{"connection_id", "version"},
		),
		throughputHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "quic_throughput_bytes_per_second",
				Help:    "QUIC throughput in bytes per second",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 20), // 1KB to 1GB/s
			},
			[]string{"connection_id"},
		),
		
		// Счетчики
		connectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_connections_total",
			Help: "Total number of QUIC connections",
		}),
		streamsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_streams_total",
			Help: "Total number of QUIC streams",
		}),
		bytesSentTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_bytes_sent_total",
			Help: "Total bytes sent over QUIC",
		}),
		bytesReceivedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_bytes_received_total",
			Help: "Total bytes received over QUIC",
		}),
		errorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_errors_total",
			Help: "Total number of QUIC errors",
		}),
		retransmitsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_retransmits_total",
			Help: "Total number of QUIC retransmits",
		}),
		handshakesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_handshakes_total",
			Help: "Total number of QUIC handshakes",
		}),
		zeroRTTTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_zero_rtt_total",
			Help: "Total number of 0-RTT connections",
		}),
		oneRTTTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_one_rtt_total",
			Help: "Total number of 1-RTT connections",
		}),
		sessionResumptionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "quic_session_resumptions_total",
			Help: "Total number of session resumptions",
		}),
		
		// Gauges
		currentConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "quic_connections_current",
			Help: "Current number of QUIC connections",
		}),
		currentStreams: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "quic_streams_current",
			Help: "Current number of QUIC streams",
		}),
		currentThroughput: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "quic_throughput_current_bytes_per_second",
			Help: "Current QUIC throughput in bytes per second",
		}),
		currentLatency: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "quic_latency_current_seconds",
			Help: "Current QUIC latency in seconds",
		}),
		packetLossRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "quic_packet_loss_rate",
			Help: "Current QUIC packet loss rate (0-1)",
		}),
		connectionDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "quic_connection_duration_seconds",
			Help: "Current QUIC connection duration in seconds",
		}),
		
		// События
		scenarioEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "quic_scenario_events_total",
				Help: "Total number of scenario events",
			},
			[]string{"scenario", "connection_id", "stream_id", "event"},
		),
		errorEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "quic_error_events_total",
				Help: "Total number of error events",
			},
			[]string{"error_type", "connection_id", "stream_id", "severity"},
		),
		protocolEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "quic_protocol_events_total",
				Help: "Total number of protocol events",
			},
			[]string{"event", "connection_id", "version", "cipher"},
		),
		networkLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "quic_network_latency_seconds",
				Help:    "Network latency in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
			},
			[]string{"profile", "connection_id", "region"},
		),
	}
	
	// Регистрируем все метрики
	registry.MustRegister(
		m.latencyHistogram, m.jitterHistogram, m.handshakeHistogram, m.throughputHistogram,
		m.connectionsTotal, m.streamsTotal, m.bytesSentTotal, m.bytesReceivedTotal,
		m.errorsTotal, m.retransmitsTotal, m.handshakesTotal, m.zeroRTTTotal,
		m.oneRTTTotal, m.sessionResumptionsTotal,
		m.currentConnections, m.currentStreams, m.currentThroughput, m.currentLatency,
		m.packetLossRate, m.connectionDuration,
		m.scenarioEvents, m.errorEvents, m.protocolEvents, m.networkLatency,
	)
	
	return m
}

// Реализации методов для записи метрик

// RecordLatency записывает задержку
func (m *PrometheusMetrics) RecordLatency(duration interface{}) {
	if d, ok := duration.(time.Duration); ok {
		m.latencyHistogram.WithLabelValues("", "").Observe(d.Seconds())
	}
}

// RecordJitter записывает джиттер
func (m *PrometheusMetrics) RecordJitter(duration interface{}) {
	if d, ok := duration.(time.Duration); ok {
		m.jitterHistogram.WithLabelValues("").Observe(d.Seconds())
	}
}

// RecordThroughput записывает пропускную способность
func (m *PrometheusMetrics) RecordThroughput(throughput float64) {
	m.throughputHistogram.WithLabelValues("").Observe(throughput)
}

// IncrementConnections увеличивает счетчик соединений
func (m *PrometheusMetrics) IncrementConnections() {
	m.connectionsTotal.Inc()
	m.currentConnections.Inc()
}

// DecrementConnections уменьшает счетчик соединений
func (m *PrometheusMetrics) DecrementConnections() {
	m.currentConnections.Dec()
}

// IncrementStreams увеличивает счетчик потоков
func (m *PrometheusMetrics) IncrementStreams() {
	m.streamsTotal.Inc()
	m.currentStreams.Inc()
}

// DecrementStreams уменьшает счетчик потоков
func (m *PrometheusMetrics) DecrementStreams() {
	m.currentStreams.Dec()
}

// AddBytesSent добавляет отправленные байты
func (m *PrometheusMetrics) AddBytesSent(bytes int64) {
	m.bytesSentTotal.Add(float64(bytes))
}

// AddBytesReceived добавляет полученные байты
func (m *PrometheusMetrics) AddBytesReceived(bytes int64) {
	m.bytesReceivedTotal.Add(float64(bytes))
}

// IncrementErrors увеличивает счетчик ошибок
func (m *PrometheusMetrics) IncrementErrors() {
	m.errorsTotal.Inc()
}

// IncrementRetransmits увеличивает счетчик ретрансмиссий
func (m *PrometheusMetrics) IncrementRetransmits() {
	m.retransmitsTotal.Inc()
}

// IncrementHandshakes увеличивает счетчик handshake
func (m *PrometheusMetrics) IncrementHandshakes() {
	m.handshakesTotal.Inc()
}

// IncrementZeroRTT увеличивает счетчик 0-RTT
func (m *PrometheusMetrics) IncrementZeroRTT() {
	m.zeroRTTTotal.Inc()
}

// IncrementOneRTT увеличивает счетчик 1-RTT
func (m *PrometheusMetrics) IncrementOneRTT() {
	m.oneRTTTotal.Inc()
}

// IncrementSessionResumptions увеличивает счетчик возобновлений сессии
func (m *PrometheusMetrics) IncrementSessionResumptions() {
	m.sessionResumptionsTotal.Inc()
}

// SetCurrentThroughput устанавливает текущую пропускную способность
func (m *PrometheusMetrics) SetCurrentThroughput(throughput float64) {
	m.currentThroughput.Set(throughput)
}

// SetCurrentLatency устанавливает текущую задержку
func (m *PrometheusMetrics) SetCurrentLatency(latency interface{}) {
	if d, ok := latency.(time.Duration); ok {
		m.currentLatency.Set(d.Seconds())
	}
}

// SetPacketLossRate устанавливает коэффициент потери пакетов
func (m *PrometheusMetrics) SetPacketLossRate(rate float64) {
	m.packetLossRate.Set(rate)
}

// SetConnectionDuration устанавливает длительность соединения
func (m *PrometheusMetrics) SetConnectionDuration(duration interface{}) {
	if d, ok := duration.(time.Duration); ok {
		m.connectionDuration.Set(d.Seconds())
	}
}

// RecordScenarioEvent записывает событие сценария
func (m *PrometheusMetrics) RecordScenarioEvent(scenario, connID, streamID, event string) {
	m.scenarioEvents.WithLabelValues(scenario, connID, streamID, event).Inc()
}

// RecordErrorEvent записывает событие ошибки
func (m *PrometheusMetrics) RecordErrorEvent(errorType, connID, streamID, severity string) {
	m.errorEvents.WithLabelValues(errorType, connID, streamID, severity).Inc()
}

// RecordProtocolEvent записывает событие протокола
func (m *PrometheusMetrics) RecordProtocolEvent(event, connID, version, cipher string) {
	m.protocolEvents.WithLabelValues(event, connID, version, cipher).Inc()
}

// RecordScenarioDuration записывает длительность сценария
func (m *PrometheusMetrics) RecordScenarioDuration(scenario, connID, result string, duration interface{}) {
	if d, ok := duration.(time.Duration); ok {
		m.scenarioEvents.WithLabelValues(scenario, connID, "", "duration").Add(d.Seconds())
	}
}

// RecordNetworkLatency записывает сетевую задержку
func (m *PrometheusMetrics) RecordNetworkLatency(profile, connID, region string, latency interface{}) {
	if d, ok := latency.(time.Duration); ok {
		m.networkLatency.WithLabelValues(profile, connID, region).Observe(d.Seconds())
	}
}

// RecordHandshakeTime записывает время handshake
func (m *PrometheusMetrics) RecordHandshakeTime(duration interface{}) {
	if d, ok := duration.(time.Duration); ok {
		m.handshakeHistogram.WithLabelValues("", "").Observe(d.Seconds())
	}
}

// RecordRTT записывает RTT
func (m *PrometheusMetrics) RecordRTT(duration interface{}) {
	if d, ok := duration.(time.Duration); ok {
		m.latencyHistogram.WithLabelValues("", "rtt").Observe(d.Seconds())
	}
}