package handler

import (
	"hospital/middleware"
	"hospital/repository"
	"hospital/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterUtilRoutes dang ky utility endpoints (Slice 10).
// Nguoi D implement - thay the ham nay bang code that.
func RegisterUtilRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// 1. Khởi tạo dây chuyền xử lý: Repo -> Service -> Handler
	utilRepo := repository.NewUtilRepo(db)
	utilService := service.NewUtilService(utilRepo)
	utilHandler := NewUtilHandler(utilService)

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

		// ========================================
		// ZONE 2: API BẢO MẬT (Bắt buộc đăng nhập)
		// ========================================
		// Tạo một nhóm con ẩn danh để áp dụng middleware bảo vệ
		authGroup := utilGroup.Group("")
		authGroup.Use(middleware.Auth()) 
		{
			// Chỉ có API SubmitFeedback là bị yêu cầu kiểm tra Token
			authGroup.POST("/feedback", utilHandler.SubmitFeedback)
		}
	}
}