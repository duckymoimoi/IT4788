package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
)

type NotifHandler struct {
	svc service.NotifService
}

func NewNotifHandler(svc service.NotifService) *NotifHandler {
	return &NotifHandler{svc: svc}
}

type notifActionRequest struct {
	NotifID uint64 `json:"notif_id" binding:"required"`
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