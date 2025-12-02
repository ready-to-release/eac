// npm.go - Build handler for npm build system
package builders

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	// Register handler for "npm" build system
	RegisterSystem("npm", BuildNpmModule)
}

// BuildNpmModule builds npm-based modules.
// Uses capabilities to determine build steps:
// - typescript: run tsc to compile, output to out/build/{moniker}/
func BuildNpmModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		Logln(logWriter, "❌ No package.json found at: %s", packageJSON)
		return 1
	}

	// Step 1: npm install (in module root for node_modules)
	Logln(logWriter, "📦 Installing dependencies")
	Logln(logWriter, "Running: npm install")

	exitCode := RunCommandWithLog(moduleRoot, logWriter, "npm", "install")
	if exitCode != 0 {
		Logln(logWriter, "❌ npm install failed")
		return exitCode
	}

	// Step 2: Compile based on capabilities
	cfg := config.Global()
	hasTypeScript := cfg != nil && cfg.ModuleTypes != nil && cfg.ModuleTypes.HasCapability(module.Type, "typescript")

	if hasTypeScript {
		Logln(logWriter, "🔨 Compiling TypeScript")
		// Compile to out/build/{moniker}/ instead of module's out/
		Logln(logWriter, "Running: npx tsc -p ./ --outDir %s", outputDir)

		exitCode = RunCommandWithLog(moduleRoot, logWriter, "npx", "tsc", "-p", "./", "--outDir", outputDir)
		if exitCode != 0 {
			Logln(logWriter, "❌ TypeScript compilation failed")
			return exitCode
		}
	}

	// Copy package.json to output directory for post-build steps
	destPackageJSON := filepath.Join(outputDir, "package.json")
	if err := CopyFile(packageJSON, destPackageJSON); err != nil {
		Logln(logWriter, "⚠️  Could not copy package.json: %v", err)
	}

	// Write build marker for dependency verification
	if err := WriteBuildMarker(outputDir); err != nil {
		Logln(logWriter, "⚠️  Could not write build marker: %v", err)
	}

	Logln(logWriter, "✅ npm module built successfully")
	Logln(logWriter, "   Output: %s", outputDir)

	return 0
}
