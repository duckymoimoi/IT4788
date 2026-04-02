package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	response "hospital/pkg"
	"hospital/service"
)

// EngineHandler xu ly HTTP request cho Engine Admin (Slice 11).
type EngineHandler struct {
	svc *service.EngineService
}

func NewEngineHandler(svc *service.EngineService) *EngineHandler {
	return &EngineHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS
// ========================================

type solveMCMFRequest struct {
	StartLocation int    `json:"start_location" binding:"required"`
	DestLocation  int    `json:"dest_location" binding:"required"`
	ModeID        string `json:"mode_id" binding:"required"`
}

type updateCostRequest struct {
	POIID  uint32  `json:"poi_id" binding:"required"`
	Weight float32 `json:"weight" binding:"required"`
}

type setParamsRequest struct {
	MaxAgents      int     `json:"max_agents"`
	TimeStepMs     int     `json:"time_step_ms"`
	CostMultiplier float64 `json:"cost_multiplier"`
}

type loadMAPFRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

// ========================================
// ENGINE ADMIN APIs [91-98]
// ========================================

// [91] POST /api/engine/solve  - Chay Dijkstra (simplified MCMF).
func (h *EngineHandler) Solve(c *gin.Context) {
	var req solveMCMFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	result, err := h.svc.SolveMCMF(req.StartLocation, req.DestLocation, req.ModeID)
	if err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, result)
}

// [92] POST /api/engine/update_cost  - Cap nhat POI weight.
func (h *EngineHandler) UpdateCost(c *gin.Context) {
	var req updateCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if err := h.svc.UpdatePOICost(req.POIID, req.Weight); err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{"updated": true, "poi_id": req.POIID, "weight": req.Weight})
}

// [93] GET /api/engine/convergence  - Trang thai hoi tu.
func (h *EngineHandler) GetConvergence(c *gin.Context) {
	info := h.svc.GetConvergence()
	response.Success(c, info)
}

// [94] POST /api/engine/set_params  - Thiet lap tham so engine.
func (h *EngineHandler) SetParams(c *gin.Context) {
	var req setParamsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	h.svc.SetParams(service.EngineParams{
		MaxAgents:      req.MaxAgents,
		TimeStepMs:     req.TimeStepMs,
		CostMultiplier: req.CostMultiplier,
	})

	response.Success(c, h.svc.GetParams())
}

// [97] GET /api/engine/health  - Health check.
func (h *EngineHandler) Health(c *gin.Context) {
	info := h.svc.HealthCheck()
	response.Success(c, info)
}

// [98] POST /api/engine/clear_cache  - Xoa Dijkstra cache.
func (h *EngineHandler) ClearCache(c *gin.Context) {
	h.svc.ClearCache()
	response.Success(c, gin.H{"cache_cleared": true})
}

// ========================================
// MAPF REPLAY APIs (bonus)
// ========================================

// POST /api/engine/load_mapf  - Load output.json.
func (h *EngineHandler) LoadMAPF(c *gin.Context) {
	var req loadMAPFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if err := h.svc.LoadMAPFOutput(req.FilePath); err != nil {
		response.Error(c, response.CodeEngineUnavailable, err.Error())
		return
	}

	response.Success(c, h.svc.GetMAPFInfo())
}

// GET /api/engine/mapf_positions?timestep=  - Vi tri agents tai timestep.
func (h *EngineHandler) GetMAPFPositions(c *gin.Context) {
	ts, _ := strconv.Atoi(c.DefaultQuery("timestep", "0"))
	positions, err := h.svc.GetMAPFPositions(ts)
	if err != nil {
		response.Error(c, response.CodeEngineUnavailable, err.Error())
		return
	}

	response.Success(c, gin.H{
		"timestep":  ts,
		"positions": positions,
	})
}

// GET /api/engine/mapf_info  - MAPF metadata.
func (h *EngineHandler) GetMAPFInfo(c *gin.Context) {
	response.Success(c, h.svc.GetMAPFInfo())
}
