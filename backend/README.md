# Dependency Graph Visualizer Backend

Бэкенд-часть приложения для визуализации графа зависимостей в JavaScript/TypeScript проектах, написанная на Go.

## Функциональность

- Сканирование структуры директорий проекта
- Анализ JavaScript/TypeScript файлов для выявления констант и зависимостей между ними
- Построение графа зависимостей
- Предоставление REST API для фронтенд-части приложения

## Структура проекта

```
backend/
  ├── main.go                  # Точка входа в приложение
  ├── models/                  # Модели данных
  │   └── models.go            # Определение основных структур (FileNode, Constant, Dependency, DependencyGraph)
  ├── handlers/                # HTTP-обработчики
  │   └── handlers.go          # Обработчики запросов API
  ├── services/                # Бизнес-логика
  │   ├── file_service.go      # Сервис для работы с файловой системой
  │   └── dependency_service.go # Сервис для анализа зависимостей
  └── utils/                   # Вспомогательные утилиты
      └── gitignore.go         # Обработка правил .gitignore
```

## Запуск приложения

### Сборка

```bash
cd backend
go build -o dependency-graph-visualizer
```

### Запуск

```bash
./dependency-graph-visualizer -path /path/to/your/js/project
```

Где `/path/to/your/js/project` - путь к JavaScript/TypeScript проекту, который вы хотите проанализировать.

## API Endpoints

### 1. Информация о проекте

```
GET /api/project-info
```

Возвращает информацию о проекте (имя и путь).

### 2. Структура файлов проекта

```
GET /api/file-tree
```

Возвращает дерево файлов проекта.

### 3. Полный граф зависимостей

```
GET /api/dependency-graph
```

Возвращает полный граф зависимостей между константами в проекте.

### 4. Зависимости конкретного файла

```
POST /api/file-dependencies
Content-Type: application/json

{
  "filePath": "path/to/file.js"
}
```

Возвращает граф зависимостей для указанного файла.

## Особенности реализации

- Поддержка JavaScript и TypeScript файлов (`.js`, `.jsx`, `.ts`, `.tsx`)
- Игнорирование файлов и директорий, указанных в `.gitignore`
- CORS поддержка для взаимодействия с фронтенд-частью
- Анализ константных выражений и их взаимосвязей

## Тестирование

Для запуска тестов выполните:

```bash
go test ./...
```

## Технологии

- Go 1.21
- Стандартная библиотека Go (без внешних зависимостей)
- HTTP сервер на основе `net/http`
