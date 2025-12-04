// mkdocs.go - Build functions for MkDocs module types
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

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
// When opts.PDFMode is true, generates PDF documentation in addition to HTML.
// opts.PDFTheme controls which PDF theme(s) to build: "dark", "light", or "all".
func BuildMkDocsModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	// Handle PDF theme builds
	if opts.PDFMode {
		theme := opts.PDFTheme
		if theme == "" {
			theme = "dark" // Default to dark theme
		}

		if theme == "all" {
			// Build both themes sequentially
			Logln(logWriter, "\n=== Building %s: %s (PDF mode - all themes) ===", module.Type, module.Moniker)

			// First build dark theme (clean=true to start fresh)
			darkOpts := opts
			darkOpts.PDFTheme = "dark"
			if exitCode := buildMkDocsWithTheme(module, workspaceRoot, outputDir, logWriter, darkOpts, true); exitCode != 0 {
				return exitCode
			}

			// Then build light theme (clean=false to preserve dark PDF)
			lightOpts := opts
			lightOpts.PDFTheme = "light"
			return buildMkDocsWithTheme(module, workspaceRoot, outputDir, logWriter, lightOpts, false)
		}

		// Single theme build (always clean for single theme)
		return buildMkDocsWithTheme(module, workspaceRoot, outputDir, logWriter, opts, true)
	}

	// Standard HTML-only build
	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	// Check for book configuration and run preprocessing if found
	// pdfMode=false for standard HTML builds
	stagingDir, bookUsed := checkAndPreprocessBook(module.Moniker, workspaceRoot, outputDir, logWriter, false)
	if bookUsed && stagingDir == "" {
		// Preprocessing failed
		return 1
	}

	// Check for mkdocs.yml at repository root
	mkdocsConfig := filepath.Join(workspaceRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		Logln(logWriter, "⚠️  No mkdocs.yml found at: %s", mkdocsConfig)
		Logln(logWriter, "ℹ️  Skipping MkDocs build")
		return 0
	}

	if opts.PDFMode {
		Logln(logWriter, "📄 Building MkDocs site with PDF export using Docker")
	} else {
		Logln(logWriter, "📚 Building MkDocs site using Docker")
	}
	Logln(logWriter, "   Config: %s", mkdocsConfig)
	Logln(logWriter, "   WorkspaceRoot: %s", workspaceRoot)

	// Patch mkdocs.yml as needed:
	// 1. For site builds: remove PDF-specific plugins (container doesn't have WeasyPrint/Chromium)
	// 2. For book builds: override docs_dir to use staging directory
	// IMPORTANT: We write to a temp file and mount it in Docker, never modifying the source file
	originalConfig, err := os.ReadFile(mkdocsConfig)
	if err != nil {
		Logln(logWriter, "❌ Failed to read mkdocs.yml: %v", err)
		return 1
	}

	patchedConfig := string(originalConfig)
	needsPatch := false

	// Remove PDF plugins for non-PDF builds
	if !opts.PDFMode {
		patchedConfig = removePDFPlugins(patchedConfig)
		needsPatch = true
	}

	// Override docs_dir for book preprocessing (use staging directory)
	if stagingDir != "" {
		// Get relative path from workspace root to staging dir
		relStagingDir, _ := filepath.Rel(workspaceRoot, stagingDir)
		relStagingDir = filepath.ToSlash(relStagingDir) // Use forward slashes for YAML
		patchedConfig = patchDocsDir(patchedConfig, relStagingDir)
		needsPatch = true
		Logln(logWriter, "   Using staging: %s", relStagingDir)
	}

	// Write patched config to temp file in output directory (never modify source)
	var patchedConfigPath string
	if needsPatch {
		patchedConfigPath = filepath.Join(outputDir, "mkdocs.yml")
		if err := os.WriteFile(patchedConfigPath, []byte(patchedConfig), 0644); err != nil {
			Logln(logWriter, "❌ Failed to write patched mkdocs.yml: %v", err)
			return 1
		}
		Logln(logWriter, "   Patched config: %s", patchedConfigPath)
	}

	// For Docker-in-Docker: use host path for volume mount
	// When running inside a container, the Docker daemon runs on the host,
	// so volume mounts need host paths instead of container paths
	hostRepoRoot := workspaceRoot
	isDinD := isDockerInDocker()
	if isDinD {
		if hostRoot := os.Getenv("R2R_HOST_REPOROOT"); hostRoot != "" {
			hostRepoRoot = hostRoot
			Logln(logWriter, "   Docker-in-Docker: using host path %s", hostRoot)
		}
	}

	// Select Docker image based on build mode:
	// - mkdocs-site: Lean Alpine image (~150MB) for HTML builds
	// - mkdocs-pdf: Full image (~800MB) with WeasyPrint + Chromium for PDF
	var imageName, containerDir string
	if opts.PDFMode {
		imageName = "cli-mkdocs-pdf:latest"
		containerDir = "mkdocs-pdf"
	} else {
		imageName = "cli-mkdocs-site:latest"
		containerDir = "mkdocs-site"
	}

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

	// Calculate relative site dir for Docker volume mount
	// In DinD mode, we need the relative path from container's workspaceRoot
	// since that maps to the host's repo root
	relSiteDir, err := filepath.Rel(workspaceRoot, siteDir)
	if err != nil {
		Logln(logWriter, "❌ Failed to calculate relative path: %v", err)
		return 1
	}

	volumeMountPath := hostRepoRoot

	dockerVolume := FormatDockerVolumePath(volumeMountPath)
	dockerSiteDir := strings.ReplaceAll(relSiteDir, "\\", "/")

	// Check for --accept-warnings flag
	acceptWarnings := false
	for _, arg := range os.Args {
		if arg == "--accept-warnings" {
			acceptWarnings = true
			break
		}
	}

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
	}

	// Mount patched mkdocs.yml over the original (if patched)
	if patchedConfigPath != "" {
		var patchedConfigMount string
		if isDinD {
			// For DinD, convert to host path
			relPatchedConfig, _ := filepath.Rel(workspaceRoot, patchedConfigPath)
			patchedConfigMount = FormatDockerVolumePath(filepath.Join(hostRepoRoot, relPatchedConfig))
		} else {
			patchedConfigMount = FormatDockerVolumePath(patchedConfigPath)
		}
		buildArgs = append(buildArgs, "-v", patchedConfigMount+":/docs/mkdocs.yml:ro")
	}

	// Enable PDF export if PDFMode is set
	// This sets the environment variable that mkdocs-with-pdf checks
	if opts.PDFMode {
		buildArgs = append(buildArgs, "-e", "ENABLE_PDF_EXPORT=1")
		Logln(logWriter, "   PDF Export: enabled")
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
		"--site-dir", dockerSiteDir,
		"--clean",
	)

	// Determine strict mode: enabled by default, disabled for PDF mode or --accept-warnings
	useStrict := !opts.PDFMode && !acceptWarnings
	if useStrict {
		buildArgs = append(buildArgs, "--strict")
	}

	Logln(logWriter, "   Image: %s", imageName)
	Logln(logWriter, "   Output: %s", siteDir)
	Logln(logWriter, "   Volume: %s:/docs", dockerVolume)
	Logln(logWriter, "   SiteDir: %s", dockerSiteDir)

	if opts.PDFMode {
		Logln(logWriter, "   Mode: non-strict (PDF mode - CSS warnings expected)")
	} else if acceptWarnings {
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

	if opts.PDFMode {
		// PDF is generated inside the site directory under /pdf/
		pdfPath := filepath.Join(siteDir, "pdf", "ready-to-release-docs.pdf")
		if _, err := os.Stat(pdfPath); err == nil {
			Logln(logWriter, "   PDF Output: %s", pdfPath)
		} else {
			Logln(logWriter, "⚠️  PDF file not found at expected location: %s", pdfPath)
		}
	}

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

// buildMkDocsWithTheme builds a PDF with a specific theme (dark or light)
// cleanBuild controls whether to use --clean flag; set false when building multiple themes
// to preserve PDFs from previous theme builds
func buildMkDocsWithTheme(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions, cleanBuild bool) int {
	theme := opts.PDFTheme
	if theme == "" {
		theme = "dark"
	}

	Logln(logWriter, "\n=== Building %s: %s (PDF %s) ===", module.Type, module.Moniker, theme)

	// Check for book configuration and run preprocessing if found
	// pdfMode=true enables link normalization for PDF compatibility
	stagingDir, bookUsed := checkAndPreprocessBook(module.Moniker, workspaceRoot, outputDir, logWriter, true)
	if bookUsed && stagingDir == "" {
		// Preprocessing failed
		return 1
	}

	// Check for mkdocs.yml at repository root
	mkdocsConfig := filepath.Join(workspaceRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		Logln(logWriter, "⚠️  No mkdocs.yml found at: %s", mkdocsConfig)
		Logln(logWriter, "ℹ️  Skipping MkDocs build")
		return 0
	}

	// Read original mkdocs.yml
	originalConfig, err := os.ReadFile(mkdocsConfig)
	if err != nil {
		Logln(logWriter, "❌ Failed to read mkdocs.yml: %v", err)
		return 1
	}

	// Patch mkdocs.yml to use the correct theme template path
	// IMPORTANT: Write to temp file in output directory, never modify source
	themePath := fmt.Sprintf("docs/assets/templates/pdf-%s", theme)
	patchedConfig := patchMkDocsConfig(string(originalConfig), themePath)

	// Override docs_dir for book preprocessing (use staging directory)
	if stagingDir != "" {
		relStagingDir, _ := filepath.Rel(workspaceRoot, stagingDir)
		relStagingDir = filepath.ToSlash(relStagingDir)
		patchedConfig = patchDocsDir(patchedConfig, relStagingDir)
		Logln(logWriter, "   Using staging: %s", relStagingDir)
	}

	patchedConfigPath := filepath.Join(outputDir, "mkdocs.yml")
	if err := os.WriteFile(patchedConfigPath, []byte(patchedConfig), 0644); err != nil {
		Logln(logWriter, "❌ Failed to write patched mkdocs.yml: %v", err)
		return 1
	}

	Logln(logWriter, "📄 Building MkDocs site with PDF export (%s theme)", theme)
	Logln(logWriter, "   Config: %s", mkdocsConfig)
	Logln(logWriter, "   Theme: %s", themePath)
	Logln(logWriter, "   WorkspaceRoot: %s", workspaceRoot)

	// For Docker-in-Docker: use host path for volume mount
	hostRepoRoot := workspaceRoot
	isDinD := isDockerInDocker()
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

	relSiteDir, err := filepath.Rel(workspaceRoot, siteDir)
	if err != nil {
		Logln(logWriter, "❌ Failed to calculate relative path: %v", err)
		return 1
	}

	volumeMountPath := hostRepoRoot
	dockerVolume := FormatDockerVolumePath(volumeMountPath)
	dockerSiteDir := strings.ReplaceAll(relSiteDir, "\\", "/")

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
		"-e", "ENABLE_PDF_EXPORT=1",
	}

	// Mount patched mkdocs.yml over the original (never modify source files)
	var patchedConfigMount string
	if isDinD {
		relPatchedConfig, _ := filepath.Rel(workspaceRoot, patchedConfigPath)
		patchedConfigMount = FormatDockerVolumePath(filepath.Join(hostRepoRoot, relPatchedConfig))
	} else {
		patchedConfigMount = FormatDockerVolumePath(patchedConfigPath)
	}
	buildArgs = append(buildArgs, "-v", patchedConfigMount+":/docs/mkdocs.yml:ro")

	Logln(logWriter, "   PDF Export: enabled")
	Logln(logWriter, "   Patched config: %s", patchedConfigPath)

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

	// Rename the PDF to include theme name
	srcPdfPath := filepath.Join(siteDir, "pdf", "ready-to-release-docs.pdf")
	dstPdfPath := filepath.Join(siteDir, "pdf", fmt.Sprintf("ready-to-release-docs-%s.pdf", theme))

	if _, err := os.Stat(srcPdfPath); err == nil {
		// Rename the PDF
		if err := os.Rename(srcPdfPath, dstPdfPath); err != nil {
			Logln(logWriter, "⚠️  Failed to rename PDF: %v", err)
			Logln(logWriter, "   PDF Output: %s", srcPdfPath)
		} else {
			Logln(logWriter, "✅ MkDocs PDF built successfully (%s theme)", theme)
			Logln(logWriter, "   PDF Output: %s", dstPdfPath)
		}
		return 0
	}

	// PDF not generated - this is an error in PDF mode
	Logln(logWriter, "❌ PDF file not found at expected location: %s", srcPdfPath)

	// List what's actually in the site directory for diagnostics
	pdfDir := filepath.Join(siteDir, "pdf")
	if entries, err := os.ReadDir(pdfDir); err == nil {
		Logln(logWriter, "   Contents of %s:", pdfDir)
		for _, e := range entries {
			Logln(logWriter, "     - %s", e.Name())
		}
	} else if os.IsNotExist(err) {
		Logln(logWriter, "   PDF directory does not exist: %s", pdfDir)
		Logln(logWriter, "   This usually means mkdocs-with-pdf plugin didn't run")
		Logln(logWriter, "   Check that ENABLE_PDF_EXPORT=1 environment is set")
	} else {
		Logln(logWriter, "   Error reading PDF directory: %v", err)
	}

	return 1
}

// patchMkDocsConfig patches the custom_template_path in mkdocs.yml
func patchMkDocsConfig(configContent string, themePath string) string {
	// Match custom_template_path line and replace its value
	re := regexp.MustCompile(`(?m)^(\s*custom_template_path:\s*).*$`)
	return re.ReplaceAllString(configContent, "${1}"+themePath)
}

// patchDocsDir patches the docs_dir in mkdocs.yml to use staging directory
func patchDocsDir(configContent string, newDocsDir string) string {
	// Match docs_dir line and replace its value
	re := regexp.MustCompile(`(?m)^(docs_dir:\s*).*$`)
	return re.ReplaceAllString(configContent, "${1}"+newDocsDir)
}

// removePDFPlugins removes PDF-specific plugins from mkdocs.yml for site builds
// This allows using the lean mkdocs-site container without WeasyPrint/Chromium
func removePDFPlugins(configContent string) string {
	lines := strings.Split(configContent, "\n")
	var result []string
	skipUntilNextPlugin := false
	pluginIndent := 0

	for _, line := range lines {
		// Check if this is a plugin entry (starts with "  - ")
		trimmed := strings.TrimLeft(line, " ")
		currentIndent := len(line) - len(trimmed)

		if strings.HasPrefix(trimmed, "- mermaid-to-svg:") || strings.HasPrefix(trimmed, "- with-pdf:") {
			// Start skipping this plugin block
			skipUntilNextPlugin = true
			pluginIndent = currentIndent
			continue
		}

		if skipUntilNextPlugin {
			// Check if we've reached the next plugin or section
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				if currentIndent <= pluginIndent && strings.HasPrefix(trimmed, "- ") {
					// Next plugin entry at same level
					skipUntilNextPlugin = false
				} else if currentIndent < pluginIndent {
					// Back to parent level (e.g., new section)
					skipUntilNextPlugin = false
				}
			}
		}

		if !skipUntilNextPlugin {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
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

	// Check if there's a book for this module
	book := cfg.GetBookByName(moniker)
	if book == nil {
		return "", false
	}

	Logln(logWriter, "📚 Book configuration found for '%s'", moniker)

	// Use persistent staging directory at out/staging/{moniker}
	// This survives across builds, enabling:
	// - Mermaid SVG caching (rendered diagrams persist)
	// - Faster incremental builds (unchanged files already staged)
	// NOTE: We don't clean this directory - files are overwritten incrementally.
	// For a clean rebuild, delete out/staging/ manually.
	stagingDir := filepath.Join(workspaceRoot, "out", "staging", moniker)

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
