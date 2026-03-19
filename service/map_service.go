package service

import (
	"errors"
	"math"
	"strings"

	"hospital/repository"
	"hospital/schema"
)

// Cac loi nghiep vu cho MapService
var (
	ErrFloorNotFound = errors.New("floor not found")
	ErrNodeNotFound  = errors.New("node not found")
	ErrEdgeNotFound  = errors.New("edge not found")
	ErrNodeCodeExist = errors.New("node code already exists")
	ErrMissingField  = errors.New("missing required field")
)

// MapService xu ly logic nghiep vu cho 14 API ban do (16-30, bo 23).
// Nhan du lieu tu handler, goi repository de truy van DB,
// xu ly logic, tra ket qua hoac loi.
type MapService struct {
	repo *repository.MapRepo
}

func NewMapService(repo *repository.MapRepo) *MapService {
	return &MapService{repo: repo}
}

// ========================================
// RETURN TYPES
// ========================================

// FloorItem la output cho moi tang trong get_floors.
type FloorItem struct {
	FloorID       uint32  `json:"floor_id"`
	BuildingID    uint32  `json:"building_id"`
	BuildingName  string  `json:"building_name"`
	BuildingCode  string  `json:"building_code"`
	FloorNumber   int8    `json:"floor_number"`
	FloorName     string  `json:"floor_name"`
	DisplayOrder  int     `json:"display_order"`
	MapImageURL   *string `json:"map_image_url"`
}

// FloorMetaResult la output cho get_meta.
type FloorMetaResult struct {
	FloorID       uint32  `json:"floor_id"`
	FloorName     string  `json:"floor_name"`
	MapImageURL   *string `json:"map_image_url"`
	ImageWidthPx  int     `json:"image_width_px"`
	ImageHeightPx int     `json:"image_height_px"`
	RealWidthM    float32 `json:"real_width_m"`
	RealHeightM   float32 `json:"real_height_m"`
}

// NodeItem la output cho moi node.
type NodeItem struct {
	NodeID               uint32   `json:"node_id"`
	FloorID              uint32   `json:"floor_id"`
	NodeCode             string   `json:"node_code"`
	NodeName             string   `json:"node_name"`
	NodeType             string   `json:"node_type"`
	PolygonCoords        string   `json:"polygon_coords"`
	CenterX              float32  `json:"center_x"`
	CenterY              float32  `json:"center_y"`
	AccessX              *float32 `json:"access_x"`
	AccessY              *float32 `json:"access_y"`
	IsLandmark           bool     `json:"is_landmark"`
	IsAccessible         bool     `json:"is_accessible"`
	WheelchairAccessible bool     `json:"wheelchair_accessible"`
}

// EdgeItem la output cho moi edge.
type EdgeItem struct {
	EdgeID               uint32   `json:"edge_id"`
	FloorID              uint32   `json:"floor_id"`
	FromNodeID           uint32   `json:"from_node_id"`
	ToNodeID             uint32   `json:"to_node_id"`
	PolygonCoords        *string  `json:"polygon_coords"`
	DistanceM            float32  `json:"distance_m"`
	Weight               float32  `json:"weight"`
	IsBidirectional      bool     `json:"is_bidirectional"`
	IsCrossFloor         bool     `json:"is_cross_floor"`
	WheelchairAccessible bool     `json:"wheelchair_accessible"`
}

// SyncFullResult la output cho sync_full.
type SyncFullResult struct {
	Buildings []schema.Building `json:"buildings"`
	Floors    []FloorItem       `json:"floors"`
	Nodes     []NodeItem        `json:"nodes"`
	Edges     []EdgeItem        `json:"edges"`
}

// ========================================
// INPUT TYPES — Admin APIs
// ========================================

// AddNodeInput la input cho admin add_node.
type AddNodeInput struct {
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

// EditNodeInput la input cho admin edit_node.
type EditNodeInput struct {
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

// AddEdgeInput la input cho admin add_edge.
type AddEdgeInput struct {
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

// SetWeightInput la input cho admin set_weight.
type SetWeightInput struct {
	EdgeID uint32  `json:"edge_id"`
	Weight float32 `json:"weight"`
}

// ========================================
// PRIVATE HELPERS
// ========================================

func nodeToItem(n schema.MapNode) NodeItem {
	return NodeItem{
		NodeID:               n.NodeID,
		FloorID:              n.FloorID,
		NodeCode:             n.NodeCode,
		NodeName:             n.NodeName,
		NodeType:             string(n.NodeType),
		PolygonCoords:        n.PolygonCoords,
		CenterX:              n.CenterX,
		CenterY:              n.CenterY,
		AccessX:              n.AccessX,
		AccessY:              n.AccessY,
		IsLandmark:           n.IsLandmark,
		IsAccessible:         n.IsAccessible,
		WheelchairAccessible: n.WheelchairAccessible,
	}
}

func edgeToItem(e schema.MapEdge) EdgeItem {
	return EdgeItem{
		EdgeID:               e.EdgeID,
		FloorID:              e.FloorID,
		FromNodeID:           e.FromNodeID,
		ToNodeID:             e.ToNodeID,
		PolygonCoords:        e.PolygonCoords,
		DistanceM:            e.DistanceM,
		Weight:               e.Weight,
		IsBidirectional:      e.IsBidirectional,
		IsCrossFloor:         e.IsCrossFloor,
		WheelchairAccessible: e.WheelchairAccessible,
	}
}

func nodesToItems(nodes []schema.MapNode) []NodeItem {
	items := make([]NodeItem, len(nodes))
	for i, n := range nodes {
		items[i] = nodeToItem(n)
	}
	return items
}

func edgesToItems(edges []schema.MapEdge) []EdgeItem {
	items := make([]EdgeItem, len(edges))
	for i, e := range edges {
		items[i] = edgeToItem(e)
	}
	return items
}

// ========================================
// PUBLIC METHODS — 8 READ APIs
// ========================================

// [16] GetFloors tra ve danh sach tang kem building info.
func (s *MapService) GetFloors() ([]FloorItem, error) {
	floors, err := s.repo.FindAllFloors()
	if err != nil {
		return nil, err
	}

	// Lay buildings de map building_name/code
	buildings, err := s.repo.FindAllBuildings()
	if err != nil {
		return nil, err
	}
	bMap := map[uint32]schema.Building{}
	for _, b := range buildings {
		bMap[b.BuildingID] = b
	}

	items := make([]FloorItem, len(floors))
	for i, f := range floors {
		b := bMap[f.BuildingID]
		items[i] = FloorItem{
			FloorID:      f.FloorID,
			BuildingID:   f.BuildingID,
			BuildingName: b.BuildingName,
			BuildingCode: b.BuildingCode,
			FloorNumber:  f.FloorNumber,
			FloorName:    f.FloorName,
			DisplayOrder: f.DisplayOrder,
			MapImageURL:  f.MapImageURL,
		}
	}
	return items, nil
}

// [17] GetNodes tra ve nodes cua 1 tang.
func (s *MapService) GetNodes(floorID uint32) ([]NodeItem, error) {
	nodes, err := s.repo.FindAllNodes(floorID)
	if err != nil {
		return nil, err
	}
	return nodesToItems(nodes), nil
}

// [18] GetEdges tra ve edges cua 1 tang.
func (s *MapService) GetEdges(floorID uint32) ([]EdgeItem, error) {
	edges, err := s.repo.FindAllEdges(floorID)
	if err != nil {
		return nil, err
	}
	return edgesToItems(edges), nil
}

// [19] GetMeta tra ve metadata tang (kich thuoc, ty le pixel->met).
func (s *MapService) GetMeta(floorID uint32) (*FloorMetaResult, error) {
	if floorID == 0 {
		return nil, ErrMissingField
	}
	floor, err := s.repo.FindFloorByID(floorID)
	if err != nil {
		return nil, err
	}
	if floor == nil {
		return nil, ErrFloorNotFound
	}
	return &FloorMetaResult{
		FloorID:       floor.FloorID,
		FloorName:     floor.FloorName,
		MapImageURL:   floor.MapImageURL,
		ImageWidthPx:  floor.ImageWidthPx,
		ImageHeightPx: floor.ImageHeightPx,
		RealWidthM:    floor.RealWidthM,
		RealHeightM:   floor.RealHeightM,
	}, nil
}

// [20] GetDepartments loc nodes theo loai/khoa.
func (s *MapService) GetDepartments(nodeType string, wardID uint32) ([]NodeItem, error) {
	nodes, err := s.repo.FindNodesByType(schema.NodeType(nodeType), wardID)
	if err != nil {
		return nil, err
	}
	return nodesToItems(nodes), nil
}

// [21] SearchLocation tim kiem node theo keyword.
func (s *MapService) SearchLocation(keyword string, floorID uint32) ([]NodeItem, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []NodeItem{}, nil
	}
	nodes, err := s.repo.SearchNodes(keyword, floorID)
	if err != nil {
		return nil, err
	}
	return nodesToItems(nodes), nil
}

// [22] GetLandmarks tra ve cac diem moc.
func (s *MapService) GetLandmarks(floorID uint32) ([]NodeItem, error) {
	nodes, err := s.repo.FindLandmarks(floorID)
	if err != nil {
		return nil, err
	}
	return nodesToItems(nodes), nil
}

// [24] SyncFull tra ve toan bo du lieu ban do.
func (s *MapService) SyncFull() (*SyncFullResult, error) {
	syncData, err := s.repo.FindSyncData()
	if err != nil {
		return nil, err
	}

	// Build floors with building info
	bMap := map[uint32]schema.Building{}
	for _, b := range syncData.Buildings {
		bMap[b.BuildingID] = b
	}
	floorItems := make([]FloorItem, len(syncData.Floors))
	for i, f := range syncData.Floors {
		b := bMap[f.BuildingID]
		floorItems[i] = FloorItem{
			FloorID:      f.FloorID,
			BuildingID:   f.BuildingID,
			BuildingName: b.BuildingName,
			BuildingCode: b.BuildingCode,
			FloorNumber:  f.FloorNumber,
			FloorName:    f.FloorName,
			DisplayOrder: f.DisplayOrder,
			MapImageURL:  f.MapImageURL,
		}
	}

	return &SyncFullResult{
		Buildings: syncData.Buildings,
		Floors:    floorItems,
		Nodes:     nodesToItems(syncData.Nodes),
		Edges:     edgesToItems(syncData.Edges),
	}, nil
}

// ========================================
// PUBLIC METHODS — 6 ADMIN APIs
// ========================================

// [25] AddNode them 1 node moi.
func (s *MapService) AddNode(input AddNodeInput) (*NodeItem, error) {
	code := strings.TrimSpace(input.NodeCode)
	name := strings.TrimSpace(input.NodeName)
	if code == "" || name == "" || input.FloorID == 0 {
		return nil, ErrMissingField
	}

	// Kiem tra floor ton tai
	floor, err := s.repo.FindFloorByID(input.FloorID)
	if err != nil {
		return nil, err
	}
	if floor == nil {
		return nil, ErrFloorNotFound
	}

	// Kiem tra trung node_code
	existing, err := s.repo.FindNodeByCode(code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrNodeCodeExist
	}

	node := &schema.MapNode{
		FloorID:              input.FloorID,
		WardID:               input.WardID,
		NodeCode:             code,
		NodeName:             name,
		NodeType:             schema.NodeType(input.NodeType),
		PolygonCoords:        input.PolygonCoords,
		CenterX:              input.CenterX,
		CenterY:              input.CenterY,
		AccessX:              input.AccessX,
		AccessY:              input.AccessY,
		IsLandmark:           input.IsLandmark,
		IsAccessible:         true,
		WheelchairAccessible: input.WheelchairAccessible,
		IsActive:             true,
	}

	if err := s.repo.CreateNode(node); err != nil {
		return nil, err
	}

	item := nodeToItem(*node)
	return &item, nil
}

// [26] EditNode cap nhat thong tin node.
func (s *MapService) EditNode(input EditNodeInput) (*NodeItem, error) {
	if input.NodeID == 0 {
		return nil, ErrMissingField
	}

	node, err := s.repo.FindNodeByID(input.NodeID)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, ErrNodeNotFound
	}

	updates := map[string]interface{}{}

	if input.NodeCode != nil {
		c := strings.TrimSpace(*input.NodeCode)
		if c != "" && c != node.NodeCode {
			// Kiem tra trung
			existing, err := s.repo.FindNodeByCode(c)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.NodeID != node.NodeID {
				return nil, ErrNodeCodeExist
			}
			updates["node_code"] = c
		}
	}
	if input.NodeName != nil {
		updates["node_name"] = strings.TrimSpace(*input.NodeName)
	}
	if input.NodeType != nil {
		updates["node_type"] = *input.NodeType
	}
	if input.PolygonCoords != nil {
		updates["polygon_coords"] = *input.PolygonCoords
	}
	if input.CenterX != nil {
		updates["center_x"] = *input.CenterX
	}
	if input.CenterY != nil {
		updates["center_y"] = *input.CenterY
	}
	if input.AccessX != nil {
		updates["access_x"] = *input.AccessX
	}
	if input.AccessY != nil {
		updates["access_y"] = *input.AccessY
	}
	if input.IsLandmark != nil {
		updates["is_landmark"] = *input.IsLandmark
	}
	if input.WheelchairAccessible != nil {
		updates["wheelchair_accessible"] = *input.WheelchairAccessible
	}
	if input.IsAccessible != nil {
		updates["is_accessible"] = *input.IsAccessible
	}

	if len(updates) > 0 {
		if err := s.repo.UpdateNode(input.NodeID, updates); err != nil {
			return nil, err
		}
	}

	// Lay lai node sau khi cap nhat
	updated, err := s.repo.FindNodeByID(input.NodeID)
	if err != nil {
		return nil, err
	}
	item := nodeToItem(*updated)
	return &item, nil
}

// [27] DelNode xoa (soft delete) node va cac edge lien quan.
func (s *MapService) DelNode(nodeID uint32) error {
	if nodeID == 0 {
		return ErrMissingField
	}

	node, err := s.repo.FindNodeByID(nodeID)
	if err != nil {
		return err
	}
	if node == nil {
		return ErrNodeNotFound
	}

	// Xoa cac edge lien quan truoc
	if err := s.repo.DeactivateEdgesByNode(nodeID); err != nil {
		return err
	}

	return s.repo.DeactivateNode(nodeID)
}

// [28] AddEdge them 1 edge moi.
func (s *MapService) AddEdge(input AddEdgeInput) (*EdgeItem, error) {
	if input.FromNodeID == 0 || input.ToNodeID == 0 || input.FloorID == 0 {
		return nil, ErrMissingField
	}

	// Kiem tra 2 node ton tai
	fromNode, err := s.repo.FindNodeByID(input.FromNodeID)
	if err != nil {
		return nil, err
	}
	if fromNode == nil {
		return nil, ErrNodeNotFound
	}

	toNode, err := s.repo.FindNodeByID(input.ToNodeID)
	if err != nil {
		return nil, err
	}
	if toNode == nil {
		return nil, ErrNodeNotFound
	}

	// Tu dong tinh distance_m neu khong cung cap
	distM := float32(0)
	if input.DistanceM != nil && *input.DistanceM > 0 {
		distM = *input.DistanceM
	} else {
		// Tinh tu access_x/y cua 2 node
		floor, _ := s.repo.FindFloorByID(input.FloorID)
		if floor != nil && floor.ImageWidthPx > 0 && floor.ImageHeightPx > 0 {
			sx := floor.RealWidthM / float32(floor.ImageWidthPx)
			sy := floor.RealHeightM / float32(floor.ImageHeightPx)
			fx, fy := fromNode.CenterX, fromNode.CenterY
			tx, ty := toNode.CenterX, toNode.CenterY
			if fromNode.AccessX != nil {
				fx = *fromNode.AccessX
			}
			if fromNode.AccessY != nil {
				fy = *fromNode.AccessY
			}
			if toNode.AccessX != nil {
				tx = *toNode.AccessX
			}
			if toNode.AccessY != nil {
				ty = *toNode.AccessY
			}
			dx := float64((tx - fx) * sx)
			dy := float64((ty - fy) * sy)
			distM = float32(math.Sqrt(dx*dx + dy*dy))
		}
	}

	weight := input.Weight
	if weight <= 0 {
		weight = 1.0
	}

	edge := &schema.MapEdge{
		FloorID:              input.FloorID,
		FromNodeID:           input.FromNodeID,
		ToNodeID:             input.ToNodeID,
		PolygonCoords:        input.PolygonCoords,
		DistanceM:            distM,
		Weight:               weight,
		IsBidirectional:      input.IsBidirectional,
		IsCrossFloor:         input.IsCrossFloor,
		WheelchairAccessible: input.WheelchairAccessible,
		IsActive:             true,
	}

	if err := s.repo.CreateEdge(edge); err != nil {
		return nil, err
	}

	item := edgeToItem(*edge)
	return &item, nil
}

// [29] DelEdge xoa (soft delete) 1 edge.
func (s *MapService) DelEdge(edgeID uint32) error {
	if edgeID == 0 {
		return ErrMissingField
	}

	edge, err := s.repo.FindEdgeByID(edgeID)
	if err != nil {
		return err
	}
	if edge == nil {
		return ErrEdgeNotFound
	}

	return s.repo.DeactivateEdge(edgeID)
}

// [30] SetWeight cap nhat trong so cua 1 edge.
func (s *MapService) SetWeight(edgeID uint32, weight float32) error {
	if edgeID == 0 {
		return ErrMissingField
	}
	if weight <= 0 {
		return ErrMissingField
	}

	edge, err := s.repo.FindEdgeByID(edgeID)
	if err != nil {
		return err
	}
	if edge == nil {
		return ErrEdgeNotFound
	}

	return s.repo.UpdateEdgeWeight(edgeID, weight)
}
