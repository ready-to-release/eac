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
	Component          string   // Specific component to build (empty = all components, for component-level parallelism)
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

	// Requirements returns system dependencies required by this handler.
	// Used for early validation (e.g., ["go", "docker"]).
	Requirements() []string

	// ValidateModule checks if a module's configuration is valid for a specific component.
	// Returns nil if valid, or an error describing the problem.
	// Called before build starts for early failure.
	ValidateModule(module *modules.ModuleContract, workspaceRoot, component string) error
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

// ComponentHandler pairs a component name with its build handler.
type ComponentHandler struct {
	Component string
	Handler   Handler
}

// GetHandlersForModule returns all handlers for a module's buildable components.
// Returns a slice of ComponentHandler pairs, one for each component that has a builder.
// Handler selection follows this priority:
//  1. Module-level handler override (build.handler in module config) - applies to primary component only
//  2. Component-type builders (from component-types.yml for each enabled component)
//
// Returns empty slice if module has no buildable components.
func GetHandlersForModule(module *modules.ModuleContract) []ComponentHandler {
	mu.RLock()
	defer mu.RUnlock()

	if module == nil {
		return nil
	}

	var result []ComponentHandler

	// Priority 1: Check for per-module handler override (applies to primary component)
	if module.GetBuildHandler() != "" {
		handlerName := module.GetBuildHandler()
		if h, ok := handlers[handlerName]; ok {
			log.Debugf("Using per-module handler override: %s -> %s", module.Moniker, handlerName)
			result = append(result, ComponentHandler{
				Component: "override",
				Handler:   h,
			})
			return result // Override takes precedence, skip component-based handlers
		}
		log.Warnf("Module %s specifies unknown handler %q, falling back to component-based selection",
			module.Moniker, handlerName)
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	// Priority 2: Find builders from all component types
	// Each component gets its own handler entry for component-level parallelism
	for _, compName := range module.GetEnabledComponents() {
		// Get the component type (may differ from name for named components)
		compTypeName := module.Components.GetComponentType(compName)
		compType := cfg.ComponentTypes.Get(compTypeName)
		if compType != nil && compType.HasBuilder() {
			builderName := compType.Builder
			if h, ok := handlers[builderName]; ok {
				log.Debugf("Adding component handler: %s (component: %s, type: %s) -> %s",
					module.Moniker, compName, compTypeName, builderName)
				result = append(result, ComponentHandler{
					Component: compName,
					Handler:   h,
				})
			}
		}
	}

	if len(result) == 0 {
		log.Debugf("No build handlers for module %s (no buildable components)", module.Moniker)
	}

	return result
}

// GetHandlerByBuilder returns the handler for a specific builder name.
func GetHandlerByBuilder(builderName string) Handler {
	mu.RLock()
	defer mu.RUnlock()
	return handlers[builderName]
}
