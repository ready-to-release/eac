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

		// Turbo mode
		{
			name:      "turbo 1.25x on 8 CPU, 16GB",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     1.25,
			want:      6, // base=5, 5*1.25=6.25->6, cap=16 (2x CPU)
		},
		{
			name:      "turbo 2.0x on 8 CPU, 16GB",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     2.0,
			want:      10, // base=5, 5*2=10, cap=16 (2x CPU)
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
