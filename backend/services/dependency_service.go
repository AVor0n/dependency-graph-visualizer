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
	inObjectDeclaration := false
	currentObjectName := ""

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Пропускаем пустые строки и комментарии
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
			continue
		}

		// Отслеживаем уровень вложенности
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		blockDepth += openBraces - closeBraces

		// Проверяем начало объявления объекта
		if strings.Contains(trimmedLine, "={") || strings.Contains(trimmedLine, "= {") {
			inObjectDeclaration = true
			// Сохраняем имя объекта, если это константа
			if strings.HasPrefix(trimmedLine, "const ") || strings.HasPrefix(trimmedLine, "export const ") {
				parts := strings.Fields(trimmedLine)
				if len(parts) >= 2 {
					if strings.HasPrefix(trimmedLine, "export const") && len(parts) >= 3 {
						currentObjectName = parts[2]
					} else {
						currentObjectName = parts[1]
					}
					currentObjectName = strings.Split(currentObjectName, "=")[0]
					currentObjectName = strings.TrimSpace(currentObjectName)
				}
			}
		}

		// Проверяем конец объявления объекта
		if inObjectDeclaration && closeBraces > 0 && blockDepth == 0 {
			inObjectDeclaration = false
			// Если это был объект-константа, добавляем его
			if currentObjectName != "" {
				// Создаем константу для объекта
				constant := models.Constant{
					Name:     currentObjectName,
					Value:    "object",
					Type:     "object",
					FilePath: filePath,
					LineNum:  lineNum + 1,
				}

				// Безопасно добавляем константу в граф
				ds.GraphMutex.Lock()
				ds.Graph.Nodes = append(ds.Graph.Nodes, constant)
				ds.ConstantMap[currentObjectName] = true
				ds.GraphMutex.Unlock()

				log.Printf("Found constant %s in file %s at line %d\n", currentObjectName, filePath, lineNum+1)
				currentObjectName = ""
			}
		}

		// Если мы внутри объявления объекта, пропускаем строку
		if inObjectDeclaration {
			continue
		}

		// Если мы находимся на верхнем уровне файла
		if blockDepth == 0 && (strings.HasPrefix(trimmedLine, "const ") || strings.HasPrefix(trimmedLine, "export const ")) &&
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
					// Пропускаем, так как уже обработали объекты выше
					continue
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
	// Проверяем существование файла
	_, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Error accessing file %s: %v\n", filePath, err)
		return
	}

	// Получаем список всех констант
	var constants []string
	ds.GraphMutex.RLock()
	for _, node := range ds.Graph.Nodes {
		constants = append(constants, node.Name)
	}
	ds.GraphMutex.RUnlock()

	// Ищем зависимости между константами в этом файле

	// Карта констант в этом файле
	fileConstants := make(map[string]models.Constant)

	// Заполняем карту констант для этого файла
	ds.GraphMutex.RLock()
	for _, node := range ds.Graph.Nodes {
		if node.FilePath == filePath {
			fileConstants[node.Name] = node
		}
	}
	ds.GraphMutex.RUnlock()

	// Для каждой константы из файла ищем зависимости
	for constName, constant := range fileConstants {
		// Ищем использование других констант в значении этой константы
		for otherName := range fileConstants {
			// Пропускаем сравнение константы с самой собой
			if constName == otherName {
				continue
			}

			// Проверяем, содержит ли значение константы имя другой константы
			if strings.Contains(constant.Value, otherName) {
				// Добавляем зависимость
				dependency := models.Dependency{
					Source: constName,
					Target: otherName,
				}

				ds.GraphMutex.Lock()
				ds.Graph.Edges = append(ds.Graph.Edges, dependency)
				ds.GraphMutex.Unlock()

				log.Printf("Found dependency: %s -> %s in file %s\n", constName, otherName, filePath)
			}
		}
	}
}

// GetFileDependencies возвращает зависимости для указанного файла
func (ds *DependencyService) GetFileDependencies(filePath string) models.DependencyGraph {
	ds.GraphMutex.RLock()
	defer ds.GraphMutex.RUnlock()

	// Если путь пустой, возвращаем весь граф
	if filePath == "" {
		return ds.Graph
	}

	fileAbsPath := filePath
	if !strings.HasPrefix(filePath, ds.FileService.ProjectPath) {
		fileAbsPath = filepath.Join(ds.FileService.ProjectPath, filePath)
	}

	// Создаем подграф для выбранного файла
	subgraph := models.DependencyGraph{
		Nodes: []models.Constant{},
		Edges: []models.Dependency{},
	}

	// Карта для отслеживания констант, которые относятся к этому файлу
	fileConstants := make(map[string]bool)

	// Добавляем все константы из выбранного файла
	for _, node := range ds.Graph.Nodes {
		if node.FilePath == fileAbsPath {
			subgraph.Nodes = append(subgraph.Nodes, node)
			fileConstants[node.Name] = true
		}
	}

	// Добавляем только зависимости между константами из этого файла
	for _, edge := range ds.Graph.Edges {
		// Добавляем ребро только если обе константы находятся в файле
		if fileConstants[edge.Source] && fileConstants[edge.Target] {
			subgraph.Edges = append(subgraph.Edges, edge)
		}
	}

	return subgraph
}
