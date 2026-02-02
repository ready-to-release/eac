package scan

import (
	"math"

	"github.com/ready-to-release/eac/go/cli/eac/impl/scan/internal"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/resolver"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// FlattenModulesToScanUnits converts modules to scan component work items.
// Each module's scannable component + scanner combination becomes a work item.
// This allows parallel execution of different scanners (trivy-vuln, semgrep, etc.)
// within the same module.
// Uses ComponentResolver for consistent component-to-tool mapping.
// Returns nil if no scannable components are found.
func FlattenModulesToScanUnits(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
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
	if sctx, ok := ctx.Config.Extra["scanContext"].(*scanContext); ok && sctx != nil {
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
		// Pass nil for scanCategories to use defaults from component-types.yml
		specs := compResolver.ResolveForScan(module, nil, cachedModules)

		// Add specs with correct index and weight
		for i := range specs {
			// Apply scanner-specific weight from getUnitWeight
			scannerType := internal.ScannerType(specs[i].ID.Tool)
			specs[i].Weight = getUnitWeight(moniker, specs[i].ID.Component, scannerType)
			specs[i].Index = globalIndex
			globalIndex++
		}

		allWork = append(allWork, specs...)
	}

	if len(allWork) == 0 {
		return nil
	}

	// Return as single layer (scans run in parallel)
	return [][]workunit.UnitSpec{allWork}
}

// getScanWeightForScanner returns the weight for a specific scanner type.
// Different scanners have different resource requirements.
func getScanWeightForScanner(scannerType internal.ScannerType) int {
	// Default weights based on scanner characteristics
	// Higher weights = more resource-intensive
	switch scannerType {
	case internal.ScannerSAST:
		// SAST (Semgrep) can be CPU-intensive on large codebases
		return 2
	case internal.ScannerVuln:
		// Vulnerability scanning involves network requests for DB updates
		return 2
	case internal.ScannerSBOM:
		// SBOM generation is relatively lightweight
		return 1
	case internal.ScannerSecrets:
		// Secret scanning is relatively fast
		return 1
	case internal.ScannerIaC:
		// IaC scanning is moderate
		return 1
	case internal.ScannerCompliance:
		// Compliance scanning is moderate
		return 1
	case internal.ScannerDAST:
		// DAST requires running services, high resource
		return 3
	default:
		return 1
	}
}

// getUnitWeight returns the scheduling weight for a component.
// Weight = base scanner weight × component amp (from config).
func getUnitWeight(moniker, componentName string, scannerType internal.ScannerType) int {
	baseWeight := getScanWeightForScanner(scannerType)

	// Get amp from config (the source of truth)
	amp := 1.0
	cfg := config.Global()
	if cfg != nil && cfg.Repository != nil {
		if module, ok := cfg.Repository.GetModule(moniker); ok && module != nil {
			amp = module.GetComponentAmp(componentName, "scan")
		}
	}

	// Apply amp to weight (ceil to ensure at least 1)
	weight := int(math.Ceil(float64(baseWeight) * amp))
	if weight < 1 {
		weight = 1
	}

	return weight
}

// CountScanComponents returns the total number of scan component work items.
func CountScanComponents(layers [][]workunit.UnitSpec) int {
	count := 0
	for _, layer := range layers {
		count += len(layer)
	}
	return count
}

// getScanUoWCount returns the total number of scannable UoWs (units of work).
func getScanUoWCount(ctx *cmdframework.ExecutionContext) int {
	layers := FlattenModulesToScanUnits(ctx)
	return CountScanComponents(layers)
}
