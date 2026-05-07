package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"hospital/repository"
	"hospital/schema"
)

// ========================================
// ERRORS
// ========================================

var (
	ErrDeviceNotFound      = errors.New("device not found")
	ErrStationNotFound     = errors.New("station not found")
	ErrNodeNotFoundDev     = errors.New("node not found")
	ErrDeviceUnavailable   = errors.New("device unavailable")
	ErrDeviceLimitExceeded = errors.New("already borrowing a device")
	ErrDeviceOwnership     = errors.New("device does not belong to you")
	ErrDeviceAlreadyDeleted = errors.New("device already deleted")
)

type DeviceService struct {
	repo *repository.DeviceRepo
}

func NewDeviceService(repo *repository.DeviceRepo) *DeviceService {
	return &DeviceService{repo: repo}
}

// ========================================
// 1. GetStations — GET /api/asset/asset_stations
// ========================================

type StationItem struct {
	StationID           uint32 `json:"station_id"`
	StationName         string `json:"station_name"`
	Capacity            int    `json:"capacity"`
	AvailableWheelchairs int   `json:"available_wheelchairs"`
}

func (s *DeviceService) GetStations() ([]StationItem, error) {
	stations, err := s.repo.FindAllStations()
	if err != nil {
		return nil, err
	}
	items := make([]StationItem, len(stations))
	for i, st := range stations {
		count, _ := s.repo.CountAvailableByStation(st.StationID, schema.DeviceTypeWheelchair)
		items[i] = StationItem{
			StationID:           st.StationID,
			StationName:         st.StationName,
			Capacity:            st.Capacity,
			AvailableWheelchairs: int(count),
		}
	}
	return items, nil
}

// ========================================
// 2. FindNearbyWheelchairs — GET /api/asset/find_wheelchairs
// ========================================

type WheelchairItem struct {
	AssetID      string  `json:"asset_id"`     // device_code
	DeviceID     uint32  `json:"device_id"`
	Status       string  `json:"status"`
	BatteryLevel *int    `json:"battery_level,omitempty"`
	Distance     float64 `json:"distance,omitempty"`
}

func (s *DeviceService) FindNearbyWheelchairs(nodeID string, radius int) ([]WheelchairItem, error) {
	// Kiểm tra node tồn tại
	exists, err := s.repo.NodeExists(nodeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNodeNotFoundDev
	}

	devices, err := s.repo.FindAvailableDevices(schema.DeviceTypeWheelchair)
	if err != nil {
		return nil, err
	}

	// Filter theo bán kính nếu cần (hiện tại trả tất cả available trong cùng floor)
	// Với grid map, "bán kính" = số bước đi. Khi chưa có pathfinder, trả tất cả.
	items := make([]WheelchairItem, 0, len(devices))
	for _, d := range devices {
		items = append(items, WheelchairItem{
			AssetID:      d.DeviceCode,
			DeviceID:     d.DeviceID,
			Status:       string(d.Status),
			BatteryLevel: d.BatteryLevel,
		})
	}
	return items, nil
}

// ========================================
// 3. GetDeviceHealth — GET /api/asset/asset_health
// ========================================

type DeviceHealthItem struct {
	AssetID      string `json:"asset_id"`
	Condition    string `json:"condition"`
	BatteryLevel string `json:"battery_level"`
	Status       string `json:"status"`
}

func (s *DeviceService) GetDeviceHealth(assetID string) (*DeviceHealthItem, error) {
	device, err := s.repo.FindDeviceByCode(assetID)
	if err != nil || device == nil {
		return nil, ErrDeviceNotFound
	}

	battStr := "N/A"
	if device.BatteryLevel != nil {
		battStr = fmt.Sprintf("%d%%", *device.BatteryLevel)
	}

	condition := string(device.Status)

	return &DeviceHealthItem{
		AssetID:      device.DeviceCode,
		Condition:    condition,
		BatteryLevel: battStr,
		Status:       string(device.Status),
	}, nil
}

// ========================================
// 4. BookAsset — POST /api/asset/book_asset
// ========================================

func (s *DeviceService) BookAsset(userID uint64, assetID string) (*schema.DeviceBooking, error) {
	// Tìm device trước — nếu không tồn tại trả 4004 luôn
	device, err := s.repo.FindDeviceByCode(assetID)
	if err != nil || device == nil {
		return nil, ErrDeviceNotFound // 4004
	}

	// Kiểm tra user đang mượn xe khác chưa
	activeBooking, _ := s.repo.FindActiveBookingByUser(userID)
	if activeBooking != nil {
		return nil, ErrDeviceLimitExceeded // 1010
	}

	// Kiểm tra trạng thái: chỉ available mới cho mượn
	if device.Status != schema.DeviceStatusAvailable {
		return nil, ErrDeviceUnavailable // 1009 (xe hỏng, đang dùng, bảo trì)
	}

	booking := &schema.DeviceBooking{
		DeviceID: device.DeviceID,
		UserID:   userID,
		Status:   schema.BookingStatusInUse,
	}

	if err := s.repo.CreateBookingTx(booking); err != nil {
		return nil, err
	}
	return booking, nil
}

// ========================================
// 5. ReleaseAsset — POST /api/asset/release_asset
// ========================================

func (s *DeviceService) ReleaseAsset(userID uint64, assetID, stationCode string) error {
	// Tìm booking đang active của user
	booking, err := s.repo.FindActiveBookingByUser(userID)
	if err != nil || booking == nil {
		return ErrDeviceOwnership // không có booking nào → không phải xe của bạn
	}

	// Verify đúng xe user đang mượn
	if booking.Device != nil && booking.Device.DeviceCode != assetID {
		return ErrDeviceOwnership // 1009
	}

	// Kiểm tra station tồn tại (bằng station_name hoặc ID)
	station, err := s.repo.FindStationByCode(stationCode)
	if err != nil || station == nil {
		return ErrStationNotFound // 4004
	}

	now := time.Now()
	return s.repo.ReturnDeviceTx(booking.BookingID, booking.DeviceID, station.StationID, now)
}

// ========================================
// 6. ReportBrokenAsset — POST /api/asset/report_broken_asset
// Returns: reportID, message, error
// ========================================

func (s *DeviceService) ReportBrokenAsset(userID uint64, assetID, reason, imgURL string) (uint64, string, error) {
	device, err := s.repo.FindDeviceByCode(assetID)
	if err != nil || device == nil {
		return 0, "", ErrDeviceNotFound // 4004
	}

	// Xe đã broken → vẫn tạo report nhưng trả message khác
	if device.Status == schema.DeviceStatusMaintenance {
		report := &schema.DeviceBrokenReport{
			DeviceID:    device.DeviceID,
			ReportedBy:  userID,
			Description: reason,
			ImageURL:    imgURL,
			Status:      schema.BrokenStatusPending,
		}
		_ = s.repo.CreateBrokenReportTx(report)
		return report.ReportID, "Tình trạng hỏng đã được ghi nhận trước đó, cảm ơn bạn đã báo cáo thêm", nil
	}

	report := &schema.DeviceBrokenReport{
		DeviceID:    device.DeviceID,
		ReportedBy:  userID,
		Description: reason,
		ImageURL:    imgURL,
		Status:      schema.BrokenStatusPending,
	}
	if err := s.repo.CreateBrokenReportTx(report); err != nil {
		return 0, "", err
	}
	return report.ReportID, "Đã ghi nhận thiết bị hỏng, nhân viên sẽ xử lý sớm", nil
}

// ========================================
// 7. RequestStaff — POST /api/staff/request_staff
// ========================================

func (s *DeviceService) RequestStaff(userID uint64, assetID, nodeID, note string) (uint64, error) {
	// Nếu có assetID → kiểm tra ownership
	if assetID != "" {
		device, err := s.repo.FindDeviceByCode(assetID)
		if err != nil || device == nil {
			return 0, ErrDeviceNotFound // 4004
		}

		// Kiểm tra user có đang mượn xe này không
		booking, _ := s.repo.FindActiveBookingByUser(userID)
		if booking == nil || booking.Device == nil || booking.Device.DeviceCode != assetID {
			return 0, ErrDeviceOwnership // 1009
		}
	}

	// Tạo staff request record
	reqID, err := s.repo.CreateStaffRequest(userID, assetID, nodeID, note)
	if err != nil {
		return 0, err
	}
	return reqID, nil
}

// ========================================
// 8. TrackAsset — GET /api/asset/track_asset
// ========================================

type AssetTrackItem struct {
	AssetID       string  `json:"asset_id"`
	MovingStatus  string  `json:"moving_status"`
	CurrentNodeID *uint32 `json:"current_node_id,omitempty"`
	Status        string  `json:"status"`
}

func (s *DeviceService) TrackAsset(userID uint64, assetID string) (*AssetTrackItem, error) {
	device, err := s.repo.FindDeviceByCode(assetID)
	if err != nil || device == nil {
		return nil, ErrDeviceNotFound // 4004
	}

	// Check ownership: chỉ user đang mượn hoặc staff/admin mới track được
	// Kiểm tra booking active
	booking, _ := s.repo.FindActiveBookingByUser(userID)
	isOwner := booking != nil && booking.Device != nil && booking.Device.DeviceCode == assetID

	// Nếu device đang in_use và user không phải owner → 1009
	if device.Status == schema.DeviceStatusInUse && !isOwner {
		return nil, ErrDeviceOwnership
	}

	movingStatus := "stationary"
	if device.Status == schema.DeviceStatusInUse {
		movingStatus = "moving"
	}

	return &AssetTrackItem{
		AssetID:       device.DeviceCode,
		MovingStatus:  movingStatus,
		CurrentNodeID: device.CurrentPoiID,
		Status:        string(device.Status),
	}, nil
}

// ========================================
// ADMIN DEVICE CRUD
// ========================================

func (s *DeviceService) AdminAddDevice(devType, status, currentNodeID string) (*schema.Device, error) {
	// Generate device code: type prefix + timestamp
	code := strings.ToUpper(string([]rune(devType)[:2])) + fmt.Sprintf("-%d", time.Now().UnixNano()%100000)

	device := &schema.Device{
		DeviceCode: code,
		DeviceType: schema.DeviceType(devType),
		Status:     schema.DeviceStatus(strings.ToLower(status)),
		IsActive:   true,
	}
	if err := s.repo.CreateDevice(device); err != nil {
		return nil, err
	}
	return device, nil
}

func (s *DeviceService) AdminEditDevice(deviceID uint32, status string) error {
	device, err := s.repo.FindDeviceByID(deviceID)
	if err != nil || device == nil {
		return ErrDeviceNotFound // → 4001 in handler
	}
	updates := map[string]interface{}{}
	if status != "" {
		updates["status"] = strings.ToLower(status)
	}
	return s.repo.UpdateDevice(deviceID, updates)
}

func (s *DeviceService) AdminDelDevice(deviceID uint32) error {
	device, err := s.repo.FindDeviceByID(deviceID)
	if err != nil || device == nil {
		return ErrDeviceAlreadyDeleted // 4001
	}
	if !device.IsActive {
		return ErrDeviceAlreadyDeleted // 4001
	}
	return s.repo.DeactivateDevice(deviceID)
}