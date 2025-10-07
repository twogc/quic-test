package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryManager управляет OpenTelemetry трейсингом и метриками
type TelemetryManager struct {
	tracer    trace.Tracer
	meter     metric.Meter
	shutdown  func(context.Context) error
}

// TelemetryConfig содержит конфигурацию для телеметрии
type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	PrometheusAddr string
	SampleRate     float64
}

// NewTelemetryManager создает новый менеджер телеметрии
func NewTelemetryManager(ctx context.Context, cfg TelemetryConfig) (*TelemetryManager, error) {
	// Создаем ресурс
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Настраиваем трейсинг
	var tp trace.TracerProvider
	if cfg.OTLPEndpoint != "" {
		// OTLP экспортер
		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}

		tp = trace.NewTracerProvider(
			trace.WithBatcher(exporter),
			trace.WithResource(res),
			trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRate)),
		)
	} else {
		// Локальный провайдер для разработки
		tp = trace.NewTracerProvider(
			trace.WithResource(res),
			trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRate)),
		)
	}

	// Настраиваем метрики
	var mp metric.MeterProvider
	if cfg.PrometheusAddr != "" {
		// Prometheus экспортер
		exporter, err := prometheus.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
		}

		mp = metric.NewMeterProvider(
			metric.WithReader(exporter),
			metric.WithResource(res),
		)
	} else {
		// Локальный провайдер
		mp = metric.NewMeterProvider(
			metric.WithResource(res),
		)
	}

	// Устанавливаем глобальные провайдеры
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Создаем трейсер и метр
	tracer := tp.Tracer(cfg.ServiceName)
	meter := mp.Meter(cfg.ServiceName)

	// Функция для graceful shutdown
	shutdown := func(ctx context.Context) error {
		var errs []error
		
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
		
		if err := mp.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown meter provider: %w", err))
		}
		
		if len(errs) > 0 {
			return fmt.Errorf("shutdown errors: %v", errs)
		}
		
		return nil
	}

	return &TelemetryManager{
		tracer:   tracer,
		meter:    meter,
		shutdown: shutdown,
	}, nil
}

// StartSpan создает новый span
func (tm *TelemetryManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tm.tracer.Start(ctx, name, opts...)
}

// CreateInt64Counter создает счетчик
func (tm *TelemetryManager) CreateInt64Counter(name, description string) (instrument.Int64Counter, error) {
	return tm.meter.Int64Counter(name, instrument.WithDescription(description))
}

// CreateFloat64Counter создает счетчик с плавающей точкой
func (tm *TelemetryManager) CreateFloat64Counter(name, description string) (instrument.Float64Counter, error) {
	return tm.meter.Float64Counter(name, instrument.WithDescription(description))
}

// CreateInt64Histogram создает гистограмму
func (tm *TelemetryManager) CreateInt64Histogram(name, description string) (instrument.Int64Histogram, error) {
	return tm.meter.Int64Histogram(name, instrument.WithDescription(description))
}

// CreateFloat64Histogram создает гистограмму с плавающей точкой
func (tm *TelemetryManager) CreateFloat64Histogram(name, description string) (instrument.Float64Histogram, error) {
	return tm.meter.Float64Histogram(name, instrument.WithDescription(description))
}

// CreateInt64Gauge создает gauge
func (tm *TelemetryManager) CreateInt64Gauge(name, description string) (instrument.Int64Gauge, error) {
	return tm.meter.Int64Gauge(name, instrument.WithDescription(description))
}

// CreateFloat64Gauge создает gauge с плавающей точкой
func (tm *TelemetryManager) CreateFloat64Gauge(name, description string) (instrument.Float64Gauge, error) {
	return tm.meter.Float64Gauge(name, instrument.WithDescription(description))
}

// Shutdown корректно завершает работу телеметрии
func (tm *TelemetryManager) Shutdown(ctx context.Context) error {
	return tm.shutdown(ctx)
}

// QUICMetrics содержит метрики для QUIC тестирования
type QUICMetrics struct {
	// Счетчики
	ConnectionsTotal     instrument.Int64Counter
	StreamsTotal         instrument.Int64Counter
	BytesSentTotal       instrument.Int64Counter
	BytesReceivedTotal   instrument.Int64Counter
	ErrorsTotal          instrument.Int64Counter
	RetransmitsTotal     instrument.Int64Counter
	HandshakesTotal      instrument.Int64Counter
	ZeroRTTTotal         instrument.Int64Counter
	OneRTTTotal          instrument.Int64Counter
	KeyUpdatesTotal      instrument.Int64Counter
	DatagramsSentTotal   instrument.Int64Counter
	DatagramsReceivedTotal instrument.Int64Counter

	// Гистограммы
	LatencyHistogram     instrument.Float64Histogram
	JitterHistogram      instrument.Float64Histogram
	HandshakeTimeHistogram instrument.Float64Histogram
	ThroughputHistogram  instrument.Float64Histogram

	// Gauges
	ActiveConnections    instrument.Int64Gauge
	ActiveStreams        instrument.Int64Gauge
	CurrentThroughput    instrument.Float64Gauge
	CurrentLatency       instrument.Float64Gauge
	PacketLossRate       instrument.Float64Gauge
}

// NewQUICMetrics создает метрики для QUIC тестирования
func NewQUICMetrics(tm *TelemetryManager) (*QUICMetrics, error) {
	// Счетчики
	connectionsTotal, err := tm.CreateInt64Counter("quic_connections_total", "Total number of QUIC connections established")
	if err != nil {
		return nil, fmt.Errorf("failed to create connections counter: %w", err)
	}

	streamsTotal, err := tm.CreateInt64Counter("quic_streams_total", "Total number of QUIC streams created")
	if err != nil {
		return nil, fmt.Errorf("failed to create streams counter: %w", err)
	}

	bytesSentTotal, err := tm.CreateInt64Counter("quic_bytes_sent_total", "Total bytes sent over QUIC connections")
	if err != nil {
		return nil, fmt.Errorf("failed to create bytes sent counter: %w", err)
	}

	bytesReceivedTotal, err := tm.CreateInt64Counter("quic_bytes_received_total", "Total bytes received over QUIC connections")
	if err != nil {
		return nil, fmt.Errorf("failed to create bytes received counter: %w", err)
	}

	errorsTotal, err := tm.CreateInt64Counter("quic_errors_total", "Total number of QUIC errors")
	if err != nil {
		return nil, fmt.Errorf("failed to create errors counter: %w", err)
	}

	retransmitsTotal, err := tm.CreateInt64Counter("quic_retransmits_total", "Total number of QUIC packet retransmissions")
	if err != nil {
		return nil, fmt.Errorf("failed to create retransmits counter: %w", err)
	}

	handshakesTotal, err := tm.CreateInt64Counter("quic_handshakes_total", "Total number of QUIC handshakes completed")
	if err != nil {
		return nil, fmt.Errorf("failed to create handshakes counter: %w", err)
	}

	zeroRTTTotal, err := tm.CreateInt64Counter("quic_zero_rtt_total", "Total number of 0-RTT connections")
	if err != nil {
		return nil, fmt.Errorf("failed to create zero RTT counter: %w", err)
	}

	oneRTTTotal, err := tm.CreateInt64Counter("quic_one_rtt_total", "Total number of 1-RTT connections")
	if err != nil {
		return nil, fmt.Errorf("failed to create one RTT counter: %w", err)
	}

	keyUpdatesTotal, err := tm.CreateInt64Counter("quic_key_updates_total", "Total number of QUIC key updates")
	if err != nil {
		return nil, fmt.Errorf("failed to create key updates counter: %w", err)
	}

	datagramsSentTotal, err := tm.CreateInt64Counter("quic_datagrams_sent_total", "Total number of QUIC datagrams sent")
	if err != nil {
		return nil, fmt.Errorf("failed to create datagrams sent counter: %w", err)
	}

	datagramsReceivedTotal, err := tm.CreateInt64Counter("quic_datagrams_received_total", "Total number of QUIC datagrams received")
	if err != nil {
		return nil, fmt.Errorf("failed to create datagrams received counter: %w", err)
	}

	// Гистограммы
	latencyHistogram, err := tm.CreateFloat64Histogram("quic_latency_seconds", "QUIC latency distribution")
	if err != nil {
		return nil, fmt.Errorf("failed to create latency histogram: %w", err)
	}

	jitterHistogram, err := tm.CreateFloat64Histogram("quic_jitter_seconds", "QUIC jitter distribution")
	if err != nil {
		return nil, fmt.Errorf("failed to create jitter histogram: %w", err)
	}

	handshakeTimeHistogram, err := tm.CreateFloat64Histogram("quic_handshake_time_seconds", "QUIC handshake time distribution")
	if err != nil {
		return nil, fmt.Errorf("failed to create handshake time histogram: %w", err)
	}

	throughputHistogram, err := tm.CreateFloat64Histogram("quic_throughput_bytes_per_second", "QUIC throughput distribution")
	if err != nil {
		return nil, fmt.Errorf("failed to create throughput histogram: %w", err)
	}

	// Gauges
	activeConnections, err := tm.CreateInt64Gauge("quic_active_connections", "Number of active QUIC connections")
	if err != nil {
		return nil, fmt.Errorf("failed to create active connections gauge: %w", err)
	}

	activeStreams, err := tm.CreateInt64Gauge("quic_active_streams", "Number of active QUIC streams")
	if err != nil {
		return nil, fmt.Errorf("failed to create active streams gauge: %w", err)
	}

	currentThroughput, err := tm.CreateFloat64Gauge("quic_current_throughput_mbps", "Current QUIC throughput in Mbps")
	if err != nil {
		return nil, fmt.Errorf("failed to create current throughput gauge: %w", err)
	}

	currentLatency, err := tm.CreateFloat64Gauge("quic_current_latency_ms", "Current QUIC latency in milliseconds")
	if err != nil {
		return nil, fmt.Errorf("failed to create current latency gauge: %w", err)
	}

	packetLossRate, err := tm.CreateFloat64Gauge("quic_packet_loss_rate_percent", "Current QUIC packet loss rate in percent")
	if err != nil {
		return nil, fmt.Errorf("failed to create packet loss rate gauge: %w", err)
	}

	return &QUICMetrics{
		ConnectionsTotal:       connectionsTotal,
		StreamsTotal:           streamsTotal,
		BytesSentTotal:         bytesSentTotal,
		BytesReceivedTotal:     bytesReceivedTotal,
		ErrorsTotal:            errorsTotal,
		RetransmitsTotal:       retransmitsTotal,
		HandshakesTotal:        handshakesTotal,
		ZeroRTTTotal:           zeroRTTTotal,
		OneRTTTotal:            oneRTTTotal,
		KeyUpdatesTotal:        keyUpdatesTotal,
		DatagramsSentTotal:     datagramsSentTotal,
		DatagramsReceivedTotal: datagramsReceivedTotal,
		LatencyHistogram:       latencyHistogram,
		JitterHistogram:        jitterHistogram,
		HandshakeTimeHistogram: handshakeTimeHistogram,
		ThroughputHistogram:    throughputHistogram,
		ActiveConnections:      activeConnections,
		ActiveStreams:          activeStreams,
		CurrentThroughput:      currentThroughput,
		CurrentLatency:         currentLatency,
		PacketLossRate:         packetLossRate,
	}, nil
}

// RecordLatency записывает задержку
func (qm *QUICMetrics) RecordLatency(ctx context.Context, latency time.Duration, attrs ...attribute.KeyValue) {
	qm.LatencyHistogram.Record(ctx, latency.Seconds(), attrs...)
}

// RecordJitter записывает джиттер
func (qm *QUICMetrics) RecordJitter(ctx context.Context, jitter time.Duration, attrs ...attribute.KeyValue) {
	qm.JitterHistogram.Record(ctx, jitter.Seconds(), attrs...)
}

// RecordHandshakeTime записывает время handshake
func (qm *QUICMetrics) RecordHandshakeTime(ctx context.Context, handshakeTime time.Duration, attrs ...attribute.KeyValue) {
	qm.HandshakeTimeHistogram.Record(ctx, handshakeTime.Seconds(), attrs...)
}

// RecordThroughput записывает пропускную способность
func (qm *QUICMetrics) RecordThroughput(ctx context.Context, throughput float64, attrs ...attribute.KeyValue) {
	qm.ThroughputHistogram.Record(ctx, throughput, attrs...)
}

// IncrementConnections увеличивает счетчик соединений
func (qm *QUICMetrics) IncrementConnections(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.ConnectionsTotal.Add(ctx, 1, attrs...)
}

// IncrementStreams увеличивает счетчик потоков
func (qm *QUICMetrics) IncrementStreams(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.StreamsTotal.Add(ctx, 1, attrs...)
}

// AddBytesSent добавляет отправленные байты
func (qm *QUICMetrics) AddBytesSent(ctx context.Context, bytes int64, attrs ...attribute.KeyValue) {
	qm.BytesSentTotal.Add(ctx, bytes, attrs...)
}

// AddBytesReceived добавляет полученные байты
func (qm *QUICMetrics) AddBytesReceived(ctx context.Context, bytes int64, attrs ...attribute.KeyValue) {
	qm.BytesReceivedTotal.Add(ctx, bytes, attrs...)
}

// IncrementErrors увеличивает счетчик ошибок
func (qm *QUICMetrics) IncrementErrors(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.ErrorsTotal.Add(ctx, 1, attrs...)
}

// IncrementRetransmits увеличивает счетчик ретрансмиссий
func (qm *QUICMetrics) IncrementRetransmits(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.RetransmitsTotal.Add(ctx, 1, attrs...)
}

// IncrementHandshakes увеличивает счетчик handshake
func (qm *QUICMetrics) IncrementHandshakes(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.HandshakesTotal.Add(ctx, 1, attrs...)
}

// IncrementZeroRTT увеличивает счетчик 0-RTT соединений
func (qm *QUICMetrics) IncrementZeroRTT(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.ZeroRTTTotal.Add(ctx, 1, attrs...)
}

// IncrementOneRTT увеличивает счетчик 1-RTT соединений
func (qm *QUICMetrics) IncrementOneRTT(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.OneRTTTotal.Add(ctx, 1, attrs...)
}

// IncrementKeyUpdates увеличивает счетчик обновлений ключей
func (qm *QUICMetrics) IncrementKeyUpdates(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.KeyUpdatesTotal.Add(ctx, 1, attrs...)
}

// IncrementDatagramsSent увеличивает счетчик отправленных датаграмм
func (qm *QUICMetrics) IncrementDatagramsSent(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.DatagramsSentTotal.Add(ctx, 1, attrs...)
}

// IncrementDatagramsReceived увеличивает счетчик полученных датаграмм
func (qm *QUICMetrics) IncrementDatagramsReceived(ctx context.Context, attrs ...attribute.KeyValue) {
	qm.DatagramsReceivedTotal.Add(ctx, 1, attrs...)
}

// SetActiveConnections устанавливает количество активных соединений
func (qm *QUICMetrics) SetActiveConnections(ctx context.Context, count int64, attrs ...attribute.KeyValue) {
	qm.ActiveConnections.Record(ctx, count, attrs...)
}

// SetActiveStreams устанавливает количество активных потоков
func (qm *QUICMetrics) SetActiveStreams(ctx context.Context, count int64, attrs ...attribute.KeyValue) {
	qm.ActiveStreams.Record(ctx, count, attrs...)
}

// SetCurrentThroughput устанавливает текущую пропускную способность
func (qm *QUICMetrics) SetCurrentThroughput(ctx context.Context, throughput float64, attrs ...attribute.KeyValue) {
	qm.CurrentThroughput.Record(ctx, throughput, attrs...)
}

// SetCurrentLatency устанавливает текущую задержку
func (qm *QUICMetrics) SetCurrentLatency(ctx context.Context, latency float64, attrs ...attribute.KeyValue) {
	qm.CurrentLatency.Record(ctx, latency, attrs...)
}

// SetPacketLossRate устанавливает текущий процент потерь пакетов
func (qm *QUICMetrics) SetPacketLossRate(ctx context.Context, lossRate float64, attrs ...attribute.KeyValue) {
	qm.PacketLossRate.Record(ctx, lossRate, attrs...)
}
