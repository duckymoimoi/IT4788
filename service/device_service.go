package service

import (
	"errors"
	"hospital/repository"
	"hospital/schema"
	"time"
)

type DeviceService struct {
	repo *repository.DeviceRepo
}

func NewDeviceService(repo *repository.DeviceRepo) *DeviceService {
	return &DeviceService{repo: repo}
}

// 1. Get Stations (#87)
func (s *DeviceService) GetStations() ([]schema.DeviceStation, error) {
	return s.repo.FindAllStations()
}

// 2. Get Wheelchairs (#83)
func (s *DeviceService) GetAvailableWheelchairs() ([]schema.Device, error) {
	return s.repo.FindAvailableDevices(schema.DeviceTypeWheelchair)
}

// 3. Get Status (#88)
func (s *DeviceService) GetDeviceStatus(deviceID uint32) (*schema.Device, error) {
	return s.repo.FindDeviceByID(deviceID)
}

// 4. Book Device (#84)
func (s *DeviceService) BookDevice(userID uint64, deviceID uint32) error {
	// Kiem tra xem user co dang muon xe khac chua
	activeBooking, _ := s.repo.FindActiveBookingByUser(userID)
	if activeBooking != nil {
		return errors.New("ban dang muon mot thiet bi khac, vui long tra truoc khi muon moi")
	}

	// Kiem tra trang thai xe
	device, err := s.repo.FindDeviceByID(deviceID)
	if err != nil {
		return errors.New("thiet bi khong ton tai")
	}
	if device.Status != schema.DeviceStatusAvailable {
		return errors.New("thiet bi nay hien khong kha dung")
	}

	// Tao record muon
	booking := &schema.DeviceBooking{
		DeviceID: deviceID,
		UserID:   userID,
		Status:   schema.BookingStatusInUse,
	}

	return s.repo.CreateBookingTx(booking)
}

// 5. Release Device (#85)
func (s *DeviceService) ReleaseDevice(userID uint64, returnStationID uint32) error {
	// Tim thong tin xe dang muon cua user
	booking, err := s.repo.FindActiveBookingByUser(userID)
	if err != nil || booking == nil {
		return errors.New("khong tim thay thiet bi dang muon")
	}

	now := time.Now()
	return s.repo.ReturnDeviceTx(booking.BookingID, booking.DeviceID, returnStationID, now)
}

// 6. Report Broken (#89)
func (s *DeviceService) ReportBrokenDevice(userID uint64, deviceID uint32, desc, imgURL string) error {
	report := &schema.DeviceBrokenReport{
		DeviceID:    deviceID,
		ReportedBy:  userID,
		Description: desc,
		ImageURL:    imgURL,
		Status:      schema.BrokenStatusPending,
	}
	return s.repo.CreateBrokenReportTx(report)
}

// 7. Request Staff (#86) - Stub cho logic goi nhan vien
func (s *DeviceService) RequestStaffSupport(userID uint64, poiID uint32, note string) error {
	// Logic thate gui thong bao Firebase (Person C lam) hoac tao record ho tro
	// Hien tai return nil de handler co the goi dung luong
	return nil
}

// 8. Track Device (#90)
func (s *DeviceService) TrackDeviceLocation(deviceID uint32) (*schema.GridPOI, error) {
	device, err := s.repo.FindDeviceByID(deviceID)
	if err != nil {
		return nil, errors.New("khong tim thay thiet bi")
	}
	if device.CurrentPOI == nil {
		return nil, errors.New("khong xac dinh duoc vi tri thiet bi luc nay")
	}
	return device.CurrentPOI, nil
}