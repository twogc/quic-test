package internal

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSSEManager(t *testing.T) {
	manager := NewSSEManager()
	
	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}
	
	if manager.clients == nil {
		t.Fatal("Expected clients map to be initialized")
	}
	
	if len(manager.clients) != 0 {
		t.Errorf("Expected 0 clients initially, got %d", len(manager.clients))
	}
}

func TestAddClient(t *testing.T) {
	manager := NewSSEManager()
	
	client := manager.AddClient("test-client")
	
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	
	if client.ID != "test-client" {
		t.Errorf("Expected client ID 'test-client', got '%s'", client.ID)
	}
	
	if client.Messages == nil {
		t.Fatal("Expected messages channel to be initialized")
	}
	
	if client.Done == nil {
		t.Fatal("Expected done channel to be initialized")
	}
	
	if len(manager.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(manager.clients))
	}
}

func TestRemoveClient(t *testing.T) {
	manager := NewSSEManager()
	
	client := manager.AddClient("test-client")
	
	// Проверяем, что клиент добавлен
	if len(manager.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(manager.clients))
	}
	
	manager.RemoveClient("test-client")
	
	// Проверяем, что клиент удален
	if len(manager.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(manager.clients))
	}
	
	// Проверяем, что каналы закрыты
	select {
	case <-client.Messages:
		// Канал закрыт, это ожидаемо
	default:
		t.Error("Expected messages channel to be closed")
	}
	
	select {
	case <-client.Done:
		// Канал закрыт, это ожидаемо
	default:
		t.Error("Expected done channel to be closed")
	}
}

func TestBroadcast(t *testing.T) {
	manager := NewSSEManager()
	
	// Добавляем клиентов
	client1 := manager.AddClient("client1")
	client2 := manager.AddClient("client2")
	
	// Тестовые данные
	data := map[string]interface{}{
		"type": "test",
		"data": "test message",
	}
	
	// Отправляем broadcast
	manager.Broadcast(data)
	
	// Проверяем, что сообщения дошли до клиентов
	select {
	case msg1 := <-client1.Messages:
		var received map[string]interface{}
		err := json.Unmarshal(msg1, &received)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		
		if received["type"] != "test" {
			t.Errorf("Expected type 'test', got %s", received["type"])
		}
		
		if received["data"] != "test message" {
			t.Errorf("Expected data 'test message', got %s", received["data"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected message to be received by client1")
	}
	
	select {
	case msg2 := <-client2.Messages:
		var received map[string]interface{}
		err := json.Unmarshal(msg2, &received)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		
		if received["type"] != "test" {
			t.Errorf("Expected type 'test', got %s", received["type"])
		}
		
		if received["data"] != "test message" {
			t.Errorf("Expected data 'test message', got %s", received["data"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected message to be received by client2")
	}
}

func TestBroadcastMetrics(t *testing.T) {
	manager := NewSSEManager()
	
	client := manager.AddClient("test-client")
	
	metrics := map[string]interface{}{
		"Success": true,
		"Errors":  5,
		"BytesSent": int64(1024000),
	}
	
	manager.BroadcastMetrics(metrics)
	
	select {
	case msg := <-client.Messages:
		var received map[string]interface{}
		err := json.Unmarshal(msg, &received)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		
		if received["type"] != "metrics" {
			t.Errorf("Expected type 'metrics', got %s", received["type"])
		}
		
		if received["data"] == nil {
			t.Error("Expected data field to be present")
		}
		
		if received["timestamp"] == nil {
			t.Error("Expected timestamp field to be present")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected metrics message to be received")
	}
}

func TestBroadcastStatus(t *testing.T) {
	manager := NewSSEManager()
	
	client := manager.AddClient("test-client")
	
	manager.BroadcastStatus(true, false)
	
	select {
	case msg := <-client.Messages:
		var received map[string]interface{}
		err := json.Unmarshal(msg, &received)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		
		if received["type"] != "status" {
			t.Errorf("Expected type 'status', got %s", received["type"])
		}
		
		if received["server_running"] != true {
			t.Errorf("Expected server_running true, got %v", received["server_running"])
		}
		
		if received["client_running"] != false {
			t.Errorf("Expected client_running false, got %v", received["client_running"])
		}
		
		if received["timestamp"] == nil {
			t.Error("Expected timestamp field to be present")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected status message to be received")
	}
}

func TestBroadcastTestConfig(t *testing.T) {
	manager := NewSSEManager()
	
	client := manager.AddClient("test-client")
	
	config := TestConfig{
		Mode:        "test",
		Addr:        ":9000",
		Connections: 2,
		Streams:     4,
	}
	
	manager.BroadcastTestConfig(config)
	
	select {
	case msg := <-client.Messages:
		var received map[string]interface{}
		err := json.Unmarshal(msg, &received)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		
		if received["type"] != "test_config" {
			t.Errorf("Expected type 'test_config', got %s", received["type"])
		}
		
		if received["data"] == nil {
			t.Error("Expected data field to be present")
		}
		
		if received["timestamp"] == nil {
			t.Error("Expected timestamp field to be present")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected test config message to be received")
	}
}

func TestBroadcastError(t *testing.T) {
	manager := NewSSEManager()
	
	client := manager.AddClient("test-client")
	
	err := &TestError{Message: "test error"}
	manager.BroadcastError(err)
	
	select {
	case msg := <-client.Messages:
		var received map[string]interface{}
		err := json.Unmarshal(msg, &received)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		
		if received["type"] != "error" {
			t.Errorf("Expected type 'error', got %s", received["type"])
		}
		
		if received["message"] != "test error" {
			t.Errorf("Expected message 'test error', got %s", received["message"])
		}
		
		if received["timestamp"] == nil {
			t.Error("Expected timestamp field to be present")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected error message to be received")
	}
}

func TestGetClientCount(t *testing.T) {
	manager := NewSSEManager()
	
	if manager.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients initially, got %d", manager.GetClientCount())
	}
	
	manager.AddClient("client1")
	manager.AddClient("client2")
	
	if manager.GetClientCount() != 2 {
		t.Errorf("Expected 2 clients, got %d", manager.GetClientCount())
	}
	
	manager.RemoveClient("client1")
	
	if manager.GetClientCount() != 1 {
		t.Errorf("Expected 1 client, got %d", manager.GetClientCount())
	}
}

func TestGetClients(t *testing.T) {
	manager := NewSSEManager()
	
	clients := manager.GetClients()
	if len(clients) != 0 {
		t.Errorf("Expected 0 clients initially, got %d", len(clients))
	}
	
	manager.AddClient("client1")
	manager.AddClient("client2")
	
	clients = manager.GetClients()
	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}
	
	// Проверяем, что оба клиента присутствуют
	found1 := false
	found2 := false
	for _, client := range clients {
		if client == "client1" {
			found1 = true
		}
		if client == "client2" {
			found2 = true
		}
	}
	
	if !found1 {
		t.Error("Expected client1 to be in clients list")
	}
	
	if !found2 {
		t.Error("Expected client2 to be in clients list")
	}
}

func TestSSEServerHandler(t *testing.T) {
	manager := NewSSEManager()
	
	req := httptest.NewRequest("GET", "/api/events", nil)
	w := httptest.NewRecorder()
	
	// Запускаем handler в отдельной goroutine
	go func() {
		manager.SSEServerHandler(w, req)
	}()
	
	// Ждем немного, чтобы handler успел запуститься
	time.Sleep(50 * time.Millisecond)
	
	// Проверяем заголовки
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", w.Header().Get("Content-Type"))
	}
	
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", w.Header().Get("Cache-Control"))
	}
	
	if w.Header().Get("Connection") != "keep-alive" {
		t.Errorf("Expected Connection 'keep-alive', got '%s'", w.Header().Get("Connection"))
	}
	
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
	
	// Проверяем, что клиент добавлен
	if manager.GetClientCount() != 1 {
		t.Errorf("Expected 1 client, got %d", manager.GetClientCount())
	}
}

// Вспомогательный тип для тестирования ошибок
type TestError struct {
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}
