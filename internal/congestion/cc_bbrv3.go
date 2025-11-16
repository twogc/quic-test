package congestion

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// BBRv3 congestion control algorithm implementation
// Based on draft-ietf-ccwg-bbr-04 specification
//
// Key improvements over BBRv2:
// - Dual-scale bandwidth model (fast/slow)
// - Loss threshold = 2%
// - β = 0.7 (cwnd reduction factor)
// - Headroom = 0.15 BDP
// - Adaptive pacing gain (Startup 2.77, Drain 0.35)
// - ProbeRTTDuration = 200ms

type bbrv3State int

const (
	bbrv3Startup bbrv3State = iota
	bbrv3Drain
	bbrv3ProbeBW
	bbrv3ProbeRTT
)

// BBRv3Parameters contains BBRv3 algorithm parameters
type BBRv3Parameters struct {
	// Loss threshold (2% according to draft)
	LossThreshold float64
	
	// β factor for cwnd reduction (0.7)
	Beta float64
	
	// Headroom as fraction of BDP (0.15)
	HeadroomFraction float64
	
	// Pacing gains
	StartupPacingGain float64 // 2.77
	DrainPacingGain  float64 // 0.35
	
	// ProbeRTT duration (200ms)
	ProbeRTTDuration time.Duration
}

// DefaultBBRv3Parameters returns default BBRv3 parameters per draft
func DefaultBBRv3Parameters() BBRv3Parameters {
	return BBRv3Parameters{
		LossThreshold:     0.02,  // 2%
		Beta:             0.7,    // β = 0.7
		HeadroomFraction: 0.15,  // 15% of BDP
		StartupPacingGain: 2.77,
		DrainPacingGain:  0.35,
		ProbeRTTDuration: 200 * time.Millisecond,
	}
}

// OptimizedBBRv3Parameters returns optimized parameters for better performance
// These parameters are tuned for inter-regional networks (RTT > 80ms)
func OptimizedBBRv3Parameters() BBRv3Parameters {
	return BBRv3Parameters{
		LossThreshold:     0.018, // 1.8% (slightly more sensitive for faster reaction)
		Beta:             0.72,    // Slightly less aggressive reduction
		HeadroomFraction: 0.12,    // 12% (allows more throughput while maintaining stability)
		StartupPacingGain: 2.77,  // Keep draft value
		DrainPacingGain:  0.32,   // Slightly more aggressive drain (faster queue clearing)
		ProbeRTTDuration: 200 * time.Millisecond,
	}
}

// BBRv3Metrics contains BBRv3-specific metrics for visualization
type BBRv3Metrics struct {
	Phase          string  `json:"phase"`            // Startup, Drain, ProbeBW, ProbeRTT
	BandwidthFast  float64 `json:"bw_fast"`         // Fast-scale bandwidth estimate (bps)
	BandwidthSlow  float64 `json:"bw_slow"`         // Slow-scale bandwidth estimate (bps)
	Bandwidth      float64 `json:"bw"`               // Current bandwidth (max of fast/slow)
	LossRateRound  float64 `json:"loss_rate_round"`  // Loss rate per round
	LossRateEMA    float64 `json:"loss_rate_ema"`    // EMA loss rate (for metrics)
	LossThreshold  float64 `json:"loss_threshold"`   // Loss threshold (2%)
	HeadroomUsage  float64 `json:"headroom_usage"`   // Headroom usage (0.0-1.0)
	InflightTarget float64 `json:"inflight_target"`  // Target inflight with headroom reserved
	PacingQuantum  int64   `json:"pacing_quantum"`   // Pacing quantum (bytes)
	SendQuantum    int64   `json:"send_quantum"`     // Send quantum (bytes)
	
	// Advanced metrics for experiments
	PacingGain            float64            `json:"pacing_gain"`             // Current pacing gain
	CWNDGain              float64            `json:"cwnd_gain"`               // Current CWND gain
	ProbeRTTMinMs         float64            `json:"probe_rtt_min_ms"`        // Minimum RTT during ProbeRTT
	BufferbloatFactor      float64            `json:"bufferbloat_factor"`     // (avg_rtt / min_rtt) - 1
	StabilityIndex         float64            `json:"stability_index"`         // Δ throughput / Δ rtt
	PhaseDurationMs       map[string]float64 `json:"phase_duration_ms"`       // Duration of each phase
	RecoveryTimeMs         float64            `json:"recovery_time_ms"`        // Time to recover from loss (ms)
	LossRecoveryEfficiency float64            `json:"loss_recovery_efficiency"` // recovered / lost
}

// BBRv3 implements the BBRv3 congestion control algorithm
type BBRv3 struct {
	state       bbrv3State
	mtu         int
	cwnd        int
	pacingBps   int64
	minRTT      time.Duration
	minRTTSince time.Time
	bwFast      float64 // Fast-scale bandwidth estimate
	bwSlow      float64 // Slow-scale bandwidth estimate
	bw          float64 // Current bandwidth estimate (max of fast/slow)
	cycleIdx    int
	lastStateTs time.Time
	pacer       *Pacer
	params      BBRv3Parameters
	
	// Loss tracking by round (not sliding window)
	roundAcked      int64 // Bytes acked in current round
	roundLost       int64 // Bytes lost in current round
	roundStartTime  time.Time
	
	// Loss rate tracking (for metrics)
	packetsSent      int64
	packetsLost      int64
	lossRateEMA      float64 // Exponential moving average
	
	// Headroom tracking
	currentInflight int64
	
	// Pacing quantum
	sendQuantum int64
	
	// Phase timing tracking
	phaseStartTimes map[bbrv3State]time.Time
	phaseDurations  map[string]time.Duration
	
	// RTT tracking for bufferbloat calculation
	recentRTTs []time.Duration // Last N RTT samples
	recentRTTIdx int
	
	// Recovery tracking
	lastLossTime     time.Time
	lastRecoveryTime time.Time
	recoveredPackets int64
	
	// Throughput delta tracking for stability
	lastThroughput float64
	lastRTT        time.Duration
	
	// Metrics for visualization
	metrics    BBRv3Metrics
	metricsMux sync.Mutex // Protects metrics from concurrent access

	// Qlog callback
	qlogCallback func(eventType string, data map[string]interface{})
}

// NewBBRv3 creates a new BBRv3 congestion controller
func NewBBRv3(mtu int, initialCWND int) *BBRv3 {
	if mtu <= 0 {
		mtu = 1460
	}
	if initialCWND <= 0 {
		initialCWND = 32 * mtu
	}

	// Use optimized parameters for better performance
	params := OptimizedBBRv3Parameters()
	now := time.Now()
	
	b := &BBRv3{
		state:         bbrv3Startup,
		mtu:           mtu,
		cwnd:          initialCWND,
		lastStateTs:   now,
		pacer:         NewPacer(mtu),
		params:        params,
		roundStartTime: now,
		sendQuantum:   int64(2 * mtu), // Initial quantum
		phaseStartTimes: make(map[bbrv3State]time.Time),
		phaseDurations:  make(map[string]time.Duration),
		recentRTTs:      make([]time.Duration, 100), // Last 100 RTT samples
	}
	
	b.phaseStartTimes[bbrv3Startup] = now
	b.updateMetrics()
	return b
}

// Name returns the algorithm name
func (b *BBRv3) Name() string {
	return "bbrv3"
}

// resetRound resets the loss tracking round
func (b *BBRv3) resetRound() {
	b.roundAcked = 0
	b.roundLost = 0
	b.roundStartTime = time.Now()
}

// roundTotal returns total bytes in current round
func (b *BBRv3) roundTotal() int64 {
	return b.roundAcked + b.roundLost
}

// fullPipeDetected checks if full pipe bandwidth is detected
func (b *BBRv3) fullPipeDetected() bool {
	// Simple heuristic: if bandwidth hasn't increased significantly
	// in last 2 seconds, consider pipe full
	return b.bwSlow > 0 && (b.bwFast/b.bwSlow) < 1.1
}

// maxDur returns the maximum of two durations
func maxDur(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// maxF returns the maximum of two float64 values
func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// OnAck handles an ACK event
func (b *BBRv3) OnAck(s Sample) (cwnd int, pacing int64) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't crash - return safe defaults
			b.logQlogEvent("panic", map[string]interface{}{
				"error": fmt.Sprintf("%v", r),
				"state": b.getStateString(),
			})
			// Return last known good values
			cwnd = b.cwnd
			pacing = b.pacingBps
		}
	}()
	
	now := time.Now()
	oldState := b.state
	oldCWND := b.cwnd
	oldPacing := b.pacingBps
	
	// Track recent RTTs for bufferbloat calculation
	if s.RTT > 0 {
		b.recentRTTs[b.recentRTTIdx%len(b.recentRTTs)] = s.RTT
		b.recentRTTIdx++
		b.lastRTT = s.RTT
	}
	
	// Update min RTT
	if s.RTT > 0 && (b.minRTT == 0 || s.RTT < b.minRTT) {
		oldMinRTT := b.minRTT
		b.minRTT = s.RTT
		b.minRTTSince = now
		
		// Track ProbeRTT minimum
		if b.state == bbrv3ProbeRTT {
			b.metrics.ProbeRTTMinMs = float64(b.minRTT.Nanoseconds()) / 1e6
		}
		
		b.logQlogEvent("rtt_update", map[string]interface{}{
			"old_rtt":      float64(oldMinRTT.Nanoseconds()) / 1e6,
			"new_rtt":      float64(s.RTT.Nanoseconds()) / 1e6,
			"min_rtt":      float64(b.minRTT.Nanoseconds()) / 1e6,
			"rtt_variance": 0.0,
			"smoothed_rtt": float64(s.RTT.Nanoseconds()) / 1e6,
		})
	}
	
	// Accumulate round acked bytes
	b.roundAcked += s.RS.BytesAcked
	
	// Update dual-scale bandwidth estimates (only if not app-limited)
	if br := s.RS.BandwidthBps(); br > 0 && !s.RS.IsAppLimited {
		// Fast-scale: immediate samples (max) with decay for stability
		if br > b.bwFast {
			b.bwFast = br
		} else {
			// Slow decay: 99.5% per sample (allows tracking decreases)
			b.bwFast = 0.995 * b.bwFast
		}
		
		// Slow-scale: exponential moving average with adaptive alpha
		if b.bwSlow == 0 {
			b.bwSlow = br
		} else {
			// EMA with adaptive alpha: more responsive when bandwidth is changing
			// If fast and slow are diverging, increase responsiveness
			ratio := b.bwFast / b.bwSlow
			alpha := 0.1 // Default
			if ratio > 1.1 || ratio < 0.9 {
				// Bandwidth is changing, be more responsive
				alpha = 0.15
			}
			b.bwSlow = (1-alpha)*b.bwSlow + alpha*br
		}
		
		// Current bandwidth is max of fast/slow
		b.bw = maxF(b.bwFast, b.bwSlow)
		
		b.logQlogEvent("bandwidth_sample", map[string]interface{}{
			"sample_bandwidth": br,
			"bw_fast":          b.bwFast,
			"bw_slow":          b.bwSlow,
			"bandwidth":        b.bw,
			"interval":         s.RS.Interval.Seconds() * 1000,
			"bytes_acked":      s.RS.BytesAcked,
			"is_app_limited":   s.RS.IsAppLimited,
			"rtt":              float64(s.RTT.Nanoseconds()) / 1e6,
		})
	}
	
	// Update pacing quantum
	b.updatePacingQuantum()
	
	// State machine
	switch b.state {
	case bbrv3Startup:
		// Aggressive growth with Startup pacing gain 2.77
		b.cwnd += max(1, int(s.RS.BytesAcked))
		b.pacingBps = int64(b.params.StartupPacingGain * b.bw)
		
		// Transition to Drain after stable bandwidth or timeout
		// Optimized: faster transition if full pipe detected early
		startupTimeout := 2 * time.Second
		if b.fullPipeDetected() {
			// If full pipe detected, reduce timeout to 1.5s for faster transition
			startupTimeout = 1 * time.Second
		}
		if now.Sub(b.lastStateTs) > startupTimeout || b.fullPipeDetected() {
			b.state = bbrv3Drain
			b.lastStateTs = now
			b.logQlogEvent("state_transition", map[string]interface{}{
				"from": "Startup",
				"to":   "Drain",
				"reason": "timeout_or_full_pipe",
			})
		}
		
	case bbrv3Drain:
		// Drain excess queued data with Drain pacing gain 0.35
		// Use inflight target (BDP with headroom reserved)
		inflightTarget := b.inflightTarget()
		b.cwnd = int(inflightTarget)
		b.pacingBps = int64(b.params.DrainPacingGain * b.bw)
		
		// Transition to ProbeBW after draining (adaptive timing)
		probePeriod := maxDur(200*time.Millisecond, 2*b.minRTT)
		if now.Sub(b.lastStateTs) > probePeriod {
			b.state = bbrv3ProbeBW
			b.cycleIdx = 0
			b.lastStateTs = now
			b.logQlogEvent("state_transition", map[string]interface{}{
				"from": "Drain",
				"to":   "ProbeBW",
				"reason": "timeout",
			})
		}
		
	case bbrv3ProbeBW:
		// ProbeBW cycle with optimized adaptive gains
		// Optimized for better balance between throughput and latency
		// Sequence: probe up (1.25) -> maintain (1.0) -> probe down (0.75) -> maintain (1.0)
		gains := []float64{1.25, 1.0, 0.75, 1.0}
		g := gains[b.cycleIdx%len(gains)]
		
		// Adaptive gain adjustment based on loss rate
		// If loss rate is low, we can be slightly more aggressive
		if b.metrics.LossRateRound < 0.01 && b.roundTotal() > 0 {
			// Boost probe up gain slightly when loss is very low
			if g == 1.25 {
				g = 1.28 // Slightly more aggressive probe up
			}
		}
		
		// Inflight target with headroom reserved (not added)
		inflightTarget := g * b.inflightTarget()
		b.cwnd = max(int(inflightTarget), 4*b.mtu) // Minimum 4 MSS
		b.pacingBps = int64(g * b.bw)
		
		// ProbeBW period depends on RTT (adaptive)
		probePeriod := maxDur(200*time.Millisecond, 2*b.minRTT)
		if now.Sub(b.lastStateTs) > probePeriod {
			b.cycleIdx++
			b.lastStateTs = now
		}
		
		// Check for ProbeRTT condition
		if b.minRTT > 0 && now.Sub(b.minRTTSince) > 10*time.Second {
			b.state = bbrv3ProbeRTT
			b.lastStateTs = now
			b.logQlogEvent("state_transition", map[string]interface{}{
				"from": "ProbeBW",
				"to":   "ProbeRTT",
				"reason": "min_rtt_stale",
			})
		}
		
	case bbrv3ProbeRTT:
		// ProbeRTT with reduced cwnd (minimum 4 MSS)
		target := 0.5 * b.BDP()
		b.cwnd = max(int(target), 4*b.mtu) // Minimum 4 MSS
		b.pacingBps = int64(0.5 * b.bw)
		
		// Return to ProbeBW after ProbeRTTDuration (200ms)
		if now.Sub(b.lastStateTs) > b.params.ProbeRTTDuration {
			b.minRTTSince = now
			b.state = bbrv3ProbeBW
			b.cycleIdx = 0
			b.lastStateTs = now
			b.logQlogEvent("state_transition", map[string]interface{}{
				"from": "ProbeRTT",
				"to":   "ProbeBW",
				"reason": "probe_rtt_complete",
			})
		}
	}
	
	// Loss threshold check by round: reduce cwnd if loss rate per round exceeds threshold
	if b.roundTotal() > 0 {
		lossRateRound := float64(b.roundLost) / float64(b.roundTotal())
		if lossRateRound > b.params.LossThreshold {
			b.cwnd = max(int(b.params.Beta*float64(b.cwnd)), 2*b.mtu)
			b.resetRound() // Start new round after reaction
			b.logQlogEvent("loss_threshold_exceeded", map[string]interface{}{
				"loss_rate_round": lossRateRound,
				"threshold":       b.params.LossThreshold,
				"round_acked":     b.roundAcked,
				"round_lost":      b.roundLost,
				"old_cwnd":        oldCWND,
				"new_cwnd":        b.cwnd,
				"beta":            b.params.Beta,
			})
		}
	}
	
	// Update loss rate EMA for metrics (not used for decisions)
	if b.packetsSent > 0 {
		currentLossRate := float64(b.packetsLost) / float64(b.packetsSent)
		if b.lossRateEMA == 0 {
			b.lossRateEMA = currentLossRate
		} else {
			b.lossRateEMA = 0.875*b.lossRateEMA + 0.125*currentLossRate
		}
		b.metrics.LossRateEMA = b.lossRateEMA
	}
	
	// Track phase durations when state changes
	if oldState != b.state {
		// Protect phase tracking with mutex
		b.metricsMux.Lock()
		// Record duration of previous phase
		if startTime, ok := b.phaseStartTimes[oldState]; ok {
			duration := now.Sub(startTime)
			b.phaseDurations[b.getStateStringFromState(oldState)] = duration
		}
		// Start timing new phase
		b.phaseStartTimes[b.state] = now
		b.metricsMux.Unlock()
		
		// Update recovery time if transitioning from recovery state
		if oldState == bbrv3ProbeRTT || (oldState == bbrv3Drain && b.state == bbrv3ProbeBW) {
			if !b.lastRecoveryTime.IsZero() && b.lastLossTime.After(b.lastRecoveryTime) {
				b.lastRecoveryTime = now
			}
		}
	}
	
	// Log state change if it occurred
	if oldState != b.state {
		b.updateMetrics()
		b.logQlogEvent("state_change", map[string]interface{}{
			"old_state":   b.getStateStringFromState(oldState),
			"new_state":   b.getStateString(),
			"reason":      "state_machine",
			"bandwidth":   b.bw,
			"bw_fast":     b.bwFast,
			"bw_slow":     b.bwSlow,
			"min_rtt":     float64(b.minRTT.Nanoseconds()) / 1e6,
			"cwnd":        b.cwnd,
			"pacing_rate": b.pacingBps,
			"loss_rate_ema": b.lossRateEMA,
			"loss_rate_round": b.metrics.LossRateRound,
		})
	}
	
	// Ensure minimum cwnd
	if b.cwnd < 2*b.mtu {
		b.cwnd = 2 * b.mtu
	}
	
	// Set pacing rate
	if b.pacingBps <= 0 && b.minRTT > 0 {
		b.pacingBps = int64(float64(b.cwnd) / b.minRTT.Seconds())
	}
	b.pacer.SetRate(b.pacingBps)
	
	// Update headroom usage metric (how close we are to using reserved headroom)
	bdp := b.BDP()
	if bdp > 0 {
		inflightTarget := b.inflightTarget()
		if b.currentInflight > 0 {
			// Usage is how much of the headroom we're using
			// 0.0 = at inflight_target, 1.0 = at BDP (headroom fully used)
			headroomSize := bdp - inflightTarget
			if headroomSize > 0 {
				headroomUsage := (float64(b.currentInflight) - inflightTarget) / headroomSize
				if headroomUsage > 1.0 {
					headroomUsage = 1.0
				} else if headroomUsage < 0.0 {
					headroomUsage = 0.0
				}
				b.metrics.HeadroomUsage = headroomUsage
			}
		}
	}
	
	// Log CWND update if it changed
	if oldCWND != b.cwnd {
		b.logQlogEvent("cwnd_update", map[string]interface{}{
			"old_cwnd":         oldCWND,
			"new_cwnd":         b.cwnd,
			"change":           b.cwnd - oldCWND,
			"reason":           "ack_processing",
			"bandwidth":        b.bw,
			"bw_fast":          b.bwFast,
			"bw_slow":          b.bwSlow,
			"min_rtt":          float64(b.minRTT.Nanoseconds()) / 1e6,
			"loss_rate_ema":    b.lossRateEMA,
			"loss_rate_round":  b.metrics.LossRateRound,
			"inflight_target":  b.inflightTarget(),
			"headroom_usage":   b.metrics.HeadroomUsage,
		})
	}
	
	// Log pacing update if it changed
	if oldPacing != b.pacingBps {
		b.logQlogEvent("pacing_update", map[string]interface{}{
			"old_rate":   oldPacing,
			"new_rate":   b.pacingBps,
			"tokens":     b.pacer.GetTokens(),
			"bandwidth":  b.bw,
			"bw_fast":    b.bwFast,
			"bw_slow":    b.bwSlow,
			"min_rtt":    float64(b.minRTT.Nanoseconds()) / 1e6,
		})
	}
	
	return b.cwnd, b.pacingBps
}

// OnLoss handles a loss event
func (b *BBRv3) OnLoss() (cwnd int, pacing int64) {
	b.packetsLost++
	b.packetsSent++ // Count lost packets as sent for loss rate calculation
	
	// Track loss time for recovery metrics
	b.lastLossTime = time.Now()
	
	// Accumulate lost bytes in current round
	// Assume average packet size for estimation if not provided
	b.roundLost += int64(b.mtu) // Conservative estimate
	
	// Note: The actual cwnd reduction happens in OnAck when loss threshold is exceeded
	// Here we just track the loss
	
	b.updateMetrics()
	b.logQlogEvent("packet_loss", map[string]interface{}{
		"packets_lost":     b.packetsLost,
		"packets_sent":     b.packetsSent,
		"loss_rate_ema":    b.lossRateEMA,
		"round_lost":       b.roundLost,
		"round_acked":      b.roundAcked,
		"loss_threshold":   b.params.LossThreshold,
		"beta":             b.params.Beta,
	})
	
	return b.cwnd, b.pacingBps
}

// OnPacketSent tracks packet sending for loss rate calculation
func (b *BBRv3) OnPacketSent() {
	b.packetsSent++
	b.currentInflight++
}

// OnPacketAcked tracks packet acknowledgment
func (b *BBRv3) OnPacketAcked() {
	if b.currentInflight > 0 {
		b.currentInflight--
	}
}

// BDP calculates the bandwidth-delay product
func (b *BBRv3) BDP() float64 {
	if b.minRTT <= 0 {
		return float64(b.cwnd)
	}
	return b.bw * b.minRTT.Seconds()
}

// bdp is an alias for BDP (for backward compatibility)
func (b *BBRv3) bdp() float64 {
	return b.BDP()
}

// inflightTarget calculates the target inflight with headroom reserved
// Headroom is RESERVED (not added), so inflight_target = BDP * (1 - headroom_fraction)
func (b *BBRv3) inflightTarget() float64 {
	bdp := b.BDP()
	return bdp * (1.0 - b.params.HeadroomFraction) // e.g. 0.85 * BDP
}

// updatePacingQuantum updates the pacing quantum
// quantum = max(2*MTU, min(64KB, pacing_rate*minRTT/8))
func (b *BBRv3) updatePacingQuantum() {
	if b.minRTT <= 0 || b.bw <= 0 {
		b.sendQuantum = int64(2 * b.mtu)
		return
	}
	
	pacingRate := b.bw
	quantumFromPacing := (pacingRate * b.minRTT.Seconds()) / 8
	minQuantum := int64(2 * b.mtu)
	maxQuantum := int64(64 * 1024) // 64KB
	
	quantum := int64(quantumFromPacing)
	if quantum < minQuantum {
		quantum = minQuantum
	}
	if quantum > maxQuantum {
		quantum = maxQuantum
	}
	
	b.sendQuantum = quantum
}

// GetState returns the current BBRv3 state
func (b *BBRv3) GetState() bbrv3State {
	return b.state
}

// GetCWND returns the current congestion window
func (b *BBRv3) GetCWND() int {
	return b.cwnd
}

// GetPacingRate returns the current pacing rate
func (b *BBRv3) GetPacingRate() int64 {
	return b.pacingBps
}

// GetBandwidth returns the current bandwidth estimate
func (b *BBRv3) GetBandwidth() float64 {
	return b.bw
}

// GetBandwidthFast returns the fast-scale bandwidth estimate
func (b *BBRv3) GetBandwidthFast() float64 {
	return b.bwFast
}

// GetBandwidthSlow returns the slow-scale bandwidth estimate
func (b *BBRv3) GetBandwidthSlow() float64 {
	return b.bwSlow
}

// GetMinRTT returns the minimum RTT
func (b *BBRv3) GetMinRTT() time.Duration {
	return b.minRTT
}

// GetLossRate returns the current packet loss rate (EMA)
func (b *BBRv3) GetLossRate() float64 {
	return b.lossRateEMA
}

// GetMetrics returns BBRv3-specific metrics for visualization
func (b *BBRv3) GetMetrics() BBRv3Metrics {
	b.metricsMux.Lock()
	defer b.metricsMux.Unlock()

	// Create a deep copy of metrics to avoid concurrent map access
	metricsCopy := b.metrics
	if b.metrics.PhaseDurationMs != nil {
		metricsCopy.PhaseDurationMs = make(map[string]float64)
		for k, v := range b.metrics.PhaseDurationMs {
			metricsCopy.PhaseDurationMs[k] = v
		}
	}
	return metricsCopy
}

// updateMetrics updates the metrics structure
func (b *BBRv3) updateMetrics() {
	b.metricsMux.Lock()
	defer b.metricsMux.Unlock()

	b.metrics.Phase = b.getStateString()
	b.metrics.BandwidthFast = b.bwFast
	b.metrics.BandwidthSlow = b.bwSlow
	b.metrics.Bandwidth = b.bw
	b.metrics.PacingQuantum = b.sendQuantum
	b.metrics.SendQuantum = b.sendQuantum
	b.metrics.InflightTarget = b.inflightTarget()
	b.metrics.LossRateEMA = b.lossRateEMA
	b.metrics.LossThreshold = b.params.LossThreshold

	// Pacing and CWND gains
	b.metrics.PacingGain = b.CalculatePacingGain()
	b.metrics.CWNDGain = b.CalculateCWNDGain()

	// ProbeRTT minimum (update if in ProbeRTT state)
	if b.state == bbrv3ProbeRTT && b.minRTT > 0 {
		b.metrics.ProbeRTTMinMs = float64(b.minRTT.Nanoseconds()) / 1e6
	}
	
	// Loss rate per round
	if b.roundTotal() > 0 {
		b.metrics.LossRateRound = float64(b.roundLost) / float64(b.roundTotal())
	}
	
	// Bufferbloat factor: calculate average RTT from recent samples
	if b.minRTT > 0 && len(b.recentRTTs) > 0 {
		var sum time.Duration
		count := 0
		for _, rtt := range b.recentRTTs {
			if rtt > 0 {
				sum += rtt
				count++
			}
		}
		if count > 0 {
			avgRTT := sum / time.Duration(count)
			b.metrics.BufferbloatFactor = b.CalculateBufferbloatFactor(avgRTT)
		}
	}
	
	// Stability index: Δ throughput / Δ rtt
	// This is approximated using recent bandwidth and RTT changes
	if b.lastRTT > 0 && b.lastThroughput > 0 && b.bw > 0 {
		throughputDelta := math.Abs(b.bw - b.lastThroughput)
		rttDeltaMs := math.Abs(float64(b.lastRTT.Nanoseconds())/1e6 - float64(b.minRTT.Nanoseconds())/1e6)
		if rttDeltaMs > 0 {
			b.metrics.StabilityIndex = throughputDelta / rttDeltaMs
		}
		b.lastThroughput = b.bw
	}
	
	// Phase durations (convert to ms) - already protected by metricsMux lock from caller
	b.metrics.PhaseDurationMs = make(map[string]float64)
	for phase, duration := range b.phaseDurations {
		b.metrics.PhaseDurationMs[phase] = float64(duration.Nanoseconds()) / 1e6
	}
	// Add current phase duration
	if startTime, ok := b.phaseStartTimes[b.state]; ok {
		currentDuration := time.Since(startTime)
		b.metrics.PhaseDurationMs[b.getStateString()] = float64(currentDuration.Nanoseconds()) / 1e6
	}
	
	// Recovery time
	if !b.lastLossTime.IsZero() && !b.lastRecoveryTime.IsZero() && b.lastRecoveryTime.After(b.lastLossTime) {
		recoveryDuration := b.lastRecoveryTime.Sub(b.lastLossTime)
		b.metrics.RecoveryTimeMs = float64(recoveryDuration.Nanoseconds()) / 1e6
	}
	
	// Loss recovery efficiency
	if b.packetsLost > 0 {
		b.metrics.LossRecoveryEfficiency = CalculateLossRecoveryEfficiency(b.recoveredPackets, b.packetsLost)
	} else {
		b.metrics.LossRecoveryEfficiency = 1.0 // No losses = perfect efficiency
	}
}

// SetQlogCallback sets the callback for qlog events
func (b *BBRv3) SetQlogCallback(callback func(eventType string, data map[string]interface{})) {
	b.qlogCallback = callback
}

// logQlogEvent logs an event to qlog
func (b *BBRv3) logQlogEvent(eventType string, data map[string]interface{}) {
	if b.qlogCallback != nil {
		b.qlogCallback(eventType, data)
	}
}

// getStateString returns the string representation of the current state
func (b *BBRv3) getStateString() string {
	return b.getStateStringFromState(b.state)
}

// getStateStringFromState returns the string representation of a state
func (b *BBRv3) getStateStringFromState(state bbrv3State) string {
	switch state {
	case bbrv3Startup:
		return "Startup"
	case bbrv3Drain:
		return "Drain"
	case bbrv3ProbeBW:
		return "ProbeBW"
	case bbrv3ProbeRTT:
		return "ProbeRTT"
	default:
		return "Unknown"
	}
}
