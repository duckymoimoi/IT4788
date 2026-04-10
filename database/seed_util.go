package database

import (
	"hospital/schema"
	"log"

	"gorm.io/gorm"
)

// SeedUtils khoi tao du lieu mau cho FAQ va App Version.
func SeedUtils(db *gorm.DB) {
	// 1. Seed FAQ (Câu hỏi thường gặp)
	var faqCount int64
	db.Model(&schema.FAQ{}).Count(&faqCount)
	if faqCount == 0 {
		log.Println("Seeding FAQs...")
		faqs := []schema.FAQ{
			{
				Category:  "Chung",
				Question:  "Làm sao để mượn xe lăn?",
				Answer:    "Bạn vào mục 'Thiết bị', chọn xe còn trống và nhấn 'Mượn'. Sau đó quét mã QR trên xe để xác nhận.",
				SortOrder: 1,
				IsActive:  true,
			},
			{
				Category:  "Khám bệnh",
				Question:  "Tôi có thể xem số thứ tự khám ở đâu?",
				Answer:    "Vào mục 'Y tế' -> 'Hàng đợi' để theo dõi vị trí hiện tại của mình trong danh sách chờ.",
				SortOrder: 2,
				IsActive:  true,
			},
		}
		db.Create(&faqs)
	}

	// 2. Seed App Version (Phiên bản App)
	var versionCount int64
	db.Model(&schema.AppVersion{}).Count(&versionCount)
	if versionCount == 0 {
		log.Println("Seeding App Version...")
		version := schema.AppVersion{
			Platform:      "android",
			VersionName:   "1.0.0",
			VersionCode:   1,
			IsForceUpdate: false,
			ChangeLog:     "Phiên bản khởi tạo đầu tiên của hệ thống.",
			DownloadURL:   "https://hospital.vn/download/android",
		}
		db.Create(&version)
	}
}