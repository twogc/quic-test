package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GetVersion читает версию из файла tag.txt
func GetVersion() (string, error) {
	// Ищем файл tag.txt в текущей директории и в родительских директориях
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Проверяем текущую директорию и родительские директории
	for {
		tagPath := filepath.Join(dir, "tag.txt")
		if _, err := os.Stat(tagPath); err == nil {
			// Файл найден, читаем его
			content, err := ioutil.ReadFile(tagPath)
			if err != nil {
				return "", fmt.Errorf("failed to read tag.txt: %w", err)
			}
			
			version := strings.TrimSpace(string(content))
			if version == "" {
				return "", fmt.Errorf("tag.txt is empty")
			}
			
			return version, nil
		}
		
		// Переходим к родительской директории
		parent := filepath.Dir(dir)
		if parent == dir {
			// Достигли корня файловой системы
			break
		}
		dir = parent
	}
	
	// Если файл не найден, возвращаем версию по умолчанию
	return "unknown", nil
}

// GetVersionInfo возвращает полную информацию о версии
func GetVersionInfo() string {
	version, err := GetVersion()
	if err != nil {
		return fmt.Sprintf("QUIC Testing Tool (version: unknown, error: %v)", err)
	}
	
	// Если версия уже начинается с "v", не добавляем еще один
	if len(version) > 0 && version[0] == 'v' {
		return fmt.Sprintf("QUIC Testing Tool %s", version)
	}
	
	return fmt.Sprintf("QUIC Testing Tool v%s", version)
}

// PrintVersion выводит информацию о версии
func PrintVersion() {
	fmt.Println(GetVersionInfo())
}
