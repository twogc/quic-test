package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quck-test/client"
	"quck-test/internal"
	"quck-test/server"
)

func main() {
	fmt.Println("\033[1;36m==============================\033[0m")
	fmt.Println("\033[1;36m  2GC CloudBridge QUICK testing\033[0m")
	fmt.Println("\033[1;36m==============================\033[0m")
	fmt.Println("Тестирование производительности и стабильности QUIC-протокола для CloudBridge 2GC")
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
	flag.Parse()

	cfg := internal.TestConfig{
		Mode:         *mode,
		Addr:         *addr,
		Streams:      *streams,
		Connections:  *connections,
		Duration:     *duration,
		PacketSize:   *packetSize,
		Rate:         *rate,
		ReportPath:   *reportPath,
		ReportFormat: *reportFormat,
		CertPath:     *certPath,
		KeyPath:      *keyPath,
		Pattern:      *pattern,
		NoTLS:        *noTLS,
		Prometheus:   *prometheus,
	}

	fmt.Printf("mode=%s, addr=%s, connections=%d, streams=%d, duration=%s, packet-size=%d, rate=%d, report=%s, report-format=%s, cert=%s, key=%s, pattern=%s, no-tls=%v, prometheus=%v\n",
		cfg.Mode, cfg.Addr, cfg.Connections, cfg.Streams, cfg.Duration.String(), cfg.PacketSize, cfg.Rate, cfg.ReportPath, cfg.ReportFormat, cfg.CertPath, cfg.KeyPath, cfg.Pattern, cfg.NoTLS, cfg.Prometheus)

	// Обработка сигналов для graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nПолучен сигнал завершения, завершаем работу...")
		os.Exit(0)
	}()

	switch cfg.Mode {
	case "server":
		fmt.Println("Запуск в режиме сервера...")
		server.Run(cfg)
	case "client":
		fmt.Println("Запуск в режиме клиента...")
		client.Run(cfg)
	case "test":
		fmt.Println("Запуск в режиме теста (сервер+клиент)...")
		// TODO: запуск сервера и клиента в одном процессе
	default:
		fmt.Println("Неизвестный режим", cfg.Mode)
		os.Exit(1)
	}
} 