// Package services provides centralized service initialization for EAC commands.
// It implements the SimpleServicesPort interface from the core contracts.
package services

import (
	"fmt"
	"path/filepath"
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workspace"
)

// Compile-time interface check.
var _ core.SimpleServicesPort = (*Services)(nil)

// Services provides core services for EAC commands.
// It implements the SimpleServicesPort interface.
type Services struct {
	workspaceRoot string
	configRoot    string
	config        core.ConfigPort
	modules       core.ModuleRegistryPort
	tools         core.ToolRegistryPort

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
func New(opts core.SimpleServicesOptions) (*Services, error) {
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
func (s *Services) Config() core.ConfigPort {
	return s.config
}

// Modules returns the module registry port.
func (s *Services) Modules() core.ModuleRegistryPort {
	return s.modules
}

// Tools returns the tool registry port.
// Returns nil if InitTools was false in options.
func (s *Services) Tools() core.ToolRegistryPort {
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
