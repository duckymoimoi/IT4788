package schema

import "time"

// ========================================
// SUPPORT  - Thong bao, SOS, Chat
// Slice 8 (notifications) + Slice 9 (sos, conversations, messages)
// ========================================

// NotifType loai thong bao.
type NotifType string

const (
	NotifTypeMedical   NotifType = "medical"
	NotifTypeBilling   NotifType = "billing"
	NotifTypeQueue     NotifType = "queue"
	NotifTypeEmergency NotifType = "emergency"
	NotifTypeSystem    NotifType = "system"
)

// SOSStatus trang thai yeu cau cap cuu.
type SOSStatus string

const (
	SOSStatusReceived SOSStatus = "received"
	SOSStatusAssigned SOSStatus = "assigned"
	SOSStatusResolved SOSStatus = "resolved"
)

// ConversationStatus trang thai cuoc hoi thoai.
type ConversationStatus string

const (
	ConversationOpen   ConversationStatus = "open"
	ConversationClosed ConversationStatus = "closed"
)

// MessageType loai tin nhan.
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVoice MessageType = "voice"
)

// SenderType nguoi gui tin nhan.
type SenderType string

const (
	SenderTypeUser  SenderType = "user"
	SenderTypeStaff SenderType = "staff"
	SenderTypeBot   SenderType = "bot"
)

// Notification thong bao day cho nguoi dung.
// Luu DB + gui push qua FCM.
// Bang: notifications [T29]
type Notification struct {
	NotifID   uint64     `gorm:"primaryKey;autoIncrement;column:notif_id" json:"notif_id"`
	UserID    uint64     `gorm:"not null;index;column:user_id" json:"user_id"`
	Title     string     `gorm:"not null;size:200;column:title" json:"title"`
	Content   string     `gorm:"type:text;column:content" json:"content"`
	NotifType NotifType  `gorm:"not null;size:20;index;column:notif_type" json:"notif_type"`
	IsRead    bool       `gorm:"not null;default:false;column:is_read" json:"is_read"`
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"not null;autoCreateTime;index;column:created_at" json:"created_at"`
	ReadAt    *time.Time `gorm:"column:read_at"`

	// Belongs-to
	User *User `gorm:"foreignKey:UserID;references:UserID"`
}

func (Notification) TableName() string {
	return "notifications"
}

// SOSRequest yeu cau cap cuu.
// Benh nhan bam SOS -> thong bao toi coordinator.
// Bang: sos_requests [T30]
type SOSRequest struct {
	SosID          uint64     `gorm:"primaryKey;autoIncrement;column:sos_id" json:"sos_id"`
	UserID         uint64     `gorm:"not null;index;column:user_id" json:"user_id"`
	GridLocation   int        `gorm:"not null;column:grid_location" json:"grid_location"`
	PosX           float64    `gorm:"column:pos_x"`
	PosY           float64    `gorm:"column:pos_y"`
	Note           string     `gorm:"type:text;column:note" json:"note"`
	Status         SOSStatus  `gorm:"not null;default:received;index;column:status" json:"status"`
	AssignedStaff  *uint64    `gorm:"column:assigned_staff_id"`
	CreatedAt      time.Time  `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	ResolvedAt     *time.Time `gorm:"column:resolved_at"`

	// Belongs-to
	User  *User  `gorm:"foreignKey:UserID;references:UserID"`
	Staff *Staff `gorm:"foreignKey:AssignedStaff;references:StaffID"`
}

func (SOSRequest) TableName() string {
	return "sos_requests"
}

// Conversation cuoc hoi thoai giua benh nhan va nhan vien.
// Moi conversation co 1 user + 1 staff.
// Bang: conversations [T31]
type Conversation struct {
	ConversationID uint64             `gorm:"primaryKey;autoIncrement;column:conversation_id" json:"conversation_id"`
	UserID         uint64             `gorm:"not null;index;column:user_id" json:"user_id"`
	StaffID        uint64             `gorm:"not null;index;column:staff_id" json:"staff_id"`
	Topic          string             `gorm:"size:200;column:topic" json:"topic"`
	Status         ConversationStatus `gorm:"not null;default:open;index;column:status" json:"status"`
	LastMessage    string             `gorm:"type:text;column:last_message" json:"last_message"`
	UnreadCount    int                `gorm:"not null;default:0;column:unread_count" json:"unread_count"`
	CreatedAt      time.Time          `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt      time.Time          `gorm:"not null;autoUpdateTime;column:updated_at" json:"updated_at"`
	ClosedAt       *time.Time         `gorm:"column:closed_at"`

	// Belongs-to
	User  *User  `gorm:"foreignKey:UserID;references:UserID"`
	Staff *Staff `gorm:"foreignKey:StaffID;references:StaffID"`

	// Has-many
	Messages []Message `gorm:"foreignKey:ConversationID"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// Message tin nhan trong cuoc hoi thoai.
// Ho tro text, image, voice.
// Bang: messages [T32]
type Message struct {
	MessageID      uint64      `gorm:"primaryKey;autoIncrement;column:message_id" json:"message_id"`
	ConversationID uint64      `gorm:"not null;index;column:conversation_id" json:"conversation_id"`
	SenderID       uint64      `gorm:"not null;column:sender_id" json:"sender_id"`
	SenderType     SenderType  `gorm:"not null;size:10;column:sender_type" json:"sender_type"`
	Type           MessageType `gorm:"not null;size:10;column:type" json:"type"`
	TextContent    string      `gorm:"type:text;column:text_content" json:"text_content"`
	MediaURL       string      `gorm:"size:255;column:media_url" json:"media_url"`
	IsDeleted      bool        `gorm:"not null;default:false;column:is_deleted" json:"is_deleted"`
	IsRead         bool        `gorm:"not null;default:false;column:is_read" json:"is_read"`
	CreatedAt      time.Time   `gorm:"not null;autoCreateTime;index;column:created_at" json:"created_at"`
	DeletedAt      *time.Time  `gorm:"column:deleted_at"`

	// Belongs-to
	Conversation *Conversation `gorm:"foreignKey:ConversationID;references:ConversationID"`
}

func (Message) TableName() string {
	return "messages"
}
