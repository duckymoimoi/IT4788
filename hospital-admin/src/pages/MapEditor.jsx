import { useState, useMemo, useCallback } from 'react';
import {
  Typography, Row, Col, Card, Select, Input, Button, Space, Tag, Descriptions,
  Modal, Form, Switch, InputNumber, message, Spin, Empty, Badge, Tooltip, Divider,
  Upload, List, Popconfirm,
} from 'antd';
import {
  EnvironmentOutlined, SearchOutlined, EditOutlined, AimOutlined,
  StarOutlined, StarFilled, ReloadOutlined, ExpandOutlined,
  InfoCircleOutlined, UploadOutlined, DownloadOutlined,
  CheckCircleOutlined, CloudUploadOutlined, FileOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  fetchFloors, fetchNodes, fetchMeta, fetchDepts,
  searchLocation, fetchLandmarks, editNode, setCapacity,
  fetchMaps, uploadMap, setActiveMap, exportMap,
} from '../api/map';
import GridCanvas from '../components/GridCanvas/GridCanvas';

const { Title, Text } = Typography;

// ─── POI type labels & colors ─────────────────────────────────
const POI_TYPE_OPTIONS = [
  { value: 'entrance', label: '🚪 Cổng', color: '#52c41a' },
  { value: 'room', label: '🏥 Phòng khám', color: '#1677ff' },
  { value: 'elevator', label: '🛗 Thang máy', color: '#faad14' },
  { value: 'canteen', label: '🍽️ Canteen', color: '#fa8c16' },
  { value: 'pharmacy', label: '💊 Nhà thuốc', color: '#13c2c2' },
  { value: 'info', label: 'ℹ️ Thông tin', color: '#722ed1' },
  { value: 'wc', label: '🚻 WC', color: '#8c8c8c' },
  { value: 'toilet', label: '🚻 Toilet', color: '#8c8c8c' },
  { value: 'stair', label: '🪜 Cầu thang', color: '#eb2f96' },
  { value: 'parking', label: '🅿️ Bãi đỗ', color: '#2f54eb' },
  { value: 'corridor', label: '🚶 Hành lang', color: '#91caff' },
  { value: 'wifi', label: '📶 WiFi', color: '#389e0d' },
];

function getTypeLabel(type) {
  return POI_TYPE_OPTIONS.find((o) => o.value === type)?.label || type;
}
function getTypeColor(type) {
  return POI_TYPE_OPTIONS.find((o) => o.value === type)?.color || '#bfbfbf';
}

// ─── POI Info Panel ──────────────────────────────────────────
function POIInfoPanel({ node, onEdit, onSetCapacity }) {
  if (!node) {
    return (
      <Card
        title={<Space><InfoCircleOutlined /> POI Info</Space>}
        style={{ height: '100%' }}
      >
        <Empty
          description="Click vào một POI trên bản đồ để xem thông tin"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      </Card>
    );
  }

  return (
    <Card
      title={
        <Space>
          <Badge color={getTypeColor(node.poi_type)} />
          <span>{node.poi_name}</span>
        </Space>
      }
      extra={
        <Space>
          <Tooltip title="Sửa metadata">
            <Button type="primary" icon={<EditOutlined />} size="small" onClick={() => onEdit(node)}>
              Edit
            </Button>
          </Tooltip>
          <Tooltip title="Đặt capacity">
            <Button icon={<AimOutlined />} size="small" onClick={() => onSetCapacity(node)}>
              Capacity
            </Button>
          </Tooltip>
        </Space>
      }
    >
      <Descriptions column={1} size="small" bordered>
        <Descriptions.Item label="POI ID">{node.poi_id}</Descriptions.Item>
        <Descriptions.Item label="Mã">{node.poi_code}</Descriptions.Item>
        <Descriptions.Item label="Loại">
          <Tag color={getTypeColor(node.poi_type)}>{getTypeLabel(node.poi_type)}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Vị trí Grid">
          ({node.grid_row}, {node.grid_col}) = <Tag>{node.grid_location}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Landmark">
          {node.is_landmark ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined style={{ color: '#d9d9d9' }} />}
        </Descriptions.Item>
        <Descriptions.Item label="Accessible">
          <Tag color={node.is_accessible ? 'green' : 'red'}>{node.is_accessible ? 'Có' : 'Không'}</Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Wheelchair">
          <Tag color={node.wheelchair_accessible ? 'blue' : 'default'}>
            {node.wheelchair_accessible ? '♿ Có' : 'Không'}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="Capacity">{node.capacity ?? '—'}</Descriptions.Item>
        <Descriptions.Item label="Weight">{node.custom_weight ?? 1}</Descriptions.Item>
        <Descriptions.Item label="Giờ mở cửa">{node.open_hours ?? '—'}</Descriptions.Item>
        <Descriptions.Item label="Chi tiết">{node.details ?? '—'}</Descriptions.Item>
      </Descriptions>
    </Card>
  );
}

// ─── Edit Node Modal ─────────────────────────────────────────
function EditNodeModal({ open, node, onClose, onSuccess }) {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: editNode,
    onSuccess: () => {
      message.success('Cập nhật POI thành công!');
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      queryClient.invalidateQueries({ queryKey: ['landmarks'] });
      onSuccess?.();
      onClose();
    },
    onError: (err) => {
      message.error(`Lỗi: ${err.response?.data?.message || err.message}`);
    },
  });

  const handleSubmit = async () => {
    const values = await form.validateFields();
    // API expects "id" (poi_code) as identifier, plus optional fields
    mutation.mutate({
      id: node.poi_code,
      name: values.poi_name,
      type: values.poi_type,
      is_landmark: values.is_landmark,
      wheelchair_accessible: values.wheelchair_accessible,
      open_hours: values.open_hours || '',
      details: values.details || '',
    });
  };

  return (
    <Modal
      title={`Sửa POI — ${node?.poi_name || ''}`}
      open={open}
      onCancel={onClose}
      onOk={handleSubmit}
      confirmLoading={mutation.isPending}
      okText="Lưu"
      cancelText="Hủy"
      destroyOnClose
      width={520}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          poi_name: node?.poi_name,
          poi_type: node?.poi_type,
          is_landmark: node?.is_landmark ?? false,
          wheelchair_accessible: node?.wheelchair_accessible ?? false,
          open_hours: node?.open_hours || '',
          details: node?.details || '',
        }}
        style={{ marginTop: 16 }}
      >
        <Form.Item label="Tên POI" name="poi_name" rules={[{ required: true, message: 'Nhập tên POI' }]}>
          <Input placeholder="VD: Phòng khám Nội khoa" />
        </Form.Item>
        <Form.Item label="Loại" name="poi_type" rules={[{ required: true }]}>
          <Select options={POI_TYPE_OPTIONS} />
        </Form.Item>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="Landmark" name="is_landmark" valuePropName="checked">
              <Switch checkedChildren="Có" unCheckedChildren="Không" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="Wheelchair" name="wheelchair_accessible" valuePropName="checked">
              <Switch checkedChildren="♿ Có" unCheckedChildren="Không" />
            </Form.Item>
          </Col>
        </Row>
        <Form.Item label="Giờ mở cửa" name="open_hours">
          <Input placeholder="VD: 07:00 - 17:00" />
        </Form.Item>
        <Form.Item label="Chi tiết" name="details">
          <Input.TextArea rows={3} placeholder="Mô tả thêm..." />
        </Form.Item>
      </Form>
    </Modal>
  );
}

// ─── Set Capacity Modal ──────────────────────────────────────
function CapacityModal({ open, node, onClose }) {
  const [value, setValue] = useState(null);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: setCapacity,
    onSuccess: () => {
      message.success('Cập nhật capacity thành công!');
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      onClose();
    },
    onError: (err) => {
      message.error(`Lỗi: ${err.response?.data?.message || err.message}`);
    },
  });

  return (
    <Modal
      title={`Set Capacity — ${node?.poi_name || ''}`}
      open={open}
      onCancel={onClose}
      onOk={() => mutation.mutate({ poi_id: node.poi_id, poi_code: node.poi_code, capacity: value })}
      confirmLoading={mutation.isPending}
      okText="Lưu"
      cancelText="Hủy"
      destroyOnClose
    >
      <div style={{ padding: '16px 0' }}>
        <Text>Capacity hiện tại: <Tag>{node?.capacity ?? 'Chưa đặt'}</Tag></Text>
        <div style={{ marginTop: 12 }}>
          <InputNumber
            min={0}
            max={9999}
            defaultValue={node?.capacity}
            onChange={setValue}
            style={{ width: '100%' }}
            placeholder="Nhập capacity mới (VD: 50)"
            size="large"
          />
        </div>
      </div>
    </Modal>
  );
}

// ─── Map Management Panel ─────────────────────────────────────
function MapManagementPanel({ activeFloor, onFloorChange }) {
  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [mapName, setMapName] = useState('');
  const queryClient = useQueryClient();

  const { data: maps, isLoading } = useQuery({
    queryKey: ['admin-maps'],
    queryFn: fetchMaps,
  });

  const activateMutation = useMutation({
    mutationFn: setActiveMap,
    onSuccess: () => {
      message.success('Đã đặt map active!');
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
      queryClient.invalidateQueries({ queryKey: ['floors'] });
    },
    onError: (err) => message.error(err.response?.data?.message || 'Lỗi'),
  });

  const handleUpload = async (file) => {
    if (!mapName.trim()) {
      message.warning('Nhập tên bản đồ trước khi upload');
      return false;
    }
    setUploading(true);
    try {
      // Read file to parse height/width from octile header
      const text = await file.text();
      const headerLines = text.split(/\r?\n/);
      let parsedRows = 0, parsedCols = 0;
      let mapStartIdx = -1;
      for (let i = 0; i < headerLines.length; i++) {
        const trimmed = headerLines[i].trim();
        if (trimmed.startsWith('height ')) parsedRows = parseInt(trimmed.split(' ')[1], 10);
        if (trimmed.startsWith('width ')) parsedCols = parseInt(trimmed.split(' ')[1], 10);
        if (trimmed === 'map') { mapStartIdx = i + 1; break; }
      }

      const fd = new FormData();
      fd.append('file', file);
      const finalName = mapName.trim() || file.name.replace(/\.map$/i, '');
      fd.append('map_name', finalName);
      if (parsedRows > 0) fd.append('rows', String(parsedRows));
      if (parsedCols > 0) fd.append('cols', String(parsedCols));

      // Parse grid to generate PNG and JSON grid_data
      if (parsedRows > 0 && parsedCols > 0 && mapStartIdx >= 0) {
        const grid = [];
        for (let r = 0; r < parsedRows; r++) {
          const row = [];
          const rawLine = headerLines[mapStartIdx + r] || '';
          for (let c = 0; c < parsedCols; c++) {
            const ch = rawLine[c] || '.';
            row.push((ch === '@' || ch === 'T') ? 1 : 0);
          }
          grid.push(row);
        }
        
        fd.append('grid_data', JSON.stringify(grid));
        
        const pngBlob = await renderGridToPNG(parsedRows, parsedCols, grid);
        const safeName = finalName.replace(/\s+/g, '_');
        const pngFile = new File([pngBlob], `${safeName}.png`, { type: 'image/png' });
        fd.append('image_file', pngFile);
      }

      await uploadMap(fd);
      message.success(`Upload thành công! ${parsedRows ? `(${parsedRows}×${parsedCols})` : ''}`);
      setUploadModalOpen(false);
      setMapName('');
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
      queryClient.invalidateQueries({ queryKey: ['floors'] });
    } catch (err) {
      message.error(err.response?.data?.message || 'Upload thất bại');
    } finally {
      setUploading(false);
    }
    return false; // prevent default upload
  };

  const handleExport = async (mapId) => {
    try {
      const res = await exportMap(mapId);
      const url = window.URL.createObjectURL(new Blob([res.data]));
      const a = document.createElement('a');
      a.href = url;
      a.download = `map_${mapId}.map`;
      a.click();
      window.URL.revokeObjectURL(url);
      message.success('Đã tải file map!');
    } catch (err) {
      message.error('Export thất bại');
    }
  };

  return (
    <>
      <Card
        title={<Space><FileOutlined /> Quản lý Map</Space>}
        size="small"
        style={{ marginTop: 16 }}
        extra={
          <Button
            type="primary"
            icon={<CloudUploadOutlined />}
            size="small"
            onClick={() => setUploadModalOpen(true)}
          >
            Upload
          </Button>
        }
      >
        {isLoading ? (
          <Spin size="small" />
        ) : !maps?.length ? (
          <Empty description="Chưa có map" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <List
            size="small"
            dataSource={maps}
            renderItem={(m) => (
              <List.Item
                style={{
                  padding: '6px 0',
                  background: m.map_id === activeFloor ? '#e6f4ff' : 'transparent',
                  borderRadius: 4,
                  paddingLeft: 8,
                }}
                actions={[
                  !m.is_active && (
                    <Popconfirm
                      key="activate"
                      title="Đặt map này làm active?"
                      onConfirm={() => activateMutation.mutate(m.map_id)}
                    >
                      <Button size="small" type="link">Active</Button>
                    </Popconfirm>
                  ),
                  <Tooltip key="export" title="Tải file .map">
                    <Button
                      size="small"
                      type="link"
                      icon={<DownloadOutlined />}
                      onClick={() => handleExport(m.map_id)}
                    />
                  </Tooltip>,
                  <Button
                    key="view"
                    size="small"
                    type="link"
                    onClick={() => onFloorChange(m.map_id)}
                  >
                    Xem
                  </Button>,
                ].filter(Boolean)}
              >
                <List.Item.Meta
                  title={
                    <Space size={4}>
                      <span style={{ fontSize: 13 }}>{m.map_name}</span>
                      {m.is_active && <Tag color="green" style={{ fontSize: 10 }}>Active</Tag>}
                    </Space>
                  }
                  description={
                    <Text style={{ fontSize: 11, color: '#999' }}>
                      {m.rows}×{m.cols} · ID:{m.map_id}
                    </Text>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </Card>

      {/* Upload Map Modal */}
      <Modal
        title="Upload file .map mới"
        open={uploadModalOpen}
        onCancel={() => { setUploadModalOpen(false); setMapName(''); }}
        footer={null}
        destroyOnClose
        width={420}
      >
        <div style={{ marginBottom: 12 }}>
          <Text strong>Tên bản đồ:</Text>
          <Input
            value={mapName}
            onChange={(e) => setMapName(e.target.value)}
            placeholder="VD: Hospital Floor 2"
            style={{ marginTop: 4 }}
          />
        </div>
        <Upload.Dragger
          accept=".map"
          maxCount={1}
          beforeUpload={handleUpload}
          showUploadList={false}
          disabled={uploading}
        >
          <p className="ant-upload-drag-icon">
            {uploading ? <Spin /> : <UploadOutlined style={{ fontSize: 32, color: '#1677ff' }} />}
          </p>
          <p className="ant-upload-text">
            {uploading ? 'Đang upload...' : 'Kéo file .map vào đây hoặc click để chọn'}
          </p>
          <p className="ant-upload-hint">Chỉ hỗ trợ file định dạng MovingAI (.map)</p>
        </Upload.Dragger>
      </Modal>
    </>
  );
}

// ─── Map Editor Page ──────────────────────────────────────────
export default function MapEditor() {
  // State
  const [selectedFloor, setSelectedFloor] = useState(null);
  const [selectedNode, setSelectedNode] = useState(null);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [capModalOpen, setCapModalOpen] = useState(false);
  const [editingNode, setEditingNode] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [highlightNodeId, setHighlightNodeId] = useState(null);
  const [showLandmarksOnly, setShowLandmarksOnly] = useState(false);

  const queryClient = useQueryClient();

  // ─── Queries ────────────────────────────────────────────────
  const { data: floors, isLoading: loadingFloors } = useQuery({
    queryKey: ['floors'],
    queryFn: fetchFloors,
  });

  // Fetch admin maps to get grid_data (wall/walkable info)
  const { data: adminMaps } = useQuery({
    queryKey: ['admin-maps'],
    queryFn: fetchMaps,
  });

  // Auto-select first floor
  const activeFloor = selectedFloor ?? floors?.[0]?.map_id ?? null;

  // Get grid_data for the active floor from admin maps
  const activeMapData = useMemo(() => {
    if (!adminMaps || !activeFloor) return null;
    return adminMaps.find((m) => m.map_id === activeFloor);
  }, [adminMaps, activeFloor]);

  const { data: meta } = useQuery({
    queryKey: ['meta', activeFloor],
    queryFn: () => fetchMeta(activeFloor),
    enabled: !!activeFloor,
  });

  const { data: allNodes, isLoading: loadingNodes } = useQuery({
    queryKey: ['nodes', activeFloor],
    queryFn: () => fetchNodes(activeFloor),
    enabled: !!activeFloor,
  });

  const { data: depts } = useQuery({
    queryKey: ['depts'],
    queryFn: fetchDepts,
  });

  const { data: landmarks } = useQuery({
    queryKey: ['landmarks', activeFloor],
    queryFn: () => fetchLandmarks(activeFloor),
    enabled: !!activeFloor,
  });

  // Filter nodes for display
  const displayNodes = useMemo(() => {
    if (!allNodes) return [];
    if (showLandmarksOnly) return allNodes.filter((n) => n.is_landmark);
    return allNodes;
  }, [allNodes, showLandmarksOnly]);

  // ─── Search ─────────────────────────────────────────────────
  const { data: searchResults, isFetching: searching } = useQuery({
    queryKey: ['search-location', searchText, activeFloor],
    queryFn: () => searchLocation(searchText, activeFloor),
    enabled: searchText.length >= 2 && !!activeFloor,
  });

  const handleSearch = useCallback((value) => {
    setSearchText(value);
    setHighlightNodeId(null);
  }, []);

  const handleSearchSelect = useCallback((poiId) => {
    setHighlightNodeId(poiId);
    const node = allNodes?.find((n) => n.poi_id === poiId);
    if (node) setSelectedNode(node);
  }, [allNodes]);

  // Local search fallback (search in loaded nodes)
  const localSearchResults = useMemo(() => {
    if (!searchText || searchText.length < 2 || !allNodes) return [];
    const q = searchText.toLowerCase();
    return allNodes.filter(
      (n) =>
        n.poi_name?.toLowerCase().includes(q) ||
        n.poi_code?.toLowerCase().includes(q) ||
        n.poi_type?.toLowerCase().includes(q)
    );
  }, [searchText, allNodes]);

  const effectiveSearchResults = searchResults?.length > 0 ? searchResults : localSearchResults;

  // ─── Handlers ──────────────────────────────────────────────
  const handleNodeClick = useCallback((node) => {
    setSelectedNode(node);
    setHighlightNodeId(null);
  }, []);

  const handleEdit = useCallback((node) => {
    setEditingNode(node);
    setEditModalOpen(true);
  }, []);

  const handleSetCapacity = useCallback((node) => {
    setEditingNode(node);
    setCapModalOpen(true);
  }, []);

  // ─── Loading ────────────────────────────────────────────────
  if (loadingFloors) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  }

  return (
    <>
      <Title level={4}>
        <EnvironmentOutlined style={{ marginRight: 8 }} />
        Map Editor
      </Title>

      {/* ─── Toolbar ─────────────────────────────────────────── */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={16} align="middle">
          {/* Floor selector */}
          <Col>
            <Space>
              <Text strong>Tầng:</Text>
              <Select
                value={activeFloor}
                onChange={(v) => {
                  setSelectedFloor(v);
                  setSelectedNode(null);
                  setHighlightNodeId(null);
                }}
                style={{ width: 240 }}
                placeholder="Chọn tầng..."
                options={(floors || []).map((f) => ({
                  value: f.map_id,
                  label: `${f.map_name} (${f.rows}×${f.cols})`,
                }))}
              />
            </Space>
          </Col>

          {/* Search */}
          <Col flex="auto">
            <Input.Search
              placeholder="Tìm POI theo tên, mã, loại..."
              allowClear
              onSearch={handleSearch}
              onChange={(e) => {
                if (!e.target.value) {
                  setSearchText('');
                  setHighlightNodeId(null);
                }
              }}
              loading={searching}
              style={{ maxWidth: 360 }}
              prefix={<SearchOutlined />}
            />
          </Col>

          {/* Landmarks toggle */}
          <Col>
            <Space>
              <Button
                type={showLandmarksOnly ? 'primary' : 'default'}
                icon={<StarFilled />}
                onClick={() => setShowLandmarksOnly(!showLandmarksOnly)}
              >
                Landmarks ({landmarks?.length || 0})
              </Button>
              <Tooltip title="Refresh dữ liệu">
                <Button
                  icon={<ReloadOutlined />}
                  onClick={() => {
                    queryClient.invalidateQueries({ queryKey: ['nodes', activeFloor] });
                    queryClient.invalidateQueries({ queryKey: ['landmarks', activeFloor] });
                    queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
                    message.info('Đang làm mới...');
                  }}
                />
              </Tooltip>
            </Space>
          </Col>
        </Row>

        {/* Search results dropdown */}
        {searchText.length >= 2 && effectiveSearchResults?.length > 0 && (
          <div style={{ marginTop: 8, padding: '8px 0' }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Kết quả tìm kiếm ({effectiveSearchResults.length}):
            </Text>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginTop: 4 }}>
              {effectiveSearchResults.slice(0, 10).map((n) => (
                <Tag
                  key={n.poi_id}
                  color={getTypeColor(n.poi_type)}
                  style={{ cursor: 'pointer' }}
                  onClick={() => handleSearchSelect(n.poi_id)}
                >
                  {n.poi_code} — {n.poi_name}
                </Tag>
              ))}
            </div>
          </div>
        )}
      </Card>

      {/* ─── Main Content ────────────────────────────────────── */}
      <Row gutter={16}>
        {/* Canvas */}
        <Col xs={24} lg={16} xl={17}>
          <Card
            bodyStyle={{ padding: 8 }}
            title={
              <Space>
                <Text strong>{meta?.map_name || 'Bản đồ'}</Text>
                {meta && (
                  <Tag color="blue">{meta.rows}×{meta.cols} grid</Tag>
                )}
                <Tag>{displayNodes?.length || 0} POIs</Tag>
                {activeMapData?.is_active && (
                  <Tag color="green">Active</Tag>
                )}
              </Space>
            }
          >
            {loadingNodes ? (
              <div style={{ display: 'flex', justifyContent: 'center', padding: 100 }}>
                <Spin size="large" tip="Đang tải bản đồ..." />
              </div>
            ) : (
              <GridCanvas
                rows={meta?.rows || 33}
                cols={meta?.cols || 57}
                gridData={activeMapData?.grid_data || null}
                nodes={displayNodes}
                selectedNodeId={selectedNode?.poi_id}
                highlightNodeId={highlightNodeId}
                onNodeClick={handleNodeClick}
                width={Math.min(1100, window.innerWidth - 500)}
                height={620}
              />
            )}
          </Card>
        </Col>

        {/* Sidebar */}
        <Col xs={24} lg={8} xl={7}>
          {/* POI Info Panel */}
          <POIInfoPanel
            node={selectedNode}
            onEdit={handleEdit}
            onSetCapacity={handleSetCapacity}
          />

          {/* Map Management Panel */}
          <MapManagementPanel
            activeFloor={activeFloor}
            onFloorChange={(id) => {
              setSelectedFloor(id);
              setSelectedNode(null);
              setHighlightNodeId(null);
            }}
          />

          {/* Quick Stats */}
          <Card title="📊 Thống kê" style={{ marginTop: 16 }} size="small">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Tổng POIs">{allNodes?.length || 0}</Descriptions.Item>
              <Descriptions.Item label="Landmarks">{landmarks?.length || 0}</Descriptions.Item>
              <Descriptions.Item label="Khoa">{depts?.length || 0}</Descriptions.Item>
              <Descriptions.Item label="Grid Size">
                {meta ? `${meta.rows} × ${meta.cols} (${meta.rows * meta.cols} cells)` : '—'}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
      </Row>

      {/* ─── Modals ──────────────────────────────────────────── */}
      {editingNode && (
        <EditNodeModal
          open={editModalOpen}
          node={editingNode}
          onClose={() => {
            setEditModalOpen(false);
            setEditingNode(null);
          }}
          onSuccess={() => {
            // Update selected node after edit
            const refreshed = queryClient.getQueryData(['nodes', activeFloor]);
            if (refreshed) {
              const updated = refreshed.find((n) => n.poi_id === editingNode.poi_id);
              if (updated) setSelectedNode(updated);
            }
          }}
        />
      )}
      {editingNode && (
        <CapacityModal
          open={capModalOpen}
          node={editingNode}
          onClose={() => {
            setCapModalOpen(false);
            setEditingNode(null);
          }}
        />
      )}
    </>
  );
}
