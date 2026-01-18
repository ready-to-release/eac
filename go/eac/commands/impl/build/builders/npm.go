// npm.go - Build handler for npm build system
package builders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	RegisterHandler(&NpmHandler{})
}

// NpmHandler builds npm-based modules.
type NpmHandler struct{}

func (h *NpmHandler) Name() string { return "npm" }

func (h *NpmHandler) Capabilities() []string { return []string{"npm_package", "typescript"} }

func (h *NpmHandler) Requirements() []string { return []string{"npm"} }

func (h *NpmHandler) ValidateModule(module *modules.ModuleContract, workspaceRoot string) error {
	moduleRoot := filepath.Join(workspaceRoot, module.GetComponentRoot("typescript"))
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found at %s", packageJSON)
	}
	return nil
}

func (h *NpmHandler) ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	return []string{"package.json"}
}

func (h *NpmHandler) Build(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.GetComponentRoot("typescript"))

	Logln(logWriter, "\n=== Building typescript: %s ===", module.Moniker)

	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		Logln(logWriter, "❌ No package.json found at: %s", packageJSON)
		return 1
	}

	Logln(logWriter, "📦 Installing dependencies")
	Logln(logWriter, "Running: npm install")

	exitCode := RunCommandWithLog(moduleRoot, logWriter, "npm", "install")
	if exitCode != 0 {
		Logln(logWriter, "❌ npm install failed")
		return exitCode
	}

	hasTypeScript := module.HasComponent("typescript")

	if hasTypeScript {
		Logln(logWriter, "🔨 Compiling TypeScript")
		Logln(logWriter, "Running: npx tsc -p ./ --outDir %s", outputDir)

		exitCode = RunCommandWithLog(moduleRoot, logWriter, "npx", "tsc", "-p", "./", "--outDir", outputDir)
		if exitCode != 0 {
			Logln(logWriter, "❌ TypeScript compilation failed")
			return exitCode
		}
	}

	destPackageJSON := filepath.Join(outputDir, "package.json")
	if err := CopyFile(packageJSON, destPackageJSON); err != nil {
		Logln(logWriter, "⚠️  Could not copy package.json: %v", err)
	}

	Logln(logWriter, "✅ npm module built successfully")
	Logln(logWriter, "   Output: %s", outputDir)

	return 0
}
