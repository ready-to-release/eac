package manifests

import (
	"os"
	"path/filepath"
	"sort"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
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
