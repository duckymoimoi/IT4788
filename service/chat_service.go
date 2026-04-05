package service

import (
	"fmt"
	"time"

	"hospital/repository"
	"hospital/schema"
)

// ChatService xu ly logic nghiep vu cho module Chat.
// Person E so huu file nay.
type ChatService struct {
	repo *repository.SupportRepo
}

func NewChatService(repo *repository.SupportRepo) *ChatService {
	return &ChatService{repo: repo}
}

// ========================================
// API #101 POST create_room
// ========================================

// CreateRoom tao phong chat moi giua benh nhan va staff.
func (s *ChatService) CreateRoom(userID, staffID uint64, topic string) (*schema.Conversation, error) {
	conv := &schema.Conversation{
		UserID:  userID,
		StaffID: staffID,
		Topic:   topic,
		Status:  schema.ConversationOpen,
	}

	if err := s.repo.CreateConversation(conv); err != nil {
		return nil, fmt.Errorf("cannot create chat room: %w", err)
	}

	return conv, nil
}

// ========================================
// API #102 GET get_rooms
// ========================================

// GetRooms lay danh sach phong chat.
// Benh nhan: tim theo user_id.
// Staff: tim theo staff_id (can map tu user_id -> staff_id).
func (s *ChatService) GetRooms(userID uint64, role string) ([]schema.Conversation, error) {
	switch role {
	case "admin", "coordinator", "staff":
		// Staff: can tim staff_id tu user_id
		staff, err := s.repo.FindStaffByUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("staff record not found for user")
		}
		return s.repo.FindConversationsByStaffID(staff.StaffID)
	default:
		// Benh nhan: tim truc tiep theo user_id
		return s.repo.FindConversationsByUserID(userID)
	}
}

// ========================================
// API #103 GET get_messages
// ========================================

// GetMessages lay lich su tin nhan trong conversation (pagination).
func (s *ChatService) GetMessages(convID uint64, page, limit int) ([]schema.Message, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// Kiem tra conversation ton tai
	_, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return nil, 0, fmt.Errorf("conversation not found")
	}

	return s.repo.FindMessagesByConversationID(convID, page, limit)
}

// ========================================
// API #104 POST send_message
// ========================================

// SendMessage gui tin nhan trong conversation.
// Tao message + cap nhat last_message va unread_count (transaction).
func (s *ChatService) SendMessage(convID, senderID uint64, senderType schema.SenderType, msgType schema.MessageType, textContent, mediaURL string) (*schema.Message, error) {
	// Kiem tra conversation ton tai va dang open
	conv, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}
	if conv.Status != schema.ConversationOpen {
		return nil, fmt.Errorf("conversation is closed")
	}

	// Validate noi dung
	if textContent == "" && mediaURL == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	msg := &schema.Message{
		ConversationID: convID,
		SenderID:       senderID,
		SenderType:     senderType,
		Type:           msgType,
		TextContent:    textContent,
		MediaURL:       mediaURL,
	}

	// Tao preview cho last_message
	preview := textContent
	if preview == "" {
		preview = "[media]"
	}
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	convUpdates := map[string]interface{}{
		"last_message": preview,
		"unread_count": conv.UnreadCount + 1,
	}

	if err := s.repo.SendMessageTransaction(msg, convUpdates); err != nil {
		return nil, fmt.Errorf("cannot send message: %w", err)
	}

	return msg, nil
}

// ========================================
// API #105 POST close_room
// ========================================

// CloseRoom dong phong chat (chi staff).
func (s *ChatService) CloseRoom(convID uint64) error {
	conv, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return fmt.Errorf("conversation not found")
	}
	if conv.Status == schema.ConversationClosed {
		return fmt.Errorf("conversation is already closed")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":    schema.ConversationClosed,
		"closed_at": &now,
	}
	return s.repo.UpdateConversation(convID, updates)
}

// ========================================
// API #106 GET get_unread_count
// ========================================

// GetUnreadCount dem so tin nhan chua doc trong conversation.
func (s *ChatService) GetUnreadCount(convID, userID uint64) (int64, error) {
	// Kiem tra conversation ton tai
	_, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return 0, fmt.Errorf("conversation not found")
	}

	return s.repo.CountUnreadMessages(convID, userID)
}

// ========================================
// API #107 POST mark_read
// ========================================

// MarkRead danh dau tat ca tin nhan da doc + reset unread_count (transaction).
func (s *ChatService) MarkRead(convID, readerID uint64) error {
	// Kiem tra conversation ton tai
	_, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return fmt.Errorf("conversation not found")
	}

	return s.repo.MarkReadTransaction(convID, readerID)
}
