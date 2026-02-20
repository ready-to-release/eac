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
	"runtime"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/tool"
)

// installPlatformMarker is written next to package.json after each managed install.
// It records the OS so a cross-platform mismatch triggers a clean reinstall on the next build.
// Listed in .gitignore so it is never committed.
const installPlatformMarker = ".npm-install-platform"

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

	// Step 1: Install dependencies when needed.
	// Reinstalls if node_modules is missing, was installed on a different OS,
	// or package-lock.json has changed since the last managed install.
	if needsNpmInstall(moduleRoot) {
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
		if err := writeInstallMarker(moduleRoot); err != nil {
			Logln(logWriter, "⚠️  Failed to write install marker: %v (next build will reinstall)", err)
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

// needsNpmInstall reports whether npm dependencies must be installed before building.
// Returns true when:
//   - node_modules does not exist
//   - the platform marker is missing or records a different OS (cross-platform mismatch)
//   - package-lock.json is newer than the marker (lock file changed since last install)
func needsNpmInstall(moduleRoot string) bool {
	if _, err := os.Stat(filepath.Join(moduleRoot, "node_modules")); os.IsNotExist(err) {
		return true
	}
	markerPath := filepath.Join(moduleRoot, installPlatformMarker)
	markerInfo, err := os.Stat(markerPath)
	if err != nil {
		return true // marker absent — node_modules was not installed by this tool
	}
	markerOS, err := os.ReadFile(markerPath)
	if err != nil || strings.TrimSpace(string(markerOS)) != runtime.GOOS {
		return true // installed on a different platform
	}
	lockFile := filepath.Join(moduleRoot, "package-lock.json")
	if lockInfo, err := os.Stat(lockFile); err == nil {
		return lockInfo.ModTime().After(markerInfo.ModTime())
	}
	return false
}

// writeInstallMarker records the current OS next to package.json so the next build
// can detect a cross-platform mismatch or a stale install.
func writeInstallMarker(moduleRoot string) error {
	return os.WriteFile(
		filepath.Join(moduleRoot, installPlatformMarker),
		[]byte(runtime.GOOS),
		0o644,
	)
}
