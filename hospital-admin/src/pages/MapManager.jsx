import React, { useState } from 'react';
import { Typography, Card, Table, Button, Modal, Upload, message, Space, Tag, Popconfirm } from 'antd';
import { UploadOutlined, PlayCircleOutlined, DownloadOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
// Import API methods here when available
// import { fetchMaps, uploadMap, setActiveMap, exportMap } from '../api/mapAdmin';

const { Title } = Typography;

const MapManager = () => {
  const queryClient = useQueryClient();
  const [isUploadModalVisible, setIsUploadModalVisible] = useState(false);

  // TODO: Implement react-query for fetching maps
  // const { data: maps, isLoading } = useQuery({ queryKey: ['maps'], queryFn: fetchMaps });
  const maps = []; // Temporary mock
  const isLoading = false;

  // TODO: Implement mutations for Upload and Set Active
  /*
  const uploadMutation = useMutation({
    mutationFn: uploadMap,
    onSuccess: () => {
      message.success('Upload thành công');
      setIsUploadModalVisible(false);
      queryClient.invalidateQueries({ queryKey: ['maps'] });
    },
    onError: () => message.error('Upload thất bại')
  });

  const setActiveMutation = useMutation({
    mutationFn: setActiveMap,
    onSuccess: () => {
      message.success('Đổi Map Active thành công');
      queryClient.invalidateQueries({ queryKey: ['maps'] });
    },
    onError: (err) => message.error('Lỗi: ' + err.message)
  });
  */

  const handleSetActive = (mapId) => {
    // setActiveMutation.mutate(mapId);
    console.log("Set active map:", mapId);
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'map_id',
      key: 'map_id',
      width: 80,
    },
    {
      title: 'Tên bản đồ',
      dataIndex: 'map_name',
      key: 'map_name',
    },
    {
      title: 'Kích thước',
      key: 'size',
      render: (_, record) => `${record.cols}x${record.rows}`,
    },
    {
      title: 'Trạng thái',
      key: 'status',
      render: (_, record) => (
        record.is_active ? <Tag color="success">Active</Tag> : <Tag color="default">Inactive</Tag>
      ),
    },
    {
      title: 'Hành động',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          {!record.is_active && (
            <Popconfirm 
              title="Bạn có chắc muốn đặt map này làm Active?" 
              onConfirm={() => handleSetActive(record.map_id)}
            >
              <Button type="primary" size="small" icon={<PlayCircleOutlined />}>Set Active</Button>
            </Popconfirm>
          )}
          <Button 
            size="small" 
            icon={<DownloadOutlined />}
            onClick={() => console.log('TODO: Export map', record.map_id)}
          >
            Export
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={2} style={{ margin: 0 }}>Quản lý File Bản Đồ</Title>
        <Button 
          type="primary" 
          icon={<UploadOutlined />} 
          onClick={() => setIsUploadModalVisible(true)}
        >
          Tải Map Mới
        </Button>
      </div>

      <Card>
        <Table 
          dataSource={maps} 
          columns={columns} 
          rowKey="map_id" 
          loading={isLoading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="Tải lên Bản Đồ Mới (.map)"
        open={isUploadModalVisible}
        onCancel={() => setIsUploadModalVisible(false)}
        footer={null}
      >
        <p>TODO: Implement Upload Form here using antd Upload component</p>
        <p>Ghi chú: Cần các trường map_name, rows, cols và file binary.</p>
      </Modal>
    </div>
  );
};

export default MapManager;
