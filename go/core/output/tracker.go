package output

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// InMemoryTracker implements UoWTracker with immediate disk writes.
// It tracks start times in memory and persists manifests to disk on completion.
type InMemoryTracker struct {
	workspaceRoot string
	ctx           workunit.Context

	mu      sync.Mutex
	started map[string]time.Time // id.Longname() -> start time
}

// NewTracker creates a new UoW tracker for the given workspace and context.
func NewTracker(workspaceRoot string, ctx workunit.Context) *InMemoryTracker {
	return &InMemoryTracker{
		workspaceRoot: workspaceRoot,
		ctx:           ctx,
		started:       make(map[string]time.Time),
	}
}

// uowDir returns the output directory for a UoW.
// Format: out/{context}/{module}/{component}_{tool}/
func (t *InMemoryTracker) uowDir(id workunit.UnitID) string {
	dirName := id.Component + "_" + id.Tool
	return filepath.Join(t.workspaceRoot, "out", string(id.Context), id.Module, dirName)
}

// RecordStart marks a UoW as started and creates its output directory.
func (t *InMemoryTracker) RecordStart(id workunit.UnitID) error {
	t.mu.Lock()
	t.started[id.Longname()] = time.Now()
	t.mu.Unlock()

	// Create output directory
	dir := t.uowDir(id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	return nil
}

// RecordComplete records completion and persists the UoW manifest.
// If RecordStart was called, the duration is computed from the start time.
// Otherwise, the manifest's Duration field is preserved.
func (t *InMemoryTracker) RecordComplete(id workunit.UnitID, manifest *UoWManifest) error {
	t.mu.Lock()
	startTime, hasStart := t.started[id.Longname()]
	if hasStart {
		delete(t.started, id.Longname())
	}
	t.mu.Unlock()

	// Set duration from start time if available
	if hasStart && manifest.Duration == 0 {
		manifest.Duration = time.Since(startTime)
	}

	// Ensure UoW identity fields match the ID
	manifest.Context = id.Context
	manifest.Module = id.Module
	manifest.Component = id.Component
	manifest.Tool = id.Tool

	// Write manifest to disk
	return manifest.Save(t.workspaceRoot)
}

// RecordCacheHit validates and returns an existing UoW manifest.
// Returns an error if the manifest is missing, invalid, or artifacts are missing/corrupt.
func (t *InMemoryTracker) RecordCacheHit(id workunit.UnitID) (*UoWManifest, error) {
	// Compute manifest path
	dirName := id.Component + "_" + id.Tool
	manifestPath := filepath.Join(t.workspaceRoot, "out", string(id.Context), id.Module, dirName, "uow.manifest.json")

	// Load manifest
	manifest, err := Load(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("cache invalid - no manifest: %w", err)
	}

	// Validate artifacts
	if len(manifest.Artifacts) > 0 {
		uowDir := t.uowDir(id)
		result := ValidateArtifacts(uowDir, manifest.Artifacts)
		if !result.Valid {
			if len(result.MissingArtifacts) > 0 {
				return nil, fmt.Errorf("cache invalid - artifacts missing: %v", result.MissingArtifacts)
			}
			if len(result.CorruptArtifacts) > 0 {
				return nil, fmt.Errorf("cache invalid - artifacts corrupt: %v", result.CorruptArtifacts)
			}
			if result.Error != nil {
				return nil, fmt.Errorf("cache invalid - validation error: %w", result.Error)
			}
			return nil, fmt.Errorf("cache invalid - artifacts invalid")
		}
	}

	return manifest, nil
}
