package service

import (
	"errors"
	"hospital/middleware"
	"hospital/repository"
	"hospital/schema"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MedicalService interface {
	GetMyTasks(c *gin.Context) ([]schema.Treatment, error)
	CheckinRoom(c *gin.Context, treatmentID uint) error
	SyncHIS(c *gin.Context) error
}

type medicalService struct {
	repo   repository.MedicalRepository
	mapRepo repository.MapRepository // Để lấy POI khi sync
	db     *gorm.DB
}

func NewMedicalService(repo repository.MedicalRepository, mapRepo repository.MapRepository, db *gorm.DB) MedicalService {
	return &medicalService{repo: repo, mapRepo: mapRepo, db: db}
}

// #61: Lấy danh sách nhiệm vụ y tế hiện tại của user
func (s *medicalService) GetMyTasks(c *gin.Context) ([]schema.Treatment, error) {
	userID := middleware.GetUserID(c) // Lấy userID từ token [cite: 47]
	// Chỉ lấy các task đang chờ hoặc đang khám [cite: 175]
	return s.repo.GetTreatmentsByUserID(userID, []string{"pending", "in_progress"})
}

// #63: Xử lý Check-in vào phòng khám
func (s *medicalService) CheckinRoom(c *gin.Context, treatmentID uint) error {
	userID := middleware.GetUserID(c)

	// Bắt buộc dùng Transaction vì thay đổi 2 bảng treatments và queues 
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Cập nhật trạng thái treatment
		err := s.repo.UpdateTreatmentStatus(tx, treatmentID, userID, "in_progress")
		if err != nil {
			return err
		}

		// Giả sử chúng ta lấy được POI ID từ treatment này
		// (Trong thực tế bạn sẽ query treatment để lấy poi_id trước)
		var t schema.Treatment
		tx.First(&t, treatmentID)

		// 2. Giảm số lượng người đang chờ trong hàng đợi [cite: 177]
		return s.repo.UpdateQueueCount(tx, t.PoiID, -1)
	})
}

// #67: Giả lập đồng bộ dữ liệu từ HIS [cite: 171, 178]
func (s *medicalService) SyncHIS(c *gin.Context) error {
	userID := middleware.GetUserID(c)
	
	// Chọn ngẫu nhiên 3-5 phòng khám loại 'room' hoặc 'pharmacy' [cite: 178]
	pois, _ := s.mapRepo.GetPOIsByType([]string{"room", "pharmacy"})
	if len(pois) == 0 {
		return errors.New("no clinic rooms found to sync")
	}

	rand.Seed(time.Now().UnixNano())
	numTasks := rand.Intn(3) + 3 // Tạo 3-5 tasks [cite: 171, 178]

	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < numTasks; i++ {
			p := pois[rand.Intn(len(pois))]
			newT := schema.Treatment{
				UserID:         userID,
				PoiID:          p.PoiID,
				TaskName:       "Khám lâm sàng " + p.PoiName,
				Status:         "pending",
				Priority:       5,
				SequenceNumber: i + 1,
			}
			if err := tx.Create(&newT).Error; err != nil {
				return err
			}
		}
		return nil
	})
}