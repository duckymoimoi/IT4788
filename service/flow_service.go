package service

import (
	"fmt"
	"time"

	"hospital/pkg/mapf"
	"hospital/repository"
	"hospital/schema"
)

// FlowService xu ly logic nghiep vu cho Flow + Simulation module (Slice 5).
type FlowService struct {
	repo    *repository.FlowRepo
	manager *mapf.AgentManager
}

func NewFlowService(repo *repository.FlowRepo) *FlowService {
	return &FlowService{
		repo:    repo,
		manager: mapf.NewAgentManager(),
	}
}

// AutoStartSimulation tu dong bat dau mo phong khi server khoi dong.
// Chay loop vo han (reset ve timestep 0 khi het makespan).
// Goi tu RegisterFlowRoutes trong goroutine rieng.
func (s *FlowService) AutoStartSimulation(outputFile string, tickRateMs int) error {
	if err := s.manager.Start(outputFile, tickRateMs); err != nil {
		return fmt.Errorf("auto-start simulation failed: %w", err)
	}
	teamSize, makespan, _, _, _ := s.manager.GetInfo()
	fmt.Printf("[SIM] Auto-started: %d agents, makespan=%d, tick=%dms (loop forever)\n", teamSize, makespan, tickRateMs)
	return nil
}

// ========================================
// FLOW APIs (12 methods)
// ========================================

// PingLocation ghi nhan vi tri hien tai cua user.
// API #46 POST ping_location
func (s *FlowService) PingLocation(userID uint64, gridLocation, gridRow, gridCol int, routeID *string) error {
	ping := &schema.UserPing{
		UserID:       userID,
		RouteID:      routeID,
		GridLocation: gridLocation,
		GridRow:      gridRow,
		GridCol:      gridCol,
	}
	return s.repo.CreatePing(ping)
}

// DensityInfo ket qua mat do tai 1 vi tri.
type DensityInfo struct {
	GridLocation int   `json:"grid_location"`
	Count        int64 `json:"count"`
	Minutes      int   `json:"window_minutes"`
}

// GetDensity lay mat do tai 1 diem trong 30 phut gan nhat.
// API #47 GET get_density
func (s *FlowService) GetDensity(gridLocation int) (*DensityInfo, error) {
	minutes := 30
	count, err := s.repo.GetDensityByLocation(gridLocation, minutes)
	if err != nil {
		return nil, err
	}

	return &DensityInfo{
		GridLocation: gridLocation,
		Count:        count,
		Minutes:      minutes,
	}, nil
}

// HeatmapEntry 1 o tren heatmap.
type HeatmapEntry struct {
	GridLocation int   `json:"grid_location"`
	Density      int64 `json:"density"`
}

// GetHeatmap lay heatmap density tu pings gan day.
// API #48 GET get_heatmap
func (s *FlowService) GetHeatmap() ([]HeatmapEntry, error) {
	// Neu simulation dang chay, lay tu AgentManager
	if s.manager.IsRunning() {
		positions := s.manager.GetAllPositions()
		densityMap := make(map[int]int64)
		for _, pos := range positions {
			densityMap[pos.Location]++
		}
		entries := make([]HeatmapEntry, 0, len(densityMap))
		for loc, count := range densityMap {
			entries = append(entries, HeatmapEntry{GridLocation: loc, Density: count})
		}
		return entries, nil
	}

	// Mac dinh: lay tu user_pings 30 phut gan nhat
	results, err := s.repo.GetDensityAll(30)
	if err != nil {
		return nil, err
	}

	entries := make([]HeatmapEntry, len(results))
	for i, r := range results {
		entries[i] = HeatmapEntry{
			GridLocation: r.GridLocation,
			Density:      r.Count,
		}
	}
	return entries, nil
}

// GetBottlenecks lay top N diem un tac.
// API #49 GET get_bottlenecks
func (s *FlowService) GetBottlenecks(limit int) ([]repository.DensityResult, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetBottlenecks(30, limit)
}

// ReportObstacle benh nhan bao cao vat can.
// API #50 POST report_obstacle
func (s *FlowService) ReportObstacle(userID uint64, gridLocation int, reportType, description string, routeID *string) (*schema.ObstacleReport, error) {
	report := &schema.ObstacleReport{
		UserID:       userID,
		RouteID:      routeID,
		GridLocation: gridLocation,
		ReportType:   reportType,
		Description:  description,
		Status:       schema.ObstacleStatusPending,
	}
	if err := s.repo.CreateObstacleReport(report); err != nil {
		return nil, err
	}
	return report, nil
}

// SetCapacity admin cap nhat capacity cua POI.
// API #51 PATCH set_capacity
func (s *FlowService) SetCapacity(poiID uint32, capacity int) error {
	if capacity < 0 {
		return fmt.Errorf("capacity must be non-negative")
	}
	return s.repo.UpdatePOICapacity(poiID, capacity)
}

// GetForecast du bao luu thong theo gio.
// API #52 GET get_forecast
func (s *FlowService) GetForecast(hours int) ([]repository.HourlyStats, error) {
	if hours <= 0 {
		hours = 24
	}
	return s.repo.GetPingsByHour(hours)
}

// SetPriority staff dat tuyen uu tien.
// API #53 POST set_priority
func (s *FlowService) SetPriority(staffID uint64, fromLocation, toLocation int, reason string, emergencyID *string) (*schema.PriorityRoute, error) {
	pr := &schema.PriorityRoute{
		EmergencyID:  emergencyID,
		SetBy:        staffID,
		FromLocation: fromLocation,
		ToLocation:   toLocation,
		Reason:       reason,
		Status:       schema.PriorityStatusActive,
	}
	if err := s.repo.CreatePriorityRoute(pr); err != nil {
		return nil, err
	}
	return pr, nil
}

// GetAlerts lay cac tuyen uu tien dang active.
// API #54 GET get_alerts
func (s *FlowService) GetAlerts() ([]schema.PriorityRoute, error) {
	return s.repo.GetActivePriorityRoutes()
}

// GetEdgeStatus lay trang thai 1 hanh lang (so nguoi dang o do).
// API #55 GET edge_status
func (s *FlowService) GetEdgeStatus(gridLocation int) (*DensityInfo, error) {
	count, err := s.repo.GetEdgeDensity(gridLocation)
	if err != nil {
		return nil, err
	}
	return &DensityInfo{
		GridLocation: gridLocation,
		Count:        count,
		Minutes:      30,
	}, nil
}

// GetStatsFlow thong ke flow theo gio cho admin.
// API #56 GET stats_flow
func (s *FlowService) GetStatsFlow(hours int) ([]repository.HourlyStats, error) {
	if hours <= 0 {
		hours = 24
	}
	return s.repo.GetPingsByHour(hours)
}

// ResetFlow xoa tat ca flow data.
// API #57 POST reset_flow
func (s *FlowService) ResetFlow() error {
	return s.repo.ResetFlowData()
}

// GetObstacles lay danh sach bao cao.
// API #50 (GET variant) - ho tro xem danh sach
func (s *FlowService) GetObstacles(status string, page, limit int) ([]schema.ObstacleReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.repo.GetObstacleReports(status, page, limit)
}

// ResolveObstacle staff xu ly bao cao vat can.
// API #51 (POST variant) - resolve
func (s *FlowService) ResolveObstacle(reportID uint64, staffID uint64, action string) error {
	var status schema.ObstacleReportStatus
	switch action {
	case "resolve":
		status = schema.ObstacleStatusResolved
	case "reject":
		status = schema.ObstacleStatusRejected
	default:
		status = schema.ObstacleStatusResolved
	}
	return s.repo.ResolveObstacle(reportID, staffID, status)
}

// ExpirePriority set tuyen uu tien het han.
func (s *FlowService) ExpirePriority(priorityID uint64) error {
	return s.repo.ExpirePriorityRoute(priorityID)
}

// ========================================
// SIMULATION APIs (3 methods)
// ========================================

// SimulationInfo thong tin phien mo phong.
type SimulationInfo struct {
	RunID       uint64                `json:"run_id,omitempty"`
	Running     bool                  `json:"running"`
	TeamSize    int                   `json:"team_size"`
	Makespan    int                   `json:"makespan"`
	CurrentTS   int                   `json:"current_timestep"`
	TickRateMs  int                   `json:"tick_rate_ms"`
	OutputFile  string                `json:"output_file"`
	Positions   []mapf.AgentState     `json:"positions,omitempty"`
}

// StartSimulation bat dau mo phong MAPF.
// API #58 POST simulate/start
func (s *FlowService) StartSimulation(mapID uint32, outputFile string, tickRateMs int) (*SimulationInfo, error) {
	// Kiem tra da co phien dang chay chua
	if s.manager.IsRunning() {
		return nil, fmt.Errorf("a simulation is already running")
	}

	// Load va bat dau mo phong
	if err := s.manager.Start(outputFile, tickRateMs); err != nil {
		return nil, err
	}

	teamSize, makespan, currentTS, actualTickRate, _ := s.manager.GetInfo()

	// Luu vao DB
	run := &schema.SimulationRun{
		MapID:      mapID,
		OutputFile: outputFile,
		TeamSize:   teamSize,
		Makespan:   makespan,
		TickRateMs: actualTickRate,
		Status:     schema.SimulationRunning,
		StartedAt:  time.Now(),
	}
	if err := s.repo.CreateSimulationRun(run); err != nil {
		s.manager.Stop()
		return nil, fmt.Errorf("cannot save simulation run: %w", err)
	}

	return &SimulationInfo{
		RunID:      run.RunID,
		Running:    true,
		TeamSize:   teamSize,
		Makespan:   makespan,
		CurrentTS:  currentTS,
		TickRateMs: actualTickRate,
		OutputFile: outputFile,
	}, nil
}

// StopSimulation dung mo phong.
// API #59 POST simulate/stop
func (s *FlowService) StopSimulation() error {
	if !s.manager.IsRunning() {
		return fmt.Errorf("no simulation is running")
	}

	s.manager.Stop()

	// Cap nhat DB
	run, err := s.repo.GetRunningSimulation()
	if err == nil && run != nil {
		s.repo.UpdateSimulationStatus(run.RunID, schema.SimulationStopped)
	}

	return nil
}

// SimulationStatus lay trang thai mo phong hien tai.
// API #60 GET simulate/status
func (s *FlowService) SimulationStatus() (*SimulationInfo, error) {
	running := s.manager.IsRunning()
	teamSize, makespan, currentTS, tickRateMs, outputFile := s.manager.GetInfo()

	info := &SimulationInfo{
		Running:    running,
		TeamSize:   teamSize,
		Makespan:   makespan,
		CurrentTS:  currentTS,
		TickRateMs: tickRateMs,
		OutputFile: outputFile,
	}

	// Lay vi tri tat ca agents
	if teamSize > 0 {
		info.Positions = s.manager.GetAllPositions()
	}

	// Lay run_id tu DB
	run, err := s.repo.GetRunningSimulation()
	if err == nil && run != nil {
		info.RunID = run.RunID
	}

	return info, nil
}
