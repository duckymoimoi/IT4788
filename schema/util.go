package schema

import "time"

// ========================================
// UTIL  - Tien ich: feedback, FAQ
// Slice 10
// ========================================

// Feedback phan hoi/danh gia tu benh nhan.
// attached_images la mang JSON chua URL anh dinh kem.
// Bang: feedbacks [T33]
type Feedback struct {
	FeedbackID     uint64    `gorm:"primaryKey;autoIncrement;column:feedback_id" json:"feedback_id"`
	UserID         uint64    `gorm:"not null;index;column:user_id" json:"user_id"`
	Rating         int       `gorm:"not null;column:rating" json:"rating"`
	Comment        string    `gorm:"type:text;column:comment" json:"comment"`
	AttachedImages string    `gorm:"type:text;column:attached_images" json:"attached_images,omitempty"` // JSON array of URLs
	CreatedAt      time.Time `gorm:"not null;autoCreateTime;column:created_at" json:"created_at,omitempty"`

	// Belongs-to
	User *User `gorm:"foreignKey:UserID;references:UserID"`
}

func (Feedback) TableName() string {
	return "feedbacks"
}

// FAQ cau hoi thuong gap.
// Admin tao, sap xep theo sort_order.
// Bang: faqs [T34]
type FAQ struct {
	FaqID     uint32 `gorm:"primaryKey;autoIncrement;column:faq_id" json:"faq_id"`
	Category  string `gorm:"not null;size:50;index;column:category" json:"category"`
	Question  string `gorm:"not null;type:text;column:question" json:"question"`
	Answer    string `gorm:"not null;type:text;column:answer" json:"answer"`
	SortOrder int    `gorm:"not null;default:0;column:sort_order" json:"sort_order"`
	IsActive  bool   `gorm:"not null;default:true;column:is_active" json:"is_active"`
}

func (FAQ) TableName() string {
	return "faqs"
}
