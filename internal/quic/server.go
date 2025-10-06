package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// QUICServer представляет QUIC сервер
type QUICServer struct {
	logger      *zap.Logger
	addr        string
	listener    *quic.Listener
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	isRunning   bool
	connections map[string]*quic.Connection
}

// QUICServerConfig конфигурация QUIC сервера
type QUICServerConfig struct {
	Addr           string        `json:"addr"`
	MaxConnections int           `json:"max_connections"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
	KeepAlive      time.Duration `json:"keep_alive"`
}

// NewQUICServer создает новый QUIC сервер
func NewQUICServer(logger *zap.Logger, config *QUICServerConfig) *QUICServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &QUICServer{
		logger:      logger,
		addr:        config.Addr,
		ctx:         ctx,
		cancel:      cancel,
		connections: make(map[string]*quic.Connection),
	}
}

// Start запускает QUIC сервер
func (qs *QUICServer) Start() error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if qs.isRunning {
		return fmt.Errorf("QUIC server is already running")
	}

	// Создаем простую TLS конфигурацию для разработки
	cert, err := qs.generateSelfSignedCert()
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		NextProtos:         []string{"quic-test"},
		InsecureSkipVerify: true,
	}

	// Создаем QUIC конфигурацию
	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	}

	// Создаем listener
	listener, err := quic.ListenAddr(qs.addr, tlsConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", qs.addr, err)
	}

	qs.listener = listener
	qs.isRunning = true

	qs.logger.Info("QUIC server started", zap.String("addr", qs.addr))

	// Запускаем обработку соединений
	go qs.handleConnections()

	return nil
}

// Stop останавливает QUIC сервер
func (qs *QUICServer) Stop() error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if !qs.isRunning {
		return nil
	}

	qs.cancel()

	if qs.listener != nil {
		if err := qs.listener.Close(); err != nil {
			qs.logger.Warn("Failed to close QUIC listener", zap.Error(err))
		}
	}

	// Закрываем все соединения
	for _, conn := range qs.connections {
		if err := (*conn).CloseWithError(0, "server shutdown"); err != nil {
			qs.logger.Warn("Failed to close QUIC connection", zap.Error(err))
		}
	}

	qs.isRunning = false
	qs.logger.Info("QUIC server stopped")

	return nil
}

// IsRunning возвращает статус сервера
func (qs *QUICServer) IsRunning() bool {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return qs.isRunning
}

// GetConnections возвращает количество активных соединений
func (qs *QUICServer) GetConnections() int {
	qs.mu.RLock()
	defer qs.mu.RUnlock()
	return len(qs.connections)
}

// handleConnections обрабатывает входящие соединения
func (qs *QUICServer) handleConnections() {
	for {
		select {
		case <-qs.ctx.Done():
			return
		default:
			conn, err := qs.listener.Accept(qs.ctx)
			if err != nil {
				if qs.ctx.Err() != nil {
					return
				}
				qs.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}

			// Добавляем соединение в список
			qs.mu.Lock()
			connID := fmt.Sprintf("%p", conn)
			qs.connections[connID] = &conn
			qs.mu.Unlock()

			qs.logger.Info("New QUIC connection accepted",
				zap.String("conn_id", connID),
				zap.String("remote_addr", conn.RemoteAddr().String()))

			// Обрабатываем соединение
			go qs.handleConnection(&conn, connID)
		}
	}
}

// handleConnection обрабатывает отдельное соединение
func (qs *QUICServer) handleConnection(conn *quic.Connection, connID string) {
	defer func() {
		// Удаляем соединение из списка
		qs.mu.Lock()
		delete(qs.connections, connID)
		qs.mu.Unlock()

		(*conn).CloseWithError(0, "connection closed")
		qs.logger.Info("QUIC connection closed", zap.String("conn_id", connID))
	}()

	// Обрабатываем потоки
	for {
		select {
		case <-qs.ctx.Done():
			return
		default:
			stream, err := (*conn).AcceptStream(qs.ctx)
			if err != nil {
				if qs.ctx.Err() != nil {
					return
				}
				qs.logger.Debug("Failed to accept stream", zap.Error(err))
				continue
			}

			// Обрабатываем поток
			go qs.handleStream(stream, connID)
		}
	}
}

// handleStream обрабатывает поток данных
func (qs *QUICServer) handleStream(stream quic.Stream, connID string) {
	defer stream.Close()

	buffer := make([]byte, 4096)

	for {
		select {
		case <-qs.ctx.Done():
			return
		default:
			n, err := stream.Read(buffer)
			if err != nil {
				qs.logger.Debug("Stream read error", zap.Error(err))
				return
			}

			qs.logger.Debug("Received data",
				zap.String("conn_id", connID),
				zap.Int("bytes", n))

			// Эхо-ответ
			_, err = stream.Write(buffer[:n])
			if err != nil {
				qs.logger.Debug("Stream write error", zap.Error(err))
				return
			}
		}
	}
}

// generateTLSConfig создает TLS конфигурацию для разработки
func (qs *QUICServer) generateTLSConfig() (*tls.Config, error) {
	// Создаем самоподписанный сертификат
	cert, err := qs.generateSelfSignedCert()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		NextProtos:         []string{"quic-test"},
		InsecureSkipVerify: true,
	}, nil
}

// generateSelfSignedCert создает самоподписанный сертификат
func (qs *QUICServer) generateSelfSignedCert() (tls.Certificate, error) {
	// Создаем приватный ключ
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Создаем сертификат
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"QUCK Test"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Кодируем в PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	// Создаем TLS сертификат
	return tls.X509KeyPair(certPEM, keyPEM)
}
