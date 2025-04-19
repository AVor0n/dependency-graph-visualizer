import { useState, useEffect } from 'react';
import styled from 'styled-components';
import GraphCanvas from './graph/GraphCanvas';
import { GraphData, D3Node } from './graph/types';
import { api, DependencyGraph as DependencyGraphType } from '../api/api';

const GraphContainer = styled.div`
  flex: 1;
  background-color: #1e1e1e;
  display: flex;
  flex-direction: column;
  color: #d4d4d4;
`;

const GraphHeader = styled.div`
  padding: 10px 15px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3e3e3e;
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const Title = styled.h2`
  margin: 0;
  font-size: 18px;
  font-weight: normal;
`;

const GraphContent = styled.div`
  flex: 1;
  overflow: hidden;
  position: relative;
`;

const PlaceholderText = styled.div`
  font-size: 16px;
  color: #777;
  text-align: center;
  padding: 20px;
`;

const LoadingSpinner = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;

  &:after {
    content: " ";
    display: block;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 6px solid #007acc;
    border-color: #007acc transparent #007acc transparent;
    animation: spinner 1.2s linear infinite;
  }

  @keyframes spinner {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;

// Функция для преобразования данных из API в формат для D3
const transformData = (data: DependencyGraphType): GraphData => {
  const nodes: D3Node[] = data.nodes.map(node => ({
    id: node.name,
    name: node.name,
    type: node.type,
    value: node.value,
    filePath: node.filePath,
    lineNum: node.lineNum,
    color: getNodeColor(node.type)
  }));

  const links = data.edges.map(edge => ({
    source: edge.source,
    target: edge.target
  }));

  return { nodes, links };
};

// Определяем цвет узла в зависимости от типа константы
const getNodeColor = (type: string): string => {
  switch (type) {
    case 'string':
      return '#4da6ff'; // голубой
    case 'number':
      return '#ff9900'; // оранжевый
    case 'boolean':
      return '#00cc00'; // зеленый
    case 'object':
      return '#cc00cc'; // фиолетовый
    case 'array':
      return '#ff3333'; // красный
    default:
      return '#aaaaaa'; // серый для неизвестных типов
  }
};

interface DependencyGraphProps {
  selectedFile?: string;
}

const DependencyGraph: React.FC<DependencyGraphProps> = ({ selectedFile }) => {
  const [graphData, setGraphData] = useState<GraphData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [hoveredNode, setHoveredNode] = useState<D3Node | null>(null);

  // Загружаем данные о зависимостях при изменении выбранного файла
  useEffect(() => {
    if (!selectedFile) {
      // Если файл не выбран, загружаем общий граф
      fetchFullGraph();
    } else {
      // Иначе загружаем данные для конкретного файла
      fetchFileGraph(selectedFile);
    }
  }, [selectedFile]);

  // Загружает полный граф зависимостей
  const fetchFullGraph = async () => {
    setLoading(true);
    setError(null);

    try {
      const data = await api.getDependencyGraph();
      const transformedData = transformData(data);
      setGraphData(transformedData);
    } catch (err) {
      console.error('Error fetching graph data:', err);
      setError('Ошибка при загрузке графа зависимостей');
    } finally {
      setLoading(false);
    }
  };

  // Загружает граф зависимостей для конкретного файла
  const fetchFileGraph = async (filePath: string) => {
    setLoading(true);
    setError(null);

    try {
      const data = await api.getFileDependencies(filePath);
      const transformedData = transformData(data);
      setGraphData(transformedData);
    } catch (err) {
      console.error('Error fetching file dependencies:', err);
      setError('Ошибка при загрузке зависимостей для файла');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <GraphContainer>
        <GraphHeader>
          <Title>Граф зависимостей констант</Title>
        </GraphHeader>
        <LoadingSpinner />
      </GraphContainer>
    );
  }

  if (error) {
    return (
      <GraphContainer>
        <GraphHeader>
          <Title>Граф зависимостей констант</Title>
        </GraphHeader>
        <PlaceholderText>{error}</PlaceholderText>
      </GraphContainer>
    );
  }

  if (!graphData || graphData.nodes.length === 0) {
    return (
      <GraphContainer>
        <GraphHeader>
          <Title>Граф зависимостей констант</Title>
        </GraphHeader>
        <PlaceholderText>
          {selectedFile
            ? 'В выбранном файле не найдено констант или зависимостей'
            : 'Выберите файл в проводнике для отображения графа зависимостей'}
        </PlaceholderText>
      </GraphContainer>
    );
  }

  return (
    <GraphContainer>
      <GraphHeader>
        <Title>
          {selectedFile
            ? `Граф зависимостей: ${selectedFile}`
            : 'Общий граф зависимостей проекта'}
        </Title>
        {hoveredNode && (
          <div>
            <strong>{hoveredNode.name}</strong> - {hoveredNode.type} (строка {hoveredNode.lineNum})
          </div>
        )}
      </GraphHeader>
      <GraphContent>
        <GraphCanvas data={graphData} onNodeHover={setHoveredNode} />
      </GraphContent>
    </GraphContainer>
  );
};

export default DependencyGraph;
