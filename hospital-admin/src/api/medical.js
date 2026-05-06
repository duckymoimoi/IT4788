import api from './client';

export const fetchTasks = () => api.get('/medical/get_tasks').then((r) => r.data.data);
export const fetchQueue = (poi_id) => api.get('/medical/get_queue', { params: { poi_id } }).then((r) => r.data.data);
export const fetchResultStatus = (treatment_id) => api.get('/medical/result_status', { params: { treatment_id } }).then((r) => r.data.data);
export const fetchPrescription = () => api.get('/medical/get_prescription').then((r) => r.data.data);
export const fetchRoomOpen = (poi_id) => api.get('/medical/room_open', { params: { poi_id } }).then((r) => r.data.data);
export const fetchHistory = () => api.get('/medical/get_history').then((r) => r.data.data);
export const syncNow = () => api.post('/medical/sync_now').then((r) => r.data);
