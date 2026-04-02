package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
	fmt.Println("  PART 5: MAP APIs (2)")
	fmt.Println(strings.Repeat("-", 50))

	r, _ := doReq("GET", base+"/map/get_floors", nil, "")
	check("[16] GET get_floors", r != nil && r.Code == 1000, "")

	r, _ = doReq("GET", base+"/map/get_nodes?map_id=0", nil, "")
	check("[17] GET get_nodes", r != nil && r.Code == 1000, "")
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
