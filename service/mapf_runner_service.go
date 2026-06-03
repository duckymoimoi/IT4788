package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"hospital/schema"
)

type MAPFRunRequest struct {
	Agents                int   `json:"agents"`
	Tasks                 int   `json:"tasks"`
	Steps                 int   `json:"steps"`
	Seed                  int64 `json:"seed"`
	PlanTimeLimitMs       int   `json:"plan_time_limit_ms"`
	PreprocessTimeLimitMs int   `json:"preprocess_time_limit_ms"`
	TimeoutSeconds        int   `json:"timeout_seconds"`
}

type MAPFRunStatus struct {
	JobID          string    `json:"job_id"`
	Status         string    `json:"status"`
	Message        string    `json:"message,omitempty"`
	MapID          uint32    `json:"map_id,omitempty"`
	MapName        string    `json:"map_name,omitempty"`
	Rows           int       `json:"rows,omitempty"`
	Cols           int       `json:"cols,omitempty"`
	WalkableCells  int       `json:"walkable_cells,omitempty"`
	WallCells      int       `json:"wall_cells,omitempty"`
	POICount       int       `json:"poi_count,omitempty"`
	UsablePOITasks int       `json:"usable_poi_tasks,omitempty"`
	Agents         int       `json:"agents,omitempty"`
	Tasks          int       `json:"tasks,omitempty"`
	Steps          int       `json:"steps,omitempty"`
	InputFile      string    `json:"input_file,omitempty"`
	OutputFile     string    `json:"output_file,omitempty"`
	LogTail        string    `json:"log_tail,omitempty"`
	StartedAt      time.Time `json:"started_at,omitempty"`
	FinishedAt     time.Time `json:"finished_at,omitempty"`
}

func (s *EngineService) StartMAPFRun(req MAPFRunRequest) (*MAPFRunStatus, error) {
	s.mu.Lock()
	if s.mapfJob != nil && s.mapfJob.Status == "running" {
		job := *s.mapfJob
		s.mu.Unlock()
		return &job, fmt.Errorf("MAPF job is already running")
	}

	req = normalizeMAPFRunRequest(req, s.params.MaxAgents)
	job := &MAPFRunStatus{
		JobID:     fmt.Sprintf("mapf_%d", time.Now().Unix()),
		Status:    "running",
		Message:   "MAPF solver is running",
		Agents:    req.Agents,
		Tasks:     req.Tasks,
		Steps:     req.Steps,
		StartedAt: time.Now(),
	}
	s.mapfJob = job
	s.mu.Unlock()

	go s.runMAPFJob(req, job.JobID)
	return job, nil
}

func (s *EngineService) GetMAPFRunStatus() *MAPFRunStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.mapfJob == nil {
		return &MAPFRunStatus{Status: "idle", Message: "No MAPF job has been started"}
	}
	job := *s.mapfJob
	return &job
}

func normalizeMAPFRunRequest(req MAPFRunRequest, defaultAgents int) MAPFRunRequest {
	if req.Agents <= 0 {
		req.Agents = defaultAgents
	}
	if req.Agents <= 0 {
		req.Agents = 100
	}
	if req.Tasks <= 0 {
		req.Tasks = req.Agents * 3
	}
	if req.Steps <= 0 {
		req.Steps = 500
	}
	if req.Seed == 0 {
		req.Seed = 4788
	}
	if req.PlanTimeLimitMs <= 0 {
		req.PlanTimeLimitMs = 5000
	}
	if req.PreprocessTimeLimitMs <= 0 {
		req.PreprocessTimeLimitMs = 300000
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 1200
	}
	return req
}

func (s *EngineService) runMAPFJob(req MAPFRunRequest, jobID string) {
	status, err := s.generateAndRunMAPF(req, jobID)
	if status == nil {
		status = &MAPFRunStatus{JobID: jobID, StartedAt: time.Now()}
	}
	if err != nil {
		status.Status = "failed"
		status.Message = err.Error()
	} else {
		status.Status = "succeeded"
		status.Message = "MAPF output generated and loaded"
		if loadErr := s.LoadMAPFOutput(status.OutputFile); loadErr != nil {
			status.Status = "failed"
			status.Message = fmt.Sprintf("output generated but cannot load: %v", loadErr)
		}
	}
	status.FinishedAt = time.Now()

	s.mu.Lock()
	s.mapfJob = status
	s.mu.Unlock()
}

func (s *EngineService) generateAndRunMAPF(req MAPFRunRequest, jobID string) (*MAPFRunStatus, error) {
	activeMap, err := s.mapRepo.FindActiveMap()
	if err != nil {
		return &MAPFRunStatus{JobID: jobID}, err
	}
	if activeMap == nil {
		return &MAPFRunStatus{JobID: jobID}, fmt.Errorf("active map not found")
	}

	var grid [][]int
	if err := json.Unmarshal([]byte(activeMap.GridData), &grid); err != nil {
		return &MAPFRunStatus{JobID: jobID, MapID: activeMap.MapID, MapName: activeMap.MapName}, fmt.Errorf("invalid active map grid_data: %w", err)
	}
	if len(grid) != activeMap.Rows {
		return &MAPFRunStatus{JobID: jobID, MapID: activeMap.MapID, MapName: activeMap.MapName}, fmt.Errorf("grid row mismatch: expected %d, got %d", activeMap.Rows, len(grid))
	}
	for i := range grid {
		if len(grid[i]) != activeMap.Cols {
			return &MAPFRunStatus{JobID: jobID, MapID: activeMap.MapID, MapName: activeMap.MapName}, fmt.Errorf("grid col mismatch at row %d: expected %d, got %d", i, activeMap.Cols, len(grid[i]))
		}
	}

	pois, err := s.mapRepo.FindAllPOIs(activeMap.MapID)
	if err != nil {
		return &MAPFRunStatus{JobID: jobID, MapID: activeMap.MapID, MapName: activeMap.MapName}, err
	}
	poiLocations, blockedPOIs := usablePOILocations(pois, grid, activeMap.Cols)
	if len(poiLocations) == 0 {
		return &MAPFRunStatus{JobID: jobID, MapID: activeMap.MapID, MapName: activeMap.MapName}, fmt.Errorf("active map has no walkable POI to create MAPF tasks")
	}

	agents, err := sampleAgentLocations(grid, activeMap.Cols, poiLocations, req.Agents, req.Seed)
	if err != nil {
		return &MAPFRunStatus{JobID: jobID, MapID: activeMap.MapID, MapName: activeMap.MapName}, err
	}
	tasks := repeatLocations(poiLocations, req.Tasks)
	walkable := countWalkable(grid)

	runDir := filepath.Join("data", "mapf_runs", fmt.Sprintf("%s_%d_%s", slugify(activeMap.MapName), activeMap.MapID, jobID))
	mapDir := filepath.Join(runDir, "maps")
	agentDir := filepath.Join(runDir, "agents")
	taskDir := filepath.Join(runDir, "tasks")
	if err := os.MkdirAll(mapDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return nil, err
	}

	baseName := slugify(activeMap.MapName)
	mapFile := filepath.Join(mapDir, baseName+".map")
	agentFile := filepath.Join(agentDir, fmt.Sprintf("%s_%d.agents", baseName, req.Agents))
	taskFile := filepath.Join(taskDir, fmt.Sprintf("%s_%d.tasks", baseName, req.Tasks))
	inputFile := filepath.Join(runDir, fmt.Sprintf("%s_%d.json", baseName, req.Agents))
	outputFile := filepath.Join(runDir, "output.json")
	logFile := filepath.Join(runDir, "lifelong.log")

	if err := writeMAPFMapFile(mapFile, activeMap.Rows, activeMap.Cols, grid); err != nil {
		return nil, err
	}
	if err := writeIntLines(agentFile, agents); err != nil {
		return nil, err
	}
	if err := writeIntLines(taskFile, tasks); err != nil {
		return nil, err
	}
	input := map[string]interface{}{
		"mapFile":              "maps/" + filepath.Base(mapFile),
		"agentFile":            "agents/" + filepath.Base(agentFile),
		"teamSize":             req.Agents,
		"taskFile":             "tasks/" + filepath.Base(taskFile),
		"numTasksReveal":       1,
		"simulation_steps":     req.Steps,
		"agentCounter":         req.Agents,
		"version":              "2026 LoRR",
		"agentSize":            1.0,
		"enableTaskFinishSpin": true,
	}
	rawInput, _ := json.MarshalIndent(input, "", "  ")
	if err := os.WriteFile(inputFile, rawInput, 0644); err != nil {
		return nil, err
	}

	status := &MAPFRunStatus{
		JobID:          jobID,
		Status:         "running",
		MapID:          activeMap.MapID,
		MapName:        activeMap.MapName,
		Rows:           activeMap.Rows,
		Cols:           activeMap.Cols,
		WalkableCells:  walkable,
		WallCells:      activeMap.Rows*activeMap.Cols - walkable,
		POICount:       len(pois),
		UsablePOITasks: len(poiLocations),
		Agents:         req.Agents,
		Tasks:          req.Tasks,
		Steps:          req.Steps,
		InputFile:      inputFile,
		OutputFile:     outputFile,
		StartedAt:      time.Now(),
	}
	if len(blockedPOIs) > 0 {
		status.Message = fmt.Sprintf("Skipped %d POIs on blocked cells", len(blockedPOIs))
	}

	solver := os.Getenv("MAPF_SOLVER_PATH")
	if solver == "" {
		solver = filepath.Join("mapf_solver", "lifelong")
	}
	if _, err := os.Stat(solver); err != nil {
		return status, fmt.Errorf("MAPF solver not found at %s", solver)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.TimeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, solver,
		"--inputFile", filepath.Base(inputFile),
		"--output", filepath.Base(outputFile),
		"--simulationTime", fmt.Sprintf("%d", req.Steps),
		"--planTimeLimit", fmt.Sprintf("%d", req.PlanTimeLimitMs),
		"--preprocessTimeLimit", fmt.Sprintf("%d", req.PreprocessTimeLimitMs),
		"--outputScreen", "1",
		"--logFile", filepath.Base(logFile),
	)
	cmd.Dir = runDir
	combined, err := cmd.CombinedOutput()
	logTail := tailString(string(combined)+"\n"+readFileBestEffort(logFile), 4000)
	status.LogTail = strings.TrimSpace(logTail)
	if ctx.Err() == context.DeadlineExceeded {
		return status, fmt.Errorf("MAPF solver timed out after %d seconds", req.TimeoutSeconds)
	}
	if err != nil {
		return status, fmt.Errorf("MAPF solver failed: %w", err)
	}
	if _, err := os.Stat(outputFile); err != nil {
		return status, fmt.Errorf("MAPF solver finished but output file was not created")
	}
	return status, nil
}

func usablePOILocations(pois []schema.GridPOI, grid [][]int, cols int) ([]int, []string) {
	locations := make([]int, 0, len(pois))
	blocked := make([]string, 0)
	seen := make(map[int]bool)
	for _, poi := range pois {
		row, col := poi.GridLocation/cols, poi.GridLocation%cols
		if row < 0 || row >= len(grid) || col < 0 || col >= len(grid[row]) || grid[row][col] != 0 {
			blocked = append(blocked, poi.POICode)
			continue
		}
		if !seen[poi.GridLocation] {
			seen[poi.GridLocation] = true
			locations = append(locations, poi.GridLocation)
		}
	}
	return locations, blocked
}

func sampleAgentLocations(grid [][]int, cols int, poiLocations []int, count int, seed int64) ([]int, error) {
	blocked := make(map[int]bool, len(poiLocations))
	for _, loc := range poiLocations {
		blocked[loc] = true
	}
	walkable := make([]int, 0)
	for r := range grid {
		for c, cell := range grid[r] {
			loc := r*cols + c
			if cell == 0 && !blocked[loc] {
				walkable = append(walkable, loc)
			}
		}
	}
	if len(walkable) < count {
		return nil, fmt.Errorf("not enough walkable cells for %d agents, got %d", count, len(walkable))
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(walkable), func(i, j int) {
		walkable[i], walkable[j] = walkable[j], walkable[i]
	})
	return walkable[:count], nil
}

func repeatLocations(locations []int, count int) []int {
	result := make([]int, 0, count)
	for len(result) < count {
		for _, loc := range locations {
			result = append(result, loc)
			if len(result) == count {
				break
			}
		}
	}
	return result
}

func writeMAPFMapFile(path string, rows, cols int, grid [][]int) error {
	var b strings.Builder
	b.WriteString("type octile\n")
	b.WriteString(fmt.Sprintf("height %d\n", rows))
	b.WriteString(fmt.Sprintf("width %d\n", cols))
	b.WriteString("map\n")
	for _, row := range grid {
		for _, cell := range row {
			if cell == 1 {
				b.WriteByte('@')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func writeIntLines(path string, values []int) error {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d\n", len(values)))
	for _, value := range values {
		b.WriteString(fmt.Sprintf("%d\n", value))
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func countWalkable(grid [][]int) int {
	count := 0
	for _, row := range grid {
		for _, cell := range row {
			if cell == 0 {
				count++
			}
		}
	}
	return count
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	value = strings.Trim(re.ReplaceAllString(value, "_"), "_")
	if value == "" {
		return "active_map"
	}
	return value
}

func readFileBestEffort(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func tailString(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[len(value)-max:]
}
