//go:build L1
// +build L1

package release

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFindModulesWithChangelogsAll tests that findModulesWithChangelogsAll finds both published and internal modules.
func TestFindModulesWithChangelogsAll(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)

	// Load module registry to get modules with changelogs
	modules := findModulesWithChangelogsAll(workspaceRoot)

	// Should find both published and internal modules
	assert.NotEmpty(t, modules, "should find modules with changelogs")

	// Check that we find published modules
	hasPublished := false
	for _, mod := range modules {
		if mod == "r2r-cli" || mod == "docs" || mod == "books" {
			hasPublished = true
			break
		}
	}
	assert.True(t, hasPublished, "should find at least one published module")

	// Check that we find internal modules
	hasInternal := false
	for _, mod := range modules {
		if mod == "eac-cli" || mod == "mcp-server" {
			hasInternal = true
			break
		}
	}
	assert.True(t, hasInternal, "should find at least one internal module")
}

// TestFilterModulesByReleaseType tests filtering modules by release type.
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
			shouldInclude: []string{"r2r-cli", "docs", "books", "ext-eac"},
			shouldExclude: []string{"eac-cli", "mcp-server"},
		},
		{
			name:          "filter internal only",
			filterType:    "internal",
			shouldInclude: []string{"eac-cli", "mcp-server", "r2r-installer", "vscode-commit"},
			shouldExclude: []string{"r2r-cli", "docs", "books"},
		},
		{
			name:          "filter bundle only",
			filterType:    "bundle",
			shouldInclude: []string{"r2r-eac-bundle"},
			shouldExclude: []string{"r2r-cli", "eac-cli"},
		},
		{
			name:          "no filter (all)",
			filterType:    "",
			shouldInclude: []string{"r2r-cli", "eac-cli", "docs", "r2r-eac-bundle"},
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

// TestGetModuleReleaseType tests getting release type for a module.
func TestGetModuleReleaseType(t *testing.T) {
	workspaceRoot := getTestWorkspaceRoot(t)

	testCases := []struct {
		module       string
		expectedType string
	}{
		{"r2r-cli", "published"},
		{"ext-eac", "published"},
		{"docs", "published"},
		{"books", "published"},
		{"r2r-eac-bundle", "bundle"},
		{"eac-cli", "internal"},
		{"mcp-server", "internal"},
		{"r2r-installer", "internal"},
		{"vscode-commit", "internal"},
		// core no longer has explicit versioning config, so it gets default "internal"
		{"core", "internal"},
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
