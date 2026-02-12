package scan

import (
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/resolver"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ResolveScanUnitSpecs converts modules to scan component work items.
// Each module's scannable component + scanner combination becomes a work item.
// This allows parallel execution of different scanners (trivy-vuln, semgrep, etc.)
// within the same module.
// Uses ComponentResolver for consistent component-to-tool mapping.
// Returns nil if no scannable components are found.
func ResolveScanUnitSpecs(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	cfg := config.Global()
	if cfg == nil {
		return nil
	}

	// Check if module registry is available
	if ctx.ModuleRegistry == nil {
		return nil
	}

	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return nil
	}

	// Get cached modules map from scan context (set during incremental detection)
	var cachedModules map[string]bool
	if sctx, ok := ctx.Config.ScanCmdContext.(*scanContext); ok && sctx != nil {
		cachedModules = sctx.cachedModules
	}

	// Create component resolver for unified resolution
	compResolver := resolver.NewComponentResolver()

	var allWork []workunit.UnitSpec
	globalIndex := 0

	for _, moniker := range monikers {
		// Get module contract
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Use ComponentResolver to get scan specs for this module
		// Pass nil for scanCategories to use defaults from blueprints.yml component-kinds
		specs := compResolver.ResolveForScan(module, nil, cachedModules)

		// Add specs with correct index (weight is already set by ComponentResolver)
		for i := range specs {
			specs[i].Index = globalIndex
			globalIndex++
		}

		allWork = append(allWork, specs...)
	}

	return allWork
}

// CountScanComponents returns the total number of scan component work items.
func CountScanComponents(units []workunit.UnitSpec) int {
	return len(units)
}

// getScanUoWCount returns the total number of scannable UoWs (units of work).
func getScanUoWCount(ctx *cmdframework.ExecutionContext) int {
	units := ResolveScanUnitSpecs(ctx)
	return CountScanComponents(units)
}
