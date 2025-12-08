// mkdocs.go - Build functions for MkDocs module types
package builders

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/books"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func init() {
	// Register handler for "mkdocs" build system
	RegisterSystem("mkdocs", BuildMkDocsModule)
	RegisterSystemArtifacts("mkdocs", ListMkDocsArtifacts)
}

// ListMkDocsArtifacts returns the artifacts that would be produced by building this MkDocs module
func ListMkDocsArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	// MkDocs produces a site directory
	return []string{"site/"}
}

// BuildMkDocsModule builds MkDocs documentation sites using Docker.
// All MkDocs modules use this handler - behavior is the same for all.
// If the module has books defined in books.yml, each book is built based on its output config:
//   - site: HTML only
//   - pdf-dark: PDF with dark theme
//   - pdf-light: PDF with light theme
//   - pdf-all: Both dark and light PDFs
//
// Multiple books for the same module are built in parallel.
func BuildMkDocsModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	// Load config to check for book configuration
	cfg, _ := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if cfg != nil {
		cfg.LoadBooks(false)
		// Check if module has ANY books defined
		allBooks := cfg.GetBooksByModule(module.Moniker)
		if len(allBooks) > 0 {
			// Module has books - filter based on requested artifacts
			var moduleBooks []*config.Book
			if len(opts.RequestedArtifacts) > 0 {
				// Filter books to only those whose artifacts are requested
				for _, book := range allBooks {
					output := book.GetOutput()
					shouldInclude := false

					// Determine artifact IDs for this book based on output mode
					// Match the logic from expandBookArtifacts in artifact_helpers.go
					switch output {
					case "pdf-dark", "pdf-light":
						// Artifact ID is the book name
						for _, reqID := range opts.RequestedArtifacts {
							if reqID == book.Name {
								shouldInclude = true
								break
							}
						}
					case "pdf-all":
						// Artifact IDs are "{book-name}-dark" and "{book-name}-light"
						darkID := fmt.Sprintf("%s-dark", book.Name)
						lightID := fmt.Sprintf("%s-light", book.Name)
						for _, reqID := range opts.RequestedArtifacts {
							if reqID == darkID || reqID == lightID {
								shouldInclude = true
								break
							}
						}
					case "site":
						// HTML site - artifact ID is "site" directory
						for _, reqID := range opts.RequestedArtifacts {
							if reqID == "site" {
								shouldInclude = true
								break
							}
						}
					}

					if shouldInclude {
						moduleBooks = append(moduleBooks, book)
					}
				}
				if len(moduleBooks) > 0 {
					Logln(logWriter, "📚 Building %d book(s) based on requested artifacts", len(moduleBooks))
				}
			} else {
				// No specific artifacts requested - use default books
				moduleBooks = cfg.GetDefaultBooksByModule(module.Moniker)
			}
			if len(moduleBooks) > 0 {
				return buildModuleBooks(module, moduleBooks, workspaceRoot, outputDir, logWriter)
			}
			// Module has books but all are non-default - skip with success
			Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)
			Logln(logWriter, "📚 All %d book(s) have default: false - skipping (use --all to build)", len(allBooks))
			Logln(logWriter, "✅ Build skipped (no default books)")
			return 0
		}
	}

	// No books configured - standard HTML-only build
	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	// Check for book configuration and run preprocessing if found
	// pdfMode=false for standard HTML builds
	stagingDir, bookUsed := checkAndPreprocessBook(module.Moniker, workspaceRoot, outputDir, logWriter, false)
	if bookUsed && stagingDir == "" {
		// Preprocessing failed
		return 1
	}

	Logln(logWriter, "📚 Building MkDocs site using Docker")

	// Check Docker availability first - fail fast if unavailable
	if !IsDockerAvailable() {
		errorMsg := "Docker is not available but required for MkDocs builds"
		if IsDockerInDocker() {
			errorMsg += "\nRunning in container: mount Docker socket with -v /var/run/docker.sock:/var/run/docker.sock"
		} else {
			errorMsg += "\nEnsure Docker is installed and the daemon is running"
		}
		Logln(logWriter, "❌ %s", errorMsg)
		return 1
	}

	Logln(logWriter, "   WorkspaceRoot: %s", workspaceRoot)

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

	// For Docker-in-Docker: use host path for volume mount
	// When running inside a container, the Docker daemon runs on the host,
	// so volume mounts need host paths instead of container paths
	hostRepoRoot := workspaceRoot
	isDinD := IsDockerInDocker()
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
			Logln(logWriter, "   Docker-in-Docker: using host path %s", hostRoot)
		}
	}

	// Use lean mkdocs-site image for HTML builds (~150MB)
	imageName := "cli-mkdocs-site:latest"
	containerDir := "mkdocs-site"

	// In DinD mode, docker build runs on host so paths must be host paths
	var dockerfilePath, contextPath string
	if isDinD {
		// Host is Windows, construct Windows paths manually
		dockerfilePath = hostRepoRoot + "\\containers\\" + containerDir + "\\Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\" + containerDir
	} else {
		dockerfilePath = filepath.Join(hostRepoRoot, "containers", containerDir, "Dockerfile")
		contextPath = filepath.Join(hostRepoRoot, "containers", containerDir)
	}

	if err := ensureMkDocsImage(imageName, dockerfilePath, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	// Calculate the site output directory (local path within container)
	siteDir := filepath.Join(outputDir, "site")

	if err := os.MkdirAll(siteDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// MkDocs resolves --site-dir relative to the config file directory when using -f
	// Since both config and site are in outputDir, site-dir is just "site"
	dockerSiteDir := "site"

	volumeMountPath := hostRepoRoot
	dockerVolume := FormatDockerVolumePath(volumeMountPath)

	// Check for --accept-warnings flag
	acceptWarnings := false
	for _, arg := range os.Args {
		if arg == "--accept-warnings" {
			acceptWarnings = true
			break
		}
	}

	// Get relative config path for Docker (config is already in mounted volume)
	relConfigPath, _ := filepath.Rel(workspaceRoot, configPath)
	dockerConfigPath := strings.ReplaceAll(relConfigPath, "\\", "/")

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
	}

	// In Docker-in-Docker mode, run as current user to avoid permission issues
	// This ensures files created in the container have the same ownership as the host
	// We use the current process's UID/GID which matches the ext-eac container user
	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		userSpec := fmt.Sprintf("%d:%d", uid, gid)
		buildArgs = append(buildArgs, "--user", userSpec)
		Logln(logWriter, "   Docker-in-Docker: running as user %s", userSpec)
	}

	buildArgs = append(buildArgs,
		imageName,
		"mkdocs", "build",
		"-f", dockerConfigPath,
		"--site-dir", dockerSiteDir,
		"--clean",
	)

	// Determine strict mode: enabled by default, disabled with --accept-warnings
	if !acceptWarnings {
		buildArgs = append(buildArgs, "--strict")
	}

	Logln(logWriter, "   Image: %s", imageName)
	Logln(logWriter, "   Output: %s", siteDir)
	Logln(logWriter, "   Volume: %s:/docs", dockerVolume)
	Logln(logWriter, "   SiteDir: %s", dockerSiteDir)

	if acceptWarnings {
		Logln(logWriter, "   Mode: accepting warnings (--accept-warnings)")
	} else {
		Logln(logWriter, "   Mode: strict (--strict flag enabled)")
	}

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", buildArgs...)

	if acceptWarnings && exitCode != 0 {
		Logln(logWriter, "⚠️  Build completed with warnings (accepted)")
		exitCode = 0
	}

	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs build failed")
		return exitCode
	}

	Logln(logWriter, "✅ MkDocs site built successfully")
	Logln(logWriter, "   HTML Output: %s", siteDir)
	return 0
}

// ensureMkDocsImage ensures the cli-mkdocs Docker image exists.
// In CI workflows, the image is pre-built by docker/build-push-action with GHA caching.
// This function checks if the image exists first; only builds if missing.
// In DinD mode, dockerfilePath and contextPath are host (Windows) paths.
func ensureMkDocsImage(imageName, dockerfilePath, contextPath string, logWriter io.Writer) error {
	// Check if image already exists (e.g., pre-built by CI workflow)
	cmd := exec.Command("docker", "image", "inspect", imageName)
	if err := cmd.Run(); err == nil {
		Logln(logWriter, "   Docker image exists: %s (using pre-built)", imageName)
		return nil
	}

	// Image doesn't exist, build it
	Logln(logWriter, "   Building Docker image: %s", imageName)

	exitCode := RunCommandWithLog("", logWriter,
		"docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		contextPath)

	if exitCode != 0 {
		return fmt.Errorf("docker build failed with exit code %d", exitCode)
	}

	return nil
}

// buildModuleBooks builds all books for a module.
// HTML books can run in parallel, but PDF books run sequentially to avoid Playwright timeouts.
// Final PDFs are copied to the module output root with naming: {book-name}-{theme}.pdf
func buildModuleBooks(module *modules.ModuleContract, moduleBooks []*config.Book, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	if len(moduleBooks) == 1 {
		// Single book - build directly to module output directory
		return buildSingleBook(module, moduleBooks[0], workspaceRoot, outputDir, logWriter)
	}

	// Separate books by output type
	var htmlBooks, pdfBooks []*config.Book
	for _, book := range moduleBooks {
		output := book.GetOutput()
		if output == "site" {
			htmlBooks = append(htmlBooks, book)
		} else {
			pdfBooks = append(pdfBooks, book)
		}
	}

	Logln(logWriter, "\n=== Building %s: %s (%d books) ===", module.Type, module.Moniker, len(moduleBooks))
	for _, book := range moduleBooks {
		Logln(logWriter, "   - %s (%s)", book.Name, book.GetOutput())
	}

	// Build HTML books in parallel (they're lightweight)
	if len(htmlBooks) > 0 {
		var wg sync.WaitGroup
		results := make(chan int, len(htmlBooks))

		for _, book := range htmlBooks {
			wg.Add(1)
			go func(b *config.Book) {
				defer wg.Done()
				bookOutputDir := filepath.Join(outputDir, b.Name)
				if err := os.MkdirAll(bookOutputDir, 0755); err != nil {
					Logln(logWriter, "❌ Failed to create output directory for book '%s': %v", b.Name, err)
					results <- 1
					return
				}
				var bookLog bytes.Buffer
				exitCode := buildSingleBook(module, b, workspaceRoot, bookOutputDir, &bookLog)
				logWriter.Write(bookLog.Bytes())
				results <- exitCode
			}(book)
		}

		wg.Wait()
		close(results)

		for exitCode := range results {
			if exitCode != 0 {
				return exitCode
			}
		}
	}

	// Build PDF books in parallel (now safe with Docker resource limits)
	// With --cpus=8 --memory=8g limits, each container won't overwhelm the system
	if len(pdfBooks) > 0 {
		Logln(logWriter, "\n🚀 Building %d PDF books in PARALLEL...", len(pdfBooks))
		var wg sync.WaitGroup
		results := make(chan int, len(pdfBooks))

		for _, book := range pdfBooks {
			wg.Add(1)
			go func(b *config.Book) {
				defer wg.Done()
				bookOutputDir := filepath.Join(outputDir, b.Name)
				if err := os.MkdirAll(bookOutputDir, 0755); err != nil {
					Logln(logWriter, "❌ Failed to create output directory for book '%s': %v", b.Name, err)
					results <- 1
					return
				}

				// Use buffered writer to avoid interleaved output
				var bookLog bytes.Buffer
				exitCode := buildSingleBook(module, b, workspaceRoot, bookOutputDir, &bookLog)

				if exitCode != 0 {
					// Write log even on failure for debugging
					logWriter.Write(bookLog.Bytes())
					results <- exitCode
					return
				}

				// Copy PDF to module output root
				bookOutput := b.GetOutput()
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
					srcPdf := filepath.Join(bookOutputDir, fmt.Sprintf("%s-%s.pdf", b.Name, theme))
					dstPdf := filepath.Join(outputDir, fmt.Sprintf("%s-%s.pdf", b.Name, theme))
					if err := copyFile(srcPdf, dstPdf); err != nil {
						Logln(&bookLog, "⚠️  Failed to copy PDF to module root: %v", err)
					} else {
						Logln(&bookLog, "   📄 %s-%s.pdf → module output root", b.Name, theme)
					}
				}

				// Write complete log atomically (build + PDF copy)
				logWriter.Write(bookLog.Bytes())
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
	}

	Logln(logWriter, "\n✅ All %d books built successfully", len(moduleBooks))
	return 0
}

// buildSingleBook builds a single book based on its output configuration
func buildSingleBook(module *modules.ModuleContract, book *config.Book, workspaceRoot string, outputDir string, logWriter io.Writer) int {
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
			stagingDir, bookUsed := preprocessBook(book, workspaceRoot, outputDir, logWriter, true)
			if bookUsed && stagingDir == "" {
				return 1 // Preprocessing failed
			}

			// First build dark theme (clean=true to start fresh)
			if exitCode := buildBookWithThemeAndStaging(module, book, workspaceRoot, outputDir, logWriter, "dark", true, stagingDir); exitCode != 0 {
				return exitCode
			}

			// Then build light theme (clean=false to preserve dark PDF)
			return buildBookWithThemeAndStaging(module, book, workspaceRoot, outputDir, logWriter, "light", false, stagingDir)
		}

		// Single theme build
		return buildBookWithTheme(module, book, workspaceRoot, outputDir, logWriter, pdfTheme)
	}

	// HTML-only build
	return buildBookHTML(module, book, workspaceRoot, outputDir, logWriter)
}

// preprocessBook runs book preprocessing and returns the staging directory.
// Uses ephemeral staging inside the build output directory (.staging/).
// This is rebuilt on every build - no caching, no stale content.
func preprocessBook(book *config.Book, workspaceRoot string, outputDir string, logWriter io.Writer, pdfMode bool) (string, bool) {
	// Use ephemeral staging inside build output directory
	// Example: out/build/repository-report/.staging/
	stagingDir := filepath.Join(outputDir, ".staging")

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
func buildBookWithTheme(module *modules.ModuleContract, book *config.Book, workspaceRoot string, outputDir string, logWriter io.Writer, theme string) int {
	stagingDir, bookUsed := preprocessBook(book, workspaceRoot, outputDir, logWriter, true)
	if bookUsed && stagingDir == "" {
		return 1
	}
	return buildBookWithThemeAndStaging(module, book, workspaceRoot, outputDir, logWriter, theme, true, stagingDir)
}

// buildBookWithThemeAndStaging builds a book as PDF using a pre-computed staging directory
func buildBookWithThemeAndStaging(module *modules.ModuleContract, book *config.Book, workspaceRoot string, outputDir string, logWriter io.Writer, theme string, cleanBuild bool, stagingDir string) int {
	// Delegate to existing PDF build logic, passing book metadata for PDF generation
	return buildMkDocsWithThemeAndStaging(module, book.Name, book.Title, book.Description, workspaceRoot, outputDir, logWriter, theme, cleanBuild, stagingDir)
}

// buildBookHTML builds a book as HTML site
func buildBookHTML(module *modules.ModuleContract, book *config.Book, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	// Preprocess the book
	stagingDir, bookUsed := preprocessBook(book, workspaceRoot, outputDir, logWriter, false)
	if bookUsed && stagingDir == "" {
		return 1
	}

	// Build using existing HTML logic with the staging directory
	return buildHTMLWithStaging(module, workspaceRoot, outputDir, logWriter, stagingDir)
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

	// Docker setup
	hostRepoRoot := workspaceRoot
	isDinD := IsDockerInDocker()
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
		}
	}

	imageName := "cli-mkdocs-site:latest"
	containerDir := "mkdocs-site"
	var dockerfilePath, contextPath string
	if isDinD {
		dockerfilePath = hostRepoRoot + "\\containers\\" + containerDir + "\\Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\" + containerDir
	} else {
		dockerfilePath = filepath.Join(hostRepoRoot, "containers", containerDir, "Dockerfile")
		contextPath = filepath.Join(hostRepoRoot, "containers", containerDir)
	}

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

// buildMkDocsWithTheme builds a PDF with a specific theme (dark or light)
// cleanBuild controls whether to use --clean flag; set false when building multiple themes
// to preserve PDFs from previous theme builds
func buildMkDocsWithTheme(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, theme string, cleanBuild bool) int {
	// Check for book configuration and run preprocessing if found
	// pdfMode=true enables link normalization for PDF compatibility
	stagingDir, bookUsed := checkAndPreprocessBook(module.Moniker, workspaceRoot, outputDir, logWriter, true)
	if bookUsed && stagingDir == "" {
		// Preprocessing failed
		return 1
	}

	// Use module moniker as default book name when no book config is present
	// No title/description available for legacy builds (pass empty strings)
	return buildMkDocsWithThemeAndStaging(module, module.Moniker, "", "", workspaceRoot, outputDir, logWriter, theme, cleanBuild, stagingDir)
}

// buildMkDocsWithThemeAndStaging builds a PDF with a specific theme using a pre-computed staging directory
// This allows multiple theme builds to share preprocessing work
// bookName is used for the final PDF naming: {bookName}-{theme}.pdf
// bookTitle and bookDescription are used for the PDF cover page
func buildMkDocsWithThemeAndStaging(module *modules.ModuleContract, bookName string, bookTitle string, bookDescription string, workspaceRoot string, outputDir string, logWriter io.Writer, theme string, cleanBuild bool, stagingDir string) int {
	if theme == "" {
		theme = "dark"
	}

	// Default bookTitle to bookName if not provided
	if bookTitle == "" {
		bookTitle = bookName
	}

	Logln(logWriter, "\n=== Building %s: %s (PDF %s) ===", module.Type, module.Moniker, theme)

	// Generate mkdocs.yml from PDF template
	// docs_dir must be relative to the config file location (outputDir), not workspaceRoot
	configPath := filepath.Join(outputDir, "mkdocs.yml")
	relStagingDir := ""
	if stagingDir != "" {
		relStagingDir, _ = filepath.Rel(outputDir, stagingDir)
		relStagingDir = filepath.ToSlash(relStagingDir)
		Logln(logWriter, "   Using staging: %s (relative to config)", relStagingDir)
	}

	outputFormat := fmt.Sprintf("pdf-%s", theme)
	configOpts := books.ConfigOptions{
		SiteName:        bookName,
		SiteDescription: fmt.Sprintf("Generated PDF documentation for %s", bookName),
		BookTitle:       bookTitle,
		BookDescription: bookDescription,
		SiteURL:         "",
		DocsDir:         relStagingDir,
		Theme:           theme,
		OutputFormat:    outputFormat,
	}
	if err := books.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		Logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return 1
	}

	Logln(logWriter, "📄 Building MkDocs site with PDF export (%s theme)", theme)
	Logln(logWriter, "   Config: %s (from template)", configPath)
	Logln(logWriter, "   Theme: pdf-%s", theme)
	Logln(logWriter, "   WorkspaceRoot: %s", workspaceRoot)

	// For Docker-in-Docker: use host path for volume mount
	hostRepoRoot := workspaceRoot
	isDinD := IsDockerInDocker()
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
			Logln(logWriter, "   Docker-in-Docker: using host path %s", hostRoot)
		}
	}

	// PDF builds always use the mkdocs-pdf image
	imageName := "cli-mkdocs-pdf:latest"
	containerDir := "mkdocs-pdf"
	var dockerfilePath, contextPath string
	if isDinD {
		dockerfilePath = hostRepoRoot + "\\containers\\" + containerDir + "\\Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\" + containerDir
	} else {
		dockerfilePath = filepath.Join(hostRepoRoot, "containers", containerDir, "Dockerfile")
		contextPath = filepath.Join(hostRepoRoot, "containers", containerDir)
	}

	if err := ensureMkDocsImage(imageName, dockerfilePath, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	// Calculate the site output directory
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
		"--cpus", "8",              // Allocate 8 CPU cores for faster rendering
		"--memory", "8g",           // 8GB RAM for Chromium and mkdocs
		"--shm-size", "2gb",        // Shared memory for Chromium (prevents crashes)
		"-e", "ENABLE_PDF_EXPORT=true",
	}

	Logln(logWriter, "   PDF Export: enabled (mkdocs-exporter)")

	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		userSpec := fmt.Sprintf("%d:%d", uid, gid)
		buildArgs = append(buildArgs, "--user", userSpec)
		Logln(logWriter, "   Docker-in-Docker: running as user %s", userSpec)
	}

	buildArgs = append(buildArgs,
		imageName,
		"mkdocs", "build",
		"-f", dockerConfigPath,
		"--site-dir", dockerSiteDir,
	)

	// Only add --clean for first build; preserve previous theme PDFs when building multiple themes
	if cleanBuild {
		buildArgs = append(buildArgs, "--clean")
	}

	Logln(logWriter, "   Image: %s", imageName)
	Logln(logWriter, "   Output: %s", siteDir)
	Logln(logWriter, "   Volume: %s:/docs", dockerVolume)
	Logln(logWriter, "   SiteDir: %s", dockerSiteDir)
	if cleanBuild {
		Logln(logWriter, "   Mode: non-strict, clean build (PDF mode)")
	} else {
		Logln(logWriter, "   Mode: non-strict, incremental (preserving previous PDFs)")
	}

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", buildArgs...)

	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs PDF build failed (%s theme)", theme)
		return exitCode
	}

	// Merge individual PDFs into single document
	// mkdocs-exporter creates individual PDFs for each page, we merge them using pypdf
	dstPdfPath := filepath.Join(siteDir, "pdf", fmt.Sprintf("%s-%s.pdf", bookName, theme))

	Logln(logWriter, "📄 Merging individual PDFs...")

	if err := mergePDFs(siteDir, dstPdfPath, hostRepoRoot, workspaceRoot, stagingDir, imageName, bookTitle, bookDescription, logWriter, isDinD); err != nil {
		Logln(logWriter, "❌ PDF merge failed: %v", err)
		return 1
	}

	// Copy PDF to build output directory for easier access
	// Format: out/build/{module}/{bookName}-{theme}.pdf
	finalPdfPath := filepath.Join(outputDir, fmt.Sprintf("%s-%s.pdf", bookName, theme))
	if err := copyFile(dstPdfPath, finalPdfPath); err != nil {
		Logln(logWriter, "⚠️  Failed to copy PDF to output: %v", err)
	} else {
		Logln(logWriter, "✅ MkDocs PDF built successfully (%s theme)", theme)
		Logln(logWriter, "   PDF Output: %s", finalPdfPath)
	}
	return 0
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

// mergePDFs merges all individual page PDFs into a single document using pypdf
// Generates a cover page, table of contents, and hierarchical bookmarks
// bookTitle and bookDescription are used for the cover page content
func mergePDFs(siteDir, outputPath, hostRepoRoot, workspaceRoot, stagingDir, imageName string, bookTitle string, bookDescription string, logWriter io.Writer, isDinD bool) error {
	// Get relative paths for Docker
	relSiteDir, err := filepath.Rel(workspaceRoot, siteDir)
	if err != nil {
		return fmt.Errorf("calculating relative site dir: %w", err)
	}
	relOutputPath, err := filepath.Rel(workspaceRoot, outputPath)
	if err != nil {
		return fmt.Errorf("calculating relative output path: %w", err)
	}
	relStagingDir, err := filepath.Rel(workspaceRoot, stagingDir)
	if err != nil {
		return fmt.Errorf("calculating relative staging dir: %w", err)
	}

	// Convert to Docker paths
	dockerSiteDir := "/docs/" + strings.ReplaceAll(relSiteDir, "\\", "/")
	dockerOutputPath := "/docs/" + strings.ReplaceAll(relOutputPath, "\\", "/")
	dockerStagingDir := "/docs/" + strings.ReplaceAll(relStagingDir, "\\", "/")

	volumeMountPath := hostRepoRoot
	dockerVolume := FormatDockerVolumePath(volumeMountPath)

	args := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
	}

	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		args = append(args, "--user", fmt.Sprintf("%d:%d", uid, gid))
	}

	// Python script to generate cover page, TOC, and merge all PDFs
	// Parses .nav.yml files for proper titles and ordering
	// Escape book title and description for Python (handle quotes and newlines)
	escapedTitle := strings.ReplaceAll(strings.ReplaceAll(bookTitle, "\\", "\\\\"), "'", "\\'")
	escapedDescription := strings.ReplaceAll(strings.ReplaceAll(bookDescription, "\\", "\\\\"), "'", "\\'")

	pythonScript := fmt.Sprintf(`
import os
import io
import yaml
from datetime import datetime
from pathlib import Path
from pypdf import PdfWriter, PdfReader
from reportlab.lib.pagesizes import A4
from reportlab.lib.colors import HexColor
from reportlab.pdfgen import canvas
from reportlab.lib.units import cm

site_dir = '%s'
output_path = '%s'
docs_dir = '%s'  # Staging directory with .nav.yml files
book_title = '''%s'''  # Book-specific title for cover page
book_description = '''%s'''  # Book-specific description for cover page

# Dark theme colors (matching PDF dark theme)
BG_COLOR = HexColor('#0d1117')
TEXT_COLOR = HexColor('#e6edf3')
ACCENT_COLOR = HexColor('#58a6ff')
MUTED_COLOR = HexColor('#8b949e')

def draw_wrapped_text(canvas, text, x, y, max_width, font='Helvetica', size=12, line_height=0.5*cm, align='center'):
    """Draw text with word wrapping."""
    canvas.setFont(font, size)
    words = text.split()
    lines = []
    current_line = []

    for word in words:
        test_line = ' '.join(current_line + [word])
        if canvas.stringWidth(test_line, font, size) <= max_width:
            current_line.append(word)
        else:
            if current_line:
                lines.append(' '.join(current_line))
            current_line = [word]

    if current_line:
        lines.append(' '.join(current_line))

    # Draw lines centered or left-aligned
    for i, line in enumerate(lines):
        if align == 'center':
            canvas.drawCentredString(x, y - (i * line_height), line)
        else:
            canvas.drawString(x, y - (i * line_height), line)

    return len(lines)  # Return number of lines drawn

def create_cover_page():
    """Create a cover page PDF with title, subtitle, and metadata."""
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=A4)
    width, height = A4

    # Dark background
    c.setFillColor(BG_COLOR)
    c.rect(0, 0, width, height, fill=True, stroke=False)

    # Brand Title (top, smaller)
    c.setFillColor(TEXT_COLOR)
    c.setFont('Helvetica-Bold', 32)
    c.drawCentredString(width/2, height - 6*cm, 'Ready-to-Release')

    # Brand Subtitle
    c.setFont('Helvetica', 20)
    c.drawCentredString(width/2, height - 7.5*cm, 'Documentation')

    # Book-specific title (larger, prominent, accent color)
    # Auto-size or wrap if title is too long
    if book_title:
        c.setFillColor(ACCENT_COLOR)
        max_title_width = width - 4*cm  # Leave 2cm margin on each side

        # Try different font sizes to fit title on one line
        title_font_size = 28
        c.setFont('Helvetica-Bold', title_font_size)
        title_width = c.stringWidth(book_title, 'Helvetica-Bold', title_font_size)

        if title_width > max_title_width:
            # Title too long, try smaller font
            title_font_size = 24
            c.setFont('Helvetica-Bold', title_font_size)
            title_width = c.stringWidth(book_title, 'Helvetica-Bold', title_font_size)

            if title_width > max_title_width:
                # Still too long, try even smaller
                title_font_size = 20
                c.setFont('Helvetica-Bold', title_font_size)
                title_width = c.stringWidth(book_title, 'Helvetica-Bold', title_font_size)

                if title_width > max_title_width:
                    # Still too long, wrap it
                    draw_wrapped_text(c, book_title, width/2, height - 10*cm, max_title_width, 'Helvetica-Bold', 18, 0.6*cm, 'center')
                else:
                    c.drawCentredString(width/2, height - 10*cm, book_title)
            else:
                c.drawCentredString(width/2, height - 10*cm, book_title)
        else:
            c.drawCentredString(width/2, height - 10*cm, book_title)

        c.setFillColor(TEXT_COLOR)  # Reset color

    # Horizontal line
    c.setStrokeColor(ACCENT_COLOR)
    c.setLineWidth(2)
    c.line(4*cm, height - 11.5*cm, width - 4*cm, height - 11.5*cm)

    # Brand description
    c.setFillColor(MUTED_COLOR)
    c.setFont('Helvetica', 14)
    c.drawCentredString(width/2, height - 13*cm, 'Everything-as-Code Platform')
    c.drawCentredString(width/2, height - 14*cm, 'for Software Delivery Flows')

    # Book-specific description (wrapped if needed)
    current_y = height - 16*cm
    if book_description:
        c.setFont('Helvetica', 12)
        c.setFillColor(TEXT_COLOR)
        lines_drawn = draw_wrapped_text(c, book_description, width/2, current_y, width - 8*cm, 'Helvetica', 12, 0.5*cm, 'center')
        current_y -= lines_drawn * 0.5*cm

    # Date at bottom
    c.setFillColor(MUTED_COLOR)
    c.setFont('Helvetica', 11)
    date_str = datetime.now().strftime('Generated: %%B %%d, %%Y')
    c.drawCentredString(width/2, 3*cm, date_str)

    c.save()
    buffer.seek(0)
    return PdfReader(buffer)

def create_toc_pages(toc_entries, content_start_page):
    """Create table of contents pages with page numbers and clickable links."""
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=A4)
    width, height = A4

    # Track current y position and page
    y = height - 3*cm
    entries_per_page = 35
    entry_count = 0
    # Store link info: (toc_page, x, y, width, height, target_page)
    links = []
    current_toc_page = 0

    def start_new_page():
        nonlocal y, current_toc_page
        c.showPage()
        current_toc_page += 1
        # Dark background
        c.setFillColor(BG_COLOR)
        c.rect(0, 0, width, height, fill=True, stroke=False)
        y = height - 3*cm

    def draw_toc_header():
        nonlocal y
        # Dark background
        c.setFillColor(BG_COLOR)
        c.rect(0, 0, width, height, fill=True, stroke=False)

        # TOC title
        c.setFillColor(TEXT_COLOR)
        c.setFont('Helvetica-Bold', 24)
        c.drawString(2*cm, height - 2*cm, 'Table of Contents')

        # Line under title
        c.setStrokeColor(ACCENT_COLOR)
        c.setLineWidth(1)
        c.line(2*cm, height - 2.5*cm, width - 2*cm, height - 2.5*cm)
        y = height - 3.5*cm

    draw_toc_header()

    for title, page_num, depth, path in toc_entries:
        if entry_count > 0 and entry_count %% entries_per_page == 0:
            start_new_page()
            # Continue TOC header on new page
            c.setFillColor(TEXT_COLOR)
            c.setFont('Helvetica-Bold', 14)
            c.drawString(2*cm, height - 2*cm, 'Table of Contents (continued)')
            c.setStrokeColor(ACCENT_COLOR)
            c.setLineWidth(0.5)
            c.line(2*cm, height - 2.3*cm, width - 2*cm, height - 2.3*cm)
            y = height - 3*cm

        # Indentation based on depth (0.4cm per level)
        indent = depth * 0.4 * cm
        x = 2*cm + indent

        # Font size and style based on depth
        if depth == 1:
            c.setFont('Helvetica-Bold', 11)
            c.setFillColor(TEXT_COLOR)
        elif depth == 2:
            c.setFont('Helvetica', 10)
            c.setFillColor(TEXT_COLOR)
        else:
            c.setFont('Helvetica', 9)
            c.setFillColor(MUTED_COLOR)

        # Draw title
        c.drawString(x, y, title)

        # Draw page number (adjusted for cover + TOC pages)
        actual_page = page_num + content_start_page
        c.setFillColor(MUTED_COLOR)
        c.setFont('Helvetica', 9)
        c.drawRightString(width - 2*cm, y, str(actual_page))

        # Store link rectangle (full width of entry line)
        # Links will be added after merging since we need final page numbers
        links.append((current_toc_page, x, y - 0.1*cm, width - 2*cm - x, 0.4*cm, actual_page - 1))

        # Dotted line between title and page number
        c.setStrokeColor(HexColor('#30363d'))
        c.setLineWidth(0.3)
        c.setDash(1, 2)
        title_width = c.stringWidth(title, 'Helvetica-Bold' if depth == 1 else 'Helvetica', 11 if depth == 1 else (10 if depth == 2 else 9))
        line_start = x + title_width + 0.3*cm
        line_end = width - 2*cm - 0.5*cm
        if line_end > line_start:
            c.line(line_start, y + 0.1*cm, line_end, y + 0.1*cm)
        c.setDash()

        y -= 0.5*cm
        entry_count += 1

    c.save()
    buffer.seek(0)
    return PdfReader(buffer), links

# Parse .nav.yml files to build proper navigation structure
def parse_nav_yml(nav_dir, base_path=''):
    """Parse .nav.yml and return ordered list of (title, path, depth) entries."""
    nav_file = os.path.join(nav_dir, '.nav.yml')
    entries = []

    if not os.path.exists(nav_file):
        return entries

    with open(nav_file, 'r') as f:
        nav_data = yaml.safe_load(f)

    if not nav_data:
        return entries

    section_title = nav_data.get('title', '')
    nav_items = nav_data.get('nav', [])

    for item in nav_items:
        if isinstance(item, str):
            # Simple file reference: "index.md" or "subdirectory/"
            if item.endswith('.md'):
                # File reference
                file_path = os.path.join(base_path, item)
                # Convert .md to site path (file.md -> file/index.pdf or index.md -> index.pdf)
                if item == 'index.md':
                    site_path = base_path if base_path else ''
                else:
                    site_path = os.path.join(base_path, item[:-3])  # Remove .md

                # Get title from frontmatter or filename
                title = item[:-3].replace('-', ' ').replace('_', ' ').title()
                if item == 'index.md' and section_title:
                    title = section_title

                entries.append((title, site_path, file_path))
            else:
                # Subdirectory reference
                subdir = item.rstrip('/')
                subdir_path = os.path.join(nav_dir, subdir)
                sub_base = os.path.join(base_path, subdir) if base_path else subdir
                sub_entries = parse_nav_yml(subdir_path, sub_base)
                entries.extend(sub_entries)
        elif isinstance(item, dict):
            # Titled section: {"Title": [...]} or {"Title": "path.md"}
            for title, content in item.items():
                if isinstance(content, str):
                    # Single file with custom title
                    file_path = os.path.join(base_path, content)
                    if content.endswith('.md'):
                        site_path = os.path.join(base_path, content[:-3])
                    else:
                        site_path = os.path.join(base_path, content)
                    entries.append((title, site_path, file_path))
                elif isinstance(content, list):
                    # Inline section with items - add section header first
                    # Use special marker for section headers (no PDF, just TOC entry)
                    entries.append((title, '__section__', '__section__'))
                    for sub_item in content:
                        if isinstance(sub_item, str) and sub_item.endswith('.md'):
                            file_path = os.path.join(base_path, sub_item)
                            site_path = os.path.join(base_path, sub_item[:-3])
                            sub_title = sub_item[:-3].split('/')[-1].replace('-', ' ').replace('_', ' ').title()
                            entries.append((sub_title, site_path, file_path))

    return entries

def get_pdf_path(site_dir, site_path):
    """Convert site path to PDF file path."""
    if not site_path:
        return os.path.join(site_dir, 'index.pdf')
    # Try directory/index.pdf first
    pdf_path = os.path.join(site_dir, site_path, 'index.pdf')
    if os.path.exists(pdf_path):
        return pdf_path
    # Try file.pdf
    pdf_path = os.path.join(site_dir, site_path + '.pdf')
    if os.path.exists(pdf_path):
        return pdf_path
    return None

# Parse navigation from .nav.yml files
print('Parsing navigation structure from .nav.yml files...')
nav_entries = parse_nav_yml(docs_dir)

# Build TOC entries with PDF paths and page numbers
toc_entries = []
pdf_files = []
page_counts = []
current_section_depth = 1  # Track depth for section headers

for title, site_path, file_path in nav_entries:
    if site_path == '__section__':
        # Section header (no PDF) - will point to next page
        next_page = sum(page_counts)
        toc_entries.append((title, next_page, current_section_depth, '__section__'))
        current_section_depth = 2  # Items after section header are indented
        continue

    pdf_path = get_pdf_path(site_dir, site_path)
    if pdf_path and os.path.exists(pdf_path):
        # Calculate depth from site_path
        if not site_path:
            depth = 0
        else:
            depth = len(site_path.replace('\\\\', '/').split('/'))

        # Use section depth for items in inline sections
        if current_section_depth == 2 and depth <= 2:
            depth = current_section_depth

        reader = PdfReader(pdf_path)
        page_count = len(reader.pages)
        page_counts.append(page_count)

        current_page = sum(page_counts[:-1])
        toc_entries.append((title, current_page, depth, site_path))
        pdf_files.append(pdf_path)

        # Reset section depth when we hit a new top-level item
        if depth <= 1:
            current_section_depth = 1

print(f'Found {len(pdf_files)} PDF files from navigation')

# Create cover page
print('Creating cover page...')
cover_reader = create_cover_page()
cover_pages = len(cover_reader.pages)

# Estimate TOC pages (roughly 35 entries per page)
toc_page_estimate = max(1, (len(toc_entries) + 34) // 35)

# Content starts after cover + TOC
content_start_page = cover_pages + toc_page_estimate + 1

# Create TOC pages
print(f'Creating table of contents ({len(toc_entries)} entries)...')
toc_reader, toc_links = create_toc_pages(toc_entries, content_start_page)
toc_pages = len(toc_reader.pages)

# Recalculate if TOC pages changed
if toc_pages != toc_page_estimate:
    content_start_page = cover_pages + toc_pages + 1
    toc_reader, toc_links = create_toc_pages(toc_entries, content_start_page)

# Final merge
writer = PdfWriter()

# Add cover page
for page in cover_reader.pages:
    writer.add_page(page)

# Add TOC pages
for page in toc_reader.pages:
    writer.add_page(page)

# Add content pages
current_page = cover_pages + toc_pages
bookmarks = []

for i, pdf_path in enumerate(pdf_files):
    title, _, depth, site_path = toc_entries[i]
    bookmarks.append((title, current_page, depth, site_path))

    reader = PdfReader(pdf_path)
    for page in reader.pages:
        writer.add_page(page)
    current_page += len(reader.pages)

# Add clickable links to TOC pages
print(f'Adding {len(toc_links)} TOC links...')
from pypdf.generic import ArrayObject, DictionaryObject, FloatObject, NameObject, NumberObject

for toc_page_idx, x, y, link_width, link_height, target_page in toc_links:
    # Get the actual page in the merged PDF (cover + toc_page_idx)
    page_num = cover_pages + toc_page_idx
    if page_num < len(writer.pages) and target_page < len(writer.pages):
        page = writer.pages[page_num]

        # Create link annotation
        link = DictionaryObject()
        link[NameObject('/Type')] = NameObject('/Annot')
        link[NameObject('/Subtype')] = NameObject('/Link')
        link[NameObject('/Rect')] = ArrayObject([
            FloatObject(x),
            FloatObject(y),
            FloatObject(x + link_width),
            FloatObject(y + link_height)
        ])
        link[NameObject('/Border')] = ArrayObject([NumberObject(0), NumberObject(0), NumberObject(0)])
        # Destination: go to target page, fit width
        link[NameObject('/Dest')] = ArrayObject([
            writer.pages[target_page].indirect_reference,
            NameObject('/XYZ'),
            NumberObject(0),
            NumberObject(842),  # A4 height
            NumberObject(0)
        ])

        # Add annotation to page
        if '/Annots' not in page:
            page[NameObject('/Annots')] = ArrayObject()
        page['/Annots'].append(link)

# Add hierarchical bookmarks/outline
print(f'Adding {len(bookmarks)} bookmarks...')

# Add TOC bookmark first
toc_bookmark = writer.add_outline_item('Table of Contents', cover_pages)

parent_stack = {}
for title, page_num, depth, path in bookmarks:
    if depth == 1:
        parent = writer.add_outline_item(title, page_num)
        parent_stack[1] = parent
        parent_stack[2] = None
        parent_stack[3] = None
    elif depth == 2 and parent_stack.get(1):
        parent = writer.add_outline_item(title, page_num, parent=parent_stack[1])
        parent_stack[2] = parent
        parent_stack[3] = None
    elif depth == 3 and parent_stack.get(2):
        parent = writer.add_outline_item(title, page_num, parent=parent_stack[2])
        parent_stack[3] = parent
    elif depth >= 4 and parent_stack.get(3):
        writer.add_outline_item(title, page_num, parent=parent_stack[3])

# Ensure output directory exists
os.makedirs(os.path.dirname(output_path), exist_ok=True)

# Write merged PDF
with open(output_path, 'wb') as f:
    writer.write(f)

print(f'Merged PDF written to {output_path}')
print(f'Total pages: {len(writer.pages)}')
print(f'  - Cover: {cover_pages} page(s)')
print(f'  - TOC: {toc_pages} page(s)')
print(f'  - Content: {current_page - cover_pages - toc_pages} pages')
print(f'TOC entries: {len(bookmarks)}')
`, dockerSiteDir, dockerOutputPath, dockerStagingDir, escapedTitle, escapedDescription)

	args = append(args, imageName, "python3", "-c", pythonScript)

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", args...)
	if exitCode != 0 {
		return fmt.Errorf("PDF merge exited with code %d", exitCode)
	}

	return nil
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
	stagingDir := filepath.Join(workspaceRoot, "out", "staging", book.Name)

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
