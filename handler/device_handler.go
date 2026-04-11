package handler

import (
	"hospital/middleware"
	response "hospital/pkg" // Package quy thuan cua Leader
	"hospital/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	svc *service.DeviceService
}

func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS (Chứa dữ liệu Client gửi lên)
// ========================================
type bookRequest struct {
	DeviceID uint32 `json:"device_id" binding:"required"`
}

type releaseRequest struct {
	ReturnStationID uint32 `json:"return_station_id" binding:"required"`
}

type reportBrokenRequest struct {
	DeviceID    uint32 `json:"device_id" binding:"required"`
	Description string `json:"description" binding:"required"`
	ImageURL    string `json:"image_url"`
}

type requestStaffRequest struct {
	PoiID uint32 `json:"poi_id" binding:"required"`
	Note  string `json:"note"`
}

// ========================================
// API HANDLERS
// ========================================

// [87] GET /api/device/stations
func (h *DeviceHandler) GetStations(c *gin.Context) {
	stations, err := h.svc.GetStations()
	if err != nil {
		response.Error(c, 5000, err.Error()) // Lỗi chung
		return
	}
	response.Success(c, stations)
}

// [83] GET /api/device/wheelchairs
func (h *DeviceHandler) GetWheelchairs(c *gin.Context) {
	devices, err := h.svc.GetAvailableWheelchairs()
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}
	response.Success(c, devices)
}

// [88] GET /api/device/status/:id
func (h *DeviceHandler) GetDeviceStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, 2003, "ID thiết bị không hợp lệ") // Invalid parameter
		return
	}

	device, err := h.svc.GetDeviceStatus(uint32(id))
	if err != nil {
		response.Error(c, 8001, "Không tìm thấy thiết bị") // Asset not found
		return
	}
	response.Success(c, device)
}

// [84] POST /api/device/book
func (h *DeviceHandler) BookDevice(c *gin.Context) {
	var req bookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c) // Lỗi 2005
		return
	}

	userID := middleware.GetUserID(c) // Lấy ID người dùng đang đăng nhập (Rule 5)

	err := h.svc.BookDevice(userID, req.DeviceID)
	if err != nil {
		response.Error(c, 1010, err.Error()) // Lỗi logic nghiệp vụ (Limit exceeded hoặc Asset bận)
		return
	}

	response.Success(c, map[string]string{"message": "Mượn thiết bị thành công"})
}

// [85] POST /api/device/release
func (h *DeviceHandler) ReleaseDevice(c *gin.Context) {
	var req releaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)

	err := h.svc.ReleaseDevice(userID, req.ReturnStationID)
	if err != nil {
		response.Error(c, 4000, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "Trả thiết bị thành công"})
}

// [89] POST /api/device/report_broken
func (h *DeviceHandler) ReportBroken(c *gin.Context) {
	var req reportBrokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)

	err := h.svc.ReportBrokenDevice(userID, req.DeviceID, req.Description, req.ImageURL)
	if err != nil {
		response.Error(c, 5000, "Không thể tạo báo cáo lỗi lúc này")
		return
	}

	response.Success(c, map[string]string{"message": "Đã ghi nhận thiết bị hỏng"})
}

// [86] POST /api/device/request_staff
func (h *DeviceHandler) RequestStaff(c *gin.Context) {
	var req requestStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)

	err := h.svc.RequestStaffSupport(userID, req.PoiID, req.Note)
	if err != nil {
		response.Error(c, 5000, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "Đã gửi yêu cầu nhân viên hỗ trợ"})
}

// [90] GET /api/device/track/:id
func (h *DeviceHandler) TrackDevice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, 2003, "ID thiết bị không hợp lệ")
		return
	}

	poi, err := h.svc.TrackDeviceLocation(uint32(id))
	if err != nil {
		response.Error(c, 8001, err.Error())
		return
	}

	response.Success(c, poi)
}