package schema

import "time"

// ========================================
// Bảng: buildings [T08]
// Tòa nhà trong khuôn viên bệnh viện.
// ========================================

type Building struct {
	BuildingID   uint32     `gorm:"primaryKey;autoIncrement;column:building_id"`
	BuildingCode string     `gorm:"uniqueIndex;not null;size:20;column:building_code"` // VD: A, B, C
	BuildingName string     `gorm:"not null;size:200;column:building_name"`
	TotalFloors  int8       `gorm:"not null;column:total_floors"`
	IsActive     bool       `gorm:"not null;default:true;column:is_active"`
	CreatedAt    time.Time  `gorm:"not null;autoCreateTime;column:created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`

	// Has-many: 1 toa nha co nhieu tang
	Floors []Floor `gorm:"foreignKey:BuildingID"`
}

func (Building) TableName() string { return "buildings" }

// ========================================
// Bảng: floors [T09]
// Tang cua toa nha. Moi tang la 1 layer ban do rieng biet.
// Chua URL anh nen + he so quy doi pixel → met.
//
// Cong thuc quy doi pixel → met:
//   meter_x = pixel_x × (RealWidthM  / ImageWidthPx)
//   meter_y = pixel_y × (RealHeightM / ImageHeightPx)
// ========================================

type Floor struct {
	FloorID       uint32  `gorm:"primaryKey;autoIncrement;column:floor_id"`
	BuildingID    uint32  `gorm:"not null;index;column:building_id"`
	FloorNumber   int8    `gorm:"not null;column:floor_number"`  // 0 = tret, 1 = lau 1...
	FloorName     string  `gorm:"not null;size:100;column:floor_name"` // VD: "Tang 1"
	DisplayOrder  int     `gorm:"not null;default:0;column:display_order"` // Thu tu bo loc tren App
	MapImageURL   *string `gorm:"size:500;column:map_image_url"`  // URL anh nen so do tang
	ImageWidthPx  int     `gorm:"not null;default:0;column:image_width_px"`
	ImageHeightPx int     `gorm:"not null;default:0;column:image_height_px"`
	RealWidthM    float32 `gorm:"not null;default:0;column:real_width_m"`  // Chieu rong thuc te (met)
	RealHeightM   float32 `gorm:"not null;default:0;column:real_height_m"` // Chieu cao thuc te (met)
	IsActive      bool    `gorm:"not null;default:true;column:is_active"`

	// Belongs-to: toa nha chua tang nay
	Building *Building `gorm:"foreignKey:BuildingID"`
}

func (Floor) TableName() string { return "floors" }

// ========================================
// Bảng: map_nodes [T10]
// Diem nut tren do thi ban do: phong kham, hanh lang,
// thang may, cau thang, diem giao nhau.
//
// Moi node mo ta bang polygon toa do pixel (ho tro moi hinh da giac).
// - polygon_coords: "x1,y1,x2,y2,x3,y3,..." (cac dinh polygon)
// - center_x/y: tam hinh hoc, dung de hien thi nhan va render ban do
//
// Phan loai node:
// - Node co polygon (isPoly=true): phong, khu vuc... => center la tam polygon
// - Node khong co polygon (isPoly=false): cua, junction, thang may...
//   => la access node (diem truy cap), dung de tinh khoang cach edge
//
// parent_node_id: Lien ket cua ra vao (entrance) voi phong cha.
//   VD: Cua P101 (entrance) -> Phong 101 (room).
//   Junction va corridor khong can parent.
// ========================================

type NodeType string

const (
	NodeTypeRoom       NodeType = "room"
	NodeTypeCorridor   NodeType = "corridor"
	NodeTypeElevator   NodeType = "elevator"
	NodeTypeEscalator  NodeType = "escalator"
	NodeTypeStairs     NodeType = "stairs"
	NodeTypeWC         NodeType = "wc"
	NodeTypeEntrance   NodeType = "entrance"
	NodeTypePharmacy   NodeType = "pharmacy"
	NodeTypeInfo       NodeType = "info"
	NodeTypeRestricted NodeType = "restricted"
	NodeTypeOther      NodeType = "other"
)

type MapNode struct {
	NodeID               uint32   `gorm:"primaryKey;autoIncrement;column:node_id"`
	FloorID              uint32   `gorm:"not null;index;column:floor_id"`
	WardID               *uint32  `gorm:"column:ward_id"`        // NULL = khong thuoc khoa nao
	ParentNodeID         *uint32  `gorm:"column:parent_node_id"` // Cua -> Phong cha (NULL = khong co)
	NodeCode             string   `gorm:"uniqueIndex;not null;size:30;column:node_code"`  // VD: P04, N_ELV1
	NodeName             string   `gorm:"not null;size:200;column:node_name"`             // FULLTEXT khi dung PostgreSQL
	NodeType             NodeType `gorm:"not null;column:node_type"`
	PolygonCoords        string   `gorm:"not null;type:text;column:polygon_coords"` // "x1,y1,x2,y2,..." (rong = access node)
	CenterX              float32  `gorm:"not null;default:0;column:center_x"` // Tam polygon hoac toa do access
	CenterY              float32  `gorm:"not null;default:0;column:center_y"`
	IsLandmark           bool     `gorm:"not null;default:false;column:is_landmark"`           // Diem moc noi bat
	IsAccessible         bool     `gorm:"not null;default:true;column:is_accessible"`          // Co the di qua
	WheelchairAccessible bool     `gorm:"not null;default:false;column:wheelchair_accessible"` // Ho tro xe lan
	IsActive             bool     `gorm:"not null;default:true;column:is_active"`

	// Belongs-to
	Floor      *Floor   `gorm:"foreignKey:FloorID"`
	Ward       *Ward    `gorm:"foreignKey:WardID"`
	ParentNode *MapNode `gorm:"foreignKey:ParentNodeID"` // Self-reference: cua -> phong
}

func (MapNode) TableName() string { return "map_nodes" }

// ========================================
// Bảng: map_edges [T11]
// Hanh lang / doan duong noi 2 node.
// Co the cung tang (is_cross_floor=false) hoac lien tang (thang may/cau thang).
//
// distance_m duoc tinh tu toa do center_x/y cua 2 node:
//   dx = (center_x_to - center_x_from) × (real_width_m  / image_width_px)
//   dy = (center_y_to - center_y_from) × (real_height_m / image_height_px)
//   distance_m = sqrt(dx² + dy²)
// Gia tri nay duoc tinh khi admin them edge va luu vao DB.
//
// weight: trong so Dijkstra. Admin co the tang weight de "huong" nguoi
// dung tranh khu vuc dong nguoi.
// ========================================

type MapEdge struct {
	EdgeID               uint32  `gorm:"primaryKey;autoIncrement;column:edge_id"`
	FloorID              uint32  `gorm:"not null;index;column:floor_id"` // Tang xuat phat (neu cross_floor)
	FromNodeID           uint32  `gorm:"not null;column:from_node_id"`
	ToNodeID             uint32  `gorm:"not null;column:to_node_id"`
	PolygonCoords        *string `gorm:"type:text;column:polygon_coords"` // NULL = duong thang noi 2 access point
	DistanceM            float32 `gorm:"not null;default:0;column:distance_m"` // Khoang cach thuc te (met)
	Weight               float32 `gorm:"not null;default:1.0;column:weight"`   // Trong so Dijkstra
	IsBidirectional      bool    `gorm:"not null;default:true;column:is_bidirectional"`   // 2 chieu
	IsCrossFloor         bool    `gorm:"not null;default:false;column:is_cross_floor"`    // Lien tang
	WheelchairAccessible bool    `gorm:"not null;default:false;column:wheelchair_accessible"`
	IsActive             bool    `gorm:"not null;default:true;column:is_active"`

	// Belongs-to
	FromNode *MapNode `gorm:"foreignKey:FromNodeID"`
	ToNode   *MapNode `gorm:"foreignKey:ToNodeID"`
}

func (MapEdge) TableName() string { return "map_edges" }
