package database

import (
	"fmt"
	"log"

	"hospital/pkg/mapf"
	"hospital/schema"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedMap seeds bản đồ grid 2D từ file .map (MovingAI format).
// Parse file -> INSERT grid_maps + seed POIs mẫu.
func SeedMap(db *gorm.DB) error {
	// Kiểm tra đã có dữ liệu chưa
	var count int64
	db.Model(&schema.GridMap{}).Count(&count)
	if count > 0 {
		log.Println("seed_map: grid_maps da co du lieu, bo qua")
		return nil
	}

	// Parse file .map
	mapFilePath := "data/warehouse_small.map"
	grid, err := mapf.LoadGridMap(mapFilePath)
	if err != nil {
		return fmt.Errorf("seed_map: parse map file loi: %w", err)
	}

	log.Printf("seed_map: Parsed map %s: %d rows x %d cols, %d walkable cells",
		mapFilePath, grid.Rows, grid.Cols, grid.CountWalkable())

	// Tạo JSON grid_data (compact)
	gridData := grid.GridDataToJSON()

	// INSERT grid_maps
	gridMap := &schema.GridMap{
		MapName:     "Hospital Main Floor",
		MapFilePath: mapFilePath,
		Rows:        grid.Rows,
		Cols:        grid.Cols,
		GridData:    gridData,
		IsActive:    true,
	}

	if err := db.Omit(clause.Associations).Create(gridMap).Error; err != nil {
		return fmt.Errorf("seed_map: insert grid_maps loi: %w", err)
	}

	log.Printf("seed_map: Inserted grid_map ID=%d (%s)", gridMap.MapID, gridMap.MapName)

	// Seed POIs mẫu
	if err := seedSamplePOIs(db, gridMap, grid); err != nil {
		return err
	}

	return nil
}

// seedSamplePOIs tạo 10+ POI mẫu trên grid.
// Chọn các ô walkable để gán POI.
func seedSamplePOIs(db *gorm.DB, gridMap *schema.GridMap, grid *mapf.GridMap) error {
	type poiSeed struct {
		Code     string
		Name     string
		Type     schema.POIType
		Row      int
		Col      int
		Landmark bool
	}

	// Tìm 1 vài ô walkable cố định để seed
	// (dùng tọa độ cụ thể từ warehouse_small.map 33x57)
	seeds := []poiSeed{
		{"ENT-01", "Cổng chính", schema.POITypeEntrance, 4, 4, true},
		{"ENT-02", "Cổng phụ", schema.POITypeEntrance, 32, 52, true},
		{"RM-101", "Phòng khám Nội khoa", schema.POITypeRoom, 4, 8, true},
		{"RM-102", "Phòng khám Ngoại khoa", schema.POITypeRoom, 4, 12, false},
		{"RM-103", "Phòng Xét nghiệm", schema.POITypeRoom, 4, 16, false},
		{"RM-104", "Phòng X-Quang", schema.POITypeRoom, 4, 20, false},
		{"RM-105", "Phòng Siêu âm", schema.POITypeRoom, 4, 24, false},
		{"PH-01", "Nhà thuốc", schema.POITypePharmacy, 4, 28, true},
		{"WC-01", "WC Tầng 1", schema.POITypeWC, 4, 32, false},
		{"CAN-01", "Canteen Bệnh viện", schema.POITypeCanteen, 4, 36, true},
		{"INFO-01", "Bàn thông tin", schema.POITypeInfo, 4, 40, true},
		{"WIFI-01", "Wifi Lobby", schema.POITypeWifi, 4, 44, false},
		{"COR-01", "Hành lang chính", schema.POITypeCorridor, 4, 48, false},
	}

	for _, s := range seeds {
		// Kiểm tra ô walkable
		if !grid.IsWalkable(s.Row, s.Col) {
			log.Printf("seed_map: WARNING - POI %s at (%d,%d) is NOT walkable, skipping", s.Code, s.Row, s.Col)
			continue
		}

		poi := &schema.GridPOI{
			MapID:        gridMap.MapID,
			POICode:      s.Code,
			POIName:      s.Name,
			POIType:      s.Type,
			GridRow:      s.Row,
			GridCol:      s.Col,
			GridLocation: s.Row*gridMap.Cols + s.Col,
			IsLandmark:   s.Landmark,
			IsAccessible: true,
			CustomWeight: 1.0,
			IsActive:     true,
		}

		if err := db.Omit(clause.Associations).Create(poi).Error; err != nil {
			return fmt.Errorf("seed_map: insert POI %s loi: %w", s.Code, err)
		}
	}

	var poiCount int64
	db.Model(&schema.GridPOI{}).Count(&poiCount)
	log.Printf("seed_map: Seeded %d POIs", poiCount)

	return nil
}
