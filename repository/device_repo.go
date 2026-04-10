package repository

import (
	"hospital/schema"
	"time"

	"gorm.io/gorm"
)

// DeviceRepo xu ly truy van database cho module Device.
// Bao gom: devices, device_stations, device_bookings, device_broken_reports.
type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

// ========================================
// DEVICE STATIONS
// ========================================

// FindAllStations lay danh sach tat ca cac tram de thiet bi dang hoat dong.
func (r *DeviceRepo) FindAllStations() ([]schema.DeviceStation, error) {
	var stations []schema.DeviceStation
	err := r.db.Where("is_active = ?", true).
		Preload("POI").
		Find(&stations).Error
	return stations, err
}

// ========================================
// DEVICES
// ========================================

// FindAvailableDevices lay danh sach thiet bi dang ranh (theo loai).
func (r *DeviceRepo) FindAvailableDevices(devType schema.DeviceType) ([]schema.Device, error) {
	var devices []schema.Device
	err := r.db.Where("device_type = ? AND status = ? AND is_active = ?", devType, schema.DeviceStatusAvailable, true).
		Preload("Station").
		Preload("CurrentPOI").
		Find(&devices).Error
	return devices, err
}

// FindDeviceByID tim thiet bi theo ID.
func (r *DeviceRepo) FindDeviceByID(deviceID uint32) (*schema.Device, error) {
	var device schema.Device
	err := r.db.Where("device_id = ?", deviceID).
		Preload("CurrentPOI").
		First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// ========================================
// DEVICE BOOKINGS
// ========================================

// FindActiveBookingByUser kiem tra xem user co dang muon thiet bi nao khong.
func (r *DeviceRepo) FindActiveBookingByUser(userID uint64) (*schema.DeviceBooking, error) {
	var booking schema.DeviceBooking
	err := r.db.Where("user_id = ? AND status = ?", userID, schema.BookingStatusInUse).
		Preload("Device").
		First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// FindBookingByID tim thong tin muon tra theo ID.
func (r *DeviceRepo) FindBookingByID(bookingID uint64) (*schema.DeviceBooking, error) {
	var booking schema.DeviceBooking
	err := r.db.Where("booking_id = ?", bookingID).First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// CreateBookingTx thuc hien giao dich muon thiet bi.
// 1. Tao lich su muon
// 2. Cap nhat trang thai thiet bi -> in_use
func (r *DeviceRepo) CreateBookingTx(booking *schema.DeviceBooking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Buoc 1: Tao record muon
		if err := tx.Create(booking).Error; err != nil {
			return err
		}

		// Buoc 2: Cap nhat thiet bi thanh dang su dung (va tam thoi xoa khoi Station)
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

// ReturnDeviceTx thuc hien giao dich tra thiet bi.
// 1. Cap nhat lich su muon -> returned
// 2. Cap nhat trang thai thiet bi -> available, gan lai vao station dich.
func (r *DeviceRepo) ReturnDeviceTx(bookingID uint64, deviceID uint32, returnStationID uint32, returnedAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Buoc 1: Chot thoi gian tra
		if err := tx.Model(&schema.DeviceBooking{}).
			Where("booking_id = ?", bookingID).
			Updates(map[string]interface{}{
				"status":            schema.BookingStatusReturned,
				"returned_at":       returnedAt,
				"return_station_id": returnStationID,
			}).Error; err != nil {
			return err
		}

		// Buoc 2: Gan xe vao tram moi, doi status thanh available
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

// CreateBrokenReportTx tao bao cao hong va khoa thiet bi lai (chuyen sang maintenance).
func (r *DeviceRepo) CreateBrokenReportTx(report *schema.DeviceBrokenReport) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Buoc 1: Tao report
		if err := tx.Create(report).Error; err != nil {
			return err
		}

		// Buoc 2: Doi status thiet bi thanh dang bao tri
		if err := tx.Model(&schema.Device{}).
			Where("device_id = ?", report.DeviceID).
			Update("status", schema.DeviceStatusMaintenance).Error; err != nil {
			return err
		}

		return nil
	})
}