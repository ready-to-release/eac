package test

import (
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/commands/base"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	moduledeps "github.com/ready-to-release/eac/go/core/module-deps"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

// extractMonikerFromPath extracts the module moniker from a package path.
// Returns empty string if mapper is nil or path doesn't map to any module.
func extractMonikerFromPath(pkgPath string, mapper *ModuleMapper) string {
	if mapper == nil {
		return ""
	}
	return mapper.GetModuleForPackagePath(pkgPath)
}

func verifyTestDependencies(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) error {
	systemDeps := testing.GetSystemDependencies(testCfg.SelectedTests)
	moduleDeps := testing.GetModuleDependencies(testCfg.SelectedTests)

	if len(systemDeps) == 0 && len(moduleDeps) == 0 {
		return nil
	}

	if ctx.Config.SkipDeps {
		return nil
	}

	// Strip @deps: prefix from system dependencies to get tool IDs
	// @deps: tags use tool IDs (e.g., @deps:go, @deps:docker, @deps:go-lint)
	toolIDs := make([]string, 0, len(systemDeps))
	for _, dep := range systemDeps {
		toolID := strings.TrimPrefix(dep, "@deps:")
		toolIDs = append(toolIDs, toolID)
	}

	// Filter out platform-incompatible tools before verification
	toolIDs = tool.FilterPlatformSupported(toolIDs)

	// Verify system dependencies using tool registry
	registry := tool.GlobalRegistry()
	sysResults := registry.VerifyAll(toolIDs)

	// missingForFilter: original tool IDs used to match @deps: tags in test references.
	// missingForReport: user-visible tool names (container tools report as "docker").
	missingForFilter := make(map[string]bool)
	missingForReport := make(map[string]bool)
	for _, result := range sysResults {
		if result.Available || result.Skipped {
			continue
		}
		// When a tool resolves to a container variant (via auto fallback or
		// explicit container binding), its real dependency is docker, not the
		// tool itself. Report "docker" instead of the tool name.
		resolvedTool, ok := registry.Get(result.ToolID)
		if ok && resolvedTool.Type == tool.ToolTypeContainer {
			missingForReport["docker"] = true
		} else {
			missingForReport[result.ToolID] = true
		}
		missingForFilter[result.ToolID] = true
	}

	if len(missingForFilter) > 0 {
		// Filter out tests that require missing tools rather than blocking the entire run.
		// Tests with unsatisfied @deps: tags are excluded; all other tests proceed normally.
		filtered, removedCount := filterTestsWithMissingDeps(testCfg.SelectedTests, missingForFilter)
		if removedCount > 0 {
			missing := make([]string, 0, len(missingForReport))
			for dep := range missingForReport {
				missing = append(missing, dep)
			}
			sort.Strings(missing)
			ctx.WriteInit("⚠ Skipping %d tests: %s not installed", removedCount, strings.Join(missing, ", "))
			testCfg.SelectedTests = filtered
			testCfg.Stats.Selected = len(filtered)
			testCfg.Stats.ModulesInScope = getUniqueModulesFromTests(filtered, testCfg.ModuleMapper)
		}
	}

	// Verify module dependencies
	modResults := moduledeps.VerifyAll(moduleDeps)
	for _, result := range modResults {
		if !result.Available {
			log.Warnf("Module dependency not available: %s", result.Dependency)
		}
	}

	return nil
}

// filterTestsWithMissingDeps removes test references that require any of the given missing tool IDs.
// Returns the filtered list and the count of removed tests.
func filterTestsWithMissingDeps(tests []testing.TestReference, missingToolIDs map[string]bool) ([]testing.TestReference, int) {
	filtered := make([]testing.TestReference, 0, len(tests))
	removed := 0
	for _, test := range tests {
		hasMissingDep := false
		for _, tag := range test.Tags {
			if strings.HasPrefix(tag, "@deps:") {
				toolID := strings.TrimPrefix(tag, "@deps:")
				if missingToolIDs[toolID] {
					hasMissingDep = true
					break
				}
			}
		}
		if hasMissingDep {
			removed++
		} else {
			filtered = append(filtered, test)
		}
	}
	return filtered, removed
}

func validateTestArtifacts(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) error {
	stats := testCfg.Stats

	if len(stats.ModulesInScope) == 0 || ctx.Config.SkipDepm {
		return nil
	}

	// Get expected build UoWs to filter validation (ignores orphaned manifests)
	expectedBuildUoWs := getExpectedBuildUoWs(ctx)

	artifactValidation := base.ValidateBuildArtifactsWithExpected(
		stats.ModulesInScope,
		ctx.EACConfig,
		ctx.WorkspaceRoot,
		ctx.ModuleRegistry,
		expectedBuildUoWs,
	)

	if artifactValidation.AllValid() {
		return nil
	}

	// Collect affected modules for the resolution command
	var affectedModules []string

	ctx.WriteInit("")
	ctx.WriteInit("┌─────────────────────────────────────────────────────────────────┐")
	ctx.WriteInit("│  Build Required                                                 │")
	ctx.WriteInit("└─────────────────────────────────────────────────────────────────┘")
	ctx.WriteInit("")

	if !artifactValidation.AllPresent {
		ctx.WriteInit("Missing build artifacts:")
		for _, moduleName := range artifactValidation.MissingFrom {
			ctx.WriteInit("  • %s", moduleName)
			affectedModules = append(affectedModules, moduleName)
		}
		ctx.WriteInit("")
	}

	if !artifactValidation.AllCurrent {
		ctx.WriteInit("Stale build artifacts (source changed since last build):")
		for _, moduleName := range artifactValidation.StaleModules {
			reason := artifactValidation.StaleReasons[moduleName]
			ctx.WriteInit("  • %s: %s", moduleName, reason)
			// Only add if not already in the list
			found := false
			for _, m := range affectedModules {
				if m == moduleName {
					found = true
					break
				}
			}
			if !found {
				affectedModules = append(affectedModules, moduleName)
			}
		}
		ctx.WriteInit("")
	}

	// Show resolution command
	ctx.WriteInit("To fix, run:")
	if len(affectedModules) == 1 {
		ctx.WriteInit("  eac build %s", affectedModules[0])
	} else if len(affectedModules) <= 3 {
		ctx.WriteInit("  eac build %s", strings.Join(affectedModules, " "))
	} else {
		ctx.WriteInit("  eac build")
	}
	ctx.WriteInit("")

	// Return informational exit - no additional error message needed
	return cmdframework.NewInformationalExit("build required")
}

// extractUniqueModulesFromPaths extracts unique module monikers from package paths.
// Used to convert test package paths to module monikers for ScopeMonikers.
func extractUniqueModulesFromPaths(paths []string, mapper *ModuleMapper) []string {
	seen := make(map[string]bool)
	var modules []string
	for _, path := range paths {
		moniker := mapper.GetModuleForPackagePath(path)
		if moniker != "" && !seen[moniker] {
			seen[moniker] = true
			modules = append(modules, moniker)
		}
	}
	sort.Strings(modules)
	return modules
}

// removeExistingModules returns items from 'from' that are not in 'existing'.
// Used to deduplicate modules between parallel and sequential layers.
func removeExistingModules(from, existing []string) []string {
	set := make(map[string]bool)
	for _, s := range existing {
		set[s] = true
	}
	var result []string
	for _, s := range from {
		if !set[s] {
			result = append(result, s)
		}
	}
	return result
}

// preComputeModuleHashes computes input hashes for all modules in scope.
// Called once before test execution to ensure detection and workers use identical hashes.
func preComputeModuleHashes(ctx *cmdframework.ExecutionContext, testCfg *TestFrameworkConfig) map[string]string {
	if ctx.ModuleRegistry == nil || testCfg.TestsByPackage == nil {
		return nil
	}

	moduleHashes := make(map[string]string)
	seen := make(map[string]bool)

	for pkgPath := range testCfg.TestsByPackage {
		moniker := testCfg.ModuleMapper.GetModuleForPackagePath(pkgPath)
		if moniker == "" || seen[moniker] {
			continue
		}
		seen[moniker] = true

		contract, ok := ctx.ModuleRegistry.Get(moniker)
		if !ok {
			continue
		}

		h, err := computeTestInputHash(ctx, contract)
		if err != nil {
			log.Debugf("Failed to pre-compute hash for %s: %v", moniker, err)
			continue
		}
		moduleHashes[moniker] = h
	}

	return moduleHashes
}

// getSuitesIncluded parses composite suite syntax.
// Handles "unit+integration" -> ["unit", "integration"]. For single suites, returns nil.
func getSuitesIncluded(suiteMoniker string) []string {
	if strings.Contains(suiteMoniker, "+") {
		return strings.Split(suiteMoniker, "+")
	}
	return nil
}
