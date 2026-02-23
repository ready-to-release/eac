// pdf.go - Unified PDF build handler
// Combines preprocessing and container rendering in a single handler.
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/commands/build/builders/mkdocs"
	"github.com/ready-to-release/eac/go/commands/build/docprep"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&PDFHandler{})
}

// PDFHandler is a unified handler for docs-pdf component type.
// It combines preprocessing (native Go) and PDF rendering (container)
// in a single handler, eliminating the need for tool_chain orchestration.
type PDFHandler struct{}

func (h *PDFHandler) Name() string { return "pdf" }

func (h *PDFHandler) Capabilities() []string { return []string{"documentation", "pdf", "container"} }

func (h *PDFHandler) Requirements() []string { return []string{"docker"} }

// IsContainer returns true as PDF generation requires Docker for the rendering step.
func (h *PDFHandler) IsContainer() bool { return true }

// IsHostInstalled returns false as PDF generation uses Docker.
func (h *PDFHandler) IsHostInstalled() bool { return false }

// ValidateModule checks if a module has valid book configuration for PDF generation.
func (h *PDFHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	// Check Docker availability
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available but required for PDF builds")
	}

	// Load config to check for book configuration
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := cfg.LoadBooks(false); err != nil {
		return fmt.Errorf("loading books config: %w", err)
	}

	// Resolve book name from component config or naming convention
	bookName := resolveBookNameForPDF(module, component)

	// Look up book by name
	book := cfg.GetBookByName(bookName)
	if book != nil {
		return nil
	}

	// Fallback: try module's book list
	moduleBooks := cfg.GetBooksByModule(module.GetMoniker())
	for _, b := range moduleBooks {
		if b.Name == bookName {
			return nil
		}
	}

	return fmt.Errorf("book '%s' not found for PDF generation (component: %s, module: %s)", bookName, component, module.GetMoniker())
}

// ListArtifacts returns artifact paths that would be produced.
// PDFs are output to site/pdf/ directory by mkdocs-exporter.
func (h *PDFHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"site/pdf/"}
}

// Build executes the unified PDF build: preprocessing + container rendering.
func (h *PDFHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
	startTime := time.Now()

	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}

	// Load book and repository configuration
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		Logln(logWriter, "❌ Failed to load config: %v", err)
		return 1
	}
	if err := cfg.LoadRepository(false); err != nil {
		Logln(logWriter, "❌ Failed to load repository config: %v", err)
		return 1
	}
	if err := cfg.LoadBooks(false); err != nil {
		Logln(logWriter, "❌ Failed to load books config: %v", err)
		return 1
	}

	// Resolve book name from component config or naming convention
	bookName := resolveBookNameForPDF(module, opts.Component)

	// Look up book
	book := cfg.GetBookByName(bookName)
	if book == nil {
		moduleBooks := cfg.GetBooksByModule(concrete.Moniker)
		for _, b := range moduleBooks {
			if b.Name == bookName {
				book = b
				break
			}
		}
	}
	if book == nil {
		Logln(logWriter, "❌ Book '%s' not found (component: %s, module: %s)", bookName, opts.Component, concrete.Moniker)
		return 1
	}

	// Determine theme from book output configuration
	// - pdf-dark: dark theme (default)
	// - pdf-light: light theme
	// - pdf-all: both themes (handled by orchestrator creating multiple UoWs)
	theme := "dark"
	output := book.GetOutput()
	if output == "pdf-light" {
		theme = "light"
	}
	// For pdf-all, the orchestrator will create separate UoWs with component names
	// like "tutorials-dark" and "tutorials-light" - we extract theme from component name
	if output == "pdf-all" && opts.Component != "" {
		if len(opts.Component) > 6 && opts.Component[len(opts.Component)-6:] == "-light" {
			theme = "light"
		}
		// Default to dark if component doesn't end with -light
	}

	Logln(logWriter, "\n=== Building PDF: %s/%s (%s theme) ===", concrete.Moniker, bookName, theme)

	// ━━━ Step 1: Preprocessing (native Go) ━━━
	Logln(logWriter, "📝 Running preprocessing...")

	stagingDir := paths.BookStagingCachePath(workspaceRoot, concrete.Moniker, bookName)

	// Clean staging on force rebuild
	if opts.ForceRebuild {
		Logln(logWriter, "   🔄 Force rebuild: clearing staging directory")
		if err := os.RemoveAll(stagingDir); err != nil && !os.IsNotExist(err) {
			Logln(logWriter, "❌ Failed to clean staging directory: %v", err)
			return 1
		}
	}

	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return 1
	}

	// Run preprocessing pipeline (PDF mode for link normalization)
	pctx := docprep.NewPreprocessContext(context.Background(), book, workspaceRoot, stagingDir, logWriter, docprep.PDFMode{ThemeName: "dark"})
	pctx.Moniker = concrete.Moniker
	pctx.DesignComponents = cfg.Repository.DesignComponents()
	pctx.DiagramComponents = cfg.Repository.DiagramComponents(concrete.Moniker)
	if err := docprep.DefaultPipeline().Execute(pctx); err != nil {
		Logln(logWriter, "❌ Preprocessing failed: %v", err)
		return 1
	}

	preprocessDuration := time.Since(startTime)
	Logln(logWriter, "   ✅ Preprocessing complete (%v)", preprocessDuration.Round(time.Millisecond))

	// ━━━ Step 2: Generate mkdocs.yml config ━━━
	configPath := filepath.Join(outputDir, "mkdocs.yml")

	pdfConcurrency := environments.GetPDFExportConcurrency()
	pageLimit := opts.ArtifactsMode.PDFPageLimit()
	configOpts := mkdocs.ConfigOptions{
		SiteName:        bookName,
		SiteDescription: fmt.Sprintf("Generated PDF documentation for %s", bookName),
		BookTitle:       book.Title,
		BookDescription: book.Description,
		DocsDir:         "/staging", // Container mount path
		Theme:           theme,
		OutputFormat:    fmt.Sprintf("pdf-%s", theme),
		PDFConcurrency:  pdfConcurrency,
		PageLimit:       pageLimit,
	}
	if err := mkdocs.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		Logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return 1
	}
	Logln(logWriter, "   Config: %s", configPath)

	// Create output directories
	siteDir := filepath.Join(outputDir, "site")
	pdfDir := filepath.Join(siteDir, "pdf")
	if err := os.MkdirAll(pdfDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// ━━━ Step 3: Invoke pdf-oci container ━━━
	Logln(logWriter, "📄 Invoking PDF render container...")
	Logln(logWriter, "   Theme: %s", theme)
	Logln(logWriter, "   Concurrency: %d (environment: %s)", pdfConcurrency, environments.DetectRuntime())

	// Ensure the pdf-oci container image is available.
	// In devbox: builds from Dockerfile if stale. In CI: pulls from GHCR and tags as local.
	pdfImageName := "pdf-oci:local"
	contextPath := filepath.Join(workspaceRoot, "containers", "pdf-oci")
	if err := ensureMkDocsImage(pdfImageName, workspaceRoot, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure pdf-oci image: %v", err)
		return 1
	}

	bridge := tool.GlobalHandlerToolBridge()

	// Get weight for resource scaling
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
			Logln(logWriter, "🔄 Retrying PDF build (attempt %d/%d)...", attempt, maxRetries)
		}

		exitCode, execErr = bridge.ExecuteTool(context.Background(), "pdf-oci", tc)
		if execErr == nil && exitCode == 0 {
			break
		}

		if attempt < maxRetries {
			Logln(logWriter, "⚠️  PDF build failed, will retry...")
		}
	}

	if execErr != nil {
		Logln(logWriter, "❌ Tool execution failed: %v", execErr)
		return 1
	}
	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs PDF build failed with exit code %d after %d attempts", exitCode, maxRetries)
		return exitCode
	}

	// ━━━ Step 4: Merge PDFs into final document ━━━
	pdfPath := filepath.Join(pdfDir, fmt.Sprintf("%s-%s.pdf", bookName, theme))

	// Use book title if available, fallback to book name
	bookTitle := book.Title
	if bookTitle == "" {
		bookTitle = bookName
	}

	Logln(logWriter, "📄 Merging individual PDFs...")

	hostRepoRoot := ResolveHostRepoRoot(workspaceRoot, logWriter)
	isDinD := IsDockerInDocker()
	imageName := "pdf-oci:local"

	if err := MergePDFs(siteDir, pdfPath, hostRepoRoot, workspaceRoot, stagingDir, imageName, bookTitle, book.Description, logWriter, isDinD); err != nil {
		Logln(logWriter, "❌ PDF merge failed: %v", err)
		return 1
	}

	// ━━━ Step 5: Copy PDF to module output root ━━━
	// The old MkDocs handler moved PDFs to the module root (out/build/{module}/)
	// for easy access. Replicate that behavior here.
	moduleOutputDir := paths.BuildOutputPath(workspaceRoot, concrete.Moniker)
	dstPdf := filepath.Join(moduleOutputDir, fmt.Sprintf("%s-%s.pdf", bookName, theme))
	if err := copyFile(pdfPath, dstPdf); err != nil {
		Logln(logWriter, "⚠️  Failed to copy PDF to module root: %v", err)
	} else {
		Logln(logWriter, "   📄 %s-%s.pdf → module output root", bookName, theme)
	}

	totalDuration := time.Since(startTime)
	Logln(logWriter, "✅ PDF build complete (%s theme)", theme)
	Logln(logWriter, "   PDF Output: %s", dstPdf)
	Logln(logWriter, "   Duration: %v", totalDuration.Round(time.Millisecond))

	return 0
}

// resolveBookNameForPDF determines the book name for a PDF component.
// Resolution order:
// 1. config.book field (explicit override)
// 2. Strip "-pdf" suffix if present (e.g., "tutorials-pdf" -> "tutorials")
// 3. Use component name as book name
// 4. Default to "site" if component name is empty
func resolveBookNameForPDF(module core.ModuleContractPort, componentName string) string {
	if componentName == "" {
		return "site"
	}

	// Try to get book name from component config
	concrete := adapters.UnwrapModule(module)
	if concrete != nil {
		if comp, ok := concrete.Components[componentName]; ok && comp != nil {
			if bookName, ok := comp.Config["book"]; ok && bookName != "" {
				return bookName
			}
		}
	}

	// Convention: strip "-pdf" suffix to get book name
	// e.g., "tutorials-pdf" -> "tutorials"
	if len(componentName) > 4 && componentName[len(componentName)-4:] == "-pdf" {
		return componentName[:len(componentName)-4]
	}

	// Fall back to component name as book name
	return componentName
}

var _ build.BuilderPort = (*PDFHandler)(nil)
