import styled from 'styled-components';
import { D3Node } from './types';

const TooltipContainer = styled.div<{ x: number; y: number; visible: boolean }>`
  position: absolute;
  left: ${props => props.x}px;
  top: ${props => props.y}px;
  background: #333;
  color: #fff;
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
  pointer-events: none;
  opacity: ${props => (props.visible ? 0.9 : 0)};
  transition: opacity 0.3s;
  z-index: 1000;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
`;

interface NodeTooltipProps {
  node: D3Node | null;
  position: { x: number; y: number };
}

const NodeTooltip: React.FC<NodeTooltipProps> = ({ node, position }) => {
  if (!node) {
    return null;
  }

  return (
    <TooltipContainer
      x={position.x + 10}
      y={position.y - 28}
      visible={!!node}
    >
      <div>
        <strong>{node.name}</strong> ({node.type})
      </div>
      <div>Значение: {node.value}</div>
      <div>Файл: {node.filePath}</div>
      <div>Строка: {node.lineNum}</div>
    </TooltipContainer>
  );
};

export default NodeTooltip;
