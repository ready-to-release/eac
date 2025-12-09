// registry.go - Self-registration system for build handlers
package builders

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// BuildOptions contains flags for controlling the build process.
type BuildOptions struct {
	TidyFirst          bool     // Run go mod tidy before building
	Version            string   // Version to inject via ldflags
	DryRun             bool     // Simulate build without actually running it
	RequestedArtifacts []string // Specific artifact IDs to build (empty = default artifacts, "*" = all)
}

// Handler is the interface for build handlers.
// Each handler is responsible for building modules of specific types.
type Handler interface {
	// Name returns the handler identifier (e.g., "go", "mkdocs", "docker")
	Name() string

	// Build executes the build for a module.
	// Returns exit code (0 = success, non-zero = failure).
	Build(module *modules.ModuleContract, workspaceRoot, outputDir string,
		logWriter io.Writer, opts BuildOptions) int

	// ListArtifacts returns artifact paths that would be produced.
	// Paths are relative to the module's output directory.
	ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string

	// Capabilities returns module capabilities this handler supports.
	// Used for dispatch matching (e.g., ["go_module", "cross_compile"]).
	Capabilities() []string

	// Requirements returns system dependencies required by this handler.
	// Used for early validation (e.g., ["go", "docker"]).
	Requirements() []string

	// ValidateModule checks if a module's configuration is valid.
	// Returns nil if valid, or an error describing the problem.
	// Called before build starts for early failure.
	ValidateModule(module *modules.ModuleContract, workspaceRoot string) error
}

var (
	mu       sync.RWMutex
	handlers = make(map[string]Handler)
	log      = logging.C()
)

// RegisterHandler registers a handler for a build dependency.
// Call this from init() in your builder file.
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

// GetHandlerForModule returns the appropriate handler for a module.
// It first checks for a per-module handler override, then finds a handler
// whose capabilities match the module's capabilities from module-types.yml.
func GetHandlerForModule(module *modules.ModuleContract, moduleType string) Handler {
	mu.RLock()
	defer mu.RUnlock()

	// Check for per-module handler override first
	if module != nil && module.GetBuildHandler() != "" {
		handlerName := module.GetBuildHandler()
		if h, ok := handlers[handlerName]; ok {
			return h
		}
	}

	// Get module capabilities from config
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		if h, ok := handlers[""]; ok {
			return h
		}
		return nil
	}

	moduleCapabilities := cfg.ModuleTypes.GetCapabilities(moduleType)

	// Find handler whose capabilities match module capabilities
	for _, h := range handlers {
		if h.Name() == "" {
			continue // Skip no-op handler for now
		}
		handlerCaps := h.Capabilities()
		if matchesCapabilities(moduleCapabilities, handlerCaps) {
			return h
		}
	}

	// Fallback: no-op handler (for types with no matching capabilities)
	if h, ok := handlers[""]; ok {
		return h
	}

	return nil
}

// matchesCapabilities returns true if the module has any capability the handler supports
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

// IsGoModuleType returns true if the module type uses Go tooling (has go_module capability)
func IsGoModuleType(moduleType string) bool {
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		return cfg.ModuleTypes.HasCapability(moduleType, "go_module")
	}
	return false
}
