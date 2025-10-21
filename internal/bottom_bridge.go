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
	Latency      float64 `json:"latency"`
	Throughput   float64 `json:"throughput"`
	Connections  int32   `json:"connections"`
	Errors       int32   `json:"errors"`
	PacketLoss   float64 `json:"packet_loss"`
	Retransmits  int32   `json:"retransmits"`
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
	connections := getInt32(metrics, "Connections", 0)
	errors := getInt32(metrics, "Errors", 0)
	packetLoss := getFloat64(metrics, "PacketLoss", 0.0)
	retransmits := getInt32(metrics, "Retransmits", 0)

	// Create request
	req := MetricsRequest{
		Latency:     latency,
		Throughput:  throughput,
		Connections: connections,
		Errors:      errors,
		PacketLoss:  packetLoss,
		Retransmits: retransmits,
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
