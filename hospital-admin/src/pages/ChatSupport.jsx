import { useState } from 'react';
import {
  Typography, Card, Row, Col, List, Badge, Tag, Space, Button,
  Empty, Spin, Popconfirm, message, Avatar, Input,
} from 'antd';
import {
  MessageOutlined,
  UserOutlined,
  CustomerServiceOutlined,
  CloseCircleOutlined,
  SearchOutlined,
  CommentOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchRooms, closeRoom } from '../api/chat';
import ChatWindow from '../components/ChatWindow/ChatWindow';

const { Title, Text } = Typography;

// ─── Room status config ───────────────────────────────────────
const ROOM_STATUS = {
  open: { color: 'green', label: 'Đang mở', icon: <CommentOutlined /> },
  closed: { color: 'default', label: 'Đã đóng', icon: <CheckCircleOutlined /> },
};

export default function ChatSupport() {
  const queryClient = useQueryClient();
  const [selectedRoom, setSelectedRoom] = useState(null);
  const [searchText, setSearchText] = useState('');

  // ── Fetch rooms ──
  const { data: roomsData, isLoading } = useQuery({
    queryKey: ['chat-rooms'],
    queryFn: fetchRooms,
    refetchInterval: 5000, // auto-refresh 5s
  });

  const rooms = roomsData?.rooms || [];

  // ── Filter rooms by search ──
  const filteredRooms = rooms.filter((room) => {
    if (!searchText) return true;
    const q = searchText.toLowerCase();
    return (
      (room.topic || '').toLowerCase().includes(q) ||
      (room.last_message || '').toLowerCase().includes(q) ||
      String(room.conversation_id).includes(q)
    );
  });

  // ── Close room mutation ──
  const closeMutation = useMutation({
    mutationFn: closeRoom,
    onSuccess: () => {
      message.success('Đã đóng phòng chat');
      queryClient.invalidateQueries({ queryKey: ['chat-rooms'] });
      if (selectedRoom) {
        setSelectedRoom({ ...selectedRoom, status: 'closed' });
      }
    },
    onError: (err) => {
      message.error(err.response?.data?.message || 'Không thể đóng phòng');
    },
  });

  // ── Select room handler ──
  const handleSelectRoom = (room) => {
    setSelectedRoom(room);
  };

  return (
    <>
      <Title level={4} style={{ marginBottom: 16 }}>
        <MessageOutlined style={{ color: '#1677ff', marginRight: 8 }} />
        Chat Support
      </Title>

      <Row gutter={16} style={{ height: 'calc(100vh - 220px)', minHeight: 500 }}>
        {/* ── Sidebar: Room List (E4) ── */}
        <Col span={7}>
          <Card
            title={
              <Space>
                <CommentOutlined />
                {`Phòng chat (${rooms.length})`}
              </Space>
            }
            bodyStyle={{ padding: 0, height: 'calc(100% - 56px)', overflowY: 'auto' }}
            style={{ height: '100%' }}
          >
            {/* Search bar */}
            <div style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0' }}>
              <Input
                placeholder="Tìm phòng..."
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
                description="Không có phòng chat nào"
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
                      onClick={() => handleSelectRoom(room)}
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
                            <Text
                              strong={!!room.unread_count}
                              style={{ fontSize: 13 }}
                              ellipsis
                            >
                              {room.topic || `Phòng #${room.conversation_id}`}
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
                          <Text
                            type="secondary"
                            style={{ fontSize: 12 }}
                            ellipsis
                          >
                            {room.last_message || 'Chưa có tin nhắn'}
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

        {/* ── Main: Chat Window (E5, E6, E7) ── */}
        <Col span={17}>
          <Card
            title={
              selectedRoom ? (
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Space>
                    <CustomerServiceOutlined style={{ color: '#1677ff' }} />
                    <span>{selectedRoom.topic || `Phòng #${selectedRoom.conversation_id}`}</span>
                    <Tag color={ROOM_STATUS[selectedRoom.status]?.color || 'green'}>
                      {ROOM_STATUS[selectedRoom.status]?.label || selectedRoom.status}
                    </Tag>
                  </Space>
                  {/* E7: Close room button */}
                  {selectedRoom.status === 'open' && (
                    <Popconfirm
                      title="Đóng phòng chat?"
                      description="Sau khi đóng, không ai có thể gửi thêm tin nhắn."
                      onConfirm={() => closeMutation.mutate(selectedRoom.conversation_id)}
                      okText="Đóng"
                      cancelText="Hủy"
                      okButtonProps={{ danger: true }}
                    >
                      <Button
                        size="small"
                        danger
                        icon={<CloseCircleOutlined />}
                        loading={closeMutation.isPending}
                      >
                        Đóng phòng
                      </Button>
                    </Popconfirm>
                  )}
                </div>
              ) : (
                <Space>
                  <MessageOutlined />
                  <span>Tin nhắn</span>
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
