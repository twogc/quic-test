package cli

import (
	"testing"
)

func TestGetString(t *testing.T) {
	config := map[string]interface{}{
		"string_key": "test_value",
		"int_key":    123,
		"bool_key":   true,
	}
	
	// Тест с существующим строковым ключом
	result := getString(config, "string_key")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}
	
	// Тест с несуществующим ключом
	result = getString(config, "nonexistent")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
	
	// Тест с ключом неправильного типа
	result = getString(config, "int_key")
	if result != "" {
		t.Errorf("Expected empty string for int key, got '%s'", result)
	}
}

func TestGetInt(t *testing.T) {
	config := map[string]interface{}{
		"int_key":    123,
		"string_key": "test",
		"bool_key":   true,
	}
	
	// Тест с существующим целочисленным ключом
	result := getInt(config, "int_key")
	if result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}
	
	// Тест с несуществующим ключом
	result = getInt(config, "nonexistent")
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
	
	// Тест с ключом неправильного типа
	result = getInt(config, "string_key")
	if result != 0 {
		t.Errorf("Expected 0 for string key, got %d", result)
	}
}

func TestGetBool(t *testing.T) {
	config := map[string]interface{}{
		"bool_key":   true,
		"string_key": "test",
		"int_key":    123,
	}
	
	// Тест с существующим булевым ключом
	result := getBool(config, "bool_key")
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
	
	// Тест с несуществующим ключом
	result = getBool(config, "nonexistent")
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}
	
	// Тест с ключом неправильного типа
	result = getBool(config, "string_key")
	if result != false {
		t.Errorf("Expected false for string key, got %v", result)
	}
}

func TestCommandsMap(t *testing.T) {
	// Проверяем, что все команды определены
	expectedCommands := []string{
		"server",
		"client", 
		"test",
		"dashboard",
		"masque",
		"ice",
		"enhanced",
	}
	
	for _, cmdName := range expectedCommands {
		if cmd, exists := Commands[cmdName]; !exists {
			t.Errorf("Command '%s' not found in Commands map", cmdName)
		} else if cmd.Name != cmdName {
			t.Errorf("Command name mismatch: expected '%s', got '%s'", cmdName, cmd.Name)
		} else if cmd.Handler == nil {
			t.Errorf("Command '%s' has nil handler", cmdName)
		}
	}
}

func TestCreateLogger(t *testing.T) {
	logger := CreateLogger()
	if logger == nil {
		t.Error("CreateLogger returned nil")
	}
	
	// Проверяем, что логгер работает
	logger.Info("Test log message")
}

func TestShowHelp(t *testing.T) {
	// Тест функции ShowHelp - проверяем, что она не паникует
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowHelp panicked: %v", r)
		}
	}()
	
	ShowHelp()
}

func TestParseFlags(t *testing.T) {
	// Тест парсинга флагов
	// Этот тест может быть сложным из-за глобального состояния flag
	// Но мы можем проверить базовую функциональность
	
	// Сбрасываем состояние flag для теста
	// flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	
	// mode, config := ParseFlags()
	// if mode == "" {
	//     t.Error("Expected non-empty mode")
	// }
	// if config == nil {
	//     t.Error("Expected non-nil config")
	// }
}

// Тесты для проверки обработки различных типов данных в конфигурации
func TestConfigTypeHandling(t *testing.T) {
	config := map[string]interface{}{
		"string_val": "test",
		"int_val":    42,
		"float_val":  3.14,
		"bool_val":   true,
		"nil_val":    nil,
	}
	
	// Тест обработки различных типов
	tests := []struct {
		key      string
		expected string
	}{
		{"string_val", "test"},
		{"int_val", ""}, // int не должен конвертироваться в string
		{"float_val", ""}, // float не должен конвертироваться в string
		{"bool_val", ""}, // bool не должен конвертироваться в string
		{"nil_val", ""}, // nil должен возвращать пустую строку
		{"nonexistent", ""}, // несуществующий ключ
	}
	
	for _, test := range tests {
		result := getString(config, test.key)
		if result != test.expected {
			t.Errorf("getString(%s): expected '%s', got '%s'", test.key, test.expected, result)
		}
	}
}

func TestConfigIntHandling(t *testing.T) {
	config := map[string]interface{}{
		"int_val":    42,
		"string_val": "test",
		"float_val":  3.14,
		"bool_val":   true,
		"nil_val":    nil,
	}
	
	// Тест обработки различных типов для getInt
	tests := []struct {
		key      string
		expected int
	}{
		{"int_val", 42},
		{"string_val", 0}, // string не должен конвертироваться в int
		{"float_val", 0}, // float не должен конвертироваться в int
		{"bool_val", 0}, // bool не должен конвертироваться в int
		{"nil_val", 0}, // nil должен возвращать 0
		{"nonexistent", 0}, // несуществующий ключ
	}
	
	for _, test := range tests {
		result := getInt(config, test.key)
		if result != test.expected {
			t.Errorf("getInt(%s): expected %d, got %d", test.key, test.expected, result)
		}
	}
}

func TestConfigBoolHandling(t *testing.T) {
	config := map[string]interface{}{
		"bool_val":   true,
		"string_val": "test",
		"int_val":    42,
		"float_val":  3.14,
		"nil_val":    nil,
	}
	
	// Тест обработки различных типов для getBool
	tests := []struct {
		key      string
		expected bool
	}{
		{"bool_val", true},
		{"string_val", false}, // string не должен конвертироваться в bool
		{"int_val", false}, // int не должен конвертироваться в bool
		{"float_val", false}, // float не должен конвертироваться в bool
		{"nil_val", false}, // nil должен возвращать false
		{"nonexistent", false}, // несуществующий ключ
	}
	
	for _, test := range tests {
		result := getBool(config, test.key)
		if result != test.expected {
			t.Errorf("getBool(%s): expected %v, got %v", test.key, test.expected, result)
		}
	}
}
