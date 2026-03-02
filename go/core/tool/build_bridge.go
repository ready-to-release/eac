// Package tool provides a build bridge that integrates the tool system.
package tool

import (
	"fmt"
	"os/exec"
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// BuildBridge provides a unified interface for resolving build handlers.
// Handlers are resolved from native handlers first, then tool-config.yml definitions.
type BuildBridge struct {
	mu sync.RWMutex

	// Native handlers (registered from commands/impl/build/builders)
	nativeHandlers map[string]build.BuilderPort

	// Tool system integration
	registry Registry
	resolver *DefaultResolver
	executor Executor

	// lookPathFunc checks if a binary exists on PATH. Defaults to exec.LookPath.
	// Override in tests to control requirement checking.
	lookPathFunc func(string) (string, error)
}

// NewBuildBridge creates a new build bridge.
func NewBuildBridge() *BuildBridge {
	return &BuildBridge{
		nativeHandlers: make(map[string]build.BuilderPort),
		lookPathFunc:   exec.LookPath,
	}
}

// RegisterNativeHandler registers a native build handler.
// Native handlers take precedence over tool-config.yml handlers.
// Call this from init() in builder files.
func (b *BuildBridge) RegisterNativeHandler(h build.BuilderPort) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nativeHandlers[h.Name()] = h
}

// ResetForTesting clears all native handler registrations.
// Use only in tests that need a clean bridge state.
func (b *BuildBridge) ResetForTesting() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nativeHandlers = make(map[string]build.BuilderPort)
}

// SetToolSystem configures the tool system for tool-config.yml defined tools.
func (b *BuildBridge) SetToolSystem(registry Registry, resolver *DefaultResolver, executor Executor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
	b.resolver = resolver
	b.executor = executor
}

// GetHandler returns a build handler by name.
// Checks native handlers first, then falls back to tool registry.
func (b *BuildBridge) GetHandler(name string) build.BuilderPort {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.getHandlerUnlocked(name)
}

// GetAllHandlers returns all available handlers (native + tool registry).
// Native handlers are only included if their binding allows and requirements are met.
func (b *BuildBridge) GetAllHandlers() map[string]build.BuilderPort {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]build.BuilderPort)

	// Add tool registry handlers first
	if b.registry != nil && b.executor != nil {
		for name, tool := range b.registry.GetAll() {
			result[name] = NewToolHandlerAdapter(tool, b.executor)
		}
	}

	// Native handlers: only add if binding allows and requirements met
	for name, h := range b.nativeHandlers {
		binding := b.getBindingForTool(name)
		if binding == ToolBindingContainer {
			continue // Skip native in container mode
		}
		if binding == ToolBindingSystem || b.areRequirementsMet(h) {
			result[name] = h // Override registry handler
		}
	}

	return result
}

// ComponentBuildHandler pairs a component name with its build handler.
type ComponentBuildHandler struct {
	Component string
	Handler   build.BuilderPort
}

// GetHandlersForModule returns build handlers for all buildable components in a module.
// Handler selection:
// 1. Module-level handler override (build.handler in module config)
// 2. Component-kind builders (from blueprints.yml component-kinds)
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
	if cfg == nil || cfg.ComponentKinds == nil {
		return nil
	}

	// Priority 2: Find builders from component types
	for _, compName := range module.GetEnabledComponents() {
		compTypeName := module.Components.GetComponentType(compName)
		compType := cfg.ComponentKinds.Get(compTypeName)
		if compType == nil || !compType.IsBuildable() {
			continue
		}

		builderName := compType.GetBuilders()[0]

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
			if tool, err := b.resolver.Resolve(compTypeName, core.ActionBuild); err == nil && tool != nil {
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
// Respects executor-mode and per-tool bindings:
//   - container mode: skip native handlers, use registry only
//   - system mode: always use native handler (fail fast if broken)
//   - auto mode: use native if requirements met, else fall through to registry
func (b *BuildBridge) getHandlerUnlocked(name string) build.BuilderPort {
	binding := b.getBindingForTool(name)

	if binding != ToolBindingContainer {
		if h, ok := b.nativeHandlers[name]; ok {
			if binding == ToolBindingSystem || b.areRequirementsMet(h) {
				return h
			}
			// Auto mode: requirements not met, fall through
		}
	}

	// Fall back to tool registry (which has its own auto/system/container logic)
	if b.registry != nil && b.executor != nil {
		if tool, ok := b.registry.Get(name); ok {
			return NewToolHandlerAdapter(tool, b.executor)
		}
	}

	return nil
}

// HasHandler checks if a handler exists by name.
// Respects binding mode and requirements for native handlers.
func (b *BuildBridge) HasHandler(name string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Check native handlers with availability gating
	if h, ok := b.nativeHandlers[name]; ok {
		binding := b.getBindingForTool(name)
		if binding != ToolBindingContainer {
			if binding == ToolBindingSystem || b.areRequirementsMet(h) {
				return true
			}
		}
	}

	// Check tool registry
	if b.registry != nil {
		if _, ok := b.registry.Get(name); ok {
			return true
		}
	}

	return false
}

// GetToolForComponent returns the tool definition for a component type.
// This is used to access tool resources for scheduling weight calculation.
// Resolution order: resolver component-tools mapping → direct registry lookup.
// Returns nil if no tool is found for the component type.
func (b *BuildBridge) GetToolForComponent(componentType string) *ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var toolID string

	// Try resolver first (for component-tools mappings)
	if b.resolver != nil {
		toolID = b.resolver.ResolveToolID(componentType, core.ActionBuild)
	}

	// If resolver didn't find it, try direct registry lookup
	// (componentType might be the tool name itself)
	if toolID == "" {
		toolID = componentType
	}

	// Look up tool in registry
	if b.registry != nil {
		if tool, ok := b.registry.Get(toolID); ok {
			return tool
		}
	}

	return nil
}

// GetHandlerForComponent returns a build handler for a component type using the resolver.
// This uses the component-tools mapping to find the correct tool (e.g., typescript → npm-build).
// Native handlers take precedence over tool registry definitions.
// Falls back to blueprints.yml component-kinds builder field when resolver is unavailable.
func (b *BuildBridge) GetHandlerForComponent(componentType string) build.BuilderPort {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var toolID string

	// Try resolver first (for component-tools mappings)
	if b.resolver != nil {
		toolID = b.resolver.ResolveToolID(componentType, core.ActionBuild)
	}

	// Fall back to blueprints.yml component-kinds builders field
	if toolID == "" {
		cfg := config.Global()
		if cfg != nil && cfg.ComponentKinds != nil {
			if compType := cfg.ComponentKinds.Get(componentType); compType != nil && compType.IsBuildable() {
				toolID = compType.GetBuilders()[0]
			}
		}
	}

	if toolID == "" {
		return nil
	}

	// Check native handler with availability gating
	if h, ok := b.nativeHandlers[toolID]; ok {
		binding := b.getBindingForTool(toolID)
		if binding != ToolBindingContainer {
			if binding == ToolBindingSystem || b.areRequirementsMet(h) {
				return h
			}
		}
	}

	// Fall back to tool registry lookup
	if b.registry == nil || b.executor == nil {
		return nil
	}
	tool, ok := b.registry.Get(toolID)
	if !ok {
		return nil
	}
	return NewToolHandlerAdapter(tool, b.executor)
}

// ResolveTool returns the tool definition for a component type and operation.
// Returns nil if no tool is configured or resolver is not available.
func (b *BuildBridge) ResolveTool(componentType string, operation core.ActionType) *ToolDefinition {
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

// getBindingForTool determines the effective binding mode for a tool name.
func (b *BuildBridge) getBindingForTool(name string) ToolBinding {
	if b.registry == nil {
		return ToolBindingAuto
	}
	reg, ok := b.registry.(*DefaultRegistry)
	if !ok {
		return ToolBindingAuto
	}
	return reg.GetBindingForTool(name)
}

// areRequirementsMet checks if a native handler's system requirements are satisfied.
// Uses two-tier verification:
//  1. Fast-reject: lookPathFunc checks binary exists on PATH (~0ms)
//  2. Full verify: registry.IsAvailable checks version requirements (cached after first call)
//
// Returns true if the handler declares no requirements.
func (b *BuildBridge) areRequirementsMet(h build.BuilderPort) bool {
	reqs := h.Requirements()
	if len(reqs) == 0 {
		return true
	}

	// Tier 1: Fast-reject via PATH check
	lookPath := b.lookPathFunc
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	for _, req := range reqs {
		if _, err := lookPath(req); err != nil {
			return false // Binary not on PATH, skip expensive verification
		}
	}

	// Tier 2: Full verification via registry (version checks, cached)
	reg, ok := b.registry.(*DefaultRegistry)
	if !ok {
		return true // No registry available, PATH check passed
	}
	for _, req := range reqs {
		if !reg.IsAvailable(req) {
			return false
		}
	}
	return true
}

// PreWarmRequirements verifies all native handler requirements in parallel.
// Call during initialization to populate the registry's verify cache.
// Subsequent areRequirementsMet calls return cached results instantly.
func (b *BuildBridge) PreWarmRequirements() {
	b.mu.RLock()
	reg, ok := b.registry.(*DefaultRegistry)
	if !ok {
		b.mu.RUnlock()
		return
	}

	// Collect unique requirements from all native handlers
	unique := make(map[string]bool)
	for _, h := range b.nativeHandlers {
		for _, req := range h.Requirements() {
			unique[req] = true
		}
	}
	b.mu.RUnlock()

	if len(unique) == 0 {
		return
	}

	// Fast-reject: skip requirements not on PATH
	lookPath := b.lookPathFunc
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	var toVerify []string
	for req := range unique {
		if _, err := lookPath(req); err == nil {
			toVerify = append(toVerify, req)
		}
	}

	if len(toVerify) == 0 {
		return
	}

	// Collect tool definitions under read lock
	var tools []*ToolDefinition

	reg.mu.RLock()
	verifier := reg.verifier
	for _, req := range toVerify {
		if _, ok := reg.verifyCache[req]; ok {
			continue // Already cached
		}
		systemID := req + ":system"
		if tool, ok := reg.tools[systemID]; ok {
			tools = append(tools, tool)
		}
	}
	reg.mu.RUnlock()

	if len(tools) == 0 {
		return
	}

	// Verify in parallel (no lock held during expensive shell commands)
	type verifyResult struct {
		toolID    string
		available bool
	}
	results := make([]verifyResult, len(tools))
	var wg sync.WaitGroup
	for i, tool := range tools {
		wg.Add(1)
		go func(idx int, td *ToolDefinition) {
			defer wg.Done()
			var avail bool
			if verifier != nil {
				avail = verifier(td)
			} else {
				vr := VerifyToolDefinition(td)
				avail = vr.Available
			}
			results[idx] = verifyResult{td.ID, avail}
		}(i, tool)
	}
	wg.Wait()

	// Write all results to cache under write lock
	reg.mu.Lock()
	for _, r := range results {
		reg.verifyCache[r.toolID] = r.available
	}
	reg.mu.Unlock()
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

// globalToolSystem holds the ToolSystem created by InitializeGlobalBridges.
var globalToolSystem *ToolSystem

// GlobalToolSystem returns the global ToolSystem, or nil if not initialized.
func GlobalToolSystem() *ToolSystem {
	return globalToolSystem
}

// SetGlobalToolSystemForTesting sets the global ToolSystem to a test instance.
// Use NewToolSystemForTesting() to create a minimal instance. Call this in
// TestMain when enabling in-process command dispatch for commands that depend
// on GlobalToolSystem().
func SetGlobalToolSystemForTesting(ts *ToolSystem) {
	globalToolSystem = ts
}

// initBridgesOnce ensures InitializeGlobalBridges body runs only once per process,
// making the second call (from phaseInitDeferred) a no-op returning the cached result.
var (
	initBridgesOnce sync.Once
	initBridgesErr  error
)

// InitializeGlobalBridges initializes all global bridges (build, lint, test, scan, serve) with tool system.
// Call this during application startup after loading configuration.
// Internally creates a ToolSystem and populates all legacy globals for backward compatibility.
// Safe to call multiple times: only the first call performs work; subsequent calls return the
// cached result immediately.
func InitializeGlobalBridges(repoRoot, configRoot string) error {
	initBridgesOnce.Do(func() {
		initBridgesErr = initializeGlobalBridgesImpl(repoRoot, configRoot)
	})
	return initBridgesErr
}

func initializeGlobalBridgesImpl(repoRoot, configRoot string) error {
	ts, err := NewToolSystem(repoRoot, configRoot, defaultContainerProvider)
	if err != nil {
		// Tool config is optional — no tool-config.yml is expected in many repos.
		// Return the error so callers can log it for diagnostic purposes.
		return fmt.Errorf("tool system init: %w", err)
	}

	globalToolSystem = ts

	// Backward compat: populate legacy globals so existing callers continue to work.
	SetGlobalRegistry(ts.Registry)
	SetGlobalExecutor(ts.Executor)

	// Wire bridges: use ToolSystem's pre-created bridges by copying handler registrations.
	// The global bridges may have native handlers registered via init(), so we wire tool
	// system into those existing bridges rather than replacing them.
	GlobalBuildBridge().SetToolSystem(ts.Registry, ts.Resolver, ts.Executor)
	GlobalLintBridge().SetToolSystem(ts.Registry, ts.Resolver, ts.Executor)
	GlobalTestBridge().SetToolSystem(ts.Registry, ts.Resolver, ts.Executor)
	GlobalScanBridge().SetToolSystem(ts.Registry, ts.Resolver, ts.Executor)
	GlobalServeBridge().SetToolSystem(ts.Registry, ts.Resolver, ts.Executor)
	GlobalDeployBridge().SetToolSystem(ts.Registry, ts.Resolver, ts.Executor)

	// Pre-warm native handler requirement cache in parallel.
	// Runs ~4-5 verification commands concurrently (~500ms wall time).
	// Only build bridge needs pre-warming — all bridges share the same registry
	// and verify cache, so results are available to lint/test/scan bridges too.
	GlobalBuildBridge().PreWarmRequirements()

	return nil
}
