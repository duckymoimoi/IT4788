import { useState } from 'react';
import {
  Typography, Card, Table, Button, Modal, Upload, message, Space, Tag,
  Popconfirm, Form, Input, InputNumber, Descriptions, Empty,
} from 'antd';
import {
  UploadOutlined, PlayCircleOutlined, DownloadOutlined,
  FileOutlined, CheckCircleOutlined, CloudUploadOutlined, EditOutlined,
  DeleteOutlined, StopOutlined, PictureOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  fetchMaps, fetchNodes, uploadMap, setActiveMap,
  exportMap, uploadOutput, deleteMap, deactivateMap,
} from '../api/map';
import {
  parseMapFile, appendMapPreviewToFormData, parseGridData,
  renderGridToPNG, downloadPngBlob,
} from '../utils/mapExport';

const { Title, Text } = Typography;

export default function MapManager() {
  const queryClient = useQueryClient();
  const [uploadMapModalOpen, setUploadMapModalOpen] = useState(false);
  const [uploadOutputModalOpen, setUploadOutputModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editingMap, setEditingMap] = useState(null);
  const [mapForm] = Form.useForm();
  const [editForm] = Form.useForm();

  // ─── Queries ────────────────────────────────────────────────
  const { data: maps, isLoading } = useQuery({
    queryKey: ['admin-maps'],
    queryFn: fetchMaps,
  });

  // ─── Mutations ──────────────────────────────────────────────
  const setActiveMutation = useMutation({
    mutationFn: (map_id) => setActiveMap(map_id),
    onSuccess: () => {
      message.success('Đã đổi Map Active thành công!');
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
      queryClient.invalidateQueries({ queryKey: ['floors'] });
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      queryClient.invalidateQueries({ queryKey: ['meta'] });
    },
    onError: (err) => message.error('Lỗi: ' + (err.response?.data?.message || err.message)),
  });

  const uploadMapMutation = useMutation({
    mutationFn: uploadMap,
    onSuccess: () => {
      message.success('Upload map thành công!');
      setUploadMapModalOpen(false);
      mapForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
    },
    onError: (err) => message.error('Upload thất bại: ' + (err.response?.data?.message || err.message)),
  });

  const uploadOutputMutation = useMutation({
    mutationFn: uploadOutput,
    onSuccess: () => {
      message.success('Upload output.json thành công!');
      setUploadOutputModalOpen(false);
    },
    onError: (err) => message.error('Upload thất bại: ' + (err.response?.data?.message || err.message)),
  });

  const deleteMapMutation = useMutation({
    mutationFn: (map_id) => deleteMap(map_id),
    onSuccess: () => {
      message.success('Đã xóa map!');
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
    },
    onError: (err) => message.error('Lỗi xóa: ' + (err.response?.data?.message || err.message)),
  });

  const deactivateMutation = useMutation({
    mutationFn: (map_id) => deactivateMap(map_id),
    onSuccess: () => {
      message.success('Đã tắt trạng thái Active!');
      queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
      queryClient.invalidateQueries({ queryKey: ['floors'] });
    },
    onError: (err) => message.error('Lỗi: ' + (err.response?.data?.message || err.message)),
  });

  // ─── Handlers ──────────────────────────────────────────────
  const handleUploadMap = async () => {
    try {
      const values = await mapForm.validateFields();
      const file = values.file?.[0]?.originFileObj;
      if (!file) return;

      const formData = new FormData();
      const mapName = values.map_name.trim();
      formData.append('map_name', mapName);
      formData.append('file', file);

      const text = await file.text();
      const parsed = parseMapFile(text);
      const rows = parsed?.height || values.rows;
      const cols = parsed?.width || values.cols;
      formData.append('rows', String(rows));
      formData.append('cols', String(cols));

      if (parsed?.grid) {
        await appendMapPreviewToFormData(formData, mapName, rows, cols, parsed.grid);
      }

      uploadMapMutation.mutate(formData);
    } catch (e) {
      message.error('Upload thất bại: ' + (e.message || 'Lỗi không xác định'));
    }
  };

  const handleExportPng = async (record) => {
    try {
      const grid = parseGridData(record.grid_data);
      if (!grid) {
        message.warning('Map chưa có grid_data — không thể xuất PNG');
        return;
      }
      const nodes = await fetchNodes(record.map_id);
      const blob = await renderGridToPNG(record.rows, record.cols, grid, nodes || []);
      const safeName = (record.map_name || `map_${record.map_id}`).replace(/\s+/g, '_');
      downloadPngBlob(blob, `${safeName}.png`);
      message.success(`Đã tải PNG (${nodes?.length || 0} POI)`);
    } catch (err) {
      message.error('Xuất PNG thất bại: ' + (err.response?.data?.message || err.message));
    }
  };

  const handleExport = async (map_file_path, map_name) => {
    try {
      // Extract filename from path like "data/warehouse_small.map" → "warehouse_small.map"
      const filename = map_file_path?.split('/').pop() || `${map_name}.map`;
      const response = await exportMap(filename);
      const blob = new Blob([response.data], { type: 'application/octet-stream' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
      message.success('Đã tải file .map');
    } catch (err) {
      message.error('Export thất bại: ' + (err.response?.data?.message || err.message));
    }
  };

  const handleUploadOutputFile = (info) => {
    const file = info.file;
    if (!file) return;
    const formData = new FormData();
    // Use active map id
    const activeMap = maps?.find((m) => m.is_active);
    if (activeMap) {
      formData.append('map_id', String(activeMap.map_id));
    }
    formData.append('file', file);
    uploadOutputMutation.mutate(formData);
  };

  const handleEditMap = async () => {
    try {
      const values = await editForm.validateFields();
      // Use editNode API pattern but for map - call a simple POST
      const res = await fetch(
        `${import.meta.env.VITE_API_BASE_URL || ''}/admin/edit_map`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${localStorage.getItem('token')}`,
          },
          body: JSON.stringify({ map_id: editingMap.map_id, map_name: values.map_name }),
        }
      );
      const data = await res.json();
      if (data.code === 1000) {
        message.success('Cập nhật tên map thành công!');
        setEditModalOpen(false);
        setEditingMap(null);
        queryClient.invalidateQueries({ queryKey: ['admin-maps'] });
        queryClient.invalidateQueries({ queryKey: ['floors'] });
      } else {
        message.error(data.message || 'Lỗi cập nhật');
      }
    } catch (e) {
      // validation or network error
      if (e.message) message.error(e.message);
    }
  };

  // ─── Table columns ─────────────────────────────────────────
  const columns = [
    {
      title: 'ID',
      dataIndex: 'map_id',
      key: 'map_id',
      width: 60,
    },
    {
      title: 'Tên bản đồ',
      dataIndex: 'map_name',
      key: 'map_name',
      render: (text, record) => (
        <Space>
          <FileOutlined />
          <Text strong>{text}</Text>
        </Space>
      ),
    },
    {
      title: 'Kích thước',
      key: 'size',
      width: 120,
      render: (_, record) => (
        <Tag color="blue">{record.rows}×{record.cols}</Tag>
      ),
    },
    {
      title: 'File',
      dataIndex: 'map_file_path',
      key: 'file',
      ellipsis: true,
      width: 200,
      render: (path) => (
        <Text type="secondary" style={{ fontSize: 12 }}>{path || '—'}</Text>
      ),
    },
    {
      title: 'Trạng thái',
      key: 'status',
      width: 100,
      render: (_, record) => (
        record.is_active
          ? <Tag icon={<CheckCircleOutlined />} color="success">Active</Tag>
          : <Tag color="default">Inactive</Tag>
      ),
    },
    {
      title: 'Ngày tạo',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (val) => val ? new Date(val).toLocaleString('vi-VN') : '—',
    },
    {
      title: 'Hành động',
      key: 'action',
      width: 400,
      render: (_, record) => (
        <Space size="small" wrap>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingMap(record);
              setEditModalOpen(true);
              setTimeout(() => editForm.setFieldsValue({ map_name: record.map_name }), 100);
            }}
          >
            Sửa
          </Button>
          {!record.is_active && (
            <Popconfirm
              title="Đặt map này làm Active?"
              description="Simulation đang chạy sẽ tự động dừng."
              onConfirm={() => setActiveMutation.mutate(record.map_id)}
            >
              <Button
                type="primary"
                size="small"
                icon={<PlayCircleOutlined />}
                loading={setActiveMutation.isPending}
              >
                Set Active
              </Button>
            </Popconfirm>
          )}
          {record.is_active && (
            <Popconfirm
              title="Tắt trạng thái Active?"
              description="Map sẽ không còn được sử dụng bởi hệ thống."
              onConfirm={() => deactivateMutation.mutate(record.map_id)}
            >
              <Button
                size="small"
                icon={<StopOutlined />}
                loading={deactivateMutation.isPending}
              >
                Deactivate
              </Button>
            </Popconfirm>
          )}
          <Button
            size="small"
            icon={<DownloadOutlined />}
            onClick={() => handleExport(record.map_file_path, record.map_name)}
          >
            .map
          </Button>
          <Button
            size="small"
            icon={<PictureOutlined />}
            onClick={() => handleExportPng(record)}
          >
            PNG
          </Button>
          {!record.is_active && (
            <Popconfirm
              title="Xóa vĩnh viễn map này?"
              description="Thao tác không thể hoàn tác. Toàn bộ POI và edge liên quan sẽ bị xóa."
              onConfirm={() => deleteMapMutation.mutate(record.map_id)}
              okText="Xóa"
              okType="danger"
            >
              <Button
                danger
                size="small"
                icon={<DeleteOutlined />}
                loading={deleteMapMutation.isPending}
              >
                Xóa
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>
          <FileOutlined style={{ marginRight: 8 }} />
          Quản lý File Bản Đồ
        </Title>
        <Space>
          <Button
            icon={<CloudUploadOutlined />}
            onClick={() => setUploadOutputModalOpen(true)}
          >
            Upload Output (MAPF)
          </Button>
          <Button
            type="primary"
            icon={<UploadOutlined />}
            onClick={() => setUploadMapModalOpen(true)}
          >
            Tải Map Mới
          </Button>
        </Space>
      </div>

      <Card>
        <Table
          dataSource={maps || []}
          columns={columns}
          rowKey="map_id"
          loading={isLoading}
          pagination={{ pageSize: 10 }}
          locale={{
            emptyText: <Empty description="Chưa có bản đồ nào" />,
          }}
        />
      </Card>

      {/* ─── Upload Map Modal ─────────────────────────────────── */}
      <Modal
        title="Tải lên Bản Đồ Mới (.map)"
        open={uploadMapModalOpen}
        onCancel={() => {
          setUploadMapModalOpen(false);
          mapForm.resetFields();
        }}
        onOk={handleUploadMap}
        confirmLoading={uploadMapMutation.isPending}
        okText="Upload"
        cancelText="Hủy"
        destroyOnClose
        width={480}
      >
        <Form form={mapForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            label="Tên bản đồ"
            name="map_name"
            rules={[{ required: true, message: 'Nhập tên bản đồ' }]}
          >
            <Input placeholder="VD: Hospital Floor 2" />
          </Form.Item>
          <Space size={16}>
            <Form.Item
              label="Rows"
              name="rows"
              rules={[{ required: true, message: 'Nhập số hàng' }]}
            >
              <InputNumber min={1} max={500} placeholder="33" />
            </Form.Item>
            <Form.Item
              label="Cols"
              name="cols"
              rules={[{ required: true, message: 'Nhập số cột' }]}
            >
              <InputNumber min={1} max={500} placeholder="57" />
            </Form.Item>
          </Space>
          <Form.Item
            label="File .map (octile format)"
            name="file"
            valuePropName="fileList"
            getValueFromEvent={(e) => e?.fileList ? e.fileList : e}
            rules={[{ required: true, message: 'Chọn file .map' }]}
          >
            <Upload
              beforeUpload={() => false}
              maxCount={1}
              accept=".map"
            >
              <Button icon={<UploadOutlined />}>Chọn file .map</Button>
            </Upload>
          </Form.Item>
        </Form>
      </Modal>

      {/* ─── Upload Output Modal ──────────────────────────────── */}
      <Modal
        title="Upload Output MAPF (output.json)"
        open={uploadOutputModalOpen}
        onCancel={() => setUploadOutputModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <div style={{ padding: '16px 0' }}>
          <Text type="secondary">
            Upload file output.json chứa MAPF pre-computed paths cho map đang active.
          </Text>
          <div style={{ marginTop: 16 }}>
            <Upload.Dragger
              beforeUpload={(file) => {
                handleUploadOutputFile({ file });
                return false;
              }}
              maxCount={1}
              accept=".json"
              showUploadList={false}
            >
              <p className="ant-upload-drag-icon">
                <CloudUploadOutlined style={{ fontSize: 36, color: '#1677ff' }} />
              </p>
              <p>Kéo thả file output.json vào đây</p>
              <p className="ant-upload-hint">Chỉ chấp nhận file .json</p>
            </Upload.Dragger>
          </div>
          {uploadOutputMutation.isPending && (
            <Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
              Đang upload...
            </Text>
          )}
        </div>
      </Modal>

      {/* ─── Edit Map Modal ────────────────────────────────────── */}
      <Modal
        title={`Sửa bản đồ — ${editingMap?.map_name || ''}`}
        open={editModalOpen}
        onCancel={() => { setEditModalOpen(false); setEditingMap(null); }}
        onOk={handleEditMap}
        okText="Lưu"
        cancelText="Hủy"
        destroyOnClose
        width={400}
      >
        <Form form={editForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            label="Tên bản đồ"
            name="map_name"
            rules={[{ required: true, message: 'Nhập tên bản đồ' }]}
          >
            <Input placeholder="VD: Hospital Floor 2" />
          </Form.Item>
          {editingMap && (
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="Map ID">{editingMap.map_id}</Descriptions.Item>
              <Descriptions.Item label="Kích thước">{editingMap.rows}×{editingMap.cols}</Descriptions.Item>
              <Descriptions.Item label="File">{editingMap.map_file_path || '—'}</Descriptions.Item>
              <Descriptions.Item label="Active">
                {editingMap.is_active ? <Tag color="green">Yes</Tag> : <Tag>No</Tag>}
              </Descriptions.Item>
            </Descriptions>
          )}
        </Form>
      </Modal>
    </div>
  );
}
