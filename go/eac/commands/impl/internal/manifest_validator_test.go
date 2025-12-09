package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// getTestWorkspaceRoot finds the workspace root for tests
func getTestWorkspaceRoot(t *testing.T) string {
	t.Helper()
	// Walk up from current directory to find .r2r
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".r2r")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find workspace root from %s", cwd)
		}
		dir = parent
	}
}

func TestManifestValidator_ValidManifest(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentDevbox,
		Moniker:    "test-module",
		Type:       "go",
		BuildTime:  time.Now(),
		GitCommit:  "abcdef1234567890abcdef1234567890abcdef12",
		Artifacts:  []ArtifactInfo{}, // go module with no artifacts (library)
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	err = validator.ValidateManifest(manifest)
	if err != nil {
		t.Errorf("valid manifest should pass validation: %v", err)
	}
}

func TestManifestValidator_MissingRequiredField(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Missing moniker
	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentDevbox,
		Type:       "go",
		BuildTime:  time.Now(),
		Artifacts:  []ArtifactInfo{},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	err = validator.ValidateManifest(manifest)
	if err == nil {
		t.Error("manifest missing moniker should fail validation")
	}
}

func TestManifestValidator_InvalidPlatform(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentDevbox,
		Moniker:    "test-module",
		Type:       "go",
		BuildTime:  time.Now(),
		Artifacts:  []ArtifactInfo{},
		Platforms: []PlatformInfo{
			{OS: "invalid-os", Arch: "amd64"},
		},
		Version: "2.0",
	}

	err = validator.ValidateManifest(manifest)
	if err == nil {
		t.Error("manifest with invalid platform should fail validation")
	}
}

func TestManifestValidator_InvalidVersion(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentDevbox,
		Moniker:    "test-module",
		Type:       "go",
		BuildTime:  time.Now(),
		Artifacts:  []ArtifactInfo{},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "1.0", // Wrong version
	}

	err = validator.ValidateManifest(manifest)
	if err == nil {
		t.Error("manifest with wrong version should fail validation")
	}
}

func TestManifestValidator_ValidArtifactWithPlatform(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentCI,
		Moniker:    "test-cli",
		Type:       "go",
		BuildTime:  time.Now(),
		GitCommit:  "abcdef1234567890abcdef1234567890abcdef12",
		Artifacts: []ArtifactInfo{
			{
				Type:     "executable",
				ID:       "windows-amd64",
				Name:     "test-cli-windows-amd64.exe",
				Path:     "out/build/test-cli/test-cli-windows-amd64.exe",
				Platform: "windows-amd64",
			},
		},
		Platforms: []PlatformInfo{
			{OS: "windows", Arch: "amd64"},
		},
		Version: "2.0",
	}

	err = validator.ValidateManifest(manifest)
	if err != nil {
		t.Errorf("valid manifest with platform artifact should pass: %v", err)
	}
}

func TestManifestValidator_WithVerifiedUnchangedAt(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	manifest := &ModuleManifest{
		BuildID:             "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent:          BuildAgentDevbox,
		Moniker:             "test-module",
		Type:                "go",
		BuildTime:           time.Now(),
		GitCommit:           "abcdef1234567890abcdef1234567890abcdef12",
		VerifiedUnchangedAt: "1234567890abcdef1234567890abcdef12345678",
		Artifacts:           []ArtifactInfo{},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	err = validator.ValidateManifest(manifest)
	if err != nil {
		t.Errorf("manifest with verified_unchanged_at should pass: %v", err)
	}
}

func TestManifestValidator_MissingBuildAgent(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Missing BuildAgent (required field)
	manifest := &ModuleManifest{
		BuildID:   "550e8400-e29b-41d4-a716-446655440000",
		Moniker:   "test-module",
		Type:      "go",
		BuildTime: time.Now(),
		Artifacts: []ArtifactInfo{},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	err = validator.ValidateManifest(manifest)
	if err == nil {
		t.Error("manifest missing build_agent should fail validation")
	}
}

func TestManifestValidator_InvalidBuildAgent(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)
	validator, err := NewManifestValidator(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: "invalid-agent", // Not "ci" or "devbox"
		Moniker:    "test-module",
		Type:       "go",
		BuildTime:  time.Now(),
		Artifacts:  []ArtifactInfo{},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	err = validator.ValidateManifest(manifest)
	if err == nil {
		t.Error("manifest with invalid build_agent should fail validation")
	}
}

func TestGetManifestValidator_Singleton(t *testing.T) {
	// Reset global for test
	globalManifestValidator = nil
	globalManifestValidatorRoot = ""

	v1, err := GetManifestValidator()
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	v2, err := GetManifestValidator()
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if v1 != v2 {
		t.Error("GetManifestValidator should return the same instance")
	}
}
