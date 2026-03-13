package schema

import "time"

// StaffRole cap phan quyen cua nhan vien trong he thong.
//
// admin:       toan quyen, quan ly ban do, nhan vien, xem thong ke.
// coordinator: giam sat luong nguoi, xu ly su co, quan ly xe lan.
// staff:       xem ban do, xem hang doi, ho tro benh nhan.
type StaffRole string

const (
	StaffRoleAdmin       StaffRole = "admin"
	StaffRoleCoordinator StaffRole = "coordinator"
	StaffRoleStaff       StaffRole = "staff"
)

// Staff luu thong tin nghiep vu va phan quyen cho nhan vien benh vien.
// Moi nhan vien bat buoc phai co row trong bang users truoc
// (voi user_type = "staff"), bang nay la extension.
// Khi tao nhan vien moi, phai INSERT ca users va staffs trong 1 transaction.
//
// Ca truc luu dang chuoi thoi gian "HH:MM", vi du "07:00", "13:00".
// Cho phep linh hoat thiet lap bat ky khung gio nao ma khong bi gioi han
// boi enum co dinh. Admin co the tu tao ca truc tuy y sau nay.
// Bang: staffs [T05]
type Staff struct {
	StaffID    uint64    `gorm:"primaryKey;autoIncrement;column:staff_id"`
	UserID     uint64    `gorm:"uniqueIndex;not null;column:user_id"`
	StaffCode  string    `gorm:"uniqueIndex;not null;size:20;column:staff_code"`
	Role       StaffRole `gorm:"not null;index;column:role"`
	WardID     *uint32   `gorm:"column:ward_id"`
	IsActive   bool      `gorm:"not null;default:true;column:is_active"`
	ShiftStart *string   `gorm:"size:5;column:shift_start"` // vi du: "07:00"
	ShiftEnd   *string   `gorm:"size:5;column:shift_end"`   // vi du: "13:00"
	CreatedAt  time.Time `gorm:"not null;autoCreateTime;column:created_at"`

	// Belongs-to: thong tin tai khoan he thong
	User *User `gorm:"foreignKey:UserID;references:UserID"`

	// Belongs-to: khoa/phong phu trach (NULL = quan ly toan vien)
	Ward *Ward `gorm:"foreignKey:WardID"`
}

func (Staff) TableName() string {
	return "staffs"
}

// Ward danh muc khoa/vien cua benh vien.
// Dung de nhom cac phong kham theo chuyen khoa va phan cong nhan vien.
//
// Luu y circular dependency:
//   staffs.ward_id  -> wards.ward_id
//   wards.head_staff_id -> staffs.staff_id
//
// Khi seed data, INSERT ward voi head_staff_id = NULL truoc,
// tao staff xong roi UPDATE wards SET head_staff_id = <id>.
// Bang: wards [T06]
type Ward struct {
	WardID      uint32  `gorm:"primaryKey;autoIncrement;column:ward_id"`
	WardCode    string  `gorm:"uniqueIndex;not null;size:20;column:ward_code"`
	WardName    string  `gorm:"not null;size:200;column:ward_name"`
	HeadStaffID *uint64 `gorm:"column:head_staff_id"`
	IsActive    bool    `gorm:"not null;default:true;column:is_active"`

	// Has-many: cac nhan vien thuoc khoa nay
	Staffs []Staff `gorm:"foreignKey:WardID"`
}

func (Ward) TableName() string {
	return "wards"
}
