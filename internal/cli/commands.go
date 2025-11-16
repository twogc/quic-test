package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"quic-test/client"
	"quic-test/internal"
	"quic-test/internal/dashboard"
	"quic-test/internal/quic"
	"quic-test/server"

	"go.uber.org/zap"
)

// Глобальный менеджер метрик
var metricsManager = dashboard.NewMetricsManager()

// Глобальный QUIC менеджер
var quicManager *quic.QUICManager

// Command представляет команду CLI
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// Commands содержит все доступные команды
var Commands = map[string]Command{
	"server": {
		Name:        "server",
		Description: "Запуск QUIC сервера",
		Handler:     runServer,
	},
	"client": {
		Name:        "client",
		Description: "Запуск QUIC клиента",
		Handler:     runClient,
	},
	"test": {
		Name:        "test",
		Description: "Запуск тестирования (сервер+клиент)",
		Handler:     runTest,
	},
	"dashboard": {
		Name:        "dashboard",
		Description: "Запуск веб-интерфейса",
		Handler:     runDashboard,
	},
	"masque": {
		Name:        "masque",
		Description: "Запуск MASQUE тестирования",
		Handler:     runMASQUE,
	},
	"ice": {
		Name:        "ice",
		Description: "Запуск ICE/STUN/TURN тестирования",
		Handler:     runICE,
	},
	"enhanced": {
		Name:        "enhanced",
		Description: "Запуск расширенного тестирования (MASQUE + ICE + QUIC)",
		Handler:     runEnhanced,
	},
}

// ParseFlags парсит флаги командной строки
func ParseFlags() (string, map[string]interface{}) {
	mode := flag.String("mode", "server", "Режим работы: server, client, test, dashboard, masque, ice, enhanced")
	version := flag.Bool("version", false, "Показать версию программы")

	// Общие флаги
	addr := flag.String("addr", "localhost:8443", "Адрес сервера")
	cert := flag.String("cert", "server.crt", "Путь к сертификату")
	key := flag.String("key", "server.key", "Путь к приватному ключу")
	prometheus := flag.Bool("prometheus", false, "Включить Prometheus метрики")

	// Флаги для клиента
	connections := flag.Int("connections", 1, "Количество соединений")
	streams := flag.Int("streams", 1, "Количество потоков")
	packetSize := flag.Int("packet-size", 1024, "Размер пакета")
	rate := flag.Int("rate", 100, "Скорость отправки (пакетов/сек)")
	pattern := flag.String("pattern", "burst", "Паттерн отправки: burst, steady, random")

	// Флаги для MASQUE
	masqueServer := flag.String("masque-server", "localhost:8443", "MASQUE сервер для тестирования")
	masqueTargets := flag.String("masque-targets", "8.8.8.8:53,1.1.1.1:53", "Целевые хосты для CONNECT-UDP (через запятую)")

	// Флаги для ICE
	iceStunServers := flag.String("ice-stun", "stun.l.google.com:19302,stun1.l.google.com:19302", "STUN серверы (через запятую)")
	iceTurnServers := flag.String("ice-turn", "", "TURN серверы (через запятую)")
	iceTurnUser := flag.String("ice-turn-user", "", "TURN username")
	iceTurnPass := flag.String("ice-turn-pass", "", "TURN password")

	flag.Parse()

	// Обработка флага --version
	if *version {
		internal.PrintVersion()
		os.Exit(0)
	}

	config := map[string]interface{}{
		"addr":           *addr,
		"cert":           *cert,
		"key":            *key,
		"prometheus":     *prometheus,
		"connections":    *connections,
		"streams":        *streams,
		"packetSize":     *packetSize,
		"rate":           *rate,
		"pattern":        *pattern,
		"masqueServer":   *masqueServer,
		"masqueTargets":  *masqueTargets,
		"iceStunServers": *iceStunServers,
		"iceTurnServers": *iceTurnServers,
		"iceTurnUser":    *iceTurnUser,
		"iceTurnPass":    *iceTurnPass,
	}

	return *mode, config
}

// CreateLogger создает логгер
func CreateLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	return logger
}

// ShowHelp показывает справку
func ShowHelp() {
	// Показываем версию
	version, err := internal.GetVersion()
	if err == nil {
		fmt.Printf("2GC Network Protocol Suite v%s - Комплексное тестирование сетевых протоколов\n", version)
	} else {
		fmt.Println("2GC Network Protocol Suite - Комплексное тестирование сетевых протоколов")
	}
	fmt.Println()
	fmt.Println("Использование:")
	fmt.Println("  quic-test -mode=<режим> [флаги]")
	fmt.Println()
	fmt.Println("Режимы:")
	for name, cmd := range Commands {
		fmt.Printf("  %-10s - %s\n", name, cmd.Description)
	}
	fmt.Println()
	fmt.Println("Флаги:")
	flag.PrintDefaults()
}

// Вспомогательные функции для парсинга конфигурации
func getString(config map[string]interface{}, key string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return ""
}

func getInt(config map[string]interface{}, key string) int {
	if val, ok := config[key].(int); ok {
		return val
	}
	return 0
}

func getBool(config map[string]interface{}, key string) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return false
}

// runServer запускает сервер
func runServer(args []string) error {
	fmt.Println("Запуск в режиме сервера...")
	
	// Парсим конфигурацию из аргументов
	_, config := ParseFlags()
	
	// Создаем конфигурацию теста
	cfg := internal.TestConfig{
		Mode:         "server",
		Addr:         getString(config, "addr"),
		CertPath:     getString(config, "cert"),
		KeyPath:      getString(config, "key"),
		Prometheus:   getBool(config, "prometheus"),
	}
	
	// Запускаем сервер
	server.Run(cfg)
	return nil
}

// runClient запускает клиент
func runClient(args []string) error {
	fmt.Println("Запуск в режиме клиента...")
	
	// Парсим конфигурацию из аргументов
	_, config := ParseFlags()
	
	// Создаем конфигурацию теста
	cfg := internal.TestConfig{
		Mode:        "client",
		Addr:        getString(config, "addr"),
		Connections: getInt(config, "connections"),
		Streams:     getInt(config, "streams"),
		PacketSize:  getInt(config, "packetSize"),
		Rate:        getInt(config, "rate"),
		Pattern:     getString(config, "pattern"),
		Prometheus:  getBool(config, "prometheus"),
	}
	
	// Запускаем клиент
	client.Run(cfg)
	return nil
}

// runTest запускает тестирование
func runTest(args []string) error {
	fmt.Println("Запуск в режиме теста (сервер+клиент)...")
	
	// Парсим конфигурацию из аргументов
	_, config := ParseFlags()
	
	// Создаем конфигурацию теста
	cfg := internal.TestConfig{
		Mode:        "test",
		Addr:        getString(config, "addr"),
		Connections: getInt(config, "connections"),
		Streams:     getInt(config, "streams"),
		PacketSize:  getInt(config, "packetSize"),
		Rate:        getInt(config, "rate"),
		Pattern:     getString(config, "pattern"),
		Prometheus:  getBool(config, "prometheus"),
		Duration:    30 * time.Second, // По умолчанию 30 секунд
	}
	
	// Запускаем тест (сервер + клиент)
	// TODO: реализовать одновременный запуск сервера и клиента
	// Пока запускаем только клиент, предполагая что сервер уже работает
	client.Run(cfg)
	return nil
}

// runDashboard запускает dashboard
func runDashboard(args []string) error {
	fmt.Println("Starting QUIC Testing Dashboard on http://localhost:9990")
	fmt.Println("Open your browser and navigate to http://localhost:9990")
	fmt.Println("🛑 Press Ctrl+C to stop the server")
	fmt.Println("🔍 Advanced analysis features available at:")
	fmt.Println("   - /api/analysis/deep - Deep protocol analysis")
	fmt.Println("   - /api/analysis/protocol - Protocol analysis")
	fmt.Println("   - /api/network/simulation - Network simulation")

	// Инициализируем QUIC менеджер
	logger := CreateLogger()
	quicConfig := &quic.QUICManagerConfig{
		ServerAddr:     ":9001", // Уникальный порт для QUIC
		MaxConnections: 10,
		MaxStreams:     100,
		ConnectTimeout: 30 * time.Second,
		IdleTimeout:    60 * time.Second,
	}
	quicManager = quic.NewQUICManager(logger, quicConfig)

	// Запускаем HTTP сервер
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
		} else {
			http.ServeFile(w, r, r.URL.Path[1:])
		}
	})

	// API endpoints
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if quicManager != nil {
			json.NewEncoder(w).Encode(quicManager.GetStatus())
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"server": map[string]interface{}{
					"running": false,
				},
				"client": map[string]interface{}{
					"running": false,
				},
			})
		}
	})

	// Metrics endpoint
	http.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metricsManager.UpdateMetrics()
		json.NewEncoder(w).Encode(metricsManager.GetMetrics())
	})

	// Config endpoint
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"server": map[string]interface{}{
				"addr": ":9001", // QUIC Testing Tool сервер (уникальный порт)
				"cert": "server.crt",
				"key":  "server.key",
			},
			"client": map[string]interface{}{
				"addr":        "localhost:9001", // QUIC Testing Tool клиент (уникальный порт)
				"connections": 1,
				"streams":     1,
				"packetSize":  1200,
				"rate":        100,
				"pattern":     "random",
			},
		})
	})

	// QUIC Server API endpoints
	http.HandleFunc("/api/server/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if quicManager != nil {
			err := quicManager.StartServer()
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "started",
				"message": "QUIC server started on port 9001",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "QUIC manager not initialized",
			})
		}
	})

	http.HandleFunc("/api/server/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if quicManager != nil {
			err := quicManager.StopServer()
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "stopped",
				"message": "QUIC server stopped",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "QUIC manager not initialized",
			})
		}
	})

	// QUIC Client API endpoints
	http.HandleFunc("/api/client/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if quicManager != nil {
			err := quicManager.StartClient()
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "started",
				"message": "QUIC client connected to localhost:9001",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "QUIC manager not initialized",
			})
		}
	})

	http.HandleFunc("/api/client/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if quicManager != nil {
			err := quicManager.StopClient()
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "stopped",
				"message": "QUIC client disconnected",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "QUIC manager not initialized",
			})
		}
	})

	// QUIC Test API endpoint
	http.HandleFunc("/api/test/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if quicManager != nil {
			// Парсим параметры теста из запроса
			var testParams struct {
				PacketSize  int `json:"packet_size"`
				PacketCount int `json:"packet_count"`
				Duration    int `json:"duration"` // в секундах
			}

			if err := json.NewDecoder(r.Body).Decode(&testParams); err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": "Invalid test parameters",
				})
				return
			}

			// Устанавливаем значения по умолчанию
			if testParams.PacketSize == 0 {
				testParams.PacketSize = 1200
			}
			if testParams.PacketCount == 0 {
				testParams.PacketCount = 10
			}
			if testParams.Duration == 0 {
				testParams.Duration = 30
			}

			// Создаем конфигурацию теста
			testConfig := &quic.TestConfig{
				PacketSize:  testParams.PacketSize,
				PacketCount: testParams.PacketCount,
				Duration:    time.Duration(testParams.Duration) * time.Second,
			}

			// Запускаем тест
			ctx := context.Background()
			err := quicManager.RunTest(ctx, testConfig)
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "started",
				"message": "QUIC test started",
				"config":  testConfig,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "QUIC manager not initialized",
			})
		}
	})

	// MASQUE API endpoints
	http.HandleFunc("/api/masque/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metricsManager.SetMASQUEActive(true)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "started",
			"message": "MASQUE testing started",
		})
	})

	http.HandleFunc("/api/masque/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metricsManager.SetMASQUEActive(false)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "stopped",
			"message": "MASQUE testing stopped",
		})
	})

	// ICE API endpoints
	http.HandleFunc("/api/ice/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metricsManager.SetICEActive(true)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "started",
			"message": "ICE testing started",
		})
	})

	http.HandleFunc("/api/ice/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metricsManager.SetICEActive(false)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "stopped",
			"message": "ICE testing stopped",
		})
	})

	// History endpoint для графиков
	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metricsManager.GetHistory())
	})

	// Запускаем сервер
	log.Fatal(http.ListenAndServe(":9990", nil))
	return nil
}

// runMASQUE запускает MASQUE тестирование
func runMASQUE(args []string) error {
	fmt.Println("🔥 Запуск MASQUE тестирования...")
	// Импортируем и вызываем функцию из cmd/masque
	// TODO: реализовать запуск MASQUE тестирования
	return nil
}

// runICE запускает ICE тестирование
func runICE(args []string) error {
	fmt.Println("🧊 Запуск ICE/STUN/TURN тестирования...")
	// Импортируем и вызываем функцию из cmd/ice
	// TODO: реализовать запуск ICE тестирования
	return nil
}

// runEnhanced запускает расширенное тестирование
func runEnhanced(args []string) error {
	fmt.Println("Запуск расширенного тестирования (MASQUE + ICE + QUIC)...")
	// Импортируем и вызываем функцию из cmd/enhanced
	// TODO: реализовать запуск расширенного тестирования
	return nil
}
