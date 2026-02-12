package workunit

import (
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/cache"
)

// ModuleChangeResult represents change detection results at module granularity.
// This bridges module-level caching (current approach) with unit-level storage.
type ModuleChangeResult struct {
	// Modules that need execution
	ChangedModules []string
	// Modules that are up-to-date
	UpToDateModules []string
	// Reason for each changed module
	ChangeReasons map[string]string
	// Whether this is a fresh run
	FreshRun bool
	// Detection time
	DetectionTime time.Duration
}

// ModuleHashProvider computes the source hash for a module.
type ModuleHashProvider func(module string) (string, error)

// DetectModuleChanges performs change detection at module granularity.
// This is a convenience method that wraps per-unit state checks into module-level results.
// For each module, it creates a representative unit ID and checks its state.
// If cacheConfig is provided and state cache is skipped, all modules are marked as changed.
func (m *StateManager) DetectModuleChanges(ctx core.ActionType, modules []string, rule InvalidationRule, hashProvider ModuleHashProvider, cacheConfig *cache.Config) (*ModuleChangeResult, error) {
	start := time.Now()
	result := &ModuleChangeResult{
		ChangeReasons: make(map[string]string),
	}

	if len(modules) == 0 {
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// If state cache is skipped, all modules need execution
	if cacheConfig != nil && cacheConfig.ShouldSkipState() {
		result.ChangedModules = append(result.ChangedModules, modules...)
		for _, module := range modules {
			result.ChangeReasons[module] = "cache skipped (--skip-cache=state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check if any prior state exists
	hasAnyState := false
	for _, module := range modules {
		// Use a representative unit ID for the module (context:module:_module:_)
		unitID := UnitID{
			Action:    ctx,
			Module:    module,
			ComponentType: "_module", // Representative component for module-level state
			ComponentName: "_module",
			Tool:      "_",
		}
		if m.Exists(unitID) {
			hasAnyState = true
			break
		}
	}

	if !hasAnyState {
		// Fresh run - all modules need execution
		result.FreshRun = true
		result.ChangedModules = append(result.ChangedModules, modules...)
		for _, module := range modules {
			result.ChangeReasons[module] = "fresh run (no prior state)"
		}
		result.DetectionTime = time.Since(start)
		return result, nil
	}

	// Check each module
	for _, module := range modules {
		unitID := UnitID{
			Action:    ctx,
			Module:    module,
			ComponentType: "_module",
			ComponentName: "_module",
			Tool:          "_",
		}
		spec := UnitSpec{ID: unitID}

		currentHash := ""
		if hashProvider != nil {
			var err error
			currentHash, err = hashProvider(module)
			if err != nil {
				result.ChangedModules = append(result.ChangedModules, module)
				result.ChangeReasons[module] = "hash error: " + err.Error()
				continue
			}
		}

		needsExec, reason := m.NeedsExecution(spec, rule, currentHash, cacheConfig)
		if needsExec {
			result.ChangedModules = append(result.ChangedModules, module)
			result.ChangeReasons[module] = reason
		} else {
			result.UpToDateModules = append(result.UpToDateModules, module)
		}
	}

	// Sort for deterministic output
	sort.Strings(result.ChangedModules)
	sort.Strings(result.UpToDateModules)

	result.DetectionTime = time.Since(start)
	return result, nil
}

// SaveModuleResult saves state for a module-level execution.
// Uses a representative unit ID for the module.
func (m *StateManager) SaveModuleResult(ctx core.ActionType, module string, passed bool, sourceHash string) error {
	unitID := UnitID{
		Action:        ctx,
		Module:        module,
		ComponentType: "_module",
		ComponentName: "_module",
		Tool:          "_",
	}

	state := &UnitState{
		ID:         unitID,
		SourceHash: sourceHash,
		Passed:     passed,
		ExecutedAt: time.Now(),
	}

	return m.Save(state)
}
