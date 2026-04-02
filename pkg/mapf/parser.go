package mapf

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ========================================
// PARSER  - Doc output.json cua MAPF solver
// Format: LoRR 2024 (rotation model MAPF_T)
// ========================================

// OutputJSON cau truc output.json cua MAPF solver.
type OutputJSON struct {
	ActionModel     string          `json:"actionModel"`
	Version         string          `json:"version"`
	TeamSize        int             `json:"teamSize"`
	NumTaskFinished int             `json:"numTaskFinished"`
	Makespan        int             `json:"makespan"`
	Start           [][]interface{} `json:"start"`
	ActualPaths     []string        `json:"actualPaths"`
	PlannerPaths    []string        `json:"plannerPaths"`
	ActualSchedule  []string        `json:"actualSchedule"`
	Events          [][]int         `json:"events"`
	Tasks           [][]interface{} `json:"tasks"`
}

// Orientation huong mat cua agent.
type Orientation int

const (
	OrientEast  Orientation = 0 // ->
	OrientSouth Orientation = 1 // ↓
	OrientWest  Orientation = 2 // <-
	OrientNorth Orientation = 3 // ↑
)

// AgentState vi tri + huong cua agent tai 1 thoi diem.
type AgentState struct {
	Row         int
	Col         int
	Location    int
	Orientation Orientation
}

// AgentTrajectory quy dao di chuyen cua 1 agent qua cac timestep.
type AgentTrajectory struct {
	AgentID  int
	StartRow int
	StartCol int
	States   []AgentState // states[t] = vi tri tai timestep t
}

// TaskInfo thong tin 1 task tu output.json.
type TaskInfo struct {
	TaskID    int
	RevealTS  int // timestep task duoc reveal
	Row       int
	Col       int
	Location  int
}

// MAPFResult ket qua parse tu output.json.
type MAPFResult struct {
	TeamSize        int
	NumTaskFinished int
	Makespan        int
	Trajectories    []AgentTrajectory
	Tasks           []TaskInfo
}

// ParseOutputJSON doc va parse file output.json.
// Tra ve MAPFResult chua trajectories cho tat ca agents.
func ParseOutputJSON(filepath string) (*MAPFResult, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("cannot read output file: %w", err)
	}

	var output OutputJSON
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("cannot parse output JSON: %w", err)
	}

	if len(output.Start) != output.TeamSize {
		return nil, fmt.Errorf("teamSize=%d but start has %d entries", output.TeamSize, len(output.Start))
	}

	if len(output.ActualPaths) != output.TeamSize {
		return nil, fmt.Errorf("teamSize=%d but actualPaths has %d entries", output.TeamSize, len(output.ActualPaths))
	}

	// Parse trajectories
	trajectories := make([]AgentTrajectory, output.TeamSize)

	for i := 0; i < output.TeamSize; i++ {
		startRow, startCol, startOrient, err := parseStart(output.Start[i])
		if err != nil {
			return nil, fmt.Errorf("agent %d start error: %w", i, err)
		}

		actions := parseActions(output.ActualPaths[i])
		states := computeStates(startRow, startCol, startOrient, actions)

		trajectories[i] = AgentTrajectory{
			AgentID:  i,
			StartRow: startRow,
			StartCol: startCol,
			States:   states,
		}
	}

	// Parse tasks
	tasks := make([]TaskInfo, 0, len(output.Tasks))
	for _, t := range output.Tasks {
		ti, err := parseTask(t)
		if err != nil {
			continue // skip invalid tasks
		}
		tasks = append(tasks, ti)
	}

	return &MAPFResult{
		TeamSize:        output.TeamSize,
		NumTaskFinished: output.NumTaskFinished,
		Makespan:        output.Makespan,
		Trajectories:    trajectories,
		Tasks:           tasks,
	}, nil
}

// parseStart doc [row, col, "E"] tu start array.
func parseStart(s []interface{}) (int, int, Orientation, error) {
	if len(s) < 3 {
		return 0, 0, 0, fmt.Errorf("start entry needs 3 elements, got %d", len(s))
	}

	row, err := toInt(s[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid row: %w", err)
	}
	col, err := toInt(s[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid col: %w", err)
	}

	orientStr, ok := s[2].(string)
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid orientation type")
	}

	orient := parseOrientation(orientStr)
	return row, col, orient, nil
}

// parseTask doc [task_id, reveal_ts, [row, col]] tu tasks array.
func parseTask(t []interface{}) (TaskInfo, error) {
	if len(t) < 3 {
		return TaskInfo{}, fmt.Errorf("task needs 3 elements")
	}

	taskID, err := toInt(t[0])
	if err != nil {
		return TaskInfo{}, err
	}
	revealTS, err := toInt(t[1])
	if err != nil {
		return TaskInfo{}, err
	}

	loc, ok := t[2].([]interface{})
	if !ok || len(loc) < 2 {
		return TaskInfo{}, fmt.Errorf("invalid task location")
	}
	row, err := toInt(loc[0])
	if err != nil {
		return TaskInfo{}, err
	}
	col, err := toInt(loc[1])
	if err != nil {
		return TaskInfo{}, err
	}

	return TaskInfo{
		TaskID:   taskID,
		RevealTS: revealTS,
		Row:      row,
		Col:      col,
	}, nil
}

// Action loai hanh dong cua agent.
type Action int

const (
	ActionForward Action = iota // F - di thang
	ActionRight                 // R - quay phai (clockwise)
	ActionLeft                  // C - quay trai (counter-clockwise)
	ActionWait                  // W - cho
)

// parseActions chuyen chuoi "F,R,C,F,W,..." thanh list actions.
func parseActions(s string) []Action {
	parts := strings.Split(s, ",")
	actions := make([]Action, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch p {
		case "F":
			actions = append(actions, ActionForward)
		case "R":
			actions = append(actions, ActionRight)
		case "C":
			actions = append(actions, ActionLeft)
		case "W":
			actions = append(actions, ActionWait)
		default:
			actions = append(actions, ActionWait)
		}
	}
	return actions
}

// parseOrientation chuyen "E"/"S"/"W"/"N" sang Orientation.
func parseOrientation(s string) Orientation {
	switch strings.ToUpper(s) {
	case "E":
		return OrientEast
	case "S":
		return OrientSouth
	case "W":
		return OrientWest
	case "N":
		return OrientNorth
	default:
		return OrientEast
	}
}

// computeStates mo phong di chuyen tu vi tri bat dau qua list actions.
// Tra ve chuoi AgentState cho moi timestep.
func computeStates(startRow, startCol int, orient Orientation, actions []Action) []AgentState {
	states := make([]AgentState, 0, len(actions)+1)

	// State tai timestep 0
	curRow, curCol, curOrient := startRow, startCol, orient
	states = append(states, AgentState{
		Row:         curRow,
		Col:         curCol,
		Orientation: curOrient,
	})

	// delta cho moi huong: E->(0,1), S->(1,0), W->(0,-1), N->(-1,0)
	dr := [4]int{0, 1, 0, -1}  // row delta theo orient
	dc := [4]int{1, 0, -1, 0}  // col delta theo orient

	for _, action := range actions {
		switch action {
		case ActionForward:
			curRow += dr[curOrient]
			curCol += dc[curOrient]
		case ActionRight:
			curOrient = (curOrient + 1) % 4
		case ActionLeft:
			curOrient = (curOrient + 3) % 4 // +3 = -1 mod 4
		case ActionWait:
			// khong lam gi
		}

		states = append(states, AgentState{
			Row:         curRow,
			Col:         curCol,
			Orientation: curOrient,
		})
	}

	// Set location cho tat ca states
	// (can cols de tinh, nhung parser khong co grid info,
	// caller se set location sau khi parse)
	return states
}

// SetLocations cap nhat Location field cho tat ca states dua tren grid cols.
func (t *AgentTrajectory) SetLocations(cols int) {
	for i := range t.States {
		t.States[i].Location = t.States[i].Row*cols + t.States[i].Col
	}
}

// SetAllLocations cap nhat Location cho tat ca trajectories va tasks.
func (r *MAPFResult) SetAllLocations(cols int) {
	for i := range r.Trajectories {
		r.Trajectories[i].SetLocations(cols)
	}
	for i := range r.Tasks {
		r.Tasks[i].Location = r.Tasks[i].Row*cols + r.Tasks[i].Col
	}
}

// GetPositionsAtTimestep tra ve vi tri cua tat ca agents tai timestep t.
func (r *MAPFResult) GetPositionsAtTimestep(t int) []AgentState {
	positions := make([]AgentState, r.TeamSize)
	for i, traj := range r.Trajectories {
		if t < len(traj.States) {
			positions[i] = traj.States[t]
		} else if len(traj.States) > 0 {
			// Agent da dung: tra ve vi tri cuoi cung
			positions[i] = traj.States[len(traj.States)-1]
		}
	}
	return positions
}

// helper: chuyen interface{} (float64 tu JSON) sang int.
func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}
