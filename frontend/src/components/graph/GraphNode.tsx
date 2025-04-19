import { useRef, useEffect } from 'react';
import * as d3 from 'd3';
import { D3Node } from '../graph/types';

interface GraphNodeProps {
  node: D3Node;
  simulation: d3.Simulation<D3Node, d3.SimulationLinkDatum<D3Node>>;
  onNodeHover: (node: D3Node | null, event?: React.MouseEvent) => void;
}

const GraphNode: React.FC<GraphNodeProps> = ({
  node,
  simulation,
  onNodeHover
}) => {
  const nodeRef = useRef<SVGCircleElement>(null);

  useEffect(() => {
    if (!nodeRef.current) return;

    const nodeSel = d3.select<SVGCircleElement, D3Node>(nodeRef.current);

    // Драг-поведение
    const drag = d3.drag<SVGCircleElement, D3Node>()
      .on('start', (event) => {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        node.fx = node.x;
        node.fy = node.y;
      })
      .on('drag', (event) => {
        node.fx = event.x;
        node.fy = event.y;
      })
      .on('end', (event) => {
        if (!event.active) simulation.alphaTarget(0);
        node.fx = null;
        node.fy = null;
      });

    nodeSel.call(drag);

    return () => {
      // Очистка при размонтировании
      nodeSel.on('.drag', null);
    };
  }, [node, simulation]);

  return (
    <circle
      ref={nodeRef}
      r={10}
      fill={node.color}
      cx={node.x || 0}
      cy={node.y || 0}
      onMouseOver={(event) => onNodeHover(node, event)}
      onMouseOut={() => onNodeHover(null)}
    />
  );
};

export default GraphNode;
