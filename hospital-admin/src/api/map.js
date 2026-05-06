import api from './client';

// Public — Map data (read-only)
export const fetchFloors = () =>
  api.get('/map/get_floors').then((r) => r.data.data);

export const fetchNodes = (map_id) =>
  api.get('/map/get_nodes', { params: { map_id } }).then((r) => r.data.data);

export const fetchEdges = (map_id) =>
  api.get('/map/get_edges', { params: { map_id } }).then((r) => r.data.data);

export const fetchMeta = (map_id) =>
  api.get('/map/get_meta', { params: { map_id } }).then((r) => r.data.data);

export const fetchDepts = () =>
  api.get('/map/get_depts').then((r) => r.data.data);

export const searchLocation = (keyword, map_id) =>
  api.get('/map/search_location', { params: { keyword, map_id } }).then((r) => r.data.data);

export const fetchLandmarks = (map_id) =>
  api.get('/map/get_landmarks', { params: { map_id } }).then((r) => r.data.data);

// Admin — chỉ sửa metadata POI, không thêm/xóa node, không sửa weight
export const editNode = (data) =>
  api.post('/admin/edit_node', data).then((r) => r.data);

export const setCapacity = (data) =>
  api.patch('/admin/set_capacity', data).then((r) => r.data);
