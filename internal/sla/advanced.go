package sla

import (
	"context"
	"fmt"
	"time"
)

// AdvancedSLAChecker предоставляет расширенные SLA проверки
type AdvancedSLAChecker struct {
	config AdvancedSLAConfig
}

// AdvancedSLAConfig содержит конфигурацию расширенных SLA
type AdvancedSLAConfig struct {
	// Latency SLA
	RTT_P95_Max    time.Duration `json:"rtt_p95_max"`
	RTT_P99_Max    time.Duration `json:"rtt_p99_max"`
	RTT_P999_Max   time.Duration `json:"rtt_p999_max"`
	
	// Jitter SLA
	Jitter_P95_Max time.Duration `json:"jitter_p95_max"`
	Jitter_P99_Max time.Duration `json:"jitter_p99_max"`
	
	// Loss SLA
	Loss_Max       float64 `json:"loss_max_percent"`
	Loss_Burst_Max float64 `json:"loss_burst_max_percent"`
	
	// Throughput SLA
	Min_Throughput_Mbps float64 `json:"min_throughput_mbps"`
	Min_Goodput_Mbps    float64 `json:"min_goodput_mbps"`
	
	// Error SLA
	Max_Errors_Total    int64   `json:"max_errors_total"`
	Max_Errors_Percent  float64 `json:"max_errors_percent"`
	Max_Retransmits     int64   `json:"max_retransmits"`
	
	// QUIC-specific SLA
	Max_Handshake_Time  time.Duration `json:"max_handshake_time"`
	Min_ZeroRTT_Success float64       `json:"min_zero_rtt_success_percent"`
	Max_Stream_Resets   int64         `json:"max_stream_resets"`
	Max_Connection_Drops int64        `json:"max_connection_drops"`
	
	// Availability SLA
	Min_Uptime_Percent  float64 `json:"min_uptime_percent"`
	Max_Downtime_Seconds int64  `json:"max_downtime_seconds"`
}

// SLAMetrics содержит метрики для SLA проверки
type SLAMetrics struct {
	// Latency metrics
	RTT_P50   time.Duration `json:"rtt_p50"`
	RTT_P95   time.Duration `json:"rtt_p95"`
	RTT_P99   time.Duration `json:"rtt_p99"`
	RTT_P999  time.Duration `json:"rtt_p999"`
	
	// Jitter metrics
	Jitter_P50  time.Duration `json:"jitter_p50"`
	Jitter_P95  time.Duration `json:"jitter_p95"`
	Jitter_P99  time.Duration `json:"jitter_p99"`
	
	// Loss metrics
	Loss_Percent     float64 `json:"loss_percent"`
	Loss_Burst_Percent float64 `json:"loss_burst_percent"`
	
	// Throughput metrics
	Throughput_Mbps float64 `json:"throughput_mbps"`
	Goodput_Mbps    float64 `json:"goodput_mbps"`
	
	// Error metrics
	Errors_Total     int64   `json:"errors_total"`
	Errors_Percent   float64 `json:"errors_percent"`
	Retransmits      int64   `json:"retransmits"`
	
	// QUIC-specific metrics
	Handshake_Time     time.Duration `json:"handshake_time"`
	ZeroRTT_Success    float64       `json:"zero_rtt_success_percent"`
	Stream_Resets      int64         `json:"stream_resets"`
	Connection_Drops   int64         `json:"connection_drops"`
	
	// Availability metrics
	Uptime_Percent     float64 `json:"uptime_percent"`
	Downtime_Seconds   int64   `json:"downtime_seconds"`
}

// SLAViolationType определяет тип нарушения SLA
type SLAViolationType string

const (
	ViolationRTT_P95     SLAViolationType = "rtt_p95"
	ViolationRTT_P99     SLAViolationType = "rtt_p99"
	ViolationRTT_P999    SLAViolationType = "rtt_p999"
	ViolationJitter_P95  SLAViolationType = "jitter_p95"
	ViolationJitter_P99  SLAViolationType = "jitter_p99"
	ViolationLoss        SLAViolationType = "loss"
	ViolationLossBurst   SLAViolationType = "loss_burst"
	ViolationThroughput  SLAViolationType = "throughput"
	ViolationGoodput     SLAViolationType = "goodput"
	ViolationErrors      SLAViolationType = "errors"
	ViolationRetransmits SLAViolationType = "retransmits"
	ViolationHandshake   SLAViolationType = "handshake_time"
	ViolationZeroRTT     SLAViolationType = "zero_rtt_success"
	ViolationStreamResets SLAViolationType = "stream_resets"
	ViolationConnDrops   SLAViolationType = "connection_drops"
	ViolationUptime      SLAViolationType = "uptime"
	ViolationDowntime    SLAViolationType = "downtime"
)

// SLAViolation описывает нарушение SLA
type SLAViolation struct {
	Type      SLAViolationType `json:"type"`
	Expected  interface{}      `json:"expected"`
	Actual    interface{}      `json:"actual"`
	Severity  string           `json:"severity"` // "warning", "critical"
	Message   string           `json:"message"`
	Timestamp time.Time        `json:"timestamp"`
	Impact    string           `json:"impact"` // "performance", "availability", "reliability"
}

// SLAResult содержит результат проверки SLA
type SLAResult struct {
	Passed     bool            `json:"passed"`
	Violations []SLAViolation  `json:"violations"`
	ExitCode   int             `json:"exit_code"`
	Summary    SLASummary      `json:"summary"`
}

// SLASummary содержит сводку по SLA
type SLASummary struct {
	TotalChecks     int `json:"total_checks"`
	PassedChecks    int `json:"passed_checks"`
	FailedChecks    int `json:"failed_checks"`
	WarningChecks   int `json:"warning_checks"`
	CriticalChecks  int `json:"critical_checks"`
}

// NewAdvancedSLAChecker создает новый расширенный SLA чекер
func NewAdvancedSLAChecker(config AdvancedSLAConfig) *AdvancedSLAChecker {
	return &AdvancedSLAChecker{
		config: config,
	}
}

// Check выполняет расширенные SLA проверки
func (c *AdvancedSLAChecker) Check(ctx context.Context, metrics SLAMetrics) *SLAResult {
	result := &SLAResult{
		Passed:     true,
		Violations: make([]SLAViolation, 0),
		ExitCode:   0,
		Summary: SLASummary{
			TotalChecks: 0,
		},
	}

	// Проверяем RTT P95
	if c.config.RTT_P95_Max > 0 {
		result.Summary.TotalChecks++
		if metrics.RTT_P95 > c.config.RTT_P95_Max {
			violation := SLAViolation{
				Type:      ViolationRTT_P95,
				Expected:  c.config.RTT_P95_Max,
				Actual:    metrics.RTT_P95,
				Severity:  "critical",
				Message:   fmt.Sprintf("RTT P95 %v exceeds SLA limit %v", metrics.RTT_P95, c.config.RTT_P95_Max),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем RTT P99
	if c.config.RTT_P99_Max > 0 {
		result.Summary.TotalChecks++
		if metrics.RTT_P99 > c.config.RTT_P99_Max {
			violation := SLAViolation{
				Type:      ViolationRTT_P99,
				Expected:  c.config.RTT_P99_Max,
				Actual:    metrics.RTT_P99,
				Severity:  "critical",
				Message:   fmt.Sprintf("RTT P99 %v exceeds SLA limit %v", metrics.RTT_P99, c.config.RTT_P99_Max),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем RTT P999
	if c.config.RTT_P999_Max > 0 {
		result.Summary.TotalChecks++
		if metrics.RTT_P999 > c.config.RTT_P999_Max {
			violation := SLAViolation{
				Type:      ViolationRTT_P999,
				Expected:  c.config.RTT_P999_Max,
				Actual:    metrics.RTT_P999,
				Severity:  "warning",
				Message:   fmt.Sprintf("RTT P999 %v exceeds SLA limit %v", metrics.RTT_P999, c.config.RTT_P999_Max),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Summary.WarningChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем Jitter P95
	if c.config.Jitter_P95_Max > 0 {
		result.Summary.TotalChecks++
		if metrics.Jitter_P95 > c.config.Jitter_P95_Max {
			violation := SLAViolation{
				Type:      ViolationJitter_P95,
				Expected:  c.config.Jitter_P95_Max,
				Actual:    metrics.Jitter_P95,
				Severity:  "critical",
				Message:   fmt.Sprintf("Jitter P95 %v exceeds SLA limit %v", metrics.Jitter_P95, c.config.Jitter_P95_Max),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем Jitter P99
	if c.config.Jitter_P99_Max > 0 {
		result.Summary.TotalChecks++
		if metrics.Jitter_P99 > c.config.Jitter_P99_Max {
			violation := SLAViolation{
				Type:      ViolationJitter_P99,
				Expected:  c.config.Jitter_P99_Max,
				Actual:    metrics.Jitter_P99,
				Severity:  "warning",
				Message:   fmt.Sprintf("Jitter P99 %v exceeds SLA limit %v", metrics.Jitter_P99, c.config.Jitter_P99_Max),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Summary.WarningChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем потери пакетов
	if c.config.Loss_Max > 0 {
		result.Summary.TotalChecks++
		if metrics.Loss_Percent > c.config.Loss_Max {
			violation := SLAViolation{
				Type:      ViolationLoss,
				Expected:  c.config.Loss_Max,
				Actual:    metrics.Loss_Percent,
				Severity:  "critical",
				Message:   fmt.Sprintf("Packet loss %.2f%% exceeds SLA limit %.2f%%", metrics.Loss_Percent, c.config.Loss_Max),
				Timestamp: time.Now(),
				Impact:    "reliability",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем всплески потерь
	if c.config.Loss_Burst_Max > 0 {
		result.Summary.TotalChecks++
		if metrics.Loss_Burst_Percent > c.config.Loss_Burst_Max {
			violation := SLAViolation{
				Type:      ViolationLossBurst,
				Expected:  c.config.Loss_Burst_Max,
				Actual:    metrics.Loss_Burst_Percent,
				Severity:  "warning",
				Message:   fmt.Sprintf("Packet loss burst %.2f%% exceeds SLA limit %.2f%%", metrics.Loss_Burst_Percent, c.config.Loss_Burst_Max),
				Timestamp: time.Now(),
				Impact:    "reliability",
			}
			result.Violations = append(result.Violations, violation)
			result.Summary.WarningChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем пропускную способность
	if c.config.Min_Throughput_Mbps > 0 {
		result.Summary.TotalChecks++
		if metrics.Throughput_Mbps < c.config.Min_Throughput_Mbps {
			violation := SLAViolation{
				Type:      ViolationThroughput,
				Expected:  c.config.Min_Throughput_Mbps,
				Actual:    metrics.Throughput_Mbps,
				Severity:  "critical",
				Message:   fmt.Sprintf("Throughput %.2f Mbps below SLA minimum %.2f Mbps", metrics.Throughput_Mbps, c.config.Min_Throughput_Mbps),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем goodput
	if c.config.Min_Goodput_Mbps > 0 {
		result.Summary.TotalChecks++
		if metrics.Goodput_Mbps < c.config.Min_Goodput_Mbps {
			violation := SLAViolation{
				Type:      ViolationGoodput,
				Expected:  c.config.Min_Goodput_Mbps,
				Actual:    metrics.Goodput_Mbps,
				Severity:  "critical",
				Message:   fmt.Sprintf("Goodput %.2f Mbps below SLA minimum %.2f Mbps", metrics.Goodput_Mbps, c.config.Min_Goodput_Mbps),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем общее количество ошибок
	if c.config.Max_Errors_Total > 0 {
		result.Summary.TotalChecks++
		if metrics.Errors_Total > c.config.Max_Errors_Total {
			violation := SLAViolation{
				Type:      ViolationErrors,
				Expected:  c.config.Max_Errors_Total,
				Actual:    metrics.Errors_Total,
				Severity:  "critical",
				Message:   fmt.Sprintf("Total errors %d exceeds SLA limit %d", metrics.Errors_Total, c.config.Max_Errors_Total),
				Timestamp: time.Now(),
				Impact:    "reliability",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем процент ошибок
	if c.config.Max_Errors_Percent > 0 {
		result.Summary.TotalChecks++
		if metrics.Errors_Percent > c.config.Max_Errors_Percent {
			violation := SLAViolation{
				Type:      ViolationErrors,
				Expected:  c.config.Max_Errors_Percent,
				Actual:    metrics.Errors_Percent,
				Severity:  "critical",
				Message:   fmt.Sprintf("Error rate %.2f%% exceeds SLA limit %.2f%%", metrics.Errors_Percent, c.config.Max_Errors_Percent),
				Timestamp: time.Now(),
				Impact:    "reliability",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем ретрансмиссии
	if c.config.Max_Retransmits > 0 {
		result.Summary.TotalChecks++
		if metrics.Retransmits > c.config.Max_Retransmits {
			violation := SLAViolation{
				Type:      ViolationRetransmits,
				Expected:  c.config.Max_Retransmits,
				Actual:    metrics.Retransmits,
				Severity:  "warning",
				Message:   fmt.Sprintf("Retransmits %d exceeds SLA limit %d", metrics.Retransmits, c.config.Max_Retransmits),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Summary.WarningChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем время handshake
	if c.config.Max_Handshake_Time > 0 {
		result.Summary.TotalChecks++
		if metrics.Handshake_Time > c.config.Max_Handshake_Time {
			violation := SLAViolation{
				Type:      ViolationHandshake,
				Expected:  c.config.Max_Handshake_Time,
				Actual:    metrics.Handshake_Time,
				Severity:  "critical",
				Message:   fmt.Sprintf("Handshake time %v exceeds SLA limit %v", metrics.Handshake_Time, c.config.Max_Handshake_Time),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем успешность 0-RTT
	if c.config.Min_ZeroRTT_Success > 0 {
		result.Summary.TotalChecks++
		if metrics.ZeroRTT_Success < c.config.Min_ZeroRTT_Success {
			violation := SLAViolation{
				Type:      ViolationZeroRTT,
				Expected:  c.config.Min_ZeroRTT_Success,
				Actual:    metrics.ZeroRTT_Success,
				Severity:  "warning",
				Message:   fmt.Sprintf("0-RTT success rate %.2f%% below SLA minimum %.2f%%", metrics.ZeroRTT_Success, c.config.Min_ZeroRTT_Success),
				Timestamp: time.Now(),
				Impact:    "performance",
			}
			result.Violations = append(result.Violations, violation)
			result.Summary.WarningChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем сбросы потоков
	if c.config.Max_Stream_Resets > 0 {
		result.Summary.TotalChecks++
		if metrics.Stream_Resets > c.config.Max_Stream_Resets {
			violation := SLAViolation{
				Type:      ViolationStreamResets,
				Expected:  c.config.Max_Stream_Resets,
				Actual:    metrics.Stream_Resets,
				Severity:  "warning",
				Message:   fmt.Sprintf("Stream resets %d exceeds SLA limit %d", metrics.Stream_Resets, c.config.Max_Stream_Resets),
				Timestamp: time.Now(),
				Impact:    "reliability",
			}
			result.Violations = append(result.Violations, violation)
			result.Summary.WarningChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем обрывы соединений
	if c.config.Max_Connection_Drops > 0 {
		result.Summary.TotalChecks++
		if metrics.Connection_Drops > c.config.Max_Connection_Drops {
			violation := SLAViolation{
				Type:      ViolationConnDrops,
				Expected:  c.config.Max_Connection_Drops,
				Actual:    metrics.Connection_Drops,
				Severity:  "critical",
				Message:   fmt.Sprintf("Connection drops %d exceeds SLA limit %d", metrics.Connection_Drops, c.config.Max_Connection_Drops),
				Timestamp: time.Now(),
				Impact:    "availability",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем uptime
	if c.config.Min_Uptime_Percent > 0 {
		result.Summary.TotalChecks++
		if metrics.Uptime_Percent < c.config.Min_Uptime_Percent {
			violation := SLAViolation{
				Type:      ViolationUptime,
				Expected:  c.config.Min_Uptime_Percent,
				Actual:    metrics.Uptime_Percent,
				Severity:  "critical",
				Message:   fmt.Sprintf("Uptime %.2f%% below SLA minimum %.2f%%", metrics.Uptime_Percent, c.config.Min_Uptime_Percent),
				Timestamp: time.Now(),
				Impact:    "availability",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Проверяем downtime
	if c.config.Max_Downtime_Seconds > 0 {
		result.Summary.TotalChecks++
		if metrics.Downtime_Seconds > c.config.Max_Downtime_Seconds {
			violation := SLAViolation{
				Type:      ViolationDowntime,
				Expected:  c.config.Max_Downtime_Seconds,
				Actual:    metrics.Downtime_Seconds,
				Severity:  "critical",
				Message:   fmt.Sprintf("Downtime %d seconds exceeds SLA limit %d seconds", metrics.Downtime_Seconds, c.config.Max_Downtime_Seconds),
				Timestamp: time.Now(),
				Impact:    "availability",
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
			result.Summary.CriticalChecks++
		} else {
			result.Summary.PassedChecks++
		}
	}

	// Определяем exit code на основе нарушений
	result.Summary.FailedChecks = result.Summary.CriticalChecks + result.Summary.WarningChecks
	
	if result.Summary.CriticalChecks > 0 {
		result.ExitCode = 2 // Критические нарушения
	} else if result.Summary.WarningChecks > 0 {
		result.ExitCode = 1 // Предупреждения
	} else {
		result.ExitCode = 0 // Все проверки пройдены
	}

	return result
}

// GetDefaultSLAConfig возвращает конфигурацию SLA по умолчанию
func GetDefaultSLAConfig() AdvancedSLAConfig {
	return AdvancedSLAConfig{
		RTT_P95_Max:           100 * time.Millisecond,
		RTT_P99_Max:           200 * time.Millisecond,
		RTT_P999_Max:          500 * time.Millisecond,
		Jitter_P95_Max:        50 * time.Millisecond,
		Jitter_P99_Max:        100 * time.Millisecond,
		Loss_Max:              1.0, // 1%
		Loss_Burst_Max:        5.0, // 5%
		Min_Throughput_Mbps:   10.0,
		Min_Goodput_Mbps:      8.0,
		Max_Errors_Total:      100,
		Max_Errors_Percent:    0.1, // 0.1%
		Max_Retransmits:       1000,
		Max_Handshake_Time:    5 * time.Second,
		Min_ZeroRTT_Success:   80.0, // 80%
		Max_Stream_Resets:     50,
		Max_Connection_Drops:  10,
		Min_Uptime_Percent:    99.9, // 99.9%
		Max_Downtime_Seconds:  60,   // 1 минута
	}
}
