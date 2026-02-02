package tool

import (
	"fmt"
	"os"
	"strings"
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

// ToolVerifier is a function that checks if a tool is available.
// Used for testing to mock system tool verification.
type ToolVerifier func(tool *ToolDefinition) bool

// DefaultRegistry is the thread-safe in-memory implementation of Registry.
type DefaultRegistry struct {
	mu          sync.RWMutex
	tools       map[string]*ToolDefinition
	bindings    map[string]ToolBinding // canonical name -> binding mode
	verifier    ToolVerifier           // custom verifier for tests (nil = use default)
	verifyCache map[string]bool        // cache for tool availability checks
}

// NewRegistry creates a new tool registry.
func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		tools:       make(map[string]*ToolDefinition),
		bindings:    make(map[string]ToolBinding),
		verifyCache: make(map[string]bool),
	}
}

// SetVerifier sets a custom tool verifier (for testing).
// Pass nil to restore default verification behavior.
func (r *DefaultRegistry) SetVerifier(v ToolVerifier) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.verifier = v
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

	return r.RegisterFromConfig(&config)
}

// RegisterFromConfig registers all tools from a ToolConfig.
// Tools are stored with CANONICAL IDs (no type suffix) in tool.ID.
// Registry stores dual keys for lookup convenience:
//   - Primary key: canonical name (e.g., "npm-build")
//   - Alias key: name with type suffix (e.g., "npm-build:system")
//
// Both keys point to the same tool, and tool.ID is always canonical.
func (r *DefaultRegistry) RegisterFromConfig(config *ToolConfig) error {
	if config == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Register system-tools with canonical ID (no suffix)
	for id, tool := range config.SystemTools {
		tool.ID = id // Canonical - NO suffix
		if tool.Type == "" {
			tool.Type = ToolTypeSystem
		}
		if err := tool.Validate(); err != nil {
			return fmt.Errorf("registering system-tool %s: %w", id, err)
		}
		// Dual-key storage: both canonical and suffixed keys point to same tool
		r.tools[id] = tool            // Primary key: "npm-build"
		r.tools[id+":system"] = tool  // Alias key for explicit lookup
	}

	// Register container-tools with canonical ID (no suffix)
	for id, tool := range config.ContainerTools {
		tool.ID = id // Canonical - NO suffix
		if tool.Type == "" {
			tool.Type = ToolTypeContainer
		}
		if err := tool.Validate(); err != nil {
			return fmt.Errorf("registering container-tool %s: %w", id, err)
		}
		// Dual-key storage: both canonical and suffixed keys point to same tool
		r.tools[id] = tool               // Primary key: "npm-build"
		r.tools[id+":container"] = tool  // Alias key for explicit lookup
	}

	// Store bindings for canonical resolution
	r.bindings = config.ToolBindings

	return nil
}

// Get returns a tool by ID or canonical name.
// If toolID contains ":system" or ":container" suffix, does exact lookup.
// Otherwise, does canonical lookup with binding rules (auto/system/container),
// with fallback to direct ID lookup for backward compatibility.
func (r *DefaultRegistry) Get(toolID string) (*ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Exact lookup if ID has type suffix
	if strings.HasSuffix(toolID, ":system") || strings.HasSuffix(toolID, ":container") {
		tool, ok := r.tools[toolID]
		if !ok {
			return nil, false
		}
		return tool.Clone(), true
	}

	// Canonical lookup with binding rules
	if tool, ok := r.getCanonicalUnlocked(toolID); ok {
		return tool, true
	}

	// Fallback: direct ID lookup (backward compatibility for tests/legacy)
	if tool, ok := r.tools[toolID]; ok {
		return tool.Clone(), true
	}

	return nil, false
}

// getCanonicalUnlocked does canonical lookup. Must be called with lock held.
func (r *DefaultRegistry) getCanonicalUnlocked(canonicalName string) (*ToolDefinition, bool) {
	binding := r.bindings[canonicalName]
	if binding == "" {
		binding = ToolBindingAuto
	}

	systemID := canonicalName + ":system"
	containerID := canonicalName + ":container"

	debug := os.Getenv("DEBUG_TOOL_RESOLVE") != ""
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] getCanonical(%s) binding=%s\n", canonicalName, binding)
	}

	switch binding {
	case ToolBindingSystem:
		if tool, ok := r.tools[systemID]; ok {
			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] -> using system tool %s\n", systemID)
			}
			return tool.Clone(), true
		}
	case ToolBindingContainer:
		if tool, ok := r.tools[containerID]; ok {
			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] -> using container tool %s\n", containerID)
			}
			return tool.Clone(), true
		}
	case ToolBindingAuto:
		// Try system first - check if available
		if tool, ok := r.tools[systemID]; ok {
			avail := r.isToolAvailableUnlocked(tool)
			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] -> system %s found, available=%v\n", systemID, avail)
			}
			if avail {
				return tool.Clone(), true
			}
		} else if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] -> system %s not found in registry\n", systemID)
		}
		// Fall back to container
		if tool, ok := r.tools[containerID]; ok {
			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] -> falling back to container %s\n", containerID)
			}
			return tool.Clone(), true
		} else if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] -> container %s not found in registry\n", containerID)
		}
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] -> no tool found for %s\n", canonicalName)
	}
	return nil, false
}

// isToolAvailableUnlocked checks if a system tool and all its requirements are available.
// Must be called with lock held. Results are cached for the lifetime of the registry.
func (r *DefaultRegistry) isToolAvailableUnlocked(tool *ToolDefinition) bool {
	if tool == nil || tool.Type != ToolTypeSystem {
		return false
	}

	// Check cache first
	if cached, ok := r.verifyCache[tool.ID]; ok {
		return cached
	}

	// Use custom verifier if set (for testing), otherwise use default
	var available bool
	if r.verifier != nil {
		available = r.verifier(tool)
	} else {
		result := VerifyToolDefinition(tool)
		available = result.Available
	}
	if !available {
		r.verifyCache[tool.ID] = false
		return false
	}

	// Verify all requirements recursively
	for _, reqName := range tool.Requirements {
		// Try to find requirement as system tool
		reqID := reqName + ":system"
		reqTool, exists := r.tools[reqID]
		if !exists {
			// Requirement not found as system tool
			r.verifyCache[tool.ID] = false
			return false
		}
		if !r.isToolAvailableUnlocked(reqTool) {
			r.verifyCache[tool.ID] = false
			return false
		}
	}

	r.verifyCache[tool.ID] = true
	return true
}

// GetAll returns a copy of all registered tools by their canonical ID.
// Only returns tools keyed by their canonical name (no :system/:container aliases).
func (r *DefaultRegistry) GetAll() map[string]*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ToolDefinition)
	seen := make(map[*ToolDefinition]bool)

	for id, tool := range r.tools {
		// Skip alias keys (those with :system or :container suffix)
		if strings.HasSuffix(id, ":system") || strings.HasSuffix(id, ":container") {
			continue
		}
		// Deduplicate by tool pointer (in case same tool is registered under multiple canonical names)
		if seen[tool] {
			continue
		}
		seen[tool] = true
		result[id] = tool.Clone()
	}
	return result
}

// ListByType returns all tools of the specified type.
// Only returns unique tools (no duplicates from alias keys).
func (r *DefaultRegistry) ListByType(toolType ToolType) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ToolDefinition
	seen := make(map[*ToolDefinition]bool)
	for id, tool := range r.tools {
		// Skip alias keys
		if strings.HasSuffix(id, ":system") || strings.HasSuffix(id, ":container") {
			continue
		}
		if tool.Type == toolType && !seen[tool] {
			seen[tool] = true
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
// Runs verifications in parallel for faster startup.
func (r *DefaultRegistry) VerifyAll(toolIDs []string) []VerifyResult {
	results := make([]VerifyResult, len(toolIDs))

	// Run verifications in parallel
	var wg sync.WaitGroup
	wg.Add(len(toolIDs))

	for i, toolID := range toolIDs {
		go func(idx int, id string) {
			defer wg.Done()
			results[idx] = r.VerifyTool(id)
		}(i, toolID)
	}

	wg.Wait()
	return results
}


// GetMissingTools returns the IDs of tools that aren't available.
// Tools that are skipped due to platform incompatibility are not considered missing.
func (r *DefaultRegistry) GetMissingTools(toolIDs []string) []string {
	var missing []string
	for _, toolID := range toolIDs {
		result := r.VerifyTool(toolID)
		// Skip platform-incompatible tools - they're not "missing", just not applicable
		if result.Skipped {
			continue
		}
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
