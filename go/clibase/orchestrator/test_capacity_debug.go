// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	// Get system info
	cpuCount := orchestrator.GetEffectiveCPUs()
	effectiveMem := orchestrator.GetEffectiveMemoryBytes()

	var ramGB int
	if effectiveMem > 0 {
		ramGB = int(effectiveMem / (1024 * 1024 * 1024))
	} else {
		memInfo, _ := mem.VirtualMemory()
		ramGB = int(memInfo.Available / (1024 * 1024 * 1024))
	}

	fmt.Fprintf(os.Stderr, "System Detection:\n")
	fmt.Fprintf(os.Stderr, "  GetEffectiveCPUs() = %d\n", cpuCount)
	fmt.Fprintf(os.Stderr, "  GetEffectiveMemoryBytes() = %d bytes (%.2f GB)\n", effectiveMem, float64(effectiveMem)/(1024*1024*1024))
	fmt.Fprintf(os.Stderr, "  Calculated ramGB = %d GB\n", ramGB)
	fmt.Fprintf(os.Stderr, "\nTesting capacity calculations:\n")

	// Test different scenarios
	scenarios := []struct {
		configMax int
		turbo     float64
		desc      string
	}{
		{0, 1.0, "Auto-detect, no turbo"},
		{0, 1.25, "Auto-detect, turbo 1.25x"},
		{0, 2.0, "Auto-detect, turbo 2.0x"},
		{0, 3.2, "Auto-detect, turbo 3.2x"},
		{5, 1.0, "Explicit roof=5"},
		{16, 1.0, "Explicit roof=16"},
	}

	for _, s := range scenarios {
		fmt.Fprintf(os.Stderr, "\n%s (configMax=%d, turbo=%.2f):\n", s.desc, s.configMax, s.turbo)
		// This will trigger our debug logging
		capacity := calculateCapacityTest(cpuCount, ramGB, s.configMax, s.turbo)
		fmt.Fprintf(os.Stderr, "  → Final capacity = %d\n", capacity)
	}
}

func calculateCapacityTest(cpuCount, ramGB, roof int, turbo float64) int {
	// --roof overrides all
	if roof > 0 {
		fmt.Fprintf(os.Stderr, "  [Using explicit roof: %d]\n", roof)
		return roof
	}

	// Auto-detect: min(CPU, RAM/3) × turbo
	ramCap := ramGB / 3
	if ramCap < 1 {
		ramCap = 1
	}

	base := cpuCount
	if ramCap < base {
		fmt.Fprintf(os.Stderr, "  [RAM-limited: ramCap=%d < cpuCount=%d, using ramCap]\n", ramCap, cpuCount)
		base = ramCap
	} else {
		fmt.Fprintf(os.Stderr, "  [CPU-limited: ramCap=%d ≥ cpuCount=%d, using cpuCount]\n", ramCap, cpuCount)
	}

	capacity := int(float64(base) * turbo)
	fmt.Fprintf(os.Stderr, "  [Base capacity=%d × turbo=%.2f = %d]\n", base, turbo, capacity)

	// Cap at RAM limit FIRST (safety), then CPU limit
	ramMax := ramGB / 3
	if capacity > ramMax && ramMax > 0 {
		fmt.Fprintf(os.Stderr, "  [Turbo exceeded RAM limit: %d → %d (ramGB=%d)]\n", capacity, ramMax, ramGB)
		capacity = ramMax
	}

	// Cap at CPU count (or 2x with turbo, max 64)
	maxCap := cpuCount
	if turbo > 1.0 {
		maxCap = cpuCount * 2
		if maxCap > 64 {
			maxCap = 64
		}
	}
	if capacity > maxCap {
		capacity = maxCap
	}

	if capacity < 1 {
		capacity = 1
	}

	return capacity
}
