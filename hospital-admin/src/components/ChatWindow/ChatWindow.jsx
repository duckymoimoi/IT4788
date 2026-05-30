import { useState, useEffect, useRef, useCallback } from 'react';
import { Input, Button, Space, Spin, Empty, Typography, Avatar, Tooltip } from 'antd';
import {
  SendOutlined,
  UserOutlined,
  CustomerServiceOutlined,
  RobotOutlined,
  LoadingOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchMessages, sendMessage, markRead, getWSUrl } from '../../api/chat';

const { Text } = Typography;

// ─── Message Bubble ───────────────────────────────────────────
function MessageBubble({ msg, isStaff }) {
  const isMe = msg.sender_type === 'staff';
  const time = msg.created_at
    ? new Date(msg.created_at).toLocaleString('vi-VN', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : '';

  const avatarIcon =
    msg.sender_type === 'staff' ? (
      <CustomerServiceOutlined />
    ) : msg.sender_type === 'bot' ? (
      <RobotOutlined />
    ) : (
      <UserOutlined />
    );

  const avatarColor =
    msg.sender_type === 'staff'
      ? '#1677ff'
      : msg.sender_type === 'bot'
        ? '#722ed1'
        : '#87d068';

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: isMe ? 'flex-end' : 'flex-start',
        marginBottom: 12,
        gap: 8,
        alignItems: 'flex-end',
      }}
    >
      {!isMe && (
        <Avatar size={28} icon={avatarIcon} style={{ backgroundColor: avatarColor, flexShrink: 0 }} />
      )}
      <div style={{ maxWidth: '65%' }}>
        <div
          style={{
            padding: '8px 14px',
            borderRadius: isMe ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
            background: isMe ? '#1677ff' : '#f0f0f0',
            color: isMe ? '#fff' : '#000',
            wordBreak: 'break-word',
            lineHeight: 1.5,
          }}
        >
          {msg.text_content || (
            <a
              href={msg.media_url}
              target="_blank"
              rel="noreferrer"
              style={{ color: isMe ? '#e6f4ff' : '#1677ff' }}
            >
              📎 Xem tệp đính kèm
            </a>
          )}
        </div>
        <div
          style={{
            fontSize: 11,
            color: '#999',
            marginTop: 2,
            textAlign: isMe ? 'right' : 'left',
          }}
        >
          {time}
        </div>
      </div>
      {isMe && (
        <Avatar size={28} icon={avatarIcon} style={{ backgroundColor: avatarColor, flexShrink: 0 }} />
      )}
    </div>
  );
}

// ─── Chat Window Component ────────────────────────────────────
export default function ChatWindow({ conversationId, isClosed }) {
  const queryClient = useQueryClient();
  const [inputText, setInputText] = useState('');
  const [wsMessages, setWsMessages] = useState([]);
  const messagesEndRef = useRef(null);
  const messagesContainerRef = useRef(null);
  const wsRef = useRef(null);
  const reconnectTimerRef = useRef(null);
  const reconnectAttempts = useRef(0);
  const MAX_RECONNECT = 5;

  // ── Fetch message history ──
  const { data: messageData, isLoading } = useQuery({
    queryKey: ['chat-messages', conversationId],
    queryFn: () => fetchMessages(conversationId, 1, 50),
    enabled: !!conversationId,
  });

  const messages = messageData?.messages || [];

  // ── Merge API messages + WS live messages ──
  const allMessages = [
    ...messages,
    ...wsMessages.filter(
      (wsMsg) => !messages.some((m) => m.message_id === wsMsg.message_id)
    ),
  ];

  // ── Auto-scroll to bottom ──
  const scrollToBottom = useCallback(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [allMessages.length, scrollToBottom]);

  // ── Mark read ──
  useEffect(() => {
    if (conversationId) {
      markRead(conversationId).catch(() => {});
    }
  }, [conversationId, allMessages.length]);

  // ── WebSocket connection ──
  useEffect(() => {
    if (!conversationId || isClosed) return;

    const connectWS = () => {
      const url = getWSUrl(conversationId);
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        reconnectAttempts.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.message_id) {
            setWsMessages((prev) => {
              if (prev.some((m) => m.message_id === data.message_id)) return prev;
              return [...prev, data];
            });
            // Mark as read since user is viewing
            markRead(conversationId).catch(() => {});
          }
        } catch {
          // ignore non-JSON messages
        }
      };

      ws.onclose = () => {
        if (reconnectAttempts.current < MAX_RECONNECT) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 10000);
          reconnectTimerRef.current = setTimeout(() => {
            reconnectAttempts.current += 1;
            connectWS();
          }, delay);
        }
      };

      ws.onerror = () => {
        ws.close();
      };
    };

    connectWS();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
      setWsMessages([]);
    };
  }, [conversationId, isClosed]);

  // ── Send message mutation ──
  const sendMutation = useMutation({
    mutationFn: sendMessage,
    onSuccess: (res) => {
      setInputText('');
      if (res.data) {
        setWsMessages((prev) => {
          if (prev.some((m) => m.message_id === res.data.message_id)) return prev;
          return [...prev, res.data];
        });
      }
      queryClient.invalidateQueries({ queryKey: ['chat-rooms'] });
    },
  });

  const handleSend = () => {
    const text = inputText.trim();
    if (!text || isClosed) return;
    sendMutation.mutate({
      conversation_id: conversationId,
      type: 'text',
      text_content: text,
    });
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  // ── Empty state ──
  if (!conversationId) {
    return (
      <div
        style={{
          height: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Empty description="Chọn một phòng chat để bắt đầu" />
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* Messages area */}
      <div
        ref={messagesContainerRef}
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '16px 12px',
          background: '#fafafa',
          borderRadius: 8,
        }}
      >
        {isLoading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin indicator={<LoadingOutlined spin />} />
          </div>
        ) : allMessages.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
            Chưa có tin nhắn. Hãy bắt đầu cuộc trò chuyện!
          </div>
        ) : (
          <>
            {allMessages.map((msg, idx) => (
              <MessageBubble key={msg.message_id || `ws-${idx}`} msg={msg} isStaff />
            ))}
            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      {/* Input bar */}
      <div
        style={{
          padding: '12px 0 0',
          borderTop: '1px solid #f0f0f0',
        }}
      >
        {isClosed ? (
          <div
            style={{
              textAlign: 'center',
              padding: '12px',
              background: '#fff7e6',
              borderRadius: 8,
              color: '#ad6800',
            }}
          >
            🔒 Phòng chat đã đóng — không thể gửi tin nhắn
          </div>
        ) : (
          <Space.Compact style={{ width: '100%' }}>
            <Input.TextArea
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Nhập tin nhắn... (Enter để gửi)"
              autoSize={{ minRows: 1, maxRows: 3 }}
              style={{ borderRadius: '8px 0 0 8px' }}
              disabled={sendMutation.isPending}
            />
            <Tooltip title="Gửi (Enter)">
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={handleSend}
                loading={sendMutation.isPending}
                disabled={!inputText.trim()}
                style={{ height: 'auto', borderRadius: '0 8px 8px 0' }}
              />
            </Tooltip>
          </Space.Compact>
        )}
      </div>
    </div>
  );
}
