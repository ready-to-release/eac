// static.go - Test functions for static/config module types
package testers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// TestStaticModule is a passthrough test for static/configuration modules.
// These modules don't have runtime tests - they are validated by the build process.
func TestStaticModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing %s: %s ===", module.Type, module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	Writeln(logWriter, "Module type '%s' has no runtime tests", module.Type)
	Writeln(logWriter, "✅ Static module - validation done at build time")
	return 0
}

// TestMkDocsSite tests the MkDocs documentation site by verifying the build output exists.
// The build is done separately with strict mode - this test just verifies the output.
func TestMkDocsSite(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing mkdocs-site: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)

	// Check that the build output exists
	buildOutputDir := filepath.Join(workspaceRoot, "out", "build", module.Moniker, "site")
	indexFile := filepath.Join(buildOutputDir, "index.html")

	Writeln(logWriter, "Checking build output: %s", buildOutputDir)

	// Verify site directory exists
	if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
		Writeln(logWriter, "\n❌ Build output not found: %s", buildOutputDir)
		Writeln(logWriter, "   Run 'build module %s' first", module.Moniker)
		return 1
	}

	// Verify index.html exists (indicates successful build)
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		Writeln(logWriter, "\n❌ index.html not found in build output")
		Writeln(logWriter, "   Build may have failed - run 'build module %s'", module.Moniker)
		return 1
	}

	Writeln(logWriter, "✅ Build output verified: %s", indexFile)
	Writeln(logWriter, "\n✅ MkDocs site validation passed")

	return 0
}
