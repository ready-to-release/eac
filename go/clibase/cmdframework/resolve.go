package cmdframework

import (
	"fmt"
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
)

// phaseResolve handles the module resolution phase:
// - Load module contracts
// - Resolve monikers (all modules if none specified)
// - Apply skip filters (from CLI flags or config)
// - Calculate execution order (with or without dependencies)
// - Build module type lookup.
func phaseResolve(ctx *ExecutionContext) error {
	// Load module contracts
	discoveryStart := time.Now()
	moduleReport, err := reports.GetModuleContracts(ctx.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}
	ctx.ModuleReport = moduleReport
	ctx.ModuleRegistry = moduleReport.Registry
	ctx.initTimings.ModuleDiscovery = time.Since(discoveryStart)

	// Resolve monikers - if none specified, use all modules
	monikers := ctx.Config.Monikers
	if len(monikers) == 0 {
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
		ctx.Config.Monikers = monikers
		log.Debugf("No modules specified, using all %d modules", len(monikers))
	}

	// Apply skip filter (available to all commands via SkipMonikers)
	monikers = applySkipFilter(monikers, ctx)
	ctx.Config.Monikers = monikers

	// Validate all monikers exist
	for _, moniker := range monikers {
		if _, exists := moduleReport.Registry.Get(moniker); !exists {
			return fmt.Errorf("module not found: %s", moniker)
		}
	}

	// Expand scope to include transitive dependencies
	orderStart := time.Now()
	scopeMonikers, err := expandScope(monikers, ctx.ModuleRegistry, ctx.Config.IncludeDepm)
	if err != nil {
		return fmt.Errorf("failed to expand module scope: %w", err)
	}
	ctx.ScopeMonikers = scopeMonikers
	ctx.initTimings.ExecutionOrder = time.Since(orderStart)

	// Build module component types lookup for all modules in scope
	for _, moniker := range scopeMonikers {
		if module, exists := moduleReport.Registry.Get(moniker); exists {
			ctx.ComponentTypesDisplay[moniker] = module.GetComponentTypesDisplay()
		}
	}

	// Update orchestrator with module component types
	ctx.Orchestrator.SetComponentTypesDisplay(ctx.ComponentTypesDisplay)

	// Send planned work to TUI for early skeleton tabs (before tool resolution)
	if ctx.Config.UseTUI {
		sendPlannedWork(ctx)
	}

	log.Debugf("Resolved %d modules in scope: %v", len(scopeMonikers), scopeMonikers)

	return nil
}

// sendPlannedWork sends predicted work items to the TUI based on module discovery.
// This enables grey skeleton tabs to appear before tool resolution completes.
// Uses DisplayOrder filtered to scope for consistent tab ordering.
func sendPlannedWork(ctx *ExecutionContext) {
	if len(ctx.ScopeMonikers) == 0 || ctx.EACConfig == nil || ctx.Orchestrator == nil {
		return
	}

	// Use DisplayOrder filtered to scope for consistent TUI tab ordering.
	orderedScope := displayOrderedScope(ctx.ScopeMonikers, ctx.EACConfig.Repository.DisplayOrder)

	ctxName := string(ctx.Config.Type) // "build", "test", "lint", "scan"
	var items []display.PlannedWorkItem

	for _, moniker := range orderedScope {
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		for _, compName := range module.GetEnabledComponents() {
			toolIDs := getToolIDsForComponentKind(ctx.EACConfig.ComponentKinds, compName, ctx.Config.Type)
			if len(toolIDs) == 0 {
				continue
			}

			for _, toolID := range toolIDs {
				// Predicted Longname: action:module:component:tool
				id := fmt.Sprintf("%s:%s:%s:%s", ctxName, moniker, compName, toolID)
				displayName := moniker + ":" + compName

				items = append(items, display.PlannedWorkItem{
					ID:            id,
					DisplayName:   displayName,
					Weight:        1, // Approximate; enriched later
					Module:        moniker,
					Component:     compName,
					ComponentType: compName,
				})
			}
		}
	}

	if len(items) > 0 {
		ctx.Orchestrator.SendPlannedWork(items)
	}
}

// getToolIDsForComponentKind returns the tool IDs for a component kind and action type.
func getToolIDsForComponentKind(kinds *config.ComponentKindsConfig, compName string, cmdType core.ActionType) []string {
	if kinds == nil || kinds.Kinds == nil {
		return nil
	}
	ct, exists := kinds.Kinds[compName]
	if !exists || ct == nil {
		return nil
	}

	return ct.ToolIDsForAction(cmdType)
}

// GetRequestedMonikers returns just the originally requested monikers (not deps).
func (ctx *ExecutionContext) GetRequestedMonikers() []string {
	return ctx.Config.Monikers
}

// GetExecutionMonikers returns all monikers to be executed (including deps).
func (ctx *ExecutionContext) GetExecutionMonikers() []string {
	if len(ctx.ScopeMonikers) > 0 {
		return ctx.ScopeMonikers
	}
	return ctx.Config.Monikers
}

// GetAddedDependencies returns monikers added as dependencies (not originally requested).
func (ctx *ExecutionContext) GetAddedDependencies() []string {
	if len(ctx.ScopeMonikers) == 0 {
		return nil
	}

	requestedSet := make(map[string]bool)
	for _, m := range ctx.Config.Monikers {
		requestedSet[m] = true
	}

	var added []string
	for _, m := range ctx.ScopeMonikers {
		if !requestedSet[m] {
			added = append(added, m)
		}
	}
	return added
}


// applySkipFilter removes modules from the list based on skip configuration.
func applySkipFilter(monikers []string, ctx *ExecutionContext) []string {
	// Check if command provided a skip list
	skipList := ctx.Config.SkipMonikers
	if len(skipList) == 0 {
		return monikers
	}

	// Build skip set for O(1) lookup
	skipSet := make(map[string]bool)
	for _, skip := range skipList {
		skipSet[skip] = true
	}

	// Filter out skipped modules
	filtered := make([]string, 0, len(monikers))
	var skipped []string
	for _, moniker := range monikers {
		if skipSet[moniker] {
			skipped = append(skipped, moniker)
		} else {
			filtered = append(filtered, moniker)
		}
	}

	// Log what was filtered
	if len(skipped) > 0 {
		log.Infof("Skipping %d module(s): %v", len(skipped), skipped)
	}

	return filtered
}

// expandScope expands the requested monikers to include transitive dependencies.
// This does NOT perform topological sorting —
// scheduling order comes from UnitSpec.DependsOn constraints, and display order
// from DisplayOrder. The returned slice is sorted alphabetically for stable logging.
func expandScope(monikers []string, reg *modules.Registry, includeDeps bool) ([]string, error) {
	scope := make(map[string]bool, len(monikers))
	for _, m := range monikers {
		scope[m] = true
	}

	if includeDeps {
		for _, m := range monikers {
			if err := addTransitiveDeps(m, reg, scope); err != nil {
				return nil, err
			}
		}
	}

	result := make([]string, 0, len(scope))
	for m := range scope {
		result = append(result, m)
	}
	sort.Strings(result)
	return result, nil
}

// addTransitiveDeps recursively adds all dependencies of a module to the scope set.
func addTransitiveDeps(moniker string, reg *modules.Registry, scope map[string]bool) error {
	mod, exists := reg.Get(moniker)
	if !exists {
		return fmt.Errorf("module '%s' not found in registry", moniker)
	}

	for _, dep := range mod.GetDependencies() {
		if _, exists := reg.Get(dep); !exists {
			return fmt.Errorf("missing dependency contract: %s -> %s", moniker, dep)
		}
		if !scope[dep] {
			scope[dep] = true
			if err := addTransitiveDeps(dep, reg, scope); err != nil {
				return err
			}
		}
	}
	return nil
}

// displayOrderedScope returns scope monikers in DisplayOrder sequence.
// Modules not in DisplayOrder are appended alphabetically at the end.
func displayOrderedScope(scope []string, displayOrder *config.DisplayOrder) []string {
	if displayOrder == nil || len(displayOrder.Modules) == 0 {
		return scope
	}

	scopeSet := make(map[string]bool, len(scope))
	for _, m := range scope {
		scopeSet[m] = true
	}

	var ordered []string
	for _, m := range displayOrder.Modules {
		if scopeSet[m] {
			ordered = append(ordered, m)
			delete(scopeSet, m)
		}
	}

	// Append any remaining (shouldn't happen, but defensive)
	if len(scopeSet) > 0 {
		var remaining []string
		for m := range scopeSet {
			remaining = append(remaining, m)
		}
		sort.Strings(remaining)
		ordered = append(ordered, remaining...)
	}

	return ordered
}
