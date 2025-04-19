// Константа с базовым URL для API
const API_BASE_URL = 'http://localhost:8080/api';

// Типы данных
export interface Constant {
  name: string;
  value: string;
  type: string;
  filePath: string;
  lineNum: number;
}

export interface Dependency {
  source: string;
  target: string;
}

export interface DependencyGraph {
  nodes: Constant[];
  edges: Dependency[];
}

export interface ProjectInfo {
  projectPath: string;
  projectName: string;
}

export interface FileNode {
  name: string;
  path: string;
  isDir: boolean;
  children?: FileNode[];
}

// Функции для работы с API
export const api = {
  // Получение информации о проекте
  async getProjectInfo(): Promise<ProjectInfo> {
    const response = await fetch(`${API_BASE_URL}/project-info`);
    if (!response.ok) {
      throw new Error('Не удалось получить информацию о проекте');
    }
    return response.json();
  },

  // Получение дерева файлов
  async getFileTree(): Promise<FileNode> {
    const response = await fetch(`${API_BASE_URL}/file-tree`);
    if (!response.ok) {
      throw new Error('Не удалось получить дерево файлов');
    }
    return response.json();
  },

  // Получение полного графа зависимостей
  async getDependencyGraph(): Promise<DependencyGraph> {
    const response = await fetch(`${API_BASE_URL}/dependency-graph`);
    if (!response.ok) {
      throw new Error('Не удалось загрузить данные графа');
    }
    return response.json();
  },

  // Получение зависимостей для конкретного файла
  async getFileDependencies(filePath: string): Promise<DependencyGraph> {
    const response = await fetch(`${API_BASE_URL}/file-dependencies`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ filePath })
    });

    if (!response.ok) {
      throw new Error('Не удалось загрузить данные зависимостей для файла');
    }
    return response.json();
  }
};

export default api;
