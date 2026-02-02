package mkdocs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

func init() {
	// Register the handler in the parent package's registry
	// This is done via the RegisterMkDocsHandlers function called from the builders package
}

// PreprocessHandler handles base-site preprocessing.
// This is the first stage of the MkDocs build pipeline:
// 1. Copy static files to staging
// 2. Apply link translations
// 3. Execute commands and insert outputs
// 4. Process diagrams (mermaid, structurizr, drawio)
// 5. Write manifest for cache validation
type PreprocessHandler struct {
	manifestStore *ManifestStore
}

// NewPreprocessHandler creates a new preprocessing handler.
func NewPreprocessHandler(workspaceRoot string) *PreprocessHandler {
	return &PreprocessHandler{
		manifestStore: NewManifestStore(workspaceRoot),
	}
}

// Name returns the handler identifier.
func (h *PreprocessHandler) Name() string { return "mkdocs-preprocess" }

// Requirements returns system dependencies (none - pure Go preprocessing).
func (h *PreprocessHandler) Requirements() []string { return nil }

// resolveBookName determines the book name for a base-site component.
// Resolution order:
// 1. config.book field (explicit override)
// 2. Strip "-base" suffix if present (e.g., "tutorials-base" -> "tutorials")
// 3. Use component name as book name (e.g., "site" -> "site")
// 4. Default to "site" if component name is empty
func resolveBookName(module interfaces.ModuleContractPort, componentName string) string {
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

	// Convention: strip "-base" suffix to get book name
	if strings.HasSuffix(componentName, "-base") {
		return strings.TrimSuffix(componentName, "-base")
	}

	// Fall back to component name as book name
	return componentName
}

// ValidateModule checks if a module has valid book configuration.
func (h *PreprocessHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	// Load config to check for book configuration
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := cfg.LoadBooks(false); err != nil {
		return fmt.Errorf("loading books config: %w", err)
	}

	// Resolve book name from component config or naming convention
	bookName := resolveBookName(module, component)

	// Look up book by name directly (not through module association)
	// This supports the new base-site/site-render architecture where books
	// are referenced by component config rather than book-type components
	book := cfg.GetBookByName(bookName)
	if book != nil {
		return nil
	}

	// Fallback: try module's book list for backwards compatibility
	moduleBooks := cfg.GetBooksByModule(module.GetMoniker())
	for _, b := range moduleBooks {
		if b.Name == bookName {
			return nil
		}
	}

	return fmt.Errorf("book '%s' not found (component: %s, module: %s)", bookName, component, module.GetMoniker())
}

// ListArtifacts returns artifact paths that would be produced.
func (h *PreprocessHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{"staged/", "mkdocs.yml", ".manifest.json"}
}

// IsContainer returns false as preprocessing runs in pure Go.
func (h *PreprocessHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as preprocessing uses Go only.
func (h *PreprocessHandler) IsHostInstalled() bool { return true }

// BuildOptions contains flags for controlling the build process.
type BuildOptions struct {
	// Component is the specific component to build (book name).
	Component string

	// Force disables cache and forces a full rebuild.
	Force bool

	// PDFMode enables PDF-specific preprocessing (link normalization).
	PDFMode bool

	// Tidy runs pre-build operations like cache refresh.
	Tidy bool

	// Reproducible forces rebuild for CI reproducibility.
	Reproducible bool

	// Metadata holds additional build parameters (e.g., theme for PDF).
	Metadata map[string]string

	// Weight is the resource multiplier for container builds.
	// Higher weight = more CPU and memory allocated.
	// Default: 1
	Weight int
}

// Build executes preprocessing for a base-site component.
func (h *PreprocessHandler) Build(
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

	// Load book configuration
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		logln(logWriter, "❌ Failed to load config: %v", err)
		return BuildResult{ExitCode: 1}
	}
	if err := cfg.LoadBooks(false); err != nil {
		logln(logWriter, "❌ Failed to load books config: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Resolve book name from component config or naming convention
	bookName := resolveBookName(module, opts.Component)

	// Look up book by name directly (not through module association)
	// This supports the new base-site/site-render architecture
	book := cfg.GetBookByName(bookName)
	if book == nil {
		// Fallback: try module's book list for backwards compatibility
		moduleBooks := cfg.GetBooksByModule(concrete.Moniker)
		for _, b := range moduleBooks {
			if b.Name == bookName {
				book = b
				break
			}
		}
	}
	if book == nil {
		logln(logWriter, "❌ Book '%s' not found (component: %s, module: %s)", bookName, opts.Component, concrete.Moniker)
		return BuildResult{ExitCode: 1}
	}

	logln(logWriter, "\n=== Preprocessing base-site: %s/%s ===", concrete.Moniker, bookName)

	// Compute input hash for cache check
	inputHash, err := h.computeInputHash(book, workspaceRoot)
	if err != nil {
		logln(logWriter, "   ⚠️  Failed to compute input hash: %v", err)
		// Continue without caching
	}

	// Check cache - skip if unchanged (unless force rebuild)
	// Use opts.Component (not bookName) to match how we save the manifest
	if !opts.Force && !opts.Reproducible && inputHash != "" {
		manifest, err := h.manifestStore.Load(concrete.Moniker, opts.Component)
		if err == nil && manifest != nil && manifest.InputHash == inputHash {
			// Check if staged directory exists
			stagingDir := paths.BookStagingCachePath(workspaceRoot, concrete.Moniker, bookName)
			if _, statErr := os.Stat(stagingDir); statErr == nil {
				logln(logWriter, "⏭️  Preprocessing skipped (cached, hash: %s...)", inputHash[:8])
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

	// Determine if this is PDF mode from book output config
	pdfMode := opts.PDFMode
	if !pdfMode {
		output := book.GetOutput()
		pdfMode = output == "pdf-dark" || output == "pdf-light" || output == "pdf-all"
	}

	// Setup staging directory
	stagingDir := paths.BookStagingCachePath(workspaceRoot, concrete.Moniker, bookName)

	// Only clean staging directory on force rebuild
	if opts.Force {
		logln(logWriter, "🔄 Force rebuild: clearing staging directory")
		if err := os.RemoveAll(stagingDir); err != nil && !os.IsNotExist(err) {
			logln(logWriter, "❌ Failed to clean staging directory: %v", err)
			return BuildResult{ExitCode: 1}
		}
	}

	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Run preprocessing pipeline
	preprocessor := books.NewPreprocessor(book, workspaceRoot, stagingDir, logWriter, pdfMode)
	if err := preprocessor.Preprocess(); err != nil {
		logln(logWriter, "❌ Preprocessing failed: %v", err)
		return BuildResult{ExitCode: 1}
	}

	// Generate mkdocs.yml config
	configPath := filepath.Join(outputDir, "mkdocs.yml")
	relStagingDir, err := filepath.Rel(outputDir, stagingDir)
	if err != nil {
		relStagingDir = stagingDir
	}
	relStagingDir = filepath.ToSlash(relStagingDir)

	outputFormat := "site"
	if pdfMode {
		outputFormat = "pdf-dark" // default theme for preprocessing
	}

	configOpts := books.ConfigOptions{
		SiteName:     book.Name,
		DocsDir:      relStagingDir,
		OutputFormat: outputFormat,
	}
	if err := books.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return BuildResult{ExitCode: 1}
	}
	logln(logWriter, "   Config: %s", configPath)

	// Compute output hash
	outputHash, err := computeStagingHash(stagingDir)
	if err != nil {
		logln(logWriter, "   ⚠️  Failed to compute output hash: %v", err)
		outputHash = "" // Continue without caching
	}

	// Write manifest - use opts.Component (not bookName) so site-render can find it
	manifest := &ComponentManifest{
		Version:    ManifestVersion,
		Type:       ComponentTypeBaseSite,
		Module:     concrete.Moniker,
		Component:  opts.Component,
		InputHash:  inputHash,
		OutputHash: outputHash,
		Timestamp:  time.Now().UTC(),
		Duration:   time.Since(startTime),
		Artifacts:  []string{"staged/", "mkdocs.yml"},
		Metadata: map[string]string{
			"book":     book.Name,
			"pdf_mode": fmt.Sprintf("%v", pdfMode),
		},
	}

	if err := h.manifestStore.Save(manifest); err != nil {
		logln(logWriter, "   ⚠️  Failed to save manifest: %v", err)
	}

	logln(logWriter, "✅ Preprocessing complete (hash: %s...)", outputHash[:min(8, len(outputHash))])

	return BuildResult{
		ExitCode:  0,
		Cached:    false,
		Manifest:  manifest,
		Artifacts: manifest.Artifacts,
		Duration:  time.Since(startTime),
	}
}

// computeInputHash computes a SHA256 hash of all inputs for cache validation.
func (h *PreprocessHandler) computeInputHash(book *config.Book, workspaceRoot string) (string, error) {
	hasher := sha256.New()

	// 1. Hash book name and config (deterministic via sorted keys would be ideal,
	// but for now we hash the name which is the primary identifier)
	hasher.Write([]byte(book.Name))
	if book.Title != "" {
		hasher.Write([]byte(book.Title))
	}
	if book.Description != "" {
		hasher.Write([]byte(book.Description))
	}

	// 2. Hash source file contents
	copySources := book.GetCopySources()
	var allFiles []string

	for _, src := range copySources {
		srcDir := filepath.Join(workspaceRoot, filepath.Dir(src.From))
		pattern := filepath.Base(src.From)

		// Simple glob matching - in production this would use the full glob logic
		err := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			matched, _ := filepath.Match(pattern, d.Name())
			if matched || pattern == "**" || pattern == "*" {
				allFiles = append(allFiles, path)
			}
			return nil
		})
		if err != nil {
			continue // Skip sources we can't read
		}
	}

	// Sort for deterministic order
	sort.Strings(allFiles)

	// Hash file contents (limit to avoid memory issues)
	const maxFilesToHash = 1000
	for i, path := range allFiles {
		if i >= maxFilesToHash {
			break
		}
		relPath, _ := filepath.Rel(workspaceRoot, path)
		hasher.Write([]byte(relPath))

		// Hash file content
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		_, _ = io.Copy(hasher, f)
		f.Close()
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// computeStagingHash computes a SHA256 hash of all files in the staging directory.
func computeStagingHash(stagingDir string) (string, error) {
	hasher := sha256.New()

	var files []string
	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	// Sort for deterministic order
	sort.Strings(files)

	for _, path := range files {
		relPath, _ := filepath.Rel(stagingDir, path)
		hasher.Write([]byte(relPath))

		f, err := os.Open(path)
		if err != nil {
			continue
		}
		_, _ = io.Copy(hasher, f)
		f.Close()
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// logln writes a formatted log line.
func logln(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, format+"\n", args...)
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
