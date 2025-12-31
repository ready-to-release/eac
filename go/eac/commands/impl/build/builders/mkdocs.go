// mkdocs.go - Build functions for MkDocs module types
package builders

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/books"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/environments"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// pdfExportSemaphore limits concurrent PDF exports to 1.
// mkdocs-exporter uses Playwright which is resource-intensive.
// Preprocessing and other operations can run in parallel.
var pdfExportSemaphore = make(chan struct{}, 1)

// getPDFConcurrency returns the Playwright concurrency for PDF exports.
// Uses repository parallelism config: CI=4, devbox=8.
// Since only one mkdocs-exporter runs at a time (via semaphore),
// we can use all available cores for internal page rendering.
func getPDFConcurrency(workspaceRoot string) int {
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		// Fallback to sensible defaults
		if environments.IsCI() {
			return 4
		}
		return 8
	}

	if environments.IsCI() {
		if cfg.Repository.Repository.Parallelism.CI > 0 {
			return cfg.Repository.Repository.Parallelism.CI
		}
		return 4
	}

	if cfg.Repository.Repository.Parallelism.Devbox > 0 {
		return cfg.Repository.Repository.Parallelism.Devbox
	}
	return 8
}

func init() {
	RegisterHandler(&MkDocsHandler{})
}

// MkDocsHandler builds MkDocs documentation sites using Docker.
type MkDocsHandler struct{}

func (h *MkDocsHandler) Name() string { return "mkdocs" }

func (h *MkDocsHandler) Capabilities() []string { return []string{"documentation", "container"} }

func (h *MkDocsHandler) Requirements() []string { return []string{"docker"} }

func (h *MkDocsHandler) ValidateModule(module *modules.ModuleContract, workspaceRoot string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available")
	}
	return nil
}

func (h *MkDocsHandler) ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	return []string{"site/"}
}

func (h *MkDocsHandler) Build(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	return buildMkDocsModule(module, workspaceRoot, outputDir, logWriter, opts)
}

// mkdocsDockerConfig holds resolved docker configuration for mkdocs builds
type mkdocsDockerConfig struct {
	ImageName      string
	ContainerDir   string
	DockerfilePath string
	ContextPath    string
}

// getMkDocsDockerConfig resolves docker configuration from module first, then type, then defaults.
// For PDF builds, always use mkdocs-pdf container (module config only applies to site builds).
// For site builds, module docker_build config is respected.
func getMkDocsDockerConfig(module *modules.ModuleContract, workspaceRoot string, isPDF bool) mkdocsDockerConfig {
	// Defaults based on output type
	var defaultImage, defaultContainer string
	if isPDF {
		defaultImage = "cli-mkdocs-pdf:latest"
		defaultContainer = "mkdocs-pdf"
	} else {
		defaultImage = "cli-mkdocs-site:latest"
		defaultContainer = "mkdocs-site"
	}

	cfg := mkdocsDockerConfig{
		ImageName:    defaultImage,
		ContainerDir: defaultContainer,
	}

	// For PDF builds, always use the PDF container - don't let module config override
	// PDF requires specific plugins (mkdocs-exporter, playwright) only in mkdocs-pdf container
	if isPDF {
		cfg.ContextPath = filepath.Join(workspaceRoot, "containers", cfg.ContainerDir)
		cfg.DockerfilePath = filepath.Join(cfg.ContextPath, "Dockerfile")
		return cfg
	}

	// For site builds: module's own docker_build config takes priority
	if module.DockerBuild != nil && len(module.DockerBuild) > 0 {
		if tags, ok := module.DockerBuild["tags"].([]interface{}); ok && len(tags) > 0 {
			if tag, ok := tags[0].(string); ok {
				cfg.ImageName = tag
			}
		} else if container, ok := module.DockerBuild["container"].(string); ok {
			cfg.ImageName = container + ":latest"
		}

		if context, ok := module.DockerBuild["context"].(string); ok {
			cfg.ContainerDir = filepath.Base(context)
			cfg.ContextPath = filepath.Join(workspaceRoot, context)
		}

		if dockerfile, ok := module.DockerBuild["dockerfile"].(string); ok {
			cfg.DockerfilePath = filepath.Join(workspaceRoot, dockerfile)
		}
	}

	// Second priority: module type config (for backwards compatibility)
	if cfg.ContextPath == "" {
		if globalCfg := config.Global(); globalCfg != nil && globalCfg.ModuleTypes != nil {
			if img := globalCfg.ModuleTypes.GetDockerImageName(module.Type); img != "" {
				cfg.ImageName = img
			}
			if dir := globalCfg.ModuleTypes.GetDockerContainerDir(module.Type); dir != "" {
				cfg.ContainerDir = filepath.Base(dir)
			}
		}
	}

	// Build paths if not set from module config
	if cfg.ContextPath == "" {
		cfg.ContextPath = filepath.Join(workspaceRoot, "containers", cfg.ContainerDir)
	}
	if cfg.DockerfilePath == "" {
		cfg.DockerfilePath = filepath.Join(cfg.ContextPath, "Dockerfile")
	}

	return cfg
}

// buildMkDocsModule builds MkDocs documentation sites using Docker.
// All MkDocs modules use this handler - behavior is the same for all.
// If the module has books defined in books.yml, each book is built based on its output config:
//   - site: HTML only
//   - pdf-dark: PDF with dark theme
//   - pdf-light: PDF with light theme
//   - pdf-all: Both dark and light PDFs
//
// Multiple books for the same module are built in parallel.
func buildMkDocsModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	// Load config to check for book configuration
	cfg, _ := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if cfg != nil {
		cfg.LoadBooks(false)
		cfg.LoadRepository(false) // Need modules to look up books by module's books list
		// Check if module has ANY books defined
		allBooks := cfg.GetBooksByModule(module.Moniker)
		if len(allBooks) > 0 {
			// Module has books - filter based on requested artifacts
			var moduleBooks []*config.Book

			// Check if --all flag was used (RequestedArtifacts contains "*")
			buildAllArtifacts := false
			for _, reqID := range opts.RequestedArtifacts {
				if reqID == "*" {
					buildAllArtifacts = true
					break
				}
			}

			if buildAllArtifacts {
				// Build all books when --all is specified
				moduleBooks = allBooks
				Logln(logWriter, "📚 Building all %d book(s) (--all specified)", len(moduleBooks))
			} else if len(opts.RequestedArtifacts) > 0 {
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

// ensureMkDocsImage ensures the cli-mkdocs Docker image is built.
// Always runs docker build - Docker's layer cache handles efficiency.
// In DinD mode, dockerfilePath and contextPath are host (Windows) paths.
func ensureMkDocsImage(imageName, dockerfilePath, contextPath string, logWriter io.Writer) error {
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
	pdfConcurrency := getPDFConcurrency(workspaceRoot)
	configOpts := books.ConfigOptions{
		SiteName:        bookName,
		SiteDescription: fmt.Sprintf("Generated PDF documentation for %s", bookName),
		BookTitle:       bookTitle,
		BookDescription: bookDescription,
		SiteURL:         "",
		DocsDir:         relStagingDir,
		Theme:           theme,
		OutputFormat:    outputFormat,
		PDFConcurrency:  pdfConcurrency,
	}
	if err := books.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		Logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return 1
	}

	Logln(logWriter, "📄 Building MkDocs site with PDF export (%s theme)", theme)
	Logln(logWriter, "   Config: %s (from template)", configPath)
	Logln(logWriter, "   Theme: pdf-%s", theme)
	Logln(logWriter, "   Concurrency: %d (environment: %s)", pdfConcurrency, environments.DetectRuntime())
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

	// Get docker configuration from module first, then type, then defaults (PDF mode)
	dockerCfg := getMkDocsDockerConfig(module, workspaceRoot, true)
	imageName := dockerCfg.ImageName

	// Use container paths for docker build context
	// The Docker CLI (running in container) needs to access these paths to tar up the context
	// Host paths are only needed for volume mounts, not build context
	dockerfilePath := dockerCfg.DockerfilePath
	contextPath := dockerCfg.ContextPath

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

	// Use all available CPUs for PDF rendering (adapts to CI runners with fewer cores)
	cpuLimit := fmt.Sprintf("%d", runtime.NumCPU())

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
		"--cpus", cpuLimit, // Allocate available CPU cores for faster rendering
		"--memory", "8g", // 8GB RAM for Chromium and mkdocs
		"--shm-size", "2gb", // Shared memory for Chromium (prevents crashes)
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
		"--verbose", // Show detailed output for debugging PDF rendering errors
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

	// Acquire semaphore - only one PDF export at a time (Playwright is resource-intensive)
	Logln(logWriter, "⏳ Waiting for PDF export slot...")
	pdfExportSemaphore <- struct{}{}
	Logln(logWriter, "🔓 Acquired PDF export slot")

	// Retry logic for PDF builds - Playwright can have transient timeouts
	maxRetries := 2
	var exitCode int
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			Logln(logWriter, "🔄 Retrying PDF build (attempt %d/%d)...", attempt, maxRetries)
		}

		exitCode = RunCommandWithLog(workspaceRoot, logWriter, "docker", buildArgs...)

		if exitCode == 0 {
			break // Success
		}

		if attempt < maxRetries {
			Logln(logWriter, "⚠️  PDF build failed, will retry...")
		}
	}

	// Release semaphore - PDF export done, merge can proceed in parallel
	<-pdfExportSemaphore
	Logln(logWriter, "🔓 Released PDF export slot")

	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs PDF build failed (%s theme) after %d attempts", theme, maxRetries)
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

	Logln(logWriter, "✅ MkDocs PDF built successfully (%s theme)", theme)
	Logln(logWriter, "   PDF Output: %s", dstPdfPath)
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

import re

def get_title_from_markdown(md_path):
    """Extract title from markdown file (frontmatter or first H1)."""
    if not os.path.exists(md_path):
        return None

    try:
        with open(md_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except:
        return None

    # Try frontmatter title
    if content.startswith('---'):
        end_idx = content.find('---', 3)
        if end_idx > 0:
            frontmatter = content[3:end_idx]
            match = re.search(r'^title:\s*["\']?([^"\'\\n]+)["\']?', frontmatter, re.MULTILINE)
            if match:
                return match.group(1).strip()

    # Try first H1
    match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
    if match:
        return match.group(1).strip()

    return None

# Dark theme colors (matching PDF dark theme)
BG_COLOR = HexColor('#0d1117')
TEXT_COLOR = HexColor('#e6edf3')
ACCENT_COLOR = HexColor('#58a6ff')
MUTED_COLOR = HexColor('#8b949e')
FOOTER_COLOR = HexColor('#6e7681')

def create_page_footer_overlay(page_num, total_pages, book_title, section_title=''):
    """Create a PDF page with footer overlay (page number, book title, section)."""
    packet = io.BytesIO()
    c = canvas.Canvas(packet, pagesize=A4)
    width, height = A4

    # Footer Y position (2cm from bottom = ~56 points)
    footer_y = 1.5 * cm

    # Page number - center
    c.setFont('Helvetica', 9)
    c.setFillColor(MUTED_COLOR)
    page_text = str(page_num)
    c.drawCentredString(width / 2, footer_y, page_text)

    # Book title - left (truncate if too long)
    c.setFont('Helvetica', 8)
    c.setFillColor(FOOTER_COLOR)
    display_title = book_title[:40] + '...' if len(book_title) > 40 else book_title
    c.drawString(2 * cm, footer_y, display_title)

    # Section title - right (truncate if too long)
    if section_title:
        display_section = section_title[:35] + '...' if len(section_title) > 35 else section_title
        c.drawRightString(width - 2 * cm, footer_y, display_section)

    c.save()
    packet.seek(0)
    return PdfReader(packet)

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
        else:  # depth 3, 4, etc. - all same format
            c.setFont('Helvetica', 9)
            c.setFillColor(TEXT_COLOR)

        # Draw title
        c.drawString(x, y, title)

        # Draw page number (content pages start at 1)
        display_page = page_num + 1
        c.setFillColor(MUTED_COLOR)
        c.setFont('Helvetica', 9)
        c.drawRightString(width - 2*cm, y, str(display_page))

        # Store link rectangle - target is absolute page in merged PDF
        # page_num is 0-indexed content offset, add content_start_page for absolute position
        absolute_target = page_num + content_start_page
        links.append((current_toc_page, x, y - 0.1*cm, width - 2*cm - x, 0.4*cm, absolute_target))

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

                # Get title from markdown file, fallback to filename
                md_file_path = os.path.join(nav_dir, item)
                title = get_title_from_markdown(md_file_path)
                if not title:
                    # Fallback to filename-based title
                    title = item[:-3].replace('-', ' ').replace('_', ' ').title()

                entries.append((title, site_path, file_path))
            else:
                # Subdirectory reference
                subdir = item.rstrip('/')
                subdir_path = os.path.join(nav_dir, subdir)
                sub_base = os.path.join(base_path, subdir) if base_path else subdir
                sub_entries = parse_nav_yml(subdir_path, sub_base)
                entries.extend(sub_entries)
        elif isinstance(item, dict):
            # Titled section: {"Title": [...]} or {"Title": "path.md"} or {"Title": "subdirectory"}
            for title, content in item.items():
                if isinstance(content, str):
                    if content.endswith('.md'):
                        # Single file with custom title
                        file_path = os.path.join(base_path, content)
                        site_path = os.path.join(base_path, content[:-3])
                        entries.append((title, site_path, file_path))
                    else:
                        # Subdirectory with custom title - add header and recurse
                        subdir = content.rstrip('/')
                        subdir_path = os.path.join(nav_dir, subdir)
                        sub_base = os.path.join(base_path, subdir) if base_path else subdir

                        # Check if subdirectory exists
                        if os.path.isdir(subdir_path):
                            # Add the directory's index.md as an entry with the custom title
                            index_site_path = sub_base
                            index_file_path = os.path.join(sub_base, 'index.md')
                            entries.append((title, index_site_path, index_file_path))

                            # Recursively parse subdirectory
                            sub_entries = parse_nav_yml(subdir_path, sub_base)
                            # Skip the subdirectory's index.md since we added it with custom title
                            for sub_entry in sub_entries:
                                sub_title, sub_site_path, sub_file_path = sub_entry
                                if sub_file_path != os.path.join(sub_base, 'index.md'):
                                    entries.append(sub_entry)
                        else:
                            # Fallback: treat as file path
                            file_path = os.path.join(base_path, content)
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
                            # Get title from markdown file, fallback to filename
                            md_file_path = os.path.join(nav_dir, sub_item)
                            sub_title = get_title_from_markdown(md_file_path)
                            if not sub_title:
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

# Content starts after cover + TOC (page numbers start at 1, not 0)
content_start_page = cover_pages + toc_page_estimate

# Create TOC pages
print(f'Creating table of contents ({len(toc_entries)} entries)...')
toc_reader, toc_links = create_toc_pages(toc_entries, content_start_page)
toc_pages = len(toc_reader.pages)

# Recalculate if TOC pages changed
if toc_pages != toc_page_estimate:
    content_start_page = cover_pages + toc_pages
    toc_reader, toc_links = create_toc_pages(toc_entries, content_start_page)

# Final merge
writer = PdfWriter()

# Add cover page
for page in cover_reader.pages:
    writer.add_page(page)

# Add TOC pages
for page in toc_reader.pages:
    writer.add_page(page)

# Add content pages and track section titles for each page
current_page = cover_pages + toc_pages
bookmarks = []
page_sections = {}  # Map page index -> section title

current_section = ''
for i, pdf_path in enumerate(pdf_files):
    title, _, depth, site_path = toc_entries[i]
    bookmarks.append((title, current_page, depth, site_path))

    # Update current section for top-level entries (depth <= 1)
    if depth <= 1:
        current_section = title

    reader = PdfReader(pdf_path)
    for j, page in enumerate(reader.pages):
        writer.add_page(page)
        # Store section title for this page (use document title for first page of section)
        page_sections[current_page + j] = title if j == 0 else current_section
    current_page += len(reader.pages)

# Overlay page footers on TOC and content pages (skip cover only)
total_pages = len(writer.pages)
toc_start = cover_pages
content_start = cover_pages + toc_pages
footer_page_count = total_pages - cover_pages  # All pages after cover get footers

print(f'Adding page footers to {toc_pages} TOC + {total_pages - content_start} content pages...')

# Add footers to TOC pages (roman numerals style: i, ii, iii...)
for toc_idx in range(toc_pages):
    page_idx = toc_start + toc_idx
    page = writer.pages[page_idx]
    # Use roman numerals for TOC pages
    roman_num = ['i', 'ii', 'iii', 'iv', 'v', 'vi', 'vii', 'viii', 'ix', 'x'][toc_idx] if toc_idx < 10 else str(toc_idx + 1)
    footer_reader = create_page_footer_overlay(
        page_num=roman_num,
        total_pages=footer_page_count,
        book_title=book_title,
        section_title='Table of Contents'
    )
    page.merge_page(footer_reader.pages[0])

# Add footers to content pages (arabic numerals: 1, 2, 3...)
for page_idx in range(content_start, total_pages):
    page = writer.pages[page_idx]
    display_page_num = page_idx - content_start + 1  # Content pages start at 1
    section = page_sections.get(page_idx, '')

    footer_reader = create_page_footer_overlay(
        page_num=display_page_num,
        total_pages=total_pages - content_start,
        book_title=book_title,
        section_title=section
    )
    page.merge_page(footer_reader.pages[0])

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
