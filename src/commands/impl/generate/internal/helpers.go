package generate

import (
	"os"
	"path/filepath"
)

// FindModulesWithResults finds all module directories with cucumber.json results
func FindModulesWithResults(testRunDir string) ([]string, error) {
	entries, err := os.ReadDir(testRunDir)
	if err != nil {
		return nil, err
	}

	var modules []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this directory has cucumber.json
		cucumberPath := filepath.Join(testRunDir, entry.Name(), "cucumber.json")
		if _, err := os.Stat(cucumberPath); err == nil {
			modules = append(modules, entry.Name())
		}
	}

	return modules, nil
}
