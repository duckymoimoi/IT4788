import api from './client';

export const fetchRooms = () =>
  api.get('/chat/get_rooms').then((r) => r.data.data);

export const fetchChatParticipants = () =>
  api.get('/chat/participants').then((r) => r.data.data);

export const createRoom = (data) =>
  api.post('/chat/create_room', data).then((r) => r.data);

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
  const configuredBase = import.meta.env.VITE_WS_URL;
  const apiBase = import.meta.env.VITE_API_BASE_URL || window.location.origin + '/api';
  const base = configuredBase || apiBase
    .replace(/^https:/, 'wss:')
    .replace(/^http:/, 'ws:')
    .replace(/\/api\/?$/, '/api/ws');
  return `${base}/chat?conversation_id=${conversationId}&token=${token}`;
};
