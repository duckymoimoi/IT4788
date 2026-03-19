package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	response "hospital/pkg"
	"hospital/service"
)

type MapHandler struct {
	svc *service.MapService
}

func NewMapHandler(svc *service.MapService) *MapHandler {
	return &MapHandler{svc: svc}
}

// ========================================
// REQUEST STRUCTS — Admin APIs
// ========================================

type addNodeRequest struct {
	FloorID              uint32   `json:"floor_id"`
	WardID               *uint32  `json:"ward_id"`
	NodeCode             string   `json:"node_code"`
	NodeName             string   `json:"node_name"`
	NodeType             string   `json:"node_type"`
	PolygonCoords        string   `json:"polygon_coords"`
	CenterX              float32  `json:"center_x"`
	CenterY              float32  `json:"center_y"`
	AccessX              *float32 `json:"access_x"`
	AccessY              *float32 `json:"access_y"`
	IsLandmark           bool     `json:"is_landmark"`
	WheelchairAccessible bool     `json:"wheelchair_accessible"`
}

type editNodeRequest struct {
	NodeID               uint32   `json:"node_id"`
	NodeCode             *string  `json:"node_code"`
	NodeName             *string  `json:"node_name"`
	NodeType             *string  `json:"node_type"`
	PolygonCoords        *string  `json:"polygon_coords"`
	CenterX              *float32 `json:"center_x"`
	CenterY              *float32 `json:"center_y"`
	AccessX              *float32 `json:"access_x"`
	AccessY              *float32 `json:"access_y"`
	IsLandmark           *bool    `json:"is_landmark"`
	WheelchairAccessible *bool    `json:"wheelchair_accessible"`
	IsAccessible         *bool    `json:"is_accessible"`
}

type addEdgeRequest struct {
	FloorID              uint32   `json:"floor_id"`
	FromNodeID           uint32   `json:"from_node_id"`
	ToNodeID             uint32   `json:"to_node_id"`
	PolygonCoords        *string  `json:"polygon_coords"`
	DistanceM            *float32 `json:"distance_m"`
	Weight               float32  `json:"weight"`
	IsBidirectional      bool     `json:"is_bidirectional"`
	IsCrossFloor         bool     `json:"is_cross_floor"`
	WheelchairAccessible bool     `json:"wheelchair_accessible"`
}

type setWeightRequest struct {
	EdgeID uint32  `json:"edge_id"`
	Weight float32 `json:"weight"`
}

// ========================================
// HELPER
// ========================================

func parseUint32Query(c *gin.Context, key string) uint32 {
	val := c.Query(key)
	if val == "" {
		return 0
	}
	n, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(n)
}

func (h *MapHandler) handleMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFloorNotFound):
		response.Error(c, response.CodeFloorNotFound, "Floor not found")
	case errors.Is(err, service.ErrNodeNotFound):
		response.Error(c, response.CodeNodeNotFound, "Node not found")
	case errors.Is(err, service.ErrEdgeNotFound):
		response.Error(c, response.CodeEdgeNotFound, "Edge not found")
	case errors.Is(err, service.ErrNodeCodeExist):
		response.Error(c, response.CodeInvalidValue, "Node code already exists")
	case errors.Is(err, service.ErrMissingField):
		response.ErrMissingParam(c)
	default:
		response.ErrUnexpected(c)
	}
}

// ========================================
// PUBLIC READ APIs (khong can token)
// ========================================

// [16] GetFloors — GET /api/map/get_floors
func (h *MapHandler) GetFloors(c *gin.Context) {
	result, err := h.svc.GetFloors()
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [17] GetNodes — GET /api/map/get_nodes?floor_id=
func (h *MapHandler) GetNodes(c *gin.Context) {
	floorID := parseUint32Query(c, "floor_id")
	result, err := h.svc.GetNodes(floorID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [18] GetEdges — GET /api/map/get_edges?floor_id=
func (h *MapHandler) GetEdges(c *gin.Context) {
	floorID := parseUint32Query(c, "floor_id")
	result, err := h.svc.GetEdges(floorID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [19] GetMeta — GET /api/map/get_meta?floor_id=
func (h *MapHandler) GetMeta(c *gin.Context) {
	floorID := parseUint32Query(c, "floor_id")
	if floorID == 0 {
		response.ErrMissingParam(c)
		return
	}
	result, err := h.svc.GetMeta(floorID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [20] GetDepartments — GET /api/map/get_depts?node_type=&ward_id=
func (h *MapHandler) GetDepartments(c *gin.Context) {
	nodeType := c.Query("node_type")
	wardID := parseUint32Query(c, "ward_id")
	result, err := h.svc.GetDepartments(nodeType, wardID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [21] SearchLocation — GET /api/map/search_location?keyword=&floor_id=
func (h *MapHandler) SearchLocation(c *gin.Context) {
	keyword := c.Query("keyword")
	floorID := parseUint32Query(c, "floor_id")
	result, err := h.svc.SearchLocation(keyword, floorID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [22] GetLandmarks — GET /api/map/get_landmarks?floor_id=
func (h *MapHandler) GetLandmarks(c *gin.Context) {
	floorID := parseUint32Query(c, "floor_id")
	result, err := h.svc.GetLandmarks(floorID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [24] SyncFull — GET /api/map/sync_full
func (h *MapHandler) SyncFull(c *gin.Context) {
	result, err := h.svc.SyncFull()
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// ========================================
// ADMIN WRITE APIs (can token + role admin)
// ========================================

// [25] AddNode — POST /api/admin/add_node
func (h *MapHandler) AddNode(c *gin.Context) {
	var req addNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	input := service.AddNodeInput{
		FloorID:              req.FloorID,
		WardID:               req.WardID,
		NodeCode:             req.NodeCode,
		NodeName:             req.NodeName,
		NodeType:             req.NodeType,
		PolygonCoords:        req.PolygonCoords,
		CenterX:              req.CenterX,
		CenterY:              req.CenterY,
		AccessX:              req.AccessX,
		AccessY:              req.AccessY,
		IsLandmark:           req.IsLandmark,
		WheelchairAccessible: req.WheelchairAccessible,
	}

	result, err := h.svc.AddNode(input)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [26] EditNode — POST /api/admin/edit_node
func (h *MapHandler) EditNode(c *gin.Context) {
	var req editNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if req.NodeID == 0 {
		response.ErrMissingParam(c)
		return
	}

	input := service.EditNodeInput{
		NodeID:               req.NodeID,
		NodeCode:             req.NodeCode,
		NodeName:             req.NodeName,
		NodeType:             req.NodeType,
		PolygonCoords:        req.PolygonCoords,
		CenterX:              req.CenterX,
		CenterY:              req.CenterY,
		AccessX:              req.AccessX,
		AccessY:              req.AccessY,
		IsLandmark:           req.IsLandmark,
		WheelchairAccessible: req.WheelchairAccessible,
		IsAccessible:         req.IsAccessible,
	}

	result, err := h.svc.EditNode(input)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [27] DelNode — DELETE /api/admin/del_node?node_id=
func (h *MapHandler) DelNode(c *gin.Context) {
	nodeID := parseUint32Query(c, "node_id")
	if nodeID == 0 {
		response.ErrMissingParam(c)
		return
	}

	if err := h.svc.DelNode(nodeID); err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, nil)
}

// [28] AddEdge — POST /api/admin/add_edge
func (h *MapHandler) AddEdge(c *gin.Context) {
	var req addEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	input := service.AddEdgeInput{
		FloorID:              req.FloorID,
		FromNodeID:           req.FromNodeID,
		ToNodeID:             req.ToNodeID,
		PolygonCoords:        req.PolygonCoords,
		DistanceM:            req.DistanceM,
		Weight:               req.Weight,
		IsBidirectional:      req.IsBidirectional,
		IsCrossFloor:         req.IsCrossFloor,
		WheelchairAccessible: req.WheelchairAccessible,
	}

	result, err := h.svc.AddEdge(input)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [29] DelEdge — DELETE /api/admin/del_edge?edge_id=
func (h *MapHandler) DelEdge(c *gin.Context) {
	edgeID := parseUint32Query(c, "edge_id")
	if edgeID == 0 {
		response.ErrMissingParam(c)
		return
	}

	if err := h.svc.DelEdge(edgeID); err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, nil)
}

// [30] SetWeight — PATCH /api/admin/set_weight
func (h *MapHandler) SetWeight(c *gin.Context) {
	var req setWeightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if req.EdgeID == 0 || req.Weight <= 0 {
		response.ErrMissingParam(c)
		return
	}

	if err := h.svc.SetWeight(req.EdgeID, req.Weight); err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, nil)
}
