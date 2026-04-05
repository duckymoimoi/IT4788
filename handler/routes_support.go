package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"hospital/middleware"
	"hospital/repository"
	"hospital/service"
)

// RegisterSupportRoutes dang ky support endpoints (Slice 9).
// Person E so huu file nay.
// 12 REST API + 1 WebSocket endpoint.
func RegisterSupportRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// ---- Repository (shared) ----
	supportRepo := repository.NewSupportRepo(db)

	// ---- Services ----
	sosSvc := service.NewSOSService(supportRepo)
	chatSvc := service.NewChatService(supportRepo)

	// ---- Handlers ----
	sosH := NewSupportHandler(sosSvc)
	chatH := NewChatHandler(chatSvc)
	wsH := NewWSHandler(chatSvc)

	// ========================================
	// SOS APIs (5 endpoints)
	// ========================================

	// [96] Benh nhan tao SOS — bat ky user nao cung duoc
	sos := api.Group("/sos")
	sos.Use(middleware.Auth())
	sos.POST("/create", sosH.CreateSOS)         // [96]
	sos.GET("/get_detail", sosH.GetSOSDetail)    // [100]

	// [97-99] Chi staff moi duoc xem DS, nhan va dong SOS
	sosStaff := api.Group("/sos")
	sosStaff.Use(middleware.Auth(), middleware.RequireStaff())
	sosStaff.GET("/get_list", sosH.GetSOSList)   // [97]
	sosStaff.POST("/respond", sosH.RespondSOS)   // [98]
	sosStaff.POST("/resolve", sosH.ResolveSOS)   // [99]

	// ========================================
	// CHAT APIs (7 endpoints)
	// ========================================

	chat := api.Group("/chat")
	chat.Use(middleware.Auth())
	chat.POST("/create_room", chatH.CreateRoom)       // [101]
	chat.GET("/get_rooms", chatH.GetRooms)             // [102]
	chat.GET("/get_messages", chatH.GetMessages)       // [103]
	chat.POST("/send_message", chatH.SendMessage)      // [104]
	chat.GET("/get_unread_count", chatH.GetUnreadCount) // [106]
	chat.POST("/mark_read", chatH.MarkRead)            // [107]

	// [105] Chi staff moi duoc dong phong
	chatStaff := api.Group("/chat")
	chatStaff.Use(middleware.Auth(), middleware.RequireStaff())
	chatStaff.POST("/close_room", chatH.CloseRoom)    // [105]

	// ========================================
	// WebSocket (1 endpoint)
	// ========================================

	// KHÔNG dung middleware.Auth() vi WS truyen token qua query param.
	// Xac thuc xu ly trong WSHandler.HandleWS().
	api.GET("/ws/chat", wsH.HandleWS)
}
