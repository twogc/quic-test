package internal

import (
	"sync"

	"quic-test/internal/congestion"
	"quic-test/internal/integration"
)

// GlobalMetricsCollector collects metrics globally across all connections
type GlobalMetricsCollector struct {
	mu sync.RWMutex
	
	// Experimental integration for CC metrics
	experimentalIntegration *integration.SimpleIntegration
	
	// Multiple flow tracking for fairness
	flowThroughputs []float64
	flowMutex       sync.Mutex
}

var globalMetricsCollector *GlobalMetricsCollector
var globalMetricsCollectorOnce sync.Once

// GetGlobalMetricsCollector returns the global metrics collector
func GetGlobalMetricsCollector() *GlobalMetricsCollector {
	globalMetricsCollectorOnce.Do(func() {
		globalMetricsCollector = &GlobalMetricsCollector{
			flowThroughputs: make([]float64, 0),
		}
	})
	return globalMetricsCollector
}

// SetExperimentalIntegration sets the experimental integration for CC metrics collection
func (gmc *GlobalMetricsCollector) SetExperimentalIntegration(si *integration.SimpleIntegration) {
	gmc.mu.Lock()
	defer gmc.mu.Unlock()
	gmc.experimentalIntegration = si
}

// GetBBRv3Metrics retrieves BBRv3 metrics from experimental integration
func (gmc *GlobalMetricsCollector) GetBBRv3Metrics() map[string]interface{} {
	gmc.mu.RLock()
	defer gmc.mu.RUnlock()
	
	if gmc.experimentalIntegration == nil {
		return nil
	}
	
	return gmc.experimentalIntegration.GetBBRv3Metrics()
}

// RecordFlowThroughput records throughput for a flow (for fairness calculation)
func (gmc *GlobalMetricsCollector) RecordFlowThroughput(throughput float64) {
	gmc.flowMutex.Lock()
	defer gmc.flowMutex.Unlock()
	
	// Keep only last 100 flows for fairness calculation
	if len(gmc.flowThroughputs) >= 100 {
		gmc.flowThroughputs = gmc.flowThroughputs[1:]
	}
	gmc.flowThroughputs = append(gmc.flowThroughputs, throughput)
}

// GetFairnessIndex calculates Jain's Fairness Index from recorded flows
func (gmc *GlobalMetricsCollector) GetFairnessIndex() float64 {
	gmc.flowMutex.Lock()
	defer gmc.flowMutex.Unlock()
	
	return congestion.JainFairnessIndex(gmc.flowThroughputs)
}

// EnhanceMetricsMap adds BBRv3 and experimental metrics to metrics map
func EnhanceMetricsMap(metricsMap map[string]interface{}) map[string]interface{} {
	// Get BBRv3 metrics if available
	gmc := GetGlobalMetricsCollector()
	bbrv3Metrics := gmc.GetBBRv3Metrics()
	if bbrv3Metrics != nil {
		metricsMap["BBRv3Metrics"] = bbrv3Metrics
	}
	
	// Add fairness index if we have multiple flows
	fairnessIndex := gmc.GetFairnessIndex()
	if fairnessIndex > 0 {
		metricsMap["FairnessIndex"] = fairnessIndex
	}
	
	return metricsMap
}

