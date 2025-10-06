package main

import (
	"context"
	"strings"
	"time"

	"quck-test/internal/masque"

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
	tester.Stop()
}
