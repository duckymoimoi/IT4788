package schema

import "time"

// ========================================
// DEVICE  - Xe lan, cang, tram sac
// Slice 7
// ========================================

// DeviceType loai thiet bi ho tro.
type DeviceType string

const (
	DeviceTypeWheelchair   DeviceType = "wheelchair"
	DeviceTypeStretcher    DeviceType = "stretcher"
	DeviceTypeHospitalCart DeviceType = "hospital_cart"
)

// DeviceStatus trang thai thiet bi.
type DeviceStatus string

const (
	DeviceStatusAvailable   DeviceStatus = "available"
	DeviceStatusInUse       DeviceStatus = "in_use"
	DeviceStatusMaintenance DeviceStatus = "maintenance"
)

// BookingStatus trang thai muon/tra thiet bi.
type BookingStatus string

const (
	BookingStatusInUse    BookingStatus = "in_use"
	BookingStatusReturned BookingStatus = "returned"
)

// BrokenReportStatus trang thai bao cao hong.
type BrokenReportStatus string

const (
	BrokenStatusPending  BrokenReportStatus = "pending"
	BrokenStatusResolved BrokenReportStatus = "resolved"
)

// Device thiet bi ho tro (xe lan, cang, ...).
// Moi device gan vao 1 station va co vi tri hien tai (current_poi_id).
// Bang: devices [T25]
type Device struct {
	DeviceID     uint32       `gorm:"primaryKey;autoIncrement;column:device_id" json:"device_id"`
	DeviceCode   string       `gorm:"uniqueIndex;not null;size:20;column:device_code" json:"device_code"`
	DeviceType   DeviceType   `gorm:"not null;size:20;column:device_type" json:"device_type"`
	StationID    *uint32      `gorm:"column:station_id"`
	CurrentPoiID *uint32      `gorm:"column:current_poi_id"`
	Status       DeviceStatus `gorm:"not null;default:available;index;column:status" json:"status"`
	BatteryLevel *int         `gorm:"column:battery_level"`
	IsActive     bool         `gorm:"not null;default:true;column:is_active" json:"is_active,omitempty"`

	// Belongs-to
	Station    *DeviceStation `gorm:"foreignKey:StationID;references:StationID"`
	CurrentPOI *GridPOI       `gorm:"foreignKey:CurrentPoiID;references:PoiID"`
}

func (Device) TableName() string {
	return "devices"
}

// DeviceStation tram de xe lan/cang.
// Moi tram dat tai 1 POI tren grid.
// Bang: device_stations [T26]
type DeviceStation struct {
	StationID   uint32 `gorm:"primaryKey;autoIncrement;column:station_id" json:"station_id"`
	PoiID       uint32 `gorm:"not null;column:poi_id" json:"poi_id"`
	StationName string `gorm:"not null;size:100;column:station_name" json:"station_name"`
	Capacity    int    `gorm:"not null;default:10;column:capacity" json:"capacity"`
	IsActive    bool   `gorm:"not null;default:true;column:is_active" json:"is_active,omitempty"`

	// Belongs-to
	POI *GridPOI `gorm:"foreignKey:PoiID;references:PoiID"`

	// Has-many
	Devices []Device `gorm:"foreignKey:StationID"`
}

func (DeviceStation) TableName() string {
	return "device_stations"
}

// DeviceBooking lich su muon/tra thiet bi.
// Bang: device_bookings [T27]
type DeviceBooking struct {
	BookingID       uint64        `gorm:"primaryKey;autoIncrement;column:booking_id" json:"booking_id"`
	DeviceID        uint32        `gorm:"not null;index;column:device_id" json:"device_id"`
	UserID          uint64        `gorm:"not null;index;column:user_id" json:"user_id"`
	StaffID         *uint64       `gorm:"column:staff_id"`
	Status          BookingStatus `gorm:"not null;default:in_use;index;column:status" json:"status"`
	BorrowedAt      time.Time     `gorm:"not null;autoCreateTime;column:borrowed_at" json:"borrowed_at"`
	ReturnedAt      *time.Time    `gorm:"column:returned_at"`
	ReturnStationID *uint32       `gorm:"column:return_station_id"`

	// Belongs-to
	Device        *Device        `gorm:"foreignKey:DeviceID;references:DeviceID"`
	User          *User          `gorm:"foreignKey:UserID;references:UserID"`
	Staff         *Staff         `gorm:"foreignKey:StaffID;references:StaffID"`
	ReturnStation *DeviceStation `gorm:"foreignKey:ReturnStationID;references:StationID"`
}

func (DeviceBooking) TableName() string {
	return "device_bookings"
}

// DeviceBrokenReport bao cao thiet bi hong.
// Benh nhan bao cao, staff xu ly.
// Bang: device_broken_reports [T28]
type DeviceBrokenReport struct {
	ReportID    uint64             `gorm:"primaryKey;autoIncrement;column:report_id" json:"report_id"`
	DeviceID    uint32             `gorm:"not null;index;column:device_id" json:"device_id"`
	ReportedBy  uint64             `gorm:"not null;column:reported_by" json:"reported_by"`
	Description string             `gorm:"type:text;column:description" json:"description"`
	ImageURL    string             `gorm:"size:255;column:image_url" json:"image_url"`
	Status      BrokenReportStatus `gorm:"not null;default:pending;index;column:status" json:"status"`
	ResolvedBy  *uint64            `gorm:"column:resolved_by"`
	CreatedAt   time.Time          `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	ResolvedAt  *time.Time         `gorm:"column:resolved_at"`

	// Belongs-to
	Device   *Device `gorm:"foreignKey:DeviceID;references:DeviceID"`
	Reporter *User   `gorm:"foreignKey:ReportedBy;references:UserID"`
	Resolver *Staff  `gorm:"foreignKey:ResolvedBy;references:StaffID"`
}

func (DeviceBrokenReport) TableName() string {
	return "device_broken_reports"
}
