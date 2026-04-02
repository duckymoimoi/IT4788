package repository

import (
	"hospital/schema"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MapRepo xử lý truy vấn database cho Map module.
// Bao gồm 2 bảng: grid_maps, grid_pois.
type MapRepo struct {
	db *gorm.DB
}

func NewMapRepo(db *gorm.DB) *MapRepo {
	return &MapRepo{db: db}
}

// ========================================
// GRID MAPS
// ========================================

// FindAllMaps trả về tất cả bản đồ đang active.
func (r *MapRepo) FindAllMaps() ([]schema.GridMap, error) {
	var maps []schema.GridMap
	err := r.db.Where("is_active = ?", true).
		Order("map_id ASC").
		Find(&maps).Error
	return maps, err
}

// FindMapByID trả về 1 bản đồ theo ID.
func (r *MapRepo) FindMapByID(mapID uint32) (*schema.GridMap, error) {
	var m schema.GridMap
	err := r.db.First(&m, "map_id = ? AND is_active = ?", mapID, true).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &m, err
}

// CreateMap tạo bản đồ mới.
func (r *MapRepo) CreateMap(m *schema.GridMap) error {
	return r.db.Omit(clause.Associations).Create(m).Error
}

// ========================================
// GRID POIS  - CRUD
// ========================================

// FindAllPOIs trả về tất cả POI đang active của 1 map.
func (r *MapRepo) FindAllPOIs(mapID uint32) ([]schema.GridPOI, error) {
	var pois []schema.GridPOI
	q := r.db.Where("is_active = ?", true)
	if mapID > 0 {
		q = q.Where("map_id = ?", mapID)
	}
	err := q.Order("poi_id ASC").Find(&pois).Error
	return pois, err
}

// FindPOIByID trả về 1 POI theo ID.
func (r *MapRepo) FindPOIByID(poiID uint32) (*schema.GridPOI, error) {
	var poi schema.GridPOI
	err := r.db.First(&poi, "poi_id = ? AND is_active = ?", poiID, true).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &poi, err
}

// FindPOIByCode tìm POI theo mã code (unique).
func (r *MapRepo) FindPOIByCode(code string) (*schema.GridPOI, error) {
	var poi schema.GridPOI
	err := r.db.First(&poi, "poi_code = ?", code).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &poi, err
}

// SearchPOIs tìm kiếm POI theo keyword (case-insensitive LIKE).
func (r *MapRepo) SearchPOIs(keyword string, mapID uint32) ([]schema.GridPOI, error) {
	var pois []schema.GridPOI
	q := r.db.Where("is_active = ? AND poi_name LIKE ?", true, "%"+keyword+"%")
	if mapID > 0 {
		q = q.Where("map_id = ?", mapID)
	}
	err := q.Order("is_landmark DESC, poi_name ASC").Limit(30).Find(&pois).Error
	return pois, err
}

// FindLandmarks trả về các POI là mốc nổi bật.
func (r *MapRepo) FindLandmarks(mapID uint32) ([]schema.GridPOI, error) {
	var pois []schema.GridPOI
	q := r.db.Where("is_active = ? AND is_landmark = ?", true, true)
	if mapID > 0 {
		q = q.Where("map_id = ?", mapID)
	}
	err := q.Order("poi_name ASC").Find(&pois).Error
	return pois, err
}

// FindPOIsByType lọc POI theo loại.
func (r *MapRepo) FindPOIsByType(poiType schema.POIType, mapID uint32) ([]schema.GridPOI, error) {
	var pois []schema.GridPOI
	q := r.db.Where("is_active = ?", true)
	if poiType != "" {
		q = q.Where("poi_type = ?", poiType)
	}
	if mapID > 0 {
		q = q.Where("map_id = ?", mapID)
	}
	err := q.Order("poi_name ASC").Find(&pois).Error
	return pois, err
}

// CreatePOI tạo 1 POI mới.
func (r *MapRepo) CreatePOI(poi *schema.GridPOI) error {
	return r.db.Omit(clause.Associations).Create(poi).Error
}

// UpdatePOI cập nhật POI.
func (r *MapRepo) UpdatePOI(poiID uint32, updates map[string]interface{}) error {
	return r.db.Model(&schema.GridPOI{}).
		Where("poi_id = ?", poiID).
		Updates(updates).Error
}

// DeactivatePOI soft delete POI.
func (r *MapRepo) DeactivatePOI(poiID uint32) error {
	return r.db.Model(&schema.GridPOI{}).
		Where("poi_id = ?", poiID).
		Update("is_active", false).Error
}

// ========================================
// WARDS  - đếm POIs thuộc mỗi khoa
// ========================================

// WardPOICount kết quả đếm POI thuộc mỗi ward.
type WardPOICount struct {
	WardID   uint32 `json:"ward_id"`
	WardName string `json:"ward_name"`
	POICount int    `json:"poi_count"`
}

// CountPOIsByWard đếm số POI thuộc mỗi khoa.
func (r *MapRepo) CountPOIsByWard() ([]WardPOICount, error) {
	var results []WardPOICount
	err := r.db.Table("wards").
		Select("wards.ward_id, wards.ward_name, COUNT(grid_pois.poi_id) as poi_count").
		Joins("LEFT JOIN grid_pois ON grid_pois.ward_id = wards.ward_id AND grid_pois.is_active = ?", true).
		Where("wards.is_active = ?", true).
		Group("wards.ward_id").
		Order("wards.ward_name ASC").
		Scan(&results).Error
	return results, err
}

// ========================================
// SYNC  - API 24 (sync_full)
// ========================================

// SyncResult chứa toàn bộ dữ liệu bản đồ.
type SyncResult struct {
	Maps []schema.GridMap `json:"maps"`
	POIs []schema.GridPOI `json:"pois"`
}

// FindSyncData lấy tất cả dữ liệu bản đồ.
func (r *MapRepo) FindSyncData(mapID uint32) (*SyncResult, error) {
	var result SyncResult
	var err error

	if mapID > 0 {
		var m schema.GridMap
		err = r.db.First(&m, "map_id = ? AND is_active = ?", mapID, true).Error
		if err != nil {
			return nil, err
		}
		result.Maps = []schema.GridMap{m}
	} else {
		result.Maps, err = r.FindAllMaps()
		if err != nil {
			return nil, err
		}
	}

	result.POIs, err = r.FindAllPOIs(mapID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Ping kiem tra ket noi database.
func (r *MapRepo) Ping() bool {
	sqlDB, err := r.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

// UpdatePOIWeight cap nhat custom_weight cua POI.
func (r *MapRepo) UpdatePOIWeight(poiID uint32, weight float32) error {
	return r.db.Model(&schema.GridPOI{}).
		Where("poi_id = ?", poiID).
		Update("custom_weight", weight).Error
}
