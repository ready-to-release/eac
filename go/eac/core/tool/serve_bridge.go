// Package tool provides a serve bridge that integrates the tool system.
package tool

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
)

// ServerType represents the type of server.
// This allows for different serve implementations (static site, live reload, etc.)
type ServerType string

const (
	ServerStaticSite  ServerType = "static-site"
	ServerMkDocsLive  ServerType = "mkdocs-live"
	ServerStructurizr ServerType = "structurizr"
)

// ParseServerType converts a string to ServerType.
func ParseServerType(s string) (ServerType, bool) {
	switch s {
	case "static-site":
		return ServerStaticSite, true
	case "mkdocs-live":
		return ServerMkDocsLive, true
	case "structurizr":
		return ServerStructurizr, true
	default:
		return "", false
	}
}

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
// All servers are resolved from tool-config.yml definitions.
type ServeBridge struct {
	mu sync.RWMutex

	// Server type to tool ID mapping (from tool-config.yml)
	serverTools map[ServerType]string

	// Tool system integration
	registry      Registry
	resolver      *DefaultResolver
	executor      Executor
	toolSystemSet bool // tracks if SetToolSystem was called
}

// NewServeBridge creates a new serve bridge.
func NewServeBridge() *ServeBridge {
	return &ServeBridge{
		serverTools: map[ServerType]string{
			// Default mappings (from tool-config.yml)
			ServerStaticSite:  "static-site",
			ServerMkDocsLive:  "mkdocs-live",
			ServerStructurizr: "structurizr",
		},
	}
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

// SetServerToolMapping sets a custom server-to-tool mapping.
// This allows overriding the default tool ID for a server type.
func (b *ServeBridge) SetServerToolMapping(serverType ServerType, toolID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.serverTools[serverType] = toolID
}

// GetServerToolID returns the tool ID mapped to a server type.
func (b *ServeBridge) GetServerToolID(serverType ServerType) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.serverTools[serverType]
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

// GetServer returns the server function for a server type from tool-config.yml.
func (b *ServeBridge) GetServer(serverType ServerType) ServeFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	executor := b.getExecutor()

	if registry != nil && executor != nil {
		toolID := b.serverTools[serverType]
		if tool, ok := registry.Get(toolID); ok {
			return b.createToolServeFuncWithExecutor(tool, executor)
		}
	}

	return nil
}

// HasServer checks if a server is available for the given type.
func (b *ServeBridge) HasServer(serverType ServerType) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry != nil {
		toolID := b.serverTools[serverType]
		if _, ok := registry.Get(toolID); ok {
			return true
		}
	}

	return false
}

// GetAllServerTypes returns all available server types.
func (b *ServeBridge) GetAllServerTypes() []ServerType {
	b.mu.RLock()
	defer b.mu.RUnlock()

	serverSet := make(map[ServerType]bool)

	// Add servers available via tool registry
	registry := b.getRegistry()
	if registry != nil {
		for st, toolID := range b.serverTools {
			if _, ok := registry.Get(toolID); ok {
				serverSet[st] = true
			}
		}
	}

	result := make([]ServerType, 0, len(serverSet))
	for st := range serverSet {
		result = append(result, st)
	}
	return result
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
