package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/avor0n/dependency-graph-visualizer/models"
)

// DependencyService представляет сервис для работы с зависимостями
type DependencyService struct {
	FileService  *FileService
	Graph        models.DependencyGraph
	ConstantMap  map[string]bool
	GraphMutex   sync.RWMutex
}

// NewDependencyService создает новый экземпляр DependencyService
func NewDependencyService(fileService *FileService) *DependencyService {
	return &DependencyService{
		FileService: fileService,
		Graph: models.DependencyGraph{
			Nodes: []models.Constant{},
			Edges: []models.Dependency{},
		},
		ConstantMap: make(map[string]bool),
	}
}

// BuildDependencyGraph строит граф зависимостей для всего проекта
func (ds *DependencyService) BuildDependencyGraph() {
	// Получаем список всех JS/TS файлов в проекте
	files := ds.FileService.GetJSTSFiles()
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

			ds.FindConstants(filePath)
		}(file)
	}

	wg.Wait() // Ждем завершения поиска констант

	fmt.Printf("Найдено %d констант\n", len(ds.Graph.Nodes))

	// Затем устанавливаем зависимости между константами
	for _, file := range files {
		wg.Add(1)
		guard <- struct{}{}

		go func(filePath string) {
			defer wg.Done()
			defer func() { <-guard }()

			ds.FindDependencies(filePath)
		}(file)
	}

	wg.Wait() // Ждем завершения поиска зависимостей

	fmt.Printf("Найдено %d зависимостей\n", len(ds.Graph.Edges))
}

// FindConstants находит константы в файле
func (ds *DependencyService) FindConstants(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	// Простой поиск объявлений констант
	lines := strings.Split(string(content), "\n")

	// Отслеживаем уровень вложенности блоков кода
	blockDepth := 0

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Отслеживаем уровень вложенности
		blockDepth += strings.Count(line, "{") - strings.Count(line, "}")

		// Игнорируем строки с закрывающими скобками (они могут уменьшить blockDepth)
		if strings.Contains(trimmedLine, "}") {
			continue
		}

		// Проверяем, что мы находимся на верхнем уровне файла
		if blockDepth > 0 {
			continue
		}

		// Поиск объявлений const
		if (strings.HasPrefix(trimmedLine, "const ") || strings.HasPrefix(trimmedLine, "export const ")) &&
		   !strings.Contains(trimmedLine, "function") && !strings.Contains(trimmedLine, "=>") {
			// Извлекаем имя константы
			parts := strings.Fields(trimmedLine)
			if len(parts) >= 2 {
				constName := parts[1]
				if strings.HasPrefix(trimmedLine, "export const") && len(parts) >= 3 {
					constName = parts[2]
				}

				// Удаляем все после знака =, : или запятой
				constName = strings.Split(constName, "=")[0]
				constName = strings.Split(constName, ":")[0]
				constName = strings.Split(constName, ",")[0]
				constName = strings.TrimSpace(constName)

				// Поиск значения константы
				value := ""
				if strings.Contains(trimmedLine, "=") {
					valuePart := strings.Split(trimmedLine, "=")[1]
					value = strings.TrimSpace(valuePart)
					// Удаляем возможную точку с запятой в конце
					value = strings.TrimSuffix(value, ";")
				}

				// Определяем тип константы
				constType := "unknown"
				if strings.Contains(trimmedLine, ":") {
					typePart := strings.Split(strings.Split(trimmedLine, ":")[1], "=")[0]
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
				constant := models.Constant{
					Name:     constName,
					Value:    value,
					Type:     constType,
					FilePath: filePath,
					LineNum:  lineNum + 1,
				}

				// Безопасно добавляем константу в граф
				ds.GraphMutex.Lock()
				ds.Graph.Nodes = append(ds.Graph.Nodes, constant)
				ds.ConstantMap[constName] = true
				ds.GraphMutex.Unlock()

				log.Printf("Found constant %s in file %s at line %d\n", constName, filePath, lineNum+1)
			}
		}
	}
}

// FindDependencies находит зависимости между константами
func (ds *DependencyService) FindDependencies(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	// Получаем список всех констант
	var constants []string
	ds.GraphMutex.RLock()
	for _, node := range ds.Graph.Nodes {
		constants = append(constants, node.Name)
	}
	ds.GraphMutex.RUnlock()

	// Ищем использование констант в файле
	for _, constant := range constants {
		// Ищем все вхождения константы в файле
		matches := strings.Count(string(content), constant)

		if matches > 0 {
			// Для каждой найденной константы ищем зависимости от других констант
			// В этом простом примере мы считаем, что константа A зависит от константы B,
			// если обе объявлены в одном файле и A использует B в своем определении

			ds.GraphMutex.RLock()
			for _, node := range ds.Graph.Nodes {
				// Проверяем, что это не та же самая константа и она объявлена в том же файле
				if node.Name != constant && node.FilePath == filePath {
					// Если значение константы содержит другую константу
					if strings.Contains(node.Value, constant) {
						// Добавляем зависимость
						dependency := models.Dependency{
							Source: node.Name,
							Target: constant,
						}

						// Освобождаем блокировку для чтения перед взятием блокировки для записи
						ds.GraphMutex.RUnlock()
						ds.GraphMutex.Lock()
						ds.Graph.Edges = append(ds.Graph.Edges, dependency)
						ds.GraphMutex.Unlock()
						ds.GraphMutex.RLock()

						log.Printf("Found dependency: %s -> %s in file %s\n", node.Name, constant, filePath)
					}
				}
			}
			ds.GraphMutex.RUnlock()
		}
	}
}

// GetFileDependencies возвращает зависимости для указанного файла
func (ds *DependencyService) GetFileDependencies(filePath string) models.DependencyGraph {
	ds.GraphMutex.RLock()
	defer ds.GraphMutex.RUnlock()

	fileAbsPath := filePath
	if !strings.HasPrefix(filePath, ds.FileService.ProjectPath) {
		fileAbsPath = filepath.Join(ds.FileService.ProjectPath, filePath)
	}

	// Создаем подграф для выбранного файла
	subgraph := models.DependencyGraph{
		Nodes: []models.Constant{},
		Edges: []models.Dependency{},
	}

	// Карта для отслеживания добавленных узлов
	addedNodes := make(map[string]bool)

	// Добавляем все константы из выбранного файла
	for _, node := range ds.Graph.Nodes {
		if node.FilePath == fileAbsPath {
			subgraph.Nodes = append(subgraph.Nodes, node)
			addedNodes[node.Name] = true
		}
	}

	// Добавляем зависимости для этих констант
	for _, edge := range ds.Graph.Edges {
		// Если исходная константа из нашего файла
		if addedNodes[edge.Source] {
			// Добавляем ребро
			subgraph.Edges = append(subgraph.Edges, edge)

			// Проверяем, есть ли целевая константа в подграфе
			if !addedNodes[edge.Target] {
				// Ищем целевую константу в общем графе
				for _, node := range ds.Graph.Nodes {
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
			for _, node := range ds.Graph.Nodes {
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
