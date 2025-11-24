package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"quic-test/client"
	"quic-test/internal"
	"quic-test/server"
)

func main() {
	// Добавляем флаг --version
	version := flag.Bool("version", false, "Показать версию программы")
	
	fmt.Println("\033[1;36m==========================================\033[0m")
	fmt.Println("\033[1;36m    2GC Network Protocol Suite\033[0m")
	fmt.Println("\033[1;36m==========================================\033[0m")
	fmt.Println("Комплексное тестирование QUIC, MASQUE, ICE/STUN/TURN и других сетевых протоколов")
	mode := flag.String("mode", "test", "Режим: server | client | test")
	addr := flag.String("addr", ":9000", "Адрес для подключения или прослушивания")
	streams := flag.Int("streams", 1, "Количество потоков на соединение")
	connections := flag.Int("connections", 1, "Количество QUIC-соединений")
	duration := flag.Duration("duration", 0, "Длительность теста (0 — до ручного завершения)")
	packetSize := flag.Int("packet-size", 1200, "Размер пакета (байт)")
	rate := flag.Int("rate", 100, "Частота отправки пакетов (в секунду)")
	reportPath := flag.String("report", "", "Путь к файлу для отчета (опционально)")
	reportFormat := flag.String("report-format", "md", "Формат отчета: csv | md | json")
	certPath := flag.String("cert", "", "Путь к TLS-сертификату (опционально)")
	keyPath := flag.String("key", "", "Путь к TLS-ключу (опционально)")
	pattern := flag.String("pattern", "random", "Шаблон данных: random | zeroes | increment")
	noTLS := flag.Bool("no-tls", false, "Отключить TLS (для тестов)")
	prometheus := flag.Bool("prometheus", false, "Экспортировать метрики Prometheus на /metrics")
	quicBottom := flag.Bool("quic-bottom", false, "Запустить QUIC Bottom для визуализации метрик")
	emulateLoss := flag.Float64("emulate-loss", 0, "Вероятность потери пакета (0..1)")
	emulateLatency := flag.Duration("emulate-latency", 0, "Дополнительная задержка перед отправкой пакета (например, 20ms)")
	emulateDup := flag.Float64("emulate-dup", 0, "Вероятность дублирования пакета (0..1)")
	
	// FEC флаги
	fecEnabled := flag.Bool("enable-fec", false, "Включить Forward Error Correction")
	fecRate := flag.Float64("fec-rate", 0.10, "Уровень избыточности FEC (0.05-0.20, например 0.05=5%, 0.10=10%, 0.20=20%)")
	// Alias для обратной совместимости
	fecEnabledAlias := flag.Bool("fec", false, "Alias для --enable-fec")
	fecRedundancyAlias := flag.Float64("fec-redundancy", 0.10, "Alias для --fec-rate")
	
	// PQC флаги
	pqcEnabled := flag.Bool("pqc", false, "Включить Post-Quantum Cryptography (симуляция)")
	pqcAlgorithm := flag.String("pqc-algorithm", "ml-kem-768", "PQC алгоритм: ml-kem-512, ml-kem-768, dilithium-2, hybrid, baseline")
	
	// SLA флаги
	slaRttP95 := flag.Duration("sla-rtt-p95", 0, "SLA: максимальный RTT p95 (например, 100ms)")
	slaLoss := flag.Float64("sla-loss", 0, "SLA: максимальная потеря пакетов (0..1, например, 0.01 для 1%)")
	slaThroughput := flag.Float64("sla-throughput", 0, "SLA: минимальная пропускная способность (KB/s)")
	slaErrors := flag.Int64("sla-errors", 0, "SLA: максимальное количество ошибок")
	
	// QUIC тюнинг флаги
	cc := flag.String("cc", "", "Алгоритм управления перегрузкой: cubic, bbr, bbrv2, bbrv3, reno")
	maxIdleTimeout := flag.Duration("max-idle-timeout", 0, "Максимальное время простоя соединения")
	handshakeTimeout := flag.Duration("handshake-timeout", 0, "Таймаут handshake")
	keepAlive := flag.Duration("keep-alive", 0, "Интервал keep-alive")
	maxStreams := flag.Int64("max-streams", 0, "Максимальное количество потоков")
	maxStreamData := flag.Int64("max-stream-data", 0, "Максимальный размер данных потока")
	enable0RTT := flag.Bool("enable-0rtt", false, "Включить 0-RTT")
	enableKeyUpdate := flag.Bool("enable-key-update", false, "Включить key update")
	enableDatagrams := flag.Bool("enable-datagrams", false, "Включить datagrams")
	maxIncomingStreams := flag.Int64("max-incoming-streams", 0, "Максимальное количество входящих потоков")
	maxIncomingUniStreams := flag.Int64("max-incoming-uni-streams", 0, "Максимальное количество входящих unidirectional потоков")
	
	// Сценарии тестирования
	scenario := flag.String("scenario", "", "Предустановленный сценарий: wifi, lte, sat, dc-eu, ru-eu, loss-burst, reorder")
	listScenarios := flag.Bool("list-scenarios", false, "Показать список доступных сценариев")
	
	// Сетевые профили
	networkProfile := flag.String("network-profile", "", "Сетевой профиль: wifi, lte, 5g, satellite, ethernet, fiber, datacenter")
	listProfiles := flag.Bool("list-profiles", false, "Показать список доступных сетевых профилей")
	
	flag.Parse()

	// Обработка флага --version
	if *version {
		internal.PrintVersion()
		os.Exit(0)
	}

	cfg := internal.TestConfig{
		Mode:           *mode,
		Addr:           *addr,
		Streams:        *streams,
		Connections:    *connections,
		Duration:       *duration,
		PacketSize:     *packetSize,
		Rate:           *rate,
		ReportPath:     *reportPath,
		ReportFormat:   *reportFormat,
		CertPath:       *certPath,
		KeyPath:        *keyPath,
		Pattern:        *pattern,
		NoTLS:          *noTLS,
		Prometheus:     *prometheus,
		EmulateLoss:    *emulateLoss,
		EmulateLatency: *emulateLatency,
		EmulateDup:     *emulateDup,
		SlaRttP95:      *slaRttP95,
		SlaLoss:        *slaLoss,
		SlaThroughput:  *slaThroughput,
		SlaErrors:      *slaErrors,
		CongestionControl: *cc,
		MaxIdleTimeout:    *maxIdleTimeout,
		HandshakeTimeout:  *handshakeTimeout,
		KeepAlive:         *keepAlive,
		MaxStreams:        *maxStreams,
		MaxStreamData:      *maxStreamData,
		Enable0RTT:        *enable0RTT,
		EnableKeyUpdate:   *enableKeyUpdate,
		EnableDatagrams:   *enableDatagrams,
		MaxIncomingStreams: *maxIncomingStreams,
		MaxIncomingUniStreams: *maxIncomingUniStreams,
		FECEnabled:       *fecEnabled || *fecEnabledAlias,
		FECRedundancy:    func() float64 {
			if *fecEnabled || *fecEnabledAlias {
				if *fecRedundancyAlias != 0.10 {
					return *fecRedundancyAlias
				}
				return *fecRate
			}
			return 0
		}(),
		PQCEnabled:       *pqcEnabled,
		PQCAlgorithm:     *pqcAlgorithm,
	}

	fmt.Printf("mode=%s, addr=%s, connections=%d, streams=%d, duration=%s, packet-size=%d, rate=%d, report=%s, report-format=%s, cert=%s, key=%s, pattern=%s, no-tls=%v, prometheus=%v\n",
		cfg.Mode, cfg.Addr, cfg.Connections, cfg.Streams, cfg.Duration.String(), cfg.PacketSize, cfg.Rate, cfg.ReportPath, cfg.ReportFormat, cfg.CertPath, cfg.KeyPath, cfg.Pattern, cfg.NoTLS, cfg.Prometheus)
	
	// Выводим SLA конфигурацию если настроена
	internal.PrintSLAConfig(cfg)
	
	// Выводим QUIC конфигурацию если настроена
	internal.PrintQUICConfig(cfg)
	
	// Запуск QUIC Bottom если запрошен
	if *quicBottom {
		fmt.Println("Starting QUIC Bottom for real-time metrics visualization...")
		go func() {
			// Запускаем QUIC Bottom в фоновом режиме
			cmd := exec.Command("./quic-bottom/target/release/quic-bottom-real")
			cmd.Dir = "."
			if err := cmd.Run(); err != nil {
				fmt.Printf("❌ Failed to start QUIC Bottom: %v\n", err)
			}
		}()
		
		// Ждем немного, чтобы QUIC Bottom запустился
		time.Sleep(2 * time.Second)
		fmt.Println("✅ QUIC Bottom started on port 8080")
	}

	// Обработка сценариев
	if *listScenarios {
		fmt.Println("Available Test Scenarios:")
		scenarios := internal.ListScenarios()
		for _, name := range scenarios {
			scenario, _ := internal.GetScenario(name)
			fmt.Printf("  - %s: %s\n", name, scenario.Description)
		}
		os.Exit(0)
	}
	
	// Обработка сетевых профилей
	if *listProfiles {
		fmt.Println("Available Network Profiles:")
		profiles := internal.ListNetworkProfiles()
		for _, name := range profiles {
			profile, _ := internal.GetNetworkProfile(name)
			fmt.Printf("  - %s: %s\n", name, profile.Description)
		}
		os.Exit(0)
	}
	
	if *scenario != "" {
		scenarioConfig, err := internal.GetScenario(*scenario)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		
		// Применяем конфигурацию сценария
		cfg = scenarioConfig.Config
		fmt.Printf("Running scenario: %s\n", scenarioConfig.Name)
	}
	
	if *networkProfile != "" {
		profile, err := internal.GetNetworkProfile(*networkProfile)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		
		// Применяем сетевой профиль
		internal.ApplyNetworkProfile(&cfg, profile)
		internal.PrintNetworkProfile(profile)
		internal.PrintProfileRecommendations(profile)
	}

	// Инициализация QUIC Bottom (используем 127.0.0.1 вместо localhost для избежания IPv6 проблем)
	internal.InitBottomBridge("http://127.0.0.1:8080", 100*time.Millisecond)
	internal.EnableBottomBridge()

	// Обработка сигналов для graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(cancelFunc context.CancelFunc) {
		<-sigs
		fmt.Println("\nПолучен сигнал завершения, завершаем работу...")
		cancelFunc() // Корректное завершение
	}(cancel)

	switch cfg.Mode {
	case "server":
		fmt.Println("Запуск в режиме сервера...")
		server.Run(cfg)
	case "client":
		fmt.Println("Запуск в режиме клиента...")
		client.Run(cfg)
	case "test":
		fmt.Println("Запуск в режиме теста (сервер+клиент)...")
		runTestMode(cfg)
	default:
		fmt.Println("Неизвестный режим", cfg.Mode)
		os.Exit(1)
	}
}

// runTestMode запускает сервер и клиент для тестирования
func runTestMode(cfg internal.TestConfig) {
	// Запускаем сервер в горутине
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		server.Run(cfg)
	}()

	// Ждем, чтобы сервер запустился
	time.Sleep(3 * time.Second)

	// Запускаем клиент
	client.Run(cfg)

	// Даем серверу время на завершение gracefully (максимум 5 секунд)
	serverTimeout := time.NewTimer(5 * time.Second)
	select {
	case <-serverDone:
		serverTimeout.Stop()
	case <-serverTimeout.C:
		fmt.Println("Server shutdown timeout, exiting...")
	}
}
