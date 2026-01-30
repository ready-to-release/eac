// Package build provides the build command implementation using cmdframework.
package build

import (
	"fmt"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

// CacheCheckResult represents the result of a cache verification for a component.
// Used by background cache detection to report results to the scheduler.
type CacheCheckResult struct {
	Moniker   string    // Display name: module:component:tool (matches TUI tabs)
	IsCached  bool      // True if cache hit confirmed
	CacheTime time.Time // When the cached artifact was originally built
	Module    string    // Module name
	Component string    // Component name
	Handler   string    // Tool/handler name
}

// CacheVerifierFunc returns a function compatible with orchestrator.CacheVerifier.
// This adapts VerifyComponentCache for use with StartBackgroundCacheDetection.
func CacheVerifierFunc() func(workspaceRoot string, spec workunit.UnitSpec, cachedModules map[string]bool, cacheTimes map[string]time.Time) (bool, time.Time) {
	return func(workspaceRoot string, spec workunit.UnitSpec, cachedModules map[string]bool, cacheTimes map[string]time.Time) (bool, time.Time) {
		result := VerifyComponentCache(workspaceRoot, spec, cachedModules, cacheTimes)
		return result.IsCached, result.CacheTime
	}
}

// VerifyComponentCache checks if a component's module is cached and verifies artifact integrity.
// This is the SAME logic as buildComponentWorker cache check, extracted for reuse.
// Operates at COMPONENT granularity to match TUI tabs.
//
// The verification process:
// 1. Check if module is in the pre-computed cachedModules set
// 2. Load the module manifest from build output (if exists)
// 3. Verify artifact integrity via hash comparison
//
// Returns a CacheCheckResult with IsCached=true if the cache is valid.
func VerifyComponentCache(
	workspaceRoot string,
	spec workunit.UnitSpec,
	cachedModules map[string]bool,
	cacheTimes map[string]time.Time,
) CacheCheckResult {
	module := spec.ID.Module

	// Build moniker to match scheduler's formatDisplayName()
	var moniker string
	if spec.ID.Tool != "" {
		moniker = fmt.Sprintf("%s:%s:%s", module, spec.ID.Component, spec.ID.Tool)
	} else {
		moniker = fmt.Sprintf("%s:%s", module, spec.ID.Component)
	}

	result := CacheCheckResult{
		Moniker:   moniker,
		Module:    module,
		Component: spec.ID.Component,
		Handler:   spec.ID.Tool,
		IsCached:  false,
	}

	// 1. Check if module is in cached set (pre-computed during init)
	if cachedModules == nil || !cachedModules[module] {
		return result
	}

	// 2. Load manifest from build output
	moduleBuildDir := paths.BuildOutputPath(workspaceRoot, module)
	manifest, err := internal.LoadModuleManifest(moduleBuildDir)
	if err != nil {
		// No manifest - trust source hash check (which already passed during init)
		// This is a valid cache hit
		result.IsCached = true
		result.CacheTime = cacheTimes[module]
		return result
	}

	// 3. Verify artifact integrity (hash check)
	if err := manifest.VerifyArtifactsIntegrity(moduleBuildDir); err != nil {
		// Artifacts changed - not actually cached
		// Don't modify cachedModules here - that's the worker's job
		return result
	}

	// 4. Confirmed cache hit
	result.IsCached = true
	result.CacheTime = cacheTimes[module]
	return result
}
