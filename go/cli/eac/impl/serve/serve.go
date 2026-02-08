// Command: serve
// Short: Start server for a module's build output
// Long: The serve command starts a Docker container to serve a module's build output.
// Long: For site-type modules (like docs), serves the HTML site.
// Long: For PDF-type modules (like books), serves the PDF directory listing.
// Long: Takes a module moniker as argument (e.g., docs, books).
// Args: module
// Flag.no-browser: type=bool, default=false, usage=Don't open browser after starting server
// Flag.port: type=int, shorthand=p, default=9000, usage=Port number for server (auto-allocated from 9000-9999 if not specified)
// Flag.stop: type=bool, default=false, usage=Stop the running server
// Flag.reload: type=bool, default=false, usage=Force reload (auto-detects container config changes)
// Flag.debug: type=bool, default=false, usage=Enable debug logging
// Flag.rebuild: type=bool, default=false, usage=Force rebuild before serving
// Flag.book: type=string, shorthand=b, default=, usage=Named book to serve (defaults to first 'site' book, or first book if no site)
// Flag.build-actual-site: type=bool, default=false, usage=Build full site before serving (disables live-reload dev mode)
package serve

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eac "github.com/ready-to-release/eac/go/adapters/eac"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/clibase/services"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

// commandFlags defines valid flags for the serve command

var log = logging.C()

func init() {
	registry.Register(Serve)
}

// Serve starts the server for a module.
func Serve() int {
	args := os.Args[2:] // Skip program name and "serve"

	// Validate flags early (before services init for fast feedback)
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags to get debug mode for services initialization
	var moduleMoniker string
	var noBrowser bool
	port := 0
	var stop bool
	var reload bool
	var debug bool
	var rebuild bool
	var namedBook string
	var buildActualSite bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--no-browser":
			noBrowser = true
		case "--stop":
			stop = true
		case "--reload":
			reload = true
		case "--rebuild":
			rebuild = true
		case "--debug":
			debug = true
		case "--port", "-p":
			if i+1 < len(args) {
				i++
				p, err := strconv.Atoi(args[i])
				if err != nil {
					log.Errorf("Error: invalid port number: %s", args[i])
					return 1
				}
				port = p
			} else {
				log.Errorf("Error: --port requires a value")
				return 1
			}
		case "--book", "-b":
			if i+1 < len(args) {
				i++
				namedBook = args[i]
			} else {
				log.Errorf("Error: --book requires a value")
				return 1
			}
		case "--build-actual-site":
			buildActualSite = true
		default:
			if strings.HasPrefix(arg, "-") {
				log.Errorf("Error: unknown flag: %s", arg)
				return 1
			}
			// First non-flag argument is module name
			if moduleMoniker == "" {
				moduleMoniker = arg
			}
		}
	}

	// Initialize services (replaces manual workspace, config, tools, logging setup)
	svc, err := services.New(core.SimpleServicesOptions{
		InitTools: true,
		DebugMode: debug,
	})
	if err != nil {
		log.Errorf("Error: failed to initialize services: %v", err)
		return 1
	}
	defer svc.Close()

	workspaceRoot := svc.WorkspaceRoot()

	// Initialize global bridges to enable tool.GetToolDefinition lookups
	if err := tool.InitializeGlobalBridges(workspaceRoot, svc.ConfigRoot()); err != nil {
		log.Errorf("Error: failed to initialize tool system: %v", err)
		return 1
	}

	if moduleMoniker == "" {
		log.Error("Error: module name is required")
		log.Info("")
		log.Info("Usage: eac serve <module> [flags]")
		if modules := listServableModulesFromConfig(svc.RawConfig()); len(modules) > 0 {
			log.Info("")
			log.Info("Available modules:")
			for _, m := range modules {
				log.Infof("  - %s", m)
			}
		}
		return 1
	}

	// Resolve module configuration
	moduleConfig, err := resolveModuleConfigFromEAC(svc.RawConfig(), workspaceRoot, moduleMoniker, namedBook)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	containerName := fmt.Sprintf("cli-serve-%s", moduleMoniker)

	// Handle --stop flag
	if stop {
		// Stop both dev and build mode containers
		devContainerName := fmt.Sprintf("cli-serve-dev-%s", moduleMoniker)
		_ = handleStop(workspaceRoot, devContainerName, moduleMoniker)
		return handleStop(workspaceRoot, containerName, moduleMoniker)
	}

	// Branch based on serve mode:
	// - Default: Development mode with live-reload (no build required)
	// - --build-actual-site: Full build then static serve (current behavior)
	if !buildActualSite && moduleConfig.IsSite {
		return serveDevMode(svc, moduleMoniker, port, noBrowser, debug, reload)
	}

	// Get Docker client
	dockerClient, err := NewDockerClient(containerName)
	if err != nil {
		log.Errorf("Failed to initialize Docker: %v", err)
		return 1
	}
	defer dockerClient.Close()

	// Check if already running
	running, info, err := dockerClient.IsRunning()
	if err != nil {
		log.Errorf("Failed to check container status: %v", err)
		return 1
	}

	if running && info != nil {
		// Check if image is stale - auto-reload if so
		imageStale, staleReason, staleErr := dockerClient.IsImageStale(workspaceRoot)
		if staleErr != nil {
			imageStale = false // Assume not stale on error
		}
		needsRestart := reload || rebuild || imageStale

		if needsRestart {
			if imageStale {
				log.Infof("Container config changed: %s", staleReason)
			}
			log.Info("Reloading server...")

			if port == 0 {
				port = info.HostPort
			}

			if err := dockerClient.StopContainer(); err != nil {
				log.Errorf("Failed to stop container: %v", err)
				return 1
			}

			if rebuild {
				if !rebuildModule(workspaceRoot, moduleMoniker, moduleConfig.ComponentName) {
					return 1
				}
			}
		} else {
			log.Infof("%s server is already running", moduleMoniker)
			log.Infof("URL: %s", info.URL)
			if !noBrowser {
				_, _ = dockerClient.OpenBrowserWithFallback(info.URL) //nolint:errcheck // best-effort browser open
			}
			return 0
		}
	}

	// Check staleness and auto-rebuild if needed
	if !rebuildIfNeeded(workspaceRoot, moduleConfig, rebuild, moduleConfig.ComponentName) {
		return 1
	}

	// Start container
	log.Infof("Starting %s server...", moduleMoniker)
	info, err = dockerClient.StartContainer(workspaceRoot, moduleConfig.ContentPath, port)
	if err != nil {
		log.Errorf("Failed to start container: %v", err)
		return 1
	}

	log.Info("")
	log.Infof("%s server is running", moduleMoniker)
	log.Infof("URL: %s", info.URL)

	if !noBrowser {
		_, _ = dockerClient.OpenBrowserWithFallback(info.URL) //nolint:errcheck // best-effort browser open
	}

	if !debug {
		log.Info("")
		log.Infof("Stop with: eac serve %s --stop", moduleMoniker)
	}

	if debug {
		log.Info("")
		log.Info("Debug mode: Streaming container logs (Press Ctrl+C to exit)")
		_ = dockerClient.StreamLogs() //nolint:errcheck // best-effort log streaming
	}

	log.Debugf("Serve completed: module=%s, url=%s", moduleMoniker, info.URL)
	return 0
}

// ModuleServeConfig holds configuration for serving a module.
type ModuleServeConfig struct {
	ModuleMoniker string
	ContentPath   string
	IsSite        bool
	ComponentName string // Component name for filtered builds (e.g., "site")
}

// resolveModuleConfigFromEAC resolves serve configuration for a module using pre-loaded config.
func resolveModuleConfigFromEAC(cfg *config.EACConfig, workspaceRoot, moduleMoniker, namedBook string) (*ModuleServeConfig, error) {
	if cfg == nil || cfg.Repository == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	module, exists := cfg.Repository.GetModule(moduleMoniker)
	if !exists || module == nil {
		return nil, fmt.Errorf("module not found: %s", moduleMoniker)
	}

	// Collect all servable content from the module
	// Servable types: site-render (HTML sites), pdf-render (PDFs), book (legacy)
	servableItems := getServableItems(module)
	if len(servableItems) == 0 {
		return nil, fmt.Errorf("module '%s' is not servable (no site, PDF, or book components)", moduleMoniker)
	}

	// Determine which item to serve
	var targetItem string
	var isSite bool
	if namedBook != "" {
		// User specified a target - validate it exists
		found := false
		for _, item := range servableItems {
			if item.name == namedBook {
				found = true
				targetItem = item.name
				isSite = item.isSite
				break
			}
		}
		if !found {
			names := make([]string, len(servableItems))
			for i, item := range servableItems {
				names[i] = item.name
			}
			return nil, fmt.Errorf("'%s' not found in module '%s' (available: %v)", namedBook, moduleMoniker, names)
		}
	} else {
		// Default: prefer "site" component, then first site-render, then first item
		for _, item := range servableItems {
			if item.name == "site" {
				targetItem = item.name
				isSite = item.isSite
				break
			}
		}
		if targetItem == "" {
			// Take first site-render type
			for _, item := range servableItems {
				if item.isSite {
					targetItem = item.name
					isSite = true
					break
				}
			}
		}
		if targetItem == "" {
			// Fallback to first item
			targetItem = servableItems[0].name
			isSite = servableItems[0].isSite
		}
	}

	// Determine content path based on type
	var contentPath string
	if isSite {
		// MkDocs outputs to site/ directory within the build output.
		// Build directories use component-tool naming (e.g., "site-site" for component=site, tool=site).
		// Resolve the actual directory from UoW manifests to handle the component-tool format.
		componentDir := resolveComponentBuildDir(workspaceRoot, moduleMoniker, targetItem)
		contentPath = filepath.Join(paths.BuildOutputPath(workspaceRoot, moduleMoniker), componentDir, "site")
	} else {
		// For non-site items (PDFs), serve the module root (contains all PDFs)
		contentPath = paths.BuildOutputPath(workspaceRoot, moduleMoniker)
	}

	return &ModuleServeConfig{
		ModuleMoniker: moduleMoniker,
		ContentPath:   contentPath,
		IsSite:        isSite,
		ComponentName: targetItem,
	}, nil
}

// servableItem represents a component that can be served.
type servableItem struct {
	name   string
	isSite bool // true for HTML sites, false for PDFs
}

// getServableItems returns all servable components from a module.
// Checks for site-render/docs-site (HTML), pdf-render/docs-pdf (PDFs), and book (legacy) types.
func getServableItems(module *config.Module) []servableItem {
	var items []servableItem

	// Check for site-render components (HTML sites)
	siteRenders := module.Components.GetComponentsByType(config.ComponentTypeSiteRender)
	for name := range siteRenders {
		items = append(items, servableItem{name: name, isSite: true})
	}

	// Check for docs-site components (HTML sites - same as site-render)
	docsSites := module.Components.GetComponentsByType(config.ComponentTypeDocsSite)
	for name := range docsSites {
		items = append(items, servableItem{name: name, isSite: true})
	}

	// Check for pdf-render components
	pdfRenders := module.Components.GetComponentsByType(config.ComponentTypePdfRender)
	for name := range pdfRenders {
		items = append(items, servableItem{name: name, isSite: false})
	}

	// Check for docs-pdf components (PDFs - same as pdf-render)
	docsPdfs := module.Components.GetComponentsByType(config.ComponentTypeDocsPdf)
	for name := range docsPdfs {
		items = append(items, servableItem{name: name, isSite: false})
	}

	// Check for legacy book components
	books := module.GetBooks()
	for _, name := range books {
		// Avoid duplicates if book name matches a pdf-render or docs-pdf
		found := false
		for _, item := range items {
			if item.name == name {
				found = true
				break
			}
		}
		if !found {
			// Books with name "site" are HTML sites, others are PDFs
			items = append(items, servableItem{name: name, isSite: name == "site"})
		}
	}

	return items
}

// listServableModulesFromConfig returns modules that can be served using pre-loaded config.
func listServableModulesFromConfig(cfg *config.EACConfig) []string {
	if cfg == nil || cfg.Repository == nil {
		return nil
	}

	var modules []string
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		// A module is servable if it has site-render, pdf-render, or book components
		if len(getServableItems(module)) > 0 {
			modules = append(modules, module.Moniker)
		}
	}
	return modules
}

// rebuildModule triggers a build for the module.
// Returns true if successful, false otherwise.
// Optional components filter which components to build.
func rebuildModule(workspaceRoot, moduleMoniker string, components ...string) bool {
	log.Infof("Rebuilding %s...", moduleMoniker)
	if err := runBuild(workspaceRoot, moduleMoniker, components...); err != nil {
		log.Errorf("Build failed: %v", err)
		return false
	}
	return true
}

// rebuildIfNeeded triggers a rebuild if forceRebuild is true or build is stale.
// Returns true if we should continue (no build needed or build succeeded).
// Returns false if build was needed and failed.
// Optional components filter which components to build.
func rebuildIfNeeded(workspaceRoot string, moduleConfig *ModuleServeConfig, forceRebuild bool, components ...string) bool {
	if forceRebuild {
		return rebuildModule(workspaceRoot, moduleConfig.ModuleMoniker, components...)
	}

	needsBuild, reason := checkStaleness(workspaceRoot, moduleConfig)
	if needsBuild {
		log.Infof("Build is stale: %s", reason)
		return rebuildModule(workspaceRoot, moduleConfig.ModuleMoniker, components...)
	}
	return true
}

// checkStaleness checks if the module build is stale.
func checkStaleness(workspaceRoot string, moduleConfig *ModuleServeConfig) (bool, string) {
	// Check if content directory exists
	if _, err := os.Stat(moduleConfig.ContentPath); os.IsNotExist(err) {
		return true, "content directory does not exist"
	}

	if moduleConfig.IsSite {
		// For sites, check if index.html exists
		indexPath := filepath.Join(moduleConfig.ContentPath, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			return true, "index.html missing"
		}
	} else {
		// For PDF directories, check if any PDF files exist
		pdfs, globErr := filepath.Glob(filepath.Join(moduleConfig.ContentPath, "*.pdf"))
		if globErr != nil || len(pdfs) == 0 {
			return true, "no PDF files found"
		}
	}

	// Check for UoW manifests to determine if module has been built
	reader := coreoutput.NewReader(workspaceRoot)
	if !reader.HasManifests(core.ActionBuild, moduleConfig.ModuleMoniker) {
		return true, moduleConfig.ModuleMoniker + " module not built (no UoW manifests)"
	}

	return false, ""
}

// resolveComponentBuildDir finds the actual build output directory for a component.
// Build directories use component-tool naming (e.g., "site-site" for component=site, tool=site).
// Falls back to the component name if no manifest is found.
func resolveComponentBuildDir(workspaceRoot, moduleMoniker, componentName string) string {
	reader := coreoutput.NewReader(workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionBuild, moduleMoniker)
	if err == nil {
		for _, m := range manifests {
			if m.Component == componentName {
				dirName := m.Component
				if m.Tool != "" {
					dirName += "-" + m.Tool
				}
				return dirName
			}
		}
	}
	return componentName
}

// runBuild executes the build command for a module.
// Optional components filter which components to build (e.g., "site").
func runBuild(workspaceRoot, moduleMoniker string, components ...string) error {
	log.Debugf("Running build command: module=%s, components=%v", moduleMoniker, components)

	args := []string{"build", moduleMoniker, "--no-tui"}
	for _, c := range components {
		if c != "" {
			args = append(args, "--component="+c)
		}
	}

	port := eac.New(workspaceRoot)
	result, err := port.Execute(context.Background(), args, &eac.ExecConfig{
		WorkspaceRoot: workspaceRoot,
		StdoutWriter:  os.Stdout,
		StderrWriter:  os.Stderr,
	})
	if err != nil {
		return err
	}
	if !result.Success() {
		return fmt.Errorf("build command failed with exit code %d", result.ExitCode)
	}
	return nil
}

// handleStop stops the running server.
func handleStop(_, containerName, moduleMoniker string) int {
	dockerClient, err := NewDockerClient(containerName)
	if err != nil {
		log.Errorf("Failed to initialize Docker: %v", err)
		return 1
	}
	defer dockerClient.Close()

	if err := dockerClient.StopContainer(); err != nil {
		if strings.Contains(err.Error(), "no container found") {
			log.Infof("%s server stopped", moduleMoniker)
			return 0
		}
		log.Errorf("Failed to stop container: %v", err)
		return 1
	}

	log.Infof("%s server stopped", moduleMoniker)
	return 0
}

// waitForServerReady polls the server URL until it responds with HTTP 200 or timeout.
// Returns nil on success, error on timeout or failure.
func waitForServerReady(url string, timeout time.Duration) error {
	log.Info("Waiting for server to be ready...")

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	deadline := time.Now().Add(timeout)
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++

		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Infof("Server ready (attempt %d)", attempt)
				return nil
			}
			// Server responded but not with 200 - MkDocs might still be building
			log.Debugf("Server responded with status %d, waiting...", resp.StatusCode)
		} else {
			log.Debugf("Connection attempt %d: %v", attempt, err)
		}

		// Progressive backoff: start fast, slow down
		sleepDuration := 500 * time.Millisecond
		if attempt > 5 {
			sleepDuration = 1 * time.Second
		}
		if attempt > 10 {
			sleepDuration = 2 * time.Second
		}

		time.Sleep(sleepDuration)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

// serveDevMode starts MkDocs in live-reload development mode.
// This is the new default behavior - no build required.
func serveDevMode(svc *services.Services, moduleMoniker string, port int, noBrowser, debug, reload bool) int {
	workspaceRoot := svc.WorkspaceRoot()

	// Resolve docs source path (not build output)
	docsPath := filepath.Join(workspaceRoot, "docs")

	// Verify docs directory exists
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		log.Errorf("Error: docs directory not found: %s", docsPath)
		return 1
	}

	containerName := fmt.Sprintf("cli-serve-dev-%s", moduleMoniker)

	// Get Docker client
	dockerClient, err := NewDockerClient(containerName)
	if err != nil {
		log.Errorf("Failed to initialize Docker: %v", err)
		return 1
	}
	defer dockerClient.Close()

	// Check if already running
	running, info, err := dockerClient.IsRunning()
	if err != nil {
		log.Errorf("Failed to check container status: %v", err)
		return 1
	}

	if running && info != nil {
		// Check if image is stale - auto-reload if so
		imageStale, staleReason, staleErr := dockerClient.IsDevImageStale(workspaceRoot)
		if staleErr != nil {
			imageStale = false // Assume not stale on error
		}
		needsRestart := reload || imageStale

		if needsRestart {
			if imageStale {
				log.Infof("Container config changed: %s", staleReason)
			}
			log.Info("Reloading development server...")

			if port == 0 {
				port = info.HostPort
			}

			if err := dockerClient.StopContainer(); err != nil {
				log.Errorf("Failed to stop container: %v", err)
				return 1
			}
		} else {
			log.Infof("Development server already running for %s", moduleMoniker)
			log.Infof("URL: %s", info.URL)
			if !noBrowser {
				_, _ = dockerClient.OpenBrowserWithFallback(info.URL)
			}
			return 0
		}
	}

	// Start MkDocs dev server
	log.Infof("Starting development server for %s...", moduleMoniker)
	log.Info("Mode: Live-reload (changes reflect immediately)")
	log.Info("")

	log.Info("Initializing container...")
	info, err = dockerClient.StartDevServer(workspaceRoot, port)
	if err != nil {
		log.Errorf("Failed to start dev server: %v", err)
		return 1
	}

	log.Infof("Container started on port %d", info.HostPort)

	// Verify server is ready and landing page is accessible
	if err := waitForServerReady(info.URL, 60*time.Second); err != nil {
		log.Errorf("Server failed to start: %v", err)
		log.Info("Check container logs with: docker logs " + containerName + "-" + strconv.Itoa(info.HostPort))
		// Stop the container since it's not working
		_ = dockerClient.StopContainer()
		return 1
	}

	log.Info("")
	log.Infof("Development server running")
	log.Infof("URL: %s", info.URL)
	log.Info("")
	log.Info("Features:")
	log.Info("  - Live reload: Edit docs/ and see changes instantly")
	log.Info("  - Link warnings: Broken links shown in terminal (non-blocking)")
	log.Info("  - Reduced features: No PDF generation, basic mermaid")

	// Note about polling-based file watching
	log.Info("")
	log.Info("File watching uses polling (works on Windows/macOS Docker Desktop)")
	log.Info("")
	log.Infof("For full site build: eac serve %s --build-actual-site", moduleMoniker)
	log.Infof("Stop with: eac serve %s --stop", moduleMoniker)

	if !noBrowser {
		_, _ = dockerClient.OpenBrowserWithFallback(info.URL)
	}

	if debug {
		log.Info("")
		log.Info("Debug mode: Streaming container logs (Press Ctrl+C to exit)")
		_ = dockerClient.StreamLogs()
	}

	return 0
}


