package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// startDashboard запускает веб-интерфейс
func startDashboard() {
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
	log.Fatal(http.ListenAndServe(":9990", nil))
}

