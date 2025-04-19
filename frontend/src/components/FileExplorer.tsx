import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { api, FileNode, ProjectInfo } from '../api/api';
import FileNodeComponent from './file-explorer/FileNode';
import StatusBar from './file-explorer/StatusBar';

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
  display: flex;
  flex-direction: column;
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

const ExplorerContent = styled.div`
  flex: 1;
  overflow-y: auto;
`;

const NodeList = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0;
`;

const LoadingIndicator = styled.div`
  padding: 20px;
  text-align: center;
  color: #999;
`;

const ErrorMessage = styled.div`
  padding: 20px;
  text-align: center;
  color: #f44336;
`;

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
    const fetchData = async () => {
      try {
        setLoading(true);
        setError('');

        // Загружаем информацию о проекте
        const projectData = await api.getProjectInfo();
        setProjectInfo(projectData);

        // Загружаем дерево файлов
        const treeData = await api.getFileTree();
        setFileTree(treeData);
      } catch (err) {
        console.error('Error fetching data:', err);
        setError('Ошибка при загрузке данных');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleSelectNode = (node: FileNode) => {
    setSelectedNode(node);

    // Если выбран файл и передан обработчик onFileSelect, вызываем его
    if (!node.isDir && onFileSelect) {
      onFileSelect(node.path);
    }
  };

  if (loading) {
    return (
      <ExplorerContainer>
        <ExplorerHeader>
          <span>ПРОЕКТ</span>
        </ExplorerHeader>
        <ExplorerContent>
          <LoadingIndicator>Загрузка...</LoadingIndicator>
        </ExplorerContent>
        <StatusBar selectedNode={null} />
      </ExplorerContainer>
    );
  }

  if (error) {
    return (
      <ExplorerContainer>
        <ExplorerHeader>
          <span>ПРОЕКТ</span>
        </ExplorerHeader>
        <ExplorerContent>
          <ErrorMessage>{error}</ErrorMessage>
        </ExplorerContent>
        <StatusBar selectedNode={null} />
      </ExplorerContainer>
    );
  }

  return (
    <ExplorerContainer>
      <ExplorerHeader>
        <span>ПРОЕКТ: {projectInfo?.projectName}</span>
      </ExplorerHeader>
      <ExplorerContent>
        {fileTree && (
          <NodeList>
            <FileNodeComponent
              node={fileTree}
              selectedPath={selectedNode?.path || ''}
              onSelectNode={handleSelectNode}
              level={0}
            />
          </NodeList>
        )}
      </ExplorerContent>
      <StatusBar selectedNode={selectedNode} />
    </ExplorerContainer>
  );
};

export default FileExplorer;
