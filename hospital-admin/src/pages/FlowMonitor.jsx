import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Table, Card, Row, Col, Statistic, Tag, Spin, Alert, Typography } from 'antd';
import { FireOutlined, DashboardOutlined, AlertOutlined } from '@ant-design/icons';
import axios from 'axios';

const { Title, Text } = Typography;

// Định nghĩa Base URL kết nối tới Production Server chung của nhóm
const BASE_URL = 'https://group3.it4788.sukkaito.id.vn/api';

// 1. Hàm call API lấy dữ liệu Heatmap lưu thông (API số 48)
const fetchHeatmapData = async () => {
  const token = localStorage.getItem('token'); // Lấy JWT token đã lưu khi đăng nhập
  const response = await axios.get(`${BASE_URL}/flow/get_heatmap`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return response.data.data;
};

// 2. Hàm call API lấy danh sách điểm tắc nghẽn (API số 49)
const fetchBottlenecks = async () => {
  const token = localStorage.getItem('token');
  const response = await axios.get(`${BASE_URL}/flow/get_bottlenecks`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return response.data.data;
};

export default function FlowMonitor() {
  // Cấu hình useQuery với cơ chế Auto-refresh mỗi 5000ms (5 giây) theo đúng UI_DESIGN.md
  const { data: heatmap, isLoading: loadingHeatmap, isError: errorHeatmap } = useQuery({
    queryKey: ['heatmap'],
    queryFn: fetchHeatmapData,
    refetchInterval: 5000, 
  });

  const { data: bottlenecks, isLoading: loadingBottlenecks } = useQuery({
    queryKey: ['bottlenecks'],
    queryFn: fetchBottlenecks,
    refetchInterval: 5000,
  });

  // Cấu hình các cột hiển thị cho bảng thống kê Top điểm tắc nghẽn
  const columns = [
    {
      title: 'Vị trí (Grid Location)',
      dataIndex: 'grid_location',
      key: 'grid_location',
      render: (loc) => <strong>Ô lưới #{loc}</strong>,
    },
    {
      title: 'Số lượt tích lũy',
      dataIndex: 'count',
      key: 'count',
      sorter: (a, b) => b.count - a.count,
      render: (count) => <span style={{ color: '#ff4d4f', fontWeight: 'bold' }}>{count} lượt</span>,
    },
    {
      title: 'Trạng thái cảnh báo',
      key: 'status',
      render: (_, record) => (
        <Tag color={record.count > 50 ? 'error' : 'warning'}>
          {record.count > 50 ? 'Tắc nghẽn cao' : 'Mật độ lớn'}
        </Tag>
      ),
    },
  ];

  if (errorHeatmap) {
    return (
      <div style={{ padding: '24px' }}>
        <Alert 
          message="Lỗi kết nối API" 
          description="Không thể đồng bộ dữ liệu luồng giao thông từ Production Server." 
          type="error" 
          showIcon 
        />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* Giữ nguyên phần Header nhận diện của Người C */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={4}>Flow Monitor</Title>
        <Text type="secondary">👤 C — Giám sát luồng người, heatmap, obstacles</Text>
      </div>

      {/* Tầng 1: Các thẻ thống kê tổng quan (Dashboard Widgets) */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col span={8}>
          <Card bordered={false}>
            <Statistic
              title="Tổng số ô lưới đang theo dõi"
              value={heatmap ? heatmap.length : 0}
              prefix={<DashboardOutlined style={{ color: '#1677ff' }} />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card bordered={false}>
            <Statistic
              title="Điểm ùn ứ hiện tại"
              value={bottlenecks ? bottlenecks.length : 0}
              prefix={<AlertOutlined style={{ color: '#faad14' }} />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card bordered={false}>
            <Statistic
              title="Trạng thái Engine luồng"
              value="Đang hoạt động"
              prefix={<FireOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a', fontSize: '18px' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Tầng 2: Chi tiết Bản đồ nhiệt và Điểm nghẽn */}
      <Row gutter={[16, 16]}>
        {/* Cột trái: Dữ liệu thô của Snapshots Heatmap */}
        <Col span={12}>
          <Card title="Dữ liệu Snapshot Bản đồ nhiệt" bordered={false} style={{ height: '100%' }}>
            {loadingHeatmap ? (
              <div style={{ textAlign: 'center', padding: '40px' }}><Spin size="large" /></div>
            ) : (
              <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
                {heatmap?.map((item, index) => (
                  <div 
                    key={index} 
                    style={{ 
                      padding: '10px 0', 
                      borderBottom: '1px solid #f0f0f0', 
                      display: 'flex', 
                      justifyContent: 'space-between',
                      alignItems: 'center'
                    }}
                  >
                    <span>Vị trí ô lưới #{item.grid_location}</span>
                    <Tag color={item.density > 0.7 ? 'red' : item.density > 0.4 ? 'orange' : 'green'}>
                      Mật độ: {(item.density * 100).toFixed(0)}%
                    </Tag>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </Col>

        {/* Cột phải: Bảng Top điểm tắc nghẽn nguy hiểm nhất */}
        <Col span={12}>
          <Card title="Cảnh báo các Điểm tắc nghẽn (Top Bottlenecks)" bordered={false}>
            <Table
              dataSource={bottlenecks}
              columns={columns}
              rowKey="grid_location"
              pagination={{ pageSize: 5 }}
              loading={loadingBottlenecks}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
}