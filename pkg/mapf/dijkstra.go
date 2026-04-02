package mapf

import (
	"container/heap"
	"math"
)

// ========================================
// DIJKSTRA  - Tim duong tren Grid 2D
// 4-direction (N, S, E, W), weight = 1.0
// ========================================

// PathResult ket qua tim duong.
type PathResult struct {
	Found    bool
	Path     []Position  // danh sach cac vi tri
	Distance float64     // tong khoang cach
}

// Position 1 diem tren grid.
type Position struct {
	Row      int
	Col      int
	Location int // row*cols + col
}

// --- Priority Queue (min-heap) cho Dijkstra ---

type dijkstraItem struct {
	location int
	dist     float64
	index    int // index trong heap
}

type priorityQueue []*dijkstraItem

func (pq priorityQueue) Len() int            { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool  { return pq[i].dist < pq[j].dist }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*dijkstraItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// 4-direction neighbors: N, S, E, W
var directions = [4][2]int{
	{-1, 0}, // North (len)
	{1, 0},  // South (xuong)
	{0, 1},  // East  (phai)
	{0, -1}, // West  (trai)
}

// Dijkstra tim duong ngan nhat tren grid tu startLoc den destLoc.
// weight luan la 1.0 cho moi buoc di.
// Tra ve PathResult gom danh sach Position va tong khoang cach.
func Dijkstra(grid *GridMap, startLoc, destLoc int) PathResult {
	if grid == nil {
		return PathResult{Found: false}
	}

	totalCells := grid.Rows * grid.Cols
	if startLoc < 0 || startLoc >= totalCells || destLoc < 0 || destLoc >= totalCells {
		return PathResult{Found: false}
	}

	startRow, startCol := grid.ToRowCol(startLoc)
	destRow, destCol := grid.ToRowCol(destLoc)

	if !grid.IsWalkable(startRow, startCol) || !grid.IsWalkable(destRow, destCol) {
		return PathResult{Found: false}
	}

	// Dijkstra
	dist := make([]float64, totalCells)
	prev := make([]int, totalCells)
	for i := range dist {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}
	dist[startLoc] = 0

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &dijkstraItem{location: startLoc, dist: 0})

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*dijkstraItem)

		if cur.location == destLoc {
			break
		}

		if cur.dist > dist[cur.location] {
			continue // outdated entry
		}

		r, c := grid.ToRowCol(cur.location)

		for _, d := range directions {
			nr, nc := r+d[0], c+d[1]
			if !grid.IsWalkable(nr, nc) {
				continue
			}

			nLoc := grid.ToLocation(nr, nc)
			newDist := dist[cur.location] + 1.0

			if newDist < dist[nLoc] {
				dist[nLoc] = newDist
				prev[nLoc] = cur.location
				heap.Push(pq, &dijkstraItem{location: nLoc, dist: newDist})
			}
		}
	}

	// Khong tim thay duong
	if math.IsInf(dist[destLoc], 1) {
		return PathResult{Found: false}
	}

	// Trace back path
	path := []Position{}
	for loc := destLoc; loc != -1; loc = prev[loc] {
		r, c := grid.ToRowCol(loc)
		path = append(path, Position{Row: r, Col: c, Location: loc})
	}

	// Reverse path (tu start -> dest)
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return PathResult{
		Found:    true,
		Path:     path,
		Distance: dist[destLoc],
	}
}

// DijkstraWithSpeed tinh estimated_time dua tren speed_factor.
// time = distance / speed_factor (moi step = 1 don vi khoang cach).
func DijkstraWithSpeed(grid *GridMap, startLoc, destLoc int, speedFactor float64) PathResult {
	result := Dijkstra(grid, startLoc, destLoc)
	if result.Found && speedFactor > 0 {
		result.Distance = result.Distance / speedFactor
	}
	return result
}
