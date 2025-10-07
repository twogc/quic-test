package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics содержит все Prometheus метрики для QUIC
type PrometheusMetrics struct {
	// Congestion Control метрики
	CCBandwidthBps    prometheus.Gauge
	CCCWNDBytes       prometheus.Gauge
	CCMinRTTMs        prometheus.Gauge
	CCState           prometheus.Gauge
	CCPacingBps       prometheus.Gauge
	
	// ACK Frequency метрики
	ACKFreqThreshold  prometheus.Gauge
	ACKMaxDelayMs     prometheus.Gauge
	ACKFrequencyCount prometheus.Counter
	
	// FEC метрики
	FECPacketsSent    prometheus.Counter
	FECPacketsRecovered prometheus.Counter
	FECRedundancy     prometheus.Gauge
	
	// Connection метрики
	ConnectionsActive prometheus.Gauge
	ConnectionsTotal  prometheus.Counter
	StreamsActive     prometheus.Gauge
	StreamsTotal      prometheus.Counter
	
	// Performance метрики
	BytesSent         prometheus.Counter
	BytesReceived     prometheus.Counter
	PacketsSent       prometheus.Counter
	PacketsReceived   prometheus.Counter
	PacketsLost       prometheus.Counter
	
	// RTT метрики
	RTTMinMs          prometheus.Gauge
	RTTMaxMs          prometheus.Gauge
	RTTMeanMs         prometheus.Gauge
	RTTPercentile95Ms prometheus.Gauge
	
	// Throughput метрики
	ThroughputBps     prometheus.Gauge
	GoodputBps        prometheus.Gauge
	
	// Состояние
	mu sync.RWMutex
}

// NewPrometheusMetrics создает новые Prometheus метрики
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		// Congestion Control метрики
		CCBandwidthBps: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_cc_bw_bps",
			Help: "Current bandwidth estimate in bytes per second",
		}),
		CCCWNDBytes: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_cc_cwnd_bytes",
			Help: "Current congestion window size in bytes",
		}),
		CCMinRTTMs: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_cc_min_rtt_ms",
			Help: "Minimum RTT in milliseconds",
		}),
		CCState: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_cc_state",
			Help: "Current congestion control state (0=Startup, 1=Drain, 2=ProbeBW, 3=ProbeRTT)",
		}),
		CCPacingBps: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_pacing_bps",
			Help: "Current pacing rate in bytes per second",
		}),
		
		// ACK Frequency метрики
		ACKFreqThreshold: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_ack_freq_threshold",
			Help: "ACK frequency threshold",
		}),
		ACKMaxDelayMs: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_ack_max_delay_ms",
			Help: "Maximum ACK delay in milliseconds",
		}),
		ACKFrequencyCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_ack_frequency_total",
			Help: "Total number of ACK frequency events",
		}),
		
		// FEC метрики
		FECPacketsSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_fec_packets_sent_total",
			Help: "Total number of FEC packets sent",
		}),
		FECPacketsRecovered: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_fec_packets_recovered_total",
			Help: "Total number of packets recovered by FEC",
		}),
		FECRedundancy: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_fec_redundancy_ratio",
			Help: "FEC redundancy ratio",
		}),
		
		// Connection метрики
		ConnectionsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_connections_active",
			Help: "Number of active connections",
		}),
		ConnectionsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_connections_total",
			Help: "Total number of connections",
		}),
		StreamsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_streams_active",
			Help: "Number of active streams",
		}),
		StreamsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_streams_total",
			Help: "Total number of streams",
		}),
		
		// Performance метрики
		BytesSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_bytes_sent_total",
			Help: "Total bytes sent",
		}),
		BytesReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_bytes_received_total",
			Help: "Total bytes received",
		}),
		PacketsSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_packets_sent_total",
			Help: "Total packets sent",
		}),
		PacketsReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_packets_received_total",
			Help: "Total packets received",
		}),
		PacketsLost: promauto.NewCounter(prometheus.CounterOpts{
			Name: "quic_packets_lost_total",
			Help: "Total packets lost",
		}),
		
		// RTT метрики
		RTTMinMs: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_rtt_min_ms",
			Help: "Minimum RTT in milliseconds",
		}),
		RTTMaxMs: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_rtt_max_ms",
			Help: "Maximum RTT in milliseconds",
		}),
		RTTMeanMs: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_rtt_mean_ms",
			Help: "Mean RTT in milliseconds",
		}),
		RTTPercentile95Ms: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_rtt_p95_ms",
			Help: "95th percentile RTT in milliseconds",
		}),
		
		// Throughput метрики
		ThroughputBps: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_throughput_bps",
			Help: "Current throughput in bytes per second",
		}),
		GoodputBps: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "quic_goodput_bps",
			Help: "Current goodput in bytes per second",
		}),
	}
}

// UpdateCCMetrics обновляет метрики congestion control
func (pm *PrometheusMetrics) UpdateCCMetrics(bandwidthBps float64, cwndBytes int, minRTTMs float64, state int, pacingBps int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.CCBandwidthBps.Set(bandwidthBps)
	pm.CCCWNDBytes.Set(float64(cwndBytes))
	pm.CCMinRTTMs.Set(minRTTMs)
	pm.CCState.Set(float64(state))
	pm.CCPacingBps.Set(float64(pacingBps))
}

// UpdateACKFrequencyMetrics обновляет метрики ACK frequency
func (pm *PrometheusMetrics) UpdateACKFrequencyMetrics(threshold int, maxDelayMs float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.ACKFreqThreshold.Set(float64(threshold))
	pm.ACKMaxDelayMs.Set(maxDelayMs)
	pm.ACKFrequencyCount.Inc()
}

// UpdateFECMetrics обновляет метрики FEC
func (pm *PrometheusMetrics) UpdateFECMetrics(packetsSent, packetsRecovered int, redundancy float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.FECPacketsSent.Add(float64(packetsSent))
	pm.FECPacketsRecovered.Add(float64(packetsRecovered))
	pm.FECRedundancy.Set(redundancy)
}

// UpdateConnectionMetrics обновляет метрики соединений
func (pm *PrometheusMetrics) UpdateConnectionMetrics(activeConnections, totalConnections, activeStreams, totalStreams int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.ConnectionsActive.Set(float64(activeConnections))
	pm.ConnectionsTotal.Add(float64(totalConnections))
	pm.StreamsActive.Set(float64(activeStreams))
	pm.StreamsTotal.Add(float64(totalStreams))
}

// UpdatePerformanceMetrics обновляет метрики производительности
func (pm *PrometheusMetrics) UpdatePerformanceMetrics(bytesSent, bytesReceived, packetsSent, packetsReceived, packetsLost int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.BytesSent.Add(float64(bytesSent))
	pm.BytesReceived.Add(float64(bytesReceived))
	pm.PacketsSent.Add(float64(packetsSent))
	pm.PacketsReceived.Add(float64(packetsReceived))
	pm.PacketsLost.Add(float64(packetsLost))
}

// UpdateRTTMetrics обновляет метрики RTT
func (pm *PrometheusMetrics) UpdateRTTMetrics(minRTT, maxRTT, meanRTT, p95RTT time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.RTTMinMs.Set(float64(minRTT.Nanoseconds()) / 1e6)
	pm.RTTMaxMs.Set(float64(maxRTT.Nanoseconds()) / 1e6)
	pm.RTTMeanMs.Set(float64(meanRTT.Nanoseconds()) / 1e6)
	pm.RTTPercentile95Ms.Set(float64(p95RTT.Nanoseconds()) / 1e6)
}

// UpdateThroughputMetrics обновляет метрики throughput
func (pm *PrometheusMetrics) UpdateThroughputMetrics(throughputBps, goodputBps int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.ThroughputBps.Set(float64(throughputBps))
	pm.GoodputBps.Set(float64(goodputBps))
}