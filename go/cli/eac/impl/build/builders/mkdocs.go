// mkdocs.go - Build functions for MkDocs module types
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	// Native handler disabled - using tool system's mkdocs-builder directly
	// RegisterHandler(&MkDocsHandler{})
}

// ResolveHostRepoRoot returns the correct repo root path for Docker operations.
// In Docker-in-Docker mode, uses R2R_HOST_REPOROOT environment variable if set.
// Exported for use by mkdocs/pdf.go handler.
func ResolveHostRepoRoot(containerRoot string, logWriter io.Writer) string {
	if !IsDockerInDocker() {
		return containerRoot
	}
	if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
		Logln(logWriter, "   Docker-in-Docker: using host path %s", hostRoot)
		return hostRoot
	}
	return containerRoot
}

// toDockerPath converts a file path to Docker-compatible format (forward slashes).
func toDockerPath(workspaceRoot, filePath string) string {
	relPath, err := filepath.Rel(workspaceRoot, filePath)
	if err != nil {
		relPath = filePath
	}
	return strings.ReplaceAll(relPath, "\\", "/")
}

// getPDFConcurrency returns the internal Playwright page concurrency for a single PDF export.
// This controls how many pages within one book are rendered in parallel.
// Uses memory-based detection to scale appropriately to system resources.
// Note: Cross-book concurrency is controlled by the component scheduler's weighted semaphore.
func getPDFConcurrency(workspaceRoot string) int {
	// Use memory-based detection from environments package
	// This already handles CI vs local and returns 1-4 based on RAM tier
	return environments.GetPDFExportConcurrency()
}

// formatDockerMemory formats bytes as Docker memory string (e.g., "4g").
// Minimum 2GB, maximum 24GB for beefy machines.
func formatDockerMemory(bytes int64) string {
	const gbSize = 1024 * 1024 * 1024
	gbVal := bytes / gbSize
	if gbVal < 2 {
		return "2g" // minimum for any build
	}
	if gbVal > 24 {
		return "24g" // cap for very large allocations
	}
	return fmt.Sprintf("%dg", gbVal)
}

// formatDockerShm calculates shared memory for Docker (1/4 of container, min 512MB, max 4GB).
// Shared memory is used by Chromium/Playwright for rendering.
func formatDockerShm(containerMemory int64) string {
	const (
		minShm = 512 * 1024 * 1024      // 512MB
		maxShm = 4 * 1024 * 1024 * 1024 // 4GB for beefy machines
		mb     = 1024 * 1024
	)

	shm := containerMemory / 4
	if shm < minShm {
		return "512m"
	}
	if shm >= maxShm {
		return "4g"
	}
	return fmt.Sprintf("%dm", shm/mb)
}

// getEffectiveWeight returns the scheduling weight for a component type.
// Returns the component type's resources.cpus from component-types.yml.
func getEffectiveWeight(componentType string) int {
	cfg := config.Global()
	if cfg != nil && cfg.ComponentTypes != nil {
		if ct := cfg.ComponentTypes.Get(componentType); ct != nil {
			return ct.GetWeight()
		}
	}
	return 1 // default weight
}

// calculateWeightedMemory calculates container memory allocation based on weight.
// Memory scales with weight: weight 1 = 2.5GB, weight 4 = 10GB, etc.
// Capped at available container memory (half of host RAM).
func calculateWeightedMemory(weight int) int64 {
	// 2.5GB per weight unit - balances beefy machines with lower-spec ones
	const baseMemoryPerWeight = 2560 * 1024 * 1024 // 2.5GB per weight unit
	maxContainerMem := environments.GetContainerMemoryBytes()

	weightedMem := int64(weight) * baseMemoryPerWeight
	if weightedMem > maxContainerMem {
		return maxContainerMem
	}
	return weightedMem
}

// MkDocsHandler builds MkDocs documentation sites using Docker.
type MkDocsHandler struct{}

func (h *MkDocsHandler) Name() string { return "mkdocs" }

func (h *MkDocsHandler) Capabilities() []string { return []string{"documentation", "container"} }

func (h *MkDocsHandler) Requirements() []string { return []string{"docker"} }

func (h *MkDocsHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available")
	}
	return nil
}

func (h *MkDocsHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{"site/"}
}

func (h *MkDocsHandler) Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}
	return buildMkDocsModule(concrete, workspaceRoot, outputDir, logWriter, opts)
}

// mkdocsDockerConfig holds resolved docker configuration for mkdocs builds.
type mkdocsDockerConfig struct {
	ImageName      string
	ContainerDir   string
	DockerfilePath string
	ContextPath    string
}

// getMkDocsDockerConfig resolves docker configuration from module first, then type, then defaults.
// For PDF builds, always use pdf-oci container (module config only applies to site builds).
// For site builds, module docker_build config is respected.
func getMkDocsDockerConfig(module *modules.ModuleContract, workspaceRoot string, isPDF bool) mkdocsDockerConfig {
	// Defaults based on output type
	// Use :local tag for local builds to match what buildx creates
	var defaultImage, defaultContainer string
	if isPDF {
		defaultImage = "pdf-oci:local"
		defaultContainer = "pdf-oci"
	} else {
		defaultImage = "site-render-oci:local"
		defaultContainer = "site-render-oci"
	}

	cfg := mkdocsDockerConfig{
		ImageName:    defaultImage,
		ContainerDir: defaultContainer,
	}

	// For PDF builds, always use the PDF container - don't let module config override
	// PDF requires specific plugins (mkdocs-exporter, playwright) only in pdf-oci container
	if isPDF {
		cfg.ContextPath = filepath.Join(workspaceRoot, "containers", cfg.ContainerDir)
		cfg.DockerfilePath = filepath.Join(cfg.ContextPath, "Dockerfile")
		return cfg
	}

	// For site builds: check if module has a dockerfile package with docker_build config
	// (This is rare - book modules typically use shared containers like site-render-oci)
	dockerfilePkg := module.Components["dockerfile"]
	if dockerfilePkg != nil && dockerfilePkg.DockerBuild != nil && len(dockerfilePkg.DockerBuild) > 0 {
		if tags, ok := dockerfilePkg.DockerBuild["tags"].([]interface{}); ok && len(tags) > 0 {
			if tag, ok := tags[0].(string); ok {
				cfg.ImageName = tag
			}
		} else if container, ok := dockerfilePkg.DockerBuild["container"].(string); ok {
			cfg.ImageName = container + ":local"
		}

		if context, ok := dockerfilePkg.DockerBuild["context"].(string); ok {
			cfg.ContainerDir = filepath.Base(context)
			cfg.ContextPath = filepath.Join(workspaceRoot, context)
		}

		if dockerfile, ok := dockerfilePkg.DockerBuild["dockerfile"].(string); ok {
			cfg.DockerfilePath = filepath.Join(workspaceRoot, dockerfile)
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
// When opts.Component is set, only that specific book is built (component-level parallelism).
// When opts.Component is empty, all books for the module are built in parallel.
func buildMkDocsModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	// Load config to check for book configuration
	cfg, loadErr := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if loadErr == nil && cfg != nil {
		_ = cfg.LoadBooks(false)      //nolint:errcheck // best-effort config load
		_ = cfg.LoadRepository(false) //nolint:errcheck // best-effort config load for books by module
		// Check if module has ANY books defined
		allBooks := cfg.GetBooksByModule(module.Moniker)
		if len(allBooks) > 0 {
			// Component-level parallelism: if a specific component is requested, build only that book
			if opts.Component != "" {
				// Find the book matching this component name
				for _, book := range allBooks {
					if book.Name == opts.Component {
						Logln(logWriter, "📚 Building component book: %s (module: %s)", book.Name, module.Moniker)
						return buildModuleBooks(module, []*config.Book{book}, workspaceRoot, outputDir, logWriter)
					}
				}
				// Component name doesn't match any book - this shouldn't happen if GetHandlersForModule is correct
				Logln(logWriter, "⚠️  Book not found for component: %s (module: %s)", opts.Component, module.Moniker)
				return 1
			}

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
			Logln(logWriter, "\n=== Building book: %s ===", module.Moniker)
			Logln(logWriter, "📚 All %d book(s) have default: false - skipping (use --all to build)", len(allBooks))
			Logln(logWriter, "✅ Build skipped (no default books)")
			return 0
		}
	}

	// No books configured - standard HTML-only build
	Logln(logWriter, "\n=== Building book: %s ===", module.Moniker)

	// Check for book configuration and run preprocessing if found
	// pdfMode=false for standard HTML builds
	stagingDir, bookUsed := checkAndPreprocessBook(module.Moniker, workspaceRoot, outputDir, logWriter, false)
	if bookUsed && stagingDir == "" {
		// Preprocessing failed
		return 1
	}

	// Check if we can skip the build using content hash caching
	if stagingDir != "" {
		canSkip, reason := ShouldSkipSiteBuild(module.Moniker, stagingDir, workspaceRoot, outputDir, opts.Reproducible)
		if canSkip {
			Logln(logWriter, "⏭️  Skipping site build: %s", reason)
			Logln(logWriter, "✅ MkDocs site cached (unchanged)")
			return exitCodeSkipped // -1 signals cached to TUI (blue tab)
		}
		Logln(logWriter, "   Cache miss: %s", reason)
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

	// Copy mkdocs macros script for footer generation
	macrosSource := filepath.Join(workspaceRoot, "containers", "site-render-oci", "mkdocs_macros.py")
	macrosTarget := filepath.Join(outputDir, "main.py")
	if macrosData, err := os.ReadFile(macrosSource); err == nil {
		if err := os.WriteFile(macrosTarget, macrosData, 0o644); err != nil {
			Logln(logWriter, "   ⚠️  Failed to copy mkdocs macros script: %v", err)
		} else {
			Logln(logWriter, "   Macros: %s", macrosTarget)
		}
	}

	// For Docker-in-Docker: use host path for volume mount
	hostRepoRoot := ResolveHostRepoRoot(workspaceRoot, logWriter)
	isDinD := IsDockerInDocker()

	// Get docker configuration from module first, then type, then defaults
	dockerCfg := getMkDocsDockerConfig(module, workspaceRoot, false)
	imageName := dockerCfg.ImageName

	// Get context path for image manager
	contextPath := dockerCfg.ContextPath

	if err := ensureMkDocsImage(imageName, workspaceRoot, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	// Calculate the site output directory (local path within container)
	siteDir := filepath.Join(outputDir, "site")

	if err := os.MkdirAll(siteDir, 0o755); err != nil {
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
	dockerConfigPath := toDockerPath(workspaceRoot, configPath)

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

	exitCode := RunCommandWithLog(context.Background(), workspaceRoot, logWriter, "docker", buildArgs...)

	if acceptWarnings && exitCode != 0 {
		Logln(logWriter, "⚠️  Build completed with warnings (accepted)")
		exitCode = 0
	}

	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs build failed")
		return exitCode
	}

	// Record successful build in cache
	if stagingDir != "" {
		if err := RecordSiteBuildComplete(module.Moniker, stagingDir, workspaceRoot, outputDir); err != nil {
			Logln(logWriter, "   ⚠️  Failed to save build cache: %v", err)
		}
	}

	Logln(logWriter, "✅ MkDocs site built successfully")
	Logln(logWriter, "   HTML Output: %s", siteDir)
	return 0
}

// ensureMkDocsImage ensures the mkdocs Docker image is available and up-to-date.
// Uses tool.ImageManager for environment-aware image resolution:
// - Devbox: Build from Dockerfile if stale
// - CI: Pull from GHCR
func ensureMkDocsImage(imageName, workspaceRoot, contextPath string, logWriter io.Writer) error {
	// Extract container name from image (e.g., "pdf-oci:local" -> "pdf-oci")
	containerName := strings.Split(imageName, ":")[0]

	// Create ImageManager for local container handling
	imgMgr := tool.NewImageManager(
		workspaceRoot,
		environments.IsCI(),
		"ready-to-release", // GHCR org
		logWriter,
	)

	// Create a synthetic tool definition for the local container
	localTool := &tool.ToolDefinition{
		ID:        containerName,
		Type:      tool.ToolTypeContainer,
		LocalPath: strings.TrimPrefix(contextPath, workspaceRoot+string(filepath.Separator)),
	}
	// Normalize path separators
	localTool.LocalPath = filepath.ToSlash(localTool.LocalPath)

	// Use ImageManager to ensure image (handles staleness, CI pull, local build)
	ctx := context.Background()
	resolvedImage, err := imgMgr.EnsureImage(ctx, localTool)
	if err != nil {
		return fmt.Errorf("ensuring image %s: %w", imageName, err)
	}

	// If ImageManager returned a different image ref (e.g., GHCR in CI), tag it as local
	if resolvedImage != "" && resolvedImage != imageName {
		Logln(logWriter, "   Tagging %s as %s", resolvedImage, imageName)
		exitCode := RunCommandWithLog(context.Background(), "", logWriter, "docker", "tag", resolvedImage, imageName)
		if exitCode != 0 {
			return fmt.Errorf("failed to tag image %s as %s", resolvedImage, imageName)
		}
	}

	return nil
}

// buildMkDocsWithTheme builds a PDF with a specific theme (dark or light)
// cleanBuild controls whether to use --clean flag; set false when building multiple themes
// to preserve PDFs from previous theme builds.
func buildMkDocsWithTheme(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, theme string, cleanBuild bool) int {
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
// bookTitle and bookDescription are used for the PDF cover page.
func buildMkDocsWithThemeAndStaging(module *modules.ModuleContract, bookName, bookTitle, bookDescription, workspaceRoot, outputDir string, logWriter io.Writer, theme string, cleanBuild bool, stagingDir string) int {
	if theme == "" {
		theme = "dark"
	}

	// Default bookTitle to bookName if not provided
	if bookTitle == "" {
		bookTitle = bookName
	}

	Logln(logWriter, "\n=== Building book: %s (PDF %s) ===", module.Moniker, theme)

	// Generate mkdocs.yml from PDF template
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
	hostRepoRoot := ResolveHostRepoRoot(workspaceRoot, logWriter)
	isDinD := IsDockerInDocker()

	// Get docker configuration from module first, then type, then defaults (PDF mode)
	dockerCfg := getMkDocsDockerConfig(module, workspaceRoot, true)
	imageName := dockerCfg.ImageName

	// Get context path for image manager
	contextPath := dockerCfg.ContextPath

	if err := ensureMkDocsImage(imageName, workspaceRoot, contextPath, logWriter); err != nil {
		Logln(logWriter, "❌ Failed to ensure Docker image: %v", err)
		return 1
	}

	// Calculate the site output directory
	siteDir := filepath.Join(outputDir, "site")
	if err := os.MkdirAll(siteDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// MkDocs resolves --site-dir relative to the config file directory when using -f
	dockerVolume := FormatDockerVolumePath(hostRepoRoot)
	dockerSiteDir := "site"
	dockerConfigPath := toDockerPath(workspaceRoot, configPath)

	// Calculate resource allocation based on work unit weight
	// Weight determines both CPU and memory: weight 4 = 4 CPUs, 8GB RAM
	weight := getEffectiveWeight("book")
	weightedMemory := calculateWeightedMemory(weight)
	memoryLimit := formatDockerMemory(weightedMemory)
	shmSize := formatDockerShm(weightedMemory)
	cpuLimit := fmt.Sprintf("%d", weight)

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
		"--cpus", cpuLimit,
		"--memory", memoryLimit,
		"--shm-size", shmSize,
		"-e", "ENABLE_PDF_EXPORT=true",
	}

	Logln(logWriter, "   PDF Export: enabled (mkdocs-exporter)")
	Logln(logWriter, "   Resources: weight=%d, cpus=%s, memory=%s, shm=%s", weight, cpuLimit, memoryLimit, shmSize)

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

	// Note: Concurrency is now controlled by the component scheduler's weighted semaphore.
	// PDF builds have higher weights (e.g., 4) to reflect their resource requirements.

	// Retry logic for PDF builds - Playwright can have transient timeouts
	maxRetries := 2
	var exitCode int
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			Logln(logWriter, "🔄 Retrying PDF build (attempt %d/%d)...", attempt, maxRetries)
		}

		exitCode = RunCommandWithLog(context.Background(), workspaceRoot, logWriter, "docker", buildArgs...)

		if exitCode == 0 {
			break // Success
		}

		if attempt < maxRetries {
			Logln(logWriter, "⚠️  PDF build failed, will retry...")
		}
	}

	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs PDF build failed (%s theme) after %d attempts", theme, maxRetries)
		return exitCode
	}

	// Merge individual PDFs into single document
	// mkdocs-exporter creates individual PDFs for each page, we merge them using pypdf
	dstPdfPath := filepath.Join(siteDir, "pdf", fmt.Sprintf("%s-%s.pdf", bookName, theme))

	Logln(logWriter, "📄 Merging individual PDFs...")

	if err := MergePDFs(siteDir, dstPdfPath, hostRepoRoot, workspaceRoot, stagingDir, imageName, bookTitle, bookDescription, logWriter, isDinD); err != nil {
		Logln(logWriter, "❌ PDF merge failed: %v", err)
		return 1
	}

	Logln(logWriter, "✅ MkDocs PDF built successfully (%s theme)", theme)
	Logln(logWriter, "   PDF Output: %s", dstPdfPath)
	return 0
}
