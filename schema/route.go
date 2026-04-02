package schema

import "time"

// ========================================
// ROUTE  - Tim duong + Lich su lo trinh
// Slice 2 (TravelMode) + Slice 3 (Route, RoutePath)
// + Slice 4 (RouteHistoryNode, RouteShare, RouteFeedback)
// ========================================

// RouteMode che do tim duong.
type RouteMode string

const (
	RouteModeDijkstra RouteMode = "dijkstra"
	RouteModeMapf     RouteMode = "mapf"
)

// RouteStatus trang thai lo trinh.
type RouteStatus string

const (
	RouteStatusActive    RouteStatus = "active"
	RouteStatusCompleted RouteStatus = "completed"
	RouteStatusCancelled RouteStatus = "cancelled"
	RouteStatusDeleted   RouteStatus = "deleted"
)

// HistoryNodeStatus trang thai di qua node.
type HistoryNodeStatus string

const (
	HistoryNodeArrived HistoryNodeStatus = "arrived"
	HistoryNodePassed  HistoryNodeStatus = "passed"
)

// TravelMode phuong thuc di chuyen.
// Bang: travel_modes [T14]
type TravelMode struct {
	ModeID      string  `gorm:"primaryKey;size:20;column:mode_id" json:"mode_id"`
	ModeName    string  `gorm:"not null;size:50;column:mode_name" json:"mode_name"`
	SpeedFactor float64 `gorm:"not null;default:1.0;column:speed_factor" json:"speed_factor"`
}

func (TravelMode) TableName() string {
	return "travel_modes"
}

// Route lo trinh tim duong cua benh nhan.
// Bang: routes [T15]
type Route struct {
	RouteID       string      `gorm:"primaryKey;size:50;column:route_id" json:"route_id"`
	UserID        uint64      `gorm:"not null;index;column:user_id" json:"user_id"`
	ModeID        string      `gorm:"not null;size:20;column:mode_id" json:"mode_id"`
	StartLocation int         `gorm:"not null;column:start_location" json:"start_location"`
	DestLocation  int         `gorm:"not null;column:dest_location" json:"dest_location"`
	RouteMode     RouteMode   `gorm:"not null;column:route_mode" json:"route_mode"`
	AgentID       *int        `gorm:"column:agent_id" json:"agent_id,omitempty"`
	TotalDistance  float64    `gorm:"not null;default:0;column:total_distance" json:"total_distance"`
	EstimatedTime float64     `gorm:"not null;default:0;column:estimated_time" json:"estimated_time"`
	Status        RouteStatus `gorm:"not null;default:active;index;column:status" json:"status"`
	CreatedAt     time.Time   `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`

	// Belongs-to (omit from JSON to avoid circular)
	User       *User       `gorm:"foreignKey:UserID;references:UserID" json:"-"`
	TravelMode *TravelMode `gorm:"foreignKey:ModeID;references:ModeID" json:"-"`
}

func (Route) TableName() string {
	return "routes"
}

// RoutePath luu tung buoc (step) cua lo trinh tren grid.
// Bang: route_paths [T16]
type RoutePath struct {
	PathID       uint64 `gorm:"primaryKey;autoIncrement;column:path_id" json:"path_id"`
	RouteID      string `gorm:"not null;index;size:50;column:route_id" json:"route_id"`
	StepOrder    int    `gorm:"not null;column:step_order" json:"step_order"`
	GridRow      int    `gorm:"not null;column:grid_row" json:"grid_row"`
	GridCol      int    `gorm:"not null;column:grid_col" json:"grid_col"`
	GridLocation int    `gorm:"not null;column:grid_location" json:"grid_location"`
	Instruction  string `gorm:"type:text;column:instruction" json:"instruction"`
	VoiceText    string `gorm:"type:text;column:voice_text" json:"voice_text,omitempty"`

	// Belongs-to
	Route *Route `gorm:"foreignKey:RouteID;references:RouteID" json:"-"`
}

func (RoutePath) TableName() string {
	return "route_paths"
}

// RouteHistoryNode ghi nhan cac node benh nhan da di qua.
// Bang: route_history_nodes [T17]
type RouteHistoryNode struct {
	HistoryID    uint64            `gorm:"primaryKey;autoIncrement;column:history_id" json:"history_id"`
	RouteID      string            `gorm:"not null;index;size:50;column:route_id" json:"route_id"`
	GridLocation int               `gorm:"not null;column:grid_location" json:"grid_location"`
	Status       HistoryNodeStatus `gorm:"not null;column:status" json:"status"`
	ArrivalTime  time.Time         `gorm:"not null;autoCreateTime;column:arrival_time" json:"arrival_time"`

	// Belongs-to
	Route *Route `gorm:"foreignKey:RouteID;references:RouteID" json:"-"`
}

func (RouteHistoryNode) TableName() string {
	return "route_history_nodes"
}

// RouteShare chia se lo trinh qua link.
// Bang: route_shares [T18]
type RouteShare struct {
	ShareID       uint64     `gorm:"primaryKey;autoIncrement;column:share_id" json:"share_id"`
	RouteID       string     `gorm:"not null;index;size:50;column:route_id" json:"route_id"`
	ShareURL      string     `gorm:"uniqueIndex;not null;size:100;column:share_url" json:"share_url"`
	ReceiverPhone string     `gorm:"size:15;column:receiver_phone" json:"receiver_phone,omitempty"`
	ExpiredAt     *time.Time `gorm:"column:expired_at" json:"expired_at,omitempty"`
	CreatedAt     time.Time  `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`

	// Belongs-to
	Route *Route `gorm:"foreignKey:RouteID;references:RouteID" json:"-"`
}

func (RouteShare) TableName() string {
	return "route_shares"
}

// RouteFeedback danh gia lo trinh sau khi hoan thanh.
// Bang: route_feedbacks [T19]
type RouteFeedback struct {
	FeedbackID uint64    `gorm:"primaryKey;autoIncrement;column:feedback_id" json:"feedback_id"`
	RouteID    string    `gorm:"uniqueIndex;not null;size:50;column:route_id" json:"route_id"`
	Rating     int       `gorm:"not null;column:rating" json:"rating"`
	Comment    string    `gorm:"type:text;column:comment" json:"comment,omitempty"`
	IsAccurate *bool     `gorm:"column:is_accurate" json:"is_accurate,omitempty"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`

	// Belongs-to
	Route *Route `gorm:"foreignKey:RouteID;references:RouteID" json:"-"`
}

func (RouteFeedback) TableName() string {
	return "route_feedbacks"
}
