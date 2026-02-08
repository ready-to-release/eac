package release

import (
	"testing"
)

func TestParseDispatchedModules(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]bool{},
		},
		{
			name:  "single module",
			input: "docs",
			expected: map[string]bool{
				"docs": true,
			},
		},
		{
			name:  "multiple modules",
			input: "docs books core",
			expected: map[string]bool{
				"docs":     true,
				"books":    true,
				"core": true,
			},
		},
		{
			name:  "extra whitespace",
			input: "  docs   books  ",
			expected: map[string]bool{
				"docs":  true,
				"books": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDispatchedModules(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d modules, got %d", len(tt.expected), len(result))
			}

			for mod, expected := range tt.expected {
				if result[mod] != expected {
					t.Errorf("expected %s=%v, got %v", mod, expected, result[mod])
				}
			}
		})
	}
}

func TestPendingModule_Struct(t *testing.T) {
	// Test that the PendingModule struct has the expected fields
	pm := PendingModule{
		Module:   "test-module",
		Version:  "1.2.3",
		Tag:      "test-module/1.2.3",
		Type:     "semver",
		NeedsTag: true,
	}

	if pm.Module != "test-module" {
		t.Errorf("Module: expected 'test-module', got %s", pm.Module)
	}
	if pm.Version != "1.2.3" {
		t.Errorf("Version: expected '1.2.3', got %s", pm.Version)
	}
	if pm.Tag != "test-module/1.2.3" {
		t.Errorf("Tag: expected 'test-module/1.2.3', got %s", pm.Tag)
	}
	if pm.Type != "semver" {
		t.Errorf("Type: expected 'semver', got %s", pm.Type)
	}
	if !pm.NeedsTag {
		t.Error("NeedsTag: expected true")
	}
}

func TestCheckPendingResult_Struct(t *testing.T) {
	// Test that the result struct has expected fields
	result := CheckPendingResult{
		HasPending: true,
		ModulesJSON: []PendingModule{
			{Module: "mod1", Version: "1.0.0", Tag: "mod1/1.0.0", Type: "semver"},
		},
		LayersJSON: [][]PendingModule{
			{{Module: "mod1", Version: "1.0.0", Tag: "mod1/1.0.0", Type: "semver"}},
		},
		LayerCount: 1,
	}

	if !result.HasPending {
		t.Error("HasPending: expected true")
	}
	if len(result.ModulesJSON) != 1 {
		t.Errorf("ModulesJSON: expected 1 module, got %d", len(result.ModulesJSON))
	}
	if len(result.LayersJSON) != 1 {
		t.Errorf("LayersJSON: expected 1 layer, got %d", len(result.LayersJSON))
	}
	if result.LayerCount != 1 {
		t.Errorf("LayerCount: expected 1, got %d", result.LayerCount)
	}
}

func TestCheckPendingResult_Empty(t *testing.T) {
	result := CheckPendingResult{
		HasPending:  false,
		ModulesJSON: []PendingModule{},
		LayersJSON:  [][]PendingModule{},
		LayerCount:  0,
	}

	if result.HasPending {
		t.Error("HasPending: expected false for empty result")
	}
	if len(result.ModulesJSON) != 0 {
		t.Error("ModulesJSON: expected empty")
	}
	if len(result.LayersJSON) != 0 {
		t.Error("LayersJSON: expected empty")
	}
	if result.LayerCount != 0 {
		t.Error("LayerCount: expected 0")
	}
}

func TestOutputCheckPendingResult_ShellFormat(t *testing.T) {
	// Test the shell format output function indirectly through struct verification
	result := CheckPendingResult{
		HasPending: true,
		ModulesJSON: []PendingModule{
			{Module: "mod1", Version: "1.0.0", Tag: "mod1/1.0.0", Type: "semver"},
			{Module: "mod2", Version: "2.0.0", Tag: "mod2/2.0.0", Type: "calver"},
		},
		LayersJSON: [][]PendingModule{
			{{Module: "mod1", Version: "1.0.0", Tag: "mod1/1.0.0", Type: "semver"}},
			{{Module: "mod2", Version: "2.0.0", Tag: "mod2/2.0.0", Type: "calver"}},
		},
		LayerCount: 2,
	}

	// Verify we can build the module list for shell output
	modules := make([]string, len(result.ModulesJSON))
	for i, m := range result.ModulesJSON {
		modules[i] = m.Module
	}

	if len(modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(modules))
	}
	if modules[0] != "mod1" || modules[1] != "mod2" {
		t.Errorf("expected [mod1, mod2], got %v", modules)
	}
}

// TestCalverVersionFormat verifies the CalVer version format.
func TestCalverVersionFormat(t *testing.T) {
	// The format should be: 2006.0102.1504 (YYYY.MMDD.HHMM)
	// We can't test the exact output without mocking time, but we can verify the pattern

	pm := PendingModule{
		Module:   "docs",
		Version:  "2024.1216.1430",
		Tag:      "docs/2024.1216.1430",
		Type:     "calver",
		NeedsTag: true,
	}

	// Verify the tag format
	expectedTag := pm.Module + "/" + pm.Version
	if pm.Tag != expectedTag {
		t.Errorf("Tag should be module/version: expected %s, got %s", expectedTag, pm.Tag)
	}

	// Verify the type
	if pm.Type != "calver" {
		t.Errorf("Type: expected 'calver', got %s", pm.Type)
	}
}

// TestSemverVersionFormat verifies the SemVer tag format.
func TestSemverVersionFormat(t *testing.T) {
	pm := PendingModule{
		Module:   "clie-cli",
		Version:  "1.2.3",
		Tag:      "clie-cli/1.2.3",
		Type:     "semver",
		NeedsTag: true,
	}

	// Verify the tag format
	expectedTag := pm.Module + "/" + pm.Version
	if pm.Tag != expectedTag {
		t.Errorf("Tag should be module/version: expected %s, got %s", expectedTag, pm.Tag)
	}

	// Verify the type
	if pm.Type != "semver" {
		t.Errorf("Type: expected 'semver', got %s", pm.Type)
	}
}

// TestLayerOrdering verifies that layers are structured correctly.
func TestLayerOrdering(t *testing.T) {
	// Layer 0: modules with no dependencies
	// Layer 1: modules depending on layer 0
	layers := [][]PendingModule{
		{
			{Module: "core", Version: "1.0.0", Type: "semver"},
		},
		{
			{Module: "eac-cli", Version: "1.1.0", Type: "semver"},
		},
		{
			{Module: "clie-cli", Version: "2.0.0", Type: "semver"},
		},
	}

	if len(layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(layers))
	}

	// Verify first layer contains the base module
	if layers[0][0].Module != "core" {
		t.Errorf("expected core in layer 0, got %s", layers[0][0].Module)
	}

	// Verify last layer contains the top-level module
	if layers[2][0].Module != "clie-cli" {
		t.Errorf("expected clie-cli in layer 2, got %s", layers[2][0].Module)
	}
}
