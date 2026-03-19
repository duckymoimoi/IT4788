package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"hospital/schema"
)

// ================================================================
// SeedMap doc file JSON tu Map Editor v2 va seed vao database.
// File JSON phai co truong "floors" voi cau truc tuong thich.
//
// Cach dung:
//   SeedMap("map_data.json")
//
// Ham nay idempotent: kiem tra building_code + floor_number truoc khi tao.
// ================================================================

// --- JSON structures ---

type mapDataJSON struct {
	Floors []floorJSON `json:"floors"`
}

type floorJSON struct {
	Meta      metaJSON      `json:"meta"`
	Nodes     []nodeJSON    `json:"nodes"`
	Edges     []edgeJSON    `json:"edges"`
	Corridors []corridorJSON `json:"corridors"` // chi dung cho editor, khong luu DB
}

type metaJSON struct {
	BuildingName  string  `json:"building_name"`
	BuildingCode  string  `json:"building_code"`
	FloorNumber   int     `json:"floor_number"`
	FloorName     string  `json:"floor_name"`
	ImageWidthPx  int     `json:"image_width_px"`
	ImageHeightPx int     `json:"image_height_px"`
	RealWidthM    float32 `json:"real_width_m"`
	RealHeightM   float32 `json:"real_height_m"`
}

type nodeJSON struct {
	NodeCode             string  `json:"node_code"`
	NodeName             string  `json:"node_name"`
	NodeType             string  `json:"node_type"`
	PolygonCoords        string  `json:"polygon_coords"`
	CenterX              float32 `json:"center_x"`
	CenterY              float32 `json:"center_y"`
	AccessX              float32 `json:"access_x"`
	AccessY              float32 `json:"access_y"`
	IsLandmark           bool    `json:"is_landmark"`
	IsAccessible         bool    `json:"is_accessible"`
	WheelchairAccessible bool    `json:"wheelchair_accessible"`
	IsActive             bool    `json:"is_active"`
}

type edgeJSON struct {
	FromCode             string   `json:"from_code"`
	ToCode               string   `json:"to_code"`
	PolygonCoords        *string  `json:"polygon_coords"`
	DistanceM            float32  `json:"distance_m"`
	Weight               float32  `json:"weight"`
	IsBidirectional      bool     `json:"is_bidirectional"`
	IsCrossFloor         bool     `json:"is_cross_floor"`
	WheelchairAccessible bool     `json:"wheelchair_accessible"`
	IsActive             bool     `json:"is_active"`
}

type corridorJSON struct {
	AX     int     `json:"ax"`
	AY     int     `json:"ay"`
	BX     int     `json:"bx"`
	BY     int     `json:"by"`
	Weight float32 `json:"weight"`
}

// SeedMap doc file JSON va tao building, floors, nodes, edges trong DB.
func SeedMap(jsonPath string) error {
	// --- Doc file ---
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("khong doc duoc file %s: %w", jsonPath, err)
	}

	var mapData mapDataJSON
	if err := json.Unmarshal(data, &mapData); err != nil {
		return fmt.Errorf("loi parse JSON: %w", err)
	}

	if len(mapData.Floors) == 0 {
		log.Println("SeedMap: file JSON khong co tang nao, bo qua")
		return nil
	}

	log.Printf("SeedMap: doc duoc %d tang tu %s", len(mapData.Floors), jsonPath)

	// --- Nhom floors theo building_code ---
	type buildingInfo struct {
		name   string
		code   string
		floors []floorJSON
	}
	buildingMap := map[string]*buildingInfo{}
	for _, fl := range mapData.Floors {
		bc := fl.Meta.BuildingCode
		if _, ok := buildingMap[bc]; !ok {
			buildingMap[bc] = &buildingInfo{
				name:   fl.Meta.BuildingName,
				code:   bc,
				floors: []floorJSON{},
			}
		}
		buildingMap[bc].floors = append(buildingMap[bc].floors, fl)
	}

	// --- Tao/tim building ---
	for _, bi := range buildingMap {
		var building schema.Building
		result := DB.Where("building_code = ?", bi.code).First(&building)
		if result.Error != nil {
			// Chua co -> tao moi
			maxFloor := int8(0)
			for _, fl := range bi.floors {
				if int8(fl.Meta.FloorNumber) > maxFloor {
					maxFloor = int8(fl.Meta.FloorNumber)
				}
			}
			building = schema.Building{
				BuildingCode: bi.code,
				BuildingName: bi.name,
				TotalFloors:  maxFloor,
				IsActive:     true,
			}
			if err := DB.Create(&building).Error; err != nil {
				return fmt.Errorf("loi tao building %s: %w", bi.code, err)
			}
			log.Printf("  Tao building: %s (%s), ID=%d", bi.name, bi.code, building.BuildingID)
		} else {
			log.Printf("  Building %s da ton tai, ID=%d", bi.code, building.BuildingID)
		}

		// --- Tao floors ---
		for _, fl := range bi.floors {
			var floor schema.Floor
			result := DB.Where("building_id = ? AND floor_number = ?", building.BuildingID, fl.Meta.FloorNumber).First(&floor)
			if result.Error != nil {
				// Chua co -> tao moi
				floor = schema.Floor{
					BuildingID:    building.BuildingID,
					FloorNumber:   int8(fl.Meta.FloorNumber),
					FloorName:     fl.Meta.FloorName,
					DisplayOrder:  fl.Meta.FloorNumber,
					ImageWidthPx:  fl.Meta.ImageWidthPx,
					ImageHeightPx: fl.Meta.ImageHeightPx,
					RealWidthM:    fl.Meta.RealWidthM,
					RealHeightM:   fl.Meta.RealHeightM,
					IsActive:      true,
				}
				if err := DB.Create(&floor).Error; err != nil {
					return fmt.Errorf("loi tao floor %s: %w", fl.Meta.FloorName, err)
				}
				log.Printf("    Tao floor: %s (number=%d), ID=%d", fl.Meta.FloorName, fl.Meta.FloorNumber, floor.FloorID)
			} else {
				log.Printf("    Floor %s da ton tai, ID=%d", fl.Meta.FloorName, floor.FloorID)
			}

			// Bo qua neu khong co node
			if len(fl.Nodes) == 0 {
				log.Printf("    Floor %s khong co node, bo qua", fl.Meta.FloorName)
				continue
			}

			// --- Kiem tra da seed node chua ---
			var existingNodeCount int64
			DB.Model(&schema.MapNode{}).Where("floor_id = ?", floor.FloorID).Count(&existingNodeCount)
			if existingNodeCount > 0 {
				log.Printf("    Floor %s da co %d nodes, bo qua seed nodes/edges", fl.Meta.FloorName, existingNodeCount)
				continue
			}

			// --- Tao nodes ---
			codeToNodeID := map[string]uint32{}

			for _, n := range fl.Nodes {
				accessX := n.AccessX
				accessY := n.AccessY
				node := schema.MapNode{
					FloorID:              floor.FloorID,
					NodeCode:             n.NodeCode,
					NodeName:             n.NodeName,
					NodeType:             schema.NodeType(n.NodeType),
					PolygonCoords:        n.PolygonCoords,
					CenterX:             n.CenterX,
					CenterY:             n.CenterY,
					AccessX:             &accessX,
					AccessY:             &accessY,
					IsLandmark:          n.IsLandmark,
					IsAccessible:        n.IsAccessible,
					WheelchairAccessible: n.WheelchairAccessible,
					IsActive:            n.IsActive,
				}
				if err := DB.Create(&node).Error; err != nil {
					return fmt.Errorf("loi tao node %s: %w", n.NodeCode, err)
				}
				codeToNodeID[n.NodeCode] = node.NodeID
			}
			log.Printf("    Tao %d nodes cho floor %s", len(fl.Nodes), fl.Meta.FloorName)

			// --- Tao edges ---
			edgeCount := 0
			for _, e := range fl.Edges {
				fromID, okFrom := codeToNodeID[e.FromCode]
				toID, okTo := codeToNodeID[e.ToCode]
				if !okFrom || !okTo {
					log.Printf("    CANH BAO: edge %s -> %s khong tim thay node, bo qua", e.FromCode, e.ToCode)
					continue
				}

				var polyCoords *string
				if e.PolygonCoords != nil && *e.PolygonCoords != "" {
					polyCoords = e.PolygonCoords
				}

				weight := e.Weight
				if weight <= 0 {
					weight = 1.0
				}

				edge := schema.MapEdge{
					FloorID:              floor.FloorID,
					FromNodeID:           fromID,
					ToNodeID:             toID,
					PolygonCoords:        polyCoords,
					DistanceM:            e.DistanceM,
					Weight:               weight,
					IsBidirectional:      e.IsBidirectional,
					IsCrossFloor:        e.IsCrossFloor,
					WheelchairAccessible: e.WheelchairAccessible,
					IsActive:            e.IsActive,
				}
				if err := DB.Create(&edge).Error; err != nil {
					return fmt.Errorf("loi tao edge %s->%s: %w", e.FromCode, e.ToCode, err)
				}
				edgeCount++
			}
			log.Printf("    Tao %d edges cho floor %s", edgeCount, fl.Meta.FloorName)
		}
	}

	log.Println("SeedMap: hoan thanh")
	return nil
}
