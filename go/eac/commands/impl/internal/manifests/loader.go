package manifests

import (
	"os"
	"path/filepath"
	"sort"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
)

// LoadAllTestManifests loads all test.manifest.json files from out/test/ directory.
// Returns manifests sorted by test time (newest first).
// Skips invalid or corrupted manifests with a warning.
func LoadAllTestManifests(repoRoot string) ([]*implinternal.TestManifest, error) {
	testOutputDir := filepath.Join(repoRoot, "out", "test")

	// Check if test output directory exists
	if _, err := os.Stat(testOutputDir); os.IsNotExist(err) {
		return []*implinternal.TestManifest{}, nil
	}

	var manifests []*implinternal.TestManifest

	// Walk through out/test/ directory
	entries, err := os.ReadDir(testOutputDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moduleTestDir := filepath.Join(testOutputDir, entry.Name())
		manifestPath := implinternal.GetTestManifestPath(moduleTestDir)

		// Check if manifest file exists
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		// Try to load manifest
		manifest, err := implinternal.LoadTestManifest(moduleTestDir)
		if err != nil {
			// Log warning but continue processing other manifests
			log.Warnf("Skipping invalid manifest at %s: %v", manifestPath, err)
			continue
		}

		manifests = append(manifests, manifest)
	}

	// Sort by test time (newest first)
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].TestTime.After(manifests[j].TestTime)
	})

	return manifests, nil
}

// LoadTestManifestsForModules loads test manifests for specific modules.
// Automatically includes test manifests from all transitive dependencies.
// For example, if ext-eac depends on eac-commands, loading manifests for ext-eac
// will also include manifests from eac-commands.
// Returns manifests sorted by test time (newest first).
// Skips modules without manifests or with invalid manifests with a warning.
func LoadTestManifestsForModules(repoRoot string, monikers []string) ([]*implinternal.TestManifest, error) {
	testOutputDir := filepath.Join(repoRoot, "out", "test")

	// Check if test output directory exists
	if _, err := os.Stat(testOutputDir); os.IsNotExist(err) {
		return []*implinternal.TestManifest{}, nil
	}

	// Expand monikers to include all transitive dependencies
	expandedMonikers, err := expandMonikersWithDependencies(repoRoot, monikers)
	if err != nil {
		// If we can't load the module registry, fall back to just the specified monikers
		log.Warnf("Failed to expand dependencies: %v, using specified modules only", err)
		expandedMonikers = monikers
	}

	var manifests []*implinternal.TestManifest

	// Load manifest for each module (original + dependencies)
	for _, moniker := range expandedMonikers {
		moduleTestDir := filepath.Join(testOutputDir, moniker)
		manifestPath := implinternal.GetTestManifestPath(moduleTestDir)

		// Check if manifest file exists
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			// Silently skip - not all modules have test manifests (e.g., container modules)
			continue
		}

		// Try to load manifest
		manifest, err := implinternal.LoadTestManifest(moduleTestDir)
		if err != nil {
			// Log warning but continue processing other manifests
			log.Warnf("Skipping invalid manifest for module %s: %v", moniker, err)
			continue
		}

		manifests = append(manifests, manifest)
	}

	// Sort by test time (newest first)
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].TestTime.After(manifests[j].TestTime)
	})

	return manifests, nil
}

// expandMonikersWithDependencies expands a list of module monikers to include
// all their transitive dependencies. Returns a deduplicated list in no particular order.
func expandMonikersWithDependencies(repoRoot string, monikers []string) ([]string, error) {
	// Load module registry
	registry, err := modules.LoadFromWorkspace(repoRoot)
	if err != nil {
		return nil, err
	}

	// Build set of all modules to include
	allModules := make(map[string]bool)
	for _, moniker := range monikers {
		allModules[moniker] = true
		// Recursively add all dependencies
		if err := addDependenciesRecursive(moniker, registry, allModules); err != nil {
			return nil, err
		}
	}

	// Convert to slice
	result := make([]string, 0, len(allModules))
	for moniker := range allModules {
		result = append(result, moniker)
	}

	return result, nil
}

// addDependenciesRecursive recursively adds all dependencies of a module.
func addDependenciesRecursive(moniker string, registry *modules.Registry, result map[string]bool) error {
	module, exists := registry.Get(moniker)
	if !exists {
		// Module not found - just log and continue (might be a valid case)
		log.Warnf("Module '%s' not found in registry", moniker)
		return nil
	}

	for _, dep := range module.DependsOn {
		// Check if dependency exists in registry
		if _, exists := registry.Get(dep); !exists {
			log.Warnf("Dependency '%s' of module '%s' not found in registry", dep, moniker)
			continue
		}

		if !result[dep] {
			result[dep] = true
			if err := addDependenciesRecursive(dep, registry, result); err != nil {
				return err
			}
		}
	}

	return nil
}
