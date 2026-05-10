import api from './client';

export const fetchRooms = () =>
  api.get('/chat/get_rooms').then((r) => r.data.data);

export const fetchMessages = (conversation_id, page = 1, limit = 30) =>
  api.get('/chat/get_messages', { params: { conversation_id, page, limit } }).then((r) => r.data.data);

export const sendMessage = (data) =>
  api.post('/chat/send_message', data).then((r) => r.data);

export const closeRoom = (conversation_id) =>
  api.post('/chat/close_room', { conversation_id }).then((r) => r.data);

export const fetchUnreadCount = (conversation_id) =>
  api.get('/chat/get_unread_count', { params: { conversation_id } }).then((r) => r.data.data);

export const markRead = (conversation_id) =>
  api.post('/chat/mark_read', { conversation_id }).then((r) => r.data);

export const getWSUrl = (conversationId) => {
  const token = localStorage.getItem('token');
  const base = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/api/ws';
  return `${base}/chat?conversation_id=${conversationId}&token=${token}`;
};
