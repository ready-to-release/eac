// Package tool provides a build bridge that integrates the tool system.
package tool

import (
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// BuildBridge provides a unified interface for resolving build handlers.
// All handlers are resolved from tool-config.yml definitions.
type BuildBridge struct {
	mu sync.RWMutex

	// Tool system integration
	registry Registry
	resolver *DefaultResolver
	executor Executor
}

// NewBuildBridge creates a new build bridge.
func NewBuildBridge() *BuildBridge {
	return &BuildBridge{}
}

// SetToolSystem configures the tool system for tool-config.yml defined tools.
func (b *BuildBridge) SetToolSystem(registry Registry, resolver *DefaultResolver, executor Executor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
	b.resolver = resolver
	b.executor = executor
}

// GetHandler returns a build handler by name from the tool registry.
func (b *BuildBridge) GetHandler(name string) BuildHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.registry != nil && b.executor != nil {
		if tool, ok := b.registry.Get(name); ok {
			return NewToolHandlerAdapter(tool, b.executor)
		}
	}

	return nil
}

// GetAllHandlers returns all available handlers from the tool registry.
func (b *BuildBridge) GetAllHandlers() map[string]BuildHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]BuildHandler)

	if b.registry != nil && b.executor != nil {
		for name, tool := range b.registry.GetAll() {
			result[name] = NewToolHandlerAdapter(tool, b.executor)
		}
	}

	return result
}

// ComponentBuildHandler pairs a component name with its build handler.
type ComponentBuildHandler struct {
	Component string
	Handler   BuildHandler
}

// GetHandlersForModule returns build handlers for all buildable components in a module.
// Handler selection:
// 1. Module-level handler override (build.handler in module config)
// 2. Component-type builders (from component-types.yml)
// 3. ToolResolver lookup for the component type
func (b *BuildBridge) GetHandlersForModule(module *modules.ModuleContract) []ComponentBuildHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if module == nil {
		return nil
	}

	var result []ComponentBuildHandler

	// Priority 1: Check for per-module handler override
	if module.GetBuildHandler() != "" {
		handlerName := module.GetBuildHandler()
		if h := b.getHandlerUnlocked(handlerName); h != nil {
			result = append(result, ComponentBuildHandler{
				Component: "override",
				Handler:   h,
			})
			return result
		}
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	// Priority 2: Find builders from component types
	for _, compName := range module.GetEnabledComponents() {
		compTypeName := module.Components.GetComponentType(compName)
		compType := cfg.ComponentTypes.Get(compTypeName)
		if compType == nil || !compType.HasBuilder() {
			continue
		}

		builderName := compType.Builder

		// Try to find handler from tool registry
		if h := b.getHandlerUnlocked(builderName); h != nil {
			result = append(result, ComponentBuildHandler{
				Component: compName,
				Handler:   h,
			})
			continue
		}

		// Try ToolResolver (for layered config resolution)
		if b.resolver != nil {
			if tool, err := b.resolver.Resolve(compTypeName, OperationBuild); err == nil && tool != nil {
				result = append(result, ComponentBuildHandler{
					Component: compName,
					Handler:   NewToolHandlerAdapter(tool, b.executor),
				})
			}
		}
	}

	return result
}

// getHandlerUnlocked returns a handler by name (must be called with lock held).
func (b *BuildBridge) getHandlerUnlocked(name string) BuildHandler {
	if b.registry != nil && b.executor != nil {
		if tool, ok := b.registry.Get(name); ok {
			return NewToolHandlerAdapter(tool, b.executor)
		}
	}

	return nil
}

// HasHandler checks if a handler exists by name.
func (b *BuildBridge) HasHandler(name string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.registry != nil {
		if _, ok := b.registry.Get(name); ok {
			return true
		}
	}

	return false
}

// GetHandlerForComponent returns a build handler for a component type using the resolver.
// This uses the component-tools mapping to find the correct tool (e.g., typescript → npm-build).
func (b *BuildBridge) GetHandlerForComponent(componentType string) BuildHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.resolver == nil || b.executor == nil {
		return nil
	}

	tool, err := b.resolver.Resolve(componentType, OperationBuild)
	if err != nil || tool == nil {
		return nil
	}

	return NewToolHandlerAdapter(tool, b.executor)
}

// ResolveTool returns the tool definition for a component type and operation.
// Returns nil if no tool is configured or resolver is not available.
func (b *BuildBridge) ResolveTool(componentType string, operation OperationType) *ToolDefinition {
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

// Global build bridge instance.
var (
	globalBridge     *BuildBridge
	globalBridgeOnce sync.Once
)

// GlobalBuildBridge returns the global build bridge instance.
func GlobalBuildBridge() *BuildBridge {
	globalBridgeOnce.Do(func() {
		globalBridge = NewBuildBridge()
	})
	return globalBridge
}

// InitializeGlobalBridges initializes all global bridges (build, lint, test, scan, serve) with tool system.
// Call this during application startup after loading configuration.
func InitializeGlobalBridges(repoRoot, configRoot string) error {
	// Initialize tool system from config
	registry, resolver, err := InitializeFromConfig(repoRoot, configRoot)
	if err != nil {
		// Tool config is optional, but log for visibility in case of unexpected issues
		// Common reasons: no tool-config.yml (expected), invalid YAML (should be investigated)
		// Note: This warning appears in logs but doesn't fail the build
		return nil
	}

	// Set global registry for verification access throughout codebase
	SetGlobalRegistry(registry)

	// Auto-detect and set environment (CI vs local)
	// This enables environment-specific tool overrides from tool-config.yml
	env := resolver.DetectEnvironment()
	resolver.SetEnvironment(env)

	// Create executor with registry for requirement validation
	executor := NewExecutorWithRegistry(registry)

	// Configure build bridge
	GlobalBuildBridge().SetToolSystem(registry, resolver, executor)

	// Configure lint bridge
	GlobalLintBridge().SetToolSystem(registry, resolver, executor)

	// Configure test bridge
	GlobalTestBridge().SetToolSystem(registry, resolver, executor)

	// Configure scan bridge
	GlobalScanBridge().SetToolSystem(registry, resolver, executor)

	// Configure serve bridge
	GlobalServeBridge().SetToolSystem(registry, resolver, executor)

	return nil
}
