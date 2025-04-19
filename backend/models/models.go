package models

// FileNode представляет узел в дереве файлов
type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

// Constant представляет константу в коде
type Constant struct {
	Name     string `json:"name"`     // Имя константы
	Value    string `json:"value"`    // Значение константы
	Type     string `json:"type"`     // Тип константы
	FilePath string `json:"filePath"` // Путь к файлу, где объявлена константа
	LineNum  int    `json:"lineNum"`  // Номер строки в файле
}

// Dependency представляет зависимость между константами
type Dependency struct {
	Source string `json:"source"` // Имя исходной константы
	Target string `json:"target"` // Имя целевой константы
}

// DependencyGraph представляет граф зависимостей
type DependencyGraph struct {
	Nodes []Constant   `json:"nodes"` // Узлы графа (константы)
	Edges []Dependency `json:"edges"` // Ребра графа (зависимости)
}
