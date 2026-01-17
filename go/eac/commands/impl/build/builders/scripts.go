// scripts.go - Build handler for scripts-package modules
package builders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	RegisterHandler(&ScriptsHandler{})
}

// ScriptsHandler copies script files from source to build output.
type ScriptsHandler struct{}

func (h *ScriptsHandler) Name() string { return "scripts" }

func (h *ScriptsHandler) Capabilities() []string { return []string{"scripts_package"} }

func (h *ScriptsHandler) Requirements() []string { return nil }

func (h *ScriptsHandler) ValidateModule(module *modules.ModuleContract, workspaceRoot string) error {
	if len(module.Files.Source) == 0 {
		return fmt.Errorf("no source patterns defined in files.source")
	}
	return nil
}

func (h *ScriptsHandler) ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	return nil // Artifacts are the copied files, tracked in manifest
}

func (h *ScriptsHandler) Build(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== %s: %s ===", module.Type, module.Moniker)

	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)
	Logln(logWriter, "Source: %s", moduleRoot)
	Logln(logWriter, "Output: %s", outputDir)

	sourcePatterns := module.Files.Source
	if len(sourcePatterns) == 0 {
		Logln(logWriter, "ℹ️  No source patterns defined, nothing to copy")
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
