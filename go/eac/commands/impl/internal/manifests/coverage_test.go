package manifests

import (
	"testing"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
)

func TestBuildSpecCoverage_SingleFeature(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-commands", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Scenario 1",
		Type:     "godog",
		Suite:    "integration",
		Status:   implinternal.TestStatusPassed,
		Tags:     []string{"@L2", "@control:ai-2"},
		FilePath: "specs/eac-commands/create-design/specification.feature",
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Scenario 2",
		Type:     "godog",
		Suite:    "integration",
		Status:   implinternal.TestStatusPassed,
		Tags:     []string{"@L2", "@control:ai-2"},
		FilePath: "specs/eac-commands/create-design/specification.feature",
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert
	if len(coverage) != 1 {
		t.Fatalf("Expected 1 coverage entry, got %d", len(coverage))
	}

	entry := coverage[0]
	if entry.FeatureName != "create-design" {
		t.Errorf("Expected feature name 'create-design', got '%s'", entry.FeatureName)
	}
	if entry.Module != "eac-commands" {
		t.Errorf("Expected module 'eac-commands', got '%s'", entry.Module)
	}
	if entry.ScenarioCount != 2 {
		t.Errorf("Expected 2 scenarios, got %d", entry.ScenarioCount)
	}
	if entry.PassedCount != 2 {
		t.Errorf("Expected 2 passed, got %d", entry.PassedCount)
	}
	if entry.FailedCount != 0 {
		t.Errorf("Expected 0 failed, got %d", entry.FailedCount)
	}
	if len(entry.Controls) != 1 {
		t.Fatalf("Expected 1 control, got %d", len(entry.Controls))
	}
	if entry.Controls[0] != "ai-2" {
		t.Errorf("Expected control 'ai-2', got '%s'", entry.Controls[0])
	}
}

func TestBuildSpecCoverage_MultipleFeatures(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-commands", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Scenario A",
		Type:     "godog",
		Status:   implinternal.TestStatusPassed,
		FilePath: "specs/eac-commands/create-design/specification.feature",
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Scenario B",
		Type:     "godog",
		Status:   implinternal.TestStatusPassed,
		FilePath: "specs/eac-commands/show-tests/specification.feature",
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert
	if len(coverage) != 2 {
		t.Errorf("Expected 2 coverage entries, got %d", len(coverage))
	}

	// Verify both features present
	features := make(map[string]bool)
	for _, entry := range coverage {
		features[entry.FeatureName] = true
	}
	if !features["create-design"] {
		t.Error("Expected to find 'create-design' feature")
	}
	if !features["show-tests"] {
		t.Error("Expected to find 'show-tests' feature")
	}
}

func TestBuildSpecCoverage_WithFailures(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Passed scenario",
		Type:     "godog",
		Status:   implinternal.TestStatusPassed,
		FilePath: "specs/eac-core/logging/specification.feature",
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Failed scenario",
		Type:     "godog",
		Status:   implinternal.TestStatusFailed,
		FilePath: "specs/eac-core/logging/specification.feature",
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Skipped scenario",
		Type:     "godog",
		Status:   implinternal.TestStatusSkipped,
		FilePath: "specs/eac-core/logging/specification.feature",
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert
	if len(coverage) != 1 {
		t.Fatalf("Expected 1 coverage entry, got %d", len(coverage))
	}

	entry := coverage[0]
	if entry.ScenarioCount != 3 {
		t.Errorf("Expected 3 scenarios, got %d", entry.ScenarioCount)
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

func TestBuildSpecCoverage_OnlyGodogTests(t *testing.T) {
	// Arrange: Mix of godog and gotest
	manifest := implinternal.NewTestManifest("eac-commands", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Godog scenario",
		Type:     "godog",
		Status:   implinternal.TestStatusPassed,
		FilePath: "specs/eac-commands/create-design/specification.feature",
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Go unit test",
		Type:     "gotest",
		Status:   implinternal.TestStatusPassed,
		FilePath: "go/eac/commands/impl/get/tests_test.go",
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert: Only godog tests should be included
	if len(coverage) != 1 {
		t.Errorf("Expected 1 coverage entry (godog only), got %d", len(coverage))
	}
}

func TestBuildSpecCoverage_MultipleControls(t *testing.T) {
	// Arrange
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Scenario 1",
		Type:     "godog",
		Status:   implinternal.TestStatusPassed,
		Tags:     []string{"@control:au-2"},
		FilePath: "specs/eac-core/logging/specification.feature",
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Scenario 2",
		Type:     "godog",
		Status:   implinternal.TestStatusPassed,
		Tags:     []string{"@control:au-3", "@control:au-12"},
		FilePath: "specs/eac-core/logging/specification.feature",
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert
	if len(coverage) != 1 {
		t.Fatalf("Expected 1 coverage entry, got %d", len(coverage))
	}

	entry := coverage[0]
	if len(entry.Controls) != 3 {
		t.Errorf("Expected 3 unique controls, got %d", len(entry.Controls))
	}

	// Verify all controls present (order doesn't matter)
	controlSet := make(map[string]bool)
	for _, c := range entry.Controls {
		controlSet[c] = true
	}
	expectedControls := []string{"au-2", "au-3", "au-12"}
	for _, expected := range expectedControls {
		if !controlSet[expected] {
			t.Errorf("Expected to find control '%s'", expected)
		}
	}
}

func TestBuildSpecCoverage_EmptyManifests(t *testing.T) {
	// Arrange
	manifests := []*implinternal.TestManifest{}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert
	if len(coverage) != 0 {
		t.Errorf("Expected 0 coverage entries, got %d", len(coverage))
	}
}

func TestBuildSpecCoverage_NoGodogTests(t *testing.T) {
	// Arrange: Only gotest tests
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "TestExample",
		Type:   "gotest",
		Status: implinternal.TestStatusPassed,
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert
	if len(coverage) != 0 {
		t.Errorf("Expected 0 coverage entries (no godog tests), got %d", len(coverage))
	}
}

func TestBuildSpecCoverage_WithEmptyFilePath(t *testing.T) {
	// Arrange - simulate container environment where FilePath might be empty
	// but Package field contains the spec path
	manifest := implinternal.NewTestManifest("eac-commands", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Test scenario 1",
		Package:  "../../../../../specs/eac-commands/create-design",
		Type:     "godog",
		Suite:    "integration",
		Status:   implinternal.TestStatusPassed,
		FilePath: "", // Empty in container
		Tags:     []string{"@L2", "@control:ai-2"},
	})
	manifest.AddTest(implinternal.TestEntry{
		Name:     "Test scenario 2",
		Package:  "../../../../../specs/eac-commands/create-design",
		Type:     "godog",
		Suite:    "integration",
		Status:   implinternal.TestStatusPassed,
		FilePath: "", // Empty in container
		Tags:     []string{"@L2", "@control:ai-3"},
	})

	manifests := []*implinternal.TestManifest{manifest}

	// Act
	coverage := BuildSpecCoverage(manifests)

	// Assert
	if len(coverage) != 1 {
		t.Fatalf("Expected 1 coverage entry, got %d", len(coverage))
	}

	entry := coverage[0]
	if entry.FeatureName != "create-design" {
		t.Errorf("Expected feature name 'create-design', got '%s'", entry.FeatureName)
	}
	if entry.ScenarioCount != 2 {
		t.Errorf("Expected 2 scenarios, got %d", entry.ScenarioCount)
	}
	if len(entry.Controls) != 2 {
		t.Errorf("Expected 2 controls, got %d", len(entry.Controls))
	}
}
