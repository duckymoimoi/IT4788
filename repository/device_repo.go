package repository

import (
	"hospital/schema"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// DeviceRepo xu ly truy van database cho module Device.
type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

// ========================================
// DEVICE STATIONS
// ========================================

func (r *DeviceRepo) FindAllStations() ([]schema.DeviceStation, error) {
	var stations []schema.DeviceStation
	err := r.db.Where("is_active = ?", true).
		Preload("POI").
		Find(&stations).Error
	return stations, err
}

// CountAvailableByStation đếm thiết bị available tại 1 station theo loại.
func (r *DeviceRepo) CountAvailableByStation(stationID uint32, devType schema.DeviceType) (int64, error) {
	var count int64
	err := r.db.Model(&schema.Device{}).
		Where("station_id = ? AND device_type = ? AND status = ? AND is_active = ?",
			stationID, devType, schema.DeviceStatusAvailable, true).
		Count(&count).Error
	return count, err
}

// FindStationByCode tìm station theo tên hoặc station_id dạng string.
func (r *DeviceRepo) FindStationByCode(code string) (*schema.DeviceStation, error) {
	var station schema.DeviceStation
	// Thử match station_name trước
	err := r.db.Where("station_name = ? AND is_active = ?", code, true).
		First(&station).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &station, err
}

// ========================================
// DEVICES
// ========================================

func (r *DeviceRepo) FindAvailableDevices(devType schema.DeviceType) ([]schema.Device, error) {
	var devices []schema.Device
	err := r.db.Where("device_type = ? AND status = ? AND is_active = ?", devType, schema.DeviceStatusAvailable, true).
		Preload("Station").
		Preload("CurrentPOI").
		Find(&devices).Error
	return devices, err
}

// FindAllDevices returns all active devices for the admin panel.
func (r *DeviceRepo) FindAllDevices(devType schema.DeviceType) ([]schema.Device, error) {
	var devices []schema.Device
	q := r.db.Where("is_active = ?", true)
	if devType != "" {
		q = q.Where("device_type = ?", devType)
	}
	err := q.Preload("Station").
		Preload("CurrentPOI").
		Order("device_id ASC").
		Find(&devices).Error
	return devices, err
}

// FindDeviceByID tìm thiết bị theo numeric ID.
func (r *DeviceRepo) FindDeviceByID(deviceID uint32) (*schema.Device, error) {
	var device schema.Device
	err := r.db.Where("device_id = ? AND is_active = ?", deviceID, true).
		Preload("CurrentPOI").
		First(&device).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &device, err
}

// ResolvePOIIdentifier supports POI code, poi_id, or grid_location input.
func (r *DeviceRepo) ResolvePOIIdentifier(identifier string) (*schema.GridPOI, error) {
	var poi schema.GridPOI
	err := r.db.Where("poi_code = ? AND is_active = ?", identifier, true).First(&poi).Error
	if err == nil {
		return &poi, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if n, convErr := strconv.Atoi(identifier); convErr == nil {
		err = r.db.Where("(poi_id = ? OR grid_location = ?) AND is_active = ?", n, n, true).First(&poi).Error
		if err == nil {
			return &poi, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	return nil, nil
}

func (r *DeviceRepo) FindStationByPOIID(poiID uint32) (*schema.DeviceStation, error) {
	var station schema.DeviceStation
	err := r.db.Where("poi_id = ? AND is_active = ?", poiID, true).First(&station).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &station, err
}

// FindDeviceByCode tìm thiết bị theo device_code (string identifier).
func (r *DeviceRepo) FindDeviceByCode(code string) (*schema.Device, error) {
	var device schema.Device
	err := r.db.Where("device_code = ?", code).
		Preload("CurrentPOI").
		First(&device).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &device, err
}

// NodeExists kiểm tra POI/node có tồn tại không (dùng poi_code).
func (r *DeviceRepo) NodeExists(poiCode string) (bool, error) {
	var count int64
	err := r.db.Model(&schema.GridPOI{}).
		Where("poi_code = ? AND is_active = ?", poiCode, true).
		Count(&count).Error
	return count > 0, err
}

// CreateDevice tạo thiết bị mới.
func (r *DeviceRepo) CreateDevice(device *schema.Device) error {
	return r.db.Create(device).Error
}

// UpdateDevice cập nhật thiết bị.
func (r *DeviceRepo) UpdateDevice(deviceID uint32, updates map[string]interface{}) error {
	return r.db.Model(&schema.Device{}).
		Where("device_id = ?", deviceID).
		Updates(updates).Error
}

// DeactivateDevice soft-delete thiết bị.
func (r *DeviceRepo) DeactivateDevice(deviceID uint32) error {
	return r.db.Model(&schema.Device{}).
		Where("device_id = ?", deviceID).
		Update("is_active", false).Error
}

// ========================================
// DEVICE BOOKINGS
// ========================================

func (r *DeviceRepo) FindActiveBookingByUser(userID uint64) (*schema.DeviceBooking, error) {
	var booking schema.DeviceBooking
	err := r.db.Where("user_id = ? AND status = ?", userID, schema.BookingStatusInUse).
		Preload("Device").
		First(&booking).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *DeviceRepo) FindBookingByID(bookingID uint64) (*schema.DeviceBooking, error) {
	var booking schema.DeviceBooking
	err := r.db.Where("booking_id = ?", bookingID).First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// CreateBookingTx tạo booking và đổi status thiết bị sang in_use.
func (r *DeviceRepo) CreateBookingTx(booking *schema.DeviceBooking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(booking).Error; err != nil {
			return err
		}
		if err := tx.Model(&schema.Device{}).
			Where("device_id = ?", booking.DeviceID).
			Updates(map[string]interface{}{
				"status":     schema.DeviceStatusInUse,
				"station_id": nil,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

// ReturnDeviceTx trả thiết bị và cập nhật booking.
func (r *DeviceRepo) ReturnDeviceTx(bookingID uint64, deviceID uint32, returnStationID uint32, returnedAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&schema.DeviceBooking{}).
			Where("booking_id = ?", bookingID).
			Updates(map[string]interface{}{
				"status":            schema.BookingStatusReturned,
				"returned_at":       returnedAt,
				"return_station_id": returnStationID,
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&schema.Device{}).
			Where("device_id = ?", deviceID).
			Updates(map[string]interface{}{
				"status":     schema.DeviceStatusAvailable,
				"station_id": returnStationID,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

// ========================================
// BROKEN REPORTS
// ========================================

func (r *DeviceRepo) CreateBrokenReportTx(report *schema.DeviceBrokenReport) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}
		if err := tx.Model(&schema.Device{}).
			Where("device_id = ?", report.DeviceID).
			Update("status", schema.DeviceStatusMaintenance).Error; err != nil {
			return err
		}
		return nil
	})
}

// ========================================
// STAFF REQUEST
// ========================================

// StaffRequest schema nhẹ, lưu inline (không cần bảng riêng phức tạp).
// Dùng bảng support_requests nếu có, hoặc tạo record đơn giản.
// Trả về request_id để test pass.
func (r *DeviceRepo) CreateStaffRequest(userID uint64, assetID, nodeID, note string) (uint64, error) {
	// Tạo record đơn giản vào bảng device_staff_requests
	// Nếu chưa có bảng → dùng raw insert
	type StaffRequest struct {
		RequestID uint64 `gorm:"primaryKey;autoIncrement;column:request_id"`
		UserID    uint64 `gorm:"column:user_id"`
		AssetID   string `gorm:"column:asset_id;size:30"`
		NodeID    string `gorm:"column:node_id;size:30"`
		Note      string `gorm:"column:note;type:text"`
		CreatedAt time.Time
	}

	req := &StaffRequest{
		UserID:    userID,
		AssetID:   assetID,
		NodeID:    nodeID,
		Note:      note,
		CreatedAt: time.Now(),
	}

	err := r.db.Table("device_staff_requests").Create(req).Error
	if err != nil {
		// Nếu bảng chưa tồn tại → tự tạo và thử lại
		r.db.Exec(`CREATE TABLE IF NOT EXISTS device_staff_requests (
			request_id BIGSERIAL PRIMARY KEY,
			user_id BIGINT,
			asset_id VARCHAR(30),
			node_id VARCHAR(30),
			note TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`)
		err = r.db.Table("device_staff_requests").Create(req).Error
	}
	return req.RequestID, err
}
