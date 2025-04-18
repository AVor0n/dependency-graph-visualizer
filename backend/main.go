package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

// Определение структуры для хранения константы
type Constant struct {
	Name     string `json:"name"`     // Имя константы
	Value    string `json:"value"`    // Значение константы
	Type     string `json:"type"`     // Тип константы
	FilePath string `json:"filePath"` // Путь к файлу, где объявлена константа
}

// Определение зависимости между константами
type Dependency struct {
	Source string `json:"source"` // Имя исходной константы
	Target string `json:"target"` // Имя целевой константы
}

// Структура для хранения графа зависимостей
type DependencyGraph struct {
	Nodes []Constant   `json:"nodes"` // Узлы графа (константы)
	Edges []Dependency `json:"edges"` // Ребра графа (зависимости)
}

// Глобальные переменные
var (
	projectPath     string
	dependencyGraph DependencyGraph
	graphMutex      sync.RWMutex    // Мьютекс для безопасной записи в граф
	constantMap     map[string]bool // Карта для быстрой проверки существования констант
)

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

	// Инициализируем карту констант
	constantMap = make(map[string]bool)

	// Анализируем зависимости перед запуском сервера
	fmt.Println("Анализ зависимостей в проекте...")
	buildDependencyGraph()

	// API endpoints
	http.HandleFunc("/api/project-info", handleProjectInfo)
	http.HandleFunc("/api/file-tree", handleFileTree)
	http.HandleFunc("/api/dependency-graph", handleDependencyGraph)
	http.HandleFunc("/api/file-dependencies", handleFileDependencies)

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

// Обработчик для получения полного графа зависимостей
func handleDependencyGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Блокировка для чтения графа
	graphMutex.RLock()
	defer graphMutex.RUnlock()

	json.NewEncoder(w).Encode(dependencyGraph)
}

// Обработчик для получения зависимостей конкретного файла
func handleFileDependencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		FilePath string `json:"filePath"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestData); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Получаем зависимости для указанного файла
	fileDependencies := getFileDependencies(requestData.FilePath)

	json.NewEncoder(w).Encode(fileDependencies)
}

// Функция для получения зависимостей конкретного файла
func getFileDependencies(filePath string) DependencyGraph {
	graphMutex.RLock()
	defer graphMutex.RUnlock()

	fileAbsPath := filepath.Join(projectPath, filePath)

	// Создаем подграф для выбранного файла
	subgraph := DependencyGraph{
		Nodes: []Constant{},
		Edges: []Dependency{},
	}

	// Карта для отслеживания добавленных узлов
	addedNodes := make(map[string]bool)

	// Добавляем все константы из выбранного файла
	for _, node := range dependencyGraph.Nodes {
		if node.FilePath == fileAbsPath {
			subgraph.Nodes = append(subgraph.Nodes, node)
			addedNodes[node.Name] = true
		}
	}

	// Добавляем зависимости для этих констант
	for _, edge := range dependencyGraph.Edges {
		// Если исходная константа из нашего файла
		if addedNodes[edge.Source] {
			// Добавляем ребро
			subgraph.Edges = append(subgraph.Edges, edge)

			// Проверяем, есть ли целевая константа в подграфе
			if !addedNodes[edge.Target] {
				// Ищем целевую константу в общем графе
				for _, node := range dependencyGraph.Nodes {
					if node.Name == edge.Target {
						subgraph.Nodes = append(subgraph.Nodes, node)
						addedNodes[node.Name] = true
						break
					}
				}
			}
		}

		// Если целевая константа из нашего файла
		if addedNodes[edge.Target] && !addedNodes[edge.Source] {
			// Добавляем ребро
			subgraph.Edges = append(subgraph.Edges, edge)

			// Ищем исходную константу в общем графе
			for _, node := range dependencyGraph.Nodes {
				if node.Name == edge.Source {
					subgraph.Nodes = append(subgraph.Nodes, node)
					addedNodes[node.Name] = true
					break
				}
			}
		}
	}

	return subgraph
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

// Функция для построения графа зависимостей всего проекта
func buildDependencyGraph() {
	// Инициализируем граф
	dependencyGraph = DependencyGraph{
		Nodes: []Constant{},
		Edges: []Dependency{},
	}

	// Получаем список всех JS/TS файлов в проекте
	files := getJSTSFiles(projectPath)
	fmt.Printf("Найдено %d JS/TS файлов\n", len(files))

	// Используем WaitGroup для синхронизации горутин
	var wg sync.WaitGroup

	// Создаем пул из n горутин для параллельной обработки
	// Ограничиваем количество одновременных горутин для экономии ресурсов
	maxGoroutines := 10
	guard := make(chan struct{}, maxGoroutines)

	// Сначала находим все константы в проекте
	for _, file := range files {
		wg.Add(1)
		guard <- struct{}{} // Блокируемся, если уже запущено maxGoroutines горутин

		go func(filePath string) {
			defer wg.Done()
			defer func() { <-guard }() // Освобождаем слот в пуле

			findConstants(filePath)
		}(file)
	}

	wg.Wait() // Ждем завершения поиска констант

	fmt.Printf("Найдено %d констант\n", len(dependencyGraph.Nodes))

	// Затем устанавливаем зависимости между константами
	for _, file := range files {
		wg.Add(1)
		guard <- struct{}{}

		go func(filePath string) {
			defer wg.Done()
			defer func() { <-guard }()

			findDependencies(filePath)
		}(file)
	}

	wg.Wait() // Ждем завершения поиска зависимостей

	fmt.Printf("Найдено %d зависимостей\n", len(dependencyGraph.Edges))
}

// Функция для рекурсивного поиска всех JS/TS файлов в директории
func getJSTSFiles(root string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Игнорируем node_modules и другие служебные директории
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == ".git" || strings.HasPrefix(info.Name(), ".")) {
			return filepath.SkipDir
		}

		// Добавляем только JS/TS файлы
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".js") ||
		                     strings.HasSuffix(info.Name(), ".jsx") ||
		                     strings.HasSuffix(info.Name(), ".ts") ||
		                     strings.HasSuffix(info.Name(), ".tsx")) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking the path %q: %v\n", root, err)
	}

	return files
}

// Функция для поиска констант в файле
func findConstants(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	// Простой поиск объявлений констант
	// В реальном приложении здесь должен быть полноценный парсер JavaScript/TypeScript
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Поиск объявлений const
		if strings.HasPrefix(line, "const ") || strings.HasPrefix(line, "export const ") {
			// Извлекаем имя константы
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				constName := parts[1]
				if strings.HasPrefix(line, "export const") && len(parts) >= 3 {
					constName = parts[2]
				}

				// Удаляем все после знака =, : или запятой
				constName = strings.Split(constName, "=")[0]
				constName = strings.Split(constName, ":")[0]
				constName = strings.Split(constName, ",")[0]
				constName = strings.TrimSpace(constName)

				// Поиск значения константы
				value := ""
				if strings.Contains(line, "=") {
					valuePart := strings.Split(line, "=")[1]
					value = strings.TrimSpace(valuePart)
					// Удаляем возможную точку с запятой в конце
					value = strings.TrimSuffix(value, ";")
				}

				// Определяем тип константы
				constType := "unknown"
				if strings.Contains(line, ":") {
					typePart := strings.Split(strings.Split(line, ":")[1], "=")[0]
					constType = strings.TrimSpace(typePart)
				} else if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
					constType = "string"
				} else if strings.HasPrefix(value, "{") {
					constType = "object"
				} else if strings.HasPrefix(value, "[") {
					constType = "array"
				} else if value == "true" || value == "false" {
					constType = "boolean"
				} else if _, err := fmt.Sscanf(value, "%f", new(float64)); err == nil {
					constType = "number"
				}

				// Создаем константу
				constant := Constant{
					Name:     constName,
					Value:    value,
					Type:     constType,
					FilePath: filePath,
				}

				// Безопасно добавляем константу в граф
				graphMutex.Lock()
				dependencyGraph.Nodes = append(dependencyGraph.Nodes, constant)
				constantMap[constName] = true
				graphMutex.Unlock()

				log.Printf("Found constant %s in file %s at line %d\n", constName, filePath, lineNum+1)
			}
		}
	}
}

// Функция для поиска зависимостей между константами
func findDependencies(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	// Получаем список всех констант
	var constants []string
	graphMutex.RLock()
	for _, node := range dependencyGraph.Nodes {
		constants = append(constants, node.Name)
	}
	graphMutex.RUnlock()

	// Ищем использование констант в файле
	for _, constant := range constants {
		// Ищем все вхождения константы в файле
		matches := strings.Count(string(content), constant)

		if matches > 0 {
			// Для каждой найденной константы ищем зависимости от других констант
			// В этом простом примере мы считаем, что константа A зависит от константы B,
			// если обе объявлены в одном файле и A использует B в своем определении

			graphMutex.RLock()
			for _, node := range dependencyGraph.Nodes {
				// Проверяем, что это не та же самая константа и она объявлена в том же файле
				if node.Name != constant && node.FilePath == filePath {
					// Если значение константы содержит другую константу
					if strings.Contains(node.Value, constant) {
						// Добавляем зависимость
						dependency := Dependency{
							Source: node.Name,
							Target: constant,
						}

						// Освобождаем блокировку для чтения перед взятием блокировки для записи
						graphMutex.RUnlock()
						graphMutex.Lock()
						dependencyGraph.Edges = append(dependencyGraph.Edges, dependency)
						graphMutex.Unlock()
						graphMutex.RLock()

						log.Printf("Found dependency: %s -> %s in file %s\n", node.Name, constant, filePath)
					}
				}
			}
			graphMutex.RUnlock()
		}
	}
}
