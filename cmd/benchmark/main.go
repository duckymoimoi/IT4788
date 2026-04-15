package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var base = "http://localhost:8080/api"

type BenchConfig struct {
	Name            string
	Method          string
	URL             string
	Body            interface{}
	Token           string
	Concurrency     int
	TotalRequests   int
	TargetLatencyMs float64
	Timeout         time.Duration
}

type BenchResult struct {
	Name          string
	TotalReqs     int
	SuccessReqs   int
	FailReqs      int
	TotalDuration time.Duration
	Latencies     []time.Duration
	RPS           float64
	P50, P95, P99 time.Duration
	Min, Max, Avg time.Duration
	TargetMs      float64
	Pass          bool
}

type apiResp struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func login(phone, password string) string {
	body, _ := json.Marshal(map[string]string{"phone_number": phone, "password": password})
	resp, err := http.Post(base+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r apiResp
	json.Unmarshal(b, &r)
	if r.Code != 1000 {
		return ""
	}
	var d map[string]interface{}
	json.Unmarshal(r.Data, &d)
	token, _ := d["token"].(string)
	return token
}

func doRequest(client *http.Client, method, url string, body interface{}, token string) (int, time.Duration, error) {
	start := time.Now()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	return resp.StatusCode, elapsed, nil
}

func runBenchmark(cfg BenchConfig) BenchResult {
	var (
		mu        sync.Mutex
		wg        sync.WaitGroup
		latencies = make([]time.Duration, 0, cfg.TotalRequests)
		success   int64
		fail      int64
	)
	sem := make(chan struct{}, cfg.Concurrency)
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Concurrency * 2,
			MaxIdleConnsPerHost: cfg.Concurrency * 2,
			MaxConnsPerHost:     cfg.Concurrency * 2,
		},
	}
	totalStart := time.Now()
	for i := 0; i < cfg.TotalRequests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			_, elapsed, err := doRequest(client, cfg.Method, cfg.URL, cfg.Body, cfg.Token)
			if err != nil {
				atomic.AddInt64(&fail, 1)
			} else {
				atomic.AddInt64(&success, 1)
			}
			mu.Lock()
			latencies = append(latencies, elapsed)
			mu.Unlock()
		}()
	}
	wg.Wait()
	totalDuration := time.Since(totalStart)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	result := BenchResult{
		Name: cfg.Name, TotalReqs: cfg.TotalRequests,
		SuccessReqs: int(success), FailReqs: int(fail),
		TotalDuration: totalDuration, Latencies: latencies, TargetMs: cfg.TargetLatencyMs,
	}
	n := len(latencies)
	if n > 0 {
		result.Min = latencies[0]
		result.Max = latencies[n-1]
		result.P50 = latencies[int(float64(n)*0.50)]
		result.P95 = latencies[int(math.Min(float64(n)*0.95, float64(n-1)))]
		result.P99 = latencies[int(math.Min(float64(n)*0.99, float64(n-1)))]
		var total time.Duration
		for _, l := range latencies {
			total += l
		}
		result.Avg = total / time.Duration(n)
	}
	result.RPS = float64(cfg.TotalRequests) / totalDuration.Seconds()
	result.Pass = result.P95.Seconds()*1000 <= cfg.TargetLatencyMs && result.FailReqs == 0
	return result
}

func printResult(r BenchResult) {
	status := "✅ PASS"
	if !r.Pass {
		status = "❌ FAIL"
	}
	fmt.Printf("\n  ┌─── %s %s\n", r.Name, status)
	fmt.Printf("  │  Requests:    %d total, %d success, %d fail\n", r.TotalReqs, r.SuccessReqs, r.FailReqs)
	fmt.Printf("  │  Duration:    %s\n", r.TotalDuration.Round(time.Millisecond))
	fmt.Printf("  │  RPS:         %.0f req/s\n", r.RPS)
	fmt.Printf("  │    Min: %s  Avg: %s  P50: %s\n", r.Min.Round(time.Microsecond), r.Avg.Round(time.Microsecond), r.P50.Round(time.Microsecond))
	fmt.Printf("  │    P95: %s  P99: %s  Max: %s  (target: ≤%.0fms)\n", r.P95.Round(time.Microsecond), r.P99.Round(time.Microsecond), r.Max.Round(time.Microsecond), r.TargetMs)
	fmt.Printf("  └───\n")
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  HOSPITAL BENCHMARK - Docker + PostgreSQL")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Print("\n  Logging in...")
	patientToken := login("0900000004", "password123")
	if patientToken == "" {
		fmt.Println(" FAILED")
		os.Exit(1)
	}
	fmt.Println(" OK")

	var results []BenchResult

	// PHASE 1: API Latency
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("  PHASE 1: API LATENCY (target: p95 < 500ms)")
	fmt.Println(strings.Repeat("-", 70))
	for _, cfg := range []BenchConfig{
		{"GET /util/faq", "GET", base + "/util/faq", nil, "", 50, 1000, 500, 0},
		{"GET /util/pharmacy", "GET", base + "/util/pharmacy", nil, "", 50, 1000, 500, 0},
		{"GET /sys/get_voice_files", "GET", base + "/sys/get_voice_files", nil, "", 50, 1000, 500, 0},
		{"GET /medical/get_tasks", "GET", base + "/medical/get_tasks", nil, patientToken, 50, 1000, 500, 0},
		{"GET /notification/get_list", "GET", base + "/notification/get_list", nil, patientToken, 50, 1000, 500, 0},
	} {
		fmt.Printf("\n  Running: %s ...", cfg.Name)
		r := runBenchmark(cfg)
		results = append(results, r)
		printResult(r)
	}

	// PHASE 2: Route Finding
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("  PHASE 2: ROUTE FINDING (target: p95 < 3000ms)")
	fmt.Println(strings.Repeat("-", 70))
	cfg := BenchConfig{
		"POST /route/preview", "POST", base + "/route/preview",
		map[string]interface{}{"start_location": 232, "dest_location": 900, "mode_id": "walking"},
		patientToken, 20, 200, 3000, 0,
	}
	fmt.Printf("\n  Running: %s ...", cfg.Name)
	r := runBenchmark(cfg)
	results = append(results, r)
	printResult(r)

	// PHASE 3: Stress Test
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("  PHASE 3: CONCURRENT STRESS TEST")
	fmt.Println(strings.Repeat("-", 70))
	for _, c := range []int{100, 500, 1000, 2000} {
		cfg := BenchConfig{
			Name: fmt.Sprintf("GET /util/faq @ %d conc", c), Method: "GET",
			URL: base + "/util/faq", Concurrency: c, TotalRequests: c * 3,
			TargetLatencyMs: 1000, Timeout: 15 * time.Second,
		}
		fmt.Printf("\n  Running: %d concurrent ...", c)
		r := runBenchmark(cfg)
		results = append(results, r)
		printResult(r)
		if float64(r.FailReqs)/float64(r.TotalReqs) > 0.1 {
			fmt.Println("  ⚠️  Fail rate > 10%, stopping")
			break
		}
	}

	// SUMMARY
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  BENCHMARK SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	passCount, failCount := 0, 0
	fmt.Printf("\n  %-40s %10s %8s %8s %s\n", "Test", "P95", "RPS", "Target", "")
	fmt.Println("  " + strings.Repeat("-", 75))
	for _, r := range results {
		s := "✅"
		if r.Pass {
			passCount++
		} else {
			failCount++
			s = "❌"
		}
		fmt.Printf("  %-40s %10s %8.0f %6.0fms  %s\n", r.Name, r.P95.Round(time.Microsecond), r.RPS, r.TargetMs, s)
	}
	fmt.Println("\n" + strings.Repeat("=", 70))
	if failCount == 0 {
		fmt.Printf("  ALL %d BENCHMARKS PASSED ✅\n", passCount)
	} else {
		fmt.Printf("  %d PASS / %d FAIL / %d TOTAL\n", passCount, failCount, passCount+failCount)
	}
	fmt.Println(strings.Repeat("=", 70))
}
