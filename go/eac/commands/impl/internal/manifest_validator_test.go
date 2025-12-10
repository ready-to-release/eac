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

func TestVerifyDockerImage_WithRegistryTags(t *testing.T) {
	// Test that when an image has registry tags, it's considered valid
	// even if the local image doesn't exist (simulates CI push scenario)
	art := ArtifactInfo{
		Type:     "image",
		ID:       "image",
		Name:     "ext-eac:latest",
		Path:     "ext-eac:latest",
		Tags:     []string{"ghcr.io/ready-to-release/ext-eac:ci", "ghcr.io/ready-to-release/ext-eac:sha-abc1234"},
		Registry: "ghcr.io",
	}

	// Mock execCommand to simulate docker not finding local image
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(name string, args ...string) execCommandInterface {
		return &mockExecCmd{
			outputFn: func() ([]byte, error) {
				// Return empty - no local image found
				return []byte(""), nil
			},
			runFn: func() error {
				// Docker is available
				return nil
			},
		}
	}

	exists, errMsg := verifyDockerImageExists(art)
	if !exists {
		t.Errorf("image with registry tags should be considered valid: %s", errMsg)
	}
}

func TestVerifyDockerImage_LocalImage(t *testing.T) {
	art := ArtifactInfo{
		Type: "image",
		ID:   "image",
		Name: "ext-eac:local",
		Path: "ext-eac:local",
	}

	// Mock execCommand to simulate docker finding local image
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(name string, args ...string) execCommandInterface {
		return &mockExecCmd{
			outputFn: func() ([]byte, error) {
				// Return image ID - local image found
				return []byte("abc123def456\n"), nil
			},
			runFn: func() error {
				return nil
			},
		}
	}

	exists, errMsg := verifyDockerImageExists(art)
	if !exists {
		t.Errorf("local image should be found: %s", errMsg)
	}
}

func TestIsRegistryTag(t *testing.T) {
	tests := []struct {
		tag      string
		registry string
		expected bool
	}{
		{"ghcr.io/ready-to-release/ext-eac:ci", "ghcr.io", true},
		{"ghcr.io/ready-to-release/ext-eac:latest", "ghcr.io", true},
		{"ext-eac:local", "ghcr.io", false},
		{"ext-eac:latest", "", false},
		{"docker.io/library/nginx:latest", "docker.io", true},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := isRegistryTag(tt.tag, tt.registry)
			if result != tt.expected {
				t.Errorf("isRegistryTag(%q, %q) = %v, want %v", tt.tag, tt.registry, result, tt.expected)
			}
		})
	}
}

// mockExecCmd implements execCommandInterface for testing
type mockExecCmd struct {
	outputFn func() ([]byte, error)
	runFn    func() error
}

func (m *mockExecCmd) Output() ([]byte, error) {
	return m.outputFn()
}

func (m *mockExecCmd) Run() error {
	return m.runFn()
}

// TestVerifyArtifactsExist_CIPushedImage simulates the exact CI scenario:
// - buildx pushes image to ghcr.io (push=true, load=false)
// - local image doesn't exist
// - artifact has registry tags from docker_build config
// - validation should pass because image was pushed to registry
func TestVerifyArtifactsExist_CIPushedImage(t *testing.T) {
	// Create a manifest exactly like CI would produce for ext-eac
	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentCI,
		Moniker:    "ext-eac",
		Type:       "container",
		BuildTime:  time.Now(),
		GitCommit:  "abcdef1234567890abcdef1234567890abcdef12",
		Artifacts: []ArtifactInfo{
			{
				Type:     "image",
				ID:       "image",
				Name:     "ext-eac:latest",
				Path:     "ext-eac:latest", // Local reference from artifact pattern
				Tags:     []string{"ghcr.io/ready-to-release/ext-eac:ci", "ghcr.io/ready-to-release/ext-eac:sha-abc1234"},
				Registry: "ghcr.io",
			},
		},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	// Mock docker to simulate CI environment where local image doesn't exist
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(name string, args ...string) execCommandInterface {
		return &mockExecCmd{
			outputFn: func() ([]byte, error) {
				// Return empty for all docker images -q queries - no local images
				return []byte(""), nil
			},
			runFn: func() error {
				// Docker daemon is available
				return nil
			},
		}
	}

	// This should NOT fail - the image was pushed to registry
	err := manifest.VerifyArtifactsExist("/tmp/fake-build-dir")
	if err != nil {
		t.Errorf("CI pushed image should pass verification: %v", err)
	}
}

// TestVerifyArtifactsExist_LocalBuildImage simulates local dev scenario:
// - buildx builds with load=true
// - local image exists
// - no registry tags (local build)
func TestVerifyArtifactsExist_LocalBuildImage(t *testing.T) {
	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentDevbox,
		Moniker:    "ext-eac",
		Type:       "container",
		BuildTime:  time.Now(),
		GitCommit:  "abcdef1234567890abcdef1234567890abcdef12",
		Artifacts: []ArtifactInfo{
			{
				Type: "image",
				ID:   "image",
				Name: "ext-eac:local",
				Path: "ext-eac:local",
				// No Tags or Registry - local build
			},
		},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(name string, args ...string) execCommandInterface {
		return &mockExecCmd{
			outputFn: func() ([]byte, error) {
				// Return image ID - local image found
				return []byte("sha256:abc123\n"), nil
			},
			runFn: func() error {
				return nil
			},
		}
	}

	err := manifest.VerifyArtifactsExist("/tmp/fake-build-dir")
	if err != nil {
		t.Errorf("local build image should pass verification: %v", err)
	}
}

// TestVerifyArtifactsExist_MissingLocalImage tests that a local build without
// registry tags fails when the image doesn't exist locally
func TestVerifyArtifactsExist_MissingLocalImage(t *testing.T) {
	manifest := &ModuleManifest{
		BuildID:    "550e8400-e29b-41d4-a716-446655440000",
		BuildAgent: BuildAgentDevbox,
		Moniker:    "ext-eac",
		Type:       "container",
		BuildTime:  time.Now(),
		GitCommit:  "abcdef1234567890abcdef1234567890abcdef12",
		Artifacts: []ArtifactInfo{
			{
				Type: "image",
				ID:   "image",
				Name: "ext-eac:local",
				Path: "ext-eac:local",
				// No Tags or Registry - local build that should exist locally
			},
		},
		Platforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
		},
		Version: "2.0",
	}

	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(name string, args ...string) execCommandInterface {
		return &mockExecCmd{
			outputFn: func() ([]byte, error) {
				// Return empty - no local image
				return []byte(""), nil
			},
			runFn: func() error {
				return nil
			},
		}
	}

	err := manifest.VerifyArtifactsExist("/tmp/fake-build-dir")
	if err == nil {
		t.Error("missing local image without registry tags should fail verification")
	}

	// Verify it's the right error type
	if _, ok := err.(*ArtifactExistenceError); !ok {
		t.Errorf("expected ArtifactExistenceError, got %T", err)
	}
}
