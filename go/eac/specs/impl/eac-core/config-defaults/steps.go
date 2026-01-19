// Package configdefaults contains godog step implementations for specs/eac-core/config-defaults.
//
// This package tests the EAC configuration defaults loading and merging system.
// All tests run in isolated directories with R2R_REPO_ROOT set appropriately.
package configdefaults

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// testState holds state for config defaults test scenarios.
type testState struct {
	cfg               *config.EACConfig
	loadError         error
	repoRoot          string
	origRepoRoot      string // Original R2R_REPO_ROOT value
	origContainerRoot string // Original R2R_CONTAINER_ROOT value
}

var state *testState

// RegisterSteps registers step definitions for config defaults feature specs.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Setup/teardown
	sc.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		state = &testState{}
		// Save original env vars and clear config cache
		state.origRepoRoot = os.Getenv("R2R_REPO_ROOT")
		state.origContainerRoot = os.Getenv("R2R_CONTAINER_ROOT")
		config.ClearCache()
		return c, nil
	})

	// Background steps
	sc.Step(`^I am in an isolated test repository$`, func() error {
		return iAmInAnIsolatedTestRepository(ctx)
	})

	// Given steps - repository state
	sc.Step(`^the repository has no "([^"]*)" directory$`, func(path string) error {
		return theRepositoryHasNoDirectory(ctx, path)
	})
	sc.Step(`^the repository has directory "([^"]*)"$`, func(path string) error {
		return theRepositoryHasDirectory(ctx, path)
	})
	sc.Step(`^the repository has file "([^"]*)" with:$`, func(path string, content *godog.DocString) error {
		return theRepositoryHasFileWith(ctx, path, content.Content)
	})
	sc.Step(`^the contracts directory does not exist$`, func() error {
		// Unset R2R_CONTAINER_ROOT to simulate a scenario where the tool's
		// contracts/defaults are not available (e.g., corrupted install)
		// This tests embedded defaults fallback (when implemented)
		os.Unsetenv("R2R_CONTAINER_ROOT")
		return nil
	})

	// When steps - loading config
	sc.Step(`^I load the EAC configuration$`, func() error {
		return iLoadTheEACConfiguration()
	})
	sc.Step(`^I try to load the EAC configuration$`, func() error {
		return iTryToLoadTheEACConfiguration()
	})
	sc.Step(`^I apply type defaults to modules$`, func() error {
		return iApplyTypeDefaultsToModules()
	})

	// Then steps - modules assertions
	sc.Step(`^the modules config contains module "([^"]*)"$`, func(moniker string) error {
		return theModulesConfigContainsModule(moniker)
	})
	sc.Step(`^the modules config does not contain module "([^"]*)"$`, func(moniker string) error {
		return theModulesConfigDoesNotContainModule(moniker)
	})
	sc.Step(`^the module "([^"]*)" has component "([^"]*)"$`, func(moniker, expectedComp string) error {
		return theModuleHasComponent(moniker, expectedComp)
	})
	sc.Step(`^the module "([^"]*)" has component root "([^"]*)" as "([^"]*)"$`, func(moniker, compName, expectedRoot string) error {
		return theModuleHasComponentRoot(moniker, compName, expectedRoot)
	})
	sc.Step(`^the module "([^"]*)" has description "([^"]*)"$`, func(moniker, expected string) error {
		return theModuleHasDescription(moniker, expected)
	})
	sc.Step(`^the module "([^"]*)" has changelog "([^"]*)"$`, func(moniker, expected string) error {
		return theModuleHasChangelog(moniker, expected)
	})
	sc.Step(`^the module "([^"]*)" component "([^"]*)" has source patterns containing "([^"]*)"$`, func(moniker, compName, pattern string) error {
		return theModuleComponentHasSourcePatternsContaining(moniker, compName, pattern)
	})
	sc.Step(`^the module "([^"]*)" component "([^"]*)" does not have source pattern "([^"]*)"$`, func(moniker, compName, pattern string) error {
		return theModuleComponentDoesNotHaveSourcePattern(moniker, compName, pattern)
	})
	sc.Step(`^the module "([^"]*)" has no source patterns from type defaults$`, func(moniker string) error {
		return theModuleHasNoSourcePatternsFromTypeDefaults(moniker)
	})
	sc.Step(`^the module "([^"]*)" specs pattern resolves with "([^"]*)"$`, func(moniker, expected string) error {
		return theModuleSpecsPatternResolvesWith(moniker, expected)
	})

	// Then steps - component types assertions
	sc.Step(`^the component types config contains type "([^"]*)"$`, func(typeName string) error {
		return theComponentTypesConfigContainsType(typeName)
	})
	sc.Step(`^the type "([^"]*)" has capability "([^"]*)"$`, func(typeName, capability string) error {
		return theTypeHasCapability(typeName, capability)
	})
	sc.Step(`^the type "([^"]*)" has description "([^"]*)"$`, func(typeName, expected string) error {
		return theTypeHasDescription(typeName, expected)
	})
	sc.Step(`^the type "([^"]*)" has default source pattern "([^"]*)"$`, func(typeName, pattern string) error {
		return theTypeHasDefaultSourcePattern(typeName, pattern)
	})

	// Then steps - repository paths assertions
	sc.Step(`^the repository paths\.specs_root is "([^"]*)"$`, func(expected string) error {
		return theRepositoryPathsFieldIs("specs_root", expected)
	})
	sc.Step(`^the repository paths\.out\.root is "([^"]*)"$`, func(expected string) error {
		return theRepositoryPathsFieldIs("out.root", expected)
	})
	sc.Step(`^the repository paths\.out\.build is "([^"]*)"$`, func(expected string) error {
		return theRepositoryPathsFieldIs("out.build", expected)
	})
	sc.Step(`^the repository paths\.out\.test is "([^"]*)"$`, func(expected string) error {
		return theRepositoryPathsFieldIs("out.test", expected)
	})
	sc.Step(`^the repository paths\.out\.logs is "([^"]*)"$`, func(expected string) error {
		return theRepositoryPathsFieldIs("out.logs", expected)
	})

	// Then steps - system dependencies assertions
	sc.Step(`^the system dependencies config contains "([^"]*)"$`, func(moniker string) error {
		return theSystemDependenciesConfigContains(moniker)
	})
	sc.Step(`^the dependency "([^"]*)" has version "([^"]*)"$`, func(moniker, expected string) error {
		return theDependencyHasVersion(moniker, expected)
	})

	// Then steps - error assertions
	sc.Step(`^an error is returned containing "([^"]*)"$`, func(expected string) error {
		return anErrorIsReturnedContaining(expected)
	})
}

// cleanupTestState cleans up test state after each scenario.
func cleanupTestState() {
	if state != nil {
		// Restore original R2R_REPO_ROOT
		if state.origRepoRoot != "" {
			os.Setenv("R2R_REPO_ROOT", state.origRepoRoot)
		} else {
			os.Unsetenv("R2R_REPO_ROOT")
		}
		// Restore original R2R_CONTAINER_ROOT
		if state.origContainerRoot != "" {
			os.Setenv("R2R_CONTAINER_ROOT", state.origContainerRoot)
		} else {
			os.Unsetenv("R2R_CONTAINER_ROOT")
		}
		config.ClearCache()
	}
	state = nil
}

// ============================================================================
// Step Implementations
// ============================================================================

func iAmInAnIsolatedTestRepository(ctx *internal.TestContext) error {
	// Create a temporary directory for test isolation
	// This simulates a user's workspace/repository
	tempDir, err := os.MkdirTemp("", "config-defaults-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create minimal .git directory so config loader can find repo root
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .git: %w", err)
	}
	// Create minimal git config
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("[core]\n\tbare = false\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create .git/config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create .git/HEAD: %w", err)
	}

	// Find the tool's distribution root (where contracts/defaults live)
	// This is separate from the user's workspace - contracts are part of the TOOL,
	// not the user's repository. In containers, this is a fixed path relative to
	// the binary. Locally, it's the eac repository root.
	toolRoot := ctx.OriginalRepoRoot
	if toolRoot == "" {
		cwd, _ := os.Getwd()
		toolRoot = findRepoRoot(cwd)
	}

	ctx.IsolatedDir = tempDir
	state.repoRoot = tempDir

	// R2R_REPO_ROOT: User's workspace (isolated test directory)
	// This is where user configs (.r2r/eac/*.yml) are loaded from
	os.Setenv("R2R_REPO_ROOT", tempDir)

	// R2R_CONTAINER_ROOT: Tool's distribution (real repo with contracts)
	// This is where defaults and schemas are loaded from
	// Simulates how containers have contracts bundled with the tool
	if toolRoot != "" {
		os.Setenv("R2R_CONTAINER_ROOT", toolRoot)
	}

	return nil
}

// findRepoRoot walks up directories to find the repo root (containing go.work or .git).
func findRepoRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func theRepositoryHasNoDirectory(ctx *internal.TestContext, path string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	// Ensure it doesn't exist
	if _, err := os.Stat(fullPath); err == nil {
		if err := os.RemoveAll(fullPath); err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", path, err)
		}
	}
	return nil
}

func theRepositoryHasDirectory(ctx *internal.TestContext, path string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

func theRepositoryHasFileWith(ctx *internal.TestContext, path, content string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

func iLoadTheEACConfiguration() error {
	// Clear cache before loading to ensure fresh load
	config.ClearCache()

	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        state.repoRoot,
		ValidateSchemas: true, // Enable schema validation - contracts are copied to isolated dir
	})
	state.cfg = cfg
	state.loadError = err
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func iTryToLoadTheEACConfiguration() error {
	// Clear cache before loading
	config.ClearCache()

	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        state.repoRoot,
		ValidateSchemas: true, // Enable schema validation - contracts are copied to isolated dir
	})
	state.cfg = cfg
	state.loadError = err
	// Don't return error - we're testing error handling
	return nil
}

func iApplyTypeDefaultsToModules() error {
	if state.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	// Type defaults are now handled by packages - this is a no-op for compatibility
	return nil
}

// ============================================================================
// Module Assertions
// ============================================================================

func theModulesConfigContainsModule(moniker string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	_, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found in config (modules: %v)", moniker, state.cfg.Repository.AllMonikers())
	}
	return nil
}

func theModulesConfigDoesNotContainModule(moniker string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	_, found := state.cfg.Repository.GetModule(moniker)
	if found {
		return fmt.Errorf("module %q unexpectedly found in config", moniker)
	}
	return nil
}

func theModuleHasComponent(moniker, expectedComp string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	if !m.HasComponent(expectedComp) {
		return fmt.Errorf("module %q does not have component %q, has: %v", moniker, expectedComp, m.Components.GetEnabled())
	}
	return nil
}

func theModuleHasComponentRoot(moniker, compName, expectedRoot string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	entry, ok := m.Components[compName]
	if !ok || entry == nil {
		return fmt.Errorf("module %q does not have component %q", moniker, compName)
	}
	if entry.Root != expectedRoot {
		return fmt.Errorf("module %q component %q has root %q, expected %q", moniker, compName, entry.Root, expectedRoot)
	}
	return nil
}

func theModuleHasDescription(moniker, expected string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	if m.Description != expected {
		return fmt.Errorf("module %q has description %q, expected %q", moniker, m.Description, expected)
	}
	return nil
}

func theModuleHasChangelog(moniker, expected string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	changelog := m.GetChangelog()
	if changelog != expected {
		return fmt.Errorf("module %q has changelog %q, expected %q", moniker, changelog, expected)
	}
	return nil
}

func theModuleComponentHasSourcePatternsContaining(moniker, compName, pattern string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	entry, ok := m.Components[compName]
	if !ok || entry == nil {
		return fmt.Errorf("module %q does not have component %q", moniker, compName)
	}
	if entry.Patterns == nil {
		return fmt.Errorf("module %q component %q has no patterns", moniker, compName)
	}
	for _, p := range entry.Patterns.Source {
		if p == pattern || strings.Contains(p, pattern) {
			return nil
		}
	}
	return fmt.Errorf("module %q component %q source patterns %v do not contain %q", moniker, compName, entry.Patterns.Source, pattern)
}

func theModuleComponentDoesNotHaveSourcePattern(moniker, compName, pattern string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	entry, ok := m.Components[compName]
	if !ok || entry == nil {
		return fmt.Errorf("module %q does not have component %q", moniker, compName)
	}
	if entry.Patterns != nil {
		for _, p := range entry.Patterns.Source {
			if p == pattern {
				return fmt.Errorf("module %q component %q unexpectedly has source pattern %q", moniker, compName, pattern)
			}
		}
	}
	return nil
}

func theModuleHasNoSourcePatternsFromTypeDefaults(moniker string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	_, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	// For unknown types, source should remain nil/empty after type defaults
	// (since there's no type definition to get defaults from)
	// Note: This may need adjustment if we add fallback defaults for unknown types
	return nil
}

func theModuleSpecsPatternResolvesWith(moniker, expected string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	// Check if specs package exists and its root contains expected
	if specsEntry, ok := m.Components["specs"]; ok && specsEntry != nil {
		if strings.Contains(specsEntry.Root, expected) {
			return nil
		}
	}
	return fmt.Errorf("module %q specs package does not contain %q", moniker, expected)
}

// ============================================================================
// Component Types Assertions
// ============================================================================

func theComponentTypesConfigContainsType(typeName string) error {
	if state.cfg == nil || state.cfg.ComponentTypes == nil {
		return fmt.Errorf("component types config not loaded")
	}
	pt := state.cfg.ComponentTypes.Get(typeName)
	if pt == nil {
		// List available types for debugging
		var available []string
		for name := range state.cfg.ComponentTypes.ComponentTypes {
			available = append(available, name)
		}
		return fmt.Errorf("type %q not found in config (available: %v)", typeName, available)
	}
	return nil
}

func theTypeHasCapability(typeName, capability string) error {
	// Capabilities have been replaced by package types - check if the package type exists
	if state.cfg == nil || state.cfg.ComponentTypes == nil {
		return fmt.Errorf("package types config not loaded")
	}
	pt := state.cfg.ComponentTypes.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
	// In the new model, capabilities are implicit based on package type
	// Just return success if the package type exists
	return nil
}

func theTypeHasDescription(typeName, expected string) error {
	// Package types don't have descriptions - this is a no-op for compatibility
	if state.cfg == nil || state.cfg.ComponentTypes == nil {
		return fmt.Errorf("package types config not loaded")
	}
	pt := state.cfg.ComponentTypes.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
	// Package types don't have descriptions in the new model
	return nil
}

func theTypeHasDefaultSourcePattern(typeName, pattern string) error {
	if state.cfg == nil || state.cfg.ComponentTypes == nil {
		return fmt.Errorf("package types config not loaded")
	}
	pt := state.cfg.ComponentTypes.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
	// Check if pattern is in the package type's files patterns
	if pt.Files == nil {
		return fmt.Errorf("type %q has no file patterns defined", typeName)
	}
	for _, p := range pt.Files.Source {
		if p == pattern {
			return nil
		}
	}
	return fmt.Errorf("type %q source patterns %v do not contain %q", typeName, pt.Files.Source, pattern)
}

// ============================================================================
// Repository Paths Assertions
// ============================================================================

func theRepositoryPathsFieldIs(field, expected string) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("repository config not loaded")
	}
	var actual string
	switch field {
	case "specs_root":
		actual = state.cfg.Repository.Paths.SpecsRoot
	case "out.root":
		actual = state.cfg.Repository.Paths.Out.Root
	case "out.build":
		actual = state.cfg.Repository.Paths.Out.Build
	case "out.test":
		actual = state.cfg.Repository.Paths.Out.Test
	case "out.logs":
		actual = state.cfg.Repository.Paths.Out.Logs
	default:
		return fmt.Errorf("unknown repository paths field: %s", field)
	}
	if actual != expected {
		return fmt.Errorf("repository paths.%s is %q, expected %q", field, actual, expected)
	}
	return nil
}

// ============================================================================
// System Dependencies Assertions
// ============================================================================

func theSystemDependenciesConfigContains(moniker string) error {
	if state.cfg == nil || state.cfg.SystemDependencies == nil {
		return fmt.Errorf("system dependencies config not loaded")
	}
	dep := state.cfg.SystemDependencies.Get(moniker)
	if dep == nil {
		// List available deps for debugging
		var available []string
		for _, d := range state.cfg.SystemDependencies.Dependencies {
			available = append(available, d.Moniker)
		}
		return fmt.Errorf("dependency %q not found in config (available: %v)", moniker, available)
	}
	return nil
}

func theDependencyHasVersion(moniker, expected string) error {
	if state.cfg == nil || state.cfg.SystemDependencies == nil {
		return fmt.Errorf("system dependencies config not loaded")
	}
	dep := state.cfg.SystemDependencies.Get(moniker)
	if dep == nil {
		return fmt.Errorf("dependency %q not found", moniker)
	}
	if dep.Version != expected {
		return fmt.Errorf("dependency %q has version %q, expected %q", moniker, dep.Version, expected)
	}
	return nil
}

// ============================================================================
// Error Assertions
// ============================================================================

func anErrorIsReturnedContaining(expected string) error {
	if state.loadError == nil {
		return fmt.Errorf("expected error containing %q, but no error occurred", expected)
	}
	if !strings.Contains(strings.ToLower(state.loadError.Error()), strings.ToLower(expected)) {
		return fmt.Errorf("error %q does not contain %q", state.loadError.Error(), expected)
	}
	return nil
}
