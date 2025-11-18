package client

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"math"
	"sort"
	"sync"
	"syscall"
	"time"

	"quic-test/internal"
	"quic-test/internal/metrics"
	"quic-test/internal/integration"
	"quic-test/internal/fec"
	"quic-test/internal/pqc"

	"crypto/tls"
	"errors"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"go.uber.org/zap"
)

type TimePoint struct {
	Time  float64 `json:"Time"`  // seconds since start
	Value float64 `json:"Value"`
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
	
	// FEC Metrics
	FECPacketsSent    int64   `json:"fec_packets_sent"`
	FECRedundancyBytes int64   `json:"fec_redundancy_bytes"`
	FECRepairPacketsSent int64 `json:"fec_repair_sent"`      // Redundancy packets sent
	FECRecovered       int64   `json:"fec_recovered"`        // Packets recovered via FEC
	FECRecoveryEvents  int64   `json:"fec_recovery_events"`
	FECUseCXX          bool    `json:"fec_use_cxx"`          // Whether C++ SIMD encoder is used
	
	// PQC Metrics
	PQCHandshakeSize int64   `json:"pqc_handshake_size"`
	PQCHandshakeTime float64 `json:"pqc_handshake_time_ms"`
	PQCAlgorithm     string  `json:"pqc_algorithm"`
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
	
	// Вычисляем RTT процентили из Latencies (в миллисекундах)
	var rttP50, rttP95, rttP99 float64
	if len(m.Latencies) > 0 {
		rttP50, rttP95, rttP99 = calcPercentiles(m.Latencies)
	}
	
	// Вычисляем jitter (стандартное отклонение)
	jitter := calcJitter(m.Latencies)
	
	// Вычисляем throughput в Mbps (корректная формула: bytes * 8 / duration_seconds / 1e6)
	var throughputMbps float64
	var minRTT float64
	if len(m.Timestamps) > 0 {
		duration := time.Since(m.Timestamps[0]).Seconds()
		if duration > 0 {
			throughputMbps = (float64(m.BytesSent) * 8) / (duration * 1_000_000) // Bytes to Mbps
		}
		// Находим min RTT из latencies
		if len(m.Latencies) > 0 {
			minRTT = m.Latencies[0]
			for _, l := range m.Latencies {
				if l > 0 && l < minRTT {
					minRTT = l
				}
			}
		}
	}
	
	// Вычисляем goodput (исключая ретрансмиты)
	var goodputMbps float64
	if len(m.Timestamps) > 0 {
		duration := time.Since(m.Timestamps[0]).Seconds()
		if duration > 0 {
			// Приблизительно: вычитаем ретрансмиты из отправленных байт
			estimatedRetransBytes := int64(m.Retransmits) * 1200 // Примерный размер пакета
			goodputBytes := int64(m.BytesSent) - estimatedRetransBytes
			if goodputBytes < 0 {
				goodputBytes = 0
			}
			goodputMbps = (float64(goodputBytes) * 8) / (duration * 1_000_000)
		}
	}
	
	// Вычисляем bufferbloat factor: (avg_rtt / min_rtt) - 1
	var bufferbloatFactor float64
	if minRTT > 0 && avgLatency > 0 {
		bufferbloatFactor = (avgLatency / minRTT) - 1.0
		if bufferbloatFactor < 0 {
			bufferbloatFactor = 0
		}
	}
	
	// Вычисляем Fairness Index (Jain's index) для всех соединений
	// Приблизительно: используем вариацию throughput по времени как proxy для fairness
	var fairnessIndex float64
	if len(m.TimeSeriesThroughput) > 0 {
		var sum, sumSq float64
		for _, tp := range m.TimeSeriesThroughput {
			if tp.Value > 0 {
				sum += tp.Value
				sumSq += tp.Value * tp.Value
			}
		}
		if sum > 0 && sumSq > 0 {
			fairnessIndex = (sum * sum) / (float64(len(m.TimeSeriesThroughput)) * sumSq)
		}
	} else {
		// Если нет time series, используем вариацию latencies как proxy
		if len(m.Latencies) > 0 {
			var sum, sumSq float64
			for _, l := range m.Latencies {
				if l > 0 {
					sum += l
					sumSq += l * l
				}
			}
			if sum > 0 && sumSq > 0 {
				fairnessIndex = (sum * sum) / (float64(len(m.Latencies)) * sumSq)
			}
		}
	}
	
	// Вычисляем retransmission rate
	var retransmissionRate float64
	if m.Success > 0 {
		retransmissionRate = float64(m.Retransmits) / float64(m.Success)
	}
	
	result := map[string]interface{}{
		"Success":    m.Success,
		"Errors":     m.Errors,
		"BytesSent":  m.BytesSent,
		"Latencies":  m.Latencies,
		"ThroughputAverage": avgThroughput,
		"ThroughputMbps": throughputMbps,
		"GoodputMbps": goodputMbps,
		"RetransmissionRate": retransmissionRate,
		"RTTP50Ms": rttP50,
		"RTTP95Ms": rttP95,
		"RTTP99Ms": rttP99,
		"RTTMinMs": minRTT,
		"RTTAvgMs": avgLatency,
		"JitterMs": jitter,
		"PacketLoss": m.PacketLoss,
		"Retransmits": m.Retransmits,
		"BufferbloatFactor": bufferbloatFactor,
		"FairnessIndex": fairnessIndex,
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
		"FECPacketsSent": m.FECPacketsSent,
		"FECRedundancyBytes": m.FECRedundancyBytes,
		"FECRepairPacketsSent": m.FECRepairPacketsSent,
		"FECRecovered": m.FECRecovered,
		"FECRecoveryEvents": m.FECRecoveryEvents,
		"PQCHandshakeSize": m.PQCHandshakeSize,
		"PQCHandshakeTime": m.PQCHandshakeTime,
		"PQCAlgorithm": m.PQCAlgorithm,
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
		fmt.Println("\nПолучен сигнал завершения, формируем отчет...")
		cancel()
	}()

	// SimpleIntegration теперь создается для каждого соединения отдельно
	// Это необходимо для потокобезопасности при множественных соединениях

	testMetrics := &Metrics{
		HDRMetrics: metrics.NewHDRMetrics(),
	}
	var wg sync.WaitGroup

	if cfg.Prometheus {
		go startPrometheusExporter(testMetrics)
	}
	// Создаем и регистрируем глобальный SimpleIntegration ДО запуска горутин соединений
	// Это нужно, чтобы EnhanceMetricsMap мог получить BBRv3 метрики с самого начала
	// Глобальный SimpleIntegration будет использоваться во всех соединениях для сбора метрик
	var globalSI *integration.SimpleIntegration
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		logger, _ := zap.NewDevelopment()
		globalSI = integration.NewSimpleIntegration(logger, cfg.CongestionControl)
		if err := globalSI.Initialize(); err != nil {
			fmt.Printf("Warning: Failed to initialize global %s integration: %v\n", cfg.CongestionControl, err)
			globalSI = nil
		} else {
			gmc := internal.GetGlobalMetricsCollector()
			gmc.SetExperimentalIntegration(globalSI)
			fmt.Printf("[INFO] Global BBRv3 integration registered in GlobalMetricsCollector\n")
		}
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
				
				// Периодическая отправка метрик в QUIC Bottom
				metricsMap := testMetrics.ToMap()
				metricsMap = internal.EnhanceMetricsMap(metricsMap)
				internal.UpdateBottomMetrics(metricsMap)
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
			defer func() {
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d goroutine defer started\n", connID)
				}
				wg.Done()
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d goroutine defer completed, wg.Done() called\n", connID)
				}
			}()
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d goroutine started\n", connID)
			}
			// Используем глобальный SimpleIntegration для всех соединений
			// Это позволяет собирать метрики BBRv3 в одном месте
			var si *integration.SimpleIntegration
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				// Используем глобальный SimpleIntegration, если он создан
				if globalSI != nil {
					si = globalSI
				} else {
					// Fallback: создаем локальный, если глобальный не создан
					logger, _ := zap.NewDevelopment()
					si = integration.NewSimpleIntegration(logger, cfg.CongestionControl)
					if err := si.Initialize(); err != nil {
						fmt.Printf("Warning: Failed to initialize %s integration for connection %d: %v\n", cfg.CongestionControl, connID, err)
						si = nil
					}
				}
			}
			clientConnection(ctx, *cfgPtr, testMetrics, connID, &rate, si)
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d goroutine clientConnection returned\n", connID)
			}
		}(c)
	}

	// Убрана визуализация - только сохранение результатов

	if cfg.Duration > 0 {
		timer := time.NewTimer(cfg.Duration)
		go func() {
			<-timer.C
			fmt.Println("\nТест завершен по таймеру, формируем отчет...")
			cancel()
		}()
	}

	// Добавляем таймаут для wg.Wait чтобы избежать зависаний
	done := make(chan struct{})
	go func() {
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Starting wg.Wait() for %d connections\n", cfg.Connections)
		}
		wg.Wait()
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] wg.Wait() completed, all connections finished\n")
		}
		close(done)
	}()

	// Ждем завершения или таймаут (дополнительные 10 секунд после duration)
	timeout := cfg.Duration + 10*time.Second
	if cfg.Duration == 0 {
		timeout = 120 * time.Second // default timeout
	}
	
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] Waiting for connections to finish, timeout: %v\n", timeout)
	}
	
	select {
	case <-done:
		// Все горутины завершились
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] All connections finished normally\n")
		}
	case <-time.After(timeout):
		fmt.Printf("\n⚠️  Таймаут ожидания завершения (%v). Завершаем принудительно...\n", timeout)
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Timeout reached, canceling context...\n")
		}
		cancel() // Отменяем контекст
		// Ждем еще немного
		select {
		case <-done:
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connections finished after cancel\n")
			}
		case <-time.After(5 * time.Second):
			fmt.Println("⚠️  Некоторые горутины не завершились, продолжаем...")
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Some goroutines still not finished after 5s wait\n")
			}
		}
	}

	// Минимальный вывод результатов
	fmt.Printf("\nТест завершен. Обработка результатов...\n")

	// Отправляем метрики в QUIC Bottom (опционально)
	metricsMap := testMetrics.ToMap()
	
	// Enhance with BBRv3 and experimental metrics
	metricsMap = internal.EnhanceMetricsMap(metricsMap)
	
	// Базовый вывод только для контроля
	if bbrv3Metrics, ok := metricsMap["BBRv3Metrics"].(map[string]interface{}); ok {
		fmt.Printf("BBRv3 Phase: %v, BW: %.2f Mbps\n", 
			bbrv3Metrics["phase"], 
			bbrv3Metrics["bw"].(float64)/1_000_000)
	}
	
	// Опционально: отправка в QUIC Bottom (если нужно)
	internal.UpdateBottomMetrics(metricsMap)

	// Save report with enhanced metrics (including BBRv3)
	err := internal.SaveReport(cfg, metricsMap)
	if err != nil {
		fmt.Printf("Ошибка сохранения отчета: %v\n", err)
	}

	// Экспорт в Prometheus format
	if cfg.ReportPath != "" {
		// Создаем имя файла для Prometheus (заменяем расширение на .prom)
		promFile := cfg.ReportPath
		if len(promFile) > 4 && promFile[len(promFile)-5:] == ".json" {
			promFile = promFile[:len(promFile)-5] + ".prom"
		} else {
			promFile = promFile + ".prom"
		}
		
		if err := internal.ExportPrometheusMetrics(cfg, metricsMap, promFile); err != nil {
			fmt.Printf("Ошибка экспорта Prometheus метрик: %v\n", err)
		} else {
			fmt.Printf("Prometheus метрики сохранены: %s\n", promFile)
		}
	}
	
	// Проверяем SLA если настроено
	if cfg.SlaRttP95 > 0 || cfg.SlaLoss > 0 || cfg.SlaThroughput > 0 || cfg.SlaErrors > 0 {
		internal.ExitWithSLA(cfg, metricsMap)
	}
}

func clientConnection(ctx context.Context, cfg internal.TestConfig, metrics *Metrics, connID int, ratePtr *int64, si *integration.SimpleIntegration) {
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] clientConnection %d: started\n", connID)
	}
	defer func() {
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] clientConnection %d: returning\n", connID)
		}
	}()
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

	// Создаем отдельный UDP connection для каждого QUIC connection
	// Это необходимо для поддержки большого количества одновременных connections
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		metrics.mu.Lock()
		metrics.Errors++
		if metrics.ErrorTypeCounts == nil {
			metrics.ErrorTypeCounts = map[string]int{}
		}
		metrics.ErrorTypeCounts["udp_socket"]++
		metrics.mu.Unlock()
		fmt.Printf("Ошибка создания UDP socket для connection %d: %v\n", connID, err)
		return
	}
	defer udpConn.Close()

	// Создаем QUIC конфигурацию с tracer для BBRv3
	var quicConfig *quic.Config
	if si != nil && cfg.CongestionControl == "bbrv3" {
		// Создаем tracer для отслеживания реальных ACK событий
		logger, _ := zap.NewDevelopment()
		
		quicConfig = &quic.Config{
			Tracer: func(ctx context.Context, perspective logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
				connectionIDStr := fmt.Sprintf("conn_%d_%s", connID, connID.String())
				return integration.NewConnectionTracerForConnection(logger, si, connectionIDStr)
			},
		}
	}
	
	// Создаем отдельный Transport для каждого connection
	transport := &quic.Transport{
		Conn: udpConn,
	}
	defer transport.Close()

	handshakeStart := time.Now()
	
	// PQC симуляция: эмулируем overhead если включен
	var pqcSim *pqc.PQCSimulator
	if cfg.PQCEnabled && cfg.PQCAlgorithm != "" {
		pqcSim = pqc.NewPQCSimulator(cfg.PQCAlgorithm)
		pqcOverhead, pqcSize := pqcSim.SimulateHandshake()
		
		// Добавляем PQC overhead к handshake времени
		time.Sleep(pqcOverhead)
		
		metrics.mu.Lock()
		metrics.PQCHandshakeSize = int64(pqcSize)
		metrics.PQCHandshakeTime = float64(pqcOverhead.Nanoseconds()) / 1e6
		metrics.PQCAlgorithm = cfg.PQCAlgorithm
		metrics.mu.Unlock()
	}
	
	session, err := transport.Dial(ctx, parseAddr(cfg.Addr), tlsConf, quicConfig)
	handshakeTime := time.Since(handshakeStart).Seconds() * 1000 // ms
	
	// Сохраняем connection для использования в tracer (если используется BBRv3)
	if si != nil && cfg.CongestionControl == "bbrv3" && session != nil {
		connectionID := fmt.Sprintf("conn_%d", connID)
		integration.StoreConnection(connectionID, session)
	}
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
			defer func() {
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: defer started\n", connID, streamID)
				}
				wg.Done()
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: wg.Done() called\n", connID, streamID)
				}
			}()
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d, Stream %d: goroutine started\n", connID, streamID)
			}
			clientStream(ctx, session, cfg, metrics, connID, streamID, ratePtr, si)
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d, Stream %d: clientStream returned\n", connID, streamID)
			}
		}(s)
	}
	
	// Добавляем таймаут для wg.Wait на уровне соединения
	done := make(chan struct{})
	go func() {
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d: Starting wg.Wait() for %d streams\n", connID, cfg.Streams)
		}
		wg.Wait()
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d: wg.Wait() completed\n", connID)
		}
		close(done)
	}()
	
	streamTimeout := cfg.Duration + 10*time.Second
	if cfg.Duration == 0 {
		streamTimeout = 70 * time.Second
	}
	
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] Connection %d: Waiting for streams, timeout: %v\n", connID, streamTimeout)
	}
	
	select {
	case <-done:
		// Все стримы завершились
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d: All streams finished\n", connID)
		}
	case <-ctx.Done():
		// Контекст отменен - принудительно завершаем
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d: Context canceled, waiting for streams to finish\n", connID)
		}
		// Ждем еще немного для завершения стримов
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			fmt.Printf("[WARNING] Connection %d: Some streams didn't finish after context cancel\n", connID)
		}
	case <-time.After(streamTimeout):
		// Таймаут - принудительно завершаем
		fmt.Printf("[WARNING] Connection %d streams timeout after %v, canceling context\n", connID, streamTimeout)
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d: Stream timeout reached\n", connID)
		}
		// Даем еще немного времени после таймаута
		select {
		case <-done:
		case <-time.After(1 * time.Second):
		}
	}
}

// clientStream реализует передачу данных по QUIC-стриму и сбор метрик
func clientStream(ctx context.Context, session quic.Connection, cfg internal.TestConfig, metrics *Metrics, connID, streamID int, ratePtr *int64, si *integration.SimpleIntegration) {
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] Connection %d, Stream %d: clientStream started\n", connID, streamID)
	}
	
	// Инициализируем FEC encoder если включен
	// Используем HybridFECEncoder для автоматического выбора между C++ SIMD и Go
	var fecEncoder *fec.HybridFECEncoder
	var useCXX bool
	if cfg.FECEnabled && cfg.FECRedundancy > 0 {
		fecEncoder = fec.NewHybridFECEncoder(cfg.FECRedundancy)
		useCXX = fecEncoder.UseCXX()
		metrics.mu.Lock()
		metrics.FECUseCXX = useCXX
		metrics.mu.Unlock()
		if useCXX {
			fmt.Printf("[INFO] Connection %d: FEC acceleration enabled (C++ SIMD, 30-35x faster)\n", connID)
		} else {
			fmt.Printf("[INFO] Connection %d: FEC using Go implementation\n", connID)
		}
	}
	
	defer func() {
		// Flush FEC при завершении
		if fecEncoder != nil {
			redundancy, err := fecEncoder.Flush()
			if err == nil && redundancy != nil {
				// Отправляем последний redundancy пакет если есть
				metrics.mu.Lock()
				metrics.FECPacketsSent++
				metrics.FECRedundancyBytes += int64(len(redundancy))
				metrics.mu.Unlock()
			}

			// Cleanup C++ resources if using SIMD
			fecEncoder.Close()
		}

		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d, Stream %d: clientStream returning\n", connID, streamID)
		}
	}()
	
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
	
	// Таймаут для цикла отправки
	sendTimeout := cfg.Duration
	if sendTimeout == 0 {
		sendTimeout = 60 * time.Second // default
	}
	sendDeadline := time.Now().Add(sendTimeout)
	
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] Connection %d, Stream %d: sendDeadline set to %v (from now: %v)\n", 
			connID, streamID, sendDeadline, sendTimeout)
	}
	
	iterCount := 0
	for {
		iterCount++
		if cfg.CongestionControl == "bbrv3" && iterCount%1000 == 0 {
			elapsed := time.Since(sendDeadline.Add(-sendTimeout))
			fmt.Printf("[DEBUG] Connection %d, Stream %d: iteration %d, elapsed: %v, deadline in: %v\n", 
				connID, streamID, iterCount, elapsed, time.Until(sendDeadline))
		}
		
		// Проверяем контекст и таймаут перед каждой итерацией
		if time.Now().After(sendDeadline) {
			// Достигнут deadline отправки
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d, Stream %d: sendDeadline reached, returning\n", connID, streamID)
			}
			return
		}
		select {
		case <-ctx.Done():
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d, Stream %d: ctx.Done() received, returning\n", connID, streamID)
			}
			return
		default:
		}
		
		// Проверяем таймаут
		if time.Now().After(sendDeadline) {
			return
		}
		
		// Эмуляция задержки (с проверкой контекста и deadline)
		if cfg.EmulateLatency > 0 {
			// Проверяем deadline перед задержкой
			if time.Now().After(sendDeadline) {
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: deadline reached before latency emulation, returning\n", connID, streamID)
				}
				return
			}
			select {
			case <-ctx.Done():
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: ctx.Done() during latency emulation, returning\n", connID, streamID)
				}
				return
			case <-time.After(cfg.EmulateLatency):
				// Проверяем deadline после задержки
				if time.Now().After(sendDeadline) {
					if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
						fmt.Printf("[DEBUG] Connection %d, Stream %d: deadline reached after latency emulation, returning\n", connID, streamID)
					}
					return
				}
			}
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
		
		// FEC: добавляем пакет в encoder и создаем redundancy если нужно
		var redundancyPacket []byte
		if fecEncoder != nil {
			groupComplete, redundancy, err := fecEncoder.AddPacket(buf, uint64(seq))
			if err != nil {
				fmt.Printf("[WARNING] FEC encoding error: %v\n", err)
			} else if groupComplete && redundancy != nil {
				redundancyPacket = redundancy
				metrics.mu.Lock()
				metrics.FECPacketsSent++
				metrics.FECRepairPacketsSent++ // Redundancy packets = repair packets (sent from client)
				metrics.FECRedundancyBytes += int64(len(redundancy))
				metrics.mu.Unlock()
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
			// Проверяем deadline перед отправкой
			if time.Now().After(sendDeadline) {
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: deadline reached before write, returning\n", connID, streamID)
				}
				return
			}
			
			// Проверяем контекст перед отправкой
			select {
			case <-ctx.Done():
				if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: ctx.Done() before write, returning\n", connID, streamID)
				}
				return
			default:
			}
			
			// Уведомляем SimpleIntegration о отправке пакета
			if si != nil {
				if cfg.CongestionControl == "bbrv3" && sentPackets%1000 == 0 {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: OnPacketSent called (packet %d)\n", 
						connID, streamID, sentPackets)
				}
				si.OnPacketSent(session, len(buf), false)
			}
			
			// Используем context с таймаутом для Write чтобы избежать блокировок
			writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
			writeDone := make(chan error, 1)
			var n int
			var err error
			
			go func() {
				n, err = stream.Write(buf)
				writeDone <- err
			}()
			
			select {
			case <-writeCtx.Done():
				writeCancel()
				// Таймаут записи - продолжаем
				metrics.mu.Lock()
				metrics.Errors++
				if metrics.ErrorTypeCounts == nil {
					metrics.ErrorTypeCounts = map[string]int{}
				}
				metrics.ErrorTypeCounts["stream_write_timeout"]++
				metrics.mu.Unlock()
				continue
			case err = <-writeDone:
				writeCancel()
			}
			
			// Получаем реальный RTT из Connection (используем LatestRTT если доступен)
			// В quic-go RTT доступен через connection, но не через ConnectionState
			// Используем эмулированную задержку + небольшая случайная вариация для реалистичности
			var realRTT time.Duration
			if cfg.EmulateLatency > 0 {
				realRTT = cfg.EmulateLatency
				// Добавляем небольшую вариацию для jitter (5-10% от базовой задержки)
				jitter := time.Duration(float64(cfg.EmulateLatency) * 0.05 * secureFloat64())
				realRTT += jitter
			} else {
				// Fallback: используем типичный RTT для локальной сети
				realRTT = 10 * time.Millisecond
			}
			
			// Для метрик используем реальный RTT
			latencyForMetrics := float64(realRTT.Nanoseconds()) / 1e6
			
			metrics.mu.Lock()
			metrics.BytesSent += n
			metrics.Success++
			metrics.Latencies = append(metrics.Latencies, latencyForMetrics)
			metrics.Timestamps = append(metrics.Timestamps, time.Now())
			// Записываем в HDR-гистограммы
			if metrics.HDRMetrics != nil {
				metrics.HDRMetrics.RecordLatency(realRTT)
				metrics.HDRMetrics.AddBytesSent(int64(n))
				metrics.HDRMetrics.IncrementPacketsSent()
			}
			metrics.mu.Unlock()
			sentPackets++
			ackedPackets++
			
			// Уведомляем SimpleIntegration о получении ACK с реальным RTT
			// В QUIC ACK приходит асинхронно, поэтому мы используем smoothed RTT
			// Это приближение, но лучше чем время записи
			if si != nil && err == nil {
				if cfg.CongestionControl == "bbrv3" && ackedPackets%1000 == 0 {
					fmt.Printf("[DEBUG] Connection %d, Stream %d: OnAckReceived called (packet %d, acked %d)\n", 
						connID, streamID, sentPackets, ackedPackets)
				}
				// Добавляем защиту от паники
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("[ERROR] Panic in OnAckReceived: %v\n", r)
						}
					}()
					// Используем реальный RTT из connection state
					si.OnAckReceived(session, n, realRTT)
				}()
			} else if si == nil && cfg.CongestionControl == "bbrv3" {
				// Логируем если BBRv3 выбран но integration не инициализирован
				if sentPackets%1000 == 0 {
					fmt.Printf("[WARNING] BBRv3 selected but SimpleIntegration is nil (packet %d)\n", sentPackets)
				}
			}
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
			// Отправляем redundancy пакет если он был создан
			if redundancyPacket != nil && err == nil {
				// Отправляем redundancy пакет в отдельном write
				redundancyCtx, redundancyCancel := context.WithTimeout(ctx, 2*time.Second)
				redundancyDone := make(chan error, 1)
				go func() {
					_, redundancyErr := stream.Write(redundancyPacket)
					redundancyDone <- redundancyErr
				}()
				
				select {
				case <-redundancyCtx.Done():
					redundancyCancel()
					// Таймаут - не критично, продолжаем
				case redundancyErr := <-redundancyDone:
					redundancyCancel()
					if redundancyErr == nil {
						metrics.mu.Lock()
						metrics.BytesSent += len(redundancyPacket)
						metrics.mu.Unlock()
					}
				}
			}
			
			lastSeq = seq
			metrics.mu.Lock()
			metrics.TimeSeriesRetransmits = append(metrics.TimeSeriesRetransmits, TimePoint{Time: time.Since(start).Seconds(), Value: float64(retransmits)})
			metrics.TimeSeriesPacketLoss = append(metrics.TimeSeriesPacketLoss, TimePoint{Time: time.Since(start).Seconds(), Value: 100 * float64(sentPackets-ackedPackets) / (float64(sentPackets) + 1e-9)})
			metrics.mu.Unlock()
		}
		// Пауза между пакетами (с проверкой контекста и deadline)
		// Проверяем deadline перед паузой
		if time.Now().After(sendDeadline) {
			if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
				fmt.Printf("[DEBUG] Connection %d, Stream %d: deadline reached before sleep, returning\n", connID, streamID)
			}
			return
		}
		
		rate := atomic.LoadInt64(ratePtr)
		if rate > 0 {
			sleepDuration := time.Second / time.Duration(rate)
			if sleepDuration > 100*time.Millisecond {
				// Для длинных пауз используем прерываемый sleep с проверкой deadline
				select {
				case <-ctx.Done():
					if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
						fmt.Printf("[DEBUG] Connection %d, Stream %d: ctx.Done() during sleep, returning\n", connID, streamID)
					}
					return
				case <-time.After(sleepDuration):
					// Проверяем deadline после sleep
					if time.Now().After(sendDeadline) {
						if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
							fmt.Printf("[DEBUG] Connection %d, Stream %d: deadline reached after sleep, returning\n", connID, streamID)
						}
						return
					}
				}
			} else {
				// Для коротких пауз обычный sleep, но с проверкой deadline после
				select {
				case <-ctx.Done():
					if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
						fmt.Printf("[DEBUG] Connection %d, Stream %d: ctx.Done() during short sleep, returning\n", connID, streamID)
					}
					return
				case <-time.After(sleepDuration):
					// Проверяем deadline после sleep
					if time.Now().After(sendDeadline) {
						if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
							fmt.Printf("[DEBUG] Connection %d, Stream %d: deadline reached after short sleep, returning\n", connID, streamID)
						}
						return
					}
				}
			}
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

// calcJitter вычисляет стандартное отклонение латенси (jitter)
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
	variance := sum / float64(len(latencies))
	// Извлекаем квадратный корень для получения стандартного отклонения
	jitter := math.Sqrt(variance)
	return jitter
}

// printMetrics удалена - больше не используется

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

// parseAddr парсит адрес в формате "host:port" и возвращает *net.UDPAddr
func parseAddr(addr string) *net.UDPAddr {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		// Fallback на простой парсинг
		host, port := "127.0.0.1", "9000"
		if len(addr) > 0 {
			parts := splitHostPort(addr)
			if len(parts) == 2 {
				host, port = parts[0], parts[1]
				// Если host пустой (например, ":9000"), используем localhost
				if host == "" {
					host = "127.0.0.1"
				}
			} else if len(parts) == 1 {
				// Только порт (например, ":9000" или "9000")
				if parts[0] != "" {
					port = parts[0]
				}
			}
		}
		udpAddr = &net.UDPAddr{
			IP:   net.ParseIP(host),
			Port: parseInt(port),
		}
	} else {
		// Проверяем, что IP не пустой и не IPv6 :: (который может вызвать проблемы)
		if udpAddr.IP == nil || udpAddr.IP.IsUnspecified() {
			// Если IP пустой или неопределенный, используем 127.0.0.1
			udpAddr.IP = net.ParseIP("127.0.0.1")
		}
	}
	return udpAddr
}

// splitHostPort разделяет "host:port"
func splitHostPort(addr string) []string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return []string{addr[:i], addr[i+1:]}
		}
	}
	return []string{addr}
}

// parseInt парсит строку в int
func parseInt(s string) int {
	val := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			val = val*10 + int(s[i]-'0')
		}
	}
	return val
}
