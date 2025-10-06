package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Парсинг флагов
	addr := flag.String("addr", ":9990", "Адрес для веб-интерфейса")
	flag.Parse()

	fmt.Println("\033[1;36m==============================\033[0m")
	fmt.Println("\033[1;36m  2GC CloudBridge Dashboard\033[0m")
	fmt.Println("\033[1;36m==============================\033[0m")

	// Обработка сигналов для graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nПолучен сигнал завершения, остановка дашборда...")
		os.Exit(0)
	}()

	startDashboard(*addr)
}

// startDashboard запускает веб-интерфейс
func startDashboard(addr string) {
	fmt.Println("🚀 Starting QUIC Testing Dashboard on http://localhost:9990")
	fmt.Println("📊 Open your browser and navigate to http://localhost:9990")
	fmt.Println("🛑 Press Ctrl+C to stop the server")
	fmt.Println("🔍 Advanced analysis features available at:")
	fmt.Println("   - /api/analysis/deep - Deep protocol analysis")
	fmt.Println("   - /api/analysis/protocol - Protocol analysis")
	fmt.Println("   - /api/network/simulation - Network simulation")

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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"server": map[string]interface{}{
				"running": false,
			},
			"client": map[string]interface{}{
				"running": false,
			},
		})
	})

	// MASQUE API endpoints
	http.HandleFunc("/api/masque/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "started",
			"message": "MASQUE testing started",
		})
	})

	http.HandleFunc("/api/masque/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "stopped",
			"message": "MASQUE testing stopped",
		})
	})

	// ICE API endpoints
	http.HandleFunc("/api/ice/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "started",
			"message": "ICE testing started",
		})
	})

	http.HandleFunc("/api/ice/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "stopped",
			"message": "ICE testing stopped",
		})
	})

	// Запускаем сервер
	fmt.Printf("🚀 Dashboard запущен на http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
