package test

import (
	"math"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/testing"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

var componentWorkLog = logging.C()

// FlattenModulesToTestUnits converts TestsByPackage to component work layers.
// Returns two layers: parallel tests first, sequential tests second.
// Returns nil if no tests to execute.
// Work items are created for each unique path:testType combination,
// allowing parallel execution of different test types (gotest, godog) within the same package.
//
// Test keys use path-based format: "go/eac/core/config:gotest"
// This differs from build/lint/scan which use component-based format: "module:component:tool".
func FlattenModulesToTestUnits(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
	testCfg, ok := ctx.Config.Extra["testConfig"].(*TestFrameworkConfig)
	if !ok || testCfg == nil {
		return nil
	}

	testsByPackage := testCfg.TestsByPackage
	if len(testsByPackage) == 0 {
		return nil
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	var parallelWork []workunit.UnitSpec
	var sequentialWork []workunit.UnitSpec

	for pkgPath, tests := range testsByPackage {
		if len(tests) == 0 {
			continue
		}

		// Get module ownership for this package path
		// Module mapping is configured via test-impl component in component-types.yml
		moduleMoniker := testCfg.ModuleMapper.GetModuleForPackagePath(pkgPath)
		if moduleMoniker == "" {
			componentWorkLog.Warnf("FlattenModulesToTestUnits: no module found for pkgPath=%s, skipping", pkgPath)
			continue
		}

		// Group tests by type to create separate work items per test type
		// This allows parallel execution of gotest and godog tests within the same package
		testsByType := groupTestsByType(tests)

		for testType, typeTests := range testsByType {
			// Check if any test of this type is sequential
			hasSequential := false
			for i := range typeTests {
				if typeTests[i].IsSequential {
					hasSequential = true
					break
				}
			}

			// Get weight (base weight × amp, calculated internally)
			// For tests, we find the component by mapping test type -> component type
			compTypeName := getTestTypeComponentType(testType)
			componentName := findComponentOfType(ctx, moduleMoniker, compTypeName)
			weight := getTestComponentWeight(moduleMoniker, componentName, typeTests)

			// Check if module is cached
			isCached := testCfg.CachedModules != nil && testCfg.CachedModules[moduleMoniker]

			// Extract spec name for BDD tests (godog, tscucumber)
			spec := ""
			if testType == "godog" || testType == "tscucumber" {
				spec = extractSpecName(pkgPath)
			}

			isContainer := tool.GlobalTestBridge().IsContainer(compTypeName)
			work := workunit.UnitSpec{
				ID: workunit.UnitID{
					Context:   workunit.ContextTest,
					Module:    moduleMoniker,
					Component: pkgPath + ":" + testType,
					Tool:      "", // Don't duplicate - testType is already in Component
					Extra:     map[string]string{"testset": testType},
					Spec:      spec, // Spec name for BDD tests (e.g., "build-module")
				},
				ComponentType:   testType,
				Weight:          weight,
				IsContainer:     isContainer,
				IsHostInstalled: !isContainer,
				DependsOn:       nil, // Tests don't have intra-module deps
				Cached:          isCached,
				Metadata:        make(map[string]any),
				Index:           0, // Will be set per-layer below
			}

			if hasSequential {
				sequentialWork = append(sequentialWork, work)
			} else {
				parallelWork = append(parallelWork, work)
			}
		}
	}

	// Set indices per-layer (Index must be relative to the layer, not global)
	for i := range parallelWork {
		parallelWork[i].Index = i
	}
	for i := range sequentialWork {
		sequentialWork[i].Index = i
	}

	// Build layers: parallel first, sequential second
	var layers [][]workunit.UnitSpec
	if len(parallelWork) > 0 {
		layers = append(layers, parallelWork)
	}
	if len(sequentialWork) > 0 {
		layers = append(layers, sequentialWork)
	}

	return layers
}

// groupTestsByType groups tests by their type (e.g., "gotest", "godog").
func groupTestsByType(tests []testing.TestReference) map[string][]testing.TestReference {
	result := make(map[string][]testing.TestReference)
	for i := range tests {
		testType := tests[i].Type
		result[testType] = append(result[testType], tests[i])
	}
	return result
}

// getTestTypeComponentType maps test type to component type for tool lookup.
func getTestTypeComponentType(testType string) string {
	switch testType {
	case "gotest", "godog":
		return "go"
	case "mocha", "tscucumber":
		return "typescript"
	default:
		return "go" // Default
	}
}

// findComponentOfType finds the first component of the given type in a module.
// Returns empty string if not found.
func findComponentOfType(ctx *cmdframework.ExecutionContext, moniker, compTypeName string) string {
	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		return ""
	}
	for name := range module.Components {
		if module.Components.GetComponentType(name) == compTypeName {
			return name
		}
	}
	return ""
}

// getTestComponentWeight returns the scheduling weight for a set of tests.
// Weight = base tool weight × component amp (from config).
func getTestComponentWeight(moniker, componentName string, tests []testing.TestReference) int {
	if len(tests) == 0 {
		return 1
	}

	// Map test type to component type for tool lookup
	compTypeName := getTestTypeComponentType(tests[0].Type)

	// Get base weight from tool resources via test bridge
	baseWeight := 1
	bridge := tool.GlobalTestBridge()
	if bridge != nil {
		if t := bridge.ResolveTool(compTypeName, tool.OperationTest); t != nil {
			baseWeight = t.Resources.Weight()
		}
	}

	// Get amp from config (the source of truth)
	amp := 1.0
	if componentName != "" {
		cfg := config.Global()
		if cfg != nil && cfg.Repository != nil {
			if module, ok := cfg.Repository.GetModule(moniker); ok && module != nil {
				amp = module.GetComponentAmp(componentName, "test")
			}
		}
	}

	// Apply amp to weight (ceil to ensure at least 1)
	weight := int(math.Ceil(float64(baseWeight) * amp))
	if weight < 1 {
		weight = 1
	}

	return weight
}

// extractSpecName extracts the spec name from a BDD pkgPath.
// For godog tests, pkgPath format is: "specname:testRoot:featurePath"
// Example: "build-module:go/eac/specs/impl/eac-commands:specs/eac-commands/build-module/specification.feature"
// Returns the spec name (first part before colon), or empty string if not found.
func extractSpecName(pkgPath string) string {
	parts := strings.SplitN(pkgPath, ":", 2)
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}
