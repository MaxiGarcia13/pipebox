// Pixel sizes match job-timeline: h-10 cards, gap-4 rows, w-10 connector.
const JOB_CARD_HEIGHT = 40;
const JOB_ROW_GAP = 16;
const CONNECTOR_WIDTH = 40;
const DOT_RADIUS = 4;

export interface JobGraphConnectorJob {
  id: string;
  row: number;
  needs: string[];
}

export interface JobGraphConnectorDot {
  x: number;
  y: number;
}

export interface JobGraphConnector {
  width: number;
  height: number;
  lines: string[];
  dots: JobGraphConnectorDot[];
}

interface GraphEdge {
  sourceId: string;
  targetId: string;
  sourceY: number;
  targetY: number;
}

interface ConnectorComponent {
  edges: GraphEdge[];
  sourceYs: number[];
  targetYs: number[];
  minY: number;
  maxY: number;
}

export function buildJobGraphConnector(
  from: JobGraphConnectorJob[],
  to: JobGraphConnectorJob[],
): JobGraphConnector {
  const height = Math.max(columnHeight(from), columnHeight(to));
  const components = groupComponents(edgesBetween(from, to, height));
  if (components.length === 0) {
    return {
      width: CONNECTOR_WIDTH,
      height,
      lines: [],
      dots: [],
    };
  }

  const { lanes, laneCount } = assignLanes(components);
  const lines: string[] = [];
  const dots: JobGraphConnectorDot[] = [];

  components.forEach((component, index) => {
    const busX = busXForLane(lanes[index] ?? 0, laneCount);
    lines.push(...linesFor(component, busX));
    dots.push({
      x: busX,
      y: dotY(component),
    });
  });

  return {
    width: CONNECTOR_WIDTH,
    height,
    lines,
    dots,
  };
}

function columnHeight(jobs: { row: number }[]): number {
  if (jobs.length === 0) {
    return 0;
  }
  const maxRow = jobs.reduce((max, job) => Math.max(max, job.row), 0);
  return (maxRow + 1) * JOB_CARD_HEIGHT + maxRow * JOB_ROW_GAP;
}

function rowCenter(row: number, jobs: { row: number }[], graphHeight: number): number {
  // Same offset as items-center on the job row: shorter columns sit mid-span.
  const offset = (graphHeight - columnHeight(jobs)) / 2;
  return offset + row * (JOB_CARD_HEIGHT + JOB_ROW_GAP) + JOB_CARD_HEIGHT / 2;
}

function edgesBetween(
  from: JobGraphConnectorJob[],
  to: JobGraphConnectorJob[],
  height: number,
): GraphEdge[] {
  const fromById = new Map(from.map((job) => [job.id, job]));
  const seen = new Set<string>();
  const edges: GraphEdge[] = [];

  for (const target of to) {
    for (const needId of target.needs) {
      const source = fromById.get(needId);
      const key = `${needId}->${target.id}`;
      if (!source || seen.has(key)) {
        continue;
      }
      seen.add(key);
      edges.push({
        sourceId: source.id,
        targetId: target.id,
        sourceY: rowCenter(source.row, from, height),
        targetY: rowCenter(target.row, to, height),
      });
    }
  }

  return edges;
}

function groupComponents(edges: GraphEdge[]): ConnectorComponent[] {
  const parent = new Map<string, string>();

  function find(key: string): string {
    let root = key;
    const rootParent = parent.get(root);
    if (rootParent == null) {
      parent.set(root, root);
      return root;
    }
    while (parent.get(root) !== root) {
      root = parent.get(root) ?? root;
    }
    let current = key;
    while (current !== root) {
      const next = parent.get(current) ?? root;
      parent.set(current, root);
      current = next;
    }
    return root;
  }

  function union(left: string, right: string): void {
    const leftRoot = find(left);
    const rightRoot = find(right);
    if (leftRoot !== rightRoot) {
      parent.set(leftRoot, rightRoot);
    }
  }

  for (const edge of edges) {
    union(`from:${edge.sourceId}`, `to:${edge.targetId}`);
  }

  const grouped = new Map<string, GraphEdge[]>();
  for (const edge of edges) {
    const root = find(`from:${edge.sourceId}`);
    const group = grouped.get(root);
    if (group) {
      group.push(edge);
    } else {
      grouped.set(root, [edge]);
    }
  }

  return [
    ...grouped.values(),
  ]
    .map((group) => toComponent(group))
    .sort((left, right) =>
      left.minY - right.minY
      || left.maxY - right.maxY
      || left.edges[0]!.sourceId.localeCompare(right.edges[0]!.sourceId)
      || left.edges[0]!.targetId.localeCompare(right.edges[0]!.targetId),
    );
}

function toComponent(edges: GraphEdge[]): ConnectorComponent {
  const sorted = [...edges].sort((left, right) =>
    left.sourceY - right.sourceY
    || left.targetY - right.targetY
    || left.sourceId.localeCompare(right.sourceId)
    || left.targetId.localeCompare(right.targetId),
  );
  const sourceYs = uniqueYs(sorted.map((edge) => edge.sourceY));
  const targetYs = uniqueYs(sorted.map((edge) => edge.targetY));
  const ys = [
    ...sourceYs,
    ...targetYs,
  ];
  return {
    edges: sorted,
    sourceYs,
    targetYs,
    minY: Math.min(...ys),
    maxY: Math.max(...ys),
  };
}

function uniqueYs(values: number[]): number[] {
  return [
    ...new Set(values),
  ].sort((left, right) => left - right);
}

function assignLanes(components: ConnectorComponent[]): { lanes: number[]; laneCount: number } {
  const occupied: { minY: number; maxY: number }[][] = [];
  const lanes = components.map((component) => {
    for (let lane = 0; lane < occupied.length; lane += 1) {
      const ranges = occupied[lane] ?? [];
      const overlaps = ranges.some((range) =>
        component.minY <= range.maxY && range.minY <= component.maxY,
      );
      if (!overlaps) {
        ranges.push({ minY: component.minY, maxY: component.maxY });
        return lane;
      }
    }
    occupied.push([
      { minY: component.minY, maxY: component.maxY },
    ]);
    return occupied.length - 1;
  });
  return { lanes, laneCount: occupied.length };
}

function busXForLane(lane: number, laneCount: number): number {
  if (laneCount === 1) {
    return CONNECTOR_WIDTH / 2;
  }
  const inset = DOT_RADIUS + 8;
  const span = CONNECTOR_WIDTH - inset * 2;
  return inset + (span * lane) / (laneCount - 1);
}

function linesFor(component: ConnectorComponent, busX: number): string[] {
  const { sourceYs, targetYs, minY, maxY } = component;
  if (sourceYs.length === 1 && targetYs.length === 1 && sourceYs[0] === targetYs[0]) {
    return [`M0 ${sourceYs[0]}H${CONNECTOR_WIDTH}`];
  }

  const lines = [
    ...sourceYs.map((y) => `M0 ${y}H${busX}`),
    ...targetYs.map((y) => `M${busX} ${y}H${CONNECTOR_WIDTH}`),
  ];
  if (minY !== maxY) {
    lines.push(`M${busX} ${minY}V${maxY}`);
  }
  return lines;
}

function dotY(component: ConnectorComponent): number {
  const { sourceYs, targetYs, minY, maxY } = component;
  if (sourceYs.length === 1 && targetYs.length === 1) {
    return (sourceYs[0]! + targetYs[0]!) / 2;
  }
  if (targetYs.length === 1) {
    return targetYs[0]!;
  }
  if (sourceYs.length === 1) {
    return sourceYs[0]!;
  }
  return (minY + maxY) / 2;
}
