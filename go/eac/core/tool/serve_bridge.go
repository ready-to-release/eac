// Package tool provides a serve bridge that integrates the tool system.
package tool

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
)

// ServeOptions contains options for serve execution.
type ServeOptions struct {
	LiveReload  bool     // Enable live reload on file changes
	OpenBrowser bool     // Automatically open browser
	WatchPaths  []string // Paths to watch for changes
}

// ServeResult captures the outcome of starting a server.
type ServeResult struct {
	Port    int    // Port the server is running on
	URL     string // Full URL to access the server
	PID     int    // Process ID (for process-based servers)
	Running bool   // Whether the server is currently running
}

// ServeFunc is the signature for server functions.
// Parameters: workspace root, module root, content path, port, log writer, serve options
// Returns: serve result, error
type ServeFunc func(workspaceRoot, moduleRoot, contentPath string, port int, logWriter io.Writer, opts ServeOptions) (*ServeResult, error)

// ServeBridge provides a unified interface for resolving serve handlers.
// All servers are resolved dynamically from tool-config.yml definitions.
// Server tools are detected by their IsServable() capability (having a Serve config).
type ServeBridge struct {
	mu sync.RWMutex

	// Tool system integration
	registry      Registry
	resolver      *DefaultResolver
	executor      Executor
	toolSystemSet bool // tracks if SetToolSystem was called
}

// NewServeBridge creates a new serve bridge.
func NewServeBridge() *ServeBridge {
	return &ServeBridge{}
}

// SetToolSystem configures the tool system for tool-config.yml defined servers.
func (b *ServeBridge) SetToolSystem(registry Registry, resolver *DefaultResolver, executor Executor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
	b.resolver = resolver
	b.executor = executor
	b.toolSystemSet = true
}

// getRegistry returns the registry to use, falling back to GlobalRegistry.
// Only falls back if SetToolSystem was not explicitly called.
func (b *ServeBridge) getRegistry() Registry {
	if b.registry != nil {
		return b.registry
	}
	// Only fall back to global if SetToolSystem was not explicitly called
	if !b.toolSystemSet {
		return GlobalRegistry()
	}
	return nil
}

// getExecutor returns the executor to use, falling back to GlobalExecutor.
// Only falls back if SetToolSystem was not explicitly called.
func (b *ServeBridge) getExecutor() Executor {
	if b.executor != nil {
		return b.executor
	}
	// Only fall back to global if SetToolSystem was not explicitly called
	if !b.toolSystemSet {
		return GlobalExecutor()
	}
	return nil
}

// GetServerByToolID returns the server function for a tool ID from tool-config.yml.
// The tool must be servable (have a Serve configuration).
func (b *ServeBridge) GetServerByToolID(toolID string) ServeFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	executor := b.getExecutor()

	if registry != nil && executor != nil {
		if tool, ok := registry.Get(toolID); ok && tool.IsServable() {
			return b.createToolServeFuncWithExecutor(tool, executor)
		}
	}

	return nil
}

// GetServerForComponent returns the server function for a component type.
// Uses the resolver to look up the server tool from component-tools config.
func (b *ServeBridge) GetServerForComponent(componentType string) ServeFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.resolver == nil {
		return nil
	}

	executor := b.getExecutor()
	if executor == nil {
		return nil
	}

	tool, err := b.resolver.Resolve(componentType, OperationServe)
	if err != nil || tool == nil {
		return nil
	}

	if !tool.IsServable() {
		return nil
	}

	return b.createToolServeFuncWithExecutor(tool, executor)
}

// GetServerTool returns the tool definition for a component type's server.
// Returns nil if no server is configured or the tool is not servable.
func (b *ServeBridge) GetServerTool(componentType string) *ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.resolver == nil {
		return nil
	}

	tool, err := b.resolver.Resolve(componentType, OperationServe)
	if err != nil || tool == nil {
		return nil
	}

	if !tool.IsServable() {
		return nil
	}

	return tool
}

// GetServerToolByID returns the tool definition by ID if it's servable.
// Returns nil if the tool doesn't exist or is not servable.
func (b *ServeBridge) GetServerToolByID(toolID string) *ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry == nil {
		return nil
	}

	tool, ok := registry.Get(toolID)
	if !ok || !tool.IsServable() {
		return nil
	}

	return tool
}

// HasServerForComponent checks if a server is available for the given component type.
func (b *ServeBridge) HasServerForComponent(componentType string) bool {
	return b.GetServerTool(componentType) != nil
}

// HasServerByToolID checks if a server tool exists and is servable.
func (b *ServeBridge) HasServerByToolID(toolID string) bool {
	return b.GetServerToolByID(toolID) != nil
}

// GetAllServableTools returns all tools that can be used as servers.
// A tool is servable if it has a Serve configuration with a container port.
func (b *ServeBridge) GetAllServableTools() []*ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry == nil {
		return nil
	}

	var servable []*ToolDefinition
	for _, tool := range registry.GetAll() {
		if tool.IsServable() {
			servable = append(servable, tool)
		}
	}
	return servable
}

// createToolServeFuncWithExecutor wraps a ToolDefinition as a ServeFunc with an explicit executor.
func (b *ServeBridge) createToolServeFuncWithExecutor(tool *ToolDefinition, executor Executor) ServeFunc {
	return func(workspaceRoot, moduleRoot, contentPath string, port int, logWriter io.Writer, opts ServeOptions) (*ServeResult, error) {
		execCtx := &ExecutionContext{
			WorkspaceRoot: workspaceRoot,
			ModuleRoot:    moduleRoot,
			OutputDir:     contentPath,
			LogWriter:     logWriter,
			Operation:     OperationServe,
			Placeholders: map[string]string{
				"{workspace}": workspaceRoot,
				"{module}":    filepath.Join(workspaceRoot, moduleRoot),
				"{content}":   contentPath,
				"{port}":      fmt.Sprintf("%d", port),
			},
		}

		ctx := context.Background()
		result, err := executor.Execute(ctx, tool, execCtx)
		if err != nil {
			return nil, fmt.Errorf("server execution failed: %w", err)
		}

		// Extract serve-specific configuration from tool
		containerPort := port
		if tool.Serve != nil && tool.Serve.ContainerPort > 0 {
			containerPort = tool.Serve.ContainerPort
		}

		return &ServeResult{
			Port:    containerPort,
			URL:     fmt.Sprintf("http://localhost:%d", port),
			Running: result.ExitCode == 0,
		}, nil
	}
}

// Global serve bridge instance.
var (
	globalServeBridge     *ServeBridge
	globalServeBridgeOnce sync.Once
)

// GlobalServeBridge returns the global serve bridge instance.
func GlobalServeBridge() *ServeBridge {
	globalServeBridgeOnce.Do(func() {
		globalServeBridge = NewServeBridge()
	})
	return globalServeBridge
}
