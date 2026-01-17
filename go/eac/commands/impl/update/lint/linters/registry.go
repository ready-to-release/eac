// registry.go - Self-registration system for lint handlers
package linters

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// LintOptions contains flags for controlling the lint process.
type LintOptions struct {
	Fix    bool   // Auto-fix issues where possible
	Config string // Override config file path
}

// Handler is the interface for lint handlers.
// Each handler is responsible for linting modules of specific types.
type Handler interface {
	// Name returns the handler identifier (e.g., "go", "typescript")
	Name() string

	// Lint executes linting for a module.
	// Returns exit code (0 = success/no issues, non-zero = issues found or error).
	Lint(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts LintOptions) int

	// Capabilities returns module capabilities this handler supports.
	// Used for dispatch matching (e.g., ["go_module"]).
	Capabilities() []string

	// Requirements returns system dependencies required by this handler.
	// Used for early validation (e.g., ["golangci-lint"]).
	Requirements() []string

	// ValidateModule checks if a module's configuration is valid for linting.
	// Returns nil if valid, or an error describing the problem.
	ValidateModule(moduleRoot, workspaceRoot string) error
}

var (
	mu       sync.RWMutex
	handlers = make(map[string]Handler)
	log      = logging.C()
)

// RegisterHandler registers a handler for linting.
// Call this from init() in your handler file.
func RegisterHandler(h Handler) {
	mu.Lock()
	defer mu.Unlock()
	handlers[h.Name()] = h
}

// GetHandler returns the handler for a given name, or nil if not found.
func GetHandler(name string) Handler {
	mu.RLock()
	defer mu.RUnlock()
	return handlers[name]
}

// GetAllHandlers returns all registered handlers.
func GetAllHandlers() map[string]Handler {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]Handler, len(handlers))
	for k, v := range handlers {
		result[k] = v
	}
	return result
}

// GetHandlerForModule returns the appropriate lint handler for a module.
// It finds a handler whose capabilities match the module's capabilities from module-types.yml.
func GetHandlerForModule(moduleType string) Handler {
	mu.RLock()
	defer mu.RUnlock()

	// Get module capabilities from config
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		return nil
	}

	moduleCapabilities := cfg.ModuleTypes.GetCapabilities(moduleType)

	// Find handler whose capabilities match module capabilities
	for _, h := range handlers {
		if h.Name() == "" {
			continue // Skip no-op handler
		}
		handlerCaps := h.Capabilities()
		if matchesCapabilities(moduleCapabilities, handlerCaps) {
			return h
		}
	}

	return nil
}

// matchesCapabilities returns true if the module has any capability the handler supports.
func matchesCapabilities(moduleCapabilities, handlerCapabilities []string) bool {
	for _, mc := range moduleCapabilities {
		for _, hc := range handlerCapabilities {
			if mc == hc {
				return true
			}
		}
	}
	return false
}

// IsGoModuleType returns true if the module type uses Go tooling (has go_module capability).
func IsGoModuleType(moduleType string) bool {
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		return cfg.ModuleTypes.HasCapability(moduleType, "go_module")
	}
	return false
}
