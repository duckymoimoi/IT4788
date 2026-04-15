package database

import (
	"hospital/schema"
	"log"

	"gorm.io/gorm"
)

// SeedDevices khoi tao du lieu mau cho module Device
func SeedDevices(db *gorm.DB) {
	var count int64
	db.Model(&schema.DeviceStation{}).Count(&count)
	if count > 0 {
		log.Println("Device data already seeded. Skipping...")
		return
	}

	log.Println("Seeding Device Stations and Devices...")

	// 4 Tram thiet bi
	stations := []schema.DeviceStation{
		{PoiID: 1, StationName: "Trạm Sảnh Chính - Tầng 1", Capacity: 15, IsActive: true},
		{PoiID: 2, StationName: "Trạm Cấp Cứu - TNGT", Capacity: 8, IsActive: true},
		{PoiID: 3, StationName: "Trạm Khoa Nội - Tầng 2", Capacity: 10, IsActive: true},
		{PoiID: 4, StationName: "Trạm Khoa Ngoại - Tầng 3", Capacity: 6, IsActive: true},
	}
	if err := db.Create(&stations).Error; err != nil {
		log.Printf("Error seeding stations: %v\n", err)
		return
	}

	// 15 Thiet bi (10 xe lan, 5 cang)
	bat100 := 100
	bat85 := 85
	bat60 := 60
	bat40 := 40
	devices := []schema.Device{
		// Xe lan - Tram 1 (Sanh chinh)
		{DeviceCode: "WL-001", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[0].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat100},
		{DeviceCode: "WL-002", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[0].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat100},
		{DeviceCode: "WL-003", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[0].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat85},
		{DeviceCode: "WL-004", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[0].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat60},
		// Xe lan - Tram 2 (Cap cuu)
		{DeviceCode: "WL-005", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[1].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat100},
		{DeviceCode: "WL-006", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[1].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat85},
		// Xe lan - Tram 3 (Khoa Noi)
		{DeviceCode: "WL-007", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[2].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat100},
		{DeviceCode: "WL-008", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[2].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat40},
		// Xe lan - Tram 4 (Khoa Ngoai)
		{DeviceCode: "WL-009", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[3].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat100},
		{DeviceCode: "WL-010", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[3].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &bat85},
		// Cang - phan bo deu
		{DeviceCode: "ST-001", DeviceType: schema.DeviceTypeStretcher, StationID: &stations[0].StationID, Status: schema.DeviceStatusAvailable},
		{DeviceCode: "ST-002", DeviceType: schema.DeviceTypeStretcher, StationID: &stations[1].StationID, Status: schema.DeviceStatusAvailable},
		{DeviceCode: "ST-003", DeviceType: schema.DeviceTypeStretcher, StationID: &stations[1].StationID, Status: schema.DeviceStatusAvailable},
		{DeviceCode: "ST-004", DeviceType: schema.DeviceTypeStretcher, StationID: &stations[2].StationID, Status: schema.DeviceStatusAvailable},
		{DeviceCode: "ST-005", DeviceType: schema.DeviceTypeStretcher, StationID: &stations[3].StationID, Status: schema.DeviceStatusAvailable},
	}

	if err := db.Create(&devices).Error; err != nil {
		log.Printf("Error seeding devices: %v\n", err)
	} else {
		log.Printf("Da tao %d tram, %d thiet bi", len(stations), len(devices))
	}
}