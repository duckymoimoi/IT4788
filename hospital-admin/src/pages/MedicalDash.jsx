import React, { useState } from 'react';
import { Typography, Row, Col, Card, Select, Button, Table, Tabs, Statistic, message, Tag } from 'antd';
import { SyncOutlined, UserOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  fetchTasks,
  fetchQueue,
  fetchPrescription,
  fetchRoomOpen,
  fetchHistory,
  syncNow
} from '../api/medical';

const { Title, Text } = Typography;

export default function MedicalDash() {
  const [selectedRoom, setSelectedRoom] = useState(null);
  const queryClient = useQueryClient();

  const roomOptions = [
    { value: 1, label: 'Room 1 - General' },
    { value: 2, label: 'Room 2 - Cardiology' },
    { value: 3, label: 'Room 3 - Neurology' },
  ];

  // Queries
  const { data: queueData, isLoading: queueLoading } = useQuery({
    queryKey: ['medicalQueue', selectedRoom],
    queryFn: () => fetchQueue(selectedRoom),
    enabled: !!selectedRoom,
    refetchInterval: 10000,
  });

  const { data: roomData, isLoading: roomLoading } = useQuery({
    queryKey: ['roomOpen', selectedRoom],
    queryFn: () => fetchRoomOpen(selectedRoom),
    enabled: !!selectedRoom,
  });

  const { data: tasksData, isLoading: tasksLoading } = useQuery({
    queryKey: ['medicalTasks'],
    queryFn: fetchTasks,
    refetchInterval: 15000,
  });

  const { data: prescriptionData, isLoading: prepLoading } = useQuery({
    queryKey: ['prescriptions'],
    queryFn: fetchPrescription,
  });

  const { data: historyData, isLoading: historyLoading } = useQuery({
    queryKey: ['medicalHistory'],
    queryFn: fetchHistory,
  });

  // Mutation
  const syncMutation = useMutation({
    mutationFn: syncNow,
    onSuccess: () => {
      message.success('HIS Sync successful');
      queryClient.invalidateQueries({ queryKey: ['medicalTasks'] });
      queryClient.invalidateQueries({ queryKey: ['medicalHistory'] });
      queryClient.invalidateQueries({ queryKey: ['prescriptions'] });
    },
    onError: () => {
      message.error('HIS Sync failed');
    }
  });

  // Table Columns
  const taskColumns = [
    { title: 'Task ID', dataIndex: 'id', key: 'id' },
    { title: 'Patient', dataIndex: 'patient_name', key: 'patient_name' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    { title: 'Status', dataIndex: 'status', key: 'status',
      render: (s) => <Tag color={s === 'completed' ? 'green' : s === 'pending' ? 'orange' : 'blue'}>{s}</Tag>
    },
  ];

  const historyColumns = [
    { title: 'Date', dataIndex: 'date', key: 'date' },
    { title: 'Diagnosis', dataIndex: 'diagnosis', key: 'diagnosis' },
    { title: 'Doctor', dataIndex: 'doctor', key: 'doctor' },
  ];

  const prescriptionColumns = [
    { title: 'Medicine', dataIndex: 'medicine_name', key: 'medicine' },
    { title: 'Dosage', dataIndex: 'dosage', key: 'dosage' },
    { title: 'Duration', dataIndex: 'duration', key: 'duration' },
  ];

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Medical Dashboard</Title>
        <Button 
          type="primary" 
          icon={<SyncOutlined spin={syncMutation.isPending} />} 
          onClick={() => syncMutation.mutate()}
          loading={syncMutation.isPending}
        >
          Sync HIS Now
        </Button>
      </div>

      <Row gutter={[16, 16]}>
        {/* Left Column: Room Status & Queue */}
        <Col span={8}>
          <Card title="Room Status & Queue" style={{ height: '100%' }}>
            <Select
              style={{ width: '100%', marginBottom: 16 }}
              placeholder="Select a Room"
              options={roomOptions}
              onChange={(val) => setSelectedRoom(val)}
              allowClear
            />
            
            {selectedRoom ? (
              <Row gutter={[0, 16]}>
                <Col span={12}>
                  <Statistic 
                    title="Current Queue" 
                    value={queueData?.queue_length || 0} 
                    prefix={<UserOutlined />} 
                    loading={queueLoading} 
                  />
                </Col>
                <Col span={12}>
                  <Statistic 
                    title="Est. Wait Time" 
                    value={queueData?.estimated_wait_time || 0} 
                    suffix="mins" 
                    prefix={<ClockCircleOutlined />} 
                    loading={queueLoading} 
                  />
                </Col>
                <Col span={24}>
                  <Text strong>Room Hours:</Text><br/>
                  {roomLoading ? <Text type="secondary">Loading...</Text> : (
                    <Text>{roomData?.open_time || '08:00'} - {roomData?.close_time || '17:00'}</Text>
                  )}
                </Col>
              </Row>
            ) : (
              <Text type="secondary">Please select a room to view its status and queue.</Text>
            )}
          </Card>
        </Col>

        {/* Right Column: Tasks, History, Prescriptions */}
        <Col span={16}>
          <Card title="Tasks & Patient Records">
            <Tabs defaultActiveKey="1" items={[
              {
                key: '1',
                label: 'Current Tasks',
                children: (
                  <Table 
                    dataSource={tasksData || []} 
                    columns={taskColumns} 
                    loading={tasksLoading} 
                    rowKey={(r) => r.id || r.task_id || Math.random()} 
                    size="small"
                    pagination={{ pageSize: 5 }}
                  />
                )
              },
              {
                key: '2',
                label: 'Medical History',
                children: (
                  <Table 
                    dataSource={historyData || []} 
                    columns={historyColumns} 
                    loading={historyLoading} 
                    rowKey={(r) => r.id || r.history_id || Math.random()} 
                    size="small"
                    pagination={{ pageSize: 5 }}
                  />
                )
              },
              {
                key: '3',
                label: 'Prescriptions',
                children: (
                  <Table 
                    dataSource={prescriptionData || []} 
                    columns={prescriptionColumns} 
                    loading={prepLoading} 
                    rowKey={(r) => r.id || r.prescription_id || Math.random()} 
                    size="small"
                    pagination={{ pageSize: 5 }}
                  />
                )
              }
            ]} />
          </Card>
        </Col>
      </Row>
    </>
  );
}
