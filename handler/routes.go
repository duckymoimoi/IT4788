package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"hospital/middleware"
	"hospital/repository"
	"hospital/service"
)

// RegisterRoutes dang ky toan bo route vao router.
// Shared service instances de dam bao 1 grid cache duy nhat.
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	// Shared repositories
	userRepo := repository.NewUserRepo(db)
	mapRepo := repository.NewMapRepo(db)
	routeRepo := repository.NewRouteRepo(db)

	// Shared services  - 1 instance duy nhat cho toan bo app
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	mapSvc := service.NewMapService(mapRepo)
	routeSvc := service.NewRouteService(routeRepo, mapRepo)
	engineSvc := service.NewEngineService(mapRepo, routeSvc)

	authH := NewAuthHandler(authSvc)
	userH := NewUserHandler(userSvc)
	sysH := NewSysHandler(userSvc)
	mapH := NewMapHandler(mapSvc)
	routeH := NewRouteHandler(routeSvc)
	engineH := NewEngineHandler(engineSvc)

	api := router.Group("/api")

	// =============================================
	// AUTH  - Public
	// =============================================
	auth := api.Group("/auth")
	auth.POST("/login", authH.Login)
	auth.POST("/signup", authH.Signup)
	auth.POST("/verify_otp", authH.VerifyOTP)
	auth.POST("/forgot_password", authH.ForgotPassword)
	auth.POST("/reset_password", authH.ResetPassword)

	// AUTH  - Private
	authPriv := api.Group("/auth")
	authPriv.Use(middleware.Auth())
	authPriv.POST("/logout", authH.Logout)
	authPriv.POST("/change_password", authH.ChangePassword)

	// =============================================
	// USER  - Private
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
	// SYS  - Public
	// =============================================
	api.GET("/sys/check_version", sysH.CheckVersion)
	api.GET("/sys/get_voice_key", sysH.GetVoiceKey)     // #79
	api.GET("/sys/get_voice_files", sysH.GetVoiceFiles) // #80

	// =============================================
	// MAP  - Public (API 16-22, 24)
	// =============================================
	mapG := api.Group("/map")
	mapG.GET("/get_floors", mapH.GetFloors)
	mapG.GET("/get_nodes", mapH.GetNodes)
	mapG.GET("/get_edges", mapH.GetEdges)
	mapG.GET("/get_meta", mapH.GetMeta)
	mapG.GET("/get_depts", mapH.GetDepartments)
	mapG.GET("/search_location", mapH.SearchLocation)
	mapG.GET("/get_landmarks", mapH.GetLandmarks)
	mapG.GET("/sync_full", mapH.SyncFull)
	// Legacy aliases used by the external Jest test suite.
	mapG.GET("/floors", mapH.GetFloors)
	mapG.GET("/nodes", mapH.GetNodes)
	mapG.GET("/edges", mapH.GetEdges)
	mapG.GET("/meta", mapH.GetMeta)
	mapG.GET("/search", mapH.SearchLocation)
	mapG.GET("/landmarks", mapH.GetLandmarks)

	// =============================================
	// ADMIN  - Private (admin only)
	// =============================================
	admin := api.Group("/admin")
	admin.Use(middleware.Auth(), middleware.RequireAdmin())
	admin.POST("/add_node", mapH.AddNode)
	admin.POST("/edit_node", mapH.EditNode)
	admin.DELETE("/del_node", mapH.DelNode)
	admin.POST("/add_edge", mapH.AddEdge)
	admin.DELETE("/del_edge", mapH.DelEdge)
	admin.PATCH("/set_weight", mapH.SetWeight)

	// Map File APIs
	admin.POST("/upload_map", mapH.UploadMap)
	admin.POST("/upload_output", mapH.UploadOutput)
	admin.POST("/set_active_map", mapH.SetActiveMap)
	admin.GET("/get_maps", mapH.GetMaps)
	admin.POST("/edit_map", mapH.EditMap)
	admin.POST("/update_grid", mapH.UpdateGrid)
	admin.GET("/export_map", mapH.ExportMap)
	admin.DELETE("/delete_map", mapH.DeleteMap)
	admin.POST("/deactivate_map", mapH.DeactivateMap)
	admin.GET("/get_request_logs", GetRequestLogs)
	admin.POST("/clear_request_logs", ClearRequestLogs)

	// Device Admin APIs
	RegisterAdminDeviceRoutes(admin, db)

	// =============================================
	// ROUTE  - Public + Private (shared routeH)
	// =============================================
	routeG := api.Group("/route")
	routeG.GET("/get_modes", routeH.GetModes)

	routePriv := api.Group("/route")
	routePriv.Use(middleware.Auth())
	routePriv.POST("/preview", routeH.Preview)
	routePriv.POST("/order", routeH.Order)
	routePriv.POST("/order_multi", routeH.OrderMulti)
	routePriv.POST("/order_unordered", routeH.OrderUnordered)
	routePriv.GET("/get_steps", routeH.GetSteps)
	routePriv.POST("/get_eta", routeH.GetETA)
	routePriv.GET("/get_active", routeH.GetActive)
	routePriv.POST("/cancel", routeH.Cancel)
	routePriv.POST("/recalculate", routeH.Recalculate)
	routePriv.POST("/pass_node", routeH.PassNode)
	routePriv.GET("/get_next", routeH.GetNext)
	routePriv.GET("/get_history", routeH.GetHistory)
	routePriv.DELETE("/clear_history", routeH.ClearHistory)
	routePriv.POST("/share", routeH.Share)
	routePriv.POST("/rate", routeH.Rate)

	// Legacy /api/routing aliases used by the external Jest test suite.
	routing := api.Group("/routing")
	routing.GET("/get_modes", routeH.GetModes)

	routingPriv := api.Group("/routing")
	routingPriv.Use(middleware.AuthCompat())
	routingPriv.POST("/route_ordered", routeH.OrderMulti)
	routingPriv.POST("/route_unordered", routeH.OrderUnordered)
	routingPriv.POST("/re_calculate", routeH.Recalculate)
	routingPriv.GET("/get_active", routeH.GetActive)
	routingPriv.POST("/cancel_route", routeH.Cancel)
	routingPriv.POST("/share_route", routeH.Share)
	routingPriv.POST("/rate_path", routeH.Rate)
	routingPriv.GET("/get_history", routeH.GetHistory)
	routingPriv.DELETE("/clear_history", routeH.ClearHistory)
	routingPriv.GET("/get_steps", routeH.GetSteps)
	routingPriv.POST("/preview_path", routeH.Preview)
	routingPriv.GET("/preview_path", routeH.Preview)
	routingPriv.POST("/get_eta", routeH.GetETA)
	routingPriv.POST("/pass_node", routeH.PassNode)

	// =============================================
	// ENGINE  - Admin only (shared engineH)
	// =============================================
	engine := api.Group("/engine")
	engine.Use(middleware.Auth(), middleware.RequireAdmin())
	engine.POST("/solve", engineH.Solve)
	engine.POST("/update_cost", engineH.UpdateCost)
	engine.GET("/convergence", engineH.GetConvergence)
	engine.POST("/set_params", engineH.SetParams)
	engine.GET("/health", engineH.Health)
	engine.POST("/clear_cache", engineH.ClearCache)
	engine.POST("/load_mapf", engineH.LoadMAPF)
	engine.GET("/mapf_positions", engineH.GetMAPFPositions)
	engine.GET("/mapf_info", engineH.GetMAPFInfo)

	// MODULE STUBS
	RegisterFlowRoutes(api, db)
	RegisterMedicalRoutes(api, db)
	RegisterNotifRoutes(api, db)
	RegisterDeviceRoutes(api, db)
	RegisterUtilRoutes(api, db)
	RegisterSupportRoutes(api, db)
}
