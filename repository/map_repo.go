package repository

import (
	"hospital/schema"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MapRepo xu ly truy van database cho toan bo Map module.
// Bao gom 4 bang: buildings, floors, map_nodes, map_edges.
// Chi chua CRUD thuan tuy, logic nghiep vu nam o service layer.
type MapRepo struct {
	db *gorm.DB
}

func NewMapRepo(db *gorm.DB) *MapRepo {
	return &MapRepo{db: db}
}

// ========================================
// BUILDINGS
// ========================================

// FindAllBuildings tra ve tat ca toa nha dang hoat dong.
func (r *MapRepo) FindAllBuildings() ([]schema.Building, error) {
	var buildings []schema.Building
	err := r.db.Where("is_active = ?", true).
		Order("building_code ASC").
		Find(&buildings).Error
	return buildings, err
}

// ========================================
// FLOORS — API 16 (get_floors), API 19 (get_meta)
// ========================================

// FindAllFloors tra ve tat ca tang dang hoat dong cua moi toa nha.
// Dung cho API get_floors: render bo loc tang tren App.
func (r *MapRepo) FindAllFloors() ([]schema.Floor, error) {
	var floors []schema.Floor
	err := r.db.Where("is_active = ?", true).
		Order("display_order ASC, floor_number ASC").
		Find(&floors).Error
	return floors, err
}

// FindFloorsByBuilding tra ve cac tang cua 1 toa nha cu the.
func (r *MapRepo) FindFloorsByBuilding(buildingID uint32) ([]schema.Floor, error) {
	var floors []schema.Floor
	err := r.db.Where("building_id = ? AND is_active = ?", buildingID, true).
		Order("display_order ASC, floor_number ASC").
		Find(&floors).Error
	return floors, err
}

// FindFloorByID tra ve 1 tang theo ID. Dung khi can lay he so quy doi pixel→met.
func (r *MapRepo) FindFloorByID(floorID uint32) (*schema.Floor, error) {
	var floor schema.Floor
	err := r.db.First(&floor, "floor_id = ? AND is_active = ?", floorID, true).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &floor, err
}

// ========================================
// MAP NODES — API 17, 20, 21, 22, 25, 26, 27
// ========================================

// FindAllNodes tra ve tat ca node dang hoat dong.
// Neu floorID > 0 thi loc theo tang. Dung cho API get_nodes.
func (r *MapRepo) FindAllNodes(floorID uint32) ([]schema.MapNode, error) {
	var nodes []schema.MapNode
	q := r.db.Where("is_active = ?", true)
	if floorID > 0 {
		q = q.Where("floor_id = ?", floorID)
	}
	err := q.Order("node_id ASC").Find(&nodes).Error
	return nodes, err
}

// FindNodesByType loc node theo node_type. Dung cho API get_depts.
// Neu wardID > 0 thi loc kem theo khoa.
func (r *MapRepo) FindNodesByType(nodeType schema.NodeType, wardID uint32) ([]schema.MapNode, error) {
	var nodes []schema.MapNode
	q := r.db.Where("is_active = ?", true)
	if nodeType != "" {
		q = q.Where("node_type = ?", nodeType)
	}
	if wardID > 0 {
		q = q.Where("ward_id = ?", wardID)
	}
	err := q.Order("node_name ASC").Find(&nodes).Error
	return nodes, err
}

// SearchNodes tim kiem node theo tu khoa (case-insensitive LIKE).
// Dung cho API search_location.
// SQLite: dung LIKE. PostgreSQL: co the nang cap len tsvector/GIN index
// chi can sua duy nhat ham nay.
func (r *MapRepo) SearchNodes(keyword string, floorID uint32) ([]schema.MapNode, error) {
	var nodes []schema.MapNode
	q := r.db.Where("is_active = ? AND node_name LIKE ?", true, "%"+keyword+"%")
	if floorID > 0 {
		q = q.Where("floor_id = ?", floorID)
	}
	err := q.Order("is_landmark DESC, node_name ASC").Limit(30).Find(&nodes).Error
	return nodes, err
}

// FindLandmarks tra ve cac diem moc (is_landmark = true).
// Dung cho API get_landmarks.
func (r *MapRepo) FindLandmarks(floorID uint32) ([]schema.MapNode, error) {
	var nodes []schema.MapNode
	q := r.db.Where("is_active = ? AND is_landmark = ?", true, true)
	if floorID > 0 {
		q = q.Where("floor_id = ?", floorID)
	}
	err := q.Order("node_name ASC").Find(&nodes).Error
	return nodes, err
}

// FindNodeByID tra ve 1 node theo ID.
func (r *MapRepo) FindNodeByID(nodeID uint32) (*schema.MapNode, error) {
	var node schema.MapNode
	err := r.db.First(&node, "node_id = ? AND is_active = ?", nodeID, true).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &node, err
}

// FindNodeByCode tra ve 1 node theo node_code. Kiem tra trung truoc khi create.
func (r *MapRepo) FindNodeByCode(code string) (*schema.MapNode, error) {
	var node schema.MapNode
	err := r.db.First(&node, "node_code = ?", code).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &node, err
}

// CreateNode tao 1 node moi. Dung Omit(Associations) nhu User.Create.
func (r *MapRepo) CreateNode(node *schema.MapNode) error {
	return r.db.Omit(clause.Associations).Create(node).Error
}

// UpdateNode cap nhat thong tin node theo map fields.
// Chi update cac truong duoc truyen vao, khong overwrite toan bo.
func (r *MapRepo) UpdateNode(nodeID uint32, updates map[string]interface{}) error {
	return r.db.Model(&schema.MapNode{}).
		Where("node_id = ?", nodeID).
		Updates(updates).Error
}

// DeactivateNode soft delete node (is_active = false).
// Kiem tra xem node co edge lien quan con active khong phai lam o service layer.
func (r *MapRepo) DeactivateNode(nodeID uint32) error {
	return r.db.Model(&schema.MapNode{}).
		Where("node_id = ?", nodeID).
		Update("is_active", false).Error
}

// ========================================
// MAP EDGES — API 18, 28, 29, 30
// ========================================

// FindAllEdges tra ve tat ca canh dang hoat dong.
// Neu floorID > 0 thi loc theo tang. Dung cho API get_edges + sync_full.
func (r *MapRepo) FindAllEdges(floorID uint32) ([]schema.MapEdge, error) {
	var edges []schema.MapEdge
	q := r.db.Where("is_active = ?", true)
	if floorID > 0 {
		q = q.Where("floor_id = ?", floorID)
	}
	err := q.Order("edge_id ASC").Find(&edges).Error
	return edges, err
}

// FindEdgeByID tra ve 1 edge theo ID.
func (r *MapRepo) FindEdgeByID(edgeID uint32) (*schema.MapEdge, error) {
	var edge schema.MapEdge
	err := r.db.First(&edge, "edge_id = ? AND is_active = ?", edgeID, true).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &edge, err
}

// FindEdgesByNode tra ve tat ca edge co lien quan den 1 node (from hoac to).
// Dung truoc khi xoa node de kiem tra edge phu thuoc.
func (r *MapRepo) FindEdgesByNode(nodeID uint32) ([]schema.MapEdge, error) {
	var edges []schema.MapEdge
	err := r.db.Where(
		"is_active = ? AND (from_node_id = ? OR to_node_id = ?)",
		true, nodeID, nodeID,
	).Find(&edges).Error
	return edges, err
}

// CreateEdge tao 1 canh hanh lang moi.
func (r *MapRepo) CreateEdge(edge *schema.MapEdge) error {
	return r.db.Omit(clause.Associations).Create(edge).Error
}

// UpdateEdgeWeight cap nhat trong so cua 1 edge. Dung cho API set_weight.
func (r *MapRepo) UpdateEdgeWeight(edgeID uint32, weight float32) error {
	return r.db.Model(&schema.MapEdge{}).
		Where("edge_id = ?", edgeID).
		Update("weight", weight).Error
}

// DeactivateEdge soft delete edge (is_active = false). Dung cho API del_edge.
func (r *MapRepo) DeactivateEdge(edgeID uint32) error {
	return r.db.Model(&schema.MapEdge{}).
		Where("edge_id = ?", edgeID).
		Update("is_active", false).Error
}

// DeactivateEdgesByNode soft delete tat ca edge lien quan khi xoa node.
// Goi khi admin del_node de giu tinh nhat quan cua do thi.
func (r *MapRepo) DeactivateEdgesByNode(nodeID uint32) error {
	return r.db.Model(&schema.MapEdge{}).
		Where("from_node_id = ? OR to_node_id = ?", nodeID, nodeID).
		Update("is_active", false).Error
}

// ========================================
// SYNC — API 24 (sync_full)
// ========================================

// SyncResult chua toan bo du lieu ban do cho 1 lan goi sync_full.
type SyncResult struct {
	Buildings []schema.Building `json:"buildings"`
	Floors    []schema.Floor    `json:"floors"`
	Nodes     []schema.MapNode  `json:"nodes"`
	Edges     []schema.MapEdge  `json:"edges"`
}

// FindSyncData lay tat ca du lieu ban do trong 1 lan truy van.
// Dung cho API sync_full: client luu cache de dung offline.
func (r *MapRepo) FindSyncData() (*SyncResult, error) {
	var result SyncResult
	var err error

	result.Buildings, err = r.FindAllBuildings()
	if err != nil {
		return nil, err
	}

	result.Floors, err = r.FindAllFloors()
	if err != nil {
		return nil, err
	}

	result.Nodes, err = r.FindAllNodes(0) // 0 = khong loc theo tang
	if err != nil {
		return nil, err
	}

	result.Edges, err = r.FindAllEdges(0)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
