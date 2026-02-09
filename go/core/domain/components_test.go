package domain

import (
	"sort"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
)

func TestAmpConfig_GetAmp(t *testing.T) {
	tests := []struct {
		name     string
		amp      *config.AmpConfig
		op       string
		expected float64
	}{
		{"nil returns 1.0", nil, "build", 1.0},
		{"empty struct returns 1.0", &config.AmpConfig{}, "build", 1.0},
		{"build 2.0", &config.AmpConfig{Build: 2.0}, "build", 2.0},
		{"lint 0.5", &config.AmpConfig{Lint: 0.5}, "lint", 0.5},
		{"test 1.5", &config.AmpConfig{Test: 1.5}, "test", 1.5},
		{"scan 3.0", &config.AmpConfig{Scan: 3.0}, "scan", 3.0},
		{"unknown op returns 1.0", &config.AmpConfig{Build: 2.0}, "unknown", 1.0},
		{"zero value returns 1.0", &config.AmpConfig{Build: 0}, "build", 1.0},
		{"negative value returns 1.0", &config.AmpConfig{Build: -1.0}, "build", 1.0},
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
		entry    *config.ComponentEntry
		op       string
		expected float64
	}{
		{"nil entry returns 1.0", nil, "build", 1.0},
		{"nil amp returns 1.0", &config.ComponentEntry{}, "build", 1.0},
		{"configured amp", &config.ComponentEntry{Amp: &config.AmpConfig{Build: 2.0}}, "build", 2.0},
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

func TestModuleComponents_GetComponentTypes(t *testing.T) {
	tests := []struct {
		name     string
		mc       config.ModuleComponents
		expected []string
	}{
		{
			name:     "nil returns nil",
			mc:       nil,
			expected: nil,
		},
		{
			name:     "empty map returns empty slice",
			mc:       config.ModuleComponents{},
			expected: []string{},
		},
		{
			name: "name is type when Type field empty",
			mc: config.ModuleComponents{
				"go": &config.ComponentEntry{Root: "go/"},
			},
			expected: []string{"go"},
		},
		{
			name: "Type field overrides name",
			mc: config.ModuleComponents{
				"python": &config.ComponentEntry{Type: "testdata", Root: "testdata/python"},
			},
			expected: []string{"testdata"},
		},
		{
			name: "deduplicates types",
			mc: config.ModuleComponents{
				"main-go":  &config.ComponentEntry{Type: "go", Root: "main/"},
				"other-go": &config.ComponentEntry{Type: "go", Root: "other/"},
			},
			expected: []string{"go"},
		},
		{
			name: "mixed explicit and implicit types",
			mc: config.ModuleComponents{
				"go":     &config.ComponentEntry{Root: "go/"},                               // implicit type: "go"
				"python": &config.ComponentEntry{Type: "testdata", Root: "testdata/python"}, // explicit type: "testdata"
			},
			expected: []string{"go", "testdata"},
		},
		{
			name: "nil entry uses name as type",
			mc: config.ModuleComponents{
				"assets": nil,
			},
			expected: []string{"assets"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mc.GetComponentTypes()
			// Sort for comparison since map iteration order is undefined
			sort.Strings(got)
			sort.Strings(tt.expected)
			if len(got) != len(tt.expected) {
				t.Errorf("GetComponentTypes() = %v, want %v", got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("GetComponentTypes() = %v, want %v", got, tt.expected)
					return
				}
			}
		})
	}
}

// TestModuleComponents_GetEnabled_ReturnsNames verifies that GetEnabled returns component names,
// not types. This is the key distinction that caused the Python dependency false positive bug.
func TestModuleComponents_GetEnabled_ReturnsNames(t *testing.T) {
	mc := config.ModuleComponents{
		"python": &config.ComponentEntry{Type: "testdata", Root: "testdata/python"},
	}

	enabled := mc.GetEnabled()
	if len(enabled) != 1 || enabled[0] != "python" {
		t.Errorf("GetEnabled() should return names ['python'], got %v", enabled)
	}

	// Verify GetComponentTypes returns the actual type
	types := mc.GetComponentTypes()
	if len(types) != 1 || types[0] != "testdata" {
		t.Errorf("GetComponentTypes() should return types ['testdata'], got %v", types)
	}
}
