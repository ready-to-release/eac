// mkdocs_book.go - Book-building functions for MkDocs modules
package builders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/paths"
)

// isForceRebuild checks if the user requested a full rebuild (--skip-cache or --force flag).
func isForceRebuild() bool {
	for _, arg := range os.Args {
		if arg == "--force" || arg == "-f" || arg == "--skip-cache" {
			return true
		}
	}
	return false
}

// exitCodeSkipped is returned by BuildSingleBook when a build is skipped due to unchanged content.
// This is distinct from success (0) to signal that no new PDF was generated.
const exitCodeSkipped = -1

// buildModuleBooks builds books for a module sequentially.
// Concurrency is managed solely by the orchestrator's weighted semaphore.
// Final PDFs are moved to the module output root with naming: {book-name}-{theme}.pdf.
func buildModuleBooks(module *modules.ModuleContract, moduleBooks []*config.Book, workspaceRoot, outputDir string, logWriter io.Writer) int {
	if len(moduleBooks) == 0 {
		return 0
	}

	if len(moduleBooks) > 1 {
		Logln(logWriter, "\n=== Building book: %s (%d books) ===", module.Moniker, len(moduleBooks))
		for _, book := range moduleBooks {
			Logln(logWriter, "   - %s (%s)", book.Name, book.GetOutput())
		}
	}

	// Pre-build: ensure drawio cache is up to date
	optimized, err := books.UpdateDrawioCache(workspaceRoot, logWriter)
	if err != nil {
		Logln(logWriter, "⚠️  Warning: drawio cache update failed: %v", err)
	} else if optimized > 0 {
		Logln(logWriter, "📊 Updated drawio cache: %d image(s) optimized", optimized)
	}

	// Build books sequentially - orchestrator manages parallelism
	allSkipped := true
	for _, book := range moduleBooks {
		// Determine output directory based on book type
		var bookOutputDir string
		if book.GetOutput() == "site" {
			bookOutputDir = outputDir
		} else {
			bookOutputDir = filepath.Join(outputDir, book.Name)
		}

		if err := os.MkdirAll(bookOutputDir, 0o755); err != nil {
			Logln(logWriter, "❌ Failed to create output directory for book '%s': %v", book.Name, err)
			return 1
		}

		exitCode := BuildSingleBook(module, book, workspaceRoot, outputDir, bookOutputDir, logWriter)

		if exitCode > 0 {
			return exitCode // Failure
		}
		if exitCode == 0 {
			allSkipped = false
			// For PDF books, move PDF to module output root
			movePDFToModuleRoot(book, bookOutputDir, outputDir, logWriter)
		}
	}

	if allSkipped && len(moduleBooks) > 0 {
		return exitCodeSkipped
	}

	if len(moduleBooks) > 1 {
		Logln(logWriter, "\n✅ All %d books built successfully", len(moduleBooks))
	}
	return 0
}

// movePDFToModuleRoot moves generated PDFs from build directory to module output root.
func movePDFToModuleRoot(book *config.Book, bookOutputDir, outputDir string, logWriter io.Writer) {
	bookOutput := book.GetOutput()
	if bookOutput == "site" {
		return
	}

	var themes []string
	switch bookOutput {
	case "pdf-dark":
		themes = []string{"dark"}
	case "pdf-light":
		themes = []string{"light"}
	case "pdf-all":
		themes = []string{"dark", "light"}
	}

	for _, theme := range themes {
		srcPdf := filepath.Join(bookOutputDir, "site", "pdf", fmt.Sprintf("%s-%s.pdf", book.Name, theme))
		dstPdf := filepath.Join(outputDir, fmt.Sprintf("%s-%s.pdf", book.Name, theme))
		if err := copyFile(srcPdf, dstPdf); err != nil {
			Logln(logWriter, "⚠️  Failed to copy PDF to module root: %v", err)
		} else {
			os.Remove(srcPdf)
			Logln(logWriter, "   📄 %s-%s.pdf → module output root", book.Name, theme)
		}
	}
}

// BuildSingleBook builds a single book based on its output configuration.
// moduleOutputDir is the module's base output directory (used for staging).
// bookOutputDir is where this book's final output goes.
// Exported for use by the evidence builder and other book-building commands.
func BuildSingleBook(module *modules.ModuleContract, book *config.Book, workspaceRoot, moduleOutputDir, bookOutputDir string, logWriter io.Writer) int {
	bookOutput := book.GetOutput()

	// Determine build mode from book config
	pdfMode := bookOutput == "pdf-dark" || bookOutput == "pdf-light" || bookOutput == "pdf-all"
	pdfTheme := ""
	switch bookOutput {
	case "pdf-dark":
		pdfTheme = "dark"
	case "pdf-light":
		pdfTheme = "light"
	case "pdf-all":
		pdfTheme = "all"
	}

	// Check Docker availability - fail fast if unavailable
	if !IsDockerAvailable() {
		errorMsg := "Docker is not available but required for book builds"
		if IsDockerInDocker() {
			errorMsg += "\nRunning in container: mount Docker socket with -v /var/run/docker.sock:/var/run/docker.sock"
		} else {
			errorMsg += "\nEnsure Docker is installed and the daemon is running"
		}
		Logln(logWriter, "❌ %s", errorMsg)
		return 1
	}

	Logln(logWriter, "\n=== Building book: %s (%s) ===", book.Name, bookOutput)

	if pdfMode {
		if pdfTheme == "all" {
			// Build both themes sequentially, sharing preprocessing
			stagingDir, ok := preprocessBook(book, workspaceRoot, module.Moniker, logWriter, true)
			if !ok {
				return 1 // Preprocessing failed
			}

			// First build dark theme (clean=true to start fresh)
			// Pass moduleOutputDir as pdfOutputDir (where PDFs end up after move)
			if exitCode := buildBookWithThemeAndStaging(module, book, workspaceRoot, moduleOutputDir, bookOutputDir, logWriter, "dark", true, stagingDir); exitCode != 0 {
				return exitCode
			}

			// Then build light theme (clean=false to preserve dark PDF)
			return buildBookWithThemeAndStaging(module, book, workspaceRoot, moduleOutputDir, bookOutputDir, logWriter, "light", false, stagingDir)
		}

		// Single theme build
		return buildBookWithTheme(module, book, workspaceRoot, module.Moniker, moduleOutputDir, bookOutputDir, logWriter, pdfTheme)
	}

	// HTML-only build
	return buildBookHTML(module, book, workspaceRoot, module.Moniker, moduleOutputDir, bookOutputDir, logWriter)
}

// preprocessBook runs book preprocessing and returns the staging directory.
// Uses persistent cache at .cache/eac/staging/{moniker}/{bookname}/ for incremental builds.
// This enables fast mtime-based incremental copies across builds.
//
// Incremental build support:
// - Staging directory is preserved across builds in .cache/eac/staging/
// - Use --force flag to clear staging and rebuild from scratch
func preprocessBook(book *config.Book, workspaceRoot, moniker string, logWriter io.Writer, pdfMode bool) (string, bool) {
	// Use persistent cache directory for incremental builds
	// Path: .cache/eac/staging/{moniker}/{book}
	// Example: .cache/eac/staging/docs:site/site/
	stagingDir := paths.BookStagingCachePath(workspaceRoot, moniker, book.Name)

	// Only clean staging directory on force rebuild (enables incremental builds)
	if isForceRebuild() {
		Logln(logWriter, "🔄 Force rebuild: clearing staging directory")
		if err := os.RemoveAll(stagingDir); err != nil && !os.IsNotExist(err) {
			Logln(logWriter, "❌ Failed to clean staging directory: %v", err)
			Logln(logWriter, "   This may indicate file locks or permission issues")
			Logln(logWriter, "   Try: rm -rf %s", stagingDir)
			return "", false
		}
	}

	// Ensure staging directory exists (may already exist from previous build)
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return "", false
	}

	Logln(logWriter, "📚 Preprocessing book: %s", book.Name)

	// Run preprocessing
	// TODO: Thread CacheConfig through for --skip-cache=asset support
	preprocessor := books.NewPreprocessor(book, workspaceRoot, stagingDir, logWriter, pdfMode, nil)
	if err := preprocessor.Preprocess(); err != nil {
		Logln(logWriter, "❌ Book preprocessing failed: %v", err)
		return "", false
	}

	return stagingDir, true
}

// buildBookWithTheme builds a book as PDF with a specific theme
// moniker is the module:component identifier for cache isolation.
// moduleOutputDir is used for final PDF location, bookOutputDir is for MkDocs output.
func buildBookWithTheme(module *modules.ModuleContract, book *config.Book, workspaceRoot, moniker, moduleOutputDir, bookOutputDir string, logWriter io.Writer, theme string) int {
	stagingDir, ok := preprocessBook(book, workspaceRoot, moniker, logWriter, true)
	if !ok {
		return 1 // Preprocessing failed
	}
	return buildBookWithThemeAndStaging(module, book, workspaceRoot, moduleOutputDir, bookOutputDir, logWriter, theme, true, stagingDir)
}

// buildBookWithThemeAndStaging builds a book as PDF using a pre-computed staging directory.
// Checks staging hash to skip PDF generation if content is unchanged (incremental build).
// pdfOutputDir is where the final PDF ends up (module output dir), bookOutputDir is MkDocs working dir.
func buildBookWithThemeAndStaging(module *modules.ModuleContract, book *config.Book, workspaceRoot, pdfOutputDir, bookOutputDir string, logWriter io.Writer, theme string, cleanBuild bool, stagingDir string) int {
	// Check for incremental build skip after preprocessing
	// The staging directory is now populated - compare its hash with previous build
	// Check PDF at pdfOutputDir (module output) where it ends up after being moved
	if !isForceRebuild() {
		canSkip, reason := ShouldSkipPDFGeneration(book.Name, theme, stagingDir, workspaceRoot, pdfOutputDir)
		if canSkip {
			Logln(logWriter, "⏩ Skipping PDF generation for %s-%s: %s", book.Name, theme, reason)
			return exitCodeSkipped
		}
		Logln(logWriter, "🔄 Generating PDF for %s-%s: %s", book.Name, theme, reason)
	}

	// Delegate to existing PDF build logic, passing book metadata for PDF generation
	exitCode := buildMkDocsWithThemeAndStaging(module, book.Name, book.Title, book.Description, workspaceRoot, bookOutputDir, logWriter, theme, cleanBuild, stagingDir)

	// Record build state on success for future incremental builds
	// Record PDF location at pdfOutputDir where it will be moved to
	if exitCode == 0 {
		if err := RecordPDFBuildComplete(book.Name, theme, stagingDir, workspaceRoot, pdfOutputDir); err != nil {
			Logln(logWriter, "⚠️  Failed to record build state: %v", err)
		}
	}

	return exitCode
}

// buildBookHTML builds a book as HTML site
// moniker is the module:component identifier for cache isolation.
// bookOutputDir is for final output.
func buildBookHTML(module *modules.ModuleContract, book *config.Book, workspaceRoot, moniker, moduleOutputDir, bookOutputDir string, logWriter io.Writer) int {
	stagingDir, ok := preprocessBook(book, workspaceRoot, moniker, logWriter, false)
	if !ok {
		return 1 // Preprocessing failed
	}

	// Check if we can skip the build (cache hit)
	if !isForceRebuild() {
		canSkip, reason := ShouldSkipSiteBuild(module.Moniker, stagingDir, workspaceRoot, bookOutputDir, false)
		if canSkip {
			Logln(logWriter, "⏭️  Skipping site build: %s", reason)
			Logln(logWriter, "✅ MkDocs site cached (unchanged)")
			return exitCodeSkipped
		}
		Logln(logWriter, "   Cache miss: %s", reason)
	}

	// Build using existing HTML logic with the staging directory
	exitCode := buildHTMLWithStaging(module, workspaceRoot, bookOutputDir, logWriter, stagingDir)

	// Record successful build for future cache hits
	if exitCode == 0 {
		if err := RecordSiteBuildComplete(module.Moniker, stagingDir, workspaceRoot, bookOutputDir); err != nil {
			Logln(logWriter, "   ⚠️  Failed to save build cache: %v", err)
		}
	}

	return exitCode
}

// buildHTMLWithStaging builds HTML site using a staging directory.
func buildHTMLWithStaging(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, stagingDir string) int {
	Logln(logWriter, "📚 Building MkDocs site using Docker")

	// Generate mkdocs.yml from site template
	// docs_dir must be relative to the config file location (outputDir), not workspaceRoot
	configPath := filepath.Join(outputDir, "mkdocs.yml")
	relStagingDir := ""
	if stagingDir != "" {
		var relErr error
		relStagingDir, relErr = filepath.Rel(outputDir, stagingDir)
		if relErr != nil {
			relStagingDir = stagingDir // Fallback to absolute
		}
		relStagingDir = filepath.ToSlash(relStagingDir)
		Logln(logWriter, "   Using staging: %s (relative to config)", relStagingDir)
	}
	configOpts := books.ConfigOptions{
		SiteName:     "Documentation",
		DocsDir:      relStagingDir,
		OutputFormat: "site",
	}
	if err := books.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		Logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return 1
	}
	Logln(logWriter, "   Config: %s (from template)", configPath)

	// Verify config file was written successfully
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		Logln(logWriter, "❌ Config file does not exist after generation: %s", configPath)
		return 1
	}

	// Copy mkdocs macros script for footer generation
	macrosSource := filepath.Join(workspaceRoot, "containers", "site-render-tool", "mkdocs_macros.py")
	macrosTarget := filepath.Join(outputDir, "main.py")
	if macrosData, err := os.ReadFile(macrosSource); err == nil {
		if err := os.WriteFile(macrosTarget, macrosData, 0o644); err != nil {
			Logln(logWriter, "   ⚠️  Failed to copy mkdocs macros script: %v", err)
		} else {
			Logln(logWriter, "   Macros: %s", macrosTarget)
		}
	}

	// Docker setup
	hostRepoRoot := workspaceRoot
	isDinD := IsDockerInDocker()
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
		}
	}

	// Get docker configuration from module first, then type, then defaults
	dockerCfg := getMkDocsDockerConfig(module, workspaceRoot, false)
	imageName := dockerCfg.ImageName

	// For docker build, use container paths (workspaceRoot) because the Docker CLI
	// runs in the container and tars up the context locally before sending to the daemon.
	// Host paths are only needed for volume mounts (docker run -v).
	dockerfilePath := dockerCfg.DockerfilePath
	contextPath := dockerCfg.ContextPath

	if err := ensureMkDocsImage(imageName, dockerfilePath, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	siteDir := filepath.Join(outputDir, "site")
	if err := os.MkdirAll(siteDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// MkDocs resolves --site-dir relative to the config file directory when using -f
	// Since both config and site are in outputDir, site-dir is just "site"
	relConfigPath, relErr := filepath.Rel(workspaceRoot, configPath)
	if relErr != nil {
		relConfigPath = configPath // Fallback to absolute
	}
	dockerVolume := FormatDockerVolumePath(hostRepoRoot)
	dockerSiteDir := "site"
	dockerConfigPath := strings.ReplaceAll(relConfigPath, "\\", "/")

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
	}

	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		buildArgs = append(buildArgs, "--user", fmt.Sprintf("%d:%d", uid, gid))
	}

	buildArgs = append(buildArgs,
		imageName,
		"mkdocs", "build",
		"-f", dockerConfigPath,
		"--site-dir", dockerSiteDir,
		"--clean",
		"--strict",
	)

	Logln(logWriter, "   Image: %s", imageName)
	Logln(logWriter, "   Output: %s", siteDir)

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", buildArgs...)
	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs build failed")
		return exitCode
	}

	Logln(logWriter, "✅ MkDocs site built successfully")
	return 0
}

// checkAndPreprocessBook checks for books.yml and runs preprocessing if a book matches
// Returns (stagingDir, bookUsed) - stagingDir is empty on error, bookUsed indicates if preprocessing was attempted
// pdfMode enables PDF-specific processing like link normalization.
func checkAndPreprocessBook(moniker, workspaceRoot, outputDir string, logWriter io.Writer, pdfMode bool) (string, bool) {
	// Load config with books
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		return "", false
	}

	if err := cfg.LoadBooks(false); err != nil {
		return "", false
	}

	// Check if there are books for this module
	moduleBooks := cfg.GetBooksByModule(moniker)
	if len(moduleBooks) == 0 {
		return "", false
	}

	// Use the first book for this module
	// TODO: Support multiple books per module with --book flag
	book := moduleBooks[0]
	if len(moduleBooks) > 1 {
		Logln(logWriter, "📚 Found %d books for module '%s', using '%s'", len(moduleBooks), moniker, book.Name)
	} else {
		Logln(logWriter, "📚 Book configuration found: '%s' (module: %s)", book.Name, moniker)
	}

	// Use persistent staging directory at .cache/eac/staging/{moniker}/{book.Name}
	// This survives across builds, enabling:
	// - Mermaid SVG caching (rendered diagrams persist)
	// - Faster incremental builds (unchanged files already staged)
	// NOTE: We don't clean this directory - files are overwritten incrementally.
	// For a clean rebuild, delete .cache/eac/staging/ manually.
	stagingDir := paths.BookStagingCachePath(workspaceRoot, moniker, book.Name)

	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return "", true
	}

	// Run preprocessing (overwrites existing files incrementally)
	// TODO: Thread CacheConfig through for --skip-cache=asset support
	preprocessor := books.NewPreprocessor(book, workspaceRoot, stagingDir, logWriter, pdfMode, nil)
	if err := preprocessor.Preprocess(); err != nil {
		Logln(logWriter, "❌ Book preprocessing failed: %v", err)
		return "", true
	}

	return stagingDir, true
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
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
