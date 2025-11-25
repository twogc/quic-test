package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// MetricData представляет одну запись метрики
type MetricData struct {
	LatencyMs float64 `json:"latency_ms"`
	Code      int     `json:"code"`
	CPU       float64 `json:"cpu"`
	RTTMs     float64 `json:"rtt_ms"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// TUIDashboard представляет TUI дашборд
type TUIDashboard struct {
	// Данные для графиков
	latencyData    []float64
	cpuData        []float64
	rttData        []float64
	errorData      []float64
	
	// Статистика
	totalRequests  int
	errorCount     int
	successCount   int
	
	// Настройки
	width          int
	height         int
	maxDataPoints  int
	fps            int
	
	// Live writer для обновлений без мерцаний
	lastUpdate     time.Time
}

// NewTUIDashboard создает новый TUI дашборд
func NewTUIDashboard() *TUIDashboard {
	d := &TUIDashboard{
		width:         120,
		height:        30,
		maxDataPoints: 360, // 120 * 3
		fps:           5,
		lastUpdate:    time.Now(),
	}
	
	// Получаем реальный размер терминала
	d.updateTermSize()
	
	return d
}

// updateTermSize обновляет размер терминала
func (d *TUIDashboard) updateTermSize() {
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 && h > 0 {
		d.width, d.height = w, h
		d.maxDataPoints = d.width * 3
	}
}

// AddMetric добавляет новую метрику
func (d *TUIDashboard) AddMetric(metric MetricData) {
	d.totalRequests++
	
	// Добавляем данные в соответствующие массивы
	d.latencyData = append(d.latencyData, metric.LatencyMs)
	d.cpuData = append(d.cpuData, metric.CPU)
	d.rttData = append(d.rttData, metric.RTTMs)
	
	// Обрабатываем ошибки
	if metric.Code >= 400 {
		d.errorCount++
		d.errorData = append(d.errorData, 1.0)
	} else {
		d.successCount++
		d.errorData = append(d.errorData, 0.0)
	}
	
	// Ограничиваем размер данных
	d.trimData()
}

// trimData обрезает данные до максимального размера
func (d *TUIDashboard) trimData() {
	if len(d.latencyData) > d.maxDataPoints {
		d.latencyData = d.latencyData[len(d.latencyData)-d.maxDataPoints:]
		d.cpuData = d.cpuData[len(d.cpuData)-d.maxDataPoints:]
		d.rttData = d.rttData[len(d.rttData)-d.maxDataPoints:]
		d.errorData = d.errorData[len(d.errorData)-d.maxDataPoints:]
	}
}

// Render отображает дашборд
func (d *TUIDashboard) Render() {
	// Очищаем экран и перемещаем курсор в начало
	fmt.Print("\x1b[2J\x1b[H")
	
	// Заголовок
	fmt.Println("\033[1;36m==============================\033[0m")
	fmt.Println("\033[1;36m  2GC CloudBridge QUIC TUI   \033[0m")
	fmt.Println("\033[1;36m==============================\033[0m")
	
	// Статистика
	d.renderStats()
	
	// Графики
	d.renderGraphs()
	
	// Легенда
	d.renderLegend()
}

// renderStats отображает статистику
func (d *TUIDashboard) renderStats() {
	successRate := 0.0
	if d.totalRequests > 0 {
		successRate = float64(d.successCount) / float64(d.totalRequests) * 100
	}
	
	fmt.Printf("\033[1;33mСтатистика:\033[0m Запросы: %d | Успех: %.1f%% | Ошибки: %d\n", 
		d.totalRequests, successRate, d.errorCount)
}

// renderGraphs отображает графики
func (d *TUIDashboard) renderGraphs() {
	if len(d.latencyData) == 0 {
		fmt.Println("\033[1;31mНет данных для отображения\033[0m")
		return
	}
	
	// График латенсий
	fmt.Println("\n\033[1;32mЛатенсия (мс):\033[0m")
	d.plotSimpleGraph(d.latencyData, "P50/P95/P99", "\033[31m")
	
	// График CPU
	fmt.Println("\n\033[1;33mCPU (%):\033[0m")
	d.plotSimpleGraph(d.cpuData, "CPU Usage", "\033[33m")
	
	// График RTT
	fmt.Println("\n\033[1;34mRTT (мс):\033[0m")
	d.plotSimpleGraph(d.rttData, "Round Trip Time", "\033[34m")
}

// renderLegend отображает легенду
func (d *TUIDashboard) renderLegend() {
	fmt.Println("\n\033[1;37mЛегенда:\033[0m")
	fmt.Println("  \033[31m●\033[0m Латенсия  \033[33m●\033[0m CPU  \033[34m●\033[0m RTT")
	fmt.Println("\n\033[1;37mУправление:\033[0m Ctrl+C для выхода")
}

// plotSimpleGraph отображает простой график
func (d *TUIDashboard) plotSimpleGraph(data []float64, caption, color string) {
	if len(data) == 0 {
		fmt.Println("  Нет данных")
		return
	}
	
	// Показываем последние N точек, а не первые
	width := d.width - 20
	if width > len(data) {
		width = len(data)
	}
	
	start := len(data) - width
	if start < 0 {
		start = 0
	}
	view := data[start:]
	
	// Находим min и max значения
	min, max := view[0], view[0]
	for _, v := range view {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	// Guard на max==min (деление на ноль)
	den := max - min
	if den <= 0 {
		den = 1 // плоская линия
	}
	
	// Простой ASCII график
	height := 8
	
	fmt.Printf("  %s\n", caption)
	fmt.Printf("  %s\n", strings.Repeat("─", width))
	
	// Создаем график
	for h := height; h >= 0; h-- {
		fmt.Print("  ")
		for w := 0; w < width; w++ {
			if w < len(view) {
				value := view[w]
				normalized := (value - min) / den
				barHeight := int(normalized * float64(height))
				
				if h <= barHeight {
					fmt.Print(color + "█" + "\033[0m")
				} else {
					fmt.Print(" ")
				}
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	
	// Показываем статистику
	p50, p95, p99 := d.calculatePercentiles(view)
	fmt.Printf("  Min: %.1f | Max: %.1f | P50: %.1f | P95: %.1f | P99: %.1f\n", 
		min, max, p50, p95, p99)
}

// calculatePercentiles вычисляет перцентили (быстрая версия без O(n²))
func (d *TUIDashboard) calculatePercentiles(data []float64) (p50, p95, p99 float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}
	
	// Создаем копию для сортировки
	s := make([]float64, len(data))
	copy(s, data)
	sort.Float64s(s)
	
	// Быстрая функция для вычисления квантилей
	q := func(alpha float64) float64 {
		if len(s) == 1 {
			return s[0]
		}
		idx := alpha * float64(len(s)-1)
		i := int(idx)
		f := idx - float64(i)
		if i+1 < len(s) {
			return s[i]*(1-f) + s[i+1]*f
		}
		return s[i]
	}
	
	return q(0.50), q(0.95), q(0.99)
}

// generateDemoData генерирует демо данные (бесконечный цикл с выходом по сигналу)
func (d *TUIDashboard) generateDemoData(stop <-chan struct{}) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for {
		select {
		case <-stop:
			return
		default:
			// Генерируем реалистичные данные с синусоидальными колебаниями
			latency := 10 + r.Float64()*50 + math.Sin(float64(time.Now().UnixNano())/8e9)*10
			cpu := 20 + r.Float64()*40 + math.Sin(float64(time.Now().UnixNano())/6e9)*15
			rtt := latency + r.Float64()*5
			
			// Иногда добавляем ошибки
			code := 200
			if r.Float64() < 0.05 { // 5% ошибок
				code = 500
			}
			
			metric := MetricData{
				LatencyMs: latency,
				Code:      code,
				CPU:       cpu,
				RTTMs:     rtt,
				Timestamp: time.Now(),
			}
			
			d.AddMetric(metric)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// processInput обрабатывает входные данные (БЕЗ вызова Render!)
func (d *TUIDashboard) processInput(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		var metric MetricData
		if err := json.Unmarshal([]byte(line), &metric); err != nil {
			log.Printf("Ошибка парсинга JSON: %v", err)
			continue
		}
		
		d.AddMetric(metric)
		// НЕ вызываем d.Render() здесь - только по тикеру!
	}
}

// isUnix проверяет, является ли текущая платформа Unix-подобной
func isUnix() bool {
	return runtime.GOOS != "windows"
}

func main() {
	var (
		demo = flag.Bool("demo", false, "Запустить демо режим")
		fps  = flag.Int("fps", 5, "FPS для обновления экрана")
	)
	flag.Parse()
	
	// Прячем курсор
	fmt.Print("\x1b[?25l")
	defer fmt.Print("\x1b[?25h")
	
	dashboard := NewTUIDashboard()
	dashboard.fps = *fps
	
	// Обработка сигналов для корректного завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Обработка ресайза терминала (только для Unix-систем)
	var resizeChan chan os.Signal
	if isUnix() {
		resizeChan = make(chan os.Signal, 1)
		signal.Notify(resizeChan, syscall.SIGWINCH)
	} else {
		// На Windows создаем пустой канал, который никогда не получит сигнал
		resizeChan = make(chan os.Signal, 1)
	}
	
	// Канал для остановки демо
	stopChan := make(chan struct{})
	
	if *demo {
		fmt.Println("Запуск демо режима...")
		go dashboard.generateDemoData(stopChan)
	} else {
		// Читаем данные из stdin с увеличенным буфером
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 0, 64*1024), 1<<20) // до 1 МБ
		go dashboard.processInput(scanner)
	}
	
	// Основной цикл обновления
	ticker := time.NewTicker(time.Second / time.Duration(dashboard.fps))
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			dashboard.Render()
			
		case <-resizeChan:
			// Обновляем размер терминала
			dashboard.updateTermSize()
			
		case <-sigChan:
			close(stopChan) // Останавливаем демо
			fmt.Println("\nЗавершение работы...")
			return
		}
	}
}