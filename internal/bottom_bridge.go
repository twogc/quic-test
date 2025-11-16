// Package bottom_bridge provides integration with QUIC Bottom TUI
package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BottomBridge handles communication with QUIC Bottom TUI
type BottomBridge struct {
	apiURL    string
	client    *http.Client
	enabled   bool
	lastSent  time.Time
	interval  time.Duration
}

// MetricsRequest represents the data sent to QUIC Bottom
type MetricsRequest struct {
	Latency           float64 `json:"latency"`
	Throughput        float64 `json:"throughput"`
	ThroughputMbps    float64 `json:"throughput_mbps,omitempty"`
	GoodputMbps       float64 `json:"goodput_mbps,omitempty"`
	Connections       int32   `json:"connections"`
	Errors            int32   `json:"errors"`
	PacketLoss        float64 `json:"packet_loss"`
	Retransmits       int32   `json:"retransmits"`
	RetransmissionRate float64 `json:"retransmission_rate,omitempty"`
	RTTP50Ms          float64 `json:"rtt_p50_ms,omitempty"`
	RTTP95Ms          float64 `json:"rtt_p95_ms,omitempty"`
	RTTP99Ms          float64 `json:"rtt_p99_ms,omitempty"`
	JitterMs          float64 `json:"jitter_ms,omitempty"`
	
	// BBRv3 specific metrics (optional, only when using BBRv3)
	BBRv3Phase              string            `json:"bbrv3_phase,omitempty"`                 // Startup, Drain, ProbeBW, ProbeRTT
	BBRv3BandwidthFast      float64           `json:"bbrv3_bw_fast,omitempty"`                // Fast-scale bandwidth (bps)
	BBRv3BandwidthSlow      float64           `json:"bbrv3_bw_slow,omitempty"`               // Slow-scale bandwidth (bps)
	BBRv3LossRateRound      float64           `json:"bbrv3_loss_rate_round,omitempty"`       // Loss rate per round
	BBRv3LossRateEMA        float64           `json:"bbrv3_loss_rate_ema,omitempty"`         // EMA loss rate
	BBRv3LossThreshold      float64           `json:"bbrv3_loss_threshold,omitempty"`         // Loss threshold (2%)
	BBRv3HeadroomUsage      float64           `json:"bbrv3_headroom_usage,omitempty"`         // Headroom usage (0.0-1.0)
	BBRv3InflightTarget     float64           `json:"bbrv3_inflight_target,omitempty"`      // Target inflight (bytes)
	BBRv3PacingQuantum      int64             `json:"bbrv3_pacing_quantum,omitempty"`        // Pacing quantum (bytes)
	BBRv3PacingGain         float64           `json:"bbrv3_pacing_gain,omitempty"`            // Current pacing gain
	BBRv3CWNDGain           float64           `json:"bbrv3_cwnd_gain,omitempty"`             // Current CWND gain
	BBRv3ProbeRTTMinMs      float64           `json:"bbrv3_probe_rtt_min_ms,omitempty"`       // Minimum RTT during ProbeRTT
	BBRv3BufferbloatFactor  float64           `json:"bbrv3_bufferbloat_factor,omitempty"`    // (avg_rtt / min_rtt) - 1
	BBRv3StabilityIndex      float64           `json:"bbrv3_stability_index,omitempty"`       // Δ throughput / Δ rtt
	BBRv3PhaseDurationMs    map[string]float64 `json:"bbrv3_phase_duration_ms,omitempty"`     // Duration of each phase
	BBRv3RecoveryTimeMs      float64           `json:"bbrv3_recovery_time_ms,omitempty"`      // Time to recover from loss
	BBRv3LossRecoveryEfficiency float64        `json:"bbrv3_loss_recovery_efficiency,omitempty"` // recovered / lost
}

// MetricsResponse represents the response from QUIC Bottom
type MetricsResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// NewBottomBridge creates a new bridge to QUIC Bottom
func NewBottomBridge(apiURL string, interval time.Duration) *BottomBridge {
	return &BottomBridge{
		apiURL:   apiURL,
		client:   &http.Client{Timeout: 5 * time.Second},
		enabled:  true,
		interval: interval,
	}
}

// UpdateMetrics sends metrics to QUIC Bottom TUI
func (bb *BottomBridge) UpdateMetrics(metrics map[string]interface{}) error {
	if !bb.enabled {
		return nil
	}

	// Check if enough time has passed since last update
	if time.Since(bb.lastSent) < bb.interval {
		return nil
	}

	// Extract metrics from the map
	latency := getFloat64(metrics, "Latency", 0.0)
	throughput := getFloat64(metrics, "ThroughputAverage", 0.0)
	throughputMbps := getFloat64(metrics, "ThroughputMbps", 0.0)
	goodputMbps := getFloat64(metrics, "GoodputMbps", 0.0)
	connections := getInt32(metrics, "Connections", 0)
	errors := getInt32(metrics, "Errors", 0)
	packetLoss := getFloat64(metrics, "PacketLoss", 0.0)
	retransmits := getInt32(metrics, "Retransmits", 0)
	retransmissionRate := getFloat64(metrics, "RetransmissionRate", 0.0)
	rttP50 := getFloat64(metrics, "RTTP50Ms", 0.0)
	rttP95 := getFloat64(metrics, "RTTP95Ms", 0.0)
	rttP99 := getFloat64(metrics, "RTTP99Ms", 0.0)
	jitterMs := getFloat64(metrics, "JitterMs", 0.0)
	
	// Extract BBRv3 metrics if available
	var bbrv3Phase string
	var bbrv3BwFast, bbrv3BwSlow, bbrv3LossRateRound, bbrv3LossRateEMA, bbrv3LossThreshold float64
	var bbrv3Headroom, bbrv3InflightTarget, bbrv3PacingGain, bbrv3CWNDGain float64
	var bbrv3ProbeRTTMin, bbrv3Bufferbloat, bbrv3Stability float64
	var bbrv3RecoveryTime, bbrv3RecoveryEfficiency float64
	var bbrv3PacingQuantum int64
	var bbrv3PhaseDuration map[string]float64
	
	if bbrv3Metrics, ok := metrics["BBRv3Metrics"]; ok {
		if bbrv3Map, ok := bbrv3Metrics.(map[string]interface{}); ok {
			if phase, ok := bbrv3Map["phase"].(string); ok {
				bbrv3Phase = phase
			}
			bbrv3BwFast = getFloat64(bbrv3Map, "bw_fast", 0.0)
			bbrv3BwSlow = getFloat64(bbrv3Map, "bw_slow", 0.0)
			bbrv3LossRateRound = getFloat64(bbrv3Map, "loss_rate_round", 0.0)
			bbrv3LossRateEMA = getFloat64(bbrv3Map, "loss_rate_ema", 0.0)
			bbrv3LossThreshold = getFloat64(bbrv3Map, "loss_threshold", 0.0)
			bbrv3Headroom = getFloat64(bbrv3Map, "headroom_usage", 0.0)
			bbrv3InflightTarget = getFloat64(bbrv3Map, "inflight_target", 0.0)
			bbrv3PacingGain = getFloat64(bbrv3Map, "pacing_gain", 0.0)
			bbrv3CWNDGain = getFloat64(bbrv3Map, "cwnd_gain", 0.0)
			bbrv3ProbeRTTMin = getFloat64(bbrv3Map, "probe_rtt_min_ms", 0.0)
			bbrv3Bufferbloat = getFloat64(bbrv3Map, "bufferbloat_factor", 0.0)
			bbrv3Stability = getFloat64(bbrv3Map, "stability_index", 0.0)
			bbrv3RecoveryTime = getFloat64(bbrv3Map, "recovery_time_ms", 0.0)
			bbrv3RecoveryEfficiency = getFloat64(bbrv3Map, "loss_recovery_efficiency", 0.0)
			
			if quantum, ok := bbrv3Map["pacing_quantum"].(float64); ok {
				bbrv3PacingQuantum = int64(quantum)
			}
			
			if phaseDur, ok := bbrv3Map["phase_duration_ms"].(map[string]interface{}); ok {
				bbrv3PhaseDuration = make(map[string]float64)
				for k, v := range phaseDur {
					if f, ok := v.(float64); ok {
						bbrv3PhaseDuration[k] = f
					}
				}
			}
		}
	}

	// Create request
	req := MetricsRequest{
		Latency:            latency,
		Throughput:         throughput,
		ThroughputMbps:     throughputMbps,
		GoodputMbps:        goodputMbps,
		Connections:        connections,
		Errors:             errors,
		PacketLoss:         packetLoss,
		Retransmits:        retransmits,
		RetransmissionRate: retransmissionRate,
		RTTP50Ms:           rttP50,
		RTTP95Ms:           rttP95,
		RTTP99Ms:           rttP99,
		JitterMs:           jitterMs,
	}
	
	// Add BBRv3 metrics if available
	if bbrv3Phase != "" {
		req.BBRv3Phase = bbrv3Phase
		req.BBRv3BandwidthFast = bbrv3BwFast
		req.BBRv3BandwidthSlow = bbrv3BwSlow
		req.BBRv3LossRateRound = bbrv3LossRateRound
		req.BBRv3LossRateEMA = bbrv3LossRateEMA
		req.BBRv3LossThreshold = bbrv3LossThreshold
		req.BBRv3HeadroomUsage = bbrv3Headroom
		req.BBRv3InflightTarget = bbrv3InflightTarget
		req.BBRv3PacingQuantum = bbrv3PacingQuantum
		req.BBRv3PacingGain = bbrv3PacingGain
		req.BBRv3CWNDGain = bbrv3CWNDGain
		req.BBRv3ProbeRTTMinMs = bbrv3ProbeRTTMin
		req.BBRv3BufferbloatFactor = bbrv3Bufferbloat
		req.BBRv3StabilityIndex = bbrv3Stability
		req.BBRv3RecoveryTimeMs = bbrv3RecoveryTime
		req.BBRv3LossRecoveryEfficiency = bbrv3RecoveryEfficiency
		if bbrv3PhaseDuration != nil {
			req.BBRv3PhaseDurationMs = bbrv3PhaseDuration
		}
	}

	// Send to QUIC Bottom
	if err := bb.sendMetrics(req); err != nil {
		// Log error but don't fail the main application
		fmt.Printf("Warning: Failed to send metrics to QUIC Bottom: %v\n", err)
		return nil
	}

	bb.lastSent = time.Now()
	return nil
}

// sendMetrics sends metrics to the QUIC Bottom API
func (bb *BottomBridge) sendMetrics(req MetricsRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %v", err)
	}

	resp, err := bb.client.Post(bb.apiURL+"/metrics", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("QUIC Bottom API returned status %d", resp.StatusCode)
	}

	// Parse response
	var metricsResp MetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&metricsResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if metricsResp.Status != "ok" {
		return fmt.Errorf("QUIC Bottom API error: %s", metricsResp.Message)
	}

	return nil
}

// Enable enables the bridge
func (bb *BottomBridge) Enable() {
	bb.enabled = true
}

// Disable disables the bridge
func (bb *BottomBridge) Disable() {
	bb.enabled = false
}

// IsEnabled returns whether the bridge is enabled
func (bb *BottomBridge) IsEnabled() bool {
	return bb.enabled
}

// SetInterval sets the update interval
func (bb *BottomBridge) SetInterval(interval time.Duration) {
	bb.interval = interval
}

// CheckHealth checks if QUIC Bottom is running
func (bb *BottomBridge) CheckHealth() error {
	resp, err := bb.client.Get(bb.apiURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to check health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("QUIC Bottom health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// Helper functions to safely extract values from interface{}
func getFloat64(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultValue
}

func getInt32(m map[string]interface{}, key string, defaultValue int32) int32 {
	if val, ok := m[key]; ok {
		if i, ok := val.(int32); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int32(f)
		}
	}
	return defaultValue
}

// Global bridge instance
var globalBottomBridge *BottomBridge

// InitBottomBridge initializes the global bridge
func InitBottomBridge(apiURL string, interval time.Duration) {
	globalBottomBridge = NewBottomBridge(apiURL, interval)
}

// UpdateBottomMetrics updates metrics via the global bridge
func UpdateBottomMetrics(metrics map[string]interface{}) {
	if globalBottomBridge != nil {
		globalBottomBridge.UpdateMetrics(metrics)
	}
}

// EnableBottomBridge enables the global bridge
func EnableBottomBridge() {
	if globalBottomBridge != nil {
		globalBottomBridge.Enable()
	}
}

// DisableBottomBridge disables the global bridge
func DisableBottomBridge() {
	if globalBottomBridge != nil {
		globalBottomBridge.Disable()
	}
}
