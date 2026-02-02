// Command: update pdf-screenshots
// Short: Extract PDF pages as images for documentation cache
// Long: Scans out/build folders for generated PDF books and extracts
// Long: each page as a PNG image. Images are stored in .cache/eac/pdf-screenshots/
// Long: organized by book name with hash marker for cache invalidation.
// Long:
// Long: Expected Output:
// Long:   - PNG images in .cache/eac/pdf-screenshots/ directory
// Long:   - Organized by book name (one subdirectory per PDF)
// Long:   - Hash marker files for cache validation
// Flag.dry-run: type=bool, default=false, usage=Show what would be done without making changes
// Flag.force: type=bool, shorthand=f, default=false, usage=Regenerate all images ignoring cache
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed progress
// Flag.dpi: type=int, default=150, usage=Image resolution (72-300)
// Flag.module: type=string, shorthand=m, usage=Process only specific module's PDFs
package pdfscreenshots

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

// commandFlags defines valid flags for the update pdf-screenshots command

var log = logging.C()

const (
	// pdfToolsImage is the Docker image for PDF operations.
	pdfToolsImage = "pdf-tool:latest"

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

func init() {
	registry.Register(UpdatePDFScreenshots)
}

// UpdatePDFScreenshots scans out/build for PDFs and extracts pages as images.
func UpdatePDFScreenshots() int {
	// Validate flags
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	dryRun := false
	force := false
	verbose := false
	dpi := defaultDPI
	moduleFilter := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		case "--force", "-f":
			force = true
		case "-v", "--verbose":
			verbose = true
		case "--dpi":
			if i+1 < len(args) {
				i++
				if d, err := strconv.Atoi(args[i]); err == nil && d >= 72 && d <= 300 {
					dpi = d
				}
			}
		case "-m", "--module":
			if i+1 < len(args) {
				i++
				moduleFilter = args[i]
			}
		}
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
	pdfs, err := scanForPDFs(buildDir, moduleFilter)
	if err != nil {
		log.Errorf("Error scanning for PDFs: %v", err)
		return 1
	}

	if len(pdfs) == 0 {
		fmt.Println("No PDFs found in out/build/")
		if moduleFilter != "" {
			fmt.Printf("  (filtered by module: %s)\n", moduleFilter)
		}
		return 0
	}

	fmt.Printf("Found %d PDF(s)\n", len(pdfs))
	if verbose {
		for _, pdf := range pdfs {
			fmt.Printf("  - %s\n", pdf.RelPath)
		}
	}

	// Calculate hashes and check cache
	cacheRoot := paths.PDFScreenshotsCachePath(repoRoot)
	cacheHits := 0
	cacheMisses := 0
	toProcess := []PDFInfo{}

	for i := range pdfs {
		pdf := &pdfs[i]
		pdf.Hash = hashFile(pdf.Path)
		// Cache dir is just the book name
		pdf.CacheDir = filepath.Join(cacheRoot, pdf.BookName)

		// Check if cache is valid (has correct hash marker)
		if !force && cacheValidWithHash(pdf.CacheDir, pdf.Hash) {
			cacheHits++
			if verbose {
				fmt.Printf("  ✓ %s [cached]\n", pdf.RelPath)
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

	if cacheMisses == 0 {
		fmt.Println("All PDFs are cached. Nothing to extract.")
		return 0
	}

	if dryRun {
		fmt.Println("\n[DRY RUN] Would extract pages from:")
		for _, pdf := range toProcess {
			fmt.Printf("  - %s -> %s/\n", pdf.RelPath, pdf.BookName)
		}
		return 0
	}

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

	// Build the pdf-tool image if needed
	fmt.Println("Ensuring pdf-tool Docker image...")
	if err := ensurePDFToolsImage(repoRoot); err != nil {
		log.Errorf("Error building pdf-tool image: %v", err)
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
		if verbose {
			fmt.Printf("  Processing %s...\n", pdf.RelPath)
		}

		// Clear old cache if exists (stale hash)
		if _, err := os.Stat(pdf.CacheDir); err == nil {
			if verbose {
				fmt.Printf("  Clearing stale cache for %s\n", pdf.BookName)
			}
			_ = os.RemoveAll(pdf.CacheDir)
		}

		// Create cache directory for this PDF
		if err := os.MkdirAll(pdf.CacheDir, 0o755); err != nil {
			log.Errorf("  ❌ Failed to create cache dir for %s: %v", pdf.RelPath, err)
			failed++
			continue
		}

		// Extract pages using pdftoppm
		if err := extractPages(dockerClient, pdf.Path, pdf.CacheDir, dpi); err != nil {
			log.Errorf("  ❌ Failed to extract %s: %v", pdf.RelPath, err)
			// Clean up partial cache
			_ = os.RemoveAll(pdf.CacheDir)
			failed++
			continue
		}

		// Write hash marker file
		hashMarker := filepath.Join(pdf.CacheDir, pdf.Hash+".cache")
		if err := os.WriteFile(hashMarker, []byte(pdf.Hash), 0o644); err != nil {
			log.Errorf("  ⚠️  Failed to write cache marker for %s: %v", pdf.RelPath, err)
		}

		// Count extracted pages
		pageCount := countPages(pdf.CacheDir)
		extracted++

		if verbose {
			fmt.Printf("  ✓ %s: %d pages -> %s/\n", pdf.RelPath, pageCount, pdf.BookName)
		}
	}

	// Summary
	fmt.Println()
	if failed > 0 {
		fmt.Printf("Completed with errors: %d extracted, %d failed\n", extracted, failed)
	} else {
		fmt.Printf("✓ Successfully extracted pages from %d PDF(s)\n", extracted)
	}
	fmt.Printf("Cache location: %s\n", cacheRoot)

	if failed > 0 {
		return 1
	}
	return 0
}

// scanForPDFs finds all PDF files in the build directory.
func scanForPDFs(buildDir, moduleFilter string) ([]PDFInfo, error) {
	var pdfs []PDFInfo

	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return pdfs, nil
	}

	err := filepath.WalkDir(buildDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".pdf") {
			return nil
		}

		relPath, relErr := filepath.Rel(buildDir, path)
		if relErr != nil {
			relPath = path // Fallback to absolute path
		}
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		module := ""
		if len(parts) > 0 {
			module = parts[0]
		}

		// Apply module filter if specified
		if moduleFilter != "" && module != moduleFilter {
			return nil
		}

		// Extract book name from filename (without .pdf extension)
		baseName := filepath.Base(path)
		bookName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		pdfs = append(pdfs, PDFInfo{
			Path:     path,
			RelPath:  relPath,
			Module:   module,
			BookName: bookName,
		})

		return nil
	})

	return pdfs, err
}

// hashFile returns the SHA256 hash of a file (first 12 characters).
func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "unknown"
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "unknown"
	}

	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}

// cacheValidWithHash checks if cache dir exists with correct hash marker.
func cacheValidWithHash(cacheDir, hash string) bool {
	// Check for hash marker file: <hash>.cache
	hashMarker := filepath.Join(cacheDir, hash+".cache")
	if _, err := os.Stat(hashMarker); err != nil {
		return false
	}

	// Also verify at least one PNG exists
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".png") {
			return true
		}
	}
	return false
}

// countPages counts the number of PNG files in a directory.
func countPages(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".png") {
			count++
		}
	}
	return count
}

// ensurePDFToolsImage builds the pdf-tool Docker image if needed.
func ensurePDFToolsImage(repoRoot string) error {
	// Check if image exists
	cmd := exec.Command("docker", "image", "inspect", pdfToolsImage)
	if err := cmd.Run(); err == nil {
		return nil // Image exists
	}

	// Build the image
	dockerfilePath := paths.ContainerDockerfilePath(repoRoot, "pdf-tool")
	buildCtx := paths.ContainersPath(repoRoot, "pdf-tool")

	fmt.Println("Building pdf-tool image...")
	cmd = exec.Command("docker", "build",
		"-t", pdfToolsImage,
		"-f", dockerfilePath,
		buildCtx)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// extractPages uses pdftoppm to extract PDF pages as PNG images.
func extractPages(_ docker.DockerClient, pdfPath, outputDir string, dpi int) error {
	pdfDir := filepath.Dir(pdfPath)
	pdfName := filepath.Base(pdfPath)

	ctx := context.Background()

	runConfig := &docker.RunConfig{
		Image: pdfToolsImage,
		Command: []string{
			"pdftoppm",
			"-png",
			"-r", fmt.Sprintf("%d", dpi),
			"/input/" + pdfName,
			"/output/page",
		},
		Mounts: []docker.MountConfig{
			{Source: pdfDir, Target: "/input", ReadOnly: true},
			{Source: outputDir, Target: "/output", ReadOnly: false},
		},
		WorkingDir:    "/workspace",
		Timeout:       10 * time.Minute,
		ContainerName: fmt.Sprintf("pdf-extract-%d", time.Now().UnixNano()),
	}

	result, err := docker.RunContainer(ctx, runConfig)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		logs := result.Stderr
		if logs == "" {
			logs = result.Stdout
		}
		return fmt.Errorf("container exited with code %d: %s", result.ExitCode, logs)
	}

	// pdftoppm outputs files like page-1.png, page-2.png
	// Rename to zero-padded format: page-01.png, page-02.png
	return renameToZeroPadded(outputDir)
}

// renameToZeroPadded renames page-1.png to page-01.png etc for proper sorting.
func renameToZeroPadded(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Find max page number for padding
	maxPage := 0
	pagePattern := regexp.MustCompile(`^page-(\d+)\.png$`)

	for _, entry := range entries {
		matches := pagePattern.FindStringSubmatch(entry.Name())
		if matches != nil {
			if num, err := strconv.Atoi(matches[1]); err == nil && num > maxPage {
				maxPage = num
			}
		}
	}

	// Determine padding width
	padWidth := 2
	if maxPage >= 100 {
		padWidth = 3
	}
	if maxPage >= 1000 {
		padWidth = 4
	}

	// Rename files
	for _, entry := range entries {
		matches := pagePattern.FindStringSubmatch(entry.Name())
		if matches != nil {
			num, convErr := strconv.Atoi(matches[1])
			if convErr != nil {
				continue
			}
			newName := fmt.Sprintf("page-%0*d.png", padWidth, num)
			if entry.Name() != newName {
				oldPath := filepath.Join(dir, entry.Name())
				newPath := filepath.Join(dir, newName)
				if err := os.Rename(oldPath, newPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
