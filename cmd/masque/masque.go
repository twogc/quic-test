package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"quic-test/internal/masque"

	"go.uber.org/zap"
)

// runMASQUETesting запускает MASQUE тестирование
func runMASQUETesting(logger *zap.Logger, masqueServer, masqueTargets string) {
	logger.Info("Starting MASQUE testing",
		zap.String("server", masqueServer),
		zap.String("targets", masqueTargets))

	// Парсим целевые хосты
	targets := strings.Split(masqueTargets, ",")
	for i, target := range targets {
		targets[i] = strings.TrimSpace(target)
	}

	// Создаем конфигурацию MASQUE
	config := &masque.MASQUEConfig{
		ServerURL:       masqueServer,
		UDPTargets:      targets,
		IPTargets:       []string{"8.8.8.8", "1.1.1.1"},
		ConnectTimeout:  30 * time.Second,
		TestTimeout:     60 * time.Second,
		ConcurrentTests: 5,
		TestDuration:    30 * time.Second,
	}

	// Создаем и запускаем MASQUE тестер
	tester := masque.NewMASQUETester(logger, config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := tester.Start(ctx); err != nil {
		logger.Fatal("Failed to start MASQUE testing", zap.Error(err))
	}

	// Ждем завершения тестирования
	<-ctx.Done()

	// Получаем результаты
	metrics := tester.GetMetrics()

	logger.Info("MASQUE testing completed",
		zap.Int64("connect_udp_successes", metrics.ConnectUDPSuccesses),
		zap.Int64("connect_ip_successes", metrics.ConnectIPSuccesses),
		zap.Float64("datagram_loss_rate", metrics.DatagramLossRate),
		zap.Float64("throughput_mbps", metrics.Throughput),
		zap.Duration("average_latency", metrics.AverageLatency))

	// Останавливаем тестер
	if err := tester.Stop(); err != nil {
		logger.Error("Failed to stop MASQUE tester", zap.Error(err))
	}
}

func main() {
	// Парсим аргументы командной строки
	masqueServer := flag.String("server", "localhost:8443", "MASQUE сервер для тестирования")
	masqueTargets := flag.String("targets", "8.8.8.8:53,1.1.1.1:53", "Целевые хосты для CONNECT-UDP (через запятую)")
	verbose := flag.Bool("verbose", false, "Подробный вывод")
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
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer logger.Sync()

	fmt.Println("🔥 Запуск MASQUE тестирования...")
	fmt.Printf("🌐 Сервер: %s\n", *masqueServer)
	fmt.Printf("Цели: %s\n", *masqueTargets)

	// Запускаем MASQUE тестирование
	runMASQUETesting(logger, *masqueServer, *masqueTargets)
}
