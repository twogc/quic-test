package internal

import (
	"net/http"
	"path/filepath"
	"strings"
)

// StaticFileSystem предоставляет доступ к встроенным статическим файлам
type StaticFileSystem struct {
	httpFS http.FileSystem
}

// NewStaticFileSystem создает новую файловую систему для статических файлов
func NewStaticFileSystem() *StaticFileSystem {
	return &StaticFileSystem{
		httpFS: http.Dir("static"),
	}
}

// Open открывает файл
func (sfs *StaticFileSystem) Open(name string) (http.File, error) {
	return sfs.httpFS.Open(name)
}

// ServeStatic обрабатывает запросы к статическим файлам
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Убираем префикс /static/ если он есть
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == "" {
		path = "index.html"
	}
	
	// Проверяем расширение файла для правильного MIME типа
	ext := filepath.Ext(path)
	var contentType string
	
	switch ext {
	case ".html":
		contentType = "text/html; charset=utf-8"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".json":
		contentType = "application/json"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	case ".ico":
		contentType = "image/x-icon"
	case ".woff":
		contentType = "font/woff"
	case ".woff2":
		contentType = "font/woff2"
	case ".ttf":
		contentType = "font/ttf"
	case ".eot":
		contentType = "application/vnd.ms-fontobject"
	default:
		contentType = "application/octet-stream"
	}
	
	w.Header().Set("Content-Type", contentType)
	
	// Открываем файл
	file, err := NewStaticFileSystem().Open(path)
	if err != nil {
		// Если файл не найден, возвращаем index.html для SPA
		if path != "index.html" {
			file, err = NewStaticFileSystem().Open("index.html")
			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}
	defer file.Close()
	
	// Копируем содержимое файла в ответ
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "File stat error", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, path, stat.ModTime(), file)
}

// GetStaticFileList возвращает список доступных статических файлов
func GetStaticFileList() ([]string, error) {
	// Простая реализация - возвращаем основные файлы
	return []string{"index.html", "css/style.css", "js/app.js"}, nil
}

// StaticFileHandler создает обработчик для статических файлов
func StaticFileHandler() http.Handler {
	return http.HandlerFunc(ServeStatic)
}
