package schema

import "time"

// ========================================
// FLOW  - Luong nguoi, mat do, canh bao
// Slice 5
// ========================================

// ObstacleReportStatus trang thai bao cao vat can.
type ObstacleReportStatus string

const (
	ObstacleStatusPending  ObstacleReportStatus = "pending"
	ObstacleStatusResolved ObstacleReportStatus = "resolved"
	ObstacleStatusRejected ObstacleReportStatus = "rejected"
)

// PriorityRouteStatus trang thai tuyen uu tien.
type PriorityRouteStatus string

const (
	PriorityStatusActive  PriorityRouteStatus = "active"
	PriorityStatusExpired PriorityRouteStatus = "expired"
)

// UserPing vi tri hien tai cua nguoi dung tren grid.
// App gui dinh ky (moi 10-30 giay) de tinh mat do.
// Bang: user_pings [T10]
type UserPing struct {
	PingID       uint64    `gorm:"primaryKey;autoIncrement;column:ping_id" json:"ping_id"`
	UserID       uint64    `gorm:"not null;index;column:user_id" json:"user_id"`
	RouteID      *string   `gorm:"size:50;column:route_id" json:"route_id"`
	GridLocation int       `gorm:"not null;index;column:grid_location" json:"grid_location"`
	GridRow      int       `gorm:"not null;column:grid_row" json:"grid_row"`
	GridCol      int       `gorm:"not null;column:grid_col" json:"grid_col"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime;index;column:created_at" json:"created_at,omitempty"`

	// Belongs-to
	User *User `gorm:"foreignKey:UserID;references:UserID"`
}

func (UserPing) TableName() string {
	return "user_pings"
}

// ObstacleReport bao cao vat can tren duong di.
// Benh nhan bao cao, coordinator xu ly.
// Bang: obstacle_reports [T11]
type ObstacleReport struct {
	ReportID     uint64               `gorm:"primaryKey;autoIncrement;column:report_id" json:"report_id"`
	UserID       uint64               `gorm:"not null;index;column:user_id" json:"user_id"`
	RouteID      *string              `gorm:"size:50;column:route_id" json:"route_id"`
	GridLocation int                  `gorm:"not null;column:grid_location" json:"grid_location"`
	ReportType   string               `gorm:"not null;size:50;column:report_type" json:"report_type"`
	Description  string               `gorm:"type:text;column:description" json:"description"`
	ResolvedBy   *uint64              `gorm:"column:resolved_by"`
	Status       ObstacleReportStatus `gorm:"not null;default:pending;index;column:status" json:"status"`
	CreatedAt    time.Time            `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	ResolvedAt   *time.Time           `gorm:"column:resolved_at"`

	// Belongs-to
	User     *User  `gorm:"foreignKey:UserID;references:UserID"`
	Resolver *Staff `gorm:"foreignKey:ResolvedBy;references:StaffID"`
}

func (ObstacleReport) TableName() string {
	return "obstacle_reports"
}

// HeatmapSnapshot anh chup mat do theo grid_location.
// Luu tru cac moc thoi gian de phan tich xu huong.
// Bang: heatmap_snapshots [T12]
type HeatmapSnapshot struct {
	HeatmapID    uint64    `gorm:"primaryKey;autoIncrement;column:heatmap_id" json:"heatmap_id"`
	GridLocation int       `gorm:"not null;index;column:grid_location" json:"grid_location"`
	DensityLevel int       `gorm:"not null;column:density_level" json:"density_level"`
	RecordedAt   time.Time `gorm:"not null;autoCreateTime;index;column:recorded_at" json:"recorded_at"`
}

func (HeatmapSnapshot) TableName() string {
	return "heatmap_snapshots"
}

// PriorityRoute tuyen uu tien khan cap.
// Coordinator dat tuyen uu tien de giai toa tat nghen.
// Bang: priority_routes [T13]
type PriorityRoute struct {
	PriorityID   uint64              `gorm:"primaryKey;autoIncrement;column:priority_id" json:"priority_id"`
	EmergencyID  *string             `gorm:"size:50;column:emergency_id" json:"emergency_id"`
	SetBy        uint64              `gorm:"not null;column:set_by" json:"set_by"`
	FromLocation int                 `gorm:"not null;column:from_location" json:"from_location"`
	ToLocation   int                 `gorm:"not null;column:to_location" json:"to_location"`
	Reason       string              `gorm:"type:text;column:reason" json:"reason"`
	Status       PriorityRouteStatus `gorm:"not null;default:active;index;column:status" json:"status"`
	ActivatedAt  time.Time           `gorm:"not null;autoCreateTime;column:activated_at" json:"activated_at"`
	ExpiredAt    *time.Time          `gorm:"column:expired_at"`

	// Belongs-to
	Staff *Staff `gorm:"foreignKey:SetBy;references:StaffID"`
}

func (PriorityRoute) TableName() string {
	return "priority_routes"
}
