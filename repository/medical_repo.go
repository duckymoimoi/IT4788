package repository

import (
	"hospital/schema"

	"gorm.io/gorm"
)

type MedicalRepository interface {
	GetTreatmentsByUserID(userID uint64, status []string) ([]schema.Treatment, error)
	GetTreatmentByID(treatmentID uint64, userID uint64) (*schema.Treatment, error)
	GetQueueByPOI(poiID uint32) (schema.Queue, error)
	UpdateTreatmentStatus(tx *gorm.DB, treatmentID uint64, userID uint64, status string) error
	UpdateQueueCount(tx *gorm.DB, poiID uint32, delta int) error
	CreatePrescription(prescription *schema.Prescription) error
	GetPrescriptionsByUser(userID uint64) ([]schema.Prescription, error)
	GetCompletedTreatments(userID uint64) ([]schema.Treatment, error)
}

type medicalRepository struct {
	db *gorm.DB
}

func NewMedicalRepository(db *gorm.DB) MedicalRepository {
	return &medicalRepository{db: db}
}

// Lấy danh sách chỉ định khám của bệnh nhân [cite: 175, 176]
func (r *medicalRepository) GetTreatmentsByUserID(userID uint64, status []string) ([]schema.Treatment, error) {
	var treatments []schema.Treatment
	err := r.db.Where("user_id = ? AND status IN ?", userID, status).
		Order("priority ASC, sequence_number ASC").Find(&treatments).Error
	return treatments, err
}

// Lấy 1 treatment theo ID + userID
func (r *medicalRepository) GetTreatmentByID(treatmentID uint64, userID uint64) (*schema.Treatment, error) {
	var t schema.Treatment
	err := r.db.Where("treatment_id = ? AND user_id = ?", treatmentID, userID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Lấy trạng thái hàng đợi tại một phòng [cite: 176]
func (r *medicalRepository) GetQueueByPOI(poiID uint32) (schema.Queue, error) {
	var queue schema.Queue
	err := r.db.Where("poi_id = ?", poiID).First(&queue).Error
	return queue, err
}

// Cập nhật trạng thái khám (Cần dùng Transaction) [cite: 47, 176, 177]
func (r *medicalRepository) UpdateTreatmentStatus(tx *gorm.DB, treatmentID uint64, userID uint64, status string) error {
	return tx.Model(&schema.Treatment{}).
		Where("treatment_id = ? AND user_id = ?", treatmentID, userID).
		Update("status", status).Error
}

// Cập nhật số lượng người chờ trong hàng đợi [cite: 177]
func (r *medicalRepository) UpdateQueueCount(tx *gorm.DB, poiID uint32, delta int) error {
	return tx.Model(&schema.Queue{}).
		Where("poi_id = ?", poiID).
		UpdateColumn("waiting_count", gorm.Expr("waiting_count + ?", delta)).Error
}

// Tạo đơn thuốc mới
func (r *medicalRepository) CreatePrescription(prescription *schema.Prescription) error {
	return r.db.Create(prescription).Error
}

// Lấy đơn thuốc theo user
func (r *medicalRepository) GetPrescriptionsByUser(userID uint64) ([]schema.Prescription, error) {
	var prescriptions []schema.Prescription
	err := r.db.Where("user_id = ?", userID).Order("issued_at DESC").Find(&prescriptions).Error
	return prescriptions, err
}

// Lấy lịch sử khám đã hoàn thành
func (r *medicalRepository) GetCompletedTreatments(userID uint64) ([]schema.Treatment, error) {
	var treatments []schema.Treatment
	err := r.db.Where("user_id = ? AND status = ?", userID, "completed").
		Order("completed_at DESC").Find(&treatments).Error
	return treatments, err
}

