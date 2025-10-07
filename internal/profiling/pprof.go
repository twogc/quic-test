package profiling

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // Регистрирует HTTP обработчики для pprof
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

// Profiler управляет профилированием приложения
type Profiler struct {
	server   *http.Server
	traceFile *os.File
	enabled  bool
}

// ProfilerConfig содержит конфигурацию профилировщика
type ProfilerConfig struct {
	Addr        string        // Адрес для pprof HTTP сервера
	TraceFile   string        // Файл для записи trace
	TraceDuration time.Duration // Длительность trace
	Enabled     bool          // Включен ли профилировщик
}

// NewProfiler создает новый профилировщик
func NewProfiler(cfg ProfilerConfig) *Profiler {
	return &Profiler{
		enabled: cfg.Enabled,
	}
}

// Start запускает профилировщик
func (p *Profiler) Start(ctx context.Context, cfg ProfilerConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Настраиваем runtime профилирование
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	// Запускаем HTTP сервер для pprof
	if cfg.Addr != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
		
		p.server = &http.Server{
			Addr:    cfg.Addr,
			Handler: mux,
		}

		go func() {
			log.Printf("Starting pprof server on %s", cfg.Addr)
			if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("pprof server error: %v", err)
			}
		}()
	}

	// Запускаем trace если указан файл
	if cfg.TraceFile != "" {
		if err := p.startTrace(cfg.TraceFile, cfg.TraceDuration); err != nil {
			return fmt.Errorf("failed to start trace: %w", err)
		}
	}

	return nil
}

// startTrace запускает запись trace
func (p *Profiler) startTrace(filename string, duration time.Duration) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create trace file: %w", err)
	}

	p.traceFile = file

	if err := trace.Start(file); err != nil {
		file.Close()
		return fmt.Errorf("failed to start trace: %w", err)
	}

	log.Printf("Started trace recording to %s", filename)

	// Останавливаем trace через указанное время
	if duration > 0 {
		go func() {
			time.Sleep(duration)
			p.StopTrace()
		}()
	}

	return nil
}

// StopTrace останавливает запись trace
func (p *Profiler) StopTrace() {
	if p.traceFile != nil {
		trace.Stop()
		p.traceFile.Close()
		p.traceFile = nil
		log.Println("Stopped trace recording")
	}
}

// Stop останавливает профилировщик
func (p *Profiler) Stop(ctx context.Context) error {
	if !p.enabled {
		return nil
	}

	// Останавливаем trace
	p.StopTrace()

	// Останавливаем HTTP сервер
	if p.server != nil {
		if err := p.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown pprof server: %w", err)
		}
	}

	return nil
}

// WriteHeapProfile записывает heap профиль в файл
func (p *Profiler) WriteHeapProfile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create heap profile file: %w", err)
	}
	defer file.Close()

	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}

	log.Printf("Heap profile written to %s", filename)
	return nil
}

// WriteCPUProfile записывает CPU профиль в файл
func (p *Profiler) WriteCPUProfile(filename string, duration time.Duration) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}
	defer file.Close()

	if err := pprof.StartCPUProfile(file); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	time.Sleep(duration)
	pprof.StopCPUProfile()

	log.Printf("CPU profile written to %s", filename)
	return nil
}

// WriteMutexProfile записывает mutex профиль в файл
func (p *Profiler) WriteMutexProfile(filename string) error {
	// Mutex профилирование доступно только через HTTP endpoint /debug/pprof/mutex
	// или через go tool pprof
	log.Printf("Mutex profile should be collected via HTTP endpoint /debug/pprof/mutex")
	return fmt.Errorf("mutex profile collection not implemented - use HTTP endpoint /debug/pprof/mutex")
}

// WriteBlockProfile записывает block профиль в файл
func (p *Profiler) WriteBlockProfile(filename string) error {
	// Block профилирование доступно только через HTTP endpoint /debug/pprof/block
	// или через go tool pprof
	log.Printf("Block profile should be collected via HTTP endpoint /debug/pprof/block")
	return fmt.Errorf("block profile collection not implemented - use HTTP endpoint /debug/pprof/block")
}

// GetMemStats возвращает статистику памяти
func (p *Profiler) GetMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// ForceGC принудительно запускает сборку мусора
func (p *Profiler) ForceGC() {
	runtime.GC()
}

// SetGCPercent устанавливает процент для GC
func (p *Profiler) SetGCPercent(percent int) int {
	// Функция SetGCPercent недоступна в некоторых версиях Go
	// Возвращаем текущее значение
	log.Printf("SetGCPercent not available in this Go version")
	return 100 // Возвращаем значение по умолчанию
}

// MemStats содержит статистику памяти
type MemStats struct {
	Alloc         uint64 `json:"alloc_bytes"`
	TotalAlloc    uint64 `json:"total_alloc_bytes"`
	Sys           uint64 `json:"sys_bytes"`
	Lookups       uint64 `json:"lookups"`
	Mallocs       uint64 `json:"mallocs"`
	Frees         uint64 `json:"frees"`
	HeapAlloc     uint64 `json:"heap_alloc_bytes"`
	HeapSys       uint64 `json:"heap_sys_bytes"`
	HeapIdle      uint64 `json:"heap_idle_bytes"`
	HeapInuse     uint64 `json:"heap_inuse_bytes"`
	HeapReleased  uint64 `json:"heap_released_bytes"`
	HeapObjects   uint64 `json:"heap_objects"`
	StackInuse    uint64 `json:"stack_inuse_bytes"`
	StackSys      uint64 `json:"stack_sys_bytes"`
	MSpanInuse    uint64 `json:"mspan_inuse_bytes"`
	MSpanSys      uint64 `json:"mspan_sys_bytes"`
	MCacheInuse   uint64 `json:"mcache_inuse_bytes"`
	MCacheSys     uint64 `json:"mcache_sys_bytes"`
	BuckHashSys   uint64 `json:"buck_hash_sys_bytes"`
	GCSys         uint64 `json:"gc_sys_bytes"`
	OtherSys      uint64 `json:"other_sys_bytes"`
	NextGC        uint64 `json:"next_gc_bytes"`
	LastGC        uint64 `json:"last_gc_ns"`
	PauseTotalNs  uint64 `json:"pause_total_ns"`
	NumGC         uint32 `json:"num_gc"`
	NumForcedGC   uint32 `json:"num_forced_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
	EnableGC      bool   `json:"enable_gc"`
	DebugGC       bool   `json:"debug_gc"`
}

// GetMemStatsStruct возвращает структурированную статистику памяти
func (p *Profiler) GetMemStatsStruct() MemStats {
	m := p.GetMemStats()
	return MemStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		EnableGC:      m.EnableGC,
		DebugGC:       m.DebugGC,
	}
}

// RuntimeStats содержит статистику runtime
type RuntimeStats struct {
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCgoCall   int64  `json:"num_cgo_call"`
	MemStats     MemStats `json:"mem_stats"`
}

// GetRuntimeStats возвращает статистику runtime
func (p *Profiler) GetRuntimeStats() RuntimeStats {
	return RuntimeStats{
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCgoCall:   runtime.NumCgoCall(),
		MemStats:     p.GetMemStatsStruct(),
	}
}
