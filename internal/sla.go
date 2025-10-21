package internal

import (
	"fmt"
	"os"
	"time"
)

// SLAExitCode определяет exit code на основе SLA проверок
type SLAExitCode int

const (
	ExitCodeSuccess SLAExitCode = 0
	ExitCodeSLAFailure SLAExitCode = 1
	ExitCodeCriticalFailure SLAExitCode = 2
)

// SLAViolationType определяет тип нарушения SLA
type SLAViolationType string

const (
	ViolationRTT SLAViolationType = "rtt_p95"
	ViolationLoss SLAViolationType = "packet_loss"
	ViolationThroughput SLAViolationType = "throughput"
	ViolationErrors SLAViolationType = "errors"
)

// SLAViolationInfo описывает нарушение SLA (используем другое имя)
type SLAViolationInfo struct {
	Type      SLAViolationType `json:"type"`
	Expected  interface{}      `json:"expected"`
	Actual    interface{}      `json:"actual"`
	Severity  string           `json:"severity"` // "warning", "critical"
	Message   string           `json:"message"`
}

// CheckSLA проверяет соответствие метрик SLA требованиям
func CheckSLA(cfg TestConfig, metrics map[string]interface{}) (bool, []SLAViolationInfo, SLAExitCode) {
	var violations []SLAViolationInfo
	hasCriticalViolations := false
	
	// Проверяем RTT p95
	if cfg.SlaRttP95 > 0 {
		latencies, _ := metrics["Latencies"].([]float64)
		if len(latencies) > 0 {
			_, p95, _ := calcPercentiles(latencies)
			actualRTT := time.Duration(p95 * float64(time.Millisecond))
			
			if actualRTT > cfg.SlaRttP95 {
				violation := SLAViolationInfo{
					Type:     ViolationRTT,
					Expected: cfg.SlaRttP95,
					Actual:   actualRTT,
					Severity: "critical",
					Message:  fmt.Sprintf("RTT p95 %v exceeds SLA limit %v", actualRTT, cfg.SlaRttP95),
				}
				violations = append(violations, violation)
				hasCriticalViolations = true
			}
		}
	}
	
	// Проверяем потерю пакетов
	if cfg.SlaLoss > 0 {
		packetLoss := getFloat64FromSchema(metrics, "PacketLoss")
		if packetLoss > cfg.SlaLoss {
			violation := SLAViolationInfo{
				Type:     ViolationLoss,
				Expected: cfg.SlaLoss,
				Actual:   packetLoss,
				Severity: "critical",
				Message:  fmt.Sprintf("Packet loss %.2f%% exceeds SLA limit %.2f%%", packetLoss*100, cfg.SlaLoss*100),
			}
			violations = append(violations, violation)
			hasCriticalViolations = true
		}
	}
	
	// Проверяем пропускную способность
	if cfg.SlaThroughput > 0 {
		throughput := getFloat64FromSchema(metrics, "ThroughputAverage")
		if throughput < cfg.SlaThroughput {
			violation := SLAViolationInfo{
				Type:     ViolationThroughput,
				Expected: cfg.SlaThroughput,
				Actual:   throughput,
				Severity: "critical",
				Message:  fmt.Sprintf("Throughput %.2f KB/s below SLA limit %.2f KB/s", throughput, cfg.SlaThroughput),
			}
			violations = append(violations, violation)
			hasCriticalViolations = true
		}
	}
	
	// Проверяем количество ошибок
	if cfg.SlaErrors > 0 {
		errors := getInt64(metrics, "Errors")
		if errors > cfg.SlaErrors {
			violation := SLAViolationInfo{
				Type:     ViolationErrors,
				Expected: cfg.SlaErrors,
				Actual:   errors,
				Severity: "critical",
				Message:  fmt.Sprintf("Error count %d exceeds SLA limit %d", errors, cfg.SlaErrors),
			}
			violations = append(violations, violation)
			hasCriticalViolations = true
		}
	}
	
	// Определяем exit code
	var exitCode SLAExitCode
	if len(violations) == 0 {
		exitCode = ExitCodeSuccess
	} else if hasCriticalViolations {
		exitCode = ExitCodeCriticalFailure
	} else {
		exitCode = ExitCodeSLAFailure
	}
	
	passed := len(violations) == 0
	return passed, violations, exitCode
}

// ExitWithSLA проверяет SLA и завершает программу с соответствующим exit code
func ExitWithSLA(cfg TestConfig, metrics map[string]interface{}) {
	passed, violations, exitCode := CheckSLA(cfg, metrics)
	
	if !passed {
		fmt.Printf("\n❌ SLA проверки не пройдены:\n")
		for _, violation := range violations {
			fmt.Printf("  - %s: %s\n", violation.Type, violation.Message)
		}
	} else {
		fmt.Printf("\n✅ Все SLA проверки пройдены успешно\n")
	}
	
	// Выводим детальную информацию о SLA
	if cfg.SlaRttP95 > 0 || cfg.SlaLoss > 0 || cfg.SlaThroughput > 0 || cfg.SlaErrors > 0 {
		fmt.Printf("\n📊 SLA Summary:\n")
		if cfg.SlaRttP95 > 0 {
			latencies, _ := metrics["Latencies"].([]float64)
			if len(latencies) > 0 {
				_, p95, _ := calcPercentiles(latencies)
				actualRTT := time.Duration(p95 * float64(time.Millisecond))
				status := "✅"
				if actualRTT > cfg.SlaRttP95 {
					status = "❌"
				}
				fmt.Printf("  %s RTT p95: %v (limit: %v)\n", status, actualRTT, cfg.SlaRttP95)
			}
		}
		
		if cfg.SlaLoss > 0 {
			packetLoss := getFloat64FromSchema(metrics, "PacketLoss")
			status := "✅"
			if packetLoss > cfg.SlaLoss {
				status = "❌"
			}
			fmt.Printf("  %s Packet Loss: %.2f%% (limit: %.2f%%)\n", status, packetLoss*100, cfg.SlaLoss*100)
		}
		
		if cfg.SlaThroughput > 0 {
			throughput := getFloat64FromSchema(metrics, "ThroughputAverage")
			status := "✅"
			if throughput < cfg.SlaThroughput {
				status = "❌"
			}
			fmt.Printf("  %s Throughput: %.2f KB/s (limit: %.2f KB/s)\n", status, throughput, cfg.SlaThroughput)
		}
		
		if cfg.SlaErrors > 0 {
			errors := getInt64(metrics, "Errors")
			status := "✅"
			if errors > cfg.SlaErrors {
				status = "❌"
			}
			fmt.Printf("  %s Errors: %d (limit: %d)\n", status, errors, cfg.SlaErrors)
		}
	}
	
	os.Exit(int(exitCode))
}

// PrintSLAConfig выводит информацию о настроенных SLA параметрах
func PrintSLAConfig(cfg TestConfig) {
	if cfg.SlaRttP95 > 0 || cfg.SlaLoss > 0 || cfg.SlaThroughput > 0 || cfg.SlaErrors > 0 {
		fmt.Printf("🎯 SLA Configuration:\n")
		if cfg.SlaRttP95 > 0 {
			fmt.Printf("  - RTT p95 limit: %v\n", cfg.SlaRttP95)
		}
		if cfg.SlaLoss > 0 {
			fmt.Printf("  - Packet loss limit: %.2f%%\n", cfg.SlaLoss*100)
		}
		if cfg.SlaThroughput > 0 {
			fmt.Printf("  - Throughput limit: %.2f KB/s\n", cfg.SlaThroughput)
		}
		if cfg.SlaErrors > 0 {
			fmt.Printf("  - Error count limit: %d\n", cfg.SlaErrors)
		}
		fmt.Println()
	}
}
