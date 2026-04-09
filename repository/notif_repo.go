package repository

import (
	"hospital/schema"
	"gorm.io/gorm"
)

type NotifRepository interface {
	GetList(userID uint, page, limit int) ([]schema.Notification, int64, error)
	MarkAsRead(notifID uint, userID uint) error
	Delete(notifID uint, userID uint) error
	Create(notif *schema.Notification) error
}

type notifRepository struct {
	db *gorm.DB
}

func NewNotifRepository(db *gorm.DB) NotifRepository {
	return &notifRepository{db: db}
}

// [71] Lấy danh sách thông báo phân trang
func (r *notifRepository) GetList(userID uint, page, limit int) ([]schema.Notification, int64, error) {
	var notifs []schema.Notification
	var total int64
	offset := (page - 1) * limit

	r.db.Model(&schema.Notification{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifs).Error
	return notifs, total, err
}

// [72] Đánh dấu đã đọc
func (r *notifRepository) MarkAsRead(notifID uint, userID uint) error {
	return r.db.Model(&schema.Notification{}).
		Where("notif_id = ? AND user_id = ?", notifID, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error
}

// [73] Xóa thông báo
func (r *notifRepository) Delete(notifID uint, userID uint) error {
	return r.db.Where("notif_id = ? AND user_id = ?", notifID, userID).Delete(&schema.Notification{}).Error
}

func (r *notifRepository) Create(notif *schema.Notification) error {
	return r.db.Create(notif).Error
}