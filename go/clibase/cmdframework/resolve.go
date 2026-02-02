package cmdframework

import (
	"fmt"
	"time"

	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
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

	// Apply skip filter (available to all commands via Extra["skipMonikers"])
	monikers = applySkipFilter(monikers, ctx)
	ctx.Config.Monikers = monikers

	// Validate all monikers exist
	for _, moniker := range monikers {
		if _, exists := moduleReport.Registry.Get(moniker); !exists {
			return fmt.Errorf("module not found: %s", moniker)
		}
	}

	// Calculate execution order
	orderStart := time.Now()
	executionPlan, err := repository.CalculateExecutionOrder(
		monikers,
		ctx.WorkspaceRoot,
		ctx.Config.IncludeDepm, // Include module dependencies
	)
	if err != nil {
		return fmt.Errorf("failed to calculate execution order: %w", err)
	}
	ctx.ExecutionPlan = executionPlan
	ctx.initTimings.ExecutionOrder = time.Since(orderStart)

	// Build module component types lookup for all modules in execution plan
	for _, moniker := range executionPlan.ExecutionOrder {
		if module, exists := moduleReport.Registry.Get(moniker); exists {
			ctx.ModuleTypes[moniker] = module.GetComponentTypesDisplay()
		}
	}

	// Update orchestrator with module component types
	ctx.Orchestrator.SetModuleTypes(ctx.ModuleTypes)

	log.Debugf("Resolved %d modules, execution order: %v",
		len(executionPlan.ExecutionOrder), executionPlan.ExecutionOrder)

	return nil
}

// GetRequestedMonikers returns just the originally requested monikers (not deps).
func (ctx *ExecutionContext) GetRequestedMonikers() []string {
	return ctx.Config.Monikers
}

// GetExecutionMonikers returns all monikers to be executed (including deps).
func (ctx *ExecutionContext) GetExecutionMonikers() []string {
	if ctx.ExecutionPlan != nil {
		return ctx.ExecutionPlan.ExecutionOrder
	}
	return ctx.Config.Monikers
}

// GetAddedDependencies returns monikers added as dependencies (not originally requested).
func (ctx *ExecutionContext) GetAddedDependencies() []string {
	if ctx.ExecutionPlan == nil {
		return nil
	}

	requestedSet := make(map[string]bool)
	for _, m := range ctx.Config.Monikers {
		requestedSet[m] = true
	}

	var added []string
	for _, m := range ctx.ExecutionPlan.ExecutionOrder {
		if !requestedSet[m] {
			added = append(added, m)
		}
	}
	return added
}

// GetLayers returns the dependency layers for layered execution.
func (ctx *ExecutionContext) GetLayers() [][]string {
	if ctx.ExecutionPlan != nil {
		return ctx.ExecutionPlan.Layers
	}
	// Single layer with all monikers for non-layered execution
	return [][]string{ctx.Config.Monikers}
}

// applySkipFilter removes modules from the list based on skip configuration.
func applySkipFilter(monikers []string, ctx *ExecutionContext) []string {
	// Check if command provided a skip list
	if ctx.Config.Extra == nil {
		return monikers
	}

	skipList, ok := ctx.Config.Extra["skipMonikers"].([]string)
	if !ok || len(skipList) == 0 {
		return monikers
	}

	// Build skip set for O(1) lookup
	skipSet := make(map[string]bool)
	for _, skip := range skipList {
		skipSet[skip] = true
	}

	// If no skip filters configured, return original list
	if len(skipSet) == 0 {
		return monikers
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
