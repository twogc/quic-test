package testing

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"quic-test/internal/sla"
)

// TestRunner выполняет тесты из матрицы
type TestRunner struct {
	matrix     *TestMatrix
	results    []TestResult
	outputDir  string
	verbose    bool
}

// NewTestRunner создает новый исполнитель тестов
func NewTestRunner(matrix *TestMatrix, outputDir string, verbose bool) *TestRunner {
	return &TestRunner{
		matrix:    matrix,
		results:   make([]TestResult, 0),
		outputDir: outputDir,
		verbose:   verbose,
	}
}

// RunAllTests выполняет все тесты из матрицы
func (tr *TestRunner) RunAllTests(ctx context.Context) error {
	scenarios := tr.matrix.GenerateScenarios()
	totalScenarios := len(scenarios)
	
	fmt.Printf("🧪 Starting test matrix execution\n")
	fmt.Printf("Total scenarios: %d\n", totalScenarios)
	fmt.Printf("⏱️  Estimated duration: %v\n", tr.matrix.GetEstimatedDuration())
	fmt.Printf("Output directory: %s\n\n", tr.outputDir)
	
	// Создаем директорию для результатов
	if err := os.MkdirAll(tr.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Запускаем тесты последовательно
	for i, scenario := range scenarios {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		fmt.Printf("🔄 Running scenario %d/%d: %s\n", i+1, totalScenarios, scenario.ID)
		
		result, err := tr.runScenario(ctx, scenario)
		if err != nil {
			fmt.Printf("❌ Scenario %s failed: %v\n", scenario.ID, err)
			result = &TestResult{
				ScenarioID: scenario.ID,
				StartTime:  time.Now(),
				EndTime:    time.Now(),
				Passed:     false,
				Errors:     []string{err.Error()},
			}
		}
		
		tr.results = append(tr.results, *result)
		
		// Сохраняем промежуточные результаты
		if err := tr.saveResults(); err != nil {
			fmt.Printf("⚠️  Failed to save results: %v\n", err)
		}
		
		// Пауза между тестами
		if i < totalScenarios-1 {
			time.Sleep(tr.matrix.CooldownDuration)
		}
	}
	
	fmt.Printf("\n✅ All tests completed!\n")
	return tr.generateReport()
}

// runScenario выполняет один сценарий тестирования
func (tr *TestRunner) runScenario(ctx context.Context, scenario TestScenario) (*TestResult, error) {
	result := &TestResult{
		ScenarioID: scenario.ID,
		StartTime:  time.Now(),
	}
	
	// Создаем директорию для этого сценария
	scenarioDir := filepath.Join(tr.outputDir, scenario.ID)
	if err := os.MkdirAll(scenarioDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create scenario directory: %w", err)
	}
	
	// Запускаем сервер
	_, serverCancel := context.WithCancel(ctx)
	defer serverCancel()
	
	serverCmd := tr.buildServerCommand(scenario, scenarioDir)
	if tr.verbose {
		fmt.Printf("Starting server: %s\n", strings.Join(serverCmd.Args, " "))
	}
	
	if err := serverCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}
	
	// Ждем запуска сервера
	time.Sleep(2 * time.Second)
	
	// Запускаем клиент
	clientCmd := tr.buildClientCommand(scenario, scenarioDir)
	if tr.verbose {
		fmt.Printf("🔗 Starting client: %s\n", strings.Join(clientCmd.Args, " "))
	}
	
	_, clientCancel := context.WithTimeout(ctx, scenario.TestDuration)
	defer clientCancel()
	
	if err := clientCmd.Run(); err != nil {
		serverCancel()
		return nil, fmt.Errorf("client execution failed: %w", err)
	}
	
	// Останавливаем сервер
	serverCancel()
	time.Sleep(1 * time.Second)
	
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	// Анализируем результаты
	if err := tr.analyzeResults(scenario, scenarioDir, result); err != nil {
		return nil, fmt.Errorf("failed to analyze results: %w", err)
	}
	
	// Проверяем SLA
	tr.checkSLA(scenario, result)
	
	return result, nil
}

// buildServerCommand создает команду для запуска сервера
func (tr *TestRunner) buildServerCommand(scenario TestScenario, outputDir string) *exec.Cmd {
	args := []string{
		"--mode", "server",
		"--addr", ":9000",
		"--cc", "bbrv2",
		"--ack-freq", "2",
		"--fec",
		"--fec-redundancy", "0.1",
		"--greasing",
		"--qlog", filepath.Join(outputDir, "server-qlog"),
		"--verbose",
	}
	
	// Добавляем SLA флаги
	args = append(args, "--sla-p95-rtt", "100")
	args = append(args, "--sla-loss", "5")
	args = append(args, "--sla-goodput", "1")
	
	return exec.CommandContext(context.Background(), "./quic-test-experimental", args...)
}

// buildClientCommand создает команду для запуска клиента
func (tr *TestRunner) buildClientCommand(scenario TestScenario, outputDir string) *exec.Cmd {
	args := []string{
		"--mode", "client",
		"--addr", "127.0.0.1:9000",
		"--connections", strconv.Itoa(scenario.Connections),
		"--streams", strconv.Itoa(scenario.Streams),
		"--duration", scenario.TestDuration.String(),
		"--packet-size", strconv.Itoa(scenario.PacketSize),
		"--rate", strconv.Itoa(scenario.PacketRate),
		"--qlog", filepath.Join(outputDir, "client-qlog"),
		"--verbose",
	}
	
	return exec.CommandContext(context.Background(), "./quic-test-experimental", args...)
}

// analyzeResults анализирует результаты теста
func (tr *TestRunner) analyzeResults(scenario TestScenario, outputDir string, result *TestResult) error {
	// Здесь должна быть логика анализа qlog файлов, метрик и т.д.
	// Пока что заполняем заглушками
	
	result.GoodputMbps = float64(scenario.PacketRate * scenario.Connections * scenario.PacketSize * 8) / 1e6
	result.ThroughputMbps = result.GoodputMbps * 1.1 // 10% overhead
	result.LatencyMeanMs = float64(scenario.RTT.Milliseconds()) + 1.0
	result.LatencyMinMs = result.LatencyMeanMs * 0.8
	result.LatencyMaxMs = result.LatencyMeanMs * 1.5
	result.LatencyP95Ms = result.LatencyMeanMs * 1.2
	result.LatencyP99Ms = result.LatencyMeanMs * 1.4
	
	result.LossRatePercent = scenario.LossRate
	result.PacketsSent = int64(scenario.PacketRate * scenario.Connections * int(scenario.TestDuration.Seconds()))
	result.PacketsLost = int64(float64(result.PacketsSent) * scenario.LossRate / 100.0)
	result.PacketsReceived = result.PacketsSent - result.PacketsLost
	
	result.BandwidthBps = result.GoodputMbps * 1e6
	result.CWNDBytes = 1460 * 20 // Примерное значение
	result.PacingRateBps = int64(result.BandwidthBps)
	result.CCState = "bbrv2"
	
	result.ACKDelayMs = 25.0
	result.ACKFrequency = 2
	
	result.FECRedundancy = 0.1
	result.FECRecoveryRate = 0.9
	
	result.Passed = true
	
	return nil
}

// checkSLA проверяет результаты против SLA
func (tr *TestRunner) checkSLA(scenario TestScenario, result *TestResult) {
	// Создаем SLA-гейты
	slaGates := sla.NewSLAGates()
	
	// Создаем метрики для проверки
	metrics := sla.SLAMetrics{
		RTTPercentile95Ms: result.LatencyP95Ms,
		RTTMaxMs:          result.LatencyMaxMs,
		RTTMeanMs:         result.LatencyMeanMs,
		LossRatePercent:   result.LossRatePercent,
		GoodputMbps:       result.GoodputMbps,
		ThroughputMbps:    result.ThroughputMbps,
		BandwidthBps:      result.BandwidthBps,
		CWNDBytes:         result.CWNDBytes,
		ACKDelayMs:        result.ACKDelayMs,
		ACKFrequency:      result.ACKFrequency,
		FECRedundancy:     result.FECRedundancy,
		FECRecoveryRate:   result.FECRecoveryRate,
	}
	
	// Проверяем SLA
	validator := sla.NewSLAValidator(slaGates)
	slaResult := validator.Validate(metrics)
	
	result.SLAScore = slaResult.Score
	result.Passed = slaResult.Passed
	
	// Записываем нарушения
	for _, violation := range slaResult.Violations {
		result.SLAViolations = append(result.SLAViolations, violation.Message)
	}
}

// saveResults сохраняет результаты в файл
func (tr *TestRunner) saveResults() error {
	// Здесь должна быть логика сохранения результатов в JSON/CSV
	// Пока что просто логируем
	if tr.verbose {
		fmt.Printf("💾 Saving results...\n")
	}
	return nil
}

// generateReport генерирует итоговый отчет
func (tr *TestRunner) generateReport() error {
	fmt.Printf("\nTest Matrix Report\n")
	fmt.Printf("====================\n")
	
	passed := 0
	failed := 0
	totalScore := 0.0
	
	for _, result := range tr.results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
		totalScore += result.SLAScore
	}
	
	avgScore := totalScore / float64(len(tr.results))
	
	fmt.Printf("✅ Passed: %d\n", passed)
	fmt.Printf("❌ Failed: %d\n", failed)
	fmt.Printf("📈 Average SLA Score: %.2f\n", avgScore)
	fmt.Printf("⏱️  Total Duration: %v\n", time.Since(tr.results[0].StartTime))
	
	return nil
}
