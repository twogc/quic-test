package main

import (
	"context"
	"strings"
	"time"

	"quic-test/internal/ice"

	"go.uber.org/zap"
)

// runICETesting запускает ICE/STUN/TURN тестирование
func runICETesting(logger *zap.Logger, stunServers, turnServers, turnUser, turnPass string) {
	logger.Info("Starting ICE testing",
		zap.String("stun_servers", stunServers),
		zap.String("turn_servers", turnServers))

	// Парсим STUN серверы
	stunList := []string{}
	if stunServers != "" {
		for _, server := range strings.Split(stunServers, ",") {
			if server != "" {
				stunList = append(stunList, strings.TrimSpace(server))
			}
		}
	}

	// Парсим TURN серверы
	turnList := []string{}
	if turnServers != "" {
		for _, server := range strings.Split(turnServers, ",") {
			if server != "" {
				turnList = append(turnList, strings.TrimSpace(server))
			}
		}
	}

	// Создаем конфигурацию ICE
	iceConfig := &ice.ICEConfig{
		StunServers:       stunList,
		TurnServers:       turnList,
		TurnUsername:      turnUser,
		TurnPassword:      turnPass,
		GatheringTimeout:  30 * time.Second,
		ConnectionTimeout: 60 * time.Second,
		TestDuration:      30 * time.Second,
		ConcurrentTests:   5,
	}

	// Создаем и запускаем ICE тестер
	tester := ice.NewICETester(logger, iceConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := tester.Start(ctx); err != nil {
		logger.Fatal("Failed to start ICE testing", zap.Error(err))
	}

	// Ждем завершения тестирования
	<-ctx.Done()

	// Получаем результаты
	metrics := tester.GetMetrics()

	logger.Info("ICE testing completed",
		zap.Int64("stun_requests", metrics.StunRequests),
		zap.Int64("stun_responses", metrics.StunResponses),
		zap.Int64("turn_allocations", metrics.TurnAllocations),
		zap.Int64("candidates_gathered", metrics.CandidatesGathered),
		zap.Int64("connections_successful", metrics.ConnectionsSuccessful))

	// Останавливаем тестер
	if err := tester.Stop(); err != nil {
		logger.Error("Failed to stop ICE tester", zap.Error(err))
	}
}
