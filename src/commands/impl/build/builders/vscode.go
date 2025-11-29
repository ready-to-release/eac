// vscode.go - Build functions for VS Code module types
package builders

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// BuildVSCodeDefault is the default build handler for VS Code modules.
func BuildVSCodeDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== Building VS Code module: %s (type: %s) ===", module.Moniker, module.Type)
	Logln(logWriter, "ℹ️  VS Code extensions are built via vsce package command")
	return 0
}

// BuildVSCodeExtension builds a VS Code extension
func BuildVSCodeExtension(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Building vscode-ext: %s ===", module.Moniker)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		Logln(logWriter, "⚠️  No package.json found at: %s", packageJSON)
		Logln(logWriter, "ℹ️  Skipping VS Code extension build")
		return 0
	}

	Logln(logWriter, "📦 Installing dependencies")
	Logln(logWriter, "Running: npm install")

	exitCode := RunCommandWithLog(moduleRoot, logWriter, "npm", "install")
	if exitCode != 0 {
		Logln(logWriter, "❌ npm install failed")
		return exitCode
	}

	Logln(logWriter, "🔨 Compiling TypeScript")
	Logln(logWriter, "Running: npm run compile")

	exitCode = RunCommandWithLog(moduleRoot, logWriter, "npm", "run", "compile")
	if exitCode != 0 {
		Logln(logWriter, "❌ npm run compile failed")
		return exitCode
	}

	Logln(logWriter, "✅ VS Code extension built successfully")

	return 0
}
