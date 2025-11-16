package scenarios

import (
	"context"
	"fmt"
	"time"
)

// QUICSpecificScenario представляет QUIC-специфичный сценарий
type QUICSpecificScenario struct {
	id          string
	name        string
	description string
	steps       []ScenarioStep
}

// ID возвращает идентификатор сценария
func (s *QUICSpecificScenario) ID() string {
	return s.id
}

// Name возвращает имя сценария
func (s *QUICSpecificScenario) Name() string {
	return s.name
}

// Description возвращает описание сценария
func (s *QUICSpecificScenario) Description() string {
	return s.description
}

// Steps возвращает шаги сценария
func (s *QUICSpecificScenario) Steps() []ScenarioStep {
	return s.steps
}

// Validate проверяет корректность сценария
func (s *QUICSpecificScenario) Validate() error {
	if s.id == "" {
		return fmt.Errorf("scenario ID cannot be empty")
	}
	if s.name == "" {
		return fmt.Errorf("scenario name cannot be empty")
	}
	if len(s.steps) == 0 {
		return fmt.Errorf("scenario must have at least one step")
	}
	return nil
}

// ScenarioStep представляет шаг сценария
type ScenarioStep struct {
	Type        string                 `yaml:"type"`
	Duration    time.Duration          `yaml:"duration"`
	Concurrency int                    `yaml:"concurrency"`
	Parameters  map[string]interface{} `yaml:"parameters"`
	Expected    map[string]interface{} `yaml:"expected"`
}

// QUICSpecificScenarios содержит все QUIC-специфичные сценарии
var QUICSpecificScenarios = map[string]*QUICSpecificScenario{
	"version_negotiation": {
		id:          "version_negotiation",
		name:        "Version Negotiation",
		description: "Тестирование переговоров версии QUIC",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    10 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"versions": []string{"v1", "v2", "draft-29"},
					"force_version_negotiation": true,
				},
				Expected: map[string]interface{}{
					"success_rate": 0.95,
					"max_handshake_time": "5s",
				},
			},
		},
	},
	
	"retry_scenario": {
		id:          "retry_scenario",
		name:        "Retry Scenario",
		description: "Тестирование механизма Retry в QUIC",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    15 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"force_retry": true,
					"retry_delay": "100ms",
				},
				Expected: map[string]interface{}{
					"retry_rate": 0.8,
					"success_rate": 0.9,
				},
			},
		},
	},
	
	"zero_rtt_load": {
		id:          "zero_rtt_load",
		name:        "0-RTT Load Test",
		description: "Нагрузочное тестирование 0-RTT соединений",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_0rtt": true,
					"session_resumption": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "1s",
				},
			},
			{
				Type:        "zero_rtt_data",
				Duration:    30 * time.Second,
				Concurrency: 20,
				Parameters: map[string]interface{}{
					"data_size": 1024,
					"packet_rate": 100,
				},
				Expected: map[string]interface{}{
					"zero_rtt_success_rate": 0.95,
					"max_latency": "50ms",
				},
			},
		},
	},
	
	"key_update_load": {
		id:          "key_update_load",
		name:        "Key Update Load Test",
		description: "Тестирование обновления ключей под нагрузкой",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_key_update": true,
					"key_update_interval": "30s",
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "streams",
				Duration:    60 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"stream_count": 100,
					"data_size": 4096,
					"packet_rate": 200,
				},
				Expected: map[string]interface{}{
					"key_updates": 2,
					"success_rate": 0.98,
				},
			},
		},
	},
	
	"mtu_probe": {
		id:          "mtu_probe",
		name:        "MTU Probe Test",
		description: "Тестирование обнаружения MTU и PMTUD",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_pmtud": true,
					"initial_mtu": 1200,
				},
				Expected: map[string]interface{}{
					"handshake_time": "3s",
				},
			},
			{
				Type:        "mtu_probe",
				Duration:    20 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"probe_sizes": []int{1200, 1400, 1500, 9000},
					"probe_interval": "5s",
				},
				Expected: map[string]interface{}{
					"mtu_discovered": 1500,
					"probe_success_rate": 0.9,
				},
			},
		},
	},
	
	"ecn_test": {
		id:          "ecn_test",
		name:        "ECN Test",
		description: "Тестирование поддержки ECN (Explicit Congestion Notification)",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_ecn": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "congestion_test",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"congestion_control": "cubic",
					"packet_rate": 500,
					"data_size": 1400,
				},
				Expected: map[string]interface{}{
					"ecn_marked_packets": 0.1,
					"congestion_events": 5,
				},
			},
		},
	},
	
	"nat_rebinding": {
		id:          "nat_rebinding",
		name:        "NAT Rebinding Test",
		description: "Тестирование переподключения при изменении NAT",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_nat_rebinding": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "nat_rebind",
				Duration:    20 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"rebind_interval": "10s",
					"rebind_delay": "2s",
				},
				Expected: map[string]interface{}{
					"rebind_success_rate": 0.9,
					"recovery_time": "5s",
				},
			},
		},
	},
	
	"flow_control_limits": {
		id:          "flow_control_limits",
		name:        "Flow Control Limits Test",
		description: "Тестирование лимитов flow control",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"max_stream_data": 1024 * 1024, // 1MB
					"max_connection_data": 10 * 1024 * 1024, // 10MB
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "flow_control_test",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"stream_count": 1000,
					"data_size": 2 * 1024 * 1024, // 2MB per stream
					"packet_rate": 1000,
				},
				Expected: map[string]interface{}{
					"flow_control_events": 10,
					"success_rate": 0.95,
				},
			},
		},
	},
	
	"datagrams_vs_streams": {
		id:          "datagrams_vs_streams",
		name:        "Datagrams vs Streams Test",
		description: "Сравнение производительности datagrams и streams",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_datagrams": true,
					"enable_streams": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "datagrams_test",
				Duration:    20 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"datagram_size": 1200,
					"datagram_rate": 500,
				},
				Expected: map[string]interface{}{
					"datagram_success_rate": 0.98,
					"datagram_latency": "10ms",
				},
			},
			{
				Type:        "streams_test",
				Duration:    20 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"stream_count": 100,
					"stream_data_size": 4096,
					"stream_rate": 200,
				},
				Expected: map[string]interface{}{
					"stream_success_rate": 0.99,
					"stream_latency": "15ms",
				},
			},
		},
	},
	
	"congestion_control_switch": {
		id:          "congestion_control_switch",
		name:        "Congestion Control Switch Test",
		description: "Тестирование переключения алгоритмов управления перегрузкой",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"congestion_control": "cubic",
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "congestion_test_cubic",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"congestion_control": "cubic",
					"packet_rate": 1000,
					"data_size": 1400,
				},
				Expected: map[string]interface{}{
					"throughput": 50.0, // Mbps
					"congestion_events": 5,
				},
			},
			{
				Type:        "congestion_test_bbr",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"congestion_control": "bbr",
					"packet_rate": 1000,
					"data_size": 1400,
				},
				Expected: map[string]interface{}{
					"throughput": 60.0, // Mbps
					"congestion_events": 3,
				},
			},
		},
	},
	
	"connection_migration": {
		id:          "connection_migration",
		name:        "Connection Migration Test",
		description: "Тестирование миграции соединения между адресами",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_connection_migration": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "migration_test",
				Duration:    30 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"migration_interval": "15s",
					"new_addresses": []string{"192.168.1.100", "192.168.1.101"},
				},
				Expected: map[string]interface{}{
					"migration_success_rate": 0.95,
					"migration_time": "2s",
				},
			},
		},
	},
}

// GetQUICSpecificScenario возвращает QUIC-специфичный сценарий по ID
func GetQUICSpecificScenario(id string) (*QUICSpecificScenario, error) {
	scenario, exists := QUICSpecificScenarios[id]
	if !exists {
		return nil, fmt.Errorf("QUIC scenario '%s' not found", id)
	}
	return scenario, nil
}

// ListQUICSpecificScenarios возвращает список всех QUIC-специфичных сценариев
func ListQUICSpecificScenarios() []string {
	scenarios := make([]string, 0, len(QUICSpecificScenarios))
	for id := range QUICSpecificScenarios {
		scenarios = append(scenarios, id)
	}
	return scenarios
}

// ScenarioExecutor выполняет сценарий
type ScenarioExecutor interface {
	ExecuteStep(ctx context.Context, step ScenarioStep) error
	GetMetrics() map[string]interface{}
}

// QUICScenarioExecutor выполняет QUIC-специфичные сценарии
type QUICScenarioExecutor struct {
	transport Transport
	metrics   MetricsCollector
}

// NewQUICScenarioExecutor создает новый исполнитель QUIC сценариев
func NewQUICScenarioExecutor(transport Transport, metrics MetricsCollector) *QUICScenarioExecutor {
	return &QUICScenarioExecutor{
		transport: transport,
		metrics:   metrics,
	}
}

// ExecuteStep выполняет шаг сценария
func (e *QUICScenarioExecutor) ExecuteStep(ctx context.Context, step ScenarioStep) error {
	switch step.Type {
	case "handshake":
		return e.executeHandshakeStep(ctx, step)
	case "zero_rtt_data":
		return e.executeZeroRTTDataStep(ctx, step)
	case "key_update_load":
		return e.executeKeyUpdateStep(ctx, step)
	case "mtu_probe":
		return e.executeMTUProbeStep(ctx, step)
	case "ecn_test":
		return e.executeECNTestStep(ctx, step)
	case "nat_rebind":
		return e.executeNATRebindStep(ctx, step)
	case "flow_control_test":
		return e.executeFlowControlTestStep(ctx, step)
	case "datagrams_test":
		return e.executeDatagramsTestStep(ctx, step)
	case "streams_test":
		return e.executeStreamsTestStep(ctx, step)
	case "congestion_test":
		return e.executeCongestionTestStep(ctx, step)
	case "migration_test":
		return e.executeMigrationTestStep(ctx, step)
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// GetMetrics возвращает собранные метрики
func (e *QUICScenarioExecutor) GetMetrics() map[string]interface{} {
	return e.metrics.GetAllMetrics()
}

// Заглушки для методов выполнения шагов
func (e *QUICScenarioExecutor) executeHandshakeStep(ctx context.Context, step ScenarioStep) error {
	// Реализация handshake шага
	return nil
}

func (e *QUICScenarioExecutor) executeZeroRTTDataStep(ctx context.Context, step ScenarioStep) error {
	// Реализация 0-RTT data шага
	return nil
}

func (e *QUICScenarioExecutor) executeKeyUpdateStep(ctx context.Context, step ScenarioStep) error {
	// Реализация key update шага
	return nil
}

func (e *QUICScenarioExecutor) executeMTUProbeStep(ctx context.Context, step ScenarioStep) error {
	// Реализация MTU probe шага
	return nil
}

func (e *QUICScenarioExecutor) executeECNTestStep(ctx context.Context, step ScenarioStep) error {
	// Реализация ECN test шага
	return nil
}

func (e *QUICScenarioExecutor) executeNATRebindStep(ctx context.Context, step ScenarioStep) error {
	// Реализация NAT rebind шага
	return nil
}

func (e *QUICScenarioExecutor) executeFlowControlTestStep(ctx context.Context, step ScenarioStep) error {
	// Реализация flow control test шага
	return nil
}

func (e *QUICScenarioExecutor) executeDatagramsTestStep(ctx context.Context, step ScenarioStep) error {
	// Реализация datagrams test шага
	return nil
}

func (e *QUICScenarioExecutor) executeStreamsTestStep(ctx context.Context, step ScenarioStep) error {
	// Реализация streams test шага
	return nil
}

func (e *QUICScenarioExecutor) executeCongestionTestStep(ctx context.Context, step ScenarioStep) error {
	// Реализация congestion test шага
	return nil
}

func (e *QUICScenarioExecutor) executeMigrationTestStep(ctx context.Context, step ScenarioStep) error {
	// Реализация migration test шага
	return nil
}

// Интерфейсы для зависимостей
type Transport interface {
	Dial(ctx context.Context, addr string) (Connection, error)
}

type Connection interface {
	OpenStream() (Stream, error)
	SendDatagram(data []byte) error
	Close() error
}

type Stream interface {
	Write(data []byte) (int, error)
	Read(data []byte) (int, error)
	Close() error
}

type MetricsCollector interface {
	GetAllMetrics() map[string]interface{}
}
