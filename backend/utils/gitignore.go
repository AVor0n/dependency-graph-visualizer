package utils

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

	for _, pattern := range gi.patterns {
		// Обрабатываем разные типы шаблонов

		// Шаблон исключения (начинается с !)
		if strings.HasPrefix(pattern, "!") {
			// Если путь соответствует шаблону исключения, он не игнорируется
			negPattern := pattern[1:]
			if matchPattern(path, negPattern) {
				return false
			}
			continue
		}

		// Обычный шаблон
		if matchPattern(path, pattern) {
			return true
		}
	}

	return false
}

// matchPattern проверяет соответствие пути шаблону .gitignore
func matchPattern(path, pattern string) bool {
	// Удаляем слеши в начале и конце
	path = strings.Trim(path, "/")
	pattern = strings.Trim(pattern, "/")

	// Если шаблон заканчивается на /, то это директория
	isDirPattern := strings.HasSuffix(pattern, "/")
	if isDirPattern {
		pattern = strings.TrimSuffix(pattern, "/")
	}

	// Преобразуем шаблон gitignore в регулярное выражение
	regexPattern := "^"

	// Если шаблон начинается с /, то он соответствует только корню проекта
	if strings.HasPrefix(pattern, "/") {
		pattern = strings.TrimPrefix(pattern, "/")
	} else {
		// Иначе шаблон может соответствовать любой части пути
		regexPattern = ".*?"
	}

	// Заменяем шаблонные символы на их регулярные выражения
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "**/", ".*?")
	pattern = strings.ReplaceAll(pattern, "**", ".*?")
	pattern = strings.ReplaceAll(pattern, "*", "[^/]*?")
	pattern = strings.ReplaceAll(pattern, "?", "[^/]")

	regexPattern += pattern

	// Если это шаблон директории, он должен соответствовать всему пути или подпапке
	if isDirPattern {
		regexPattern += "(/.*)?$"
	} else {
		regexPattern += "$"
	}

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		log.Printf("Ошибка при компиляции регулярного выражения для шаблона %s: %v\n", pattern, err)
		return false
	}

	return regex.MatchString(path)
}
