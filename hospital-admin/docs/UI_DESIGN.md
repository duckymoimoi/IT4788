# Hướng dẫn UI/UX — Design System

## Theme chung

```
Primary Color:   #1677ff (Ant Design Blue)
Success:         #52c41a
Warning:         #faad14
Error:           #ff4d4f
Background:      #f5f5f5
Sidebar BG:      #001529 (dark navy)
Card BG:         #ffffff
Text Primary:    #000000e0
Text Secondary:  #00000073
```

## Layout chung

```
┌──────────────────────────────────────────────────┐
│ Header (64px)                          [A] Admin ▼│
├────────┬─────────────────────────────────────────┤
│        │                                         │
│ Side   │  Page Content                           │
│ bar    │                                         │
│ (240px)│  ┌─ Breadcrumb ──────────────────────┐  │
│        │  │ Dashboard > Flow Monitor          │  │
│ - Dash │  └───────────────────────────────────┘  │
│ - Map  │                                         │
│ - Flow │  ┌─ Content Area ────────────────────┐  │
│ - Sim  │  │                                   │  │
│ - Med  │  │  (each page renders here)         │  │
│ - Dev  │  │                                   │  │
│ - SOS  │  └───────────────────────────────────┘  │
│ - Chat │                                         │
│ - Eng  │                                         │
│ - Set  │                                         │
│        │                                         │
├────────┴─────────────────────────────────────────┤
│ Footer: Hospital Navigator v1.0                   │
└──────────────────────────────────────────────────┘
```

## Sidebar Menu

```jsx
const menuItems = [
  { key: '/',          icon: <DashboardOutlined />,    label: 'Dashboard' },
  { key: '/map',       icon: <EnvironmentOutlined />,  label: 'Map Editor' },
  { key: '/flow',      icon: <HeatMapOutlined />,      label: 'Flow Monitor' },
  { key: '/sim',       icon: <PlayCircleOutlined />,   label: 'Simulation' },
  { key: '/medical',   icon: <MedicineBoxOutlined />,  label: 'Medical' },
  { key: '/device',    icon: <LaptopOutlined />,       label: 'Device' },
  { key: '/sos',       icon: <AlertOutlined />,        label: 'SOS' },
  { key: '/chat',      icon: <MessageOutlined />,      label: 'Chat' },
  { key: '/engine',    icon: <SettingOutlined />,      label: 'Engine' },
  { key: '/settings',  icon: <ToolOutlined />,         label: 'Settings' },
];
```

## Component patterns

### 1. Page Layout

Mỗi trang dùng pattern này:

```jsx
import { Card, Row, Col, Typography } from 'antd';

export default function PageName() {
  return (
    <>
      <Typography.Title level={4}>Page Title</Typography.Title>
      <Row gutter={[16, 16]}>
        <Col span={16}>
          <Card title="Main Content">
            {/* Table, Canvas, etc. */}
          </Card>
        </Col>
        <Col span={8}>
          <Card title="Side Panel">
            {/* Stats, filters, etc. */}
          </Card>
        </Col>
      </Row>
    </>
  );
}
```

### 2. Data Table

```jsx
import { useQuery } from '@tanstack/react-query';
import { Table, Tag } from 'antd';
import { fetchObstacles } from '../api/flow';

export default function ObstacleTable() {
  const { data, isLoading } = useQuery({
    queryKey: ['obstacles'],
    queryFn: fetchObstacles,
    refetchInterval: 10000, // auto-refresh 10s
  });

  const columns = [
    { title: 'ID', dataIndex: 'report_id', key: 'id' },
    { title: 'Location', dataIndex: 'grid_location', key: 'loc' },
    { title: 'Status', dataIndex: 'status', key: 'status',
      render: (s) => <Tag color={s === 'pending' ? 'orange' : 'green'}>{s}</Tag>
    },
  ];

  return <Table dataSource={data} columns={columns} loading={isLoading} rowKey="report_id" />;
}
```

### 3. API Hook Pattern

```jsx
// Luôn dùng TanStack Query, không dùng useEffect + fetch
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// GET → useQuery
const { data } = useQuery({ queryKey: ['heatmap'], queryFn: fetchHeatmap });

// POST/PATCH/DELETE → useMutation + invalidate
const queryClient = useQueryClient();
const mutation = useMutation({
  mutationFn: resolveObstacle,
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['obstacles'] }),
});
```

### 4. Auto-refresh

```jsx
// Dùng refetchInterval của TanStack Query
const { data } = useQuery({
  queryKey: ['heatmap'],
  queryFn: fetchHeatmap,
  refetchInterval: 5000,  // 5 giây
});
```

## Quy tắc responsive

- Desktop-first (min-width: 1200px)
- Sidebar collapse ở < 768px
- Không cần mobile layout (admin panel dùng trên PC)

## Card KPI Pattern (Dashboard)

```jsx
<Row gutter={16}>
  <Col span={6}>
    <Card>
      <Statistic title="Online Users" value={120} prefix={<UserOutlined />} />
    </Card>
  </Col>
  <Col span={6}>
    <Card>
      <Statistic title="Active SOS" value={3} valueStyle={{ color: '#ff4d4f' }} />
    </Card>
  </Col>
</Row>
```
