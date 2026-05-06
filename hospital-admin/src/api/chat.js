import api from './client';

export const fetchRooms = () => api.get('/chat/get_rooms').then((r) => r.data.data);
export const fetchMessages = (room_id) => api.get('/chat/get_messages', { params: { room_id } }).then((r) => r.data.data);
export const sendMessage = (data) => api.post('/chat/send_message', data).then((r) => r.data);
export const closeRoom = (room_id) => api.post('/chat/close_room', { room_id }).then((r) => r.data);
export const fetchUnreadCount = () => api.get('/chat/get_unread_count').then((r) => r.data.data);
