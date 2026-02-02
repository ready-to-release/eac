// Package orchestrator provides parallel execution with dynamic capacity.
//
// # Dynamic Capacity System
//
// The capacity (max concurrent builds) is calculated dynamically based on
// available RAM, detected from the appropriate source:
//
//	Priority: Docker → WSL → Host
//
// # Formula
//
//	Capacity = (Available RAM / 256MB) × turbo
//
// Examples with 6GB RAM:
//   - Normal (turbo=1): 22 slots
//   - Turbo 2x: 44 slots
//   - Turbo 4x: 64 slots (capped)
//
// # Memory Detection
//
// On Windows with WSL2/Docker, the host "available RAM" is misleading because
// WSL2 reserves memory from Windows. We query the appropriate memory pool:
//
//   - Docker: `docker info --format {{.MemTotal}}`
//   - WSL: `wsl -e cat /proc/meminfo`
//   - Host: gopsutil VirtualMemory().Available
//
// # Memory Instrumentation
//
// Each build logs memory before/after for tuning:
//
//	[memory] before: used=2.89GB avail=2.89GB total=5.78GB (50.0%)
//	[memory] after: used=3.10GB avail=2.68GB total=5.78GB (53.6%) delta=+210MB
package orchestrator

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
)

const (
	// memoryPerSlot is the memory unit for calculating pressure capacity.
	// Based on build log analysis showing ~0MB memory delta per build,
	// containerized builds have negligible memory impact on host.
	// Set to 256MB to allow more parallelism (CPU/IO is the real constraint).
	memoryPerSlot = 256 * 1024 * 1024
)

// GetWSLMemoryBytes returns WSL2 total memory by querying the WSL CLI.
// Returns 0 if not on Windows, WSL is not available, or the command fails.
//
// Uses `wsl -e cat /proc/meminfo` to get actual memory available inside WSL,
// which reflects the WSL2 memory limit (from .wslconfig or defaults).
func GetWSLMemoryBytes() uint64 {
	// Only relevant on Windows
	if runtime.GOOS != "windows" {
		return 0
	}

	ctx, cancel := config.WithDockerQueryContext(context.Background())
	defer cancel()

	// Query /proc/meminfo inside WSL to get total memory
	cmd := exec.CommandContext(ctx, "wsl", "-e", "cat", "/proc/meminfo")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	return parseMeminfo(string(output))
}

// parseMeminfo extracts MemTotal from /proc/meminfo output.
// Format: "MemTotal:        6039696 kB"
func parseMeminfo(output string) uint64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			// Extract the value - format is "MemTotal:        6039696 kB"
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0
				}
				// Convert from kB to bytes
				return kb * 1024
			}
		}
	}
	return 0
}

// GetDockerMemoryBytes returns Docker daemon's total memory via `docker info`.
// Returns 0 if Docker is not available or the command fails.
//
// On Windows with WSL2 backend, this reflects the WSL2 memory limit,
// making it the most accurate source for containerized build capacity.
func GetDockerMemoryBytes() uint64 {
	ctx, cancel := config.WithDockerQueryContext(context.Background())
	defer cancel()

	// Try formatted output first (most efficient)
	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.MemTotal}}")
	output, err := cmd.Output()
	if err == nil {
		memStr := strings.TrimSpace(string(output))
		if memBytes, err := strconv.ParseUint(memStr, 10, 64); err == nil && memBytes > 0 {
			return memBytes
		}
	}

	// Fallback: parse unformatted output (some Docker versions may not support --format well)
	cmd = exec.CommandContext(ctx, "docker", "info")
	output, err = cmd.Output()
	if err != nil {
		return 0
	}

	return parseDockerInfoMemory(string(output))
}

// parseDockerInfoMemory extracts total memory from docker info output.
// Handles format like "Total Memory: 5.762GiB" or "Total Memory: 5762MiB"
func parseDockerInfoMemory(output string) uint64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Total Memory:") {
			// Extract the value after the colon
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			value := strings.TrimSpace(parts[1])
			return parseDockerMemoryValue(value)
		}
	}
	return 0
}

// parseDockerMemoryValue parses Docker memory format like "5.762GiB", "5762MiB"
func parseDockerMemoryValue(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Handle GiB, MiB, TiB, KiB (binary prefixes used by Docker)
	re := regexp.MustCompile(`^([\d.]+)\s*(GiB|MiB|TiB|KiB|GB|MB|TB|KB|G|M|T|K)?$`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0
	}

	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	suffix := ""
	if len(matches) > 2 {
		suffix = strings.ToUpper(matches[2])
	}

	var multiplier float64 = 1
	switch {
	case strings.HasPrefix(suffix, "T"):
		multiplier = 1024 * 1024 * 1024 * 1024
	case strings.HasPrefix(suffix, "G"):
		multiplier = 1024 * 1024 * 1024
	case strings.HasPrefix(suffix, "M"):
		multiplier = 1024 * 1024
	case strings.HasPrefix(suffix, "K"):
		multiplier = 1024
	}

	return uint64(num * multiplier)
}

// MemoryStats holds memory usage information for instrumentation.
type MemoryStats struct {
	TotalBytes     uint64
	AvailableBytes uint64
	UsedBytes      uint64
	UsedPercent    float64
}

// GetMemoryStats returns current memory usage statistics.
// Uses Docker memory if available, falls back to WSL, then host memory.
func GetMemoryStats() MemoryStats {
	// Try to get stats from inside the execution environment
	if stats := getDockerMemoryStats(); stats.TotalBytes > 0 {
		return stats
	}
	if stats := getWSLMemoryStats(); stats.TotalBytes > 0 {
		return stats
	}
	return getHostMemoryStats()
}

// getDockerMemoryStats gets memory stats via docker info.
func getDockerMemoryStats() MemoryStats {
	total := GetDockerMemoryBytes()
	if total == 0 {
		return MemoryStats{}
	}
	// Docker doesn't easily expose available memory, estimate from host
	host := getHostMemoryStats()
	used := host.UsedBytes
	if used > total {
		used = total / 2 // Fallback estimate
	}
	return MemoryStats{
		TotalBytes:     total,
		AvailableBytes: total - used,
		UsedBytes:      used,
		UsedPercent:    float64(used) / float64(total) * 100,
	}
}

// getWSLMemoryStats gets memory stats from WSL /proc/meminfo.
func getWSLMemoryStats() MemoryStats {
	if runtime.GOOS != "windows" {
		return MemoryStats{}
	}

	ctx, cancel := config.WithDockerQueryContext(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "wsl", "-e", "cat", "/proc/meminfo")
	output, err := cmd.Output()
	if err != nil {
		return MemoryStats{}
	}

	return parseMeminfoStats(string(output))
}

// parseMeminfoStats extracts memory stats from /proc/meminfo output.
func parseMeminfoStats(output string) MemoryStats {
	var stats MemoryStats
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		// Values are in kB
		valueBytes := value * 1024

		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			stats.TotalBytes = valueBytes
		case strings.HasPrefix(line, "MemAvailable:"):
			stats.AvailableBytes = valueBytes
		}
	}

	if stats.TotalBytes > 0 {
		stats.UsedBytes = stats.TotalBytes - stats.AvailableBytes
		stats.UsedPercent = float64(stats.UsedBytes) / float64(stats.TotalBytes) * 100
	}
	return stats
}

// getHostMemoryStats gets memory stats from the host system.
func getHostMemoryStats() MemoryStats {
	// Import is at package level via gopsutil in component_scheduler.go
	// We'll use a simple approach here
	ctx, cancel := config.WithDockerQueryContext(context.Background())
	defer cancel()

	// Try platform-specific commands
	if runtime.GOOS == "windows" {
		// Use wmic on Windows
		cmd := exec.CommandContext(ctx, "wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize", "/Value")
		output, err := cmd.Output()
		if err == nil {
			return parseWmicMemory(string(output))
		}
	}

	return MemoryStats{}
}

// parseWmicMemory parses wmic memory output.
func parseWmicMemory(output string) MemoryStats {
	var stats MemoryStats
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TotalVisibleMemorySize=") {
			val := strings.TrimPrefix(line, "TotalVisibleMemorySize=")
			if kb, err := strconv.ParseUint(strings.TrimSpace(val), 10, 64); err == nil {
				stats.TotalBytes = kb * 1024
			}
		} else if strings.HasPrefix(line, "FreePhysicalMemory=") {
			val := strings.TrimPrefix(line, "FreePhysicalMemory=")
			if kb, err := strconv.ParseUint(strings.TrimSpace(val), 10, 64); err == nil {
				stats.AvailableBytes = kb * 1024
			}
		}
	}

	if stats.TotalBytes > 0 {
		stats.UsedBytes = stats.TotalBytes - stats.AvailableBytes
		stats.UsedPercent = float64(stats.UsedBytes) / float64(stats.TotalBytes) * 100
	}
	return stats
}

// FormatBytes formats bytes as a human-readable string.
func FormatBytes(bytes uint64) string {
	const (
		gb = 1024 * 1024 * 1024
		mb = 1024 * 1024
	)
	if bytes >= gb {
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(gb))
	}
	return fmt.Sprintf("%.0fMB", float64(bytes)/float64(mb))
}

// GetDockerCPUs returns the number of CPUs available to Docker.
// Returns 0 if Docker is not available or the command fails.
func GetDockerCPUs() int {
	ctx, cancel := config.WithDockerQueryContext(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.NCPU}}")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	cpus, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0
	}
	return cpus
}

// GetEffectiveCPUs returns the appropriate CPU count for scheduling.
// Tries Docker first, then falls back to runtime.NumCPU().
func GetEffectiveCPUs() int {
	if cpus := GetDockerCPUs(); cpus > 0 {
		return cpus
	}
	return runtime.NumCPU()
}

// GetEffectiveMemoryBytes returns the appropriate memory pool for scheduling.
// Priority:
// 1. Docker memory (most accurate for containerized builds, reflects WSL2 limit)
// 2. WSL memory (queried via WSL CLI)
// 3. 0 (caller should fall back to host available RAM)
func GetEffectiveMemoryBytes() uint64 {
	// Try Docker first - this reflects WSL2 memory on Windows with WSL2 backend
	if mem := GetDockerMemoryBytes(); mem > 0 {
		return mem
	}

	// Try WSL CLI
	if mem := GetWSLMemoryBytes(); mem > 0 {
		return mem
	}

	// Return 0 to signal caller should use host available RAM
	return 0
}
