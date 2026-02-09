//go:build L1
// +build L1

package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/core/domain/modules"
	eactesting "github.com/ready-to-release/eac/go/core/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReleaseTypeWorkflowValidation_PublishedModules tests that published SemVer modules
// are allowed to trigger release workflows and have changelogs.
// Published SemVer modules: clie, eac-ext (docs is CalVer, no changelog)
func TestReleaseTypeWorkflowValidation_PublishedModules(t *testing.T) {
	// Published SemVer modules that should be allowed to release
	publishedModules := []string{"clie", "eac-ext"}

	// Use mock registry with predictable data
	registryOpts := make([]eactesting.RegistryOption, 0, len(publishedModules))
	for _, moniker := range publishedModules {
		registryOpts = append(registryOpts,
			eactesting.WithModule(moniker,
				eactesting.WithVersioning(),
				eactesting.WithChangelog("release/"+moniker+"/CHANGELOG.md"),
				eactesting.WithReleaseType("published"),
			),
		)
	}
	moduleRegistry := eactesting.NewMockRegistry(registryOpts...)

	for _, moniker := range publishedModules {
		t.Run(moniker, func(t *testing.T) {
			moduleContract, exists := moduleRegistry.Get(moniker)
			require.True(t, exists, "mock module should always exist")

			releaseType := moduleContract.Versioning.ReleaseType
			assert.Equal(t, "published", releaseType,
				"module %s should have release_type: published", moniker)

			// Verify changelog is in release/ folder
			changelogPath := moduleContract.GetChangelogPath()
			assert.True(t, isInReleaseFolder(changelogPath),
				"published module %s should have changelog in release/ folder, got: %s", moniker, changelogPath)
		})
	}
}

// TestReleaseTypeWorkflowValidation_BundleModules tests that CalVer bundle modules
// have no changelog (auto-managed releases).
func TestReleaseTypeWorkflowValidation_BundleModules(t *testing.T) {
	// Bundle modules - clie-eac-bundle is CalVer, so no changelog
	bundleModules := []string{"clie-eac-bundle"}

	// Use mock registry with CalVer bundle (no changelog)
	registryOpts := make([]eactesting.RegistryOption, 0, len(bundleModules))
	for _, moniker := range bundleModules {
		registryOpts = append(registryOpts,
			eactesting.WithModule(moniker,
				eactesting.WithVersioning(), // WithVersioning defaults to CalVer
				eactesting.WithReleaseType("bundle"),
			),
		)
	}
	moduleRegistry := eactesting.NewMockRegistry(registryOpts...)

	for _, moniker := range bundleModules {
		t.Run(moniker, func(t *testing.T) {
			moduleContract, exists := moduleRegistry.Get(moniker)
			require.True(t, exists, "mock module should always exist")

			releaseType := moduleContract.Versioning.ReleaseType
			assert.Equal(t, "bundle", releaseType,
				"module %s should have release_type: bundle", moniker)

			// CalVer bundle has no changelog
			changelogPath := moduleContract.GetChangelogPath()
			assert.Empty(t, changelogPath,
				"CalVer bundle module %s should have no changelog, got: %s", moniker, changelogPath)
		})
	}
}

// TestReleaseTypeWorkflowValidation_AllModules tests that all modules with versioning
// have valid release_type values.
func TestReleaseTypeWorkflowValidation_AllModules(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	require.NoError(t, err)

	allModules := moduleRegistry.AllMonikers()
	validReleaseTypes := map[string]bool{
		"published": true,
		"internal":  true,
		"bundle":    true,
		"none":      true,
	}

	for _, moniker := range allModules {
		moduleContract, exists := moduleRegistry.Get(moniker)
		require.True(t, exists)

		// Skip modules without versioning
		if moduleContract.Versioning == nil {
			continue
		}

		releaseType := moduleContract.Versioning.ReleaseType

		// Every module with versioning should have a valid release_type
		assert.True(t, validReleaseTypes[releaseType],
			"module %s has invalid release_type: %s", moniker, releaseType)

		// Verify consistency with changelog location
		changelogPath := moduleContract.GetChangelogPath()
		inReleaseFolder := isInReleaseFolder(changelogPath)

		// CalVer modules have no changelogs regardless of release type
		if moduleContract.Versioning.Scheme == "CalVer" {
			assert.Empty(t, changelogPath,
				"CalVer module %s should have no changelog, got: %s", moniker, changelogPath)
			continue
		}

		switch releaseType {
		case "published", "bundle":
			assert.True(t, inReleaseFolder,
				"module %s with release_type: %s should have changelog in release/, got: %s",
				moniker, releaseType, changelogPath)
		case "internal":
			assert.False(t, inReleaseFolder,
				"module %s with release_type: internal should NOT have changelog in release/, got: %s",
				moniker, changelogPath)
		case "none":
			// No changelog location requirement for none
		}
	}
}

// TestReleaseTypeWorkflowValidation_OnlyPublishedHaveReleaseWorkflows tests that
// only published/bundle modules should have release workflows.
func TestReleaseTypeWorkflowValidation_OnlyPublishedHaveReleaseWorkflows(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	require.NoError(t, err)

	// Check which modules have release workflows
	workflowsDir := filepath.Join(workspaceRoot, ".github", "workflows")
	entries, err := os.ReadDir(workflowsDir)
	require.NoError(t, err)

	releaseWorkflows := make(map[string]bool)
	for _, entry := range entries {
		name := entry.Name()
		// Match release-{module}.yaml
		if len(name) > 8 && name[:8] == "release-" && (name[len(name)-5:] == ".yaml" || name[len(name)-4:] == ".yml") {
			// Extract module name
			moduleName := name[8:]
			if moduleName[len(moduleName)-5:] == ".yaml" {
				moduleName = moduleName[:len(moduleName)-5]
			} else if moduleName[len(moduleName)-4:] == ".yml" {
				moduleName = moduleName[:len(moduleName)-4]
			}
			releaseWorkflows[moduleName] = true
		}
	}

	t.Logf("Found release workflows for: %v", releaseWorkflows)

	// Check all modules
	allModules := moduleRegistry.AllMonikers()
	for _, moniker := range allModules {
		moduleContract, exists := moduleRegistry.Get(moniker)
		require.True(t, exists)

		if moduleContract.Versioning == nil {
			continue
		}

		releaseType := moduleContract.Versioning.ReleaseType
		hasWorkflow := releaseWorkflows[moniker]

		switch releaseType {
		case "published", "bundle":
			// Published/bundle modules should have release workflows
			if !hasWorkflow {
				t.Logf("WARNING: Module %s has release_type: %s but no release workflow", moniker, releaseType)
			}
		case "internal", "none":
			// Internal/none modules should NOT have release workflows
			if hasWorkflow {
				t.Errorf("Module %s has release_type: %s but has a release workflow (this should not happen)", moniker, releaseType)
			}
		}
	}
}

// isInReleaseFolder checks if a path starts with "release/" and is in a subdirectory.
// Only paths like release/<module>/CHANGELOG.md are in release folder.
func isInReleaseFolder(path string) bool {
	// Normalize path separators
	normalizedPath := filepath.ToSlash(path)

	// Check for release/ prefix and subdirectory
	if len(normalizedPath) < 8 || normalizedPath[:8] != "release/" {
		return false
	}

	// Must have at least two slashes (release/<module>/CHANGELOG.md)
	return strings.Count(normalizedPath, "/") >= 2
}
