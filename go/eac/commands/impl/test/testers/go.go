// go.go - Test handler for Go build system
package testers

import (
	"io"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	// Register handler for "go" build dependency
	// The unified "go" type uses this via its build_deps contract
	RegisterSystem("go", TestGoModule)
}

// TestGoModule is the test handler for all Go modules.
// It delegates to the test suite orchestrator which handles:
// - Discovery of both unit tests (*_test.go) and BDD specs (specs/{moniker}/**)
// - Suite-based filtering (unit, integration, acceptance, etc.)
// - Parallel execution and reporting
func TestGoModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing %s: %s ===", module.Type, module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	return RunTestSuiteForModule(module.Moniker, suiteName)
}
