package service

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"hospital/pkg/mapf"
	"hospital/repository"
	"hospital/schema"
)

// Errors
var (
	ErrMapNotFound   = errors.New("map not found")
	ErrNodeNotFound  = errors.New("poi not found")
	ErrEdgeNotFound  = errors.New("edge not found")
	ErrNodeCodeExist = errors.New("poi code already exists")
	ErrMissingField  = errors.New("missing required field")
	ErrCellNotFree   = errors.New("grid cell is not walkable")
)

// MapService xử lý logic nghiệp vụ cho Map module (15 API: #16-#30).
type MapService struct {
	repo *repository.MapRepo
}

func NewMapService(repo *repository.MapRepo) *MapService {
	return &MapService{repo: repo}
}

// ========================================
// RETURN TYPES
// ========================================

// MapItem là output cho mỗi bản đồ trong get_floors.
type MapItem struct {
	MapID       uint32  `json:"map_id"`
	MapName     string  `json:"map_name"`
	Rows        int     `json:"rows"`
	Cols        int     `json:"cols"`
	MapImageURL *string `json:"map_image_url"`
}

// MapMetaResult là output cho get_meta.
type MapMetaResult struct {
	MapID       uint32  `json:"map_id"`
	MapName     string  `json:"map_name"`
	Rows        int     `json:"rows"`
	Cols        int     `json:"cols"`
	GridData    string  `json:"grid_data"`
	MapImageURL *string `json:"map_image_url"`
}

// POIItem là output cho mỗi POI.
type POIItem struct {
	POIID                uint32  `json:"poi_id"`
	MapID                uint32  `json:"map_id"`
	WardID               *uint32 `json:"ward_id"`
	POICode              string  `json:"poi_code"`
	POIName              string  `json:"poi_name"`
	POIType              string  `json:"poi_type"`
	GridRow              int     `json:"grid_row"`
	GridCol              int     `json:"grid_col"`
	GridLocation         int     `json:"grid_location"`
	IsLandmark           bool    `json:"is_landmark"`
	IsAccessible         bool    `json:"is_accessible"`
	WheelchairAccessible bool    `json:"wheelchair_accessible"`
	CustomWeight         float32 `json:"custom_weight"`
	Capacity             *int    `json:"capacity"`
	Details              *string `json:"details"`
	OpenHours            *string `json:"open_hours"`
}

// SyncFullResult là output cho sync_full.
type SyncFullResult struct {
	Maps []MapItem `json:"maps"`
	POIs []POIItem `json:"pois"`
}

// ========================================
// INPUT TYPES  - Admin APIs
// ========================================

type AddNodeInput struct {
	MapID                uint32  `json:"map_id"`
	WardID               *uint32 `json:"ward_id"`
	POICode              string  `json:"poi_code"`
	POIName              string  `json:"poi_name"`
	POIType              string  `json:"poi_type"`
	GridRow              int     `json:"grid_row"`
	GridCol              int     `json:"grid_col"`
	IsLandmark           bool    `json:"is_landmark"`
	WheelchairAccessible bool    `json:"wheelchair_accessible"`
	Capacity             *int    `json:"capacity"`
	Details              *string `json:"details"`
	OpenHours            *string `json:"open_hours"`
}

type EditNodeInput struct {
	POICode              string   `json:"poi_code"` // Used as primary lookup
	GridRow              *int     `json:"grid_row"`
	GridCol              *int     `json:"grid_col"`
	POIName              *string  `json:"poi_name"`
	POIType              *string  `json:"poi_type"`
	IsLandmark           *bool    `json:"is_landmark"`
	WheelchairAccessible *bool    `json:"wheelchair_accessible"`
	IsAccessible         *bool    `json:"is_accessible"`
	Capacity             *int     `json:"capacity"`
	Details              *string  `json:"details"`
	OpenHours            *string  `json:"open_hours"`
	CustomWeight         *float32 `json:"custom_weight"`
}

type SetWeightInput struct {
	POIID  uint32  `json:"poi_id"`
	Weight float32 `json:"weight"`
}

// ========================================
// PRIVATE HELPERS
// ========================================

func poiToItem(p schema.GridPOI) POIItem {
	return POIItem{
		POIID:                p.POIID,
		MapID:                p.MapID,
		WardID:               p.WardID,
		POICode:              p.POICode,
		POIName:              p.POIName,
		POIType:              string(p.POIType),
		GridRow:              p.GridRow,
		GridCol:              p.GridCol,
		GridLocation:         p.GridLocation,
		IsLandmark:           p.IsLandmark,
		IsAccessible:         p.IsAccessible,
		WheelchairAccessible: p.WheelchairAccessible,
		CustomWeight:         p.CustomWeight,
		Capacity:             p.Capacity,
		Details:              p.Details,
		OpenHours:            p.OpenHours,
	}
}

func poisToItems(pois []schema.GridPOI) []POIItem {
	items := make([]POIItem, len(pois))
	for i, p := range pois {
		items[i] = poiToItem(p)
	}
	return items
}

func mapToItem(m schema.GridMap) MapItem {
	return MapItem{
		MapID:       m.MapID,
		MapName:     m.MapName,
		Rows:        m.Rows,
		Cols:        m.Cols,
		MapImageURL: m.MapImageURL,
	}
}

// ========================================
// READ APIs
// ========================================

// [16] GetFloors trả về danh sách bản đồ (thay thế floor).
func (s *MapService) GetFloors() ([]MapItem, error) {
	maps, err := s.repo.FindAllMaps()
	if err != nil {
		return nil, err
	}
	items := make([]MapItem, len(maps))
	for i, m := range maps {
		items[i] = mapToItem(m)
	}
	return items, nil
}

// [17] GetNodes trả về POIs của 1 map.
func (s *MapService) GetNodes(mapID uint32) ([]POIItem, error) {
	pois, err := s.repo.FindAllPOIs(mapID)
	if err != nil {
		return nil, err
	}
	return poisToItems(pois), nil
}

// EdgeItem 1 cạnh giữa 2 ô walkable liền kề.
type EdgeItem struct {
	FromRow      int `json:"from_row"`
	FromCol      int `json:"from_col"`
	FromLocation int `json:"from_location"`
	ToRow        int `json:"to_row"`
	ToCol        int `json:"to_col"`
	ToLocation   int `json:"to_location"`
}

// [18] GetEdges tính edges từ grid adjacency (4 hướng: N, S, E, W).
// Trả về danh sách cạnh giữa các ô walkable liền kề.
func (s *MapService) GetEdges(mapID uint32) (interface{}, error) {
	if mapID == 0 {
		return nil, ErrMissingField
	}

	m, err := s.repo.FindMapByID(mapID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMapNotFound
	}

	// Parse grid_data JSON: [[0,1,0],[1,0,0],...]
	var grid [][]int
	if err := json.Unmarshal([]byte(m.GridData), &grid); err != nil {
		return nil, errors.New("invalid grid_data format")
	}

	rows := len(grid)
	if rows == 0 {
		return map[string]interface{}{
			"map_id": mapID,
			"total":  0,
			"edges":  []EdgeItem{},
		}, nil
	}
	cols := len(grid[0])

	// 4-direction: N, S, E, W
	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, 1}, {0, -1}}

	edges := make([]EdgeItem, 0, rows*cols) // ước lượng

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if grid[r][c] != 0 {
				continue // obstacle
			}
			for _, d := range dirs {
				nr, nc := r+d[0], c+d[1]
				if nr < 0 || nr >= rows || nc < 0 || nc >= cols {
					continue
				}
				if grid[nr][nc] != 0 {
					continue
				}
				// Chỉ lưu 1 chiều (from < to) để tránh trùng
				fromLoc := r*cols + c
				toLoc := nr*cols + nc
				if fromLoc < toLoc {
					edges = append(edges, EdgeItem{
						FromRow: r, FromCol: c, FromLocation: fromLoc,
						ToRow: nr, ToCol: nc, ToLocation: toLoc,
					})
				}
			}
		}
	}

	return map[string]interface{}{
		"map_id": mapID,
		"total":  len(edges),
		"edges":  edges,
	}, nil
}

// [19] GetMeta trả về metadata bản đồ.
func (s *MapService) GetMeta(mapID uint32) (*MapMetaResult, error) {
	if mapID == 0 {
		return nil, ErrMissingField
	}
	m, err := s.repo.FindMapByID(mapID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMapNotFound
	}
	return &MapMetaResult{
		MapID:       m.MapID,
		MapName:     m.MapName,
		Rows:        m.Rows,
		Cols:        m.Cols,
		GridData:    m.GridData,
		MapImageURL: m.MapImageURL,
	}, nil
}

// [20] GetDepartments trả về wards kèm đếm số POI.
func (s *MapService) GetDepartments(nodeType string, wardID uint32) (interface{}, error) {
	if nodeType != "" || wardID > 0 {
		pois, err := s.repo.FindPOIsByType(schema.POIType(nodeType), 0)
		if err != nil {
			return nil, err
		}
		return poisToItems(pois), nil
	}
	return s.repo.CountPOIsByWard()
}

// [21] SearchLocation tìm kiếm POI theo keyword.
func (s *MapService) SearchLocation(keyword string, mapID uint32) ([]POIItem, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []POIItem{}, nil
	}
	pois, err := s.repo.SearchPOIs(keyword, mapID)
	if err != nil {
		return nil, err
	}
	return poisToItems(pois), nil
}

// [22] GetLandmarks trả về các điểm mốc.
func (s *MapService) GetLandmarks(mapID uint32) ([]POIItem, error) {
	pois, err := s.repo.FindLandmarks(mapID)
	if err != nil {
		return nil, err
	}
	return poisToItems(pois), nil
}

// [24] SyncFull trả về toàn bộ dữ liệu bản đồ.
func (s *MapService) SyncFull(mapID uint32) (*SyncFullResult, error) {
	syncData, err := s.repo.FindSyncData(mapID)
	if err != nil {
		return nil, err
	}

	mapItems := make([]MapItem, len(syncData.Maps))
	for i, m := range syncData.Maps {
		mapItems[i] = mapToItem(m)
	}

	return &SyncFullResult{
		Maps: mapItems,
		POIs: poisToItems(syncData.POIs),
	}, nil
}

// ========================================
// ADMIN APIs
// ========================================

// [25] AddNode thêm 1 POI mới.
func (s *MapService) AddNode(input AddNodeInput) (*POIItem, error) {
	code := strings.TrimSpace(input.POICode)
	name := strings.TrimSpace(input.POIName)
	if code == "" || name == "" || input.MapID == 0 {
		return nil, ErrMissingField
	}

	// Kiểm tra map tồn tại
	m, err := s.repo.FindMapByID(input.MapID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMapNotFound
	}
	if s.repo.IsSimulationRunning(input.MapID) {
		return nil, errors.New("cannot add node: simulation is currently running on this map")
	}

	// Kiểm tra trùng code
	existing, err := s.repo.FindPOIByCode(code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrNodeCodeExist
	}

	gridLocation := input.GridRow*m.Cols + input.GridCol

	poi := &schema.GridPOI{
		MapID:                input.MapID,
		WardID:               input.WardID,
		POICode:              code,
		POIName:              name,
		POIType:              schema.POIType(input.POIType),
		GridRow:              input.GridRow,
		GridCol:              input.GridCol,
		GridLocation:         gridLocation,
		IsLandmark:           input.IsLandmark,
		IsAccessible:         true,
		WheelchairAccessible: input.WheelchairAccessible,
		CustomWeight:         1.0,
		Capacity:             input.Capacity,
		Details:              input.Details,
		OpenHours:            input.OpenHours,
		IsActive:             true,
	}

	if err := s.repo.CreatePOI(poi); err != nil {
		return nil, err
	}

	item := poiToItem(*poi)
	return &item, nil
}

// [26] EditNode cập nhật POI.
func (s *MapService) EditNode(input EditNodeInput) (*POIItem, error) {
	if input.POICode == "" {
		return nil, ErrMissingField
	}

	poi, err := s.repo.FindPOIByCode(input.POICode)
	if err != nil {
		return nil, err
	}
	if poi == nil {
		return nil, ErrNodeNotFound
	}

	if s.repo.IsSimulationRunning(poi.MapID) {
		return nil, errors.New("cannot edit node: simulation is currently running on this map")
	}

	updates := map[string]interface{}{}

	if input.POIName != nil {
		updates["poi_name"] = strings.TrimSpace(*input.POIName)
	}
	if input.POIType != nil {
		updates["poi_type"] = *input.POIType
	}
	if input.GridRow != nil {
		updates["grid_row"] = *input.GridRow
		// Note: Location depends on map cols; assume logic handles update in repository/storage layer.
	}
	if input.GridCol != nil {
		updates["grid_col"] = *input.GridCol
	}
	if input.IsLandmark != nil {
		updates["is_landmark"] = *input.IsLandmark
	}
	if input.WheelchairAccessible != nil {
		updates["wheelchair_accessible"] = *input.WheelchairAccessible
	}
	if input.IsAccessible != nil {
		updates["is_accessible"] = *input.IsAccessible
	}
	if input.CustomWeight != nil {
		updates["custom_weight"] = *input.CustomWeight
	}
	if input.Capacity != nil {
		updates["capacity"] = *input.Capacity
	}
	if input.Details != nil {
		updates["details"] = *input.Details
	}
	if input.OpenHours != nil {
		updates["open_hours"] = *input.OpenHours
	}

	if len(updates) > 0 {
		if err := s.repo.UpdatePOI(poi.POIID, updates); err != nil {
			return nil, err
		}
		// Refresh
		updatedPoi, _ := s.repo.FindPOIByID(poi.POIID)
		if updatedPoi != nil {
			poi = updatedPoi
		}
	}

	item := poiToItem(*poi)
	return &item, nil
}

// [27] DelNode xóa (soft delete) POI.
func (s *MapService) DelNode(poiCode string) error {
	if poiCode == "" {
		return ErrMissingField
	}
	poi, err := s.repo.FindPOIByCode(poiCode)
	if err != nil {
		return err
	}
	if poi == nil {
		return ErrNodeNotFound
	}
	if s.repo.IsSimulationRunning(poi.MapID) {
		return errors.New("cannot delete node: simulation is currently running on this map")
	}
	return s.repo.DeactivatePOI(poi.POIID)
}

// MapExists kiểm tra map có tồn tại không (dùng trong handler validation).
func (s *MapService) MapExists(mapID uint32) (bool, error) {
	m, err := s.repo.FindMapByID(mapID)
	if err != nil {
		return false, err
	}
	return m != nil, nil
}

// [28] AddEdge thêm edge thủ công vào bảng map_steps.
func (s *MapService) AddEdge(mapID uint32, startNode, endNode string, distance float32) (uint32, error) {
	if mapID == 0 || startNode == "" || endNode == "" {
		return 0, ErrMissingField
	}
	if distance <= 0 {
		return 0, errors.New("distance must be positive")
	}

	// Kiểm tra map tồn tại
	m, err := s.repo.FindMapByID(mapID)
	if err != nil {
		return 0, err
	}
	if m == nil {
		return 0, ErrMapNotFound
	}

	edgeID, err := s.repo.CreateEdge(mapID, startNode, endNode, distance)
	if err != nil {
		return 0, err
	}
	return edgeID, nil
}

// [29] DelEdge xóa edge thủ công.
func (s *MapService) DelEdge(edgeID uint32) error {
	if edgeID == 0 {
		return ErrMissingField
	}
	exists, err := s.repo.EdgeExists(edgeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrEdgeNotFound
	}
	return s.repo.DeleteEdge(edgeID)
}

// [30] SetWeight cập nhật custom_weight của POI.
func (s *MapService) SetWeight(poiID uint32, weight float32) error {
	if poiID == 0 {
		return ErrMissingField
	}
	if weight <= 0 {
		return ErrMissingField
	}

	poi, err := s.repo.FindPOIByID(poiID)
	if err != nil {
		return err
	}
	if poi == nil {
		return ErrNodeNotFound
	}
	if s.repo.IsSimulationRunning(poi.MapID) {
		return errors.New("cannot set weight: simulation is currently running on this map")
	}

	return s.repo.UpdatePOI(poiID, map[string]interface{}{
		"custom_weight": weight,
	})
}

// ========================================
// MAP FILE APIs
// ========================================

// UploadMap luu thong tin map moi vao DB
func (s *MapService) UploadMap(mapName string, mapFilePath string, rows int, cols int, gridData string, mapImageURL *string) (*schema.GridMap, error) {
	if mapName == "" || mapFilePath == "" {
		return nil, ErrMissingField
	}
	rows, cols, gridData = normalizeMapFileData(mapFilePath, rows, cols, gridData)
	m := &schema.GridMap{
		MapName:     mapName,
		MapFilePath: mapFilePath,
		Rows:        rows,
		Cols:        cols,
		GridData:    gridData,
		MapImageURL: mapImageURL,
		IsActive:    false, // Mac dinh la false khi moi upload
	}
	if err := s.repo.CreateMap(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetMaps lay tat ca maps
func (s *MapService) GetMaps() ([]schema.GridMap, error) {
	maps, err := s.repo.GetAllMaps()
	if err != nil {
		return nil, err
	}
	for i := range maps {
		if maps[i].Rows > 0 && maps[i].Cols > 0 && maps[i].GridData != "" && maps[i].GridData != "[]" {
			continue
		}
		rows, cols, gridData := normalizeMapFileData(maps[i].MapFilePath, maps[i].Rows, maps[i].Cols, maps[i].GridData)
		if rows == maps[i].Rows && cols == maps[i].Cols && gridData == maps[i].GridData {
			continue
		}
		updates := map[string]interface{}{
			"rows":      rows,
			"cols":      cols,
			"grid_data": gridData,
		}
		if err := s.repo.UpdateMap(maps[i].MapID, updates); err == nil {
			maps[i].Rows = rows
			maps[i].Cols = cols
			maps[i].GridData = gridData
		}
	}
	return maps, nil
}

// SetActiveMap set map active va kiem tra simulation
func (s *MapService) SetActiveMap(mapID uint32) error {
	m, err := s.repo.FindMapByIDAnyStatus(mapID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrMapNotFound
	}
	// Check if any simulation is currently running
	// If simulation is running, we cannot change active map
	if s.repo.IsSimulationRunning(m.MapID) {
		return errors.New("cannot change active map: simulation is currently running. Please stop it first.")
	}
	return s.repo.SetActiveMap(mapID)
}

// EditMap cap nhat ten ban do.
func (s *MapService) EditMap(mapID uint32, mapName string) error {
	if mapID == 0 {
		return ErrMissingField
	}
	mapName = strings.TrimSpace(mapName)
	if mapName == "" {
		return ErrMissingField
	}
	m, err := s.repo.FindMapByIDAnyStatus(mapID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrMapNotFound
	}
	return s.repo.UpdateMapName(mapID, mapName)
}

// UpdateGrid cap nhat grid_data va ten cua ban do tu web editor.
func (s *MapService) UpdateGrid(mapID uint32, gridData string, mapName string) error {
	if mapID == 0 || gridData == "" {
		return ErrMissingField
	}
	m, err := s.repo.FindMapByIDAnyStatus(mapID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrMapNotFound
	}
	// Parse gridData to update rows/cols
	var grid [][]int
	if err := json.Unmarshal([]byte(gridData), &grid); err != nil {
		return ErrMissingField
	}
	rows := len(grid)
	cols := 0
	if rows > 0 {
		cols = len(grid[0])
	}
	updates := map[string]interface{}{
		"grid_data": gridData,
		"rows":      rows,
		"cols":      cols,
	}
	if mapName = strings.TrimSpace(mapName); mapName != "" {
		updates["map_name"] = mapName
	}
	return s.repo.UpdateMap(mapID, updates)
}

// UpdateMapFiles cap nhat file .map, anh, va grid_data cho map da ton tai.
func (s *MapService) UpdateMapFiles(mapID uint32, mapName string, mapFilePath string, rows int, cols int, gridData string, mapImageURL *string) error {
	if mapID == 0 {
		return ErrMissingField
	}
	m, err := s.repo.FindMapByIDAnyStatus(mapID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrMapNotFound
	}

	rows, cols, gridData = normalizeMapFileData(mapFilePath, rows, cols, gridData)

	updates := map[string]interface{}{
		"map_file_path": mapFilePath,
		"rows":          rows,
		"cols":          cols,
		"grid_data":     gridData,
	}
	if strings.TrimSpace(mapName) != "" {
		updates["map_name"] = strings.TrimSpace(mapName)
	}
	if mapImageURL != nil {
		updates["map_image_url"] = *mapImageURL
	}
	return s.repo.UpdateMap(mapID, updates)
}

func normalizeMapFileData(mapFilePath string, rows int, cols int, gridData string) (int, int, string) {
	if rows > 0 && cols > 0 && gridData != "" && gridData != "[]" {
		return rows, cols, gridData
	}
	if _, err := os.Stat(mapFilePath); err != nil {
		return rows, cols, gridData
	}
	grid, err := mapf.LoadGridMap(mapFilePath)
	if err != nil {
		return rows, cols, gridData
	}
	if rows <= 0 {
		rows = grid.Rows
	}
	if cols <= 0 {
		cols = grid.Cols
	}
	if gridData == "" || gridData == "[]" {
		gridData = grid.GridDataToJSON()
	}
	return rows, cols, gridData
}

// DeleteMap xóa map theo ID (hard delete). Không cho phép xóa map đang active.
func (s *MapService) DeleteMap(mapID uint32) error {
	if mapID == 0 {
		return ErrMissingField
	}
	m, err := s.repo.FindMapByIDAnyStatus(mapID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrMapNotFound
	}
	if m.IsActive {
		return errors.New("cannot delete active map; deactivate it first")
	}
	if s.repo.IsSimulationRunning(mapID) {
		return errors.New("cannot delete map: simulation is currently running")
	}
	return s.repo.DeleteMap(mapID)
}

// DeactivateMap set is_active = false cho map (không cần map thay thế).
func (s *MapService) DeactivateMap(mapID uint32) error {
	if mapID == 0 {
		return ErrMissingField
	}
	m, err := s.repo.FindMapByIDAnyStatus(mapID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrMapNotFound
	}
	if s.repo.IsSimulationRunning(mapID) {
		return errors.New("cannot deactivate map: simulation is currently running")
	}
	return s.repo.DeactivateMap(mapID)
}
