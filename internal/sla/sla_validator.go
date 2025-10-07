package sla

import (
	"fmt"
	"strings"
)

// SLAValidator проверяет метрики против SLA-гейтов
type SLAValidator struct {
	gates *SLAGates
}

// NewSLAValidator создает новый валидатор SLA
func NewSLAValidator(gates *SLAGates) *SLAValidator {
	return &SLAValidator{
		gates: gates,
	}
}

// Validate проверяет метрики против SLA-гейтов
func (v *SLAValidator) Validate(metrics SLAMetrics) *SLAResult {
	result := &SLAResult{
		Passed:     true,
		Score:      1.0,
		Violations: make([]SLAViolation, 0),
		Metrics:    metrics,
		Summary:    "",
	}
	
	// Проверяем RTT метрики
	v.validateRTT(metrics, result)
	
	// Проверяем Loss метрики
	v.validateLoss(metrics, result)
	
	// Проверяем Throughput метрики
	v.validateThroughput(metrics, result)
	
	// Проверяем Congestion Control метрики
	v.validateCongestionControl(metrics, result)
	
	// Проверяем ACK метрики
	v.validateACK(metrics, result)
	
	// Проверяем FEC метрики
	v.validateFEC(metrics, result)
	
	// Вычисляем общий score
	v.calculateScore(result)
	
	// Генерируем summary
	v.generateSummary(result)
	
	return result
}

// validateRTT проверяет RTT метрики
func (v *SLAValidator) validateRTT(metrics SLAMetrics, result *SLAResult) {
	// Проверяем 95-й процентиль RTT
	if metrics.RTTPercentile95Ms > v.gates.P95RTTMs {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "rtt_p95",
			Expected: v.gates.P95RTTMs,
			Actual:   metrics.RTTPercentile95Ms,
			Severity: "critical",
			Message:  fmt.Sprintf("95th percentile RTT %.2fms exceeds limit %.2fms", metrics.RTTPercentile95Ms, v.gates.P95RTTMs),
		})
		result.Passed = false
	}
	
	// Проверяем максимальный RTT
	if metrics.RTTMaxMs > v.gates.MaxRTTMs {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "rtt_max",
			Expected: v.gates.MaxRTTMs,
			Actual:   metrics.RTTMaxMs,
			Severity: "critical",
			Message:  fmt.Sprintf("Maximum RTT %.2fms exceeds limit %.2fms", metrics.RTTMaxMs, v.gates.MaxRTTMs),
		})
		result.Passed = false
	}
	
	// Проверяем средний RTT
	if metrics.RTTMeanMs > v.gates.MeanRTTMs {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "rtt_mean",
			Expected: v.gates.MeanRTTMs,
			Actual:   metrics.RTTMeanMs,
			Severity: "warning",
			Message:  fmt.Sprintf("Mean RTT %.2fms exceeds limit %.2fms", metrics.RTTMeanMs, v.gates.MeanRTTMs),
		})
	}
}

// validateLoss проверяет метрики потерь
func (v *SLAValidator) validateLoss(metrics SLAMetrics, result *SLAResult) {
	// Проверяем процент потерь
	if metrics.LossRatePercent > v.gates.LossRatePercent {
		severity := "warning"
		if metrics.LossRatePercent > v.gates.MaxLossRate {
			severity = "critical"
			result.Passed = false
		}
		
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "loss_rate",
			Expected: v.gates.LossRatePercent,
			Actual:   metrics.LossRatePercent,
			Severity: severity,
			Message:  fmt.Sprintf("Loss rate %.2f%% exceeds limit %.2f%%", metrics.LossRatePercent, v.gates.LossRatePercent),
		})
	}
}

// validateThroughput проверяет метрики пропускной способности
func (v *SLAValidator) validateThroughput(metrics SLAMetrics, result *SLAResult) {
	// Проверяем goodput
	if metrics.GoodputMbps < v.gates.MinGoodputMbps {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "goodput",
			Expected: v.gates.MinGoodputMbps,
			Actual:   metrics.GoodputMbps,
			Severity: "critical",
			Message:  fmt.Sprintf("Goodput %.2f Mbps below minimum %.2f Mbps", metrics.GoodputMbps, v.gates.MinGoodputMbps),
		})
		result.Passed = false
	}
	
	// Проверяем throughput
	if metrics.ThroughputMbps < v.gates.MinThroughputMbps {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "throughput",
			Expected: v.gates.MinThroughputMbps,
			Actual:   metrics.ThroughputMbps,
			Severity: "warning",
			Message:  fmt.Sprintf("Throughput %.2f Mbps below minimum %.2f Mbps", metrics.ThroughputMbps, v.gates.MinThroughputMbps),
		})
	}
}

// validateCongestionControl проверяет метрики congestion control
func (v *SLAValidator) validateCongestionControl(metrics SLAMetrics, result *SLAResult) {
	// Проверяем пропускную способность
	if metrics.BandwidthBps < v.gates.MinBandwidthBps {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "bandwidth",
			Expected: v.gates.MinBandwidthBps,
			Actual:   metrics.BandwidthBps,
			Severity: "warning",
			Message:  fmt.Sprintf("Bandwidth %.0f bps below minimum %.0f bps", metrics.BandwidthBps, v.gates.MinBandwidthBps),
		})
	}
	
	// Проверяем congestion window
	if metrics.CWNDBytes < int64(v.gates.MinCWNDBytes) {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "cwnd_min",
			Expected: float64(v.gates.MinCWNDBytes),
			Actual:   float64(metrics.CWNDBytes),
			Severity: "warning",
			Message:  fmt.Sprintf("CWND %d bytes below minimum %d bytes", metrics.CWNDBytes, v.gates.MinCWNDBytes),
		})
	}
	
	if metrics.CWNDBytes > int64(v.gates.MaxCWNDBytes) {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "cwnd_max",
			Expected: float64(v.gates.MaxCWNDBytes),
			Actual:   float64(metrics.CWNDBytes),
			Severity: "warning",
			Message:  fmt.Sprintf("CWND %d bytes above maximum %d bytes", metrics.CWNDBytes, v.gates.MaxCWNDBytes),
		})
	}
}

// validateACK проверяет метрики ACK
func (v *SLAValidator) validateACK(metrics SLAMetrics, result *SLAResult) {
	// Проверяем задержку ACK
	if metrics.ACKDelayMs > v.gates.MaxACKDelayMs {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "ack_delay",
			Expected: v.gates.MaxACKDelayMs,
			Actual:   metrics.ACKDelayMs,
			Severity: "warning",
			Message:  fmt.Sprintf("ACK delay %.2fms exceeds limit %.2fms", metrics.ACKDelayMs, v.gates.MaxACKDelayMs),
		})
	}
	
	// Проверяем частоту ACK
	if metrics.ACKFrequency < v.gates.MinACKFrequency {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "ack_frequency",
			Expected: float64(v.gates.MinACKFrequency),
			Actual:   float64(metrics.ACKFrequency),
			Severity: "warning",
			Message:  fmt.Sprintf("ACK frequency %d below minimum %d", metrics.ACKFrequency, v.gates.MinACKFrequency),
		})
	}
}

// validateFEC проверяет метрики FEC
func (v *SLAValidator) validateFEC(metrics SLAMetrics, result *SLAResult) {
	// Проверяем избыточность FEC
	if metrics.FECRedundancy > v.gates.MaxFECRedundancy {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "fec_redundancy",
			Expected: v.gates.MaxFECRedundancy,
			Actual:   metrics.FECRedundancy,
			Severity: "warning",
			Message:  fmt.Sprintf("FEC redundancy %.2f%% exceeds limit %.2f%%", metrics.FECRedundancy*100, v.gates.MaxFECRedundancy*100),
		})
	}
	
	// Проверяем скорость восстановления FEC
	if metrics.FECRecoveryRate < v.gates.MinFECRecoveryRate {
		result.Violations = append(result.Violations, SLAViolation{
			Metric:   "fec_recovery",
			Expected: v.gates.MinFECRecoveryRate,
			Actual:   metrics.FECRecoveryRate,
			Severity: "warning",
			Message:  fmt.Sprintf("FEC recovery rate %.2f%% below minimum %.2f%%", metrics.FECRecoveryRate*100, v.gates.MinFECRecoveryRate*100),
		})
	}
}

// calculateScore вычисляет общий score (0.0 - 1.0)
func (v *SLAValidator) calculateScore(result *SLAResult) {
	if len(result.Violations) == 0 {
		result.Score = 1.0
		return
	}
	
	// Штрафы за нарушения
	criticalPenalty := 0.0
	warningPenalty := 0.0
	infoPenalty := 0.0
	
	for _, violation := range result.Violations {
		switch violation.Severity {
		case "critical":
			criticalPenalty += 0.3
		case "warning":
			warningPenalty += 0.1
		case "info":
			infoPenalty += 0.05
		}
	}
	
	// Вычисляем score
	result.Score = 1.0 - criticalPenalty - warningPenalty - infoPenalty
	if result.Score < 0.0 {
		result.Score = 0.0
	}
}

// generateSummary генерирует текстовое описание результата
func (v *SLAValidator) generateSummary(result *SLAResult) {
	if result.Passed {
		result.Summary = fmt.Sprintf("✅ SLA PASSED (Score: %.2f) - All metrics within acceptable limits", result.Score)
	} else {
		criticalCount := 0
		warningCount := 0
		infoCount := 0
		
		for _, violation := range result.Violations {
			switch violation.Severity {
			case "critical":
				criticalCount++
			case "warning":
				warningCount++
			case "info":
				infoCount++
			}
		}
		
		result.Summary = fmt.Sprintf("❌ SLA FAILED (Score: %.2f) - %d critical, %d warning, %d info violations", 
			result.Score, criticalCount, warningCount, infoCount)
	}
}

// GetDetailedReport возвращает детальный отчет о нарушениях
func (v *SLAValidator) GetDetailedReport(result *SLAResult) string {
	var report strings.Builder
	
	report.WriteString(fmt.Sprintf("SLA Validation Report\n"))
	report.WriteString(fmt.Sprintf("====================\n"))
	report.WriteString(fmt.Sprintf("Result: %s\n", result.Summary))
	report.WriteString(fmt.Sprintf("Score: %.2f/1.0\n\n", result.Score))
	
	if len(result.Violations) == 0 {
		report.WriteString("✅ No violations found - all metrics within SLA limits\n")
		return report.String()
	}
	
	report.WriteString("Violations:\n")
	report.WriteString("-----------\n")
	
	for i, violation := range result.Violations {
		report.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, violation.Severity, violation.Metric))
		report.WriteString(fmt.Sprintf("   Expected: %.2f, Actual: %.2f\n", violation.Expected, violation.Actual))
		report.WriteString(fmt.Sprintf("   %s\n\n", violation.Message))
	}
	
	return report.String()
}

