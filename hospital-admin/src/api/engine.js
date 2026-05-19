import api from './client';

export const fetchEngineHealth = () =>
  api.get('/engine/health').then((r) => r.data.data);

export const solve = (data) =>
  api.post('/engine/solve', data).then((r) => r.data.data);

export const setParams = (data) =>
  api.post('/engine/set_params', data).then((r) => r.data);

export const fetchConvergence = () =>
  api.get('/engine/convergence').then((r) => r.data.data);

export const clearCache = () =>
  api.post('/engine/clear_cache').then((r) => r.data);

export const loadMapf = (file_path) =>
  api.post('/engine/load_mapf', { file_path }).then((r) => r.data);

export const fetchMapfPositions = (timestep = 0) =>
  api.get('/engine/mapf_positions', { params: { timestep } }).then((r) => r.data.data);

export const fetchMapfInfo = () =>
  api.get('/engine/mapf_info').then((r) => r.data.data);
