package test

import (
	"math"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ResolveTestUnitSpecs converts TestsByPackage to component work specs.
// Returns parallel tests first, followed by sequential tests.
// Returns nil if no tests to execute.
// Work items are created for each unique component:tool combination,
// allowing parallel execution of different test types (gotest, godog) within the same package.
//
// Test keys use the same format as build/lint/scan: "module:component:tool".
func ResolveTestUnitSpecs(ctx *cmdframework.ExecutionContext) []workunit.UnitSpec {
	testCfg, ok := ctx.Config.TestCmdConfig.(*TestFrameworkConfig)
	if !ok || testCfg == nil {
		return nil
	}

	testsByPackage := testCfg.TestsByPackage
	if len(testsByPackage) == 0 {
		return nil
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentKinds == nil {
		return nil
	}

	// Initialize component-to-pkgPath mapping for clean directory names
	if testCfg.ComponentToPkgPath == nil {
		testCfg.ComponentToPkgPath = make(map[string]string)
	}

	// Initialize UoW tag map for manifest writing
	if testCfg.UoWTags == nil {
		testCfg.UoWTags = make(map[string]workunit.TagSummary)
	}

	var parallelWork []workunit.UnitSpec
	var sequentialWork []workunit.UnitSpec

	for pkgPath, tests := range testsByPackage {
		if len(tests) == 0 {
			continue
		}

		// Get module ownership for this package path
		// Module mapping is configured via test-impl component in blueprints.yml component-kinds
		moduleMoniker := testCfg.ModuleMapper.GetModuleForPackagePath(pkgPath)
		if moduleMoniker == "" {
			log.Warnf("ResolveTestUnitSpecs: no module found for pkgPath=%s, skipping", pkgPath)
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
			compTypeName := getTestTypeComponentType(ctx.ToolSystem, testType)
			componentName := findComponentForTests(ctx, moduleMoniker, compTypeName, pkgPath)
			weight := getTestComponentWeight(ctx.ToolSystem, moduleMoniker, componentName, typeTests)

			// Compute testname - unique identifier within module:component
			// For BDD tests: use spec name (e.g., "build-module")
			// For unit tests: use path-based name (e.g., "impl-build")
			spec := ""
			testname := ""
			if testType == "godog" || testType == "tscucumber" {
				spec = extractSpecName(pkgPath)
				testname = spec
			} else {
				// Path-based testname unique within module
				// e.g., "go/cli/eac/impl/build" -> "impl-build"
				testname = uniqueComponentName(pkgPath, moduleMoniker, testCfg.ModuleMapper)
			}

			// Skip if testname could not be determined (configuration error)
			if testname == "" {
				log.Warnf("ResolveTestUnitSpecs: skipping tests at pkgPath=%s - could not determine unique testname", pkgPath)
				continue
			}

			// Compute tag summary for this UoW group
			tagSummary := testing.AggregateTagsForTests(typeTests)

			// Tool is the test type (gotest, godog, etc.)
			toolName := testType
			if toolName == "" {
				toolName = "none"
			}

			// Build UnitID with structure: test:module:componentType:componentName:tool:testname
			// - ComponentType = component TYPE from blueprints.yml component-kinds (e.g., "go", "gherkin")
			// - ComponentName = component instance name within module (e.g., "go")
			// - Extra["testname"] = unique test identifier within module:component
			// This gives: Longname = test:eac:go:go:gotest:impl-build
			//             DirName  = go-gotest-impl-build
			unitID := workunit.UnitID{
				Action:        core.ActionTest,
				Module:        moduleMoniker,
				ComponentType: compTypeName,  // Component TYPE from blueprints.yml component-kinds
				ComponentName: componentName, // Component instance name from findComponentOfType
				Tool:          toolName,
				Spec:          spec,
				Extra:         map[string]string{"testname": testname},
			}

			// Check UoW-level cache first, fall back to module-level
			isCached := false
			if testCfg.CachedUoWs != nil && testCfg.CachedUoWs[unitID.Longname()] {
				isCached = true
			} else if testCfg.CachedModules != nil && testCfg.CachedModules[moduleMoniker] {
				isCached = true
			}

			isContainer := false
			if ctx.ToolSystem != nil && ctx.ToolSystem.TestBridge != nil {
				isContainer = ctx.ToolSystem.TestBridge.IsContainer(compTypeName)
			}
			work := workunit.UnitSpec{
				ID:             unitID, // Use the same UnitID for consistency
				ComponentType:  testType,
				Weight:         weight,
				PoolAllocation: core.AllocationForWeight(weight, isContainer),
				DependsOn:      nil, // Tests don't have intra-module deps
				Cached:         isCached,
				Metadata:       map[string]any{"pkgPath": pkgPath}, // Full path for test lookup
				Index:          0,                                  // Will be set per-layer below
				Tags:           tagSummary,
			}

			// Store mapping from testname to pkgPath for worker lookup
			// Key is testname (unique within module:component)
			testCfg.ComponentToPkgPath[testname] = pkgPath

			// Store tag summary keyed by UoW longname for manifest writing
			testCfg.UoWTags[unitID.Longname()] = tagSummary

			if hasSequential {
				sequentialWork = append(sequentialWork, work)
			} else {
				parallelWork = append(parallelWork, work)
			}
		}
	}

	// Combine parallel and sequential work into a single flat slice
	// Parallel work comes first, followed by sequential work
	var specs []workunit.UnitSpec
	specs = append(specs, parallelWork...)
	specs = append(specs, sequentialWork...)

	// Set indices globally
	for i := range specs {
		specs[i].Index = i
	}

	return specs
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
// Uses the ToolSystem's config test-type-mapping, with fallback to defaults.
// If ts is nil, falls back to built-in defaults via a minimal ToolSystem.
func getTestTypeComponentType(ts *tool.ToolSystem, testType string) string {
	if ts != nil {
		return ts.GetTestTypeComponentType(testType)
	}
	// Nil ToolSystem: use a zero-config instance for built-in fallback behavior
	empty := tool.NewToolSystemForTesting()
	return empty.GetTestTypeComponentType(testType)
}

// findComponentOfType finds the first component of the given type in a module.
// Returns empty string if not found.
// findComponentForTests finds the correct component for a test package path.
// For BDD tests (pkgPath format: "specname:testRoot:featurePath"), it uses the
// feature file path to match against component roots via longest-prefix matching.
// This ensures deterministic results when a module has multiple components of the same type.
// Falls back to findComponentOfType for non-BDD or unresolved cases.
func findComponentForTests(ctx *cmdframework.ExecutionContext, moniker, compTypeName, pkgPath string) string {
	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		return ""
	}

	// For BDD tests, extract the feature file path and match against component roots
	parts := strings.SplitN(pkgPath, ":", 3)
	if len(parts) == 3 {
		featurePath := filepath.ToSlash(parts[2])
		var bestName string
		var bestLen int
		for name := range module.Components {
			if module.Components.GetComponentType(name) != compTypeName {
				continue
			}
			root := filepath.ToSlash(module.Components.GetComponentRoot(name))
			if root == "" {
				continue
			}
			if strings.HasPrefix(featurePath, root+"/") || featurePath == root {
				if len(root) > bestLen {
					bestLen = len(root)
					bestName = name
				}
			}
		}
		if bestName != "" {
			return bestName
		}
	}

	// Fallback: find any component of the matching type (deterministic: shortest name wins)
	return findComponentOfType(ctx, moniker, compTypeName)
}

func findComponentOfType(ctx *cmdframework.ExecutionContext, moniker, compTypeName string) string {
	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		return ""
	}
	var best string
	for name := range module.Components {
		if module.Components.GetComponentType(name) == compTypeName {
			if best == "" || name < best {
				best = name
			}
		}
	}
	return best
}

// getTestComponentWeight returns the scheduling weight for a set of tests.
// Weight = base tool weight × component amp (from config).
func getTestComponentWeight(ts *tool.ToolSystem, moniker, componentName string, tests []testing.TestReference) int {
	if len(tests) == 0 {
		return 1
	}

	// Map test type to component type for tool lookup
	compTypeName := getTestTypeComponentType(ts, tests[0].Type)

	// Get base weight from tool resources via test bridge
	baseWeight := 1
	if ts != nil && ts.TestBridge != nil {
		if t := ts.TestBridge.ResolveTool(compTypeName, core.ActionTest); t != nil {
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
// Example: "build-module:go/eac/specs/impl/eac:specs/eac/build-module/specification.feature"
// Returns the spec name (first part before colon), or empty string if not found.
// The result is sanitized for use in file paths.
func extractSpecName(pkgPath string) string {
	parts := strings.SplitN(pkgPath, ":", 2)
	if len(parts) >= 1 && parts[0] != "" {
		return sanitizeTestname(parts[0])
	}
	return ""
}

// sanitizeTestname makes a testname safe for use in filesystem paths.
// Replaces characters that are problematic on Windows and Unix:
// - Colons, slashes, backslashes are replaced with dashes
// - Other unsafe characters (<>"|?*) are replaced with underscores
// - Spaces are replaced with dashes
// - Multiple consecutive dashes are collapsed
// - Leading/trailing dashes are trimmed
func sanitizeTestname(name string) string {
	if name == "" {
		return name
	}

	// Replace path separators and colons with dashes
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")

	// Replace spaces with dashes
	name = strings.ReplaceAll(name, " ", "-")

	// Replace other unsafe characters with underscores
	unsafeChars := []string{"<", ">", "\"", "|", "?", "*"}
	for _, c := range unsafeChars {
		name = strings.ReplaceAll(name, c, "_")
	}

	// Collapse multiple consecutive dashes into one
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Trim leading and trailing dashes
	name = strings.Trim(name, "-")

	return name
}

// uniqueComponentName generates a unique component name from a package path.
// Uses longest prefix matching to find the most specific component root.
// This ensures deterministic results regardless of Go map iteration order.
//
// Example: For pkgPath "go/cli/eac/impl/validate/risk_profile" with roots:
//   - "go/cli/eac"
//   - "go/cli/eac/impl"
//
// The "go/cli/eac/impl" root wins (longer prefix), producing "validate-risk_profile"
// Returns empty string if the path cannot be mapped to a unique name (caller should skip).
func uniqueComponentName(pkgPath, moduleMoniker string, mapper *ModuleMapper) string {
	normalizedPath := filepath.ToSlash(pkgPath)

	if mapper == nil || mapper.registry == nil {
		log.Warnf("uniqueComponentName: mapper unavailable for pkgPath=%s, moduleMoniker=%s", pkgPath, moduleMoniker)
		return ""
	}

	module, exists := mapper.registry.Get(moduleMoniker)
	if !exists {
		log.Warnf("uniqueComponentName: module not found: moniker=%s, pkgPath=%s", moduleMoniker, pkgPath)
		return ""
	}

	// Use longest prefix match to ensure deterministic results
	// (Go map iteration order is undefined, so "first match" would be non-deterministic)
	var bestRoot string
	var bestSuffix string

	for _, root := range module.GetComponentRoots() {
		if root == "" || root == "/" {
			continue
		}
		root = filepath.ToSlash(root)

		// Prefix match: pkgPath starts with root/
		if strings.HasPrefix(normalizedPath, root+"/") {
			suffix := strings.TrimPrefix(normalizedPath, root+"/")
			if suffix != "" && (bestRoot == "" || len(root) > len(bestRoot)) {
				bestRoot = root
				bestSuffix = suffix
			}
		}

		// Exact match: pkgPath equals root
		if normalizedPath == root {
			if bestRoot == "" || len(root) > len(bestRoot) {
				bestRoot = root
				bestSuffix = filepath.Base(root)
			}
		}
	}

	if bestSuffix == "" {
		log.Warnf("uniqueComponentName: no component root matched for pkgPath=%s in module=%s (roots=%v)",
			pkgPath, moduleMoniker, module.GetComponentRoots())
		return ""
	}

	return sanitizeTestname(bestSuffix)
}
