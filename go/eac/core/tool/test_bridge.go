// Package tool provides a test bridge that integrates the tool system.
package tool

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// TestFunc is the signature for module test functions.
// Parameters: module contract, workspace root, output directory, log writer, report format, suite name
// Returns: exit code.
type TestFunc func(*modules.ModuleContract, string, string, io.Writer, string, string) int

// TestBridge provides a unified interface for resolving test handlers.
// All handlers are resolved from tool-config.yml definitions.
type TestBridge struct {
	mu sync.RWMutex

	// Component type to handler mapping
	componentTestHandlers map[string]string

	// Tool system integration
	registry Registry
	resolver *DefaultResolver
	executor Executor
}

// NewTestBridge creates a new test bridge.
func NewTestBridge() *TestBridge {
	return &TestBridge{
		// All mappings come from component-tools in tool-config.yml via the resolver.
		// No hardcoded fallbacks - tool names should not be embedded in code.
		componentTestHandlers: map[string]string{},
	}
}

// SetToolSystem configures the tool system for tool-config.yml defined tools.
func (b *TestBridge) SetToolSystem(registry Registry, resolver *DefaultResolver, executor Executor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
	b.resolver = resolver
	b.executor = executor
}

// SetComponentMapping sets or updates a component type to handler mapping.
func (b *TestBridge) SetComponentMapping(compType, handlerName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.componentTestHandlers[compType] = handlerName
}

// GetTestFunc returns the appropriate test function for a module.
// It matches module component types to test handlers from tool-config.yml.
// Resolution order:
// 1. Resolver (component-tools mapping from tool-config.yml)
// 2. Hardcoded componentTestHandlers fallback
func (b *TestBridge) GetTestFunc(module *modules.ModuleContract) TestFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if module == nil {
		return noOpTestFunc
	}

	// Get all component types for this module
	componentTypes := module.Components.GetComponentTypes()

	// Priority 1: Try resolver for each component type (uses component-tools mapping)
	if b.resolver != nil && b.executor != nil {
		for _, compType := range componentTypes {
			if tool, err := b.resolver.Resolve(compType, OperationTest); err == nil && tool != nil {
				return b.createToolTestFunc(tool)
			}
		}
	}

	// Priority 2: Fall back to hardcoded mappings
	for compType, handlerName := range b.componentTestHandlers {
		if module.HasComponent(compType) {
			// Try YAML-defined tool
			if b.registry != nil && b.executor != nil {
				if tool, ok := b.registry.Get(handlerName); ok {
					return b.createToolTestFunc(tool)
				}
			}
		}
	}

	return noOpTestFunc
}

// createToolTestFunc wraps a ToolDefinition as a TestFunc.
func (b *TestBridge) createToolTestFunc(tool *ToolDefinition) TestFunc {
	return func(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, reportFormat, suiteName string) int {
		adapter := NewTestHandlerAdapter(tool, b.executor)
		opts := TestOptions{
			Verbose: false,
		}
		return adapter.Test(module, workspaceRoot, outputDir, logWriter, opts)
	}
}

// noOpTestFunc is the default no-op test function.
func noOpTestFunc(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, reportFormat, suiteName string) int {
	return 0
}

// HasHandler checks if a handler exists for the given name.
func (b *TestBridge) HasHandler(name string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.registry != nil {
		if _, ok := b.registry.Get(name); ok {
			return true
		}
	}

	return false
}

// ResolveTool returns the tool definition for a component type and operation.
// Returns nil if no tool is configured or resolver is not available.
func (b *TestBridge) ResolveTool(componentType string, operation OperationType) *ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.resolver == nil {
		return nil
	}

	t, err := b.resolver.Resolve(componentType, operation)
	if err != nil {
		return nil
	}

	return t
}

// IsContainer returns true if the test handler for the given component type runs in a container.
func (b *TestBridge) IsContainer(componentType string) bool {
	t := b.ResolveTool(componentType, OperationTest)
	if t == nil {
		return false
	}
	return t.Type == ToolTypeContainer
}

// Global test bridge instance.
var (
	globalTestBridge     *TestBridge
	globalTestBridgeOnce sync.Once
)

// GlobalTestBridge returns the global test bridge instance.
func GlobalTestBridge() *TestBridge {
	globalTestBridgeOnce.Do(func() {
		globalTestBridge = NewTestBridge()
	})
	return globalTestBridge
}
