package runner

import (
	"context"
	"io"
	"time"
)

// Transport представляет транспортный протокол
type Transport interface {
	Name() string
	Dial(ctx context.Context, addr string, opts TransportOptions) (Connection, error)
	Listen(ctx context.Context, addr string, opts TransportOptions) (Listener, error)
}

// Connection представляет соединение
type Connection interface {
	OpenStream(ctx context.Context) (io.ReadWriteCloser, error)
	SendDatagram(data []byte) error
	ReceiveDatagram() ([]byte, error)
	Close() error
	Stats() ConnectionStats
}

// Listener представляет слушатель соединений
type Listener interface {
	Accept(ctx context.Context) (Connection, error)
	Close() error
	Addr() string
}

// TransportOptions содержит опции для транспорта
type TransportOptions struct {
	TLSConfig    interface{} // TLS конфигурация
	QUICConfig   interface{} // QUIC конфигурация
	Timeout      time.Duration
	KeepAlive    time.Duration
	MaxStreams   int64
	Enable0RTT   bool
	EnableDatagrams bool
}

// ConnectionStats содержит статистику соединения
type ConnectionStats struct {
	BytesSent     int64
	BytesReceived int64
	PacketsSent   int64
	PacketsLost   int64
	RTT           time.Duration
	Jitter        time.Duration
	StreamsOpen   int64
	StreamsClosed int64
}

// Checker проверяет SLA и инварианты
type Checker interface {
	Name() string
	Check(ctx context.Context, result *RunResult) *CheckResult
}

// CheckResult содержит результат проверки
type CheckResult struct {
	Passed     bool
	Violations []Violation
	Metrics    map[string]interface{}
}

// Violation описывает нарушение SLA
type Violation struct {
	Type      string      `json:"type"`
	Expected  interface{} `json:"expected"`
	Actual    interface{} `json:"actual"`
	Severity  string      `json:"severity"` // "warning", "critical"
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
}

// Reporter генерирует отчеты
type Reporter interface {
	Name() string
	WriteJSON(ctx context.Context, result *RunResult, w io.Writer) error
	WriteCSV(ctx context.Context, result *RunResult, w io.Writer) error
	WriteMarkdown(ctx context.Context, result *RunResult, w io.Writer) error
}

// Scenario представляет тестовый сценарий
type Scenario interface {
	ID() string
	Name() string
	Description() string
	Steps() []ScenarioStep
	Validate() error
}

// ScenarioStep представляет шаг сценария
type ScenarioStep struct {
	Type        string                 `yaml:"type"`        // "handshake", "streams", "datagrams", "key_update", "nat_rebind"
	Duration    time.Duration          `yaml:"duration"`
	Concurrency int                    `yaml:"concurrency"`
	Parameters  map[string]interface{} `yaml:"parameters"`
	Expected    map[string]interface{} `yaml:"expected"`
}

// RunResult содержит результат выполнения теста
type RunResult struct {
	ID          string                 `json:"id"`
	ScenarioID  string                 `json:"scenario_id"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Environment Environment            `json:"environment"`
	Parameters  map[string]interface{} `json:"parameters"`
	Metrics     Metrics                `json:"metrics"`
	Events      []Event                `json:"events"`
	SLA         SLAResult              `json:"sla"`
	Passed      bool                   `json:"passed"`
	ExitCode    int                    `json:"exit_code"`
}

// Environment содержит информацию об окружении
type Environment struct {
	GoVersion    string            `json:"go_version"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	Network      NetworkInfo       `json:"network"`
	Resources    ResourceInfo      `json:"resources"`
	Environment  map[string]string `json:"environment"`
}

// NetworkInfo содержит информацию о сети
type NetworkInfo struct {
	Interface string  `json:"interface"`
	MTU       int     `json:"mtu"`
	RTT       float64 `json:"rtt_ms"`
	Bandwidth float64 `json:"bandwidth_mbps"`
	Loss      float64 `json:"loss_percent"`
}

// ResourceInfo содержит информацию о ресурсах
type ResourceInfo struct {
	CPUs       int     `json:"cpus"`
	MemoryMB   int     `json:"memory_mb"`
	DiskSpace  int64   `json:"disk_space_bytes"`
	LoadAvg    float64 `json:"load_avg"`
}

// Metrics содержит все метрики теста
type Metrics struct {
	Latency    LatencyMetrics    `json:"latency"`
	Throughput ThroughputMetrics `json:"throughput"`
	Network    NetworkMetrics    `json:"network"`
	QUIC       QUICMetrics       `json:"quic"`
	TLS        TLSMetrics        `json:"tls"`
	Errors     ErrorMetrics      `json:"errors"`
}

// LatencyMetrics содержит метрики задержки
type LatencyMetrics struct {
	P50   float64 `json:"p50_ms"`
	P90   float64 `json:"p90_ms"`
	P95   float64 `json:"p95_ms"`
	P99   float64 `json:"p99_ms"`
	P999  float64 `json:"p999_ms"`
	Min   float64 `json:"min_ms"`
	Max   float64 `json:"max_ms"`
	Mean  float64 `json:"mean_ms"`
	Jitter float64 `json:"jitter_ms"`
}

// ThroughputMetrics содержит метрики пропускной способности
type ThroughputMetrics struct {
	Goodput float64 `json:"goodput_mbps"`
	Total   float64 `json:"total_mbps"`
	Min     float64 `json:"min_mbps"`
	Max     float64 `json:"max_mbps"`
	Mean    float64 `json:"mean_mbps"`
}

// NetworkMetrics содержит сетевые метрики
type NetworkMetrics struct {
	PacketsSent     int64   `json:"packets_sent"`
	PacketsReceived int64   `json:"packets_received"`
	PacketsLost     int64   `json:"packets_lost"`
	LossPercent     float64 `json:"loss_percent"`
	Retransmits     int64   `json:"retransmits"`
	OutOfOrder      int64   `json:"out_of_order"`
	Duplicates      int64   `json:"duplicates"`
}

// QUICMetrics содержит QUIC-специфичные метрики
type QUICMetrics struct {
	ConnectionsEstablished int64   `json:"connections_established"`
	StreamsOpened         int64   `json:"streams_opened"`
	StreamsClosed         int64   `json:"streams_closed"`
	DatagramsSent         int64   `json:"datagrams_sent"`
	DatagramsReceived     int64   `json:"datagrams_received"`
	ZeroRTTConnections    int64   `json:"zero_rtt_connections"`
	OneRTTConnections     int64   `json:"one_rtt_connections"`
	KeyUpdates            int64   `json:"key_updates"`
	VersionNegotiations   int64   `json:"version_negotiations"`
	Retries               int64   `json:"retries"`
	HandshakeTime         float64 `json:"handshake_time_ms"`
	CongestionWindow      float64 `json:"congestion_window_bytes"`
}

// TLSMetrics содержит TLS-специфичные метрики
type TLSMetrics struct {
	Version              string `json:"version"`
	CipherSuite          string `json:"cipher_suite"`
	SessionResumptions   int64  `json:"session_resumptions"`
	CertificateValid     bool   `json:"certificate_valid"`
	OCSPStapling         bool   `json:"ocsp_stapling"`
	ALPN                 string `json:"alpn"`
}

// ErrorMetrics содержит метрики ошибок
type ErrorMetrics struct {
	TotalErrors      int64            `json:"total_errors"`
	ErrorTypes       map[string]int64 `json:"error_types"`
	ConnectionErrors int64            `json:"connection_errors"`
	StreamErrors     int64            `json:"stream_errors"`
	TimeoutErrors    int64            `json:"timeout_errors"`
}

// Event представляет событие во время теста
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Severity  string                 `json:"severity"`
}

// SLAResult содержит результат проверки SLA
type SLAResult struct {
	Passed     bool        `json:"passed"`
	Violations []Violation `json:"violations"`
	ExitCode   int         `json:"exit_code"`
}
