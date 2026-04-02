package repository

import (
	"hospital/schema"

	"gorm.io/gorm"
)

// RouteRepo xu ly truy van database cho Route module.
// Bao gom: travel_modes, routes, route_paths,
// route_history_nodes, route_shares, route_feedbacks.
type RouteRepo struct {
	db *gorm.DB
}

func NewRouteRepo(db *gorm.DB) *RouteRepo {
	return &RouteRepo{db: db}
}

// ========================================
// TRAVEL MODES
// ========================================

// FindAllModes tra ve tat ca phuong thuc di chuyen.
func (r *RouteRepo) FindAllModes() ([]schema.TravelMode, error) {
	var modes []schema.TravelMode
	err := r.db.Find(&modes).Error
	return modes, err
}

// FindModeByID tim travel_mode theo mode_id.
func (r *RouteRepo) FindModeByID(modeID string) (*schema.TravelMode, error) {
	var mode schema.TravelMode
	err := r.db.Where("mode_id = ?", modeID).First(&mode).Error
	if err != nil {
		return nil, err
	}
	return &mode, nil
}

// ========================================
// ROUTES
// ========================================

// CreateRoute tao lo trinh moi.
func (r *RouteRepo) CreateRoute(route *schema.Route) error {
	return r.db.Create(route).Error
}

// FindRouteByID tim route theo route_id.
func (r *RouteRepo) FindRouteByID(routeID string) (*schema.Route, error) {
	var route schema.Route
	err := r.db.Where("route_id = ?", routeID).First(&route).Error
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// FindActiveRoute tim route dang active cua user.
func (r *RouteRepo) FindActiveRoute(userID uint64) (*schema.Route, error) {
	var route schema.Route
	err := r.db.Where("user_id = ? AND status = ?", userID, schema.RouteStatusActive).
		First(&route).Error
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// FindRouteHistory tim lich su route cua user (pagination).
func (r *RouteRepo) FindRouteHistory(userID uint64, page, limit int) ([]schema.Route, int64, error) {
	var routes []schema.Route
	var total int64

	q := r.db.Model(&schema.Route{}).Where("user_id = ? AND status != ?", userID, schema.RouteStatusDeleted)
	q.Count(&total)

	err := q.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&routes).Error
	return routes, total, err
}

// UpdateRouteStatus cap nhat trang thai route.
func (r *RouteRepo) UpdateRouteStatus(routeID string, status schema.RouteStatus) error {
	return r.db.Model(&schema.Route{}).
		Where("route_id = ?", routeID).
		Update("status", status).Error
}

// UpdateRoute cap nhat route (distance, time, etc.).
func (r *RouteRepo) UpdateRoute(routeID string, updates map[string]interface{}) error {
	return r.db.Model(&schema.Route{}).
		Where("route_id = ?", routeID).
		Updates(updates).Error
}

// SoftDeleteRouteHistory xoa mem tat ca route cua user.
func (r *RouteRepo) SoftDeleteRouteHistory(userID uint64) error {
	return r.db.Model(&schema.Route{}).
		Where("user_id = ? AND status IN ?", userID,
			[]schema.RouteStatus{schema.RouteStatusCompleted, schema.RouteStatusCancelled}).
		Update("status", schema.RouteStatusDeleted).Error
}

// ========================================
// ROUTE PATHS
// ========================================

// CreatePaths tao nhieu buoc di cung luc (bulk insert).
func (r *RouteRepo) CreatePaths(paths []schema.RoutePath) error {
	if len(paths) == 0 {
		return nil
	}
	return r.db.Create(&paths).Error
}

// FindPathsByRouteID lay cac buoc di cua route.
func (r *RouteRepo) FindPathsByRouteID(routeID string) ([]schema.RoutePath, error) {
	var paths []schema.RoutePath
	err := r.db.Where("route_id = ?", routeID).
		Order("step_order ASC").
		Find(&paths).Error
	return paths, err
}

// FindRemainingSteps lay cac buoc con lai tu step_order > current.
func (r *RouteRepo) FindRemainingSteps(routeID string, currentStep int) ([]schema.RoutePath, error) {
	var paths []schema.RoutePath
	err := r.db.Where("route_id = ? AND step_order > ?", routeID, currentStep).
		Order("step_order ASC").
		Find(&paths).Error
	return paths, err
}

// DeletePathsByRouteID xoa tat ca buoc cua route (cho recalculate).
func (r *RouteRepo) DeletePathsByRouteID(routeID string) error {
	return r.db.Where("route_id = ?", routeID).
		Delete(&schema.RoutePath{}).Error
}

// ========================================
// ROUTE HISTORY NODES
// ========================================

// CreateHistoryNode ghi nhan node da di qua.
func (r *RouteRepo) CreateHistoryNode(node *schema.RouteHistoryNode) error {
	return r.db.Create(node).Error
}

// UpdatePreviousNodes cap nhat status cac node truoc thanh "passed".
func (r *RouteRepo) UpdatePreviousNodes(routeID string, currentHistoryID uint64) error {
	return r.db.Model(&schema.RouteHistoryNode{}).
		Where("route_id = ? AND history_id < ? AND status = ?",
			routeID, currentHistoryID, schema.HistoryNodeArrived).
		Update("status", schema.HistoryNodePassed).Error
}

// ========================================
// ROUTE SHARES
// ========================================

// CreateShare tao share link cho route.
func (r *RouteRepo) CreateShare(share *schema.RouteShare) error {
	return r.db.Create(share).Error
}

// FindShareByURL tim share theo URL.
func (r *RouteRepo) FindShareByURL(shareURL string) (*schema.RouteShare, error) {
	var share schema.RouteShare
	err := r.db.Where("share_url = ?", shareURL).
		Preload("Route").
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// ========================================
// ROUTE FEEDBACKS
// ========================================

// CreateFeedback tao danh gia cho route.
func (r *RouteRepo) CreateFeedback(fb *schema.RouteFeedback) error {
	return r.db.Create(fb).Error
}

// ========================================
// TRANSACTIONS
// ========================================

// CreateRouteWithPaths tao route + paths trong 1 transaction.
func (r *RouteRepo) CreateRouteWithPaths(route *schema.Route, paths []schema.RoutePath) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(route).Error; err != nil {
			return err
		}
		if len(paths) > 0 {
			if err := tx.Create(&paths).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RecalculateRouteTransaction xoa paths cu, tao moi, cap nhat route trong 1 transaction.
func (r *RouteRepo) RecalculateRouteTransaction(routeID string, newPaths []schema.RoutePath, updates map[string]interface{}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Xoa paths cu
		if err := tx.Where("route_id = ?", routeID).Delete(&schema.RoutePath{}).Error; err != nil {
			return err
		}
		// Tao paths moi
		if len(newPaths) > 0 {
			if err := tx.Create(&newPaths).Error; err != nil {
				return err
			}
		}
		// Cap nhat route
		if err := tx.Model(&schema.Route{}).Where("route_id = ?", routeID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}
