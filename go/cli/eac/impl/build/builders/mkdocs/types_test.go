package mkdocs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifestStore_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "mkdocs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewManifestStore(tmpDir)

	// Create a test manifest
	manifest := &ComponentManifest{
		Version:    ManifestVersion,
		Type:       ComponentTypeBaseSite,
		Module:     "docs",
		Component:  "base-site",
		InputHash:  "abc123def456",
		OutputHash: "789xyz000111",
		Timestamp:  time.Now().UTC().Truncate(time.Second),
		Duration:   5 * time.Second,
		Artifacts:  []string{"staged/", "mkdocs.yml"},
		Metadata: map[string]string{
			"book": "site",
		},
	}

	// Save the manifest
	if err := store.Save(manifest); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Verify file was created
	manifestPath := store.ManifestPath("docs", "base-site")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatalf("manifest file not created at %s", manifestPath)
	}

	// Load the manifest
	loaded, err := store.Load("docs", "base-site")
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	// Verify loaded manifest matches
	if loaded.Version != manifest.Version {
		t.Errorf("version mismatch: got %s, want %s", loaded.Version, manifest.Version)
	}
	if loaded.Type != manifest.Type {
		t.Errorf("type mismatch: got %s, want %s", loaded.Type, manifest.Type)
	}
	if loaded.Module != manifest.Module {
		t.Errorf("module mismatch: got %s, want %s", loaded.Module, manifest.Module)
	}
	if loaded.Component != manifest.Component {
		t.Errorf("component mismatch: got %s, want %s", loaded.Component, manifest.Component)
	}
	if loaded.InputHash != manifest.InputHash {
		t.Errorf("input hash mismatch: got %s, want %s", loaded.InputHash, manifest.InputHash)
	}
	if loaded.OutputHash != manifest.OutputHash {
		t.Errorf("output hash mismatch: got %s, want %s", loaded.OutputHash, manifest.OutputHash)
	}
	if len(loaded.Artifacts) != len(manifest.Artifacts) {
		t.Errorf("artifacts length mismatch: got %d, want %d", len(loaded.Artifacts), len(manifest.Artifacts))
	}
}

func TestManifestStore_LoadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mkdocs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewManifestStore(tmpDir)

	// Load non-existent manifest should return nil, nil
	loaded, err := store.Load("nonexistent", "component")
	if err != nil {
		t.Errorf("expected nil error for non-existent manifest, got: %v", err)
	}
	if loaded != nil {
		t.Errorf("expected nil manifest for non-existent, got: %+v", loaded)
	}
}

func TestManifestPath(t *testing.T) {
	store := NewManifestStore("/workspace")

	path := store.ManifestPath("docs", "base-site")
	expected := filepath.Join("/workspace", ".cache", "eac", "components", "docs", "base-site", ".manifest.json")
	if path != expected {
		t.Errorf("manifest path mismatch:\n  got:  %s\n  want: %s", path, expected)
	}
}

func TestComponentTypes(t *testing.T) {
	tests := []struct {
		name     string
		ct       ComponentType
		expected string
	}{
		{"base-site", ComponentTypeBaseSite, "base-site"},
		{"site-render", ComponentTypeSiteRender, "site-render"},
		{"pdf-render", ComponentTypePDFRender, "pdf-render"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.ct) != tt.expected {
				t.Errorf("got %s, want %s", tt.ct, tt.expected)
			}
		})
	}
}
