package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
)

// FlowHandler xu ly HTTP request cho Flow + Simulation module (Slice 5).
type FlowHandler struct {
	svc *service.FlowService
}

func NewFlowHandler(svc *service.FlowService) *FlowHandler {
	return &FlowHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS
// ========================================

type pingRequest struct {
	GridLocation int     `json:"grid_location" binding:"required"`
	GridRow      int     `json:"grid_row"      binding:"required"`
	GridCol      int     `json:"grid_col"      binding:"required"`
	RouteID      *string `json:"route_id"`
}

type reportObstacleRequest struct {
	GridLocation int     `json:"grid_location" binding:"required"`
	ReportType   string  `json:"report_type"   binding:"required"`
	Description  string  `json:"description"`
	RouteID      *string `json:"route_id"`
}

type setCapacityRequest struct {
	POIID    uint32 `json:"poi_id"   binding:"required"`
	Capacity int    `json:"capacity" binding:"required"`
}

type setPriorityRequest struct {
	FromLocation int     `json:"from_location" binding:"required"`
	ToLocation   int     `json:"to_location"   binding:"required"`
	Reason       string  `json:"reason"        binding:"required"`
	EmergencyID  *string `json:"emergency_id"`
}

type resolveObstacleRequest struct {
	ReportID uint64 `json:"report_id" binding:"required"`
	Action   string `json:"action"` // "resolve" hoac "reject", default "resolve"
}

type expirePriorityRequest struct {
	PriorityID uint64 `json:"priority_id" binding:"required"`
}

type startSimRequest struct {
	MapID      uint32 `json:"map_id"      binding:"required"`
	OutputFile string `json:"output_file" binding:"required"`
	TickRateMs int    `json:"tick_rate_ms"`
}

// ========================================
// FLOW APIs  (#46 - #57)
// ========================================

// [46] POST /api/flow/ping_location
func (h *FlowHandler) PingLocation(c *gin.Context) {
	var req pingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	if err := h.svc.PingLocation(userID, req.GridLocation, req.GridRow, req.GridCol, req.RouteID); err != nil {
		response.Error(c, response.CodeInvalidLocationData, err.Error())
		return
	}

	response.Success(c, gin.H{"pinged": true})
}

// [47] GET /api/flow/get_density?grid_location=
func (h *FlowHandler) GetDensity(c *gin.Context) {
	locStr := c.Query("grid_location")
	if locStr == "" {
		response.ErrMissingParam(c)
		return
	}

	gridLocation, err := strconv.Atoi(locStr)
	if err != nil {
		response.Error(c, response.CodeInvalidLocationData, "Invalid grid_location")
		return
	}

	result, err := h.svc.GetDensity(gridLocation)
	if err != nil {
		response.Error(c, response.CodeDensityUnavailable, err.Error())
		return
	}

	response.Success(c, result)
}

// [48] GET /api/flow/get_heatmap
func (h *FlowHandler) GetHeatmap(c *gin.Context) {
	entries, err := h.svc.GetHeatmap()
	if err != nil {
		response.Error(c, response.CodeDensityUnavailable, err.Error())
		return
	}

	response.Success(c, entries)
}

// [49] GET /api/flow/get_bottlenecks?limit=
func (h *FlowHandler) GetBottlenecks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	results, err := h.svc.GetBottlenecks(limit)
	if err != nil {
		response.Error(c, response.CodeDensityUnavailable, err.Error())
		return
	}

	response.Success(c, results)
}

// [50] POST /api/flow/report_obstacle
func (h *FlowHandler) ReportObstacle(c *gin.Context) {
	var req reportObstacleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	report, err := h.svc.ReportObstacle(userID, req.GridLocation, req.ReportType, req.Description, req.RouteID)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, report)
}

// GetObstacles GET /api/flow/get_obstacles?status=&page=&limit=
func (h *FlowHandler) GetObstacles(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	reports, total, err := h.svc.GetObstacles(status, page, limit)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{
		"reports": reports,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// ResolveObstacle POST /api/flow/resolve_obstacle
func (h *FlowHandler) ResolveObstacle(c *gin.Context) {
	var req resolveObstacleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	// Lay staff_id tu context (user_id trung voi staff.user_id)
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	action := req.Action
	if action == "" {
		action = "resolve"
	}

	if err := h.svc.ResolveObstacle(req.ReportID, userID, action); err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{"resolved": true})
}

// [51] PATCH /api/admin/set_capacity
func (h *FlowHandler) SetCapacity(c *gin.Context) {
	var req setCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if err := h.svc.SetCapacity(req.POIID, req.Capacity); err != nil {
		response.ErrBadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"updated": true})
}

// [52] GET /api/flow/get_forecast?hours=
func (h *FlowHandler) GetForecast(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

	stats, err := h.svc.GetForecast(hours)
	if err != nil {
		response.Error(c, response.CodeDensityUnavailable, err.Error())
		return
	}

	response.Success(c, stats)
}

// [53] POST /api/flow/set_priority
func (h *FlowHandler) SetPriority(c *gin.Context) {
	var req setPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	pr, err := h.svc.SetPriority(userID, req.FromLocation, req.ToLocation, req.Reason, req.EmergencyID)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, pr)
}

// [54] GET /api/flow/get_alerts
func (h *FlowHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.svc.GetAlerts()
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, alerts)
}

// ExpirePriority POST /api/flow/expire_priority
func (h *FlowHandler) ExpirePriority(c *gin.Context) {
	var req expirePriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if err := h.svc.ExpirePriority(req.PriorityID); err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{"expired": true})
}

// [55] GET /api/flow/edge_status?grid_location=
func (h *FlowHandler) EdgeStatus(c *gin.Context) {
	locStr := c.Query("grid_location")
	if locStr == "" {
		response.ErrMissingParam(c)
		return
	}

	gridLocation, err := strconv.Atoi(locStr)
	if err != nil {
		response.Error(c, response.CodeInvalidLocationData, "Invalid grid_location")
		return
	}

	result, err := h.svc.GetEdgeStatus(gridLocation)
	if err != nil {
		response.Error(c, response.CodeDensityUnavailable, err.Error())
		return
	}

	response.Success(c, result)
}

// [56] GET /api/admin/stats_flow?hours=
func (h *FlowHandler) StatsFlow(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

	stats, err := h.svc.GetStatsFlow(hours)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, stats)
}

// [57] POST /api/admin/reset_flow
func (h *FlowHandler) ResetFlow(c *gin.Context) {
	if err := h.svc.ResetFlow(); err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{"reset": true})
}

// ========================================
// SIMULATION APIs  (#58 - #60)
// ========================================

// [58] POST /api/simulate/start
func (h *FlowHandler) StartSimulation(c *gin.Context) {
	var req startSimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if req.TickRateMs <= 0 {
		req.TickRateMs = 1000
	}

	info, err := h.svc.StartSimulation(req.MapID, req.OutputFile, req.TickRateMs)
	if err != nil {
		response.Error(c, response.CodeEngineUnavailable, err.Error())
		return
	}

	response.Success(c, info)
}

// [59] POST /api/simulate/stop
func (h *FlowHandler) StopSimulation(c *gin.Context) {
	if err := h.svc.StopSimulation(); err != nil {
		response.Error(c, response.CodeEngineUnavailable, err.Error())
		return
	}

	response.Success(c, gin.H{"stopped": true})
}

// [60] GET /api/simulate/status
func (h *FlowHandler) SimulationStatus(c *gin.Context) {
	info, err := h.svc.SimulationStatus()
	if err != nil {
		response.Error(c, response.CodeEngineUnavailable, err.Error())
		return
	}

	response.Success(c, info)
}
