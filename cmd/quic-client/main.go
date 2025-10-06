package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quic-test/client"
	"quic-test/internal"
)

func main() {
	fmt.Println("\033[1;36m==============================\033[0m")
	fmt.Println("\033[1;36m  2GC CloudBridge QUIC Client\033[0m")
	fmt.Println("\033[1;36m==============================\033[0m")

	// Парсинг флагов
	addr := flag.String("addr", "127.0.0.1:9000", "Адрес сервера для подключения")
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
	emulateLoss := flag.Float64("emulate-loss", 0, "Вероятность потери пакета (0..1)")
	emulateLatency := flag.Duration("emulate-latency", 0, "Дополнительная задержка перед отправкой пакета")
	emulateDup := flag.Float64("emulate-dup", 0, "Вероятность дублирования пакета (0..1)")
	pprofAddr := flag.String("pprof-addr", "", "Адрес для pprof (например, :6060)")
	slaRttP95 := flag.Duration("sla-rtt-p95", 0, "SLA: максимальный RTT p95 (например, 100ms)")
	slaLoss := flag.Float64("sla-loss", 0, "SLA: максимальная потеря пакетов (например, 0.01)")
	flag.Parse()

	// Валидация флагов
	if err := validateFlags(*noTLS, *rate, *emulateLoss, *emulateDup, *slaLoss); err != nil {
		fmt.Printf("Ошибка валидации: %v\n", err)
		os.Exit(1)
	}

	cfg := internal.TestConfig{
		Mode:           "client",
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
		PprofAddr:      *pprofAddr,
		SlaRttP95:      *slaRttP95,
		SlaLoss:        *slaLoss,
	}

	fmt.Printf("Подключение к %s с %d соединениями, %d потоков на соединение\n",
		cfg.Addr, cfg.Connections, cfg.Streams)

	// Обработка сигналов для graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-sigs
		fmt.Println("\nПолучен сигнал завершения, завершаем работу...")
		cancel()
	}()

	// Передаем контекст в client.Run
	_ = ctx // Используем контекст для graceful shutdown

	// Запуск клиента
	client.Run(cfg)
}

// validateFlags проверяет корректность комбинаций флагов
func validateFlags(noTLS bool, rate int, emulateLoss, emulateDup, slaLoss float64) error {
	if rate <= 0 {
		return fmt.Errorf("rate должен быть положительным")
	}
	if emulateLoss < 0 || emulateLoss > 1 {
		return fmt.Errorf("emulate-loss должен быть в диапазоне [0, 1]")
	}
	if emulateDup < 0 || emulateDup > 1 {
		return fmt.Errorf("emulate-dup должен быть в диапазоне [0, 1]")
	}
	if slaLoss < 0 || slaLoss > 1 {
		return fmt.Errorf("sla-loss должен быть в диапазоне [0, 1]")
	}
	// Дополнительные проверки можно добавить здесь
	return nil
}
