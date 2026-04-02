package service

import (
	"errors"
	"strings"

	"hospital/repository"
	"hospital/schema"
)

// Errors
var (
	ErrMapNotFound   = errors.New("map not found")
	ErrNodeNotFound  = errors.New("poi not found")
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
	POIID                uint32   `json:"poi_id"`
	POICode              *string  `json:"poi_code"`
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

// [18] GetEdges - grid-based: edges tính tự động từ adjacency.
// Trả về code 2003 "edges auto-computed from grid".
func (s *MapService) GetEdges(mapID uint32) (map[string]interface{}, error) {
	return map[string]interface{}{
		"message": "edges are auto-computed from grid adjacency",
		"map_id":  mapID,
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
	if input.POIID == 0 {
		return nil, ErrMissingField
	}

	poi, err := s.repo.FindPOIByID(input.POIID)
	if err != nil {
		return nil, err
	}
	if poi == nil {
		return nil, ErrNodeNotFound
	}

	updates := map[string]interface{}{}

	if input.POICode != nil {
		c := strings.TrimSpace(*input.POICode)
		if c != "" && c != poi.POICode {
			existing, err := s.repo.FindPOIByCode(c)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.POIID != poi.POIID {
				return nil, ErrNodeCodeExist
			}
			updates["poi_code"] = c
		}
	}
	if input.POIName != nil {
		updates["poi_name"] = strings.TrimSpace(*input.POIName)
	}
	if input.POIType != nil {
		updates["poi_type"] = *input.POIType
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
		if err := s.repo.UpdatePOI(input.POIID, updates); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.FindPOIByID(input.POIID)
	if err != nil {
		return nil, err
	}
	item := poiToItem(*updated)
	return &item, nil
}

// [27] DelNode xóa (soft delete) POI.
func (s *MapService) DelNode(poiID uint32) error {
	if poiID == 0 {
		return ErrMissingField
	}
	poi, err := s.repo.FindPOIByID(poiID)
	if err != nil {
		return err
	}
	if poi == nil {
		return ErrNodeNotFound
	}
	return s.repo.DeactivatePOI(poiID)
}

// [28] AddEdge - grid-based: không hỗ trợ thêm edge thủ công.
func (s *MapService) AddEdge() error {
	return nil // trả code 2003
}

// [29] DelEdge - grid-based: không hỗ trợ xóa edge thủ công.
func (s *MapService) DelEdge() error {
	return nil // trả code 2003
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

	return s.repo.UpdatePOI(poiID, map[string]interface{}{
		"custom_weight": weight,
	})
}
