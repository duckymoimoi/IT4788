package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"hospital/middleware"
	"hospital/repository"
	"hospital/service"
)

// RegisterRoutes dang ky toan bo route vao router.
// Duoc goi 1 lan within main.go khi khoi dong server.
//
// Khoi tao theo thu tu: repo -> service -> handler -> route group.
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	// Khoi tao cac tang
	userRepo := repository.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)

	authH := NewAuthHandler(authSvc)
	userH := NewUserHandler(userSvc)
	sysH := NewSysHandler(userSvc)

	api := router.Group("/api")

	// =============================================
	// AUTH — Public routes (khong can token)
	// =============================================
	auth := api.Group("/auth")
	auth.POST("/login", authH.Login)
	auth.POST("/signup", authH.Signup)
	auth.POST("/verify_otp", authH.VerifyOTP)
	auth.POST("/forgot_password", authH.ForgotPassword)
	auth.POST("/reset_password", authH.ResetPassword)

	// =============================================
	// AUTH — Private routes (can token)
	// =============================================
	authPriv := api.Group("/auth")
	authPriv.Use(middleware.Auth())
	authPriv.POST("/logout", authH.Logout)
	authPriv.POST("/change_password", authH.ChangePassword)

	// =============================================
	// USER — Private routes (can token)
	// =============================================
	user := api.Group("/user")
	user.Use(middleware.Auth())
	user.GET("/get_profile", userH.GetProfile)
	user.POST("/set_profile", userH.SetProfile)
	user.POST("/set_devtoken", userH.SetDevToken)
	user.GET("/get_settings", userH.GetSettings)
	user.POST("/set_settings", userH.SetSettings)
	user.DELETE("/delete_account", userH.DeleteAccount)

	// =============================================
	// SYS — Public routes (khong can token)
	// =============================================
	api.GET("/sys/check_version", sysH.CheckVersion)
}
