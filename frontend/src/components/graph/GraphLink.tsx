import { D3Link } from './types';

interface GraphLinkProps {
  link: D3Link;
  hoveredNodeId: string | null;
}

const GraphLink: React.FC<GraphLinkProps> = ({ link, hoveredNodeId }) => {
  const source = typeof link.source === 'object' ? link.source : { x: 0, y: 0 };
  const target = typeof link.target === 'object' ? link.target : { x: 0, y: 0 };

  const isHighlighted = hoveredNodeId && (
    (typeof link.source === 'object' && link.source.id === hoveredNodeId) ||
    (typeof link.target === 'object' && link.target.id === hoveredNodeId) ||
    (typeof link.source === 'string' && link.source === hoveredNodeId) ||
    (typeof link.target === 'string' && link.target === hoveredNodeId)
  );

  return (
    <line
      x1={source.x || 0}
      y1={source.y || 0}
      x2={target.x || 0}
      y2={target.y || 0}
      stroke={isHighlighted ? '#fff' : '#999'}
      strokeOpacity={isHighlighted ? 1 : 0.6}
      strokeWidth={isHighlighted ? 2 : 1}
      markerEnd="url(#arrow)"
    />
  );
};

export default GraphLink;
