import React, { useState, useMemo } from 'react';
import { Typography, Table, Tag, Button, Space, Input, Select, Modal, Card, Empty, Tooltip, Badge, message } from 'antd';
import { DeleteOutlined, EyeOutlined, SearchOutlined, ApiOutlined, ReloadOutlined, CopyOutlined } from '@ant-design/icons';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../api/client';

const { Title, Text } = Typography;
const { TextArea } = Input;

const METHOD_COLORS = {
  GET: 'blue',
  POST: 'green',
  PUT: 'orange',
  PATCH: 'purple',
  DELETE: 'red',
};

const STATUS_COLOR = (status) => {
  if (!status) return 'default';
  if (status >= 200 && status < 300) return 'success';
  if (status >= 400 && status < 500) return 'warning';
  if (status >= 500) return 'error';
  return 'default';
};

function formatJSON(data) {
  if (!data) return '';
  try {
    if (typeof data === 'string') {
      const parsed = JSON.parse(data);
      return JSON.stringify(parsed, null, 2);
    }
    return JSON.stringify(data, null, 2);
  } catch {
    return String(data);
  }
}

// Fetch server-side request logs
const fetchRequestLogs = async () => {
  const res = await api.get('/admin/get_request_logs?limit=300');
  return res.data.data?.logs ?? [];
};

export default function ApiLogger() {
  const queryClient = useQueryClient();

  const { data: logs = [], isLoading, refetch } = useQuery({
    queryKey: ['requestLogs'],
    queryFn: fetchRequestLogs,
    refetchInterval: 5000, // Auto-refresh mỗi 5 giây
  });

  const [selectedLog, setSelectedLog] = useState(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [filterMethod, setFilterMethod] = useState(null);
  const [filterStatus, setFilterStatus] = useState(null);

  const filteredLogs = useMemo(() => {
    return logs.filter((log) => {
      const url = log.path + (log.query ? '?' + log.query : '');
      if (searchText && !url.toLowerCase().includes(searchText.toLowerCase())) return false;
      if (filterMethod && log.method !== filterMethod) return false;
      if (filterStatus === 'success' && (log.status < 200 || log.status >= 300)) return false;
      if (filterStatus === 'error' && log.status >= 200 && log.status < 400) return false;
      return true;
    });
  }, [logs, searchText, filterMethod, filterStatus]);

  const handleView = (record) => {
    setSelectedLog(record);
    setModalOpen(true);
  };

  const handleClear = async () => {
    try {
      await api.post('/admin/clear_request_logs');
      message.success('Đã xóa toàn bộ log');
      queryClient.invalidateQueries({ queryKey: ['requestLogs'] });
    } catch {
      message.error('Không thể xóa log');
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    message.success('Đã copy!');
  };

  const successCount = logs.filter((l) => l.status >= 200 && l.status < 300).length;
  const errorCount = logs.filter((l) => l.status >= 400).length;

  const columns = [
    {
      title: '#',
      dataIndex: 'id',
      key: 'id',
      width: 60,
      render: (id) => <Text type="secondary" style={{ fontSize: 12 }}>{id}</Text>,
    },
    {
      title: 'Thời gian',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 100,
      render: (t) => <Text style={{ fontSize: 12, fontFamily: 'monospace' }}>{new Date(t).toLocaleTimeString('vi-VN')}</Text>,
    },
    {
      title: 'Method',
      dataIndex: 'method',
      key: 'method',
      width: 80,
      filters: [
        { text: 'GET', value: 'GET' },
        { text: 'POST', value: 'POST' },
        { text: 'DELETE', value: 'DELETE' },
        { text: 'PUT', value: 'PUT' },
        { text: 'PATCH', value: 'PATCH' },
      ],
      onFilter: (value, record) => record.method === value,
      render: (m) => <Tag color={METHOD_COLORS[m] || 'default'}>{m}</Tag>,
    },
    {
      title: 'Path',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
      render: (path, record) => (
        <Tooltip title={path + (record.query ? '?' + record.query : '')}>
          <Text code style={{ fontSize: 12 }}>{path}</Text>
          {record.query && <Text type="secondary" style={{ fontSize: 11 }}>?{record.query.substring(0, 30)}{record.query.length > 30 ? '...' : ''}</Text>}
        </Tooltip>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (s) => <Tag color={STATUS_COLOR(s)}>{s}</Tag>,
    },
    {
      title: 'Duration',
      dataIndex: 'duration_ms',
      key: 'duration_ms',
      width: 90,
      sorter: (a, b) => a.duration_ms - b.duration_ms,
      render: (d) => {
        const color = d > 1000 ? '#ff4d4f' : d > 300 ? '#faad14' : '#52c41a';
        return <Text style={{ fontSize: 12, color, fontWeight: d > 1000 ? 'bold' : 'normal' }}>{d}ms</Text>;
      },
    },
    {
      title: 'Client',
      dataIndex: 'client_ip',
      key: 'client_ip',
      width: 120,
      render: (ip) => <Text style={{ fontSize: 11 }}>{ip}</Text>,
    },
    {
      title: '',
      key: 'action',
      width: 50,
      render: (_, record) => (
        <Button type="link" icon={<EyeOutlined />} onClick={() => handleView(record)} />
      ),
    },
  ];

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Space>
          <Title level={4} style={{ margin: 0 }}>
            <ApiOutlined style={{ marginRight: 8 }} />
            API Logger
          </Title>
          <Badge count={logs.length} overflowCount={999} style={{ backgroundColor: '#1677ff' }} />
          <Text type="secondary" style={{ fontSize: 12 }}>Server-side logs (auto-refresh 5s)</Text>
        </Space>
        <Space>
          <Tag color="success">{successCount} OK</Tag>
          <Tag color="error">{errorCount} Errors</Tag>
          <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isLoading}>
            Refresh
          </Button>
          <Button icon={<DeleteOutlined />} danger onClick={handleClear} disabled={logs.length === 0}>
            Clear All
          </Button>
        </Space>
      </div>

      {/* Filters */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Input
            prefix={<SearchOutlined />}
            placeholder="Filter by path..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
            style={{ width: 300 }}
          />
          <Select
            placeholder="Method"
            value={filterMethod}
            onChange={setFilterMethod}
            allowClear
            style={{ width: 120 }}
            options={[
              { value: 'GET', label: 'GET' },
              { value: 'POST', label: 'POST' },
              { value: 'PUT', label: 'PUT' },
              { value: 'PATCH', label: 'PATCH' },
              { value: 'DELETE', label: 'DELETE' },
            ]}
          />
          <Select
            placeholder="Status"
            value={filterStatus}
            onChange={setFilterStatus}
            allowClear
            style={{ width: 120 }}
            options={[
              { value: 'success', label: '2xx Success' },
              { value: 'error', label: '4xx/5xx Error' },
            ]}
          />
          <Text type="secondary">{filteredLogs.length} / {logs.length} requests</Text>
        </Space>
      </Card>

      {/* Table */}
      {filteredLogs.length === 0 && !isLoading ? (
        <Card>
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={logs.length === 0
              ? "Chưa có request nào đến server. Hãy gọi API từ app hoặc Postman, log sẽ xuất hiện tại đây."
              : "Không tìm thấy request phù hợp với bộ lọc."
            }
          />
        </Card>
      ) : (
        <Table
          dataSource={filteredLogs}
          columns={columns}
          rowKey="id"
          size="small"
          loading={isLoading}
          pagination={{ pageSize: 20, showSizeChanger: true, pageSizeOptions: [10, 20, 50, 100] }}
          rowClassName={(record) => record.status >= 400 ? 'api-log-error-row' : ''}
        />
      )}

      {/* Detail Modal */}
      <Modal
        title={
          <Space>
            <Tag color={METHOD_COLORS[selectedLog?.method]}>{selectedLog?.method}</Tag>
            <Text code>{selectedLog?.path}</Text>
            {selectedLog?.query && <Text type="secondary">?{selectedLog.query}</Text>}
            {selectedLog?.status && <Tag color={STATUS_COLOR(selectedLog.status)}>{selectedLog.status}</Tag>}
            {selectedLog?.duration_ms != null && <Text type="secondary">{selectedLog.duration_ms}ms</Text>}
          </Space>
        }
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        width={800}
      >
        {selectedLog && (
          <div>
            <Space style={{ marginBottom: 16 }}>
              <Text type="secondary">{new Date(selectedLog.timestamp).toLocaleString('vi-VN')}</Text>
              <Tag>{selectedLog.client_ip}</Tag>
              {selectedLog.user_agent && (
                <Tooltip title={selectedLog.user_agent}>
                  <Tag color="default">UA</Tag>
                </Tooltip>
              )}
            </Space>

            {/* Request Headers */}
            {selectedLog.headers && Object.keys(selectedLog.headers).length > 0 && (
              <div style={{ marginTop: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                  <Text strong>Request Headers</Text>
                  <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(formatJSON(selectedLog.headers))}>Copy</Button>
                </div>
                <TextArea
                  value={formatJSON(selectedLog.headers)}
                  readOnly
                  autoSize={{ minRows: 2, maxRows: 6 }}
                  style={{ fontFamily: 'monospace', fontSize: 12 }}
                />
              </div>
            )}

            {/* Request Body */}
            {selectedLog.request_body && (
              <div style={{ marginTop: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                  <Text strong>Request Body</Text>
                  <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(formatJSON(selectedLog.request_body))}>Copy</Button>
                </div>
                <TextArea
                  value={formatJSON(selectedLog.request_body)}
                  readOnly
                  autoSize={{ minRows: 2, maxRows: 10 }}
                  style={{ fontFamily: 'monospace', fontSize: 12 }}
                />
              </div>
            )}

            {/* Response Body */}
            {selectedLog.response_body && (
              <div style={{ marginTop: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                  <Text strong>Response Body</Text>
                  <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(formatJSON(selectedLog.response_body))}>Copy</Button>
                </div>
                <TextArea
                  value={formatJSON(selectedLog.response_body)}
                  readOnly
                  autoSize={{ minRows: 3, maxRows: 15 }}
                  style={{ fontFamily: 'monospace', fontSize: 12, backgroundColor: selectedLog.status >= 400 ? '#fff2f0' : '#f6ffed' }}
                />
              </div>
            )}
          </div>
        )}
      </Modal>

      <style>{`
        .api-log-error-row { background: #fff2f0 !important; }
        .api-log-error-row:hover > td { background: #ffebe8 !important; }
      `}</style>
    </>
  );
}
