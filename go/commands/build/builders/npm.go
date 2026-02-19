// npm.go - Build handler for npm/TypeScript modules.
//
// Ensures project dependencies are installed before running the build.
// Uses npm ci (if lock file exists) or npm install as a pre-build step,
// then delegates to npm run build.
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&NpmHandler{})
}

// NpmHandler builds TypeScript/JavaScript modules using npm.
type NpmHandler struct{}

func (h *NpmHandler) Name() string { return "npm-build" }

func (h *NpmHandler) Capabilities() []string { return []string{"npm-build"} }

func (h *NpmHandler) Requirements() []string { return []string{"npm"} }

func (h *NpmHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	componentRoot := filepath.Join(workspaceRoot, module.GetComponentRoot(component))
	packageJSON := filepath.Join(componentRoot, "package.json")
	if _, err := os.Stat(packageJSON); err != nil {
		return fmt.Errorf("package.json not found in %s", componentRoot)
	}
	return nil
}

func (h *NpmHandler) IsContainer() bool     { return false }
func (h *NpmHandler) IsHostInstalled() bool { return true }

func (h *NpmHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"out/"}
}

func (h *NpmHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	componentRoot := module.GetComponentRoot(opts.Component)
	moduleRoot := filepath.Join(workspaceRoot, componentRoot)

	Logln(logWriter, "\n=== Building npm: %s ===", module.GetMoniker())

	// Step 1: Install dependencies if node_modules is missing
	nodeModules := filepath.Join(moduleRoot, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		lockFile := filepath.Join(moduleRoot, "package-lock.json")
		if _, err := os.Stat(lockFile); err == nil {
			Logln(logWriter, "Running: npm ci")
			if exitCode := RunCommandWithLog(context.Background(), moduleRoot, logWriter, "npm", "ci"); exitCode != 0 {
				Logln(logWriter, "❌ npm ci failed")
				return exitCode
			}
		} else {
			Logln(logWriter, "Running: npm install")
			if exitCode := RunCommandWithLog(context.Background(), moduleRoot, logWriter, "npm", "install"); exitCode != 0 {
				Logln(logWriter, "❌ npm install failed")
				return exitCode
			}
		}
	}

	// Step 2: Run build
	Logln(logWriter, "Running: npm run build")
	if exitCode := RunCommandWithLog(context.Background(), moduleRoot, logWriter, "npm", "run", "build"); exitCode != 0 {
		Logln(logWriter, "❌ npm run build failed")
		return exitCode
	}

	Logln(logWriter, "✅ npm build completed successfully")
	return 0
}
