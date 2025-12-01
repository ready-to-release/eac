// static.go - Test handlers for non-Go build dependencies
package testers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

func init() {
	// Build dependency handlers - contracts define which types use which build deps
	RegisterSystem("mkdocs", TestMkDocsModule)
	RegisterSystem("npm", TestNpmModule)
	RegisterSystem("docker", TestStaticModule)
	RegisterSystem("", TestStaticModule) // Modules with no build deps
}

// TestStaticModule is a passthrough for modules without runtime tests.
// These modules are validated by the build process.
func TestStaticModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing %s: %s ===", module.Type, module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)
	Writeln(logWriter, "Module type '%s' has no runtime tests", module.Type)
	Writeln(logWriter, "✅ Static module - validation done at build time")
	return 0
}

// TestNpmModule runs npm test for npm-based modules.
func TestNpmModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Writeln(logWriter, "\n=== Testing %s: %s ===", module.Type, module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		Writeln(logWriter, "⚠️  No package.json found, skipping tests")
		return 0
	}

	Writeln(logWriter, "Running: npm test")
	return RunTestCommand(moduleRoot, logWriter, "npm", "test")
}

// TestMkDocsModule tests MkDocs sites by verifying the build output exists.
func TestMkDocsModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, reportFormat string, suiteName string) int {
	Writeln(logWriter, "\n=== Testing %s: %s ===", module.Type, module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)

	// Check that the build output exists
	buildOutputDir := filepath.Join(workspaceRoot, "out", "build", module.Moniker, "site")
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
