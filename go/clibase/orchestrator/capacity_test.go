package orchestrator

import "testing"

func TestCalculateCapacity(t *testing.T) {
	tests := []struct {
		name      string
		cpuCount  int
		ramGB     int
		configMax int
		turbo     float64
		want      int
	}{
		// Basic cases - RAM/3 because each weight unit uses ~2.5GB + overhead
		{
			name:      "16 CPU, 32GB RAM, no config, no turbo",
			cpuCount:  16,
			ramGB:     32,
			configMax: 0,
			turbo:     1.0,
			want:      10, // min(16, 32/3=10) = 10
		},
		{
			name:      "8 CPU, 16GB RAM, no config, no turbo",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     1.0,
			want:      5, // min(8, 16/3=5) = 5
		},
		{
			name:      "4 CPU, 8GB RAM, no config, no turbo",
			cpuCount:  4,
			ramGB:     8,
			configMax: 0,
			turbo:     1.0,
			want:      2, // min(4, 8/3=2) = 2
		},

		// RAM-limited cases
		{
			name:      "16 CPU, 8GB RAM - RAM limited",
			cpuCount:  16,
			ramGB:     8,
			configMax: 0,
			turbo:     1.0,
			want:      2, // min(16, 8/3=2) = 2
		},
		{
			name:      "8 CPU, 4GB RAM - RAM limited",
			cpuCount:  8,
			ramGB:     4,
			configMax: 0,
			turbo:     1.0,
			want:      1, // min(8, 4/3=1) = 1
		},

		// CPU-limited cases
		{
			name:      "4 CPU, 32GB RAM - CPU limited",
			cpuCount:  4,
			ramGB:     32,
			configMax: 0,
			turbo:     1.0,
			want:      4, // min(4, 32/3=10) = 4
		},
		{
			name:      "2 CPU, 16GB RAM - CPU limited",
			cpuCount:  2,
			ramGB:     16,
			configMax: 0,
			turbo:     1.0,
			want:      2, // min(2, 16/3=5) = 2
		},

		// configMax (--roof) fully controls capacity
		{
			name:      "configMax higher than detected - uses configMax",
			cpuCount:  8,
			ramGB:     16,
			configMax: 16,
			turbo:     1.0,
			want:      16, // --roof=16 overrides detected=5
		},
		{
			name:      "configMax lower than detected - uses configMax",
			cpuCount:  16,
			ramGB:     32,
			configMax: 4,
			turbo:     1.0,
			want:      4, // detected=10, configMax=4, uses 4
		},
		{
			name:      "configMax equals detected",
			cpuCount:  8,
			ramGB:     16,
			configMax: 5,
			turbo:     1.0,
			want:      5,
		},

		// Turbo mode - now capped at RAM limit to prevent overcommit
		{
			name:      "turbo 1.25x on 8 CPU, 16GB",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     1.25,
			want:      5, // base=5, 5*1.25=6.25->6, but capped at ramGB/3=5 to prevent overcommit
		},
		{
			name:      "turbo 2.0x on 8 CPU, 16GB",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     2.0,
			want:      5, // base=5, 5*2=10, but capped at ramGB/3=5 to prevent overcommit
		},
		{
			name:      "turbo 2.0x on 8 CPU with low configMax",
			cpuCount:  8,
			ramGB:     16,
			configMax: 8,
			turbo:     2.0,
			want:      8, // base=5, 5*2=10, but configMax=8 caps it
		},

		// Edge cases
		{
			name:      "minimum capacity is 1",
			cpuCount:  1,
			ramGB:     1,
			configMax: 0,
			turbo:     1.0,
			want:      1,
		},
		{
			name:      "very low RAM",
			cpuCount:  8,
			ramGB:     1,
			configMax: 0,
			turbo:     1.0,
			want:      1, // 1GB/3 = 0, but min is 1
		},
		{
			name:      "turbo with 64 cap",
			cpuCount:  64,
			ramGB:     192,
			configMax: 0,
			turbo:     2.0,
			want:      64, // base=min(64, 64)=64, 64*2=128, but max is 64
		},
		{
			name:      "configMax of 1 forces sequential",
			cpuCount:  16,
			ramGB:     32,
			configMax: 1,
			turbo:     1.0,
			want:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCapacity(tt.cpuCount, tt.ramGB, tt.configMax, tt.turbo)
			if got != tt.want {
				t.Errorf("calculateCapacity(%d, %d, %d, %.2f) = %d, want %d",
					tt.cpuCount, tt.ramGB, tt.configMax, tt.turbo, got, tt.want)
			}
		})
	}
}

func TestCalculateCapacity_RoofFullyControls(t *testing.T) {
	// This test validates that --roof fully controls capacity,
	// allowing both raising AND lowering from auto-detected values

	// Case 1: Small machine with high --roof - should raise capacity
	smallMachineCPU := 8
	smallMachineRAM := 16 // GB
	highRoof := 16

	got := calculateCapacity(smallMachineCPU, smallMachineRAM, highRoof, 1.0)
	if got != 16 {
		t.Errorf("Small machine with --roof=16: got capacity=%d, want 16 (roof overrides detected)", got)
	}

	// Case 2: Large machine with low --roof - should lower capacity
	largeMachineCPU := 32
	largeMachineRAM := 64 // GB
	lowRoof := 8

	got = calculateCapacity(largeMachineCPU, largeMachineRAM, lowRoof, 1.0)
	if got != 8 {
		t.Errorf("Large machine with --roof=8: got capacity=%d, want 8 (roof overrides detected)", got)
	}

	// Case 3: No roof (0) - should use auto-detection (RAM/3)
	// 16GB/3 = 5, min(8, 5) = 5
	got = calculateCapacity(smallMachineCPU, smallMachineRAM, 0, 1.0)
	if got != 5 {
		t.Errorf("Small machine with no --roof: got capacity=%d, want 5 (auto-detected from 16GB/3)", got)
	}
}

// TestCalculateCapacity_16GB_8CPU validates the exact system configuration
// described in the bug report (16GB RAM, 8 CPU cores).
// This test should PASS after the fix - it verifies capacity=5 with turbo=1.0.
func TestCalculateCapacity_16GB_8CPU(t *testing.T) {
	// User's exact system: 16GB RAM, 8 CPU cores
	cpuCount := 8
	ramGB := 16
	configMax := 0 // No --roof flag
	turbo := 1.0   // No turbo

	capacity := calculateCapacity(cpuCount, ramGB, configMax, turbo)

	// Expected: min(CPU=8, RAM/3=5) × turbo=1.0 = 5
	// This allows 1 PDF build at a time (weight=4, uses ~8GB)
	if capacity != 5 {
		t.Errorf("16GB/8CPU should give capacity=5, got %d", capacity)
		t.Logf("Expected calculation: min(8, 16/3=5) × 1.0 = 5")
		t.Logf("This ensures only 1 PDF (weight=4, ~8GB) runs at a time")
	}
}

// TestCalculateCapacity_TurboDoesNotOvercommit validates that the turbo multiplier
// respects RAM limits and does not cause memory overcommit.
// This test FAILS before the fix because turbo currently bypasses RAM limits.
func TestCalculateCapacity_TurboDoesNotOvercommit(t *testing.T) {
	// User's system: 16GB RAM, 8 CPU cores
	cpuCount := 8
	ramGB := 16
	configMax := 0 // No --roof flag

	tests := []struct {
		name        string
		turbo       float64
		maxExpected int
		reason      string
	}{
		{
			name:        "turbo=1.0 baseline",
			turbo:       1.0,
			maxExpected: 5,
			reason:      "base capacity is 5 (16GB/3=5)",
		},
		{
			name:        "turbo=2.0 should not exceed RAM limit",
			turbo:       2.0,
			maxExpected: 5,
			reason:      "turbo should not exceed RAM/3=5 to prevent memory overcommit",
		},
		{
			name:        "turbo=3.2 should not exceed RAM limit",
			turbo:       3.2,
			maxExpected: 5,
			reason:      "observed capacity=16 causes 4 PDFs (32GB) on 16GB system - MUST FAIL",
		},
		{
			name:        "turbo=4.0 should not exceed RAM limit",
			turbo:       4.0,
			maxExpected: 5,
			reason:      "even extreme turbo should respect RAM/3=5 limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := calculateCapacity(cpuCount, ramGB, configMax, tt.turbo)

			// Critical assertion: turbo should NEVER exceed RAM-based capacity
			// RAM capacity = 16GB / 3 = 5 (each weight unit = ~2.5GB + overhead)
			// If capacity > 5, we can schedule 4+ PDFs (16 weight units)
			// That would require 32GB on a 16GB system → overcommit → swapping → timeout
			if capacity > tt.maxExpected {
				t.Errorf("Turbo should not exceed RAM limit: capacity=%d > %d (RAM/3)", capacity, tt.maxExpected)
				t.Logf("Reason: %s", tt.reason)
				t.Logf("With capacity=%d, we could schedule %d PDFs (weight=4 each)", capacity, capacity/4)
				t.Logf("That requires ~%dGB RAM on a 16GB system → OVERCOMMIT", (capacity/4)*8)
			}
		})
	}
}

// TestCalculateCapacity_RoofOverridesAll validates that the --roof flag
// overrides ALL other calculations including turbo.
// This test should PASS - it verifies --roof is the ultimate override.
func TestCalculateCapacity_RoofOverridesAll(t *testing.T) {
	cpuCount := 8
	ramGB := 16
	roof := 10   // User explicitly requests capacity=10
	turbo := 1.0 // Turbo doesn't matter when roof is set

	capacity := calculateCapacity(cpuCount, ramGB, roof, turbo)

	// --roof should override auto-detection completely
	if capacity != 10 {
		t.Errorf("--roof should override: expected 10, got %d", capacity)
		t.Logf("Detected values: CPU=%d, RAM/3=%d, but --roof=%d should win", cpuCount, ramGB/3, roof)
	}

	// Test with turbo - roof still wins
	capacity = calculateCapacity(cpuCount, ramGB, roof, 2.0)
	if capacity != 10 {
		t.Errorf("--roof should override even with turbo: expected 10, got %d", capacity)
	}
}

// TestDetectAvailableCapacity_UsesNewDetection validates that detectAvailableCapacity
// now calls the new GetEffectiveCPUs() and GetEffectiveMemoryBytes() functions.
// This test FAILS before the fix because the old code uses runtime.NumCPU() directly.
//
// NOTE: This test is challenging because GetEffectiveCPUs/GetEffectiveMemoryBytes
// have external dependencies (Docker, WSL). For now, we document the expected behavior.
// A complete fix would require refactoring detectAvailableCapacity to accept
// injected functions for testing.
func TestDetectAvailableCapacity_UsesNewDetection(t *testing.T) {
	t.Skip("TODO: This test requires refactoring detectAvailableCapacity to accept injected functions")

	// After fix, detectAvailableCapacity should call:
	// - GetEffectiveCPUs() instead of runtime.NumCPU()
	// - GetEffectiveMemoryBytes() instead of mem.VirtualMemory()
	//
	// Expected behavior after fix:
	// func detectAvailableCapacity(configMax int, turbo float64) int {
	//     cpuCount := GetEffectiveCPUs()
	//     if cpuCount < 1 {
	//         cpuCount = 4 // Fallback
	//     }
	//
	//     var ramGB int
	//     effectiveMem := GetEffectiveMemoryBytes()
	//     if effectiveMem > 0 {
	//         ramGB = int(effectiveMem / (1024 * 1024 * 1024))
	//     } else {
	//         // Fallback: use host available RAM (not total)
	//         memInfo, err := mem.VirtualMemory()
	//         if err == nil {
	//             ramGB = int(memInfo.Available / (1024 * 1024 * 1024))
	//         } else {
	//             ramGB = 8
	//         }
	//     }
	//
	//     return calculateCapacity(cpuCount, ramGB, configMax, turbo)
	// }
}
