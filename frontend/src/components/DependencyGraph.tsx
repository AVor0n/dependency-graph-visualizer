import React from 'react';
import styled from 'styled-components';

const GraphContainer = styled.div`
  flex: 1;
  background-color: #1e1e1e;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #d4d4d4;
`;

const PlaceholderText = styled.div`
  font-size: 16px;
  color: #777;
  text-align: center;
`;

interface DependencyGraphProps {
  selectedFile?: string;
}

const DependencyGraph: React.FC<DependencyGraphProps> = ({ selectedFile }) => {
  return (
    <GraphContainer>
      <PlaceholderText>
        {selectedFile ? (
          `Здесь будет отображаться граф зависимостей для файла: ${selectedFile}`
        ) : (
          'Выберите файл в проводнике для отображения графа зависимостей'
        )}
      </PlaceholderText>
    </GraphContainer>
  );
};

export default DependencyGraph;
