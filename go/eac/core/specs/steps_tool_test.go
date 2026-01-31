// Package specs contains godog step implementations for eac-core features.
//
// This file contains step definitions for the tool system tests.
// These tests verify tool resolution - how component types are mapped
// to specific tools for build, test, lint, and scan operations.
package specs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/eac/godog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// toolTestState holds state for tool system test scenarios.
type toolTestState struct {
	cfg               *config.EACConfig
	toolConfig        *tool.ToolConfig
	loadError         error
	repoRoot          string
	origRepoRoot      string // Original R2R_REPO_ROOT value
	origContainerRoot string // Original R2R_CONTAINER_ROOT value
}

var toolState *toolTestState

// registerToolSteps registers step definitions for tool system feature specs.
func registerToolSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	registerToolHooks(sc)
	registerToolBackgroundSteps(sc, ctx)
	registerToolWhenSteps(sc, ctx)
	registerToolThenSteps(sc)
}

func registerToolHooks(sc *godog.ScenarioContext) {
	sc.Before(beforeToolScenario)
	sc.After(afterToolScenario)
}

func beforeToolScenario(c context.Context, _ *godog.Scenario) (context.Context, error) {
	toolState = &toolTestState{
		origRepoRoot:      os.Getenv("R2R_REPO_ROOT"),
		origContainerRoot: os.Getenv("R2R_CONTAINER_ROOT"),
	}
	config.ClearCache()
	return c, nil
}

func afterToolScenario(c context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	cleanupToolTestState()
	return c, nil
}

func registerToolBackgroundSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Hook to perform tool-specific setup after isolation is established
	sc.StepContext().After(func(c context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		if strings.Contains(st.Text, "isolated test repository") && status == godog.StepPassed && ctx.IsolatedDir != "" {
			// Perform tool-specific setup after common isolation
			toolSetupAfterIsolation(ctx)
		}
		return c, nil
	})

	sc.Step(`^I copy the test layout "([^"]*)"$`, func(layoutName string) error {
		return eacgodog.CopyTestLayout(ctx, layoutName, false)
	})
}

func registerToolWhenSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// The "I load the EAC configuration" step is registered once in steps_config_test.go
	// Tool tests use a hook to perform tool-specific config loading after the step runs
	sc.StepContext().After(func(c context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		if strings.Contains(st.Text, "load the EAC configuration") && status == godog.StepPassed {
			// Copy config from cfgState to toolState (config step populates cfgState)
			if cfgState != nil && cfgState.cfg != nil && toolState != nil {
				toolState.cfg = cfgState.cfg
				if toolState.repoRoot == "" {
					toolState.repoRoot = cfgState.repoRoot
				}
			}
			// Load tool config after EAC config is loaded
			if toolState != nil && toolState.repoRoot != "" {
				if loadErr := loadToolToolConfig(); loadErr != nil {
					return c, loadErr
				}
			}
		}
		return c, nil
	})
}

func registerToolThenSteps(sc *godog.ScenarioContext) {
	// Note: "the module ... has component ..." is registered in steps_config_test.go
	sc.Step(`^the configuration has (\d+) modules$`, theConfigurationHasNModules)
	sc.Step(`^the builder for component type "([^"]*)" is "([^"]*)"$`, theBuilderForComponentTypeIs)
	sc.Step(`^the tester for component type "([^"]*)" is "([^"]*)"$`, theTesterForComponentTypeIs)
	sc.Step(`^the component type "([^"]*)" has scanner "([^"]*)"$`, theComponentTypeHasScanner)
	sc.Step(`^the component type "([^"]*)" has extension "([^"]*)"$`, theComponentTypeHasExtension)
}

// cleanupToolTestState cleans up test state after each scenario.
func cleanupToolTestState() {
	if toolState == nil {
		return
	}
	restoreToolEnvVar("R2R_REPO_ROOT", toolState.origRepoRoot)
	restoreToolEnvVar("R2R_CONTAINER_ROOT", toolState.origContainerRoot)
	config.ClearCache()
	toolState = nil
}

// restoreToolEnvVar restores an environment variable to its original value.
func restoreToolEnvVar(key, origValue string) {
	if origValue != "" {
		os.Setenv(key, origValue)
	} else {
		os.Unsetenv(key)
	}
}

// ============================================================================
// Step Implementations
// ============================================================================

// toolSetupAfterIsolation performs tool-specific setup after the common isolation step.
func toolSetupAfterIsolation(ctx *eacgodog.TestContext) {
	if ctx.IsolatedDir == "" {
		return
	}

	toolRoot := ctx.OriginalRepoRoot
	if toolRoot == "" {
		cwd, _ := os.Getwd()
		toolRoot = findToolTestRepoRoot(cwd)
	}

	toolState.repoRoot = ctx.IsolatedDir
	os.Setenv("R2R_REPO_ROOT", ctx.IsolatedDir)
	if toolRoot != "" {
		os.Setenv("R2R_CONTAINER_ROOT", toolRoot)
	}
}

// findToolTestRepoRoot walks up directories to find the repo root.
func findToolTestRepoRoot(start string) string {
	for dir := start; ; dir = filepath.Dir(dir) {
		if isToolTestRepoRoot(dir) {
			return dir
		}
		if filepath.Dir(dir) == dir {
			return ""
		}
	}
}

func isToolTestRepoRoot(dir string) bool {
	markers := []string{"go.work", ".git"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}

func toolILoadTheEACConfiguration() error {
	config.ClearCache()
	if err := loadToolEACConfig(); err != nil {
		return err
	}
	return loadToolToolConfig()
}

func loadToolEACConfig() error {
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        toolState.repoRoot,
		ValidateSchemas: true,
	})
	toolState.cfg = cfg
	toolState.loadError = err
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func loadToolToolConfig() error {
	toolCfg, err := tool.LoadToolConfig(toolState.repoRoot, filepath.Join(toolState.repoRoot, ".r2r", "eac"))
	if err != nil {
		return fmt.Errorf("failed to load tool config: %w", err)
	}
	toolState.toolConfig = toolCfg
	return nil
}

// ============================================================================
// Module Assertions
// ============================================================================

func toolTheModuleHasComponent(moniker, compType string) error {
	m, err := getToolModule(moniker)
	if err != nil {
		return err
	}
	if !m.HasComponent(compType) {
		return fmt.Errorf("module %q does not have component %q, has: %v", moniker, compType, m.Components.GetEnabled())
	}
	return nil
}

func getToolModule(moniker string) (*config.Module, error) {
	if err := requireToolLoadedConfig(); err != nil {
		return nil, err
	}
	m, found := toolState.cfg.Repository.GetModule(moniker)
	if !found {
		return nil, fmt.Errorf("module %q not found", moniker)
	}
	return m, nil
}

func theConfigurationHasNModules(count int) error {
	if err := requireToolLoadedConfig(); err != nil {
		return err
	}
	if actual := len(toolState.cfg.Repository.Modules); actual != count {
		return fmt.Errorf("expected %d modules, got %d", count, actual)
	}
	return nil
}

func requireToolLoadedConfig() error {
	if toolState.cfg == nil || toolState.cfg.Repository == nil {
		return fmt.Errorf("config not loaded")
	}
	return nil
}

// ============================================================================
// Tool Resolution Assertions
// ============================================================================

func getToolComponentTools(compType string) (*tool.ToolAssignment, error) {
	if toolState.toolConfig == nil {
		return nil, fmt.Errorf("tool config not loaded")
	}
	ct, ok := toolState.toolConfig.ComponentTools[compType]
	if !ok || ct == nil {
		return nil, fmt.Errorf("no component tools defined for %q", compType)
	}
	return ct, nil
}

func theBuilderForComponentTypeIs(compType, expected string) error {
	ct, err := getToolComponentTools(compType)
	if err != nil {
		return err
	}
	if ct.Builder != expected {
		return fmt.Errorf("builder for %q is %q, expected %q", compType, ct.Builder, expected)
	}
	return nil
}

func theTesterForComponentTypeIs(compType, expected string) error {
	ct, err := getToolComponentTools(compType)
	if err != nil {
		return err
	}
	if ct.Tester != expected {
		return fmt.Errorf("tester for %q is %q, expected %q", compType, ct.Tester, expected)
	}
	return nil
}

func theComponentTypeHasScanner(compType, scannerID string) error {
	ct, err := getToolComponentTools(compType)
	if err != nil {
		return err
	}
	if !slices.Contains(ct.Scanners, scannerID) {
		return fmt.Errorf("component type %q does not have scanner %q, has: %v", compType, scannerID, ct.Scanners)
	}
	return nil
}

// ============================================================================
// Component Type Assertions
// ============================================================================

func theComponentTypeHasExtension(compType, ext string) error {
	ct, err := getToolComponentType(compType)
	if err != nil {
		return err
	}
	if !slices.Contains(ct.Extensions, ext) {
		return fmt.Errorf("component type %q does not have extension %q, has: %v", compType, ext, ct.Extensions)
	}
	return nil
}

func getToolComponentType(compType string) (*config.ComponentType, error) {
	if toolState.cfg == nil || toolState.cfg.ComponentTypes == nil {
		return nil, fmt.Errorf("component types config not loaded")
	}
	ct := toolState.cfg.ComponentTypes.Get(compType)
	if ct == nil {
		return nil, fmt.Errorf("component type %q not found", compType)
	}
	return ct, nil
}
