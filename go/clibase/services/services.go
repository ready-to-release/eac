// Package services provides centralized service initialization for EAC commands.
// It implements the SimpleServicesPort interface from the core contracts.
package services

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workspace"
)

// ============================================================================
// Config Adapter - wraps *config.EACConfig to implement ConfigPort
// ============================================================================

// configAdapter wraps *config.EACConfig to implement interfaces.ConfigPort.
type configAdapter struct {
	cfg *config.EACConfig
}

// Compile-time check
var _ interfaces.ConfigPort = (*configAdapter)(nil)

func (a *configAdapter) GetRepoRoot() string   { return a.cfg.RepoRoot }
func (a *configAdapter) GetConfigRoot() string { return a.cfg.ConfigRoot }

func (a *configAdapter) GetRepository() interfaces.RepositoryConfigPort {
	if a.cfg.Repository == nil {
		return nil
	}
	return &repositoryAdapter{cfg: a.cfg}
}

func (a *configAdapter) GetEnvironments() interfaces.EnvironmentsConfigPort {
	if a.cfg.Environments == nil {
		return nil
	}
	return &environmentsAdapter{envs: a.cfg.Environments}
}

func (a *configAdapter) GetTestingTags() interfaces.TestingTagsConfigPort {
	if a.cfg.TestingTags == nil {
		return nil
	}
	return &testingTagsAdapter{tags: a.cfg.TestingTags}
}

func (a *configAdapter) GetTestSuites() interfaces.TestSuitesConfigPort {
	if a.cfg.TestSuites == nil {
		return nil
	}
	return &testSuitesAdapter{suites: a.cfg.TestSuites}
}

func (a *configAdapter) GetComponentTypes() interfaces.ComponentTypesConfigPort {
	if a.cfg.ComponentTypes == nil {
		return nil
	}
	return &componentTypesAdapter{types: a.cfg.ComponentTypes}
}

// repositoryAdapter wraps config.RepositoryConfig
type repositoryAdapter struct {
	cfg *config.EACConfig
}

func (a *repositoryAdapter) AllMonikers() []string {
	monikers := make([]string, len(a.cfg.Repository.Modules))
	for i := range a.cfg.Repository.Modules {
		monikers[i] = a.cfg.Repository.Modules[i].Moniker
	}
	return monikers
}

func (a *repositoryAdapter) GetModule(moniker string) (interfaces.ModuleContractPort, bool) {
	for i := range a.cfg.Repository.Modules {
		if a.cfg.Repository.Modules[i].Moniker == moniker {
			return &moduleAdapter{module: &a.cfg.Repository.Modules[i], wsRoot: a.cfg.RepoRoot}, true
		}
	}
	return nil, false
}

func (a *repositoryAdapter) GetBuildOutputPath(workspaceRoot, moniker string) string {
	return paths.BuildOutputPath(workspaceRoot, moniker)
}

func (a *repositoryAdapter) GetTestOutputPath(workspaceRoot, moniker string) string {
	return paths.TestOutputPath(workspaceRoot, moniker)
}

func (a *repositoryAdapter) GetScanOutputPath(workspaceRoot, moniker string) string {
	return paths.ComponentScanOutputPath(workspaceRoot, moniker, "")
}

func (a *repositoryAdapter) GetLintOutputPath(workspaceRoot, moniker string) string {
	return paths.LintOutputPath(workspaceRoot, moniker)
}

func (a *repositoryAdapter) TestOutputDir() string {
	return a.cfg.Repository.Paths.Out.Test
}

func (a *repositoryAdapter) GetBuildArtifacts(moniker string, includeAll bool) []interfaces.ArtifactPort {
	artifacts := a.cfg.GetBuildArtifacts(moniker, includeAll)
	result := make([]interfaces.ArtifactPort, len(artifacts))
	for i := range artifacts {
		result[i] = &artifactAdapter{a: &artifacts[i]}
	}
	return result
}

func (a *repositoryAdapter) GetBuildArtifactIDs(moniker string, buildAll bool) []string {
	return a.cfg.GetBuildArtifactIDs(moniker, buildAll)
}

func (a *repositoryAdapter) GetBooksByModule(moniker string) []interfaces.BookConfigPort {
	books := a.cfg.GetBooksByModule(moniker)
	result := make([]interfaces.BookConfigPort, len(books))
	for i := range books {
		result[i] = &bookAdapter{b: books[i]}
	}
	return result
}

func (a *repositoryAdapter) GetDefaultBooksByModule(moniker string) []interfaces.BookConfigPort {
	books := a.cfg.GetDefaultBooksByModule(moniker)
	result := make([]interfaces.BookConfigPort, len(books))
	for i := range books {
		result[i] = &bookAdapter{b: books[i]}
	}
	return result
}

// moduleAdapter wraps config.Module for minimal interface compliance
type moduleAdapter struct {
	module *config.Module
	wsRoot string
}

func (a *moduleAdapter) GetMoniker() string                       { return a.module.Moniker }
func (a *moduleAdapter) GetName() string                          { return a.module.Name }
func (a *moduleAdapter) GetDescription() string                   { return a.module.Description }
func (a *moduleAdapter) HasComponent(compType string) bool        { return a.module.Components != nil && a.module.Components[compType] != nil }
func (a *moduleAdapter) GetComponentRoot(compType string) string  {
	if a.module.Components == nil {
		return ""
	}
	if entry := a.module.Components[compType]; entry != nil {
		return entry.Root
	}
	return ""
}
func (a *moduleAdapter) GetComponentRoots() map[string]string {
	if a.module.Components == nil {
		return nil
	}
	roots := make(map[string]string)
	for name, entry := range a.module.Components {
		if entry != nil {
			roots[name] = entry.Root
		}
	}
	return roots
}
func (a *moduleAdapter) GetComponentTypesDisplay() string { return "" }
func (a *moduleAdapter) GetComponentAmp(compName, op string) float64 { return 1.0 }
func (a *moduleAdapter) GetDependsOn() []string           { return a.module.DependsOn }
func (a *moduleAdapter) GetVersioningScheme() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.Scheme
}
func (a *moduleAdapter) GetReleaseType() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.ReleaseType
}
func (a *moduleAdapter) GetChangelog() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.Changelog
}
func (a *moduleAdapter) HasVersioning() bool { return a.module.Versioning != nil }
func (a *moduleAdapter) GetMetadata() map[string]interface{} {
	if a.module.Metadata == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range a.module.Metadata {
		result[k] = v
	}
	return result
}
func (a *moduleAdapter) GetContentHash() (string, error) { return "", nil }

// artifactAdapter wraps config.Artifact
type artifactAdapter struct {
	a *config.Artifact
}

func (a *artifactAdapter) GetID() string      { return a.a.ID }
func (a *artifactAdapter) GetType() string    { return a.a.Type }
func (a *artifactAdapter) GetPattern() string { return a.a.Pattern }

// bookAdapter wraps config.Book
type bookAdapter struct {
	b *config.Book
}

func (a *bookAdapter) GetName() string { return a.b.Name }
func (a *bookAdapter) GetType() string { return a.b.Template } // Template serves as type
func (a *bookAdapter) GetRoot() string { return "" }           // Books don't have root in this sense

// environmentsAdapter wraps config.EnvironmentsConfig
type environmentsAdapter struct {
	envs *config.EnvironmentsConfig
}

func (a *environmentsAdapter) GetEnvironment(name string) (interfaces.EnvironmentPort, bool) {
	env, ok := a.envs.GetEnvironment(name)
	if !ok {
		return nil, false
	}
	return &environmentAdapter{e: *env}, true
}

func (a *environmentsAdapter) ListEnvironments() []string {
	names := make([]string, len(a.envs.Environments))
	for i := range a.envs.Environments {
		names[i] = a.envs.Environments[i].Moniker
	}
	return names
}

type environmentAdapter struct {
	e config.Environment
}

func (a *environmentAdapter) GetName() string               { return a.e.Moniker }
func (a *environmentAdapter) GetVariables() map[string]string { return nil } // Environments don't have Variables field

// testingTagsAdapter wraps config.TestingTagsConfig
type testingTagsAdapter struct {
	tags *config.TestingTagsConfig
}

func (a *testingTagsAdapter) GetSkipReasons() map[string]string {
	// Convert SkipReason slice to map
	result := make(map[string]string)
	for _, sr := range a.tags.SkipReasons {
		result[sr.Code] = sr.Description
	}
	return result
}
func (a *testingTagsAdapter) GetTags() []string {
	result := make([]string, 0, len(a.tags.Tags))
	for _, tag := range a.tags.Tags {
		result = append(result, tag.Tag)
	}
	return result
}

// testSuitesAdapter wraps config.TestSuitesConfig
type testSuitesAdapter struct {
	suites *config.TestSuitesConfig
}

func (a *testSuitesAdapter) ListDefault() []string { return a.suites.ListDefault() }
func (a *testSuitesAdapter) GetSuiteLTags(suiteName string) []string {
	return a.suites.GetSuiteLTags(suiteName)
}
func (a *testSuitesAdapter) GetSuites() []interfaces.TestSuitePort {
	result := make([]interfaces.TestSuitePort, len(a.suites.Suites))
	for i := range a.suites.Suites {
		result[i] = &testSuiteDefAdapter{s: &a.suites.Suites[i]}
	}
	return result
}

type testSuiteDefAdapter struct {
	s *config.TestSuiteDef
}

func (a *testSuiteDefAdapter) GetName() string   { return a.s.Name }
func (a *testSuiteDefAdapter) GetTags() []string { return nil }                  // Tags are stored differently in selectors
func (a *testSuiteDefAdapter) IsDefault() bool   { return !a.s.ExtendedSuite }   // ExtendedSuite=true means NOT default

// componentTypesAdapter wraps config.ComponentTypesConfig
type componentTypesAdapter struct {
	types *config.ComponentTypesConfig
}

func (a *componentTypesAdapter) Get(name string) interfaces.ComponentTypePort {
	ct := a.types.Get(name)
	if ct == nil {
		return nil
	}
	return &componentTypeAdapter{ct: ct, name: name}
}

func (a *componentTypesAdapter) List() []string {
	if a.types == nil || a.types.ComponentTypes == nil {
		return nil
	}
	result := make([]string, 0, len(a.types.ComponentTypes))
	for name := range a.types.ComponentTypes {
		result = append(result, name)
	}
	return result
}

type componentTypeAdapter struct {
	ct   *config.ComponentType
	name string
}

func (a *componentTypeAdapter) GetName() string { return a.name }
func (a *componentTypeAdapter) GetPatterns() interfaces.ComponentPatternsPort {
	if a.ct.Files == nil {
		return nil
	}
	return &componentPatternsAdapter{f: a.ct.Files}
}
func (a *componentTypeAdapter) HasBuild() bool { return a.ct.HasBuilder() }

type componentPatternsAdapter struct {
	f *config.ComponentTypeFiles
}

func (a *componentPatternsAdapter) GetSource() []string { return a.f.Source }
func (a *componentPatternsAdapter) GetTests() []string  { return a.f.Tests }
func (a *componentPatternsAdapter) GetConfig() []string { return a.f.Config }

// ============================================================================
// Tool Registry Adapter
// ============================================================================

// toolRegistryAdapter wraps tool.DefaultRegistry
type toolRegistryAdapter struct {
	registry *tool.DefaultRegistry
}

func (a *toolRegistryAdapter) Get(canonicalName string) (interfaces.ToolDefinitionPort, bool) {
	t, ok := a.registry.Get(canonicalName)
	if !ok {
		return nil, false
	}
	return &toolDefinitionAdapter{t: t}, true
}

func (a *toolRegistryAdapter) GetForComponent(compType string, op interfaces.OperationType) []interfaces.ToolDefinitionPort {
	return nil // Not implemented in tool.DefaultRegistry
}

func (a *toolRegistryAdapter) Has(canonicalName string) bool {
	return a.registry.Has(canonicalName)
}

func (a *toolRegistryAdapter) AllCanonical() []string {
	all := a.registry.GetAll()
	result := make([]string, 0, len(all))
	for id := range all {
		result = append(result, id)
	}
	return result
}

type toolDefinitionAdapter struct {
	t *tool.ToolDefinition
}

func (a *toolDefinitionAdapter) GetID() string          { return a.t.ID }
func (a *toolDefinitionAdapter) GetCanonicalID() string { return a.t.ID }
func (a *toolDefinitionAdapter) GetDescription() string { return a.t.Description }
func (a *toolDefinitionAdapter) GetType() interfaces.ToolType {
	if a.t.Type == tool.ToolTypeSystem {
		return interfaces.ToolTypeSystem
	}
	return interfaces.ToolTypeContainer
}
func (a *toolDefinitionAdapter) GetBinary() string       { return a.t.Binary }
func (a *toolDefinitionAdapter) GetImage() string        { return a.t.Image }
func (a *toolDefinitionAdapter) GetCommand() []string    { return a.t.Command }
func (a *toolDefinitionAdapter) GetEntrypoint() []string { return a.t.Entrypoint }
func (a *toolDefinitionAdapter) GetEnv() map[string]string { return a.t.Env }
func (a *toolDefinitionAdapter) GetWorkDir() string      { return a.t.WorkDir }
func (a *toolDefinitionAdapter) GetResources() interfaces.ToolResourcesPort { return nil }
func (a *toolDefinitionAdapter) GetVerify() interfaces.ToolVerifyPort       { return nil }
func (a *toolDefinitionAdapter) DisplayName() string                        { return a.t.DisplayName() }
func (a *toolDefinitionAdapter) ShortName() string                          { return a.t.ShortName() }

// Compile-time interface check.
var _ interfaces.SimpleServicesPort = (*Services)(nil)

// Services provides core services for EAC commands.
// It implements the SimpleServicesPort interface.
type Services struct {
	workspaceRoot string
	configRoot    string
	config        interfaces.ConfigPort
	modules       interfaces.ModuleRegistryPort
	tools         interfaces.ToolRegistryPort

	// Cleanup functions to run on Close(), in reverse order
	cleanups []func() error
	mu       sync.Mutex
	closed   bool
}

// New creates a new Services instance with the given options.
//
// Steps:
// 1. Find workspace root using workspace.Detect()
// 2. Set config root to {workspaceRoot}/.eac
// 3. Load EAC config using config.Load()
// 4. If InitTools=true, initialize tool registry
// 5. Build module registry from config
// 6. If DebugMode=true, configure logging
// 7. Track cleanup functions to run in Close()
func New(opts interfaces.SimpleServicesOptions) (*Services, error) {
	// Step 1: Find workspace root
	ws, err := workspace.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect workspace: %w", err)
	}
	workspaceRoot := ws.Root

	// Step 2: Set config root
	configRoot := paths.EACConfigPath(workspaceRoot)

	// Step 3: Load EAC config
	loadOpts := config.LoadOptions{
		RepoRoot:        workspaceRoot,
		ValidateSchemas: true,
		LazyLoad:        false,
	}
	cfg, err := config.Load(loadOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create config adapter (local adapter, not from adapters package)
	configPort := &configAdapter{cfg: cfg}

	// Step 5: Build module registry from config
	// Use LoadFromWorkspace which properly converts config to domain models
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load modules: %w", err)
	}
	modulesPort := adapters.AdaptRegistry(registry)

	// Create services instance
	svc := &Services{
		workspaceRoot: workspaceRoot,
		configRoot:    configRoot,
		config:        configPort,
		modules:       modulesPort,
	}

	// Step 4: Initialize tools if requested
	if opts.InitTools {
		toolRegistry := tool.GlobalRegistry()

		// Load tool-config.yml if it exists (optional - don't fail if missing)
		toolConfigPath := filepath.Join(configRoot, "tool-config.yml")
		_ = toolRegistry.RegisterFromYAML(toolConfigPath)

		// Create tool registry adapter (local adapter)
		svc.tools = &toolRegistryAdapter{registry: toolRegistry}
	}

	// Note: DebugMode is passed through for future use but not currently
	// implemented - logging configuration happens at the command level

	return svc, nil
}

// WorkspaceRoot returns the repository root path.
func (s *Services) WorkspaceRoot() string {
	return s.workspaceRoot
}

// ConfigRoot returns the .eac configuration directory path.
func (s *Services) ConfigRoot() string {
	return s.configRoot
}

// Config returns the configuration access port.
func (s *Services) Config() interfaces.ConfigPort {
	return s.config
}

// Modules returns the module registry port.
func (s *Services) Modules() interfaces.ModuleRegistryPort {
	return s.modules
}

// Tools returns the tool registry port.
// Returns nil if InitTools was false in options.
func (s *Services) Tools() interfaces.ToolRegistryPort {
	return s.tools
}

// AddCleanup registers a cleanup function to be called on Close().
// Cleanup functions are called in reverse order of registration.
func (s *Services) AddCleanup(cleanup func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanups = append(s.cleanups, cleanup)
}

// RawConfig returns the underlying EACConfig for commands that need
// access to concrete types. Returns nil if the config adapter doesn't
// wrap an EACConfig. Prefer using Config() port-based access when possible.
func (s *Services) RawConfig() *config.EACConfig {
	if ca, ok := s.config.(*configAdapter); ok {
		return ca.cfg
	}
	return nil
}

// Close cleans up any resources held by the services.
// Cleanup functions are called in reverse order of registration.
// Close is idempotent - calling it multiple times is safe.
func (s *Services) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Idempotent - only close once
	if s.closed {
		return nil
	}
	s.closed = true

	// Run cleanups in reverse order
	var lastErr error
	for i := len(s.cleanups) - 1; i >= 0; i-- {
		if err := s.cleanups[i](); err != nil {
			lastErr = err
		}
	}

	// Clear cleanups slice
	s.cleanups = nil

	return lastErr
}
