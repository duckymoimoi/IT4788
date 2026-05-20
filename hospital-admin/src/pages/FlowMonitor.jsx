import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Alert,
  Card,
  Col,
  Empty,
  Row,
  Space,
  Spin,
  Statistic,
  Table,
  Tag,
  Typography,
} from 'antd';
import {
  AlertOutlined,
  DashboardOutlined,
  EnvironmentOutlined,
  FireOutlined,
} from '@ant-design/icons';
import { fetchBottlenecks, fetchHeatmap } from '../api/flow';
import { fetchMaps, fetchNodes } from '../api/map';
import GridCanvas from '../components/GridCanvas/GridCanvas';

const { Title, Text } = Typography;

const REFRESH_INTERVAL_MS = 5000;

function formatDensity(value) {
  const density = Number(value || 0);
  return density.toLocaleString('vi-VN');
}

export default function FlowMonitor() {
  const {
    data: maps,
    isLoading: loadingMaps,
    isError: errorMaps,
  } = useQuery({
    queryKey: ['admin-maps'],
    queryFn: fetchMaps,
  });

  const activeMap = useMemo(() => {
    if (!maps?.length) return null;
    return maps.find((m) => m.is_active) || maps[0];
  }, [maps]);

  const activeMapId = activeMap?.map_id ?? null;

  const {
    data: nodes,
    isLoading: loadingNodes,
  } = useQuery({
    queryKey: ['nodes', activeMapId],
    queryFn: () => fetchNodes(activeMapId),
    enabled: !!activeMapId,
  });

  const {
    data: heatmap,
    isLoading: loadingHeatmap,
    isError: errorHeatmap,
  } = useQuery({
    queryKey: ['heatmap'],
    queryFn: fetchHeatmap,
    refetchInterval: REFRESH_INTERVAL_MS,
  });

  const {
    data: bottlenecks,
    isLoading: loadingBottlenecks,
  } = useQuery({
    queryKey: ['bottlenecks'],
    queryFn: () => fetchBottlenecks(10),
    refetchInterval: REFRESH_INTERVAL_MS,
  });

  const normalizedHeatmap = useMemo(() => {
    return (heatmap || [])
      .map((item) => ({
        grid_location: Number(item.grid_location),
        density: Number(item.density || item.count || 0),
      }))
      .filter((item) => Number.isFinite(item.grid_location) && item.density > 0);
  }, [heatmap]);

  const maxDensity = useMemo(() => {
    return normalizedHeatmap.reduce((max, item) => Math.max(max, item.density), 0);
  }, [normalizedHeatmap]);

  const totalDensity = useMemo(() => {
    return normalizedHeatmap.reduce((sum, item) => sum + item.density, 0);
  }, [normalizedHeatmap]);

  const topHeatmapRows = useMemo(() => {
    return [...normalizedHeatmap]
      .sort((a, b) => b.density - a.density)
      .slice(0, 30);
  }, [normalizedHeatmap]);

  const canvasWidth = Math.max(360, Math.min(1120, window.innerWidth - 360));

  const tableColumns = [
    {
      title: 'Grid location',
      dataIndex: 'grid_location',
      key: 'grid_location',
      render: (loc) => {
        const row = activeMap?.cols ? Math.floor(loc / activeMap.cols) : null;
        const col = activeMap?.cols ? loc % activeMap.cols : null;
        return (
          <Space direction="vertical" size={0}>
            <Text strong>Cell #{loc}</Text>
            {row != null && col != null && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                row {row}, col {col}
              </Text>
            )}
          </Space>
        );
      },
    },
    {
      title: 'Density',
      dataIndex: 'density',
      key: 'density',
      sorter: (a, b) => b.density - a.density,
      render: (density) => (
        <Text strong style={{ color: '#cf1322' }}>
          {formatDensity(density)}
        </Text>
      ),
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => {
        const ratio = maxDensity > 0 ? record.density / maxDensity : 0;
        const status = ratio >= 0.7 ? 'High' : ratio >= 0.35 ? 'Medium' : 'Low';
        const color = ratio >= 0.7 ? 'red' : ratio >= 0.35 ? 'orange' : 'green';
        return <Tag color={color}>{status}</Tag>;
      },
    },
  ];

  const bottleneckColumns = [
    {
      title: 'Grid location',
      dataIndex: 'grid_location',
      key: 'grid_location',
      render: (loc) => <Text strong>Cell #{loc}</Text>,
    },
    {
      title: 'Count',
      dataIndex: 'count',
      key: 'count',
      sorter: (a, b) => b.count - a.count,
      render: (count) => (
        <Text strong style={{ color: '#cf1322' }}>
          {formatDensity(count)}
        </Text>
      ),
    },
    {
      title: 'Alert',
      key: 'status',
      render: (_, record) => {
        const ratio = maxDensity > 0 ? Number(record.count || 0) / maxDensity : 0;
        return (
          <Tag color={ratio >= 0.7 ? 'error' : 'warning'}>
            {ratio >= 0.7 ? 'Congestion' : 'Dense'}
          </Tag>
        );
      },
    },
  ];

  if (errorHeatmap || errorMaps) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          message="Khong the tai du lieu flow"
          description="Kiem tra ket noi API flow/map va thu refresh lai man hinh."
          type="error"
          showIcon
        />
      </div>
    );
  }

  return (
    <div style={{ padding: 24, background: '#f5f5f5', minHeight: '100vh' }}>
      <div style={{ marginBottom: 24 }}>
        <Title level={4} style={{ margin: 0 }}>
          Flow Monitor
        </Title>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} md={8}>
          <Card bordered={false}>
            <Statistic
              title="Tracked cells"
              value={normalizedHeatmap.length}
              prefix={<DashboardOutlined style={{ color: '#1677ff' }} />}
            />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card bordered={false}>
            <Statistic
              title="Total density"
              value={totalDensity}
              prefix={<FireOutlined style={{ color: '#cf1322' }} />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card bordered={false}>
            <Statistic
              title="Bottlenecks"
              value={bottlenecks?.length || 0}
              prefix={<AlertOutlined style={{ color: '#fa8c16' }} />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        bordered={false}
        bodyStyle={{ padding: 8 }}
        title={
          <Space wrap>
            <EnvironmentOutlined />
            <Text strong>{activeMap?.map_name || 'Active map'}</Text>
            {activeMap && <Tag color="blue">{activeMap.rows}x{activeMap.cols}</Tag>}
            {activeMap?.is_active && <Tag color="green">Active</Tag>}
            <Tag color="red">Max density: {formatDensity(maxDensity)}</Tag>
          </Space>
        }
        extra={
          <Space size={8}>
            <Text type="secondary" style={{ fontSize: 12 }}>Low</Text>
            <span style={{ display: 'inline-block', width: 44, height: 10, background: '#ffe58f', borderRadius: 4 }} />
            <span style={{ display: 'inline-block', width: 44, height: 10, background: '#ff9c6e', borderRadius: 4 }} />
            <span style={{ display: 'inline-block', width: 44, height: 10, background: '#cf1322', borderRadius: 4 }} />
            <Text type="secondary" style={{ fontSize: 12 }}>High</Text>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        {loadingMaps || loadingNodes || loadingHeatmap ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 96 }}>
            <Spin size="large" tip="Dang tai heatmap..." />
          </div>
        ) : !activeMap ? (
          <Empty description="Chua co active map de hien thi heatmap" />
        ) : (
          <GridCanvas
            rows={activeMap.rows || 33}
            cols={activeMap.cols || 57}
            gridData={activeMap.grid_data || null}
            nodes={nodes || []}
            heatmapData={normalizedHeatmap}
            width={canvasWidth}
            height={620}
          />
        )}
      </Card>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="Top density cells" bordered={false}>
            <Table
              dataSource={topHeatmapRows}
              columns={tableColumns}
              rowKey="grid_location"
              pagination={{ pageSize: 8 }}
              loading={loadingHeatmap}
              size="small"
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Bottleneck alerts" bordered={false}>
            <Table
              dataSource={bottlenecks || []}
              columns={bottleneckColumns}
              rowKey="grid_location"
              pagination={{ pageSize: 8 }}
              loading={loadingBottlenecks}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
}
