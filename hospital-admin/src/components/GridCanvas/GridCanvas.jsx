import { useState, useRef, useCallback, useMemo, useEffect } from 'react';
import { Stage, Layer, Rect, Circle, Text as KonvaText, Image as KonvaImage, Line } from 'react-konva';
import { Button, Tooltip } from 'antd';
import { DownloadOutlined } from '@ant-design/icons';

// ─── Constants ────────────────────────────────────────────────
const MIN_SCALE = 0.15;
const MAX_SCALE = 6;

// Color map for POI types
const POI_COLORS = {
  entrance: '#52c41a',
  room: '#1677ff',
  elevator: '#faad14',
  canteen: '#fa8c16',
  pharmacy: '#13c2c2',
  info: '#722ed1',
  toilet: '#8c8c8c',
  wc: '#8c8c8c',
  stair: '#eb2f96',
  stairs: '#eb2f96',
  parking: '#2f54eb',
  corridor: '#91caff',
  wifi: '#389e0d',
  other: '#bfbfbf',
  default: '#bfbfbf',
};

function getPOIColor(poiType) {
  return POI_COLORS[poiType] || POI_COLORS.default;
}

// ─── Pre-render grid background as offscreen canvas ───────────
// This is critical for large grids (140×500 = 70K cells).
// Instead of creating 70K Konva.Rect, we paint once on a canvas.
function buildGridImage(rows, cols, cellSize, heatmapData, pathCells, gridDataStr) {
  const canvas = document.createElement('canvas');
  canvas.width = cols * cellSize;
  canvas.height = rows * cellSize;
  const ctx = canvas.getContext('2d');

  // Base grid
  ctx.fillStyle = '#f0f0f0';
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  // Grid lines (subtle)
  ctx.strokeStyle = '#e8e8e8';
  ctx.lineWidth = 0.5;
  for (let r = 0; r <= rows; r++) {
    ctx.beginPath();
    ctx.moveTo(0, r * cellSize);
    ctx.lineTo(cols * cellSize, r * cellSize);
    ctx.stroke();
  }
  for (let c = 0; c <= cols; c++) {
    ctx.beginPath();
    ctx.moveTo(c * cellSize, 0);
    ctx.lineTo(c * cellSize, rows * cellSize);
    ctx.stroke();
  }

  // Draw walls from gridData
  if (gridDataStr) {
    try {
      // Handle both JSON string and already-parsed array
      const grid = typeof gridDataStr === 'string' ? JSON.parse(gridDataStr) : gridDataStr;
      if (Array.isArray(grid)) {
        ctx.fillStyle = '#595959'; // Wall color (darker for better contrast)
        for (let r = 0; r < Math.min(rows, grid.length); r++) {
          for (let c = 0; c < Math.min(cols, grid[r].length); c++) {
            if (grid[r][c] === 1) { // 1 = obstacle/wall
              ctx.fillRect(c * cellSize, r * cellSize, cellSize, cellSize);
            }
          }
        }
      }
    } catch (e) {
      console.error('Failed to parse gridData', e);
    }
  }

  // Heatmap overlay
  if (heatmapData && heatmapData.length > 0) {
    const maxDensity = Math.max(...heatmapData.map((d) => d.density), 1);
    for (const d of heatmapData) {
      const r = Math.floor(d.grid_location / cols);
      const c = d.grid_location % cols;
      const intensity = Math.min(d.density / maxDensity, 1);
      ctx.fillStyle = `rgba(255, ${Math.round(255 - intensity * 200)}, 50, ${0.3 + intensity * 0.5})`;
      ctx.fillRect(c * cellSize, r * cellSize, cellSize, cellSize);
    }
  }

  // Path cells
  const pathSet = new Set(pathCells || []);
  if (pathSet.size > 0) {
    ctx.fillStyle = 'rgba(186, 224, 255, 0.7)';
    ctx.strokeStyle = '#1677ff';
    ctx.lineWidth = 1;
    for (const loc of pathSet) {
      const r = Math.floor(loc / cols);
      const c = loc % cols;
      ctx.fillRect(c * cellSize, r * cellSize, cellSize, cellSize);
      ctx.strokeRect(c * cellSize, r * cellSize, cellSize, cellSize);
    }
  }

  return canvas;
}

// ─── GridCanvas Component ─────────────────────────────────────
export default function GridCanvas({
  rows = 33,
  cols = 57,
  gridData = null,
  nodes = [],
  heatmapData = [],
  pathCells = [],
  agentPositions = [],
  selectedNodeId = null,
  highlightNodeId = null,
  onNodeClick,
  onCellClick,
  onCellPaint,
  editMode = false,
  width = 1100,
  height = 620,
}) {
  const stageRef = useRef(null);
  const [stagePos, setStagePos] = useState({ x: 0, y: 0 });
  const [scale, setScale] = useState(1);
  const [hoveredNode, setHoveredNode] = useState(null);
  const [tooltipPos, setTooltipPos] = useState({ x: 0, y: 0 });
  const [isPainting, setIsPainting] = useState(false);
  const lastPaintedCell = useRef(null);

  // Dynamic cell size: fit large grids
  const cellSize = useMemo(() => {
    // Target: grid fits within canvas at scale ~1
    const maxCellW = (width - 20) / cols;
    const maxCellH = (height - 20) / rows;
    const cs = Math.min(maxCellW, maxCellH);
    // Clamp: min 2px (huge grids), max 24px (small grids)
    return Math.max(2, Math.min(24, Math.floor(cs)));
  }, [rows, cols, width, height]);

  // Build node lookup: grid_location → node
  const nodeMap = useMemo(() => {
    const map = new Map();
    (nodes || []).forEach((n) => {
      if (n.grid_location != null) {
        map.set(n.grid_location, n);
      }
    });
    return map;
  }, [nodes]);

  // Pre-render grid background image (offscreen canvas)
  const gridImage = useMemo(() => {
    return buildGridImage(rows, cols, cellSize, heatmapData, pathCells, gridData);
  }, [rows, cols, cellSize, heatmapData, pathCells, gridData]);

  // Zoom with mouse wheel
  const handleWheel = useCallback((e) => {
    e.evt.preventDefault();
    const stage = stageRef.current;
    if (!stage) return;

    const oldScale = stage.scaleX();
    const pointer = stage.getPointerPosition();

    const mousePointTo = {
      x: (pointer.x - stage.x()) / oldScale,
      y: (pointer.y - stage.y()) / oldScale,
    };

    const direction = e.evt.deltaY > 0 ? -1 : 1;
    const factor = oldScale > 2 ? 0.3 : 0.15;
    const newScale = Math.min(MAX_SCALE, Math.max(MIN_SCALE, oldScale + direction * factor));

    setScale(newScale);
    setStagePos({
      x: pointer.x - mousePointTo.x * newScale,
      y: pointer.y - mousePointTo.y * newScale,
    });
  }, []);

  // Fit grid to container on mount or when dimensions change
  useEffect(() => {
    const gridW = cols * cellSize;
    const gridH = rows * cellSize;
    const scaleX = (width - 10) / gridW;
    const scaleY = (height - 10) / gridH;
    const fitScale = Math.min(scaleX, scaleY, 3);
    setScale(fitScale);
    setStagePos({
      x: (width - gridW * fitScale) / 2,
      y: (height - gridH * fitScale) / 2,
    });
  }, [rows, cols, cellSize, width, height]);

  // ─── POI Markers (interactive — only ~10-20 elements) ────────
  const poiMarkers = useMemo(() => {
    return (nodes || []).map((n) => {
      const cx = n.grid_col * cellSize + cellSize / 2;
      const cy = n.grid_row * cellSize + cellSize / 2;
      const isSelected = n.poi_id === selectedNodeId;
      const isHighlighted = n.poi_id === highlightNodeId;
      const radius = Math.max(cellSize * 0.6, 5);

      return (
        <Circle
          key={`poi-${n.poi_id}`}
          x={cx}
          y={cy}
          radius={radius}
          fill={getPOIColor(n.poi_type)}
          stroke={isHighlighted ? '#ff4d4f' : isSelected ? '#000' : '#fff'}
          strokeWidth={isHighlighted || isSelected ? 3 : 1.5}
          shadowColor="rgba(0,0,0,0.25)"
          shadowBlur={4}
          shadowOffset={{ x: 1, y: 1 }}
          onMouseEnter={(e) => {
            const stage = e.target.getStage();
            const pointer = stage.getPointerPosition();
            setHoveredNode(n);
            setTooltipPos({ x: pointer.x, y: pointer.y });
            stage.container().style.cursor = 'pointer';
          }}
          onMouseLeave={(e) => {
            setHoveredNode(null);
            e.target.getStage().container().style.cursor = 'default';
          }}
          onClick={() => {
            if (onNodeClick) onNodeClick(n);
          }}
        />
      );
    });
  }, [nodes, cellSize, selectedNodeId, highlightNodeId, onNodeClick]);

  // ─── POI Labels (landmarks only, for clarity) ────────────────
  const poiLabels = useMemo(() => {
    // Only show labels if cellSize >= 6 (otherwise too small)
    if (cellSize < 6) return [];
    return (nodes || []).filter((n) => n.is_landmark).map((n) => (
      <KonvaText
        key={`label-${n.poi_id}`}
        x={n.grid_col * cellSize + cellSize + 3}
        y={n.grid_row * cellSize - 2}
        text={n.poi_code || n.poi_name?.slice(0, 10)}
        fontSize={Math.max(8, Math.min(11, cellSize * 0.7))}
        fill="#333"
        fontStyle="bold"
      />
    ));
  }, [nodes, cellSize]);

  // ─── Agent Dots (MAPF) ──────────────────────────────────────
  const AGENT_COLORS = ['#f5222d', '#1890ff', '#52c41a', '#faad14', '#722ed1', '#eb2f96', '#13c2c2', '#fa541c', '#2f54eb', '#a0d911'];

  const agentDots = useMemo(() => {
    return (agentPositions || []).map((agent, idx) => {
      const loc = agent.location ?? agent.grid_location;
      if (loc == null) return null;
      const r = Math.floor(loc / cols);
      const c = loc % cols;
      return (
        <Circle
          key={`agent-${agent.agent_id ?? idx}`}
          x={c * cellSize + cellSize / 2}
          y={r * cellSize + cellSize / 2}
          radius={Math.max(cellSize / 2.5, 3)}
          fill={AGENT_COLORS[idx % AGENT_COLORS.length]}
          stroke="#fff"
          strokeWidth={1.5}
          shadowColor="rgba(0,0,0,0.3)"
          shadowBlur={4}
        />
      );
    });
  }, [agentPositions, cols, cellSize]);

  // ─── Cell coordinate helper ─────────────────────────────────
  const getCellFromPointer = useCallback(() => {
    const stage = stageRef.current;
    if (!stage) return null;
    const pointer = stage.getPointerPosition();
    if (!pointer) return null;
    const x = (pointer.x - stagePos.x) / scale;
    const y = (pointer.y - stagePos.y) / scale;
    const col = Math.floor(x / cellSize);
    const row = Math.floor(y / cellSize);
    if (row >= 0 && row < rows && col >= 0 && col < cols) return { row, col };
    return null;
  }, [stagePos, scale, cellSize, rows, cols]);

  // ─── Click on empty grid area ───────────────────────────────
  const handleStageClick = useCallback((e) => {
    if (editMode) return; // painting handles clicks in editMode
    if (e.target === e.target.getStage() || e.target.attrs?.image) {
      if (!onCellClick) return;
      const cell = getCellFromPointer();
      if (cell) onCellClick(cell.row * cols + cell.col, cell.row, cell.col);
    }
  }, [editMode, onCellClick, getCellFromPointer, cols]);

  // ─── Paint handlers (editMode) ──────────────────────────────
  const handlePaintStart = useCallback((e) => {
    if (!editMode || !onCellPaint) return;
    if (e.target !== e.target.getStage() && !e.target.attrs?.image) return;
    setIsPainting(true);
    const cell = getCellFromPointer();
    if (cell) {
      lastPaintedCell.current = `${cell.row},${cell.col}`;
      onCellPaint(cell.row, cell.col);
    }
  }, [editMode, onCellPaint, getCellFromPointer]);

  const handlePaintMove = useCallback(() => {
    if (!isPainting || !editMode || !onCellPaint) return;
    const cell = getCellFromPointer();
    if (cell) {
      const key = `${cell.row},${cell.col}`;
      if (key !== lastPaintedCell.current) {
        lastPaintedCell.current = key;
        onCellPaint(cell.row, cell.col);
      }
    }
  }, [isPainting, editMode, onCellPaint, getCellFromPointer]);

  const handlePaintEnd = useCallback(() => {
    setIsPainting(false);
    lastPaintedCell.current = null;
  }, []);

  const handleExport = useCallback(() => {
    if (!stageRef.current) return;
    const uri = stageRef.current.toDataURL({ pixelRatio: 2 });
    const link = document.createElement('a');
    link.download = `hospital_map_${rows}x${cols}_${new Date().getTime()}.png`;
    link.href = uri;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }, [rows, cols]);

  return (
    <div style={{ position: 'relative', border: '1px solid #d9d9d9', borderRadius: 8, overflow: 'hidden', background: editMode ? '#fff' : '#fafafa' }}>
      <Stage
        ref={stageRef}
        width={width}
        height={height}
        scaleX={scale}
        scaleY={scale}
        x={stagePos.x}
        y={stagePos.y}
        draggable={!editMode}
        onWheel={handleWheel}
        onDragEnd={(e) => {
          if (!editMode) setStagePos({ x: e.target.x(), y: e.target.y() });
        }}
        onClick={handleStageClick}
        onMouseDown={handlePaintStart}
        onMouseMove={handlePaintMove}
        onMouseUp={handlePaintEnd}
        onMouseLeave={handlePaintEnd}
      >
        {/* Layer 1: Pre-rendered grid background (single canvas image) */}
        <Layer listening={false}>
          <KonvaImage image={gridImage} x={0} y={0} />
        </Layer>

        {/* Layer 2: Interactive POI markers + labels */}
        <Layer>
          {poiMarkers}
          {poiLabels}
        </Layer>

        {/* Layer 3: Agent dots (MAPF) */}
        {agentDots.length > 0 && (
          <Layer>
            {agentDots}
          </Layer>
        )}
      </Stage>

      {/* Tooltip overlay (HTML — avoids Konva re-render) */}
      {hoveredNode && (
        <div
          style={{
            position: 'absolute',
            left: Math.min(tooltipPos.x + 12, width - 230),
            top: Math.max(tooltipPos.y - 50, 4),
            background: 'rgba(0,0,0,0.85)',
            color: '#fff',
            padding: '8px 14px',
            borderRadius: 8,
            fontSize: 12,
            pointerEvents: 'none',
            zIndex: 10,
            maxWidth: 240,
            lineHeight: 1.6,
            boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
          }}
        >
          <div style={{ fontWeight: 700, fontSize: 13 }}>{hoveredNode.poi_name}</div>
          <div style={{ opacity: 0.75, marginTop: 2 }}>
            {hoveredNode.poi_code} · {hoveredNode.poi_type}
          </div>
          <div style={{ opacity: 0.6 }}>
            Grid: ({hoveredNode.grid_row}, {hoveredNode.grid_col}) = {hoveredNode.grid_location}
          </div>
          {hoveredNode.capacity && (
            <div style={{ opacity: 0.6 }}>Capacity: {hoveredNode.capacity}</div>
          )}
        </div>
      )}

      {/* Grid info badge & Export button */}
      <div
        style={{
          position: 'absolute',
          bottom: 6,
          right: 10,
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}
      >
        <Tooltip title="Xuất ảnh bản đồ (PNG)">
          <Button 
            icon={<DownloadOutlined />} 
            size="small" 
            onClick={handleExport}
            style={{ boxShadow: '0 2px 4px rgba(0,0,0,0.2)' }}
          />
        </Tooltip>
        <div
          style={{
            background: 'rgba(0,0,0,0.55)',
            color: '#fff',
            padding: '2px 8px',
            borderRadius: 4,
            fontSize: 10,
            pointerEvents: 'none',
          }}
        >
          {rows}×{cols} · cell={cellSize}px · zoom={Math.round(scale * 100)}%
        </div>
      </div>
    </div>
  );
}
