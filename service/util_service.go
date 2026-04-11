package service

import (
	"errors"
	"hospital/repository"
	"hospital/schema"
)

type UtilService struct {
	repo *repository.UtilRepo
}

func NewUtilService(repo *repository.UtilRepo) *UtilService {
	return &UtilService{repo: repo}
}

// 1. FAQ & Hướng dẫn
func (s *UtilService) GetFAQs(category string) ([]schema.FAQ, error) {
	return s.repo.FindActiveFAQs(category)
}

// 2. Feedback (Phản hồi)
func (s *UtilService) SubmitFeedback(userID uint64, rating int, comment, images string) error {
	if rating < 1 || rating > 5 {
		return errors.New("diem danh gia phai tu 1 den 5")
	}

	feedback := &schema.Feedback{
		UserID:         userID,
		Rating:         rating,
		Comment:        comment,
		AttachedImages: images,
	}
	return s.repo.CreateFeedback(feedback)
}

func (s *UtilService) GetFeedbackSummary() (map[string]interface{}, error) {
	total, avg, err := s.repo.GetFeedbackSummary()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_feedbacks": total,
		"average_rating":  avg,
	}, nil
}

// 3. Kiểm tra phiên bản (Check Version)
func (s *UtilService) CheckVersion(platform string, clientCode int) (map[string]interface{}, error) {
	latest, err := s.repo.GetLatestVersion(platform)
	if err != nil {
		return nil, errors.New("khong lay duoc thong tin phien ban")
	}

	status := "up_to_date"
	if clientCode < latest.VersionCode {
		if latest.IsForceUpdate {
			status = "force_update" // Bat buoc cap nhat (VD: Ban cu bi loi nghiem trong)
		} else {
			status = "update_available" // Co the cap nhat hoac bo qua
		}
	}

	return map[string]interface{}{
		"status":          status,
		"latest_version":  latest.VersionName,
		"change_log":      latest.ChangeLog,
		"download_url":    latest.DownloadURL,
	}, nil
}

// 4. Các dữ liệu tĩnh (Static Data) - Không cần truy vấn DB
func (s *UtilService) GetLanguages() []map[string]string {
	return []map[string]string{
		{"code": "vi", "name": "Tiếng Việt"},
		{"code": "en", "name": "English"},
	}
}

func (s *UtilService) GetAboutInfo() map[string]string {
	return map[string]string{
		"hospital_name": "Bệnh viện Đa khoa Trung Tâm",
		"description":   "Hệ thống điều hướng và quản lý bệnh viện thông minh.",
		"version":       "1.0.0",
	}
}

func (s *UtilService) GetContactInfo() map[string]string {
	return map[string]string{
		"hotline": "1900-1234",
		"email":   "support@hospital.vn",
		"address": "123 Đường Y Tế, Quận 1, TP.HCM",
	}
}