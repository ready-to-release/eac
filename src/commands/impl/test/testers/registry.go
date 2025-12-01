// registry.go - Self-registration system for test handlers
package testers

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// TestFunc is the signature for module test functions.
// Parameters: module contract, workspace root, output directory, log writer, report format, suite name
// Returns: exit code
type TestFunc func(*modules.ModuleContract, string, string, io.Writer, string, string) int

var (
	mu             sync.RWMutex
	systemHandlers = make(map[string]TestFunc)
)

// RegisterSystem registers a handler for a build dependency.
// Call this from init() in your tester file.
// The primary build_dep is looked up from module-types.yml contract.
func RegisterSystem(buildDep string, fn TestFunc) {
	mu.Lock()
	defer mu.Unlock()
	systemHandlers[buildDep] = fn
}

// GetTestFunc returns the appropriate test function for a module type.
// It uses dispatch rules from handlers.yml to determine which handler to use,
// falling back to the primary build_dep from module-types.yml.
func GetTestFunc(moduleType string) TestFunc {
	mu.RLock()
	defer mu.RUnlock()

	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		// No config available, use static handler
		if fn, ok := systemHandlers[""]; ok {
			return fn
		}
		return func(*modules.ModuleContract, string, string, io.Writer, string, string) int {
			return 0
		}
	}

	// Get module capabilities and primary build dep
	capabilities := cfg.ModuleTypes.GetCapabilities(moduleType)
	primaryDep := cfg.ModuleTypes.GetPrimaryBuildDep(moduleType)

	// Use handlers config dispatch rules if available
	var handlerName string
	if cfg.Handlers != nil {
		handlerName = cfg.Handlers.GetTestHandler(moduleType, capabilities, primaryDep)
	} else {
		// Legacy fallback: use primary build dep
		handlerName = primaryDep
	}

	// Look up the handler
	if handlerName != "" {
		if fn, ok := systemHandlers[handlerName]; ok {
			return fn
		}
	}

	// Fallback: static module test (no-op for types with no build deps)
	if fn, ok := systemHandlers[""]; ok {
		return fn
	}

	// Panic-safe fallback (should never reach here)
	return func(*modules.ModuleContract, string, string, io.Writer, string, string) int {
		return 0
	}
}

// RunTestSuiteForModule is a callback that will be set by the test package
// to allow testers to run test suites without circular imports.
// This enables testers like TestGoCLI to delegate to the full test suite infrastructure.
var RunTestSuiteForModule func(moniker string, suiteName string) int
