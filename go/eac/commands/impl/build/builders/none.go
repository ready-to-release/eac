// none.go - Build handler for modules with no build step (empty build_deps)
package builders

import (
	"io"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	// Register handler for modules with no build dependencies
	// These modules have no build step - validation is done separately
	RegisterSystem("", BuildNoneModule)
}

// BuildNoneModule is a no-op build function for modules that don't require building.
// Module types with empty build_deps include config files, scripts, templates, etc.
// If validation is needed, it should be handled by the validate command, not build.
func BuildNoneModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== %s: %s ===", module.Type, module.Moniker)
	Logln(logWriter, "ℹ️  No build step required for this module type")
	return 0
}
