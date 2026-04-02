package schema

import "time"

// ========================================
// MEDICAL  - Y te gia lap
// Slice 6: treatments, prescriptions, queues
// ========================================

// TaskType loai chi dinh kham.
type TaskType string

const (
	TaskTypeExam    TaskType = "exam"
	TaskTypeLab     TaskType = "lab"
	TaskTypeImaging TaskType = "imaging"
	TaskTypeMed     TaskType = "medication"
)

// TaskStatus trang thai chi dinh.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusSkipped    TaskStatus = "skipped"
)

// PrescriptionStatus trang thai don thuoc.
type PrescriptionStatus string

const (
	PrescriptionPending   PrescriptionStatus = "pending"
	PrescriptionDispensed PrescriptionStatus = "dispensed"
)

// Treatment chi dinh kham/xet nghiem cho benh nhan.
// Moi treatment gan vao 1 POI (phong kham) va 1 ward (khoa).
// Bang: treatments [T22]
type Treatment struct {
	TreatmentID    uint64     `gorm:"primaryKey;autoIncrement;column:treatment_id" json:"treatment_id"`
	UserID         uint64     `gorm:"not null;index;column:user_id" json:"user_id"`
	PoiID          uint32     `gorm:"not null;column:poi_id" json:"poi_id"`
	WardID         uint32     `gorm:"not null;column:ward_id" json:"ward_id"`
	TaskType       TaskType   `gorm:"not null;size:20;column:task_type" json:"task_type"`
	TaskName       string     `gorm:"not null;size:200;column:task_name" json:"task_name"`
	Priority       int        `gorm:"not null;default:0;column:priority" json:"priority"`
	SequenceNumber int        `gorm:"not null;default:1;column:sequence_number" json:"sequence_number"`
	Status         TaskStatus `gorm:"not null;default:pending;index;column:status" json:"status"`
	Note           string     `gorm:"type:text;column:note" json:"note"`
	HasResult      bool       `gorm:"not null;default:false;column:has_result" json:"has_result"`
	CreatedAt      time.Time  `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;autoUpdateTime;column:updated_at" json:"updated_at"`
	CheckinAt      *time.Time `gorm:"column:checkin_at"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`

	// Belongs-to
	User *User    `gorm:"foreignKey:UserID;references:UserID"`
	POI  *GridPOI `gorm:"foreignKey:PoiID;references:PoiID"`
	Ward *Ward    `gorm:"foreignKey:WardID;references:WardID"`
}

func (Treatment) TableName() string {
	return "treatments"
}

// Prescription don thuoc cua benh nhan.
// items_json la mang JSON chua ten thuoc, lieu luong, so luong.
// Bang: prescriptions [T23]
type Prescription struct {
	PrescriptionID uint64             `gorm:"primaryKey;autoIncrement;column:prescription_id" json:"prescription_id"`
	UserID         uint64             `gorm:"not null;index;column:user_id" json:"user_id"`
	IssuedBy       uint64             `gorm:"not null;column:issued_by" json:"issued_by"`
	PharmacyPoiID  *uint32            `gorm:"column:pharmacy_poi_id"`
	ItemsJSON      string             `gorm:"type:text;column:items_json" json:"items_json"`
	Status         PrescriptionStatus `gorm:"not null;default:pending;index;column:status" json:"status"`
	IssuedAt       time.Time          `gorm:"not null;autoCreateTime;column:issued_at" json:"issued_at"`
	DispensedAt    *time.Time         `gorm:"column:dispensed_at"`

	// Belongs-to
	User     *User    `gorm:"foreignKey:UserID;references:UserID"`
	Issuer   *Staff   `gorm:"foreignKey:IssuedBy;references:StaffID"`
	Pharmacy *GridPOI `gorm:"foreignKey:PharmacyPoiID;references:PoiID"`
}

func (Prescription) TableName() string {
	return "prescriptions"
}

// Queue hang doi tai 1 phong kham (POI).
// Moi POI chi co 1 queue (poi_id UNIQUE).
// Bang: queues [T24]
type Queue struct {
	QueueID        uint32    `gorm:"primaryKey;autoIncrement;column:queue_id" json:"queue_id"`
	PoiID          uint32    `gorm:"uniqueIndex;not null;column:poi_id" json:"poi_id"`
	CurrentNumber  int       `gorm:"not null;default:0;column:current_number" json:"current_number"`
	WaitingCount   int       `gorm:"not null;default:0;column:waiting_count" json:"waiting_count"`
	AvgWaitMinutes float64   `gorm:"not null;default:0;column:avg_wait_minutes" json:"avg_wait_minutes"`
	UpdatedAt      time.Time `gorm:"not null;autoUpdateTime;column:updated_at" json:"updated_at,omitempty"`

	// Belongs-to
	POI *GridPOI `gorm:"foreignKey:PoiID;references:PoiID"`
}

func (Queue) TableName() string {
	return "queues"
}
