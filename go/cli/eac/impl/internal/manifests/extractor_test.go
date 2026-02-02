package manifests

import (
	"testing"

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
)

func TestExtractTestResults_SingleTest(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:       "TestExample",
		Package:    "go/eac/core/testing",
		Type:       "gotest",
		Suite:      "unit",
		Status:     implinternal.TestStatusPassed,
		DurationMs: 123,
		Tags:       []string{"@L1", "@deps:go"},
		FilePath:   "go/eac/core/testing/example_test.go",
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	results := ExtractTestResults(manifests)

	// Assert
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Name != "TestExample" {
		t.Errorf("Expected name 'TestExample', got '%s'", result.Name)
	}
	if result.Module != "core" {
		t.Errorf("Expected module 'core', got '%s'", result.Module)
	}
	if result.Type != "gotest" {
		t.Errorf("Expected type 'gotest', got '%s'", result.Type)
	}
	if result.Suite != "unit" {
		t.Errorf("Expected suite 'unit', got '%s'", result.Suite)
	}
	if result.Status != "passed" {
		t.Errorf("Expected status 'passed', got '%s'", result.Status)
	}
	if result.DurationMs != 123 {
		t.Errorf("Expected duration 123, got %d", result.DurationMs)
	}
	if len(result.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(result.Tags))
	}
}

func TestExtractTestResults_GodogTest_ExtractsControlTags(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-cli", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:       "Generate message from commits",
		Package:    "specs/eac-cli/create-squash-message",
		Type:       "godog",
		Suite:      "integration",
		Status:     implinternal.TestStatusPassed,
		DurationMs: 1234,
		Tags:       []string{"@L2", "@control:ai-2", "@deps:go", "@control:sa-3"},
		FilePath:   "specs/eac-cli/create-squash-message/specification.feature",
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	results := ExtractTestResults(manifests)

	// Assert
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Type != "godog" {
		t.Errorf("Expected type 'godog', got '%s'", result.Type)
	}

	// Check control tags extracted
	if len(result.ControlTags) != 2 {
		t.Fatalf("Expected 2 control tags, got %d", len(result.ControlTags))
	}
	expectedControls := map[string]bool{"ai-2": true, "sa-3": true}
	for _, control := range result.ControlTags {
		if !expectedControls[control] {
			t.Errorf("Unexpected control tag: %s", control)
		}
	}

	// Check feature name extracted
	if result.FeatureName != "create-squash-message" {
		t.Errorf("Expected feature name 'create-squash-message', got '%s'", result.FeatureName)
	}

	// Check feature path
	if result.FeaturePath != "specs/eac-cli/create-squash-message/specification.feature" {
		t.Errorf("Expected feature path set, got '%s'", result.FeaturePath)
	}
}

func TestExtractTestResults_MultipleManifests(t *testing.T) {
	// Arrange
	manifest1 := implinternal.NewTestManifest("core", "go", "abc123")
	manifest1.AddTest(implinternal.TestEntry{
		Name:   "Test1",
		Type:   "gotest",
		Suite:  "unit",
		Status: implinternal.TestStatusPassed,
	})
	manifest1.AddTest(implinternal.TestEntry{
		Name:   "Test2",
		Type:   "gotest",
		Suite:  "unit",
		Status: implinternal.TestStatusPassed,
	})

	manifest2 := implinternal.NewTestManifest("eac-cli", "go", "abc123")
	manifest2.AddTest(implinternal.TestEntry{
		Name:   "Test3",
		Type:   "godog",
		Suite:  "integration",
		Status: implinternal.TestStatusPassed,
	})

	manifests := []*implinternal.TestManifest{manifest1, manifest2}

	// Act
	results := ExtractTestResults(manifests)

	// Assert
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify modules are correct
	moduleCount := make(map[string]int)
	for _, r := range results {
		moduleCount[r.Module]++
	}
	if moduleCount["core"] != 2 {
		t.Errorf("Expected 2 tests from core, got %d", moduleCount["core"])
	}
	if moduleCount["eac-cli"] != 1 {
		t.Errorf("Expected 1 test from eac-cli, got %d", moduleCount["eac-cli"])
	}
}

func TestExtractTestResults_EmptyManifests(t *testing.T) {
	// Arrange
	manifests := []*implinternal.TestManifest{}

	// Act
	results := ExtractTestResults(manifests)

	// Assert
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestExtractTestResults_NoControlTags(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "TestWithoutControls",
		Type:   "gotest",
		Suite:  "unit",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@L1", "@deps:go"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	results := ExtractTestResults(manifests)

	// Assert
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if len(result.ControlTags) != 0 {
		t.Errorf("Expected 0 control tags, got %d", len(result.ControlTags))
	}
}

func TestExtractControlTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected []string
	}{
		{
			name:     "Single control tag",
			tags:     []string{"@L2", "@control:ai-2", "@deps:go"},
			expected: []string{"ai-2"},
		},
		{
			name:     "Multiple control tags",
			tags:     []string{"@control:au-2", "@L1", "@control:au-3", "@control:au-12"},
			expected: []string{"au-2", "au-3", "au-12"},
		},
		{
			name:     "No control tags",
			tags:     []string{"@L1", "@deps:go", "@ov"},
			expected: []string{},
		},
		{
			name:     "Empty tags",
			tags:     []string{},
			expected: []string{},
		},
		{
			name:     "Malformed control tag (no value)",
			tags:     []string{"@control:", "@L1"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractControlTags(tt.tags)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d control tags, got %d", len(tt.expected), len(result))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected control tag '%s' at index %d, got '%s'", expected, i, result[i])
				}
			}
		})
	}
}

func TestExtractFeatureName(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedName string
	}{
		{
			name:         "Standard feature path",
			filePath:     "specs/eac-cli/create-design/specification.feature",
			expectedName: "create-design",
		},
		{
			name:         "Feature with underscores",
			filePath:     "specs/core/logging_audit/specification.feature",
			expectedName: "logging_audit",
		},
		{
			name:         "Feature with custom name",
			filePath:     "specs/eac-cli/show-tests/custom.feature",
			expectedName: "show-tests",
		},
		{
			name:         "Non-feature path",
			filePath:     "go/eac/core/testing/example_test.go",
			expectedName: "",
		},
		{
			name:         "Empty path",
			filePath:     "",
			expectedName: "",
		},
		{
			name:         "Relative path from cucumber.json",
			filePath:     "../../../../../specs/eac-cli/create-design/specification.feature",
			expectedName: "create-design",
		},
		{
			name:         "Absolute Windows path",
			filePath:     "C:\\source\\ready-to-release\\eac\\specs\\eac-cli\\get-commands\\specification.feature",
			expectedName: "get-commands",
		},
		{
			name:         "Absolute Linux path",
			filePath:     "/app/specs/eac-cli/work-merge/specification.feature",
			expectedName: "work-merge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFeatureName(tt.filePath)
			if result != tt.expectedName {
				t.Errorf("Expected feature name '%s', got '%s'", tt.expectedName, result)
			}
		})
	}
}
