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
		// Basic cases - balanced CPU/RAM
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

		// configMax as ceiling (the bug fix case)
		{
			name:      "configMax higher than detected - uses detected",
			cpuCount:  8,
			ramGB:     16,
			configMax: 16,
			turbo:     1.0,
			want:      8, // detected=8, configMax=16, uses 8
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

		// Turbo mode
		{
			name:      "turbo 1.25x on 8 CPU",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     1.25,
			want:      10, // base=8, 8*1.25=10, cap=16 (2x CPU)
		},
		{
			name:      "turbo 2.0x on 8 CPU",
			cpuCount:  8,
			ramGB:     16,
			configMax: 0,
			turbo:     2.0,
			want:      16, // base=8, 8*2=16, cap=16 (2x CPU)
		},
		{
			name:      "turbo 2.0x on 8 CPU with low configMax",
			cpuCount:  8,
			ramGB:     16,
			configMax: 10,
			turbo:     2.0,
			want:      10, // base=8, 8*2=16, but configMax=10 caps it
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
			ramGB:     128,
			configMax: 0,
			turbo:     2.0,
			want:      64, // base=64, 64*2=128, but max is 64
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

func TestCalculateCapacity_ConfigMaxAsCeiling(t *testing.T) {
	// This test specifically validates the bug fix:
	// configMax should be a ceiling, not an override

	// Simulate a small machine (half resources of a typical dev machine)
	smallMachineCPU := 8
	smallMachineRAM := 16 // GB

	// Repo config designed for larger machine
	repoConfigParallelism := 16

	// Expected: should use detected capacity (8), not config (16)
	got := calculateCapacity(smallMachineCPU, smallMachineRAM, repoConfigParallelism, 1.0)
	if got != 8 {
		t.Errorf("On small machine with configMax=16, got capacity=%d, want 8", got)
	}

	// Large machine should also respect configMax as ceiling
	largeMachineCPU := 32
	largeMachineRAM := 64 // GB

	// Expected: should use configMax (16), not detected (32)
	got = calculateCapacity(largeMachineCPU, largeMachineRAM, repoConfigParallelism, 1.0)
	if got != 16 {
		t.Errorf("On large machine with configMax=16, got capacity=%d, want 16", got)
	}
}
