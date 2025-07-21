package internal

// ServerMetrics хранит метрики сервера.
type ServerMetrics struct {
	TotalConnections int
	TotalStreams     int
	TotalBytes       int64
	TotalErrors      int
}

// ClientMetrics хранит метрики клиента.
type ClientMetrics struct {
	LatencyMs      []float64
	ThroughputBps  []float64
	Errors         int
	Success        int
	AvgPacketSize  float64
} 