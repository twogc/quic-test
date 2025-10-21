package client

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"quic-test/internal"
	"quic-test/internal/metrics"

	// "quic-test/internal/report" // удалить

	"crypto/tls"
	"errors"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/guptarohit/asciigraph"
	"github.com/olekukonko/tablewriter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quic-go/quic-go"
)

type TimePoint struct {
	Time  float64 // seconds since start
	Value float64
}

// TUIMetric представляет метрику для TUI дашборда
type TUIMetric struct {
	LatencyMs float64 `json:"latency_ms"`
	Code      int     `json:"code"`
	CPU       float64 `json:"cpu"`
	RTTMs     float64 `json:"rtt_ms"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// Metrics хранит метрики теста
type Metrics struct {
	mu         sync.Mutex
	Success    int
	Errors     int
	BytesSent  int
	Latencies  []float64
	Timestamps []time.Time
	Throughput []float64
	// Time series for latency and throughput
	TimeSeriesLatency    []TimePoint
	TimeSeriesThroughput []TimePoint

	// --- Advanced QUIC/TLS metrics ---
	PacketLoss             float64 // %
	Retransmits            int
	HandshakeTimes         []float64 // ms
	TLSVersion             string
	CipherSuite            string
	SessionResumptionCount int
	ZeroRTTCount           int
	OneRTTCount            int
	OutOfOrderCount        int
	FlowControlEvents      int
	KeyUpdateEvents        int
	ErrorTypeCounts        map[string]int // error type -> count
	// Time series for new metrics
	TimeSeriesPacketLoss    []TimePoint
	TimeSeriesRetransmits   []TimePoint
	TimeSeriesHandshakeTime []TimePoint
	
	// HDR Histograms for precise metrics
	HDRMetrics *metrics.HDRMetrics
}

// ToMap конвертирует метрики в map для совместимости с SLA проверками
func (m *Metrics) ToMap() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Вычисляем средние значения
	var avgLatency float64
	if len(m.Latencies) > 0 {
		sum := 0.0
		for _, l := range m.Latencies {
			sum += l
		}
		avgLatency = sum / float64(len(m.Latencies))
	}
	
	var avgThroughput float64
	if len(m.Throughput) > 0 {
		sum := 0.0
		for _, t := range m.Throughput {
			sum += t
		}
		avgThroughput = sum / float64(len(m.Throughput))
	}
	
	result := map[string]interface{}{
		"Success":    m.Success,
		"Errors":     m.Errors,
		"BytesSent":  m.BytesSent,
		"Latencies":  m.Latencies,
		"ThroughputAverage": avgThroughput,
		"PacketLoss": m.PacketLoss,
		"Retransmits": m.Retransmits,
		"TLSVersion": m.TLSVersion,
		"CipherSuite": m.CipherSuite,
		"SessionResumptionCount": m.SessionResumptionCount,
		"ZeroRTTCount": m.ZeroRTTCount,
		"OneRTTCount": m.OneRTTCount,
		"HandshakeTime": avgLatency,
		"KeyUpdateEvents": m.KeyUpdateEvents,
		"FlowControlEvents": m.FlowControlEvents,
		"ErrorTypeCounts": m.ErrorTypeCounts,
		"TimeSeriesLatency": m.TimeSeriesLatency,
		"TimeSeriesThroughput": m.TimeSeriesThroughput,
		"TimeSeriesPacketLoss": m.TimeSeriesPacketLoss,
		"TimeSeriesRetransmits": m.TimeSeriesRetransmits,
		"TimeSeriesHandshakeTime": m.TimeSeriesHandshakeTime,
	}
	
	// Добавляем HDR-метрики если доступны
	if m.HDRMetrics != nil {
		result["HDRLatencyStats"] = m.HDRMetrics.GetLatencyStats()
		result["HDRJitterStats"] = m.HDRMetrics.GetJitterStats()
		result["HDRHandshakeStats"] = m.HDRMetrics.GetHandshakeStats()
		result["HDRThroughputStats"] = m.HDRMetrics.GetThroughputStats()
		result["HDRNetworkStats"] = m.HDRMetrics.GetNetworkStats()
	}
	
	return result
}

// Run запускает клиентский тест
func Run(cfg internal.TestConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nПолучен сигнал завершения, формируем отчёт...")
		cancel()
	}()

	testMetrics := &Metrics{
		HDRMetrics: metrics.NewHDRMetrics(),
	}
	var wg sync.WaitGroup

	if cfg.Prometheus {
		go startPrometheusExporter(testMetrics)
	}
	startTime := time.Now()
	// Time series collector
	go func() {
		var lastCount int
		var lastBytes int
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				testMetrics.mu.Lock()
				now := time.Since(startTime).Seconds()
				lat := 0.0
				if len(testMetrics.Latencies) > lastCount {
					sum := 0.0
					for _, l := range testMetrics.Latencies[lastCount:] {
						sum += l
					}
					lat = sum / float64(len(testMetrics.Latencies[lastCount:]))
				}
				testMetrics.TimeSeriesLatency = append(testMetrics.TimeSeriesLatency, TimePoint{Time: now, Value: lat})
				bytesNow := testMetrics.BytesSent
				throughput := float64(bytesNow-lastBytes) / 1024.0
				testMetrics.TimeSeriesThroughput = append(testMetrics.TimeSeriesThroughput, TimePoint{Time: now, Value: throughput})
				lastCount = len(testMetrics.Latencies)
				lastBytes = bytesNow
				testMetrics.mu.Unlock()
			}
		}
	}()

	// --- Ramp-up/ramp-down сценарий ---
	var rate int64 = int64(cfg.Rate)
	cfgPtr := &cfg // чтобы менять Rate по указателю
	go func() {
		minRate := int64(1)
		maxRate := int64(cfg.Rate)
		if maxRate < 10 {
			maxRate = 100 // по умолчанию ramp-up до 100 pps
		}
		step := (maxRate - minRate) / 10
		if step < 1 {
			step = 1
		}
		for {
			// Ramp-up
			for r := minRate; r <= maxRate; r += step {
				atomic.StoreInt64(&rate, r)
				time.Sleep(1 * time.Second)
			}
			// Ramp-down
			for r := maxRate; r >= minRate; r -= step {
				atomic.StoreInt64(&rate, r)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	for c := 0; c < cfg.Connections; c++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()
			clientConnection(ctx, *cfgPtr, testMetrics, connID, &rate)
		}(c)
	}

	// Визуализация метрик (заглушка)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				printMetrics(testMetrics, &rate, false)
			}
		}
	}()

	if cfg.Duration > 0 {
		timer := time.NewTimer(cfg.Duration)
		go func() {
			<-timer.C
			fmt.Println("\nТест завершён по таймеру, формируем отчёт...")
			cancel()
		}()
	}

	wg.Wait()

	// Финальный красивый вывод
	printMetrics(testMetrics, &rate, true)

	// Отправляем метрики в QUIC Bottom
	metricsMap := testMetrics.ToMap()
	internal.UpdateBottomMetrics(metricsMap)

	err := internal.SaveReport(cfg, testMetrics)
	if err != nil {
		fmt.Println("Ошибка сохранения отчёта:", err)
	}
	
	// Проверяем SLA если настроено
	if cfg.SlaRttP95 > 0 || cfg.SlaLoss > 0 || cfg.SlaThroughput > 0 || cfg.SlaErrors > 0 {
		internal.ExitWithSLA(cfg, metricsMap)
	}
}

func clientConnection(ctx context.Context, cfg internal.TestConfig, metrics *Metrics, connID int, ratePtr *int64) {
	var tlsConf *tls.Config
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			metrics.mu.Lock()
			metrics.Errors++
			if metrics.ErrorTypeCounts == nil {
				metrics.ErrorTypeCounts = map[string]int{}
			}
			metrics.ErrorTypeCounts["tls_load_cert"]++
			metrics.mu.Unlock()
			fmt.Println("Ошибка загрузки сертификата:", err)
			return
		}
		tlsConf = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-test"},
		}
	} else {
		// Используем единую функцию для генерации TLS конфигурации
		tlsConf = internal.GenerateTLSConfig(cfg.NoTLS)
	}

	handshakeStart := time.Now()
	session, err := quic.DialAddr(ctx, cfg.Addr, tlsConf, nil)
	handshakeTime := time.Since(handshakeStart).Seconds() * 1000 // ms
	metrics.mu.Lock()
	metrics.HandshakeTimes = append(metrics.HandshakeTimes, handshakeTime)
	metrics.TimeSeriesHandshakeTime = append(metrics.TimeSeriesHandshakeTime, TimePoint{Time: time.Since(handshakeStart).Seconds(), Value: handshakeTime})
	// Записываем handshake время в HDR-гистограммы
	if metrics.HDRMetrics != nil {
		metrics.HDRMetrics.RecordHandshakeTime(time.Duration(handshakeTime) * time.Millisecond)
	}
	if err != nil {
		metrics.Errors++
		if metrics.ErrorTypeCounts == nil {
			metrics.ErrorTypeCounts = map[string]int{}
		}
		metrics.ErrorTypeCounts["quic_handshake"]++
		metrics.mu.Unlock()
		fmt.Println("Ошибка соединения:", err)
		return
	}
	// TLS negotiated params
	state := session.ConnectionState()
	metrics.TLSVersion = tlsVersionString(state.TLS.Version)
	metrics.CipherSuite = cipherSuiteString(state.TLS.CipherSuite)
	if state.TLS.DidResume {
		metrics.SessionResumptionCount++
	}
	if state.Used0RTT {
		metrics.ZeroRTTCount++
	} else {
		metrics.OneRTTCount++
	}
	metrics.mu.Unlock()
	defer func() {
		if err := session.CloseWithError(0, "client done"); err != nil {
			fmt.Printf("Warning: failed to close session: %v\n", err)
		}
	}()

	var wg sync.WaitGroup
	for s := 0; s < cfg.Streams; s++ {
		wg.Add(1)
		go func(streamID int) {
			defer wg.Done()
			clientStream(ctx, session, cfg, metrics, connID, streamID, ratePtr)
		}(s)
	}
	wg.Wait()
}

// clientStream реализует передачу данных по QUIC-стриму и сбор метрик
func clientStream(ctx context.Context, session quic.Connection, cfg internal.TestConfig, metrics *Metrics, connID, streamID int, ratePtr *int64) {
	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		metrics.mu.Lock()
		metrics.Errors++
		if metrics.ErrorTypeCounts == nil {
			metrics.ErrorTypeCounts = map[string]int{}
		}
		metrics.ErrorTypeCounts["open_stream"]++
		metrics.mu.Unlock()
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			fmt.Printf("Warning: failed to close stream: %v\n", err)
		}
	}()

	// Инициализация map для ошибок
	metrics.mu.Lock()
	if metrics.ErrorTypeCounts == nil {
		metrics.ErrorTypeCounts = map[string]int{}
	}
	metrics.mu.Unlock()

	packetSize := cfg.PacketSize
	pattern := cfg.Pattern
	sentPackets := 0
	ackedPackets := 0
	retransmits := 0
	outOfOrder := 0
	var lastSeq int64 = -1
	var seq int64
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// Эмуляция задержки
		if cfg.EmulateLatency > 0 {
			time.Sleep(cfg.EmulateLatency)
		}
		// Эмуляция потери пакета
		if cfg.EmulateLoss > 0 && secureFloat64() < cfg.EmulateLoss {
			metrics.mu.Lock()
			metrics.ErrorTypeCounts["emulated_loss"]++
			metrics.mu.Unlock()
			continue // пропускаем отправку
		}
		// Формируем пакет с seq
		buf := makePacket(packetSize, pattern)
		seq++
		if len(buf) >= 8 {
			for i := 0; i < 8; i++ {
				buf[i] = byte(seq >> (8 * i))
			}
		}
		// Дублирование пакета
		dupCount := 1
		if cfg.EmulateDup > 0 && secureFloat64() < cfg.EmulateDup {
			dupCount = 2
			metrics.mu.Lock()
			metrics.ErrorTypeCounts["emulated_dup"]++
			metrics.mu.Unlock()
		}
		for d := 0; d < dupCount; d++ {
			startWrite := time.Now()
			n, err := stream.Write(buf)
			latency := time.Since(startWrite).Seconds() * 1000
			metrics.mu.Lock()
			metrics.BytesSent += n
			metrics.Success++
			metrics.Latencies = append(metrics.Latencies, latency)
			metrics.Timestamps = append(metrics.Timestamps, time.Now())
			// Записываем в HDR-гистограммы
			if metrics.HDRMetrics != nil {
				metrics.HDRMetrics.RecordLatency(time.Duration(latency) * time.Millisecond)
				metrics.HDRMetrics.AddBytesSent(int64(n))
				metrics.HDRMetrics.IncrementPacketsSent()
			}
			metrics.mu.Unlock()
			sentPackets++
			ackedPackets++
			if err != nil {
				metrics.mu.Lock()
				metrics.Errors++
				if metrics.ErrorTypeCounts == nil {
					metrics.ErrorTypeCounts = map[string]int{}
				}
				metrics.ErrorTypeCounts["stream_write"]++
				retransmits++
				metrics.Retransmits++
				var se *quic.StreamError
				var te *quic.TransportError
				if errors.As(err, &se) {
					if uint64(se.ErrorCode) == flowControlErrorCode {
						metrics.FlowControlEvents++
						metrics.ErrorTypeCounts["flow_control"]++
					}
				}
				if errors.As(err, &te) {
					if uint64(te.ErrorCode) == keyUpdateErrorCode {
						metrics.KeyUpdateEvents++
						metrics.ErrorTypeCounts["key_update"]++
					}
				}
				metrics.mu.Unlock()
				continue
			}
			if lastSeq != -1 && seq != lastSeq+1 {
				outOfOrder++
				metrics.mu.Lock()
				metrics.OutOfOrderCount++
				metrics.mu.Unlock()
			}
			lastSeq = seq
			metrics.mu.Lock()
			metrics.TimeSeriesRetransmits = append(metrics.TimeSeriesRetransmits, TimePoint{Time: time.Since(start).Seconds(), Value: float64(retransmits)})
			metrics.TimeSeriesPacketLoss = append(metrics.TimeSeriesPacketLoss, TimePoint{Time: time.Since(start).Seconds(), Value: 100 * float64(sentPackets-ackedPackets) / (float64(sentPackets) + 1e-9)})
			metrics.mu.Unlock()
		}
		// Пауза между пакетами
		rate := atomic.LoadInt64(ratePtr)
		if rate > 0 {
			time.Sleep(time.Second / time.Duration(rate))
		}
	}
}

func makePacket(size int, pattern string) []byte {
	buf := make([]byte, size)
	switch pattern {
	case "zeroes":
		// already zeroed
	case "increment":
		for i := range buf {
			buf[i] = byte(i % 256)
		}
	default:
		_, _ = rand.Read(buf)
	}
	return buf
}

// calcPercentiles вычисляет p50, p95, p99 для латенси
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

// calcJitter вычисляет стандартное отклонение латенси
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

func printMetrics(metrics *Metrics, ratePtr *int64, final bool) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	if !final {
		fmt.Print("\033[H\033[2J") // очистка экрана и курсор в левый верхний угол
	}
	fmt.Println("\033[1;36m  2GC CloudBridge QUIC testing Client\033[0m")

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"Success", "Errors", "BytesSent", "Avg Latency (ms)", "Throughput", "Uptime (s)", "Rate (pps)"}
	table.Append(headers)

	avgLatency := 0.0
	if len(metrics.Latencies) > 0 {
		sum := 0.0
		for _, l := range metrics.Latencies {
			sum += l
		}
		avgLatency = sum / float64(len(metrics.Latencies))
	}

	// Percentiles & jitter
	p50, p95, p99 := calcPercentiles(metrics.Latencies)
	jitter := calcJitter(metrics.Latencies)

	uptime := 0.0
	if len(metrics.Timestamps) > 0 {
		uptime = time.Since(metrics.Timestamps[0]).Seconds()
	}

	throughput := 0.0
	if uptime > 0 {
		throughput = float64(metrics.BytesSent) / 1024.0 / uptime
	}

	rate := int64(0)
	if ratePtr != nil {
		rate = atomic.LoadInt64(ratePtr)
	}

	row := []string{
		green(fmt.Sprintf("%d", metrics.Success)),
		red(fmt.Sprintf("%d", metrics.Errors)),
		blue(fmt.Sprintf("%.2f KB", float64(metrics.BytesSent)/1024)),
		yellow(fmt.Sprintf("%.2f", avgLatency)),
		blue(fmt.Sprintf("%.2f KB/s", throughput)),
		fmt.Sprintf("%.0f", uptime),
		fmt.Sprintf("%d", rate),
	}
	table.Append(row)
	table.Render()

	fmt.Printf("Percentiles: p50=%.2f ms, p95=%.2f ms, p99=%.2f ms\n", p50, p95, p99)
	fmt.Printf("Jitter: %.2f ms\n", jitter)

	if len(metrics.Latencies) > 0 {
		fmt.Println(yellow(asciigraph.Plot(metrics.Latencies, asciigraph.Height(8), asciigraph.Width(60), asciigraph.Caption("Latency ms"))))
	}
}

func startPrometheusExporter(metrics *Metrics) {
	success := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_success_total",
		Help: "Total successful packets sent",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return float64(metrics.Success)
	})
	errors := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_errors_total",
		Help: "Total errors",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return float64(metrics.Errors)
	})
	bytesSent := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_bytes_sent",
		Help: "Total bytes sent",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return float64(metrics.BytesSent)
	})
	avgLatency := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_avg_latency_ms",
		Help: "Average latency in ms",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		if len(metrics.Latencies) == 0 {
			return 0
		}
		sum := 0.0
		for _, l := range metrics.Latencies {
			sum += l
		}
		return sum / float64(len(metrics.Latencies))
	})
	throughput := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_throughput_kbps",
		Help: "Current throughput in KB/s",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		uptime := 0.0
		if len(metrics.Timestamps) > 0 {
			uptime = time.Since(metrics.Timestamps[0]).Seconds()
		}
		if uptime > 0 {
			return float64(metrics.BytesSent) / 1024.0 / uptime
		}
		return 0
	})

	prometheus.MustRegister(success, errors, bytesSent, avgLatency, throughput)
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Prometheus endpoint доступен на :2112/metrics")
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Printf("Failed to start Prometheus server: %v", err)
	}
}

// Вспомогательные функции для TLSVersion/CipherSuite
func tlsVersionString(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	default:
		return fmt.Sprintf("0x%x", v)
	}
}
func cipherSuiteString(cs uint16) string {
	switch cs {
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	default:
		return fmt.Sprintf("0x%x", cs)
	}
}

// secureFloat64 генерирует криптографически стойкое случайное число от 0 до 1
func secureFloat64() float64 {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based seed if crypto/rand fails
		return float64(time.Now().UnixNano()%1000) / 1000.0
	}
	return float64(binary.BigEndian.Uint64(b)) / float64(^uint64(0))
}

// Коды ошибок из RFC 9000/QUIC:
const (
	flowControlErrorCode = 0x3 // FlowControlError
	keyUpdateErrorCode   = 0xE // KeyUpdateError
)
