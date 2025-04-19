package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDependencyService(t *testing.T) {
	fileService := &FileService{
		ProjectPath: "/test/path",
		GitIgnore:   nil,
	}

	dependencyService := NewDependencyService(fileService)

	if dependencyService == nil {
		t.Fatalf("Ожидается ненулевой DependencyService")
	}

	if dependencyService.FileService != fileService {
		t.Errorf("Ожидается FileService=%v, получено: %v", fileService, dependencyService.FileService)
	}

	if len(dependencyService.Graph.Nodes) != 0 {
		t.Errorf("Ожидается пустой список узлов, получено: %d узлов", len(dependencyService.Graph.Nodes))
	}

	if len(dependencyService.Graph.Edges) != 0 {
		t.Errorf("Ожидается пустой список ребер, получено: %d ребер", len(dependencyService.Graph.Edges))
	}

	if dependencyService.ConstantMap == nil {
		t.Errorf("Ожидается инициализированная карта констант")
	}
}

func TestFindConstants(t *testing.T) {
	// Создаем временную директорию и тестовый файл
	tempDir, err := os.MkdirTemp("", "dep-service-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем файл с константами для тестирования
	testFile := filepath.Join(tempDir, "constants.js")
	testContent := `
// Это комментарий
const SIMPLE_CONST = "simple value";
const NUMBER_CONST = 42;
const BOOLEAN_CONST = true;

// Константа с другой константой в значении
const DEPENDENT_CONST = "prefix_" + SIMPLE_CONST;

// Функция (не должна быть распознана как константа)
const myFunction = () => {
    return "not a constant";
};

// Константа с типом (TypeScript)
const TYPED_CONST: string = "typed value";

// Объект
const OBJECT_CONST = {
    key: "value",
    nested: {
        key: "nested value"
    }
};

// Массив
const ARRAY_CONST = ["item1", "item2", SIMPLE_CONST];

// Экспортируемая константа
export const EXPORTED_CONST = "exported value";
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Инициализируем сервисы
	fileService := NewFileService(tempDir, nil)
	dependencyService := NewDependencyService(fileService)

	// Запускаем поиск констант
	dependencyService.FindConstants(testFile)

	// Проверяем результаты
	expectedConstants := 8 // Количество констант в тестовом файле (без функции)
	if len(dependencyService.Graph.Nodes) != expectedConstants {
		t.Errorf("Ожидается %d констант, получено: %d", expectedConstants, len(dependencyService.Graph.Nodes))
	}

	// Проверяем, что функция не была распознана как константа
	for _, node := range dependencyService.Graph.Nodes {
		if node.Name == "myFunction" {
			t.Errorf("Функция myFunction была неправильно распознана как константа")
		}
	}

	// Проверяем наличие конкретных констант
	constantNames := map[string]bool{
		"SIMPLE_CONST":    false,
		"NUMBER_CONST":    false,
		"BOOLEAN_CONST":   false,
		"DEPENDENT_CONST": false,
		"TYPED_CONST":     false,
		"OBJECT_CONST":    false,
		"ARRAY_CONST":     false,
		"EXPORTED_CONST":  false,
	}

	for _, node := range dependencyService.Graph.Nodes {
		if _, exists := constantNames[node.Name]; exists {
			constantNames[node.Name] = true
		}
	}

	for name, found := range constantNames {
		if name != "myFunction" && !found {
			t.Errorf("Константа %s не была найдена", name)
		}
	}
}

func TestFindDependencies(t *testing.T) {
	// Создаем временную директорию и тестовый файл
	tempDir, err := os.MkdirTemp("", "dep-service-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем файл с зависимостями для тестирования
	testFile := filepath.Join(tempDir, "dependencies.js")
	testContent := `
const BASE_URL = "https://example.com";
const API_PATH = "/api";
const API_URL = BASE_URL + API_PATH;

const USER_ENDPOINT = API_URL + "/user";
const USER_ID = 123;
const USER_API = USER_ENDPOINT + "/" + USER_ID;
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Инициализируем сервисы
	fileService := NewFileService(tempDir, nil)
	dependencyService := NewDependencyService(fileService)

	// Сначала ищем константы
	dependencyService.FindConstants(testFile)

	// Затем ищем зависимости
	dependencyService.FindDependencies(testFile)

	// Проверяем результаты
	expectedConstants := 6  // Общее количество констант
	expectedDependencies := 5  // Ожидаемое количество зависимостей

	if len(dependencyService.Graph.Nodes) != expectedConstants {
		t.Errorf("Ожидается %d констант, получено: %d", expectedConstants, len(dependencyService.Graph.Nodes))
	}

	if len(dependencyService.Graph.Edges) != expectedDependencies {
		t.Errorf("Ожидается %d зависимостей, получено: %d", expectedDependencies, len(dependencyService.Graph.Edges))
	}

	// Проверяем конкретные зависимости
	expectedEdges := []struct {
		source string
		target string
	}{
		{"API_URL", "BASE_URL"},
		{"API_URL", "API_PATH"},
		{"USER_ENDPOINT", "API_URL"},
		{"USER_API", "USER_ENDPOINT"},
		{"USER_API", "USER_ID"},
	}

	for _, expected := range expectedEdges {
		found := false
		for _, edge := range dependencyService.Graph.Edges {
			if edge.Source == expected.source && edge.Target == expected.target {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Ожидаемая зависимость не найдена: %s -> %s", expected.source, expected.target)
		}
	}
}

func TestGetFileDependencies(t *testing.T) {
	// Создаем временную директорию и тестовые файлы
	tempDir, err := os.MkdirTemp("", "file-dep-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем первый файл с константами
	file1 := filepath.Join(tempDir, "file1.js")
	content1 := `
const COMMON_CONST = "common";
const FILE1_CONST = COMMON_CONST + "_file1";
`

	// Создаем второй файл с константами
	file2 := filepath.Join(tempDir, "file2.js")
	content2 := `
const COMMON_CONST = "common";
const FILE2_CONST = COMMON_CONST + "_file2";
`

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Инициализируем сервисы
	fileService := NewFileService(tempDir, nil)
	dependencyService := NewDependencyService(fileService)

	// Анализируем оба файла
	dependencyService.FindConstants(file1)
	dependencyService.FindConstants(file2)
	dependencyService.FindDependencies(file1)
	dependencyService.FindDependencies(file2)

	// Получаем зависимости для первого файла
	file1Dependencies := dependencyService.GetFileDependencies(file1)

	// Проверяем результаты
	expectedNodes := 2  // COMMON_CONST и FILE1_CONST
	expectedEdges := 1  // FILE1_CONST -> COMMON_CONST

	if len(file1Dependencies.Nodes) != expectedNodes {
		t.Errorf("Ожидается %d узлов для file1.js, получено: %d", expectedNodes, len(file1Dependencies.Nodes))
	}

	if len(file1Dependencies.Edges) != expectedEdges {
		t.Errorf("Ожидается %d ребер для file1.js, получено: %d", expectedEdges, len(file1Dependencies.Edges))
	}

	// Проверяем, что зависимости из file2.js не включены
	for _, node := range file1Dependencies.Nodes {
		if node.Name == "FILE2_CONST" {
			t.Errorf("Константа FILE2_CONST не должна быть включена в зависимости file1.js")
		}
	}

	// Получаем зависимости для второго файла
	file2Dependencies := dependencyService.GetFileDependencies(file2)

	// Проверяем результаты для второго файла
	if len(file2Dependencies.Nodes) != expectedNodes {
		t.Errorf("Ожидается %d узлов для file2.js, получено: %d", expectedNodes, len(file2Dependencies.Nodes))
	}

	if len(file2Dependencies.Edges) != expectedEdges {
		t.Errorf("Ожидается %d ребер для file2.js, получено: %d", expectedEdges, len(file2Dependencies.Edges))
	}

	// Проверяем, что зависимости из file1.js не включены
	for _, node := range file2Dependencies.Nodes {
		if node.Name == "FILE1_CONST" {
			t.Errorf("Константа FILE1_CONST не должна быть включена в зависимости file2.js")
		}
	}
}

func TestBuildDependencyGraph(t *testing.T) {
	// Создаем временную директорию
	tempDir, err := os.MkdirTemp("", "build-graph-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем файлы с константами и зависимостями
	files := map[string]string{
		"config.js": `
const API_URL = "https://example.com/api";
const TIMEOUT = 5000;
`,
		"users.js": `
const API_URL = "https://example.com/api";
const USERS_ENDPOINT = API_URL + "/users";
`,
		"app.js": `
const API_URL = "https://example.com/api";
const TIMEOUT = 5000;
const USER_URL = API_URL + "/user";
`,
	}

	for fileName, content := range files {
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Не удалось создать файл %s: %v", fileName, err)
		}
	}

	// Инициализируем сервисы
	fileService := NewFileService(tempDir, nil)
	dependencyService := NewDependencyService(fileService)

	// Строим граф зависимостей
	dependencyService.BuildDependencyGraph()

	// Проверяем результаты
	// Учитывая, что API_URL определен в трех файлах, а TIMEOUT в двух,
	// общее количество констант должно быть 7
	expectedConstants := 7
	if len(dependencyService.Graph.Nodes) != expectedConstants {
		t.Errorf("Ожидается %d констант, получено: %d", expectedConstants, len(dependencyService.Graph.Nodes))
	}

	// Должно быть 2 зависимости: USERS_ENDPOINT -> API_URL и USER_URL -> API_URL
	expectedDependencies := 2
	if len(dependencyService.Graph.Edges) != expectedDependencies {
		t.Errorf("Ожидается %d зависимостей, получено: %d", expectedDependencies, len(dependencyService.Graph.Edges))
	}

	// Проверяем конкретные зависимости
	dependencyExists := func(source, target string) bool {
		for _, edge := range dependencyService.Graph.Edges {
			if edge.Source == source && edge.Target == target {
				return true
			}
		}
		return false
	}

	if !dependencyExists("USERS_ENDPOINT", "API_URL") {
		t.Errorf("Ожидаемая зависимость не найдена: USERS_ENDPOINT -> API_URL")
	}

	if !dependencyExists("USER_URL", "API_URL") {
		t.Errorf("Ожидаемая зависимость не найдена: USER_URL -> API_URL")
	}
}
