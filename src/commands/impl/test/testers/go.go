// go.go - Test handler for Go build system
package testers

import (
	"io"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

func init() {
	// Register handler for "go" build system
	// All go-* types (go-cli, go-library, go-commands, etc.) use this via their build_system contract
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
