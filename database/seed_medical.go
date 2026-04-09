package database

import (
	"log"
	"time"

	"hospital/schema"

	"gorm.io/gorm"
)

// SeedMedical tao du lieu mau cho module Y te (Slice 6).
// Ham nay duoc goi tu Seed() trong seed.go
func SeedMedical(db *gorm.DB) {
	log.Println("Bat dau seed du lieu y te...")

	// 0. Dam bao co POI type "room" de SyncHIS hoat dong
	// Tim POI type "room" co san tu SeedMap
	var roomPOIs []schema.GridPOI
	db.Where("poi_type = ? AND is_active = ?", "room", true).Find(&roomPOIs)

	// Neu chua co, tao 2 phong kham mau
	if len(roomPOIs) == 0 {
		// Tim map_id dau tien
		var mapID uint32 = 1
		var gridMap schema.GridMap
		if db.First(&gridMap).Error == nil {
			mapID = gridMap.MapID
		}

		newPOIs := []schema.GridPOI{
			{
				MapID: mapID, POICode: "MED001", POIName: "Phong Kham Noi",
				POIType: schema.POITypeRoom, GridRow: 5, GridCol: 5,
				GridLocation: 5*57 + 5, IsActive: true, CustomWeight: 1.0,
			},
			{
				MapID: mapID, POICode: "MED002", POIName: "Phong Kham Ngoai",
				POIType: schema.POITypeRoom, GridRow: 6, GridCol: 6,
				GridLocation: 6*57 + 6, IsActive: true, CustomWeight: 1.0,
			},
		}
		for i := range newPOIs {
			db.Where("poi_code = ?", newPOIs[i].POICode).FirstOrCreate(&newPOIs[i])
		}
		// Reload
		db.Where("poi_type = ? AND is_active = ?", "room", true).Find(&roomPOIs)
		log.Printf("Da tao %d POI phong kham mau", len(newPOIs))
	}

	// Lay POI IDs thuc te tu database (khong hardcode)
	var poiID1, poiID2 uint32
	if len(roomPOIs) >= 1 {
		poiID1 = roomPOIs[0].POIID
	}
	if len(roomPOIs) >= 2 {
		poiID2 = roomPOIs[1].POIID
	} else {
		poiID2 = poiID1
	}

	// 1. Seed danh sach hang doi (Queues - T24)
	queues := []schema.Queue{
		{PoiID: poiID1, CurrentNumber: 5, WaitingCount: 3, AvgWaitMinutes: 15, UpdatedAt: time.Now()},
		{PoiID: poiID2, CurrentNumber: 12, WaitingCount: 8, AvgWaitMinutes: 40, UpdatedAt: time.Now()},
	}

	for _, q := range queues {
		if err := db.Where("poi_id = ?", q.PoiID).FirstOrCreate(&q).Error; err != nil {
			log.Printf("CANH BAO: Seed queue cho POI %d loi: %v", q.PoiID, err)
		}
	}

	// 2. Seed cac chi dinh kham mau (Treatments - T22)
	// Su dung UserID = 4 (Pham Thi Benh Nhan) da duoc tao o Buoc 2 trong seed.go
	sampleTreatments := []schema.Treatment{
		{
			UserID:   4,
			PoiID:    poiID1,
			WardID:   3, // Khoa Noi (WardID 3 trong seed.go)
			TaskType: schema.TaskTypeExam,
			TaskName: "Kham noi tong quat",
			Status:   schema.TaskStatusPending,
		},
	}

	for _, t := range sampleTreatments {
		var count int64
		db.Model(&schema.Treatment{}).Where("user_id = ? AND task_name = ?", t.UserID, t.TaskName).Count(&count)
		if count == 0 {
			db.Create(&t)
		}
	}

	log.Println("Seed du lieu y te hoan thanh")
}