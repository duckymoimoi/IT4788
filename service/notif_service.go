package service

import (
	"hospital/repository"
	"hospital/schema"
	"log"
)

type NotifService interface {
	GetNotifications(userID uint64, page, limit int) ([]schema.Notification, int64, error)
	ReadNotification(notifID uint64, userID uint64) error
	DeleteNotification(notifID uint64, userID uint64) error
	// Hàm nội bộ để các module khác gọi
	SendNotification(userID uint64, title, content, notifType string) error
}

type notifService struct {
	repo repository.NotifRepository
}

func NewNotifService(repo repository.NotifRepository) NotifService {
	return &notifService{repo: repo}
}

func (s *notifService) GetNotifications(userID uint64, page, limit int) ([]schema.Notification, int64, error) {
	return s.repo.GetList(userID, page, limit)
}

func (s *notifService) ReadNotification(notifID uint64, userID uint64) error {
	return s.repo.MarkAsRead(notifID, userID)
}

func (s *notifService) DeleteNotification(notifID uint64, userID uint64) error {
	return s.repo.Delete(notifID, userID)
}

// Logic gửi thông báo quan trọng
func (s *notifService) SendNotification(userID uint64, title, content, notifType string) error {
	notif := schema.Notification{
		UserID:    userID,
		Title:     title,
		Content:   content,
		NotifType: schema.NotifType(notifType),
		IsRead:    false,
	}
	
	// 1. Lưu vào Database
	if err := s.repo.Create(&notif); err != nil {
		return err
	}

	// 2. Giả lập gửi qua FCM (Firebase)
	// Trong thực tế bạn sẽ lấy FCM Token từ bảng fcm_tokens và gọi API Google
	log.Printf("PUSH NOTIFICATION [%s] to User %d: %s", title, userID, content)
	
	return nil
}