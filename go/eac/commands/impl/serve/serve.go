// Command: serve
// Short: Start server for a module's build output
// Long: The serve command starts a Docker container to serve a module's build output.
// Long: For site-type modules (like docs), serves the HTML site.
// Long: For PDF-type modules (like books), serves the PDF directory listing.
// Arg.module: type=string, required=true, usage=Module moniker to serve (e.g., docs, books)
// Flag.no-browser: type=bool, default=false, usage=Don't open browser after starting server
// Flag.port: type=int, shorthand=p, default=9000, usage=Port number for server (auto-allocated from 9000-9999 if not specified)
// Flag.stop: type=bool, default=false, usage=Stop the running server
// Flag.reload: type=bool, default=false, usage=Force reload (auto-detects container config changes)
// Flag.debug: type=bool, default=false, usage=Enable debug logging
// Flag.rebuild: type=bool, default=false, usage=Force rebuild before serving
// Flag.book: type=string, shorthand=b, default=, usage=Named book to serve (defaults to first 'site' book, or first book if no site)
package serve

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/buildstate"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"go.uber.org/zap"
)

var log = logging.C()

func init() {
	registry.Register(Serve)
}

// printHelp displays help information
func printHelp(workspaceRoot string) {
	log.Info("NAME")
	log.Info("    serve - Start server for a module's build output")
	log.Info("")
	log.Info("SYNOPSIS")
	log.Info("    eac serve <module> [flags]")
	log.Info("")
	log.Info("DESCRIPTION")
	log.Info("    Serves a module's build output via nginx container.")
	log.Info("    For site-type books: serves HTML site")
	log.Info("    For PDF books: serves directory listing with PDFs")
	log.Info("")
	log.Info("    By default serves the first 'site' book if one exists,")
	log.Info("    otherwise serves the first book in the module.")
	log.Info("")
	log.Info("ARGUMENTS")
	log.Info("    module              Module moniker to serve (required)")
	log.Info("")
	log.Info("FLAGS")
	log.Info("    -b, --book          Named book to serve (defaults to first site)")
	log.Info("    --no-browser        Don't open browser after starting server")
	log.Info("    -p, --port          Port number (default: auto-allocated 9000-9999)")
	log.Info("    --stop              Stop the running server")
	log.Info("    --reload            Force reload")
	log.Info("    --rebuild           Force rebuild before serving")
	log.Info("    --debug             Enable debug logging")
	log.Info("    -h, --help          Show this help message")
	log.Info("")
	log.Info("EXAMPLES")
	log.Info("    eac serve docs                  # Serve documentation site")
	log.Info("    eac serve books                 # Serve PDF books")
	log.Info("    eac serve docs --port 9001     # Serve on specific port")
	log.Info("    eac serve docs --rebuild       # Force rebuild before serving")
	log.Info("    eac serve docs --stop          # Stop the running server")
	log.Info("")

	// List servable modules
	if modules := listServableModules(workspaceRoot); len(modules) > 0 {
		log.Info("SERVABLE MODULES")
		for _, m := range modules {
			log.Infof("    %s", m)
		}
		log.Info("")
	}
}

// Serve starts the server for a module
func Serve() int {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	args := os.Args[2:] // Skip program name and "serve"

	var moduleMoniker string
	var noBrowser bool
	var port int = 0
	var stop bool
	var reload bool
	var debug bool
	var rebuild bool
	var namedBook string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			printHelp(workspaceRoot)
			return 0
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

	if moduleMoniker == "" {
		log.Error("Error: module name is required")
		log.Info("")
		log.Info("Usage: eac serve <module> [flags]")
		if modules := listServableModules(workspaceRoot); len(modules) > 0 {
			log.Info("")
			log.Info("Available modules:")
			for _, m := range modules {
				log.Infof("  - %s", m)
			}
		}
		return 1
	}

	// Initialize logger
	var logger *logging.Logger
	if debug {
		logger, err = logging.NewWithDebug("serve", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("serve", workspaceRoot)
	}
	if err != nil {
		log.Errorf("Error initializing logger: %v", err)
		return 1
	}
	defer logger.Sync()

	// Resolve module configuration
	moduleConfig, err := resolveModuleConfig(workspaceRoot, moduleMoniker, namedBook)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	containerName := fmt.Sprintf("cli-serve-%s", moduleMoniker)

	// Handle --stop flag
	if stop {
		return handleStop(workspaceRoot, logger, containerName, moduleMoniker)
	}

	// Get Docker client
	dockerClient, err := NewDockerClient(logger, containerName)
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
		imageStale, staleReason, _ := dockerClient.IsImageStale(workspaceRoot)
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
				log.Infof("Rebuilding %s...", moduleMoniker)
				if err := runBuild(workspaceRoot, moduleMoniker, logger); err != nil {
					log.Errorf("Build failed: %v", err)
					return 1
				}
			}
		} else {
			log.Infof("%s server is already running", moduleMoniker)
			log.Infof("URL: %s", info.URL)
			if !noBrowser {
				dockerClient.OpenBrowserWithFallback(info.URL)
			}
			return 0
		}
	}

	// Check staleness and auto-rebuild if needed
	if rebuild {
		log.Infof("Rebuilding %s...", moduleMoniker)
		if err := runBuild(workspaceRoot, moduleMoniker, logger); err != nil {
			log.Errorf("Build failed: %v", err)
			return 1
		}
	} else {
		needsBuild, reason := checkStaleness(workspaceRoot, moduleConfig)
		if needsBuild {
			log.Infof("Build is stale: %s", reason)
			log.Infof("Building %s...", moduleMoniker)
			if err := runBuild(workspaceRoot, moduleMoniker, logger); err != nil {
				log.Errorf("Build failed: %v", err)
				return 1
			}
		}
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
		dockerClient.OpenBrowserWithFallback(info.URL)
	}

	if !debug {
		log.Info("")
		log.Infof("Stop with: eac serve %s --stop", moduleMoniker)
	}

	if debug {
		log.Info("")
		log.Info("Debug mode: Streaming container logs (Press Ctrl+C to exit)")
		dockerClient.StreamLogs()
	}

	logger.Info("Serve completed", zap.String("module", moduleMoniker), zap.String("url", info.URL))
	return 0
}

// ModuleServeConfig holds configuration for serving a module
type ModuleServeConfig struct {
	ModuleMoniker string
	ContentPath   string
	IsSite        bool
}

// resolveModuleConfig resolves serve configuration for a module
func resolveModuleConfig(workspaceRoot string, moduleMoniker string, namedBook string) (*ModuleServeConfig, error) {
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.LoadRepository(false); err != nil {
		return nil, fmt.Errorf("failed to load repository: %w", err)
	}

	module, exists := cfg.Repository.GetModule(moduleMoniker)
	if !exists || module == nil {
		return nil, fmt.Errorf("module not found: %s", moduleMoniker)
	}

	// Check if module is servable (container type)
	if module.Type != "container" {
		return nil, fmt.Errorf("module '%s' is not servable (type: %s, expected: container)", moduleMoniker, module.Type)
	}

	// Check module has books
	if len(module.Books) == 0 {
		return nil, fmt.Errorf("module '%s' has no books to serve", moduleMoniker)
	}

	// Determine which book to serve
	var targetBook string
	if namedBook != "" {
		// User specified a book - validate it exists
		found := false
		for _, bookName := range module.Books {
			if bookName == namedBook {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("book '%s' not found in module '%s' (available: %v)", namedBook, moduleMoniker, module.Books)
		}
		targetBook = namedBook
	} else {
		// Default: first "site" book, or first book if no site
		for _, bookName := range module.Books {
			if bookName == "site" {
				targetBook = "site"
				break
			}
		}
		if targetBook == "" {
			targetBook = module.Books[0]
		}
	}

	// Determine content path based on book type
	isSite := targetBook == "site"
	var contentPath string
	if isSite {
		contentPath = filepath.Join(paths.BuildOutputPath(workspaceRoot, moduleMoniker), "site")
	} else {
		// For non-site books, serve the module root (contains all PDFs)
		contentPath = paths.BuildOutputPath(workspaceRoot, moduleMoniker)
	}

	return &ModuleServeConfig{
		ModuleMoniker: moduleMoniker,
		ContentPath:   contentPath,
		IsSite:        isSite,
	}, nil
}

// listServableModules returns modules that can be served
func listServableModules(workspaceRoot string) []string {
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		return nil
	}

	if err := cfg.LoadRepository(false); err != nil {
		return nil
	}

	var modules []string
	for _, module := range cfg.Repository.Modules {
		if module.Type == "container" && len(module.Books) > 0 {
			modules = append(modules, module.Moniker)
		}
	}
	return modules
}

// checkStaleness checks if the module build is stale
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
		pdfs, _ := filepath.Glob(filepath.Join(moduleConfig.ContentPath, "*.pdf"))
		if len(pdfs) == 0 {
			return true, "no PDF files found"
		}
	}

	// Check build state
	state, err := buildstate.Load(workspaceRoot)
	if err != nil || state == nil {
		return true, "no build state found"
	}

	if _, exists := state.Modules[moduleConfig.ModuleMoniker]; !exists {
		return true, moduleConfig.ModuleMoniker + " module not in build state"
	}

	return false, ""
}

// runBuild executes the build command for a module
func runBuild(workspaceRoot string, moduleMoniker string, logger *logging.Logger) error {
	binaryPath := paths.CommandsBinaryPath(workspaceRoot)

	logger.Debug("Running build command",
		zap.String("binary", binaryPath),
		zap.String("module", moduleMoniker))

	cmd := exec.Command(binaryPath, "build", moduleMoniker, "--no-tui")
	cmd.Dir = workspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// handleStop stops the running server
func handleStop(workspaceRoot string, logger *logging.Logger, containerName string, moduleMoniker string) int {
	dockerClient, err := NewDockerClient(logger, containerName)
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

// DockerClient wraps the internal serve package for module serving
type DockerClient struct {
	logger        *logging.Logger
	containerName string
	ctx           context.Context
}

// NewDockerClient creates a new Docker client
func NewDockerClient(logger *logging.Logger, containerName string) (*DockerClient, error) {
	return &DockerClient{
		logger:        logger,
		containerName: containerName,
		ctx:           context.Background(),
	}, nil
}

// Close closes the client (no-op for this wrapper)
func (c *DockerClient) Close() {}

// IsRunning checks if the container is running
func (c *DockerClient) IsRunning() (bool, *serve.ServeResult, error) {
	result, running, err := serve.IsServing(c.ctx, c.containerName)
	return running, result, err
}

// StartContainer starts the serve container
func (c *DockerClient) StartContainer(workspaceRoot string, contentPath string, port int) (*serve.ServeResult, error) {
	dockerfile := filepath.Join(workspaceRoot, "containers/static-site/Dockerfile")
	contextPath := filepath.Dir(dockerfile)

	serveConfig := &serve.ServeConfig{
		Name:  c.containerName,
		Image: "cli-static-site:latest",
		BuildInfo: &serve.BuildInfo{
			Dockerfile:  dockerfile,
			ContextPath: contextPath,
		},
		ContentPath:   contentPath,
		ContainerPath: "/usr/share/nginx/html",
		ContainerPort: 8000,
		PreferredPort: port,
		RestartPolicy: "unless-stopped",
	}

	return serve.StartServe(c.ctx, serveConfig)
}

// StopContainer stops the container
func (c *DockerClient) StopContainer() error {
	return serve.StopServe(c.ctx, c.containerName)
}

// IsImageStale checks if the container image is stale
func (c *DockerClient) IsImageStale(workspaceRoot string) (bool, string, error) {
	dockerfile := filepath.Join(workspaceRoot, "containers/static-site/Dockerfile")
	contextPath := filepath.Dir(dockerfile)

	serveConfig := &serve.ServeConfig{
		Image: "cli-static-site:latest",
		BuildInfo: &serve.BuildInfo{
			Dockerfile:  dockerfile,
			ContextPath: contextPath,
		},
	}

	return serve.CheckImageStale(c.ctx, serveConfig)
}

// OpenBrowserWithFallback opens the browser
func (c *DockerClient) OpenBrowserWithFallback(url string) (bool, error) {
	return serve.OpenBrowserWithFallback(url)
}

// StreamLogs streams container logs
func (c *DockerClient) StreamLogs() error {
	// Not implemented in simplified version
	return nil
}
