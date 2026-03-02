package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	tools "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// ToolLoader provides namespace-based lazy loading of tool definitions.
// Only bootstrap tools are loaded at startup; other namespaces load on-demand.
type ToolLoader struct {
	configRoot string

	config     *tools.ToolsConfig
	namespaces map[tools.Namespace][]string
	loaded     map[tools.Namespace]bool
	toolDefs   map[string]*tools.ToolDefinition

	mu sync.RWMutex
}

// NewToolLoader creates a ToolLoader that loads tools from the eac-tools contract.
func NewToolLoader(configRoot string) (*ToolLoader, error) {
	l := &ToolLoader{
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
	// Load contract defaults from embedded filesystem
	defaultData, err := tools.FS.ReadFile(tools.DefaultPath(ToolConfigFileName))
	if err != nil {
		return fmt.Errorf("reading embedded tool config defaults: %w", err)
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

	base.Namespaces = mergeMap(base.Namespaces, override.Namespaces)
	base.SystemTools = mergeMap(base.SystemTools, override.SystemTools)
	base.ContainerTools = mergeMap(base.ContainerTools, override.ContainerTools)
	base.Bindings = mergeMap(base.Bindings, override.Bindings)
	base.ComponentTools = mergeMap(base.ComponentTools, override.ComponentTools)
	base.Environments = mergeMap(base.Environments, override.Environments)
	base.Caches = mergeMap(base.Caches, override.Caches)
}

// mergeMap copies all entries from override into base, initializing base if nil.
func mergeMap[K comparable, V any](base, override map[K]V) map[K]V {
	if len(override) == 0 {
		return base
	}
	if base == nil {
		base = make(map[K]V, len(override))
	}
	for k, v := range override {
		base[k] = v
	}
	return base
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
// Tools that are listed in the namespace but not defined in config are skipped
// (namespaces may reference optional tools).
func (l *ToolLoader) EnsureNamespace(ns tools.Namespace) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loaded[ns] {
		return nil
	}

	toolIDs := l.namespaces[ns]
	for _, id := range toolIDs {
		_ = l.loadToolUnlocked(id) // Optional tools may not exist
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
		tool.IDValue = id
		if tool.TypeValue == "" {
			tool.TypeValue = tools.ToolTypeSystem
		}
		l.toolDefs[id] = tool
		return nil
	}

	// Check container tools
	if tool, ok := l.config.ContainerTools[id]; ok {
		tool.IDValue = id
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
