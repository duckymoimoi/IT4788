import React, { useState } from 'react';
import { Typography, Row, Col, Card, Table, Form, Input, Button, Modal, message, Select, Tag, Space, Divider } from 'antd';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  fetchStations,
  fetchWheelchairs,
  fetchDeviceStatus,
  fetchDeviceTrack,
  reportBroken,
  requestStaff,
  addDevice,
  editDevice,
  delDevice
} from '../api/device';

const { Title, Text } = Typography;

export default function DeviceMonitor() {
  const queryClient = useQueryClient();
  const [trackingId, setTrackingId] = useState('');
  const [reportForm] = Form.useForm();
  const [deviceForm] = Form.useForm();
  
  const [isDeviceModalOpen, setIsDeviceModalOpen] = useState(false);
  const [editingDevice, setEditingDevice] = useState(null);

  // Queries
  const { data: stationsData, isLoading: stationsLoading } = useQuery({
    queryKey: ['stations'],
    queryFn: fetchStations,
  });

  const { data: wheelchairsData, isLoading: wheelchairsLoading } = useQuery({
    queryKey: ['wheelchairs'],
    queryFn: fetchWheelchairs,
    refetchInterval: 10000,
  });

  const { data: trackData, isLoading: trackLoading } = useQuery({
    queryKey: ['deviceTrack', trackingId],
    queryFn: () => Promise.all([fetchDeviceStatus(trackingId), fetchDeviceTrack(trackingId)]).then(([status, track]) => ({ status, track })),
    enabled: !!trackingId,
  });

  // Mutations
  const reportMutation = useMutation({
    mutationFn: reportBroken,
    onSuccess: () => {
      message.success('Report submitted successfully');
      reportForm.resetFields();
    },
    onError: () => message.error('Failed to submit report')
  });

  const staffMutation = useMutation({
    mutationFn: requestStaff,
    onSuccess: () => message.success('Staff requested successfully'),
    onError: () => message.error('Enter a device code first, then request staff')
  });

  const saveDeviceMutation = useMutation({
    mutationFn: (data) => {
      const payload = {
        type: data.type,
        status: data.status,
        current_node_id: data.current_node_id,
      };
      return editingDevice ? editDevice({ ...payload, id: editingDevice.id }) : addDevice(payload);
    },
    onSuccess: () => {
      message.success(`Device ${editingDevice ? 'updated' : 'added'} successfully`);
      setIsDeviceModalOpen(false);
      setEditingDevice(null);
      deviceForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['wheelchairs'] });
      queryClient.invalidateQueries({ queryKey: ['stations'] });
    },
    onError: () => message.error('Failed to save device')
  });

  const deleteDeviceMutation = useMutation({
    mutationFn: (id) => delDevice({ id }),
    onSuccess: () => {
      message.success('Device deleted');
      queryClient.invalidateQueries({ queryKey: ['wheelchairs'] });
      queryClient.invalidateQueries({ queryKey: ['stations'] });
    },
    onError: () => message.error('Failed to delete device')
  });

  // Tables
  const stationColumns = [
    { title: 'ID', dataIndex: 'station_id', key: 'station_id', width: 80 },
    { title: 'Name', dataIndex: 'station_name', key: 'station_name' },
    { title: 'Capacity', dataIndex: 'capacity', key: 'capacity', width: 100 },
    { title: 'Available Wheelchairs', dataIndex: 'available_wheelchairs', key: 'available_wheelchairs', width: 170 },
  ];

  const wheelchairColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
    { title: 'Code', dataIndex: 'device_code', key: 'device_code' },
    { title: 'Type', dataIndex: 'device_type', key: 'device_type',
      render: (type) => <Tag>{type}</Tag>
    },
    { title: 'Status', dataIndex: 'status', key: 'status',
      render: (s) => <Tag color={s === 'available' ? 'green' : s === 'in_use' ? 'blue' : 'orange'}>{s}</Tag>
    },
    { title: 'Current Node', dataIndex: 'current_node_id', key: 'current_node_id' },
    { title: 'Station', dataIndex: 'station_name', key: 'station_name', render: (v) => v || '—' },
    { title: 'Actions', key: 'actions', render: (_, record) => (
        <Space>
          <a onClick={() => {
            setEditingDevice(record);
            deviceForm.setFieldsValue({
              type: record.device_type,
              status: record.status,
              current_node_id: record.current_node_id,
            });
            setIsDeviceModalOpen(true);
          }}>Edit</a>
          <a style={{ color: 'red' }} onClick={() => deleteDeviceMutation.mutate(record.id)}>Delete</a>
        </Space>
      )
    }
  ];

  const onReportFinish = (values) => {
    reportMutation.mutate({
      asset_id: values.asset_id,
      reason: values.reason,
    });
  };

  const onDeviceFinish = (values) => {
    saveDeviceMutation.mutate(values);
  };

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Device Monitor</Title>
        <Space>
          <Button
            onClick={() => staffMutation.mutate({ asset_id: trackingId, node_id: trackData?.track?.current_node_id || '1' })}
            disabled={!trackingId}
          >
            Request Staff
          </Button>
          <Button type="primary" onClick={() => { setEditingDevice(null); deviceForm.resetFields(); setIsDeviceModalOpen(true); }}>
            Add Device
          </Button>
        </Space>
      </div>

      <Row gutter={[16, 16]}>
        <Col span={16}>
          <Card title="Stations & Wheelchairs">
            <Title level={5}>Stations</Title>
            <Table 
              dataSource={stationsData || []} 
              columns={stationColumns} 
              loading={stationsLoading} 
              rowKey="id" 
              size="small"
              pagination={{ pageSize: 5 }}
              style={{ marginBottom: 24 }}
            />
            
            <Title level={5}>Wheelchairs</Title>
            <Table 
              dataSource={wheelchairsData || []} 
              columns={wheelchairColumns} 
              loading={wheelchairsLoading} 
              rowKey="id" 
              size="small"
              pagination={{ pageSize: 5 }}
            />
          </Card>
        </Col>

        <Col span={8}>
          <Card title="Tracking & Reports" style={{ marginBottom: 16 }}>
            <Title level={5}>Track Device</Title>
            <Input.Search 
              placeholder="Enter Device ID" 
              allowClear 
              enterButton="Track" 
              onSearch={setTrackingId}
            />
            {trackingId && (
              <div style={{ marginTop: 16 }}>
                {trackLoading ? <Text type="secondary">Loading...</Text> : trackData ? (
                  <div>
                    <Text strong>Status:</Text> <Tag>{trackData.status?.status || 'Unknown'}</Tag><br/>
                    <Text strong>Location:</Text> {trackData.track?.current_node_id ? `Node ${trackData.track.current_node_id}` : 'N/A'}
                  </div>
                ) : <Text type="danger">Device not found or error</Text>}
              </div>
            )}

            <Divider />

            <Title level={5}>Report Broken Device</Title>
            <Form form={reportForm} layout="vertical" onFinish={onReportFinish}>
              <Form.Item name="asset_id" label="Device Code" rules={[{ required: true }]}>
                <Input placeholder="e.g. WL-001" />
              </Form.Item>
              <Form.Item name="reason" label="Issue Description" rules={[{ required: true }]}>
                <Input.TextArea rows={2} placeholder="What is broken?" />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={reportMutation.isPending} block>
                Submit Report
              </Button>
            </Form>
          </Card>
        </Col>
      </Row>

      {/* Add/Edit Device Modal */}
      <Modal
        title={editingDevice ? "Edit Device" : "Add New Device"}
        open={isDeviceModalOpen}
        onCancel={() => setIsDeviceModalOpen(false)}
        footer={null}
      >
        <Form form={deviceForm} layout="vertical" onFinish={onDeviceFinish}>
          <Form.Item name="type" label="Type" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="wheelchair">Wheelchair</Select.Option>
              <Select.Option value="stretcher">Stretcher</Select.Option>
              <Select.Option value="hospital_cart">Hospital Cart</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="current_node_id" label="Current Node" rules={[{ required: true }]}>
            <Input placeholder="POI code, POI ID, or grid location" />
          </Form.Item>
          <Form.Item name="status" label="Status" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="available">Available</Select.Option>
              <Select.Option value="in_use">In Use</Select.Option>
              <Select.Option value="maintenance">Maintenance</Select.Option>
            </Select>
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={saveDeviceMutation.isPending} block>
            {editingDevice ? "Update" : "Create"}
          </Button>
        </Form>
      </Modal>
    </>
  );
}
