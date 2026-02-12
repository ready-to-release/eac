// Package specs contains godog step implementations for core features.
//
// This file contains step definitions for the config defaults tests.
// These tests verify configuration loading and defaults system.
package specs

import (
	"github.com/ready-to-release/eac/go/core/environments"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// configTestState holds state for config defaults test scenarios.
type configTestState struct {
	cfg               *config.EACConfig
	loadError         error
	repoRoot          string
	origRepoRoot      string // Original CLIE_REPO_ROOT value
	origContainerRoot string // Original CLIE_CONTAINER_ROOT value
}

// cfgState is per-scenario state, reset in registerConfigSteps Before/After hooks.
// Godog runs scenarios sequentially, so a package-level var is safe here.
var cfgState *configTestState

// registerConfigSteps registers step definitions for config defaults feature specs.
func registerConfigSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Setup/teardown
	sc.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		cfgState = &configTestState{}
		// Save original env vars and clear config cache
		cfgState.origRepoRoot = os.Getenv(environments.EnvCLIERepoRoot)
		cfgState.origContainerRoot = os.Getenv(environments.EnvCLIEContainerRoot)
		config.ClearCache()
		return c, nil
	})

	sc.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cleanupConfigTestState()
		return c, nil
	})

	// Hook to perform config-specific setup after isolation is established
	sc.StepContext().After(func(c context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		if strings.Contains(st.Text, "isolated test repository") && status == godog.StepPassed && ctx.IsolatedDir != "" {
			// Perform config-specific setup after common isolation
			configSetupAfterIsolation(ctx)
		}
		return c, nil
	})

	// Given steps - repository state
	sc.Step(`^the repository has no "([^"]*)" directory$`, func(path string) error {
		return configTheRepositoryHasNoDirectory(ctx, path)
	})
	sc.Step(`^the repository has directory "([^"]*)"$`, func(path string) error {
		return configTheRepositoryHasDirectory(ctx, path)
	})
	sc.Step(`^the repository has file "([^"]*)" with:$`, func(path string, content *godog.DocString) error {
		return configTheRepositoryHasFileWith(ctx, path, content.Content)
	})
	sc.Step(`^the contracts directory does not exist$`, func() error {
		os.Unsetenv(environments.EnvCLIEContainerRoot)
		return nil
	})

	// When steps - loading config
	sc.Step(`^I load the EAC configuration$`, configILoadTheEACConfiguration)
	sc.Step(`^I try to load the EAC configuration$`, configITryToLoadTheEACConfiguration)
	sc.Step(`^I apply type defaults to modules$`, configIApplyTypeDefaultsToModules)

	// Then steps - modules assertions
	sc.Step(`^the modules config contains module "([^"]*)"$`, configTheModulesConfigContainsModule)
	sc.Step(`^the modules config does not contain module "([^"]*)"$`, configTheModulesConfigDoesNotContainModule)
	sc.Step(`^the module "([^"]*)" has component "([^"]*)"$`, configTheModuleHasComponent)
	sc.Step(`^the module "([^"]*)" has component root "([^"]*)" as "([^"]*)"$`, configTheModuleHasComponentRoot)
	sc.Step(`^the module "([^"]*)" has description "([^"]*)"$`, configTheModuleHasDescription)
	sc.Step(`^the module "([^"]*)" has changelog "([^"]*)"$`, configTheModuleHasChangelog)
	sc.Step(`^the module "([^"]*)" component "([^"]*)" has source patterns containing "([^"]*)"$`, configTheModuleComponentHasSourcePatternsContaining)
	sc.Step(`^the module "([^"]*)" component "([^"]*)" does not have source pattern "([^"]*)"$`, configTheModuleComponentDoesNotHaveSourcePattern)
	sc.Step(`^the module "([^"]*)" has no source patterns from type defaults$`, configTheModuleHasNoSourcePatternsFromTypeDefaults)
	sc.Step(`^the module "([^"]*)" specs pattern resolves with "([^"]*)"$`, configTheModuleSpecsPatternResolvesWith)
	sc.Step(`^the module "([^"]*)" component "([^"]*)" has root "([^"]*)"$`, configTheModuleComponentHasRoot)

	// Then steps - component types assertions
	sc.Step(`^the component types config contains type "([^"]*)"$`, configTheComponentKindsConfigContainsType)
	sc.Step(`^the type "([^"]*)" has capability "([^"]*)"$`, configTheTypeHasCapability)
	sc.Step(`^the type "([^"]*)" has description "([^"]*)"$`, configTheTypeHasDescription)
	sc.Step(`^the type "([^"]*)" has default source pattern "([^"]*)"$`, configTheTypeHasDefaultSourcePattern)
	sc.Step(`^the type "([^"]*)" has builder "([^"]*)"$`, configTheTypeHasBuilder)
	sc.Step(`^the type "([^"]*)" has extension "([^"]*)"$`, configTheTypeHasExtension)

	// Then steps - repository paths assertions
	sc.Step(`^the repository paths\.specs_root is "([^"]*)"$`, func(expected string) error {
		return configTheRepositoryPathsFieldIs("specs_root", expected)
	})
	sc.Step(`^the repository paths\.out\.root is "([^"]*)"$`, func(expected string) error {
		return configTheRepositoryPathsFieldIs("out.root", expected)
	})
	sc.Step(`^the repository paths\.out\.build is "([^"]*)"$`, func(expected string) error {
		return configTheRepositoryPathsFieldIs("out.build", expected)
	})
	sc.Step(`^the repository paths\.out\.test is "([^"]*)"$`, func(expected string) error {
		return configTheRepositoryPathsFieldIs("out.test", expected)
	})
	sc.Step(`^the repository paths\.out\.logs is "([^"]*)"$`, func(expected string) error {
		return configTheRepositoryPathsFieldIs("out.logs", expected)
	})

	// Then steps - system dependencies assertions
	sc.Step(`^the system dependencies config contains "([^"]*)"$`, configTheSystemDependenciesConfigContains)
	sc.Step(`^the dependency "([^"]*)" has version "([^"]*)"$`, configTheDependencyHasVersion)

	// Then steps - error assertions
	sc.Step(`^an error is returned containing "([^"]*)"$`, configAnErrorIsReturnedContaining)
}

// cleanupConfigTestState cleans up test state after each scenario.
func cleanupConfigTestState() {
	if cfgState != nil {
		if cfgState.origRepoRoot != "" {
			os.Setenv(environments.EnvCLIERepoRoot, cfgState.origRepoRoot)
		} else {
			os.Unsetenv(environments.EnvCLIERepoRoot)
		}
		if cfgState.origContainerRoot != "" {
			os.Setenv(environments.EnvCLIEContainerRoot, cfgState.origContainerRoot)
		} else {
			os.Unsetenv(environments.EnvCLIEContainerRoot)
		}
		config.ClearCache()
	}
	cfgState = nil
}

// ============================================================================
// Step Implementations
// ============================================================================

// configSetupAfterIsolation performs config-specific setup after the common isolation step.
func configSetupAfterIsolation(ctx *eacgodog.TestContext) {
	if ctx.IsolatedDir == "" {
		return
	}

	toolRoot := ctx.OriginalRepoRoot
	if toolRoot == "" {
		cwd, _ := os.Getwd()
		toolRoot = findConfigRepoRoot(cwd)
	}

	cfgState.repoRoot = ctx.IsolatedDir

	os.Setenv(environments.EnvCLIERepoRoot, ctx.IsolatedDir)
	if toolRoot != "" {
		os.Setenv(environments.EnvCLIEContainerRoot, toolRoot)
	}
}

func findConfigRepoRoot(start string) string {
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

func configTheRepositoryHasNoDirectory(ctx *eacgodog.TestContext, path string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if _, err := os.Stat(fullPath); err == nil {
		if err := os.RemoveAll(fullPath); err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", path, err)
		}
	}
	return nil
}

func configTheRepositoryHasDirectory(ctx *eacgodog.TestContext, path string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if err := os.MkdirAll(fullPath, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

func configTheRepositoryHasFileWith(ctx *eacgodog.TestContext, path, content string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

func configILoadTheEACConfiguration() error {
	config.ClearCache()
	repoRoot := cfgState.repoRoot
	if repoRoot == "" {
		// Fall back to finding the repo root from cwd
		cwd, _ := os.Getwd()
		repoRoot = findConfigRepoRoot(cwd)
		cfgState.repoRoot = repoRoot
	}
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        repoRoot,
		ValidateSchemas: true,
	})
	cfgState.cfg = cfg
	cfgState.loadError = err
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	_ = tool.InitializeGlobalBridges(repoRoot, cfg.ConfigRoot)
	return nil
}

func configITryToLoadTheEACConfiguration() error {
	config.ClearCache()
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        cfgState.repoRoot,
		ValidateSchemas: true,
	})
	cfgState.cfg = cfg
	cfgState.loadError = err
	return nil
}

func configIApplyTypeDefaultsToModules() error {
	if cfgState.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	return nil
}

// ============================================================================
// Module Assertions
// ============================================================================

func configTheModulesConfigContainsModule(moniker string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	_, found := cfgState.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found in config (modules: %v)", moniker, cfgState.cfg.Repository.AllMonikers())
	}
	return nil
}

func configTheModulesConfigDoesNotContainModule(moniker string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	_, found := cfgState.cfg.Repository.GetModule(moniker)
	if found {
		return fmt.Errorf("module %q unexpectedly found in config", moniker)
	}
	return nil
}

func configTheModuleHasComponent(moniker, expectedComp string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := cfgState.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	if !m.HasComponent(expectedComp) {
		return fmt.Errorf("module %q does not have component %q, has: %v", moniker, expectedComp, m.Components.GetEnabled())
	}
	return nil
}

func configTheModuleHasComponentRoot(moniker, compName, expectedRoot string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := cfgState.cfg.Repository.GetModule(moniker)
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

func configTheModuleHasDescription(moniker, expected string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := cfgState.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	if m.Description != expected {
		return fmt.Errorf("module %q has description %q, expected %q", moniker, m.Description, expected)
	}
	return nil
}

func configTheModuleHasChangelog(moniker, expected string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := cfgState.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	changelog := m.GetChangelog()
	if changelog != expected {
		return fmt.Errorf("module %q has changelog %q, expected %q", moniker, changelog, expected)
	}
	return nil
}

func configTheModuleComponentHasSourcePatternsContaining(moniker, compName, pattern string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := cfgState.cfg.Repository.GetModule(moniker)
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

func configTheModuleComponentDoesNotHaveSourcePattern(moniker, compName, pattern string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := cfgState.cfg.Repository.GetModule(moniker)
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

func configTheModuleHasNoSourcePatternsFromTypeDefaults(moniker string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	_, found := cfgState.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	return nil
}

func configTheModuleSpecsPatternResolvesWith(moniker, expected string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("modules config not loaded")
	}
	m, found := cfgState.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	if specsEntry, ok := m.Components["specs"]; ok && specsEntry != nil {
		if strings.Contains(specsEntry.Root, expected) {
			return nil
		}
	}
	return fmt.Errorf("module %q specs package does not contain %q", moniker, expected)
}

func configTheModuleComponentHasRoot(moniker, comp, expectedRoot string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("config not loaded")
	}
	module, found := cfgState.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	entry, ok := module.Components[comp]
	if !ok || entry == nil {
		return fmt.Errorf("module %q does not have component %q", moniker, comp)
	}
	if entry.Root != expectedRoot {
		return fmt.Errorf("module %q component %q has root %q, expected %q", moniker, comp, entry.Root, expectedRoot)
	}
	return nil
}

// ============================================================================
// Component Types Assertions
// ============================================================================

func configTheComponentKindsConfigContainsType(typeName string) error {
	if cfgState.cfg == nil || cfgState.cfg.ComponentKinds == nil {
		return fmt.Errorf("component types config not loaded")
	}
	pt := cfgState.cfg.ComponentKinds.Get(typeName)
	if pt == nil {
		var available []string
		for name := range cfgState.cfg.ComponentKinds.Kinds {
			available = append(available, name)
		}
		return fmt.Errorf("type %q not found in config (available: %v)", typeName, available)
	}
	return nil
}

func configTheTypeHasCapability(typeName, capability string) error {
	if cfgState.cfg == nil || cfgState.cfg.ComponentKinds == nil {
		return fmt.Errorf("package types config not loaded")
	}
	pt := cfgState.cfg.ComponentKinds.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
	return nil
}

func configTheTypeHasDescription(typeName, expected string) error {
	if cfgState.cfg == nil || cfgState.cfg.ComponentKinds == nil {
		return fmt.Errorf("package types config not loaded")
	}
	pt := cfgState.cfg.ComponentKinds.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
	return nil
}

func configTheTypeHasDefaultSourcePattern(typeName, pattern string) error {
	if cfgState.cfg == nil || cfgState.cfg.ComponentKinds == nil {
		return fmt.Errorf("package types config not loaded")
	}
	pt := cfgState.cfg.ComponentKinds.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
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

func configTheTypeHasBuilder(typeName, builder string) error {
	if cfgState.cfg == nil || cfgState.cfg.ComponentKinds == nil {
		return fmt.Errorf("component types config not loaded")
	}
	pt := cfgState.cfg.ComponentKinds.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
	builders := pt.GetBuilders()
	actualBuilder := ""
	if len(builders) > 0 {
		actualBuilder = builders[0]
	}
	if actualBuilder != builder {
		return fmt.Errorf("type %q has builder %q, expected %q", typeName, actualBuilder, builder)
	}
	return nil
}

func configTheTypeHasExtension(typeName, ext string) error {
	if cfgState.cfg == nil || cfgState.cfg.ComponentKinds == nil {
		return fmt.Errorf("component types config not loaded")
	}
	pt := cfgState.cfg.ComponentKinds.Get(typeName)
	if pt == nil {
		return fmt.Errorf("type %q not found", typeName)
	}
	for _, e := range pt.Extensions {
		if e == ext {
			return nil
		}
	}
	return fmt.Errorf("type %q extensions %v do not contain %q", typeName, pt.Extensions, ext)
}

// ============================================================================
// Repository Paths Assertions
// ============================================================================

func configTheRepositoryPathsFieldIs(field, expected string) error {
	if cfgState.cfg == nil || cfgState.cfg.Repository == nil {
		return fmt.Errorf("repository config not loaded")
	}
	var actual string
	switch field {
	case "specs_root":
		actual = cfgState.cfg.Repository.Paths.SpecsRoot
	case "out.root":
		actual = cfgState.cfg.Repository.Paths.Out.Root
	case "out.build":
		actual = cfgState.cfg.Repository.Paths.Out.Build
	case "out.test":
		actual = cfgState.cfg.Repository.Paths.Out.Test
	case "out.logs":
		actual = cfgState.cfg.Repository.Paths.Out.Logs
	default:
		return fmt.Errorf("unknown repository paths field: %s", field)
	}
	if actual != expected {
		return fmt.Errorf("repository paths.%s is %q, expected %q", field, actual, expected)
	}
	return nil
}

// ============================================================================
// Tool Registry Assertions
// ============================================================================

func configTheSystemDependenciesConfigContains(moniker string) error {
	registry := tool.GlobalRegistry()
	if !registry.Has(moniker) {
		var available []string
		for toolID := range registry.GetAll() {
			available = append(available, toolID)
		}
		return fmt.Errorf("tool %q not found in registry (available: %v)", moniker, available)
	}
	return nil
}

func configTheDependencyHasVersion(moniker, expected string) error {
	registry := tool.GlobalRegistry()
	toolDef, ok := registry.Get(moniker)
	if !ok {
		return fmt.Errorf("tool %q not found in registry", moniker)
	}
	if toolDef.Version != expected {
		return fmt.Errorf("tool %q has version %q, expected %q", moniker, toolDef.Version, expected)
	}
	return nil
}

// ============================================================================
// Error Assertions
// ============================================================================

func configAnErrorIsReturnedContaining(expected string) error {
	if cfgState.loadError == nil {
		return fmt.Errorf("expected error containing %q, but no error occurred", expected)
	}
	if !strings.Contains(strings.ToLower(cfgState.loadError.Error()), strings.ToLower(expected)) {
		return fmt.Errorf("error %q does not contain %q", cfgState.loadError.Error(), expected)
	}
	return nil
}
