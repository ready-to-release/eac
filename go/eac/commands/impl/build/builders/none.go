// none.go - Build handler for modules with no build step (empty build_deps)
package builders

import (
	"io"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	// Register handler for modules with no build dependencies
	// These modules have no build step but still write a marker for artifact validation
	RegisterSystem("", BuildNoneModule)
	RegisterSystemArtifacts("", ListNoneArtifacts)
}

// ListNoneArtifacts returns the build marker artifact
func ListNoneArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	return []string{BuildMarkerFilename}
}

// BuildNoneModule writes a build marker for modules that don't require building.
// Module types with empty build_deps include config files, templates, etc.
// The marker ensures these modules participate in the artifact validation system.
func BuildNoneModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== %s: %s ===", module.Type, module.Moniker)
	Logln(logWriter, "ℹ️  No build step required for %s", module.Type)

	// Write build marker for artifact validation
	if err := WriteBuildMarker(outputDir); err != nil {
		Logln(logWriter, "⚠️  Failed to write build marker: %v", err)
	}

	return 0
}
