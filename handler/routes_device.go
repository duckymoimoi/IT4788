package handler

import (
	"hospital/middleware" // Import package middleware cua Leader
	"hospital/repository"
	"hospital/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterDeviceRoutes dang ky device endpoints (Slice 7).
func RegisterDeviceRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// 1. Khởi tạo bộ 3: Repo -> Service -> Handler ngay tại đây
	deviceRepo := repository.NewDeviceRepo(db)
	deviceService := service.NewDeviceService(deviceRepo)
	deviceHandler := NewDeviceHandler(deviceService)

	// 2. Tạo nhóm route /device
	deviceGroup := api.Group("/device")
	
	// 3. Áp dụng middleware kiểm tra đăng nhập cho toàn bộ API mượn/trả thiết bị
	// (Lưu ý: Tùy theo Leader đặt tên hàm là Auth() hay RequireAuth(), bạn chỉnh lại cho khớp nhé)
	deviceGroup.Use(middleware.Auth()) 
	{
		// GET Methods
		deviceGroup.GET("/stations", deviceHandler.GetStations)
		deviceGroup.GET("/wheelchairs", deviceHandler.GetWheelchairs)
		deviceGroup.GET("/status/:id", deviceHandler.GetDeviceStatus)
		deviceGroup.GET("/track/:id", deviceHandler.TrackDevice)

		// POST Methods
		deviceGroup.POST("/book", deviceHandler.BookDevice)
		deviceGroup.POST("/release", deviceHandler.ReleaseDevice)
		deviceGroup.POST("/report_broken", deviceHandler.ReportBroken)
		deviceGroup.POST("/request_staff", deviceHandler.RequestStaff)
	}
}