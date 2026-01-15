// mkdocs_book.go - Book-building functions for MkDocs modules
package builders

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/books"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// buildModuleBooks builds all books for a module in parallel.
// Preprocessing runs in parallel; PDF exports are serialized via semaphore.
// Final PDFs are moved to the module output root with naming: {book-name}-{theme}.pdf
func buildModuleBooks(module *modules.ModuleContract, moduleBooks []*config.Book, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	if len(moduleBooks) > 1 {
		Logln(logWriter, "\n=== Building %s: %s (%d books) ===", module.Type, module.Moniker, len(moduleBooks))
		for _, book := range moduleBooks {
			Logln(logWriter, "   - %s (%s)", book.Name, book.GetOutput())
		}
		Logln(logWriter, "\n🚀 Building %d books in parallel (PDF exports serialized)...", len(moduleBooks))
	}

	// Pre-build: ensure drawio cache is up to date
	// This is fast if already cached (just hash comparison)
	optimized, err := books.UpdateDrawioCache(workspaceRoot, logWriter)
	if err != nil {
		Logln(logWriter, "⚠️  Warning: drawio cache update failed: %v", err)
		// Non-fatal - continue with build
	} else if optimized > 0 {
		Logln(logWriter, "📊 Updated drawio cache: %d image(s) optimized", optimized)
	}

	// Build books in parallel (PDF exports serialized via pdfExportSemaphore)
	var wg sync.WaitGroup
	results := make(chan int, len(moduleBooks))

	for _, book := range moduleBooks {
		wg.Add(1)
		go func(b *config.Book) {
			defer wg.Done()

			// Determine output directory based on book type:
			// - Site books: output directly to module dir (MkDocs creates site/ subdirectory)
			// - PDF books: use isolated subdirectory (PDF is moved to module root after build)
			var bookOutputDir string
			if b.GetOutput() == "site" {
				bookOutputDir = outputDir
			} else {
				bookOutputDir = filepath.Join(outputDir, b.Name)
			}

			if err := os.MkdirAll(bookOutputDir, 0755); err != nil {
				Logln(logWriter, "❌ Failed to create output directory for book '%s': %v", b.Name, err)
				results <- 1
				return
			}

			// Use buffered writer for parallel builds to avoid interleaved output
			var bookLog bytes.Buffer
			var bookLogWriter io.Writer = &bookLog
			if len(moduleBooks) == 1 {
				// Single book - write directly to main log
				bookLogWriter = logWriter
			}

			// Build the book (PDF exports will be serialized via semaphore)
			exitCode := BuildSingleBook(module, b, workspaceRoot, outputDir, bookOutputDir, bookLogWriter)

			if exitCode != 0 {
				if len(moduleBooks) > 1 {
					logWriter.Write(bookLog.Bytes())
				}
				results <- exitCode
				return
			}

			// For PDF books, move PDF to module output root
			bookOutput := b.GetOutput()
			if bookOutput != "site" {
				themes := []string{}
				switch bookOutput {
				case "pdf-dark":
					themes = []string{"dark"}
				case "pdf-light":
					themes = []string{"light"}
				case "pdf-all":
					themes = []string{"dark", "light"}
				}

				for _, theme := range themes {
					// Move PDF from site/pdf/ to module root
					// Use copy+delete instead of rename to handle cross-user permission issues
					// (Docker creates files as different user than the Go process)
					srcPdf := filepath.Join(bookOutputDir, "site", "pdf", fmt.Sprintf("%s-%s.pdf", b.Name, theme))
					dstPdf := filepath.Join(outputDir, fmt.Sprintf("%s-%s.pdf", b.Name, theme))
					if err := copyFile(srcPdf, dstPdf); err != nil {
						Logln(bookLogWriter, "⚠️  Failed to copy PDF to module root: %v", err)
					} else {
						// Remove source after successful copy
						os.Remove(srcPdf)
						Logln(bookLogWriter, "   📄 %s-%s.pdf → module output root", b.Name, theme)
					}
				}
			}

			// Write complete log atomically for parallel builds
			if len(moduleBooks) > 1 {
				logWriter.Write(bookLog.Bytes())
			}
			results <- 0
		}(book)
	}

	wg.Wait()
	close(results)

	for exitCode := range results {
		if exitCode != 0 {
			return exitCode
		}
	}

	if len(moduleBooks) > 1 {
		Logln(logWriter, "\n✅ All %d books built successfully", len(moduleBooks))
	}
	return 0
}

// BuildSingleBook builds a single book based on its output configuration.
// moduleOutputDir is the module's base output directory (used for staging).
// bookOutputDir is where this book's final output goes.
// Exported for use by the evidence builder and other book-building commands.
func BuildSingleBook(module *modules.ModuleContract, book *config.Book, workspaceRoot string, moduleOutputDir string, bookOutputDir string, logWriter io.Writer) int {
	bookOutput := book.GetOutput()

	// Check Docker availability first - fail fast if unavailable
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

	Logln(logWriter, "\n=== Building book: %s (%s) ===", book.Name, bookOutput)

	if pdfMode {
		if pdfTheme == "all" {
			// Build both themes sequentially, sharing preprocessing
			// Staging is always at module output root for isolation
			stagingDir, ok := preprocessBook(book, workspaceRoot, moduleOutputDir, logWriter, true)
			if !ok {
				return 1 // Preprocessing failed
			}

			// First build dark theme (clean=true to start fresh)
			if exitCode := buildBookWithThemeAndStaging(module, book, workspaceRoot, bookOutputDir, logWriter, "dark", true, stagingDir); exitCode != 0 {
				return exitCode
			}

			// Then build light theme (clean=false to preserve dark PDF)
			return buildBookWithThemeAndStaging(module, book, workspaceRoot, bookOutputDir, logWriter, "light", false, stagingDir)
		}

		// Single theme build
		return buildBookWithTheme(module, book, workspaceRoot, moduleOutputDir, bookOutputDir, logWriter, pdfTheme)
	}

	// HTML-only build
	return buildBookHTML(module, book, workspaceRoot, moduleOutputDir, bookOutputDir, logWriter)
}

// preprocessBook runs book preprocessing and returns the staging directory.
// Uses book-specific staging directory at the module output root: staging/<bookname>/
// This enables parallel builds with isolated preprocessing for each book.
// The staging directory is placed at moduleOutputDir/staging/<bookname> to ensure
// all books share the same base level, regardless of their individual output paths.
func preprocessBook(book *config.Book, workspaceRoot string, moduleOutputDir string, logWriter io.Writer, pdfMode bool) (string, bool) {
	// Use book-specific staging directory for isolation during parallel builds
	// Always relative to module output root for consistent isolation
	// Example: out/build/docs/staging/site/ or out/build/docs/staging/pdf/
	stagingDir := filepath.Join(moduleOutputDir, "staging", book.Name)

	// Clean staging directory on every build (ensures fresh content)
	if err := os.RemoveAll(stagingDir); err != nil && !os.IsNotExist(err) {
		Logln(logWriter, "❌ Failed to clean staging directory: %v", err)
		Logln(logWriter, "   This may indicate file locks or permission issues")
		Logln(logWriter, "   Try: rm -rf %s", stagingDir)
		return "", false
	}

	// Create fresh staging directory
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return "", false
	}

	Logln(logWriter, "📚 Preprocessing book: %s", book.Name)

	// Run preprocessing
	preprocessor := books.NewPreprocessor(book, workspaceRoot, stagingDir, logWriter, pdfMode)
	if err := preprocessor.Preprocess(); err != nil {
		Logln(logWriter, "❌ Book preprocessing failed: %v", err)
		return "", false
	}

	return stagingDir, true
}

// buildBookWithTheme builds a book as PDF with a specific theme
// moduleOutputDir is used for staging, bookOutputDir is for final output
func buildBookWithTheme(module *modules.ModuleContract, book *config.Book, workspaceRoot string, moduleOutputDir string, bookOutputDir string, logWriter io.Writer, theme string) int {
	stagingDir, ok := preprocessBook(book, workspaceRoot, moduleOutputDir, logWriter, true)
	if !ok {
		return 1 // Preprocessing failed
	}
	return buildBookWithThemeAndStaging(module, book, workspaceRoot, bookOutputDir, logWriter, theme, true, stagingDir)
}

// buildBookWithThemeAndStaging builds a book as PDF using a pre-computed staging directory
func buildBookWithThemeAndStaging(module *modules.ModuleContract, book *config.Book, workspaceRoot string, bookOutputDir string, logWriter io.Writer, theme string, cleanBuild bool, stagingDir string) int {
	// Delegate to existing PDF build logic, passing book metadata for PDF generation
	return buildMkDocsWithThemeAndStaging(module, book.Name, book.Title, book.Description, workspaceRoot, bookOutputDir, logWriter, theme, cleanBuild, stagingDir)
}

// buildBookHTML builds a book as HTML site
// moduleOutputDir is used for staging, bookOutputDir is for final output
func buildBookHTML(module *modules.ModuleContract, book *config.Book, workspaceRoot string, moduleOutputDir string, bookOutputDir string, logWriter io.Writer) int {
	// Preprocess the book - staging is always at module output root for isolation
	stagingDir, ok := preprocessBook(book, workspaceRoot, moduleOutputDir, logWriter, false)
	if !ok {
		return 1 // Preprocessing failed
	}

	// Build using existing HTML logic with the staging directory
	return buildHTMLWithStaging(module, workspaceRoot, bookOutputDir, logWriter, stagingDir)
}

// buildHTMLWithStaging builds HTML site using a staging directory
func buildHTMLWithStaging(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, stagingDir string) int {
	Logln(logWriter, "📚 Building MkDocs site using Docker")

	// Generate mkdocs.yml from site template
	// docs_dir must be relative to the config file location (outputDir), not workspaceRoot
	configPath := filepath.Join(outputDir, "mkdocs.yml")
	relStagingDir := ""
	if stagingDir != "" {
		relStagingDir, _ = filepath.Rel(outputDir, stagingDir)
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

	// Copy mkdocs macros script for footer generation
	macrosSource := filepath.Join(workspaceRoot, "containers", "mkdocs-site", "mkdocs_macros.py")
	macrosTarget := filepath.Join(outputDir, "main.py")
	if macrosData, err := os.ReadFile(macrosSource); err == nil {
		if err := os.WriteFile(macrosTarget, macrosData, 0644); err != nil {
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
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// MkDocs resolves --site-dir relative to the config file directory when using -f
	// Since both config and site are in outputDir, site-dir is just "site"
	relConfigPath, _ := filepath.Rel(workspaceRoot, configPath)
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
// pdfMode enables PDF-specific processing like link normalization
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

	// Use persistent staging directory at out/staging/{book.Name}
	// This survives across builds, enabling:
	// - Mermaid SVG caching (rendered diagrams persist)
	// - Faster incremental builds (unchanged files already staged)
	// NOTE: We don't clean this directory - files are overwritten incrementally.
	// For a clean rebuild, delete out/staging/ manually.
	stagingDir := paths.StagingPath(workspaceRoot, book.Name)

	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return "", true
	}

	// Run preprocessing (overwrites existing files incrementally)
	preprocessor := books.NewPreprocessor(book, workspaceRoot, stagingDir, logWriter, pdfMode)
	if err := preprocessor.Preprocess(); err != nil {
		Logln(logWriter, "❌ Book preprocessing failed: %v", err)
		return "", true
	}

	return stagingDir, true
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
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
