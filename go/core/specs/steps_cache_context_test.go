// Package specs contains godog step implementations for core features.
//
// This file contains the cache test context types, per-scenario state,
// and the step registration function that wires all cache steps together.
package specs

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// cacheContext holds test state for cache invalidation scenarios.
type cacheContext struct {
	parsedYAML     map[string]interface{}
	mockedCIStatus map[string]mockedModuleCI
	changedFiles   []string
	currentHeadSHA string
	moduleRegistry *modules.Registry // cached per-scenario to avoid repeated disk I/O
}

type mockedModuleCI struct {
	LastSuccessSHA   string
	HasFilesChanged  bool
	HasValidCIAtHead bool
	NoHistory        bool
}

// cacheCtx is per-scenario state, reset in registerCacheSteps Before/After hooks.
// Godog runs scenarios sequentially, so a package-level var is safe here.
var cacheCtx cacheContext

func resetCacheContext() {
	cacheCtx = cacheContext{
		mockedCIStatus: make(map[string]mockedModuleCI),
	}
	resetContainerMockContext()
}

// getOrLoadRegistry returns the cached module registry, loading from disk if needed.
// This avoids repeated YAML parsing within the same scenario.
func getOrLoadRegistry(isolatedDir string) (*modules.Registry, error) {
	if cacheCtx.moduleRegistry != nil {
		return cacheCtx.moduleRegistry, nil
	}
	reg, err := modules.LoadFromWorkspaceNoValidation(isolatedDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load module registry: %w", err)
	}
	cacheCtx.moduleRegistry = reg
	return reg, nil
}

// registerCacheSteps registers step definitions for cache invalidation feature specs.
func registerCacheSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	resetCacheContext()

	// Enable in-process domain dispatch for cache invalidation tests.
	// This avoids subprocess overhead (~200-500ms per call on Windows).
	ctx.CommandDispatcher = makeCoreInProcessDispatcher(ctx)

	// Hook to capture HEAD SHA after isolation is set up
	sc.After(func(ctx2 context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		resetCacheContext()
		return ctx2, nil
	})

	// Background / Setup - use hook to capture HEAD SHA when isolation is set up
	sc.StepContext().Before(func(ctx2 context.Context, st *godog.Step) (context.Context, error) {
		if strings.Contains(st.Text, "isolated test repository") && ctx.IsolatedDir != "" {
			captureHeadSHA(ctx)
		}
		return ctx2, nil
	})

	sc.Step(`^the repository has a multi-module structure with:$`, func(table *godog.Table) error {
		return setupMultiModuleStructure(ctx, table)
	})
	sc.Step(`^there is no build state file$`, func() error {
		return ensureNoBuildState(ctx)
	})
	sc.Step(`^I have built all modules successfully$`, func() error {
		return buildAllModules(ctx)
	})
	sc.Step(`^no files have been modified$`, func() error {
		return ensureNoModifications(ctx)
	})

	// File manipulation
	sc.Step(`^I append "([^"]*)" to file "([^"]*)"$`, func(content, filePath string) error {
		return appendToFile(ctx, content, filePath)
	})
	sc.Step(`^I delete the file "([^"]*)"$`, func(filePath string) error {
		return deleteFile(ctx, filePath)
	})
	sc.Step(`^I delete the build state directory$`, func() error {
		return ensureNoBuildState(ctx)
	})

	// Exit code
	sc.Step(`^the exit code is not (\d+)$`, func(notExpected int) error {
		if ctx.ExitCode == notExpected {
			return fmt.Errorf("expected exit code NOT to be %d, but it was\nOutput: %s", notExpected, ctx.CommandOutput)
		}
		return nil
	})

	// YAML output assertions
	sc.Step(`^the YAML output field "([^"]*)" is "([^"]*)"$`, func(fieldPath, expected string) error {
		return yamlFieldEquals(ctx, fieldPath, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" contains "([^"]*)"$`, func(fieldPath, expected string) error {
		return yamlFieldContains(ctx, fieldPath, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" does not contain "([^"]*)"$`, func(fieldPath, expected string) error {
		return yamlFieldNotContains(ctx, fieldPath, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" is empty$`, func(fieldPath string) error {
		return yamlFieldEmpty(ctx, fieldPath)
	})
	sc.Step(`^the YAML output field "([^"]*)" contains exactly (\d+) occurrence of "([^"]*)"$`, func(fieldPath string, count int, expected string) error {
		return yamlFieldContainsExactly(ctx, fieldPath, count, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" indicates under (\d+) milliseconds$`, func(fieldPath string, maxMs int) error {
		return yamlFieldUnderMs(ctx, fieldPath, maxMs)
	})

	// Output assertions
	sc.Step(`^the output indicates "([^"]*)" would be built$`, func(module string) error {
		return outputIndicatesWouldBuild(ctx, module)
	})
	sc.Step(`^the output indicates "([^"]*)" is up-to-date or would be skipped$`, func(module string) error {
		return outputIndicatesSkipped(ctx, module)
	})
	sc.Step(`^the output indicates "([^"]*)" would be linted$`, func(module string) error {
		return outputIndicatesWouldLint(ctx, module)
	})
	sc.Step(`^the output indicates "([^"]*)" would be skipped$`, func(module string) error {
		return outputIndicatesSkipped(ctx, module)
	})
	sc.Step(`^the output contains "([^"]*)"$`, func(expected string) error {
		return eacgodog.OutputContains(ctx, expected)
	})

	// CI mocking
	sc.Step(`^the mocked CI status shows:$`, func(table *godog.Table) error {
		return setupMockedCIStatus(ctx, table)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" has valid CI at current HEAD$`, func(module string) error {
		return mockValidCIAtHead(ctx, module)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" has valid CI at HEAD$`, func(module string) error {
		return mockValidCIAtHead(ctx, module)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" has no successful runs$`, func(module string) error {
		return mockNoCIHistory(ctx, module)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" CI at different SHA$`, func(module string) error {
		return mockCIAtDifferentSHA(ctx, module)
	})
	sc.Step(`^I run "([^"]*)" with mocked CI$`, func(cmdLine string) error {
		return runCommandWithMockedCI(ctx, cmdLine)
	})
	sc.Step(`^the only changed file is "([^"]*)"$`, func(filePath string) error {
		cacheCtx.changedFiles = []string{filePath}
		return nil
	})
	sc.Step(`^the changed files are "([^"]*)" and "([^"]*)"$`, func(file1, file2 string) error {
		cacheCtx.changedFiles = []string{file1, file2}
		return nil
	})

	// Lint state
	sc.Step(`^I have linted "([^"]*)" successfully$`, func(module string) error {
		return lintModuleSuccessfully(ctx, module)
	})
	sc.Step(`^I have a lint state showing "([^"]*)" failed$`, func(module string) error {
		return setLintStateFailed(ctx, module)
	})
	sc.Step(`^no files in "([^"]*)" have been modified$`, func(dir string) error {
		return ensureNoModificationsInDir(ctx, dir)
	})

	// Edge cases
	sc.Step(`^I have built modules "([^"]*)" and "([^"]*)"$`, func(mod1, mod2 string) error {
		return buildSpecificModules(ctx, mod1, mod2)
	})
	sc.Step(`^a new module "([^"]*)" is configured with go_root "([^"]*)"$`, func(moniker, goRoot string) error {
		return addNewModule(ctx, moniker, goRoot)
	})
	sc.Step(`^modules are configured with circular dependency:$`, func(table *godog.Table) error {
		return setupCircularDependency(ctx, table)
	})
	sc.Step(`^the build state file contains invalid JSON "([^"]*)"$`, func(content string) error {
		return corruptBuildState(ctx, content)
	})

	// Container change detection
	sc.Step(`^module "([^"]*)" has container components:$`, func(moduleName string, table *godog.Table) error {
		return setupContainerModule(ctx, moduleName, table)
	})
	sc.Step(`^the mocked container registry has no tags$`, func() error {
		return mockContainerRegistryNoTags()
	})
	sc.Step(`^the mocked container registry shows:$`, func(table *godog.Table) error {
		return mockContainerRegistryFromTable(table)
	})
	sc.Step(`^the mocked container registry returns error for "([^"]*)"$`, func(component string) error {
		return mockContainerRegistryError(component)
	})
	sc.Step(`^I commit a change to "([^"]*)"$`, func(filePath string) error {
		return commitChangeToFile(ctx, filePath)
	})
	sc.Step(`^I run "([^"]*)" with mocked container registry$`, func(cmdLine string) error {
		return runCommandWithMockedContainerRegistry(ctx, cmdLine)
	})

	// Container-specific YAML assertions
	sc.Step(`^the YAML output field "([^"]*)" has (\d+) entries$`, func(fieldPath string, n int) error {
		return yamlFieldHasNEntries(ctx, fieldPath, n)
	})
	sc.Step(`^the YAML output field "([^"]*)" contains component "([^"]*)"$`, func(fieldPath, component string) error {
		return yamlFieldContainsComponent(ctx, fieldPath, component)
	})
	sc.Step(`^each entry in "([^"]*)" has reason "([^"]*)"$`, func(fieldPath, reason string) error {
		return yamlArrayAllHaveReason(ctx, fieldPath, reason)
	})
	sc.Step(`^the component "([^"]*)" in "([^"]*)" has reason containing "([^"]*)"$`, func(component, fieldPath, substring string) error {
		return yamlComponentReasonContains(ctx, fieldPath, component, substring)
	})
}
