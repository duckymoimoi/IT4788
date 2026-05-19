import { useState } from 'react';
import {
  Typography, Card, Table, Tag, Space, Row, Col, Statistic, Tabs,
  Button, Modal, Descriptions, Badge, Popconfirm, message, Spin, Empty,
  Timeline,
} from 'antd';
import {
  AlertOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  TeamOutlined,
  ReloadOutlined,
  EyeOutlined,
  PhoneOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchSOSList, fetchSOSDetail, respondSOS, resolveSOS } from '../api/sos';

const { Title } = Typography;

// ─── Status config ────────────────────────────────────────────
const STATUS_CONFIG = {
  received: { color: 'red', icon: <ExclamationCircleOutlined />, label: 'Chờ xử lý' },
  assigned: { color: 'orange', icon: <ClockCircleOutlined />, label: 'Đang xử lý' },
  resolved: { color: 'green', icon: <CheckCircleOutlined />, label: 'Đã giải quyết' },
};

// ─── KPI Cards ────────────────────────────────────────────────
function KPICards({ sosList }) {
  const total = sosList?.length || 0;
  const received = sosList?.filter((s) => s.status === 'received').length || 0;
  const assigned = sosList?.filter((s) => s.status === 'assigned').length || 0;
  const resolved = sosList?.filter((s) => s.status === 'resolved').length || 0;

  return (
    <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="Tổng SOS"
            value={total}
            prefix={<AlertOutlined />}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="Chờ xử lý"
            value={received}
            prefix={<ExclamationCircleOutlined />}
            valueStyle={{ color: '#ff4d4f' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="Đang xử lý"
            value={assigned}
            prefix={<ClockCircleOutlined />}
            valueStyle={{ color: '#faad14' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="Đã giải quyết"
            value={resolved}
            prefix={<CheckCircleOutlined />}
            valueStyle={{ color: '#52c41a' }}
          />
        </Card>
      </Col>
    </Row>
  );
}

// ─── Detail Modal ─────────────────────────────────────────────
function SOSDetailModal({ sosId, open, onClose }) {
  const { data: detail, isLoading } = useQuery({
    queryKey: ['sos-detail', sosId],
    queryFn: () => fetchSOSDetail(sosId),
    enabled: !!sosId && open,
  });

  return (
    <Modal
      title={
        <Space>
          <AlertOutlined style={{ color: '#ff4d4f' }} />
          {`Chi tiết SOS #${sosId || ''}`}
        </Space>
      }
      open={open}
      onCancel={onClose}
      footer={null}
      width={600}
    >
      {isLoading ? (
        <Spin style={{ display: 'block', margin: '40px auto' }} />
      ) : !detail ? (
        <Empty description="Không tìm thấy SOS" />
      ) : (
        <>
          <Descriptions column={2} bordered size="small" style={{ marginBottom: 20 }}>
            <Descriptions.Item label="SOS ID">{detail.sos_id}</Descriptions.Item>
            <Descriptions.Item label="User ID">{detail.user_id}</Descriptions.Item>
            <Descriptions.Item label="Vị trí (Grid)">{detail.grid_location}</Descriptions.Item>
            <Descriptions.Item label="Trạng thái">
              <Tag
                color={STATUS_CONFIG[detail.status]?.color || 'default'}
                icon={STATUS_CONFIG[detail.status]?.icon}
              >
                {STATUS_CONFIG[detail.status]?.label || detail.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Tọa độ X">{detail.pos_x || '—'}</Descriptions.Item>
            <Descriptions.Item label="Tọa độ Y">{detail.pos_y || '—'}</Descriptions.Item>
            <Descriptions.Item label="Ghi chú" span={2}>
              {detail.note || 'Không có ghi chú'}
            </Descriptions.Item>
            {detail.assigned_staff_id && (
              <Descriptions.Item label="Staff phụ trách" span={2}>
                <TeamOutlined /> Staff #{detail.assigned_staff_id}
              </Descriptions.Item>
            )}
          </Descriptions>

          <Card size="small" title="Timeline" style={{ marginTop: 8 }}>
            <Timeline
              items={[
                {
                  color: 'red',
                  children: (
                    <>
                      <strong>Tạo SOS</strong>
                      <br />
                      <span style={{ color: '#888', fontSize: 12 }}>
                        {detail.created_at
                          ? new Date(detail.created_at).toLocaleString('vi-VN')
                          : '—'}
                      </span>
                    </>
                  ),
                },
                ...(detail.assigned_staff_id
                  ? [
                      {
                        color: 'orange',
                        children: (
                          <>
                            <strong>Staff #{detail.assigned_staff_id} nhận xử lý</strong>
                          </>
                        ),
                      },
                    ]
                  : []),
                ...(detail.resolved_at
                  ? [
                      {
                        color: 'green',
                        children: (
                          <>
                            <strong>Đã giải quyết</strong>
                            <br />
                            <span style={{ color: '#888', fontSize: 12 }}>
                              {new Date(detail.resolved_at).toLocaleString('vi-VN')}
                            </span>
                          </>
                        ),
                      },
                    ]
                  : []),
              ]}
            />
          </Card>
        </>
      )}
    </Modal>
  );
}

// ─── Main Page ────────────────────────────────────────────────
export default function SOSDashboard() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('all');
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [detailId, setDetailId] = useState(null);
  const [detailOpen, setDetailOpen] = useState(false);

  // ── Data fetching ──
  const { data, isLoading } = useQuery({
    queryKey: ['sos-list', page, pageSize],
    queryFn: () => fetchSOSList(page, pageSize),
    refetchInterval: 10000, // auto-refresh 10s
  });

  const sosList = data?.sos_list || [];
  const total = data?.total || 0;

  // ── Filter by tab ──
  const filteredList =
    activeTab === 'all' ? sosList : sosList.filter((s) => s.status === activeTab);

  // ── Mutations ──
  const respondMutation = useMutation({
    mutationFn: respondSOS,
    onSuccess: () => {
      message.success('Đã nhận xử lý SOS');
      queryClient.invalidateQueries({ queryKey: ['sos-list'] });
    },
    onError: (err) => {
      message.error(err.response?.data?.message || 'Không thể nhận SOS');
    },
  });

  const resolveMutation = useMutation({
    mutationFn: resolveSOS,
    onSuccess: () => {
      message.success('Đã giải quyết SOS');
      queryClient.invalidateQueries({ queryKey: ['sos-list'] });
    },
    onError: (err) => {
      message.error(err.response?.data?.message || 'Không thể đóng SOS');
    },
  });

  // ── Table columns ──
  const columns = [
    {
      title: 'ID',
      dataIndex: 'sos_id',
      key: 'sos_id',
      width: 70,
      sorter: (a, b) => a.sos_id - b.sos_id,
    },
    {
      title: 'User',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 80,
      render: (id) => <Badge status="processing" text={`#${id}`} />,
    },
    {
      title: 'Vị trí',
      dataIndex: 'grid_location',
      key: 'grid_location',
      width: 90,
      render: (loc) => <Tag>{`Ô ${loc}`}</Tag>,
    },
    {
      title: 'Ghi chú',
      dataIndex: 'note',
      key: 'note',
      ellipsis: true,
      render: (note) => note || <span style={{ color: '#bbb' }}>—</span>,
    },
    {
      title: 'Trạng thái',
      dataIndex: 'status',
      key: 'status',
      width: 130,
      render: (status) => {
        const cfg = STATUS_CONFIG[status] || {};
        return (
          <Tag color={cfg.color} icon={cfg.icon}>
            {cfg.label || status}
          </Tag>
        );
      },
    },
    {
      title: 'Thời gian',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (t) => (t ? new Date(t).toLocaleString('vi-VN') : '—'),
      sorter: (a, b) => new Date(a.created_at) - new Date(b.created_at),
      defaultSortOrder: 'descend',
    },
    {
      title: 'Hành động',
      key: 'action',
      width: 250,
      render: (_, record) => (
        <Space size="small">
          <Button
            size="small"
            icon={<EyeOutlined />}
            onClick={() => {
              setDetailId(record.sos_id);
              setDetailOpen(true);
            }}
          >
            Chi tiết
          </Button>

          {record.status === 'received' && (
            <Popconfirm
              title="Bạn muốn nhận xử lý SOS này?"
              onConfirm={() => respondMutation.mutate({ sos_id: record.sos_id })}
              okText="Nhận"
              cancelText="Hủy"
            >
              <Button
                size="small"
                type="primary"
                icon={<PhoneOutlined />}
                loading={respondMutation.isPending}
                style={{ background: '#faad14', borderColor: '#faad14' }}
              >
                Nhận xử lý
              </Button>
            </Popconfirm>
          )}

          {record.status === 'assigned' && (
            <Popconfirm
              title="Xác nhận đã giải quyết SOS này?"
              onConfirm={() => resolveMutation.mutate({ sos_id: record.sos_id })}
              okText="Đã xong"
              cancelText="Hủy"
            >
              <Button
                size="small"
                type="primary"
                icon={<CheckCircleOutlined />}
                loading={resolveMutation.isPending}
              >
                Giải quyết
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  // ── Tab items ──
  const tabItems = [
    { key: 'all', label: `Tất cả (${sosList.length})` },
    {
      key: 'received',
      label: (
        <span>
          <Badge color="red" /> Chờ xử lý ({sosList.filter((s) => s.status === 'received').length})
        </span>
      ),
    },
    {
      key: 'assigned',
      label: (
        <span>
          <Badge color="orange" /> Đang xử lý ({sosList.filter((s) => s.status === 'assigned').length})
        </span>
      ),
    },
    {
      key: 'resolved',
      label: (
        <span>
          <Badge color="green" /> Đã giải quyết ({sosList.filter((s) => s.status === 'resolved').length})
        </span>
      ),
    },
  ];

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>
          <AlertOutlined style={{ color: '#ff4d4f', marginRight: 8 }} />
          SOS Dashboard
        </Title>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => queryClient.invalidateQueries({ queryKey: ['sos-list'] })}
        >
          Làm mới
        </Button>
      </div>

      <KPICards sosList={sosList} />

      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />

        <Table
          dataSource={filteredList}
          columns={columns}
          rowKey="sos_id"
          loading={isLoading}
          pagination={{
            current: page,
            pageSize,
            total: activeTab === 'all' ? total : filteredList.length,
            onChange: (p) => setPage(p),
            showTotal: (t) => `Tổng ${t} SOS`,
            showSizeChanger: false,
          }}
          size="middle"
          locale={{ emptyText: <Empty description="Không có yêu cầu SOS nào" /> }}
          rowClassName={(record) =>
            record.status === 'received' ? 'sos-row-urgent' : ''
          }
        />
      </Card>

      <SOSDetailModal
        sosId={detailId}
        open={detailOpen}
        onClose={() => {
          setDetailOpen(false);
          setDetailId(null);
        }}
      />

      <style>{`
        .sos-row-urgent {
          background: #fff2f0 !important;
        }
        .sos-row-urgent:hover > td {
          background: #ffebe8 !important;
        }
      `}</style>
    </>
  );
}
