// mkdocs.go - Build functions for MkDocs module types
package builders

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gofrs/flock"
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
// Multiple books for the same module are built in parallel.
func BuildMkDocsModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	// Load config to check for book configuration
	cfg, _ := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if cfg != nil {
		cfg.LoadBooks(false)
		// Use filtered books unless --all flag is set
		var moduleBooks []*config.Book
		if opts.BuildAll {
			moduleBooks = cfg.GetBooksByModule(module.Moniker)
		} else {
			moduleBooks = cfg.GetDefaultBooksByModule(module.Moniker)
		}
		if len(moduleBooks) > 0 {
			return buildModuleBooks(module, moduleBooks, workspaceRoot, outputDir, logWriter)
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

	// Check for mkdocs.yml at repository root
	mkdocsConfig := filepath.Join(workspaceRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		Logln(logWriter, "⚠️  No mkdocs.yml found at: %s", mkdocsConfig)
		Logln(logWriter, "ℹ️  Skipping MkDocs build")
		return 0
	}

	Logln(logWriter, "📚 Building MkDocs site using Docker")
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

	// Remove PDF plugins for HTML-only builds
	patchedConfig = removePDFPlugins(patchedConfig)
	needsPatch = true

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

	// Build PDF books sequentially (Playwright/Docker are resource-intensive)
	for _, book := range pdfBooks {
		bookOutputDir := filepath.Join(outputDir, book.Name)
		if err := os.MkdirAll(bookOutputDir, 0755); err != nil {
			Logln(logWriter, "❌ Failed to create output directory for book '%s': %v", book.Name, err)
			return 1
		}

		exitCode := buildSingleBook(module, book, workspaceRoot, bookOutputDir, logWriter)
		if exitCode != 0 {
			return exitCode
		}

		// Copy PDF to module output root
		bookOutput := book.GetOutput()
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
			srcPdf := filepath.Join(bookOutputDir, fmt.Sprintf("%s-%s.pdf", book.Name, theme))
			dstPdf := filepath.Join(outputDir, fmt.Sprintf("%s-%s.pdf", book.Name, theme))
			if err := copyFile(srcPdf, dstPdf); err != nil {
				Logln(logWriter, "⚠️  Failed to copy PDF to module root: %v", err)
			} else {
				Logln(logWriter, "   📄 %s-%s.pdf → module output root", book.Name, theme)
			}
		}
	}

	Logln(logWriter, "\n✅ All %d books built successfully", len(moduleBooks))
	return 0
}

// buildSingleBook builds a single book based on its output configuration
func buildSingleBook(module *modules.ModuleContract, book *config.Book, workspaceRoot string, outputDir string, logWriter io.Writer) int {
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
// Acquires a lock on the staging directory to prevent concurrent preprocessing.
func preprocessBook(book *config.Book, workspaceRoot string, outputDir string, logWriter io.Writer, pdfMode bool) (string, bool) {
	// Ensure staging base directory exists
	stagingBase := filepath.Join(workspaceRoot, "out", "staging")
	if err := os.MkdirAll(stagingBase, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create staging base directory: %v", err)
		return "", true
	}

	// Acquire lock for this book's staging directory
	lock, err := acquireStagingLock(book.Name, stagingBase)
	if err != nil {
		Logln(logWriter, "❌ Failed to acquire staging lock for book '%s': %v", book.Name, err)
		return "", true
	}
	// Note: Lock is intentionally NOT released here - it persists for the build duration
	// The lock file will be cleaned up when the process exits or on next successful build
	_ = lock // Suppress unused warning - lock is held via file descriptor

	// Use persistent staging directory at out/staging/{book-name}
	stagingDir := filepath.Join(stagingBase, book.Name)

	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		Logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return "", true
	}

	Logln(logWriter, "📚 Preprocessing book: %s", book.Name)

	// Run preprocessing
	preprocessor := books.NewPreprocessor(book, workspaceRoot, stagingDir, logWriter, pdfMode)
	if err := preprocessor.Preprocess(); err != nil {
		Logln(logWriter, "❌ Book preprocessing failed: %v", err)
		return "", true
	}

	return stagingDir, true
}

// acquireStagingLock acquires an exclusive lock for a book's staging directory.
// Returns the lock handle on success. The lock should be held for the duration of the build.
func acquireStagingLock(bookName string, stagingBase string) (*flock.Flock, error) {
	lockPath := filepath.Join(stagingBase, fmt.Sprintf(".lock-%s", bookName))

	lock := flock.New(lockPath)

	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return nil, fmt.Errorf("book '%s' staging is already in use by another process", bookName)
	}

	return lock, nil
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
	// Delegate to existing PDF build logic, passing book name for output naming
	return buildMkDocsWithThemeAndStaging(module, book.Name, workspaceRoot, outputDir, logWriter, theme, cleanBuild, stagingDir)
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
	// Check for mkdocs.yml at repository root
	mkdocsConfig := filepath.Join(workspaceRoot, "mkdocs.yml")
	if _, err := os.Stat(mkdocsConfig); os.IsNotExist(err) {
		Logln(logWriter, "⚠️  No mkdocs.yml found at: %s", mkdocsConfig)
		Logln(logWriter, "ℹ️  Skipping MkDocs build")
		return 0
	}

	Logln(logWriter, "📚 Building MkDocs site using Docker")
	Logln(logWriter, "   Config: %s", mkdocsConfig)

	// Read and patch mkdocs.yml
	originalConfig, err := os.ReadFile(mkdocsConfig)
	if err != nil {
		Logln(logWriter, "❌ Failed to read mkdocs.yml: %v", err)
		return 1
	}

	patchedConfig := removePDFPlugins(string(originalConfig))

	// Override docs_dir for staging
	if stagingDir != "" {
		relStagingDir, _ := filepath.Rel(workspaceRoot, stagingDir)
		relStagingDir = filepath.ToSlash(relStagingDir)
		patchedConfig = patchDocsDir(patchedConfig, relStagingDir)
		Logln(logWriter, "   Using staging: %s", relStagingDir)
	}

	// Write patched config
	patchedConfigPath := filepath.Join(outputDir, "mkdocs.yml")
	if err := os.WriteFile(patchedConfigPath, []byte(patchedConfig), 0644); err != nil {
		Logln(logWriter, "❌ Failed to write patched mkdocs.yml: %v", err)
		return 1
	}

	// Docker setup
	hostRepoRoot := workspaceRoot
	isDinD := isDockerInDocker()
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

	relSiteDir, _ := filepath.Rel(workspaceRoot, siteDir)
	dockerVolume := FormatDockerVolumePath(hostRepoRoot)
	dockerSiteDir := strings.ReplaceAll(relSiteDir, "\\", "/")

	buildArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
	}

	// Mount patched config
	var patchedConfigMount string
	if isDinD {
		relPatchedConfig, _ := filepath.Rel(workspaceRoot, patchedConfigPath)
		patchedConfigMount = FormatDockerVolumePath(filepath.Join(hostRepoRoot, relPatchedConfig))
	} else {
		patchedConfigMount = FormatDockerVolumePath(patchedConfigPath)
	}
	buildArgs = append(buildArgs, "-v", patchedConfigMount+":/docs/mkdocs.yml:ro")

	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		buildArgs = append(buildArgs, "--user", fmt.Sprintf("%d:%d", uid, gid))
	}

	buildArgs = append(buildArgs,
		imageName,
		"mkdocs", "build",
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
	return buildMkDocsWithThemeAndStaging(module, module.Moniker, workspaceRoot, outputDir, logWriter, theme, cleanBuild, stagingDir)
}

// buildMkDocsWithThemeAndStaging builds a PDF with a specific theme using a pre-computed staging directory
// This allows multiple theme builds to share preprocessing work
// bookName is used for the final PDF naming: {bookName}-{theme}.pdf
func buildMkDocsWithThemeAndStaging(module *modules.ModuleContract, bookName string, workspaceRoot string, outputDir string, logWriter io.Writer, theme string, cleanBuild bool, stagingDir string) int {
	if theme == "" {
		theme = "dark"
	}

	Logln(logWriter, "\n=== Building %s: %s (PDF %s) ===", module.Type, module.Moniker, theme)

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
	// When using staging directory, use staging-relative path (assets are copied to staging/assets/)
	themePath := fmt.Sprintf("assets/templates/pdf-%s", theme)
	if stagingDir == "" {
		themePath = fmt.Sprintf("docs/assets/templates/pdf-%s", theme)
	}
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
		"-e", "ENABLE_PDF_EXPORT=true",
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

	Logln(logWriter, "   PDF Export: enabled (mkdocs-exporter)")
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

	// Merge individual PDFs into single document
	// mkdocs-exporter creates individual PDFs for each page, we merge them using pypdf
	dstPdfPath := filepath.Join(siteDir, "pdf", fmt.Sprintf("%s-%s.pdf", bookName, theme))

	Logln(logWriter, "📄 Merging individual PDFs...")

	if err := mergePDFs(siteDir, dstPdfPath, hostRepoRoot, workspaceRoot, stagingDir, imageName, logWriter, isDinD); err != nil {
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
func mergePDFs(siteDir, outputPath, hostRepoRoot, workspaceRoot, stagingDir, imageName string, logWriter io.Writer, isDinD bool) error {
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

# Dark theme colors (matching PDF dark theme)
BG_COLOR = HexColor('#0d1117')
TEXT_COLOR = HexColor('#e6edf3')
ACCENT_COLOR = HexColor('#58a6ff')
MUTED_COLOR = HexColor('#8b949e')

def create_cover_page():
    """Create a cover page PDF with title, subtitle, and metadata."""
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=A4)
    width, height = A4

    # Dark background
    c.setFillColor(BG_COLOR)
    c.rect(0, 0, width, height, fill=True, stroke=False)

    # Title
    c.setFillColor(TEXT_COLOR)
    c.setFont('Helvetica-Bold', 36)
    c.drawCentredString(width/2, height - 8*cm, 'Ready-to-Release')

    # Subtitle
    c.setFont('Helvetica', 24)
    c.drawCentredString(width/2, height - 10*cm, 'Documentation')

    # Horizontal line
    c.setStrokeColor(ACCENT_COLOR)
    c.setLineWidth(2)
    c.line(4*cm, height - 12*cm, width - 4*cm, height - 12*cm)

    # Description
    c.setFillColor(MUTED_COLOR)
    c.setFont('Helvetica', 14)
    c.drawCentredString(width/2, height - 14*cm, 'Everything-as-Code Platform')
    c.drawCentredString(width/2, height - 15*cm, 'for Software Delivery Flows')

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
`, dockerSiteDir, dockerOutputPath, dockerStagingDir)

	args = append(args, imageName, "python3", "-c", pythonScript)

	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", args...)
	if exitCode != 0 {
		return fmt.Errorf("PDF merge exited with code %d", exitCode)
	}

	return nil
}

// patchMkDocsConfig patches the custom_template_path in mkdocs.yml
func patchMkDocsConfig(configContent string, themePath string) string {
	// Match custom_template_path line and replace its value
	re := regexp.MustCompile(`(?m)^(\s*custom_template_path:\s*).*$`)
	return re.ReplaceAllString(configContent, "${1}"+themePath)
}

// patchDocsDir patches the docs_dir in mkdocs.yml to use staging directory
// Also fixes stylesheet paths to work with the new docs_dir
func patchDocsDir(configContent string, newDocsDir string) string {
	// Match docs_dir line and replace its value
	re := regexp.MustCompile(`(?m)^(docs_dir:\s*).*$`)
	result := re.ReplaceAllString(configContent, "${1}"+newDocsDir)

	// Fix stylesheet paths: docs/assets/... -> {stagingDir}/assets/...
	// The exporter plugin resolves stylesheets relative to project root, not docs_dir
	// So we need to provide the full path from project root to staging assets
	result = patchStylesheetPaths(result, newDocsDir)

	return result
}

// patchStylesheetPaths fixes stylesheet paths for staging directory builds
// Changes docs/assets/... to {stagingDir}/assets/... since assets are copied to staging/assets/
// The exporter plugin resolves paths relative to project root, not relative to docs_dir
func patchStylesheetPaths(configContent string, stagingDir string) string {
	// Match stylesheet entries with docs/assets prefix
	re := regexp.MustCompile(`(?m)^(\s*-\s*)docs/assets/(.*)$`)
	return re.ReplaceAllString(configContent, "${1}"+stagingDir+"/assets/${2}")
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

		if strings.HasPrefix(trimmed, "- mermaid-to-svg:") || strings.HasPrefix(trimmed, "- with-pdf:") || strings.HasPrefix(trimmed, "- exporter:") {
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

// regenPDFWithWeasyPrint re-renders a PDF from processed HTML using WeasyPrint directly
// This bypasses mkdocs-with-pdf's internal WeasyPrint call, allowing us to use HTML
// with embedded CSS (which fixes WeasyPrint's image embedding bug in large documents)
func regenPDFWithWeasyPrint(htmlPath, pdfPath, hostRepoRoot, workspaceRoot, imageName string, logWriter io.Writer, isDinD bool) error {
	// Calculate relative paths for Docker
	relHTMLPath, err := filepath.Rel(workspaceRoot, htmlPath)
	if err != nil {
		return fmt.Errorf("calculating relative HTML path: %w", err)
	}
	relPDFPath, err := filepath.Rel(workspaceRoot, pdfPath)
	if err != nil {
		return fmt.Errorf("calculating relative PDF path: %w", err)
	}

	// Convert to Docker paths (forward slashes)
	dockerHTMLPath := "/docs/" + strings.ReplaceAll(relHTMLPath, "\\", "/")
	dockerPDFPath := "/docs/" + strings.ReplaceAll(relPDFPath, "\\", "/")

	volumeMountPath := hostRepoRoot
	dockerVolume := FormatDockerVolumePath(volumeMountPath)

	// Build Docker command to run WeasyPrint directly
	args := []string{
		"run", "--rm",
		"-v", dockerVolume + ":/docs",
		"-w", "/docs",
	}

	// In DinD mode, run as current user
	if isDinD {
		uid := os.Getuid()
		gid := os.Getgid()
		args = append(args, "--user", fmt.Sprintf("%d:%d", uid, gid))
	}

	// Run WeasyPrint via Python
	pythonCmd := fmt.Sprintf(`
import weasyprint
doc = weasyprint.HTML(filename='%s', base_url='/docs/')
doc.write_pdf('%s')
print('WeasyPrint regeneration complete')
`, dockerHTMLPath, dockerPDFPath)

	args = append(args, imageName, "python3", "-c", pythonCmd)

	Logln(logWriter, "   Running WeasyPrint on processed HTML...")
	exitCode := RunCommandWithLog(workspaceRoot, logWriter, "docker", args...)

	if exitCode != 0 {
		return fmt.Errorf("WeasyPrint exited with code %d", exitCode)
	}

	return nil
}
