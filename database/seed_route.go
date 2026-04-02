package database

import (
	"log"

	"hospital/schema"

	"gorm.io/gorm"
)

// SeedRoute tao du lieu mau cho bang travel_modes.
// Goi sau khi AutoMigrate da tao bang.
func SeedRoute(db *gorm.DB) {
	modes := []schema.TravelMode{
		{ModeID: "walking", ModeName: "Đi bộ", SpeedFactor: 1.0},
		{ModeID: "wheelchair", ModeName: "Xe lăn", SpeedFactor: 0.7},
		{ModeID: "stretcher", ModeName: "Cáng", SpeedFactor: 0.5},
		{ModeID: "hospital_cart", ModeName: "Xe đẩy bệnh viện", SpeedFactor: 1.5},
	}

	for _, mode := range modes {
		result := db.Where("mode_id = ?", mode.ModeID).FirstOrCreate(&mode)
		if result.Error != nil {
			log.Printf("Seed travel_mode '%s' loi: %v", mode.ModeID, result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("Seed travel_mode '%s'  - %s (speed=%.1f)", mode.ModeID, mode.ModeName, mode.SpeedFactor)
		}
	}
}
