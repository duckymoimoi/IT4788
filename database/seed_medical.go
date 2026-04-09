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

	// 1. Seed danh sach hang doi (Queues - T24)
	queues := []schema.Queue{
		{PoiID: 10, CurrentNumber: 5, WaitingCount: 3, AvgWaitMinutes: 15, UpdatedAt: time.Now()},
		{PoiID: 11, CurrentNumber: 12, WaitingCount: 8, AvgWaitMinutes: 40, UpdatedAt: time.Now()},
	}

	for _, q := range queues {
		// Kiem tra ton tai truoc khi tao de tranh trung lap (Idempotency) 
		if err := db.Where("poi_id = ?", q.PoiID).FirstOrCreate(&q).Error; err != nil {
			log.Printf("CANH BAO: Seed queue cho POI %d loi: %v", q.PoiID, err)
		}
	}

	// 2. Seed cac chi dinh kham mau (Treatments - T22)
	// Su dung UserID = 4 (Pham Thi Benh Nhan) da duoc tao o Buoc 2 trong seed.go
	sampleTreatments := []schema.Treatment{
		{
			UserID: 4, 
			PoiID: 10, 
			WardID: 3, // Khoa Noi (WardID 3 trong seed.go) [cite: 304]
			TaskType: "examination", 
			TaskName: "Kham noi tong quat", 
			Status: "pending",
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