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
		// Basic cases - RAM/2 because each weight unit uses ~1.5-2GB
		{
			name:      "16 CPU, 32GB RAM, no config, no turbo",
			cpuCount:  16,
			ramGB:     32,
			configMax: 0,
			turbo:     1.0,
			want:      16, // min(16, 32/2=16) = 16
		},
		{
			name:      "8 CPU, 16GB RAM, no config, no turbo",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     1.0,
			want:      8, // min(8, 16/2=8) = 8
		},
		{
			name:      "4 CPU, 8GB RAM, no config, no turbo",
			cpuCount:  4,
			ramGB:     8,
			configMax: 0,
			turbo:     1.0,
			want:      4, // min(4, 8/2=4) = 4
		},

		// RAM-limited cases
		{
			name:      "16 CPU, 8GB RAM - RAM limited",
			cpuCount:  16,
			ramGB:     8,
			configMax: 0,
			turbo:     1.0,
			want:      4, // min(16, 8/2=4) = 4
		},
		{
			name:      "8 CPU, 4GB RAM - RAM limited",
			cpuCount:  8,
			ramGB:     4,
			configMax: 0,
			turbo:     1.0,
			want:      2, // min(8, 4/2=2) = 2
		},

		// CPU-limited cases
		{
			name:      "4 CPU, 32GB RAM - CPU limited",
			cpuCount:  4,
			ramGB:     32,
			configMax: 0,
			turbo:     1.0,
			want:      4, // min(4, 32/2=16) = 4
		},
		{
			name:      "2 CPU, 16GB RAM - CPU limited",
			cpuCount:  2,
			ramGB:     16,
			configMax: 0,
			turbo:     1.0,
			want:      2, // min(2, 16/2=8) = 2
		},

		// configMax (--roof) fully controls capacity
		{
			name:      "configMax higher than detected - uses configMax",
			cpuCount:  8,
			ramGB:     16,
			configMax: 16,
			turbo:     1.0,
			want:      16, // --roof=16 overrides detected=8
		},
		{
			name:      "configMax lower than detected - uses configMax",
			cpuCount:  16,
			ramGB:     32,
			configMax: 4,
			turbo:     1.0,
			want:      4, // detected=16, configMax=4, uses 4
		},
		{
			name:      "configMax equals detected",
			cpuCount:  8,
			ramGB:     16,
			configMax: 8,
			turbo:     1.0,
			want:      8,
		},

		// Turbo mode - now capped at RAM limit to prevent overcommit
		{
			name:      "turbo 1.25x on 8 CPU, 16GB",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     1.25,
			want:      8, // base=8, 8*1.25=10, but capped at ramGB/2=8 to prevent overcommit
		},
		{
			name:      "turbo 2.0x on 8 CPU, 16GB",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     2.0,
			want:      8, // base=8, 8*2=16, but capped at ramGB/2=8 to prevent overcommit
		},
		{
			name:      "turbo 2.0x on 8 CPU with low configMax",
			cpuCount:  8,
			ramGB:     16,
			configMax: 6,
			turbo:     2.0,
			want:      6, // --roof=6 takes precedence
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
			want:      1, // 1GB/2 = 0, but min is 1
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

	// Case 3: No roof (0) - should use auto-detection (RAM/2)
	// 16GB/2 = 8, min(8, 8) = 8
	got = calculateCapacity(smallMachineCPU, smallMachineRAM, 0, 1.0)
	if got != 8 {
		t.Errorf("Small machine with no --roof: got capacity=%d, want 8 (auto-detected from min(8 CPU, 16GB/2=8))", got)
	}
}

// TestCalculateCapacity_16GB_8CPU validates the exact system configuration
// described in the bug report (16GB RAM, 8 CPU cores).
// This test should PASS after the fix - it verifies capacity=8 with turbo=1.0.
func TestCalculateCapacity_16GB_8CPU(t *testing.T) {
	// User's exact system: 16GB RAM, 8 CPU cores
	cpuCount := 8
	ramGB := 16
	configMax := 0 // No --roof flag
	turbo := 1.0   // No turbo

	capacity := calculateCapacity(cpuCount, ramGB, configMax, turbo)

	// Expected: min(CPU=8, RAM/2=8) × turbo=1.0 = 8
	// With 16GB RAM and 8 CPUs, we can run 8 parallel lightweight tasks
	// or 2 PDF builds (weight=4 each, ~4GB each)
	if capacity != 8 {
		t.Errorf("16GB/8CPU should give capacity=8, got %d", capacity)
		t.Logf("Expected calculation: min(8, 16/2=8) × 1.0 = 8")
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
			maxExpected: 8,
			reason:      "base capacity is 8 (min(8 CPU, 16GB/2=8))",
		},
		{
			name:        "turbo=2.0 should not exceed RAM limit",
			turbo:       2.0,
			maxExpected: 8,
			reason:      "turbo should not exceed RAM/2=8 to prevent memory overcommit",
		},
		{
			name:        "turbo=3.2 should not exceed RAM limit",
			turbo:       3.2,
			maxExpected: 8,
			reason:      "turbo is capped by RAM/2=8 limit",
		},
		{
			name:        "turbo=4.0 should not exceed RAM limit",
			turbo:       4.0,
			maxExpected: 8,
			reason:      "even extreme turbo should respect RAM/2=8 limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capacity := calculateCapacity(cpuCount, ramGB, configMax, tt.turbo)

			// Critical assertion: turbo should NEVER exceed RAM-based capacity
			// RAM capacity = 16GB / 2 = 8 (each weight unit = ~1.5-2GB)
			// If capacity > 8, we can schedule more than RAM supports
			if capacity > tt.maxExpected {
				t.Errorf("Turbo should not exceed RAM limit: capacity=%d > %d (RAM/2)", capacity, tt.maxExpected)
				t.Logf("Reason: %s", tt.reason)
				t.Logf("With capacity=%d, we could schedule %d PDFs (weight=4 each)", capacity, capacity/4)
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
		t.Logf("Detected values: CPU=%d, RAM/2=%d, but --roof=%d should win", cpuCount, ramGB/2, roof)
	}

	// Test with turbo - roof still wins
	capacity = calculateCapacity(cpuCount, ramGB, roof, 2.0)
	if capacity != 10 {
		t.Errorf("--roof should override even with turbo: expected 10, got %d", capacity)
	}
}

// TestDetectAvailableCapacity_UsesInjectedDetector validates that detectAvailableCapacityWith
// uses the injected CapacityDetector for CPU and memory detection.
func TestDetectAvailableCapacity_UsesInjectedDetector(t *testing.T) {
	detector := &mockCapacityDetector{cpus: 8, memGB: 16}

	// Auto-detect mode (configMax = 0)
	capacity := detectAvailableCapacityWith(detector, 0, 1.0)
	// min(8 CPU, 16/2=8 RAM) * 1.0 = 8
	if capacity != 8 {
		t.Errorf("expected capacity 8 (CPU-limited), got %d", capacity)
	}

	// RAM-limited scenario
	detector.cpus = 16
	detector.memGB = 8
	capacity = detectAvailableCapacityWith(detector, 0, 1.0)
	// min(16 CPU, 8/2=4 RAM) * 1.0 = 4
	if capacity != 4 {
		t.Errorf("expected capacity 4 (RAM-limited), got %d", capacity)
	}

	// With --roof override
	capacity = detectAvailableCapacityWith(detector, 10, 1.0)
	if capacity != 10 {
		t.Errorf("expected capacity 10 (roof override), got %d", capacity)
	}

	// With turbo
	detector.cpus = 4
	detector.memGB = 32
	capacity = detectAvailableCapacityWith(detector, 0, 1.5)
	// min(4 CPU, 32/2=16 RAM) * 1.5 = 6, capped at 8 (2x CPU with turbo)
	if capacity != 6 {
		t.Errorf("expected capacity 6 (turbo), got %d", capacity)
	}

	// Fallback: 0 CPUs should use fallback of 4
	detector.cpus = 0
	detector.memGB = 16
	capacity = detectAvailableCapacityWith(detector, 0, 1.0)
	// min(4 fallback CPU, 16/2=8 RAM) * 1.0 = 4
	if capacity != 4 {
		t.Errorf("expected capacity 4 (CPU fallback), got %d", capacity)
	}
}

type mockCapacityDetector struct {
	cpus  int
	memGB int
}

func (m *mockCapacityDetector) EffectiveCPUs() int    { return m.cpus }
func (m *mockCapacityDetector) EffectiveMemoryGB() int { return m.memGB }
