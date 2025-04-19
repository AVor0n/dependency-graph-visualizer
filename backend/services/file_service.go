package services

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/avor0n/dependency-graph-visualizer/models"
	"github.com/avor0n/dependency-graph-visualizer/utils"
)

// FileService представляет сервис для работы с файловой системой
type FileService struct {
	ProjectPath string
	GitIgnore   *utils.GitIgnore
}

// NewFileService создает новый экземпляр FileService
func NewFileService(projectPath string, gitIgnore *utils.GitIgnore) *FileService {
	return &FileService{
		ProjectPath: projectPath,
		GitIgnore:   gitIgnore,
	}
}

// ScanDirectory рекурсивно сканирует директорию и возвращает структуру
func (fs *FileService) ScanDirectory(relativePath string) models.FileNode {
	absPath := filepath.Join(fs.ProjectPath, relativePath)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		return models.FileNode{}
	}

	nodeName := fileInfo.Name()
	if relativePath == "" {
		// Если это корневая директория, используем последнюю часть пути
		nodeName = filepath.Base(fs.ProjectPath)
	}

	node := models.FileNode{
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
		// Проверяем, соответствует ли файл правилам .gitignore
		childRelPath := filepath.Join(relativePath, entry.Name())
		if fs.GitIgnore != nil && fs.GitIgnore.IsIgnored(childRelPath) {
			continue
		}

		childNode := fs.ScanDirectory(childRelPath)
		node.Children = append(node.Children, childNode)
	}

	return node
}

// GetJSTSFiles получает список всех JS/TS файлов в проекте
func (fs *FileService) GetJSTSFiles() []string {
	var files []string

	err := filepath.Walk(fs.ProjectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Получаем относительный путь для проверки .gitignore
		relPath, err := filepath.Rel(fs.ProjectPath, path)
		if err != nil {
			relPath = path
		}

		// Проверяем, соответствует ли файл правилам .gitignore
		if fs.GitIgnore != nil && fs.GitIgnore.IsIgnored(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
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
		log.Printf("Error walking the path %q: %v\n", fs.ProjectPath, err)
	}

	return files
}
