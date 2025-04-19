import React, { useState } from 'react';
import styled from 'styled-components';
import { FiFolder, FiFile, FiChevronDown, FiChevronRight } from 'react-icons/fi';
import { FileNode } from '../../api/api';

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

const NodeList = styled.ul`
  list-style: none;
  padding: 0;
  margin: 0;
`;

interface FileNodeComponentProps {
  node: FileNode;
  selectedPath: string;
  onSelectNode: (node: FileNode) => void;
  level: number; // Уровень вложенности для отступов
}

const FileNodeComponent: React.FC<FileNodeComponentProps> = ({
  node,
  selectedPath,
  onSelectNode,
  level
}) => {
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
              <FileNodeComponent
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

export default FileNodeComponent;
