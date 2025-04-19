package models

import (
	"encoding/json"
	"testing"
)

func TestFileNodeStructure(t *testing.T) {
	// Создаем тестовую структуру дерева файлов
	fileTree := FileNode{
		Name:  "root",
		Path:  "/",
		IsDir: true,
		Children: []FileNode{
			{
				Name:  "folder1",
				Path:  "/folder1",
				IsDir: true,
				Children: []FileNode{
					{
						Name:  "file1.js",
						Path:  "/folder1/file1.js",
						IsDir: false,
					},
				},
			},
			{
				Name:  "file2.js",
				Path:  "/file2.js",
				IsDir: false,
			},
		},
	}

	// Проверяем, что структура создана правильно
	if fileTree.Name != "root" {
		t.Errorf("Ожидается имя корневого узла 'root', получено: %s", fileTree.Name)
	}

	if len(fileTree.Children) != 2 {
		t.Errorf("Ожидается 2 дочерних узла, получено: %d", len(fileTree.Children))
	}

	// Проверяем сериализацию/десериализацию JSON
	jsonData, err := json.Marshal(fileTree)
	if err != nil {
		t.Fatalf("Ошибка сериализации в JSON: %v", err)
	}

	var deserializedTree FileNode
	err = json.Unmarshal(jsonData, &deserializedTree)
	if err != nil {
		t.Fatalf("Ошибка десериализации из JSON: %v", err)
	}

	// Проверяем, что структура сохранилась после десериализации
	if deserializedTree.Name != fileTree.Name {
		t.Errorf("Ожидается имя корневого узла '%s', получено: %s", fileTree.Name, deserializedTree.Name)
	}

	if len(deserializedTree.Children) != len(fileTree.Children) {
		t.Errorf("Ожидается %d дочерних узлов, получено: %d", len(fileTree.Children), len(deserializedTree.Children))
	}
}

func TestDependencyGraphStructure(t *testing.T) {
	// Создаем тестовую структуру графа зависимостей
	graph := DependencyGraph{
		Nodes: []Constant{
			{
				Name:     "CONST1",
				Value:    "value1",
				Type:     "string",
				FilePath: "/path/to/file1.js",
				LineNum:  10,
			},
			{
				Name:     "CONST2",
				Value:    "CONST1 + 'value2'",
				Type:     "string",
				FilePath: "/path/to/file1.js",
				LineNum:  20,
			},
		},
		Edges: []Dependency{
			{
				Source: "CONST2",
				Target: "CONST1",
			},
		},
	}

	// Проверяем, что структура создана правильно
	if len(graph.Nodes) != 2 {
		t.Errorf("Ожидается 2 узла, получено: %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 1 {
		t.Errorf("Ожидается 1 ребро, получено: %d", len(graph.Edges))
	}

	// Проверяем сериализацию/десериализацию JSON
	jsonData, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("Ошибка сериализации в JSON: %v", err)
	}

	var deserializedGraph DependencyGraph
	err = json.Unmarshal(jsonData, &deserializedGraph)
	if err != nil {
		t.Fatalf("Ошибка десериализации из JSON: %v", err)
	}

	// Проверяем, что структура сохранилась после десериализации
	if len(deserializedGraph.Nodes) != len(graph.Nodes) {
		t.Errorf("Ожидается %d узлов, получено: %d", len(graph.Nodes), len(deserializedGraph.Nodes))
	}

	if len(deserializedGraph.Edges) != len(graph.Edges) {
		t.Errorf("Ожидается %d ребер, получено: %d", len(graph.Edges), len(deserializedGraph.Edges))
	}

	// Проверяем содержимое узлов
	if deserializedGraph.Nodes[0].Name != graph.Nodes[0].Name {
		t.Errorf("Ожидается имя узла '%s', получено: %s", graph.Nodes[0].Name, deserializedGraph.Nodes[0].Name)
	}

	// Проверяем содержимое ребер
	if deserializedGraph.Edges[0].Source != graph.Edges[0].Source ||
	   deserializedGraph.Edges[0].Target != graph.Edges[0].Target {
		t.Errorf("Ребро не соответствует ожидаемому: ожидается %v, получено: %v",
			graph.Edges[0], deserializedGraph.Edges[0])
	}
}
