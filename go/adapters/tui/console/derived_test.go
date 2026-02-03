package console

import (
	"testing"
)

func TestDerivedCounts_Empty(t *testing.T) {
	m := Model{
		uowStates: make(map[string]*UoWState),
	}

	counts := m.DeriveCounts()

	if counts.Total != 0 {
		t.Errorf("Total: got %d, want 0", counts.Total)
	}
	if counts.Running != 0 {
		t.Errorf("Running: got %d, want 0", counts.Running)
	}
	if counts.Done != 0 {
		t.Errorf("Done: got %d, want 0", counts.Done)
	}
	if counts.Cached != 0 {
		t.Errorf("Cached: got %d, want 0", counts.Cached)
	}
	if counts.Failed != 0 {
		t.Errorf("Failed: got %d, want 0", counts.Failed)
	}
	if counts.Pending != 0 {
		t.Errorf("Pending: got %d, want 0", counts.Pending)
	}
}

func TestDerivedCounts_AllPending(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:c1": {Status: UoWPending, Weight: 1},
			"m1:c2": {Status: UoWPending, Weight: 1},
			"m2:c1": {Status: UoWPending, Weight: 1},
		},
	}

	counts := m.DeriveCounts()

	if counts.Total != 3 {
		t.Errorf("Total: got %d, want 3", counts.Total)
	}
	if counts.Pending != 3 {
		t.Errorf("Pending: got %d, want 3", counts.Pending)
	}
	if counts.Running != 0 {
		t.Errorf("Running: got %d, want 0", counts.Running)
	}
}

func TestDerivedCounts_WeightAwareRunning(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:c1": {Status: UoWRunning, Weight: 5},
			"m1:c2": {Status: UoWRunning, Weight: 3},
			"m2:c1": {Status: UoWComplete, Weight: 2},
		},
	}

	counts := m.DeriveCounts()

	if counts.Total != 3 {
		t.Errorf("Total: got %d, want 3", counts.Total)
	}
	// Running should sum weights: 5 + 3 = 8
	if counts.Running != 8 {
		t.Errorf("Running (weight-aware): got %d, want 8", counts.Running)
	}
	if counts.Done != 1 {
		t.Errorf("Done: got %d, want 1", counts.Done)
	}
}

func TestDerivedCounts_ZeroWeightDefaults(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:c1": {Status: UoWRunning, Weight: 0}, // Zero weight should default to 1
			"m1:c2": {Status: UoWRunning, Weight: -1}, // Negative weight should default to 1
		},
	}

	counts := m.DeriveCounts()

	// Running should default zero/negative weights to 1: 1 + 1 = 2
	if counts.Running != 2 {
		t.Errorf("Running (with default weights): got %d, want 2", counts.Running)
	}
}

func TestDerivedCounts_MixedStates(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:go":     {Status: UoWComplete, Weight: 1},
			"m1:docker": {Status: UoWComplete, Weight: 2},
			"m2:go":     {Status: UoWRunning, Weight: 3},
			"m2:docker": {Status: UoWPending, Weight: 1},
			"m3:go":     {Status: UoWSkipped, Weight: 1},
			"m3:docker": {Status: UoWFailed, Weight: 1},
		},
	}

	counts := m.DeriveCounts()

	if counts.Total != 6 {
		t.Errorf("Total: got %d, want 6", counts.Total)
	}
	if counts.Done != 2 {
		t.Errorf("Done: got %d, want 2", counts.Done)
	}
	if counts.Running != 3 { // Weight of running unit
		t.Errorf("Running: got %d, want 3", counts.Running)
	}
	if counts.Pending != 1 {
		t.Errorf("Pending: got %d, want 1", counts.Pending)
	}
	if counts.Cached != 1 {
		t.Errorf("Cached: got %d, want 1", counts.Cached)
	}
	if counts.Failed != 1 {
		t.Errorf("Failed: got %d, want 1", counts.Failed)
	}
}

func TestDerivedCounts_AllComplete(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:c1": {Status: UoWComplete, Weight: 1},
			"m1:c2": {Status: UoWComplete, Weight: 1},
			"m2:c1": {Status: UoWComplete, Weight: 1},
		},
	}

	counts := m.DeriveCounts()

	if counts.Total != 3 {
		t.Errorf("Total: got %d, want 3", counts.Total)
	}
	if counts.Done != 3 {
		t.Errorf("Done: got %d, want 3", counts.Done)
	}
	if counts.Running != 0 {
		t.Errorf("Running: got %d, want 0", counts.Running)
	}
}

func TestDerivedCounts_AllCached(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:c1": {Status: UoWSkipped, Weight: 1},
			"m1:c2": {Status: UoWSkipped, Weight: 1},
		},
	}

	counts := m.DeriveCounts()

	if counts.Total != 2 {
		t.Errorf("Total: got %d, want 2", counts.Total)
	}
	if counts.Cached != 2 {
		t.Errorf("Cached: got %d, want 2", counts.Cached)
	}
}

func TestDerivedCounts_Completed(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:c1": {Status: UoWComplete, Weight: 1},
			"m1:c2": {Status: UoWSkipped, Weight: 1},
			"m2:c1": {Status: UoWFailed, Weight: 1},
			"m2:c2": {Status: UoWRunning, Weight: 1},
			"m3:c1": {Status: UoWPending, Weight: 1},
		},
	}

	counts := m.DeriveCounts()

	// Completed = Done + Cached + Failed = 1 + 1 + 1 = 3
	if counts.Completed() != 3 {
		t.Errorf("Completed(): got %d, want 3", counts.Completed())
	}
}

func TestDerivedCounts_ProgressPercent(t *testing.T) {
	tests := []struct {
		name     string
		states   map[string]*UoWState
		expected float64
	}{
		{
			name:     "empty",
			states:   map[string]*UoWState{},
			expected: 0,
		},
		{
			name: "all complete",
			states: map[string]*UoWState{
				"m1:c1": {Status: UoWComplete},
				"m1:c2": {Status: UoWComplete},
			},
			expected: 100,
		},
		{
			name: "half complete",
			states: map[string]*UoWState{
				"m1:c1": {Status: UoWComplete},
				"m1:c2": {Status: UoWRunning, Weight: 1},
			},
			expected: 50,
		},
		{
			name: "mixed finished",
			states: map[string]*UoWState{
				"m1:c1": {Status: UoWComplete},
				"m1:c2": {Status: UoWSkipped},
				"m1:c3": {Status: UoWFailed},
				"m1:c4": {Status: UoWPending},
			},
			expected: 75, // 3 out of 4 finished
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{uowStates: tt.states}
			counts := m.DeriveCounts()
			got := counts.ProgressPercent()

			if got != tt.expected {
				t.Errorf("ProgressPercent(): got %.2f, want %.2f", got, tt.expected)
			}
		})
	}
}

// Tests for capacity tracking (Roof, PressureTarget, Running)
// The three-value capacity model:
// - Running: items currently executing (sum of weights)
// - Roof: hard ceiling - actual allocation (workers spawned at start)
// - PressureTarget: dynamic optimal capacity (may be lower under memory pressure)

func TestCapacityInfo_EffectiveRoof(t *testing.T) {
	tests := []struct {
		name           string
		roof           int
		pressureTarget int
		running        int
		wantRoof       int
	}{
		{
			name:           "roof from status",
			roof:           24,
			pressureTarget: 16,
			running:        12,
			wantRoof:       24,
		},
		{
			name:           "roof zero uses max of running and pressureTarget",
			roof:           0,
			pressureTarget: 16,
			running:        12,
			wantRoof:       16, // max(12, 16)
		},
		{
			name:           "roof zero running exceeds pressureTarget",
			roof:           0,
			pressureTarget: 16,
			running:        24,
			wantRoof:       24, // max(24, 16)
		},
		{
			name:           "all zeros",
			roof:           0,
			pressureTarget: 0,
			running:        0,
			wantRoof:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := CapacityInfo{
				Roof:           tt.roof,
				PressureTarget: tt.pressureTarget,
				Running:        tt.running,
			}
			got := info.EffectiveRoof()
			if got != tt.wantRoof {
				t.Errorf("EffectiveRoof(): got %d, want %d", got, tt.wantRoof)
			}
		})
	}
}

func TestCapacityInfo_IsUnderPressure(t *testing.T) {
	tests := []struct {
		name           string
		roof           int
		pressureTarget int
		want           bool
	}{
		{
			name:           "under pressure - target less than roof",
			roof:           24,
			pressureTarget: 16,
			want:           true,
		},
		{
			name:           "no pressure - target equals roof",
			roof:           24,
			pressureTarget: 24,
			want:           false,
		},
		{
			name:           "no pressure - target greater than roof",
			roof:           16,
			pressureTarget: 24,
			want:           false,
		},
		{
			name:           "no pressure - zero values",
			roof:           0,
			pressureTarget: 0,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := CapacityInfo{
				Roof:           tt.roof,
				PressureTarget: tt.pressureTarget,
			}
			got := info.IsUnderPressure()
			if got != tt.want {
				t.Errorf("IsUnderPressure(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCapacityInfo_HeadroomAvailable(t *testing.T) {
	tests := []struct {
		name    string
		roof    int
		running int
		want    int
	}{
		{
			name:    "partial usage",
			roof:    24,
			running: 16,
			want:    8,
		},
		{
			name:    "full usage",
			roof:    24,
			running: 24,
			want:    0,
		},
		{
			name:    "over capacity (soft limit exceeded)",
			roof:    16,
			running: 24,
			want:    0, // Never negative
		},
		{
			name:    "no running",
			roof:    24,
			running: 0,
			want:    24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := CapacityInfo{
				Roof:    tt.roof,
				Running: tt.running,
			}
			got := info.HeadroomAvailable()
			if got != tt.want {
				t.Errorf("HeadroomAvailable(): got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRenderCapacityLamps(t *testing.T) {
	tests := []struct {
		name       string
		running    int
		roof       int
		asciiMode  bool
		wantFilled int
		wantEmpty  int
	}{
		{
			name:       "partial capacity",
			running:    6,
			roof:       10,
			asciiMode:  false,
			wantFilled: 6,
			wantEmpty:  4,
		},
		{
			name:       "full capacity",
			running:    10,
			roof:       10,
			asciiMode:  false,
			wantFilled: 10,
			wantEmpty:  0,
		},
		{
			name:       "over capacity (running > roof)",
			running:    24,
			roof:       16,
			asciiMode:  false,
			wantFilled: 16, // Capped at roof
			wantEmpty:  0,
		},
		{
			name:       "empty",
			running:    0,
			roof:       10,
			asciiMode:  false,
			wantFilled: 0,
			wantEmpty:  10,
		},
		{
			name:       "ascii mode",
			running:    3,
			roof:       5,
			asciiMode:  true,
			wantFilled: 3,
			wantEmpty:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filled, empty := RenderCapacityLamps(tt.running, tt.roof, tt.asciiMode)

			// Count the lamps in the returned strings
			filledCount := len(filled)
			emptyCount := len(empty)

			// Unicode lamps are multi-byte, ASCII are single-byte
			if !tt.asciiMode {
				// "●" is 3 bytes in UTF-8
				filledCount = len([]rune(filled))
				emptyCount = len([]rune(empty))
			}

			if filledCount != tt.wantFilled {
				t.Errorf("filled count: got %d, want %d (string: %q)", filledCount, tt.wantFilled, filled)
			}
			if emptyCount != tt.wantEmpty {
				t.Errorf("empty count: got %d, want %d (string: %q)", emptyCount, tt.wantEmpty, empty)
			}
		})
	}
}

func TestFormatCapacityDisplay(t *testing.T) {
	tests := []struct {
		name           string
		running        int
		roof           int
		pressureTarget int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "normal operation - no pressure",
			running:        12,
			roof:           24,
			pressureTarget: 24,
			wantContains:   []string{"12", "24"},
			wantNotContain: []string{"▼"}, // No pressure indicator
		},
		{
			name:           "under pressure",
			running:        24,
			roof:           24,
			pressureTarget: 16,
			wantContains:   []string{"24", "24", "▼16"}, // Shows pressure target
			wantNotContain: []string{},
		},
		{
			name:           "at capacity",
			running:        24,
			roof:           24,
			pressureTarget: 24,
			wantContains:   []string{"24", "24"},
			wantNotContain: []string{"▼"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := CapacityInfo{
				Running:        tt.running,
				Roof:           tt.roof,
				PressureTarget: tt.pressureTarget,
			}
			result := FormatCapacityDisplay(info, false)

			for _, want := range tt.wantContains {
				if !containsSubstring(result, want) {
					t.Errorf("result should contain %q, got: %s", want, result)
				}
			}
			for _, notWant := range tt.wantNotContain {
				if containsSubstring(result, notWant) {
					t.Errorf("result should NOT contain %q, got: %s", notWant, result)
				}
			}
		})
	}
}

// containsSubstring checks if s contains substr (helper for tests)
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsSubstringImpl(s, substr)))
}

func containsSubstringImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
