package tool

import (
	"context"
	"fmt"

	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

// ToolSystem holds all tool infrastructure as explicit state.
// Created once during initialization; passed to consumers via
// constructor injection or stored as a field on ExecutionContext.
type ToolSystem struct {
	Config   *ToolConfig
	Registry *DefaultRegistry
	Resolver *DefaultResolver
	Executor *DefaultExecutor

	// Bridges (unified access for each action type)
	BuildBridge   *BuildBridge
	LintBridge    *LintBridge
	TestBridge    *TestBridge
	ScanBridge    *ScanBridge
	ServeBridge   *ServeBridge
	HandlerBridge *HandlerToolBridge

	// Category resolver
	Categories *CategoryResolver
}

// NewToolSystem creates and wires a complete tool system from config.
// This replaces InitializeGlobalBridges as the canonical creation path.
func NewToolSystem(repoRoot, configRoot string, containerProvider ContainerProvider) (*ToolSystem, error) {
	registry, resolver, toolConfig, err := InitializeFromConfig(repoRoot, configRoot)
	if err != nil {
		return nil, err
	}

	// Auto-detect environment
	env := resolver.DetectEnvironment()
	resolver.SetEnvironment(env)

	// Create executor
	executor := NewExecutorWithRegistry(registry)
	if toolConfig.Credentials != nil {
		executor.SetCredentials(toolConfig.Credentials)
	}
	if containerProvider != nil {
		executor.SetContainerProvider(containerProvider)
	}

	// Create handler bridge with registry and executor
	handlerBridge := &HandlerToolBridge{
		registry: registry,
		executor: executor,
	}

	// Create bridges
	ts := &ToolSystem{
		Config:        toolConfig,
		Registry:      registry,
		Resolver:      resolver,
		Executor:      executor,
		BuildBridge:   NewBuildBridge(),
		LintBridge:    NewLintBridge(),
		TestBridge:    NewTestBridge(),
		ScanBridge:    NewScanBridge(),
		ServeBridge:   NewServeBridge(),
		HandlerBridge: handlerBridge,
		Categories:    newCategoryResolver(),
	}

	// Wire bridges
	ts.BuildBridge.SetToolSystem(registry, resolver, executor)
	ts.LintBridge.SetToolSystem(registry, resolver, executor)
	ts.TestBridge.SetToolSystem(registry, resolver, executor)
	ts.ScanBridge.SetToolSystem(registry, resolver, executor)
	ts.ServeBridge.SetToolSystem(registry, resolver, executor)

	return ts, nil
}

// NewToolSystemForTesting creates a minimal ToolSystem with an empty registry.
// Use in tests to avoid loading real tool config.
func NewToolSystemForTesting() *ToolSystem {
	registry := NewRegistry()
	executor := NewExecutorWithRegistry(registry)

	return &ToolSystem{
		Config:        &ToolConfig{},
		Registry:      registry,
		Resolver:      NewResolver(registry),
		Executor:      executor,
		BuildBridge:   NewBuildBridge(),
		LintBridge:    NewLintBridge(),
		TestBridge:    NewTestBridge(),
		ScanBridge:    NewScanBridge(),
		ServeBridge:   NewServeBridge(),
		HandlerBridge: &HandlerToolBridge{registry: registry, executor: executor},
		Categories:    newCategoryResolver(),
	}
}

// GetOrAdhoc returns a tool from the registry, or creates an ad-hoc system tool.
func (ts *ToolSystem) GetOrAdhoc(binary string) *ToolDefinition {
	return ts.Registry.GetOrAdhoc(binary)
}

// Execute runs a tool with the given context.
func (ts *ToolSystem) Execute(ctx context.Context, tool *ToolDefinition, execCtx *ExecutionContext) (*ExecutionResult, error) {
	return ts.Executor.Execute(ctx, tool, execCtx)
}

// GetToolImage returns the Docker image for a tool ID.
// Returns empty string if tool is not found.
func (ts *ToolSystem) GetToolImage(toolID string) string {
	tool, ok := ts.Registry.Get(toolID)
	if !ok {
		return ""
	}
	return tool.FullImage()
}

// GetToolImageWithDefault returns the Docker image for a tool ID, or a default.
func (ts *ToolSystem) GetToolImageWithDefault(toolID, defaultImage string) string {
	if img := ts.GetToolImage(toolID); img != "" {
		return img
	}
	return defaultImage
}

// GetToolDefinition returns the tool definition by ID.
// Returns nil and false if not found.
func (ts *ToolSystem) GetToolDefinition(toolID string) (*ToolDefinition, bool) {
	return ts.Registry.Get(toolID)
}

// ScannerToolIDForCategory returns the tool ID for a scanner category.
func (ts *ToolSystem) ScannerToolIDForCategory(category string) string {
	return ts.Categories.ScannerToolIDForCategory(category)
}

// ServerToolIDForType returns the tool ID for a server type.
func (ts *ToolSystem) ServerToolIDForType(serverType string) string {
	return ts.Categories.ServerToolIDForType(serverType)
}

// GetTestTypeComponentType returns the component type for a test type.
// Uses the tool config's test-type-mapping first, then falls back
// to the adapter registry.
func (ts *ToolSystem) GetTestTypeComponentType(testType string) string {
	// 1. Try tool config (data-driven from YAML)
	if ts.Config != nil && ts.Config.TestTypeMapping != nil {
		if compType, ok := ts.Config.TestTypeMapping[testType]; ok {
			return compType
		}
	}

	// 2. Try adapter registry via provider
	if compType := coretesting.GetComponentTypeFromRegistry(testType); compType != "" {
		return compType
	}

	// 3. Return the test type itself as component type (no hardcoded fallback)
	return testType
}

// FilterPlatformSupported filters tool IDs to only those supported on the current platform.
func (ts *ToolSystem) FilterPlatformSupported(toolIDs []string) []string {
	var supported []string
	for _, toolID := range toolIDs {
		tool, ok := ts.Registry.Get(toolID)
		if !ok {
			supported = append(supported, toolID)
			continue
		}
		if IsPlatformSupported(tool) {
			supported = append(supported, toolID)
		}
	}
	return supported
}

// IsAvailable checks if a tool is available.
func (ts *ToolSystem) IsAvailable(toolID string) bool {
	return ts.Registry.IsAvailable(toolID)
}

// Verify checks if a tool is available and meets version requirements.
func (ts *ToolSystem) Verify(toolID string) VerifyResult {
	return ts.Registry.VerifyTool(toolID)
}

// VerifyAll checks multiple tools and returns results for each.
func (ts *ToolSystem) VerifyAll(toolIDs []string) []VerifyResult {
	return ts.Registry.VerifyAll(toolIDs)
}

// GetMissingDependencies returns list of unavailable tools.
func (ts *ToolSystem) GetMissingDependencies(toolIDs []string) []string {
	return ts.Registry.GetMissingTools(toolIDs)
}

// Close releases resources held by the tool system.
func (ts *ToolSystem) Close() error {
	if ts.Executor != nil {
		return ts.Executor.Close()
	}
	return nil
}

// newCategoryResolver creates a new non-global CategoryResolver instance.
func newCategoryResolver() *CategoryResolver {
	r := &CategoryResolver{}
	r.initDefaultMappings()
	return r
}

// ExecuteToolByName looks up a tool by name and executes it.
// This is a convenience method combining GetOrAdhoc + Execute.
func (ts *ToolSystem) ExecuteToolByName(ctx context.Context, toolName string, execCtx *ExecutionContext) (*ExecutionResult, error) {
	tool := ts.GetOrAdhoc(toolName)
	if tool == nil {
		return nil, fmt.Errorf("tool %q not found", toolName)
	}
	return ts.Execute(ctx, tool, execCtx)
}

// RunTool executes a tool by name with simple args. Returns stdout.
// Non-zero exit codes are returned as errors.
func (ts *ToolSystem) RunTool(ctx context.Context, toolName, workDir string, args ...string) ([]byte, error) {
	result, err := ts.ExecuteToolByName(ctx, toolName, &ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	})
	if err != nil {
		return nil, fmt.Errorf("%s execution failed: %w", toolName, err)
	}
	if result.ExitCode != 0 {
		return result.Stderr, fmt.Errorf("%s command failed (exit %d): %s",
			toolName, result.ExitCode, string(result.Stderr))
	}
	return result.Stdout, nil
}

// RunToolCombined executes a tool returning stdout + exit code.
// Non-zero exit codes are NOT treated as errors — useful for commands
// where non-zero exit has semantic meaning.
func (ts *ToolSystem) RunToolCombined(ctx context.Context, toolName, workDir string, args ...string) ([]byte, int, error) {
	result, err := ts.ExecuteToolByName(ctx, toolName, &ExecutionContext{
		WorkspaceRoot: workDir,
		ModuleRoot:    workDir,
		ArgsOverrides: args,
	})
	if err != nil {
		return nil, -1, fmt.Errorf("%s execution failed: %w", toolName, err)
	}
	return result.Stdout, result.ExitCode, nil
}
