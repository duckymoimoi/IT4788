package handler

import (
	"hospital/middleware"
	"hospital/repository"
	"hospital/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterMedicalRoutes dang ky medical endpoints (Slice 6).
// Khoi tao repo, service, handler ben trong de giu signature nhat quan voi routes.go.
func RegisterMedicalRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// 1. Khoi tao Repository
	medicalRepo := repository.NewMedicalRepository(db)
	mapRepo := repository.NewMapRepo(db)

	// 2. Khoi tao Service
	medicalSvc := service.NewMedicalService(medicalRepo, mapRepo, db)

	// 3. Khoi tao Handler
	h := NewMedicalHandler(medicalSvc)

	// 4. Nhom cac API Medical
	medical := api.Group("/medical")
	medical.Use(middleware.Auth())
	{
		medical.GET("/get_tasks", h.GetTasks)             // #61
		medical.GET("/get_queue", h.GetQueue)              // #62
		medical.POST("/checkin_room", h.CheckinRoom)       // #63
		medical.POST("/checkout_room", h.CheckoutRoom)     // #64
		medical.GET("/result_status", h.GetResultStatus)   // #65
		medical.GET("/get_prescription", h.GetPrescription) // #66
		medical.POST("/sync_now", h.SyncNow)               // #67
		medical.GET("/room_open", h.GetRoomOpen)           // #68
		medical.POST("/cancel_task", h.CancelTask)         // #69
		medical.GET("/get_history", h.GetHistory)          // #70
	}
}