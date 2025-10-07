package client

import (
	"sync"
	"time"

	"quic-test/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AdvancedPrometheusExporter предоставляет продвинутые метрики Prometheus для клиента
type AdvancedPrometheusExporter struct {
	// Основные метрики
	metrics *metrics.PrometheusMetrics

	// Дополнительные метрики клиента
	clientMetrics *ClientMetrics

	// Счетчики по типам тестов
	testTypeCounters *prometheus.CounterVec

	// Гистограммы по типам данных
	dataPatternHistograms *prometheus.HistogramVec

	// Метрики по соединениям
	connectionMetrics *prometheus.GaugeVec

	// Метрики по потокам
	streamMetrics *prometheus.GaugeVec

	mu sync.RWMutex
}

// ClientMetrics содержит метрики клиента
type ClientMetrics struct {
	TestType        string
	DataPattern     string
	ConnectionCount int
	StreamCount     int
	StartTime       time.Time
	LastUpdate      time.Time
}

// NewAdvancedPrometheusExporter создает новый экспортер метрик
func NewAdvancedPrometheusExporter() *AdvancedPrometheusExporter {
	return &AdvancedPrometheusExporter{
		metrics: metrics.NewPrometheusMetrics(),
		clientMetrics: &ClientMetrics{
			StartTime: time.Now(),
		},
		testTypeCounters: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_client_test_type_total",
			Help: "Total tests by type",
		}, []string{"test_type", "data_pattern", "connection_id"}),
		dataPatternHistograms: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "quic_client_data_pattern_duration_seconds",
			Help:    "Data pattern test duration",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0},
		}, []string{"data_pattern", "connection_id", "result"}),
		connectionMetrics: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "quic_client_connection_info",
			Help: "Connection information",
		}, []string{"connection_id", "remote_addr", "tls_version", "cipher_suite"}),
		streamMetrics: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "quic_client_stream_info",
			Help: "Stream information",
		}, []string{"stream_id", "connection_id", "stream_type", "state"}),
	}
}

// NewAdvancedPrometheusExporterWithRegistry создает новый экспортер метрик с указанным registry
func NewAdvancedPrometheusExporterWithRegistry(registry prometheus.Registerer) *AdvancedPrometheusExporter {
	testTypeCounters := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quic_client_test_type_total",
		Help: "Total tests by type",
	}, []string{"test_type", "data_pattern", "connection_id"})
	
	dataPatternHistograms := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "quic_client_data_pattern_duration_seconds",
		Help:    "Data pattern test duration",
		Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0, 1000.0},
	}, []string{"data_pattern", "connection_id", "result"})
	
	connectionMetrics := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quic_client_connection_info",
		Help: "Connection information",
	}, []string{"connection_id", "remote_addr", "tls_version", "cipher_suite"})
	
	streamMetrics := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quic_client_stream_info",
		Help: "Stream information",
	}, []string{"stream_id", "connection_id", "stream_type", "state"})
	
	// Регистрируем метрики
	registry.MustRegister(testTypeCounters, dataPatternHistograms, connectionMetrics, streamMetrics)
	
	return &AdvancedPrometheusExporter{
		metrics: metrics.NewPrometheusMetricsWithRegistry(registry),
		clientMetrics: &ClientMetrics{
			StartTime: time.Now(),
		},
		testTypeCounters: testTypeCounters,
		dataPatternHistograms: dataPatternHistograms,
		connectionMetrics: connectionMetrics,
		streamMetrics: streamMetrics,
	}
}

// UpdateTestType обновляет тип теста
func (ape *AdvancedPrometheusExporter) UpdateTestType(testType, dataPattern string) {
	ape.mu.Lock()
	defer ape.mu.Unlock()

	ape.clientMetrics.TestType = testType
	ape.clientMetrics.DataPattern = dataPattern
	ape.clientMetrics.LastUpdate = time.Now()
}

// RecordTestExecution записывает выполнение теста
func (ape *AdvancedPrometheusExporter) RecordTestExecution(connectionID string, duration time.Duration, result string) {
	ape.mu.RLock()
	testType := ape.clientMetrics.TestType
	dataPattern := ape.clientMetrics.DataPattern
	ape.mu.RUnlock()

	// Записываем в основные метрики
	ape.metrics.RecordScenarioDuration(testType, connectionID, result, duration)

	// Записываем в специфичные для клиента метрики
	ape.testTypeCounters.WithLabelValues(testType, dataPattern, connectionID).Inc()
	ape.dataPatternHistograms.WithLabelValues(dataPattern, connectionID, result).Observe(duration.Seconds())
}

// RecordConnectionInfo записывает информацию о соединении
func (ape *AdvancedPrometheusExporter) RecordConnectionInfo(connectionID, remoteAddr, tlsVersion, cipherSuite string) {
	ape.connectionMetrics.WithLabelValues(connectionID, remoteAddr, tlsVersion, cipherSuite).Set(1)
}

// RecordStreamInfo записывает информацию о потоке
func (ape *AdvancedPrometheusExporter) RecordStreamInfo(streamID, connectionID, streamType, state string) {
	ape.streamMetrics.WithLabelValues(streamID, connectionID, streamType, state).Set(1)
}

// RecordLatency записывает задержку
func (ape *AdvancedPrometheusExporter) RecordLatency(latency time.Duration) {
	ape.metrics.RecordLatency(latency)
}

// RecordJitter записывает джиттер
func (ape *AdvancedPrometheusExporter) RecordJitter(jitter time.Duration) {
	ape.metrics.RecordJitter(jitter)
}

// RecordThroughput записывает пропускную способность
func (ape *AdvancedPrometheusExporter) RecordThroughput(throughput float64) {
	ape.metrics.RecordThroughput(throughput)
}

// RecordHandshakeTime записывает время handshake
func (ape *AdvancedPrometheusExporter) RecordHandshakeTime(duration time.Duration) {
	ape.metrics.RecordHandshakeTime(duration)
}

// RecordRTT записывает RTT
func (ape *AdvancedPrometheusExporter) RecordRTT(rtt time.Duration) {
	ape.metrics.RecordRTT(rtt)
}

// IncrementConnections увеличивает счетчик соединений
func (ape *AdvancedPrometheusExporter) IncrementConnections() {
	ape.metrics.IncrementConnections()
	ape.mu.Lock()
	ape.clientMetrics.ConnectionCount++
	ape.mu.Unlock()
}

// DecrementConnections уменьшает счетчик соединений
func (ape *AdvancedPrometheusExporter) DecrementConnections() {
	ape.metrics.DecrementConnections()
	ape.mu.Lock()
	ape.clientMetrics.ConnectionCount--
	ape.mu.Unlock()
}

// IncrementStreams увеличивает счетчик потоков
func (ape *AdvancedPrometheusExporter) IncrementStreams() {
	ape.metrics.IncrementStreams()
	ape.mu.Lock()
	ape.clientMetrics.StreamCount++
	ape.mu.Unlock()
}

// DecrementStreams уменьшает счетчик потоков
func (ape *AdvancedPrometheusExporter) DecrementStreams() {
	ape.metrics.DecrementStreams()
	ape.mu.Lock()
	ape.clientMetrics.StreamCount--
	ape.mu.Unlock()
}

// AddBytesSent добавляет отправленные байты
func (ape *AdvancedPrometheusExporter) AddBytesSent(bytes int64) {
	ape.metrics.AddBytesSent(bytes)
}

// AddBytesReceived добавляет полученные байты
func (ape *AdvancedPrometheusExporter) AddBytesReceived(bytes int64) {
	ape.metrics.AddBytesReceived(bytes)
}

// IncrementErrors увеличивает счетчик ошибок
func (ape *AdvancedPrometheusExporter) IncrementErrors() {
	ape.metrics.IncrementErrors()
}

// IncrementRetransmits увеличивает счетчик ретрансмиссий
func (ape *AdvancedPrometheusExporter) IncrementRetransmits() {
	ape.metrics.IncrementRetransmits()
}

// IncrementHandshakes увеличивает счетчик handshake
func (ape *AdvancedPrometheusExporter) IncrementHandshakes() {
	ape.metrics.IncrementHandshakes()
}

// IncrementZeroRTT увеличивает счетчик 0-RTT
func (ape *AdvancedPrometheusExporter) IncrementZeroRTT() {
	ape.metrics.IncrementZeroRTT()
}

// IncrementOneRTT увеличивает счетчик 1-RTT
func (ape *AdvancedPrometheusExporter) IncrementOneRTT() {
	ape.metrics.IncrementOneRTT()
}

// IncrementSessionResumptions увеличивает счетчик возобновлений сессии
func (ape *AdvancedPrometheusExporter) IncrementSessionResumptions() {
	ape.metrics.IncrementSessionResumptions()
}

// SetCurrentThroughput устанавливает текущую пропускную способность
func (ape *AdvancedPrometheusExporter) SetCurrentThroughput(throughput float64) {
	ape.metrics.SetCurrentThroughput(throughput)
}

// SetCurrentLatency устанавливает текущую задержку
func (ape *AdvancedPrometheusExporter) SetCurrentLatency(latency time.Duration) {
	ape.metrics.SetCurrentLatency(latency)
}

// SetPacketLossRate устанавливает коэффициент потерь пакетов
func (ape *AdvancedPrometheusExporter) SetPacketLossRate(rate float64) {
	ape.metrics.SetPacketLossRate(rate)
}

// SetConnectionDuration устанавливает длительность соединения
func (ape *AdvancedPrometheusExporter) SetConnectionDuration(duration time.Duration) {
	ape.metrics.SetConnectionDuration(duration)
}

// RecordScenarioEvent записывает событие сценария
func (ape *AdvancedPrometheusExporter) RecordScenarioEvent(scenario, connectionID, streamID, result string) {
	ape.metrics.RecordScenarioEvent(scenario, connectionID, streamID, result)
}

// RecordErrorEvent записывает событие ошибки
func (ape *AdvancedPrometheusExporter) RecordErrorEvent(errorType, connectionID, streamID, severity string) {
	ape.metrics.RecordErrorEvent(errorType, connectionID, streamID, severity)
}

// RecordProtocolEvent записывает событие протокола
func (ape *AdvancedPrometheusExporter) RecordProtocolEvent(eventType, connectionID, tlsVersion, cipherSuite string) {
	ape.metrics.RecordProtocolEvent(eventType, connectionID, tlsVersion, cipherSuite)
}

// RecordNetworkLatency записывает сетевую задержку по профилю
func (ape *AdvancedPrometheusExporter) RecordNetworkLatency(networkProfile, connectionID, region string, latency time.Duration) {
	ape.metrics.RecordNetworkLatency(networkProfile, connectionID, region, latency)
}

// GetClientMetrics возвращает текущие метрики клиента
func (ape *AdvancedPrometheusExporter) GetClientMetrics() *ClientMetrics {
	ape.mu.RLock()
	defer ape.mu.RUnlock()

	return &ClientMetrics{
		TestType:        ape.clientMetrics.TestType,
		DataPattern:     ape.clientMetrics.DataPattern,
		ConnectionCount: ape.clientMetrics.ConnectionCount,
		StreamCount:     ape.clientMetrics.StreamCount,
		StartTime:       ape.clientMetrics.StartTime,
		LastUpdate:      ape.clientMetrics.LastUpdate,
	}
}
