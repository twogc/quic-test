package main

import (
	"context"
	"strings"
	"time"

	"quck-test/internal/ice"
	"quck-test/internal/integration"
	"quck-test/internal/masque"

	"go.uber.org/zap"
)

// runEnhancedTesting запускает расширенное тестирование
func runEnhancedTesting(logger *zap.Logger, masqueServer, masqueTargets, iceStunServers, iceTurnServers, iceTurnUser, iceTurnPass string) {
	logger.Info("Starting enhanced testing (MASQUE + ICE + QUIC)")

	// Парсим MASQUE целевые хосты
	masqueTargetsList := strings.Split(masqueTargets, ",")
	for i, target := range masqueTargetsList {
		masqueTargetsList[i] = strings.TrimSpace(target)
	}

	// Парсим STUN серверы
	stunList := []string{}
	if iceStunServers != "" {
		for _, server := range strings.Split(iceStunServers, ",") {
			if server != "" {
				stunList = append(stunList, strings.TrimSpace(server))
			}
		}
	}

	// Парсим TURN серверы
	turnList := []string{}
	if iceTurnServers != "" {
		for _, server := range strings.Split(iceTurnServers, ",") {
			if server != "" {
				turnList = append(turnList, strings.TrimSpace(server))
			}
		}
	}

	// Создаем конфигурации
	masqueConfig := &masque.MASQUEConfig{
		ServerURL:       masqueServer,
		UDPTargets:      masqueTargetsList,
		IPTargets:       []string{"8.8.8.8", "1.1.1.1"},
		ConnectTimeout:  30 * time.Second,
		TestTimeout:     60 * time.Second,
		ConcurrentTests: 5,
		TestDuration:    30 * time.Second,
	}

	iceConfig := &ice.ICEConfig{
		StunServers:       stunList,
		TurnServers:       turnList,
		TurnUsername:      iceTurnUser,
		TurnPassword:      iceTurnPass,
		GatheringTimeout:  30 * time.Second,
		ConnectionTimeout: 60 * time.Second,
		TestDuration:      30 * time.Second,
		ConcurrentTests:   5,
	}

	// Создаем конфигурацию для расширенного тестирования
	config := &integration.EnhancedConfig{
		MASQUE:          *masqueConfig,
		ICE:             *iceConfig,
		TestDuration:    60 * time.Second,
		ConcurrentTests: 5,
		EnableMASQUE:    true,
		EnableICE:       true,
	}

	// Создаем расширенный тестер
	tester := integration.NewEnhancedTester(logger, config)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := tester.Start(ctx); err != nil {
		logger.Fatal("Failed to start enhanced testing", zap.Error(err))
	}

	// Ждем завершения тестирования
	<-ctx.Done()

	// Получаем результаты
	metrics := tester.GetMetrics()

	logger.Info("Enhanced testing completed",
		zap.Int64("total_tests", metrics.TotalTests),
		zap.Int64("successful_tests", metrics.SuccessfulTests),
		zap.Float64("success_rate", metrics.SuccessRate),
		zap.Duration("test_duration", metrics.TestDuration))

	// Останавливаем тестер
	tester.Stop()
}
