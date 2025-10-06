package internal

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// SaveReport сохраняет отчёт по завершении теста в выбранном формате
func SaveReport(cfg TestConfig, metrics any) error {
	format := strings.ToLower(cfg.ReportFormat)
	if format == "" {
		format = "md"
	}
	filename := cfg.ReportPath
	if filename == "" {
		filename = fmt.Sprintf("report.%s", format)
	}

	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(makeReportJSON(cfg, metrics), "", "  ")
	case "csv":
		return saveCSV(filename, makeReportCSV(cfg, metrics))
	case "md":
		data = []byte(makeReportMarkdown(cfg, metrics))
	default:
		data = []byte(makeReportMarkdown(cfg, metrics))
	}

	if format != "csv" {
		err = os.WriteFile(filename, data, 0644)
	}
	if err != nil {
		return fmt.Errorf("ошибка сохранения отчёта: %w", err)
	}
	fmt.Println("\nОтчёт сохранён:", filename)
	return nil
}

// --- Заглушки для сериализации ---

func makeReportJSON(cfg TestConfig, metrics any) any {
	return map[string]any{
		"params": cfg,
		"metrics": metrics,
	}
}

func makeReportCSV(cfg TestConfig, metrics any) [][]string {
	// TODO: реализовать сериализацию в CSV
	return [][]string{{"param", "value"}, {"mode", cfg.Mode}}
}

func saveCSV(filename string, rows [][]string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: failed to close file %s: %v\n", filename, err)
		}
	}()
	w := csv.NewWriter(f)
	defer w.Flush()
	return w.WriteAll(rows)
}

func makeReportMarkdown(cfg TestConfig, metrics any) string {
	m, ok := metrics.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("# 2GC CloudBridge QUICK testing\n\n**Параметры:** \"%+v\"\n\n**Метрики:** \"%+v\"\n", cfg, metrics)
	}
	latencies, _ := m["Latencies"].([]float64)
	p50, p95, p99 := calcPercentiles(latencies)
	jitter := calcJitter(latencies)
	avg := avgLatency(latencies)

	tsLatency, _ := m["TimeSeriesLatency"].([]interface{})
	tsThroughput, _ := m["TimeSeriesThroughput"].([]interface{})
	tsPacketLoss, _ := m["TimeSeriesPacketLoss"].([]interface{})
	tsRetransmits, _ := m["TimeSeriesRetransmits"].([]interface{})
	tsHandshakeTime, _ := m["TimeSeriesHandshakeTime"].([]interface{})

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`# 2GC CloudBridge QUICK testing\n\n**Параметры:** "%+v"\n\n**Метрики:**\n\n- Success: %v\n- Errors: %v\n- BytesSent: %v\n- Avg Latency: %.2f ms\n- p50: %.2f ms\n- p95: %.2f ms\n- p99: %.2f ms\n- Jitter: %.2f ms\n- PacketLoss: %v %%\n- Retransmits: %v\n- TLSVersion: %v\n- CipherSuite: %v\n- SessionResumptionCount: %v\n- 0-RTT: %v\n- 1-RTT: %v\n- OutOfOrder: %v\n- FlowControlEvents: %v\n- KeyUpdateEvents: %v\n- ErrorTypeCounts: %v\n`, cfg, m["Success"], m["Errors"], m["BytesSent"], avg, p50, p95, p99, jitter, m["PacketLoss"], m["Retransmits"], m["TLSVersion"], m["CipherSuite"], m["SessionResumptionCount"], m["ZeroRTTCount"], m["OneRTTCount"], m["OutOfOrderCount"], m["FlowControlEvents"], m["KeyUpdateEvents"], m["ErrorTypeCounts"]))

	buf.WriteString("\n## Временные ряды (Time Series)\n")
	buf.WriteString("\n### Latency (ms)\n")
	buf.WriteString("| Time (s) | Latency (ms) |\n|---|---|\n")
	for _, v := range tsLatency {
		point, ok := v.(map[string]interface{})
		if ok {
			buf.WriteString(fmt.Sprintf("| %.0f | %.2f |\n", point["Time"].(float64), point["Value"].(float64)))
		}
	}
	buf.WriteString("\n### Throughput (KB/s)\n| Time (s) | Throughput (KB/s) |\n|---|---|\n")
	for _, v := range tsThroughput {
		point, ok := v.(map[string]interface{})
		if ok {
			buf.WriteString(fmt.Sprintf("| %.0f | %.2f |\n", point["Time"].(float64), point["Value"].(float64)))
		}
	}
	buf.WriteString("\n### Packet Loss (%)\n| Time (s) | Packet Loss (%) |\n|---|---|\n")
	for _, v := range tsPacketLoss {
		point, ok := v.(map[string]interface{})
		if ok {
			buf.WriteString(fmt.Sprintf("| %.0f | %.2f |\n", point["Time"].(float64), point["Value"].(float64)))
		}
	}
	buf.WriteString("\n### Retransmits\n| Time (s) | Retransmits |\n|---|---|\n")
	for _, v := range tsRetransmits {
		point, ok := v.(map[string]interface{})
		if ok {
			buf.WriteString(fmt.Sprintf("| %.0f | %.0f |\n", point["Time"].(float64), point["Value"].(float64)))
		}
	}
	buf.WriteString("\n### Handshake Time (ms)\n| Time (s) | Handshake Time (ms) |\n|---|---|\n")
	for _, v := range tsHandshakeTime {
		point, ok := v.(map[string]interface{})
		if ok {
			buf.WriteString(fmt.Sprintf("| %.0f | %.2f |\n", point["Time"].(float64), point["Value"].(float64)))
		}
	}
	// ASCII-графики
	buf.WriteString("\n#### Latency Graph (ASCII)\n\n```")
	var latencyVals []float64
	for _, v := range tsLatency {
		point, ok := v.(map[string]interface{})
		if ok {
			latencyVals = append(latencyVals, point["Value"].(float64))
		}
	}
	buf.WriteString("\n" + asciigraphPlot(latencyVals, "Latency ms") + "\n")
	buf.WriteString("```")
	buf.WriteString("\n#### Throughput Graph (ASCII)\n\n```")
	var throughputVals []float64
	for _, v := range tsThroughput {
		point, ok := v.(map[string]interface{})
		if ok {
			throughputVals = append(throughputVals, point["Value"].(float64))
		}
	}
	buf.WriteString("\n" + asciigraphPlot(throughputVals, "Throughput KB/s") + "\n")
	buf.WriteString("```")
	buf.WriteString("\n#### Packet Loss Graph (ASCII)\n\n```")
	var lossVals []float64
	for _, v := range tsPacketLoss {
		point, ok := v.(map[string]interface{})
		if ok {
			lossVals = append(lossVals, point["Value"].(float64))
		}
	}
	buf.WriteString("\n" + asciigraphPlot(lossVals, "Packet Loss %") + "\n")
	buf.WriteString("```")
	buf.WriteString("\n#### Retransmits Graph (ASCII)\n\n```")
	var retransVals []float64
	for _, v := range tsRetransmits {
		point, ok := v.(map[string]interface{})
		if ok {
			retransVals = append(retransVals, point["Value"].(float64))
		}
	}
	buf.WriteString("\n" + asciigraphPlot(retransVals, "Retransmits") + "\n")
	buf.WriteString("```")
	buf.WriteString("\n#### Handshake Time Graph (ASCII)\n\n```")
	var hsVals []float64
	for _, v := range tsHandshakeTime {
		point, ok := v.(map[string]interface{})
		if ok {
			hsVals = append(hsVals, point["Value"].(float64))
		}
	}
	buf.WriteString("\n" + asciigraphPlot(hsVals, "Handshake Time ms") + "\n")
	buf.WriteString("```")
	return buf.String()
}

// ascii-график (заглушка, если нет asciigraph)
func asciigraphPlot(data []float64, caption string) string {
	if len(data) == 0 {
		return ""
	}
	// Можно подключить asciigraph, если доступен, иначе простая заглушка
	max := data[0]
	min := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	return fmt.Sprintf("%s: min=%.2f max=%.2f (graph suppressed)\n", caption, min, max)
}

// calcPercentiles и calcJitter (дублируем для отчёта)
func calcPercentiles(latencies []float64) (p50, p95, p99 float64) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}
	copyLat := make([]float64, len(latencies))
	copy(copyLat, latencies)
	sort.Float64s(copyLat)
	idx := func(p float64) int {
		return int(p*float64(len(copyLat)-1) + 0.5)
	}
	p50 = copyLat[idx(0.50)]
	p95 = copyLat[idx(0.95)]
	p99 = copyLat[idx(0.99)]
	return
}
func calcJitter(latencies []float64) float64 {
	if len(latencies) == 0 {
		return 0
	}
	mean := 0.0
	for _, l := range latencies {
		mean += l
	}
	mean /= float64(len(latencies))
	var sum float64
	for _, l := range latencies {
		d := l - mean
		sum += d * d
	}
	return (sum / float64(len(latencies)))
}
func avgLatency(latencies []float64) float64 {
	if len(latencies) == 0 {
		return 0
	}
	sum := 0.0
	for _, l := range latencies {
		sum += l
	}
	return sum / float64(len(latencies))
} 