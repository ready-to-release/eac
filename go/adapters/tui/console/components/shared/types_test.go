package shared

import (
	"testing"
	"time"
)

func TestUnitStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status UnitStatus
		want   string
	}{
		{
			name:   "pending status",
			status: UnitPending,
			want:   "Pending",
		},
		{
			name:   "queued status",
			status: UnitQueued,
			want:   "Queued",
		},
		{
			name:   "running status",
			status: UnitRunning,
			want:   "Running",
		},
		{
			name:   "success status",
			status: UnitSuccess,
			want:   "Success",
		},
		{
			name:   "skipped status",
			status: UnitSkipped,
			want:   "Skipped",
		},
		{
			name:   "failed status",
			status: UnitFailed,
			want:   "Failed",
		},
		{
			name:   "unknown status",
			status: UnitStatus(99),
			want:   "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("UnitStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnitStatus_Icon(t *testing.T) {
	tests := []struct {
		name   string
		status UnitStatus
		want   string
	}{
		{
			name:   "pending icon",
			status: UnitPending,
			want:   "o",
		},
		{
			name:   "queued icon",
			status: UnitQueued,
			want:   "*",
		},
		{
			name:   "running icon",
			status: UnitRunning,
			want:   ">",
		},
		{
			name:   "success icon",
			status: UnitSuccess,
			want:   "V",
		},
		{
			name:   "skipped icon",
			status: UnitSkipped,
			want:   "=",
		},
		{
			name:   "failed icon",
			status: UnitFailed,
			want:   "X",
		},
		{
			name:   "unknown icon",
			status: UnitStatus(99),
			want:   "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.Icon()
			if got != tt.want {
				t.Errorf("UnitStatus.Icon() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnitState_Duration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		unit      UnitState
		checkFunc func(t *testing.T, d time.Duration)
	}{
		{
			name: "zero start time returns zero",
			unit: UnitState{},
			checkFunc: func(t *testing.T, d time.Duration) {
				if d != 0 {
					t.Errorf("Duration() = %v, want 0", d)
				}
			},
		},
		{
			name: "completed unit returns fixed duration",
			unit: UnitState{
				StartTime: now.Add(-10 * time.Second),
				EndTime:   now,
			},
			checkFunc: func(t *testing.T, d time.Duration) {
				if d != 10*time.Second {
					t.Errorf("Duration() = %v, want 10s", d)
				}
			},
		},
		{
			name: "running unit returns elapsed time",
			unit: UnitState{
				StartTime: now.Add(-5 * time.Second),
				// EndTime is zero (not set)
			},
			checkFunc: func(t *testing.T, d time.Duration) {
				// Should be approximately 5 seconds (with some tolerance for test execution)
				if d < 5*time.Second || d > 6*time.Second {
					t.Errorf("Duration() = %v, want approximately 5s", d)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.unit.Duration()
			tt.checkFunc(t, got)
		})
	}
}

func TestDirection_Values(t *testing.T) {
	// Verify direction constants have expected values
	tests := []struct {
		name string
		dir  Direction
		want int
	}{
		{"DirUp", DirUp, 0},
		{"DirDown", DirDown, 1},
		{"DirLeft", DirLeft, 2},
		{"DirRight", DirRight, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.dir) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, int(tt.dir), tt.want)
			}
		})
	}
}
