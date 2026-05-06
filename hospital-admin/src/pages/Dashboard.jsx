import { Row, Col, Card, Statistic, Typography, Table, Tag, Space, Spin, Result } from 'antd';
import {
  UserOutlined,
  AlertOutlined,
  HeartOutlined,
  DashboardOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import {
  BarChart, Bar, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Cell,
} from 'recharts';
import { fetchEngineHealth } from '../api/engine';
import { fetchAlerts, fetchBottlenecks, fetchHeatmap, fetchStatsFlow } from '../api/flow';

const { Title } = Typography;

// ─── KPI Cards (A7) ───────────────────────────────────────────
function KPICards({ engineHealth, heatmapData }) {
  const totalDensity = (heatmapData || []).reduce((sum, d) => sum + d.density, 0);

  return (
    <Row gutter={[16, 16]}>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="Engine Status"
            value={engineHealth?.status === 'ok' ? 'Online' : 'Offline'}
            prefix={<DashboardOutlined />}
            valueStyle={{ color: engineHealth?.status === 'ok' ? '#52c41a' : '#ff4d4f' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="DB Connected"
            value={engineHealth?.db_connected ? 'Yes' : 'No'}
            prefix={<CheckCircleOutlined />}
            valueStyle={{ color: engineHealth?.db_connected ? '#52c41a' : '#ff4d4f' }}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="Total Flow"
            value={totalDensity}
            prefix={<UserOutlined />}
          />
        </Card>
      </Col>
      <Col xs={12} sm={6}>
        <Card>
          <Statistic
            title="MAPF Agents"
            value={engineHealth?.agent_count || 0}
            prefix={<HeartOutlined />}
          />
        </Card>
      </Col>
    </Row>
  );
}

// ─── Density Chart (A8) ───────────────────────────────────────
function DensityChart({ data }) {
  const chartData = (data || []).map((d) => ({
    hour: `${d.hour}:00`,
    count: d.count,
  }));

  return (
    <Card title="📊 Mật độ người 24h" style={{ marginTop: 16 }}>
      {chartData.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          Chưa có dữ liệu — Cần bật Simulation trước
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="hour" />
            <YAxis />
            <Tooltip />
            <Bar dataKey="count" fill="#1677ff" radius={[4, 4, 0, 0]}>
              {chartData.map((_, index) => (
                <Cell key={index} fill={index % 2 === 0 ? '#1677ff' : '#69b1ff'} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      )}
    </Card>
  );
}

// ─── Alerts Panel (A9) ────────────────────────────────────────
function AlertsPanel({ alerts, bottlenecks }) {
  const alertColumns = [
    { title: 'ID', dataIndex: 'priority_id', key: 'id', width: 60 },
    {
      title: 'Route',
      key: 'route',
      render: (_, r) => `${r.from_location} → ${r.to_location}`,
    },
    { title: 'Reason', dataIndex: 'reason', key: 'reason', ellipsis: true },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (s) => (
        <Tag color={s === 'active' ? 'red' : 'green'} icon={s === 'active' ? <WarningOutlined /> : <CheckCircleOutlined />}>
          {s}
        </Tag>
      ),
    },
    {
      title: 'Time',
      dataIndex: 'activated_at',
      key: 'time',
      render: (t) => new Date(t).toLocaleString('vi-VN'),
    },
  ];

  return (
    <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
      <Col span={14}>
        <Card
          title={
            <Space>
              <AlertOutlined style={{ color: '#ff4d4f' }} />
              Priority Routes
            </Space>
          }
        >
          <Table
            dataSource={alerts || []}
            columns={alertColumns}
            rowKey="priority_id"
            pagination={false}
            size="small"
            locale={{ emptyText: 'Không có tuyến ưu tiên nào' }}
          />
        </Card>
      </Col>
      <Col span={10}>
        <Card title="🔥 Bottlenecks (Top 5)">
          {(bottlenecks || []).length === 0 ? (
            <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
              Không phát hiện tắc nghẽn
            </div>
          ) : (
            <Table
              dataSource={bottlenecks}
              columns={[
                { title: 'Location', dataIndex: 'grid_location', key: 'loc' },
                { title: 'Density', dataIndex: 'count', key: 'density' },
              ]}
              rowKey="grid_location"
              pagination={false}
              size="small"
            />
          )}
        </Card>
      </Col>
    </Row>
  );
}

// ─── Mini Heatmap (A10) ───────────────────────────────────────
function MiniHeatmap({ data }) {
  if (!data || data.length === 0) {
    return (
      <Card title="🗺️ Mini Heatmap" style={{ marginTop: 16 }}>
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          Chưa có dữ liệu heatmap — Cần bật Simulation
        </div>
      </Card>
    );
  }

  const maxDensity = Math.max(...data.map((d) => d.density), 1);
  const chartData = data.slice(0, 20).map((d) => ({
    location: `Ô ${d.grid_location}`,
    density: d.density,
  }));

  return (
    <Card title="🗺️ Mini Heatmap (Top 20 cells)" style={{ marginTop: 16 }}>
      <ResponsiveContainer width="100%" height={chartData.length * 25 + 50}>
        <BarChart data={chartData} layout="vertical" margin={{ top: 5, right: 20, left: 20, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis type="number" />
          <YAxis dataKey="location" type="category" width={80} tick={{ fontSize: 11 }} interval={0} />
          <Tooltip />
          <Bar dataKey="density" radius={[0, 4, 4, 0]}>
            {chartData.map((entry, index) => (
              <Cell
                key={index}
                fill={`rgba(255, ${Math.round(255 - (entry.density / maxDensity) * 200)}, 0, 0.85)`}
              />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </Card>
  );
}

// ─── Dashboard Page ───────────────────────────────────────────
export default function Dashboard() {
  const { data: engineHealth, isLoading: loadingEngine } = useQuery({
    queryKey: ['engine-health'],
    queryFn: fetchEngineHealth,
  });

  const { data: statsFlow, isLoading: loadingStats } = useQuery({
    queryKey: ['stats-flow'],
    queryFn: () => fetchStatsFlow(24),
  });

  const { data: alerts } = useQuery({
    queryKey: ['alerts'],
    queryFn: fetchAlerts,
    refetchInterval: 10000,
  });

  const { data: bottlenecks } = useQuery({
    queryKey: ['bottlenecks'],
    queryFn: () => fetchBottlenecks(5),
    refetchInterval: 10000,
  });

  const { data: heatmapData } = useQuery({
    queryKey: ['heatmap'],
    queryFn: fetchHeatmap,
    refetchInterval: 10000,
  });

  if (loadingEngine && loadingStats) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  }

  return (
    <>
      <Title level={4}>Dashboard</Title>
      <KPICards engineHealth={engineHealth} heatmapData={heatmapData} />
      <DensityChart data={statsFlow} />
      <AlertsPanel alerts={alerts} bottlenecks={bottlenecks} />
      <MiniHeatmap data={heatmapData} />
    </>
  );
}
