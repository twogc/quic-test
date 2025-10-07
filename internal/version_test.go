package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Создаем временный файл tag.txt
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Переходим в временную директорию
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Создаем файл tag.txt с версией
	tagFile := filepath.Join(tempDir, "tag.txt")
	err = os.WriteFile(tagFile, []byte("v1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create tag.txt: %v", err)
	}

	// Тестируем чтение версии
	version, err := GetVersion()
	if err != nil {
		t.Errorf("GetVersion() failed: %v", err)
	}
	if version != "v1.2.3" {
		t.Errorf("Expected version 'v1.2.3', got '%s'", version)
	}
}

func TestGetVersionEmptyFile(t *testing.T) {
	// Создаем временный файл tag.txt
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Переходим в временную директорию
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Создаем пустой файл tag.txt
	tagFile := filepath.Join(tempDir, "tag.txt")
	err = os.WriteFile(tagFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create tag.txt: %v", err)
	}

	// Тестируем чтение пустой версии
	version, err := GetVersion()
	if err == nil {
		t.Error("Expected error for empty tag.txt, got nil")
	}
	if version != "" {
		t.Errorf("Expected empty version, got '%s'", version)
	}
}

func TestGetVersionNotFound(t *testing.T) {
	// Создаем временную директорию без tag.txt
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Переходим в временную директорию
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Тестируем чтение версии без файла
	version, err := GetVersion()
	if err != nil {
		t.Errorf("GetVersion() failed: %v", err)
	}
	if version != "unknown" {
		t.Errorf("Expected version 'unknown', got '%s'", version)
	}
}

func TestGetVersionInfo(t *testing.T) {
	// Создаем временный файл tag.txt
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Переходим в временную директорию
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Создаем файл tag.txt с версией
	tagFile := filepath.Join(tempDir, "tag.txt")
	err = os.WriteFile(tagFile, []byte("2.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create tag.txt: %v", err)
	}

	// Тестируем получение информации о версии
	versionInfo := GetVersionInfo()
	expected := "QUIC Testing Tool v2.0.0"
	if versionInfo != expected {
		t.Errorf("Expected '%s', got '%s'", expected, versionInfo)
	}
}
