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
	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/clibase/services"
	"github.com/ready-to-release/eac/go/commands/repository/internal/helputil"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

type serveCommand struct{}

var _ core.SimpleCommandPort = (*serveCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&serveCommand{},
	}
}

func (c *serveCommand) Name() string { return "serve" }

func (c *serveCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "serve",
		Short:         "Start development servers for documentation and visualization",
		IsParent:      true,
		SubcommandGroups: []core.SubcommandGroup{
			{Name: "Servers", Subcommands: []string{"docs", "design", "gource"}},
		},
		Examples: []string{"eac serve docs", "eac serve design src-auth", "eac serve gource"},
	}
}

func (c *serveCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	cmd, _ := registry.Global().Get("serve")
	helputil.PrintHelp(os.Stdout, cmd, registry.Global())
	return 1
}

// DocsFlags returns the flag specs for the serve docs subcommand.
func DocsFlags() []core.FlagSpec {
	return []core.FlagSpec{
		{Name: "no-browser", Type: "bool", DefaultValue: "false", Usage: "Don't open browser after starting server"},
		{Name: "port", Shorthand: "p", Type: "int", DefaultValue: "9000", Usage: "Port number for server (auto-allocated from 9000-9999 if not specified)"},
		{Name: "stop", Type: "bool", DefaultValue: "false", Usage: "Stop the running server"},
		{Name: "debug", Type: "bool", DefaultValue: "false", Usage: "Enable debug logging"},
		{Name: "static", Type: "bool", DefaultValue: "false", Usage: "Serve pre-built site (requires 'eac build docs' first)"},
	}
}

var log = logging.C()

// serveFlags holds all parsed command-line flags for the serve command.
type serveFlags struct {
	moduleMoniker string
	noBrowser     bool
	port          int
	stop          bool
	debug         bool
	static        bool
}

// parseServeFlags parses and validates the command-line arguments for serve.
// Returns the parsed flags, or an error if validation or parsing fails.
func parseServeFlags(args []string) (*serveFlags, error) {
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		return nil, err
	}

	f := &serveFlags{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--no-browser":
			f.noBrowser = true
		case "--stop":
			f.stop = true
		case "--debug":
			f.debug = true
		case "--static":
			f.static = true
		case "--port", "-p":
			if i+1 < len(args) {
				i++
				p, err := strconv.Atoi(args[i])
				if err != nil {
					return nil, fmt.Errorf("invalid port number: %s", args[i])
				}
				f.port = p
			} else {
				return nil, fmt.Errorf("--port requires a value")
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			if f.moduleMoniker == "" {
				f.moduleMoniker = arg
			}
		}
	}

	return f, nil
}

// initServeServices initializes the services layer and tool bridges.
// Returns the services instance and workspace root, or an error.
func initServeServices(debug bool) (*services.Services, string, error) {
	svc, err := services.New(core.SimpleServicesOptions{
		InitTools: true,
		DebugMode: debug,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize services: %v", err)
	}

	workspaceRoot := svc.WorkspaceRoot()

	if err := tool.InitializeGlobalBridges(workspaceRoot, svc.ConfigRoot()); err != nil {
		svc.Close()
		return nil, "", fmt.Errorf("failed to initialize tool system: %v", err)
	}

	return svc, workspaceRoot, nil
}

// serveStatic serves a pre-built site via nginx. Fails if the build output doesn't exist.
func serveStatic(workspaceRoot string, moduleConfig *ModuleServeConfig, f *serveFlags) int {
	// Verify build output exists
	indexPath := filepath.Join(moduleConfig.ContentPath, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		log.Errorf("No built site found at %s", moduleConfig.ContentPath)
		log.Error("Run 'eac build docs' first, then 'eac serve docs --static'.")
		return 1
	}

	containerName := fmt.Sprintf("cli-serve-%s", f.moduleMoniker)

	dockerClient, err := NewDockerClient(containerName)
	if err != nil {
		log.Errorf("Failed to initialize Docker: %v", err)
		return 1
	}
	defer dockerClient.Close()

	// Check if already running
	running, info, checkErr := dockerClient.IsRunning()
	if checkErr != nil {
		log.Errorf("Failed to check container status: %v", checkErr)
		return 1
	}
	if running && info != nil {
		log.Infof("%s static server is already running", f.moduleMoniker)
		log.Infof("URL: %s", info.URL)
		if !f.noBrowser {
			_, _ = dockerClient.OpenBrowserWithFallback(info.URL)
		}
		return 0
	}

	log.Infof("Starting %s static server...", f.moduleMoniker)
	info, err = dockerClient.StartContainer(workspaceRoot, moduleConfig.ContentPath, f.port)
	if err != nil {
		log.Errorf("Failed to start container: %v", err)
		return 1
	}

	log.Infof("Static server running: %s", info.URL)
	log.Infof("Stop with: eac serve %s --stop", f.moduleMoniker)

	if !f.noBrowser {
		_, _ = dockerClient.OpenBrowserWithFallback(info.URL)
	}
	if f.debug {
		log.Info("Debug mode: Streaming container logs (Press Ctrl+C to exit)")
		_ = dockerClient.StreamLogs()
	}

	return 0
}

// Serve starts the server for a module.
func Serve() int {
	args := os.Args[2:] // Skip program name and "serve"

	// Parse and validate flags
	f, err := parseServeFlags(args)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Initialize services
	svc, workspaceRoot, err := initServeServices(f.debug)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer svc.Close()

	// Validate module moniker
	if f.moduleMoniker == "" {
		log.Error("Error: module name is required")
		return 1
	}

	// Resolve module configuration
	moduleConfig, err := resolveModuleConfigFromEAC(svc.RawConfig(), workspaceRoot, f.moduleMoniker, "")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Handle --stop flag
	if f.stop {
		devContainerName := fmt.Sprintf("cli-serve-dev-%s", f.moduleMoniker)
		_ = handleStop(workspaceRoot, devContainerName, f.moduleMoniker)
		containerName := fmt.Sprintf("cli-serve-%s", f.moduleMoniker)
		return handleStop(workspaceRoot, containerName, f.moduleMoniker)
	}

	// Branch: --static serves pre-built output, default is live-reload dev mode
	if f.static {
		return serveStatic(workspaceRoot, moduleConfig, f)
	}
	return serveDevMode(svc, f.moduleMoniker, f.port, f.noBrowser, f.debug)
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


// resolveComponentBuildDir finds the actual build output directory for a component.
// Build directories use component_tool naming (e.g., "site_site" for component=site, tool=site).
// Falls back to the component name if no manifest is found.
func resolveComponentBuildDir(workspaceRoot, moduleMoniker, componentName string) string {
	reader := coreoutput.NewReader(workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionBuild, moduleMoniker)
	if err == nil {
		for _, m := range manifests {
			if m.Component == componentName {
				dirName := m.Component
				if m.Tool != "" {
					dirName += "_" + m.Tool
				}
				return dirName
			}
		}
	}
	return componentName
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

// startDevServerContainer starts the MkDocs dev server container and waits for it to be ready.
// Streams container logs in real-time so build errors are immediately visible.
// Returns the container info on success, or an error.
func startDevServerContainer(dockerClient *DockerClient, workspaceRoot, stagingDir string, port int) (*docker.ServeResult, error) {
	log.Info("Initializing container...")
	info, err := dockerClient.StartDevServer(workspaceRoot, stagingDir, port)
	if err != nil {
		return nil, fmt.Errorf("failed to start dev server: %v", err)
	}

	log.Infof("Container started on port %d", info.HostPort)

	// Stream container logs in the background so the user sees mkdocs output in real-time
	logCtx, logCancel := context.WithCancel(context.Background())
	logDone := make(chan struct{})
	go func() {
		defer close(logDone)
		_ = dockerClient.StreamLogsCtx(logCtx)
	}()

	if err := waitForServerReady(info.URL, 60*time.Second); err != nil {
		logCancel()
		<-logDone
		_ = dockerClient.StopContainer()
		return nil, fmt.Errorf("server failed to start: %v", err)
	}

	logCancel()
	<-logDone
	return info, nil
}

// printDevServerStatus prints the status information for a running dev server.
func printDevServerStatus(moduleMoniker string, url string) {
	log.Info("")
	log.Infof("Development server running")
	log.Infof("URL: %s", url)
	log.Info("")
	log.Info("Features:")
	log.Info("  - Live reload: Edit docs/ and see changes instantly")
	log.Info("  - Link warnings: Broken links shown in terminal (non-blocking)")
	log.Info("  - Reduced features: No PDF generation, basic mermaid")

	log.Info("")
	log.Info("File watching uses polling (works on Windows/macOS Docker Desktop)")
	log.Info("")
	log.Infof("For full site build: eac serve %s --static", moduleMoniker)
	log.Infof("Stop with: eac serve %s --stop", moduleMoniker)
}

// serveDevMode starts MkDocs in live-reload development mode.
// Pre-processes command markers into a staging directory, then serves via MkDocs container.
func serveDevMode(svc *services.Services, moduleMoniker string, port int, noBrowser, debug bool) int {
	workspaceRoot := svc.WorkspaceRoot()

	// Verify docs directory exists
	docsPath := filepath.Join(workspaceRoot, "docs")
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		log.Errorf("Error: docs directory not found: %s", docsPath)
		return 1
	}

	// Prepare staging directory (mirror docs + expand command markers)
	log.Info("Preparing documentation staging...")
	stagingDir, err := PrepareDevStaging(context.Background(), workspaceRoot, moduleMoniker, func(format string, args ...any) {
		log.Infof(format, args...)
	})
	if err != nil {
		log.Errorf("Failed to prepare staging: %v", err)
		return 1
	}

	containerName := fmt.Sprintf("cli-serve-dev-%s", moduleMoniker)

	// Stop any stale static serve container for this module to avoid confusion
	staticName := fmt.Sprintf("cli-serve-%s", moduleMoniker)
	if staticClient, err := NewDockerClient(staticName); err == nil {
		if running, _, _ := staticClient.IsRunning(); running {
			log.Info("Stopping stale static server...")
			_ = staticClient.StopContainer()
		}
		staticClient.Close()
	}

	dockerClient, err := NewDockerClient(containerName)
	if err != nil {
		log.Errorf("Failed to initialize Docker: %v", err)
		return 1
	}
	defer dockerClient.Close()

	// Stop existing dev container if running (staging has been refreshed)
	if running, _, _ := dockerClient.IsRunning(); running {
		log.Info("Stopping existing dev server for refresh...")
		_ = dockerClient.StopContainer()
	}

	// Start MkDocs dev server
	log.Info("Starting development server...")
	log.Info("Mode: Live-reload (changes reflect immediately)")
	log.Info("")

	info, err := startDevServerContainer(dockerClient, workspaceRoot, stagingDir, port)
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	printDevServerStatus(moduleMoniker, info.URL)

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
