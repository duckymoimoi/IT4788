package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/schema"
	"hospital/service"
)

type DeviceHandler struct {
	svc *service.DeviceService
}

func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS
// ========================================

type bookAssetRequest struct {
	AssetID string `json:"asset_id"`
}

type releaseAssetRequest struct {
	AssetID   string `json:"asset_id"`
	StationID string `json:"station_id"`
}

type reportBrokenRequest struct {
	AssetID  string `json:"asset_id"`
	Reason   string `json:"reason"`
	ImageURL string `json:"image_url"`
}

type requestStaffRequest struct {
	AssetID string `json:"asset_id"`
	NodeID  string `json:"node_id"`
	Note    string `json:"note"`
}

// Admin CRUD structs
type addDeviceRequest struct {
	Type          string `json:"type"`
	Status        string `json:"status"`
	CurrentNodeID string `json:"current_node_id"`
	NodeID        string `json:"node_id"`
}

type editDeviceRequest struct {
	ID            uint32 `json:"id"`
	Status        string `json:"status"`
	CurrentNodeID string `json:"current_node_id"`
	NodeID        string `json:"node_id"`
}

type delDeviceRequest struct {
	ID uint32 `json:"id"`
}

// ========================================
// VALID STATUS ENUMS (admin context)
// ========================================
var validAdminDeviceStatuses = map[string]bool{
	"available":   true,
	"maintenance": true,
	"in_use":      true,
	"broken":      true,
}

// ========================================
// PUBLIC ASSET APIs
// ========================================

// GET /api/asset/asset_stations
func (h *DeviceHandler) GetStations(c *gin.Context) {
	stations, err := h.svc.GetStations()
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, stations)
}

// GET /api/asset/find_wheelchairs?node_id=&radius=
func (h *DeviceHandler) GetWheelchairs(c *gin.Context) {
	nodeID := c.Query("node_id")
	if nodeID == "" {
		response.ErrMissingParam(c)
		return
	}

	// radius optional, default 200
	radius := 200
	if rStr := c.Query("radius"); rStr != "" {
		if r, err := strconv.Atoi(rStr); err == nil && r > 0 {
			radius = r
		}
	}

	devices, err := h.svc.FindNearbyWheelchairs(nodeID, radius)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, devices)
}

// GET /api/asset/asset_health?asset_id=
func (h *DeviceHandler) GetDeviceStatus(c *gin.Context) {
	assetID := c.Query("asset_id")
	if assetID == "" {
		response.ErrMissingParam(c)
		return
	}

	result, err := h.svc.GetDeviceHealth(assetID)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, []interface{}{result})
}

// POST /api/asset/book_asset
func (h *DeviceHandler) BookDevice(c *gin.Context) {
	var req bookAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.AssetID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)

	booking, err := h.svc.BookAsset(userID, req.AssetID)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, []interface{}{map[string]interface{}{
		"booking_id": booking.BookingID,
		"asset_id":   req.AssetID,
		"status":     booking.Status,
	}})
}

// POST /api/asset/release_asset
func (h *DeviceHandler) ReleaseDevice(c *gin.Context) {
	var req releaseAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.AssetID == "" || req.StationID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)

	if err := h.svc.ReleaseAsset(userID, req.AssetID, req.StationID); err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, nil)
}

// POST /api/asset/report_broken_asset
func (h *DeviceHandler) ReportBroken(c *gin.Context) {
	var req reportBrokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.AssetID == "" {
		response.ErrMissingParam(c)
		return
	}
	if req.Reason == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)

	reportID, msg, err := h.svc.ReportBrokenAsset(userID, req.AssetID, req.Reason, req.ImageURL)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	c.JSON(200, map[string]interface{}{
		"code":    1000,
		"message": msg,
		"data":    []interface{}{map[string]interface{}{"report_id": reportID}},
	})
}

// POST /api/staff/request_staff
func (h *DeviceHandler) RequestStaff(c *gin.Context) {
	var req requestStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.NodeID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)

	requestID, err := h.svc.RequestStaff(userID, req.AssetID, req.NodeID, req.Note)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	c.JSON(200, map[string]interface{}{
		"code":    1000,
		"message": "Đã điều phối nhân viên hỗ trợ bạn",
		"data":    []interface{}{map[string]interface{}{"request_id": requestID}},
	})
}

// GET /api/asset/track_asset?asset_id=
func (h *DeviceHandler) TrackDevice(c *gin.Context) {
	assetID := c.Query("asset_id")
	if assetID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)

	result, err := h.svc.TrackAsset(userID, assetID)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, []interface{}{result})
}

// ========================================
// ADMIN DEVICE APIS
// ========================================

// GET /api/admin/admin_get_devices
func (h *DeviceHandler) AdminGetDevices(c *gin.Context) {
	devType := c.Query("type")
	devices, err := h.svc.AdminListDevices(devType)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, devices)
}

// POST /api/admin/admin_add_device
func (h *DeviceHandler) AdminAddDevice(c *gin.Context) {
	var req addDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	// Validate required
	if req.Type == "" {
		response.ErrMissingParam(c)
		return
	}
	if req.Status == "" {
		response.ErrMissingParam(c)
		return
	}
	currentNodeID := req.CurrentNodeID
	if currentNodeID == "" {
		currentNodeID = req.NodeID
	}
	if currentNodeID == "" {
		response.ErrMissingParam(c)
		return
	}

	// Validate status enum: admin chỉ được set available hoặc maintenance
	if !validAdminDeviceStatuses[strings.ToLower(req.Status)] {
		response.ErrInvalidValue(c)
		return
	}

	device, err := h.svc.AdminAddDevice(req.Type, req.Status, currentNodeID)
	if err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, map[string]interface{}{"id": device.DeviceID, "device_code": device.DeviceCode})
}

// POST /api/admin/admin_edit_device
func (h *DeviceHandler) AdminEditDevice(c *gin.Context) {
	var req editDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.ID == 0 {
		response.ErrMissingParam(c)
		return
	}

	// Validate status enum if provided
	if req.Status != "" && !validAdminDeviceStatuses[strings.ToLower(req.Status)] {
		response.ErrInvalidValue(c)
		return
	}

	currentNodeID := req.CurrentNodeID
	if currentNodeID == "" {
		currentNodeID = req.NodeID
	}

	if err := h.svc.AdminEditDevice(req.ID, req.Status, currentNodeID); err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, nil)
}

// POST /api/admin/admin_del_device
func (h *DeviceHandler) AdminDelDevice(c *gin.Context) {
	var req delDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.ID == 0 {
		response.ErrMissingParam(c)
		return
	}

	if err := h.svc.AdminDelDevice(req.ID); err != nil {
		h.handleDeviceError(c, err)
		return
	}
	response.Success(c, nil)
}

// ========================================
// ERROR HANDLER
// ========================================

func (h *DeviceHandler) handleDeviceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrDeviceNotFound):
		response.SuccessWithCode(c, 4004, nil)
	case errors.Is(err, service.ErrStationNotFound):
		response.SuccessWithCode(c, 4004, nil)
	case errors.Is(err, service.ErrNodeNotFoundDev):
		response.SuccessWithCode(c, 4004, nil)
	case errors.Is(err, service.ErrDeviceUnavailable):
		response.SuccessWithCode(c, 1009, nil) // Xe không khả dụng (hỏng, đang dùng)
	case errors.Is(err, service.ErrDeviceLimitExceeded):
		response.SuccessWithCode(c, 1010, nil) // Đang mượn xe khác
	case errors.Is(err, service.ErrDeviceOwnership):
		response.SuccessWithCode(c, 1009, nil) // Không phải xe của bạn
	case errors.Is(err, service.ErrDeviceAlreadyDeleted):
		response.SuccessWithCode(c, 4001, nil)
	default:
		response.ErrUnexpected(c)
	}
}

// ========================================
// HELPER — parse DeviceType from string
// ========================================

func toDeviceType(s string) schema.DeviceType {
	switch strings.ToLower(s) {
	case "wheelchair":
		return schema.DeviceTypeWheelchair
	case "stretcher":
		return schema.DeviceTypeStretcher
	case "hospital_cart":
		return schema.DeviceTypeHospitalCart
	default:
		return schema.DeviceType(s)
	}
}
