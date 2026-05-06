import api from './client';

export const fetchHeatmap = () =>
  api.get('/flow/get_heatmap').then((r) => r.data.data);

export const fetchDensity = (gridLocation) =>
  api.get('/flow/get_density', { params: { grid_location: gridLocation } }).then((r) => r.data.data);

export const fetchBottlenecks = (limit = 10) =>
  api.get('/flow/get_bottlenecks', { params: { limit } }).then((r) => r.data.data);

export const fetchForecast = (hours = 24) =>
  api.get('/flow/get_forecast', { params: { hours } }).then((r) => r.data.data);

export const fetchAlerts = () =>
  api.get('/flow/get_alerts').then((r) => r.data.data);

export const fetchObstacles = (status, page = 1, limit = 20) =>
  api.get('/flow/get_obstacles', { params: { status, page, limit } }).then((r) => r.data.data);

export const resolveObstacle = ({ report_id, action }) =>
  api.post('/flow/resolve_obstacle', { report_id, action }).then((r) => r.data);

export const setPriority = (data) =>
  api.post('/flow/set_priority', data).then((r) => r.data);

export const expirePriority = (priority_id) =>
  api.post('/flow/expire_priority', { priority_id }).then((r) => r.data);

export const fetchStatsFlow = (hours = 24) =>
  api.get('/admin/stats_flow', { params: { hours } }).then((r) => r.data.data);

export const resetFlow = () =>
  api.post('/admin/reset_flow').then((r) => r.data);

// Simulation
export const startSimulation = (data) =>
  api.post('/simulate/start', data).then((r) => r.data.data);

export const stopSimulation = () =>
  api.post('/simulate/stop').then((r) => r.data);

export const fetchSimStatus = () =>
  api.get('/simulate/status').then((r) => r.data.data);
