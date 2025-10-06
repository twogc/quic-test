package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestPrometheusMetricsBasic(t *testing.T) {
	// Создаем отдельный registry для тестов
	registry := prometheus.NewRegistry()
	
	// Создаем простые метрики для тестирования
	connectionsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_quic_connections_total",
		Help: "Total number of QUIC connections established",
	})
	
	latencyHistogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_quic_latency_seconds",
		Help:    "QUIC latency distribution",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	})
	
	// Регистрируем метрики
	registry.MustRegister(connectionsTotal)
	registry.MustRegister(latencyHistogram)
	
	// Тестируем метрики
	connectionsTotal.Inc()
	latencyHistogram.Observe(0.05) // 50ms
	
	// Проверяем, что метрики работают
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}
	
	if len(metrics) == 0 {
		t.Error("No metrics collected")
	}
}

func TestPrometheusMetricsWithLabels(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	// Создаем векторные метрики
	scenarioCounters := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_quic_scenario_total",
		Help: "Total operations by scenario",
	}, []string{"scenario", "connection_id", "result"})
	
	// Регистрируем метрики
	registry.MustRegister(scenarioCounters)
	
	// Тестируем векторные метрики
	scenarioCounters.WithLabelValues("latency", "conn1", "success").Inc()
	scenarioCounters.WithLabelValues("throughput", "conn2", "error").Inc()
	
	// Проверяем, что метрики работают
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}
	
	if len(metrics) == 0 {
		t.Error("No metrics collected")
	}
}

func TestPrometheusMetricsHistogramBuckets(t *testing.T) {
	registry := prometheus.NewRegistry()
	
	// Создаем гистограмму с бакетами для QUIC
	latencyHistogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_quic_latency_seconds",
		Help:    "QUIC latency distribution",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	})
	
	registry.MustRegister(latencyHistogram)
	
	// Тестируем различные значения задержки
	testLatencies := []time.Duration{
		1 * time.Millisecond,   // 0.001s
		5 * time.Millisecond,   // 0.005s
		10 * time.Millisecond,  // 0.01s
		25 * time.Millisecond,  // 0.025s
		50 * time.Millisecond,  // 0.05s
		100 * time.Millisecond, // 0.1s
		250 * time.Millisecond, // 0.25s
		500 * time.Millisecond, // 0.5s
		1 * time.Second,        // 1.0s
		2 * time.Second,        // 2.0s
		5 * time.Second,        // 5.0s
		10 * time.Second,       // 10.0s
	}
	
	for _, latency := range testLatencies {
		latencyHistogram.Observe(latency.Seconds())
	}
	
	// Проверяем, что метрики работают
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}
	
	if len(metrics) == 0 {
		t.Error("No metrics collected")
	}
}
