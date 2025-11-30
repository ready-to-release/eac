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
// It looks up the primary build_dep from the module-types contract and returns
// the registered handler for that build dependency.
// Special case: documentation + container capability uses mkdocs handler.
func GetBuildFunc(moduleType string) BuildFunc {
	mu.RLock()
	defer mu.RUnlock()

	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		// Special case: documentation sites with container capability use mkdocs builder
		if cfg.ModuleTypes.HasCapability(moduleType, "documentation") &&
			cfg.ModuleTypes.HasCapability(moduleType, "container") {
			if fn, ok := systemHandlers["mkdocs"]; ok {
				return fn
			}
		}

		// Standard case: look up primary build dep from contracts
		primaryDep := cfg.ModuleTypes.GetPrimaryBuildDep(moduleType)
		if primaryDep != "" {
			if fn, ok := systemHandlers[primaryDep]; ok {
				return fn
			}
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
