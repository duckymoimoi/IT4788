package schema

import "time"

// UserType phan loai nguoi dung trong he thong.
// patient = benh nhan / nguoi nha, staff = nhan vien benh vien.
type UserType string

const (
	UserTypePatient UserType = "patient"
	UserTypeStaff   UserType = "staff"
)

// UserStatus trang thai hoat dong cua tai khoan.
type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusPending UserStatus = "pending"
	UserStatusBanned  UserStatus = "banned"
	UserStatusDeleted UserStatus = "deleted"
)

// Gender gioi tinh nguoi dung.
type Gender string

const (
	GenderMale   Gender = "M"
	GenderFemale Gender = "F"
	GenderOther  Gender = "O"
)

// User luu thong tin tai khoan cua tat ca nguoi dung (ca benh nhan lan nhan vien).
// Nhan vien se co them row tuong ung trong bang staffs.
// Bang: users [T01]
type User struct {
	UserID       uint64     `gorm:"primaryKey;autoIncrement;column:user_id"`
	PhoneNumber  string     `gorm:"uniqueIndex;not null;size:15;column:phone_number"`
	PasswordHash string     `gorm:"not null;size:255;column:password_hash"`
	FullName     string     `gorm:"not null;size:100;column:full_name"`
	UserType     UserType   `gorm:"not null;default:patient;column:user_type"`
	DateOfBirth  *time.Time `gorm:"column:date_of_birth"`
	Gender       *Gender    `gorm:"column:gender"`
	AvatarURL    *string    `gorm:"size:500;column:avatar_url"`
	Status       UserStatus `gorm:"not null;default:active;column:status"`
	CreatedAt    time.Time  `gorm:"not null;autoCreateTime;column:created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`

	// Has-one: moi user co dung 1 dong setting
	Setting *UserSetting `gorm:"foreignKey:UserID"`

	// Has-many: 1 user co the co nhieu FCM token (nhieu thiet bi)
	FCMTokens []FCMToken `gorm:"foreignKey:UserID"`

	// Has-one: chi co khi UserType = "staff"
	Staff *Staff `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

// OTPType loai ma OTP de tranh dung nham ma signup cho reset_password.
type OTPType string

const (
	OTPTypeSignup        OTPType = "signup"
	OTPTypeResetPassword OTPType = "reset_password"
	OTPTypeChangePhone   OTPType = "change_phone"
)

// OTPCode luu ma xac thuc OTP.
// Trong MVP, server tra otp_code plain text qua field DebugCode trong response
// thay vi gui SMS that. Bang nay van can thiet de quan ly trang thai,
// tranh su dung lai ma da dung (replay attack).
// Bang: otp_codes [T02]
type OTPCode struct {
	OTPID       uint64    `gorm:"primaryKey;autoIncrement;column:otp_id"`
	PhoneNumber string    `gorm:"not null;size:15;index:idx_phone_type;column:phone_number"`
	OTPCode     string    `gorm:"not null;size:8;column:otp_code"`
	Type        OTPType   `gorm:"not null;index:idx_phone_type;column:type"`
	ExpiredAt   time.Time `gorm:"not null;column:expired_at"`
	IsUsed      bool      `gorm:"not null;default:false;column:is_used"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime;column:created_at"`
}

func (OTPCode) TableName() string {
	return "otp_codes"
}

// TravelMode che do di chuyen anh huong den thuat toan tim duong.
// wheelchair: chi chon edge co wheelchair_accessible = true.
// stretcher: chon hanh lang du rong, tranh thang bo.
type TravelMode string

const (
	TravelModeWalk       TravelMode = "walk"
	TravelModeWheelchair TravelMode = "wheelchair"
	TravelModeStretcher  TravelMode = "stretcher"
)

// UserSetting luu cau hinh ca nhan cua nguoi dung.
// Quan he one-to-one voi users. Row nay tu dong tao khi user dang ky.
// Bang: user_settings [T03]
type UserSetting struct {
	SettingID            uint64     `gorm:"primaryKey;autoIncrement;column:setting_id"`
	UserID               uint64     `gorm:"uniqueIndex;not null;column:user_id"`
	VoiceGuidanceEnabled bool       `gorm:"not null;default:true;column:voice_guidance_enabled"`
	NotificationEnabled  bool       `gorm:"not null;default:true;column:notification_enabled"`
	TravelMode           TravelMode `gorm:"not null;default:walk;column:travel_mode"`
	Language             string     `gorm:"not null;default:vi;size:10;column:language"`
	UpdatedAt            *time.Time `gorm:"column:updated_at"`
}

func (UserSetting) TableName() string {
	return "user_settings"
}

// DevicePlatform nen tang thiet bi di dong.
type DevicePlatform string

const (
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformAndroid DevicePlatform = "android"
)

// FCMToken luu Firebase Cloud Messaging token cua tung thiet bi.
// Mot user co the co nhieu row (nhieu thiet bi). Trong MVP chua gui
// push notification nhung van luu de giam sat loai thiet bi va san sang
// tich hop FCM sau nay ma khong phai doi schema.
// Bang: fcm_tokens [T04]
type FCMToken struct {
	TokenID        uint64         `gorm:"primaryKey;autoIncrement;column:token_id"`
	UserID         uint64         `gorm:"not null;index;column:user_id"`
	FCMToken       string         `gorm:"not null;uniqueIndex;size:500;column:fcm_token"`
	DevicePlatform DevicePlatform `gorm:"not null;index;column:device_platform"`
	DeviceModel    *string        `gorm:"size:100;column:device_model"`
	AppVersion     *string        `gorm:"size:20;column:app_version"`
	IsActive       bool           `gorm:"not null;default:true;column:is_active"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime;column:created_at"`
	LastUsedAt     *time.Time     `gorm:"column:last_used_at"`
}

func (FCMToken) TableName() string {
	return "fcm_tokens"
}
