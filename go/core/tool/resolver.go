package tool

import (
	"fmt"
	"os"
	"runtime"
	"sync"
)

// Resolver resolves tool assignments for component types and operations.
// It implements layered configuration: CLI > Environment > Project > Defaults.
type Resolver interface {
	// Resolve returns the tool for a component type and operation.
	Resolve(componentType string, operation OperationType) (*ToolDefinition, error)

	// ResolveAll returns all tools for a component type.
	ResolveAll(componentType string) map[OperationType]*ToolDefinition

	// ResolveMultiple returns all tools for operations that support multiple tools.
	// For example, a component might have multiple linters or scanners.
	ResolveMultiple(componentType string, operation OperationType) ([]*ToolDefinition, error)

	// SetOverride sets a CLI-level override for an operation.
	SetOverride(componentType string, operation OperationType, toolID string)

	// ClearOverrides removes all CLI-level overrides.
	ClearOverrides()

	// SetEnvironment sets the current environment (e.g., "ci", "local").
	// This affects which environment-specific overrides are applied.
	SetEnvironment(env string)
}

// DefaultResolver implements layered configuration resolution.
// Resolution order: CLI overrides > Environment config > Project config > Defaults.
type DefaultResolver struct {
	mu sync.RWMutex

	registry Registry

	// Configuration layers (lowest to highest priority)
	defaults      map[string]*ToolAssignment // Built-in defaults
	projectConfig map[string]*ToolAssignment // From project's tool-config.yml
	envConfigs    map[string]map[string]*ToolAssignment // Environment-specific configs
	cliOverrides  map[string]map[OperationType]string   // CLI flag overrides

	// Current environment (e.g., "ci", "local")
	currentEnv string
}

// NewResolver creates a new tool resolver.
func NewResolver(registry Registry) *DefaultResolver {
	return &DefaultResolver{
		registry:      registry,
		defaults:      make(map[string]*ToolAssignment),
		projectConfig: make(map[string]*ToolAssignment),
		envConfigs:    make(map[string]map[string]*ToolAssignment),
		cliOverrides:  make(map[string]map[OperationType]string),
	}
}

// LoadDefaults loads default tool assignments.
func (r *DefaultResolver) LoadDefaults(assignments map[string]*ToolAssignment) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for compType, assignment := range assignments {
		r.defaults[compType] = assignment
	}
}

// LoadProjectConfig loads project-level tool assignments.
func (r *DefaultResolver) LoadProjectConfig(assignments map[string]*ToolAssignment) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for compType, assignment := range assignments {
		r.projectConfig[compType] = assignment
	}
}

// LoadEnvironmentConfig loads environment-specific tool assignments.
func (r *DefaultResolver) LoadEnvironmentConfig(envName string, assignments map[string]*ToolAssignment) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.envConfigs[envName] == nil {
		r.envConfigs[envName] = make(map[string]*ToolAssignment)
	}
	for compType, assignment := range assignments {
		r.envConfigs[envName][compType] = assignment
	}
}

// LoadFromConfig loads all configuration layers from a ToolConfig.
func (r *DefaultResolver) LoadFromConfig(config *ToolConfig, isDefaults bool) {
	if config == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Load component-tools to appropriate layer
	target := r.projectConfig
	if isDefaults {
		target = r.defaults
	}

	for compType, assignment := range config.ComponentTools {
		target[compType] = assignment
	}

	// Load environment configs
	for envName, envConfig := range config.Environments {
		if r.envConfigs[envName] == nil {
			r.envConfigs[envName] = make(map[string]*ToolAssignment)
		}
		for compType, assignment := range envConfig.ComponentTools {
			r.envConfigs[envName][compType] = assignment
		}
	}
}

// SetEnvironment sets the current environment name.
// This determines which environment-specific overrides are applied.
func (r *DefaultResolver) SetEnvironment(env string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentEnv = env
}

// DetectEnvironment auto-detects the environment from common CI environment variables
// and platform characteristics.
func (r *DefaultResolver) DetectEnvironment() string {
	// Check for common CI environment variables (highest priority)
	ciEnvVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_URL",
		"CIRCLECI",
		"TRAVIS",
		"AZURE_PIPELINES",
	}

	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return "ci"
		}
	}

	// Check for Windows (npm has PATH issues, use containers)
	if runtime.GOOS == "windows" {
		return "windows"
	}

	return "local"
}

// SetOverride sets a CLI-level override for a specific component type and operation.
func (r *DefaultResolver) SetOverride(componentType string, operation OperationType, toolID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cliOverrides[componentType] == nil {
		r.cliOverrides[componentType] = make(map[OperationType]string)
	}
	r.cliOverrides[componentType][operation] = toolID
}

// ClearOverrides removes all CLI-level overrides.
func (r *DefaultResolver) ClearOverrides() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cliOverrides = make(map[string]map[OperationType]string)
}

// Resolve returns the tool for a component type and operation.
// Resolution order: CLI > Environment > Project > Defaults.
func (r *DefaultResolver) Resolve(componentType string, operation OperationType) (*ToolDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	toolID := r.resolveToolID(componentType, operation)
	if toolID == "" {
		return nil, fmt.Errorf("no tool configured for %s/%s", componentType, operation)
	}

	tool, ok := r.registry.Get(toolID)
	if !ok {
		if dr, ok := r.registry.(*DefaultRegistry); ok {
			mode := dr.GetExecutorMode()
			if mode != ExecutorModeAuto {
				return nil, fmt.Errorf(
					"tool %q not found for %s/%s (executor-mode=%s; tool may lack a %s variant)",
					toolID, componentType, operation, mode, mode)
			}
		}
		return nil, fmt.Errorf("tool %q not found in registry (for %s/%s)", toolID, componentType, operation)
	}

	return tool, nil
}

// resolveToolID resolves the tool ID through the configuration layers.
// Must be called with read lock held.
func (r *DefaultResolver) resolveToolID(componentType string, operation OperationType) string {
	// 1. Check CLI overrides (highest priority)
	if overrides, ok := r.cliOverrides[componentType]; ok {
		if toolID, ok := overrides[operation]; ok && toolID != "" {
			return toolID
		}
	}

	// 2. Check environment config
	if r.currentEnv != "" {
		if envConfig, ok := r.envConfigs[r.currentEnv]; ok {
			if assignment, ok := envConfig[componentType]; ok {
				if toolID := assignment.GetToolID(operation); toolID != "" {
					return toolID
				}
			}
		}
	}

	// 3. Check project config
	if assignment, ok := r.projectConfig[componentType]; ok {
		if toolID := assignment.GetToolID(operation); toolID != "" {
			return toolID
		}
	}

	// 4. Check defaults (lowest priority)
	if assignment, ok := r.defaults[componentType]; ok {
		return assignment.GetToolID(operation)
	}

	return ""
}

// ResolveAll returns all tools for a component type across all operations.
func (r *DefaultResolver) ResolveAll(componentType string) map[OperationType]*ToolDefinition {
	result := make(map[OperationType]*ToolDefinition)

	for _, op := range AllOperations() {
		tool, err := r.Resolve(componentType, op)
		if err == nil && tool != nil {
			result[op] = tool
		}
	}

	return result
}

// ResolveMultiple returns all tools for operations that support multiple tools.
func (r *DefaultResolver) ResolveMultiple(componentType string, operation OperationType) ([]*ToolDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	toolIDs := r.resolveToolIDs(componentType, operation)
	if len(toolIDs) == 0 {
		return nil, fmt.Errorf("no tools configured for %s/%s", componentType, operation)
	}

	var tools []*ToolDefinition
	for _, toolID := range toolIDs {
		tool, ok := r.registry.Get(toolID)
		if !ok {
			if dr, ok := r.registry.(*DefaultRegistry); ok {
				mode := dr.GetExecutorMode()
				if mode != ExecutorModeAuto {
					return nil, fmt.Errorf(
						"tool %q not found (executor-mode=%s; tool may lack a %s variant)",
						toolID, mode, mode)
				}
			}
			return nil, fmt.Errorf("tool %q not found in registry", toolID)
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// resolveToolIDs resolves all tool IDs for operations that support multiple tools.
// Must be called with read lock held.
func (r *DefaultResolver) resolveToolIDs(componentType string, operation OperationType) []string {
	// For multi-tool operations, check each layer for multiple tools
	// Single CLI override replaces all
	if overrides, ok := r.cliOverrides[componentType]; ok {
		if toolID, ok := overrides[operation]; ok && toolID != "" {
			return []string{toolID}
		}
	}

	// Check environment config
	if r.currentEnv != "" {
		if envConfig, ok := r.envConfigs[r.currentEnv]; ok {
			if assignment, ok := envConfig[componentType]; ok {
				if toolIDs := assignment.GetToolIDs(operation); len(toolIDs) > 0 {
					return toolIDs
				}
			}
		}
	}

	// Check project config
	if assignment, ok := r.projectConfig[componentType]; ok {
		if toolIDs := assignment.GetToolIDs(operation); len(toolIDs) > 0 {
			return toolIDs
		}
	}

	// Check defaults
	if assignment, ok := r.defaults[componentType]; ok {
		return assignment.GetToolIDs(operation)
	}

	return nil
}

// HasTool checks if there's a tool configured for the given component type and operation.
func (r *DefaultResolver) HasTool(componentType string, operation OperationType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resolveToolID(componentType, operation) != ""
}

// ResolveToolID returns the tool ID for a component type and operation without
// requiring the tool to exist in the registry. This is useful for checking if
// the tool ID matches a native handler before falling back to registry lookup.
func (r *DefaultResolver) ResolveToolID(componentType string, operation OperationType) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resolveToolID(componentType, operation)
}

// ListConfiguredComponents returns all component types that have any tool configured.
func (r *DefaultResolver) ListConfiguredComponents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)

	for compType := range r.defaults {
		seen[compType] = true
	}
	for compType := range r.projectConfig {
		seen[compType] = true
	}
	for _, envConfig := range r.envConfigs {
		for compType := range envConfig {
			seen[compType] = true
		}
	}

	result := make([]string, 0, len(seen))
	for compType := range seen {
		result = append(result, compType)
	}
	return result
}
