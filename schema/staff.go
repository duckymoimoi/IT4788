package schema

import "time"

type StaffRole string

const (
	StaffRoleAdmin       StaffRole = "admin"
	StaffRoleCoordinator StaffRole = "coordinator"
	StaffRoleStaff       StaffRole = "staff"
)

// Staff nhan vien benh vien. Bang: staffs [T05]
type Staff struct {
	StaffID    uint64    `gorm:"primaryKey;autoIncrement;column:staff_id" json:"staff_id"`
	UserID     uint64    `gorm:"uniqueIndex;not null;column:user_id" json:"user_id"`
	StaffCode  string    `gorm:"uniqueIndex;not null;size:20;column:staff_code" json:"staff_code"`
	Role       StaffRole `gorm:"not null;index;column:role" json:"role"`
	WardID     *uint32   `gorm:"column:ward_id" json:"ward_id,omitempty"`
	IsActive   bool      `gorm:"not null;default:true;column:is_active" json:"is_active"`
	ShiftStart *string   `gorm:"size:5;column:shift_start" json:"shift_start,omitempty"`
	ShiftEnd   *string   `gorm:"size:5;column:shift_end" json:"shift_end,omitempty"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`

	User *User `gorm:"foreignKey:UserID;references:UserID" json:"-"`
	Ward *Ward `gorm:"foreignKey:WardID" json:"-"`
}

func (Staff) TableName() string { return "staffs" }

// Ward khoa/phong benh vien. Bang: wards [T06]
type Ward struct {
	WardID      uint32  `gorm:"primaryKey;autoIncrement;column:ward_id" json:"ward_id"`
	WardCode    string  `gorm:"uniqueIndex;not null;size:20;column:ward_code" json:"ward_code"`
	WardName    string  `gorm:"not null;size:200;column:ward_name" json:"ward_name"`
	HeadStaffID *uint64 `gorm:"column:head_staff_id" json:"head_staff_id,omitempty"`
	IsActive    bool    `gorm:"not null;default:true;column:is_active" json:"is_active"`

	Staffs []Staff `gorm:"foreignKey:WardID" json:"staffs,omitempty"`
}

func (Ward) TableName() string { return "wards" }
