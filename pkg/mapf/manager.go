package mapf

import (
	"fmt"
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

	// Auto-advance timestep trong goroutine rieng
	m.ticker = time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	go m.advanceLoop()

	return nil
}

// advanceLoop tu dong tang timestep theo ticker.
// Dung khi het makespan hoac bi Stop().
func (m *AgentManager) advanceLoop() {
	for {
		select {
		case <-m.ticker.C:
			m.mu.Lock()
			if m.currentTS < m.result.Makespan {
				m.currentTS++
			} else {
				// Mo phong hoan thanh
				m.running = false
				m.ticker.Stop()
				m.mu.Unlock()
				return
			}
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
