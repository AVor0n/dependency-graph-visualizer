import React from 'react';
import styled from 'styled-components';
import { FileNode } from '../../api/api';

const StatusBarContainer = styled.div`
  padding: 5px 10px;
  font-size: 12px;
  background-color: #007acc;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

interface StatusBarProps {
  selectedNode: FileNode | null;
}

const StatusBar: React.FC<StatusBarProps> = ({ selectedNode }) => {
  if (!selectedNode) {
    return (
      <StatusBarContainer>
        <span>Нет выбранного файла</span>
      </StatusBarContainer>
    );
  }

  return (
    <StatusBarContainer>
      <span>
        {selectedNode.isDir ? 'Папка' : 'Файл'}: {selectedNode.name}
      </span>
      <span>
        {selectedNode.path}
      </span>
    </StatusBarContainer>
  );
};

export default StatusBar;
