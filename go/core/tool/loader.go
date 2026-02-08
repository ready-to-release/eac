package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/ready-to-release/eac/go/core/paths"

	tools "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// ToolLoader provides namespace-based lazy loading of tool definitions.
// Only bootstrap tools are loaded at startup; other namespaces load on-demand.
type ToolLoader struct {
	repoRoot   string
	configRoot string

	config     *tools.ToolsConfig
	namespaces map[tools.Namespace][]string
	loaded     map[tools.Namespace]bool
	toolDefs   map[string]*tools.ToolDefinition

	mu sync.RWMutex
}

// NewToolLoader creates a ToolLoader that loads tools from the eac-tools contract.
func NewToolLoader(repoRoot, configRoot string) (*ToolLoader, error) {
	l := &ToolLoader{
		repoRoot:   repoRoot,
		configRoot: configRoot,
		namespaces: make(map[tools.Namespace][]string),
		loaded:     make(map[tools.Namespace]bool),
		toolDefs:   make(map[string]*tools.ToolDefinition),
	}

	// Load config from contract defaults
	if err := l.loadConfig(); err != nil {
		return nil, err
	}

	// Always load bootstrap namespace at startup
	if err := l.EnsureNamespace(tools.NSBootstrap); err != nil {
		return nil, fmt.Errorf("loading bootstrap namespace: %w", err)
	}

	return l, nil
}

// loadConfig loads the tools configuration from contract defaults and user overrides.
func (l *ToolLoader) loadConfig() error {
	// Load contract defaults
	defaultPath := filepath.Join(l.repoRoot, "contracts", "core", paths.DefaultsVersion, "schemas", "defaults", ToolConfigFileName)
	defaultData, err := os.ReadFile(defaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tool config defaults not found: %s", defaultPath)
		}
		return fmt.Errorf("reading tool config defaults: %w", err)
	}

	var config tools.ToolsConfig
	if err := yaml.Unmarshal(defaultData, &config); err != nil {
		return fmt.Errorf("parsing tools defaults: %w", err)
	}

	// Load user overrides (optional)
	userPath := filepath.Join(l.configRoot, "tools.yml")
	if userData, err := os.ReadFile(userPath); err == nil {
		var userConfig tools.ToolsConfig
		if err := yaml.Unmarshal(userData, &userConfig); err != nil {
			return fmt.Errorf("parsing user tools config: %w", err)
		}
		mergeToolsConfig(&config, &userConfig)
	}

	l.config = &config
	l.namespaces = config.Namespaces

	return nil
}

// mergeToolsConfig merges user overrides into base config.
func mergeToolsConfig(base, override *tools.ToolsConfig) {
	if override == nil {
		return
	}

	// Merge namespaces (override replaces)
	for ns, toolIDs := range override.Namespaces {
		base.Namespaces[ns] = toolIDs
	}

	// Merge system tools
	for id, tool := range override.SystemTools {
		if tool.IDValue == "" {
			tool.IDValue = id
		}
		if base.SystemTools == nil {
			base.SystemTools = make(map[string]*tools.ToolDefinition)
		}
		base.SystemTools[id] = tool
	}

	// Merge container tools
	for id, tool := range override.ContainerTools {
		if tool.IDValue == "" {
			tool.IDValue = id
		}
		if base.ContainerTools == nil {
			base.ContainerTools = make(map[string]*tools.ToolDefinition)
		}
		base.ContainerTools[id] = tool
	}

	// Merge bindings
	for id, binding := range override.Bindings {
		if base.Bindings == nil {
			base.Bindings = make(map[string]tools.Binding)
		}
		base.Bindings[id] = binding
	}

	// Merge component tools
	for compType, assignment := range override.ComponentTools {
		if base.ComponentTools == nil {
			base.ComponentTools = make(map[string]*tools.ToolAssignment)
		}
		base.ComponentTools[compType] = assignment
	}

	// Merge environments
	for name, env := range override.Environments {
		if base.Environments == nil {
			base.Environments = make(map[string]*tools.EnvironmentConfig)
		}
		base.Environments[name] = env
	}

	// Merge caches
	for name, cache := range override.Caches {
		if base.Caches == nil {
			base.Caches = make(map[string]*tools.CacheConfig)
		}
		base.Caches[name] = cache
	}
}

// GetTool returns a tool definition by ID.
// Returns nil if the tool is not found or not yet loaded.
func (l *ToolLoader) GetTool(id string) (tools.ToolDefPort, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	tool, ok := l.toolDefs[id]
	if !ok {
		return nil, false
	}
	return tool, true
}

// ListTools returns all currently loaded tool IDs.
func (l *ToolLoader) ListTools() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]string, 0, len(l.toolDefs))
	for id := range l.toolDefs {
		result = append(result, id)
	}
	return result
}

// ListToolsByNamespace returns tool IDs within a specific namespace.
func (l *ToolLoader) ListToolsByNamespace(ns tools.Namespace) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.namespaces[ns]
}

// GetBinding returns the binding mode for a tool.
func (l *ToolLoader) GetBinding(toolID string) tools.Binding {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.config != nil && l.config.Bindings != nil {
		if binding, ok := l.config.Bindings[toolID]; ok {
			return binding
		}
	}
	return tools.BindingAuto
}

// GetComponentTools returns tool assignments for a component type.
func (l *ToolLoader) GetComponentTools(componentType string) (tools.ToolConfigAssignmentPort, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.config == nil || l.config.ComponentTools == nil {
		return nil, false
	}
	assignment, ok := l.config.ComponentTools[componentType]
	return assignment, ok
}

// EnsureNamespace loads tools in the given namespace if not already loaded.
func (l *ToolLoader) EnsureNamespace(ns tools.Namespace) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loaded[ns] {
		return nil
	}

	toolIDs := l.namespaces[ns]
	for _, id := range toolIDs {
		if err := l.loadToolUnlocked(id); err != nil {
			// Log but don't fail - tool might just not exist
			continue
		}
	}

	l.loaded[ns] = true
	return nil
}

// loadToolUnlocked loads a single tool definition. Must be called with lock held.
func (l *ToolLoader) loadToolUnlocked(id string) error {
	if l.config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Check system tools first
	if tool, ok := l.config.SystemTools[id]; ok {
		if tool.IDValue == "" {
			tool.IDValue = id
		}
		if tool.TypeValue == "" {
			tool.TypeValue = tools.ToolTypeSystem
		}
		l.toolDefs[id] = tool
		return nil
	}

	// Check container tools
	if tool, ok := l.config.ContainerTools[id]; ok {
		if tool.IDValue == "" {
			tool.IDValue = id
		}
		if tool.TypeValue == "" {
			tool.TypeValue = tools.ToolTypeContainer
		}
		l.toolDefs[id] = tool
		return nil
	}

	return fmt.Errorf("tool not found: %s", id)
}

// IsNamespaceLoaded returns true if the namespace has been loaded.
func (l *ToolLoader) IsNamespaceLoaded(ns tools.Namespace) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loaded[ns]
}

// EnsureNamespaceForComponent loads the namespace required for a component type.
func (l *ToolLoader) EnsureNamespaceForComponent(componentType string) error {
	ns := tools.ComponentToNamespace(componentType)
	return l.EnsureNamespace(ns)
}

// Config returns the underlying tools configuration.
// Use with caution - this exposes internal state.
func (l *ToolLoader) Config() *tools.ToolsConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// LoadAll loads all tools from all namespaces.
// This is useful for validation or when all tools need to be available.
func (l *ToolLoader) LoadAll() error {
	for ns := range l.namespaces {
		if err := l.EnsureNamespace(ns); err != nil {
			return fmt.Errorf("loading namespace %s: %w", ns, err)
		}
	}
	return nil
}

// Verify ToolLoader implements ToolConfigPort.
var _ tools.ToolConfigPort = (*ToolLoader)(nil)
