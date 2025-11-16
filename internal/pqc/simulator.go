package pqc

import (
	"crypto/rand"
	"sync"
	"time"
)

// PQCSimulator симулирует Post-Quantum Cryptography overhead
// В реальной реализации PQC увеличивает размер handshake и время обработки
type PQCSimulator struct {
	algorithm     string // "ml-kem-512", "ml-kem-768", "dilithium-2", etc.
	handshakeSize int    // Размер handshake в байтах (эмулированный)
	handshakeTime time.Duration // Время handshake (эмулированное)
	mu            sync.RWMutex
	metrics       *PQCMetrics
}

// PQCMetrics метрики PQC
type PQCMetrics struct {
	HandshakesCompleted int64   `json:"handshakes_completed"`
	TotalHandshakeSize  int64   `json:"total_handshake_size"`
	AvgHandshakeTime    float64 `json:"avg_handshake_time_ms"`
	MaxHandshakeTime    float64 `json:"max_handshake_time_ms"`
}

// NewPQCSimulator создает новый PQC симулятор
func NewPQCSimulator(algorithm string) *PQCSimulator {
	// Эмулируем размеры handshake для разных PQC алгоритмов
	// Реальные размеры: ML-KEM-512 ~800 bytes, ML-KEM-768 ~1184 bytes, Dilithium-2 ~1312 bytes
	handshakeSizes := map[string]int{
		"ml-kem-512":   800,
		"ml-kem-768":   1184,
		"dilithium-2":  1312,
		"hybrid":       2000, // Hybrid: ECDHE + ML-KEM
		"baseline":     200,  // Baseline ECDHE
	}
	
	size := handshakeSizes[algorithm]
	if size == 0 {
		size = handshakeSizes["ml-kem-768"] // Default
	}
	
	return &PQCSimulator{
		algorithm:     algorithm,
		handshakeSize: size,
		metrics:       &PQCMetrics{},
	}
}

// SimulateHandshake симулирует PQC handshake
// Возвращает эмулированное время handshake и размер
func (p *PQCSimulator) SimulateHandshake() (time.Duration, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Эмулируем дополнительное время обработки PQC
	// Реальное время: +5-15ms для ML-KEM, +10-30ms для Dilithium
	baseTime := 10 * time.Millisecond // Base TLS handshake
	pqcOverhead := 5 * time.Millisecond // PQC overhead (simplified)
	
	if p.algorithm == "dilithium-2" {
		pqcOverhead = 15 * time.Millisecond
	} else if p.algorithm == "hybrid" {
		pqcOverhead = 10 * time.Millisecond
	}
	
	// Добавляем небольшую вариацию
	variation := time.Duration(float64(pqcOverhead) * 0.2 * (randFloat64() - 0.5))
	totalTime := baseTime + pqcOverhead + variation
	
	// Обновляем метрики
	p.metrics.HandshakesCompleted++
	p.metrics.TotalHandshakeSize += int64(p.handshakeSize)
	
	avgTime := float64(totalTime.Nanoseconds()) / 1e6
	if avgTime > p.metrics.MaxHandshakeTime {
		p.metrics.MaxHandshakeTime = avgTime
	}
	
	// Обновляем среднее время
	if p.metrics.HandshakesCompleted > 0 {
		totalTimeMs := p.metrics.AvgHandshakeTime * float64(p.metrics.HandshakesCompleted-1)
		p.metrics.AvgHandshakeTime = (totalTimeMs + avgTime) / float64(p.metrics.HandshakesCompleted)
	}
	
	return totalTime, p.handshakeSize
}

// GetMetrics возвращает метрики PQC
func (p *PQCSimulator) GetMetrics() *PQCMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	metrics := *p.metrics
	return &metrics
}

// GetAlgorithm возвращает используемый алгоритм
func (p *PQCSimulator) GetAlgorithm() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.algorithm
}

// GetHandshakeSize возвращает размер handshake
func (p *PQCSimulator) GetHandshakeSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.handshakeSize
}

// randFloat64 возвращает случайное число 0.0-1.0
func randFloat64() float64 {
	b := make([]byte, 8)
	rand.Read(b)
	var val uint64
	for i := 0; i < 8; i++ {
		val |= uint64(b[i]) << (8 * i)
	}
	return float64(val) / float64(^uint64(0))
}

// CompareWithBaseline сравнивает PQC с baseline (ECDHE)
func (p *PQCSimulator) CompareWithBaseline() map[string]interface{} {
	baselineSize := 200 // ECDHE handshake size
	baselineTime := 10.0 // ms
	
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	sizeIncrease := ((float64(p.handshakeSize) - float64(baselineSize)) / float64(baselineSize)) * 100
	timeIncrease := ((p.metrics.AvgHandshakeTime - baselineTime) / baselineTime) * 100
	
	return map[string]interface{}{
		"algorithm":        p.algorithm,
		"handshake_size":    p.handshakeSize,
		"size_increase_pct": sizeIncrease,
		"avg_time_ms":       p.metrics.AvgHandshakeTime,
		"time_increase_pct": timeIncrease,
		"baseline_size":     baselineSize,
		"baseline_time_ms":  baselineTime,
	}
}

