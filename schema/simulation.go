package schema

import "time"

// ========================================
// SIMULATION  - MAPF simulation runs
// Slice 5 (phan mo phong)
// ========================================

// SimulationStatus trang thai phien mo phong.
type SimulationStatus string

const (
	SimulationRunning SimulationStatus = "running"
	SimulationStopped SimulationStatus = "stopped"
	SimulationDone    SimulationStatus = "done"
)

// SimulationRun phien mo phong MAPF.
// Moi phien load output.json, khoi tao AgentManager voi tickRate.
// Bang: simulation_runs [T20]
type SimulationRun struct {
	RunID         uint64           `gorm:"primaryKey;autoIncrement;column:run_id" json:"run_id"`
	MapID         uint32           `gorm:"not null;column:map_id" json:"map_id"`
	OutputFile    string           `gorm:"not null;size:255;column:output_file" json:"output_file"`
	TeamSize      int              `gorm:"not null;column:team_size" json:"team_size"`
	Makespan      int              `gorm:"not null;column:makespan" json:"makespan"`
	TickRateMs    int              `gorm:"not null;default:1000;column:tick_rate_ms" json:"tick_rate_ms"`
	TasksFinished int              `gorm:"not null;default:0;column:tasks_finished" json:"tasks_finished"`
	Status        SimulationStatus `gorm:"not null;default:running;index;column:status" json:"status"`
	StartedAt     time.Time        `gorm:"not null;autoCreateTime;column:started_at" json:"started_at"`
	EndedAt       *time.Time       `gorm:"column:ended_at"`

	// Belongs-to
	GridMap *GridMap `gorm:"foreignKey:MapID;references:MapID"`
}

func (SimulationRun) TableName() string {
	return "simulation_runs"
}

// PatientAgent gan benh nhan vao agent MAPF.
// Moi benh nhan chi co the gan vao 1 agent tai 1 thoi diem
// (released_at != NULL thi agent da duoc giai phong).
// Bang: patient_agents [T21]
type PatientAgent struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RunID      uint64     `gorm:"not null;index;column:run_id" json:"run_id"`
	UserID     uint64     `gorm:"not null;index;column:user_id" json:"user_id"`
	AgentID    int        `gorm:"not null;column:agent_id" json:"agent_id"`
	AssignedAt time.Time  `gorm:"not null;autoCreateTime;column:assigned_at" json:"assigned_at"`
	ReleasedAt *time.Time `gorm:"column:released_at"`

	// Belongs-to
	SimulationRun *SimulationRun `gorm:"foreignKey:RunID;references:RunID"`
	User          *User          `gorm:"foreignKey:UserID;references:UserID"`
}

func (PatientAgent) TableName() string {
	return "patient_agents"
}
