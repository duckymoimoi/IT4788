package service

import (
	"fmt"
	"sync"

	"hospital/pkg/mapf"
	"hospital/repository"
)

// EngineService xu ly logic Engine Admin (Slice 11).
// Quan ly Dijkstra cache, engine params, MAPF replay.
type EngineService struct {
	mapRepo     *repository.MapRepo
	routeSvc    *RouteService

	mu          sync.RWMutex
	params      EngineParams
	convergence *ConvergenceInfo
	mapfResult  *mapf.MAPFResult // loaded MAPF output
}

// EngineParams cau hinh engine (luu trong RAM).
type EngineParams struct {
	MaxAgents      int     `json:"max_agents"`
	TimeStepMs     int     `json:"time_step_ms"`
	CostMultiplier float64 `json:"cost_multiplier"`
}

// ConvergenceInfo thong tin hoi tu (trả từ RAM).
type ConvergenceInfo struct {
	Iteration  int     `json:"iteration"`
	Cost       float64 `json:"cost"`
	Converged  bool    `json:"converged"`
}

// HealthInfo ket qua health check.
type HealthInfo struct {
	Status      string `json:"status"`
	DBConnected bool   `json:"db_connected"`
	GridLoaded  bool   `json:"grid_loaded"`
	MAPFLoaded  bool   `json:"mapf_loaded"`
	AgentCount  int    `json:"agent_count"`
}

func NewEngineService(mapRepo *repository.MapRepo, routeSvc *RouteService) *EngineService {
	return &EngineService{
		mapRepo:  mapRepo,
		routeSvc: routeSvc,
		params: EngineParams{
			MaxAgents:      100,
			TimeStepMs:     500,
			CostMultiplier: 1.0,
		},
		convergence: &ConvergenceInfo{
			Iteration: 0,
			Cost:      0,
			Converged: false,
		},
	}
}

// ========================================
// API #91  - POST solve_mcmf
// ========================================

// SolveMCMF chay Dijkstra giua 2 diem, tra ve ket qua.
// Day la simplified version  - khong chay MCMF that, chi Dijkstra.
func (s *EngineService) SolveMCMF(startLoc, destLoc int, modeID string) (*PreviewResult, error) {
	return s.routeSvc.PreviewRoute(startLoc, destLoc, modeID)
}

// ========================================
// API #92  - POST update_cost
// ========================================

// UpdatePOICost cap nhat custom_weight cho POI.
// Admin dung de thay doi cost dua tren density data.
func (s *EngineService) UpdatePOICost(poiID uint32, weight float32) error {
	return s.mapRepo.UpdatePOIWeight(poiID, weight)
}

// ========================================
// API #93  - GET get_convergence
// ========================================

// GetConvergence tra ve thong tin hoi tu hien tai.
func (s *EngineService) GetConvergence() *ConvergenceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.convergence
}

// ========================================
// API #94  - POST set_params
// ========================================

// SetParams cap nhat engine params.
func (s *EngineService) SetParams(params EngineParams) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if params.MaxAgents > 0 {
		s.params.MaxAgents = params.MaxAgents
	}
	if params.TimeStepMs > 0 {
		s.params.TimeStepMs = params.TimeStepMs
	}
	if params.CostMultiplier > 0 {
		s.params.CostMultiplier = params.CostMultiplier
	}
}

// GetParams tra ve params hien tai.
func (s *EngineService) GetParams() EngineParams {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// ========================================
// API #97  - GET health
// ========================================

// HealthCheck kiem tra trang thai he thong.
func (s *EngineService) HealthCheck() *HealthInfo {
	info := &HealthInfo{
		Status: "ok",
	}

	// Check DB
	info.DBConnected = s.mapRepo.Ping()

	// Check grid cache
	if s.routeSvc.gridCache != nil {
		info.GridLoaded = true
	}

	// Check MAPF
	s.mu.RLock()
	if s.mapfResult != nil {
		info.MAPFLoaded = true
		info.AgentCount = s.mapfResult.TeamSize
	}
	s.mu.RUnlock()

	if !info.DBConnected {
		info.Status = "degraded"
	}

	return info
}

// ========================================
// API #98  - POST clear_cache
// ========================================

// ClearCache xoa Dijkstra grid cache.
func (s *EngineService) ClearCache() {
	s.routeSvc.ClearGridCache()
}

// ========================================
// MAPF Replay (bonus  - dùng parser.go)
// ========================================

// LoadMAPFOutput load file output.json de replay.
func (s *EngineService) LoadMAPFOutput(filePath string) error {
	result, err := mapf.ParseOutputJSON(filePath)
	if err != nil {
		return fmt.Errorf("cannot parse MAPF output: %w", err)
	}
	s.mu.Lock()
	s.mapfResult = result
	s.mu.Unlock()
	return nil
}

// GetMAPFPositions tra ve vi tri tat ca agents tai timestep.
func (s *EngineService) GetMAPFPositions(timestep int) ([]mapf.AgentState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.mapfResult == nil {
		return nil, fmt.Errorf("no MAPF output loaded")
	}
	return s.mapfResult.GetPositionsAtTimestep(timestep), nil
}

// GetMAPFInfo tra ve metadata cua MAPF output.
func (s *EngineService) GetMAPFInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.mapfResult == nil {
		return map[string]interface{}{"loaded": false}
	}
	return map[string]interface{}{
		"loaded":           true,
		"team_size":        s.mapfResult.TeamSize,
		"makespan":         s.mapfResult.Makespan,
		"num_task_finished": s.mapfResult.NumTaskFinished,
		"total_tasks":      len(s.mapfResult.Tasks),
	}
}
