package domain

import (
	"sort"
	"testing"
)

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

func TestModuleComponents_GetComponentTypes(t *testing.T) {
	tests := []struct {
		name     string
		mc       ModuleComponents
		expected []string
	}{
		{
			name:     "nil returns nil",
			mc:       nil,
			expected: nil,
		},
		{
			name:     "empty map returns empty slice",
			mc:       ModuleComponents{},
			expected: []string{},
		},
		{
			name: "name is type when Type field empty",
			mc: ModuleComponents{
				"go": &ComponentEntry{Root: "go/"},
			},
			expected: []string{"go"},
		},
		{
			name: "Type field overrides name",
			mc: ModuleComponents{
				"python": &ComponentEntry{Type: "testdata", Root: "testdata/python"},
			},
			expected: []string{"testdata"},
		},
		{
			name: "deduplicates types",
			mc: ModuleComponents{
				"main-go":  &ComponentEntry{Type: "go", Root: "main/"},
				"other-go": &ComponentEntry{Type: "go", Root: "other/"},
			},
			expected: []string{"go"},
		},
		{
			name: "mixed explicit and implicit types",
			mc: ModuleComponents{
				"go":     &ComponentEntry{Root: "go/"},                               // implicit type: "go"
				"python": &ComponentEntry{Type: "testdata", Root: "testdata/python"}, // explicit type: "testdata"
			},
			expected: []string{"go", "testdata"},
		},
		{
			name: "nil entry uses name as type",
			mc: ModuleComponents{
				"markdown": nil,
			},
			expected: []string{"markdown"},
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
	mc := ModuleComponents{
		"python": &ComponentEntry{Type: "testdata", Root: "testdata/python"},
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
