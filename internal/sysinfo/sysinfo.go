// Package sysinfo gathers system information for display on the QNAP LCD.
// It prefers reading from /proc and using stdlib over executing external commands.
package sysinfo

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Hostname returns the system hostname.
func Hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

// OSInfo returns a string like "linux (amd64)".
func OSInfo() string {
	return fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
}

// Uptime reads /proc/uptime and returns a human-readable duration string.
func Uptime() (string, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "", err
	}
	return parseUptime(string(data))
}

// LoadAvg reads /proc/loadavg and returns the 1/5/15 minute averages as a string.
func LoadAvg() (string, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "", err
	}
	return parseLoadAvg(string(data))
}

// TrueNASVersion runs the TrueNAS CLI to get the system version.
// Returns two strings suitable for LCD line 1 and line 2.
// Returns an error if the cli command is not available.
func TrueNASVersion() (line1, line2 string, err error) {
	out, err := exec.Command("cli", "-c", "system version").Output()
	if err != nil {
		return "", "", err
	}
	return parseTrueNASVersion(strings.TrimSpace(string(out)))
}

// parseUptime parses /proc/uptime content ("12345.67 23456.78") and
// returns a formatted duration like "1d 5h 30m".
func parseUptime(content string) (string, error) {
	var uptimeSecs, idleSecs float64
	_, err := fmt.Sscanf(content, "%f %f", &uptimeSecs, &idleSecs)
	if err != nil {
		return "", fmt.Errorf("parse uptime: %w", err)
	}
	return formatDuration(int(uptimeSecs)), nil
}

// parseLoadAvg parses /proc/loadavg content ("0.05 0.10 0.15 1/234 5678")
// and returns the three load averages as a space-separated string.
func parseLoadAvg(content string) (string, error) {
	fields := strings.Fields(content)
	if len(fields) < 3 {
		return "", fmt.Errorf("parse loadavg: unexpected format: %q", content)
	}
	return fields[0] + " " + fields[1] + " " + fields[2], nil
}

// parseTrueNASVersion parses the output of "cli -c 'system version'"
// (e.g., "TrueNAS-SCALE-25.10.1") into two display lines.
func parseTrueNASVersion(output string) (string, string, error) {
	if output == "" {
		return "", "", fmt.Errorf("empty version string")
	}
	parts := strings.Split(output, "-")
	if len(parts) < 2 {
		return output, "", nil
	}
	line1 := strings.Join(parts[:len(parts)-1], "-")
	line2 := parts[len(parts)-1]
	return line1, line2, nil
}

// formatDuration converts seconds into a compact human-readable string.
func formatDuration(totalSecs int) string {
	days := totalSecs / 86400
	hours := (totalSecs % 86400) / 3600
	mins := (totalSecs % 3600) / 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}
