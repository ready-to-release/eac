// static.go - Test handlers for non-Go build dependencies
package testers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// TestStaticModule is a passthrough for modules without runtime tests.
// These modules are validated by the build process.
func TestStaticModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, reportFormat, suiteName string) int {
	Writeln(logWriter, "\n=== Testing static: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	Writeln(logWriter, "Static module has no runtime tests")
	Writeln(logWriter, "✅ Static module - validation done at build time")
	return 0
}

// TestMkDocsModule tests MkDocs sites by verifying the build output exists.
func TestMkDocsModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, reportFormat, suiteName string) int {
	Writeln(logWriter, "\n=== Testing book: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)

	// Check that the build output exists
	buildOutputDir := filepath.Join(workspaceRoot, paths.OutBuildRelPath, module.Moniker, "site")
	indexFile := filepath.Join(buildOutputDir, "index.html")

	Writeln(logWriter, "Checking build output: %s", buildOutputDir)

	// Verify site directory exists
	if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
		Writeln(logWriter, "\n❌ Build output not found: %s", buildOutputDir)
		Writeln(logWriter, "   Run 'build %s' first", module.Moniker)
		return 1
	}

	// Verify index.html exists (indicates successful build)
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		Writeln(logWriter, "\n❌ index.html not found in build output")
		Writeln(logWriter, "   Build may have failed - run 'build %s'", module.Moniker)
		return 1
	}

	Writeln(logWriter, "✅ Build output verified: %s", indexFile)
	Writeln(logWriter, "\n✅ MkDocs site validation passed")

	return 0
}
