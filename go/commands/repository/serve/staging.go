package serve

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ready-to-release/eac/go/commands/build/docprep/content"
	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// cmdMarkerPresent quickly detects any book:cmd marker in a file.
var cmdMarkerPresent = regexp.MustCompile(`<!--\s*book:`)

// PrepareDevStaging creates or refreshes the staging directory for dev mode.
// It mirrors the source docs tree, expands command markers, and writes
// the MkDocs wrapper config. Returns the staging directory path.
func PrepareDevStaging(
	ctx context.Context,
	workspaceRoot, moniker string,
	logf func(string, ...any),
) (string, error) {
	stagingDir := filepath.Join(paths.ServeOutputPath(workspaceRoot), moniker)
	stagingDocsDir := filepath.Join(stagingDir, "docs")
	sourceDocsDir := filepath.Join(workspaceRoot, "docs")

	if err := os.MkdirAll(stagingDocsDir, 0o755); err != nil {
		return "", fmt.Errorf("creating staging docs dir: %w", err)
	}

	logf("  Mirroring docs tree into staging...")
	if err := mirrorDocsTree(sourceDocsDir, stagingDocsDir); err != nil {
		return "", fmt.Errorf("mirroring docs tree: %w", err)
	}

	// Expand command markers if eac binary exists
	cmdBinary := paths.CommandsBinaryPath(workspaceRoot)
	if _, err := os.Stat(cmdBinary); err == nil {
		logf("  Expanding command markers...")
		if err := expandMarkersInStaging(ctx, workspaceRoot, stagingDocsDir, logf); err != nil {
			logf("  Warning: marker expansion failed: %v", err)
			// Non-fatal: pages will show without expanded command help
		}
	} else {
		logf("  Skipping marker expansion (eac binary not found at %s)", cmdBinary)
	}

	if err := writeStagingConfig(stagingDir, workspaceRoot); err != nil {
		return "", fmt.Errorf("writing staging config: %w", err)
	}

	return stagingDir, nil
}

// mirrorDocsTree copies all files from sourceDir into targetDir.
// Uses full copies (no symlinks) for Windows Docker Desktop compatibility.
// Files containing book:cmd markers are always re-copied (need fresh expansion).
// Other files use mtime-based skipping to speed up re-runs.
func mirrorDocsTree(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(sourceDir, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(targetDir, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		// Always re-copy files with markers (they need fresh expansion)
		if filepath.Ext(path) == ".md" && fileContainsMarker(path) {
			return copyFile(path, target)
		}

		// For other files, skip if target is up-to-date (mtime-based)
		srcInfo, err := d.Info()
		if err != nil {
			return copyFile(path, target)
		}
		if dstInfo, statErr := os.Stat(target); statErr == nil {
			if dstInfo.Size() == srcInfo.Size() && !srcInfo.ModTime().After(dstInfo.ModTime()) {
				return nil // target is up-to-date
			}
		}

		return copyFile(path, target)
	})
}

// expandMarkersInStaging processes book:cmd markers in all .md files in stagingDocsDir.
func expandMarkersInStaging(
	ctx context.Context,
	workspaceRoot, stagingDocsDir string,
	logf func(string, ...any),
) error {
	idx, err := staging.NewFileIndex(stagingDocsDir)
	if err != nil {
		return fmt.Errorf("indexing staging docs: %w", err)
	}

	executor := content.ToolCommandExecutor{}
	return content.ProcessCommandMarkers(
		ctx,
		idx,
		stagingDocsDir,
		workspaceRoot,
		executor,
		logf,
		func(format string, args ...any) {
			logf("  WARN: "+format, args...)
		},
	)
}

// writeStagingConfig writes the MkDocs wrapper config and macros file to stagingDir.
// The config inherits from /source/docs/mkdocs.yml (container mount path).
func writeStagingConfig(stagingDir, workspaceRoot string) error {
	// Write mkdocs.yml wrapper that inherits from the source config
	configContent := `INHERIT: /source/docs/mkdocs.yml
docs_dir: /workspace/docs
dev_addr: '0.0.0.0:8000'
`
	configPath := filepath.Join(stagingDir, "mkdocs.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		return fmt.Errorf("writing mkdocs.yml: %w", err)
	}

	// Copy mkdocs_macros.py as main.py (MkDocs macros plugin looks for main.py beside config)
	macrosSrc := filepath.Join(workspaceRoot, "containers", "mkdocs-dev-oci", "mkdocs_macros.py")
	macrosDst := filepath.Join(stagingDir, "main.py")
	if err := copyFile(macrosSrc, macrosDst); err != nil {
		return fmt.Errorf("copying macros file: %w", err)
	}

	return nil
}

// fileContainsMarker checks if a file contains any book:cmd marker.
func fileContainsMarker(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return cmdMarkerPresent.Match(data)
}

// copyFile copies src to dst, creating parent directories as needed.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
