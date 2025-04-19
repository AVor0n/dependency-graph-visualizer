import { useState, useEffect, useRef } from 'react';
import styled from 'styled-components';
import * as d3 from 'd3';
import GraphNode from './GraphNode';
import GraphLink from './GraphLink';
import NodeTooltip from './NodeTooltip';
import { D3Node, D3Link, GraphData } from './types';

const Canvas = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
`;

interface GraphCanvasProps {
  data: GraphData;
  onNodeHover?: (node: D3Node | null) => void;
}

const GraphCanvas: React.FC<GraphCanvasProps> = ({ data, onNodeHover }) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [simulation, setSimulation] = useState<d3.Simulation<D3Node, d3.SimulationLinkDatum<D3Node>> | null>(null);
  const [hoveredNode, setHoveredNode] = useState<D3Node | null>(null);
  const [tooltipPosition, setTooltipPosition] = useState<{ x: number; y: number }>({ x: 0, y: 0 });

  // Инициализация симуляции при изменении данных
  useEffect(() => {
    if (!data || !svgRef.current) return;

    const svg = d3.select(svgRef.current);
    const width = svgRef.current.clientWidth || 800;
    const height = svgRef.current.clientHeight || 600;

    // Создаем симуляцию
    const sim = d3.forceSimulation<D3Node>(data.nodes)
      .force('link', d3.forceLink<D3Node, D3Link>(data.links).id(d => d.id).distance(100))
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(30));

    setSimulation(sim);

    // Добавляем зум
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 4])
      .on('zoom', (event: d3.D3ZoomEvent<SVGSVGElement, unknown>) => {
        svg.select('g.graph-container').attr('transform', event.transform.toString());
      });

    svg.call(zoom);

    return () => {
      sim.stop();
    };
  }, [data]);

  // Обработчик наведения на узел
  const handleNodeHover = (node: D3Node | null, event?: React.MouseEvent) => {
    setHoveredNode(node);
    if (node && event) {
      setTooltipPosition({ x: event.clientX, y: event.clientY });
    }

    // Передаем информацию родительскому компоненту, если определен обработчик
    if (onNodeHover) {
      onNodeHover(node);
    }
  };

  return (
    <Canvas>
      <svg ref={svgRef} width="100%" height="100%">
        <defs>
          <marker
            id="arrow"
            viewBox="0 -5 10 10"
            refX={25}
            refY={0}
            markerWidth={6}
            markerHeight={6}
            orient="auto"
          >
            <path d="M0,-5L10,0L0,5" fill="#999" />
          </marker>
        </defs>
        <g className="graph-container">
          <g className="links">
            {data.links.map((link, i) => (
              <GraphLink
                key={`link-${i}`}
                link={link}
                hoveredNodeId={hoveredNode?.id || null}
              />
            ))}
          </g>
          <g className="nodes">
            {data.nodes.map(node => (
              <GraphNode
                key={node.id}
                node={node}
                simulation={simulation!}
                onNodeHover={handleNodeHover}
              />
            ))}
          </g>
          <g className="node-labels">
            {data.nodes.map(node => (
              <text
                key={`text-${node.id}`}
                x={node.x || 0}
                y={node.y || 0}
                dx={12}
                dy=".35em"
                fill="#d4d4d4"
                fontSize={12}
                pointerEvents="none"
                opacity={hoveredNode ? (hoveredNode.id === node.id ? 1 : 0.2) : 1}
              >
                {node.name}
              </text>
            ))}
          </g>
        </g>
      </svg>
      <NodeTooltip node={hoveredNode} position={tooltipPosition} />
    </Canvas>
  );
};

export default GraphCanvas;
