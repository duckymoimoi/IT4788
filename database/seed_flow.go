package database

import (
	"log"
	"time"

	"hospital/schema"

	"gorm.io/gorm"
)

// SeedFlow tao du lieu mau cho module Flow (Slice 5).
// Goi sau khi AutoMigrate da tao bang.
func SeedFlow(db *gorm.DB) {
	// Kiem tra da seed chua
	var count int64
	db.Model(&schema.UserPing{}).Count(&count)
	if count > 0 {
		log.Println("Flow data da co, bo qua seed")
		return
	}

	log.Println("Seed flow data...")

	now := time.Now()

	// --- User Pings: vi tri hien tai cua user ---
	// UserID 4,5 la benh nhan (tu seed.go)
	pings := []schema.UserPing{
		{UserID: 4, GridLocation: 100, GridRow: 2, GridCol: 10, CreatedAt: now.Add(-5 * time.Minute)},
		{UserID: 4, GridLocation: 105, GridRow: 2, GridCol: 15, CreatedAt: now.Add(-4 * time.Minute)},
		{UserID: 4, GridLocation: 110, GridRow: 2, GridCol: 20, CreatedAt: now.Add(-3 * time.Minute)},
		{UserID: 5, GridLocation: 200, GridRow: 4, GridCol: 10, CreatedAt: now.Add(-5 * time.Minute)},
		{UserID: 5, GridLocation: 205, GridRow: 4, GridCol: 15, CreatedAt: now.Add(-4 * time.Minute)},
		{UserID: 5, GridLocation: 210, GridRow: 4, GridCol: 20, CreatedAt: now.Add(-3 * time.Minute)},
		{UserID: 4, GridLocation: 100, GridRow: 2, GridCol: 10, CreatedAt: now.Add(-2 * time.Minute)},
		{UserID: 5, GridLocation: 100, GridRow: 2, GridCol: 10, CreatedAt: now.Add(-1 * time.Minute)},
		{UserID: 4, GridLocation: 300, GridRow: 6, GridCol: 10, CreatedAt: now},
		{UserID: 5, GridLocation: 300, GridRow: 6, GridCol: 10, CreatedAt: now},
	}
	if err := db.Create(&pings).Error; err != nil {
		log.Printf("Seed user_pings loi: %v", err)
	} else {
		log.Printf("Da tao %d user_pings", len(pings))
	}

	// --- Obstacle Reports ---
	obstacles := []schema.ObstacleReport{
		{
			UserID:       4,
			GridLocation: 150,
			ReportType:   "wet_floor",
			Description:  "San uot gan phong kham 101",
			Status:       schema.ObstacleStatusPending,
			CreatedAt:    now.Add(-10 * time.Minute),
		},
		{
			UserID:       5,
			GridLocation: 200,
			ReportType:   "construction",
			Description:  "Hanh lang dang sua chua",
			Status:       schema.ObstacleStatusResolved,
			CreatedAt:    now.Add(-1 * time.Hour),
		},
		{
			UserID:       4,
			GridLocation: 300,
			ReportType:   "elevator_broken",
			Description:  "Thang may T2 bi hong",
			Status:       schema.ObstacleStatusPending,
			CreatedAt:    now.Add(-5 * time.Minute),
		},
	}
	if err := db.Create(&obstacles).Error; err != nil {
		log.Printf("Seed obstacle_reports loi: %v", err)
	} else {
		log.Printf("Da tao %d obstacle_reports", len(obstacles))
	}

	// --- Heatmap Snapshots ---
	snapshots := []schema.HeatmapSnapshot{
		{GridLocation: 100, DensityLevel: 5, RecordedAt: now.Add(-30 * time.Minute)},
		{GridLocation: 100, DensityLevel: 8, RecordedAt: now.Add(-15 * time.Minute)},
		{GridLocation: 200, DensityLevel: 3, RecordedAt: now.Add(-30 * time.Minute)},
		{GridLocation: 200, DensityLevel: 4, RecordedAt: now.Add(-15 * time.Minute)},
		{GridLocation: 300, DensityLevel: 10, RecordedAt: now.Add(-30 * time.Minute)},
		{GridLocation: 300, DensityLevel: 12, RecordedAt: now.Add(-15 * time.Minute)},
		{GridLocation: 150, DensityLevel: 2, RecordedAt: now.Add(-15 * time.Minute)},
	}
	if err := db.Create(&snapshots).Error; err != nil {
		log.Printf("Seed heatmap_snapshots loi: %v", err)
	} else {
		log.Printf("Da tao %d heatmap_snapshots", len(snapshots))
	}

	// --- Priority Routes ---
	// StaffID 1 = admin, 2 = coordinator (tu seed.go)
	priorities := []schema.PriorityRoute{
		{
			SetBy:        2, // coordinator
			FromLocation: 100,
			ToLocation:   300,
			Reason:       "Don duong cap cuu tu tang 1 den phong mo",
			Status:       schema.PriorityStatusActive,
			ActivatedAt:  now.Add(-20 * time.Minute),
		},
		{
			SetBy:        1, // admin
			FromLocation: 200,
			ToLocation:   400,
			Reason:       "Uu tien benh nhan nang",
			Status:       schema.PriorityStatusExpired,
			ActivatedAt:  now.Add(-2 * time.Hour),
		},
	}
	if err := db.Create(&priorities).Error; err != nil {
		log.Printf("Seed priority_routes loi: %v", err)
	} else {
		log.Printf("Da tao %d priority_routes", len(priorities))
	}

	log.Println("Seed flow data hoan thanh")
}
