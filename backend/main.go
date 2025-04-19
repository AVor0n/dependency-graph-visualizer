package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/avor0n/dependency-graph-visualizer/handlers"
	"github.com/avor0n/dependency-graph-visualizer/services"
	"github.com/avor0n/dependency-graph-visualizer/utils"
)

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
	projectPath, err := filepath.Abs(*projectPathPtr)
	if err != nil {
		fmt.Println("Ошибка при получении абсолютного пути:", err)
		os.Exit(1)
	}

	fmt.Printf("Запуск визуализатора для проекта: %s\n", projectPath)

	// Загружаем правила .gitignore
	gitIgnore := utils.LoadGitIgnore(projectPath)

	// Инициализируем сервисы
	fileService := services.NewFileService(projectPath, gitIgnore)
	dependencyService := services.NewDependencyService(fileService)

	// Анализируем зависимости перед запуском сервера
	fmt.Println("Анализ зависимостей в проекте...")
	dependencyService.BuildDependencyGraph()

	// Инициализируем обработчики с указателями на сервисы
	handler := &handlers.Handler{
		FileService:       fileService,
		DependencyService: dependencyService,
		ProjectPath:       projectPath,
	}

	// Регистрируем API endpoints
	http.HandleFunc("/api/project-info", handler.HandleProjectInfo)
	http.HandleFunc("/api/file-tree", handler.HandleFileTree)
	http.HandleFunc("/api/dependency-graph", handler.HandleDependencyGraph)
	http.HandleFunc("/api/file-dependencies", handler.HandleFileDependencies)

	// Указываем статическую директорию для фронтенда
	fs := http.FileServer(http.Dir("../frontend/dist"))
	http.Handle("/", handlers.EnableCORS(fs))

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
