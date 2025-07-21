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

	"quck-test/internal"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
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
}

// Run запускает сервер с параметрами из TestConfig
func Run(cfg internal.TestConfig) {
	metrics := &serverMetrics{Start: time.Now()}

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
		listener.Close()
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

	for {
		select {
		case <-done:
			return
		case <-time.After(2 * time.Second):
			printServerMetrics(metrics)
		}
	}
}

func handleConn(conn quic.Connection, metrics *serverMetrics) {
	defer conn.CloseWithError(0, "bye")
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
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			metrics.mu.Lock()
			metrics.Bytes += int64(n)
			metrics.mu.Unlock()
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
	if cfg.NoTLS {
		return &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"quic-test"}}
	}
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			log.Fatalf("Ошибка загрузки сертификата: %v", err)
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"quic-test"}}
	}
	certPEM, keyPEM := internal.GenerateSelfSignedTLS()
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalf("Ошибка генерации self-signed сертификата: %v", err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, NextProtos: []string{"quic-test"}}
}

func printServerMetrics(metrics *serverMetrics) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	fmt.Print("\033[H\033[2J")
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"Connections", "Streams", "Bytes", "Errors", "Uptime (s)"}
	table.Append(headers)
	uptime := time.Since(metrics.Start).Seconds()
	row := []string{
		green(fmt.Sprintf("%d", metrics.Connections)),
		blue(fmt.Sprintf("%d", metrics.Streams)),
		blue(fmt.Sprintf("%.2f KB", float64(metrics.Bytes)/1024)),
		red(fmt.Sprintf("%d", metrics.Errors)),
		yellow(fmt.Sprintf("%.0f", uptime)),
	}
	table.Append(row)
	table.Render()
}

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
	http.ListenAndServe(":2113", nil)
} 