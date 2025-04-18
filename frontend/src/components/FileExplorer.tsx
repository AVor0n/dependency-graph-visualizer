import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { FiFolder, FiFile, FiChevronDown, FiChevronRight } from 'react-icons/fi';

// Типы данных
interface FileNode {
  name: string;
  path: string;
  isDir: boolean;
  children?: FileNode[];
}

interface ProjectInfo {
  projectPath: string;
  projectName: string;
}

// Стили
const ExplorerContainer = styled.div`
  background-color: #1e1e1e;
  color: #d4d4d4;
  width: 300px;
  height: 100%;
  overflow-y: auto;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  font-size: 14px;
  user-select: none;
`;

const ExplorerHeader = styled.div`
  padding: 10px;
  font-weight: bold;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: #75beff;
  background-color: #252526;
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const NodeList = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0;
`;

const NodeItem = styled.li`
  margin: 0;
  padding: 0;
`;

interface NodeContentProps {
  isSelected: boolean;
}

const NodeContent = styled.div<NodeContentProps>`
  display: flex;
  align-items: center;
  padding: 4px 8px;
  cursor: pointer;
  &:hover {
    background-color: #2a2a2a;
  }
  background-color: ${props => props.isSelected ? '#37373d' : 'transparent'};
`;

const NodeLabel = styled.span`
  margin-left: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const IconContainer = styled.span`
  margin-right: 4px;
  display: flex;
  align-items: center;
`;

const ChildrenContainer = styled.div`
  margin-left: 16px;
`;

const StatusBar = styled.div`
  padding: 5px 10px;
  font-size: 12px;
  background-color: #007acc;
  color: white;
`;

// Компонент для отдельного узла (файла или папки)
const FileNode: React.FC<{
  node: FileNode;
  selectedPath: string;
  onSelectNode: (node: FileNode) => void;
  level: number; // Уровень вложенности для отступов
}> = ({ node, selectedPath, onSelectNode, level }) => {
  const [expanded, setExpanded] = useState(false);
  const isSelected = node.path === selectedPath;

  const toggleExpand = (e: React.MouseEvent) => {
    e.stopPropagation();
    setExpanded(!expanded);
  };

  const handleClick = () => {
    onSelectNode(node);
  };

  return (
    <NodeItem>
      <NodeContent isSelected={isSelected} onClick={handleClick}>
        {node.isDir && (
          <IconContainer onClick={toggleExpand}>
            {expanded ? <FiChevronDown size={16} /> : <FiChevronRight size={16} />}
          </IconContainer>
        )}
        <IconContainer>
          {node.isDir ? <FiFolder color="#75beff" size={16} /> : <FiFile color="#cccccc" size={16} />}
        </IconContainer>
        <NodeLabel>{node.name}</NodeLabel>
      </NodeContent>
      {node.isDir && expanded && node.children && (
        <ChildrenContainer>
          <NodeList>
            {node.children.map((child) => (
              <FileNode
                key={child.path}
                node={child}
                selectedPath={selectedPath}
                onSelectNode={onSelectNode}
                level={level + 1}
              />
            ))}
          </NodeList>
        </ChildrenContainer>
      )}
    </NodeItem>
  );
};

// Основной компонент FileExplorer
interface FileExplorerProps {
  onFileSelect?: (filePath: string) => void;
}

const FileExplorer: React.FC<FileExplorerProps> = ({ onFileSelect }) => {
  const [fileTree, setFileTree] = useState<FileNode | null>(null);
  const [projectInfo, setProjectInfo] = useState<ProjectInfo | null>(null);
  const [selectedNode, setSelectedNode] = useState<FileNode | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    // Загружаем информацию о проекте
    const fetchProjectInfo = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/project-info');
        if (!response.ok) {
          throw new Error('Не удалось получить информацию о проекте');
        }
        const data = await response.json();
        setProjectInfo(data);
      } catch (err) {
        setError('Ошибка при загрузке информации о проекте');
        console.error(err);
      }
    };

    // Загружаем дерево файлов
    const fetchFileTree = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/file-tree');
        if (!response.ok) {
          throw new Error('Не удалось получить дерево файлов');
        }
        const data = await response.json();
        setFileTree(data);
      } catch (err) {
        setError('Ошибка при загрузке дерева файлов');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchProjectInfo();
    fetchFileTree();
  }, []);

  const handleSelectNode = (node: FileNode) => {
    setSelectedNode(node);

    // Если выбран файл и передан обработчик onFileSelect, вызываем его
    if (!node.isDir && onFileSelect) {
      onFileSelect(node.path);
    }
  };

  if (loading) {
    return <ExplorerContainer>Загрузка...</ExplorerContainer>;
  }

  if (error) {
    return <ExplorerContainer>{error}</ExplorerContainer>;
  }

  return (
    <ExplorerContainer>
      <ExplorerHeader>
        <span>ПРОЕКТ: {projectInfo?.projectName}</span>
      </ExplorerHeader>
      {fileTree && (
        <NodeList>
          <FileNode
            node={fileTree}
            selectedPath={selectedNode?.path || ''}
            onSelectNode={handleSelectNode}
            level={0}
          />
        </NodeList>
      )}
      <StatusBar>
        {selectedNode && `${selectedNode.isDir ? 'Папка' : 'Файл'}: ${selectedNode.name}`}
      </StatusBar>
    </ExplorerContainer>
  );
};

export default FileExplorer;
