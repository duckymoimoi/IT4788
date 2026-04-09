package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/schema"
	"hospital/service"
)

// ChatHandler xu ly HTTP request cho module Chat.
// Person E so huu file nay.
type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

// ========================================
// REQUEST STRUCTS
// ========================================

type createRoomRequest struct {
	StaffID uint64 `json:"staff_id" binding:"required"`
	Topic   string `json:"topic"`
}

type sendMessageRequest struct {
	ConversationID uint64 `json:"conversation_id" binding:"required"`
	Type           string `json:"type" binding:"required"` // text, image, voice
	TextContent    string `json:"text_content"`
	MediaURL       string `json:"media_url"`
}

type closeRoomRequest struct {
	ConversationID uint64 `json:"conversation_id" binding:"required"`
}

type markReadRequest struct {
	ConversationID uint64 `json:"conversation_id" binding:"required"`
}

// ========================================
// [101] POST /api/chat/create_room
// ========================================

// CreateRoom benh nhan tao phong chat voi staff.
func (h *ChatHandler) CreateRoom(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	conv, err := h.chatSvc.CreateRoom(userID, req.StaffID, req.Topic)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, conv)
}

// ========================================
// [102] GET /api/chat/get_rooms
// ========================================

// GetRooms lay danh sach phong chat cua user.
func (h *ChatHandler) GetRooms(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	role := middleware.GetRole(c)

	rooms, err := h.chatSvc.GetRooms(userID, role)
	if err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"rooms": rooms,
	})
}

// ========================================
// [103] GET /api/chat/get_messages?conversation_id=&page=&limit=
// ========================================

// GetMessages lay lich su tin nhan.
func (h *ChatHandler) GetMessages(c *gin.Context) {
	convIDStr := c.Query("conversation_id")
	if convIDStr == "" {
		response.ErrMissingParam(c)
		return
	}

	convID, err := strconv.ParseUint(convIDStr, 10, 64)
	if err != nil {
		response.ErrInvalidType(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	msgs, total, err := h.chatSvc.GetMessages(convID, page, limit)
	if err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"messages": msgs,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// ========================================
// [104] POST /api/chat/send_message
// ========================================

// SendMessage gui tin nhan trong conversation.
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	// Xac dinh sender_type dua vao role
	role := middleware.GetRole(c)
	senderType := schema.SenderTypeUser
	if role == "admin" || role == "coordinator" || role == "staff" {
		senderType = schema.SenderTypeStaff
	}

	// Parse message type
	msgType := schema.MessageType(req.Type)

	msg, err := h.chatSvc.SendMessage(
		req.ConversationID, userID, senderType,
		msgType, req.TextContent, req.MediaURL,
	)
	if err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, msg)
}

// ========================================
// [105] POST /api/chat/close_room
// ========================================

// CloseRoom dong phong chat (chi staff).
func (h *ChatHandler) CloseRoom(c *gin.Context) {
	var req closeRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if err := h.chatSvc.CloseRoom(req.ConversationID); err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"closed": true})
}

// ========================================
// [106] GET /api/chat/get_unread_count?conversation_id=
// ========================================

// GetUnreadCount dem tin nhan chua doc.
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	convIDStr := c.Query("conversation_id")
	if convIDStr == "" {
		response.ErrMissingParam(c)
		return
	}

	convID, err := strconv.ParseUint(convIDStr, 10, 64)
	if err != nil {
		response.ErrInvalidType(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	count, err := h.chatSvc.GetUnreadCount(convID, userID)
	if err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"unread_count": count,
	})
}

// ========================================
// [107] POST /api/chat/mark_read
// ========================================

// MarkRead danh dau da doc tat ca tin nhan.
func (h *ChatHandler) MarkRead(c *gin.Context) {
	var req markReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	if err := h.chatSvc.MarkRead(req.ConversationID, userID); err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"marked": true})
}
