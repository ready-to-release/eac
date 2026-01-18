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

// PackageHandler pairs a package name with its build handler.
type PackageHandler struct {
	Package string
	Handler Handler
}

// GetHandlersForModule returns all handlers for a module's buildable packages.
// Returns a slice of PackageHandler pairs, one for each package that has a builder.
// Handler selection follows this priority:
//  1. Module-level handler override (build.handler in module config) - applies to primary package only
//  2. Package-type builders (from package-types.yml for each enabled package)
//
// Returns empty slice if module has no buildable packages.
func GetHandlersForModule(module *modules.ModuleContract) []PackageHandler {
	mu.RLock()
	defer mu.RUnlock()

	if module == nil {
		return nil
	}

	var result []PackageHandler

	// Priority 1: Check for per-module handler override (applies to primary package)
	if module.GetBuildHandler() != "" {
		handlerName := module.GetBuildHandler()
		if h, ok := handlers[handlerName]; ok {
			log.Debugf("Using per-module handler override: %s -> %s", module.Moniker, handlerName)
			result = append(result, PackageHandler{
				Package: "override",
				Handler: h,
			})
			return result // Override takes precedence, skip package-based handlers
		}
		log.Warnf("Module %s specifies unknown handler %q, falling back to package-based selection",
			module.Moniker, handlerName)
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	// Priority 2: Find builders from all component types
	// Track which builder names we've already added to avoid duplicates
	seenBuilders := make(map[string]bool)

	for _, compName := range module.GetEnabledComponents() {
		// Get the component type (may differ from name for named components)
		compTypeName := module.Components.GetComponentType(compName)
		compType := cfg.ComponentTypes.Get(compTypeName)
		if compType != nil && compType.HasBuilder() {
			builderName := compType.Builder
			// Skip if we've already added this builder (e.g., typescript and javascript both use npm)
			if seenBuilders[builderName] {
				continue
			}
			if h, ok := handlers[builderName]; ok {
				log.Debugf("Adding component-based handler: %s (component: %s, type: %s) -> %s",
					module.Moniker, compName, compTypeName, builderName)
				result = append(result, PackageHandler{
					Package: compName,
					Handler: h,
				})
				seenBuilders[builderName] = true
			}
		}
	}

	if len(result) == 0 {
		log.Debugf("No build handlers for module %s (no buildable packages)", module.Moniker)
	}

	return result
}

// GetHandlerForModule returns the primary handler for a module.
// This is a convenience function that returns the first handler from GetHandlersForModule.
// Use GetHandlersForModule to get all handlers for multi-package modules.
//
// Handler selection follows this priority:
//  1. Module-level handler override (build.handler in module config)
//  2. Package-type builder (from package-types.yml based on first buildable package)
//  3. No-op handler (module has no buildable packages)
//
// The moduleType parameter is deprecated and ignored.
func GetHandlerForModule(module *modules.ModuleContract, moduleType string) Handler {
	pkgHandlers := GetHandlersForModule(module)
	if len(pkgHandlers) == 0 {
		mu.RLock()
		defer mu.RUnlock()
		return handlers[""] // no-op handler
	}
	return pkgHandlers[0].Handler
}

// GetHandlerByBuilder returns the handler for a specific builder name.
func GetHandlerByBuilder(builderName string) Handler {
	mu.RLock()
	defer mu.RUnlock()
	return handlers[builderName]
}
