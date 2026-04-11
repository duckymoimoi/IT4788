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

	// 1. Tao cac Tram thiet bi (Gia su POI ID 1 va 2 da duoc tao boi Person A)
	stations := []schema.DeviceStation{
		{PoiID: 1, StationName: "Tram Sanh Chinh - Tang 1", Capacity: 10, IsActive: true},
		{PoiID: 2, StationName: "Tram Cap Cuu - TNGT", Capacity: 5, IsActive: true},
	}
	if err := db.Create(&stations).Error; err != nil {
		log.Printf("Error seeding stations: %v\n", err)
		return
	}

	// 2. Tao Thiet bi (3 Xe lan, 2 Cang)
	battery100 := 100
	devices := []schema.Device{
		{DeviceCode: "WL-001", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[0].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &battery100},
		{DeviceCode: "WL-002", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[0].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &battery100},
		{DeviceCode: "WL-003", DeviceType: schema.DeviceTypeWheelchair, StationID: &stations[1].StationID, Status: schema.DeviceStatusAvailable, BatteryLevel: &battery100},
		{DeviceCode: "ST-001", DeviceType: schema.DeviceTypeStretcher, StationID: &stations[1].StationID, Status: schema.DeviceStatusAvailable},
		{DeviceCode: "ST-002", DeviceType: schema.DeviceTypeStretcher, StationID: &stations[1].StationID, Status: schema.DeviceStatusAvailable},
	}

	if err := db.Create(&devices).Error; err != nil {
		log.Printf("Error seeding devices: %v\n", err)
	} else {
		log.Println("Seed Devices completed!")
	}
}