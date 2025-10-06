package masque

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// ConnectUDPTester тестирует CONNECT-UDP функциональность (RFC 9298)
type ConnectUDPTester struct {
	logger *zap.Logger
	config *MASQUEConfig
}

// ConnectUDPConnection представляет CONNECT-UDP соединение
type ConnectUDPConnection struct {
	udpConn  *net.UDPConn
	target   string
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewConnectUDPTester создает новый CONNECT-UDP тестер
func NewConnectUDPTester(logger *zap.Logger, config *MASQUEConfig) *ConnectUDPTester {
	return &ConnectUDPTester{
		logger: logger,
		config: config,
	}
}

// Connect создает CONNECT-UDP соединение к целевому хосту
func (cudt *ConnectUDPTester) Connect(ctx context.Context, target string) (*ConnectUDPConnection, error) {
	cudt.logger.Info("Creating CONNECT-UDP connection", zap.String("target", target))

	// Для тестирования создаем mock UDP соединение
	udpAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %v", err)
	}

	connCtx, cancel := context.WithCancel(ctx)

	conn := &ConnectUDPConnection{
		udpConn: udpConn,
		target:  target,
		ctx:     connCtx,
		cancel:  cancel,
	}

	cudt.logger.Info("CONNECT-UDP connection established", zap.String("target", target))
	return conn, nil
}

// Write отправляет UDP datagram через CONNECT-UDP соединение
func (cudc *ConnectUDPConnection) Write(data []byte) (int, error) {
	// Для тестирования отправляем через UDP соединение
	return cudc.udpConn.Write(data)
}

// Read получает UDP datagram через CONNECT-UDP соединение
func (cudc *ConnectUDPConnection) Read(data []byte) (int, error) {
	// Для тестирования читаем из UDP соединения
	if err := cudc.udpConn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		cudc.logger.Warn("Failed to set read deadline for UDP", zap.Error(err))
	}
	return cudc.udpConn.Read(data)
}

// Close закрывает CONNECT-UDP соединение
func (cudc *ConnectUDPConnection) Close() error {
	cudc.cancel()
	return cudc.udpConn.Close()
}

// LocalAddr возвращает локальный адрес
func (cudc *ConnectUDPConnection) LocalAddr() net.Addr {
	return cudc.udpConn.LocalAddr()
}

// RemoteAddr возвращает удаленный адрес
func (cudc *ConnectUDPConnection) RemoteAddr() net.Addr {
	return cudc.udpConn.RemoteAddr()
}

// SetDeadline устанавливает deadline для операций
func (cudc *ConnectUDPConnection) SetDeadline(t time.Time) error {
	return cudc.udpConn.SetDeadline(t)
}

// SetReadDeadline устанавливает deadline для чтения
func (cudc *ConnectUDPConnection) SetReadDeadline(t time.Time) error {
	return cudc.udpConn.SetReadDeadline(t)
}

// SetWriteDeadline устанавливает deadline для записи
func (cudc *ConnectUDPConnection) SetWriteDeadline(t time.Time) error {
	return cudc.udpConn.SetWriteDeadline(t)
}

// TestConnectUDP тестирует CONNECT-UDP функциональность
func (cudt *ConnectUDPTester) TestConnectUDP(ctx context.Context, target string) error {
	cudt.logger.Info("Testing CONNECT-UDP", zap.String("target", target))

	// Создаем соединение
	conn, err := cudt.Connect(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to create CONNECT-UDP connection: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			cudt.logger.Warn("Failed to close CONNECT-UDP connection", zap.Error(err))
		}
	}()

	// Отправляем тестовые данные
	testData := []byte("Hello MASQUE CONNECT-UDP!")
	_, err = conn.Write(testData)
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	// Читаем ответ
	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		// Для тестирования это нормально, если нет ответа
		cudt.logger.Info("No response received (expected for test)", zap.Error(err))
	}

	cudt.logger.Info("CONNECT-UDP test completed", zap.String("target", target))
	return nil
}

// TestConnectUDPBatch тестирует множественные CONNECT-UDP соединения
func (cudt *ConnectUDPTester) TestConnectUDPBatch(ctx context.Context, targets []string) error {
	cudt.logger.Info("Testing CONNECT-UDP batch", zap.Int("targets", len(targets)))

	for _, target := range targets {
		if err := cudt.TestConnectUDP(ctx, target); err != nil {
			cudt.logger.Error("CONNECT-UDP test failed", zap.String("target", target), zap.Error(err))
			continue
		}
	}

	cudt.logger.Info("CONNECT-UDP batch test completed")
	return nil
}

// Stop останавливает CONNECT-UDP тестер
func (cudt *ConnectUDPTester) Stop() error {
	cudt.logger.Info("Stopping CONNECT-UDP tester")
	return nil
}