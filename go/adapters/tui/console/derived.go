package console

import (
	"fmt"
	"strings"
)

// DerivedCounts contains statistics computed from uowStates.
// This is the single source of truth for progress tracking.
// Call DeriveCounts() in View() rather than maintaining separate counters.
type DerivedCounts struct {
	Total   int // Total number of UoWs registered
	Running int // Sum of weights for running UoWs (weight-aware)
	Done    int // Completed successfully (exit code 0)
	Cached  int // Skipped/cached (exit code < 0)
	Failed  int // Failed (exit code > 0)
	Pending int // Waiting to start
}

// CapacityInfo holds the three-value capacity model for display.
// This separates running count from capacity limits, enabling clear visualization
// of the scheduler's actual behavior under memory pressure.
type CapacityInfo struct {
	// Running is the sum of weights currently executing.
	// This is derived from UoW states, not stored separately.
	Running int

	// Roof is the hard ceiling - actual peak allocation.
	// Set at scheduler start based on initial capacity calculation.
	// Workers spawned = min(work items, roof).
	Roof int

	// PressureTarget is the dynamic optimal capacity.
	// Recalculated every 2s based on RAM/CPU pressure.
	// When PressureTarget < Roof, new work is blocked until Running < PressureTarget.
	PressureTarget int
}

// EffectiveRoof returns the roof value to use for display.
// If Roof is 0 (not set), falls back to max(Running, PressureTarget).
func (c CapacityInfo) EffectiveRoof() int {
	if c.Roof > 0 {
		return c.Roof
	}
	// Fallback: use max of running and pressure target
	if c.Running > c.PressureTarget {
		return c.Running
	}
	return c.PressureTarget
}

// IsUnderPressure returns true if the system is throttling due to resource pressure.
// This happens when PressureTarget < Roof.
func (c CapacityInfo) IsUnderPressure() bool {
	return c.PressureTarget > 0 && c.Roof > 0 && c.PressureTarget < c.Roof
}

// HeadroomAvailable returns available capacity (Roof - Running).
// Never negative - returns 0 if running exceeds roof (soft limit exceeded).
func (c CapacityInfo) HeadroomAvailable() int {
	headroom := c.Roof - c.Running
	if headroom < 0 {
		return 0
	}
	return headroom
}

// RenderCapacityLamps returns filled and empty lamp strings for capacity visualization.
// Filled lamps (●) represent running capacity, empty lamps (○) represent available headroom.
// In ASCII mode, uses # for filled and . for empty.
func RenderCapacityLamps(running, roof int, asciiMode bool) (filled, empty string) {
	filledChar := "●"
	emptyChar := "○"
	if asciiMode {
		filledChar = "#"
		emptyChar = "."
	}

	// Clamp running to roof for display
	filledCount := running
	if filledCount > roof {
		filledCount = roof
	}
	if filledCount < 0 {
		filledCount = 0
	}

	emptyCount := roof - filledCount
	if emptyCount < 0 {
		emptyCount = 0
	}

	return strings.Repeat(filledChar, filledCount), strings.Repeat(emptyChar, emptyCount)
}

// RenderActiveLamps returns lamp strings for the Active display with 3 states.
// Uses a fixed 16 lamps normalized to the roof capacity.
// - active: filled lamps for currently running (filled orange)
// - loft: unfilled lamps for available capacity within pressure target (unfilled orange)
// - blocked: unfilled lamps for capacity above pressure target (grey/unallocatable)
func RenderActiveLamps(running, pressureTarget, roof int, asciiMode bool) (active, loft, blocked string) {
	filledChar := "●"
	emptyChar := "○"
	if asciiMode {
		filledChar = "#"
		emptyChar = "o"
	}

	// Handle edge cases
	if roof <= 0 {
		return "", "", strings.Repeat(emptyChar, stableLampsCount)
	}

	// Normalize to stableLampsCount lamps
	// Scale factor: how many "units" each lamp represents
	scale := float64(roof) / float64(stableLampsCount)

	// Calculate lamp counts (normalized to 16)
	activeLamps := int(float64(running)/scale + 0.5)
	if activeLamps > stableLampsCount {
		activeLamps = stableLampsCount
	}
	if activeLamps < 0 {
		activeLamps = 0
	}

	// Pressure target lamps (cap at roof)
	targetLamps := int(float64(pressureTarget)/scale + 0.5)
	if targetLamps > stableLampsCount {
		targetLamps = stableLampsCount
	}
	if targetLamps < activeLamps {
		targetLamps = activeLamps
	}

	// Loft = available within pressure target
	loftLamps := targetLamps - activeLamps

	// Blocked = above pressure target (unallocatable)
	blockedLamps := stableLampsCount - targetLamps

	return strings.Repeat(filledChar, activeLamps),
		strings.Repeat(emptyChar, loftLamps),
		strings.Repeat(emptyChar, blockedLamps)
}

// FormatCapacityDisplay formats the capacity display string.
// Format: "●●●●●●○○○○ 6/10" or "●●●●●●●●●● 24/24 [▼16]" under pressure.
func FormatCapacityDisplay(info CapacityInfo, asciiMode bool) string {
	roof := info.EffectiveRoof()
	filled, empty := RenderCapacityLamps(info.Running, roof, asciiMode)

	result := fmt.Sprintf("%s%s %d/%d", filled, empty, info.Running, roof)

	// Add pressure indicator if system is throttling
	if info.IsUnderPressure() {
		pressureIcon := "▼"
		if asciiMode {
			pressureIcon = "v"
		}
		result += fmt.Sprintf(" [%s%d]", pressureIcon, info.PressureTarget)
	}

	return result
}

// DeriveCounts computes statistics from uowStates.
// Call once per View() - O(n) is fast for <1000 units.
// This replaces the counter fields (uowTotal, uowDone, uowCached, uowFailed).
func (m Model) DeriveCounts() DerivedCounts {
	var c DerivedCounts
	c.Total = len(m.uowStates)

	for _, state := range m.uowStates {
		switch state.Status {
		case UoWPending:
			c.Pending++
		case UoWRunning:
			w := state.Weight
			if w <= 0 {
				w = 1 // Default weight to 1
			}
			c.Running += w
		case UoWComplete:
			c.Done++
		case UoWSkipped:
			c.Cached++
		case UoWFailed:
			c.Failed++
		}
	}
	return c
}

// Completed returns the count of finished UoWs (done + cached + failed).
func (c DerivedCounts) Completed() int {
	return c.Done + c.Cached + c.Failed
}

// ProgressPercent returns completion percentage (0-100).
// Returns 0 if no UoWs are registered.
func (c DerivedCounts) ProgressPercent() float64 {
	if c.Total == 0 {
		return 0
	}
	return float64(c.Completed()) / float64(c.Total) * 100
}
