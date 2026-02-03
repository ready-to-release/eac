package test

import (
	"math"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

var componentWorkLog = logging.C()

// ResolveTestUnitSpecs converts TestsByPackage to component work layers.
// Returns two layers: parallel tests first, sequential tests second.
// Returns nil if no tests to execute.
// Work items are created for each unique component:tool combination,
// allowing parallel execution of different test types (gotest, godog) within the same package.
//
// Test keys use the same format as build/lint/scan: "module:component:tool".
func ResolveTestUnitSpecs(ctx *cmdframework.ExecutionContext) [][]workunit.UnitSpec {
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

	// Initialize component-to-pkgPath mapping for clean directory names
	if testCfg.ComponentToPkgPath == nil {
		testCfg.ComponentToPkgPath = make(map[string]string)
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
			componentWorkLog.Warnf("ResolveTestUnitSpecs: no module found for pkgPath=%s, skipping", pkgPath)
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
			// For BDD tests, use spec name as the component identifier
			// For unit tests, use a unique path-based name to avoid collisions
			spec := ""
			cleanName := ""
			if testType == "godog" || testType == "tscucumber" {
				spec = extractSpecName(pkgPath)
				cleanName = spec // e.g., "docs-drawio-cache"
			} else {
				// For regular tests, use unique component name based on path
				// This avoids collisions when multiple packages have the same basename
				// e.g., "go/cli/eac/impl/internal" -> "impl/internal" -> "impl-internal"
				cleanName = uniqueComponentName(pkgPath, moduleMoniker, testCfg.ModuleMapper)
			}

			// Tool is the test type (gotest, godog, etc.) - consistent with build/lint/scan
			toolName := testType
			if toolName == "" {
				toolName = "none"
			}

			isContainer := tool.GlobalTestBridge().IsContainer(compTypeName)
			work := workunit.UnitSpec{
				ID: workunit.UnitID{
					Context:   workunit.ContextTest,
					Module:    moduleMoniker,
					Component: cleanName, // Clean component name for directory creation
					Tool:      toolName,  // Test tool (gotest, godog, etc.)
					Extra:     map[string]string{"testset": testType},
					Spec:      spec, // Spec name for BDD tests (e.g., "build-module")
				},
				ComponentType: testType,
				Weight:        weight,
				Container:     isContainer,
				HostInstalled: !isContainer,
				DependsOn:     nil, // Tests don't have intra-module deps
				Cached:        isCached,
				Metadata:      map[string]any{"pkgPath": pkgPath}, // Full path for test lookup
				Index:         0,                                  // Will be set per-layer below
			}

			// Store mapping from component name to pkgPath for worker lookup
			testCfg.ComponentToPkgPath[cleanName] = pkgPath

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
// Example: "build-module:go/eac/specs/impl/eac-cli:specs/eac-cli/build-module/specification.feature"
// Returns the spec name (first part before colon), or empty string if not found.
func extractSpecName(pkgPath string) string {
	parts := strings.SplitN(pkgPath, ":", 2)
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// uniqueComponentName generates a unique component name from a package path.
// Uses the path relative to the module root to avoid collisions between packages
// with the same basename (e.g., multiple "internal" packages).
// Example: "go/cli/eac/impl/internal" for module "eac-cli" -> "impl-internal"
func uniqueComponentName(pkgPath, moduleMoniker string, mapper *ModuleMapper) string {
	// Normalize path
	normalizedPath := filepath.ToSlash(pkgPath)

	// Get module info to find the component root
	if mapper != nil && mapper.registry != nil {
		if module, exists := mapper.registry.Get(moduleMoniker); exists {
			// Try to find a matching component root
			for _, root := range module.GetComponentRoots() {
				if root == "" || root == "/" {
					continue
				}
				root = filepath.ToSlash(root)
				// Check if pkgPath is under this component root
				if strings.HasPrefix(normalizedPath, root+"/") {
					// Extract the suffix after the root
					suffix := strings.TrimPrefix(normalizedPath, root+"/")
					if suffix != "" {
						// Replace slashes with dashes for a clean component name
						return strings.ReplaceAll(suffix, "/", "-")
					}
				}
				// Exact match (package is at the root itself)
				if normalizedPath == root {
					return filepath.Base(root)
				}
			}
		}
	}

	// Fallback: use basename (may not be unique, but better than nothing)
	// This handles edge cases where module mapping isn't available
	return filepath.Base(pkgPath)
}
