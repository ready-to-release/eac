package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateArtifactsExist(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a test artifact file (executable)
	artifactPath := filepath.Join(tmpDir, "out", "build", "test-cli", "test-cli.exe")
	os.MkdirAll(filepath.Dir(artifactPath), 0755)
	os.WriteFile(artifactPath, []byte("test"), 0644)

	manifest := &ModuleManifest{
		Moniker: "test-cli",
		Artifacts: []ArtifactInfo{
			{
				Type: "executable",
				ID:   "existing",
				Name: "test-cli.exe",
				Path: "out/build/test-cli/test-cli.exe",
			},
			{
				Type: "executable",
				ID:   "missing",
				Name: "missing.exe",
				Path: "out/build/test-cli/missing.exe",
			},
		},
	}

	missing := validateArtifactsExist(tmpDir, manifest)

	if len(missing) != 1 {
		t.Errorf("expected 1 missing artifact, got %d", len(missing))
	}
	if len(missing) > 0 && missing[0] != "missing" {
		t.Errorf("expected missing artifact 'missing', got '%s'", missing[0])
	}
}

func TestValidateArtifactsExist_AllPresent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create artifact
	artifactPath := filepath.Join(tmpDir, "out", "build", "test-module", "test-module.exe")
	os.MkdirAll(filepath.Dir(artifactPath), 0755)
	os.WriteFile(artifactPath, []byte("executable"), 0644)

	manifest := &ModuleManifest{
		Moniker: "test-module",
		Artifacts: []ArtifactInfo{
			{
				Type: "executable",
				ID:   "windows-amd64",
				Name: "test-module.exe",
				Path: "out/build/test-module/test-module.exe",
			},
		},
	}

	missing := validateArtifactsExist(tmpDir, manifest)

	if len(missing) != 0 {
		t.Errorf("expected no missing artifacts, got %v", missing)
	}
}

func TestValidateArtifactsExist_EmptyArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &ModuleManifest{
		Moniker:   "test-module",
		Artifacts: []ArtifactInfo{},
	}

	missing := validateArtifactsExist(tmpDir, manifest)

	if len(missing) != 0 {
		t.Errorf("expected no missing artifacts for empty list, got %v", missing)
	}
}

func TestLoadAndValidateModule_ManifestNotFound(t *testing.T) {
	// Reset global validator
	globalManifestValidator = nil

	validator, err := GetManifestValidator()
	if err != nil {
		t.Fatalf("failed to get validator: %v", err)
	}

	tmpDir := t.TempDir()

	// Create a minimal config mock
	result := loadAndValidateModule(tmpDir, "nonexistent", nil, validator)

	if result.Error == "" {
		t.Error("expected error for nonexistent manifest")
	}
	if result.Manifest != nil {
		t.Error("manifest should be nil for nonexistent")
	}
}

func TestManifestWithVerifiedUnchangedAt(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest directory
	manifestDir := filepath.Join(tmpDir, "out", "build", "test-module")
	os.MkdirAll(manifestDir, 0755)

	// Create a valid manifest
	manifest := &ModuleManifest{
		BuildID:             "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent:          BuildAgentDevbox,
		Moniker:             "test-module",
		Type:                "go-library",
		BuildTime:           time.Now(),
		GitCommit:           "abcdef1234567890abcdef1234567890abcdef12",
		VerifiedUnchangedAt: "1111111111111111111111111111111111111111",
		Artifacts:           []ArtifactInfo{},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	// Save it
	if err := manifest.Save(manifestDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Load it back
	loaded, err := LoadModuleManifest(manifestDir)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if loaded.VerifiedUnchangedAt != "1111111111111111111111111111111111111111" {
		t.Errorf("verified_unchanged_at not preserved: got %s", loaded.VerifiedUnchangedAt)
	}

	// Validate it
	validator, _ := GetManifestValidator()
	if err := validator.ValidateManifest(loaded); err != nil {
		t.Errorf("loaded manifest should be valid: %v", err)
	}
}
