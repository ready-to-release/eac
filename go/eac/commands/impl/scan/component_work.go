package scan

import (
	"math"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

// FlattenModulesToScanComponentWork converts modules to scan component work items.
// Each module's scannable component + scanner combination becomes a work item.
// This allows parallel execution of different scanners (trivy-vuln, semgrep, etc.)
// within the same module.
// Returns nil if no scannable components are found.
func FlattenModulesToScanComponentWork(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
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

	var allWork []workunit.UnitSpec
	globalIndex := 0

	for _, moniker := range monikers {
		// Get module contract
		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		// Get enabled components for this module
		enabledComponents := module.GetEnabledComponents()

		// Process each component
		for _, componentName := range enabledComponents {
			compTypeName := module.Components.GetComponentType(componentName)
			compType := cfg.ComponentTypes.Get(compTypeName)
			if compType == nil || !compType.IsScannable() {
				continue
			}

			// Get scanners from component type configuration
			scannerStrings := compType.GetScanners()
			if len(scannerStrings) == 0 {
				continue
			}

			// Create one work item per scanner (allows parallel scanner execution)
			seenScanners := make(map[string]bool)
			for _, scannerStr := range scannerStrings {
				if seenScanners[scannerStr] {
					continue
				}
				seenScanners[scannerStr] = true

				toolID := tool.ScannerToolIDForCategory(scannerStr)
				if toolID == "" {
					continue
				}
				scannerType := internal.ScannerType(toolID)

				// Get weight (base weight × amp, calculated internally)
				weight := getComponentWeight(moniker, componentName, scannerType)

				isContainer := tool.GlobalScanBridge().IsContainer(string(scannerType))
				work := workunit.UnitSpec{
					ID: workunit.UnitID{
						Context:   workunit.ContextScan,
						Module:    moniker,
						Component: componentName,
						Tool:      string(scannerType), // Use scanner type instead of generic "scan"
					},
					ComponentType:   compTypeName,
					Weight:          weight,
					IsContainer:     isContainer,
					IsHostInstalled: !isContainer,
					DependsOn:       nil, // Scans have no intra-module dependencies
					Cached:          false,
					Metadata:        make(map[string]any),
					Index:           globalIndex,
				}

				allWork = append(allWork, work)
				globalIndex++
			}
		}
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

// getComponentWeight returns the scheduling weight for a component.
// Weight = base scanner weight × component amp (from config).
func getComponentWeight(moniker, componentName string, scannerType internal.ScannerType) int {
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

// getScanComponentCount returns the total number of scannable components.
func getScanComponentCount(ctx *cmdframework.ExecutionContext) int {
	layers := FlattenModulesToScanComponentWork(ctx)
	return CountScanComponents(layers)
}
