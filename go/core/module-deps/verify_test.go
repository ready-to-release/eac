package moduledeps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Verify — format validation (no filesystem needed)
// ---------------------------------------------------------------------------

func TestVerify_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "no prefix",
			input: "invalid",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "wrong prefix",
			input: "@dep:something",
		},
		{
			name:  "partial prefix",
			input: "@depm",
		},
		{
			name:  "similar but incorrect prefix",
			input: "depm:foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Verify(tt.input)

			assert.Equal(t, tt.input, result.Dependency, "Dependency field should echo back the input")
			assert.False(t, result.Available, "result should not be available for invalid format")
			assert.Error(t, result.Error, "result should contain an error for invalid format")
			assert.Contains(t, result.Error.Error(), "invalid module dependency format")
		})
	}
}

func TestVerify_EmptyMoniker(t *testing.T) {
	// "@depm:" with no moniker passes the prefix check but will fail at init()
	// because it cannot find a repository root in a test environment.
	result := Verify("@depm:")

	assert.Equal(t, "@depm:", result.Dependency)
	// The moniker is empty string, so the checker will be created with moniker=""
	// and init() will fail because we are not in a valid repo context.
	// That means Available=false.
	assert.False(t, result.Available)
}

func TestVerify_ValidPrefixButNonexistentModule(t *testing.T) {
	// Has valid @depm: prefix so no format error, but init() will fail
	// since we're not in a real repository context.
	result := Verify("@depm:nonexistent-module")

	assert.Equal(t, "@depm:nonexistent-module", result.Dependency)
	assert.False(t, result.Available)
	// No format error — the failure comes from init() / loadModuleContract
	assert.Nil(t, result.Error, "format is valid so Error should be nil; availability failure is signaled via Available=false")
}

// ---------------------------------------------------------------------------
// VerifyAll
// ---------------------------------------------------------------------------

func TestVerifyAll(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []string
		wantCount    int
	}{
		{
			name:         "multiple invalid format dependencies",
			dependencies: []string{"invalid-a", "invalid-b", "invalid-c"},
			wantCount:    3,
		},
		{
			name:         "empty slice",
			dependencies: []string{},
			wantCount:    0,
		},
		{
			name:         "mixed formats all producing results",
			dependencies: []string{"no-prefix", "@depm:nonexistent"},
			wantCount:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := VerifyAll(tt.dependencies)

			assert.Len(t, results, tt.wantCount)

			// Each result should have its Dependency field set to the input
			for i, r := range results {
				assert.Equal(t, tt.dependencies[i], r.Dependency)
			}
		})
	}
}

func TestVerifyAll_AllInvalidFormatsHaveErrors(t *testing.T) {
	deps := []string{"bad1", "bad2"}
	results := VerifyAll(deps)

	for _, r := range results {
		assert.False(t, r.Available)
		assert.Error(t, r.Error)
		assert.Contains(t, r.Error.Error(), "invalid module dependency format")
	}
}

func TestVerifyAll_PreservesDependencyOrder(t *testing.T) {
	deps := []string{"z-last", "a-first", "m-middle"}
	results := VerifyAll(deps)

	require.Len(t, results, 3)
	assert.Equal(t, "z-last", results[0].Dependency)
	assert.Equal(t, "a-first", results[1].Dependency)
	assert.Equal(t, "m-middle", results[2].Dependency)
}

// ---------------------------------------------------------------------------
// IsAvailable — wrapper around Verify
// ---------------------------------------------------------------------------

func TestIsAvailable_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "no prefix", input: "nope"},
		{name: "empty string", input: ""},
		{name: "wrong prefix", input: "@dep:foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, IsAvailable(tt.input))
		})
	}
}

func TestIsAvailable_ValidPrefixButUnavailableModule(t *testing.T) {
	// Valid prefix but the module cannot be resolved outside a real repo
	assert.False(t, IsAvailable("@depm:nonexistent"))
}

// ---------------------------------------------------------------------------
// GetMissingDependencies
// ---------------------------------------------------------------------------

func TestGetMissingDependencies(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []string
		wantMissing  []string
	}{
		{
			name:         "all invalid format means all missing",
			dependencies: []string{"bad-a", "bad-b"},
			wantMissing:  []string{"bad-a", "bad-b"},
		},
		{
			name:         "empty input returns empty",
			dependencies: []string{},
			wantMissing:  []string{},
		},
		{
			name:         "single invalid dependency",
			dependencies: []string{"no-prefix"},
			wantMissing:  []string{"no-prefix"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing := GetMissingDependencies(tt.dependencies)

			assert.Equal(t, tt.wantMissing, missing)
		})
	}
}

func TestGetMissingDependencies_NilInput(t *testing.T) {
	missing := GetMissingDependencies(nil)
	assert.Empty(t, missing)
}

// ---------------------------------------------------------------------------
// ModuleChecker.GetName
// ---------------------------------------------------------------------------

func TestModuleCheckerGetName(t *testing.T) {
	tests := []struct {
		name    string
		moniker string
		want    string
	}{
		{
			name:    "simple moniker",
			moniker: "eac-cli",
			want:    "Module: eac-cli",
		},
		{
			name:    "empty moniker",
			moniker: "",
			want:    "Module: ",
		},
		{
			name:    "moniker with hyphens",
			moniker: "my-cool-module",
			want:    "Module: my-cool-module",
		},
		{
			name:    "moniker with dots",
			moniker: "v1.2.3",
			want:    "Module: v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &ModuleChecker{moniker: tt.moniker}
			assert.Equal(t, tt.want, checker.GetName())
		})
	}
}

// ---------------------------------------------------------------------------
// ModuleChecker — Checker interface compliance
// ---------------------------------------------------------------------------

func TestModuleCheckerImplementsChecker(t *testing.T) {
	// Compile-time check that ModuleChecker satisfies the Checker interface.
	var _ Checker = (*ModuleChecker)(nil)
}

// ---------------------------------------------------------------------------
// ModuleChecker.checkSourceRootExists — filesystem tests
// ---------------------------------------------------------------------------

func TestCheckSourceRootExists_RootExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory that acts as a component root
	goRoot := filepath.Join(tmpDir, "go", "cli", "my-module")
	require.NoError(t, os.MkdirAll(goRoot, 0o755))

	checker := &ModuleChecker{
		moniker:  "my-module",
		repoRoot: tmpDir,
	}

	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "my-module",
		Components: domain.ModuleComponents{
			"go": &domain.ComponentEntry{
				Root: "go/cli/my-module",
			},
		},
	}, tmpDir)

	assert.True(t, checker.checkSourceRootExists(module))
}

func TestCheckSourceRootExists_NoRootsExist(t *testing.T) {
	tmpDir := t.TempDir()

	checker := &ModuleChecker{
		moniker:  "missing-module",
		repoRoot: tmpDir,
	}

	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "missing-module",
		Components: domain.ModuleComponents{
			"go": &domain.ComponentEntry{
				Root: "go/cli/does-not-exist",
			},
			"specs": &domain.ComponentEntry{
				Root: "specs/does-not-exist",
			},
		},
	}, tmpDir)

	assert.False(t, checker.checkSourceRootExists(module))
}

func TestCheckSourceRootExists_OneOfMultipleRootsExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Only create the specs root, not the go root
	specsRoot := filepath.Join(tmpDir, "specs", "my-module")
	require.NoError(t, os.MkdirAll(specsRoot, 0o755))

	checker := &ModuleChecker{
		moniker:  "my-module",
		repoRoot: tmpDir,
	}

	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "my-module",
		Components: domain.ModuleComponents{
			"go": &domain.ComponentEntry{
				Root: "go/cli/nonexistent",
			},
			"specs": &domain.ComponentEntry{
				Root: "specs/my-module",
			},
		},
	}, tmpDir)

	assert.True(t, checker.checkSourceRootExists(module))
}

func TestCheckSourceRootExists_NoComponents(t *testing.T) {
	tmpDir := t.TempDir()

	checker := &ModuleChecker{
		moniker:  "empty-module",
		repoRoot: tmpDir,
	}

	module := modules.NewModuleContract(domain.BaseContract{
		Moniker:    "empty-module",
		Components: domain.ModuleComponents{},
	}, tmpDir)

	assert.False(t, checker.checkSourceRootExists(module))
}

func TestCheckSourceRootExists_NilComponents(t *testing.T) {
	tmpDir := t.TempDir()

	checker := &ModuleChecker{
		moniker:  "nil-module",
		repoRoot: tmpDir,
	}

	module := modules.NewModuleContract(domain.BaseContract{
		Moniker:    "nil-module",
		Components: nil,
	}, tmpDir)

	assert.False(t, checker.checkSourceRootExists(module))
}

func TestCheckSourceRootExists_ComponentWithEmptyRoot(t *testing.T) {
	tmpDir := t.TempDir()

	checker := &ModuleChecker{
		moniker:  "empty-root-module",
		repoRoot: tmpDir,
	}

	// A component entry with an empty root is skipped by GetAllRoots
	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "empty-root-module",
		Components: domain.ModuleComponents{
			"go": &domain.ComponentEntry{
				Root: "",
			},
		},
	}, tmpDir)

	assert.False(t, checker.checkSourceRootExists(module))
}

func TestCheckSourceRootExists_RootIsFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file (not a directory) at the root path
	filePath := filepath.Join(tmpDir, "go", "cli")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte("not a dir"), 0o644))

	checker := &ModuleChecker{
		moniker:  "file-root-module",
		repoRoot: tmpDir,
	}

	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "file-root-module",
		Components: domain.ModuleComponents{
			"go": &domain.ComponentEntry{
				Root: "go/cli",
			},
		},
	}, tmpDir)

	// os.Stat succeeds for files too, so the function returns true
	// (it checks existence, not that it's a directory)
	assert.True(t, checker.checkSourceRootExists(module))
}

// ---------------------------------------------------------------------------
// ModuleChecker.checkAnyPlatformExists — filesystem tests
// ---------------------------------------------------------------------------

func TestCheckAnyPlatformExists_NoBuildDir(t *testing.T) {
	tmpDir := t.TempDir()

	checker := &ModuleChecker{
		moniker:  "test-mod",
		repoRoot: tmpDir,
	}

	artifact := config.Artifact{
		ID:        "test-bin",
		Type:      config.ArtifactTypeExecutable,
		Pattern:   "{moniker}-{os}-{arch}",
		Platforms: []string{"linux"},
	}

	// Build directory does not exist; no artifacts can be found.
	buildDir := filepath.Join(tmpDir, "build", "test-mod")
	result := checker.checkAnyPlatformExists(artifact, buildDir, nil)
	assert.False(t, result)
}

func TestCheckAnyPlatformExists_DefaultPlatforms(t *testing.T) {
	tmpDir := t.TempDir()

	checker := &ModuleChecker{
		moniker:  "test-mod",
		repoRoot: tmpDir,
	}

	// No Platforms specified — the function defaults to linux, windows, darwin
	artifact := config.Artifact{
		ID:      "test-bin",
		Type:    config.ArtifactTypeExecutable,
		Pattern: "{moniker}-{os}-{arch}",
	}

	buildDir := filepath.Join(tmpDir, "build", "test-mod")
	result := checker.checkAnyPlatformExists(artifact, buildDir, nil)
	assert.False(t, result, "should be false when build directory does not exist")
}

func TestCheckAnyPlatformExists_MultiplePlatformsNoneExist(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build", "test-mod")
	require.NoError(t, os.MkdirAll(buildDir, 0o755))

	checker := &ModuleChecker{
		moniker:  "test-mod",
		repoRoot: tmpDir,
	}

	artifact := config.Artifact{
		ID:        "test-bin",
		Type:      config.ArtifactTypeExecutable,
		Pattern:   "{moniker}-{os}-{arch}",
		Platforms: []string{"linux", "darwin"},
	}

	result := checker.checkAnyPlatformExists(artifact, buildDir, nil)
	assert.False(t, result, "should be false when no artifact files exist on disk")
}

// ---------------------------------------------------------------------------
// Result struct zero value
// ---------------------------------------------------------------------------

func TestResultZeroValue(t *testing.T) {
	var r Result

	assert.Empty(t, r.Dependency)
	assert.False(t, r.Available)
	assert.Empty(t, r.Version)
	assert.NoError(t, r.Error)
}
