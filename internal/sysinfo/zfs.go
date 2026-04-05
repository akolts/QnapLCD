package sysinfo

import (
	"fmt"
	"os/exec"
	"strings"
)

// PoolInfo holds information about a ZFS pool.
type PoolInfo struct {
	Name   string
	Size   string
	Alloc  string
	Health string
}

// Line1 returns the first LCD line for this pool: "name (HEALTH)".
func (p PoolInfo) Line1() string {
	return fmt.Sprintf("%s (%s)", p.Name, p.Health)
}

// Line2 returns the second LCD line for this pool: "ALLOC of SIZE".
func (p PoolInfo) Line2() string {
	return fmt.Sprintf("%s of %s", p.Alloc, p.Size)
}

// ZFSPools runs "zpool list" and returns information about all pools.
// Returns an error if the zpool command is not available.
func ZFSPools() ([]PoolInfo, error) {
	out, err := exec.Command("zpool", "list", "-H", "-o", "name,size,alloc,health").Output()
	if err != nil {
		return nil, err
	}
	return parseZPoolList(string(out))
}

// parseZPoolList parses tab-separated output from
// "zpool list -H -o name,size,alloc,health".
func parseZPoolList(output string) ([]PoolInfo, error) {
	var pools []PoolInfo
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			return nil, fmt.Errorf("unexpected zpool output: %q", line)
		}
		pools = append(pools, PoolInfo{
			Name:   fields[0],
			Size:   fields[1],
			Alloc:  fields[2],
			Health: fields[3],
		})
	}
	return pools, nil
}
