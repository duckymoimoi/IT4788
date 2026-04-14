package service

import (
	"fmt"
	"os"
	"sync"

	"hospital/pkg/mapf"
	"hospital/repository"
	"hospital/schema"

	"github.com/google/uuid"
)

// RouteService xu ly logic nghiep vu cho Route module.
// Tich hop Dijkstra de tim duong real-time.
type RouteService struct {
	repo      *repository.RouteRepo
	mapRepo   *repository.MapRepo
	gridOnce  sync.Once
	gridCache *mapf.GridMap
	gridErr   error
	mu        sync.RWMutex
}

func NewRouteService(repo *repository.RouteRepo, mapRepo *repository.MapRepo) *RouteService {
	return &RouteService{
		repo:    repo,
		mapRepo: mapRepo,
	}
}

// ========================================
// TRAVEL MODES (Slice 2)
// ========================================

// GetAllModes tra ve tat ca phuong thuc di chuyen.
func (s *RouteService) GetAllModes() ([]schema.TravelMode, error) {
	return s.repo.FindAllModes()
}

// ========================================
// ROUTE CORE (Slice 3)
// ========================================

// PreviewRoute tim duong bang Dijkstra, tra ve path KHONG luu DB.
// API #37 POST preview
func (s *RouteService) PreviewRoute(startLoc, destLoc int, modeID string) (*PreviewResult, error) {
	// Lay speed factor
	mode, err := s.repo.FindModeByID(modeID)
	if err != nil {
		return nil, fmt.Errorf("travel mode '%s' not found", modeID)
	}

	// Load grid (cache)
	grid, err := s.getGrid()
	if err != nil {
		return nil, err
	}

	// Chay Dijkstra
	result := mapf.DijkstraWithSpeed(grid, startLoc, destLoc, mode.SpeedFactor)
	if !result.Found {
		return nil, fmt.Errorf("no path found from %d to %d", startLoc, destLoc)
	}

	// Chuyen path sang response
	steps := make([]StepInfo, len(result.Path))
	for i, p := range result.Path {
		steps[i] = StepInfo{
			StepOrder:    i,
			GridRow:      p.Row,
			GridCol:      p.Col,
			GridLocation: p.Row*grid.Cols + p.Col,
		}
	}

	return &PreviewResult{
		Distance:      result.Distance,
		EstimatedTime: result.Distance / mode.SpeedFactor,
		Steps:         steps,
		ModeID:        modeID,
		SpeedFactor:   mode.SpeedFactor,
	}, nil
}

// OrderRoute tim duong + luu vao DB.
// API #31 POST ordered
func (s *RouteService) OrderRoute(userID uint64, startLoc, destLoc int, modeID string) (*schema.Route, []schema.RoutePath, error) {
	// Validate: start != dest
	if startLoc == destLoc {
		return nil, nil, fmt.Errorf("start and destination cannot be the same")
	}

	// Preview truoc
	preview, err := s.PreviewRoute(startLoc, destLoc, modeID)
	if err != nil {
		return nil, nil, err
	}

	// Tao route + paths trong 1 transaction
	routeID := uuid.New().String()
	route := &schema.Route{
		RouteID:       routeID,
		UserID:        userID,
		ModeID:        modeID,
		StartLocation: startLoc,
		DestLocation:  destLoc,
		RouteMode:     schema.RouteModeDijkstra,
		TotalDistance:  preview.Distance,
		EstimatedTime: preview.EstimatedTime,
		Status:        schema.RouteStatusActive,
	}

	paths := make([]schema.RoutePath, len(preview.Steps))
	for i, step := range preview.Steps {
		paths[i] = schema.RoutePath{
			RouteID:      routeID,
			StepOrder:    step.StepOrder,
			GridRow:      step.GridRow,
			GridCol:      step.GridCol,
			GridLocation: step.GridLocation,
			Instruction:  generateInstruction(i, preview.Steps),
			VoiceText:    getVoiceKey(i, preview.Steps),
		}
	}

	if err := s.repo.CreateRouteWithPaths(route, paths); err != nil {
		return nil, nil, fmt.Errorf("cannot create route: %w", err)
	}

	return route, paths, nil
}

// GetSteps lay cac buoc di cua route.
// API #36 GET get_steps
func (s *RouteService) GetSteps(routeID string) ([]schema.RoutePath, error) {
	return s.repo.FindPathsByRouteID(routeID)
}

// GetETA tinh thoi gian con lai.
// API #38 POST get_eta
func (s *RouteService) GetETA(routeID string, currentStep int) (*ETAResult, error) {
	route, err := s.repo.FindRouteByID(routeID)
	if err != nil {
		return nil, fmt.Errorf("route not found")
	}

	mode, err := s.repo.FindModeByID(route.ModeID)
	if err != nil {
		return nil, fmt.Errorf("travel mode not found")
	}

	remaining, err := s.repo.FindRemainingSteps(routeID, currentStep)
	if err != nil {
		return nil, err
	}

	remainingDist := float64(len(remaining))
	eta := remainingDist / mode.SpeedFactor

	return &ETAResult{
		RouteID:         routeID,
		CurrentStep:     currentStep,
		RemainingSteps:  len(remaining),
		RemainingDist:   remainingDist,
		EstimatedTime:   eta,
		SpeedFactor:     mode.SpeedFactor,
	}, nil
}

// ========================================
// AUTHORIZATION HELPER
// ========================================

// VerifyRouteOwner kiem tra routeID co thuoc userID khong.
func (s *RouteService) VerifyRouteOwner(routeID string, userID uint64) error {
	route, err := s.repo.FindRouteByID(routeID)
	if err != nil {
		return fmt.Errorf("route not found")
	}
	if route.UserID != userID {
		return fmt.Errorf("unauthorized: route does not belong to user")
	}
	return nil
}

// ========================================
// ROUTE MO RONG (Slice 4)
// ========================================

// GetActiveRoute lay route dang active cua user.
// API #34 GET get_active
func (s *RouteService) GetActiveRoute(userID uint64) (*schema.Route, []schema.RoutePath, error) {
	route, err := s.repo.FindActiveRoute(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("no active route")
	}
	paths, err := s.repo.FindPathsByRouteID(route.RouteID)
	if err != nil {
		return nil, nil, err
	}
	return route, paths, nil
}

// CancelRoute huy lo trinh.
// API #35 POST cancel
func (s *RouteService) CancelRoute(routeID string, userID uint64) error {
	route, err := s.repo.FindRouteByID(routeID)
	if err != nil {
		return fmt.Errorf("route not found")
	}
	if route.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	if route.Status != schema.RouteStatusActive {
		return fmt.Errorf("route is not active")
	}
	return s.repo.UpdateRouteStatus(routeID, schema.RouteStatusCancelled)
}

// RecalculateRoute tinh lai lo trinh tu vi tri hien tai.
// API #33 POST recalculate
func (s *RouteService) RecalculateRoute(routeID string, userID uint64, currentLocation int) (*schema.Route, []schema.RoutePath, error) {
	route, err := s.repo.FindRouteByID(routeID)
	if err != nil {
		return nil, nil, fmt.Errorf("route not found")
	}
	if route.UserID != userID {
		return nil, nil, fmt.Errorf("unauthorized")
	}

	// Tim duong moi tu vi tri hien tai
	preview, err := s.PreviewRoute(currentLocation, route.DestLocation, route.ModeID)
	if err != nil {
		return nil, nil, err
	}

	// Tao path moi
	paths := make([]schema.RoutePath, len(preview.Steps))
	for i, step := range preview.Steps {
		paths[i] = schema.RoutePath{
			RouteID:      routeID,
			StepOrder:    step.StepOrder,
			GridRow:      step.GridRow,
			GridCol:      step.GridCol,
			GridLocation: step.GridLocation,
			Instruction:  generateInstruction(i, preview.Steps),
			VoiceText:    getVoiceKey(i, preview.Steps),
		}
	}

	updates := map[string]interface{}{
		"start_location": currentLocation,
		"total_distance": preview.Distance,
		"estimated_time": preview.EstimatedTime,
	}

	// Transaction: xoa paths cu + tao moi + cap nhat route
	if err := s.repo.RecalculateRouteTransaction(routeID, paths, updates); err != nil {
		return nil, nil, err
	}

	return route, paths, nil
}

// PassNode ghi nhan benh nhan da di qua node.
// API #43 POST pass_node
func (s *RouteService) PassNode(routeID string, gridLocation int) error {
	node := &schema.RouteHistoryNode{
		RouteID:      routeID,
		GridLocation: gridLocation,
		Status:       schema.HistoryNodeArrived,
	}
	if err := s.repo.CreateHistoryNode(node); err != nil {
		return err
	}
	// Cap nhat cac node truoc thanh "passed"
	return s.repo.UpdatePreviousNodes(routeID, node.HistoryID)
}

// GetNextSteps lay cac buoc tiep theo tu buoc hien tai.
// API #44 GET get_next
func (s *RouteService) GetNextSteps(routeID string, currentStep int, limit int) ([]schema.RoutePath, error) {
	steps, err := s.repo.FindRemainingSteps(routeID, currentStep)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(steps) > limit {
		steps = steps[:limit]
	}
	return steps, nil
}

// GetHistory lay lich su route cua user.
// API #39 GET get_history
func (s *RouteService) GetHistory(userID uint64, page, limit int) ([]schema.Route, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.repo.FindRouteHistory(userID, page, limit)
}

// ClearHistory xoa mem lich su.
// API #40 DELETE clear_history
func (s *RouteService) ClearHistory(userID uint64) error {
	return s.repo.SoftDeleteRouteHistory(userID)
}

// ShareRoute tao link chia se.
// API #41 POST share
func (s *RouteService) ShareRoute(routeID string, receiverPhone string) (*schema.RouteShare, error) {
	share := &schema.RouteShare{
		RouteID:       routeID,
		ShareURL:      uuid.New().String(),
		ReceiverPhone: receiverPhone,
	}
	if err := s.repo.CreateShare(share); err != nil {
		return nil, err
	}
	return share, nil
}

// RatePath danh gia lo trinh.
// API #42 POST rate_path
func (s *RouteService) RatePath(routeID string, rating int, comment string, isAccurate *bool) error {
	fb := &schema.RouteFeedback{
		RouteID:    routeID,
		Rating:     rating,
		Comment:    comment,
		IsAccurate: isAccurate,
	}
	return s.repo.CreateFeedback(fb)
}

// ========================================
// HELPERS
// ========================================

// getGrid load va cache grid map (thread-safe voi sync.Once).
func (s *RouteService) getGrid() (*mapf.GridMap, error) {
	s.mu.RLock()
	if s.gridCache != nil {
		defer s.mu.RUnlock()
		return s.gridCache, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	// Double check
	if s.gridCache != nil {
		return s.gridCache, nil
	}

	// Doc path tu env, mac dinh data/warehouse_small.map
	mapPath := os.Getenv("GRID_MAP_PATH")
	if mapPath == "" {
		mapPath = "data/warehouse_small.map"
	}

	grid, err := mapf.LoadGridMap(mapPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load grid from %s: %w", mapPath, err)
	}
	s.gridCache = grid
	return grid, nil
}

// ClearGridCache xoa cache grid (dung khi admin upload map moi).
func (s *RouteService) ClearGridCache() {
	s.mu.Lock()
	s.gridCache = nil
	s.mu.Unlock()
}

// generateInstruction tao chi dan don gian cho moi buoc.
func generateInstruction(stepIdx int, steps []StepInfo) string {
	if stepIdx == 0 {
		return "Bắt đầu tại vị trí hiện tại"
	}
	if stepIdx == len(steps)-1 {
		return "Đã đến đích"
	}

	prev := steps[stepIdx-1]
	curr := steps[stepIdx]

	dr := curr.GridRow - prev.GridRow
	dc := curr.GridCol - prev.GridCol

	// Kiểm tra xem hướng có thay đổi so với bước trước không
	if stepIdx >= 2 {
		prevPrev := steps[stepIdx-2]
		prevDr := prev.GridRow - prevPrev.GridRow
		prevDc := prev.GridCol - prevPrev.GridCol
		if dr == prevDr && dc == prevDc {
			return "Tiếp tục đi thẳng"
		}
	}

	switch {
	case dr == -1:
		return "Đi lên (Bắc)"
	case dr == 1:
		return "Đi xuống (Nam)"
	case dc == 1:
		return "Rẽ phải (Đông)"
	case dc == -1:
		return "Rẽ trái (Tây)"
	default:
		return "Tiếp tục đi thẳng"
	}
}

// getVoiceKey tra ve key cua file audio tuong ung voi buoc di.
// Client dung key nay de map voi danh sach file tu API get_voice_files.
func getVoiceKey(stepIdx int, steps []StepInfo) string {
	if stepIdx == 0 {
		return "go_straight"
	}
	if stepIdx == len(steps)-1 {
		return "arrived"
	}

	prev := steps[stepIdx-1]
	curr := steps[stepIdx]

	dr := curr.GridRow - prev.GridRow
	dc := curr.GridCol - prev.GridCol

	// Nếu hướng không đổi → đi thẳng
	if stepIdx >= 2 {
		prevPrev := steps[stepIdx-2]
		prevDr := prev.GridRow - prevPrev.GridRow
		prevDc := prev.GridCol - prevPrev.GridCol
		if dr == prevDr && dc == prevDc {
			return "go_straight"
		}
	}

	switch {
	case dr == -1:
		return "go_straight" // đi lên = tiến thẳng
	case dr == 1:
		return "go_straight" // đi xuống = tiến thẳng
	case dc == 1:
		return "turn_right"
	case dc == -1:
		return "turn_left"
	default:
		return "go_straight"
	}
}

// PreviewResult ket qua preview route.
type PreviewResult struct {
	Distance      float64    `json:"distance"`
	EstimatedTime float64    `json:"estimated_time"`
	Steps         []StepInfo `json:"steps"`
	ModeID        string     `json:"mode_id"`
	SpeedFactor   float64    `json:"speed_factor"`
}

// StepInfo thong tin 1 buoc di.
type StepInfo struct {
	StepOrder    int    `json:"step_order"`
	GridRow      int    `json:"grid_row"`
	GridCol      int    `json:"grid_col"`
	GridLocation int    `json:"grid_location"`
}

// ETAResult ket qua tinh ETA.
type ETAResult struct {
	RouteID        string  `json:"route_id"`
	CurrentStep    int     `json:"current_step"`
	RemainingSteps int     `json:"remaining_steps"`
	RemainingDist  float64 `json:"remaining_distance"`
	EstimatedTime  float64 `json:"estimated_time"`
	SpeedFactor    float64 `json:"speed_factor"`
}
