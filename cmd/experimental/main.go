package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"quic-test/internal/experimental"
	"quic-test/internal/sla"

	"go.uber.org/zap"
)

func main() {
	// Базовые флаги
	addr := flag.String("addr", ":9000", "Address to listen/connect")
	mode := flag.String("mode", "test", "Mode: server, client, test")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	
	// Экспериментальные флаги QUIC
	cc := flag.String("cc", "cubic", "Congestion control: cubic, bbr, bbrv2, reno")
	qlog := flag.String("qlog", "", "qlog output directory")
	ackFreq := flag.Int("ack-freq", 0, "ACK frequency (0=auto)")
	maxAckDelay := flag.Duration("max-ack-delay", 25*time.Millisecond, "Max ACK delay")
	
	// Multipath QUIC
	multipath := flag.String("mp", "", "Multipath addresses (comma-separated)")
	mpStrategy := flag.String("mp-strategy", "round-robin", "Multipath strategy: round-robin, lowest-rtt, highest-bw")
	
	// FEC для datagrams
	fec := flag.Bool("fec", false, "Enable FEC for datagrams")
	fecRedundancy := flag.Float64("fec-redundancy", 0.1, "FEC redundancy factor (0.1 = 10%)")
	
	// Greasing
	greasing := flag.Bool("greasing", true, "Enable QUIC bit greasing (RFC 9287)")
	
	// Производительность
	gso := flag.Bool("gso", true, "Enable UDP GSO (if supported)")
	gro := flag.Bool("gro", true, "Enable UDP GRO (if supported)")
	socketBuffer := flag.Int("socket-buffer", 1024*1024, "Socket buffer size (bytes)")
	
	// Наблюдаемость
	tracing := flag.Bool("tracing", true, "Enable OpenTelemetry tracing")
	metricsInterval := flag.Duration("metrics-interval", 1*time.Second, "Metrics collection interval")
	
	// Тестовые параметры
	connections := flag.Int("connections", 1, "Number of connections")
	streams := flag.Int("streams", 1, "Number of streams per connection")
	duration := flag.Duration("duration", 30*time.Second, "Test duration")
	packetSize := flag.Int("packet-size", 1200, "Packet size (bytes)")
	rate := flag.Int("rate", 100, "Packets per second")
	
	// SLA-гейты для CI
	slaP95RTT := flag.Float64("sla-p95-rtt", 0, "SLA: 95th percentile RTT limit (ms, 0=disabled)")
	slaLoss := flag.Float64("sla-loss", 0, "SLA: Loss rate limit (%, 0=disabled)")
	slaGoodput := flag.Float64("sla-goodput", 0, "SLA: Minimum goodput (Mbps, 0=disabled)")
	slaMaxRTT := flag.Float64("sla-max-rtt", 0, "SLA: Maximum RTT limit (ms, 0=disabled)")
	slaMeanRTT := flag.Float64("sla-mean-rtt", 0, "SLA: Mean RTT limit (ms, 0=disabled)")
	slaThroughput := flag.Float64("sla-throughput", 0, "SLA: Minimum throughput (Mbps, 0=disabled)")
	slaBandwidth := flag.Float64("sla-bandwidth", 0, "SLA: Minimum bandwidth (bps, 0=disabled)")
	slaACKDelay := flag.Float64("sla-ack-delay", 0, "SLA: Maximum ACK delay (ms, 0=disabled)")
	
	// SLA профили
	slaProfile := flag.String("sla-profile", "", "SLA profile: strict, normal, lenient")
	
	flag.Parse()

	// Создаем логгер
	var logger *zap.Logger
	var err error
	
	if *verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	fmt.Println("🚀 2GC Network Protocol Suite - Experimental QUIC")
	fmt.Println("================================================")
	
	// Создаем экспериментальную конфигурацию
	expConfig := &experimental.ExperimentalConfig{
		// Базовые настройки
		Addr:           *addr,
		Mode:           *mode,
		Connections:    *connections,
		Streams:        *streams,
		Duration:       *duration,
		PacketSize:     *packetSize,
		Rate:           *rate,
		
		// Экспериментальные настройки QUIC
		CongestionControl: *cc,
		QlogDir:          *qlog,
		ACKFrequency:     *ackFreq,
		MaxACKDelay:      *maxAckDelay,
		
		// Multipath
		Multipath:        parseMultipathAddresses(*multipath),
		MultipathStrategy: *mpStrategy,
		
		// FEC
		EnableFEC:        *fec,
		FECRedundancy:    *fecRedundancy,
		
		// Greasing
		EnableGreasing:   *greasing,
		
		// Производительность
		EnableGSO:        *gso,
		EnableGRO:        *gro,
		SocketBufferSize: *socketBuffer,
		
		// Наблюдаемость
		EnableTracing:     *tracing,
		MetricsInterval:  *metricsInterval,
	}
	
	// Валидация конфигурации
	if err := expConfig.Validate(); err != nil {
		logger.Fatal("Invalid configuration", zap.Error(err))
	}
	
	// Настройка SLA-гейтов
	var slaGates *sla.SLAGates
	if *slaProfile != "" {
		switch *slaProfile {
		case "strict":
			slaGates = sla.NewSLAGatesStrict()
			fmt.Println("🔒 Using STRICT SLA profile")
		case "lenient":
			slaGates = sla.NewSLAGatesLenient()
			fmt.Println("🔓 Using LENIENT SLA profile")
		case "normal":
			slaGates = sla.NewSLAGates()
			fmt.Println("⚖️  Using NORMAL SLA profile")
		default:
			logger.Fatal("Invalid SLA profile", zap.String("profile", *slaProfile))
		}
	} else {
		// Создаем SLA-гейты с пользовательскими значениями
		slaGates = sla.NewSLAGates()
		
		// Применяем пользовательские значения
		if *slaP95RTT > 0 {
			slaGates.P95RTTMs = *slaP95RTT
		}
		if *slaLoss > 0 {
			slaGates.LossRatePercent = *slaLoss
		}
		if *slaGoodput > 0 {
			slaGates.MinGoodputMbps = *slaGoodput
		}
		if *slaMaxRTT > 0 {
			slaGates.MaxRTTMs = *slaMaxRTT
		}
		if *slaMeanRTT > 0 {
			slaGates.MeanRTTMs = *slaMeanRTT
		}
		if *slaThroughput > 0 {
			slaGates.MinThroughputMbps = *slaThroughput
		}
		if *slaBandwidth > 0 {
			slaGates.MinBandwidthBps = *slaBandwidth
		}
		if *slaACKDelay > 0 {
			slaGates.MaxACKDelayMs = *slaACKDelay
		}
		
		// Проверяем, есть ли хотя бы один SLA флаг
		hasSLA := *slaP95RTT > 0 || *slaLoss > 0 || *slaGoodput > 0 || *slaMaxRTT > 0 || 
		         *slaMeanRTT > 0 || *slaThroughput > 0 || *slaBandwidth > 0 || *slaACKDelay > 0
		
		if hasSLA {
			fmt.Println("🎯 Using CUSTOM SLA gates")
		}
	}
	
	// Выводим конфигурацию
	expConfig.Print()
	
	// Создаем экспериментальный менеджер
	expManager := experimental.NewExperimentalManager(logger, expConfig)
	
	// Обработка сигналов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigs
		logger.Info("Received shutdown signal")
		cancel()
	}()
	
	// Запускаем в зависимости от режима
	switch *mode {
	case "server":
		logger.Info("Starting experimental QUIC server")
		if err := expManager.StartServer(ctx); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	case "client":
		logger.Info("Starting experimental QUIC client")
		// Создаем контекст с таймаутом для клиента
		clientCtx, clientCancel := context.WithTimeout(ctx, expConfig.Duration)
		defer clientCancel()
		
		if err := expManager.StartClient(clientCtx); err != nil {
			logger.Fatal("Failed to start client", zap.Error(err))
		}
	case "test":
		logger.Info("Starting experimental QUIC test")
		if err := expManager.RunTest(ctx); err != nil {
			logger.Fatal("Failed to run test", zap.Error(err))
		}
	default:
		logger.Fatal("Unknown mode", zap.String("mode", *mode))
	}
	
	// Ждем завершения
	<-ctx.Done()
	logger.Info("Experimental QUIC test completed")
}

// parseMultipathAddresses парсит адреса для multipath
func parseMultipathAddresses(mp string) []string {
	if mp == "" {
		return nil
	}
	
	addresses := strings.Split(mp, ",")
	for i, addr := range addresses {
		addresses[i] = strings.TrimSpace(addr)
	}
	
	return addresses
}
