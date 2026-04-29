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
// REQUEST STRUCTS  - Admin APIs
// ========================================

type addNodeRequest struct {
	MapID                uint32  `json:"map_id"`
	WardID               *uint32 `json:"ward_id"`
	POICode              string  `json:"poi_code"`
	POIName              string  `json:"poi_name"`
	POIType              string  `json:"poi_type"`
	GridRow              int     `json:"grid_row"`
	GridCol              int     `json:"grid_col"`
	IsLandmark           bool    `json:"is_landmark"`
	WheelchairAccessible bool    `json:"wheelchair_accessible"`
	Capacity             *int    `json:"capacity"`
	Details              *string `json:"details"`
	OpenHours            *string `json:"open_hours"`
}

type editNodeRequest struct {
	POIID                uint32   `json:"poi_id"`
	POICode              *string  `json:"poi_code"`
	POIName              *string  `json:"poi_name"`
	POIType              *string  `json:"poi_type"`
	IsLandmark           *bool    `json:"is_landmark"`
	WheelchairAccessible *bool    `json:"wheelchair_accessible"`
	IsAccessible         *bool    `json:"is_accessible"`
	Capacity             *int     `json:"capacity"`
	Details              *string  `json:"details"`
	OpenHours            *string  `json:"open_hours"`
	CustomWeight         *float32 `json:"custom_weight"`
}

type setWeightRequest struct {
	POIID  uint32  `json:"poi_id" binding:"required"`
	Weight float32 `json:"weight" binding:"required"`
}

// ========================================
// PUBLIC APIs [16-24]
// ========================================

// [16] GET /api/map/get_floors
func (h *MapHandler) GetFloors(c *gin.Context) {
	items, err := h.svc.GetFloors()
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, items)
}

// [17] GET /api/map/get_nodes?map_id=
func (h *MapHandler) GetNodes(c *gin.Context) {
	mapID := parseUint32(c.Query("map_id"))
	items, err := h.svc.GetNodes(mapID)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, items)
}

// [18] GET /api/map/get_edges?map_id=
func (h *MapHandler) GetEdges(c *gin.Context) {
	mapID := parseUint32(c.Query("map_id"))
	result, err := h.svc.GetEdges(mapID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, result)
}

// [19] GET /api/map/get_meta?map_id=
func (h *MapHandler) GetMeta(c *gin.Context) {
	mapID := parseUint32(c.Query("map_id"))
	meta, err := h.svc.GetMeta(mapID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, meta)
}

// [20] GET /api/map/get_depts?node_type=&ward_id=
func (h *MapHandler) GetDepartments(c *gin.Context) {
	nodeType := c.Query("node_type")
	wardID := parseUint32(c.Query("ward_id"))
	result, err := h.svc.GetDepartments(nodeType, wardID)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, result)
}

// [21] GET /api/map/search_location?keyword=&map_id=
func (h *MapHandler) SearchLocation(c *gin.Context) {
	keyword := c.Query("keyword")
	mapID := parseUint32(c.Query("map_id"))
	items, err := h.svc.SearchLocation(keyword, mapID)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, items)
}

// [22] GET /api/map/get_landmarks?map_id=
func (h *MapHandler) GetLandmarks(c *gin.Context) {
	mapID := parseUint32(c.Query("map_id"))
	items, err := h.svc.GetLandmarks(mapID)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, items)
}

// [24] GET /api/map/sync_full?map_id=
func (h *MapHandler) SyncFull(c *gin.Context) {
	mapID := parseUint32(c.Query("map_id"))
	result, err := h.svc.SyncFull(mapID)
	if err != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, result)
}

// ========================================
// ADMIN APIs [25-30]
// ========================================

// [25] POST /api/admin/add_node
func (h *MapHandler) AddNode(c *gin.Context) {
	var req addNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	item, err := h.svc.AddNode(service.AddNodeInput{
		MapID:                req.MapID,
		WardID:               req.WardID,
		POICode:              req.POICode,
		POIName:              req.POIName,
		POIType:              req.POIType,
		GridRow:              req.GridRow,
		GridCol:              req.GridCol,
		IsLandmark:           req.IsLandmark,
		WheelchairAccessible: req.WheelchairAccessible,
		Capacity:             req.Capacity,
		Details:              req.Details,
		OpenHours:            req.OpenHours,
	})
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, item)
}

// [26] POST /api/admin/edit_node
func (h *MapHandler) EditNode(c *gin.Context) {
	var req editNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	item, err := h.svc.EditNode(service.EditNodeInput{
		POIID:                req.POIID,
		POICode:              req.POICode,
		POIName:              req.POIName,
		POIType:              req.POIType,
		IsLandmark:           req.IsLandmark,
		WheelchairAccessible: req.WheelchairAccessible,
		IsAccessible:         req.IsAccessible,
		Capacity:             req.Capacity,
		Details:              req.Details,
		OpenHours:            req.OpenHours,
		CustomWeight:         req.CustomWeight,
	})
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, item)
}

// [27] DELETE /api/admin/del_node?poi_id=
func (h *MapHandler) DelNode(c *gin.Context) {
	poiID := parseUint32(c.Query("poi_id"))
	if poiID == 0 {
		// Fallback: try JSON body
		var req struct {
			POIID uint32 `json:"poi_id"`
		}
		_ = c.ShouldBindJSON(&req)
		poiID = req.POIID
	}

	err := h.svc.DelNode(poiID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, nil)
}

// [28] POST /api/admin/add_edge  - grid-based: không hỗ trợ
func (h *MapHandler) AddEdge(c *gin.Context) {
	response.SuccessWithCode(c, 2003, map[string]string{
		"message": "edges are auto-computed from grid adjacency, manual edge creation is not supported",
	})
}

// [29] DELETE /api/admin/del_edge  - grid-based: không hỗ trợ
func (h *MapHandler) DelEdge(c *gin.Context) {
	response.SuccessWithCode(c, 2003, map[string]string{
		"message": "edges are auto-computed from grid adjacency, manual edge deletion is not supported",
	})
}

// [30] PATCH /api/admin/set_weight
func (h *MapHandler) SetWeight(c *gin.Context) {
	var req setWeightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	err := h.svc.SetWeight(req.POIID, req.Weight)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, nil)
}

// ========================================
// HELPERS
// ========================================

func parseUint32(s string) uint32 {
	v, _ := strconv.ParseUint(s, 10, 32)
	return uint32(v)
}

func (h *MapHandler) handleMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrMapNotFound):
		response.ErrNotFound(c)
	case errors.Is(err, service.ErrNodeNotFound):
		response.ErrNotFound(c)
	case errors.Is(err, service.ErrNodeCodeExist):
		response.ErrorResponse(c, 409, 4009, "POI code already exists")
	case errors.Is(err, service.ErrMissingField):
		response.ErrBodyInvalid(c)
	case errors.Is(err, service.ErrCellNotFree):
		response.ErrorResponse(c, 400, 4010, "Grid cell is not walkable")
	default:
		response.ErrUnexpected(c)
	}
}
