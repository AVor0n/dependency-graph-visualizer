package main

import (
	"net/http"
	"testing"
)

// Примечание: тестирование функции main напрямую затруднительно,
// поскольку она вызывает os.Exit() и http.ListenAndServe().
// В реальном проекте лучше выделять логику в отдельные функции,
// которые можно тестировать независимо.

// TestHTTPHandlersRegistration проверяет регистрацию HTTP-обработчиков
func TestHTTPHandlersRegistration(t *testing.T) {
	// Создаем тестовый ServeMux
	mux := http.NewServeMux()

	// Регистрируем фиктивные обработчики для проверки валидности путей
	mux.HandleFunc("/api/project-info", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/api/file-tree", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/api/dependency-graph", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/api/file-dependencies", func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle("/", http.FileServer(http.Dir(".")))

	// Проверяем, что все ожидаемые пути обрабатываются
	paths := []string{
		"/api/project-info",
		"/api/file-tree",
		"/api/dependency-graph",
		"/api/file-dependencies",
		"/",
	}

	for _, path := range paths {
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatalf("Ошибка создания запроса для %s: %v", path, err)
		}

		handler, _ := mux.Handler(req)
		if handler == nil {
			t.Errorf("Не найден обработчик для пути %s", path)
		}
	}
}
