// registry.go - Self-registration system for test handlers
package testers

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

// TestFunc is the signature for module test functions.
// Parameters: module contract, workspace root, output directory, log writer, report format, suite name
// Returns: exit code.
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

// Component type to test handler mapping.
var componentTestHandlers = map[string]string{
	"go":         "go",
	"npm":        "npm",
	"typescript": "npm",
}

// GetTestFunc returns the appropriate test function for a module.
// It matches module component types to test handlers.
func GetTestFunc(module *modules.ModuleContract) TestFunc {
	mu.RLock()
	defer mu.RUnlock()

	if module == nil {
		if fn, ok := systemHandlers[""]; ok {
			return fn
		}
		return func(*modules.ModuleContract, string, string, io.Writer, string, string) int {
			return 0
		}
	}

	// Check module component types and find matching handler
	for compType, handlerName := range componentTestHandlers {
		if module.HasComponent(compType) {
			if fn, ok := systemHandlers[handlerName]; ok {
				return fn
			}
		}
	}

	// Fallback: static module test (no-op for types with no test handler)
	if fn, ok := systemHandlers[""]; ok {
		return fn
	}

	return func(*modules.ModuleContract, string, string, io.Writer, string, string) int {
		return 0
	}
}
