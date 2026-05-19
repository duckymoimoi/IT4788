package handler

import (
	"github.com/gin-gonic/gin"
	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
	"strconv"
)

type NotifHandler struct {
	svc service.NotifService
}

func NewNotifHandler(svc service.NotifService) *NotifHandler {
	return &NotifHandler{svc: svc}
}

type notifActionRequest struct {
	NotifID uint64 `json:"notif_id"`
	UserID  uint64 `json:"user_id"`
}

// [71] GET /api/notif/get_list
func (h *NotifHandler) GetList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	notifs, total, err := h.svc.GetNotifications(userID, page, limit)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{
		"notifications": notifs,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// [72] POST /api/notif/set_read
func (h *NotifHandler) SetRead(c *gin.Context) {
	var req notifActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.ReadNotification(req.NotifID, userID); err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

// [73] DELETE /api/notif/delete
func (h *NotifHandler) Delete(c *gin.Context) {
	var req notifActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteNotification(req.NotifID, userID); err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

// GetNotificationCompat supports the legacy test-suite endpoint:
// GET /api/notif/get_notification?index=&count=&user_id=
func (h *NotifHandler) GetNotificationCompat(c *gin.Context) {
	userID, ok := compatUserID(c, "query")
	if !ok {
		response.ErrNotAuthenticated(c)
		return
	}

	indexStr := c.Query("index")
	countStr := c.Query("count")
	if indexStr == "" || countStr == "" {
		response.ErrMissingParam(c)
		return
	}

	index, errIndex := strconv.Atoi(indexStr)
	count, errCount := strconv.Atoi(countStr)
	if errIndex != nil || errCount != nil {
		response.ErrInvalidType(c)
		return
	}
	if index < 0 || count <= 0 {
		response.ErrInvalidValue(c)
		return
	}

	page := index/count + 1
	notifs, _, err := h.svc.GetNotifications(userID, page, count)
	if err != nil {
		response.ErrInternalError(c)
		return
	}
	response.Success(c, notifs)
}

// ReadNotificationCompat supports POST /api/notif/read_notification.
func (h *NotifHandler) ReadNotificationCompat(c *gin.Context) {
	var req notifActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.NotifID == 0 {
		response.ErrMissingParam(c)
		return
	}
	if req.NotifID == 0 {
		response.ErrMissingParam(c)
		return
	}

	userID, ok := compatUserIDFromBody(c, req.UserID)
	if !ok {
		response.ErrNotAuthenticated(c)
		return
	}
	if err := h.svc.ReadNotification(req.NotifID, userID); err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

// DeleteNotificationCompat supports POST /api/notif/del_notification.
func (h *NotifHandler) DeleteNotificationCompat(c *gin.Context) {
	var req notifActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.NotifID == 0 {
		response.ErrMissingParam(c)
		return
	}
	if req.NotifID == 0 {
		response.ErrMissingParam(c)
		return
	}

	userID, ok := compatUserIDFromBody(c, req.UserID)
	if !ok {
		response.ErrNotAuthenticated(c)
		return
	}
	if err := h.svc.DeleteNotification(req.NotifID, userID); err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func compatUserIDFromBody(c *gin.Context, bodyUserID uint64) (uint64, bool) {
	if bodyUserID != 0 {
		return bodyUserID, c.GetHeader("token") != ""
	}
	return compatUserID(c, "header")
}

func compatUserID(c *gin.Context, source string) (uint64, bool) {
	if c.GetHeader("token") == "" {
		return 0, false
	}
	var value string
	if source == "query" {
		value = c.Query("user_id")
	}
	if value == "" {
		value = c.GetHeader("user_id")
	}
	if value == "" {
		value = c.GetHeader("token")
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}
