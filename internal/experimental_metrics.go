package internal

import (
	"math"
	"time"

	"quic-test/internal/congestion"
)

// ExperimentalMetricsCollector collects advanced metrics for BBRv3 experiments
type ExperimentalMetricsCollector struct {
	// Throughput and goodput tracking
	bytesAcked         int64
	retransmittedBytes int64
	startTime          time.Time
	
	// RTT samples for percentiles and jitter
	rttSamples []time.Duration
	
	// Multiple flow tracking for fairness
	flowThroughputs []float64
	
	// Recovery tracking
	lastLossEvent   time.Time
	recoveryEvents  []time.Time
}

// NewExperimentalMetricsCollector creates a new metrics collector
func NewExperimentalMetricsCollector() *ExperimentalMetricsCollector {
	return &ExperimentalMetricsCollector{
		rttSamples:     make([]time.Duration, 0, 1000),
		flowThroughputs: make([]float64, 0),
		recoveryEvents: make([]time.Time, 0),
		startTime:      time.Now(),
	}
}

// RecordRTT records an RTT sample
func (emc *ExperimentalMetricsCollector) RecordRTT(rtt time.Duration) {
	emc.rttSamples = append(emc.rttSamples, rtt)
	// Keep only last 1000 samples
	if len(emc.rttSamples) > 1000 {
		emc.rttSamples = emc.rttSamples[len(emc.rttSamples)-1000:]
	}
}

// RecordBytesAcked records acknowledged bytes
func (emc *ExperimentalMetricsCollector) RecordBytesAcked(bytes int64) {
	emc.bytesAcked += bytes
}

// RecordRetransmittedBytes records retransmitted bytes
func (emc *ExperimentalMetricsCollector) RecordRetransmittedBytes(bytes int64) {
	emc.retransmittedBytes += bytes
}

// RecordLossEvent records a loss event for recovery tracking
func (emc *ExperimentalMetricsCollector) RecordLossEvent() {
	emc.lastLossEvent = time.Now()
}

// RecordRecoveryEvent records a recovery event
func (emc *ExperimentalMetricsCollector) RecordRecoveryEvent() {
	if !emc.lastLossEvent.IsZero() {
		emc.recoveryEvents = append(emc.recoveryEvents, time.Now())
	}
}

// RecordFlowThroughput records throughput for a flow (for fairness calculation)
func (emc *ExperimentalMetricsCollector) RecordFlowThroughput(throughput float64) {
	emc.flowThroughputs = append(emc.flowThroughputs, throughput)
}

// GetThroughput calculates throughput in Mbps
func (emc *ExperimentalMetricsCollector) GetThroughput() float64 {
	duration := time.Since(emc.startTime).Seconds()
	if duration <= 0 {
		return 0.0
	}
	// Convert bytes/s to Mbps
	return (float64(emc.bytesAcked) * 8) / (duration * 1024 * 1024)
}

// GetGoodput calculates goodput in Mbps (excluding retransmissions)
func (emc *ExperimentalMetricsCollector) GetGoodput() float64 {
	duration := time.Since(emc.startTime).Seconds()
	if duration <= 0 {
		return 0.0
	}
	goodputBytes := emc.bytesAcked - emc.retransmittedBytes
	if goodputBytes < 0 {
		goodputBytes = 0
	}
	// Convert bytes/s to Mbps
	return (float64(goodputBytes) * 8) / (duration * 1024 * 1024)
}

// GetRetransmissionRate calculates retransmission rate
func (emc *ExperimentalMetricsCollector) GetRetransmissionRate() float64 {
	if emc.bytesAcked == 0 {
		return 0.0
	}
	return float64(emc.retransmittedBytes) / float64(emc.bytesAcked)
}

// GetRTTPercentiles calculates p50, p95, p99 from RTT samples
func (emc *ExperimentalMetricsCollector) GetRTTPercentiles() (p50, p95, p99 time.Duration) {
	return congestion.CalculateRTTPercentiles(emc.rttSamples)
}

// GetJitter calculates jitter (standard deviation) from RTT samples
func (emc *ExperimentalMetricsCollector) GetJitter() time.Duration {
	return congestion.CalculateJitter(emc.rttSamples)
}

// GetFairnessIndex calculates Jain's Fairness Index from flow throughputs
func (emc *ExperimentalMetricsCollector) GetFairnessIndex() float64 {
	return congestion.JainFairnessIndex(emc.flowThroughputs)
}

// GetAverageRecoveryTime calculates average recovery time
func (emc *ExperimentalMetricsCollector) GetAverageRecoveryTime() time.Duration {
	if len(emc.recoveryEvents) == 0 || emc.lastLossEvent.IsZero() {
		return 0
	}
	
	// Calculate average time from loss to recovery
	var total time.Duration
	count := 0
	for _, recoveryTime := range emc.recoveryEvents {
		if recoveryTime.After(emc.lastLossEvent) {
			total += recoveryTime.Sub(emc.lastLossEvent)
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

// GetMetricsMap converts collected metrics to a map for integration
func (emc *ExperimentalMetricsCollector) GetMetricsMap() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// Throughput metrics
	metrics["ThroughputMbps"] = emc.GetThroughput()
	metrics["GoodputMbps"] = emc.GetGoodput()
	metrics["RetransmissionRate"] = emc.GetRetransmissionRate()
	
	// RTT metrics
	p50, p95, p99 := emc.GetRTTPercentiles()
	metrics["RTTP50Ms"] = float64(p50.Nanoseconds()) / 1e6
	metrics["RTTP95Ms"] = float64(p95.Nanoseconds()) / 1e6
	metrics["RTTP99Ms"] = float64(p99.Nanoseconds()) / 1e6
	
	jitter := emc.GetJitter()
	metrics["JitterMs"] = float64(jitter.Nanoseconds()) / 1e6
	
	// Fairness
	metrics["FairnessIndex"] = emc.GetFairnessIndex()
	
	// Recovery
	avgRecovery := emc.GetAverageRecoveryTime()
	metrics["RecoveryTimeMs"] = float64(avgRecovery.Nanoseconds()) / 1e6
	
	return metrics
}

// CalculateRTTPercentiles calculates percentiles from RTT samples (helper)
func CalculateRTTPercentiles(samples []float64) (p50, p95, p99 float64) {
	if len(samples) == 0 {
		return 0, 0, 0
	}
	
	// Copy and sort
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	n := len(sorted)
	p50Idx := int(float64(n) * 0.50)
	p95Idx := int(float64(n) * 0.95)
	p99Idx := int(float64(n) * 0.99)
	
	if p50Idx >= n {
		p50Idx = n - 1
	}
	if p95Idx >= n {
		p95Idx = n - 1
	}
	if p99Idx >= n {
		p99Idx = n - 1
	}
	
	return sorted[p50Idx], sorted[p95Idx], sorted[p99Idx]
}

// CalculateMeanRTT calculates mean RTT from samples
func CalculateMeanRTT(samples []time.Duration) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	var sum time.Duration
	for _, rtt := range samples {
		sum += rtt
	}
	return sum / time.Duration(len(samples))
}

// CalculateRTTStdDev calculates standard deviation of RTT
func CalculateRTTStdDev(samples []time.Duration, mean time.Duration) float64 {
	if len(samples) == 0 {
		return 0.0
	}
	
	meanMs := float64(mean.Nanoseconds()) / 1e6
	var variance float64
	
	for _, rtt := range samples {
		rttMs := float64(rtt.Nanoseconds()) / 1e6
		diff := rttMs - meanMs
		variance += diff * diff
	}
	variance /= float64(len(samples))
	
	return math.Sqrt(variance)
}

