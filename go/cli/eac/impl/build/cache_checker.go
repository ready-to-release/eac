// Package build provides the build command implementation using cmdframework.
package build

import (
	"context"
	"fmt"
	"time"

	"github.com/ready-to-release/eac/go/core/execution"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
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

// BuildCacheVerifier implements execution.CacheVerifier for build commands.
// It verifies cache status by checking:
// 1. Pre-computed cachedModules set (from incremental detection)
// 2. Module manifest existence
// 3. Artifact integrity via hash comparison
type BuildCacheVerifier struct {
	workspaceRoot string
	cachedModules map[string]bool
	cacheTimes    map[string]time.Time
}

// NewBuildCacheVerifier creates a new BuildCacheVerifier.
func NewBuildCacheVerifier(workspaceRoot string, cachedModules map[string]bool, cacheTimes map[string]time.Time) *BuildCacheVerifier {
	return &BuildCacheVerifier{
		workspaceRoot: workspaceRoot,
		cachedModules: cachedModules,
		cacheTimes:    cacheTimes,
	}
}

// Verify implements execution.CacheVerifier.
// It checks if a work unit's output is cached and valid.
func (v *BuildCacheVerifier) Verify(ctx context.Context, unit workunit.UnitSpec) (execution.CacheResult, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return execution.CacheResult{}, ctx.Err()
	default:
	}

	result := VerifyComponentCache(v.workspaceRoot, unit, v.cachedModules, v.cacheTimes)
	return execution.CacheResult{
		Cached:    result.IsCached,
		CacheTime: result.CacheTime,
	}, nil
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

	// 2. Check UoW manifests and verify artifact integrity
	reader := coreoutput.NewReader(workspaceRoot)
	if !reader.HasManifests(workunit.ContextBuild, module) {
		// No manifests - trust source hash check (which already passed during init)
		// This is a valid cache hit
		result.IsCached = true
		result.CacheTime = cacheTimes[module]
		return result
	}

	// 3. Verify artifact integrity (hash check)
	if err := reader.VerifyModuleIntegrity(workunit.ContextBuild, module); err != nil {
		// Artifacts changed - not actually cached
		// Don't modify cachedModules here - that's the worker's job
		return result
	}

	// 4. Confirmed cache hit
	result.IsCached = true
	result.CacheTime = cacheTimes[module]
	return result
}
