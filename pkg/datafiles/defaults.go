package datafiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hospital/pkg/mapf"
)

const (
	runtimeDataDir = "data"
	seedDataDir    = "seed_data"
)

// EnsureDefaultDataFile repairs persistent Docker volumes that shadow the
// image's bundled map/simulation files.
func EnsureDefaultDataFile(name string) (string, error) {
	target := filepath.Join(runtimeDataDir, name)
	source := filepath.Join(seedDataDir, name)

	if _, err := os.Stat(source); err != nil {
		return target, nil
	}
	if defaultDataFileOK(target, name) {
		return target, nil
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return target, fmt.Errorf("read seed data %s: %w", source, err)
	}
	if err := os.MkdirAll(runtimeDataDir, 0755); err != nil {
		return target, fmt.Errorf("create data dir: %w", err)
	}
	if err := os.WriteFile(target, data, 0644); err != nil {
		return target, fmt.Errorf("write runtime data %s: %w", target, err)
	}
	return target, nil
}

func defaultDataFileOK(path string, name string) bool {
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		return false
	}

	switch {
	case strings.HasSuffix(name, ".map"):
		grid, err := mapf.LoadGridMap(path)
		return err == nil && grid.Rows > 0 && grid.Cols > 0
	case strings.HasSuffix(name, ".json"):
		result, err := mapf.ParseOutputJSON(path)
		return err == nil && result != nil && result.TeamSize > 0 && result.Makespan > 0
	default:
		return true
	}
}
