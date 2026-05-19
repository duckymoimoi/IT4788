import api from './client';

export const fetchSOSList = (page = 1, limit = 20) =>
  api.get('/sos/get_list', { params: { page, limit } }).then((r) => r.data.data);

export const fetchSOSDetail = (sos_id) =>
  api.get('/sos/get_detail', { params: { sos_id } }).then((r) => r.data.data);

export const respondSOS = (data) => api.post('/sos/respond', data).then((r) => r.data);

export const resolveSOS = (data) => api.post('/sos/resolve', data).then((r) => r.data);
