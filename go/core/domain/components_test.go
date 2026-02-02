package domain

import "testing"

func TestAmpConfig_GetAmp(t *testing.T) {
	tests := []struct {
		name     string
		amp      *AmpConfig
		op       string
		expected float64
	}{
		{"nil returns 1.0", nil, "build", 1.0},
		{"empty struct returns 1.0", &AmpConfig{}, "build", 1.0},
		{"build 2.0", &AmpConfig{Build: 2.0}, "build", 2.0},
		{"lint 0.5", &AmpConfig{Lint: 0.5}, "lint", 0.5},
		{"test 1.5", &AmpConfig{Test: 1.5}, "test", 1.5},
		{"scan 3.0", &AmpConfig{Scan: 3.0}, "scan", 3.0},
		{"unknown op returns 1.0", &AmpConfig{Build: 2.0}, "unknown", 1.0},
		{"zero value returns 1.0", &AmpConfig{Build: 0}, "build", 1.0},
		{"negative value returns 1.0", &AmpConfig{Build: -1.0}, "build", 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.amp.GetAmp(tt.op)
			if got != tt.expected {
				t.Errorf("GetAmp(%s) = %v, want %v", tt.op, got, tt.expected)
			}
		})
	}
}

func TestComponentEntry_GetAmpForOperation(t *testing.T) {
	tests := []struct {
		name     string
		entry    *ComponentEntry
		op       string
		expected float64
	}{
		{"nil entry returns 1.0", nil, "build", 1.0},
		{"nil amp returns 1.0", &ComponentEntry{}, "build", 1.0},
		{"configured amp", &ComponentEntry{Amp: &AmpConfig{Build: 2.0}}, "build", 2.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetAmpForOperation(tt.op)
			if got != tt.expected {
				t.Errorf("GetAmpForOperation(%s) = %v, want %v", tt.op, got, tt.expected)
			}
		})
	}
}
