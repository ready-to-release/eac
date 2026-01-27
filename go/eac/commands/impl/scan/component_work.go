package scan

import (
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

// FlattenModulesToScanComponentWork converts modules to scan component work items.
// Each module's scannable components (go, typescript, dockerfile, etc.) become work items.
// Returns nil if no scannable components are found.
func FlattenModulesToScanComponentWork(ctx *cmdframework.ExecutionContext) [][]orchestrator.ComponentWork {
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

	var allWork []orchestrator.ComponentWork
	globalIndex := 0

	for _, moniker := range monikers {
		// Get module contract
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Get enabled components for this module
		enabledComponents := module.GetEnabledComponents()

		// Filter to only scannable components
		var scannableComponents []string
		for _, componentName := range enabledComponents {
			compTypeName := module.Components.GetComponentType(componentName)
			compType := cfg.ComponentTypes.Get(compTypeName)
			if compType != nil && compType.IsScannable() {
				scannableComponents = append(scannableComponents, componentName)
			}
		}

		if len(scannableComponents) == 0 {
			// Module has no scannable components - skip it entirely
			// No placeholder needed since scan only operates on scannable components
			continue
		}

		// Create work item for each scannable component
		for _, componentName := range scannableComponents {
			// Get component type (may differ from name for named components)
			compTypeName := module.Components.GetComponentType(componentName)

			// Determine weight based on component type
			// Heavier scans (container, SAST on large codebases) get higher weight
			weight := getScanWeight(compTypeName)

			work := orchestrator.ComponentWork{
				Module:        moniker,
				Component:     componentName,
				ComponentType: compTypeName,
				Handler:       "scan",
				Weight:        weight,
				BuildAfter:    nil, // Scans have no intra-module dependencies
				Index:         globalIndex,
			}

			allWork = append(allWork, work)
			globalIndex++
		}
	}

	if len(allWork) == 0 {
		return nil
	}

	// Return as single layer (scans run in parallel)
	return [][]orchestrator.ComponentWork{allWork}
}

// getScanWeight returns the weight for scanning a component type.
// Higher weight = more resource-intensive scan.
func getScanWeight(componentType string) int {
	switch componentType {
	case "dockerfile":
		return 3 // Container scans are heavy (pull images, scan layers)
	case "go", "typescript":
		return 2 // SAST scans are moderately heavy
	default:
		return 1 // Light scans (config files, etc.)
	}
}

// CountScanComponents returns the total number of scan component work items.
func CountScanComponents(layers [][]orchestrator.ComponentWork) int {
	count := 0
	for _, layer := range layers {
		count += len(layer)
	}
	return count
}

// getScanComponentCount returns the total number of scannable components.
func getScanComponentCount(ctx *cmdframework.ExecutionContext) int {
	layers := FlattenModulesToScanComponentWork(ctx)
	return CountScanComponents(layers)
}
