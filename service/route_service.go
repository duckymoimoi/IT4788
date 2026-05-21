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
	gridPath  string
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
	routingGrid, err := s.routingGrid(grid, startLoc, destLoc)
	if err != nil {
		return nil, err
	}

	// Chay Dijkstra
	result := mapf.DijkstraWithSpeed(routingGrid, startLoc, destLoc, mode.SpeedFactor)
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
			GridLocation: p.Row*routingGrid.Cols + p.Col,
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
		TotalDistance: preview.Distance,
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

// OrderMultiRoute tim duong qua nhieu diem (theo thu tu).
// API: POST order_multi
func (s *RouteService) OrderMultiRoute(userID uint64, startLoc int, destLocs []int, modeID string) (*schema.Route, []schema.RoutePath, error) {
	if len(destLocs) == 0 {
		return nil, nil, fmt.Errorf("no destinations provided")
	}

	var totalDistance float64
	var totalTime float64
	var allSteps []StepInfo

	currentStart := startLoc
	stepCounter := 0

	for i, target := range destLocs {
		if currentStart == target {
			continue // skip if same
		}

		preview, err := s.PreviewRoute(currentStart, target, modeID)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot find path to target %d: %w", target, err)
		}

		totalDistance += preview.Distance
		totalTime += preview.EstimatedTime

		// Ghep step, bo qua step[0] neu khong phai la chang dau tien de tranh trung lap node giao
		startStepIdx := 0
		if i > 0 && len(allSteps) > 0 {
			startStepIdx = 1
		}

		for j := startStepIdx; j < len(preview.Steps); j++ {
			step := preview.Steps[j]
			step.StepOrder = stepCounter
			allSteps = append(allSteps, step)
			stepCounter++
		}

		currentStart = target
	}

	if len(allSteps) == 0 {
		return nil, nil, fmt.Errorf("empty route or all destinations are same as start")
	}

	routeID := uuid.New().String()
	route := &schema.Route{
		RouteID:       routeID,
		UserID:        userID,
		ModeID:        modeID,
		StartLocation: startLoc,
		DestLocation:  destLocs[len(destLocs)-1],
		RouteMode:     schema.RouteModeDijkstra,
		TotalDistance: totalDistance,
		EstimatedTime: totalTime,
		Status:        schema.RouteStatusActive,
	}

	paths := make([]schema.RoutePath, len(allSteps))
	for i, step := range allSteps {
		paths[i] = schema.RoutePath{
			RouteID:      routeID,
			StepOrder:    step.StepOrder,
			GridRow:      step.GridRow,
			GridCol:      step.GridCol,
			GridLocation: step.GridLocation,
			Instruction:  generateInstruction(i, allSteps),
			VoiceText:    getVoiceKey(i, allSteps),
		}
	}

	if err := s.repo.CreateRouteWithPaths(route, paths); err != nil {
		return nil, nil, fmt.Errorf("cannot create route: %w", err)
	}

	return route, paths, nil
}

// OrderUnorderedRoute tim duong qua nhieu diem (nearest-neighbor).
// API: POST order_unordered
func (s *RouteService) OrderUnorderedRoute(userID uint64, startLoc int, destLocs []int, modeID string) (*schema.Route, []schema.RoutePath, error) {
	if len(destLocs) == 0 {
		return nil, nil, fmt.Errorf("no destinations provided")
	}

	// Copy mang de khong lam hong mang goc
	unvisited := make([]int, len(destLocs))
	copy(unvisited, destLocs)

	var orderedTargets []int
	currentStart := startLoc

	for len(unvisited) > 0 {
		bestIdx := -1
		bestDist := -1.0
		// Tim diem gan nhat
		for i, target := range unvisited {
			if currentStart == target {
				bestIdx = i
				bestDist = 0
				break
			}
			preview, err := s.PreviewRoute(currentStart, target, modeID)
			if err == nil {
				if bestIdx == -1 || preview.Distance < bestDist {
					bestIdx = i
					bestDist = preview.Distance
				}
			}
		}

		if bestIdx == -1 {
			return nil, nil, fmt.Errorf("cannot find path to remaining targets")
		}

		orderedTargets = append(orderedTargets, unvisited[bestIdx])
		currentStart = unvisited[bestIdx]

		// Remove tu unvisited
		unvisited = append(unvisited[:bestIdx], unvisited[bestIdx+1:]...)
	}

	// Goi lai OrderMultiRoute voi danh sach da sap xep
	return s.OrderMultiRoute(userID, startLoc, orderedTargets, modeID)
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
		RouteID:        routeID,
		CurrentStep:    currentStep,
		RemainingSteps: len(remaining),
		RemainingDist:  remainingDist,
		EstimatedTime:  eta,
		SpeedFactor:    mode.SpeedFactor,
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
	mapPath := s.gridMapPath()
	s.mu.RLock()
	if s.gridCache != nil && s.gridPath == mapPath {
		defer s.mu.RUnlock()
		return s.gridCache, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	// Double check
	if s.gridCache != nil && s.gridPath == mapPath {
		return s.gridCache, nil
	}

	grid, err := mapf.LoadGridMap(mapPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load grid from %s: %w", mapPath, err)
	}
	s.gridCache = grid
	s.gridPath = mapPath
	return grid, nil
}

// ClearGridCache xoa cache grid (dung khi admin upload map moi).
func (s *RouteService) ClearGridCache() {
	s.mu.Lock()
	s.gridCache = nil
	s.gridPath = ""
	s.mu.Unlock()
}

// routingGrid blocks active POI cells so paths do not pass through rooms or
// service points. The requested start/destination cells stay open because a
// route is allowed to begin or end at that POI.
func (s *RouteService) routingGrid(base *mapf.GridMap, startLoc, destLoc int) (*mapf.GridMap, error) {
	grid := base.Clone()
	if grid == nil {
		return nil, fmt.Errorf("grid is not loaded")
	}

	pois, err := s.activePOIsForGrid(base.Name)
	if err != nil {
		return nil, err
	}

	for _, poi := range pois {
		loc := grid.ToLocation(poi.GridRow, poi.GridCol)
		if loc == startLoc || loc == destLoc {
			continue
		}
		grid.SetCell(poi.GridRow, poi.GridCol, 1)
	}

	return grid, nil
}

func (s *RouteService) activePOIsForGrid(mapPath string) ([]schema.GridPOI, error) {
	maps, err := s.mapRepo.FindAllMaps()
	if err != nil {
		return nil, err
	}
	if len(maps) == 0 {
		return nil, nil
	}

	mapID := maps[len(maps)-1].MapID
	for i := len(maps) - 1; i >= 0; i-- {
		if maps[i].MapFilePath == mapPath {
			mapID = maps[i].MapID
			return s.mapRepo.FindAllPOIs(mapID)
		}
	}
	for i := len(maps) - 1; i >= 0; i-- {
		if maps[i].MapFilePath != "" {
			mapID = maps[i].MapID
			break
		}
	}
	return s.mapRepo.FindAllPOIs(mapID)
}

func (s *RouteService) gridMapPath() string {
	if mapPath := os.Getenv("GRID_MAP_PATH"); mapPath != "" {
		return mapPath
	}
	maps, err := s.mapRepo.FindAllMaps()
	if err == nil && len(maps) > 0 {
		// FindAllMaps is ascending by map_id; if old data has multiple active maps,
		// prefer the latest active map instead of silently using the seed map.
		for i := len(maps) - 1; i >= 0; i-- {
			if maps[i].MapFilePath != "" {
				return maps[i].MapFilePath
			}
		}
	}
	return "data/warehouse_small.map"
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
	StepOrder    int `json:"step_order"`
	GridRow      int `json:"grid_row"`
	GridCol      int `json:"grid_col"`
	GridLocation int `json:"grid_location"`
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
