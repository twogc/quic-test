package integration

import (
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"go.uber.org/zap"
)

// QUICTracer реализует logging.Tracer для перехвата событий QUIC
type QUICTracer struct {
	logger            *zap.Logger
	simpleIntegration *SimpleIntegration
}

// NewQUICTracer создает новый QUIC tracer
func NewQUICTracer(logger *zap.Logger, si *SimpleIntegration) *logging.Tracer {
	tracer := &logging.Tracer{}
	return tracer
}

// NewConnectionTracerForConnection создает ConnectionTracer для соединения
func NewConnectionTracerForConnection(logger *zap.Logger, si *SimpleIntegration, connectionID string) *logging.ConnectionTracer {
	return NewConnectionTracer(logger, si, connectionID)
}

// connectionStorage хранит connection для каждого connectionID
var connectionStorage = make(map[string]interface{})
var connectionStorageMu sync.RWMutex

// StoreConnection сохраняет connection для использования в tracer
func StoreConnection(connectionID string, conn interface{}) {
	connectionStorageMu.Lock()
	defer connectionStorageMu.Unlock()
	connectionStorage[connectionID] = conn
}

// GetConnection получает connection по ID
func GetConnection(connectionID string) interface{} {
	connectionStorageMu.RLock()
	defer connectionStorageMu.RUnlock()
	return connectionStorage[connectionID]
}

// ConnectionTracer отслеживает события конкретного соединения
func NewConnectionTracer(logger *zap.Logger, si *SimpleIntegration, connectionID string) *logging.ConnectionTracer {
	var lastBytesInFlight logging.ByteCount
	var lastUpdateTime time.Time
	
	ct := &logging.ConnectionTracer{
		UpdatedMetrics: func(rttStats *logging.RTTStats, cwnd, bytesInFlight logging.ByteCount, packetsInFlight int) {
			if si == nil || rttStats == nil {
				return
			}
			
			now := time.Now()
			
			// Получаем реальный RTT из статистики
			smoothedRTT := rttStats.SmoothedRTT()
			minRTT := rttStats.MinRTT()
			latestRTT := rttStats.LatestRTT()
			
			// Используем latest RTT если доступен, иначе smoothed
			var rtt time.Duration
			if latestRTT > 0 {
				rtt = latestRTT
			} else if smoothedRTT > 0 {
				rtt = smoothedRTT
			} else if minRTT > 0 {
				rtt = minRTT
			}
			
			if rtt > 0 {
				// Вычисляем сколько байт было подтверждено (уменьшение bytesInFlight)
				var ackedBytes int
				if lastBytesInFlight > 0 && bytesInFlight < lastBytesInFlight {
					ackedBytes = int(lastBytesInFlight - bytesInFlight)
				} else if bytesInFlight == 0 && lastBytesInFlight > 0 {
					// Все подтверждено
					ackedBytes = int(lastBytesInFlight)
				} else {
					// Если нет уменьшения bytesInFlight, используем cwnd как приближение
					ackedBytes = int(cwnd / 10) // Примерно 10% от cwnd за раз
					if ackedBytes == 0 {
						ackedBytes = 1200 // Минимум 1 пакет
					}
				}
				
				// Уведомляем SimpleIntegration о реальном ACK
				if ackedBytes > 0 {
					// Получаем connection из хранилища
					conn := GetConnection(connectionID)
					if conn != nil {
						// Вычисляем интервал с последнего обновления
						interval := time.Duration(0)
						if !lastUpdateTime.IsZero() {
							interval = now.Sub(lastUpdateTime)
						}
						
						// Если интервал достаточно большой (больше 1ms), уведомляем об ACK
						if interval > 1*time.Millisecond || !lastUpdateTime.IsZero() {
							logger.Debug("ACK detected via UpdatedMetrics",
								zap.String("connection_id", connectionID),
								zap.Duration("rtt", rtt),
								zap.Int("acked_bytes", ackedBytes),
								zap.Int64("bytes_in_flight", int64(bytesInFlight)),
								zap.Int64("cwnd", int64(cwnd)),
								zap.Duration("interval", interval))
							
							// Уведомляем SimpleIntegration о реальном ACK
							if quicConn, ok := conn.(quic.Connection); ok {
								si.OnAckReceived(quicConn, ackedBytes, rtt)
							}
						}
					}
				}
				
				lastBytesInFlight = bytesInFlight
				lastUpdateTime = now
			}
		},
		
		AcknowledgedPacket: func(encLevel logging.EncryptionLevel, pn logging.PacketNumber) {
			if si != nil {
				logger.Debug("Packet acknowledged",
					zap.String("connection_id", connectionID),
					zap.Uint64("packet_number", uint64(pn)))
			}
		},
		
		LostPacket: func(encLevel logging.EncryptionLevel, pn logging.PacketNumber, reason logging.PacketLossReason) {
			if si != nil {
				logger.Info("Packet lost",
					zap.String("connection_id", connectionID),
					zap.Uint64("packet_number", uint64(pn)),
					zap.String("reason", string(reason)))
				
				// Уведомляем о потере пакета
				// si.OnLossDetected(conn, estimatedPacketSize)
			}
		},
		
		SentLongHeaderPacket: func(hdr *logging.ExtendedHeader, size logging.ByteCount, ecn logging.ECN, ack *logging.AckFrame, frames []logging.Frame) {
			if si != nil {
				logger.Debug("Long header packet sent",
					zap.String("connection_id", connectionID),
					zap.Uint64("packet_number", uint64(hdr.PacketNumber)),
					zap.Int64("size", int64(size)))
			}
		},
		
		SentShortHeaderPacket: func(hdr *logging.ShortHeader, size logging.ByteCount, ecn logging.ECN, ack *logging.AckFrame, frames []logging.Frame) {
			if si != nil {
				logger.Debug("Short header packet sent",
					zap.String("connection_id", connectionID),
					zap.Uint64("packet_number", uint64(hdr.PacketNumber)),
					zap.Int64("size", int64(size)))
			}
		},
		
		Close: func() {
			logger.Debug("Connection tracer closed",
				zap.String("connection_id", connectionID))
		},
	}
	
	return ct
}
