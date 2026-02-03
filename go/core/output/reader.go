package output

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

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

// uowManifestPath returns the expected path for a UoW manifest.
func (r *DiskOutputReader) uowManifestPath(ctx workunit.Context, module, component, tool string) string {
	dirName := component + "_" + tool
	return filepath.Join(r.workspaceRoot, "out", string(ctx), module, dirName, "uow.manifest.json")
}

// uowDir returns the directory for a UoW.
func (r *DiskOutputReader) uowDir(ctx workunit.Context, module, component, tool string) string {
	dirName := component + "_" + tool
	return filepath.Join(r.workspaceRoot, "out", string(ctx), module, dirName)
}

// GetUoW loads a single UoW manifest from disk.
func (r *DiskOutputReader) GetUoW(ctx workunit.Context, module, component, tool string) (*UoWManifest, error) {
	path := r.uowManifestPath(ctx, module, component, tool)
	return Load(path)
}

// GetComponent computes a component view by aggregating its UoWs.
func (r *DiskOutputReader) GetComponent(ctx workunit.Context, module, component string) (*ComponentView, error) {
	view := &ComponentView{
		Module:    module,
		Component: component,
		Status:    StatusPending,
		UoWs:      []UoWManifest{},
	}

	// Find all UoW directories for this component
	// Pattern: out/{ctx}/{module}/{component}_*/uow.manifest.json
	moduleDir := filepath.Join(r.workspaceRoot, "out", string(ctx), module)
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return view, nil // No UoWs for this module
		}
		return nil, err
	}

	prefix := component + "_"
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}

		manifestPath := filepath.Join(moduleDir, entry.Name(), "uow.manifest.json")
		manifest, err := Load(manifestPath)
		if err != nil {
			continue // Skip invalid manifests
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
func (r *DiskOutputReader) GetModule(ctx workunit.Context, module string) (*ModuleView, error) {
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
func (r *DiskOutputReader) ListUoWs(ctx workunit.Context, module string) ([]*UoWManifest, error) {
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

// ValidateUoW checks if a UoW's output is valid.
func (r *DiskOutputReader) ValidateUoW(ctx workunit.Context, module, component, tool string) ValidationResult {
	result := ValidationResult{
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{},
	}

	// Check manifest exists and is valid
	manifestPath := r.uowManifestPath(ctx, module, component, tool)
	manifest, err := Load(manifestPath)
	if err != nil {
		result.Valid = false
		result.ManifestExists = false
		result.ManifestValid = false
		result.ArtifactsValid = false
		result.Error = err
		return result
	}

	result.ManifestExists = true
	result.ManifestValid = true

	// Validate artifacts
	if len(manifest.Artifacts) == 0 {
		result.Valid = true
		result.ArtifactsValid = true
		return result
	}

	uowDir := r.uowDir(ctx, module, component, tool)
	artifactResult := ValidateArtifacts(uowDir, manifest.Artifacts)

	result.Valid = artifactResult.Valid
	result.ArtifactsValid = artifactResult.ArtifactsValid
	result.MissingArtifacts = artifactResult.MissingArtifacts
	result.CorruptArtifacts = artifactResult.CorruptArtifacts
	result.Error = artifactResult.Error

	return result
}

// ValidateModule checks if all expected UoWs for a module are valid.
func (r *DiskOutputReader) ValidateModule(ctx workunit.Context, module string, expectedUoWs []workunit.UnitID) ValidationResult {
	result := ValidationResult{
		Valid:            true,
		ManifestExists:   true,
		ManifestValid:    true,
		ArtifactsValid:   true,
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{},
	}

	// If no expected UoWs, validation passes
	if len(expectedUoWs) == 0 {
		return result
	}

	// Check each expected UoW
	for _, id := range expectedUoWs {
		uowResult := r.ValidateUoW(ctx, id.Module, id.Component, id.Tool)

		if !uowResult.Valid {
			result.Valid = false

			if !uowResult.ManifestExists {
				result.ManifestExists = false
				result.MissingArtifacts = append(result.MissingArtifacts, id.Component+"_"+id.Tool+"/uow.manifest.json")
			}
			if !uowResult.ManifestValid {
				result.ManifestValid = false
			}
			if !uowResult.ArtifactsValid {
				result.ArtifactsValid = false
				result.MissingArtifacts = append(result.MissingArtifacts, uowResult.MissingArtifacts...)
				result.CorruptArtifacts = append(result.CorruptArtifacts, uowResult.CorruptArtifacts...)
			}
			if uowResult.Error != nil && result.Error == nil {
				result.Error = uowResult.Error
			}
		}
	}

	return result
}

// HasManifests returns true if any UoW manifests exist for the module.
func (r *DiskOutputReader) HasManifests(ctx workunit.Context, module string) bool {
	manifests, err := r.ListUoWs(ctx, module)
	return err == nil && len(manifests) > 0
}

// VerifyModuleIntegrity checks all UoW artifacts for a module.
// Returns nil if all artifacts pass hash verification.
// Returns error describing first failed artifact otherwise.
// This replaces legacy ModuleManifest.VerifyArtifactsIntegrity().
func (r *DiskOutputReader) VerifyModuleIntegrity(ctx workunit.Context, module string) error {
	manifests, err := r.ListUoWs(ctx, module)
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		return nil // No manifests = nothing to verify
	}

	for _, manifest := range manifests {
		uowDir := r.uowDir(ctx, module, manifest.Component, manifest.Tool)
		result := ValidateArtifacts(uowDir, manifest.Artifacts)
		if !result.Valid {
			uowName := manifest.Component + "_" + manifest.Tool
			if len(result.MissingArtifacts) > 0 {
				return &IntegrityError{
					UoW:     uowName,
					Message: "missing artifacts: " + strings.Join(result.MissingArtifacts, ", "),
				}
			}
			if len(result.CorruptArtifacts) > 0 {
				return &IntegrityError{
					UoW:     uowName,
					Message: "hash mismatch: " + strings.Join(result.CorruptArtifacts, ", "),
				}
			}
			if result.Error != nil {
				return &IntegrityError{
					UoW:     uowName,
					Message: result.Error.Error(),
				}
			}
		}
	}

	return nil
}

// GetBuildID returns a stable identifier linking tests to builds.
// This is computed from the concatenation of all UoW input hashes for the module.
// Returns empty string if no manifests exist.
func (r *DiskOutputReader) GetBuildID(ctx workunit.Context, module string) string {
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

// IntegrityError represents an artifact integrity validation failure.
type IntegrityError struct {
	UoW     string
	Message string
}

func (e *IntegrityError) Error() string {
	return "UoW " + e.UoW + ": " + e.Message
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
