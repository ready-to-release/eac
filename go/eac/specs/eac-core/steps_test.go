// Package eaccoretoolsystem contains godog step implementations for specs/eac-core/tool-system.
//
// This package tests the tool resolution system - how component types are mapped
// to specific tools for build, test, lint, and scan operations.
package eaccoretoolsystem

import (
	"fmt"
	"slices"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	eacgodog "github.com/ready-to-release/eac/go/eac/godog"
)

// toolSystemContext holds state for tool system test scenarios.
type toolSystemContext struct {
	sharedCtx  *eacgodog.TestContext
	cfg        *config.EACConfig
	toolConfig *tool.ToolConfig
	loadError  error
}

// RegisterSteps registers step definitions for tool system feature specs.
func RegisterSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	tsCtx := &toolSystemContext{sharedCtx: ctx}

	// Background step - "I am in an isolated test repository" is in common steps
	sc.Step(`^I copy the test layout "([^"]*)"$`, func(layoutName string) error {
		return eacgodog.CopyTestLayout(ctx, layoutName, false)
	})

	// When steps
	sc.Step(`^I load the EAC configuration$`, tsCtx.iLoadTheEACConfiguration)

	// Then steps - module assertions
	sc.Step(`^the module "([^"]*)" has component "([^"]*)"$`, tsCtx.theModuleHasComponent)
	sc.Step(`^the configuration has (\d+) modules$`, tsCtx.theConfigurationHasNModules)

	// Then steps - tool resolution
	sc.Step(`^the builder for component type "([^"]*)" is "([^"]*)"$`, tsCtx.theBuilderForComponentTypeIs)
	sc.Step(`^the tester for component type "([^"]*)" is "([^"]*)"$`, tsCtx.theTesterForComponentTypeIs)
	sc.Step(`^the component type "([^"]*)" has scanner "([^"]*)"$`, tsCtx.theComponentTypeHasScanner)
	sc.Step(`^the component type "([^"]*)" has extension "([^"]*)"$`, tsCtx.theComponentTypeHasExtension)
}

// ============================================================================
// Step Implementations
// ============================================================================

func (c *toolSystemContext) iLoadTheEACConfiguration() error {
	config.ClearCache()

	repoRoot := c.sharedCtx.IsolatedDir
	if repoRoot == "" {
		repoRoot = c.sharedCtx.OriginalRepoRoot
	}

	// Load EAC config
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        repoRoot,
		ValidateSchemas: true,
	})
	c.cfg = cfg
	c.loadError = err
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load tool config
	toolCfg, err := tool.LoadToolConfig(repoRoot, "")
	if err != nil {
		return fmt.Errorf("failed to load tool config: %w", err)
	}
	c.toolConfig = toolCfg

	return nil
}

// ============================================================================
// Module Assertions
// ============================================================================

func (c *toolSystemContext) theModuleHasComponent(moniker, compType string) error {
	m, err := c.getModule(moniker)
	if err != nil {
		return err
	}
	if !m.HasComponent(compType) {
		return fmt.Errorf("module %q does not have component %q, has: %v", moniker, compType, m.Components.GetEnabled())
	}
	return nil
}

func (c *toolSystemContext) getModule(moniker string) (*config.Module, error) {
	if err := c.requireLoadedConfig(); err != nil {
		return nil, err
	}
	m, found := c.cfg.Repository.GetModule(moniker)
	if !found {
		return nil, fmt.Errorf("module %q not found", moniker)
	}
	return m, nil
}

func (c *toolSystemContext) theConfigurationHasNModules(count int) error {
	if err := c.requireLoadedConfig(); err != nil {
		return err
	}
	if actual := len(c.cfg.Repository.Modules); actual != count {
		return fmt.Errorf("expected %d modules, got %d", count, actual)
	}
	return nil
}

func (c *toolSystemContext) requireLoadedConfig() error {
	if c.cfg == nil || c.cfg.Repository == nil {
		return fmt.Errorf("config not loaded")
	}
	return nil
}

// ============================================================================
// Tool Resolution Assertions
// ============================================================================

func (c *toolSystemContext) getComponentTools(compType string) (*tool.ToolAssignment, error) {
	if c.toolConfig == nil {
		return nil, fmt.Errorf("tool config not loaded")
	}
	ct, ok := c.toolConfig.ComponentTools[compType]
	if !ok || ct == nil {
		return nil, fmt.Errorf("no component tools defined for %q", compType)
	}
	return ct, nil
}

func (c *toolSystemContext) theBuilderForComponentTypeIs(compType, expected string) error {
	ct, err := c.getComponentTools(compType)
	if err != nil {
		return err
	}
	if ct.Builder != expected {
		return fmt.Errorf("builder for %q is %q, expected %q", compType, ct.Builder, expected)
	}
	return nil
}

func (c *toolSystemContext) theTesterForComponentTypeIs(compType, expected string) error {
	ct, err := c.getComponentTools(compType)
	if err != nil {
		return err
	}
	if ct.Tester != expected {
		return fmt.Errorf("tester for %q is %q, expected %q", compType, ct.Tester, expected)
	}
	return nil
}

func (c *toolSystemContext) theComponentTypeHasScanner(compType, scannerID string) error {
	ct, err := c.getComponentTools(compType)
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

func (c *toolSystemContext) theComponentTypeHasExtension(compType, ext string) error {
	ct, err := c.getComponentType(compType)
	if err != nil {
		return err
	}
	if !slices.Contains(ct.Extensions, ext) {
		return fmt.Errorf("component type %q does not have extension %q, has: %v", compType, ext, ct.Extensions)
	}
	return nil
}

func (c *toolSystemContext) getComponentType(compType string) (*config.ComponentType, error) {
	if c.cfg == nil || c.cfg.ComponentTypes == nil {
		return nil, fmt.Errorf("component types config not loaded")
	}
	ct := c.cfg.ComponentTypes.Get(compType)
	if ct == nil {
		return nil, fmt.Errorf("component type %q not found", compType)
	}
	return ct, nil
}
