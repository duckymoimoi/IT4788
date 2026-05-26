/**
 * Map file + PNG preview export utilities.
 * PNG (lossless) — sharp grid walls; JPEG would blur cell edges.
 */

const MAX_CANVAS_DIM = 8192;
const MAX_CELL_PX = 24;

/** Loại POI khuyến nghị cho bản đồ bệnh viện (dùng kiểm tra đủ POI). */
export const RECOMMENDED_POI_TYPES = ['entrance', 'elevator', 'room'];

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

export function getPOIColor(poiType) {
  return POI_COLORS[poiType] || POI_COLORS.default;
}

/** Pixels per cell, scaled so the longest side fits within canvas limits. */
export function computeExportCellSize(rows, cols) {
  const maxDim = Math.max(rows, cols, 1);
  let cellSize = Math.min(MAX_CELL_PX, Math.floor(MAX_CANVAS_DIM / maxDim));
  const logicalW = cols * cellSize;
  const logicalH = rows * cellSize;
  if (Math.max(logicalW, logicalH) > MAX_CANVAS_DIM) {
    const scale = MAX_CANVAS_DIM / Math.max(logicalW, logicalH);
    cellSize = Math.max(1, Math.floor(cellSize * scale));
  }
  return cellSize;
}

function normalizeExportNode(node) {
  if (!node) return null;
  const grid_row = node.grid_row ?? node.row;
  const grid_col = node.grid_col ?? node.col;
  if (grid_row == null || grid_col == null) return null;
  return {
    grid_row,
    grid_col,
    poi_type: node.poi_type ?? node.type ?? 'other',
    poi_code: node.poi_code ?? node.code ?? '',
    poi_name: node.poi_name ?? node.name ?? '',
    is_landmark: !!node.is_landmark,
  };
}

function drawPoisOnContext(ctx, nodes, cellSize) {
  const list = (nodes || []).map(normalizeExportNode).filter(Boolean);
  if (!list.length) return;

  for (const n of list) {
    const cx = n.grid_col * cellSize + cellSize / 2;
    const cy = n.grid_row * cellSize + cellSize / 2;
    const radius = Math.max(cellSize * 0.38, 2.5);

    ctx.beginPath();
    ctx.arc(cx, cy, radius, 0, Math.PI * 2);
    ctx.fillStyle = getPOIColor(n.poi_type);
    ctx.fill();
    ctx.strokeStyle = '#ffffff';
    ctx.lineWidth = Math.max(1, cellSize * 0.1);
    ctx.stroke();
  }

  if (cellSize >= 8) {
    const fontSize = Math.max(8, Math.min(14, Math.floor(cellSize * 0.65)));
    ctx.font = `bold ${fontSize}px sans-serif`;
    ctx.fillStyle = '#262626';
    ctx.textBaseline = 'middle';

    for (const n of list) {
      const label = n.poi_code || (n.poi_name ? n.poi_name.slice(0, 12) : '');
      if (!label) continue;
      const cx = n.grid_col * cellSize + cellSize / 2;
      const cy = n.grid_row * cellSize + cellSize / 2;
      const tx = cx + cellSize * 0.55;
      const ty = cy;
      if (n.is_landmark || cellSize >= 14) {
        ctx.fillText(label, tx, ty);
      }
    }
  }
}

/**
 * Đánh giá map đã có đủ POI (theo GET /map/get_nodes).
 * @returns {{ poi_count, landmark_count, missing_types, status, is_complete }}
 */
export function assessMapPoiCompleteness(nodes) {
  const list = nodes || [];
  const poi_count = list.length;
  const typeSet = new Set(list.map((n) => n.poi_type));
  const missing_types = RECOMMENDED_POI_TYPES.filter((t) => !typeSet.has(t));
  const landmark_count = list.filter((n) => n.is_landmark).length;

  let status = 'empty';
  if (poi_count === 0) {
    status = 'empty';
  } else if (missing_types.length === 0) {
    status = 'complete';
  } else {
    status = 'partial';
  }

  return {
    poi_count,
    landmark_count,
    missing_types,
    status,
    is_complete: status === 'complete',
  };
}

export function gridToMapFile(rows, cols, grid) {
  const lines = ['type octile', `height ${rows}`, `width ${cols}`, 'map'];
  for (let r = 0; r < rows; r++) {
    let line = '';
    for (let c = 0; c < cols; c++) {
      line += grid[r]?.[c] === 1 ? '@' : '.';
    }
    lines.push(line);
  }
  return lines.join('\n');
}

export function parseMapFile(text) {
  const lines = text.split(/\r?\n/);
  let height = 0;
  let width = 0;
  let mapStartIdx = -1;
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();
    if (line.startsWith('height ')) height = parseInt(line.split(' ')[1], 10);
    if (line.startsWith('width ')) width = parseInt(line.split(' ')[1], 10);
    if (line === 'map') {
      mapStartIdx = i + 1;
      break;
    }
  }
  if (!height || !width || mapStartIdx < 0) return null;
  const grid = [];
  for (let r = 0; r < height; r++) {
    const row = [];
    const rawLine = lines[mapStartIdx + r] || '';
    for (let c = 0; c < width; c++) {
      const ch = rawLine[c] || '.';
      row.push(ch === '@' || ch === 'T' ? 1 : 0);
    }
    grid.push(row);
  }
  return { height, width, grid };
}

export function parseGridData(gridData) {
  if (!gridData) return null;
  try {
    const grid = typeof gridData === 'string' ? JSON.parse(gridData) : gridData;
    return Array.isArray(grid) ? grid : null;
  } catch {
    return null;
  }
}

/**
 * Render grid + optional POI markers to a sharp PNG blob.
 */
export function renderGridToPNG(rows, cols, grid, nodes = []) {
  const cellSize = computeExportCellSize(rows, cols);
  const logicalW = cols * cellSize;
  const logicalH = rows * cellSize;

  const canvas = document.createElement('canvas');
  canvas.width = logicalW;
  canvas.height = logicalH;

  const ctx = canvas.getContext('2d');
  if (!ctx) {
    return Promise.resolve(new Blob([], { type: 'image/png' }));
  }

  ctx.imageSmoothingEnabled = false;

  ctx.fillStyle = '#f0f0f0';
  ctx.fillRect(0, 0, logicalW, logicalH);

  ctx.fillStyle = '#595959';
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      if (grid[r]?.[c] === 1) {
        ctx.fillRect(c * cellSize, r * cellSize, cellSize, cellSize);
      }
    }
  }

  if (cellSize >= 6) {
    ctx.strokeStyle = '#e8e8e8';
    ctx.lineWidth = 0.5;
    for (let r = 0; r <= rows; r++) {
      ctx.beginPath();
      ctx.moveTo(0, r * cellSize);
      ctx.lineTo(logicalW, r * cellSize);
      ctx.stroke();
    }
    for (let c = 0; c <= cols; c++) {
      ctx.beginPath();
      ctx.moveTo(c * cellSize, 0);
      ctx.lineTo(c * cellSize, logicalH);
      ctx.stroke();
    }
  }

  drawPoisOnContext(ctx, nodes, cellSize);

  return new Promise((resolve) => {
    canvas.toBlob((blob) => resolve(blob || new Blob([], { type: 'image/png' })), 'image/png');
  });
}

/** Trigger browser download of PNG blob. */
export function downloadPngBlob(blob, filename) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.download = filename;
  link.href = url;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/** Attach grid_data + PNG preview (with POIs) to FormData for upload_map. */
export async function appendMapPreviewToFormData(formData, mapName, rows, cols, grid, nodes = []) {
  if (!grid || rows <= 0 || cols <= 0) return;
  formData.append('grid_data', JSON.stringify(grid));
  const pngBlob = await renderGridToPNG(rows, cols, grid, nodes);
  const safeName = mapName.trim().replace(/\s+/g, '_');
  formData.append('image_file', new File([pngBlob], `${safeName}.png`, { type: 'image/png' }));
}
