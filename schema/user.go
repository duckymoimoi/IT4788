package schema

import "time"

type UserType string

const (
	UserTypePatient UserType = "patient"
	UserTypeStaff   UserType = "staff"
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusPending UserStatus = "pending"
	UserStatusBanned  UserStatus = "banned"
	UserStatusDeleted UserStatus = "deleted"
)

type Gender string

const (
	GenderMale   Gender = "M"
	GenderFemale Gender = "F"
	GenderOther  Gender = "O"
)

// User tai khoan nguoi dung. Bang: users [T01]
type User struct {
	UserID       uint64     `gorm:"primaryKey;autoIncrement;column:user_id" json:"user_id"`
	PhoneNumber  string     `gorm:"uniqueIndex;not null;size:15;column:phone_number" json:"phone_number"`
	PasswordHash string     `gorm:"not null;size:255;column:password_hash" json:"-"`
	FullName     string     `gorm:"not null;size:100;column:full_name" json:"full_name"`
	UserType     UserType   `gorm:"not null;default:patient;column:user_type" json:"user_type"`
	DateOfBirth  *time.Time `gorm:"column:date_of_birth" json:"date_of_birth,omitempty"`
	Gender       *Gender    `gorm:"column:gender" json:"gender,omitempty"`
	AvatarURL    *string    `gorm:"size:500;column:avatar_url" json:"avatar_url,omitempty"`
	Status       UserStatus `gorm:"not null;default:active;column:status" json:"status"`
	CreatedAt    time.Time  `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`

	// Relations
	Setting   *UserSetting `gorm:"foreignKey:UserID" json:"setting,omitempty"`
	FCMTokens []FCMToken   `gorm:"foreignKey:UserID" json:"-"`
	Staff     *Staff       `gorm:"foreignKey:UserID" json:"staff,omitempty"`
}

func (User) TableName() string { return "users" }

type OTPType string

const (
	OTPTypeSignup        OTPType = "signup"
	OTPTypeResetPassword OTPType = "reset_password"
	OTPTypeChangePhone   OTPType = "change_phone"
)

// OTPCode ma xac thuc. Bang: otp_codes [T02]
type OTPCode struct {
	OTPID       uint64    `gorm:"primaryKey;autoIncrement;column:otp_id" json:"otp_id"`
	PhoneNumber string    `gorm:"not null;size:15;index:idx_phone_type;column:phone_number" json:"phone_number"`
	OTPCode     string    `gorm:"not null;size:8;column:otp_code" json:"-"`
	Type        OTPType   `gorm:"not null;index:idx_phone_type;column:type" json:"type"`
	ExpiredAt   time.Time `gorm:"not null;column:expired_at" json:"expired_at"`
	IsUsed      bool      `gorm:"not null;default:false;column:is_used" json:"is_used"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
}

func (OTPCode) TableName() string { return "otp_codes" }

type TravelModeEnum string

const (
	TravelModeWalk       TravelModeEnum = "walk"
	TravelModeWheelchair TravelModeEnum = "wheelchair"
	TravelModeStretcher  TravelModeEnum = "stretcher"
)

// UserSetting cau hinh ca nhan. Bang: user_settings [T03]
type UserSetting struct {
	SettingID            uint64         `gorm:"primaryKey;autoIncrement;column:setting_id" json:"setting_id"`
	UserID               uint64         `gorm:"uniqueIndex;not null;column:user_id" json:"user_id"`
	VoiceGuidanceEnabled bool           `gorm:"not null;default:true;column:voice_guidance_enabled" json:"voice_guidance_enabled"`
	NotificationEnabled  bool           `gorm:"not null;default:true;column:notification_enabled" json:"notification_enabled"`
	TravelMode           TravelModeEnum `gorm:"not null;default:walk;column:travel_mode" json:"travel_mode"`
	Language             string         `gorm:"not null;default:vi;size:10;column:language" json:"language"`
	Theme                string         `gorm:"not null;default:light;size:20;column:theme" json:"theme"`
	UpdatedAt            *time.Time     `gorm:"column:updated_at" json:"updated_at,omitempty"`
}

func (UserSetting) TableName() string { return "user_settings" }

type DevicePlatform string

const (
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformAndroid DevicePlatform = "android"
)

// FCMToken firebase cloud messaging token. Bang: fcm_tokens [T04]
type FCMToken struct {
	TokenID        uint64         `gorm:"primaryKey;autoIncrement;column:token_id" json:"token_id"`
	UserID         uint64         `gorm:"not null;index;column:user_id" json:"user_id"`
	FCMToken       string         `gorm:"not null;uniqueIndex;size:500;column:fcm_token" json:"fcm_token"`
	DevicePlatform DevicePlatform `gorm:"not null;index;column:device_platform" json:"device_platform"`
	DeviceModel    *string        `gorm:"size:100;column:device_model" json:"device_model,omitempty"`
	AppVersion     *string        `gorm:"size:20;column:app_version" json:"app_version,omitempty"`
	IsActive       bool           `gorm:"not null;default:true;column:is_active" json:"is_active"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	LastUsedAt     *time.Time     `gorm:"column:last_used_at" json:"last_used_at,omitempty"`
}

func (FCMToken) TableName() string { return "fcm_tokens" }
