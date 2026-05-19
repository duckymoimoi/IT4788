import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, Button, Table, Tag, Space, notification, Alert, Spin, Typography } from 'antd';
import { PlayCircleOutlined, StopOutlined, CheckCircleOutlined } from '@ant-design/icons';
import axios from 'axios';

const { Title, Text } = Typography;

// Đường dẫn Base URL kết nối trực tiếp đến Production Server của nhóm
const BASE_URL = 'https://group3.it4788.sukkaito.id.vn/api';

// --- TẦNG CALL API (KHỚP HOÀN TOÀN VỚI TÀI LIỆU SWAGGER & API REFERENCE) ---

const fetchSimStatus = async () => {
  const token = localStorage.getItem('token');
  const res = await axios.get(`${BASE_URL}/simulate/status`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return res.data.data;
};

const fetchObstacles = async () => {
  const token = localStorage.getItem('token');
  const res = await axios.get(`${BASE_URL}/flow/get_obstacles`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return res.data.data;
};

const startSimulation = async () => {
  const token = localStorage.getItem('token');
  return axios.post(`${BASE_URL}/simulate/start`, {
    map_id: 1,
    output_file: "warehouse_small.json",
    tick_rate_ms: 1000
  }, { headers: { Authorization: `Bearer ${token}` } });
};

const stopSimulation = async () => {
  const token = localStorage.getItem('token');
  return axios.post(`${BASE_URL}/simulate/stop`, {}, {
    headers: { Authorization: `Bearer ${token}` }
  });
};

const resolveObstacle = async (reportID) => {
  const token = localStorage.getItem('token');
  return axios.post(`${BASE_URL}/flow/resolve_obstacle`, {
    report_id: reportID,
    action: "resolve"
  }, { headers: { Authorization: `Bearer ${token}` } });
};

// --- COMPONENT CHÍNH ---

export default function SimControl() {
  const queryClient = useQueryClient();

  // 1. useQuery lấy trạng thái hoạt động của Simulation (Tự động cập nhật mỗi 3 giây)
  const { data: simStatus, isLoading: loadingStatus } = useQuery({
    queryKey: ['simStatus'],
    queryFn: fetchSimStatus,
    refetchInterval: 3000 // Tự động refetch để cập nhật trạng thái bật/tắt nút bấm
  });

  // 2. useQuery lấy danh sách các sự cố/vật cản hành lang bệnh viện
  const { data: obstacles, isLoading: loadingObstacles } = useQuery({
    queryKey: ['obstacles'],
    queryFn: fetchObstacles
  });

  // --- QUẢN LÝ CÁC MUTATION + INVALIDATE CACHE THEO ĐÚNG HƯỚNG DẪN UI_DESIGN.MD ---

  const startMutation = useMutation({
    mutationFn: startSimulation,
    onSuccess: () => {
      notification.success({ 
        message: 'Mô phỏng thành công', 
        description: 'Hệ thống giả lập di chuyển MAPF đã được kích hoạt.' 
      });
      queryClient.invalidateQueries({ queryKey: ['simStatus'] });
    }
  });

  const stopMutation = useMutation({
    mutationFn: stopSimulation,
    onSuccess: () => {
      notification.warning({ 
        message: 'Mô phỏng tạm dừng', 
        description: 'Đã phát lệnh dừng toàn bộ tiến trình giả lập luồng giao thông.' 
      });
      queryClient.invalidateQueries({ queryKey: ['simStatus'] });
    }
  });

  const resolveMutation = useMutation({
    mutationFn: resolveObstacle,
    onSuccess: () => {
      notification.success({ 
        message: 'Xử lý thành công', 
        description: 'Đã gỡ bỏ vật cản. Các đoạn hành lang tương ứng đã thông suốt.' 
      });
      queryClient.invalidateQueries({ queryKey: ['obstacles'] });
    }
  });

  // Cấu hình các cột cho bảng quản lý vật cản
  const columns = [
    {
      title: 'Mã số',
      dataIndex: 'report_id',
      key: 'report_id',
    },
    {
      title: 'Vị trí sự cố',
      dataIndex: 'grid_location',
      key: 'grid_location',
      render: (loc) => <Tag color="blue">Ô lưới #{loc}</Tag>
    },
    {
      title: 'Loại vật cản',
      dataIndex: 'report_type',
      key: 'report_type',
      render: (type) => {
        const mapping = { wet_floor: 'Sàn ướt trơn trượt', broken_elevator: 'Thang máy hỏng', blocked: 'Hành lang bị chặn' };
        return <span>{mapping[type] || type}</span>;
      }
    },
    {
      title: 'Mô tả chi tiết',
      dataIndex: 'description',
      key: 'description',
      render: (desc) => desc || <Text type="secondary">Không có mô tả</Text>
    },
    {
      title: 'Trạng thái',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={status === 'resolved' ? 'success' : 'error'}>
          {status === 'resolved' ? 'Đã xử lý' : 'Chưa giải quyết'}
        </Tag>
      )
    },
    {
      title: 'Thao tác điều phối',
      key: 'action',
      render: (_, record) => (
        record.status !== 'resolved' ? (
          <Button 
            type="primary" 
            size="small"
            onClick={() => resolveMutation.mutate(record.report_id)}
            loading={resolveMutation.isPending}
          >
            Giải quyết sự cố
          </Button>
        ) : <span style={{ color: '#52c41a', fontWeight: 'bold' }}>Nghiệp vụ hoàn thành</span>
      )
    }
  ];

  return (
    <div style={{ padding: '24px', background: '#f5f5f5', minHeight: '100vh' }}>
      {/* Giữ nguyên phần Tiêu đề định danh của Người C */}
      <div style={{ marginBottom: '24px' }}>
        <Title level={4}>Simulation Control</Title>
        <Text type="secondary">👤 C — Điều khiển mô phỏng MAPF</Text>
      </div>

      {/* Tầng 1: Khối điều khiển kích hoạt kịch bản mô phỏng */}
      <Card title="Quản trị Tiến trình Mô phỏng (Simulation)" style={{ marginBottom: '24px' }} bordered={false}>
        {loadingStatus ? <Spin /> : (
          <Space size="large">
            <div>
              Trạng thái Engine:{' '}
              {simStatus?.is_running ? (
                <Tag color="success">ĐANG CHẠY GIẢ LẬP</Tag>
              ) : (
                <Tag color="default">ĐANG DỪNG</Tag>
              )}
            </div>
            <Button 
              type="primary" 
              icon={<PlayCircleOutlined />} 
              disabled={simStatus?.is_running}
              onClick={() => startMutation.mutate()}
              loading={startMutation.isPending}
            >
              Kích hoạt Mô phỏng
            </Button>
            <Button 
              type="primary" 
              danger 
              icon={<StopOutlined />} 
              disabled={!simStatus?.is_running}
              onClick={() => stopMutation.mutate()}
              loading={stopMutation.isPending}
            >
              Dừng mô phỏng
            </Button>
          </Space>
        )}
      </Card>

      {/* Tầng 2: Khối quản lý danh sách vật cản chướng ngại vật */}
      <Card title="Danh sách Sự cố & Báo cáo Vật cản đường đi" bordered={false}>
        <Alert 
          message="Nguyên lý tương tác hệ thống:" 
          description="Các báo cáo sự cố chưa giải quyết sẽ trực tiếp đẩy trọng số (weight) của hành lang lên cao, ép thuật toán Dijkstra tìm đường đi vòng để bảo đảm an toàn cho bệnh nhân." 
          type="info" 
          showIcon 
          style={{ marginBottom: '16px' }}
        />
        <Table 
          dataSource={obstacles} 
          columns={columns} 
          rowKey="report_id"
          loading={loadingObstacles}
          pagination={{ pageSize: 5 }}
        />
      </Card>
    </div>
  );
}