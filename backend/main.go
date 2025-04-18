package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

// Глобальная переменная для хранения пути к проекту
var projectPath string

// Middleware для CORS
func enableCors(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func main() {
	// Определяем флаг командной строки для пути к проекту
	projectPathPtr := flag.String("path", "", "Путь к JavaScript/TypeScript проекту")
	flag.Parse()

	if *projectPathPtr == "" {
		fmt.Println("Ошибка: необходимо указать путь к проекту с помощью флага -path")
		fmt.Println("Пример: ./dependency-graph-visualizer -path /path/to/js/project")
		os.Exit(1)
	}

	// Проверяем, существует ли директория
	fileInfo, err := os.Stat(*projectPathPtr)
	if err != nil {
		fmt.Println("Ошибка: указанный путь не существует")
		os.Exit(1)
	}

	if !fileInfo.IsDir() {
		fmt.Println("Ошибка: указанный путь не является директорией")
		os.Exit(1)
	}

	// Сохраняем абсолютный путь к проекту
	projectPath, err = filepath.Abs(*projectPathPtr)
	if err != nil {
		fmt.Println("Ошибка при получении абсолютного пути:", err)
		os.Exit(1)
	}

	fmt.Printf("Запуск визуализатора для проекта: %s\n", projectPath)

	// API endpoints
	http.HandleFunc("/api/project-info", handleProjectInfo)
	http.HandleFunc("/api/file-tree", handleFileTree)

	// Указываем статическую директорию для фронтенда
	fs := http.FileServer(http.Dir("../frontend/dist"))
	http.Handle("/", enableCors(fs))

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Обработчик для получения информации о проекте
func handleProjectInfo(w http.ResponseWriter, r *http.Request) {
	response := struct {
		ProjectPath string `json:"projectPath"`
		ProjectName string `json:"projectName"`
	}{
		ProjectPath: projectPath,
		ProjectName: filepath.Base(projectPath),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// Обработчик для получения дерева файлов
func handleFileTree(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Получаем структуру директории
	rootNode := scanDirectory(projectPath, "")

	json.NewEncoder(w).Encode(rootNode)
}

// Рекурсивно сканирует директорию и возвращает структуру
func scanDirectory(rootPath string, relativePath string) FileNode {
	absPath := filepath.Join(rootPath, relativePath)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		return FileNode{}
	}

	nodeName := fileInfo.Name()
	if relativePath == "" {
		// Если это корневая директория, используем последнюю часть пути
		nodeName = filepath.Base(rootPath)
	}

	node := FileNode{
		Name:  nodeName,
		Path:  relativePath,
		IsDir: fileInfo.IsDir(),
	}

	if !fileInfo.IsDir() {
		return node
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return node
	}

	for _, entry := range entries {
		childRelPath := filepath.Join(relativePath, entry.Name())
		childNode := scanDirectory(rootPath, childRelPath)
		node.Children = append(node.Children, childNode)
	}

	return node
}
