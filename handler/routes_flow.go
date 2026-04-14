package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"hospital/middleware"
	"hospital/repository"
	"hospital/service"
)

// RegisterFlowRoutes dang ky flow + simulation endpoints (Slice 5).
// Nguoi B implement  - 15 API.
func RegisterFlowRoutes(api *gin.RouterGroup, db *gorm.DB) {
	repo := repository.NewFlowRepo(db)
	svc := service.NewFlowService(repo)
	h := NewFlowHandler(svc)

	// =============================================
	// FLOW  - Public (khong can auth)
	// =============================================
	flow := api.Group("/flow")
	flow.GET("/get_density", h.GetDensity)           // #47
	flow.GET("/get_heatmap", h.GetHeatmap)            // #48
	flow.GET("/get_bottlenecks", h.GetBottlenecks)    // #49
	flow.GET("/get_forecast", h.GetForecast)           // #52
	flow.GET("/get_alerts", h.GetAlerts)               // #54
	flow.GET("/edge_status", h.EdgeStatus)             // #55

	// =============================================
	// FLOW  - Private (can auth benh nhan/staff)
	// =============================================
	flowPriv := api.Group("/flow")
	flowPriv.Use(middleware.Auth())
	flowPriv.POST("/ping_location", h.PingLocation)       // #46
	flowPriv.POST("/report_obstacle", h.ReportObstacle)   // #50
	flowPriv.GET("/get_obstacles", h.GetObstacles)         // danh sach bao cao
	flowPriv.POST("/set_priority", h.SetPriority)          // #53
	flowPriv.POST("/expire_priority", h.ExpirePriority)    // het han priority

	// =============================================
	// FLOW  - Staff only (can staff/coordinator/admin)
	// =============================================
	flowStaff := api.Group("/flow")
	flowStaff.Use(middleware.Auth(), middleware.RequireStaff())
	flowStaff.POST("/resolve_obstacle", h.ResolveObstacle) // xu ly obstacle

	// =============================================
	// ADMIN  - Flow admin endpoints
	// =============================================
	admin := api.Group("/admin")
	admin.Use(middleware.Auth(), middleware.RequireAdmin())
	admin.PATCH("/set_capacity", h.SetCapacity)   // #51
	admin.GET("/stats_flow", h.StatsFlow)          // #56
	admin.POST("/reset_flow", h.ResetFlow)         // #57

	// =============================================
	// SIMULATE  - Admin only
	// =============================================
	sim := api.Group("/simulate")
	sim.Use(middleware.Auth(), middleware.RequireAdmin())
	sim.POST("/start", h.StartSimulation)    // #58
	sim.POST("/stop", h.StopSimulation)       // #59
	sim.GET("/status", h.SimulationStatus)    // #60
}
