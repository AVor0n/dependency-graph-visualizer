// @ts-nocheck - отключение проверки типов для этого файла
import React, { useState, useEffect, useRef, useCallback } from 'react';
import styled from 'styled-components';
import * as d3 from 'd3';

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

const GraphCanvas = styled.div`
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

// Информация о константе
interface Constant {
  name: string;
  value: string;
  type: string;
  filePath: string;
}

// Информация о зависимости между константами
interface Dependency {
  source: string;
  target: string;
}

// Структура графа зависимостей
interface DependencyGraph {
  nodes: Constant[];
  edges: Dependency[];
}

// Структура для D3 графа
interface D3Node extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  type: string;
  value: string;
  filePath: string;
  color: string;
}

interface D3Link extends d3.SimulationLinkDatum<D3Node> {
  source: string | D3Node;
  target: string | D3Node;
}

interface GraphData {
  nodes: D3Node[];
  links: D3Link[];
}

interface DependencyGraphProps {
  selectedFile?: string;
}

const DependencyGraph: React.FC<DependencyGraphProps> = ({ selectedFile }) => {
  const [graphData, setGraphData] = useState<GraphData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hoverNode, setHoverNode] = useState<D3Node | null>(null);

  const svgRef = useRef<SVGSVGElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  // Функция для преобразования данных из API в формат для D3
  const transformData = (data: DependencyGraph): GraphData => {
    const nodes: D3Node[] = data.nodes.map(node => ({
      id: node.name,
      name: node.name,
      type: node.type,
      value: node.value,
      filePath: node.filePath,
      color: getNodeColor(node.type)
    }));

    const links: D3Link[] = data.edges.map(edge => ({
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
      const response = await fetch('http://localhost:8080/api/dependency-graph');
      if (!response.ok) {
        throw new Error('Не удалось загрузить данные графа');
      }

      const data: DependencyGraph = await response.json();
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
      const response = await fetch('http://localhost:8080/api/file-dependencies', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ filePath })
      });

      if (!response.ok) {
        throw new Error('Не удалось загрузить данные зависимостей для файла');
      }

      const data: DependencyGraph = await response.json();
      const transformedData = transformData(data);
      setGraphData(transformedData);
    } catch (err) {
      console.error('Error fetching file dependencies:', err);
      setError('Ошибка при загрузке зависимостей для файла');
    } finally {
      setLoading(false);
    }
  };

  // Отрисовка графа при изменении данных
  useEffect(() => {
    if (!graphData || !svgRef.current) return;

    // Очищаем предыдущий граф
    d3.select(svgRef.current).selectAll('*').remove();

    const svg = d3.select(svgRef.current);
    const width = svgRef.current.clientWidth;
    const height = svgRef.current.clientHeight;

    // Создаем симуляцию с силами
    const simulation = d3.forceSimulation<D3Node, D3Link>(graphData.nodes)
      .force('link', d3.forceLink<D3Node, D3Link>(graphData.links).id(d => d.id).distance(100))
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(30));

    // Создаем стрелки для ребер
    svg.append('defs').selectAll('marker')
      .data(['end'])
      .enter().append('marker')
      .attr('id', 'arrow')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 25)
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('path')
      .attr('fill', '#999')
      .attr('d', 'M0,-5L10,0L0,5');

    // Создаем ребра
    const link = svg.append('g')
      .selectAll('line')
      .data(graphData.links)
      .enter().append('line')
      .attr('stroke', '#999')
      .attr('stroke-opacity', 0.6)
      .attr('stroke-width', 1)
      .attr('marker-end', 'url(#arrow)');

    // Создаем узлы
    const node = svg.append('g')
      .selectAll('circle')
      .data(graphData.nodes)
      .enter().append('circle')
      .attr('r', 10)
      .attr('fill', d => d.color)
      .call(d3.drag<SVGCircleElement, D3Node>()
        .on('start', dragstarted)
        .on('drag', dragged)
        .on('end', dragended) as any);

    // Добавляем текст к узлам
    const text = svg.append('g')
      .selectAll('text')
      .data(graphData.nodes)
      .enter().append('text')
      .attr('dx', 12)
      .attr('dy', '.35em')
      .text(d => d.name)
      .style('fill', '#d4d4d4')
      .style('font-size', '12px')
      .style('pointer-events', 'none');

    // Создаем подсказку
    if (!tooltipRef.current) {
      const tooltipDiv = document.createElement('div');
      tooltipDiv.style.position = 'absolute';
      tooltipDiv.style.padding = '10px';
      tooltipDiv.style.background = '#333';
      tooltipDiv.style.color = '#fff';
      tooltipDiv.style.borderRadius = '4px';
      tooltipDiv.style.pointerEvents = 'none';
      tooltipDiv.style.opacity = '0';
      tooltipDiv.style.transition = 'opacity 0.3s';
      tooltipDiv.style.zIndex = '1000';

      const container = svgRef.current.parentNode;
      if (container) {
        container.appendChild(tooltipDiv);
        tooltipRef.current = tooltipDiv;
      }
    }

    // Обработчики событий для интерактивности
    node
      .on('mouseover', function(event, d) {
        setHoverNode(d);

        // Увеличиваем размер узла при наведении
        d3.select(this)
          .transition()
          .duration(200)
          .attr('r', 15);

        // Показываем подсказку
        if (tooltipRef.current) {
          tooltipRef.current.innerHTML = `
            <div><strong>${d.name}</strong> (${d.type})</div>
            <div>Значение: ${d.value}</div>
          `;
          tooltipRef.current.style.opacity = '0.9';
          tooltipRef.current.style.left = (event.pageX + 10) + 'px';
          tooltipRef.current.style.top = (event.pageY - 28) + 'px';
        }

        // Выделяем связанные ребра
        link
          .style('stroke', function(l: any) {
            if (l.source === d || l.target === d) {
              return '#fff';
            } else {
              return '#999';
            }
          })
          .style('stroke-opacity', function(l: any) {
            if (l.source === d || l.target === d) {
              return 1;
            } else {
              return 0.2;
            }
          })
          .style('stroke-width', function(l: any) {
            if (l.source === d || l.target === d) {
              return 2;
            } else {
              return 1;
            }
          });

        // Выделяем связанные узлы
        node
          .style('opacity', function(o: D3Node) {
            const isConnected = graphData?.links.some(
              l => (l.source === d && l.target === o) || (l.source === o && l.target === d)
            );
            return isConnected || o === d ? 1 : 0.2;
          });

        // Выделяем связанные тексты
        text
          .style('opacity', function(o: D3Node) {
            const isConnected = graphData?.links.some(
              l => (l.source === d && l.target === o) || (l.source === o && l.target === d)
            );
            return isConnected || o === d ? 1 : 0.2;
          });
      })
      .on('mouseout', function() {
        setHoverNode(null);

        // Возвращаем размер узла
        d3.select(this)
          .transition()
          .duration(200)
          .attr('r', 10);

        // Скрываем подсказку
        if (tooltipRef.current) {
          tooltipRef.current.style.opacity = '0';
        }

        // Возвращаем нормальный вид ребер
        link
          .style('stroke', '#999')
          .style('stroke-opacity', 0.6)
          .style('stroke-width', 1);

        // Возвращаем нормальный вид узлов
        node.style('opacity', 1);

        // Возвращаем нормальный вид текстов
        text.style('opacity', 1);
      })
      .on('click', function(event, d) {
        // Останавливаем распространение события
        event.stopPropagation();

        // Если узел уже выбран, снимаем выделение
        if (d === selectedNode) {
          setSelectedNode(null);
          return;
        }

        // Выбираем узел
        setSelectedNode(d);

        // Центрируем граф на выбранном узле
        const transform = d3.zoomTransform(svg.node() as any);
        const scale = transform.k;
        const x = -d.x * scale + width / 2;
        const y = -d.y * scale + height / 2;

        svg.transition()
          .duration(750)
          .call(zoom.transform, d3.zoomIdentity.translate(x, y).scale(scale));
      });

    // Обработчики для перетаскивания узлов
    function dragstarted(event: any, d: D3Node) {
      if (!event.active) simulation.alphaTarget(0.3).restart();
      d.fx = d.x;
      d.fy = d.y;
    }

    function dragged(event: any, d: D3Node) {
      d.fx = event.x;
      d.fy = event.y;
    }

    function dragended(event: any, d: D3Node) {
      if (!event.active) simulation.alphaTarget(0);
      d.fx = null;
      d.fy = null;
    }

    // Добавляем зум
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 4])
      .on('zoom', (event) => {
        svg.selectAll('g').attr('transform', event.transform.toString());
      });

    svg.call(zoom);

    // Обновляем позиции при каждом тике симуляции
    simulation.on('tick', () => {
      link
        .attr('x1', d => (d.source as D3Node).x!)
        .attr('y1', d => (d.source as D3Node).y!)
        .attr('x2', d => (d.target as D3Node).x!)
        .attr('y2', d => (d.target as D3Node).y!);

      node
        .attr('cx', d => d.x!)
        .attr('cy', d => d.y!);

      text
        .attr('x', d => d.x!)
        .attr('y', d => d.y!);
    });

    // Устанавливаем ссылку на узел для D3 в состоянии компонента
    let selectedNode: D3Node | null = null;

    // Очистка при размонтировании
    return () => {
      simulation.stop();
      if (tooltipRef.current && tooltipRef.current.parentNode) {
        tooltipRef.current.parentNode.removeChild(tooltipRef.current);
        tooltipRef.current = null;
      }
    };
  }, [graphData]);

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
        {hoverNode && (
          <div>
            <strong>{hoverNode.name}</strong> - {hoverNode.type}
          </div>
        )}
      </GraphHeader>
      <GraphCanvas>
        <svg ref={svgRef} width="100%" height="100%" />
      </GraphCanvas>
    </GraphContainer>
  );
};

export default DependencyGraph;
