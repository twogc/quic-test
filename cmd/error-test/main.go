package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quic-test/internal/experimental"

	"go.uber.org/zap"
)

func main() {
	// Базовые флаги
	scenario := flag.String("scenario", "", "Error testing scenario (use -list-scenarios to see available)")
	listScenarios := flag.Bool("list-scenarios", false, "List available error testing scenarios")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	
	// Настройки тестирования
	duration := flag.Duration("duration", 5*time.Minute, "Test duration")
	concurrent := flag.Int("concurrent", 5, "Number of concurrent tests")
	randomSeed := flag.Int64("seed", 0, "Random seed (0 = auto)")
	
	// Настройки ошибок
	networkErrors := flag.Bool("network-errors", true, "Enable network error simulation")
	packetLoss := flag.Float64("packet-loss", 0.05, "Packet loss rate (0.0-1.0)")
	packetDup := flag.Float64("packet-dup", 0.02, "Packet duplication rate (0.0-1.0)")
	packetReorder := flag.Bool("packet-reorder", true, "Enable packet reordering")
	packetCorrupt := flag.Float64("packet-corrupt", 0.01, "Packet corruption rate (0.0-1.0)")
	
	// QUIC ошибки
	quicErrors := flag.Bool("quic-errors", true, "Enable QUIC protocol errors")
	streamErrors := flag.Bool("stream-errors", true, "Enable stream errors")
	handshakeErrors := flag.Bool("handshake-errors", true, "Enable handshake errors")
	versionErrors := flag.Bool("version-errors", true, "Enable version errors")
	
	// Экспериментальные ошибки
	ackFreqErrors := flag.Bool("ack-freq-errors", true, "Enable ACK frequency errors")
	ccErrors := flag.Bool("cc-errors", true, "Enable congestion control errors")
	multipathErrors := flag.Bool("multipath-errors", true, "Enable multipath errors")
	fecErrors := flag.Bool("fec-errors", true, "Enable FEC errors")
	
	// Задержки
	latencyVar := flag.Duration("latency-var", 50*time.Millisecond, "Latency variation")
	jitterVar := flag.Duration("jitter-var", 10*time.Millisecond, "Jitter variation")
	connectionDrops := flag.Bool("connection-drops", true, "Enable connection drops")
	
	// Выходные данные
	output := flag.String("output", "", "Output file for results (JSON format)")
	report := flag.Bool("report", true, "Generate detailed report")
	
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

	fmt.Println("🧪 2GC Network Protocol Suite - Error Testing")
	fmt.Println("=============================================")
	
	// Показываем доступные сценарии
	if *listScenarios {
		fmt.Println("Available Error Testing Scenarios:")
		scenarios := experimental.ListErrorScenarios()
		for _, name := range scenarios {
			scenario, _ := experimental.GetErrorScenario(name)
			fmt.Printf("  - %s: %s\n", name, scenario.Description)
		}
		os.Exit(0)
	}
	
	// Создаем конфигурацию тестирования
	config := &experimental.ErrorTestingConfig{
		Duration:        *duration,
		ConcurrentTests: *concurrent,
		RandomSeed:      *randomSeed,
		
		// Сетевые ошибки
		NetworkErrors:     *networkErrors,
		PacketLoss:        *packetLoss,
		PacketDuplication: *packetDup,
		PacketReordering:  *packetReorder,
		PacketCorruption:  *packetCorrupt,
		
		// Задержки
		LatencyVariation: *latencyVar,
		JitterVariation:  *jitterVar,
		ConnectionDrops:  *connectionDrops,
		
		// QUIC ошибки
		QUICErrors:      *quicErrors,
		StreamErrors:       *streamErrors,
		HandshakeErrors:   *handshakeErrors,
		VersionErrors:     *versionErrors,
		
		// Экспериментальные ошибки
		ACKFrequencyErrors: *ackFreqErrors,
		CCErrors:          *ccErrors,
		MultipathErrors:   *multipathErrors,
		FECErrors:         *fecErrors,
	}
	
	// Создаем систему тестирования
	errorSuite := experimental.NewErrorTestingSuite(logger, config)
	
	// Добавляем сценарии
	if *scenario != "" {
		// Используем предустановленный сценарий
		scenarioConfig, err := experimental.GetErrorScenario(*scenario)
		if err != nil {
			logger.Fatal("Failed to get error scenario", zap.Error(err))
		}
		errorSuite.AddScenario(scenarioConfig)
	} else {
		// Создаем пользовательский сценарий
		customScenario := &experimental.ErrorScenario{
			Name:        "Custom Error Test",
			Description: "User-defined error testing scenario",
			Config:      config,
			Duration:    *duration,
			Expected: &experimental.ExpectedErrorResults{
				MaxErrorRate:     0.20,
				MinRecoveryTime:  5 * time.Second,
				MaxLatency:       500 * time.Millisecond,
				MinThroughput:    50.0,
				MaxPacketLoss:    0.20,
			},
		}
		errorSuite.AddScenario(customScenario)
	}
	
	// Выводим конфигурацию
	printConfig(config, *scenario)
	
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
	
	// Запускаем тестирование
	logger.Info("Starting error testing",
		zap.Duration("duration", *duration),
		zap.Int("concurrent_tests", *concurrent),
		zap.String("scenario", *scenario))
	
	if err := errorSuite.Start(ctx); err != nil {
		logger.Fatal("Failed to start error testing", zap.Error(err))
	}
	
	// Ждем завершения
	<-ctx.Done()
	
	// Останавливаем тестирование
	errorSuite.Stop()
	
	// Получаем результаты
	results := errorSuite.GetResults()
	
	// Выводим результаты
	printResults(results)
	
	// Сохраняем результаты в файл
	if *output != "" {
		if err := saveResults(results, *output); err != nil {
			logger.Error("Failed to save results", zap.Error(err))
		} else {
			logger.Info("Results saved", zap.String("file", *output))
		}
	}
	
	// Генерируем отчет
	if *report {
		generateReport(results)
	}
	
	logger.Info("Error testing completed")
}

// printConfig выводит конфигурацию тестирования
func printConfig(config *experimental.ErrorTestingConfig, scenario string) {
	fmt.Printf("🔧 Error Testing Configuration:\n")
	fmt.Printf("  - Duration: %v\n", config.Duration)
	fmt.Printf("  - Concurrent Tests: %d\n", config.ConcurrentTests)
	fmt.Printf("  - Random Seed: %d\n", config.RandomSeed)
	
	if scenario != "" {
		fmt.Printf("  - Scenario: %s\n", scenario)
	} else {
		fmt.Printf("  - Scenario: Custom\n")
	}
	
	fmt.Printf("\nNetwork Errors:\n")
	fmt.Printf("  - Network Errors: %v\n", config.NetworkErrors)
	fmt.Printf("  - Packet Loss: %.1f%%\n", config.PacketLoss*100)
	fmt.Printf("  - Packet Duplication: %.1f%%\n", config.PacketDuplication*100)
	fmt.Printf("  - Packet Reordering: %v\n", config.PacketReordering)
	fmt.Printf("  - Packet Corruption: %.1f%%\n", config.PacketCorruption*100)
	fmt.Printf("  - Latency Variation: %v\n", config.LatencyVariation)
	fmt.Printf("  - Jitter Variation: %v\n", config.JitterVariation)
	fmt.Printf("  - Connection Drops: %v\n", config.ConnectionDrops)
	
	fmt.Printf("\n🔧 QUIC Protocol Errors:\n")
	fmt.Printf("  - QUIC Errors: %v\n", config.QUICErrors)
	fmt.Printf("  - Stream Errors: %v\n", config.StreamErrors)
	fmt.Printf("  - Handshake Errors: %v\n", config.HandshakeErrors)
	fmt.Printf("  - Version Errors: %v\n", config.VersionErrors)
	
	fmt.Printf("\n🚀 Experimental Errors:\n")
	fmt.Printf("  - ACK Frequency Errors: %v\n", config.ACKFrequencyErrors)
	fmt.Printf("  - Congestion Control Errors: %v\n", config.CCErrors)
	fmt.Printf("  - Multipath Errors: %v\n", config.MultipathErrors)
	fmt.Printf("  - FEC Errors: %v\n", config.FECErrors)
	
	fmt.Println()
}

// printResults выводит результаты тестирования
func printResults(results *experimental.ErrorTestingResults) {
	fmt.Printf("Error Testing Results:\n")
	fmt.Printf("  - Duration: %v\n", results.Duration)
	fmt.Printf("  - Total Tests: %d\n", results.TotalTests)
	fmt.Printf("  - Passed Tests: %d\n", results.PassedTests)
	fmt.Printf("  - Failed Tests: %d\n", results.FailedTests)
	fmt.Printf("  - Success Rate: %.1f%%\n", results.SuccessRate)
	
	if results.RecoveryMetrics != nil {
		fmt.Printf("\n🔄 Recovery Metrics:\n")
		fmt.Printf("  - Average Recovery Time: %v\n", results.RecoveryMetrics.AverageRecoveryTime)
		fmt.Printf("  - Max Recovery Time: %v\n", results.RecoveryMetrics.MaxRecoveryTime)
		fmt.Printf("  - Recovery Success Rate: %.1f%%\n", results.RecoveryMetrics.RecoverySuccessRate)
		fmt.Printf("  - Failed Recoveries: %d\n", results.RecoveryMetrics.FailedRecoveries)
	}
	
	if results.PerformanceImpact != nil {
		fmt.Printf("\n📈 Performance Impact:\n")
		fmt.Printf("  - Throughput Reduction: %.1f%%\n", results.PerformanceImpact.ThroughputReduction)
		fmt.Printf("  - Latency Increase: %v\n", results.PerformanceImpact.LatencyIncrease)
		fmt.Printf("  - Packet Loss Increase: %.1f%%\n", results.PerformanceImpact.PacketLossIncrease)
		fmt.Printf("  - Connection Drops: %d\n", results.PerformanceImpact.ConnectionDrops)
	}
	
	if len(results.ErrorBreakdown) > 0 {
		fmt.Printf("\n🔍 Error Breakdown:\n")
		for errorType, breakdown := range results.ErrorBreakdown {
			fmt.Printf("  - %s: %d (%.1f%%) [%s] [%s]\n",
				errorType,
				breakdown.Count,
				breakdown.Rate,
				breakdown.Severity,
				breakdown.Impact)
		}
	}
	
	fmt.Println()
}

// saveResults сохраняет результаты в файл
func saveResults(results *experimental.ErrorTestingResults, filename string) error {
	// Здесь можно добавить сохранение в JSON формате
	// Для простоты пока просто создаем файл
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Простое сохранение в текстовом формате
	fmt.Fprintf(file, "Error Testing Results\n")
	fmt.Fprintf(file, "=====================\n\n")
	fmt.Fprintf(file, "Duration: %v\n", results.Duration)
	fmt.Fprintf(file, "Total Tests: %d\n", results.TotalTests)
	fmt.Fprintf(file, "Passed Tests: %d\n", results.PassedTests)
	fmt.Fprintf(file, "Failed Tests: %d\n", results.FailedTests)
	fmt.Fprintf(file, "Success Rate: %.1f%%\n", results.SuccessRate)
	
	return nil
}

// generateReport генерирует детальный отчет
func generateReport(results *experimental.ErrorTestingResults) {
	fmt.Printf("Detailed Error Testing Report:\n")
	fmt.Printf("================================\n\n")
	
	// Анализ результатов
	if results.SuccessRate >= 90.0 {
		fmt.Printf("✅ EXCELLENT: Error handling is very robust (%.1f%% success rate)\n", results.SuccessRate)
	} else if results.SuccessRate >= 80.0 {
		fmt.Printf("✅ GOOD: Error handling is robust (%.1f%% success rate)\n", results.SuccessRate)
	} else if results.SuccessRate >= 70.0 {
		fmt.Printf("⚠️  FAIR: Error handling needs improvement (%.1f%% success rate)\n", results.SuccessRate)
	} else {
		fmt.Printf("❌ POOR: Error handling needs significant improvement (%.1f%% success rate)\n", results.SuccessRate)
	}
	
	// Рекомендации
	fmt.Printf("\n💡 Recommendations:\n")
	if results.SuccessRate < 80.0 {
		fmt.Printf("  - Review error handling mechanisms\n")
		fmt.Printf("  - Improve recovery procedures\n")
		fmt.Printf("  - Add more robust error detection\n")
	}
	
	if results.RecoveryMetrics != nil && results.RecoveryMetrics.AverageRecoveryTime > 5*time.Second {
		fmt.Printf("  - Optimize recovery time (current: %v)\n", results.RecoveryMetrics.AverageRecoveryTime)
	}
	
	if results.PerformanceImpact != nil && results.PerformanceImpact.ThroughputReduction > 20.0 {
		fmt.Printf("  - Optimize performance under error conditions (%.1f%% reduction)\n", results.PerformanceImpact.ThroughputReduction)
	}
	
	fmt.Println()
}
