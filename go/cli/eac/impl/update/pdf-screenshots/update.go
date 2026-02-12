package pdfscreenshots

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type updatePDFScreenshotsCommand struct{}

var _ core.SimpleCommandPort = (*updatePDFScreenshotsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updatePDFScreenshotsCommand{},
	}
}

func (c *updatePDFScreenshotsCommand) Name() string { return "update pdf-screenshots" }

func (c *updatePDFScreenshotsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-pdf-screenshots",
		Short:         "Extract PDF pages as images for documentation cache",
		Long:          "Scans out/build folders for generated PDF books and extracts\neach page as a PNG image. Images are stored in .cache/eac/pdf-screenshots/\norganized by book name with hash marker for cache invalidation.\n\nExpected Output:\n  - PNG images in .cache/eac/pdf-screenshots/ directory\n  - Organized by book name (one subdirectory per PDF)\n  - Hash marker files for cache validation",
		Flags: []core.FlagSpec{
			{Name: "dry-run", Type: "bool", DefaultValue: "false", Usage: "Show what would be done without making changes"},
			{Name: "force", Shorthand: "f", Type: "bool", DefaultValue: "false", Usage: "Regenerate all images ignoring cache"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show detailed progress"},
			{Name: "dpi", Type: "int", DefaultValue: "150", Usage: "Image resolution (72-300)"},
			{Name: "module", Shorthand: "m", Type: "string", Usage: "Process only specific module's PDFs"},
		},
	}
}

func (c *updatePDFScreenshotsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdatePDFScreenshots()
}

var log = logging.C()

var rePageNumber = regexp.MustCompile(`^page-(\d+)\.png$`)

const (
	// pdfToolsImage is the Docker image for PDF operations.
	pdfToolsImage = "pdf-cli-oci:latest"

	// defaultDPI is the default resolution for extracted images.
	defaultDPI = 150
)

// PDFInfo holds information about a found PDF file.
type PDFInfo struct {
	Path     string // Absolute path to PDF
	RelPath  string // Relative path from out/build
	Module   string // Module name (first component of RelPath)
	BookName string // Base name of PDF without extension (for cache dir naming)
	Hash     string // SHA256 hash of file content (first 12 chars)
	CacheDir string // Path to cache directory for this PDF
}

// pdfConfig holds the parsed command configuration.
type pdfConfig struct {
	dryRun       bool
	force        bool
	verbose      bool
	dpi          int
	moduleFilter string
}

// UpdatePDFScreenshots scans out/build for PDFs and extracts pages as images.
func UpdatePDFScreenshots() int {
	cfg, err := parsePDFConfig()
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Get repo root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	fmt.Println("Updating PDF screenshot cache...")

	// Scan for PDFs in out/build/
	buildDir := filepath.Join(repoRoot, paths.OutBuildRelPath)
	pdfs, err := scanForPDFs(buildDir, cfg.moduleFilter)
	if err != nil {
		log.Errorf("Error scanning for PDFs: %v", err)
		return 1
	}

	if len(pdfs) == 0 {
		fmt.Println("No PDFs found in out/build/")
		if cfg.moduleFilter != "" {
			fmt.Printf("  (filtered by module: %s)\n", cfg.moduleFilter)
		}
		return 0
	}

	fmt.Printf("Found %d PDF(s)\n", len(pdfs))
	if cfg.verbose {
		for _, pdf := range pdfs {
			fmt.Printf("  - %s\n", pdf.RelPath)
		}
	}

	// Calculate hashes and check cache
	cacheRoot := paths.PDFScreenshotsCachePath(repoRoot)
	toProcess := checkCache(pdfs, cacheRoot, cfg)

	if len(toProcess) == 0 {
		fmt.Println("All PDFs are cached. Nothing to extract.")
		return 0
	}

	if cfg.dryRun {
		fmt.Println("\n[DRY RUN] Would extract pages from:")
		for _, pdf := range toProcess {
			fmt.Printf("  - %s -> %s/\n", pdf.RelPath, pdf.BookName)
		}
		return 0
	}

	return processExtraction(toProcess, cacheRoot, repoRoot, cfg)
}

// parsePDFConfig parses command flags into a pdfConfig.
func parsePDFConfig() (*pdfConfig, error) {
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	cfg := &pdfConfig{
		dpi: defaultDPI,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			cfg.dryRun = true
		case "--force", "-f":
			cfg.force = true
		case "-v", "--verbose":
			cfg.verbose = true
		case "--dpi":
			if i+1 < len(args) {
				i++
				if d, err := strconv.Atoi(args[i]); err == nil && d >= 72 && d <= 300 {
					cfg.dpi = d
				}
			}
		case "-m", "--module":
			if i+1 < len(args) {
				i++
				cfg.moduleFilter = args[i]
			}
		}
	}

	return cfg, nil
}

// checkCache computes hashes and identifies PDFs that need extraction.
func checkCache(pdfs []PDFInfo, cacheRoot string, cfg *pdfConfig) []PDFInfo {
	cacheHits := 0
	cacheMisses := 0
	var toProcess []PDFInfo

	for i := range pdfs {
		pdf := &pdfs[i]
		pdf.Hash = hashFile(pdf.Path)
		pdf.CacheDir = filepath.Join(cacheRoot, pdf.BookName)

		if !cfg.force && cacheValidWithHash(pdf.CacheDir, pdf.Hash) {
			cacheHits++
			if cfg.verbose {
				fmt.Printf("  [cached] %s\n", pdf.RelPath)
			}
		} else {
			cacheMisses++
			toProcess = append(toProcess, *pdf)
		}
	}

	fmt.Printf("Cache status: %d hits, %d misses", cacheHits, cacheMisses)
	if len(pdfs) > 0 {
		fmt.Printf(" (%.1f%% hit rate)", float64(cacheHits)/float64(len(pdfs))*100)
	}
	fmt.Println()

	return toProcess
}

// processExtraction handles Docker setup and PDF page extraction.
func processExtraction(toProcess []PDFInfo, cacheRoot, repoRoot string, cfg *pdfConfig) int {
	// Check Docker availability
	if !docker.IsDockerAvailable() {
		fmt.Fprintln(os.Stderr, "Error: Docker is not available but required for PDF extraction")
		fmt.Fprintln(os.Stderr, "Ensure Docker is installed and the daemon is running")
		return 1
	}

	// Create Docker client
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Errorf("Error creating Docker client: %v", err)
		return 1
	}
	defer dockerClient.Close()

	// Build the pdf-cli-oci image if needed
	fmt.Println("Ensuring pdf-cli-oci Docker image...")
	if err := ensurePDFToolsImage(repoRoot); err != nil {
		log.Errorf("Error building pdf-cli-oci image: %v", err)
		return 1
	}

	// Ensure cache root exists
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		log.Errorf("Error creating cache directory: %v", err)
		return 1
	}

	// Process each PDF
	fmt.Printf("Extracting pages from %d PDF(s)...\n", len(toProcess))
	extracted := 0
	failed := 0

	for _, pdf := range toProcess {
		if cfg.verbose {
			fmt.Printf("  Processing %s...\n", pdf.RelPath)
		}

		// Clear old cache if exists (stale hash)
		if _, err := os.Stat(pdf.CacheDir); err == nil {
			if cfg.verbose {
				fmt.Printf("  Clearing stale cache for %s\n", pdf.BookName)
			}
			_ = os.RemoveAll(pdf.CacheDir)
		}

		// Create cache directory for this PDF
		if err := os.MkdirAll(pdf.CacheDir, 0o755); err != nil {
			log.Errorf("  Failed to create cache dir for %s: %v", pdf.RelPath, err)
			failed++
			continue
		}

		// Extract pages using pdftoppm
		if err := extractPages(dockerClient, pdf.Path, pdf.CacheDir, cfg.dpi); err != nil {
			log.Errorf("  Failed to extract %s: %v", pdf.RelPath, err)
			_ = os.RemoveAll(pdf.CacheDir)
			failed++
			continue
		}

		// Write hash marker file
		hashMarker := filepath.Join(pdf.CacheDir, pdf.Hash+".cache")
		if err := os.WriteFile(hashMarker, []byte(pdf.Hash), 0o644); err != nil {
			log.Errorf("  Failed to write cache marker for %s: %v", pdf.RelPath, err)
		}

		// Count extracted pages
		pageCount := countPages(pdf.CacheDir)
		extracted++

		if cfg.verbose {
			fmt.Printf("  %s: %d pages -> %s/\n", pdf.RelPath, pageCount, pdf.BookName)
		}
	}

	// Summary
	fmt.Println()
	if failed > 0 {
		fmt.Printf("Completed with errors: %d extracted, %d failed\n", extracted, failed)
	} else {
		fmt.Printf("Successfully extracted pages from %d PDF(s)\n", extracted)
	}
	fmt.Printf("Cache location: %s\n", cacheRoot)

	if failed > 0 {
		return 1
	}
	return 0
}
