// drawio.go - Build handler for rendering .drawio.png files via the drawio-tool container.
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	h := &DrawioRenderHandler{}
	RegisterHandler(h)
	tool.GlobalBuildBridge().RegisterNativeHandler(h)
}

// DrawioRenderHandler renders .drawio.png files via the drawio-tool container.
// Produces rendered PNGs at out/build/docs/drawio/drawio/ preserving relative paths.
// Uses .cache/eac/drawio/ for incremental builds (skip unchanged files).
type DrawioRenderHandler struct{}

func (h *DrawioRenderHandler) Name() string { return "drawio-render" }

func (h *DrawioRenderHandler) Requirements() []string { return []string{"docker"} }

func (h *DrawioRenderHandler) IsContainer() bool { return true }

func (h *DrawioRenderHandler) IsHostInstalled() bool { return false }

func (h *DrawioRenderHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for drawio rendering)")
	}
	return nil
}

func (h *DrawioRenderHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{"drawio/"}
}

func (h *DrawioRenderHandler) Build(
	module interfaces.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions,
) int {
	Logln(logWriter, "=== Rendering drawio images via container ===")

	// Source files live under docs/assets/
	docsAssetsDir := filepath.Join(workspaceRoot, "docs", "assets")
	cacheDir := paths.DrawioAccelCachePath(workspaceRoot)
	outputDrawioDir := filepath.Join(outputDir, "drawio")

	// 1. Discover all .drawio.png source files
	images, err := books.FindDrawioImages(docsAssetsDir)
	if err != nil {
		Logln(logWriter, "Error scanning for drawio images: %v", err)
		return 1
	}

	if len(images) == 0 {
		Logln(logWriter, "No drawio images found")
		_ = os.MkdirAll(outputDrawioDir, 0o755)
		return 0
	}

	Logln(logWriter, "Found %d drawio image(s)", len(images))

	// 2. Ensure directories exist
	_ = os.MkdirAll(cacheDir, 0o755)
	_ = os.MkdirAll(outputDrawioDir, 0o755)

	// 3. For each image: check cache, render via container if needed, copy to output
	rendered := 0
	cached := 0

	for _, img := range images {
		cacheKey := books.DrawioCacheKey{
			SourcePath: img.SourceFile,
			SourceHash: img.Hash,
			MaxWidth:   books.MaxImageWidthPDF,
		}
		hashStr := books.HashDrawioKey(cacheKey)

		// Cache file path: .cache/eac/drawio/{identifier}_{hash}.png
		identifier := paths.SanitizeForCacheName(img.RelPath)
		shortHash := hashStr[:paths.CacheHashLength]
		cacheFilename := fmt.Sprintf("%s_%s.png", identifier, shortHash)
		cachePath := filepath.Join(cacheDir, cacheFilename)

		// Output path preserves relative path from docs/assets/
		outputPath := filepath.Join(outputDrawioDir, img.RelPath)
		_ = os.MkdirAll(filepath.Dir(outputPath), 0o755)

		// Check cache hit
		if _, statErr := os.Stat(cachePath); statErr == nil && !opts.ForceRebuild {
			if cpErr := CopyFile(cachePath, outputPath); cpErr != nil {
				Logln(logWriter, "Error copying cached drawio %s: %v", img.RelPath, cpErr)
				return 1
			}
			cached++
			continue
		}

		// Cache miss — render via drawio-tool container
		bridge := tool.GlobalHandlerToolBridge()
		tc := &tool.ToolContext{
			WorkspaceRoot: workspaceRoot,
			OutputDir:     cacheDir,
			LogWriter:     logWriter,
		}

		// Convert host paths to container paths
		// Mount: {workspace} -> /docs, so /docs in container = workspace root
		containerInput, relErr := filepath.Rel(workspaceRoot, img.SourceFile)
		if relErr != nil {
			Logln(logWriter, "Error computing relative path for %s: %v", img.RelPath, relErr)
			return 1
		}
		containerInputPath := "/docs/" + filepath.ToSlash(containerInput)
		containerOutputPath := "/output/" + cacheFilename

		// Substitute into command: {input_path}, {output_path}, {max_width}
		// Note: {output} is reserved for tc.OutputDir (mount resolution)
		tc.Variables = map[string]string{
			"input_path":  containerInputPath,
			"output_path": containerOutputPath,
			"max_width":   fmt.Sprintf("%d", books.MaxImageWidthPDF),
		}

		exitCode, execErr := bridge.ExecuteTool(
			context.Background(),
			"drawio",
			tc,
		)
		if execErr != nil {
			Logln(logWriter, "Error: drawio render failed for %s: %v", img.RelPath, execErr)
			return 1
		}
		if exitCode != 0 {
			Logln(logWriter, "Error: drawio render failed for %s (exit code %d)", img.RelPath, exitCode)
			return 1
		}
		rendered++

		// Copy from cache to output
		if cpErr := CopyFile(cachePath, outputPath); cpErr != nil {
			Logln(logWriter, "Error copying rendered drawio %s: %v", img.RelPath, cpErr)
			return 1
		}
	}

	total := len(images)
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(cached) / float64(total) * 100
	}
	Logln(logWriter, "Drawio: %d cached, %d rendered (%.1f%% hit rate)", cached, rendered, hitRate)
	Logln(logWriter, "Output: %s (%d files)", outputDrawioDir, total)

	return 0
}
