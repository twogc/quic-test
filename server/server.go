package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"quic-test/internal"
	"quic-test/internal/fec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	quic "github.com/quic-go/quic-go"
)

// serverMetrics хранит метрики сервера
type serverMetrics struct {
	mu          sync.Mutex
	Connections int
	Streams     int
	Bytes       int64
	Errors      int
	Start       time.Time
	FECDecoder  *fec.FECDecoder // FEC decoder для восстановления пакетов
}

// Run запускает сервер с параметрами из TestConfig
func Run(cfg internal.TestConfig) {
	metrics := &serverMetrics{
		Start:      time.Now(),
		FECDecoder: fec.NewFECDecoder(), // Инициализируем FEC decoder если нужно
	}
	
	// Периодическая очистка старых групп FEC
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if metrics.FECDecoder != nil {
				metrics.FECDecoder.CleanupGroups()
			}
		}
	}()

	if cfg.Prometheus {
		go startPrometheusExporter(metrics)
	}

	tlsConf := makeTLSConfig(cfg)
	listener, err := quic.ListenAddr(cfg.Addr, tlsConf, &quic.Config{})
	if err != nil {
		log.Fatalf("Ошибка запуска QUIC сервера: %v", err)
	}
	log.Printf("QUIC сервер слушает %s", cfg.Addr)

	done := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Остановка сервера...")
		if err := listener.Close(); err != nil {
			log.Printf("Warning: failed to close listener: %v\n", err)
		}
		close(done)
	}()

	go func() {
		for {
			conn, err := listener.Accept(context.Background())
			if err != nil {
				metrics.mu.Lock()
				metrics.Errors++
				metrics.mu.Unlock()
				break
			}
			metrics.mu.Lock()
			metrics.Connections++
			metrics.mu.Unlock()
			go handleConn(conn, metrics)
		}
	}()

	// Ожидание завершения
	<-done
}

func handleConn(conn quic.Connection, metrics *serverMetrics) {
	defer func() {
		if err := conn.CloseWithError(0, "bye"); err != nil {
			log.Printf("Warning: failed to close connection: %v\n", err)
		}
	}()
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			metrics.mu.Lock()
			metrics.Errors++
			metrics.mu.Unlock()
			return
		}
		metrics.mu.Lock()
		metrics.Streams++
		metrics.mu.Unlock()
		go handleStream(stream, metrics)
	}
}

func handleStream(stream quic.Stream, metrics *serverMetrics) {
	buf := make([]byte, 4096)
	packetID := uint64(0)
	groupID := uint64(0)
	
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			// Проверяем, является ли это FEC repair пакетом (начинается с 0xFE 0xC0)
			if n >= 11 && buf[0] == 0xFE && buf[1] == 0xC0 {
				// Это FEC repair пакет
				if metrics.FECDecoder != nil {
					recovered, recoveredList := metrics.FECDecoder.AddRedundancyPacket(buf[:n])
					if recovered && len(recoveredList) > 0 {
						// Успешно восстановлены пакеты
						for _, rec := range recoveredList {
							metrics.mu.Lock()
							metrics.Bytes += int64(len(rec.Data))
							metrics.mu.Unlock()
						}
					}
				}
			} else {
				// Обычный пакет
				metrics.mu.Lock()
				metrics.Bytes += int64(n)
				metrics.mu.Unlock()
				
				// Добавляем в FEC decoder для возможного восстановления
				if metrics.FECDecoder != nil {
					metrics.FECDecoder.AddPacket(buf[:n], packetID, groupID)
					packetID++
					if packetID >= 10 {
						packetID = 0
						groupID++
					}
				}
			}
		}
		if err != nil {
			if err.Error() != "EOF" {
				metrics.mu.Lock()
				metrics.Errors++
				metrics.mu.Unlock()
			}
			return
		}
	}
}

func makeTLSConfig(cfg internal.TestConfig) *tls.Config {
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			log.Fatalf("Ошибка загрузки сертификата: %v", err)
		}
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"quic-test"},
			MinVersion:   tls.VersionTLS12,
		}
	}
	
	// Используем единую функцию для генерации TLS конфигурации
	return internal.GenerateTLSConfig(cfg.NoTLS)
}

// printServerMetrics удалена - больше не используется

func startPrometheusExporter(metrics *serverMetrics) {
	connections := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_server_connections_total",
		Help: "Total connections",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return float64(metrics.Connections)
	})
	streams := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_server_streams_total",
		Help: "Total streams",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return float64(metrics.Streams)
	})
	bytes := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_server_bytes_total",
		Help: "Total bytes received",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return float64(metrics.Bytes)
	})
	errors := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_server_errors_total",
		Help: "Total errors",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return float64(metrics.Errors)
	})
	uptime := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_server_uptime_seconds",
		Help: "Server uptime in seconds",
	}, func() float64 {
		metrics.mu.Lock()
		defer metrics.mu.Unlock()
		return time.Since(metrics.Start).Seconds()
	})

	prometheus.MustRegister(connections, streams, bytes, errors, uptime)
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Prometheus endpoint сервера доступен на :2113/metrics")
	if err := http.ListenAndServe(":2113", nil); err != nil {
		log.Printf("Failed to start Prometheus server: %v", err)
	}
}
