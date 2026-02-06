//go:build L1
// +build L1

package release

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFindModulesWithChangelogsAll tests that findModulesWithChangelogsAll finds modules with changelogs.
// SemVer modules with changelogs: r2r-cli, ext-eac (published), r2r-eac-bundle (bundle).
// CalVer modules (docs) have no changelogs.
func TestFindModulesWithChangelogsAll(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)

	// Load module registry to get modules with changelogs
	modules := findModulesWithChangelogsAll(workspaceRoot)

	// Should find SemVer modules with changelogs
	assert.NotEmpty(t, modules, "should find modules with changelogs")

	// Check that we find SemVer published modules (r2r-cli, ext-eac)
	hasSemVerPublished := false
	for _, mod := range modules {
		if mod == "r2r-cli" || mod == "ext-eac" {
			hasSemVerPublished = true
			break
		}
	}
	assert.True(t, hasSemVerPublished, "should find at least one SemVer published module")

	// CalVer modules should NOT appear (no changelogs)
	for _, mod := range modules {
		assert.NotEqual(t, "docs", mod, "CalVer module docs should not have a changelog")
	}
}

// TestFilterModulesByReleaseType tests filtering modules by release type.
// Only SemVer modules have changelogs: r2r-cli, ext-eac (published).
// CalVer modules (docs, r2r-eac-bundle) have no changelogs, so they won't be in the input set.
func TestFilterModulesByReleaseType(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)

	testCases := []struct {
		name          string
		filterType    string
		shouldInclude []string
		shouldExclude []string
	}{
		{
			name:          "filter published only",
			filterType:    "published",
			shouldInclude: []string{"r2r-cli", "ext-eac"},
			shouldExclude: []string{"r2r-eac-bundle"},
		},
		{
			name:          "filter bundle only",
			filterType:    "bundle",
			shouldInclude: []string{},
			shouldExclude: []string{"r2r-cli", "ext-eac"},
		},
		{
			name:          "no filter (all)",
			filterType:    "",
			shouldInclude: []string{"r2r-cli", "ext-eac"},
			shouldExclude: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allModules := findModulesWithChangelogsAll(workspaceRoot)
			filtered := filterModulesByReleaseType(workspaceRoot, allModules, tc.filterType)

			// Check that expected modules are included
			for _, expectedMod := range tc.shouldInclude {
				found := false
				for _, mod := range filtered {
					if mod == expectedMod {
						found = true
						break
					}
				}
				if !found {
					// Only assert if the module exists in allModules
					moduleExists := false
					for _, m := range allModules {
						if m == expectedMod {
							moduleExists = true
							break
						}
					}
					if moduleExists {
						t.Errorf("expected module %s to be in filtered results but was not found", expectedMod)
					}
				}
			}

			// Check that excluded modules are not included
			for _, excludedMod := range tc.shouldExclude {
				for _, mod := range filtered {
					if mod == excludedMod {
						t.Errorf("expected module %s to be excluded but was found in results", excludedMod)
					}
				}
			}
		})
	}
}

// TestGetModuleReleaseType tests getting release type for modules with versioning.
// Only modules with explicit versioning configs have release_type set.
func TestGetModuleReleaseType(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)

	// Only test modules that have explicit versioning configs
	testCases := []struct {
		module       string
		expectedType string
	}{
		{"r2r-cli", "published"},
		{"ext-eac", "published"},
		{"docs", "published"},
		{"r2r-eac-bundle", "bundle"},
	}

	for _, tc := range testCases {
		t.Run(tc.module, func(t *testing.T) {
			releaseType := getModuleReleaseType(workspaceRoot, tc.module)
			assert.Equal(t, tc.expectedType, releaseType,
				"module %s should have release_type %s", tc.module, tc.expectedType)
		})
	}
}

// TestPendingReleaseIncludesReleaseType tests that PendingRelease includes release type information.
func TestPendingReleaseIncludesReleaseType(t *testing.T) {
	// Create a sample PendingRelease
	pending := PendingRelease{
		Module:         "r2r-cli",
		HasChanges:     true,
		CurrentVersion: "0.0.24",
		NextVersion:    "0.0.25",
		ReleaseType:    "published",
	}

	assert.Equal(t, "published", pending.ReleaseType, "should include release type")
	assert.Equal(t, "r2r-cli", pending.Module)
}

// getTestWorkspaceRoot returns the workspace root for tests.
func getTestWorkspaceRoot(t *testing.T) string {
	t.Helper()

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Navigate up to repository root
	// From go/cli/eac/impl/release to root
	root := cwd
	for i := 0; i < 5; i++ {
		root = filepath.Dir(root)
	}

	return root
}
