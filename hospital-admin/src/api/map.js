import api from './client';
import { assessMapPoiCompleteness } from '../utils/mapExport';

// ─── Public — Map data (read-only) ───────────────────────────
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

export const fetchSyncFull = (map_id) =>
  api.get('/map/sync_full', { params: { map_id } }).then((r) => r.data.data);

// ─── Admin — Map management ──────────────────────────────────
export const fetchMaps = () =>
  api.get('/admin/get_maps').then((r) => r.data.data);

/**
 * Kiểm tra tình trạng mọi map đã có trong DB:
 * GET /admin/get_maps + GET /map/get_nodes?map_id= cho từng map.
 */
export async function fetchMapsPoiStatus() {
  const maps = await fetchMaps();
  if (!maps?.length) return [];

  return Promise.all(
    maps.map(async (m) => {
      let nodes = [];
      try {
        nodes = (await fetchNodes(m.map_id)) || [];
      } catch {
        nodes = [];
      }
      return {
        map_id: m.map_id,
        map_name: m.map_name,
        is_active: m.is_active,
        rows: m.rows,
        cols: m.cols,
        map_file_path: m.map_file_path,
        map_image_url: m.map_image_url,
        has_grid_data: !!(m.grid_data && m.grid_data !== '[]'),
        has_preview_image: !!m.map_image_url,
        ...assessMapPoiCompleteness(nodes),
      };
    }),
  );
}

export const setActiveMap = (map_id) =>
  api.post('/admin/set_active_map', { map_id }).then((r) => r.data);

export const uploadMap = (formData) =>
  api.post('/admin/upload_map', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000,
  }).then((r) => r.data);

export const uploadOutput = (formData) =>
  api.post('/admin/upload_output', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000,
  }).then((r) => r.data);

export const exportMap = (filename) =>
  api.get('/admin/export_map', {
    params: { filename },
    responseType: 'blob',
  });

export const deleteMap = (map_id) =>
  api.delete('/admin/delete_map', { data: { map_id } }).then((r) => r.data);

export const deactivateMap = (map_id) =>
  api.post('/admin/deactivate_map', { map_id }).then((r) => r.data);

// Admin — POI metadata (only metadata, NOT position)
export const editNode = (data) =>
  api.post('/admin/edit_node', data).then((r) => r.data);

// Uses edit_node under the hood — set_capacity route doesn't exist
export const setCapacity = ({ poi_id, poi_code, capacity }) =>
  api.post('/admin/edit_node', { id: poi_code, capacity }).then((r) => r.data);
