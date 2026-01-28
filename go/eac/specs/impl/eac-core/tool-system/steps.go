// Package toolsystem contains godog step implementations for specs/eac-core/tool-system.
//
// This package tests the tool resolution system - how component types are mapped
// to specific tools for build, test, lint, and scan operations.
package toolsystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// testState holds state for tool system test scenarios.
type testState struct {
	cfg               *config.EACConfig
	toolConfig        *tool.ToolConfig
	loadError         error
	repoRoot          string
	origRepoRoot      string // Original R2R_REPO_ROOT value
	origContainerRoot string // Original R2R_CONTAINER_ROOT value
}

var state *testState

// RegisterSteps registers step definitions for tool system feature specs.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	registerHooks(sc)
	registerBackgroundSteps(sc, ctx)
	registerWhenSteps(sc)
	registerThenSteps(sc)
}

func registerHooks(sc *godog.ScenarioContext) {
	sc.Before(beforeScenario)
	sc.After(afterScenario)
}

func beforeScenario(c context.Context, _ *godog.Scenario) (context.Context, error) {
	state = &testState{
		origRepoRoot:      os.Getenv("R2R_REPO_ROOT"),
		origContainerRoot: os.Getenv("R2R_CONTAINER_ROOT"),
	}
	config.ClearCache()
	return c, nil
}

func afterScenario(c context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	cleanupTestState()
	return c, nil
}

func registerBackgroundSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	sc.Step(`^I am in an isolated test repository$`, func() error {
		return iAmInAnIsolatedTestRepository(ctx)
	})
	sc.Step(`^I copy the test layout "([^"]*)"$`, func(layoutName string) error {
		return internal.CopyTestLayout(ctx, layoutName, false)
	})
}

func registerWhenSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I load the EAC configuration$`, iLoadTheEACConfiguration)
}

func registerThenSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the module "([^"]*)" has component "([^"]*)"$`, theModuleHasComponent)
	sc.Step(`^the configuration has (\d+) modules$`, theConfigurationHasNModules)
	sc.Step(`^the builder for component type "([^"]*)" is "([^"]*)"$`, theBuilderForComponentTypeIs)
	sc.Step(`^the tester for component type "([^"]*)" is "([^"]*)"$`, theTesterForComponentTypeIs)
	sc.Step(`^the component type "([^"]*)" has scanner "([^"]*)"$`, theComponentTypeHasScanner)
	sc.Step(`^the component type "([^"]*)" has extension "([^"]*)"$`, theComponentTypeHasExtension)
}

// cleanupTestState cleans up test state after each scenario.
func cleanupTestState() {
	if state == nil {
		return
	}
	restoreEnvVar("R2R_REPO_ROOT", state.origRepoRoot)
	restoreEnvVar("R2R_CONTAINER_ROOT", state.origContainerRoot)
	config.ClearCache()
	state = nil
}

// restoreEnvVar restores an environment variable to its original value.
func restoreEnvVar(key, origValue string) {
	if origValue != "" {
		os.Setenv(key, origValue)
	} else {
		os.Unsetenv(key)
	}
}

// ============================================================================
// Step Implementations
// ============================================================================

func iAmInAnIsolatedTestRepository(ctx *internal.TestContext) error {
	tempDir, err := createIsolatedGitRepo()
	if err != nil {
		return err
	}
	setupTestEnvironment(ctx, tempDir, resolveToolRoot(ctx))
	return nil
}

func setupTestEnvironment(ctx *internal.TestContext, tempDir, toolRoot string) {
	ctx.IsolatedDir = tempDir
	state.repoRoot = tempDir
	os.Setenv("R2R_REPO_ROOT", tempDir)
	if toolRoot != "" {
		os.Setenv("R2R_CONTAINER_ROOT", toolRoot)
	}
}

func createIsolatedGitRepo() (string, error) {
	tempDir, err := os.MkdirTemp("", "tool-system-test-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	if err := initMinimalGitDir(tempDir); err != nil {
		return "", err
	}
	return tempDir, nil
}

func initMinimalGitDir(repoDir string) error {
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .git: %w", err)
	}
	return writeGitFiles(gitDir)
}

func writeGitFiles(gitDir string) error {
	if err := writeGitFile(gitDir, "config", "[core]\n\tbare = false\n"); err != nil {
		return err
	}
	return writeGitFile(gitDir, "HEAD", "ref: refs/heads/main\n")
}

func writeGitFile(gitDir, name, content string) error {
	if err := os.WriteFile(filepath.Join(gitDir, name), []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create .git/%s: %w", name, err)
	}
	return nil
}

func resolveToolRoot(ctx *internal.TestContext) string {
	if ctx.OriginalRepoRoot != "" {
		return ctx.OriginalRepoRoot
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	toolRoot := findRepoRoot(cwd)
	ctx.OriginalRepoRoot = toolRoot
	return toolRoot
}

// findRepoRoot walks up directories to find the repo root.
func findRepoRoot(start string) string {
	for dir := start; ; dir = filepath.Dir(dir) {
		if isRepoRoot(dir) {
			return dir
		}
		if filepath.Dir(dir) == dir {
			return ""
		}
	}
}

func isRepoRoot(dir string) bool {
	markers := []string{"go.work", ".git"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}

func iLoadTheEACConfiguration() error {
	config.ClearCache()
	if err := loadEACConfig(); err != nil {
		return err
	}
	return loadToolConfig()
}

func loadEACConfig() error {
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        state.repoRoot,
		ValidateSchemas: true,
	})
	state.cfg = cfg
	state.loadError = err
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func loadToolConfig() error {
	toolCfg, err := tool.LoadToolConfig(state.repoRoot, filepath.Join(state.repoRoot, ".r2r", "eac"))
	if err != nil {
		return fmt.Errorf("failed to load tool config: %w", err)
	}
	state.toolConfig = toolCfg
	return nil
}

// ============================================================================
// Module Assertions
// ============================================================================

func theModuleHasComponent(moniker, compType string) error {
	m, err := getModule(moniker)
	if err != nil {
		return err
	}
	if !m.HasComponent(compType) {
		return fmt.Errorf("module %q does not have component %q, has: %v", moniker, compType, m.Components.GetEnabled())
	}
	return nil
}

func getModule(moniker string) (*config.Module, error) {
	if err := requireLoadedConfig(); err != nil {
		return nil, err
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return nil, fmt.Errorf("module %q not found", moniker)
	}
	return m, nil
}

func theConfigurationHasNModules(count int) error {
	if err := requireLoadedConfig(); err != nil {
		return err
	}
	if actual := len(state.cfg.Repository.Modules); actual != count {
		return fmt.Errorf("expected %d modules, got %d", count, actual)
	}
	return nil
}

func requireLoadedConfig() error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("config not loaded")
	}
	return nil
}

// ============================================================================
// Tool Resolution Assertions
// ============================================================================

func getComponentTools(compType string) (*tool.ToolAssignment, error) {
	if state.toolConfig == nil {
		return nil, fmt.Errorf("tool config not loaded")
	}
	ct, ok := state.toolConfig.ComponentTools[compType]
	if !ok || ct == nil {
		return nil, fmt.Errorf("no component tools defined for %q", compType)
	}
	return ct, nil
}

func theBuilderForComponentTypeIs(compType, expected string) error {
	ct, err := getComponentTools(compType)
	if err != nil {
		return err
	}
	if ct.Builder != expected {
		return fmt.Errorf("builder for %q is %q, expected %q", compType, ct.Builder, expected)
	}
	return nil
}

func theTesterForComponentTypeIs(compType, expected string) error {
	ct, err := getComponentTools(compType)
	if err != nil {
		return err
	}
	if ct.Tester != expected {
		return fmt.Errorf("tester for %q is %q, expected %q", compType, ct.Tester, expected)
	}
	return nil
}

func theComponentTypeHasScanner(compType, scannerID string) error {
	ct, err := getComponentTools(compType)
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
	ct, err := getComponentType(compType)
	if err != nil {
		return err
	}
	if !slices.Contains(ct.Extensions, ext) {
		return fmt.Errorf("component type %q does not have extension %q, has: %v", compType, ext, ct.Extensions)
	}
	return nil
}

func getComponentType(compType string) (*config.ComponentType, error) {
	if state.cfg == nil || state.cfg.ComponentTypes == nil {
		return nil, fmt.Errorf("component types config not loaded")
	}
	ct := state.cfg.ComponentTypes.Get(compType)
	if ct == nil {
		return nil, fmt.Errorf("component type %q not found", compType)
	}
	return ct, nil
}
