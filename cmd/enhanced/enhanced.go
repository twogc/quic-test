package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"quic-test/internal/ice"
	"quic-test/internal/integration"
	"quic-test/internal/masque"

	"go.uber.org/zap"
)

func main() {
	// Флаги командной строки
	masqueServer := flag.String("masque-server", "", "MASQUE server URL")
	masqueTargets := flag.String("masque-targets", "", "MASQUE target hosts (comma-separated)")
	iceStunServers := flag.String("ice-stun", "", "ICE STUN servers (comma-separated)")
	iceTurnServers := flag.String("ice-turn", "", "ICE TURN servers (comma-separated)")
	iceTurnUser := flag.String("ice-turn-user", "", "ICE TURN username")
	iceTurnPass := flag.String("ice-turn-pass", "", "ICE TURN password")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	
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

	// Запускаем расширенное тестирование
	runEnhancedTesting(logger, *masqueServer, *masqueTargets, *iceStunServers, *iceTurnServers, *iceTurnUser, *iceTurnPass)
}

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
	if err := tester.Stop(); err != nil {
		logger.Error("Failed to stop enhanced tester", zap.Error(err))
	}
}
