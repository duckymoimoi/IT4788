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
			{Category: "Chung", Question: "Làm sao để mượn xe lăn?", Answer: "Bạn vào mục 'Thiết bị', chọn xe còn trống và nhấn 'Mượn'. Sau đó quét mã QR trên xe để xác nhận.", SortOrder: 1, IsActive: true},
			{Category: "Khám bệnh", Question: "Tôi có thể xem số thứ tự khám ở đâu?", Answer: "Vào mục 'Y tế' -> 'Hàng đợi' để theo dõi vị trí hiện tại của mình trong danh sách chờ.", SortOrder: 2, IsActive: true},
			{Category: "Chỉ đường", Question: "Làm thế nào để tìm đường đến phòng khám?", Answer: "Nhấn nút 'Tìm đường' trên màn hình chính, nhập tên phòng hoặc mã phòng, hệ thống sẽ vẽ lộ trình đi bộ chi tiết cho bạn.", SortOrder: 3, IsActive: true},
			{Category: "Chỉ đường", Question: "Ứng dụng có hỗ trợ chỉ đường bằng giọng nói không?", Answer: "Có! Hệ thống tự động phát giọng nói tiếng Việt khi bạn đi theo lộ trình. Bạn sẽ nghe 'Rẽ trái', 'Đi thẳng', v.v.", SortOrder: 4, IsActive: true},
			{Category: "Khám bệnh", Question: "Tôi cần mang theo giấy tờ gì khi đến khám?", Answer: "Bạn cần mang CMND/CCCD, thẻ BHYT (nếu có), và sổ khám bệnh cũ (nếu tái khám).", SortOrder: 5, IsActive: true},
			{Category: "Chung", Question: "Bệnh viện có Wi-Fi miễn phí không?", Answer: "Có! Vào mục 'Tiện ích' -> 'Wi-Fi' để xem tên mạng và mật khẩu cho từng khu vực.", SortOrder: 6, IsActive: true},
			{Category: "Chung", Question: "Căn tin bệnh viện ở đâu?", Answer: "Căn tin nằm ở tầng 1 khu nhà B. Bạn có thể dùng chức năng 'Tìm đường' với từ khóa 'căn tin' để được hướng dẫn.", SortOrder: 7, IsActive: true},
			{Category: "Thiết bị", Question: "Xe lăn bị hư thì báo ở đâu?", Answer: "Nhấn vào thiết bị đang mượn, chọn 'Báo hỏng'. Nhân viên kỹ thuật sẽ đến hỗ trợ trong vòng 10 phút.", SortOrder: 8, IsActive: true},
			{Category: "Khám bệnh", Question: "Tôi có thể hủy lịch khám được không?", Answer: "Vào 'Y tế' -> danh sách nhiệm vụ -> chọn lịch cần hủy -> nhấn 'Hủy'. Lưu ý: chỉ hủy được trước giờ khám 2 tiếng.", SortOrder: 9, IsActive: true},
			{Category: "Cấp cứu", Question: "Cần giúp đỡ khẩn cấp thì làm sao?", Answer: "Nhấn nút SOS màu đỏ trên màn hình chính. Thông tin vị trí của bạn sẽ được gửi ngay đến đội ngũ y tế trực.", SortOrder: 10, IsActive: true},
		}
		db.Create(&faqs)
		log.Printf("Da tao %d FAQs", len(faqs))
	}

	// 2. Seed App Version (Phiên bản App) - đã có trong seed.go, bỏ qua nếu trùng
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