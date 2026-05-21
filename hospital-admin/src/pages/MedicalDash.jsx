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
import { fetchSyncFull } from '../api/map';

const { Title, Text } = Typography;

export default function MedicalDash() {
  const [selectedRoom, setSelectedRoom] = useState(null);
  const queryClient = useQueryClient();

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

  const tasks = tasksData || [];

  const { data: mapData, isLoading: roomsLoading } = useQuery({
    queryKey: ['medicalRoomPois'],
    queryFn: () => fetchSyncFull(),
  });

  const roomOptions = React.useMemo(() => {
    const taskPoiIds = new Set(tasks.map((task) => task.poi_id).filter(Boolean));
    const pois = mapData?.pois || [];
    const optionsByID = new Map();

    pois
      .filter((poi) => poi.poi_type === 'room' || taskPoiIds.has(poi.poi_id))
      .forEach((poi) => {
        optionsByID.set(poi.poi_id, {
          value: poi.poi_id,
          label: `${poi.poi_name || poi.poi_code || 'Room'} (POI #${poi.poi_id})`,
        });
      });

    taskPoiIds.forEach((poiID) => {
      if (!optionsByID.has(poiID)) {
        optionsByID.set(poiID, {
          value: poiID,
          label: `Room POI #${poiID}`,
        });
      }
    });

    return Array.from(optionsByID.values()).sort((a, b) => a.value - b.value);
  }, [mapData, tasks]);

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
    { title: 'Task ID', dataIndex: 'treatment_id', key: 'treatment_id', width: 90 },
    { title: 'Task Name', dataIndex: 'task_name', key: 'task_name' },
    { title: 'Patient ID', dataIndex: 'user_id', key: 'user_id', width: 100 },
    { title: 'Room POI', dataIndex: 'poi_id', key: 'poi_id', width: 100 },
    {
      title: 'Type',
      dataIndex: 'task_type',
      key: 'task_type',
      width: 110,
      render: (type) => type || '-',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (s) => <Tag color={s === 'completed' ? 'green' : s === 'pending' ? 'orange' : 'blue'}>{s || '-'}</Tag>
    },
  ];

  const historyColumns = [
    { title: 'Task ID', dataIndex: 'treatment_id', key: 'treatment_id', width: 90 },
    { title: 'Task Name', dataIndex: 'task_name', key: 'task_name' },
    { title: 'Type', dataIndex: 'task_type', key: 'task_type', width: 110 },
    {
      title: 'Completed At',
      dataIndex: 'completed_at',
      key: 'completed_at',
      render: (value) => value ? new Date(value).toLocaleString() : '-',
    },
  ];

  const prescriptionColumns = [
    { title: 'Prescription ID', dataIndex: 'prescription_id', key: 'prescription_id', width: 130 },
    {
      title: 'Items',
      dataIndex: 'items_json',
      key: 'items_json',
      render: (value) => {
        try {
          const items = JSON.parse(value || '[]');
          return items.map((item) => `${item.name || 'Medicine'} (${item.dosage || 'no dosage'})`).join(', ');
        } catch {
          return value || '-';
        }
      },
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (s) => <Tag color={s === 'dispensed' ? 'green' : 'orange'}>{s || '-'}</Tag>,
    },
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
              loading={roomsLoading || tasksLoading}
              onChange={(val) => setSelectedRoom(val)}
              allowClear
            />
            
            {selectedRoom ? (
              <Row gutter={[0, 16]}>
                <Col span={12}>
                  <Statistic 
                    title="Current Queue" 
                    value={queueData?.waiting_count || 0} 
                    prefix={<UserOutlined />} 
                    loading={queueLoading} 
                  />
                </Col>
                <Col span={12}>
                  <Statistic 
                    title="Est. Wait Time" 
                    value={queueData?.avg_wait_minutes || 0} 
                    suffix="mins" 
                    prefix={<ClockCircleOutlined />} 
                    loading={queueLoading} 
                  />
                </Col>
                <Col span={24}>
                  <Text strong>Room Hours:</Text><br/>
                  {roomLoading ? <Text type="secondary">Loading...</Text> : (
                    <Text>{roomData?.open || '08:00'} - {roomData?.close || '17:00'}</Text>
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
                    dataSource={tasks} 
                    columns={taskColumns} 
                    loading={tasksLoading} 
                    rowKey={(r) => r.treatment_id} 
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
                    rowKey={(r) => r.treatment_id} 
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
                    rowKey={(r) => r.prescription_id} 
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
