package tool

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Global registry instance for verification throughout the codebase.
var (
	globalRegistry   *DefaultRegistry
	globalRegistryMu sync.RWMutex
)

// GlobalRegistry returns the global tool registry instance.
// Creates a new empty registry if one hasn't been set via SetGlobalRegistry.
func GlobalRegistry() *DefaultRegistry {
	globalRegistryMu.RLock()
	r := globalRegistry
	globalRegistryMu.RUnlock()

	if r != nil {
		return r
	}

	// Double-check with write lock
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()

	if globalRegistry == nil {
		globalRegistry = NewRegistry()
	}
	return globalRegistry
}

// SetGlobalRegistry allows overriding the global registry (for testing or initialization).
func SetGlobalRegistry(r *DefaultRegistry) {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()
	globalRegistry = r
}

// Registry stores and retrieves tool definitions.
type Registry interface {
	// Register adds a tool to the registry.
	Register(tool *ToolDefinition) error

	// RegisterFromYAML loads and registers tools from a YAML file.
	RegisterFromYAML(path string) error

	// Get returns a tool by ID, or nil and false if not found.
	Get(toolID string) (*ToolDefinition, bool)

	// GetAll returns all registered tools.
	GetAll() map[string]*ToolDefinition

	// ListByType returns all tools of a given type.
	ListByType(toolType ToolType) []*ToolDefinition

	// Validate checks if a tool can be executed (requirements met).
	Validate(toolID string) error

	// ValidateAll validates all registered tools.
	ValidateAll() []error
}

// DefaultRegistry is the thread-safe in-memory implementation of Registry.
type DefaultRegistry struct {
	mu    sync.RWMutex
	tools map[string]*ToolDefinition
}

// NewRegistry creates a new tool registry.
func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		tools: make(map[string]*ToolDefinition),
	}
}

// Register adds a tool to the registry.
// Returns an error if the tool definition is invalid.
func (r *DefaultRegistry) Register(tool *ToolDefinition) error {
	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	if err := tool.Validate(); err != nil {
		return fmt.Errorf("invalid tool definition: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[tool.ID] = tool
	return nil
}

// RegisterFromYAML loads tools from a YAML file and registers them.
// The file should contain a ToolConfig structure.
func (r *DefaultRegistry) RegisterFromYAML(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading tool config %s: %w", path, err)
	}

	var config ToolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parsing tool config %s: %w", path, err)
	}

	// Register all tools from the config
	for id, tool := range config.Tools {
		if tool.ID == "" {
			tool.ID = id // Use map key as ID if not set
		}
		if err := r.Register(tool); err != nil {
			return fmt.Errorf("registering tool %s from %s: %w", id, path, err)
		}
	}

	return nil
}

// RegisterFromConfig registers all tools from a ToolConfig.
func (r *DefaultRegistry) RegisterFromConfig(config *ToolConfig) error {
	if config == nil {
		return nil
	}

	for id, tool := range config.Tools {
		if tool.ID == "" {
			tool.ID = id
		}
		if err := r.Register(tool); err != nil {
			return fmt.Errorf("registering tool %s: %w", id, err)
		}
	}

	return nil
}

// Get returns a tool by ID.
// Returns the tool and true if found, or nil and false if not found.
func (r *DefaultRegistry) Get(toolID string) (*ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[toolID]
	if !ok {
		return nil, false
	}

	// Return a clone to prevent external modification
	return tool.Clone(), true
}

// GetAll returns a copy of all registered tools.
func (r *DefaultRegistry) GetAll() map[string]*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ToolDefinition, len(r.tools))
	for id, tool := range r.tools {
		result[id] = tool.Clone()
	}
	return result
}

// ListByType returns all tools of the specified type.
func (r *DefaultRegistry) ListByType(toolType ToolType) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ToolDefinition
	for _, tool := range r.tools {
		if tool.Type == toolType {
			result = append(result, tool.Clone())
		}
	}
	return result
}

// Validate checks if a tool can be executed.
// This includes checking that all requirements are satisfied.
func (r *DefaultRegistry) Validate(toolID string) error {
	tool, ok := r.Get(toolID)
	if !ok {
		return fmt.Errorf("tool not found: %s", toolID)
	}

	// Check basic definition validity
	if err := tool.Validate(); err != nil {
		return err
	}

	// Requirements are validated during tool verification via VerifyTool()
	// which uses the tool registry to recursively check all requirements.

	return nil
}

// ValidateAll validates all registered tools.
func (r *DefaultRegistry) ValidateAll() []error {
	r.mu.RLock()
	toolIDs := make([]string, 0, len(r.tools))
	for id := range r.tools {
		toolIDs = append(toolIDs, id)
	}
	r.mu.RUnlock()

	var errs []error
	for _, id := range toolIDs {
		if err := r.Validate(id); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Count returns the number of registered tools.
func (r *DefaultRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// Clear removes all tools from the registry.
func (r *DefaultRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = make(map[string]*ToolDefinition)
}

// Has returns true if a tool with the given ID is registered.
func (r *DefaultRegistry) Has(toolID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[toolID]
	return ok
}

// Remove removes a tool from the registry.
// Returns true if the tool was removed, false if it didn't exist.
func (r *DefaultRegistry) Remove(toolID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tools[toolID]; !ok {
		return false
	}
	delete(r.tools, toolID)
	return true
}

// VerifyTool checks if a tool is available and meets version requirements.
func (r *DefaultRegistry) VerifyTool(toolID string) VerifyResult {
	tool, ok := r.Get(toolID)
	if !ok {
		return VerifyResult{
			ToolID:    toolID,
			Available: false,
			Error:     fmt.Errorf("unknown tool: %s", toolID),
		}
	}
	return VerifyToolDefinition(tool)
}

// VerifyAll checks multiple tools and returns results for each.
func (r *DefaultRegistry) VerifyAll(toolIDs []string) []VerifyResult {
	results := make([]VerifyResult, len(toolIDs))
	for i, toolID := range toolIDs {
		results[i] = r.VerifyTool(toolID)
	}
	return results
}

// GetMissingTools returns the IDs of tools that aren't available.
func (r *DefaultRegistry) GetMissingTools(toolIDs []string) []string {
	var missing []string
	for _, toolID := range toolIDs {
		result := r.VerifyTool(toolID)
		if !result.Available {
			missing = append(missing, toolID)
		}
	}
	return missing
}

// IsAvailable checks if a tool is available (convenience function).
func (r *DefaultRegistry) IsAvailable(toolID string) bool {
	return r.VerifyTool(toolID).Available
}
