package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ManifestPath returns the expected path for this manifest file.
// Format: {workspaceRoot}/out/{context}/{module}/{component}_{tool}/uow.manifest.json
func (m *UoWManifest) ManifestPath(workspaceRoot string) string {
	dirName := m.Component + "_" + m.Tool
	return filepath.Join(workspaceRoot, "out", string(m.Context), m.Module, dirName, "uow.manifest.json")
}

// Save writes the manifest to disk at its expected path.
// Creates the directory structure if it doesn't exist.
func (m *UoWManifest) Save(workspaceRoot string) error {
	manifestPath := m.ManifestPath(workspaceRoot)
	manifestDir := filepath.Dir(manifestPath)

	// Create directory structure
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("create manifest directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	// Write file
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	return nil
}

// Load reads and parses a manifest from the given path.
func Load(path string) (*UoWManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("manifest file is empty")
	}

	var manifest UoWManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	return &manifest, nil
}
