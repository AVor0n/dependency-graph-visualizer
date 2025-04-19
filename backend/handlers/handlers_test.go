package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/avor0n/dependency-graph-visualizer/models"
	"github.com/avor0n/dependency-graph-visualizer/utils"
)

// MockFileService - мок-структура для FileService
type MockFileService struct {
	ProjectPath      string
	GitIgnore        *utils.GitIgnore
	ScanDirectoryFunc func(relativePath string) models.FileNode
}

func (m *MockFileService) ScanDirectory(relativePath string) models.FileNode {
	if m.ScanDirectoryFunc != nil {
		return m.ScanDirectoryFunc(relativePath)
	}
	return models.FileNode{}
}

func (m *MockFileService) GetJSTSFiles() []string {
	return []string{}
}

// MockDependencyService - мок-структура для DependencyService
type MockDependencyService struct {
	FileService            *MockFileService
	Graph                  models.DependencyGraph
	GetFileDependenciesFunc func(filePath string) models.DependencyGraph
}

func (m *MockDependencyService) GetFileDependencies(filePath string) models.DependencyGraph {
	if m.GetFileDependenciesFunc != nil {
		return m.GetFileDependenciesFunc(filePath)
	}

	// Если путь пустой, возвращаем весь граф
	if filePath == "" {
		return m.Graph
	}

	return models.DependencyGraph{}
}

func (m *MockDependencyService) BuildDependencyGraph() {
	// Пустая реализация для интерфейса
}

func TestNewHandler(t *testing.T) {
	// Создаем мок-сервисы
	mockFileService := &MockFileService{
		ProjectPath: "/test/path",
	}

	mockDependencyService := &MockDependencyService{
		FileService: mockFileService,
	}

	projectPath := "/test/path"

	// Создаем обработчик
	handler := &Handler{
		FileService:       mockFileService,
		DependencyService: mockDependencyService,
		ProjectPath:       projectPath,
	}

	// Проверяем, что обработчик создан правильно
	if handler == nil {
		t.Fatalf("Ожидается ненулевой Handler")
	}

	if handler.FileService != mockFileService {
		t.Errorf("Ожидается FileService=%v, получено: %v", mockFileService, handler.FileService)
	}

	if handler.DependencyService != mockDependencyService {
		t.Errorf("Ожидается DependencyService=%v, получено: %v", mockDependencyService, handler.DependencyService)
	}

	if handler.ProjectPath != projectPath {
		t.Errorf("Ожидается ProjectPath=%s, получено: %s", projectPath, handler.ProjectPath)
	}
}

func TestEnableCORS(t *testing.T) {
	// Создаем тестовый обработчик, который просто устанавливает статус код 200
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Оборачиваем в EnableCORS
	corsHandler := EnableCORS(testHandler)

	// Создаем тестовый запрос и рекордер для ответа
	req := httptest.NewRequest("GET", "http://example.com/api", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	corsHandler.ServeHTTP(rec, req)

	// Проверяем заголовки CORS
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "POST, GET, OPTIONS, PUT, DELETE",
		"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization",
	}

	for header, expected := range expectedHeaders {
		actual := rec.Header().Get(header)
		if actual != expected {
			t.Errorf("Для заголовка %s ожидается %s, получено: %s", header, expected, actual)
		}
	}

	// Проверяем, что для метода OPTIONS возвращается статус 200 без вызова основного обработчика
	req = httptest.NewRequest("OPTIONS", "http://example.com/api", nil)
	rec = httptest.NewRecorder()

	corsHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Для метода OPTIONS ожидается статус %d, получено: %d", http.StatusOK, rec.Code)
	}
}

func TestHandleProjectInfo(t *testing.T) {
	// Создаем тестовый обработчик
	handler := &Handler{
		ProjectPath: "/test/project/path",
	}

	// Создаем тестовый запрос и рекордер для ответа
	req := httptest.NewRequest("GET", "/api/project-info", nil)
	rec := httptest.NewRecorder()

	// Выполняем обработчик
	handler.HandleProjectInfo(rec, req)

	// Проверяем заголовки
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Ожидается Content-Type=application/json, получено: %s", contentType)
	}

	if corsHeader := rec.Header().Get("Access-Control-Allow-Origin"); corsHeader != "*" {
		t.Errorf("Ожидается Access-Control-Allow-Origin=*, получено: %s", corsHeader)
	}

	// Декодируем ответ
	var response struct {
		ProjectPath string `json:"projectPath"`
		ProjectName string `json:"projectName"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	// Проверяем значения
	if response.ProjectPath != handler.ProjectPath {
		t.Errorf("Ожидается ProjectPath=%s, получено: %s", handler.ProjectPath, response.ProjectPath)
	}

	expected := "path" // Последняя часть пути /test/project/path
	if response.ProjectName != expected {
		t.Errorf("Ожидается ProjectName=%s, получено: %s", expected, response.ProjectName)
	}
}

func TestHandleFileTree(t *testing.T) {
	// Создаем мок FileService, который возвращает тестовое дерево файлов
	mockFileService := &MockFileService{
		ScanDirectoryFunc: func(relativePath string) models.FileNode {
			return models.FileNode{
				Name:  "root",
				Path:  "",
				IsDir: true,
				Children: []models.FileNode{
					{
						Name:  "file.js",
						Path:  "file.js",
						IsDir: false,
					},
					{
						Name:  "folder",
						Path:  "folder",
						IsDir: true,
						Children: []models.FileNode{
							{
								Name:  "nested.js",
								Path:  "folder/nested.js",
								IsDir: false,
							},
						},
					},
				},
			}
		},
	}

	// Создаем тестовый обработчик с моком
	handler := &Handler{
		FileService: mockFileService,
	}

	// Создаем тестовый запрос и рекордер для ответа
	req := httptest.NewRequest("GET", "/api/file-tree", nil)
	rec := httptest.NewRecorder()

	// Выполняем обработчик
	handler.HandleFileTree(rec, req)

	// Проверяем заголовки
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Ожидается Content-Type=application/json, получено: %s", contentType)
	}

	if corsHeader := rec.Header().Get("Access-Control-Allow-Origin"); corsHeader != "*" {
		t.Errorf("Ожидается Access-Control-Allow-Origin=*, получено: %s", corsHeader)
	}

	// Декодируем ответ
	var response models.FileNode
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	// Проверяем структуру ответа
	if response.Name != "root" {
		t.Errorf("Ожидается Name=root, получено: %s", response.Name)
	}

	if len(response.Children) != 2 {
		t.Errorf("Ожидается 2 дочерних узла, получено: %d", len(response.Children))
	}
}

func TestHandleDependencyGraph(t *testing.T) {
	// Создаем мок DependencyService с тестовым графом
	mockDependencyService := &MockDependencyService{
		Graph: models.DependencyGraph{
			Nodes: []models.Constant{
				{
					Name:     "CONST1",
					Value:    "value1",
					Type:     "string",
					FilePath: "/path/to/file.js",
					LineNum:  10,
				},
				{
					Name:     "CONST2",
					Value:    "value2",
					Type:     "string",
					FilePath: "/path/to/file.js",
					LineNum:  20,
				},
			},
			Edges: []models.Dependency{
				{
					Source: "CONST2",
					Target: "CONST1",
				},
			},
		},
	}

	// Создаем тестовый обработчик с моком
	handler := &Handler{
		DependencyService: mockDependencyService,
	}

	// Создаем тестовый запрос и рекордер для ответа
	req := httptest.NewRequest("GET", "/api/dependency-graph", nil)
	rec := httptest.NewRecorder()

	// Выполняем обработчик
	handler.HandleDependencyGraph(rec, req)

	// Проверяем заголовки
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Ожидается Content-Type=application/json, получено: %s", contentType)
	}

	// Декодируем ответ
	var response models.DependencyGraph
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	// Проверяем структуру ответа
	expectedNodes := 2
	if len(response.Nodes) != expectedNodes {
		t.Errorf("Ожидается %d узлов, получено: %d", expectedNodes, len(response.Nodes))
	}

	expectedEdges := 1
	if len(response.Edges) != expectedEdges {
		t.Errorf("Ожидается %d ребер, получено: %d", expectedEdges, len(response.Edges))
	}
}

func TestHandleFileDependencies(t *testing.T) {
	// Создаем мок DependencyService
	mockDependencyService := &MockDependencyService{
		GetFileDependenciesFunc: func(filePath string) models.DependencyGraph {
			return models.DependencyGraph{
				Nodes: []models.Constant{
					{
						Name:     "CONST1",
						Value:    "value1",
						Type:     "string",
						FilePath: filePath,
						LineNum:  10,
					},
				},
				Edges: []models.Dependency{},
			}
		},
	}

	// Создаем тестовый обработчик с моком
	handler := &Handler{
		DependencyService: mockDependencyService,
	}

	// Случай 1: Неправильный метод запроса
	req := httptest.NewRequest("GET", "/api/file-dependencies", nil)
	rec := httptest.NewRecorder()

	handler.HandleFileDependencies(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Для метода GET ожидается статус %d, получено: %d",
			http.StatusMethodNotAllowed, rec.Code)
	}

	// Случай 2: Правильный метод, но неправильный JSON
	req = httptest.NewRequest("POST", "/api/file-dependencies",
		strings.NewReader(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	handler.HandleFileDependencies(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Для неправильного JSON ожидается статус %d, получено: %d",
			http.StatusBadRequest, rec.Code)
	}

	// Случай 3: Правильный запрос
	req = httptest.NewRequest("POST", "/api/file-dependencies",
		strings.NewReader(`{"filePath": "/path/to/file.js"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	handler.HandleFileDependencies(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Ожидается статус %d, получено: %d", http.StatusOK, rec.Code)
	}

	// Декодируем ответ
	var response models.DependencyGraph
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	// Проверяем структуру ответа
	expectedNodes := 1
	if len(response.Nodes) != expectedNodes {
		t.Errorf("Ожидается %d узлов, получено: %d", expectedNodes, len(response.Nodes))
	}
}
