package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var base = "http://localhost:8080/api"

type apiResp struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func login() string {
	body, _ := json.Marshal(map[string]string{"phone_number": "0900000001", "password": "password123"})
	resp, _ := http.Post(base+"/auth/login", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r apiResp
	json.Unmarshal(b, &r)
	var d map[string]interface{}
	json.Unmarshal(r.Data, &d)
	return d["token"].(string)
}

type AgentPos struct {
	AgentID  int `json:"agent_id"`
	Location int `json:"location"`
	Row      int `json:"row"`
	Col      int `json:"col"`
}

type SimStatus struct {
	Running   bool       `json:"running"`
	TeamSize  int        `json:"team_size"`
	Makespan  int        `json:"makespan"`
	CurrentTS int        `json:"current_timestep"`
	TickRate  int        `json:"tick_rate_ms"`
	Positions []AgentPos `json:"positions"`
}

func getStatus(token string) *SimStatus {
	req, _ := http.NewRequest("GET", base+"/simulate/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r apiResp
	json.Unmarshal(b, &r)
	var s SimStatus
	json.Unmarshal(r.Data, &s)
	return &s
}

func main() {
	fmt.Println("=== AGENT MOVEMENT TRACKER (1 phút) ===")
	fmt.Println()

	token := login()
	fmt.Println("Logged in as admin")
	fmt.Println()

	// Track agent 0 over 1 minute
	startTime := time.Now()
	pollInterval := 2 * time.Second
	lastTS := -1
	pollCount := 0

	fmt.Printf("%-6s  %-4s  %-12s  %-6s  %-6s  %-6s  %s\n",
		"Time", "T", "Agent 0 Pos", "Row", "Col", "Running", "All Agents Locations")
	fmt.Println("------  ----  ------------  ------  ------  -------  --------------------")

	for time.Since(startTime) < 62*time.Second {
		s := getStatus(token)
		if s == nil {
			fmt.Println("ERROR: cannot get status")
			time.Sleep(pollInterval)
			continue
		}

		elapsed := time.Since(startTime).Seconds()
		pollCount++

		// Chi in khi timestep thay doi
		if s.CurrentTS != lastTS {
			// Agent 0 info
			var a0 AgentPos
			allLocs := ""
			for i, p := range s.Positions {
				if p.AgentID == 0 {
					a0 = p
				}
				if i > 0 {
					allLocs += ", "
				}
				allLocs += fmt.Sprintf("A%d@(%d,%d)", p.AgentID, p.Row, p.Col)
			}

			loopMarker := ""
			if lastTS > s.CurrentTS && lastTS > 0 {
				loopMarker = " ↩ LOOP!"
			}

			fmt.Printf("%-6.1fs  t=%-2d  loc=%-8d  r=%-4d  c=%-4d  %-7v  %s%s\n",
				elapsed, s.CurrentTS, a0.Location, a0.Row, a0.Col, s.Running, allLocs, loopMarker)

			lastTS = s.CurrentTS
		}

		time.Sleep(pollInterval)
	}

	fmt.Println()
	fmt.Printf("=== Ket thuc: %d polls, last timestep: %d ===\n", pollCount, lastTS)
}
