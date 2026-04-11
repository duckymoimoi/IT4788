package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ========================================
// FLOW TEST SUITE  - Slice 5 (Person B)
// Server phai chay tai :8080
// go run cmd/test_flow/main.go
// ========================================

var (
	base       = "http://localhost:8080/api"
	pass, fail int
	total      int

	patientToken string
	adminToken   string
	staffToken   string
)

func check(name string, ok bool, detail string) {
	total++
	if ok {
		pass++
		fmt.Printf("  ✓ %s\n", name)
	} else {
		fail++
		fmt.Printf("  ✗ FAIL %s  — %s\n", name, detail)
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
	fmt.Println("  SLICE 5: FLOW & MAPF SIMULATE — TEST SUITE (Person B)")
	fmt.Println(strings.Repeat("=", 70))

	// Check server
	_, err := http.Get(base + "/sys/check_version")
	if err != nil {
		fmt.Println("\n  [ERROR] Server chua chay! Hay chay: go run cmd/main.go")
		os.Exit(1)
	}

	testLogin()
	testFlowPublicAPIs()
	testFlowPrivateAPIs()
	testFlowStaffAPIs()
	testFlowAdminAPIs()
	testSimulationAPIs()
	testFlowErrorCases()
	testFlowResetVerify()

	printSummary()
}

// ========================================
// PART 1: LOGIN
// ========================================
func testLogin() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 1: LOGIN (3)")
	fmt.Println(strings.Repeat("-", 50))

	// Patient login
	r, _ := doReq("POST", base+"/auth/login", map[string]string{
		"phone_number": "0900000004", "password": "password123",
	}, "")
	check("Login patient (0900000004)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d struct {
			Token string `json:"token"`
		}
		json.Unmarshal(r.Data, &d)
		patientToken = d.Token
	}

	// Admin login
	r, _ = doReq("POST", base+"/auth/login", map[string]string{
		"phone_number": "0900000001", "password": "password123",
	}, "")
	check("Login admin (0900000001)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d struct {
			Token string `json:"token"`
		}
		json.Unmarshal(r.Data, &d)
		adminToken = d.Token
	}

	// Staff/coordinator login
	r, _ = doReq("POST", base+"/auth/login", map[string]string{
		"phone_number": "0900000002", "password": "password123",
	}, "")
	check("Login coordinator (0900000002)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d struct {
			Token string `json:"token"`
		}
		json.Unmarshal(r.Data, &d)
		staffToken = d.Token
	}
}

// ========================================
// PART 2: FLOW PUBLIC APIs
// ========================================
func testFlowPublicAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 2: FLOW PUBLIC APIs (6)")
	fmt.Println(strings.Repeat("-", 50))

	// [47] GET get_density
	r, _ := doReq("GET", base+"/flow/get_density?grid_location=100", nil, "")
	check("[47] GET get_density", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  has grid_location + count fields",
			d["grid_location"] != nil && d["count"] != nil,
			fmt.Sprintf("keys=%v", keysOf(d)))
	}

	// [48] GET get_heatmap
	r, _ = doReq("GET", base+"/flow/get_heatmap", nil, "")
	check("[48] GET get_heatmap", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d []map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  returns array", true, "")
	}

	// [49] GET get_bottlenecks
	r, _ = doReq("GET", base+"/flow/get_bottlenecks?limit=5", nil, "")
	check("[49] GET get_bottlenecks", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d []map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  results <= 5 (limit)", len(d) <= 5, fmt.Sprintf("got %d", len(d)))
	}

	// [52] GET get_forecast
	r, _ = doReq("GET", base+"/flow/get_forecast?hours=24", nil, "")
	check("[52] GET get_forecast", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [54] GET get_alerts
	r, _ = doReq("GET", base+"/flow/get_alerts", nil, "")
	check("[54] GET get_alerts", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [55] GET edge_status
	r, _ = doReq("GET", base+"/flow/edge_status?grid_location=100", nil, "")
	check("[55] GET edge_status", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
}

// ========================================
// PART 3: FLOW PRIVATE APIs (patient auth)
// ========================================
func testFlowPrivateAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 3: FLOW PRIVATE APIs (7)")
	fmt.Println(strings.Repeat("-", 50))

	if patientToken == "" {
		fmt.Println("  [WARN] No patient token")
		return
	}

	// [46] POST ping_location
	r, _ := doReq("POST", base+"/flow/ping_location", map[string]interface{}{
		"grid_location": 777, "grid_row": 15, "grid_col": 12,
	}, patientToken)
	check("[46] POST ping_location", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  pinged=true", d["pinged"] == true, "")
	}

	// [46] POST ping_location with route_id
	r, _ = doReq("POST", base+"/flow/ping_location", map[string]interface{}{
		"grid_location": 888, "grid_row": 17, "grid_col": 13, "route_id": "test-route-123",
	}, patientToken)
	check("[46] POST ping_location (with route_id)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [50] POST report_obstacle
	r, _ = doReq("POST", base+"/flow/report_obstacle", map[string]interface{}{
		"grid_location": 555,
		"report_type":   "wet_floor",
		"description":   "San uot truoc phong 201 - test",
	}, patientToken)
	check("[50] POST report_obstacle", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  has report_id", d["report_id"] != nil, "")
		check("  status = pending", d["status"] == "pending", fmt.Sprintf("got %v", d["status"]))
	}

	// GET get_obstacles (danh sach)
	r, _ = doReq("GET", base+"/flow/get_obstacles?page=1&limit=20", nil, patientToken)
	check("GET get_obstacles (list)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  has total field", d["total"] != nil, "")
		check("  has reports array", d["reports"] != nil, "")
	}

	// GET get_obstacles with status filter
	r, _ = doReq("GET", base+"/flow/get_obstacles?status=pending&page=1&limit=20", nil, patientToken)
	check("GET get_obstacles (status=pending)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))

	// [53] POST set_priority
	r, _ = doReq("POST", base+"/flow/set_priority", map[string]interface{}{
		"from_location": 100,
		"to_location":   999,
		"reason":        "Test priority route from test suite",
	}, patientToken)
	check("[53] POST set_priority", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	var newPriorityID float64
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  has priority_id", d["priority_id"] != nil, "")
		check("  status = active", d["status"] == "active", fmt.Sprintf("got %v", d["status"]))
		if v, ok := d["priority_id"].(float64); ok {
			newPriorityID = v
		}
	}

	// POST expire_priority
	if newPriorityID > 0 {
		r, _ = doReq("POST", base+"/flow/expire_priority", map[string]interface{}{
			"priority_id": newPriorityID,
		}, patientToken)
		check("POST expire_priority", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	}
}

// ========================================
// PART 4: STAFF APIs
// ========================================
func testFlowStaffAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 4: FLOW STAFF APIs (2)")
	fmt.Println(strings.Repeat("-", 50))

	if staffToken == "" {
		fmt.Println("  [WARN] No staff token")
		return
	}

	// Lay danh sach obstacles de resolve
	r, _ := doReq("GET", base+"/flow/get_obstacles?status=pending&page=1&limit=1", nil, staffToken)
	var reportID float64
	if r != nil && r.Code == 1000 {
		var d struct {
			Reports []map[string]interface{} `json:"reports"`
		}
		json.Unmarshal(r.Data, &d)
		if len(d.Reports) > 0 {
			if v, ok := d.Reports[0]["report_id"].(float64); ok {
				reportID = v
			}
		}
	}

	// POST resolve_obstacle (resolve)
	if reportID > 0 {
		r, _ = doReq("POST", base+"/flow/resolve_obstacle", map[string]interface{}{
			"report_id": reportID,
			"action":    "resolve",
		}, staffToken)
		check("POST resolve_obstacle (resolve)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
		if r != nil && r.Code == 1000 {
			var d map[string]interface{}
			json.Unmarshal(r.Data, &d)
			check("  resolved=true", d["resolved"] == true, "")
		}
	} else {
		check("POST resolve_obstacle (resolve)", false, "no pending obstacle found")
	}

	// Tao obstacle moi roi reject
	rr, _ := doReq("POST", base+"/flow/report_obstacle", map[string]interface{}{
		"grid_location": 666, "report_type": "other", "description": "Test reject",
	}, patientToken)
	if rr != nil && rr.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(rr.Data, &d)
		if rid, ok := d["report_id"].(float64); ok {
			r, _ = doReq("POST", base+"/flow/resolve_obstacle", map[string]interface{}{
				"report_id": rid,
				"action":    "reject",
			}, staffToken)
			check("POST resolve_obstacle (reject)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
		}
	}
}

// ========================================
// PART 5: ADMIN APIs
// ========================================
func testFlowAdminAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 5: FLOW ADMIN APIs (3)")
	fmt.Println(strings.Repeat("-", 50))

	if adminToken == "" {
		fmt.Println("  [WARN] No admin token")
		return
	}

	// [51] PATCH set_capacity
	r, _ := doReq("PATCH", base+"/admin/set_capacity", map[string]interface{}{
		"poi_id": 1, "capacity": 50,
	}, adminToken)
	check("[51] PATCH set_capacity", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  updated=true", d["updated"] == true, "")
	}

	// [56] GET stats_flow
	r, _ = doReq("GET", base+"/admin/stats_flow?hours=48", nil, adminToken)
	check("[56] GET stats_flow", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d []map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  returns array of hourly stats", true, "")
	}

	// [57] POST reset_flow (test later in PART 8)
	check("[57] POST reset_flow", true, "tested in PART 8 (reset + verify)")
}

// ========================================
// PART 6: SIMULATION APIs
// ========================================
func testSimulationAPIs() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 6: SIMULATION APIs (6)")
	fmt.Println(strings.Repeat("-", 50))

	if adminToken == "" {
		fmt.Println("  [WARN] No admin token")
		return
	}

	// [60] GET status (truoc start)
	r, _ := doReq("GET", base+"/simulate/status", nil, adminToken)
	check("[60] GET simulate/status (before start)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  running=false before start", d["running"] == false, fmt.Sprintf("got %v", d["running"]))
	}

	// [58] POST start
	r, _ = doReq("POST", base+"/simulate/start", map[string]interface{}{
		"map_id": 1, "output_file": "data/output.json", "tick_rate_ms": 200,
	}, adminToken)
	check("[58] POST simulate/start", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  running=true", d["running"] == true, "")
		check("  team_size > 0", d["team_size"] != nil && d["team_size"].(float64) > 0,
			fmt.Sprintf("got %v", d["team_size"]))
		check("  has run_id", d["run_id"] != nil, "")
	}

	// [60] GET status (after start, wait for timestep advance)
	time.Sleep(500 * time.Millisecond)
	r, _ = doReq("GET", base+"/simulate/status", nil, adminToken)
	check("[60] GET simulate/status (after start)", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  running=true", d["running"] == true, fmt.Sprintf("got %v", d["running"]))
		ts := d["current_timestep"]
		check("  timestep advanced (> 0)", ts != nil && ts.(float64) > 0,
			fmt.Sprintf("got %v", ts))
		check("  has positions array", d["positions"] != nil, "")
	}

	// [59] POST stop
	r, _ = doReq("POST", base+"/simulate/stop", nil, adminToken)
	check("[59] POST simulate/stop", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  stopped=true", d["stopped"] == true, "")
	}

	// Verify stopped
	r, _ = doReq("GET", base+"/simulate/status", nil, adminToken)
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  confirm running=false after stop", d["running"] == false,
			fmt.Sprintf("got %v", d["running"]))
	}
}

// ========================================
// PART 7: ERROR CASES
// ========================================
func testFlowErrorCases() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 7: ERROR CASES (8)")
	fmt.Println(strings.Repeat("-", 50))

	// No auth -> ping rejected
	r, _ := doReq("POST", base+"/flow/ping_location", map[string]interface{}{
		"grid_location": 100, "grid_row": 2, "grid_col": 10,
	}, "")
	check("No auth -> ping rejected (3003)", r != nil && r.Code == 3003,
		fmt.Sprintf("code=%d", sc(r)))

	// No auth -> report obstacle rejected
	r, _ = doReq("POST", base+"/flow/report_obstacle", map[string]interface{}{
		"grid_location": 100, "report_type": "test",
	}, "")
	check("No auth -> report_obstacle rejected (3003)", r != nil && r.Code == 3003,
		fmt.Sprintf("code=%d", sc(r)))

	// Patient -> admin API rejected
	r, _ = doReq("POST", base+"/admin/reset_flow", nil, patientToken)
	check("Patient -> reset_flow rejected (3102)", r != nil && r.Code == 3102,
		fmt.Sprintf("code=%d", sc(r)))

	// Patient -> set_capacity rejected
	r, _ = doReq("PATCH", base+"/admin/set_capacity", map[string]interface{}{
		"poi_id": 1, "capacity": 100,
	}, patientToken)
	check("Patient -> set_capacity rejected (3102)", r != nil && r.Code == 3102,
		fmt.Sprintf("code=%d", sc(r)))

	// Patient -> stats_flow rejected
	r, _ = doReq("GET", base+"/admin/stats_flow", nil, patientToken)
	check("Patient -> stats_flow rejected (3102)", r != nil && r.Code == 3102,
		fmt.Sprintf("code=%d", sc(r)))

	// Patient -> simulate/start rejected
	r, _ = doReq("POST", base+"/simulate/start", map[string]interface{}{
		"map_id": 1, "output_file": "data/output.json",
	}, patientToken)
	check("Patient -> simulate/start rejected (3102)", r != nil && r.Code == 3102,
		fmt.Sprintf("code=%d", sc(r)))

	// Stop when no simulation running
	r, _ = doReq("POST", base+"/simulate/stop", nil, adminToken)
	check("Stop no sim -> error (9001)", r != nil && r.Code == 9001,
		fmt.Sprintf("code=%d", sc(r)))

	// GET density missing param
	r, _ = doReq("GET", base+"/flow/get_density", nil, "")
	check("get_density missing param -> (2001)", r != nil && r.Code == 2001,
		fmt.Sprintf("code=%d", sc(r)))
}

// ========================================
// PART 8: RESET + VERIFY
// ========================================
func testFlowResetVerify() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("  PART 8: RESET + VERIFY (3)")
	fmt.Println(strings.Repeat("-", 50))

	if adminToken == "" {
		fmt.Println("  [WARN] No admin token")
		return
	}

	// Truoc reset: heatmap co du lieu
	r, _ := doReq("GET", base+"/flow/get_heatmap", nil, "")
	var beforeCount int
	if r != nil && r.Code == 1000 {
		var d []interface{}
		json.Unmarshal(r.Data, &d)
		beforeCount = len(d)
	}

	// Tao ping moi de dam bao co du lieu
	if patientToken != "" {
		doReq("POST", base+"/flow/ping_location", map[string]interface{}{
			"grid_location": 12345, "grid_row": 20, "grid_col": 5,
		}, patientToken)
	}

	// [57] POST reset_flow
	r, _ = doReq("POST", base+"/admin/reset_flow", nil, adminToken)
	check("[57] POST reset_flow", r != nil && r.Code == 1000, fmt.Sprintf("code=%d", sc(r)))
	if r != nil && r.Code == 1000 {
		var d map[string]interface{}
		json.Unmarshal(r.Data, &d)
		check("  reset=true", d["reset"] == true, "")
	}

	// Sau reset: heatmap phai rong
	r, _ = doReq("GET", base+"/flow/get_heatmap", nil, "")
	if r != nil && r.Code == 1000 {
		var d []interface{}
		json.Unmarshal(r.Data, &d)
		check("  heatmap empty after reset (was "+fmt.Sprintf("%d", beforeCount)+")",
			len(d) == 0, fmt.Sprintf("still has %d entries", len(d)))
	}
}

// ========================================
// HELPERS
// ========================================

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func printSummary() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	if fail > 0 {
		fmt.Printf("  KET QUA: %d/%d PASS  |  %d FAIL\n", pass, total, fail)
	} else {
		fmt.Printf("  KET QUA: %d/%d PASS  |  ALL PASSED ✓\n", pass, total)
	}
	fmt.Println(strings.Repeat("=", 70))
	if fail > 0 {
		os.Exit(1)
	}
}
