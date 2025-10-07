package integration

import (
	"testing"

	"go.uber.org/zap"
)

func TestSimpleIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Создаем интеграцию
	integration := NewSimpleIntegration(logger, "bbrv2")
	
	// Инициализируем
	err := integration.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize integration: %v", err)
	}
	
	// Проверяем, что интеграция активна
	if !integration.IsActive() {
		t.Error("Integration should be active")
	}
	
	// Тестируем остановку
	err = integration.Stop()
	if err != nil {
		t.Fatalf("Failed to stop integration: %v", err)
	}
	
	// Проверяем, что интеграция неактивна
	if integration.IsActive() {
		t.Error("Integration should not be active after stop")
	}
}

func TestSimpleIntegrationStartStop(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Создаем интеграцию
	integration := NewSimpleIntegration(logger, "cubic")
	
	// Инициализируем
	err := integration.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize integration: %v", err)
	}
	
	// Проверяем, что интеграция активна после инициализации
	if !integration.IsActive() {
		t.Error("Integration should be active after initialization")
	}
	
	// Останавливаем
	err = integration.Stop()
	if err != nil {
		t.Fatalf("Failed to stop integration: %v", err)
	}
	
	// Проверяем, что интеграция неактивна
	if integration.IsActive() {
		t.Error("Integration should not be active after stop")
	}
}
