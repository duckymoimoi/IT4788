package schema

import "time"

// ========================================
// POI TYPE  - loại điểm quan tâm trên grid
// ========================================

type POIType string

const (
	POITypeRoom      POIType = "room"
	POITypeCorridor  POIType = "corridor"
	POITypeElevator  POIType = "elevator"
	POITypeStairs    POIType = "stairs"
	POITypeWC        POIType = "wc"
	POITypeEntrance  POIType = "entrance"
	POITypePharmacy  POIType = "pharmacy"
	POITypeCanteen   POIType = "canteen"
	POITypeParking   POIType = "parking"
	POITypeWifi      POIType = "wifi"
	POITypeInfo      POIType = "info"
	POITypeOther     POIType = "other"
)

// GridMap ban do grid 2D (MovingAI format).
// Bang: grid_maps [T08]
type GridMap struct {
	MapID       uint32     `gorm:"primaryKey;autoIncrement;column:map_id" json:"map_id"`
	MapName     string     `gorm:"not null;size:200;column:map_name" json:"map_name"`
	MapFilePath string     `gorm:"not null;size:500;column:map_file_path" json:"map_file_path"`
	Rows        int        `gorm:"not null;column:rows" json:"rows"`
	Cols        int        `gorm:"not null;column:cols" json:"cols"`
	GridData    string     `gorm:"not null;type:text;column:grid_data" json:"grid_data"`
	MapImageURL *string    `gorm:"size:500;column:map_image_url" json:"map_image_url,omitempty"`
	IsActive    bool       `gorm:"not null;default:true;column:is_active" json:"is_active"`
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`

	// Has-many
	POIs []GridPOI `gorm:"foreignKey:MapID" json:"pois,omitempty"`
}

func (GridMap) TableName() string { return "grid_maps" }

// GridPOI diem quan tam tren grid.
// Bang: grid_pois [T09]
type GridPOI struct {
	POIID                uint32  `gorm:"primaryKey;autoIncrement;column:poi_id" json:"poi_id"`
	MapID                uint32  `gorm:"not null;index;column:map_id" json:"map_id"`
	WardID               *uint32 `gorm:"column:ward_id" json:"ward_id,omitempty"`
	POICode              string  `gorm:"uniqueIndex;not null;size:30;column:poi_code" json:"poi_code"`
	POIName              string  `gorm:"not null;size:200;column:poi_name" json:"poi_name"`
	POIType              POIType `gorm:"not null;column:poi_type" json:"poi_type"`
	GridRow              int     `gorm:"not null;column:grid_row" json:"grid_row"`
	GridCol              int     `gorm:"not null;column:grid_col" json:"grid_col"`
	GridLocation         int     `gorm:"not null;index;column:grid_location" json:"grid_location"`
	IsLandmark           bool    `gorm:"not null;default:false;column:is_landmark" json:"is_landmark"`
	IsAccessible         bool    `gorm:"not null;default:true;column:is_accessible" json:"is_accessible"`
	WheelchairAccessible bool    `gorm:"not null;default:false;column:wheelchair_accessible" json:"wheelchair_accessible"`
	CustomWeight         float32 `gorm:"not null;default:1.0;column:custom_weight" json:"custom_weight"`
	Capacity             *int    `gorm:"column:capacity" json:"capacity,omitempty"`
	Details              *string `gorm:"type:text;column:details" json:"details,omitempty"`
	OpenHours            *string `gorm:"size:100;column:open_hours" json:"open_hours,omitempty"`
	IsActive             bool    `gorm:"not null;default:true;column:is_active" json:"is_active"`

	// Belongs-to
	Map  *GridMap `gorm:"foreignKey:MapID" json:"-"`
	Ward *Ward    `gorm:"foreignKey:WardID" json:"-"`
}

func (GridPOI) TableName() string { return "grid_pois" }
