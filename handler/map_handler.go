package handler

import (
	"encoding/json"
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
	ID                   string  `json:"id" binding:"required"` // corresponds to POICode
	MapID                uint32  `json:"map_id" binding:"required"`
	Name                 string  `json:"name" binding:"required"`
	Type                 string  `json:"type" binding:"required"`
	X                    int     `json:"x"`
	Y                    int     `json:"y"`
	WardID               *uint32 `json:"ward_id"`
	IsLandmark           bool    `json:"is_landmark"`
	WheelchairAccessible bool    `json:"wheelchair_accessible"`
	Capacity             *int    `json:"capacity"`
	Details              *string `json:"details"`
	OpenHours            *string `json:"open_hours"`
}

type editNodeRequest struct {
	ID                   string   `json:"id" binding:"required"` // corresponds to POICode
	X                    *int     `json:"x"`
	Y                    *int     `json:"y"`
	Name                 *string  `json:"name"`
	Type                 *string  `json:"type"`
	IsLandmark           *bool    `json:"is_landmark"`
	WheelchairAccessible *bool    `json:"wheelchair_accessible"`
	IsAccessible         *bool    `json:"is_accessible"`
	Capacity             *int     `json:"capacity"`
	Details              *string  `json:"details"`
	OpenHours            *string  `json:"open_hours"`
	CustomWeight         *float32 `json:"custom_weight"`
}

type delNodeRequest struct {
	ID string `json:"id" binding:"required"`
}

type addEdgeRequest struct {
	MapID     uint32   `json:"map_id"`
	StartNode string   `json:"start_node"`
	EndNode   string   `json:"end_node"`
	Distance  *float32 `json:"distance"`
}

type editEdgeRequest struct {
	ID       uint32   `json:"id"`
	Distance *float32 `json:"distance"`
}

type delEdgeRequest struct {
	ID uint32 `json:"id"`
}

type setWeightRequest struct {
	EdgeID *uint32  `json:"edge_id"`
	Weight *float32 `json:"weight"`
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
	mapIDStr := c.Query("map_id")
	if mapIDStr == "" {
		response.ErrMissingParam(c)
		return
	}
	mapIDInt, err := strconv.ParseInt(mapIDStr, 10, 64)
	if err != nil || mapIDStr[0] == '-' {
		response.ErrInvalidType(c)
		return
	}
	if mapIDInt <= 0 || mapIDInt > 2147483647 {
		response.ErrInvalidValue(c)
		return
	}
	mapID := uint32(mapIDInt)

	// Verify map exists
	if exists, err := h.svc.MapExists(mapID); err != nil {
		response.ErrUnexpected(c)
		return
	} else if !exists {
		response.SuccessWithCode(c, 4001, nil)
		return
	}

	items, err2 := h.svc.GetNodes(mapID)
	if err2 != nil {
		response.ErrUnexpected(c)
		return
	}
	response.Success(c, items)
}

// [18] GET /api/map/get_edges?map_id=
func (h *MapHandler) GetEdges(c *gin.Context) {
	mapIDStr := c.Query("map_id")
	if mapIDStr == "" {
		response.ErrMissingParam(c)
		return
	}
	mapIDInt, err := strconv.ParseInt(mapIDStr, 10, 64)
	if err != nil || mapIDStr[0] == '-' {
		response.ErrInvalidType(c)
		return
	}
	if mapIDInt <= 0 || mapIDInt > 2147483647 {
		response.ErrInvalidValue(c)
		return
	}
	mapID := uint32(mapIDInt)

	result, err2 := h.svc.GetEdges(mapID)
	if err2 != nil {
		h.handleMapError(c, err2)
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
	// Parse raw body to distinguish missing vs wrong type
	var rawBody map[string]interface{}
	if err := c.ShouldBindJSON(&rawBody); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	// Validate required fields
	if _, ok := rawBody["id"]; !ok {
		response.ErrMissingParam(c)
		return
	}
	if _, ok := rawBody["map_id"]; !ok {
		response.ErrMissingParam(c)
		return
	}
	if _, ok := rawBody["name"]; !ok {
		response.ErrMissingParam(c)
		return
	}
	if _, ok := rawBody["type"]; !ok {
		response.ErrMissingParam(c)
		return
	}

	// Type validation: id must be string
	idVal, idOk := rawBody["id"].(string)
	if !idOk || idVal == "" {
		response.ErrInvalidType(c)
		return
	}

	// Re-bind into typed struct for convenience
	body, _ := json.Marshal(rawBody)
	var req addNodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	item, err := h.svc.AddNode(service.AddNodeInput{
		MapID:                req.MapID,
		WardID:               req.WardID,
		POICode:              req.ID,
		POIName:              req.Name,
		POIType:              req.Type,
		GridRow:              req.Y,
		GridCol:              req.X,
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
		POICode:              req.ID,
		POIName:              req.Name,
		POIType:              req.Type,
		GridRow:              req.Y,
		GridCol:              req.X,
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

// [27] DELETE /api/admin/del_node
func (h *MapHandler) DelNode(c *gin.Context) {
	var req delNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	err := h.svc.DelNode(req.ID)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, nil)
}

// [28] POST /api/admin/add_edge
func (h *MapHandler) AddEdge(c *gin.Context) {
	var req addEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	// Validate required fields
	if req.MapID == 0 {
		response.ErrMissingParam(c)
		return
	}
	if req.StartNode == "" || req.EndNode == "" {
		response.ErrMissingParam(c)
		return
	}
	if req.Distance == nil {
		response.ErrMissingParam(c)
		return
	}
	if *req.Distance <= 0 {
		response.ErrInvalidValue(c)
		return
	}

	edgeID, err := h.svc.AddEdge(req.MapID, req.StartNode, req.EndNode, *req.Distance)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, map[string]interface{}{"id": edgeID})
}

// [29] DELETE /api/admin/del_edge
func (h *MapHandler) DelEdge(c *gin.Context) {
	var req delEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}
	if req.ID == 0 {
		response.ErrMissingParam(c)
		return
	}

	if err := h.svc.DelEdge(req.ID); err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, nil)
}

// [30] PATCH /api/admin/set_weight
func (h *MapHandler) SetWeight(c *gin.Context) {
	// Parse raw to distinguish missing vs wrong type vs invalid value
	var rawBody map[string]interface{}
	if err := c.ShouldBindJSON(&rawBody); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	// Check missing fields
	rawEdgeID, hasEdgeID := rawBody["edge_id"]
	rawWeight, hasWeight := rawBody["weight"]
	if !hasEdgeID {
		response.ErrMissingParam(c)
		return
	}
	if !hasWeight {
		response.ErrMissingParam(c)
		return
	}

	// Type check: edge_id must be number
	edgeIDFloat, ok := rawEdgeID.(float64)
	if !ok {
		response.ErrInvalidType(c)
		return
	}
	edgeID := uint32(edgeIDFloat)
	if edgeID == 0 {
		response.ErrMissingParam(c)
		return
	}

	// Type check: weight must be number
	weightFloat, ok := rawWeight.(float64)
	if !ok {
		response.ErrInvalidType(c)
		return
	}
	weight := float32(weightFloat)

	// Value validation
	if weight <= 0 {
		response.ErrInvalidValue(c)
		return
	}

	if err := h.svc.SetWeight(edgeID, weight); err != nil {
		h.handleMapError(c, err)
		return
	}

	response.Success(c, nil)
}

// ========================================
// MAP FILE APIs
// ========================================

// [31] POST /api/admin/upload_map
func (h *MapHandler) UploadMap(c *gin.Context) {
	// 1. Get file
	file, err := c.FormFile("file")
	if err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	mapName := c.PostForm("map_name")
	if mapName == "" {
		mapName = file.Filename
	}
	rows := parseUint32(c.PostForm("rows"))
	cols := parseUint32(c.PostForm("cols"))

	// 2. Save file
	filePath := "data/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		response.ErrInternalError(c)
		return
	}

	// 2.5 Save image file if provided
	var mapImageURL *string
	imageFile, err := c.FormFile("image_file")
	if err == nil && imageFile != nil {
		imagePath := "data/" + imageFile.Filename
		if err := c.SaveUploadedFile(imageFile, imagePath); err == nil {
			// URL path matches the static route we defined in main.go
			url := "/data/" + imageFile.Filename 
			mapImageURL = &url
		}
	}

	// 3. Save to DB
	// We read the file content to save to DB GridData
	// In a real app we would parse the .map file, but for now we just use a dummy grid data or the file content
	gridData := "[]" // Placeholder for actual grid data

	m, err := h.svc.UploadMap(mapName, filePath, int(rows), int(cols), gridData, mapImageURL)
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, m)
}

// [32] POST /api/admin/upload_output
func (h *MapHandler) UploadOutput(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	filePath := "data/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		response.ErrInternalError(c)
		return
	}

	response.Success(c, map[string]string{"file_path": filePath})
}

// [33] POST /api/admin/set_active_map
func (h *MapHandler) SetActiveMap(c *gin.Context) {
	var req struct {
		MapID uint32 `json:"map_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrBodyInvalid(c)
		return
	}

	if err := h.svc.SetActiveMap(req.MapID); err != nil {
		if err.Error() == "cannot change active map: simulation is currently running. Please stop it first." {
			response.ErrorResponse(c, 400, 4011, err.Error())
			return
		}
		h.handleMapError(c, err)
		return
	}

	response.Success(c, nil)
}

// [34] GET /api/admin/get_maps
func (h *MapHandler) GetMaps(c *gin.Context) {
	maps, err := h.svc.GetMaps()
	if err != nil {
		h.handleMapError(c, err)
		return
	}
	response.Success(c, maps)
}

// [35] GET /api/admin/export_map?filename=
func (h *MapHandler) ExportMap(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		response.ErrMissingParam(c)
		return
	}
	c.FileAttachment("data/"+filename, filename)
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
		response.SuccessWithCode(c, 4001, nil)
	case errors.Is(err, service.ErrNodeNotFound):
		response.SuccessWithCode(c, 4001, nil)
	case errors.Is(err, service.ErrEdgeNotFound):
		response.SuccessWithCode(c, 4001, nil)
	case errors.Is(err, service.ErrNodeCodeExist):
		response.ErrorResponse(c, 409, 4009, "POI code already exists")
	case errors.Is(err, service.ErrMissingField):
		response.ErrMissingParam(c)
	case errors.Is(err, service.ErrCellNotFree):
		response.ErrorResponse(c, 400, 4010, "Grid cell is not walkable")
	default:
		response.ErrUnexpected(c)
	}
}
