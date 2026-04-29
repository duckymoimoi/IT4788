package mapf

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ========================================
// AGENT MANAGER  - Mo phong MAPF trong RAM
// Slice 5: Doc output.json, auto-advance timestep
// ========================================

// AgentManager quan ly phien mo phong MAPF trong RAM.
// Doc output.json qua ParseOutputJSON -> auto-advance timestep.
// Thread-safe bang sync.Mutex.
type AgentManager struct {
	mu          sync.Mutex
	result      *MAPFResult
	running     bool
	currentTS   int
	tickRateMs  int
	ticker      *time.Ticker
	stopCh      chan struct{}
	outputFile  string

	// Frequency tracking voi time window tu dong reset.
	// freqMap dem so lan agent xuat hien tai moi grid_location.
	// Tu dong reset sau moi windowDuration de tiet kiem RAM.
	freqMap        map[int]int64
	totalTicks     int64         // tong tick trong window hien tai
	windowStart    time.Time     // thoi diem bat dau window hien tai
	windowDuration time.Duration // do dai moi window (default 5 phut)
}

// NewAgentManager tao AgentManager moi (chua running).
func NewAgentManager() *AgentManager {
	return &AgentManager{}
}

// Start bat dau mo phong tu file output.json.
// tickRateMs la thoi gian giua moi timestep (ms).
// Neu da running, tra loi loi.
func (m *AgentManager) Start(outputFile string, tickRateMs int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("simulation already running")
	}

	// Parse output.json
	result, err := ParseOutputJSON(outputFile)
	if err != nil {
		return fmt.Errorf("cannot parse output file: %w", err)
	}

	if tickRateMs <= 0 {
		tickRateMs = 1000 // default 1 giay
	}

	m.result = result
	m.outputFile = outputFile
	m.currentTS = 0
	m.tickRateMs = tickRateMs
	m.running = true
	m.stopCh = make(chan struct{})
	m.freqMap = make(map[int]int64)
	m.totalTicks = 0
	m.windowStart = time.Now()
	m.windowDuration = 5 * time.Minute // Reset moi 5 phut

	// Auto-advance timestep trong goroutine rieng
	m.ticker = time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	go m.advanceLoop()

	return nil
}

// advanceLoop tu dong tang timestep theo ticker.
// Khi het makespan -> quay lai timestep 0 (loop vo han).
// Chi dung khi bi Stop().
func (m *AgentManager) advanceLoop() {
	for {
		select {
		case <-m.ticker.C:
			m.mu.Lock()
			if m.currentTS < m.result.Makespan {
				m.currentTS++
			} else {
				m.currentTS = 0
			}

			// Auto-reset frequency khi het window de tiet kiem RAM
			if time.Since(m.windowStart) >= m.windowDuration {
				m.freqMap = make(map[int]int64)
				m.totalTicks = 0
				m.windowStart = time.Now()
			}

			// Ghi nhan tan suat trong window hien tai
			positions := m.result.GetPositionsAtTimestep(m.currentTS)
			for _, pos := range positions {
				m.freqMap[pos.Location]++
			}
			m.totalTicks++

			m.mu.Unlock()
		case <-m.stopCh:
			return
		}
	}
}

// Stop dung mo phong.
func (m *AgentManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	if m.ticker != nil {
		m.ticker.Stop()
	}
	close(m.stopCh)
}

// GetAllPositions tra ve vi tri tat ca agents tai timestep hien tai.
func (m *AgentManager) GetAllPositions() []AgentState {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.result == nil {
		return nil
	}

	return m.result.GetPositionsAtTimestep(m.currentTS)
}

// GetPositionsAt tra ve vi tri tat ca agents tai timestep chi dinh.
func (m *AgentManager) GetPositionsAt(timestep int) []AgentState {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.result == nil {
		return nil
	}

	return m.result.GetPositionsAtTimestep(timestep)
}

// IsRunning kiem tra mo phong dang chay khong.
func (m *AgentManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// GetCurrentTimestep tra ve timestep hien tai.
func (m *AgentManager) GetCurrentTimestep() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentTS
}

// GetInfo tra ve thong tin mo phong: teamSize, makespan, currentTS, tickRateMs, outputFile.
func (m *AgentManager) GetInfo() (teamSize, makespan, currentTS, tickRateMs int, outputFile string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.result == nil {
		return 0, 0, 0, 0, ""
	}

	return m.result.TeamSize, m.result.Makespan, m.currentTS, m.tickRateMs, m.outputFile
}

// GetResult tra ve ket qua parse (chi doc, khong sua).
func (m *AgentManager) GetResult() *MAPFResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.result
}

// ========================================
// FREQUENCY TRACKING - Do luong throughput
// ========================================

// FrequencyEntry 1 o tren ban do tan suat.
type FrequencyEntry struct {
	Location  int   `json:"grid_location"`
	Frequency int64 `json:"frequency"`
}

// GetFrequencyMap tra ve ban do tan suat tich luy, sap xep giam dan.
// Moi entry the hien so lan bat ky agent nao da xuat hien tai o do.
func (m *AgentManager) GetFrequencyMap() []FrequencyEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries := make([]FrequencyEntry, 0, len(m.freqMap))
	for loc, freq := range m.freqMap {
		entries = append(entries, FrequencyEntry{Location: loc, Frequency: freq})
	}

	// Sap xep giam dan theo tan suat
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Frequency > entries[j].Frequency
	})

	return entries
}

// GetLocationFrequency tra ve tan suat tai 1 vi tri cu the.
func (m *AgentManager) GetLocationFrequency(gridLocation int) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.freqMap[gridLocation]
}

// GetTotalTicks tra ve tong so tick da chay.
func (m *AgentManager) GetTotalTicks() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalTicks
}

// ResetFrequency xoa ban do tan suat (khong dung simulation).
func (m *AgentManager) ResetFrequency() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.freqMap = make(map[int]int64)
	m.totalTicks = 0
}

// GetWindowDuration tra ve do dai window hien tai.
func (m *AgentManager) GetWindowDuration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.windowDuration
}

// SetWindowDuration thay doi do dai window (ap dung tu window tiep theo).
func (m *AgentManager) SetWindowDuration(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d < 1*time.Minute {
		d = 1 * time.Minute // toi thieu 1 phut
	}
	m.windowDuration = d
}

// GetWindowElapsed tra ve thoi gian da troi tu dau window hien tai.
func (m *AgentManager) GetWindowElapsed() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return time.Since(m.windowStart)
}
