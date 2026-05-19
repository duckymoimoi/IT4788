import { useState } from 'react';
import { Layout, Menu, theme, Dropdown, Avatar, Space, Typography } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  DashboardOutlined,
  EnvironmentOutlined,
  HeatMapOutlined,
  PlayCircleOutlined,
  MedicineBoxOutlined,
  LaptopOutlined,
  AlertOutlined,
  MessageOutlined,
  SettingOutlined,
  ToolOutlined,
  UserOutlined,
  LogoutOutlined,
  FileOutlined,
  EditOutlined,
  CodeOutlined,
} from '@ant-design/icons';
import useAuthStore from '../../stores/authStore';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: 'Dashboard' },
  {
    key: 'map-group',
    icon: <EnvironmentOutlined />,
    label: 'Bản đồ',
    children: [
      { key: '/map-editor', label: 'Map Editor' },
      { key: '/map-builder', icon: <EditOutlined />, label: 'Map Builder' },
      { key: '/map-manager', icon: <FileOutlined />, label: 'Map Manager' },
    ],
  },
  { key: '/flow', icon: <HeatMapOutlined />, label: 'Flow Monitor' },
  { key: '/sim', icon: <PlayCircleOutlined />, label: 'Simulation' },
  { key: '/medical', icon: <MedicineBoxOutlined />, label: 'Medical' },
  { key: '/device', icon: <LaptopOutlined />, label: 'Device' },
  { key: '/sos', icon: <AlertOutlined />, label: 'SOS' },
  { key: '/chat', icon: <MessageOutlined />, label: 'Chat' },
  { key: '/engine', icon: <SettingOutlined />, label: 'Engine' },
  { key: '/settings', icon: <ToolOutlined />, label: 'Settings' },
  { key: '/api-logs', icon: <CodeOutlined />, label: 'API Logger' },
];

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const userMenuItems = [
    {
      key: 'user-info',
      label: (
        <Space direction="vertical" size={0}>
          <Text strong>{user?.full_name || 'Admin'}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>{user?.role || 'admin'}</Text>
        </Space>
      ),
      disabled: true,
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Đăng xuất',
      danger: true,
      onClick: handleLogout,
    },
  ];

  // Determine which sidebar key is active (handles nested children)
  const selectedKey = (() => {
    const path = location.pathname;
    let bestMatch = '/';
    let bestLen = 0;
    for (const item of menuItems) {
      if (item.children) {
        for (const c of item.children) {
          if (path === c.key || path.startsWith(c.key + '/')) {
            if (c.key.length > bestLen) { bestMatch = c.key; bestLen = c.key.length; }
          }
        }
      } else {
        if (item.key === '/') { if (path === '/') { bestMatch = '/'; bestLen = 999; } }
        else if ((path === item.key || path.startsWith(item.key + '/')) && item.key.length > bestLen) {
          bestMatch = item.key; bestLen = item.key.length;
        }
      }
    }
    return bestMatch;
  })();

  const openKeys = menuItems
    .filter((item) => item.children?.some((c) => location.pathname.startsWith(c.key)))
    .map((item) => item.key);

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={(value) => setCollapsed(value)}
        style={{ overflow: 'auto', height: '100vh', position: 'sticky', top: 0, left: 0 }}
      >
        <div
          style={{
            height: 48,
            margin: 12,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            overflow: 'hidden',
          }}
        >
          <Text
            strong
            style={{ color: '#fff', fontSize: collapsed ? 18 : 16, whiteSpace: 'nowrap' }}
          >
            {collapsed ? 'H' : 'Hospital Admin'}
          </Text>
        </div>
        <Menu
          theme="dark"
          selectedKeys={[selectedKey]}
          defaultOpenKeys={openKeys}
          mode="inline"
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 24px',
            background: colorBgContainer,
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
            <Space style={{ cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#1677ff' }} />
              <Text strong>{user?.full_name || 'Admin'}</Text>
            </Space>
          </Dropdown>
        </Header>
        <Content style={{ margin: 16 }}>
          <div
            style={{
              padding: 24,
              minHeight: 360,
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
            }}
          >
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
}
