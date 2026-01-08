package manifests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
)

func TestLoadAllTestManifests_SingleModule(t *testing.T) {
	// Arrange: Create temp directory with one manifest
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "out", "test", "eac-core")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	manifest.AddTest(implinternal.TestEntry{
		Name:   "TestExample",
		Type:   "gotest",
		Suite:  "unit",
		Status: implinternal.TestStatusPassed,
	})
	if err := manifest.Save(testDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Act
	manifests, err := LoadAllTestManifests(tmpDir)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(manifests) != 1 {
		t.Errorf("Expected 1 manifest, got %d", len(manifests))
	}
	if len(manifests) > 0 && manifests[0].Moniker != "eac-core" {
		t.Errorf("Expected moniker 'eac-core', got '%s'", manifests[0].Moniker)
	}
}

func TestLoadAllTestManifests_MultipleModules(t *testing.T) {
	// Arrange: Create temp directory with multiple manifests
	tmpDir := t.TempDir()

	modules := []string{"eac-core", "eac-commands", "vscode-ext-commit"}
	for _, mod := range modules {
		testDir := filepath.Join(tmpDir, "out", "test", mod)
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test dir for %s: %v", mod, err)
		}

		manifest := implinternal.NewTestManifest(mod, "go", "abc123")
		manifest.AddTest(implinternal.TestEntry{
			Name:   "Test" + mod,
			Type:   "gotest",
			Suite:  "unit",
			Status: implinternal.TestStatusPassed,
		})
		if err := manifest.Save(testDir); err != nil {
			t.Fatalf("Failed to save manifest for %s: %v", mod, err)
		}
	}

	// Act
	manifests, err := LoadAllTestManifests(tmpDir)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(manifests) != 3 {
		t.Errorf("Expected 3 manifests, got %d", len(manifests))
	}

	// Verify all modules are present
	foundMonikers := make(map[string]bool)
	for _, m := range manifests {
		foundMonikers[m.Moniker] = true
	}
	for _, mod := range modules {
		if !foundMonikers[mod] {
			t.Errorf("Expected to find module '%s'", mod)
		}
	}
}

func TestLoadAllTestManifests_NoManifests(t *testing.T) {
	// Arrange: Empty directory
	tmpDir := t.TempDir()

	// Act
	manifests, err := LoadAllTestManifests(tmpDir)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for empty directory, got: %v", err)
	}
	if len(manifests) != 0 {
		t.Errorf("Expected 0 manifests, got %d", len(manifests))
	}
}

func TestLoadAllTestManifests_SkipsInvalidManifests(t *testing.T) {
	// Arrange: Create directory with one valid and one invalid manifest
	tmpDir := t.TempDir()

	// Valid manifest
	validDir := filepath.Join(tmpDir, "out", "test", "eac-core")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	manifest := implinternal.NewTestManifest("eac-core", "go", "abc123")
	if err := manifest.Save(validDir); err != nil {
		t.Fatalf("Failed to save valid manifest: %v", err)
	}

	// Invalid manifest (corrupted JSON)
	invalidDir := filepath.Join(tmpDir, "out", "test", "eac-commands")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid test dir: %v", err)
	}
	invalidPath := filepath.Join(invalidDir, "test.manifest.json")
	if err := os.WriteFile(invalidPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid manifest: %v", err)
	}

	// Act
	manifests, err := LoadAllTestManifests(tmpDir)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (should skip invalid), got: %v", err)
	}
	if len(manifests) != 1 {
		t.Errorf("Expected 1 valid manifest, got %d", len(manifests))
	}
	if len(manifests) > 0 && manifests[0].Moniker != "eac-core" {
		t.Errorf("Expected moniker 'eac-core', got '%s'", manifests[0].Moniker)
	}
}

func TestLoadAllTestManifests_SortsNewestFirst(t *testing.T) {
	// Arrange: Create manifests with different timestamps
	tmpDir := t.TempDir()

	// Oldest
	oldDir := filepath.Join(tmpDir, "out", "test", "module-old")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	oldManifest := implinternal.NewTestManifest("module-old", "go", "abc123")
	oldManifest.TestTime = time.Now().Add(-2 * time.Hour)
	if err := oldManifest.Save(oldDir); err != nil {
		t.Fatalf("Failed to save old manifest: %v", err)
	}

	// Newest
	newDir := filepath.Join(tmpDir, "out", "test", "module-new")
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	newManifest := implinternal.NewTestManifest("module-new", "go", "abc123")
	newManifest.TestTime = time.Now()
	if err := newManifest.Save(newDir); err != nil {
		t.Fatalf("Failed to save new manifest: %v", err)
	}

	// Act
	manifests, err := LoadAllTestManifests(tmpDir)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(manifests) != 2 {
		t.Errorf("Expected 2 manifests, got %d", len(manifests))
	}
	if len(manifests) >= 2 {
		// First should be newest
		if manifests[0].Moniker != "module-new" {
			t.Errorf("Expected first manifest to be 'module-new', got '%s'", manifests[0].Moniker)
		}
		// Second should be oldest
		if manifests[1].Moniker != "module-old" {
			t.Errorf("Expected second manifest to be 'module-old', got '%s'", manifests[1].Moniker)
		}
	}
}
