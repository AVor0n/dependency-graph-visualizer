package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/avor0n/dependency-graph-visualizer/utils"
)

func TestNewFileService(t *testing.T) {
	projectPath := "/test/path"
	gitIgnore := &utils.GitIgnore{}

	fileService := NewFileService(projectPath, gitIgnore)

	if fileService == nil {
		t.Fatalf("Ожидается ненулевой FileService")
	}

	if fileService.ProjectPath != projectPath {
		t.Errorf("Ожидается ProjectPath=%s, получено: %s", projectPath, fileService.ProjectPath)
	}

	if fileService.GitIgnore != gitIgnore {
		t.Errorf("Ожидается GitIgnore=%v, получено: %v", gitIgnore, fileService.GitIgnore)
	}
}

func TestScanDirectory(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "file-service-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем структуру директорий и файлов для тестирования
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Не удалось создать поддиректорию: %v", err)
	}

	testFile1 := filepath.Join(tempDir, "test.js")
	if err := os.WriteFile(testFile1, []byte("test content"), 0644); err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	testFile2 := filepath.Join(subDir, "test2.js")
	if err := os.WriteFile(testFile2, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Создаем .gitignore который игнорирует subdir/
	gitIgnorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitIgnorePath, []byte("subdir/"), 0644); err != nil {
		t.Fatalf("Не удалось создать файл .gitignore: %v", err)
	}

	// Случай 1: Без использования .gitignore
	fileService := NewFileService(tempDir, nil)
	rootNode := fileService.ScanDirectory("")

	// Проверяем, что корневой узел создан правильно
	if rootNode.Name != filepath.Base(tempDir) {
		t.Errorf("Ожидается имя корневого узла '%s', получено: %s", filepath.Base(tempDir), rootNode.Name)
	}

	if !rootNode.IsDir {
		t.Errorf("Ожидается, что корневой узел является директорией")
	}

	// В корневой директории должно быть 3 элемента (subdir, test.js, .gitignore)
	expectedChildren := 3
	if len(rootNode.Children) != expectedChildren {
		t.Errorf("Ожидается %d дочерних узлов, получено: %d", expectedChildren, len(rootNode.Children))
	}

	// Случай 2: С использованием .gitignore
	gitIgnore := utils.LoadGitIgnore(tempDir)
	fileService = NewFileService(tempDir, gitIgnore)
	rootNodeWithIgnore := fileService.ScanDirectory("")

	// В корневой директории должно быть 2 элемента (test.js и .gitignore),
	// subdir должна быть проигнорирована
	expectedChildrenWithIgnore := 2
	if len(rootNodeWithIgnore.Children) != expectedChildrenWithIgnore {
		t.Errorf("Ожидается %d дочерних узлов с .gitignore, получено: %d",
			expectedChildrenWithIgnore, len(rootNodeWithIgnore.Children))
	}
}

func TestGetJSTSFiles(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "js-ts-files-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем структуру директорий и файлов для тестирования
	subDir := filepath.Join(tempDir, "src")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Не удалось создать поддиректорию: %v", err)
	}

	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	if err := os.Mkdir(nodeModulesDir, 0755); err != nil {
		t.Fatalf("Не удалось создать директорию node_modules: %v", err)
	}

	// Создаем тестовые файлы различных типов
	testFiles := map[string]string{
		filepath.Join(tempDir, "test.js"):             "JS file",
		filepath.Join(tempDir, "test.jsx"):            "JSX file",
		filepath.Join(tempDir, "test.ts"):             "TS file",
		filepath.Join(tempDir, "test.tsx"):            "TSX file",
		filepath.Join(tempDir, "test.txt"):            "Text file (должен быть проигнорирован)",
		filepath.Join(subDir, "component.jsx"):        "Component JSX",
		filepath.Join(subDir, "component.css"):        "CSS file (должен быть проигнорирован)",
		filepath.Join(nodeModulesDir, "module.js"):    "JS in node_modules (должен быть проигнорирован)",
	}

	for filePath, content := range testFiles {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Не удалось создать тестовый файл %s: %v", filePath, err)
		}
	}

	// Создаем .gitignore который игнорирует *.txt
	gitIgnorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitIgnorePath, []byte("*.txt"), 0644); err != nil {
		t.Fatalf("Не удалось создать файл .gitignore: %v", err)
	}

	// Случай 1: Без использования .gitignore
	fileService := NewFileService(tempDir, nil)
	files := fileService.GetJSTSFiles()

	// Ожидаем 5 JS/TS файлов (4 в корневой директории + 1 в subDir)
	// node_modules должны быть проигнорированы по умолчанию
	expectedFiles := 5
	if len(files) != expectedFiles {
		t.Errorf("Ожидается %d JS/TS файлов без учета .gitignore, получено: %d", expectedFiles, len(files))
	}

	// Случай 2: С использованием .gitignore
	gitIgnore := utils.LoadGitIgnore(tempDir)
	fileService = NewFileService(tempDir, gitIgnore)
	filesWithIgnore := fileService.GetJSTSFiles()

	// Количество файлов должно остаться тем же, так как .gitignore игнорирует только .txt,
	// которые и так не включаются в результат
	expectedFilesWithIgnore := 5
	if len(filesWithIgnore) != expectedFilesWithIgnore {
		t.Errorf("Ожидается %d JS/TS файлов с учетом .gitignore, получено: %d",
			expectedFilesWithIgnore, len(filesWithIgnore))
	}

	// Проверяем, что все найденные файлы имеют правильные расширения
	for _, file := range filesWithIgnore {
		ext := filepath.Ext(file)
		validExt := ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx"
		if !validExt {
			t.Errorf("Найден файл с недопустимым расширением: %s", file)
		}
	}
}
