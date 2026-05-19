import api from './client';

export const fetchStations = () => api.get('/device/stations').then((r) => r.data.data);
export const fetchWheelchairs = () => api.get('/device/wheelchairs').then((r) => r.data.data);
export const fetchDeviceStatus = (id) => api.get(`/device/status/${id}`).then((r) => r.data.data);
export const fetchDeviceTrack = (id) => api.get(`/device/track/${id}`).then((r) => r.data.data);
export const reportBroken = (data) => api.post('/device/report_broken', data).then((r) => r.data);
export const requestStaff = (data) => api.post('/device/request_staff', data).then((r) => r.data);
export const addDevice = (data) => api.post('/admin/add_device', data).then((r) => r.data);
export const editDevice = (data) => api.patch('/admin/edit_device', data).then((r) => r.data);
export const delDevice = (data) => api.delete('/admin/del_device', { data }).then((r) => r.data);
