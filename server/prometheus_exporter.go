package server

import (
	"sync"
	"time"

	"quic-test/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AdvancedPrometheusExporter предоставляет продвинутые метрики Prometheus для сервера
type AdvancedPrometheusExporter struct {
	// Основные метрики
	metrics *metrics.PrometheusMetrics

	// Дополнительные метрики сервера
	serverMetrics *ServerMetrics

	// Счетчики по типам запросов
	requestTypeCounters *prometheus.CounterVec

	// Гистограммы по обработке запросов
	requestProcessingHistograms *prometheus.HistogramVec

	// Метрики по соединениям
	connectionMetrics *prometheus.GaugeVec

	// Метрики по потокам
	streamMetrics *prometheus.GaugeVec

	// Метрики по обработке данных
	dataProcessingMetrics *prometheus.CounterVec

	mu sync.RWMutex
}

// ServerMetrics содержит метрики сервера
type ServerMetrics struct {
	ServerAddr         string
	MaxConnections     int
	CurrentConnections int
	CurrentStreams     int
	StartTime          time.Time
	LastUpdate         time.Time
	Uptime             time.Duration
}

// NewAdvancedPrometheusExporter создает новый экспортер метрик для сервера
func NewAdvancedPrometheusExporter(serverAddr string) *AdvancedPrometheusExporter {
	return &AdvancedPrometheusExporter{
		metrics: metrics.NewPrometheusMetrics(),
		serverMetrics: &ServerMetrics{
			ServerAddr: serverAddr,
			StartTime:  time.Now(),
		},
		requestTypeCounters: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_server_request_type_total",
			Help: "Total requests by type",
		}, []string{"request_type", "connection_id", "stream_id", "result"}),
		requestProcessingHistograms: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "quic_server_request_processing_duration_seconds",
			Help:    "Request processing duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}, []string{"request_type", "connection_id", "result"}),
		connectionMetrics: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "quic_server_connection_info",
			Help: "Server connection information",
		}, []string{"connection_id", "remote_addr", "tls_version", "cipher_suite", "state"}),
		streamMetrics: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "quic_server_stream_info",
			Help: "Server stream information",
		}, []string{"stream_id", "connection_id", "stream_type", "state", "direction"}),
		dataProcessingMetrics: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_server_data_processing_total",
			Help: "Data processing metrics",
		}, []string{"operation", "connection_id", "stream_id", "data_type"}),
	}
}

// UpdateServerInfo обновляет информацию о сервере
func (ape *AdvancedPrometheusExporter) UpdateServerInfo(maxConnections int) {
	ape.mu.Lock()
	defer ape.mu.Unlock()

	ape.serverMetrics.MaxConnections = maxConnections
	ape.serverMetrics.LastUpdate = time.Now()
	ape.serverMetrics.Uptime = time.Since(ape.serverMetrics.StartTime)
}

// RecordRequestProcessing записывает обработку запроса
func (ape *AdvancedPrometheusExporter) RecordRequestProcessing(requestType, connectionID string, duration time.Duration, result string) {
	// Записываем в основные метрики
		ape.metrics.RecordScenarioDuration(duration)

	// Записываем в специфичные для сервера метрики
	ape.requestTypeCounters.WithLabelValues(requestType, connectionID, "", result).Inc()
	ape.requestProcessingHistograms.WithLabelValues(requestType, connectionID, result).Observe(duration.Seconds())
}

// RecordConnectionInfo записывает информацию о соединении
func (ape *AdvancedPrometheusExporter) RecordConnectionInfo(connectionID, remoteAddr, tlsVersion, cipherSuite, state string) {
	ape.connectionMetrics.WithLabelValues(connectionID, remoteAddr, tlsVersion, cipherSuite, state).Set(1)
}

// RecordStreamInfo записывает информацию о потоке
func (ape *AdvancedPrometheusExporter) RecordStreamInfo(streamID, connectionID, streamType, state, direction string) {
	ape.streamMetrics.WithLabelValues(streamID, connectionID, streamType, state, direction).Set(1)
}

// RecordDataProcessing записывает обработку данных
func (ape *AdvancedPrometheusExporter) RecordDataProcessing(operation, connectionID, streamID, dataType string, bytes int64) {
	ape.dataProcessingMetrics.WithLabelValues(operation, connectionID, streamID, dataType).Add(float64(bytes))
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
	ape.metrics.RecordThroughput(int64(throughput))
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
	ape.serverMetrics.CurrentConnections++
	ape.mu.Unlock()
}

// DecrementConnections уменьшает счетчик соединений
func (ape *AdvancedPrometheusExporter) DecrementConnections() {
	ape.metrics.DecrementConnections()
	ape.mu.Lock()
	ape.serverMetrics.CurrentConnections--
	ape.mu.Unlock()
}

// IncrementStreams увеличивает счетчик потоков
func (ape *AdvancedPrometheusExporter) IncrementStreams() {
	ape.metrics.IncrementStreams()
	ape.mu.Lock()
	ape.serverMetrics.CurrentStreams++
	ape.mu.Unlock()
}

// DecrementStreams уменьшает счетчик потоков
func (ape *AdvancedPrometheusExporter) DecrementStreams() {
	ape.metrics.DecrementStreams()
	ape.mu.Lock()
	ape.serverMetrics.CurrentStreams--
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
	ape.metrics.SetCurrentThroughput(int64(throughput))
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
	ape.metrics.RecordScenarioEvent(scenario)
}

// RecordErrorEvent записывает событие ошибки
func (ape *AdvancedPrometheusExporter) RecordErrorEvent(errorType, connectionID, streamID, severity string) {
	ape.metrics.RecordErrorEvent(errorType)
}

// RecordProtocolEvent записывает событие протокола
func (ape *AdvancedPrometheusExporter) RecordProtocolEvent(eventType, connectionID, tlsVersion, cipherSuite string) {
	ape.metrics.RecordProtocolEvent(eventType)
}

// RecordNetworkLatency записывает сетевую задержку по профилю
func (ape *AdvancedPrometheusExporter) RecordNetworkLatency(networkProfile, connectionID, region string, latency time.Duration) {
	ape.metrics.RecordNetworkLatency(latency)
}

// GetServerMetrics возвращает текущие метрики сервера
func (ape *AdvancedPrometheusExporter) GetServerMetrics() *ServerMetrics {
	ape.mu.RLock()
	defer ape.mu.RUnlock()

	return &ServerMetrics{
		ServerAddr:         ape.serverMetrics.ServerAddr,
		MaxConnections:     ape.serverMetrics.MaxConnections,
		CurrentConnections: ape.serverMetrics.CurrentConnections,
		CurrentStreams:     ape.serverMetrics.CurrentStreams,
		StartTime:          ape.serverMetrics.StartTime,
		LastUpdate:         ape.serverMetrics.LastUpdate,
		Uptime:             time.Since(ape.serverMetrics.StartTime),
	}
}
