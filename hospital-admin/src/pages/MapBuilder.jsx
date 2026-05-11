import { useState, useMemo, useCallback } from 'react';
import {
  Typography, Row, Col, Card, Button, Space, Tag, Input, InputNumber,
  Modal, Form, Select, Switch, message, Spin, Empty, Divider, Radio, Tooltip,
} from 'antd';
import {
  BorderOutlined, EditOutlined, DeleteOutlined, PlusOutlined,
  SaveOutlined, ArrowLeftOutlined, EnvironmentOutlined,
  DragOutlined, ClearOutlined, AimOutlined,
} from '@ant-design/icons';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { fetchMaps, fetchNodes, uploadMap } from '../api/map';
import api from '../api/client';
import GridCanvas from '../components/GridCanvas/GridCanvas';

const { Title, Text } = Typography;

const POI_TYPES = [
  { value: 'entrance', label: '🚪 Cổng' },
  { value: 'room', label: '🏥 Phòng khám' },
  { value: 'elevator', label: '🛗 Thang máy' },
  { value: 'canteen', label: '🍽️ Canteen' },
  { value: 'pharmacy', label: '💊 Nhà thuốc' },
  { value: 'info', label: 'ℹ️ Thông tin' },
  { value: 'wc', label: '🚻 WC' },
  { value: 'stair', label: '🪜 Cầu thang' },
  { value: 'parking', label: '🅿️ Bãi đỗ' },
  { value: 'corridor', label: '🚶 Hành lang' },
  { value: 'wifi', label: '📶 WiFi' },
  { value: 'other', label: '⬜ Khác' },
];

// ─── Helper: generate .map file content ───────────────────────
function gridToMapFile(rows, cols, grid) {
  const lines = ['type octile', `height ${rows}`, `width ${cols}`, 'map'];
  for (let r = 0; r < rows; r++) {
    let line = '';
    for (let c = 0; c < cols; c++) {
      line += (grid[r]?.[c] === 1) ? '@' : '.';
    }
    lines.push(line);
  }
  return lines.join('\n');
}

// ─── Main Component ──────────────────────────────────────────
export default function MapBuilder() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const editMapId = searchParams.get('map_id');

  // Setup state
  const [mode, setMode] = useState(editMapId ? 'loading' : 'setup'); // 'setup' | 'loading' | 'editor'
  const [mapName, setMapName] = useState('');
  const [rows, setRows] = useState(20);
  const [cols, setCols] = useState(30);
  const [grid, setGrid] = useState(null); // 2D array: 0=walkable, 1=wall
  const [activeTool, setActiveTool] = useState('wall'); // 'wall' | 'eraser' | 'poi' | 'pan'
  const [localPois, setLocalPois] = useState([]); // POIs placed in editor
  const [poiModalOpen, setPoiModalOpen] = useState(false);
  const [pendingPoiCell, setPendingPoiCell] = useState(null);
  const [saving, setSaving] = useState(false);
  const [editingMapId, setEditingMapId] = useState(editMapId ? Number(editMapId) : null);
  const [poiForm] = Form.useForm();

  // Load existing maps list (for setup screen)
  const { data: maps, isLoading: loadingMaps } = useQuery({
    queryKey: ['admin-maps'],
    queryFn: fetchMaps,
    enabled: mode === 'setup',
  });

  // Auto-load if editing existing map
  const { } = useQuery({
    queryKey: ['load-map-for-edit', editMapId],
    queryFn: async () => {
      const allMaps = await fetchMaps();
      const map = allMaps?.find((m) => m.map_id === Number(editMapId));
      if (!map) { message.error('Map not found'); return null; }
      setMapName(map.map_name);
      setRows(map.rows);
      setCols(map.cols);
      const g = map.grid_data ? JSON.parse(map.grid_data) : Array.from({ length: map.rows }, () => Array(map.cols).fill(0));
      setGrid(g);
      // Load existing POIs
      const nodes = await fetchNodes(map.map_id);
      setLocalPois((nodes || []).map((n) => ({
        id: n.poi_id, code: n.poi_code, name: n.poi_name,
        type: n.poi_type, row: n.grid_row, col: n.grid_col,
        is_landmark: n.is_landmark, isExisting: true,
      })));
      setMode('editor');
      return map;
    },
    enabled: !!editMapId && mode === 'loading',
  });

  // ─── Create new empty grid ─────────────────────────────────
  const handleCreateNew = () => {
    if (!mapName.trim()) { message.warning('Nhập tên bản đồ'); return; }
    if (rows < 5 || cols < 5) { message.warning('Kích thước tối thiểu 5×5'); return; }
    const g = Array.from({ length: rows }, () => Array(cols).fill(0));
    setGrid(g);
    setLocalPois([]);
    setEditingMapId(null);
    setMode('editor');
  };

  // ─── Load existing map for editing ─────────────────────────
  const handleLoadMap = async (map) => {
    setMapName(map.map_name);
    setRows(map.rows);
    setCols(map.cols);
    const g = map.grid_data ? JSON.parse(map.grid_data) : Array.from({ length: map.rows }, () => Array(map.cols).fill(0));
    setGrid(g);
    setEditingMapId(map.map_id);
    try {
      const nodes = await fetchNodes(map.map_id);
      setLocalPois((nodes || []).map((n) => ({
        id: n.poi_id, code: n.poi_code, name: n.poi_name,
        type: n.poi_type, row: n.grid_row, col: n.grid_col,
        is_landmark: n.is_landmark, isExisting: true,
      })));
    } catch { setLocalPois([]); }
    setMode('editor');
  };

  // ─── Cell paint handler ────────────────────────────────────
  const handleCellPaint = useCallback((row, col) => {
    if (activeTool === 'wall') {
      setGrid((prev) => { const n = prev.map((r) => [...r]); n[row][col] = 1; return n; });
    } else if (activeTool === 'eraser') {
      setGrid((prev) => { const n = prev.map((r) => [...r]); n[row][col] = 0; return n; });
    } else if (activeTool === 'poi') {
      // Check if there's already a POI at this cell
      const existing = localPois.find((p) => p.row === row && p.col === col);
      if (existing) { message.info('Ô này đã có POI'); return; }
      // Check if cell is walkable
      if (grid[row]?.[col] === 1) { message.warning('Không thể đặt POI trên tường'); return; }
      setPendingPoiCell({ row, col });
      setPoiModalOpen(true);
    }
  }, [activeTool, localPois, grid]);

  // ─── Add POI ───────────────────────────────────────────────
  const handleAddPoi = async () => {
    try {
      const values = await poiForm.validateFields();
      // Check duplicate code
      if (localPois.some((p) => p.code === values.poi_code)) {
        message.error('Mã POI đã tồn tại'); return;
      }
      setLocalPois((prev) => [...prev, {
        id: Date.now(), code: values.poi_code, name: values.poi_name,
        type: values.poi_type, row: pendingPoiCell.row, col: pendingPoiCell.col,
        is_landmark: values.is_landmark || false, isExisting: false,
      }]);
      setPoiModalOpen(false);
      poiForm.resetFields();
      setPendingPoiCell(null);
      message.success('Đã thêm POI');
    } catch { /* validation error */ }
  };

  // ─── Remove POI ────────────────────────────────────────────
  const handleRemovePoi = (poiId) => {
    setLocalPois((prev) => prev.filter((p) => p.id !== poiId));
  };

  // ─── Grid data string for GridCanvas ───────────────────────
  const gridDataStr = useMemo(() => grid ? JSON.stringify(grid) : null, [grid]);

  // ─── Convert local POIs to node format for GridCanvas ──────
  const canvasNodes = useMemo(() => localPois.map((p) => ({
    poi_id: p.id, poi_code: p.code, poi_name: p.name,
    poi_type: p.type, grid_row: p.row, grid_col: p.col,
    grid_location: p.row * cols + p.col, is_landmark: p.is_landmark,
  })), [localPois, cols]);

  // ─── Save map ──────────────────────────────────────────────
  const handleSave = async () => {
    if (!mapName.trim()) { message.warning('Nhập tên bản đồ'); return; }
    setSaving(true);
    try {
      if (editingMapId) {
        // Update existing map grid_data
        await api.post('/admin/update_grid', {
          map_id: editingMapId,
          grid_data: JSON.stringify(grid),
          map_name: mapName.trim(),
        });
        message.success('Cập nhật map thành công!');
      } else {
        // Generate .map file and upload
        const content = gridToMapFile(rows, cols, grid);
        const blob = new Blob([content], { type: 'text/plain' });
        const file = new File([blob], `${mapName.trim().replace(/\s+/g, '_')}.map`);
        const fd = new FormData();
        fd.append('file', file);
        fd.append('map_name', mapName.trim());
        const result = await uploadMap(fd);
        const newMapId = result?.data?.map_id;
        if (newMapId) setEditingMapId(newMapId);
        message.success('Tạo map mới thành công!');
      }

      // Save new POIs (those not existing in DB)
      const newPois = localPois.filter((p) => !p.isExisting);
      for (const p of newPois) {
        try {
          await api.post('/admin/add_node', {
            map_id: editingMapId || 1,
            poi_code: p.code, poi_name: p.name, poi_type: p.type,
            grid_row: p.row, grid_col: p.col,
            is_landmark: p.is_landmark,
          });
        } catch (e) {
          console.warn('Failed to add POI:', p.code, e);
        }
      }

      queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
      queryClient.invalidateQueries({ queryKey: ['floors'] });
    } catch (err) {
      message.error('Lỗi: ' + (err.response?.data?.message || err.message));
    } finally {
      setSaving(false);
    }
  };

  // ─── Wall count stats ──────────────────────────────────────
  const stats = useMemo(() => {
    if (!grid) return { walls: 0, walkable: 0 };
    let walls = 0;
    for (const row of grid) for (const cell of row) if (cell === 1) walls++;
    return { walls, walkable: rows * cols - walls };
  }, [grid, rows, cols]);

  // ═══════════════════════════════════════════════════════════
  // RENDER: Setup Screen
  // ═══════════════════════════════════════════════════════════
  if (mode === 'setup') {
    return (
      <div style={{ maxWidth: 900, margin: '0 auto' }}>
        <Title level={4}><EditOutlined /> Map Builder</Title>
        <Row gutter={24}>
          {/* Create New */}
          <Col xs={24} md={12}>
            <Card title="🆕 Tạo Map Mới" style={{ height: '100%' }}>
              <Form layout="vertical">
                <Form.Item label="Tên bản đồ" required>
                  <Input value={mapName} onChange={(e) => setMapName(e.target.value)} placeholder="VD: Hospital Floor 1" />
                </Form.Item>
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item label="Rows (hàng)">
                      <InputNumber min={5} max={200} value={rows} onChange={setRows} style={{ width: '100%' }} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item label="Cols (cột)">
                      <InputNumber min={5} max={500} value={cols} onChange={setCols} style={{ width: '100%' }} />
                    </Form.Item>
                  </Col>
                </Row>
                <Text type="secondary">Grid: {rows}×{cols} = {rows * cols} ô</Text>
                <Button type="primary" block style={{ marginTop: 16 }} icon={<PlusOutlined />} onClick={handleCreateNew}>
                  Tạo Map
                </Button>
              </Form>
            </Card>
          </Col>
          {/* Edit Existing */}
          <Col xs={24} md={12}>
            <Card title="📂 Sửa Map Có Sẵn" style={{ height: '100%' }}>
              {loadingMaps ? <Spin /> : !maps?.length ? (
                <Empty description="Chưa có map nào" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              ) : (
                <div style={{ maxHeight: 300, overflow: 'auto' }}>
                  {maps.map((m) => (
                    <div
                      key={m.map_id}
                      style={{
                        padding: '8px 12px', marginBottom: 6, border: '1px solid #d9d9d9',
                        borderRadius: 6, cursor: 'pointer', display: 'flex',
                        justifyContent: 'space-between', alignItems: 'center',
                      }}
                      onClick={() => handleLoadMap(m)}
                    >
                      <div>
                        <Text strong>{m.map_name}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 11 }}>
                          {m.rows}×{m.cols} · ID:{m.map_id}
                        </Text>
                      </div>
                      <Space>
                        {m.is_active && <Tag color="green">Active</Tag>}
                        <Button size="small" icon={<EditOutlined />}>Sửa</Button>
                      </Space>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          </Col>
        </Row>
      </div>
    );
  }

  if (mode === 'loading') {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  }

  // ═══════════════════════════════════════════════════════════
  // RENDER: Editor
  // ═══════════════════════════════════════════════════════════
  return (
    <>
      {/* Toolbar */}
      <Card size="small" style={{ marginBottom: 12 }}>
        <Row gutter={12} align="middle">
          <Col>
            <Button icon={<ArrowLeftOutlined />} onClick={() => { setMode('setup'); setGrid(null); }}>
              Quay lại
            </Button>
          </Col>
          <Col flex="auto">
            <Space>
              <Input
                value={mapName}
                onChange={(e) => setMapName(e.target.value)}
                style={{ width: 240, fontWeight: 600 }}
                prefix={<EnvironmentOutlined />}
              />
              <Tag color="blue">{rows}×{cols}</Tag>
              <Tag>🧱 {stats.walls} tường</Tag>
              <Tag>🟢 {stats.walkable} đi được</Tag>
              <Tag color="purple">📍 {localPois.length} POIs</Tag>
            </Space>
          </Col>
          <Col>
            <Button type="primary" icon={<SaveOutlined />} loading={saving} onClick={handleSave}>
              {editingMapId ? 'Lưu thay đổi' : 'Tạo & Lưu'}
            </Button>
          </Col>
        </Row>

        {/* Tool Palette */}
        <Divider style={{ margin: '8px 0' }} />
        <Space size="middle">
          <Text strong style={{ fontSize: 12 }}>Công cụ:</Text>
          <Radio.Group value={activeTool} onChange={(e) => setActiveTool(e.target.value)} buttonStyle="solid">
            <Radio.Button value="wall"><BorderOutlined /> Vẽ Tường</Radio.Button>
            <Radio.Button value="eraser"><ClearOutlined /> Xóa Tường</Radio.Button>
            <Radio.Button value="poi"><AimOutlined /> Đặt POI</Radio.Button>
            <Radio.Button value="pan"><DragOutlined /> Di Chuyển</Radio.Button>
          </Radio.Group>
          <Text type="secondary" style={{ fontSize: 11 }}>
            {activeTool === 'wall' && '🖌️ Click/kéo để vẽ tường (ô xám)'}
            {activeTool === 'eraser' && '🧹 Click/kéo để xóa tường (thành đường đi)'}
            {activeTool === 'poi' && '📍 Click vào ô trống để đặt điểm POI'}
            {activeTool === 'pan' && '✋ Kéo để di chuyển bản đồ, cuộn để zoom'}
          </Text>
        </Space>
      </Card>

      {/* Editor Area */}
      <Row gutter={12}>
        {/* Canvas */}
        <Col xs={24} lg={18}>
          <Card bodyStyle={{ padding: 4 }}>
            <GridCanvas
              rows={rows}
              cols={cols}
              gridData={gridDataStr}
              nodes={canvasNodes}
              editMode={activeTool !== 'pan'}
              onCellPaint={handleCellPaint}
              onCellClick={(loc, r, c) => {
                if (activeTool === 'poi') handleCellPaint(r, c);
              }}
              width={Math.min(1000, window.innerWidth - 420)}
              height={Math.min(650, window.innerHeight - 260)}
            />
          </Card>
        </Col>

        {/* POI Sidebar */}
        <Col xs={24} lg={6}>
          <Card
            title={<Space><AimOutlined /> POIs ({localPois.length})</Space>}
            size="small"
            style={{ maxHeight: 650, overflow: 'auto' }}
          >
            {localPois.length === 0 ? (
              <Empty description="Chọn tool POI và click trên map" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              localPois.map((p) => (
                <div
                  key={p.id}
                  style={{
                    padding: '6px 8px', marginBottom: 4, borderRadius: 4,
                    border: '1px solid #f0f0f0', fontSize: 12,
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  }}
                >
                  <div>
                    <div style={{ fontWeight: 600 }}>{p.name}</div>
                    <Text type="secondary" style={{ fontSize: 10 }}>
                      {p.code} · {p.type} · ({p.row},{p.col})
                    </Text>
                    {p.is_landmark && <Tag color="gold" style={{ fontSize: 9, marginLeft: 4 }}>★</Tag>}
                    {p.isExisting && <Tag style={{ fontSize: 9, marginLeft: 4 }}>DB</Tag>}
                  </div>
                  {!p.isExisting && (
                    <Button
                      size="small" type="text" danger
                      icon={<DeleteOutlined />}
                      onClick={() => handleRemovePoi(p.id)}
                    />
                  )}
                </div>
              ))
            )}
          </Card>
        </Col>
      </Row>

      {/* Add POI Modal */}
      <Modal
        title={`Thêm POI tại (${pendingPoiCell?.row}, ${pendingPoiCell?.col})`}
        open={poiModalOpen}
        onCancel={() => { setPoiModalOpen(false); setPendingPoiCell(null); poiForm.resetFields(); }}
        onOk={handleAddPoi}
        okText="Thêm"
        cancelText="Hủy"
        destroyOnClose
        width={400}
      >
        <Form form={poiForm} layout="vertical" style={{ marginTop: 12 }}>
          <Form.Item label="Mã POI" name="poi_code" rules={[{ required: true, message: 'Nhập mã POI (VD: ENT-01)' }]}>
            <Input placeholder="VD: ENT-01, RM-101" />
          </Form.Item>
          <Form.Item label="Tên POI" name="poi_name" rules={[{ required: true, message: 'Nhập tên' }]}>
            <Input placeholder="VD: Cổng chính, Phòng khám Nội" />
          </Form.Item>
          <Form.Item label="Loại" name="poi_type" rules={[{ required: true }]} initialValue="room">
            <Select options={POI_TYPES} />
          </Form.Item>
          <Form.Item label="Landmark" name="is_landmark" valuePropName="checked" initialValue={false}>
            <Switch checkedChildren="Có" unCheckedChildren="Không" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
