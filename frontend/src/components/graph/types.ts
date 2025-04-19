import * as d3 from 'd3';

// Типы данных для работы с графом зависимостей
export interface D3Node extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  type: string;
  value: string;
  filePath: string;
  lineNum: number;
  color: string;
}

export interface D3Link extends d3.SimulationLinkDatum<D3Node> {
  source: string | D3Node;
  target: string | D3Node;
}

export interface GraphData {
  nodes: D3Node[];
  links: D3Link[];
}
