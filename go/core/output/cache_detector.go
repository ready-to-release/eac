package output

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// UoWChangeResult represents batch change detection results at UoW granularity.
// This replaces module-level change detection with finer-grained UoW-level detection.
type UoWChangeResult struct {
	// Changed contains UoWs that need execution.
	Changed []workunit.UnitID
	// UpToDate contains UoWs that are cached and can be skipped.
	UpToDate []workunit.UnitID
	// ChangeReasons maps UoW longname to the reason it needs execution.
	ChangeReasons map[string]string
	// FreshRun indicates no prior manifests exist for any requested UoW.
	FreshRun bool
	// DetectionTime is how long the detection process took.
	DetectionTime time.Duration
}

// InputHashProvider computes the current input hash for a UoW.
// This is used to compare against the manifest's stored InputHash.
type InputHashProvider func(id workunit.UnitID) (string, error)

// DependencyResolver returns the list of module dependencies for a given module.
// Used to check cross-module build invalidation for modules that consume
// outputs from other modules (e.g., eac-ext consuming eac-cli's binary).
// Returns nil if the module has no dependencies or is not found.
type DependencyResolver func(module string) []string

// DetectUoWChanges checks which UoWs need execution based on their manifests.
// For each expected UoW, it:
// 1. Loads the existing manifest (if any)
// 2. Compares the manifest's InputHash with the current input hash
// 3. Validates that artifacts are still intact
// 4. Returns whether the UoW is up-to-date or needs re-execution
//
// A UoW is marked as changed if:
// - No manifest exists (fresh/new UoW)
// - Previous execution failed (ExitCode != 0)
// - Input hash differs (source changed)
// - Artifacts are missing or corrupted
// - Hash provider returns an error
func (r *DiskOutputReader) DetectUoWChanges(
	ctx core.ActionType,
	expectedUoWs []workunit.UnitID,
	getInputHash InputHashProvider,
	depResolver DependencyResolver,
) (*UoWChangeResult, error) {
	start := time.Now()
	result := &UoWChangeResult{
		Changed:       []workunit.UnitID{},
		UpToDate:      []workunit.UnitID{},
		ChangeReasons: make(map[string]string),
	}

	if len(expectedUoWs) == 0 {
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check if any manifests exist (to detect fresh run)
	hasAnyManifest := false
	for _, id := range expectedUoWs {
		if _, err := r.GetUoW(id); err == nil {
			hasAnyManifest = true
			break
		}
	}

	if !hasAnyManifest {
		// Fresh run - all UoWs need execution
		result.FreshRun = true
		for _, id := range expectedUoWs {
			result.Changed = append(result.Changed, id)
			result.ChangeReasons[id.Longname()] = "fresh run (no prior manifest)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check each UoW individually
	for _, id := range expectedUoWs {
		changed, reason := r.checkUoWChanged(ctx, id, getInputHash, depResolver)
		if changed {
			result.Changed = append(result.Changed, id)
			result.ChangeReasons[id.Longname()] = reason
		} else {
			result.UpToDate = append(result.UpToDate, id)
		}
	}

	result.DetectionTime = time.Since(start)
	return result, nil
}

// checkUoWChanged determines if a single UoW needs execution.
// Returns (true, reason) if changed, (false, "") if up-to-date.
func (r *DiskOutputReader) checkUoWChanged(
	ctx core.ActionType,
	id workunit.UnitID,
	getInputHash InputHashProvider,
	depResolver DependencyResolver,
) (bool, string) {
	// Load existing manifest
	manifest, err := r.GetUoW(id)
	if err != nil {
		return true, "no prior manifest"
	}

	// Check if previous execution failed
	if manifest.ExitCode != 0 {
		return true, "previous failure"
	}

	// NoOp UoWs are always up-to-date if manifest exists and succeeded.
	// These are placeholder UoWs for modules with no buildable components,
	// no tests, no lintable code, or non-scannable components.
	if manifest.NoOp {
		return false, "" // Up-to-date
	}

	// Handle nil hash provider
	if getInputHash == nil {
		return true, "no hash provider"
	}

	// Compute current input hash
	currentHash, err := getInputHash(id)
	if err != nil {
		return true, "hash error: " + err.Error()
	}

	// Compare hashes
	if manifest.InputHash != currentHash {
		return true, "source changed"
	}

	// Validate artifacts if any exist
	if len(manifest.Artifacts) > 0 {
		validationResult := r.ValidateUoW(id)
		if !validationResult.Valid {
			return true, "artifacts invalid"
		}
	}

	// For test/lint/scan contexts, check if this module's own build was re-executed.
	if ctx == core.ActionTest || ctx == core.ActionLint || ctx == core.ActionScan {
		if r.hasNewerBuildManifest(id.Module, manifest.ExecutedAt) {
			return true, "build invalidated"
		}
	}

	// For build context, check if any dependency module was rebuilt since this
	// manifest was created. This handles cross-module output consumption
	// (e.g., eac-ext consuming a pre-built binary from eac-cli).
	if ctx == core.ActionBuild && depResolver != nil {
		for _, depModule := range depResolver(id.Module) {
			if r.hasNewerBuildManifest(depModule, manifest.ExecutedAt) {
				return true, "dependency build invalidated (" + depModule + ")"
			}
		}
	}

	// All checks passed - UoW is up-to-date
	return false, ""
}

// hasNewerBuildManifest checks if any build manifest for the given module
// was executed after the specified time.
func (r *DiskOutputReader) hasNewerBuildManifest(module string, after time.Time) bool {
	manifests, err := r.ListUoWs(core.ActionBuild, module)
	if err != nil {
		return false
	}
	for _, m := range manifests {
		if m.ExecutedAt.After(after) {
			return true
		}
	}
	return false
}

// IsModuleChanged returns true if any UoW in the module needs execution.
// This aggregates UoW-level change detection to module-level status.
//
// A module is considered "changed" if:
// - No manifests exist for the module (fresh run)
// - Any UoW within the module has a changed input hash
// - Any UoW within the module previously failed
// - Any UoW within the module has invalid artifacts
//
// Returns:
// - changed: true if the module needs (re)execution
// - reason: description of why the module changed (empty if not changed)
// - err: any error encountered during detection
func (r *DiskOutputReader) IsModuleChanged(
	ctx core.ActionType,
	module string,
	getInputHash InputHashProvider,
) (changed bool, reason string, err error) {
	// List all existing manifests for this module
	manifests, err := r.ListUoWs(ctx, module)
	if err != nil {
		return true, "error listing manifests: " + err.Error(), nil
	}

	// No manifests = module is new/changed
	if len(manifests) == 0 {
		return true, "no prior manifests", nil
	}

	// Check each manifest
	for _, m := range manifests {
		id := workunit.UnitID{
			Action:    ctx,
			Module:    module,
			Component: m.Component,
			Tool:      m.Tool,
		}

		// Check if previous execution failed
		if m.ExitCode != 0 {
			return true, "previous failure in " + m.Component, nil
		}

		// NoOp UoWs are always up-to-date if manifest exists and succeeded.
		// Skip hash comparison for NoOp manifests.
		if m.NoOp {
			continue // Up-to-date, check next manifest
		}

		// Handle nil hash provider
		if getInputHash == nil {
			return true, "no hash provider", nil
		}

		// Compute current hash
		currentHash, err := getInputHash(id)
		if err != nil {
			return true, "hash error for " + m.Component + ": " + err.Error(), nil
		}

		// Compare hashes
		if m.InputHash != currentHash {
			return true, "source changed in " + m.Component, nil
		}

		// Note: We don't validate artifacts here for performance.
		// Full artifact validation can be done via ValidateUoW if needed.
	}

	// All UoWs are up-to-date
	return false, "", nil
}
