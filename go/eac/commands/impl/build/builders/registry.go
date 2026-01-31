// Package builders provides build handlers using native handlers and the pluggable tool system.
//
// Native handlers (Go, MkDocs, Docker, etc.) contain specialized build logic.
// The tool system is used as a fallback for simple tools defined in tool-config.yml.
package builders

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/contracts/eac-core-interfaces"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// Handler is the interface for build handlers.
// Each handler is responsible for building modules of specific types.
// Handler accepts interfaces.ModuleContractPort for decoupled module access.
type Handler interface {
	// Name returns the handler identifier (e.g., "go", "mkdocs", "docker")
	Name() string

	// Build executes the build for a module.
	// Returns exit code (0 = success, non-zero = failure).
	Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string,
		logWriter io.Writer, opts BuildOptions) int

	// ListArtifacts returns artifact paths that would be produced.
	// Paths are relative to the module's output directory.
	ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string

	// Requirements returns system dependencies required by this handler.
	// Used for early validation (e.g., ["go", "docker"]).
	Requirements() []string

	// ValidateModule checks if a module's configuration is valid for a specific component.
	// Returns nil if valid, or an error describing the problem.
	ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error

	// IsContainer returns true if this handler runs in a Docker container.
	IsContainer() bool

	// IsHostInstalled returns true if this handler runs using host-installed tools (not containers).
	IsHostInstalled() bool
}

// BuildOptions contains flags for controlling the build process.
type BuildOptions = tool.BuildOptions

// ComponentHandler pairs a component name with its build handler.
type ComponentHandler struct {
	Component string
	Handler   Handler
}

var (
	mu       sync.RWMutex
	handlers = make(map[string]Handler)
)

// RegisterHandler registers a native handler.
// Call this from init() in your builder file.
func RegisterHandler(h Handler) {
	mu.Lock()
	defer mu.Unlock()
	handlers[h.Name()] = h
}

// GetHandler returns the handler for a given name.
// First checks native handlers, then falls back to tool system.
func GetHandler(name string) Handler {
	mu.RLock()
	h := handlers[name]
	mu.RUnlock()

	if h != nil {
		return h
	}

	// Fall back to tool system
	return tool.GlobalBuildBridge().GetHandler(name)
}

// GetAllHandlers returns all registered handlers (native + tool system).
func GetAllHandlers() map[string]Handler {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]Handler, len(handlers))
	for k, v := range handlers {
		result[k] = v
	}

	// Add tool system handlers (don't override native)
	for name, th := range tool.GlobalBuildBridge().GetAllHandlers() {
		if _, exists := result[name]; !exists {
			result[name] = th
		}
	}

	return result
}

// HasHandler checks if a handler exists by name.
func HasHandler(name string) bool {
	mu.RLock()
	_, exists := handlers[name]
	mu.RUnlock()

	if exists {
		return true
	}

	return tool.GlobalBuildBridge().HasHandler(name)
}

// GetHandlersForModule returns all handlers for a module's buildable components.
// Uses native handlers when available, falls back to tool system's resolver for
// tools configured in component-tools mapping (e.g., npm-build for typescript).
func GetHandlersForModule(module *modules.ModuleContract) []ComponentHandler {
	if module == nil {
		return nil
	}

	var result []ComponentHandler

	// Priority 1: Check for per-module handler override
	if module.GetBuildHandler() != "" {
		handlerName := module.GetBuildHandler()
		if h := GetHandler(handlerName); h != nil {
			result = append(result, ComponentHandler{
				Component: "override",
				Handler:   h,
			})
			return result
		}
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentTypes == nil {
		return nil
	}

	// Priority 2: Find handlers from component types
	for _, compName := range module.GetEnabledComponents() {
		compTypeName := module.Components.GetComponentType(compName)
		compType := cfg.ComponentTypes.Get(compTypeName)
		if compType == nil || !compType.HasBuilder() {
			continue
		}

		builderName := compType.Builder

		// Try native handler first (for specialized handlers like go, mkdocs, docker)
		mu.RLock()
		h := handlers[builderName]
		mu.RUnlock()

		if h != nil {
			result = append(result, ComponentHandler{
				Component: compName,
				Handler:   h,
			})
			continue
		}

		// Fall back to tool system's resolver which uses component-tools mapping
		// This allows npm-build to be used for typescript even though
		// component-types.yml specifies builder: npm
		if toolHandler := tool.GlobalBuildBridge().GetHandlerForComponent(compTypeName); toolHandler != nil {
			result = append(result, ComponentHandler{
				Component: compName,
				Handler:   toolHandler,
			})
		}
	}

	return result
}

// GetHandlerByBuilder returns the handler for a specific builder name.
func GetHandlerByBuilder(builderName string) Handler {
	return GetHandler(builderName)
}
