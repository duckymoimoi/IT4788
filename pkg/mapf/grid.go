package mapf

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GridMap biểu diễn bản đồ grid 2D từ file .map (MovingAI format).
// Grid[row][col] = 0 nếu walkable, 1 nếu obstacle.
type GridMap struct {
	Name string
	Rows int
	Cols int
	Grid [][]int
}

// LoadGridMap đọc file .map (MovingAI format) và trả về GridMap.
//
// Format:
//
//	type octile
//	height 140
//	width 500
//	map
//	@@@@....@@@....   (@ = obstacle, . = walkable, E/S = walkable special)
func LoadGridMap(filepath string) (*GridMap, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("cannot open map file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Tăng buffer size cho map rộng (warehouse_large: 500 cols)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var rows, cols int
	var grid [][]int

	phase := "header" // header | map

	for scanner.Scan() {
		line := scanner.Text()

		if phase == "header" {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "height") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					rows, _ = strconv.Atoi(parts[1])
				}
			} else if strings.HasPrefix(line, "width") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					cols, _ = strconv.Atoi(parts[1])
				}
			} else if line == "map" {
				phase = "map"
				grid = make([][]int, 0, rows)
			}
			continue
		}

		// Phase: map  - đọc từng dòng grid
		if len(grid) >= rows {
			break
		}
		row := make([]int, cols)
		for c := 0; c < cols && c < len(line); c++ {
			ch := line[c]
			switch ch {
			case '.', 'E', 'S':
				row[c] = 0 // walkable
			default:
				row[c] = 1 // obstacle (@, T, hoặc bất kỳ)
			}
		}
		grid = append(grid, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading map file: %w", err)
	}

	if len(grid) != rows {
		return nil, fmt.Errorf("expected %d rows, got %d", rows, len(grid))
	}

	return &GridMap{
		Name: filepath,
		Rows: rows,
		Cols: cols,
		Grid: grid,
	}, nil
}

// IsWalkable kiểm tra 1 ô có đi được không.
func (g *GridMap) IsWalkable(row, col int) bool {
	if row < 0 || row >= g.Rows || col < 0 || col >= g.Cols {
		return false
	}
	return g.Grid[row][col] == 0
}

// ToLocation chuyển (row, col) -> location index.
func (g *GridMap) ToLocation(row, col int) int {
	return row*g.Cols + col
}

// ToRowCol chuyển location index -> (row, col).
func (g *GridMap) ToRowCol(location int) (int, int) {
	return location / g.Cols, location % g.Cols
}

// GridDataToJSON chuyển grid thành JSON string cho lưu DB.
// Trả về compact format: "[[0,1,0],[1,0,0],...]"
func (g *GridMap) GridDataToJSON() string {
	var sb strings.Builder
	sb.WriteString("[")
	for r, row := range g.Grid {
		if r > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("[")
		for c, cell := range row {
			if c > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(strconv.Itoa(cell))
		}
		sb.WriteString("]")
	}
	sb.WriteString("]")
	return sb.String()
}

// CountWalkable đếm ô đi được trên grid.
func (g *GridMap) CountWalkable() int {
	count := 0
	for _, row := range g.Grid {
		for _, cell := range row {
			if cell == 0 {
				count++
			}
		}
	}
	return count
}
