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

// RegisterSystem registers a handler for a build system.
// Call this from init() in your builder file.
// The build_system is looked up from module-types.yml contract.
func RegisterSystem(buildSystem string, fn BuildFunc) {
	mu.Lock()
	defer mu.Unlock()
	systemHandlers[buildSystem] = fn
}

// GetBuildFunc returns the appropriate build function for a module type.
// It looks up the build_system from the module-types contract and returns
// the registered handler for that build system.
func GetBuildFunc(moduleType string) BuildFunc {
	mu.RLock()
	defer mu.RUnlock()

	// Look up build system from contracts
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		buildSystem := cfg.ModuleTypes.GetBuildSystem(moduleType)
		if fn, ok := systemHandlers[buildSystem]; ok {
			return fn
		}
	}

	// Fallback: no-op (for unknown types)
	if fn, ok := systemHandlers["none"]; ok {
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
