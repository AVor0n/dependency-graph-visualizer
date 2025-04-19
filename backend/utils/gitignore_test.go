package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGitIgnore(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "gitignore-test")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Случай 1: Директория без .gitignore
	gitIgnore := LoadGitIgnore(tempDir)
	if gitIgnore != nil {
		t.Errorf("Ожидается nil при отсутствии .gitignore, получено: %v", gitIgnore)
	}

	// Случай 2: Директория с пустым .gitignore
	gitIgnorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitIgnorePath, []byte(""), 0644); err != nil {
		t.Fatalf("Не удалось создать файл .gitignore: %v", err)
	}

	gitIgnore = LoadGitIgnore(tempDir)
	if gitIgnore == nil {
		t.Errorf("Ожидается ненулевой GitIgnore для пустого .gitignore")
	} else if len(gitIgnore.patterns) != 0 {
		t.Errorf("Ожидается 0 шаблонов, получено: %d", len(gitIgnore.patterns))
	}

	// Случай 3: Директория с .gitignore и шаблонами
	patterns := []string{
		"# Комментарий, который должен быть проигнорирован",
		"",
		"node_modules/",
		"*.log",
		"!important.log",
	}

	if err := os.WriteFile(gitIgnorePath, []byte(joinLines(patterns...)), 0644); err != nil {
		t.Fatalf("Не удалось обновить файл .gitignore: %v", err)
	}

	gitIgnore = LoadGitIgnore(tempDir)
	if gitIgnore == nil {
		t.Errorf("Ожидается ненулевой GitIgnore для .gitignore с шаблонами")
	} else {
		// Должно быть 3 шаблона (пустые строки и комментарии игнорируются)
		expectedPatterns := 3
		if len(gitIgnore.patterns) != expectedPatterns {
			t.Errorf("Ожидается %d шаблонов, получено: %d", expectedPatterns, len(gitIgnore.patterns))
		}
	}
}

func TestIsIgnored(t *testing.T) {
	// Создаем GitIgnore с известными шаблонами
	gitIgnore := &GitIgnore{
		patterns: []string{
			"node_modules/",
			"*.log",
			"!important.log",
			"/dist",
		},
	}

	// Тестовые случаи: path -> должен ли быть проигнорирован
	tests := []struct {
		path     string
		expected bool
	}{
		{"node_modules/file.js", true},             // Соответствует node_modules/
		{"path/to/node_modules/file.js", true},     // Соответствует node_modules/ в поддиректории
		{"file.log", true},                         // Соответствует *.log
		{"path/to/file.log", true},                 // Соответствует *.log в поддиректории
		{"important.log", false},                   // Соответствует !important.log (исключение)
		{"path/to/important.log", false},           // Соответствует !important.log в поддиректории
		{"dist/file.js", true},                     // Соответствует /dist (только в корне)
		{"path/to/dist/file.js", false},            // Не соответствует /dist в поддиректории
		{"node_modulesx/file.js", false},           // Не соответствует node_modules/
		{"file.txt", false},                        // Не соответствует ни одному шаблону
	}

	for _, test := range tests {
		result := gitIgnore.IsIgnored(test.path)
		if result != test.expected {
			t.Errorf("Для пути %q: ожидалось IsIgnored=%v, получено: %v", test.path, test.expected, result)
		}
	}

	// Тест с nil GitIgnore
	var nilGitIgnore *GitIgnore
	if nilGitIgnore.IsIgnored("any/path") {
		t.Errorf("Ожидается false для nil GitIgnore")
	}
}

func TestMatchPattern(t *testing.T) {
	// Тестовые случаи: path, pattern -> должен ли соответствовать
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"file.txt", "*.txt", true},               // Простое соответствие по расширению
		{"path/to/file.txt", "*.txt", true},       // Соответствие в поддиректории
		{"file.txt", "file.*", true},              // Соответствие по имени файла
		{"file.log", "*.log", true},               // Соответствие по расширению
		{"a/b/c/file.log", "*.log", true},         // Соответствие в глубокой поддиректории
		{"dir/", "dir/", true},                    // Соответствие директории
		{"path/to/dir/", "dir/", true},            // Соответствие директории в поддиректории
		{"path/to/dir/file.txt", "dir/", true},    // Соответствие файла в директории
		{"file.jpg", "*.txt", false},              // Несоответствие расширения
		{"file.txt", "file.jpg", false},           // Несоответствие имени файла
	}

	for _, test := range tests {
		result := matchPattern(test.path, test.pattern)
		if result != test.expected {
			t.Errorf("Для пути %q и шаблона %q: ожидалось matchPattern=%v, получено: %v",
				test.path, test.pattern, test.expected, result)
		}
	}
}

// Вспомогательная функция для объединения строк с переносами строк
func joinLines(lines ...string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
