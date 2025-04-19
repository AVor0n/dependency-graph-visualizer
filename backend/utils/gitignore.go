package utils

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GitIgnore представляет правила игнорирования файлов
type GitIgnore struct {
	patterns []string
}

// LoadGitIgnore загружает правила .gitignore из директории проекта
func LoadGitIgnore(projectPath string) *GitIgnore {
	gitIgnorePath := filepath.Join(projectPath, ".gitignore")

	// Проверяем, существует ли файл .gitignore
	if _, err := os.Stat(gitIgnorePath); os.IsNotExist(err) {
		log.Println("Файл .gitignore не найден, игнорирование файлов не будет применяться")
		return nil
	}

	// Открываем файл .gitignore
	file, err := os.Open(gitIgnorePath)
	if err != nil {
		log.Printf("Ошибка при открытии файла .gitignore: %v\n", err)
		return nil
	}
	defer file.Close()

	// Создаем объект GitIgnore
	gitIgnore := &GitIgnore{
		patterns: []string{},
	}

	// Читаем файл построчно
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Добавляем шаблон
		gitIgnore.patterns = append(gitIgnore.patterns, line)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Ошибка при чтении файла .gitignore: %v\n", err)
		return nil
	}

	log.Printf("Загружено %d правил из .gitignore\n", len(gitIgnore.patterns))
	return gitIgnore
}

// IsIgnored проверяет, соответствует ли путь правилам .gitignore
func (gi *GitIgnore) IsIgnored(path string) bool {
	// Если gitignore не загружен, ничего не игнорируем
	if gi == nil {
		return false
	}

	// Нормализуем путь для сравнения
	path = filepath.ToSlash(path)

	// По умолчанию не игнорируем путь
	ignored := false

	for _, pattern := range gi.patterns {
		// Обрабатываем разные типы шаблонов

		// Шаблон исключения (начинается с !)
		if strings.HasPrefix(pattern, "!") {
			negPattern := pattern[1:]
			if matchPattern(path, negPattern) {
				// Если путь соответствует шаблону исключения, он не игнорируется
				return false
			}
			continue
		}

		// Проверяем, соответствует ли путь текущему шаблону
		if matchPattern(path, pattern) {
			ignored = true
		}
	}

	return ignored
}

// matchPattern проверяет соответствие пути шаблону .gitignore
func matchPattern(path, pattern string) bool {
	// Нормализуем путь и шаблон
	path = strings.Trim(path, "/")

	// Если шаблон заканчивается на /, то это директория
	isDirPattern := strings.HasSuffix(pattern, "/")
	if isDirPattern {
		pattern = strings.TrimSuffix(pattern, "/")
	}

	// Проверка для шаблона директории
	if isDirPattern {
		// Путь соответствует, если он содержит эту директорию в любом месте
		return strings.HasSuffix(path, pattern) ||
			   strings.Contains(path, pattern+"/") ||
			   path == pattern
	}

	// Проверка для шаблонов, начинающихся с /
	if strings.HasPrefix(pattern, "/") {
		// Шаблон с / в начале соответствует только файлам в корне проекта
		pattern = strings.TrimPrefix(pattern, "/")
		return path == pattern || strings.HasPrefix(path, pattern+"/")
	}

	// Проверка для обычных шаблонов с расширением (*.ext)
	if strings.HasPrefix(pattern, "*") {
		ext := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, ext) || strings.Contains(path, "/"+pattern)
	}

	// Проверка для шаблонов с именем файла (file.*)
	if strings.HasSuffix(pattern, "*") {
		basename := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(filepath.Base(path), basename) ||
			   strings.Contains(path, pattern)
	}

	// Точное совпадение или совпадение в пути
	return path == pattern ||
		   strings.HasSuffix(path, "/"+pattern) ||
		   strings.Contains(path, "/"+pattern+"/")
}
