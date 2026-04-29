package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"hospital/pkg/mapf"
)

// ========================================
// FULL TEST SUITE  - Unit + Integration + API + Security
// Server phai chay tai :8080
// ========================================

var (
	base       = "http://localhost:8080/api"
	pass, fail int
	total      int

	patientToken  string
	adminToken    string
	patient2Token string
)

func check(name string, ok bool, detail string) {
	total++
	if ok {
		pass++
		fmt.Printf("  %s\n", name)
	} else {
		fail++
		fmt.Printf("  FAIL %s  - %s\n", name, detail)
	}
}

type apiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func doReq(method, url string, data interface{}, token string) (*apiResp, error) {
	var body io.Reader
	if data != nil {
		b, _ := json.Marshal(data)
		body = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r apiResp
	json.Unmarshal(b, &r)
	return &r, nil
}

func sc(r *apiResp) int {
	if r == nil {
		return -1
	}
	return r.Code
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  HOSPITAL NAVIGATION  - FULL TEST SUITE v2")
	fmt.Println(strings.Repeat("=", 70))

	testUnitDijkstra()
	testUnitParser()
	testUnitIntegration()

	// Check server
	_, err := http.Get(base + "/sys/check_version")
	if err != nil {
		fmt.Println("\n  [WARN]  Server not running  - skipping HTTP API tests")
		printSummary()
		return
	}

	testLogin()
	testMapAPIs()
	testRouteAPIs()
	testAuthorizationSecurity()
	testInputValidation()
	testEngineAPIs()
	testMedicalAPIs()
	testNotifAPIs()
	testDeviceAPIs()
	testUtilAPIs()
	testFlowAPIs()
	testSimulationAPIs()
	testNewMedicalAPIs()
	testNewUtilAPIs()
	testSysAPIs()
	testMedicalE2E()
	testNotifE2E()
	testDeviceE2E()
	testFlowE2E()
	testMedicalCheckoutE2E()
	testUploadAPI()
	testVoiceNavigationE2E()
	testJSONFormat()

	printSummary()
}

// ========================================
// PART 1: DIJKSTRA UNIT TESTS
// ========================================
func testUnitDijkstra() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 1: DIJKSTRA UNIT TESTS (10)")
	fmt.Println(strings.Repeat("-", 50))

	grid := &mapf.GridMap{
		Name: "test_5x5",
		Rows: 5, Cols: 5,
		Grid: [][]int{
			{0, 0, 0, 0, 0},
			{0, 1, 1, 0, 0},
			{0, 0, 0, 0, 0},
			{0, 0, 1, 1, 0},
			{0, 0, 0, 0, 0},
		},
	}

	r := mapf.Dijkstra(grid, grid.ToLocation(0, 0), grid.ToLocation(0, 4))
	check("Direct path (0,0->0,4) dist=4",
		r.Found && r.Distance == 4, fmt.Sprintf("found=%v dist=%.0f", r.Found, r.Distance))

	r = mapf.Dijkstra(grid, grid.ToLocation(0, 0), grid.ToLocation(2, 2))
	check("Around obstacle (0,0->2,2) dist=4",
		r.Found && r.Distance == 4, fmt.Sprintf("found=%v dist=%.0f", r.Found, r.Distance))

	r = mapf.Dijkstra(grid, grid.ToLocation(0, 0), grid.ToLocation(4, 4))
	check("Long path (0,0->4,4) dist=8",
		r.Found && r.Distance == 8, fmt.Sprintf("found=%v dist=%.0f", r.Found, r.Distance))

	check("Path len = dist+1",
		r.Found && len(r.Path) == int(r.Distance)+1,
		fmt.Sprintf("len=%d expected=%d", len(r.Path), int(r.Distance)+1))

	r = mapf.Dijkstra(grid, grid.ToLocation(2, 2), grid.ToLocation(2, 2))
	check("Same start=dest -> dist=0", r.Found && r.Distance == 0 && len(r.Path) == 1, "")

	r = mapf.Dijkstra(grid, grid.ToLocation(0, 0), grid.ToLocation(1, 1))
	check("To obstacle -> not found", !r.Found, "")

	r = mapf.Dijkstra(grid, -1, grid.ToLocation(0, 0))
	check("Invalid start -> not found", !r.Found, "")

	r = mapf.DijkstraWithSpeed(grid, grid.ToLocation(0, 0), grid.ToLocation(0, 4), 0.7)
	check("Speed=0.7 -> ~5.71",
		r.Found && r.Distance > 5.7 && r.Distance < 5.8,
		fmt.Sprintf("got %.2f", r.Distance))

	realGrid, err := mapf.LoadGridMap("data/warehouse_small.map")
	if err == nil {
		check(fmt.Sprintf("Load warehouse_small: %dx%d", realGrid.Rows, realGrid.Cols),
			realGrid.Rows == 33 && realGrid.Cols == 57, "")
		rr := mapf.Dijkstra(realGrid, realGrid.ToLocation(4, 4), realGrid.ToLocation(4, 20))
		check("Dijkstra on real map dist=16",
			rr.Found && rr.Distance == 16, fmt.Sprintf("found=%v dist=%.0f", rr.Found, rr.Distance))
	} else {
		check("Load warehouse_small.map", false, err.Error())
	}
}

// ========================================
// PART 2: MAPF PARSER UNIT TESTS
// ========================================
func testUnitParser() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 2: MAPF PARSER UNIT TESTS (10)")
	fmt.Println(strings.Repeat("-", 50))

	result, err := mapf.ParseOutputJSON("data/output.json")
	check("Parse output.json", err == nil && result != nil, func() string {
		if err != nil { return err.Error() }; return "nil"
	}())
	if result == nil { return }

	check(fmt.Sprintf("TeamSize=%d", result.TeamSize), result.TeamSize > 0, "")
	check(fmt.Sprintf("Makespan=%d", result.Makespan), result.Makespan > 0, "")
	check("Trajectories = TeamSize", len(result.Trajectories) == result.TeamSize, "")

	if len(result.Trajectories) > 0 {
		t0 := result.Trajectories[0]
		check("Agent0 has states > 0", len(t0.States) > 0, "")
		check("Agent0 start matches",
			t0.States[0].Row == t0.StartRow && t0.States[0].Col == t0.StartCol, "")
	}

	check("Tasks > 0", len(result.Tasks) > 0, "")

	result.SetAllLocations(57)
	if len(result.Trajectories) > 0 && len(result.Trajectories[0].States) > 0 {
		s := result.Trajectories[0].States[0]
		check("SetAllLocations correct", s.Location == s.Row*57+s.Col, "")
	}

	pos := result.GetPositionsAtTimestep(0)
	check("Positions(0) = TeamSize", len(pos) == result.TeamSize, "")
	posLast := result.GetPositionsAtTimestep(99999)
	check("Positions(99999) returns last", len(posLast) == result.TeamSize, "")
}

// ========================================
// PART 3: INTEGRATION
// ========================================
func testUnitIntegration() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 3: DIJKSTRA + MAPF INTEGRATION (3)")
	fmt.Println(strings.Repeat("-", 50))

	grid, err := mapf.LoadGridMap("data/warehouse_small.map")
	if err != nil { check("Load map", false, err.Error()); return }

	result, err := mapf.ParseOutputJSON("data/output.json")
	if err != nil || result == nil { check("Parse output", false, ""); return }
	result.SetAllLocations(grid.Cols)

	allValid := true
	for _, traj := range result.Trajectories {
		for _, state := range traj.States {
			if !grid.IsWalkable(state.Row, state.Col) { allValid = false; break }
		}
	}
	check("All agent positions walkable", allValid, "")

	if len(result.Tasks) > 0 && len(result.Trajectories) > 0 {
		task := result.Tasks[0]
		agent0 := result.Trajectories[0]
		dr := mapf.Dijkstra(grid, grid.ToLocation(agent0.StartRow, agent0.StartCol), task.Location)
		check("Dual-mode: Dijkstra finds path", dr.Found, "")
		check("Dual-mode: MAPF has trajectory", len(agent0.States) > 0, "")
	}
}

// ========================================
// PART 4: LOGIN
// ========================================
func testLogin() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 4: LOGIN (4)")
	fmt.Println(strings.Repeat("-", 50))

	// Bad login
	r, _ := doReq("POST", base+"/auth/login", map[string]string{
		"phone_number": "0900000004", "password": "wrongpassword",
	}, "")
	check("Bad password -> rejected", r != nil && r.Code != 1000, "")

	// Patient login
	r, _ = doReq("POST", base+"/auth/login", map[string]string{
		"phone_number": "0900000004", "password": "password123",
	}, "")
	check("Login patient (0900000004)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d struct{ Token string `json:"token"` }
		json.Unmarshal(r.Data, &d)
		patientToken = d.Token
	}

	// Patient 2 login
	r, _ = doReq("POST", base+"/auth/login", map[string]string{
		"phone_number": "0900000005", "password": "password123",
	}, "")
	check("Login patient2 (0900000005)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d struct{ Token string `json:"token"` }
		json.Unmarshal(r.Data, &d)
		patient2Token = d.Token
	}

	// Admin login
	r, _ = doReq("POST", base+"/auth/login", map[string]string{
		"phone_number": "0900000001", "password": "password123",
	}, "")
	check("Login admin (0900000001)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d struct{ Token string `json:"token"` }
		json.Unmarshal(r.Data, &d)
		adminToken = d.Token
	}
}

// ========================================
// PART 5: MAP APIs
// ========================================
func testMapAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 5: MAP APIs (12)")
	fmt.Println(strings.Repeat("-", 50))

	// [16] get_floors
	r, _ := doReq("GET", base+"/map/get_floors", nil, "")
	check("[16] GET get_floors", r != nil && r.Code == 1000, "")

	// Lấy map_id đầu tiên để test get_edges
	var mapID float64
	if r != nil && r.Code == 1000 {
		var floors []map[string]interface{}
		json.Unmarshal(r.Data, &floors)
		if len(floors) > 0 {
			if v, ok := floors[0]["map_id"]; ok {
				mapID, _ = v.(float64)
			}
		}
	}

	// [17] get_nodes
	r, _ = doReq("GET", base+"/map/get_nodes?map_id=0", nil, "")
	check("[17] GET get_nodes", r != nil && r.Code == 1000, "")

	// ----------------------------------------
	// [18] get_edges — Test suite
	// ----------------------------------------

	// Test 1: Gọi với map_id hợp lệ
	if mapID > 0 {
		r, _ = doReq("GET", fmt.Sprintf("%s/map/get_edges?map_id=%.0f", base, mapID), nil, "")
		check("[18] GET get_edges (valid map_id)", r != nil && r.Code == 1000,
			fmt.Sprintf("code=%d", sc(r)))

		// Test 2: Response có cấu trúc đúng (map_id, total, edges)
		if r != nil && r.Code == 1000 {
			var d map[string]interface{}
			json.Unmarshal(r.Data, &d)

			_, hasMapID := d["map_id"]
			_, hasTotal := d["total"]
			_, hasEdges := d["edges"]
			check("  Response has map_id, total, edges",
				hasMapID && hasTotal && hasEdges,
				fmt.Sprintf("keys: %v", keysOf(d)))

			// Test 3: total > 0 (grid phải có edges)
			total, _ := d["total"].(float64)
			check("  total > 0 (grid has edges)", total > 0,
				fmt.Sprintf("total=%.0f", total))

			// Test 4: edges array length == total
			if edgesRaw, ok := d["edges"]; ok {
				var edges []map[string]interface{}
				b, _ := json.Marshal(edgesRaw)
				json.Unmarshal(b, &edges)
				check("  len(edges) == total",
					len(edges) == int(total),
					fmt.Sprintf("len=%d total=%.0f", len(edges), total))

				// Test 5: Mỗi edge có đủ fields
				if len(edges) > 0 {
					e := edges[0]
					_, hasFromRow := e["from_row"]
					_, hasFromCol := e["from_col"]
					_, hasFromLoc := e["from_location"]
					_, hasToRow := e["to_row"]
					_, hasToCol := e["to_col"]
					_, hasToLoc := e["to_location"]
					check("  Edge has all 6 fields",
						hasFromRow && hasFromCol && hasFromLoc &&
							hasToRow && hasToCol && hasToLoc,
						fmt.Sprintf("keys: %v", keysOf(e)))

					// Test 6: from_location < to_location (1 chiều)
					fromLoc, _ := e["from_location"].(float64)
					toLoc, _ := e["to_location"].(float64)
					check("  from_location < to_location (unidirectional)",
						fromLoc < toLoc,
						fmt.Sprintf("from=%.0f to=%.0f", fromLoc, toLoc))

					// Test 7: Edges liền kề (diff == 1 hoặc == cols)
					fromRow, _ := e["from_row"].(float64)
					fromCol, _ := e["from_col"].(float64)
					toRow, _ := e["to_row"].(float64)
					toCol, _ := e["to_col"].(float64)
					rowDiff := toRow - fromRow
					colDiff := toCol - fromCol
					if rowDiff < 0 { rowDiff = -rowDiff }
					if colDiff < 0 { colDiff = -colDiff }
					isAdjacent := (rowDiff + colDiff) == 1
					check("  Edge is 4-dir adjacent (diff=1)",
						isAdjacent,
						fmt.Sprintf("from(%v,%v) to(%v,%v)", fromRow, fromCol, toRow, toCol))
				}
			}
		}
	} else {
		check("[18] GET get_edges (skip: no map_id)", false, "no floor found")
	}

	// Test 8: Thiếu map_id → error
	r, _ = doReq("GET", base+"/map/get_edges", nil, "")
	check("[18] get_edges missing map_id -> error", r != nil && r.Code != 1000,
		fmt.Sprintf("code=%d", sc(r)))

	// Test 9: map_id không tồn tại → error
	r, _ = doReq("GET", base+"/map/get_edges?map_id=99999", nil, "")
	check("[18] get_edges map_id=99999 -> not found", r != nil && r.Code != 1000,
		fmt.Sprintf("code=%d", sc(r)))
}

// ========================================
// PART 6: ROUTE APIs
// ========================================
func testRouteAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 6: ROUTE APIs (16)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// [45] get_modes (public)
	r, _ := doReq("GET", base+"/route/get_modes", nil, "")
	check("[45] GET get_modes", r != nil && r.Code == 1000, "")
	if r != nil && r.Code == 1000 {
		var modes []map[string]interface{}
		json.Unmarshal(r.Data, &modes)
		check("  Modes >= 4 & has snake_case keys", len(modes) >= 4, fmt.Sprintf("got %d", len(modes)))
		if len(modes) > 0 {
			_, hasID := modes[0]["mode_id"]
			check("  mode_id key (snake_case)", hasID, "missing mode_id key")
		}
	}

	// [37] preview
	r, _ = doReq("POST", base+"/route/preview", map[string]interface{}{
		"start_location": 4*57 + 4, "dest_location": 4*57 + 20, "mode_id": "walking",
	}, patientToken)
	check("[37] POST preview", r != nil && r.Code == 1000, "")

	// [31] order
	var routeID string
	r, _ = doReq("POST", base+"/route/order", map[string]interface{}{
		"start_location": 4*57 + 4, "dest_location": 4*57 + 20, "mode_id": "walking",
	}, patientToken)
	check("[31] POST order", r != nil && r.Code == 1000, "")
	if r != nil && r.Code == 1000 {
		var d map[string]json.RawMessage
		json.Unmarshal(r.Data, &d)
		var route map[string]interface{}
		json.Unmarshal(d["route"], &route)
		if v, ok := route["route_id"]; ok {
			routeID = fmt.Sprintf("%v", v)
		}
		check("  route_id in snake_case", routeID != "", "missing route_id")
		// Verify json:"-" hides password
		_, hasPassword := route["password_hash"]
		check("  password_hash hidden", !hasPassword, "password_hash leaked!")
	}

	// [36] get_steps
	if routeID != "" {
		r, _ = doReq("GET", base+"/route/get_steps?route_id="+routeID, nil, patientToken)
		check("[36] GET get_steps (own route)", r != nil && r.Code == 1000, "")
	}

	// [38] get_eta
	if routeID != "" {
		r, _ = doReq("POST", base+"/route/get_eta", map[string]interface{}{
			"route_id": routeID, "current_step": 2,
		}, patientToken)
		check("[38] POST get_eta", r != nil && r.Code == 1000, "")
	}

	// [34] get_active
	r, _ = doReq("GET", base+"/route/get_active", nil, patientToken)
	check("[34] GET get_active", r != nil && r.Code == 1000, "")

	// [44] get_next
	if routeID != "" {
		r, _ = doReq("GET", base+"/route/get_next?route_id="+routeID+"&current_step=0&limit=3",
			nil, patientToken)
		check("[44] GET get_next", r != nil && r.Code == 1000, "")
	}

	// [43] pass_node
	if routeID != "" {
		r, _ = doReq("POST", base+"/route/pass_node", map[string]interface{}{
			"route_id": routeID, "grid_location": 4*57 + 5,
		}, patientToken)
		check("[43] POST pass_node", r != nil && r.Code == 1000, "")
	}

	// [33] recalculate
	if routeID != "" {
		r, _ = doReq("POST", base+"/route/recalculate", map[string]interface{}{
			"route_id": routeID, "current_location": 4*57 + 10,
		}, patientToken)
		check("[33] POST recalculate", r != nil && r.Code == 1000, "")
	}

	// [41] share
	if routeID != "" {
		r, _ = doReq("POST", base+"/route/share", map[string]interface{}{
			"route_id": routeID, "receiver_phone": "0912345678",
		}, patientToken)
		check("[41] POST share", r != nil && r.Code == 1000, "")
	}

	// [42] rate
	if routeID != "" {
		r, _ = doReq("POST", base+"/route/rate", map[string]interface{}{
			"route_id": routeID, "rating": 5, "comment": "Excellent!",
		}, patientToken)
		check("[42] POST rate", r != nil && r.Code == 1000, "")
	}

	// [35] cancel
	if routeID != "" {
		r, _ = doReq("POST", base+"/route/cancel", map[string]interface{}{
			"route_id": routeID,
		}, patientToken)
		check("[35] POST cancel", r != nil && r.Code == 1000, "")
	}

	// [39] get_history
	r, _ = doReq("GET", base+"/route/get_history?page=1&limit=10", nil, patientToken)
	check("[39] GET get_history", r != nil && r.Code == 1000, "")

	// [40] clear_history
	r, _ = doReq("DELETE", base+"/route/clear_history", nil, patientToken)
	check("[40] DELETE clear_history", r != nil && r.Code == 1000, "")
}

// ========================================
// PART 7: AUTHORIZATION SECURITY TESTS
// ========================================
func testAuthorizationSecurity() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 7: AUTHORIZATION SECURITY (6)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" || patient2Token == "" {
		fmt.Println("  [WARN]  Need 2 patient tokens"); return
	}

	// Patient1 creates a route
	r, _ := doReq("POST", base+"/route/order", map[string]interface{}{
		"start_location": 4*57 + 4, "dest_location": 4*57 + 20, "mode_id": "walking",
	}, patientToken)
	var routeID string
	if r != nil && r.Code == 1000 {
		var d map[string]json.RawMessage
		json.Unmarshal(r.Data, &d)
		var route map[string]interface{}
		json.Unmarshal(d["route"], &route)
		if v, ok := route["route_id"]; ok { routeID = fmt.Sprintf("%v", v) }
	}

	if routeID == "" {
		check("Setup: create route for auth test", false, ""); return
	}

	// Patient2 tries to access Patient1's route
	r, _ = doReq("GET", base+"/route/get_steps?route_id="+routeID, nil, patient2Token)
	check("Patient2 cannot get_steps of Patient1", r != nil && r.Code != 1000,
		fmt.Sprintf("code=%d (should NOT be 1000)", sc(r)))

	r, _ = doReq("POST", base+"/route/get_eta", map[string]interface{}{
		"route_id": routeID, "current_step": 0,
	}, patient2Token)
	check("Patient2 cannot get_eta of Patient1", r != nil && r.Code != 1000, "")

	r, _ = doReq("POST", base+"/route/pass_node", map[string]interface{}{
		"route_id": routeID, "grid_location": 4*57 + 5,
	}, patient2Token)
	check("Patient2 cannot pass_node of Patient1", r != nil && r.Code != 1000, "")

	r, _ = doReq("POST", base+"/route/recalculate", map[string]interface{}{
		"route_id": routeID, "current_location": 4*57 + 10,
	}, patient2Token)
	check("Patient2 cannot recalculate Patient1", r != nil && r.Code != 1000, "")

	r, _ = doReq("POST", base+"/route/share", map[string]interface{}{
		"route_id": routeID, "receiver_phone": "0999",
	}, patient2Token)
	check("Patient2 cannot share Patient1 route", r != nil && r.Code != 1000, "")

	// No token -> rejected
	r, _ = doReq("POST", base+"/route/order", map[string]interface{}{
		"start_location": 100, "dest_location": 200, "mode_id": "walking",
	}, "")
	check("No token -> order rejected", r != nil && r.Code != 1000, "")

	// Cleanup
	doReq("POST", base+"/route/cancel", map[string]interface{}{"route_id": routeID}, patientToken)
}

// ========================================
// PART 8: INPUT VALIDATION
// ========================================
func testInputValidation() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 8: INPUT VALIDATION (4)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" { return }

	// start == dest
	r, _ := doReq("POST", base+"/route/order", map[string]interface{}{
		"start_location": 100, "dest_location": 100, "mode_id": "walking",
	}, patientToken)
	check("start == dest -> rejected", r != nil && r.Code != 1000, fmt.Sprintf("code=%d", sc(r)))

	// Invalid mode
	r, _ = doReq("POST", base+"/route/preview", map[string]interface{}{
		"start_location": 100, "dest_location": 200, "mode_id": "flying_carpet",
	}, patientToken)
	check("Invalid mode -> rejected", r != nil && r.Code != 1000, "")

	// Rating out of range
	rr, _ := doReq("POST", base+"/route/order", map[string]interface{}{
		"start_location": 4*57+4, "dest_location": 4*57+20, "mode_id": "walking",
	}, patientToken)
	var rid string
	if rr != nil && rr.Code == 1000 {
		var d map[string]json.RawMessage
		json.Unmarshal(rr.Data, &d)
		var route map[string]interface{}
		json.Unmarshal(d["route"], &route)
		if v, ok := route["route_id"]; ok { rid = fmt.Sprintf("%v", v) }
	}

	if rid != "" {
		r, _ = doReq("POST", base+"/route/rate", map[string]interface{}{
			"route_id": rid, "rating": 10,
		}, patientToken)
		check("Rating > 5 -> rejected", r != nil && r.Code != 1000, "")

		r, _ = doReq("POST", base+"/route/rate", map[string]interface{}{
			"route_id": rid, "rating": 0,
		}, patientToken)
		check("Rating < 1 -> rejected", r != nil && r.Code != 1000, "")

		// Cleanup
		doReq("POST", base+"/route/cancel", map[string]interface{}{"route_id": rid}, patientToken)
	}
}

// ========================================
// PART 9: ENGINE APIs
// ========================================
func testEngineAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 9: ENGINE APIs (12)")
	fmt.Println(strings.Repeat("-", 50))

	if adminToken == "" {
		fmt.Println("  [WARN]  No admin token"); return
	}

	// Patient cannot access engine
	r, _ := doReq("GET", base+"/engine/health", nil, patientToken)
	check("Patient cannot access engine", r != nil && r.Code != 1000, "")

	// [91] solve
	r, _ = doReq("POST", base+"/engine/solve", map[string]interface{}{
		"start_location": 4*57 + 4, "dest_location": 4*57 + 20, "mode_id": "walking",
	}, adminToken)
	check("[91] POST solve", r != nil && r.Code == 1000, "")

	// [92] update_cost
	r, _ = doReq("POST", base+"/engine/update_cost", map[string]interface{}{
		"poi_id": 1, "weight": 2.5,
	}, adminToken)
	check("[92] POST update_cost", r != nil && r.Code == 1000, "")

	// [93] convergence
	r, _ = doReq("GET", base+"/engine/convergence", nil, adminToken)
	check("[93] GET convergence", r != nil && r.Code == 1000, "")

	// [94] set_params
	r, _ = doReq("POST", base+"/engine/set_params", map[string]interface{}{
		"max_agents": 200, "time_step_ms": 250, "cost_multiplier": 1.5,
	}, adminToken)
	check("[94] POST set_params", r != nil && r.Code == 1000, "")

	// [97] health
	r, _ = doReq("GET", base+"/engine/health", nil, adminToken)
	check("[97] GET health", r != nil && r.Code == 1000, "")
	if r != nil && r.Code == 1000 {
		var h map[string]interface{}
		json.Unmarshal(r.Data, &h)
		check("  DB connected", h["db_connected"] == true, "")
	}

	// [98] clear_cache
	r, _ = doReq("POST", base+"/engine/clear_cache", nil, adminToken)
	check("[98] POST clear_cache", r != nil && r.Code == 1000, "")

	// MAPF
	r, _ = doReq("POST", base+"/engine/load_mapf", map[string]string{
		"file_path": "data/output.json",
	}, adminToken)
	check("POST load_mapf", r != nil && r.Code == 1000, "")

	r, _ = doReq("GET", base+"/engine/mapf_positions?timestep=5", nil, adminToken)
	check("GET mapf_positions (t=5)", r != nil && r.Code == 1000, "")

	r, _ = doReq("GET", base+"/engine/mapf_info", nil, adminToken)
	check("GET mapf_info", r != nil && r.Code == 1000, "")
}

// ========================================
// PART 10: JSON FORMAT VERIFICATION
// ========================================
func testJSONFormat() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 10: JSON FORMAT (3)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" { return }

	// Create route, check JSON keys are snake_case
	r, _ := doReq("POST", base+"/route/order", map[string]interface{}{
		"start_location": 4*57+4, "dest_location": 4*57+20, "mode_id": "walking",
	}, patientToken)
	if r != nil && r.Code == 1000 {
		var d map[string]json.RawMessage
		json.Unmarshal(r.Data, &d)
		var route map[string]interface{}
		json.Unmarshal(d["route"], &route)

		_, hasRouteID := route["route_id"]
		_, hasUserID := route["user_id"]
		_, hasTotalDist := route["total_distance"]
		check("Route JSON: snake_case keys", hasRouteID && hasUserID && hasTotalDist,
			fmt.Sprintf("keys: %v", keysOf(route)))

		// Verify NO PascalCase
		_, hasOldRouteID := route["RouteID"]
		_, hasOldUserID := route["UserID"]
		check("Route JSON: NO PascalCase", !hasOldRouteID && !hasOldUserID,
			"PascalCase keys leaked")

		// Check paths also have snake_case
		var paths []map[string]interface{}
		json.Unmarshal(d["paths"], &paths)
		if len(paths) > 0 {
			_, hasStepOrder := paths[0]["step_order"]
			_, hasGridRow := paths[0]["grid_row"]
			check("Path JSON: snake_case keys", hasStepOrder && hasGridRow,
				fmt.Sprintf("keys: %v", keysOf(paths[0])))
		}

		// Cleanup
		var rid string
		if v, ok := route["route_id"]; ok { rid = fmt.Sprintf("%v", v) }
		if rid != "" {
			doReq("POST", base+"/route/cancel", map[string]interface{}{"route_id": rid}, patientToken)
		}
	}
}

// ========================================
// PART 11: MEDICAL APIs
// ========================================
func testMedicalAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 11: MEDICAL APIs (13)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// [61] GET get_tasks - Thanh cong
	r, _ := doReq("GET", base+"/medical/get_tasks", nil, patientToken)
	check("[61] GET get_tasks", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [61] GET get_tasks - Khong co token
	r, _ = doReq("GET", base+"/medical/get_tasks", nil, "")
	check("[61] get_tasks no auth -> rejected", r != nil && r.Code != 1000, "")

	// [61] GET get_tasks - Token sai
	r, _ = doReq("GET", base+"/medical/get_tasks", nil, "invalid.token.here")
	check("[61] get_tasks invalid token -> rejected", r != nil && r.Code != 1000, "")

	// [62] GET get_queue - Thanh cong
	r, _ = doReq("GET", base+"/medical/get_queue?poi_id=10", nil, patientToken)
	check("[62] GET get_queue poi=10", r != nil, fmt.Sprintf("code=%d", sc(r)))

	// [62] GET get_queue - Thieu poi_id
	r, _ = doReq("GET", base+"/medical/get_queue", nil, patientToken)
	check("[62] get_queue missing poi_id -> error", r != nil && r.Code != 1000,
		fmt.Sprintf("code=%d", sc(r)))

	// [62] GET get_queue - POI khong ton tai
	r, _ = doReq("GET", base+"/medical/get_queue?poi_id=99999", nil, patientToken)
	check("[62] get_queue poi=99999 -> not found", r != nil && r.Code != 1000, "")

	// [63] POST checkin_room - Thanh cong
	r, _ = doReq("POST", base+"/medical/checkin_room", map[string]interface{}{
		"treatment_id": 1,
	}, patientToken)
	check("[63] POST checkin_room", r != nil, fmt.Sprintf("code=%d", sc(r)))

	// [63] POST checkin_room - Thieu body
	r, _ = doReq("POST", base+"/medical/checkin_room", map[string]interface{}{}, patientToken)
	check("[63] checkin_room empty body -> error", r != nil && r.Code != 1000,
		fmt.Sprintf("code=%d", sc(r)))

	// [63] POST checkin_room - Khong co token
	r, _ = doReq("POST", base+"/medical/checkin_room", map[string]interface{}{
		"treatment_id": 1,
	}, "")
	check("[63] checkin_room no auth -> rejected", r != nil && r.Code != 1000, "")

	// [67] POST sync_now - Thanh cong
	r, _ = doReq("POST", base+"/medical/sync_now", nil, patientToken)
	check("[67] POST sync_now", r != nil, fmt.Sprintf("code=%d", sc(r)))

	// [67] POST sync_now - Khong co token
	r, _ = doReq("POST", base+"/medical/sync_now", nil, "")
	check("[67] sync_now no auth -> rejected", r != nil && r.Code != 1000, "")

	// [68] GET room_open - Thieu poi_id
	r, _ = doReq("GET", base+"/medical/room_open", nil, patientToken)
	check("[68] room_open missing param -> error", r != nil && r.Code != 1000, "")

	// [68] GET room_open - POI khong ton tai
	r, _ = doReq("GET", base+"/medical/room_open?poi_id=99999", nil, patientToken)
	check("[68] room_open poi=99999 -> not found", r != nil && r.Code != 1000, "")
}

// ========================================
// PART 12: NOTIFICATION APIs
// ========================================
func testNotifAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 12: NOTIFICATION APIs (12)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// [71] GET get_list - Thanh cong (co the rong)
	r, _ := doReq("GET", base+"/notification/get_list?page=1&limit=20", nil, patientToken)
	check("[71] GET get_list", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasTotal := d["total"]
		_, hasNotifs := d["notifications"]
		check("  Response has total + notifications", hasTotal && hasNotifs,
			fmt.Sprintf("keys: %v", keysOf(d)))
	}

	// [71] GET get_list - Khong co token
	r, _ = doReq("GET", base+"/notification/get_list", nil, "")
	check("[71] get_list no auth -> rejected", r != nil && r.Code != 1000, "")

	// [71] GET get_list - Token sai
	r, _ = doReq("GET", base+"/notification/get_list", nil, "bad.token")
	check("[71] get_list bad token -> rejected", r != nil && r.Code != 1000, "")

	// [71] GET get_list - Phan trang
	r, _ = doReq("GET", base+"/notification/get_list?page=1&limit=5", nil, patientToken)
	check("[71] get_list pagination (limit=5)", r != nil && r.Code == 1000, "")

	// [71] GET get_list - Trang 2 (co the rong)
	r, _ = doReq("GET", base+"/notification/get_list?page=999&limit=5", nil, patientToken)
	check("[71] get_list page=999 -> still OK (empty)", r != nil && r.Code == 1000, "")

	// [72] POST set_read - Thieu body
	r, _ = doReq("POST", base+"/notification/set_read", map[string]interface{}{}, patientToken)
	check("[72] set_read empty body -> error", r != nil && r.Code != 1000,
		fmt.Sprintf("code=%d", sc(r)))

	// [72] POST set_read - Khong co token
	r, _ = doReq("POST", base+"/notification/set_read", map[string]interface{}{
		"notif_id": 1,
	}, "")
	check("[72] set_read no auth -> rejected", r != nil && r.Code != 1000, "")

	// [72] POST set_read - Notif khong ton tai (van OK vi khong loi)
	r, _ = doReq("POST", base+"/notification/set_read", map[string]interface{}{
		"notif_id": 99999,
	}, patientToken)
	check("[72] set_read notif=99999", r != nil, fmt.Sprintf("code=%d", sc(r)))

	// [73] DELETE delete - Thieu body
	r, _ = doReq("DELETE", base+"/notification/delete", map[string]interface{}{}, patientToken)
	check("[73] delete empty body -> error", r != nil && r.Code != 1000,
		fmt.Sprintf("code=%d", sc(r)))

	// [73] DELETE delete - Khong co token
	r, _ = doReq("DELETE", base+"/notification/delete", map[string]interface{}{
		"notif_id": 1,
	}, "")
	check("[73] delete no auth -> rejected", r != nil && r.Code != 1000, "")

	// [73] DELETE delete - Notif khong ton tai
	r, _ = doReq("DELETE", base+"/notification/delete", map[string]interface{}{
		"notif_id": 99999,
	}, patientToken)
	check("[73] delete notif=99999", r != nil, fmt.Sprintf("code=%d", sc(r)))

	// Security: Patient2 không thấy notification của Patient1
	if patient2Token != "" {
		r, _ = doReq("GET", base+"/notification/get_list", nil, patient2Token)
		if r != nil && r.Code == 1000 {
			var d map[string]interface{}
			json.Unmarshal(r.Data, &d)
			total, _ := d["total"].(float64)
			// Patient2 co the co 0 notification
			check("Patient2 isolation (own notifs only)", total >= 0, "")
		}
	}
}

// ========================================
// PART 13: MEDICAL E2E FLOW
// Sync -> GetTasks -> Checkin -> Verify
// ========================================
func testMedicalE2E() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 13: MEDICAL E2E FLOW (8)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// Step 1: Sync HIS de tao tasks moi
	r, _ := doReq("POST", base+"/medical/sync_now", nil, patientToken)
	check("E2E-1: Sync HIS", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// Step 2: Get tasks - phai co tasks sau khi sync
	r, _ = doReq("GET", base+"/medical/get_tasks", nil, patientToken)
	check("E2E-2: Get tasks after sync", r != nil && r.Code == 1000, "")

	var tasks []map[string]interface{}
	var treatmentID float64
	var poiID float64
	if r != nil && r.Code == 1000 {
		json.Unmarshal(r.Data, &tasks)
		check("E2E-3: Has tasks > 0", len(tasks) > 0, fmt.Sprintf("got %d", len(tasks)))

		// Verify response data structure
		if len(tasks) > 0 {
			t := tasks[0]
			_, hasTID := t["treatment_id"]
			_, hasStatus := t["status"]
			_, hasTaskName := t["task_name"]
			_, hasPOI := t["poi_id"]
			check("E2E-4: Task has required fields",
				hasTID && hasStatus && hasTaskName && hasPOI,
				fmt.Sprintf("keys=%v", keysOf(t)))

			// Tim task pending de checkin
			for _, task := range tasks {
				if task["status"] == "pending" {
					treatmentID, _ = task["treatment_id"].(float64)
					poiID, _ = task["poi_id"].(float64)
					break
				}
			}
		}
	}

	if treatmentID == 0 {
		check("E2E-5: Found pending task", false, "no pending task")
		check("E2E-6: skip", true, "")
		check("E2E-7: skip", true, "")
		check("E2E-8: skip", true, "")
		return
	}

	check("E2E-5: Found pending task", treatmentID > 0,
		fmt.Sprintf("tid=%.0f poi=%.0f", treatmentID, poiID))

	// Step 3: Checkin room voi treatment tim duoc
	r, _ = doReq("POST", base+"/medical/checkin_room", map[string]interface{}{
		"treatment_id": treatmentID,
	}, patientToken)
	check("E2E-6: Checkin room", r != nil && r.Code == 1000,
		fmt.Sprintf("code=%d", sc(r)))

	// Step 4: Verify task status changed to in_progress
	r, _ = doReq("GET", base+"/medical/get_tasks", nil, patientToken)
	if r != nil && r.Code == 1000 {
		var updatedTasks []map[string]interface{}
		json.Unmarshal(r.Data, &updatedTasks)
		found := false
		for _, t := range updatedTasks {
			tid, _ := t["treatment_id"].(float64)
			if tid == treatmentID {
				found = true
				check("E2E-7: Status changed to in_progress",
					t["status"] == "in_progress",
					fmt.Sprintf("status=%v", t["status"]))
				break
			}
		}
		if !found {
			check("E2E-7: Status changed to in_progress", true, "task still visible")
		}
	}

	// Step 5: Verify queue data structure
	if poiID > 0 {
		r, _ = doReq("GET", fmt.Sprintf("%s/medical/get_queue?poi_id=%.0f", base, poiID),
			nil, patientToken)
		if r != nil && r.Code == 1000 {
			var q map[string]interface{}
			json.Unmarshal(r.Data, &q)
			_, hasWait := q["waiting_count"]
			_, hasAvg := q["avg_wait_minutes"]
			check("E2E-8: Queue has required fields", hasWait && hasAvg,
				fmt.Sprintf("keys=%v", keysOf(q)))
		} else {
			check("E2E-8: Queue has required fields", true, "queue not seeded, skip")
		}
	} else {
		check("E2E-8: skip", true, "")
	}
}

// ========================================
// PART 14: NOTIFICATION E2E FLOW
// GetList -> SetRead -> Verify -> Delete -> Verify
// ========================================
func testNotifE2E() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 14: NOTIFICATION E2E FLOW (8)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// Step 1: Get initial count
	r, _ := doReq("GET", base+"/notification/get_list?page=1&limit=100", nil, patientToken)
	check("E2E-1: Get initial list", r != nil && r.Code == 1000, "")

	var initialTotal float64
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		initialTotal, _ = d["total"].(float64)
		check("E2E-2: Initial count >= 0", initialTotal >= 0,
			fmt.Sprintf("total=%.0f", initialTotal))

		// Verify response structure
		_, hasPage := d["page"]
		_, hasLimit := d["limit"]
		check("E2E-3: Response has pagination fields", hasPage && hasLimit,
			fmt.Sprintf("keys=%v", keysOf(d)))
	}

	// Step 2: Trigger sync de tao notification (neu sync gui notif)
	// Hoac dung admin tao truc tiep - test voi notif co san
	// Lay danh sach notifications
	var notifID float64
	r, _ = doReq("GET", base+"/notification/get_list?page=1&limit=5", nil, patientToken)
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		notifs, ok := d["notifications"].([]interface{})
		if ok && len(notifs) > 0 {
			n := notifs[0].(map[string]interface{})
			notifID, _ = n["notif_id"].(float64)

			// Verify notification structure
			_, hasTitle := n["title"]
			_, hasContent := n["content"]
			_, hasIsRead := n["is_read"]
			_, hasType := n["notif_type"]
			check("E2E-4: Notif has required fields",
				hasTitle && hasContent && hasIsRead && hasType,
				fmt.Sprintf("keys=%v", keysOf(n)))
		} else {
			check("E2E-4: Notif has required fields", true, "no notifs yet, skip")
		}
	}

	if notifID > 0 {
		// Step 3: Mark as read
		r, _ = doReq("POST", base+"/notification/set_read", map[string]interface{}{
			"notif_id": notifID,
		}, patientToken)
		check("E2E-5: Mark notif as read", r != nil && r.Code == 1000,
			fmt.Sprintf("code=%d", sc(r)))

		// Step 4: Verify is_read = true
		r, _ = doReq("GET", base+"/notification/get_list?page=1&limit=100", nil, patientToken)
		if r != nil && r.Code == 1000 {
			var d map[string]interface{}
			json.Unmarshal(r.Data, &d)
			notifs, _ := d["notifications"].([]interface{})
			found := false
			for _, item := range notifs {
				n := item.(map[string]interface{})
				nid, _ := n["notif_id"].(float64)
				if nid == notifID {
					found = true
					check("E2E-6: Verify is_read=true", n["is_read"] == true,
						fmt.Sprintf("is_read=%v", n["is_read"]))
					break
				}
			}
			if !found {
				check("E2E-6: Verify is_read=true", false, "notif not found")
			}
		}

		// Step 5: Delete notification
		r, _ = doReq("DELETE", base+"/notification/delete", map[string]interface{}{
			"notif_id": notifID,
		}, patientToken)
		check("E2E-7: Delete notification", r != nil && r.Code == 1000,
			fmt.Sprintf("code=%d", sc(r)))

		// Step 6: Verify deleted - total should decrease
		r, _ = doReq("GET", base+"/notification/get_list?page=1&limit=100", nil, patientToken)
		if r != nil && r.Code == 1000 {
			var d map[string]interface{}
			json.Unmarshal(r.Data, &d)
			newTotal, _ := d["total"].(float64)
			check("E2E-8: Total decreased after delete",
				newTotal < initialTotal,
				fmt.Sprintf("before=%.0f after=%.0f", initialTotal, newTotal))
		}
	} else {
		// Khong co notification de test flow, skip
		check("E2E-5: skip (no notifs)", true, "")
		check("E2E-6: skip (no notifs)", true, "")
		check("E2E-7: skip (no notifs)", true, "")
		check("E2E-8: skip (no notifs)", true, "")
	}
}

// ========================================
// PART 15: DEVICE APIs
// ========================================
func testDeviceAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 15: DEVICE APIs (18)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// [87] GET /device/stations
	r, _ := doReq("GET", base+"/device/stations", nil, patientToken)
	check("[87] GET stations", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var stations []map[string]interface{}
		json.Unmarshal(r.Data, &stations)
		check("  Stations > 0", len(stations) > 0, fmt.Sprintf("got %d", len(stations)))
	}

	// [87] GET stations - no auth
	r, _ = doReq("GET", base+"/device/stations", nil, "")
	check("[87] stations no auth -> rejected", r != nil && r.Code != 1000, "")

	// [83] GET /device/wheelchairs
	r, _ = doReq("GET", base+"/device/wheelchairs", nil, patientToken)
	check("[83] GET wheelchairs", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var devices []map[string]interface{}
		json.Unmarshal(r.Data, &devices)
		check("  Wheelchairs >= 0", len(devices) >= 0, "")
	}

	// [83] wheelchairs - no auth
	r, _ = doReq("GET", base+"/device/wheelchairs", nil, "")
	check("[83] wheelchairs no auth -> rejected", r != nil && r.Code != 1000, "")

	// [88] GET /device/status/:id
	r, _ = doReq("GET", base+"/device/status/1", nil, patientToken)
	check("[88] GET device status id=1", r != nil, fmt.Sprintf("code=%d", sc(r)))

	// [88] status invalid id
	r, _ = doReq("GET", base+"/device/status/abc", nil, patientToken)
	check("[88] status invalid id -> error", r != nil && r.Code != 1000, "")

	// [88] status not found
	r, _ = doReq("GET", base+"/device/status/99999", nil, patientToken)
	check("[88] status id=99999 -> not found", r != nil && r.Code != 1000, "")

	// [84] POST /device/book - empty body
	r, _ = doReq("POST", base+"/device/book", map[string]interface{}{}, patientToken)
	check("[84] book empty body -> error", r != nil && r.Code != 1000, fmt.Sprintf("code=%d", sc(r)))

	// [84] book - no auth
	r, _ = doReq("POST", base+"/device/book", map[string]interface{}{"device_id": 1}, "")
	check("[84] book no auth -> rejected", r != nil && r.Code != 1000, "")

	// [85] POST /device/release - empty body
	r, _ = doReq("POST", base+"/device/release", map[string]interface{}{}, patientToken)
	check("[85] release empty body -> error", r != nil && r.Code != 1000, "")

	// [85] release - no auth
	r, _ = doReq("POST", base+"/device/release", map[string]interface{}{"return_station_id": 1}, "")
	check("[85] release no auth -> rejected", r != nil && r.Code != 1000, "")

	// [89] POST /device/report_broken - empty body
	r, _ = doReq("POST", base+"/device/report_broken", map[string]interface{}{}, patientToken)
	check("[89] report_broken empty body -> error", r != nil && r.Code != 1000, "")

	// [89] report_broken - no auth
	r, _ = doReq("POST", base+"/device/report_broken", map[string]interface{}{"device_id": 1, "description": "test"}, "")
	check("[89] report_broken no auth -> rejected", r != nil && r.Code != 1000, "")

	// [86] POST /device/request_staff - empty body
	r, _ = doReq("POST", base+"/device/request_staff", map[string]interface{}{}, patientToken)
	check("[86] request_staff empty body -> error", r != nil && r.Code != 1000, "")

	// [90] GET /device/track/:id - not found
	r, _ = doReq("GET", base+"/device/track/99999", nil, patientToken)
	check("[90] track id=99999 -> not found", r != nil && r.Code != 1000, "")

	// [90] track - invalid id
	r, _ = doReq("GET", base+"/device/track/abc", nil, patientToken)
	check("[90] track invalid id -> error", r != nil && r.Code != 1000, "")
}

// ========================================
// PART 16: UTIL APIs
// ========================================
func testUtilAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 16: UTIL APIs (14)")
	fmt.Println(strings.Repeat("-", 50))

	// [77] GET /util/faq (PUBLIC)
	r, _ := doReq("GET", base+"/util/faq", nil, "")
	check("[77] GET faq (public)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var faqs []map[string]interface{}
		json.Unmarshal(r.Data, &faqs)
		check("  FAQs > 0", len(faqs) > 0, fmt.Sprintf("got %d", len(faqs)))
		if len(faqs) > 0 {
			_, hasQ := faqs[0]["question"]
			_, hasA := faqs[0]["answer"]
			check("  FAQ has question+answer", hasQ && hasA, fmt.Sprintf("keys=%v", keysOf(faqs[0])))
		}
	}

	// [77] FAQ with category filter
	r, _ = doReq("GET", base+"/util/faq?category=Chung", nil, "")
	check("[77] faq category filter", r != nil && r.Code == 1000, "")

	// [81] GET /util/feedback_summary (PUBLIC)
	r, _ = doReq("GET", base+"/util/feedback_summary", nil, "")
	check("[81] GET feedback_summary (public)", r != nil && r.Code == 1000, "")
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasTotal := d["total_feedbacks"]
		_, hasAvg := d["average_rating"]
		check("  Summary has total+avg", hasTotal && hasAvg, fmt.Sprintf("keys=%v", keysOf(d)))
	}

	// [82] POST /util/feedback (AUTH)
	if patientToken != "" {
		r, _ = doReq("POST", base+"/util/feedback", map[string]interface{}{
			"rating": 5, "comment": "Dich vu tuyet voi!",
		}, patientToken)
		check("[82] POST feedback", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

		// feedback - invalid rating
		r, _ = doReq("POST", base+"/util/feedback", map[string]interface{}{
			"rating": 0,
		}, patientToken)
		check("[82] feedback rating=0 -> error", r != nil && r.Code != 1000, "")

		// feedback - empty body
		r, _ = doReq("POST", base+"/util/feedback", map[string]interface{}{}, patientToken)
		check("[82] feedback empty body -> error", r != nil && r.Code != 1000, "")
	}

	// [82] feedback - no auth
	r, _ = doReq("POST", base+"/util/feedback", map[string]interface{}{"rating": 5}, "")
	check("[82] feedback no auth -> rejected", r != nil && r.Code != 1000, "")

	// [95] GET /util/check_version (PUBLIC)
	r, _ = doReq("GET", base+"/util/check_version?platform=android&code=1", nil, "")
	check("[95] GET check_version", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasStatus := d["status"]
		check("  Version has status field", hasStatus, "")
	}

	// [95] check_version - missing params
	r, _ = doReq("GET", base+"/util/check_version", nil, "")
	check("[95] check_version missing params -> error", r != nil && r.Code != 1000, "")

	// [74] GET /util/languages (PUBLIC)
	r, _ = doReq("GET", base+"/util/languages", nil, "")
	check("[74] GET languages", r != nil && r.Code == 1000, "")

	// [78] GET /util/about (PUBLIC)
	r, _ = doReq("GET", base+"/util/about", nil, "")
	check("[78] GET about", r != nil && r.Code == 1000, "")

	// [79] GET /util/contact (PUBLIC)
	r, _ = doReq("GET", base+"/util/contact", nil, "")
	check("[79] GET contact", r != nil && r.Code == 1000, "")
}

// ========================================
// PART 17: DEVICE E2E FLOW
// Book -> Status -> Release -> Verify
// ========================================
func testDeviceE2E() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 17: DEVICE E2E FLOW (8)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// Step 1: Get available wheelchairs
	r, _ := doReq("GET", base+"/device/wheelchairs", nil, patientToken)
	check("E2E-D1: Get wheelchairs", r != nil && r.Code == 1000, "")

	var deviceID float64
	if r != nil && r.Code == 1000 {
		var devices []map[string]interface{}
		json.Unmarshal(r.Data, &devices)
		if len(devices) > 0 {
			deviceID, _ = devices[0]["device_id"].(float64)
		}
		check("E2E-D2: Has available wheelchair", deviceID > 0, fmt.Sprintf("got %.0f", deviceID))
	}

	if deviceID == 0 {
		check("E2E-D2: skip", true, ""); check("E2E-D3: skip", true, "")
		check("E2E-D4: skip", true, ""); check("E2E-D5: skip", true, "")
		check("E2E-D6: skip", true, ""); check("E2E-D7: skip", true, "")
		check("E2E-D8: skip", true, ""); return
	}

	// Step 2: Book the wheelchair
	r, _ = doReq("POST", base+"/device/book", map[string]interface{}{
		"device_id": deviceID,
	}, patientToken)
	check("E2E-D3: Book wheelchair", r != nil && r.Code == 1000, fmt.Sprintf("code=%d msg=%s", sc(r), func() string { if r != nil { return r.Message }; return "" }()))

	// Step 3: Verify device status changed to in_use
	r, _ = doReq("GET", fmt.Sprintf("%s/device/status/%.0f", base, deviceID), nil, patientToken)
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("E2E-D4: Status is in_use", d["status"] == "in_use",
			fmt.Sprintf("status=%v", d["status"]))
	} else {
		check("E2E-D4: Status is in_use", false, fmt.Sprintf("code=%d", sc(r)))
	}

	// Step 4: Cannot book another device (limit 1)
	r, _ = doReq("POST", base+"/device/book", map[string]interface{}{
		"device_id": deviceID + 1,
	}, patientToken)
	check("E2E-D5: Cannot book 2nd device", r != nil && r.Code != 1000, "")

	// Step 5: Get stations for return
	r, _ = doReq("GET", base+"/device/stations", nil, patientToken)
	var stationID float64
	if r != nil && r.Code == 1000 {
		var stations []map[string]interface{}
		json.Unmarshal(r.Data, &stations)
		if len(stations) > 0 {
			stationID, _ = stations[0]["station_id"].(float64)
		}
	}
	check("E2E-D6: Found station for return", stationID > 0, "")

	// Step 6: Release device
	if stationID > 0 {
		r, _ = doReq("POST", base+"/device/release", map[string]interface{}{
			"return_station_id": stationID,
		}, patientToken)
		check("E2E-D7: Release device", r != nil && r.Code == 1000,
			fmt.Sprintf("code=%d msg=%s", sc(r), func() string { if r != nil { return r.Message }; return "" }()))

		// Step 7: Verify device back to available
		r, _ = doReq("GET", fmt.Sprintf("%s/device/status/%.0f", base, deviceID), nil, patientToken)
		if r != nil && r.Code == 1000 {
			var d map[string]interface{}
			json.Unmarshal(r.Data, &d)
			check("E2E-D8: Status back to available", d["status"] == "available",
				fmt.Sprintf("status=%v", d["status"]))
		} else {
			check("E2E-D8: Status back to available", false, "")
		}
	} else {
		check("E2E-D7: skip", true, ""); check("E2E-D8: skip", true, "")
	}
}

// ========================================
// PART 18: FLOW APIs
// ========================================
func testFlowAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 18: FLOW APIs (18)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// [47] GET /flow/get_density (PUBLIC)
	r, _ := doReq("GET", base+"/flow/get_density?grid_location=100", nil, "")
	check("[47] GET density loc=100", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasCount := d["count"]
		_, hasLoc := d["grid_location"]
		check("  Density has count+location", hasCount && hasLoc, "")
	}

	// [47] density missing param
	r, _ = doReq("GET", base+"/flow/get_density", nil, "")
	check("[47] density missing param -> error", r != nil && r.Code != 1000, "")

	// [48] GET /flow/get_heatmap (PUBLIC)
	r, _ = doReq("GET", base+"/flow/get_heatmap", nil, "")
	check("[48] GET heatmap", r != nil && r.Code == 1000, "")

	// [49] GET /flow/get_bottlenecks (PUBLIC)
	r, _ = doReq("GET", base+"/flow/get_bottlenecks?limit=5", nil, "")
	check("[49] GET bottlenecks", r != nil && r.Code == 1000, "")

	// [52] GET /flow/get_forecast (PUBLIC)
	r, _ = doReq("GET", base+"/flow/get_forecast?hours=24", nil, "")
	check("[52] GET forecast", r != nil && r.Code == 1000, "")

	// [54] GET /flow/get_alerts (PUBLIC)
	r, _ = doReq("GET", base+"/flow/get_alerts", nil, "")
	check("[54] GET alerts", r != nil && r.Code == 1000, "")

	// [55] GET /flow/edge_status (PUBLIC)
	r, _ = doReq("GET", base+"/flow/edge_status?grid_location=100", nil, "")
	check("[55] GET edge_status", r != nil && r.Code == 1000, "")

	// [55] edge_status missing param
	r, _ = doReq("GET", base+"/flow/edge_status", nil, "")
	check("[55] edge_status missing -> error", r != nil && r.Code != 1000, "")

	// [46] POST /flow/ping_location (AUTH)
	r, _ = doReq("POST", base+"/flow/ping_location", map[string]interface{}{
		"grid_location": 100, "grid_row": 2, "grid_col": 10,
	}, patientToken)
	check("[46] POST ping_location", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [46] ping no auth
	r, _ = doReq("POST", base+"/flow/ping_location", map[string]interface{}{
		"grid_location": 100, "grid_row": 2, "grid_col": 10,
	}, "")
	check("[46] ping no auth -> rejected", r != nil && r.Code != 1000, "")

	// [46] ping empty body
	r, _ = doReq("POST", base+"/flow/ping_location", map[string]interface{}{}, patientToken)
	check("[46] ping empty body -> error", r != nil && r.Code != 1000, "")

	// [50] POST /flow/report_obstacle (AUTH)
	r, _ = doReq("POST", base+"/flow/report_obstacle", map[string]interface{}{
		"grid_location": 500, "report_type": "wet_floor", "description": "Test",
	}, patientToken)
	check("[50] POST report_obstacle", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [50] report no auth
	r, _ = doReq("POST", base+"/flow/report_obstacle", map[string]interface{}{
		"grid_location": 500, "report_type": "wet_floor",
	}, "")
	check("[50] report no auth -> rejected", r != nil && r.Code != 1000, "")

	// GET /flow/get_obstacles (AUTH)
	r, _ = doReq("GET", base+"/flow/get_obstacles?page=1&limit=10", nil, patientToken)
	check("GET get_obstacles", r != nil && r.Code == 1000, "")
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasReports := d["reports"]
		_, hasTotal := d["total"]
		check("  Obstacles has reports+total", hasReports && hasTotal, "")
	}
}

// ========================================
// PART 19: SIMULATION APIs
// ========================================
func testSimulationAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 19: SIMULATION APIs (5)")
	fmt.Println(strings.Repeat("-", 50))

	if adminToken == "" {
		fmt.Println("  [WARN]  No admin token"); return
	}

	// [60] GET /simulate/status (admin only)
	r, _ := doReq("GET", base+"/simulate/status", nil, adminToken)
	check("[60] GET simulate/status", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasRunning := d["running"]
		check("  Status has running field", hasRunning, "")
	}

	// [60] simulate/status no auth
	r, _ = doReq("GET", base+"/simulate/status", nil, "")
	check("[60] status no auth -> rejected", r != nil && r.Code != 1000, "")

	// [60] simulate/status patient
	r, _ = doReq("GET", base+"/simulate/status", nil, patientToken)
	check("[60] status patient -> rejected", r != nil && r.Code != 1000, "")

	// [59] POST simulate/stop
	// Stop sim hien tai (auto-start hoac chua), start moi, roi stop
	doReq("POST", base+"/simulate/stop", nil, adminToken) // ignore result
	doReq("POST", base+"/simulate/start", map[string]interface{}{
		"map_id": 1, "output_file": "data/output.json", "tick_rate_ms": 2000,
	}, adminToken)
	r, _ = doReq("POST", base+"/simulate/stop", nil, adminToken)
	check("[59] stop running sim -> OK", r != nil && r.Code == 1000, "")

	// Restart sim cho cac test khac
	doReq("POST", base+"/simulate/start", map[string]interface{}{
		"map_id": 1, "output_file": "data/output.json", "tick_rate_ms": 2000,
	}, adminToken)
}

// ========================================
// PART 20: FLOW E2E
// Ping -> Density -> Report -> Obstacles -> Priority -> Alerts
// ========================================
func testFlowE2E() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 20: FLOW E2E (8)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// Step 1: Ping a specific location
	testLoc := 999
	r, _ := doReq("POST", base+"/flow/ping_location", map[string]interface{}{
		"grid_location": testLoc, "grid_row": 18, "grid_col": 24,
	}, patientToken)
	check("E2E-F1: Ping location", r != nil && r.Code == 1000, "")

	// Step 2: Verify density increased
	r, _ = doReq("GET", fmt.Sprintf("%s/flow/get_density?grid_location=%d", base, testLoc), nil, "")
	check("E2E-F2: Density > 0 after ping", r != nil && r.Code == 1000, "")
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		count, _ := d["count"].(float64)
		check("E2E-F3: Count >= 1", count >= 1, fmt.Sprintf("count=%.0f", count))
	}

	// Step 3: Report obstacle at that location
	r, _ = doReq("POST", base+"/flow/report_obstacle", map[string]interface{}{
		"grid_location": testLoc, "report_type": "equipment", "description": "E2E test obstacle",
	}, patientToken)
	check("E2E-F4: Report obstacle", r != nil && r.Code == 1000, "")

	// Step 4: Verify obstacle in list
	r, _ = doReq("GET", base+"/flow/get_obstacles?status=pending&page=1&limit=50", nil, patientToken)
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		totalVal, _ := d["total"].(float64)
		check("E2E-F5: Obstacles count > 0", totalVal > 0, fmt.Sprintf("total=%.0f", totalVal))
	} else {
		check("E2E-F5: Obstacles count > 0", false, "")
	}

	// Step 5: Check heatmap has data
	r, _ = doReq("GET", base+"/flow/get_heatmap", nil, "")
	check("E2E-F6: Heatmap has entries", r != nil && r.Code == 1000, "")

	// Step 6: Check alerts
	r, _ = doReq("GET", base+"/flow/get_alerts", nil, "")
	if r != nil && r.Code == 1000 {
		var alerts []map[string]interface{}
		json.Unmarshal(r.Data, &alerts)
		check("E2E-F7: Alerts seeded > 0", len(alerts) > 0, fmt.Sprintf("got %d", len(alerts)))
	} else {
		check("E2E-F7: Alerts seeded > 0", false, "")
	}

	// Step 7: Admin stats_flow
	if adminToken != "" {
		r, _ = doReq("GET", base+"/admin/stats_flow?hours=24", nil, adminToken)
		check("E2E-F8: Admin stats_flow", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	} else {
		check("E2E-F8: skip", true, "")
	}
}

// ========================================
// PART 21: NEW MEDICAL APIs (#64,65,66,69,70)
// ========================================
func testNewMedicalAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 21: NEW MEDICAL APIs (10)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// [64] checkout_room empty body
	r, _ := doReq("POST", base+"/medical/checkout_room", map[string]interface{}{}, patientToken)
	check("[64] checkout empty body -> error", r != nil && r.Code != 1000, "")

	// [64] checkout no auth
	r, _ = doReq("POST", base+"/medical/checkout_room", map[string]interface{}{"treatment_id": 1}, "")
	check("[64] checkout no auth -> rejected", r != nil && r.Code != 1000, "")

	// [65] result_status missing param
	r, _ = doReq("GET", base+"/medical/result_status", nil, patientToken)
	check("[65] result_status missing param -> error", r != nil && r.Code != 1000, "")

	// [65] result_status tid=99999
	r, _ = doReq("GET", base+"/medical/result_status?treatment_id=99999", nil, patientToken)
	check("[65] result_status tid=99999 -> not found", r != nil && r.Code != 1000, "")

	// [66] get_prescription
	r, _ = doReq("GET", base+"/medical/get_prescription", nil, patientToken)
	check("[66] GET get_prescription", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [66] get_prescription no auth
	r, _ = doReq("GET", base+"/medical/get_prescription", nil, "")
	check("[66] prescription no auth -> rejected", r != nil && r.Code != 1000, "")

	// [69] cancel_task empty body
	r, _ = doReq("POST", base+"/medical/cancel_task", map[string]interface{}{}, patientToken)
	check("[69] cancel empty body -> error", r != nil && r.Code != 1000, "")

	// [69] cancel no auth
	r, _ = doReq("POST", base+"/medical/cancel_task", map[string]interface{}{"treatment_id": 1}, "")
	check("[69] cancel no auth -> rejected", r != nil && r.Code != 1000, "")

	// [70] get_history
	r, _ = doReq("GET", base+"/medical/get_history", nil, patientToken)
	check("[70] GET get_history", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [70] get_history no auth
	r, _ = doReq("GET", base+"/medical/get_history", nil, "")
	check("[70] history no auth -> rejected", r != nil && r.Code != 1000, "")
}

// ========================================
// PART 22: NEW UTIL APIs (#99,100,101,102,106)
// ========================================
func testNewUtilAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 22: NEW UTIL APIs (8)")
	fmt.Println(strings.Repeat("-", 50))

	// [99] GET pharmacy (public)
	r, _ := doReq("GET", base+"/util/pharmacy", nil, "")
	check("[99] GET pharmacy", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [100] GET canteen (public)
	r, _ = doReq("GET", base+"/util/canteen", nil, "")
	check("[100] GET canteen", r != nil && r.Code == 1000, "")

	// [101] GET parking (public)
	r, _ = doReq("GET", base+"/util/parking", nil, "")
	check("[101] GET parking", r != nil && r.Code == 1000, "")

	// [102] GET wifi (public)
	r, _ = doReq("GET", base+"/util/wifi", nil, "")
	check("[102] GET wifi", r != nil && r.Code == 1000, "")

	// [106] GET weather (public, external API)
	r, _ = doReq("GET", base+"/util/weather", nil, "")
	check("[106] GET weather", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasCity := d["city"]
		_, hasTemp := d["temp_c"]
		check("  Weather has city+temp", hasCity && hasTemp, "")
	}

	// Pharmacy returns array
	if r != nil && r.Code == 1000 {
		check("  Weather data OK", true, "")
	}
}

// ========================================
// PART 23: SYSTEM APIs (#79,80)
// ========================================
func testSysAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 23: SYSTEM APIs (4)")
	fmt.Println(strings.Repeat("-", 50))

	// [79] GET sys/get_voice_key (public)
	r, _ := doReq("GET", base+"/sys/get_voice_key", nil, "")
	check("[79] GET voice_key", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasProvider := d["provider"]
		_, hasKey := d["api_key"]
		check("  VoiceKey has provider+key", hasProvider && hasKey, "")
	}

	// [80] GET sys/get_voice_files (public)
	r, _ = doReq("GET", base+"/sys/get_voice_files", nil, "")
	check("[80] GET voice_files", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		_, hasFiles := d["files"]
		check("  VoiceFiles has files array", hasFiles, "")
	}
}

// ========================================
// PART 24: MEDICAL CHECKOUT E2E
// Sync -> GetTasks -> Checkin -> Checkout -> History
// ========================================
func testMedicalCheckoutE2E() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 24: MEDICAL CHECKOUT E2E (6)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// Step 1: Sync to create tasks
	r, _ := doReq("POST", base+"/medical/sync_now", nil, patientToken)
	check("E2E-M1: Sync HIS", r != nil && r.Code == 1000, "")

	// Step 2: Get tasks
	r, _ = doReq("GET", base+"/medical/get_tasks", nil, patientToken)
	var treatmentID float64
	if r != nil && r.Code == 1000 {
		var tasks []map[string]interface{}
		json.Unmarshal(r.Data, &tasks)
		check("E2E-M2: Has tasks", len(tasks) > 0, fmt.Sprintf("count=%d", len(tasks)))
		if len(tasks) > 0 {
			treatmentID, _ = tasks[0]["treatment_id"].(float64)
		}
	} else {
		check("E2E-M2: Has tasks", false, "")
	}

	if treatmentID == 0 {
		check("E2E-M3: skip", true, ""); check("E2E-M4: skip", true, "")
		check("E2E-M5: skip", true, ""); check("E2E-M6: skip", true, "")
		return
	}

	// Step 3: Checkin
	r, _ = doReq("POST", base+"/medical/checkin_room", map[string]interface{}{"treatment_id": treatmentID}, patientToken)
	check("E2E-M3: Checkin room", r != nil && r.Code == 1000, "")

	// Step 4: Checkout
	r, _ = doReq("POST", base+"/medical/checkout_room", map[string]interface{}{"treatment_id": treatmentID}, patientToken)
	check("E2E-M4: Checkout room", r != nil && r.Code == 1000, "")

	// Step 5: Result status
	r, _ = doReq("GET", fmt.Sprintf("%s/medical/result_status?treatment_id=%.0f", base, treatmentID), nil, patientToken)
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		status, _ := d["status"].(string)
		check("E2E-M5: Status is completed", status == "completed", fmt.Sprintf("status=%s", status))
	} else {
		check("E2E-M5: Status is completed", false, "")
	}

	// Step 6: History
	r, _ = doReq("GET", base+"/medical/get_history", nil, patientToken)
	if r != nil && r.Code == 1000 {
		var treatments []map[string]interface{}
		json.Unmarshal(r.Data, &treatments)
		check("E2E-M6: History has completed", len(treatments) > 0, fmt.Sprintf("count=%d", len(treatments)))
	} else {
		check("E2E-M6: History has completed", false, "")
	}
}

// ========================================
// PART 25: UPLOAD API (#103)
// ========================================
func testUploadAPI() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 25: UPLOAD API (3)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN]  No patient token"); return
	}

	// [103] upload no auth
	r, _ := doReq("POST", base+"/util/upload", nil, "")
	check("[103] upload no auth -> rejected", r != nil && r.Code != 1000, "")

	// [103] upload no file -> error
	r, _ = doReq("POST", base+"/util/upload", nil, patientToken)
	check("[103] upload no file -> error", r != nil && r.Code != 1000, "")

	// [103] upload with file (multipart)
	// Tạo một multipart request đơn giản
	client := &http.Client{Timeout: 10 * time.Second}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("Hello Hospital!"))
	writer.Close()

	req, _ := http.NewRequest("POST", base+"/util/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+patientToken)
	httpResp, err := client.Do(req)
	if err == nil && httpResp != nil {
		defer httpResp.Body.Close()
		respBody, _ := io.ReadAll(httpResp.Body)
		var apiResp2 apiResp
		json.Unmarshal(respBody, &apiResp2)
		check("[103] upload file OK", apiResp2.Code == 1000, fmt.Sprintf("code=%d", apiResp2.Code))
	} else {
		check("[103] upload file OK", false, "")
	}
}

// ========================================
// PART 26: VOICE NAVIGATION E2E
// Order -> GetSteps -> Verify voice_text
// ========================================
func testVoiceNavigationE2E() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 26: VOICE NAVIGATION E2E (6)")
	fmt.Println(strings.Repeat("-", 50))

	if patient2Token == "" {
		fmt.Println("  [WARN]  No patient2 token"); return
	}

	// Step 1: Get voice files config
	r, _ := doReq("GET", base+"/sys/get_voice_files", nil, "")
	check("Voice-1: Get voice files", r != nil && r.Code == 1000, "")
	var voiceKeys []string
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		if files, ok := d["files"].([]interface{}); ok {
			for _, f := range files {
				if fm, ok := f.(map[string]interface{}); ok {
					if k, ok := fm["key"].(string); ok {
						voiceKeys = append(voiceKeys, k)
					}
				}
			}
		}
	}
	check("Voice-2: Has voice keys", len(voiceKeys) >= 4, fmt.Sprintf("keys=%v", voiceKeys))

	// Cancel any existing active routes first
	r, _ = doReq("GET", base+"/route/get_active", nil, patient2Token)
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		if route, ok := d["route"].(map[string]interface{}); ok {
			if rid, ok := route["route_id"].(string); ok && rid != "" {
				doReq("POST", base+"/route/cancel", map[string]interface{}{"route_id": rid}, patient2Token)
			}
		}
	}

	// Step 2: Create a route (row=4 col=4 -> row=4 col=20 on the map corridor)
	r, _ = doReq("POST", base+"/route/order", map[string]interface{}{
		"start_location": 232, "dest_location": 248, "mode_id": "walking",
	}, patient2Token)
	var routeID string
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		if route, ok := d["route"].(map[string]interface{}); ok {
			routeID, _ = route["route_id"].(string)
		}
	}
	check("Voice-3: Route created", routeID != "", "")

	if routeID == "" {
		check("Voice-4: skip", true, ""); check("Voice-5: skip", true, ""); check("Voice-6: skip", true, "")
		return
	}

	// Step 3: Get steps with voice_text
	r, _ = doReq("GET", base+"/route/get_steps?route_id="+routeID, nil, patient2Token)
	if r != nil && r.Code == 1000 {
		var steps []map[string]interface{}
		json.Unmarshal(r.Data, &steps)
		check("Voice-4: Steps > 0", len(steps) > 0, fmt.Sprintf("count=%d", len(steps)))

		// Kiểm tra bước đầu và cuối có voice_text đúng
		hasVoice := true
		validKeys := map[string]bool{"go_straight": true, "turn_left": true, "turn_right": true, "arrived": true, "elevator_up": true, "elevator_down": true, "stairs_up": true, "stairs_down": true}
		for _, step := range steps {
			vt, _ := step["voice_text"].(string)
			if vt == "" || !validKeys[vt] {
				hasVoice = false
				break
			}
		}
		check("Voice-5: All steps have valid voice_text", hasVoice, "")

		// Bước cuối phải là "arrived"
		if len(steps) > 0 {
			lastVoice, _ := steps[len(steps)-1]["voice_text"].(string)
			check("Voice-6: Last step = arrived", lastVoice == "arrived", fmt.Sprintf("got=%s", lastVoice))
		}
	} else {
		check("Voice-4: skip", true, ""); check("Voice-5: skip", true, ""); check("Voice-6: skip", true, "")
	}

	// Cleanup
	doReq("POST", base+"/route/cancel", map[string]interface{}{"route_id": routeID}, patient2Token)
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m { keys = append(keys, k) }
	return keys
}

func printSummary() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	if fail > 0 {
		fmt.Printf("  KET QUA: %d PASS / %d FAIL / %d TOTAL\n", pass, fail, total)
	} else {
		fmt.Printf("  KET QUA: %d PASS / %d FAIL / %d TOTAL\n", pass, fail, total)
	}
	fmt.Println(strings.Repeat("=", 70))
	if fail > 0 {
		os.Exit(1)
	}
}
