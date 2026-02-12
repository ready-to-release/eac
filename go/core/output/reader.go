package output

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// DiskOutputReader implements OutputReader by scanning disk.
// It computes aggregation views on-the-fly from UoW manifests.
type DiskOutputReader struct {
	workspaceRoot string
}

// NewReader creates a new OutputReader for the given workspace.
func NewReader(workspaceRoot string) *DiskOutputReader {
	return &DiskOutputReader{
		workspaceRoot: workspaceRoot,
	}
}

// uowManifestPath returns the expected path for a UoW manifest using UnitID.
func (r *DiskOutputReader) uowManifestPath(id workunit.UnitID) string {
	return filepath.Join(r.workspaceRoot, "out", string(id.Action), id.Module, id.DirName(), "uow.manifest.json")
}

// uowDir returns the directory for a UoW using UnitID.
func (r *DiskOutputReader) uowDir(id workunit.UnitID) string {
	return filepath.Join(r.workspaceRoot, "out", string(id.Action), id.Module, id.DirName())
}

// GetUoW loads a single UoW manifest from disk using a UnitID.
func (r *DiskOutputReader) GetUoW(id workunit.UnitID) (*UoWManifest, error) {
	path := r.uowManifestPath(id)
	return Load(path)
}

// GetComponent computes a component view by aggregating its UoWs.
func (r *DiskOutputReader) GetComponent(ctx core.ActionType, module, component string) (*ComponentView, error) {
	view := &ComponentView{
		Module:    module,
		Component: component,
		Status:    StatusPending,
		UoWs:      []UoWManifest{},
	}

	// Find all UoW directories for this component
	// Pattern: out/{ctx}/{module}/{component}[-*]/uow.manifest.json
	// Directory names are: component (no extras) or component-extra1-extra2 (with extras)
	moduleDir := filepath.Join(r.workspaceRoot, "out", string(ctx), module)
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return view, nil // No UoWs for this module
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Match directories that are exactly component or start with component-
		dirName := entry.Name()
		if dirName != component && !strings.HasPrefix(dirName, component+"-") {
			continue
		}

		manifestPath := filepath.Join(moduleDir, dirName, "uow.manifest.json")
		manifest, err := Load(manifestPath)
		if err != nil {
			continue // Skip invalid manifests
		}

		// Double-check the manifest's component matches
		if manifest.Component != component {
			continue
		}

		view.UoWs = append(view.UoWs, *manifest)
		for _, artifact := range manifest.Artifacts {
			view.TotalSize += artifact.Size
		}
	}

	// Compute status from UoW exit codes
	view.Status = computeStatusFromUoWs(view.UoWs)

	return view, nil
}

// GetModule computes a module view by aggregating all components.
func (r *DiskOutputReader) GetModule(ctx core.ActionType, module string) (*ModuleView, error) {
	view := &ModuleView{
		Module:     module,
		Status:     StatusPending,
		Components: []ComponentView{},
	}

	// Find all UoW directories in this module
	// Pattern: out/{ctx}/{module}/*/uow.manifest.json
	moduleDir := filepath.Join(r.workspaceRoot, "out", string(ctx), module)
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return view, nil // No UoWs for this module
		}
		return nil, err
	}

	// Group UoWs by component
	components := make(map[string]*ComponentView)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(moduleDir, entry.Name(), "uow.manifest.json")
		manifest, err := Load(manifestPath)
		if err != nil {
			continue // Skip invalid manifests
		}

		compView, exists := components[manifest.Component]
		if !exists {
			compView = &ComponentView{
				Module:    module,
				Component: manifest.Component,
				UoWs:      []UoWManifest{},
			}
			components[manifest.Component] = compView
		}

		compView.UoWs = append(compView.UoWs, *manifest)
		for _, artifact := range manifest.Artifacts {
			compView.TotalSize += artifact.Size
		}
	}

	// Build component list and compute statuses
	for _, compView := range components {
		compView.Status = computeStatusFromUoWs(compView.UoWs)
		view.Components = append(view.Components, *compView)
		view.TotalSize += compView.TotalSize
	}

	// Compute module status from components
	view.Status = computeStatusFromComponents(view.Components)

	return view, nil
}

// ListUoWs returns all UoW manifests for a module.
func (r *DiskOutputReader) ListUoWs(ctx core.ActionType, module string) ([]*UoWManifest, error) {
	var manifests []*UoWManifest

	// Find all UoW directories in this module
	moduleDir := filepath.Join(r.workspaceRoot, "out", string(ctx), module)
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return manifests, nil // No UoWs for this module
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(moduleDir, entry.Name(), "uow.manifest.json")
		manifest, err := Load(manifestPath)
		if err != nil {
			continue // Skip invalid manifests
		}

		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

// HasManifests returns true if any UoW manifests exist for the module.
func (r *DiskOutputReader) HasManifests(ctx core.ActionType, module string) bool {
	manifests, err := r.ListUoWs(ctx, module)
	return err == nil && len(manifests) > 0
}

// GetBuildID returns a stable identifier linking tests to builds.
// This is computed from the concatenation of all UoW input hashes for the module.
// Returns empty string if no manifests exist.
func (r *DiskOutputReader) GetBuildID(ctx core.ActionType, module string) string {
	manifests, err := r.ListUoWs(ctx, module)
	if err != nil || len(manifests) == 0 {
		return ""
	}

	// Combine all input hashes to create a stable build identifier
	var hashes []string
	for _, m := range manifests {
		if m.InputHash != "" {
			// Use prefix of hash for brevity
			hash := m.InputHash
			if len(hash) > 16 {
				hash = hash[:16]
			}
			hashes = append(hashes, hash)
		}
	}

	if len(hashes) == 0 {
		return ""
	}

	// Sort for determinism
	sort.Strings(hashes)
	return strings.Join(hashes, "-")
}

// computeStatusFromUoWs computes the aggregate status from a list of UoWs.
func computeStatusFromUoWs(uows []UoWManifest) Status {
	if len(uows) == 0 {
		return StatusPending
	}

	for _, uow := range uows {
		if uow.ExitCode != 0 {
			return StatusFailed
		}
	}

	return StatusCompleted
}

// computeStatusFromComponents computes the aggregate status from component statuses.
func computeStatusFromComponents(components []ComponentView) Status {
	if len(components) == 0 {
		return StatusPending
	}

	hasFailed := false
	hasInProgress := false
	hasPending := false
	allCached := true

	for _, comp := range components {
		switch comp.Status {
		case StatusFailed:
			hasFailed = true
			allCached = false
		case StatusInProgress:
			hasInProgress = true
			allCached = false
		case StatusPending:
			hasPending = true
			allCached = false
		case StatusCompleted:
			allCached = false
		case StatusCached:
			// Keep allCached as is
		}
	}

	// Priority: Failed > InProgress > Pending > Cached > Completed
	if hasFailed {
		return StatusFailed
	}
	if hasInProgress {
		return StatusInProgress
	}
	if hasPending {
		return StatusPending
	}
	if allCached {
		return StatusCached
	}
	return StatusCompleted
}
