package quic

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"sync"
	"time"

	quic "github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// QUICClient представляет QUIC клиент
type QUICClient struct {
	logger      *zap.Logger
	serverAddr  string
	conn        *quic.Conn
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	isConnected bool
	streams     map[quic.StreamID]*quic.Stream
}

// QUICClientConfig конфигурация QUIC клиента
type QUICClientConfig struct {
	ServerAddr     string        `json:"server_addr"`
	MaxStreams     int           `json:"max_streams"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
}

// NewQUICClient создает новый QUIC клиент
func NewQUICClient(logger *zap.Logger, config *QUICClientConfig) *QUICClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &QUICClient{
		logger:     logger,
		serverAddr: config.ServerAddr,
		ctx:        ctx,
		cancel:     cancel,
		streams:    make(map[quic.StreamID]*quic.Stream),
	}
}

// Connect подключается к QUIC серверу
func (qc *QUICClient) Connect() error {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if qc.isConnected {
		return fmt.Errorf("QUIC client is already connected")
	}

	// Создаем QUIC конфигурацию
	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	}

	// Подключаемся к серверу
	conn, err := quic.DialAddr(qc.ctx, qc.serverAddr, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-test"},
	}, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", qc.serverAddr, err)
	}

	qc.conn = conn
	qc.isConnected = true

	qc.logger.Info("QUIC client connected", zap.String("server", qc.serverAddr))

	// Запускаем обработку потоков
	go qc.handleStreams()

	return nil
}

// Disconnect отключается от сервера
func (qc *QUICClient) Disconnect() error {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if !qc.isConnected {
		return nil
	}

	qc.cancel()

	if qc.conn != nil {
		if err := (*qc.conn).CloseWithError(0, "client disconnect"); err != nil {
			qc.logger.Warn("Failed to close QUIC connection", zap.Error(err))
		}
	}

	// Закрываем все потоки
	for _, stream := range qc.streams {
		if err := stream.Close(); err != nil {
			qc.logger.Warn("Failed to close QUIC stream", zap.Error(err))
		}
	}

	qc.isConnected = false
	qc.logger.Info("QUIC client disconnected")

	return nil
}

// IsConnected возвращает статус подключения
func (qc *QUICClient) IsConnected() bool {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.isConnected
}

// GetStreams возвращает количество активных потоков
func (qc *QUICClient) GetStreams() int {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return len(qc.streams)
}

// SendData отправляет данные на сервер
func (qc *QUICClient) SendData(data []byte) error {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	if !qc.isConnected || qc.conn == nil {
		return fmt.Errorf("client is not connected")
	}

	// Создаем новый поток
	stream, err := (*qc.conn).OpenStreamSync(qc.ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}

	// Добавляем поток в список
	qc.mu.Lock()
	qc.streams[stream.StreamID()] = stream
	qc.mu.Unlock()

	// Отправляем данные
	_, err = stream.Write(data)
	if err != nil {
		qc.mu.Lock()
		delete(qc.streams, stream.StreamID())
		qc.mu.Unlock()
		if err := stream.Close(); err != nil {
			qc.logger.Warn("Failed to close QUIC stream on error", zap.Error(err))
		}
		return fmt.Errorf("failed to send data: %v", err)
	}

	qc.logger.Debug("Data sent to server",
		zap.Int("bytes", len(data)),
		zap.Uint64("stream_id", uint64(stream.StreamID())))

	return nil
}

// SendTestData отправляет тестовые данные
func (qc *QUICClient) SendTestData(packetSize int, count int) error {
	if !qc.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	qc.logger.Info("Sending test data",
		zap.Int("packet_size", packetSize),
		zap.Int("count", count))

	for i := 0; i < count; i++ {
		// Генерируем тестовые данные
		data := make([]byte, packetSize)
		for j := range data {
			data[j] = byte(i + j)
		}

		if err := qc.SendData(data); err != nil {
			qc.logger.Error("Failed to send test data",
				zap.Int("packet", i),
				zap.Error(err))
			return err
		}

		// Небольшая задержка между пакетами
		time.Sleep(10 * time.Millisecond)
	}

	qc.logger.Info("Test data sent successfully", zap.Int("packets", count))
	return nil
}

// handleStreams обрабатывает входящие потоки
func (qc *QUICClient) handleStreams() {
	for {
		select {
		case <-qc.ctx.Done():
			return
		default:
			if qc.conn == nil {
				return
			}

			stream, err := (*qc.conn).AcceptStream(qc.ctx)
			if err != nil {
				if qc.ctx.Err() != nil {
					return
				}
				qc.logger.Debug("Failed to accept stream", zap.Error(err))
				continue
			}

			// Добавляем поток в список
			qc.mu.Lock()
			qc.streams[stream.StreamID()] = stream
			qc.mu.Unlock()

			// Обрабатываем поток
			go qc.handleStream(stream)
		}
	}
}

// handleStream обрабатывает отдельный поток
func (qc *QUICClient) handleStream(stream *quic.Stream) {
	defer func() {
		// Удаляем поток из списка
		qc.mu.Lock()
		delete(qc.streams, stream.StreamID())
		qc.mu.Unlock()

		if err := stream.Close(); err != nil {
			qc.logger.Warn("Failed to close QUIC stream in handler", zap.Error(err))
		}
		qc.logger.Debug("Stream closed", zap.Uint64("stream_id", uint64(stream.StreamID())))
	}()

	buffer := make([]byte, 4096)

	for {
		select {
		case <-qc.ctx.Done():
			return
		default:
			n, err := stream.Read(buffer)
			if err != nil {
				if err == io.EOF {
					qc.logger.Debug("Stream EOF", zap.Uint64("stream_id", uint64(stream.StreamID())))
					return
				}
				qc.logger.Debug("Stream read error", zap.Error(err))
				return
			}

			qc.logger.Debug("Received data from server",
				zap.Uint64("stream_id", uint64(stream.StreamID())),
				zap.Int("bytes", n))
		}
	}
}
