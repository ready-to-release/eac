// scripts.go - Build handler for scripts-package modules
package builders

import (
	"io"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

func init() {
	// Register handler for scripts-package modules
	RegisterSystem("scripts", BuildScriptsModule)
}

// BuildScriptsModule copies script files from source to build output.
// It uses the module's files.source patterns to determine which files to copy,
// preserving directory structure relative to the module root.
func BuildScriptsModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== %s: %s ===", module.Type, module.Moniker)

	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)
	Logln(logWriter, "Source: %s", moduleRoot)
	Logln(logWriter, "Output: %s", outputDir)

	// Get source patterns from module contract
	sourcePatterns := module.Files.Source
	if len(sourcePatterns) == 0 {
		Logln(logWriter, "ℹ️  No source patterns defined, nothing to copy")
		return 0
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	filesCopied := 0

	// Process each source pattern
	for _, pattern := range sourcePatterns {
		Logln(logWriter, "📁 Pattern: %s", pattern)

		// Find matching files
		matches, err := doublestar.Glob(os.DirFS(moduleRoot), pattern)
		if err != nil {
			Logln(logWriter, "   ⚠️  Invalid pattern: %v", err)
			continue
		}

		for _, match := range matches {
			srcPath := filepath.Join(moduleRoot, match)

			// Skip directories (we only copy files)
			info, err := os.Stat(srcPath)
			if err != nil || info.IsDir() {
				continue
			}

			// Preserve directory structure in output
			dstPath := filepath.Join(outputDir, match)

			// Ensure parent directory exists
			dstDir := filepath.Dir(dstPath)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				Logln(logWriter, "   ❌ Failed to create directory %s: %v", dstDir, err)
				return 1
			}

			// Copy file
			if err := CopyFile(srcPath, dstPath); err != nil {
				Logln(logWriter, "   ❌ Failed to copy %s: %v", match, err)
				return 1
			}

			filesCopied++
		}
	}

	Logln(logWriter, "✅ Copied %d files to build output", filesCopied)

	// Write build marker
	if err := WriteBuildMarker(outputDir); err != nil {
		Logln(logWriter, "⚠️  Failed to write build marker: %v", err)
	}

	return 0
}
