package metrics

import (
	"quic-test/internal/congestion"
	"time"
)

// CCIntegration интегрирует Prometheus метрики с congestion control
type CCIntegration struct {
	metrics *PrometheusMetrics
	sc      *congestion.SendController
}

// NewCCIntegration создает новую интеграцию
func NewCCIntegration(metrics *PrometheusMetrics, sc *congestion.SendController) *CCIntegration {
	return &CCIntegration{
		metrics: metrics,
		sc:      sc,
	}
}

// UpdateMetrics обновляет все метрики congestion control
func (cci *CCIntegration) UpdateMetrics() {
	// Получаем текущие значения
	cwnd := cci.sc.GetCWND()
	pacingRate := cci.sc.GetPacingRate()
	bandwidth := cci.sc.GetBandwidth()
	minRTT := cci.sc.GetMinRTT()
	state := cci.sc.GetState()
	
	// Обновляем метрики
	cci.metrics.UpdateCCMetrics(
		bandwidth,
		cwnd,
		float64(minRTT.Nanoseconds())/1e6, // конвертируем в миллисекунды
		int(state),
		pacingRate,
	)
}

// StartMetricsCollection запускает периодический сбор метрик
func (cci *CCIntegration) StartMetricsCollection(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for range ticker.C {
			cci.UpdateMetrics()
		}
	}()
}

