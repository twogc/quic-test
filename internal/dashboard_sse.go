package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SSEClient представляет клиента SSE
type SSEClient struct {
	ID       string
	Messages chan []byte
	Done     chan bool
}

// SSEManager управляет SSE соединениями
type SSEManager struct {
	mu      sync.RWMutex
	clients map[string]*SSEClient
}

// NewSSEManager создает новый менеджер SSE
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[string]*SSEClient),
	}
}

// AddClient добавляет нового клиента SSE
func (m *SSEManager) AddClient(id string) *SSEClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	client := &SSEClient{
		ID:       id,
		Messages: make(chan []byte, 100),
		Done:     make(chan bool),
	}
	
	m.clients[id] = client
	return client
}

// RemoveClient удаляет клиента SSE
func (m *SSEManager) RemoveClient(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if client, exists := m.clients[id]; exists {
		close(client.Messages)
		close(client.Done)
		delete(m.clients, id)
	}
}

// Broadcast отправляет сообщение всем клиентам
func (m *SSEManager) Broadcast(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, client := range m.clients {
		select {
		case client.Messages <- jsonData:
		default:
			// Клиент не готов к получению сообщений
		}
	}
}

// SSEServerHandler обрабатывает SSE соединения
func (m *SSEManager) SSEServerHandler(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовки для SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")
	
	// Создаем уникальный ID для клиента
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	client := m.AddClient(clientID)
	
	// Уведомляем о подключении
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"client_id\":\"%s\"}\n\n", clientID)
	w.(http.Flusher).Flush()
	
	// Отправляем heartbeat каждые 30 секунд
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case message := <-client.Messages:
			fmt.Fprintf(w, "data: %s\n\n", message)
			w.(http.Flusher).Flush()
			
		case <-ticker.C:
			// Отправляем heartbeat
			heartbeat := map[string]interface{}{
				"type":      "heartbeat",
				"timestamp": time.Now().Unix(),
			}
			jsonData, _ := json.Marshal(heartbeat)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			w.(http.Flusher).Flush()
			
		case <-client.Done:
			return
			
		case <-r.Context().Done():
			m.RemoveClient(clientID)
			return
		}
	}
}

// BroadcastMetrics отправляет метрики всем подключенным клиентам
func (m *SSEManager) BroadcastMetrics(metrics map[string]interface{}) {
	message := map[string]interface{}{
		"type":      "metrics",
		"data":      metrics,
		"timestamp": time.Now().Unix(),
	}
	m.Broadcast(message)
}

// BroadcastStatus отправляет статус всем подключенным клиентам
func (m *SSEManager) BroadcastStatus(serverRunning, clientRunning bool) {
	message := map[string]interface{}{
		"type":            "status",
		"server_running": serverRunning,
		"client_running":   clientRunning,
		"timestamp":       time.Now().Unix(),
	}
	m.Broadcast(message)
}

// BroadcastTestConfig отправляет конфигурацию теста всем подключенным клиентам
func (m *SSEManager) BroadcastTestConfig(config TestConfig) {
	message := map[string]interface{}{
		"type":      "test_config",
		"data":      config,
		"timestamp": time.Now().Unix(),
	}
	m.Broadcast(message)
}

// BroadcastError отправляет ошибку всем подключенным клиентам
func (m *SSEManager) BroadcastError(err error) {
	message := map[string]interface{}{
		"type":      "error",
		"message":   err.Error(),
		"timestamp": time.Now().Unix(),
	}
	m.Broadcast(message)
}

// GetClientCount возвращает количество подключенных клиентов
func (m *SSEManager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return len(m.clients)
}

// GetClients возвращает список подключенных клиентов
func (m *SSEManager) GetClients() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	clients := make([]string, 0, len(m.clients))
	for id := range m.clients {
		clients = append(clients, id)
	}
	return clients
}
