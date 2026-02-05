package mkdocs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

// PDFRenderHandler handles PDF generation from base-site output.
// This is an alternative render stage (parallel to site-render):
// 1. Load base-site manifest to get staged content location
// 2. Run MkDocs with PDF export in Docker
// 3. Merge individual PDFs into single document
// 4. Write manifest for cache validation
type PDFRenderHandler struct {
	manifestStore *ManifestStore
}

// NewPDFRenderHandler creates a new PDF render handler.
func NewPDFRenderHandler(workspaceRoot string) *PDFRenderHandler {
	return &PDFRenderHandler{
		manifestStore: NewManifestStore(workspaceRoot),
	}
}

// Name returns the handler identifier.
func (h *PDFRenderHandler) Name() string { return "pdf-oci" }

// Requirements returns system dependencies (Docker required).
func (h *PDFRenderHandler) Requirements() []string { return []string{"docker"} }

// ValidateModule checks if a module has valid base-site dependency.
func (h *PDFRenderHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	// Resolve the base-site component this pdf-render depends on
	baseSiteComp := resolveBaseSiteComponent(module, component, h.manifestStore)

	// Check if base-site manifest exists
	manifest, err := h.manifestStore.Load(module.GetMoniker(), baseSiteComp)
	if err != nil {
		return fmt.Errorf("failed to load base-site manifest for %s: %w", baseSiteComp, err)
	}
	if manifest == nil {
		return fmt.Errorf("base-site '%s' not built for module %s (component: %s)", baseSiteComp, module.GetMoniker(), component)
	}
	return nil
}

// ListArtifacts returns artifact paths that would be produced.
// PDFs are output to site/pdf/ directory by mkdocs-exporter.
func (h *PDFRenderHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{"site/pdf/"}
}

// IsContainer returns true as PDF rendering requires Docker.
func (h *PDFRenderHandler) IsContainer() bool { return true }

// IsHostInstalled returns false as PDF rendering uses Docker.
func (h *PDFRenderHandler) IsHostInstalled() bool { return false }

// Build executes PDF generation from base-site output.
func (h *PDFRenderHandler) Build(
	module interfaces.ModuleContractPort,
	workspaceRoot, outputDir string,
	logWriter io.Writer,
	opts BuildOptions,
) BuildResult {
	startTime := time.Now()

	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		logln(logWriter, "Error: invalid module type")
		return BuildResult{ExitCode: 1}
	}

	// Resolve base-site component this pdf-render depends on
	baseSiteComponent := resolveBaseSiteComponent(module, opts.Component, h.manifestStore)

	// Resolve book name from the base-site component config
	bookName := resolveBookName(module, baseSiteComponent)

	// Determine theme (default to dark)
	theme := "dark"
	if md, ok := opts.Metadata["theme"]; ok {
		theme = md
	}

	logln(logWriter, "\n=== Rendering PDF: %s/%s (%s theme, base: %s) ===", concrete.Moniker, bookName, theme, baseSiteComponent)

	// Load base-site manifest
	baseManifest, err := h.manifestStore.Load(concrete.Moniker, baseSiteComponent)
	if err != nil {
		logln(logWriter, "❌ Failed to load base-site manifest: %v", err)
		return BuildResult{ExitCode: 1}
	}
	if baseManifest == nil {
		logln(logWriter, "❌ Base-site not built - run preprocessing first")
		return BuildResult{ExitCode: 1}
	}

	// Compute input hash for cache check
	containerHash := getContainerImageHash(workspaceRoot, "pdf-oci")
	inputHash := computePDFRenderInputHash(baseManifest.OutputHash, theme, containerHash)

	// Check cache - skip if base-site and theme unchanged
	pdfPath := filepath.Join(outputDir, fmt.Sprintf("%s-%s.pdf", bookName, theme))
	if !opts.Force && !opts.Reproducible {
		componentName := fmt.Sprintf("%s-%s", bookName, theme)
		manifest, err := h.manifestStore.Load(concrete.Moniker, componentName)
		if err == nil && manifest != nil && manifest.InputHash == inputHash {
			if _, statErr := os.Stat(pdfPath); statErr == nil {
				logln(logWriter, "⏭️  PDF rendering skipped (cached, hash: %s...)", inputHash[:8])
				return BuildResult{
					ExitCode:  ExitCodeSkipped,
					Cached:    true,
					Manifest:  manifest,
					Artifacts: manifest.Artifacts,
					Duration:  time.Since(startTime),
				}
			}
		}
	}

	// Resolve staging directory from base-site manifest
	// Use the book name from manifest metadata (staging is keyed by book, not component)
	stagingBookName := baseManifest.Metadata["book"]
	if stagingBookName == "" {
		stagingBookName = baseSiteComponent // fallback
	}
	stagingDir := paths.BookStagingCachePath(workspaceRoot, concrete.Moniker, stagingBookName)
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		logln(logWriter, "❌ Staging directory not found: %s", stagingDir)
		return BuildResult{ExitCode: 1}
	}

	// Load book config for title/description
	var bookTitle, bookDescription string
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err == nil {
		_ = cfg.LoadBooks(false)
		moduleBooks := cfg.GetBooksByModule(concrete.Moniker)
		for _, b := range moduleBooks {
			if b.Name == bookName {
				bookTitle = b.Title
				bookDescription = b.Description
				break
			}
		}
	}
	if bookTitle == "" {
		bookTitle = bookName
	}

	// Generate mkdocs.yml config for PDF
	// Use container mount path /staging (where staging dir is mounted in Docker)
	configPath := filepath.Join(outputDir, "mkdocs.yml")

	pdfConcurrency := environments.GetPDFExportConcurrency()
	pageLimit := opts.ArtifactsMode.PDFPageLimit()
	configOpts := books.ConfigOptions{
		SiteName:        bookName,
		SiteDescription: fmt.Sprintf("Generated PDF documentation for %s", bookName),
		BookTitle:       bookTitle,
		BookDescription: bookDescription,
		DocsDir:         "/staging", // Container mount path for staging directory
		Theme:           theme,
		OutputFormat:    fmt.Sprintf("pdf-%s", theme),
		PDFConcurrency:  pdfConcurrency,
		PageLimit:       pageLimit,
	}
	if err := books.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Create output directories
	siteDir := filepath.Join(outputDir, "site")
	pdfDir := filepath.Join(siteDir, "pdf")
	if err := os.MkdirAll(pdfDir, 0o755); err != nil {
		logln(logWriter, "❌ Failed to create output directory: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Run MkDocs PDF build via tool bridge
	bridge := tool.GlobalHandlerToolBridge()

	// Get weight for resource scaling (default 1)
	weight := opts.Weight
	if weight <= 0 {
		weight = 1
	}

	tc := &tool.ToolContext{
		WorkspaceRoot: workspaceRoot,
		StagingDir:    stagingDir,
		OutputDir:     outputDir,
		ConfigPath:    configPath,
		LogWriter:     logWriter,
		Weight:        weight,
		Variables: map[string]string{
			"theme": theme,
		},
	}

	// Retry logic for PDF builds - Playwright can have transient timeouts
	maxRetries := 2
	var exitCode int
	var execErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			logln(logWriter, "🔄 Retrying PDF build (attempt %d/%d)...", attempt, maxRetries)
		}

		exitCode, execErr = bridge.ExecuteTool(context.Background(), "pdf-oci", tc)
		if execErr == nil && exitCode == 0 {
			break
		}

		if attempt < maxRetries {
			logln(logWriter, "⚠️  PDF build failed, will retry...")
		}
	}

	if execErr != nil {
		logln(logWriter, "❌ Tool execution failed: %v", execErr)
		return BuildResult{ExitCode: 1}
	}
	if exitCode != 0 {
		logln(logWriter, "❌ MkDocs PDF build failed with exit code %d", exitCode)
		return BuildResult{ExitCode: exitCode}
	}

	// Move PDF to final location
	// The exporter plugin generates combined.pdf in the pdf/ subdirectory
	srcPdf := filepath.Join(pdfDir, "combined.pdf")
	if _, err := os.Stat(srcPdf); os.IsNotExist(err) {
		logln(logWriter, "❌ PDF not generated at expected path: %s", srcPdf)
		return BuildResult{ExitCode: 1}
	}

	if err := copyFileIfDifferent(srcPdf, pdfPath); err != nil {
		logln(logWriter, "❌ Failed to copy PDF: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Note: Cover, TOC, and page numbers are now added by mkdocs-exporter's aggregator
	// inside the container - no separate merge step needed

	// Compute output hash
	outputHash, _ := computeFileHashPath(pdfPath)

	// Write manifest
	componentName := fmt.Sprintf("%s-%s", bookName, theme)
	manifest := &ComponentManifest{
		Version:    ManifestVersion,
		Type:       ComponentTypePDFRender,
		Module:     concrete.Moniker,
		Component:  componentName,
		InputHash:  inputHash,
		OutputHash: outputHash,
		Timestamp:  time.Now().UTC(),
		Duration:   time.Since(startTime),
		Dependencies: map[string]string{
			"base-site": baseManifest.OutputHash,
		},
		Artifacts: []string{filepath.Base(pdfPath)},
		Metadata: map[string]string{
			"theme":     theme,
			"book_name": bookName,
		},
	}

	if err := h.manifestStore.Save(manifest); err != nil {
		logln(logWriter, "   ⚠️  Failed to save manifest: %v", err)
	}

	logln(logWriter, "✅ PDF rendering complete")
	logln(logWriter, "   PDF Output: %s", pdfPath)

	return BuildResult{
		ExitCode:  0,
		Cached:    false,
		Manifest:  manifest,
		Artifacts: manifest.Artifacts,
		Duration:  time.Since(startTime),
	}
}

// computePDFRenderInputHash computes the cache key for PDF rendering.
func computePDFRenderInputHash(baseSiteOutputHash, theme, containerHash string) string {
	return fmt.Sprintf("%s:%s:%s",
		baseSiteOutputHash[:min(16, len(baseSiteOutputHash))],
		theme,
		containerHash[:min(8, len(containerHash))])
}


// copyFileIfDifferent copies src to dst if they differ.
func copyFileIfDifferent(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// computeFileHashPath computes SHA256 hash of a file.
func computeFileHashPath(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return computeFileHash(data)
}
