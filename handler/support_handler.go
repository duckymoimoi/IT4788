package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
)

// SupportHandler xu ly HTTP request cho module SOS.
// Person E so huu file nay.
type SupportHandler struct {
	sosSvc *service.SOSService
}

func NewSupportHandler(sosSvc *service.SOSService) *SupportHandler {
	return &SupportHandler{sosSvc: sosSvc}
}

// ========================================
// REQUEST STRUCTS
// ========================================

type createSOSRequest struct {
	GridLocation int     `json:"grid_location" binding:"required"`
	PosX         float64 `json:"pos_x"`
	PosY         float64 `json:"pos_y"`
	Note         string  `json:"note"`
}

type respondSOSRequest struct {
	SosID uint64 `json:"sos_id" binding:"required"`
}

type resolveSOSRequest struct {
	SosID uint64 `json:"sos_id" binding:"required"`
}

// ========================================
// [96] POST /api/sos/create
// ========================================

// CreateSOS benh nhan gui tin hieu khan cap.
func (h *SupportHandler) CreateSOS(c *gin.Context) {
	var req createSOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	sos, err := h.sosSvc.CreateSOS(userID, req.GridLocation, req.PosX, req.PosY, req.Note)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, sos)
}

// ========================================
// [97] GET /api/sos/get_list?page=&limit=
// ========================================

// GetSOSList danh sach SOS (staff xem).
func (h *SupportHandler) GetSOSList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, total, err := h.sosSvc.GetSOSList(page, limit)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{
		"sos_list": list,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// ========================================
// [98] POST /api/sos/respond
// ========================================

// RespondSOS staff nhan xu ly SOS.
func (h *SupportHandler) RespondSOS(c *gin.Context) {
	var req respondSOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	if err := h.sosSvc.RespondSOS(req.SosID, userID); err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"responded": true})
}

// ========================================
// [99] POST /api/sos/resolve
// ========================================

// ResolveSOS dong vu viec SOS.
func (h *SupportHandler) ResolveSOS(c *gin.Context) {
	var req resolveSOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	if err := h.sosSvc.ResolveSOS(req.SosID, userID); err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"resolved": true})
}

// ========================================
// [100] GET /api/sos/get_detail?sos_id=
// ========================================

// GetSOSDetail chi tiet 1 SOS case.
func (h *SupportHandler) GetSOSDetail(c *gin.Context) {
	sosIDStr := c.Query("sos_id")
	if sosIDStr == "" {
		response.ErrMissingParam(c)
		return
	}

	sosID, err := strconv.ParseUint(sosIDStr, 10, 64)
	if err != nil {
		response.ErrInvalidType(c)
		return
	}

	sos, err := h.sosSvc.GetSOSDetail(sosID)
	if err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, sos)
}
