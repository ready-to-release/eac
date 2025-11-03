package contracts

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader provides generic contract loading functionality
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

// LoadYAML loads and parses a YAML file into the provided structure
func (l *Loader) LoadYAML(relativePath string, target interface{}) error {
	fullPath := filepath.Join(l.workspaceRoot, relativePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return NewContractError("load", fullPath, err, "contract not found")
	}

	// Read file
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return NewContractError("load", fullPath, err, fmt.Sprintf("failed to read file: %v", err))
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, target); err != nil {
		return NewContractError("parse", fullPath, err, fmt.Sprintf("failed to parse YAML: %v", err))
	}

	return nil
}

// LoadYAMLPattern loads all YAML files matching a glob pattern
func (l *Loader) LoadYAMLPattern(pattern string, loader func(string) error) error {
	fullPattern := filepath.Join(l.workspaceRoot, pattern)

	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return NewContractError("glob", fullPattern, err, fmt.Sprintf("failed to glob pattern: %v", err))
	}

	if len(matches) == 0 {
		return NewContractError("glob", fullPattern, nil, "no files matched pattern")
	}

	for _, match := range matches {
		// Convert back to relative path
		relPath, err := filepath.Rel(l.workspaceRoot, match)
		if err != nil {
			return NewContractError("path", match, err, fmt.Sprintf("failed to compute relative path: %v", err))
		}

		if err := loader(relPath); err != nil {
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
