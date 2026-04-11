package handler

import (
	"hospital/middleware"
	"hospital/repository"
	"hospital/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterUtilRoutes dang ky utility endpoints (Slice 10).
func RegisterUtilRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// 1. Khởi tạo dây chuyền xử lý
	utilRepo := repository.NewUtilRepo(db)
	mapRepo := repository.NewMapRepo(db)
	utilService := service.NewUtilService(utilRepo)
	utilHandler := NewUtilHandler(utilService, mapRepo)

	// 2. Tạo nhóm route lớn /util
	utilGroup := api.Group("/util")
	{
		// ========================================
		// ZONE 1: API CÔNG KHAI (Không cần đăng nhập)
		// ========================================
		utilGroup.GET("/faq", utilHandler.GetFAQ)
		utilGroup.GET("/feedback_summary", utilHandler.GetFeedbackSummary)
		utilGroup.GET("/check_version", utilHandler.CheckVersion)
		utilGroup.GET("/languages", utilHandler.GetLanguages)
		utilGroup.GET("/about", utilHandler.GetAbout)
		utilGroup.GET("/contact", utilHandler.GetContact)
		utilGroup.GET("/pharmacy", utilHandler.GetPharmacy)     // #99
		utilGroup.GET("/canteen", utilHandler.GetCanteen)       // #100
		utilGroup.GET("/parking", utilHandler.GetParking)       // #101
		utilGroup.GET("/wifi", utilHandler.GetWifi)             // #102
		utilGroup.GET("/weather", utilHandler.GetWeather)       // #106

		// ========================================
		// ZONE 2: API BẢO MẬT (Bắt buộc đăng nhập)
		// ========================================
		authGroup := utilGroup.Group("")
		authGroup.Use(middleware.Auth())
		{
			authGroup.POST("/feedback", utilHandler.SubmitFeedback)
		}
	}
}