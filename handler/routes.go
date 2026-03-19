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
	mapRepo := repository.NewMapRepo(db)

	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	mapSvc := service.NewMapService(mapRepo)

	authH := NewAuthHandler(authSvc)
	userH := NewUserHandler(userSvc)
	sysH := NewSysHandler(userSvc)
	mapH := NewMapHandler(mapSvc)

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

	// =============================================
	// MAP — Public routes (khong can token)
	// API 16-22, 24
	// =============================================
	mapG := api.Group("/map")
	mapG.GET("/get_floors", mapH.GetFloors)       // [16]
	mapG.GET("/get_nodes", mapH.GetNodes)          // [17]
	mapG.GET("/get_edges", mapH.GetEdges)          // [18]
	mapG.GET("/get_meta", mapH.GetMeta)            // [19]
	mapG.GET("/get_depts", mapH.GetDepartments)    // [20]
	mapG.GET("/search_location", mapH.SearchLocation) // [21]
	mapG.GET("/get_landmarks", mapH.GetLandmarks)  // [22]
	mapG.GET("/sync_full", mapH.SyncFull)          // [24]

	// =============================================
	// ADMIN — Private routes (can token + role admin)
	// API 25-30
	// =============================================
	admin := api.Group("/admin")
	admin.Use(middleware.Auth(), middleware.RequireAdmin())
	admin.POST("/add_node", mapH.AddNode)          // [25]
	admin.POST("/edit_node", mapH.EditNode)        // [26]
	admin.DELETE("/del_node", mapH.DelNode)        // [27]
	admin.POST("/add_edge", mapH.AddEdge)          // [28]
	admin.DELETE("/del_edge", mapH.DelEdge)        // [29]
	admin.PATCH("/set_weight", mapH.SetWeight)     // [30]
}
