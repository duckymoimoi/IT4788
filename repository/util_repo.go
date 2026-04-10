package repository

import (
	"hospital/schema"

	"gorm.io/gorm"
)

// UtilRepo xu ly truy van database cho module Utilities.
// Bao gom: faqs, feedbacks, app_versions.
type UtilRepo struct {
	db *gorm.DB
}

func NewUtilRepo(db *gorm.DB) *UtilRepo {
	return &UtilRepo{db: db}
}

// ========================================
// FAQ (Câu hỏi thường gặp)
// ========================================

// FindActiveFAQs lay danh sach cau hoi thuong gap dang active.
// Co the loc theo category (neu category != "").
func (r *UtilRepo) FindActiveFAQs(category string) ([]schema.FAQ, error) {
	var faqs []schema.FAQ

	// Luon chi lay cac cau hoi dang hien thi (is_active = true)
	query := r.db.Where("is_active = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Sap xep theo truong sort_order (uu tien tu nho den lon)
	err := query.Order("sort_order ASC").Find(&faqs).Error
	return faqs, err
}

// ========================================
// FEEDBACK (Phản hồi người dùng)
// ========================================

// CreateFeedback luu danh gia moi cua benh nhan vao database.
func (r *UtilRepo) CreateFeedback(feedback *schema.Feedback) error {
	return r.db.Create(feedback).Error
}

// GetFeedbackSummary lay thong ke tong quan ve danh gia (Tong so luot, diem trung binh).
func (r *UtilRepo) GetFeedbackSummary() (int64, float64, error) {
	var totalCount int64
	var avgRating float64

	// Dem tong so luot danh gia
	err := r.db.Model(&schema.Feedback{}).Count(&totalCount).Error
	if err != nil {
		return 0, 0, err
	}

	// Tinh diem trung binh neu co it nhat 1 danh gia
	if totalCount > 0 {
		r.db.Model(&schema.Feedback{}).Select("AVG(rating)").Row().Scan(&avgRating)
	}

	return totalCount, avgRating, nil
}

// ========================================
// APP VERSIONS (Cập nhật ứng dụng)
// ========================================

// GetLatestVersion lay phien ban moi nhat cua mot nen tang (android hoac ios).
func (r *UtilRepo) GetLatestVersion(platform string) (*schema.AppVersion, error) {
	var version schema.AppVersion
	
	err := r.db.Where("platform = ?", platform).
		Order("version_code DESC"). // Lay phien ban co version_code to nhat (moi nhat)
		First(&version).Error
		
	if err != nil {
		return nil, err
	}
	return &version, nil
}