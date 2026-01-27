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

	// Given steps - copy test layouts
	sc.Step(`^I copy the test layout "([^"]*)"$`, func(layoutName string) error {
		return iCopyTheTestLayout(ctx, layoutName)
	})

	// When steps - loading config
	sc.Step(`^I load the EAC configuration$`, func() error {
		return iLoadTheEACConfiguration()
	})

	// Then steps - module assertions
	sc.Step(`^the module "([^"]*)" has component "([^"]*)"$`, func(moniker, compType string) error {
		return theModuleHasComponent(moniker, compType)
	})
	sc.Step(`^the configuration has (\d+) modules$`, func(count int) error {
		return theConfigurationHasNModules(count)
	})

	// Then steps - tool resolution assertions
	sc.Step(`^the builder for component type "([^"]*)" is "([^"]*)"$`, func(compType, toolID string) error {
		return theBuilderForComponentTypeIs(compType, toolID)
	})
	sc.Step(`^the tester for component type "([^"]*)" is "([^"]*)"$`, func(compType, toolID string) error {
		return theTesterForComponentTypeIs(compType, toolID)
	})
	sc.Step(`^the component type "([^"]*)" has scanner "([^"]*)"$`, func(compType, scannerID string) error {
		return theComponentTypeHasScanner(compType, scannerID)
	})

	// Then steps - component type assertions
	sc.Step(`^the component type "([^"]*)" has extension "([^"]*)"$`, func(compType, ext string) error {
		return theComponentTypeHasExtension(compType, ext)
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
	tempDir, err := os.MkdirTemp("", "tool-system-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create minimal .git directory so config loader can find repo root
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .git: %w", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("[core]\n\tbare = false\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create .git/config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		return fmt.Errorf("failed to create .git/HEAD: %w", err)
	}

	// Find the tool's distribution root (where contracts/defaults live)
	toolRoot := ctx.OriginalRepoRoot
	if toolRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		toolRoot = findRepoRoot(cwd)
		// Set OriginalRepoRoot so CopyTestLayout can find templates
		ctx.OriginalRepoRoot = toolRoot
	}

	ctx.IsolatedDir = tempDir
	state.repoRoot = tempDir

	// R2R_REPO_ROOT: User's workspace (isolated test directory)
	os.Setenv("R2R_REPO_ROOT", tempDir)

	// R2R_CONTAINER_ROOT: Tool's distribution (real repo with contracts)
	if toolRoot != "" {
		os.Setenv("R2R_CONTAINER_ROOT", toolRoot)
	}

	return nil
}

// findRepoRoot walks up directories to find the repo root.
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

func iCopyTheTestLayout(ctx *internal.TestContext, layoutName string) error {
	return internal.CopyTestLayout(ctx, layoutName, false)
}

func iLoadTheEACConfiguration() error {
	// Clear cache before loading to ensure fresh load
	config.ClearCache()

	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        state.repoRoot,
		ValidateSchemas: true,
	})
	state.cfg = cfg
	state.loadError = err
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load tool configuration
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
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("config not loaded")
	}
	m, found := state.cfg.Repository.GetModule(moniker)
	if !found {
		return fmt.Errorf("module %q not found", moniker)
	}
	if !m.HasComponent(compType) {
		return fmt.Errorf("module %q does not have component %q, has: %v", moniker, compType, m.Components.GetEnabled())
	}
	return nil
}

func theConfigurationHasNModules(count int) error {
	if state.cfg == nil || state.cfg.Repository == nil {
		return fmt.Errorf("config not loaded")
	}
	actual := len(state.cfg.Repository.Modules)
	if actual != count {
		return fmt.Errorf("expected %d modules, got %d", count, actual)
	}
	return nil
}

// ============================================================================
// Tool Resolution Assertions
// ============================================================================

func theBuilderForComponentTypeIs(compType, expectedToolID string) error {
	if state.toolConfig == nil {
		return fmt.Errorf("tool config not loaded")
	}
	componentTools, ok := state.toolConfig.ComponentTools[compType]
	if !ok || componentTools == nil {
		return fmt.Errorf("no component tools defined for %q", compType)
	}
	if componentTools.Builder != expectedToolID {
		return fmt.Errorf("builder for %q is %q, expected %q", compType, componentTools.Builder, expectedToolID)
	}
	return nil
}

func theTesterForComponentTypeIs(compType, expectedToolID string) error {
	if state.toolConfig == nil {
		return fmt.Errorf("tool config not loaded")
	}
	componentTools, ok := state.toolConfig.ComponentTools[compType]
	if !ok || componentTools == nil {
		return fmt.Errorf("no component tools defined for %q", compType)
	}
	if componentTools.Tester != expectedToolID {
		return fmt.Errorf("tester for %q is %q, expected %q", compType, componentTools.Tester, expectedToolID)
	}
	return nil
}

func theComponentTypeHasScanner(compType, scannerID string) error {
	if state.toolConfig == nil {
		return fmt.Errorf("tool config not loaded")
	}
	componentTools, ok := state.toolConfig.ComponentTools[compType]
	if !ok || componentTools == nil {
		return fmt.Errorf("no component tools defined for %q", compType)
	}
	for _, s := range componentTools.Scanners {
		if s == scannerID {
			return nil
		}
	}
	return fmt.Errorf("component type %q does not have scanner %q, has: %v", compType, scannerID, componentTools.Scanners)
}

// ============================================================================
// Component Type Assertions
// ============================================================================

func theComponentTypeHasExtension(compType, ext string) error {
	if state.cfg == nil || state.cfg.ComponentTypes == nil {
		return fmt.Errorf("component types config not loaded")
	}
	ct := state.cfg.ComponentTypes.Get(compType)
	if ct == nil {
		return fmt.Errorf("component type %q not found", compType)
	}
	for _, e := range ct.Extensions {
		if e == ext {
			return nil
		}
	}
	return fmt.Errorf("component type %q does not have extension %q, has: %v", compType, ext, ct.Extensions)
}
