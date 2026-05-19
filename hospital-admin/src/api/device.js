import api from './client';

export const fetchStations = () => api.get('/device/stations').then((r) => r.data.data);
export const fetchWheelchairs = () => api.get('/device/wheelchairs').then((r) => r.data.data);
export const fetchDeviceStatus = (id) => api.get(`/device/status/${id}`).then((r) => r.data.data);
export const fetchDeviceTrack = (id) => api.get(`/device/track/${id}`).then((r) => r.data.data);
