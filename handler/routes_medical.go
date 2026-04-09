package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"hospital/service"
)

// RegisterMedicalRoutes dang ky medical endpoints (Slice 6).
func RegisterMedicalRoutes(api *gin.RouterGroup, db *gorm.DB, svc service.MedicalService) {
	h := NewMedicalHandler(svc)

	medical := api.Group("/medical")
	{
		medical.GET("/get_tasks", h.GetTasks)          // #61
		medical.GET("/get_queue", h.GetQueue)          // #62
		medical.POST("/checkin_room", h.CheckinRoom)    // #63
		medical.POST("/sync_now", h.SyncNow)            // #67
		medical.GET("/room_open", h.GetRoomOpen)        // #68
	}
}