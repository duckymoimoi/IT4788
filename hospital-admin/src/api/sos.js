import api from './client';

export const fetchSOSList = () => api.get('/sos/get_list').then((r) => r.data.data);
export const fetchSOSDetail = (id) => api.get('/sos/get_detail', { params: { id } }).then((r) => r.data.data);
export const respondSOS = (data) => api.post('/sos/respond', data).then((r) => r.data);
export const resolveSOS = (data) => api.post('/sos/resolve', data).then((r) => r.data);
