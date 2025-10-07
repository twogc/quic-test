package experimental

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// QlogTracer реализует qlog трассировку для QUIC
type QlogTracer struct {
	logger    *zap.Logger
	outputDir string
	mu        sync.RWMutex
	files     map[string]*os.File
	events    map[string][]QlogEvent
}

// QlogEvent представляет событие в qlog
type QlogEvent struct {
	Time     time.Time              `json:"time"`
	Category string                 `json:"category"`
	Event    string                 `json:"event"`
	Data     map[string]interface{} `json:"data"`
}

// QlogConnection представляет соединение в qlog
type QlogConnection struct {
	ConnectionID string    `json:"connection_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Events       []QlogEvent `json:"events"`
}

// NewQlogTracer создает новый qlog трассировщик
func NewQlogTracer(logger *zap.Logger, outputDir string) *QlogTracer {
	// Создаем директорию если не существует
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Warn("Failed to create qlog directory", zap.Error(err))
	}
	
	return &QlogTracer{
		logger:    logger,
		outputDir: outputDir,
		files:     make(map[string]*os.File),
		events:    make(map[string][]QlogEvent),
	}
}

// StartConnection начинает трассировку соединения
func (qt *QlogTracer) StartConnection(connectionID string) error {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	
	// Создаем файл для соединения
	filename := fmt.Sprintf("connection_%s.qlog", connectionID)
	filepath := filepath.Join(qt.outputDir, filename)
	
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create qlog file: %v", err)
	}
	
	qt.files[connectionID] = file
	qt.events[connectionID] = make([]QlogEvent, 0)
	
	// Записываем заголовок qlog
	header := map[string]interface{}{
		"qlog_version": "draft-02",
		"title":        "2GC QUIC Test qlog",
		"description":  "Experimental QUIC connection trace",
		"traces": []map[string]interface{}{
			{
				"common_fields": map[string]interface{}{
					"ODCID": connectionID,
				},
				"vantage_point": map[string]interface{}{
					"type": "server",
				},
				"events": []QlogEvent{},
			},
		},
	}
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(header); err != nil {
		file.Close()
		delete(qt.files, connectionID)
		return fmt.Errorf("failed to write qlog header: %v", err)
	}
	
	qt.logger.Info("Started qlog tracing",
		zap.String("connection_id", connectionID),
		zap.String("file", filepath))
	
	return nil
}

// LogEvent записывает событие в qlog
func (qt *QlogTracer) LogEvent(connectionID, category, event string, data map[string]interface{}) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	
	// Проверяем, что соединение активно
	if _, exists := qt.events[connectionID]; !exists {
		qt.logger.Warn("Attempted to log event for unknown connection",
			zap.String("connection_id", connectionID))
		return
	}
	
	qlogEvent := QlogEvent{
		Time:     time.Now(),
		Category: category,
		Event:    event,
		Data:     data,
	}
	
	qt.events[connectionID] = append(qt.events[connectionID], qlogEvent)
	
	// Записываем событие в файл
	if file, exists := qt.files[connectionID]; exists {
		eventJSON, _ := json.Marshal(qlogEvent)
		file.WriteString(fmt.Sprintf(",\n%s", string(eventJSON)))
		file.Sync()
	}
}

// LogPacketSent записывает событие отправки пакета
func (qt *QlogTracer) LogPacketSent(connectionID string, packetNumber uint64, packetSize int, packetType string) {
	data := map[string]interface{}{
		"packet_number": packetNumber,
		"packet_size":   packetSize,
		"packet_type":   packetType,
	}
	qt.LogEvent(connectionID, "transport", "packet_sent", data)
}

// LogPacketReceived записывает событие получения пакета
func (qt *QlogTracer) LogPacketReceived(connectionID string, packetNumber uint64, packetSize int, packetType string) {
	data := map[string]interface{}{
		"packet_number": packetNumber,
		"packet_size":   packetSize,
		"packet_type":   packetType,
	}
	qt.LogEvent(connectionID, "transport", "packet_received", data)
}

// LogPacketLost записывает событие потери пакета
func (qt *QlogTracer) LogPacketLost(connectionID string, packetNumber uint64, packetType string) {
	data := map[string]interface{}{
		"packet_number": packetNumber,
		"packet_type":   packetType,
	}
	qt.LogEvent(connectionID, "recovery", "packet_lost", data)
}

// LogACKSent записывает событие отправки ACK
func (qt *QlogTracer) LogACKSent(connectionID string, ackRanges []map[string]uint64) {
	data := map[string]interface{}{
		"ack_ranges": ackRanges,
	}
	qt.LogEvent(connectionID, "transport", "packet_sent", data)
}

// LogConnectionState записывает изменение состояния соединения
func (qt *QlogTracer) LogConnectionState(connectionID, state string, details map[string]interface{}) {
	data := map[string]interface{}{
		"state": state,
	}
	for k, v := range details {
		data[k] = v
	}
	qt.LogEvent(connectionID, "connectivity", "connection_state_changed", data)
}

// LogCongestionControl записывает события управления перегрузкой
func (qt *QlogTracer) LogCongestionControl(connectionID string, cwnd int64, ssthresh int64, algorithm string) {
	data := map[string]interface{}{
		"congestion_window":     cwnd,
		"slow_start_threshold": ssthresh,
		"algorithm":            algorithm,
	}
	qt.LogEvent(connectionID, "recovery", "congestion_state_changed", data)
}

// EndConnection завершает трассировку соединения
func (qt *QlogTracer) EndConnection(connectionID string) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	
	if file, exists := qt.files[connectionID]; exists {
		// Закрываем JSON объект
		file.WriteString("\n]")
		file.Close()
		delete(qt.files, connectionID)
	}
	
	qt.logger.Info("Ended qlog tracing",
		zap.String("connection_id", connectionID),
		zap.Int("events_count", len(qt.events[connectionID])))
	
	delete(qt.events, connectionID)
}

// Close закрывает все файлы qlog
func (qt *QlogTracer) Close() {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	
	for connectionID, file := range qt.files {
		file.WriteString("\n]")
		file.Close()
		qt.logger.Info("Closed qlog file", zap.String("connection_id", connectionID))
	}
	
	qt.files = make(map[string]*os.File)
	qt.events = make(map[string][]QlogEvent)
	
	qt.logger.Info("Qlog tracer closed")
}

// GetConnectionEvents возвращает события для соединения
func (qt *QlogTracer) GetConnectionEvents(connectionID string) []QlogEvent {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	
	if events, exists := qt.events[connectionID]; exists {
		return events
	}
	return nil
}

// GetStats возвращает статистику qlog трассировки
func (qt *QlogTracer) GetStats() map[string]interface{} {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	
	stats := map[string]interface{}{
		"active_connections": len(qt.files),
		"total_events":       0,
	}
	
	for _, events := range qt.events {
		stats["total_events"] = stats["total_events"].(int) + len(events)
	}
	
	return stats
}

