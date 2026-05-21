import { useState } from 'react';
import {
  Typography, Card, Row, Col, List, Badge, Tag, Space, Button,
  Empty, Spin, Popconfirm, message, Avatar, Input, Modal, Form, Select,
} from 'antd';
import {
  MessageOutlined,
  UserOutlined,
  CustomerServiceOutlined,
  CloseCircleOutlined,
  SearchOutlined,
  CommentOutlined,
  CheckCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchRooms, closeRoom, createRoom, fetchChatParticipants } from '../api/chat';
import ChatWindow from '../components/ChatWindow/ChatWindow';

const { Title, Text } = Typography;

const ROOM_STATUS = {
  open: { color: 'green', label: 'Open', icon: <CommentOutlined /> },
  closed: { color: 'default', label: 'Closed', icon: <CheckCircleOutlined /> },
};

export default function ChatSupport() {
  const queryClient = useQueryClient();
  const [selectedRoom, setSelectedRoom] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [form] = Form.useForm();

  const { data: roomsData, isLoading } = useQuery({
    queryKey: ['chat-rooms'],
    queryFn: fetchRooms,
    refetchInterval: 5000,
  });

  const { data: participantsData, isLoading: isParticipantsLoading } = useQuery({
    queryKey: ['chat-participants'],
    queryFn: fetchChatParticipants,
    enabled: createOpen,
  });

  const rooms = roomsData?.rooms || [];
  const patients = participantsData?.patients || [];
  const staffs = participantsData?.staffs || [];

  const patientOptions = patients.map((patient) => ({
    value: patient.user_id,
    label: `${patient.full_name} - ${patient.phone_number}`,
  }));

  const staffOptions = staffs.map((staff) => ({
    value: staff.staff_id,
    label: `${staff.user?.full_name || staff.staff_code} - ${staff.role}`,
  }));

  const filteredRooms = rooms.filter((room) => {
    if (!searchText) return true;
    const q = searchText.toLowerCase();
    return (
      (room.topic || '').toLowerCase().includes(q) ||
      (room.last_message || '').toLowerCase().includes(q) ||
      String(room.conversation_id).includes(q) ||
      String(room.user_id).includes(q)
    );
  });

  const closeMutation = useMutation({
    mutationFn: closeRoom,
    onSuccess: () => {
      message.success('Room closed');
      queryClient.invalidateQueries({ queryKey: ['chat-rooms'] });
      if (selectedRoom) {
        setSelectedRoom({ ...selectedRoom, status: 'closed' });
      }
    },
    onError: (err) => {
      message.error(err.response?.data?.message || 'Cannot close room');
    },
  });

  const createMutation = useMutation({
    mutationFn: createRoom,
    onSuccess: (res) => {
      message.success('Room created');
      setCreateOpen(false);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['chat-rooms'] });
      if (res?.data) {
        setSelectedRoom(res.data);
      }
    },
    onError: (err) => {
      message.error(err.response?.data?.message || 'Cannot create room');
    },
  });

  const openCreateModal = () => {
    setCreateOpen(true);
  };

  return (
    <>
      <Title level={4} style={{ marginBottom: 16 }}>
        <MessageOutlined style={{ color: '#1677ff', marginRight: 8 }} />
        Chat Support
      </Title>

      <Modal
        title="Create Chat Room"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={() => form.submit()}
        okText="Create"
        cancelText="Cancel"
        confirmLoading={createMutation.isPending}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => createMutation.mutate(values)}
        >
          <Form.Item
            label="Patient"
            name="user_id"
            rules={[{ required: true, message: 'Select a patient' }]}
          >
            <Select
              showSearch
              loading={isParticipantsLoading}
              options={patientOptions}
              placeholder="Select a patient"
              optionFilterProp="label"
            />
          </Form.Item>
          <Form.Item
            label="Assigned staff"
            name="staff_id"
            rules={[{ required: true, message: 'Select a staff member' }]}
          >
            <Select
              showSearch
              loading={isParticipantsLoading}
              options={staffOptions}
              placeholder="Select a staff member"
              optionFilterProp="label"
            />
          </Form.Item>
          <Form.Item label="Topic" name="topic">
            <Input placeholder="Example: Navigation support" maxLength={200} />
          </Form.Item>
        </Form>
      </Modal>

      <Row gutter={16} style={{ height: 'calc(100vh - 220px)', minHeight: 500 }}>
        <Col span={7}>
          <Card
            title={
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space>
                  <CommentOutlined />
                  <span>{`Rooms (${rooms.length})`}</span>
                </Space>
                <Button type="primary" size="small" icon={<PlusOutlined />} onClick={openCreateModal}>
                  New
                </Button>
              </div>
            }
            bodyStyle={{ padding: 0, height: 'calc(100% - 56px)', overflowY: 'auto' }}
            style={{ height: '100%' }}
          >
            <div style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0' }}>
              <Input
                placeholder="Search rooms..."
                prefix={<SearchOutlined />}
                size="small"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                allowClear
              />
            </div>

            {isLoading ? (
              <div style={{ textAlign: 'center', padding: 40 }}>
                <Spin />
              </div>
            ) : filteredRooms.length === 0 ? (
              <Empty
                description="No chat rooms"
                style={{ padding: 40 }}
                imageStyle={{ height: 60 }}
              />
            ) : (
              <List
                dataSource={filteredRooms}
                renderItem={(room) => {
                  const isSelected = selectedRoom?.conversation_id === room.conversation_id;
                  const statusCfg = ROOM_STATUS[room.status] || ROOM_STATUS.open;

                  return (
                    <List.Item
                      onClick={() => setSelectedRoom(room)}
                      style={{
                        padding: '12px 16px',
                        cursor: 'pointer',
                        background: isSelected ? '#e6f4ff' : 'transparent',
                        borderLeft: isSelected ? '3px solid #1677ff' : '3px solid transparent',
                        transition: 'all 0.2s',
                      }}
                      onMouseEnter={(e) => {
                        if (!isSelected) e.currentTarget.style.background = '#fafafa';
                      }}
                      onMouseLeave={(e) => {
                        if (!isSelected) e.currentTarget.style.background = 'transparent';
                      }}
                    >
                      <List.Item.Meta
                        avatar={
                          <Badge count={room.unread_count || 0} size="small" offset={[-2, 2]}>
                            <Avatar
                              icon={<UserOutlined />}
                              style={{ backgroundColor: isSelected ? '#1677ff' : '#d9d9d9' }}
                            />
                          </Badge>
                        }
                        title={
                          <Space size={4}>
                            <Text strong={!!room.unread_count} style={{ fontSize: 13 }} ellipsis>
                              {room.topic || `Room #${room.conversation_id}`}
                            </Text>
                            <Tag
                              color={statusCfg.color}
                              style={{ fontSize: 10, lineHeight: '16px', padding: '0 4px' }}
                            >
                              {statusCfg.label}
                            </Tag>
                          </Space>
                        }
                        description={
                          <Text type="secondary" style={{ fontSize: 12 }} ellipsis>
                            {room.last_message || `Patient #${room.user_id}`}
                          </Text>
                        }
                      />
                    </List.Item>
                  );
                }}
              />
            )}
          </Card>
        </Col>

        <Col span={17}>
          <Card
            title={
              selectedRoom ? (
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Space>
                    <CustomerServiceOutlined style={{ color: '#1677ff' }} />
                    <span>{selectedRoom.topic || `Room #${selectedRoom.conversation_id}`}</span>
                    <Tag color={ROOM_STATUS[selectedRoom.status]?.color || 'green'}>
                      {ROOM_STATUS[selectedRoom.status]?.label || selectedRoom.status}
                    </Tag>
                  </Space>
                  {selectedRoom.status === 'open' && (
                    <Popconfirm
                      title="Close chat room?"
                      description="No one can send more messages after this room is closed."
                      onConfirm={() => closeMutation.mutate(selectedRoom.conversation_id)}
                      okText="Close"
                      cancelText="Cancel"
                      okButtonProps={{ danger: true }}
                    >
                      <Button
                        size="small"
                        danger
                        icon={<CloseCircleOutlined />}
                        loading={closeMutation.isPending}
                      >
                        Close
                      </Button>
                    </Popconfirm>
                  )}
                </div>
              ) : (
                <Space>
                  <MessageOutlined />
                  <span>Messages</span>
                </Space>
              )
            }
            bodyStyle={{
              padding: selectedRoom ? '0 16px 16px' : 16,
              height: 'calc(100% - 56px)',
              display: 'flex',
              flexDirection: 'column',
            }}
            style={{ height: '100%' }}
          >
            <ChatWindow
              conversationId={selectedRoom?.conversation_id || null}
              isClosed={selectedRoom?.status === 'closed'}
            />
          </Card>
        </Col>
      </Row>
    </>
  );
}
