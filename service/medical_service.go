package service

import (
	"errors"
	"hospital/middleware"
	"hospital/repository"
	"hospital/schema"
	"math/rand"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MedicalService interface {
	GetMyTasks(c *gin.Context) ([]schema.Treatment, error)
	GetQueueStatus(poiID uint32) (schema.Queue, error)
	CheckinRoom(c *gin.Context, treatmentID uint64) error
	CheckoutRoom(c *gin.Context, treatmentID uint64) error
	GetResultStatus(c *gin.Context, treatmentID uint64) (gin.H, error)
	GetPrescriptions(c *gin.Context) ([]schema.Prescription, error)
	CancelTask(c *gin.Context, treatmentID uint64) error
	GetHistory(c *gin.Context) ([]schema.Treatment, error)
	SyncHIS(c *gin.Context) error
	GetRoomOpeningHours(poiID uint32) (gin.H, error)
}

type medicalService struct {
	repo    repository.MedicalRepository
	mapRepo *repository.MapRepo
	db      *gorm.DB
}

func NewMedicalService(repo repository.MedicalRepository, mapRepo *repository.MapRepo, db *gorm.DB) MedicalService {
	return &medicalService{repo: repo, mapRepo: mapRepo, db: db}
}

// #61: Lấy danh sách nhiệm vụ y tế hiện tại của user
func (s *medicalService) GetMyTasks(c *gin.Context) ([]schema.Treatment, error) {
	userID := middleware.GetUserID(c) // Lấy userID từ token [cite: 47]
	// Chỉ lấy các task đang chờ hoặc đang khám [cite: 175]
	return s.repo.GetTreatmentsByUserID(userID, []string{"pending", "in_progress"})
}

// #62: Lấy trạng thái hàng đợi tại phòng khám
func (s *medicalService) GetQueueStatus(poiID uint32) (schema.Queue, error) {
	return s.repo.GetQueueByPOI(poiID)
}

// #63: Xử lý Check-in vào phòng khám
func (s *medicalService) CheckinRoom(c *gin.Context, treatmentID uint64) error {
	userID := middleware.GetUserID(c)

	// Bắt buộc dùng Transaction vì thay đổi 2 bảng treatments và queues 
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Cập nhật trạng thái treatment
		err := s.repo.UpdateTreatmentStatus(tx, treatmentID, userID, "in_progress")
		if err != nil {
			return err
		}

		// Lấy POI ID từ treatment
		var t schema.Treatment
		tx.First(&t, treatmentID)

		// 2. Giảm số lượng người đang chờ trong hàng đợi [cite: 177]
		return s.repo.UpdateQueueCount(tx, t.PoiID, -1)
	})
}

// #67: Giả lập đồng bộ dữ liệu từ HIS [cite: 171, 178]
func (s *medicalService) SyncHIS(c *gin.Context) error {
	userID := middleware.GetUserID(c)
	
	// Chọn ngẫu nhiên 3-5 phòng khám loại 'room' [cite: 178]
	pois, _ := s.mapRepo.FindPOIsByType(schema.POITypeRoom, 0)
	if len(pois) == 0 {
		return errors.New("no clinic rooms found to sync")
	}

	numTasks := rand.Intn(3) + 3 // Tạo 3-5 tasks [cite: 171, 178]

	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < numTasks; i++ {
			p := pois[rand.Intn(len(pois))]
			newT := schema.Treatment{
				UserID:         userID,
				PoiID:          p.POIID,
				TaskName:       "Kham lam sang " + p.POIName,
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

// #68: Lấy giờ mở cửa phòng khám (giả lập)
func (s *medicalService) GetRoomOpeningHours(poiID uint32) (gin.H, error) {
	poi, err := s.mapRepo.FindPOIByID(poiID)
	if err != nil || poi == nil {
		return nil, errors.New("room not found")
	}
	return gin.H{
		"poi_id":   poi.POIID,
		"poi_name": poi.POIName,
		"open":     "07:00",
		"close":    "17:00",
	}, nil
}

// #64: Checkout hoàn thành khám
func (s *medicalService) CheckoutRoom(c *gin.Context, treatmentID uint64) error {
	userID := middleware.GetUserID(c)
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Cập nhật treatment -> completed
		err := s.repo.UpdateTreatmentStatus(tx, treatmentID, userID, "completed")
		if err != nil {
			return err
		}
		// 2. Lấy POI ID và tăng current_number
		var t schema.Treatment
		tx.First(&t, treatmentID)
		return tx.Model(&schema.Queue{}).
			Where("poi_id = ?", t.PoiID).
			UpdateColumn("current_number", gorm.Expr("current_number + 1")).Error
	})
}

// #65: Xem trạng thái kết quả khám
func (s *medicalService) GetResultStatus(c *gin.Context, treatmentID uint64) (gin.H, error) {
	userID := middleware.GetUserID(c)
	t, err := s.repo.GetTreatmentByID(treatmentID, userID)
	if err != nil {
		return nil, errors.New("treatment not found")
	}
	return gin.H{
		"treatment_id": t.TreatmentID,
		"task_name":    t.TaskName,
		"status":       t.Status,
		"has_result":   t.HasResult,
	}, nil
}

// #66: Lấy đơn thuốc của bệnh nhân
func (s *medicalService) GetPrescriptions(c *gin.Context) ([]schema.Prescription, error) {
	userID := middleware.GetUserID(c)
	return s.repo.GetPrescriptionsByUser(userID)
}

// #69: Hủy chỉ định khám
func (s *medicalService) CancelTask(c *gin.Context, treatmentID uint64) error {
	userID := middleware.GetUserID(c)
	return s.db.Model(&schema.Treatment{}).
		Where("treatment_id = ? AND user_id = ? AND status = ?", treatmentID, userID, "pending").
		Update("status", "skipped").Error
}

// #70: Lịch sử khám đã hoàn thành
func (s *medicalService) GetHistory(c *gin.Context) ([]schema.Treatment, error) {
	userID := middleware.GetUserID(c)
	return s.repo.GetCompletedTreatments(userID)
}