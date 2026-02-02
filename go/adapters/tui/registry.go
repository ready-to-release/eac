package tui

import (
	"strings"
	"sync"
)

// registryImpl holds command-to-TUI bindings.
// Bindings are evaluated at lookup time with a defined resolution order.
type registryImpl struct {
	mu         sync.RWMutex
	exact      map[string]ConsoleFactory // Exact matches: "build", "test"
	prefix     map[string]ConsoleFactory // Prefix matches: "work" -> "work *"
	defaultTUI ConsoleFactory            // Fallback for unmatched commands
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *registryImpl {
	return &registryImpl{
		exact:  make(map[string]ConsoleFactory),
		prefix: make(map[string]ConsoleFactory),
	}
}

// Register adds a TUI binding to the registry.
// Pattern formats:
//   - "build"       - Exact match for "build" command
//   - "work create" - Exact match for "work create" subcommand
//   - "work *"      - Prefix match for all "work" subcommands
//   - "*"           - Default fallback (lowest priority)
func (r *registryImpl) Register(pattern string, factory ConsoleFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if pattern == "*" {
		r.defaultTUI = factory
		return
	}

	if strings.HasSuffix(pattern, " *") {
		// Prefix binding: "work *" -> matches "work create", "work merge", etc.
		prefix := strings.TrimSuffix(pattern, " *")
		r.prefix[prefix] = factory
		return
	}

	// Exact binding
	r.exact[pattern] = factory
}

// NewForCommand creates a TUI for the given command path.
// commandPath is space-separated: "build", "work create", "test core".
//
// Resolution order:
//  1. Exact match: "work create"
//  2. Prefix match: "work" (from "work *") - only for subcommands
//  3. Parent command: "work" (if "work create" has no binding)
//  4. Default TUI: "*"
func (r *registryImpl) NewForCommand(commandPath string, config Config) Console {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Try exact match first
	if factory, ok := r.exact[commandPath]; ok {
		return factory(config)
	}

	// Parse command path into parts
	parts := strings.Fields(commandPath)

	// 2. Try prefix match for subcommands (only if there are multiple parts)
	if len(parts) > 1 {
		// Check if the parent has a prefix binding (e.g., "work *" for "work create")
		parent := parts[0]
		if factory, ok := r.prefix[parent]; ok {
			return factory(config)
		}
	}

	// 3. Try hierarchical fallback (parent commands)
	for i := len(parts) - 1; i >= 0; i-- {
		prefix := strings.Join(parts[:i+1], " ")

		// Skip the full path (already tried in step 1)
		if prefix == commandPath {
			continue
		}

		// Check for prefix binding at this level
		if factory, ok := r.prefix[prefix]; ok {
			return factory(config)
		}

		// Check for exact match at this level
		if factory, ok := r.exact[prefix]; ok {
			return factory(config)
		}
	}

	// 4. Use default TUI
	if r.defaultTUI != nil {
		return r.defaultTUI(config)
	}

	// 5. Panic if no default registered (programming error)
	panic("no TUI registered for command: " + commandPath + " (and no default)")
}

// MustHaveDefault panics if no default TUI is registered.
// Called during application startup to fail fast.
func (r *registryImpl) MustHaveDefault() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.defaultTUI == nil {
		panic("no default TUI registered - call Register(\"*\", factory) during init")
	}
}

// ListBindings returns all registered binding patterns (for debugging/help).
func (r *registryImpl) ListBindings() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var bindings []string
	for pattern := range r.exact {
		bindings = append(bindings, pattern)
	}
	for pattern := range r.prefix {
		bindings = append(bindings, pattern+" *")
	}
	if r.defaultTUI != nil {
		bindings = append(bindings, "*")
	}
	return bindings
}

// Global registry for static registration via init() functions.
var globalRegistry = NewRegistry()

// Register adds a TUI binding to the global registry.
// This is typically called from init() functions in TUI implementation packages.
func Register(pattern string, factory ConsoleFactory) {
	globalRegistry.Register(pattern, factory)
}

// NewForCommand creates a TUI using the global registry.
func NewForCommand(commandPath string, config Config) Console {
	return globalRegistry.NewForCommand(commandPath, config)
}

// MustHaveDefault verifies the global registry has a default TUI.
func MustHaveDefault() {
	globalRegistry.MustHaveDefault()
}

// ListBindings returns all bindings in the global registry.
func ListBindings() []string {
	return globalRegistry.ListBindings()
}
