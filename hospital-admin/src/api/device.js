import api from './client';

const first = (data) => Array.isArray(data) ? data[0] : data;

export const fetchStations = () => api.get('/asset/asset_stations').then((r) => r.data.data);
export const fetchDevices = () => api.get('/admin/get_devices').then((r) => r.data.data);
export const fetchWheelchairs = () => fetchDevices();
export const fetchDeviceStatus = (id) =>
  api.get('/asset/asset_health', { params: { asset_id: id } }).then((r) => first(r.data.data));
export const fetchDeviceTrack = (id) =>
  api.get('/asset/track_asset', { params: { asset_id: id } }).then((r) => first(r.data.data));
export const reportBroken = (data) => api.post('/asset/report_broken_asset', data).then((r) => r.data);
export const requestStaff = (data) => api.post('/staff/request_staff', data).then((r) => r.data);
export const addDevice = (data) => api.post('/admin/add_device', data).then((r) => r.data);
export const editDevice = (data) => api.post('/admin/edit_device', data).then((r) => r.data);
export const delDevice = (data) => api.delete('/admin/del_device', { data }).then((r) => r.data);
