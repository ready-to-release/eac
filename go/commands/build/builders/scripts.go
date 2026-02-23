// scripts.go - Build handler for scripts-package modules
package builders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&ScriptsHandler{})
}

// ScriptsHandler copies script files from source to build output.
type ScriptsHandler struct{}

func (h *ScriptsHandler) Name() string { return "scripts" }

func (h *ScriptsHandler) Capabilities() []string { return []string{"scripts_package", "pwsh", "bash"} }

func (h *ScriptsHandler) Requirements() []string { return nil }

func (h *ScriptsHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		return fmt.Errorf("invalid module type")
	}
	// Check if any package has source patterns
	for _, pkg := range concrete.Components {
		if pkg != nil && pkg.Patterns != nil && len(pkg.Patterns.Source) > 0 {
			return nil
		}
	}
	return fmt.Errorf("no source patterns defined in any package")
}

// IsContainer returns false as scripts handler copies files on the host.
func (h *ScriptsHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as scripts handler copies files using the host filesystem.
func (h *ScriptsHandler) IsHostInstalled() bool { return true }

func (h *ScriptsHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return nil // Artifacts are the copied files, tracked in manifest
}

func (h *ScriptsHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}
	Logln(logWriter, "\n=== scripts: %s ===", concrete.Moniker)

	// Get the specific component being built
	comp := concrete.Components[opts.Component]
	if comp == nil {
		Logln(logWriter, "❌ Component %s not found in module %s", opts.Component, concrete.Moniker)
		return 1
	}

	moduleRoot := filepath.Join(workspaceRoot, comp.Root)
	Logln(logWriter, "Source: %s", moduleRoot)
	Logln(logWriter, "Output: %s", outputDir)

	// Get source patterns from the specific component
	var sourcePatterns []string
	if comp.Patterns != nil {
		sourcePatterns = comp.Patterns.Source
	}
	if len(sourcePatterns) == 0 {
		Logln(logWriter, "ℹ️  No source patterns defined for component %s, nothing to copy", opts.Component)
		return 0
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	filesCopied := 0

	for _, pattern := range sourcePatterns {
		Logln(logWriter, "📁 Pattern: %s", pattern)

		matches, err := doublestar.Glob(os.DirFS(moduleRoot), pattern)
		if err != nil {
			Logln(logWriter, "   ⚠️  Invalid pattern: %v", err)
			continue
		}

		for _, match := range matches {
			srcPath := filepath.Join(moduleRoot, match)

			info, err := os.Stat(srcPath)
			if err != nil || info.IsDir() {
				continue
			}

			dstPath := filepath.Join(outputDir, match)
			dstDir := filepath.Dir(dstPath)
			if err := os.MkdirAll(dstDir, 0o755); err != nil {
				Logln(logWriter, "   ❌ Failed to create directory %s: %v", dstDir, err)
				return 1
			}

			if err := CopyFile(srcPath, dstPath); err != nil {
				Logln(logWriter, "   ❌ Failed to copy %s: %v", match, err)
				return 1
			}

			filesCopied++
		}
	}

	Logln(logWriter, "✅ Copied %d files to build output", filesCopied)
	return 0
}

var _ build.BuilderPort = (*ScriptsHandler)(nil)
