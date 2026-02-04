package cells

import "testing"

func TestUnitStatus_Colors(t *testing.T) {
	tests := []struct {
		name       string
		status     UnitStatus
		wantBorder string
		wantText   string
		wantBg     string
	}{
		{
			name:       "pending returns gray colors",
			status:     UnitPending,
			wantBorder: "238",
			wantText:   "245",
			wantBg:     "234",
		},
		{
			name:       "queued returns gray colors (same as pending)",
			status:     UnitQueued,
			wantBorder: "238",
			wantText:   "245",
			wantBg:     "234",
		},
		{
			name:       "running returns orange/yellow colors",
			status:     UnitRunning,
			wantBorder: "214",
			wantText:   "214",
			wantBg:     "94",
		},
		{
			name:       "complete returns green colors",
			status:     UnitComplete,
			wantBorder: "40",
			wantText:   "40",
			wantBg:     "22",
		},
		{
			name:       "skipped returns cyan/blue colors",
			status:     UnitSkipped,
			wantBorder: "75",
			wantText:   "75",
			wantBg:     "23",
		},
		{
			name:       "failed returns red colors",
			status:     UnitFailed,
			wantBorder: "196",
			wantText:   "196",
			wantBg:     "52",
		},
		{
			name:       "unknown status returns gray colors (default)",
			status:     UnitStatus(999), // Invalid status value
			wantBorder: "238",
			wantText:   "245",
			wantBg:     "234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBorder, gotText, gotBg := tt.status.Colors()

			if gotBorder != tt.wantBorder {
				t.Errorf("Colors() border = %q, want %q", gotBorder, tt.wantBorder)
			}
			if gotText != tt.wantText {
				t.Errorf("Colors() text = %q, want %q", gotText, tt.wantText)
			}
			if gotBg != tt.wantBg {
				t.Errorf("Colors() bg = %q, want %q", gotBg, tt.wantBg)
			}
		})
	}
}
