// drawio.go - Build handler for rendering .drawio.png files via the drawio-oci container.
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/commands/build/docprep/caching"
	"github.com/ready-to-release/eac/go/commands/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/clibase/caching/itemcache"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&DrawioRenderHandler{})
}

// DrawioRenderHandler renders .drawio.png files via the drawio-oci container.
// Produces rendered PNGs at out/build/docs/drawio/drawio/ preserving relative paths.
// Uses .cache/eac/drawio/ for incremental builds (skip unchanged files).
type DrawioRenderHandler struct{}

func (h *DrawioRenderHandler) Name() string { return "drawio-render" }

func (h *DrawioRenderHandler) Requirements() []string { return []string{"docker"} }

func (h *DrawioRenderHandler) IsContainer() bool { return true }

func (h *DrawioRenderHandler) IsHostInstalled() bool { return false }

func (h *DrawioRenderHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for drawio rendering)")
	}
	return nil
}

func (h *DrawioRenderHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"drawio/"}
}

func (h *DrawioRenderHandler) Build(
	module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, rawOpts any,
) int {
	opts, _ := rawOpts.(BuildOptions)
	Logln(logWriter, "=== Rendering drawio images via container ===")

	docsAssetsDir := filepath.Join(workspaceRoot, "docs", "assets")
	cacheDir := filepath.Join(paths.DrawioAccelCachePath(workspaceRoot), module.GetMoniker(), filepath.Base(outputDir))
	outputDrawioDir := filepath.Join(outputDir, "drawio")

	// 1. Discover all .drawio.png source files
	images, err := diagrams.FindDrawioImages(docsAssetsDir)
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

	// 2. Convert to itemcache.Item slice and build lookup map
	imagesByKey := make(map[string]diagrams.DrawioImage, len(images))
	items := make([]itemcache.Item, 0, len(images))
	for _, img := range images {
		cacheKey := caching.DrawioCacheKey{
			SourcePath: img.SourceFile,
			SourceHash: img.Hash,
			MaxWidth:   diagrams.MaxImageWidthPDF,
		}
		hashStr := caching.HashDrawioKey(cacheKey)
		identifier := paths.SanitizeForCacheName(img.RelPath)
		shortHash := hashStr[:paths.CacheHashLength]

		item := itemcache.Item{
			Key:           img.RelPath,
			ContentHash:   shortHash,
			CacheFilename: fmt.Sprintf("%s_%s.png", identifier, shortHash),
			OutputRelPath: img.RelPath,
		}
		items = append(items, item)
		imagesByKey[img.RelPath] = img
	}

	// 3. Build function renders each miss via drawio-oci container
	buildFunc := func(misses []itemcache.Item, dir string) (int, error) {
		return h.renderMisses(misses, imagesByKey, workspaceRoot, dir, logWriter)
	}

	// 4. Execute with cache
	cache := itemcache.New(cacheDir)
	result, err := cache.Execute(items, outputDrawioDir, buildFunc, opts.ForceRebuild)
	if err != nil {
		Logln(logWriter, "Error: %v", err)
		return 1
	}

	Logln(logWriter, "Drawio: %d cached, %d rendered (%.1f%% hit rate)",
		result.CacheHits, result.Built, result.HitRate)
	Logln(logWriter, "Output: %s (%d files)", outputDrawioDir, result.TotalItems)

	return 0
}

// renderMisses renders cache-missed drawio images via the drawio-oci container.
func (h *DrawioRenderHandler) renderMisses(
	misses []itemcache.Item, imagesByKey map[string]diagrams.DrawioImage,
	workspaceRoot, cacheDir string, logWriter io.Writer,
) (int, error) {
	bridge := tool.GlobalHandlerToolBridge()
	rendered := 0

	for _, item := range misses {
		img := imagesByKey[item.Key]

		tc := &tool.ToolContext{
			WorkspaceRoot: workspaceRoot,
			OutputDir:     cacheDir,
			LogWriter:     logWriter,
		}

		containerInput, relErr := filepath.Rel(workspaceRoot, img.SourceFile)
		if relErr != nil {
			return rendered, fmt.Errorf("computing relative path for %s: %w", img.RelPath, relErr)
		}
		containerInputPath := "/docs/" + filepath.ToSlash(containerInput)
		containerOutputPath := "/output/" + item.CacheFilename

		tc.Variables = map[string]string{
			"input_path":  containerInputPath,
			"output_path": containerOutputPath,
			"max_width":   fmt.Sprintf("%d", diagrams.MaxImageWidthPDF),
		}

		exitCode, execErr := bridge.ExecuteTool(
			context.Background(),
			"drawio",
			tc,
		)
		if execErr != nil {
			return rendered, fmt.Errorf("drawio render failed for %s: %w", img.RelPath, execErr)
		}
		if exitCode != 0 {
			return rendered, fmt.Errorf("drawio render failed for %s (exit code %d)", img.RelPath, exitCode)
		}
		rendered++
	}

	return rendered, nil
}

var _ build.BuilderPort = (*DrawioRenderHandler)(nil)
