package manifests

import (
	"testing"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
)

func TestBuildControlSummary_SingleControl(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-commands", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 1",
		Type:   "godog",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:ai-2"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 2",
		Type:   "godog",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:ai-2"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert
	if len(summary) != 1 {
		t.Fatalf("Expected 1 control summary, got %d", len(summary))
	}

	entry := summary[0]
	if entry.ControlID != "ai-2" {
		t.Errorf("Expected control ID 'ai-2', got '%s'", entry.ControlID)
	}
	if entry.TestCount != 2 {
		t.Errorf("Expected 2 tests, got %d", entry.TestCount)
	}
	if entry.PassedCount != 2 {
		t.Errorf("Expected 2 passed, got %d", entry.PassedCount)
	}
	if entry.FailedCount != 0 {
		t.Errorf("Expected 0 failed, got %d", entry.FailedCount)
	}
	if entry.ModuleCount != 1 {
		t.Errorf("Expected 1 module, got %d", entry.ModuleCount)
	}
	if len(entry.Modules) != 1 || entry.Modules[0] != "eac-commands" {
		t.Errorf("Expected modules ['eac-commands'], got %v", entry.Modules)
	}
}

func TestBuildControlSummary_MultipleControls(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 1",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:au-2"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 2",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:au-3"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 3",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:au-12"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert
	if len(summary) != 3 {
		t.Errorf("Expected 3 control summaries, got %d", len(summary))
	}

	// Verify all controls present
	controls := make(map[string]bool)
	for _, entry := range summary {
		controls[entry.ControlID] = true
	}
	expectedControls := []string{"au-2", "au-3", "au-12"}
	for _, expected := range expectedControls {
		if !controls[expected] {
			t.Errorf("Expected to find control '%s'", expected)
		}
	}
}

func TestBuildControlSummary_AcrossModules(t *testing.T) {
	// Arrange
	manifest1 := implinternal.NewTestManifest("eac-commands", "go", "abc123")
	manifest1.AddTest(implinternal.TestEntry{
		Name:   "Test 1",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:ai-2"},
	})
	manifest1.AddTest(implinternal.TestEntry{
		Name:   "Test 2",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:ai-2"},
	})

	manifest2 := implinternal.NewTestManifest("vscode-ext-commit", "go", "abc123")
	manifest2.AddTest(implinternal.TestEntry{
		Name:   "Test 3",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:ai-2"},
	})

	manifests := []*implinternal.TestManifest{manifest1, manifest2}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert
	if len(summary) != 1 {
		t.Fatalf("Expected 1 control summary, got %d", len(summary))
	}

	entry := summary[0]
	if entry.ControlID != "ai-2" {
		t.Errorf("Expected control ID 'ai-2', got '%s'", entry.ControlID)
	}
	if entry.TestCount != 3 {
		t.Errorf("Expected 3 tests, got %d", entry.TestCount)
	}
	if entry.ModuleCount != 2 {
		t.Errorf("Expected 2 modules, got %d", entry.ModuleCount)
	}
	if len(entry.Modules) != 2 {
		t.Fatalf("Expected 2 modules in list, got %d", len(entry.Modules))
	}

	// Verify both modules present (order doesn't matter)
	moduleSet := make(map[string]bool)
	for _, m := range entry.Modules {
		moduleSet[m] = true
	}
	if !moduleSet["eac-commands"] {
		t.Error("Expected module 'eac-commands' in list")
	}
	if !moduleSet["vscode-ext-commit"] {
		t.Error("Expected module 'vscode-ext-commit' in list")
	}
}

func TestBuildControlSummary_WithFailures(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Passed test",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:au-2"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Failed test",
		Status: implinternal.TestStatusFailed,
		Tags:   []string{"@control:au-2"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Skipped test",
		Status: implinternal.TestStatusSkipped,
		Tags:   []string{"@control:au-2"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert
	if len(summary) != 1 {
		t.Fatalf("Expected 1 control summary, got %d", len(summary))
	}

	entry := summary[0]
	if entry.TestCount != 3 {
		t.Errorf("Expected 3 tests, got %d", entry.TestCount)
	}
	if entry.PassedCount != 1 {
		t.Errorf("Expected 1 passed, got %d", entry.PassedCount)
	}
	if entry.FailedCount != 1 {
		t.Errorf("Expected 1 failed, got %d", entry.FailedCount)
	}
	if entry.SkippedCount != 1 {
		t.Errorf("Expected 1 skipped, got %d", entry.SkippedCount)
	}
}

func TestBuildControlSummary_TestWithMultipleControlTags(t *testing.T) {
	// Arrange: Single test with multiple control tags
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Multi-control test",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:au-2", "@control:au-3"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert: Test should be counted for both controls
	if len(summary) != 2 {
		t.Fatalf("Expected 2 control summaries, got %d", len(summary))
	}

	// Both controls should have 1 test
	for _, entry := range summary {
		if entry.TestCount != 1 {
			t.Errorf("Expected control '%s' to have 1 test, got %d", entry.ControlID, entry.TestCount)
		}
		if entry.PassedCount != 1 {
			t.Errorf("Expected control '%s' to have 1 passed, got %d", entry.ControlID, entry.PassedCount)
		}
	}
}

func TestBuildControlSummary_NoControlTags(t *testing.T) {
	// Arrange: Tests without control tags
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test without controls",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@L1", "@deps:go"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert: No summaries should be generated
	if len(summary) != 0 {
		t.Errorf("Expected 0 control summaries, got %d", len(summary))
	}
}

func TestBuildControlSummary_EmptyManifests(t *testing.T) {
	// Arrange
	manifests := []*implinternal.TestManifest{}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert
	if len(summary) != 0 {
		t.Errorf("Expected 0 control summaries, got %d", len(summary))
	}
}

func TestBuildControlSummary_SortedByControlID(t *testing.T) {
	// Arrange: Controls added in non-alphabetical order
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 1",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:cm-2"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 2",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:ai-2"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:   "Test 3",
		Status: implinternal.TestStatusPassed,
		Tags:   []string{"@control:sc-7"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	summary := BuildControlSummary(manifests)

	// Assert: Should be sorted alphabetically
	if len(summary) != 3 {
		t.Fatalf("Expected 3 control summaries, got %d", len(summary))
	}

	expectedOrder := []string{"ai-2", "cm-2", "sc-7"}
	for i, expected := range expectedOrder {
		if summary[i].ControlID != expected {
			t.Errorf("Expected control '%s' at index %d, got '%s'", expected, i, summary[i].ControlID)
		}
	}
}
