package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hospital/middleware"
	response "hospital/pkg"
	"hospital/service"
)

// RouteHandler xu ly HTTP request cho Route module.
type RouteHandler struct {
	svc *service.RouteService
}

func NewRouteHandler(svc *service.RouteService) *RouteHandler {
	return &RouteHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS
// ========================================

type previewRequest struct {
	StartLocation int    `json:"start_location"`
	DestLocation  int    `json:"dest_location"`
	ModeID        string `json:"mode_id"`
	StartNode     string `json:"start_node"`
	TargetNode    string `json:"target_node"`
	TransportMode string `json:"transport_mode"`
}

type orderRequest struct {
	StartLocation int    `json:"start_location"`
	DestLocation  int    `json:"dest_location"`
	ModeID        string `json:"mode_id"`
	StartNode     string `json:"start_node"`
	TargetNode    string `json:"target_node"`
	TransportMode string `json:"transport_mode"`
}

type orderMultiRequest struct {
	StartLocation   int      `json:"start_location"`
	TargetLocations []int    `json:"target_locations"`
	ModeID          string   `json:"mode_id"`
	StartNode       string   `json:"start_node"`
	TargetNodes     []string `json:"target_nodes"`
	TransportMode   string   `json:"transport_mode"`
}

type cancelRequest struct {
	RouteID string `json:"route_id"`
}

type recalcRequest struct {
	RouteID         string `json:"route_id"`
	CurrentLocation int    `json:"current_location"`
	CurrentNode     string `json:"current_node"`
}

type etaRequest struct {
	RouteID     string `json:"route_id"`
	CurrentStep int    `json:"current_step"`
	CurrentNode string `json:"current_node"`
}

type passNodeRequest struct {
	RouteID      string `json:"route_id"`
	GridLocation int    `json:"grid_location"`
	NodeID       string `json:"node_id"`
}

type shareRequest struct {
	RouteID        string `json:"route_id"`
	ReceiverPhone  string `json:"receiver_phone"`
	RecipientPhone string `json:"recipient_phone"`
}

type rateRequest struct {
	RouteID    string `json:"route_id" binding:"required"`
	Rating     int    `json:"rating" binding:"required"`
	Comment    string `json:"comment"`
	IsAccurate *bool  `json:"is_accurate"`
}

// ========================================
// SLICE 2 APIS
// ========================================

// [45] GET /api/route/get_modes
func (h *RouteHandler) GetModes(c *gin.Context) {
	modes, err := h.svc.GetAllModes()
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, modes)
}

// ========================================
// SLICE 3 APIS  - Route Core
// ========================================

// [37] POST /api/route/preview
func (h *RouteHandler) Preview(c *gin.Context) {
	var req previewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	normalizeRoutePoints(&req.StartLocation, &req.DestLocation, req.StartNode, req.TargetNode)
	normalizeMode(&req.ModeID, req.TransportMode)
	if req.StartLocation == 0 || req.DestLocation == 0 || req.ModeID == "" {
		response.ErrMissingParam(c)
		return
	}

	result, err := h.svc.PreviewRoute(req.StartLocation, req.DestLocation, req.ModeID)
	if err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, result)
}

// [31] POST /api/route/order
func (h *RouteHandler) Order(c *gin.Context) {
	var req orderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	normalizeRoutePoints(&req.StartLocation, &req.DestLocation, req.StartNode, req.TargetNode)
	normalizeMode(&req.ModeID, req.TransportMode)
	if req.StartLocation == 0 || req.DestLocation == 0 || req.ModeID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	route, paths, err := h.svc.OrderRoute(userID, req.StartLocation, req.DestLocation, req.ModeID)
	if err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, gin.H{
		"route": route,
		"paths": paths,
	})
}

// POST /api/route/order_multi
func (h *RouteHandler) OrderMulti(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req orderMultiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	normalizeMultiRoute(&req)
	if req.StartLocation == 0 || len(req.TargetLocations) == 0 || req.ModeID == "" {
		response.ErrMissingParam(c)
		return
	}

	route, paths, err := h.svc.OrderMultiRoute(userID, req.StartLocation, req.TargetLocations, req.ModeID)
	if err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, gin.H{
		"route": route,
		"paths": paths,
	})
}

// POST /api/route/order_unordered
func (h *RouteHandler) OrderUnordered(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	var req orderMultiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	normalizeMultiRoute(&req)
	if req.StartLocation == 0 || len(req.TargetLocations) == 0 || req.ModeID == "" {
		response.ErrMissingParam(c)
		return
	}

	route, paths, err := h.svc.OrderUnorderedRoute(userID, req.StartLocation, req.TargetLocations, req.ModeID)
	if err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, gin.H{
		"route": route,
		"paths": paths,
	})
}

// [36] GET /api/route/get_steps?route_id=
func (h *RouteHandler) GetSteps(c *gin.Context) {
	routeID := c.Query("route_id")
	if routeID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.VerifyRouteOwner(routeID, userID); err != nil {
		response.Error(c, response.CodeNotAccess, err.Error())
		return
	}

	paths, err := h.svc.GetSteps(routeID)
	if err != nil {
		response.ErrNotFound(c)
		return
	}

	response.Success(c, paths)
}

// [38] POST /api/route/get_eta
func (h *RouteHandler) GetETA(c *gin.Context) {
	var req etaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.RouteID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.VerifyRouteOwner(req.RouteID, userID); err != nil {
		response.Error(c, response.CodeNotAccess, err.Error())
		return
	}

	result, err := h.svc.GetETA(req.RouteID, req.CurrentStep)
	if err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, result)
}

// ========================================
// SLICE 4 APIS  - Route Mo Rong
// ========================================

// [34] GET /api/route/get_active
func (h *RouteHandler) GetActive(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	route, paths, err := h.svc.GetActiveRoute(userID)
	if err != nil {
		response.Error(c, response.CodePathNotFound, "No active route")
		return
	}

	response.Success(c, gin.H{
		"route": route,
		"paths": paths,
	})
}

// [35] POST /api/route/cancel
func (h *RouteHandler) Cancel(c *gin.Context) {
	var req cancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.RouteID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.CancelRoute(req.RouteID, userID); err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, gin.H{"cancelled": true})
}

// [33] POST /api/route/recalculate
func (h *RouteHandler) Recalculate(c *gin.Context) {
	var req recalcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.CurrentLocation == 0 {
		req.CurrentLocation = parseLegacyLocation(req.CurrentNode)
	}
	if req.RouteID == "" || req.CurrentLocation == 0 {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	route, paths, err := h.svc.RecalculateRoute(req.RouteID, userID, req.CurrentLocation)
	if err != nil {
		response.Error(c, response.CodePathNotFound, err.Error())
		return
	}

	response.Success(c, gin.H{
		"route": route,
		"paths": paths,
	})
}

// [43] POST /api/route/pass_node
func (h *RouteHandler) PassNode(c *gin.Context) {
	var req passNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.GridLocation == 0 {
		req.GridLocation = parseLegacyLocation(req.NodeID)
	}
	if req.RouteID == "" || req.GridLocation == 0 {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.VerifyRouteOwner(req.RouteID, userID); err != nil {
		response.Error(c, response.CodeNotAccess, err.Error())
		return
	}

	if err := h.svc.PassNode(req.RouteID, req.GridLocation); err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{"recorded": true})
}

// [44] GET /api/route/get_next?route_id=&current_step=&limit=
func (h *RouteHandler) GetNext(c *gin.Context) {
	routeID := c.Query("route_id")
	if routeID == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.VerifyRouteOwner(routeID, userID); err != nil {
		response.Error(c, response.CodeNotAccess, err.Error())
		return
	}

	currentStep, _ := strconv.Atoi(c.DefaultQuery("current_step", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	steps, err := h.svc.GetNextSteps(routeID, currentStep, limit)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, steps)
}

// [39] GET /api/route/get_history?page=&limit=
func (h *RouteHandler) GetHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	routes, total, err := h.svc.GetHistory(userID, page, limit)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{
		"routes": routes,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// [40] DELETE /api/route/clear_history
func (h *RouteHandler) ClearHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.ErrNotAuthenticated(c)
		return
	}

	if err := h.svc.ClearHistory(userID); err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{"cleared": true})
}

// [41] POST /api/route/share
func (h *RouteHandler) Share(c *gin.Context) {
	var req shareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.ReceiverPhone == "" {
		req.ReceiverPhone = req.RecipientPhone
	}
	if req.RouteID == "" || req.ReceiverPhone == "" {
		response.ErrMissingParam(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.VerifyRouteOwner(req.RouteID, userID); err != nil {
		response.Error(c, response.CodeNotAccess, err.Error())
		return
	}

	share, err := h.svc.ShareRoute(req.RouteID, req.ReceiverPhone)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, share)
}

// [42] POST /api/route/rate
func (h *RouteHandler) Rate(c *gin.Context) {
	var req rateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		response.ErrInvalidValue(c)
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.svc.VerifyRouteOwner(req.RouteID, userID); err != nil {
		response.Error(c, response.CodeNotAccess, err.Error())
		return
	}

	if err := h.svc.RatePath(req.RouteID, req.Rating, req.Comment, req.IsAccurate); err != nil {
		response.ErrUnexpected(c)
		return
	}

	response.Success(c, gin.H{"rated": true})
}

func normalizeRoutePoints(start *int, dest *int, startAlias string, destAlias string) {
	if *start == 0 {
		*start = parseLegacyLocation(startAlias)
	}
	if *dest == 0 {
		*dest = parseLegacyLocation(destAlias)
	}
}

func normalizeMode(mode *string, alias string) {
	if *mode == "" {
		*mode = alias
	}
}

func normalizeMultiRoute(req *orderMultiRequest) {
	if req.StartLocation == 0 {
		req.StartLocation = parseLegacyLocation(req.StartNode)
	}
	if len(req.TargetLocations) == 0 && len(req.TargetNodes) > 0 {
		for _, node := range req.TargetNodes {
			if loc := parseLegacyLocation(node); loc > 0 {
				req.TargetLocations = append(req.TargetLocations, loc)
			}
		}
	}
	normalizeMode(&req.ModeID, req.TransportMode)
}

func parseLegacyLocation(value string) int {
	if value == "" {
		return 0
	}
	if n, err := strconv.Atoi(value); err == nil {
		return n
	}
	digits := ""
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}
	if digits == "" {
		return 0
	}
	n, _ := strconv.Atoi(digits)
	return n
}
