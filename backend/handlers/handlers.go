package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/avor0n/dependency-graph-visualizer/models"
	"github.com/avor0n/dependency-graph-visualizer/services"
)

// FileServiceInterface определяет интерфейс для FileService
type FileServiceInterface interface {
	ScanDirectory(relativePath string) models.FileNode
	GetJSTSFiles() []string
}

// DependencyServiceInterface определяет интерфейс для DependencyService
type DependencyServiceInterface interface {
	GetFileDependencies(filePath string) models.DependencyGraph
	BuildDependencyGraph()
}

// Handler представляет обработчики HTTP запросов
type Handler struct {
	FileService       FileServiceInterface
	DependencyService DependencyServiceInterface
	ProjectPath       string
}

// NewHandler создает новый экземпляр Handler
func NewHandler(fileService services.FileService, dependencyService services.DependencyService, projectPath string) *Handler {
	return &Handler{
		FileService:       &fileService,
		DependencyService: &dependencyService,
		ProjectPath:       projectPath,
	}
}

// EnableCORS добавляет CORS-заголовки к ответу
func EnableCORS(handler http.Handler) http.Handler {
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

// HandleProjectInfo обрабатывает запрос информации о проекте
func (h *Handler) HandleProjectInfo(w http.ResponseWriter, r *http.Request) {
	response := struct {
		ProjectPath string `json:"projectPath"`
		ProjectName string `json:"projectName"`
	}{
		ProjectPath: h.ProjectPath,
		ProjectName: filepath.Base(h.ProjectPath),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleFileTree обрабатывает запрос дерева файлов
func (h *Handler) HandleFileTree(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Получаем структуру директории
	rootNode := h.FileService.ScanDirectory("")

	json.NewEncoder(w).Encode(rootNode)
}

// HandleDependencyGraph обрабатывает запрос полного графа зависимостей
func (h *Handler) HandleDependencyGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Здесь мы не можем напрямую получить Graph из интерфейса DependencyServiceInterface
	// Вместо этого мы можем получить полный граф, передав пустой путь к файлу
	graph := h.DependencyService.GetFileDependencies("")

	json.NewEncoder(w).Encode(graph)
}

// HandleFileDependencies обрабатывает запрос зависимостей конкретного файла
func (h *Handler) HandleFileDependencies(w http.ResponseWriter, r *http.Request) {
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
	fileDependencies := h.DependencyService.GetFileDependencies(requestData.FilePath)

	json.NewEncoder(w).Encode(fileDependencies)
}
