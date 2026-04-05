package repository

import (
	"hospital/schema"

	"gorm.io/gorm"
)

// SupportRepo xu ly truy van database cho module SOS + Chat.
// Bao gom: sos_requests, conversations, messages.
// Person E so huu file nay.
type SupportRepo struct {
	db *gorm.DB
}

func NewSupportRepo(db *gorm.DB) *SupportRepo {
	return &SupportRepo{db: db}
}

// ========================================
// SOS REQUESTS
// ========================================

// CreateSOS tao yeu cau SOS moi.
func (r *SupportRepo) CreateSOS(sos *schema.SOSRequest) error {
	return r.db.Create(sos).Error
}

// FindSOSByID tim SOS theo sos_id.
func (r *SupportRepo) FindSOSByID(sosID uint64) (*schema.SOSRequest, error) {
	var sos schema.SOSRequest
	err := r.db.Where("sos_id = ?", sosID).First(&sos).Error
	if err != nil {
		return nil, err
	}
	return &sos, nil
}

// FindSOSList lay danh sach SOS (staff xem, pagination).
func (r *SupportRepo) FindSOSList(page, limit int) ([]schema.SOSRequest, int64, error) {
	var list []schema.SOSRequest
	var total int64

	q := r.db.Model(&schema.SOSRequest{})
	q.Count(&total)

	err := q.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&list).Error
	return list, total, err
}

// UpdateSOS cap nhat SOS (assign staff, resolve).
func (r *SupportRepo) UpdateSOS(sosID uint64, updates map[string]interface{}) error {
	return r.db.Model(&schema.SOSRequest{}).
		Where("sos_id = ?", sosID).
		Updates(updates).Error
}

// ========================================
// CONVERSATIONS
// ========================================

// CreateConversation tao phong chat moi.
func (r *SupportRepo) CreateConversation(conv *schema.Conversation) error {
	return r.db.Create(conv).Error
}

// FindConversationByID tim conversation theo conversation_id.
func (r *SupportRepo) FindConversationByID(convID uint64) (*schema.Conversation, error) {
	var conv schema.Conversation
	err := r.db.Where("conversation_id = ?", convID).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// FindConversationsByUserID lay DS phong chat cua benh nhan.
func (r *SupportRepo) FindConversationsByUserID(userID uint64) ([]schema.Conversation, error) {
	var convs []schema.Conversation
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&convs).Error
	return convs, err
}

// FindConversationsByStaffID lay DS phong chat cua staff (dung staff_id, khong phai user_id).
func (r *SupportRepo) FindConversationsByStaffID(staffID uint64) ([]schema.Conversation, error) {
	var convs []schema.Conversation
	err := r.db.Where("staff_id = ?", staffID).
		Order("updated_at DESC").
		Find(&convs).Error
	return convs, err
}

// UpdateConversation cap nhat conversation (dong phong, last_message, unread_count).
func (r *SupportRepo) UpdateConversation(convID uint64, updates map[string]interface{}) error {
	return r.db.Model(&schema.Conversation{}).
		Where("conversation_id = ?", convID).
		Updates(updates).Error
}

// ========================================
// STAFF HELPER
// ========================================

// FindStaffByUserID tim staff record tu user_id.
// Can vi conversations.staff_id luu staff_id (khong phai user_id).
func (r *SupportRepo) FindStaffByUserID(userID uint64) (*schema.Staff, error) {
	var staff schema.Staff
	err := r.db.Where("user_id = ?", userID).First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// ========================================
// MESSAGES
// ========================================

// CreateMessage luu tin nhan moi.
func (r *SupportRepo) CreateMessage(msg *schema.Message) error {
	return r.db.Create(msg).Error
}

// FindMessagesByConversationID lay lich su tin nhan (pagination).
func (r *SupportRepo) FindMessagesByConversationID(convID uint64, page, limit int) ([]schema.Message, int64, error) {
	var msgs []schema.Message
	var total int64

	q := r.db.Model(&schema.Message{}).
		Where("conversation_id = ? AND is_deleted = ?", convID, false)
	q.Count(&total)

	err := q.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&msgs).Error
	return msgs, total, err
}

// CountUnreadMessages dem so tin nhan chua doc trong conversation.
// Doc = tin nhan ma nguoi khac gui, minh chua doc.
func (r *SupportRepo) CountUnreadMessages(convID uint64, readerID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&schema.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = ? AND is_deleted = ?",
			convID, readerID, false, false).
		Count(&count).Error
	return count, err
}

// MarkMessagesAsRead danh dau tat ca tin nhan chua doc la da doc.
func (r *SupportRepo) MarkMessagesAsRead(convID uint64, readerID uint64) error {
	return r.db.Model(&schema.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND is_read = ?",
			convID, readerID, false).
		Update("is_read", true).Error
}

// ========================================
// TRANSACTIONS
// ========================================

// SendMessageTransaction tao message + cap nhat conversation trong 1 transaction.
func (r *SupportRepo) SendMessageTransaction(msg *schema.Message, convUpdates map[string]interface{}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		if err := tx.Model(&schema.Conversation{}).
			Where("conversation_id = ?", msg.ConversationID).
			Updates(convUpdates).Error; err != nil {
			return err
		}
		return nil
	})
}

// MarkReadTransaction danh dau da doc + reset unread_count trong 1 transaction.
func (r *SupportRepo) MarkReadTransaction(convID uint64, readerID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Danh dau tat ca tin nhan chua doc
		if err := tx.Model(&schema.Message{}).
			Where("conversation_id = ? AND sender_id != ? AND is_read = ?",
				convID, readerID, false).
			Update("is_read", true).Error; err != nil {
			return err
		}
		// Reset unread_count
		if err := tx.Model(&schema.Conversation{}).
			Where("conversation_id = ?", convID).
			Update("unread_count", 0).Error; err != nil {
			return err
		}
		return nil
	})
}
