package database

import (
	"log"
	"time"

	"hospital/schema"

	"gorm.io/gorm"
)

// SeedMedical tao du lieu mau cho module Y te (Slice 6).
func SeedMedical(db *gorm.DB) {
	log.Println("Bat dau seed du lieu y te...")

	// 0. Dam bao co POI type "room" de SyncHIS hoat dong
	var roomPOIs []schema.GridPOI
	db.Where("poi_type = ? AND is_active = ?", "room", true).Find(&roomPOIs)

	if len(roomPOIs) == 0 {
		var mapID uint32 = 1
		var gridMap schema.GridMap
		if db.First(&gridMap).Error == nil {
			mapID = gridMap.MapID
		}
		newPOIs := []schema.GridPOI{
			{MapID: mapID, POICode: "MED001", POIName: "Phong Kham Noi", POIType: schema.POITypeRoom, GridRow: 5, GridCol: 5, GridLocation: 5*57 + 5, IsActive: true, CustomWeight: 1.0},
			{MapID: mapID, POICode: "MED002", POIName: "Phong Kham Ngoai", POIType: schema.POITypeRoom, GridRow: 6, GridCol: 6, GridLocation: 6*57 + 6, IsActive: true, CustomWeight: 1.0},
		}
		for i := range newPOIs {
			db.Where("poi_code = ?", newPOIs[i].POICode).FirstOrCreate(&newPOIs[i])
		}
		db.Where("poi_type = ? AND is_active = ?", "room", true).Find(&roomPOIs)
		log.Printf("Da tao %d POI phong kham mau", len(newPOIs))
	}

	var poiID1, poiID2 uint32
	if len(roomPOIs) >= 1 {
		poiID1 = roomPOIs[0].POIID
	}
	if len(roomPOIs) >= 2 {
		poiID2 = roomPOIs[1].POIID
	} else {
		poiID2 = poiID1
	}

	// 1. Seed hang doi
	queues := []schema.Queue{
		{PoiID: poiID1, CurrentNumber: 15, WaitingCount: 8, AvgWaitMinutes: 20, UpdatedAt: time.Now()},
		{PoiID: poiID2, CurrentNumber: 22, WaitingCount: 12, AvgWaitMinutes: 35, UpdatedAt: time.Now()},
	}
	for _, q := range queues {
		if err := db.Where("poi_id = ?", q.PoiID).FirstOrCreate(&q).Error; err != nil {
			log.Printf("CANH BAO: Seed queue cho POI %d loi: %v", q.PoiID, err)
		}
	}

	// 2. Seed treatments cho nhieu benh nhan
	var treatmentCount int64
	db.Model(&schema.Treatment{}).Count(&treatmentCount)
	if treatmentCount == 0 {
		treatments := []schema.Treatment{
			// Benh nhan 4 (Pham Thi Benh Nhan) - 3 chi dinh
			{UserID: 4, PoiID: poiID1, WardID: 3, TaskType: schema.TaskTypeExam, TaskName: "Kham noi tong quat", Status: schema.TaskStatusPending},
			{UserID: 4, PoiID: poiID2, WardID: 4, TaskType: schema.TaskTypeLab, TaskName: "Xet nghiem mau", Status: schema.TaskStatusPending},
			{UserID: 4, PoiID: poiID1, WardID: 3, TaskType: schema.TaskTypeExam, TaskName: "Do huyet ap", Status: schema.TaskStatusCompleted},
			// Benh nhan 5 (Hoang Van Test) - 2 chi dinh
			{UserID: 5, PoiID: poiID2, WardID: 4, TaskType: schema.TaskTypeExam, TaskName: "Kham ngoai khoa", Status: schema.TaskStatusPending},
			{UserID: 5, PoiID: poiID1, WardID: 1, TaskType: schema.TaskTypeLab, TaskName: "Xet nghiem nuoc tieu", Status: schema.TaskStatusPending},
			// Benh nhan 6 (Dao Minh Tuan) - 2 chi dinh
			{UserID: 6, PoiID: poiID1, WardID: 3, TaskType: schema.TaskTypeExam, TaskName: "Kham tim mach", Status: schema.TaskStatusPending},
			{UserID: 6, PoiID: poiID2, WardID: 2, TaskType: schema.TaskTypeImaging, TaskName: "Chup X-Quang nguc", Status: schema.TaskStatusPending},
			// Benh nhan 7 (Vu Thi Lan)
			{UserID: 7, PoiID: poiID1, WardID: 3, TaskType: schema.TaskTypeExam, TaskName: "Kham san phu khoa", Status: schema.TaskStatusInProgress},
			// Benh nhan 8 (Bui Duc Manh) - nguoi gia
			{UserID: 8, PoiID: poiID1, WardID: 3, TaskType: schema.TaskTypeExam, TaskName: "Kham noi tiet", Status: schema.TaskStatusPending},
			{UserID: 8, PoiID: poiID2, WardID: 2, TaskType: schema.TaskTypeImaging, TaskName: "Sieu am bung", Status: schema.TaskStatusPending},
			{UserID: 8, PoiID: poiID1, WardID: 1, TaskType: schema.TaskTypeLab, TaskName: "Xet nghiem duong huyet", Status: schema.TaskStatusPending},
		}
		db.Create(&treatments)
		log.Printf("Da tao %d treatments", len(treatments))

		// 3. Seed don thuoc mau
		prescriptions := []schema.Prescription{
			{UserID: 4, IssuedBy: 2, ItemsJSON: `[{"name":"Paracetamol 500mg","dosage":"2 vien/lan, 3 lan/ngay","qty":30,"note":"Uong sau an"},{"name":"Vitamin C 1000mg","dosage":"1 vien/ngay","qty":30,"note":"Uong buoi sang"}]`, Status: schema.PrescriptionPending},
			{UserID: 7, IssuedBy: 2, ItemsJSON: `[{"name":"Acid folic 5mg","dosage":"1 vien/ngay","qty":90,"note":"Uong truoc an sang"}]`, Status: schema.PrescriptionPending},
			{UserID: 8, IssuedBy: 3, ItemsJSON: `[{"name":"Metformin 500mg","dosage":"1 vien/ngay","qty":60,"note":"Uong sau an sang"},{"name":"Glimepiride 2mg","dosage":"1 vien/ngay","qty":30,"note":"Uong truoc an"}]`, Status: schema.PrescriptionPending},
		}
		db.Create(&prescriptions)
		log.Printf("Da tao %d prescriptions", len(prescriptions))
	}

	// 4. Seed notifications cho nhieu user
	var notifCount int64
	db.Model(&schema.Notification{}).Count(&notifCount)
	if notifCount == 0 {
		now := time.Now()
		notifications := []schema.Notification{
			// Benh nhan 4
			{UserID: 4, Title: "Lịch khám hôm nay", Content: "Bạn có lịch khám Nội tổng quát lúc 9:00 sáng tại Phòng 101", NotifType: "reminder", IsRead: false, CreatedAt: now.Add(-2 * time.Hour)},
			{UserID: 4, Title: "Kết quả xét nghiệm", Content: "Kết quả xét nghiệm máu của bạn đã có. Vui lòng liên hệ bác sĩ để nhận kết quả.", NotifType: "result", IsRead: false, CreatedAt: now.Add(-1 * time.Hour)},
			{UserID: 4, Title: "Nhắc nhở uống thuốc", Content: "Đã đến giờ uống Paracetamol 500mg (2 viên). Uống sau ăn.", NotifType: "medicine", IsRead: true, CreatedAt: now.Add(-30 * time.Minute)},
			// Benh nhan 5
			{UserID: 5, Title: "Đặt lịch thành công", Content: "Bạn đã đặt lịch khám Ngoại khoa ngày mai lúc 8:30.", NotifType: "booking", IsRead: false, CreatedAt: now.Add(-3 * time.Hour)},
			{UserID: 5, Title: "Hàng đợi cập nhật", Content: "Phòng khám Ngoại: còn 3 người trước bạn (khoảng 15 phút).", NotifType: "queue", IsRead: false, CreatedAt: now.Add(-10 * time.Minute)},
			// Benh nhan 6
			{UserID: 6, Title: "Chào mừng!", Content: "Chào mừng bạn đến với ứng dụng Bệnh viện Thông minh.", NotifType: "system", IsRead: true, CreatedAt: now.Add(-24 * time.Hour)},
			{UserID: 6, Title: "Lịch chụp X-Quang", Content: "Bạn có lịch chụp X-Quang ngực lúc 10:30 tại Phòng CĐHA.", NotifType: "reminder", IsRead: false, CreatedAt: now.Add(-1 * time.Hour)},
			// Benh nhan 8 (nguoi gia)
			{UserID: 8, Title: "Nhắc nhở khám định kỳ", Content: "Bạn cần đo đường huyết định kỳ. Vui lòng đến phòng Xét nghiệm tầng 1.", NotifType: "reminder", IsRead: false, CreatedAt: now.Add(-4 * time.Hour)},
			{UserID: 8, Title: "Hướng dẫn xe lăn", Content: "Bạn có thể mượn xe lăn miễn phí tại Trạm Sảnh Chính.", NotifType: "system", IsRead: false, CreatedAt: now.Add(-6 * time.Hour)},
			// System
			{UserID: 4, Title: "Bảo trì hệ thống", Content: "Hệ thống sẽ bảo trì từ 23:00-01:00 đêm nay.", NotifType: "system", IsRead: true, CreatedAt: now.Add(-48 * time.Hour)},
		}
		db.Create(&notifications)
		log.Printf("Da tao %d notifications", len(notifications))
	}

	log.Println("Seed du lieu y te hoan thanh")
}