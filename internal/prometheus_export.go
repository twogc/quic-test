package internal

import (
	"fmt"
	"os"
	"time"
)

// ExportPrometheusMetrics экспортирует метрики в Prometheus text exposition format
func ExportPrometheusMetrics(cfg TestConfig, metrics map[string]interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create prometheus file: %w", err)
	}
	defer file.Close()

	// Заголовок с HELP и TYPE
	file.WriteString("# HELP quic_test_duration_seconds Test duration in seconds\n")
	file.WriteString("# TYPE quic_test_duration_seconds gauge\n")
	
	file.WriteString("# HELP quic_test_connections_total Number of connections\n")
	file.WriteString("# TYPE quic_test_connections_total gauge\n")
	
	file.WriteString("# HELP quic_test_bytes_sent_total Total bytes sent\n")
	file.WriteString("# TYPE quic_test_bytes_sent_total counter\n")
	
	file.WriteString("# HELP quic_test_packets_sent_total Total packets sent\n")
	file.WriteString("# TYPE quic_test_packets_sent_total counter\n")
	
	file.WriteString("# HELP quic_test_errors_total Total errors\n")
	file.WriteString("# TYPE quic_test_errors_total counter\n")
	
	file.WriteString("# HELP quic_test_latency_p50_ms Latency p50 in milliseconds\n")
	file.WriteString("# TYPE quic_test_latency_p50_ms gauge\n")
	
	file.WriteString("# HELP quic_test_latency_p95_ms Latency p95 in milliseconds\n")
	file.WriteString("# TYPE quic_test_latency_p95_ms gauge\n")
	
	file.WriteString("# HELP quic_test_latency_p99_ms Latency p99 in milliseconds\n")
	file.WriteString("# TYPE quic_test_latency_p99_ms gauge\n")
	
	file.WriteString("# HELP quic_test_jitter_ms Jitter in milliseconds\n")
	file.WriteString("# TYPE quic_test_jitter_ms gauge\n")
	
	file.WriteString("# HELP quic_test_throughput_mbps Throughput in Mbps\n")
	file.WriteString("# TYPE quic_test_throughput_mbps gauge\n")
	
	file.WriteString("# HELP quic_test_packet_loss_percent Packet loss percentage\n")
	file.WriteString("# TYPE quic_test_packet_loss_percent gauge\n")
	
	file.WriteString("# HELP quic_test_retransmission_rate_percent Retransmission rate percentage\n")
	file.WriteString("# TYPE quic_test_retransmission_rate_percent gauge\n")

	// Базовые метрики (используем функции из schema.go)
	bytesSent := getInt64(metrics, "BytesSent")
	success := getInt(metrics, "Success")
	errors := getInt(metrics, "Errors")
	
	durationSec := float64(cfg.Duration.Seconds())
	if durationSec == 0 {
		durationSec = 60.0 // default
	}
	
	rttP50 := getFloat64FromSchema(metrics, "RTTP50Ms")
	rttP95 := getFloat64FromSchema(metrics, "RTTP95Ms")
	rttP99 := getFloat64FromSchema(metrics, "RTTP99Ms")
	jitter := getFloat64FromSchema(metrics, "JitterMs")
	throughputMbps := getFloat64FromSchema(metrics, "ThroughputMbps")
	packetLoss := getFloat64FromSchema(metrics, "PacketLoss") * 100
	retransmissionRate := getFloat64FromSchema(metrics, "RetransmissionRate") * 100

	// Записываем метрики
	file.WriteString(fmt.Sprintf("quic_test_duration_seconds{cc=\"%s\"} %.2f\n", cfg.CongestionControl, durationSec))
	file.WriteString(fmt.Sprintf("quic_test_connections_total{cc=\"%s\"} %d\n", cfg.CongestionControl, cfg.Connections))
	file.WriteString(fmt.Sprintf("quic_test_bytes_sent_total{cc=\"%s\"} %d\n", cfg.CongestionControl, bytesSent))
	file.WriteString(fmt.Sprintf("quic_test_packets_sent_total{cc=\"%s\"} %d\n", cfg.CongestionControl, success))
	file.WriteString(fmt.Sprintf("quic_test_errors_total{cc=\"%s\"} %d\n", cfg.CongestionControl, errors))
	file.WriteString(fmt.Sprintf("quic_test_latency_p50_ms{cc=\"%s\"} %.3f\n", cfg.CongestionControl, rttP50))
	file.WriteString(fmt.Sprintf("quic_test_latency_p95_ms{cc=\"%s\"} %.3f\n", cfg.CongestionControl, rttP95))
	file.WriteString(fmt.Sprintf("quic_test_latency_p99_ms{cc=\"%s\"} %.3f\n", cfg.CongestionControl, rttP99))
	file.WriteString(fmt.Sprintf("quic_test_jitter_ms{cc=\"%s\"} %.3f\n", cfg.CongestionControl, jitter))
	file.WriteString(fmt.Sprintf("quic_test_throughput_mbps{cc=\"%s\"} %.3f\n", cfg.CongestionControl, throughputMbps))
	file.WriteString(fmt.Sprintf("quic_test_packet_loss_percent{cc=\"%s\"} %.3f\n", cfg.CongestionControl, packetLoss))
	file.WriteString(fmt.Sprintf("quic_test_retransmission_rate_percent{cc=\"%s\"} %.3f\n", cfg.CongestionControl, retransmissionRate))

	// BBRv3 специфичные метрики
	if bbrv3Metrics, ok := metrics["BBRv3Metrics"].(map[string]interface{}); ok {
		file.WriteString("\n# BBRv3 specific metrics\n")
		file.WriteString("# HELP quic_bbrv3_phase_current Current BBRv3 phase\n")
		file.WriteString("# TYPE quic_bbrv3_phase_current gauge\n")
		file.WriteString("# HELP quic_bbrv3_bandwidth_fast_bps BBRv3 fast bandwidth estimate\n")
		file.WriteString("# TYPE quic_bbrv3_bandwidth_fast_bps gauge\n")
		file.WriteString("# HELP quic_bbrv3_bandwidth_slow_bps BBRv3 slow bandwidth estimate\n")
		file.WriteString("# TYPE quic_bbrv3_bandwidth_slow_bps gauge\n")
		file.WriteString("# HELP quic_bbrv3_loss_rate_round_percent BBRv3 loss rate per round\n")
		file.WriteString("# TYPE quic_bbrv3_loss_rate_round_percent gauge\n")
		file.WriteString("# HELP quic_bbrv3_headroom_usage_percent BBRv3 headroom usage\n")
		file.WriteString("# TYPE quic_bbrv3_headroom_usage_percent gauge\n")
		file.WriteString("# HELP quic_bbrv3_pacing_gain BBRv3 pacing gain\n")
		file.WriteString("# TYPE quic_bbrv3_pacing_gain gauge\n")

		phase := getString(bbrv3Metrics, "phase")
		phaseValue := 0.0
		switch phase {
		case "Startup":
			phaseValue = 1.0
		case "Drain":
			phaseValue = 2.0
		case "ProbeBW":
			phaseValue = 3.0
		case "ProbeRTT":
			phaseValue = 4.0
		}

		bwFast := getFloat64FromSchema(bbrv3Metrics, "bw_fast")
		bwSlow := getFloat64FromSchema(bbrv3Metrics, "bw_slow")
		lossRateRound := getFloat64FromSchema(bbrv3Metrics, "loss_rate_round") * 100
		headroomUsage := getFloat64FromSchema(bbrv3Metrics, "headroom_usage") * 100
		pacingGain := getFloat64FromSchema(bbrv3Metrics, "pacing_gain")

		file.WriteString(fmt.Sprintf("quic_bbrv3_phase_current{cc=\"%s\"} %.0f\n", cfg.CongestionControl, phaseValue))
		file.WriteString(fmt.Sprintf("quic_bbrv3_bandwidth_fast_bps{cc=\"%s\"} %.0f\n", cfg.CongestionControl, bwFast))
		file.WriteString(fmt.Sprintf("quic_bbrv3_bandwidth_slow_bps{cc=\"%s\"} %.0f\n", cfg.CongestionControl, bwSlow))
		file.WriteString(fmt.Sprintf("quic_bbrv3_loss_rate_round_percent{cc=\"%s\"} %.4f\n", cfg.CongestionControl, lossRateRound))
		file.WriteString(fmt.Sprintf("quic_bbrv3_headroom_usage_percent{cc=\"%s\"} %.2f\n", cfg.CongestionControl, headroomUsage))
		file.WriteString(fmt.Sprintf("quic_bbrv3_pacing_gain{cc=\"%s\"} %.2f\n", cfg.CongestionControl, pacingGain))
	}

	file.WriteString(fmt.Sprintf("\n# Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	
	return nil
}

