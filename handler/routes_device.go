package handler

import (
	"hospital/middleware"
	"hospital/repository"
	"hospital/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterDeviceRoutes đăng ký tất cả device/asset endpoints.
func RegisterDeviceRoutes(api *gin.RouterGroup, db *gorm.DB) {
	deviceRepo := repository.NewDeviceRepo(db)
	deviceService := service.NewDeviceService(deviceRepo)
	deviceHandler := NewDeviceHandler(deviceService)

	// ── /api/asset/* ────────────────────────────────────
	assetGroup := api.Group("/asset")
	assetGroup.Use(middleware.Auth())
	{
		assetGroup.GET("/asset_stations", deviceHandler.GetStations)
		assetGroup.GET("/find_wheelchairs", deviceHandler.GetWheelchairs)
		assetGroup.GET("/asset_health", deviceHandler.GetDeviceStatus)
		assetGroup.GET("/track_asset", deviceHandler.TrackDevice)

		assetGroup.POST("/book_asset", deviceHandler.BookDevice)
		assetGroup.POST("/release_asset", deviceHandler.ReleaseDevice)
		assetGroup.POST("/report_broken_asset", deviceHandler.ReportBroken)
	}

	// ── /api/staff/* ─────────────────────────────────────
	staffGroup := api.Group("/staff")
	staffGroup.Use(middleware.Auth())
	{
		staffGroup.POST("/request_staff", deviceHandler.RequestStaff)
	}
}

// RegisterAdminDeviceRoutes đăng ký admin endpoints cho device (gọi từ admin group).
func RegisterAdminDeviceRoutes(adminGroup *gin.RouterGroup, db *gorm.DB) {
	deviceRepo := repository.NewDeviceRepo(db)
	deviceService := service.NewDeviceService(deviceRepo)
	deviceHandler := NewDeviceHandler(deviceService)

	adminGroup.POST("/admin_add_device", deviceHandler.AdminAddDevice)
	adminGroup.POST("/admin_edit_device", deviceHandler.AdminEditDevice)
	adminGroup.POST("/admin_del_device", deviceHandler.AdminDelDevice)
}