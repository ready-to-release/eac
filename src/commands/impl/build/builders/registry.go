// registry.go - Self-registration system for build handlers
package builders

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// BuildOptions contains flags for controlling the build process.
type BuildOptions struct {
	TidyFirst     bool   // Run go mod tidy before building
	Version       string // Version to inject via ldflags
	Compressed    bool   // Strip debug info with -ldflags "-s -w" (for releases)
	CompressedUPX bool   // Also apply UPX compression after build
}

// BuildFunc is the signature for module build functions.
type BuildFunc func(*modules.ModuleContract, string, string, io.Writer, BuildOptions) int

var (
	mu             sync.RWMutex
	systemHandlers = make(map[string]BuildFunc)
)

// RegisterSystem registers a handler for a build dependency.
// Call this from init() in your builder file.
// The primary build_dep is looked up from module-types.yml contract.
func RegisterSystem(buildDep string, fn BuildFunc) {
	mu.Lock()
	defer mu.Unlock()
	systemHandlers[buildDep] = fn
}

// GetBuildFunc returns the appropriate build function for a module type.
// It uses dispatch rules from handlers.yml to determine which handler to use,
// falling back to the primary build_dep from module-types.yml.
func GetBuildFunc(moduleType string) BuildFunc {
	mu.RLock()
	defer mu.RUnlock()

	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		// No config available, use no-op handler
		if fn, ok := systemHandlers[""]; ok {
			return fn
		}
		return func(*modules.ModuleContract, string, string, io.Writer, BuildOptions) int {
			return 0
		}
	}

	// Get module capabilities and primary build dep
	capabilities := cfg.ModuleTypes.GetCapabilities(moduleType)
	primaryDep := cfg.ModuleTypes.GetPrimaryBuildDep(moduleType)

	// Use handlers config dispatch rules if available
	var handlerName string
	if cfg.Handlers != nil {
		handlerName = cfg.Handlers.GetBuildHandler(moduleType, capabilities, primaryDep)
	} else {
		// Legacy fallback: special case for documentation + container
		if cfg.ModuleTypes.HasCapability(moduleType, "documentation") &&
			cfg.ModuleTypes.HasCapability(moduleType, "container") {
			handlerName = "mkdocs"
		} else {
			handlerName = primaryDep
		}
	}

	// Look up the handler
	if handlerName != "" {
		if fn, ok := systemHandlers[handlerName]; ok {
			return fn
		}
	}

	// Fallback: no-op (for types with no build deps)
	if fn, ok := systemHandlers[""]; ok {
		return fn
	}

	// Panic-safe fallback (should never reach here)
	return func(*modules.ModuleContract, string, string, io.Writer, BuildOptions) int {
		return 0
	}
}

// IsGoModuleType returns true if the module type uses Go tooling (has go_module capability)
func IsGoModuleType(moduleType string) bool {
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		return cfg.ModuleTypes.HasCapability(moduleType, "go_module")
	}
	return false
}
