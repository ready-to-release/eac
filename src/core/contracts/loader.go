package contracts

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EACConfigRelPath is the relative path from repo root to EAC repository configuration.
// This is where repository-specific EAC contracts are stored (modules, tags, environments).
// Note: This constant is duplicated here to avoid import cycle with repository package.
// The canonical definition is in repository.EACConfigRelPath.
const EACConfigRelPath = ".r2r/eac/repository"

// Loader handles loading contract YAML files
type Loader struct {
	workspaceRoot string
}

// NewLoader creates a new contract loader
func NewLoader(workspaceRoot string) *Loader {
	return &Loader{
		workspaceRoot: workspaceRoot,
	}
}

// GetWorkspaceRoot returns the workspace root directory
func (l *Loader) GetWorkspaceRoot() string {
	return l.workspaceRoot
}

// LoadYAML loads a YAML file into the target struct
func (l *Loader) LoadYAML(relativePath string, target interface{}) error {
	fullPath := filepath.Join(l.workspaceRoot, relativePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewContractError("load", fullPath, err, "contract not found")
		}
		return fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse YAML from %s: %w", fullPath, err)
	}

	return nil
}

// LoadYAMLPattern loads all YAML files matching a glob pattern
// The callback function is called for each matching file with its relative path
func (l *Loader) LoadYAMLPattern(pattern string, callback func(relPath string) error) error {
	// Construct full pattern path
	fullPattern := filepath.Join(l.workspaceRoot, pattern)

	// Find all matching files
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return fmt.Errorf("failed to match pattern %s: %w", pattern, err)
	}

	// Process each match
	for _, match := range matches {
		// Get relative path
		relPath, err := filepath.Rel(l.workspaceRoot, match)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", match, err)
		}

		// Call callback with relative path
		if err := callback(relPath); err != nil {
			return err
		}
	}

	return nil
}

// FileExists checks if a file exists at the given relative path
func (l *Loader) FileExists(relativePath string) bool {
	fullPath := filepath.Join(l.workspaceRoot, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}
