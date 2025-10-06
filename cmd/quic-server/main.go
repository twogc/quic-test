package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quck-test/internal"
	"quck-test/server"
)

func main() {
	fmt.Println("\033[1;36m==============================\033[0m")
	fmt.Println("\033[1;36m  2GC CloudBridge QUIC Server\033[0m")
	fmt.Println("\033[1;36m==============================\033[0m")

	// Парсинг флагов
	addr := flag.String("addr", ":9000", "Адрес для прослушивания")
	certPath := flag.String("cert", "", "Путь к TLS-сертификату (опционально)")
	keyPath := flag.String("key", "", "Путь к TLS-ключу (опционально)")
	noTLS := flag.Bool("no-tls", false, "Отключить TLS (для тестов)")
	prometheus := flag.Bool("prometheus", false, "Экспортировать метрики Prometheus на /metrics")
	pprofAddr := flag.String("pprof-addr", "", "Адрес для pprof (например, :6060)")
	flag.Parse()

	// Валидация флагов
	if err := validateFlags(*noTLS, *certPath, *keyPath); err != nil {
		fmt.Printf("Ошибка валидации: %v\n", err)
		os.Exit(1)
	}

	cfg := internal.TestConfig{
		Mode:       "server",
		Addr:       *addr,
		CertPath:   *certPath,
		KeyPath:    *keyPath,
		NoTLS:      *noTLS,
		Prometheus: *prometheus,
		PprofAddr:  *pprofAddr,
	}

	fmt.Printf("Запуск QUIC сервера на %s\n", cfg.Addr)
	if cfg.Prometheus {
		fmt.Println("Prometheus метрики будут доступны на /metrics")
	}
	if cfg.PprofAddr != "" {
		fmt.Printf("pprof будет доступен на %s/debug/pprof\n", cfg.PprofAddr)
	}

	// Обработка сигналов для graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nПолучен сигнал завершения, остановка сервера...")
		os.Exit(0)
	}()

	// Запуск сервера
	server.Run(cfg)
}

// validateFlags проверяет корректность комбинаций флагов
func validateFlags(noTLS bool, certPath, keyPath string) error {
	if !noTLS && certPath != "" && keyPath == "" {
		return fmt.Errorf("если указан cert, должен быть указан key")
	}
	if !noTLS && certPath == "" && keyPath != "" {
		return fmt.Errorf("если указан key, должен быть указан cert")
	}
	return nil
}
