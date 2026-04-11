package repository

import (
	"time"

	"hospital/schema"

	"gorm.io/gorm"
)

// FlowRepo xu ly truy van database cho Flow module (Slice 5).
// Bao gom: user_pings, obstacle_reports, heatmap_snapshots,
// priority_routes, simulation_runs, patient_agents.
type FlowRepo struct {
	db *gorm.DB
}

func NewFlowRepo(db *gorm.DB) *FlowRepo {
	return &FlowRepo{db: db}
}

// ========================================
// USER PINGS
// ========================================

// CreatePing ghi nhan vi tri hien tai cua user len grid.
func (r *FlowRepo) CreatePing(ping *schema.UserPing) error {
	return r.db.Create(ping).Error
}

// DensityResult ket qua dem mat do theo vi tri.
type DensityResult struct {
	GridLocation int   `json:"grid_location"`
	Count        int64 `json:"count"`
}

// GetDensityByLocation dem so luong ping tai 1 vi tri trong N phut gan nhat.
func (r *FlowRepo) GetDensityByLocation(gridLocation int, minutes int) (int64, error) {
	var count int64
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)

	err := r.db.Model(&schema.UserPing{}).
		Where("grid_location = ? AND created_at > ?", gridLocation, since).
		Count(&count).Error
	return count, err
}

// GetDensityAll dem mat do tat ca vi tri trong N phut gan nhat.
// Tra ve danh sach [{grid_location, count}] sap xep theo count giam dan.
func (r *FlowRepo) GetDensityAll(minutes int) ([]DensityResult, error) {
	var results []DensityResult
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)

	err := r.db.Model(&schema.UserPing{}).
		Select("grid_location, COUNT(*) as count").
		Where("created_at > ?", since).
		Group("grid_location").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// GetBottlenecks lay top N diem un tac nhat trong N phut.
func (r *FlowRepo) GetBottlenecks(minutes int, limit int) ([]DensityResult, error) {
	var results []DensityResult
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)

	err := r.db.Model(&schema.UserPing{}).
		Select("grid_location, COUNT(*) as count").
		Where("created_at > ?", since).
		Group("grid_location").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// GetPingsByHour thong ke so ping theo gio (forecast/stats).
type HourlyStats struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

func (r *FlowRepo) GetPingsByHour(hours int) ([]HourlyStats, error) {
	var results []HourlyStats
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	err := r.db.Model(&schema.UserPing{}).
		Select("strftime('%H', created_at) as hour, COUNT(*) as count").
		Where("created_at > ?", since).
		Group("hour").
		Order("hour ASC").
		Find(&results).Error
	return results, err
}

// GetEdgeDensity dem so ping tai 1 vi tri cu the.
func (r *FlowRepo) GetEdgeDensity(gridLocation int) (int64, error) {
	var count int64
	since := time.Now().Add(-30 * time.Minute)

	err := r.db.Model(&schema.UserPing{}).
		Where("grid_location = ? AND created_at > ?", gridLocation, since).
		Count(&count).Error
	return count, err
}

// DeleteOldPings xoa tat ca pings cu hon threshold.
func (r *FlowRepo) DeleteOldPings(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&schema.UserPing{})
	return result.RowsAffected, result.Error
}

// DeleteAllPings xoa tat ca pings.
func (r *FlowRepo) DeleteAllPings() (int64, error) {
	result := r.db.Where("1 = 1").Delete(&schema.UserPing{})
	return result.RowsAffected, result.Error
}

// ========================================
// OBSTACLE REPORTS
// ========================================

// CreateObstacleReport bao cao vat can moi.
func (r *FlowRepo) CreateObstacleReport(report *schema.ObstacleReport) error {
	return r.db.Create(report).Error
}

// GetObstacleReports lay danh sach bao cao (loc theo status neu co).
func (r *FlowRepo) GetObstacleReports(status string, page, limit int) ([]schema.ObstacleReport, int64, error) {
	var reports []schema.ObstacleReport
	var total int64

	q := r.db.Model(&schema.ObstacleReport{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)

	err := q.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&reports).Error
	return reports, total, err
}

// ResolveObstacle staff xu ly bao cao.
func (r *FlowRepo) ResolveObstacle(reportID uint64, staffID uint64, newStatus schema.ObstacleReportStatus) error {
	now := time.Now()
	return r.db.Model(&schema.ObstacleReport{}).
		Where("report_id = ?", reportID).
		Updates(map[string]interface{}{
			"status":      newStatus,
			"resolved_by": staffID,
			"resolved_at": &now,
		}).Error
}

// ========================================
// HEATMAP SNAPSHOTS
// ========================================

// GetHeatmapSnapshots lay snapshots gan nhat.
func (r *FlowRepo) GetHeatmapSnapshots(minutes int) ([]schema.HeatmapSnapshot, error) {
	var snapshots []schema.HeatmapSnapshot
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)

	err := r.db.Where("recorded_at > ?", since).
		Order("recorded_at DESC").
		Find(&snapshots).Error
	return snapshots, err
}

// CreateHeatmapSnapshot luu 1 snapshot.
func (r *FlowRepo) CreateHeatmapSnapshot(snapshot *schema.HeatmapSnapshot) error {
	return r.db.Create(snapshot).Error
}

// DeleteAllHeatmapSnapshots xoa tat ca snapshots.
func (r *FlowRepo) DeleteAllHeatmapSnapshots() (int64, error) {
	result := r.db.Where("1 = 1").Delete(&schema.HeatmapSnapshot{})
	return result.RowsAffected, result.Error
}

// ========================================
// PRIORITY ROUTES
// ========================================

// CreatePriorityRoute tao tuyen uu tien moi.
func (r *FlowRepo) CreatePriorityRoute(pr *schema.PriorityRoute) error {
	return r.db.Create(pr).Error
}

// GetActivePriorityRoutes lay cac tuyen uu tien dang active.
func (r *FlowRepo) GetActivePriorityRoutes() ([]schema.PriorityRoute, error) {
	var routes []schema.PriorityRoute
	err := r.db.Where("status = ?", schema.PriorityStatusActive).
		Order("activated_at DESC").
		Find(&routes).Error
	return routes, err
}

// DeletePriorityRoute xoa tuyen uu tien (hoac set expired).
func (r *FlowRepo) ExpirePriorityRoute(priorityID uint64) error {
	now := time.Now()
	return r.db.Model(&schema.PriorityRoute{}).
		Where("priority_id = ?", priorityID).
		Updates(map[string]interface{}{
			"status":     schema.PriorityStatusExpired,
			"expired_at": &now,
		}).Error
}

// ========================================
// SIMULATION RUNS
// ========================================

// CreateSimulationRun tao phien mo phong moi.
func (r *FlowRepo) CreateSimulationRun(run *schema.SimulationRun) error {
	return r.db.Create(run).Error
}

// GetRunningSimulation lay phien dang chay.
func (r *FlowRepo) GetRunningSimulation() (*schema.SimulationRun, error) {
	var run schema.SimulationRun
	err := r.db.Where("status = ?", schema.SimulationRunning).First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// UpdateSimulationStatus cap nhat trang thai phien mo phong.
func (r *FlowRepo) UpdateSimulationStatus(runID uint64, status schema.SimulationStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == schema.SimulationStopped || status == schema.SimulationDone {
		now := time.Now()
		updates["ended_at"] = &now
	}
	return r.db.Model(&schema.SimulationRun{}).
		Where("run_id = ?", runID).
		Updates(updates).Error
}

// GetSimulationByID lay phien theo ID.
func (r *FlowRepo) GetSimulationByID(runID uint64) (*schema.SimulationRun, error) {
	var run schema.SimulationRun
	err := r.db.Where("run_id = ?", runID).First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// ========================================
// GRID POIS (doc/update tu Map module)
// ========================================

// UpdatePOICapacity cap nhat capacity cua 1 POI.
func (r *FlowRepo) UpdatePOICapacity(poiID uint32, capacity int) error {
	return r.db.Model(&schema.GridPOI{}).
		Where("poi_id = ?", poiID).
		Update("capacity", capacity).Error
}

// ========================================
// TRANSACTIONS
// ========================================

// ResetFlowData xoa tat ca flow data trong 1 transaction.
func (r *FlowRepo) ResetFlowData() error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&schema.UserPing{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&schema.HeatmapSnapshot{}).Error; err != nil {
			return err
		}
		return nil
	})
}
