// Command: serve docs
// Short: Start or stop MkDocs server
// Long: The serve docs command manages the MkDocs documentation server using Docker.
// Long: It can start a server on a specified port, stop a running server, and open the documentation in your browser.
// Flag.no-browser: type=bool, default=false, usage=Don't open browser after starting server
// Flag.port: type=int, shorthand=p, default=9000, usage=Port number for MkDocs server (auto-allocated from 9000-9999 if not specified)
// Flag.stop: type=bool, default=false, usage=Stop the running MkDocs server
// Flag.debug: type=bool, default=false, usage=Enable debug mode with log streaming
package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	docsInternal "github.com/ready-to-release/eac/go/eac/commands/impl/docs/helper"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"go.uber.org/zap"
)

var log = logging.C()

func init() {
	registry.Register(ServeDocs)
}

// printHelp displays help information for the serve docs command
func printHelp() {
	log.Info("NAME")
	log.Info("    serve docs - Start or stop MkDocs server")
	log.Info("")
	log.Info("SYNOPSIS")
	log.Info("    eac serve docs [flags]")
	log.Info("")
	log.Info("DESCRIPTION")
	log.Info("    The serve docs command manages the MkDocs documentation server using Docker.")
	log.Info("    It can start a server on a specified port, stop a running server, and open")
	log.Info("    the documentation in your browser.")
	log.Info("")
	log.Info("FLAGS")
	log.Info("    --no-browser        Don't open browser after starting server")
	log.Info("    -p, --port          Port number for MkDocs server (default: auto-allocated 9000-9999)")
	log.Info("    --stop              Stop the running MkDocs server")
	log.Info("    --debug             Enable debug mode with log streaming")
	log.Info("    --skip-validation   Skip Docker validation (for testing)")
	log.Info("    -h, --help          Show this help message")
	log.Info("")
	log.Info("EXAMPLES")
	log.Info("    eac serve docs                  # Start server with auto-allocated port")
	log.Info("    eac serve docs --port 9001      # Start server on specific port")
	log.Info("    eac serve docs --no-browser     # Start without opening browser")
	log.Info("    eac serve docs --stop           # Stop the running server")
	log.Info("")
}

// writeDebugFile writes content to a debug file when debug mode is enabled.
// Files are written to out/logs/docs/<filename> in the workspace root.
func writeDebugFile(workspaceRoot string, logger *logging.Logger, filename string, content string) {
	if !logger.IsDebugMode() {
		return
	}

	debugDir := paths.CommandLogsPath(workspaceRoot, "docs")
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		logger.Warn("Failed to create debug directory", zap.Error(err))
		return
	}

	debugFile := filepath.Join(debugDir, filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		logger.Warn("Failed to write debug file", zap.String("file", debugFile), zap.Error(err))
	} else {
		logger.Debug("Saved debug file", zap.String("file", debugFile))
	}
}

// ServeDocs starts or stops MkDocs server
func ServeDocs() int {
	// Get workspace root early for logging
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	args := os.Args[3:] // Skip "go", "run", ".", "docs", and "serve"

	var noBrowser bool
	var port int = 0 // 0 means auto-allocate from 9000-9999 range
	var stop bool
	var debug bool
	var skipValidation bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--help", "-h":
			printHelp()
			return 0
		case "--no-browser":
			noBrowser = true
		case "--stop":
			stop = true
		case "--debug":
			debug = true
		case "--skip-validation":
			skipValidation = true
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
		default:
			log.Errorf("Error: unknown flag: %s", arg)
			return 1
		}
	}

	// Initialize logger based on debug mode
	var logger *logging.Logger
	if debug {
		logger, err = logging.NewWithDebug("docs", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("docs", workspaceRoot)
	}
	if err != nil {
		log.Errorf("Error initializing logger: %v", err)
		return 1
	}
	defer logger.Sync()

	logger.Debug("DocsServe command started",
		zap.Bool("noBrowser", noBrowser),
		zap.Int("port", port),
		zap.Bool("stop", stop),
		zap.Bool("debug", debug),
		zap.Bool("skipValidation", skipValidation))

	// Skip validation mode (for tests)
	if skipValidation {
		logger.Debug("Skipping Docker validation (test mode)")

		// Use state file to track mock "running" state
		stateFile := filepath.Join(workspaceRoot, "out", "test", ".mkdocs-mock-state")

		if stop {
			// Remove state file if it exists
			os.Remove(stateFile)
			log.Info("✅ MkDocs documentation server stopped")
			return 0
		}

		// Check if mock server is "already running"
		if _, err := os.Stat(stateFile); err == nil {
			// State file exists - server is "already running"
			// Read the port from state file
			stateData, _ := os.ReadFile(stateFile)
			runningPort := 9000
			if len(stateData) > 0 {
				fmt.Sscanf(string(stateData), "%d", &runningPort)
			}

			// If user requested a specific port and it's different, error
			if port != 0 && port != runningPort {
				log.Infof("❌ MkDocs is already running on port %d", runningPort)
				log.Infof("📚 Running at: http://localhost:%d", runningPort)
				log.Info("")
				log.Info("💡 To use a different port:")
				log.Info("  1. Stop the running container: go run . docs serve --stop")
				log.Infof("  2. Start with new port: go run . docs serve --port %d", port)
				return 1
			}

			// Server already running on expected port
			log.Info("ℹ️  MkDocs is already running")
			log.Infof("📚 Documentation: http://localhost:%d", runningPort)
			return 0
		}

		// Not running - start it
		if port == 0 {
			port = 9000
		}

		// Create state file directory if it doesn't exist
		os.MkdirAll(filepath.Dir(stateFile), 0755)

		// Write port to state file
		os.WriteFile(stateFile, []byte(fmt.Sprintf("%d", port)), 0644)

		log.Info("🚀 Starting MkDocs documentation server")
		log.Info("")
		log.Info("✅ MkDocs documentation server is running")
		log.Infof("📚 Documentation: http://localhost:%d", port)
		return 0
	}

	// Handle --stop flag
	if stop {
		return handleDocsStop(workspaceRoot, logger)
	}

	// Get Docker client
	dockerClient, err := docsInternal.NewClient(logger)
	if err != nil {
		logger.Error("Failed to initialize Docker client", zap.Error(err))
		log.Infof("❌ Failed to initialize: %v", err)
		return 1
	}
	defer dockerClient.Close()

	// Write debug info about Docker client
	if debug {
		writeDebugFile(workspaceRoot, logger, "docker-status.txt",
			fmt.Sprintf("Docker client initialized successfully\nWorkspace: %s\n", workspaceRoot))
	}

	// Check if already running
	running, info, err := dockerClient.IsRunning()
	if err != nil {
		logger.Error("Failed to check container status", zap.Error(err))
		log.Infof("❌ Failed to check container status: %v", err)
		return 1
	}

	if running && info != nil {
		// If user requested a specific port and it's different from running container, error
		if port != 0 && info.Port != port {
			logger.Error("MkDocs container already running on different port",
				zap.Int("runningPort", info.Port),
				zap.Int("requestedPort", port))
			log.Infof("❌ MkDocs is already running on port %d", info.Port)
			log.Infof("📚 Running at: %s", info.URL)
			log.Info("")
			log.Info("💡 To use a different port:")
			log.Info("  1. Stop the running container: go run . docs serve --stop")
			log.Infof("  2. Start with new port: go run . docs serve --port %d", port)
			return 1
		}

		// Container is running on the expected port (or port was auto-allocated)
		logger.Info("MkDocs container already running",
			zap.String("url", info.URL),
			zap.Int("port", info.Port))
		log.Info("ℹ️  MkDocs is already running")
		log.Infof("📚 Documentation: %s", info.URL)

		if !noBrowser {
			opened, err := dockerClient.OpenBrowserWithFallback(info.URL)
			if err != nil {
				logger.Warn("Failed to open browser", zap.Error(err))
				log.Info("")
				log.Infof("⚠️  Failed to open browser: %v", err)
				log.Infof("📖 Please open manually: %s", info.URL)
			} else if !opened {
				logger.Debug("Browser opening skipped (DinD mode or no display)")
				log.Infof("📖 Open in your browser: %s", info.URL)
			}
		}
		return 0
	}

	// Start container
	logger.Info("Starting MkDocs documentation server", zap.Int("port", port))
	log.Info("🚀 Starting MkDocs documentation server")

	// Write debug info about container configuration
	if debug {
		configJSON, _ := json.MarshalIndent(map[string]interface{}{
			"port":       port,
			"noBrowser":  noBrowser,
			"workspace":  workspaceRoot,
		}, "", "  ")
		writeDebugFile(workspaceRoot, logger, "container-config.json", string(configJSON))
	}

	info, err = dockerClient.StartContainer(port)
	if err != nil {
		if info != nil {
			logger.Warn("Container started with warnings", zap.Error(err), zap.String("url", info.URL))
			log.Infof("⚠️  %v", err)
			log.Infof("📖 Try accessing manually: %s", info.URL)
		} else {
			logger.Error("Failed to start container", zap.Error(err))
			log.Infof("❌ Failed to start container: %v", err)
			return 1
		}
	}

	// Display success
	logger.Info("MkDocs server started successfully",
		zap.String("url", info.URL),
		zap.Int("port", info.Port),
		zap.String("containerName", info.Name))
	log.Info("")
	log.Info("✅ MkDocs documentation server is running")
	log.Infof("📚 Documentation: %s", info.URL)

	// Open browser (skipped in DinD mode)
	if !noBrowser {
		opened, err := dockerClient.OpenBrowserWithFallback(info.URL)
		if err != nil {
			logger.Warn("Failed to open browser", zap.Error(err))
			log.Info("")
			log.Infof("⚠️  Failed to open browser: %v", err)
			log.Infof("📖 Please open manually: %s", info.URL)
		} else if !opened {
			logger.Debug("Browser opening skipped (DinD mode or no display)")
			log.Infof("📖 Open in your browser: %s", info.URL)
		} else {
			logger.Debug("Browser opened successfully")
		}
	}

	// Show tips
	if !debug {
		log.Info("")
		log.Info("💡 Tips:")
		log.Info("  • Container will keep running until stopped")
		log.Info("  • Stop with: go run . docs serve --stop")
		log.Infof("  • Or: docker stop %s", info.Name)
		log.Infof("  • View logs: docker logs %s", info.Name)
	}

	// Stream logs if debug mode
	if debug {
		logger.Info("Starting log streaming (debug mode)")
		log.Info("")
		log.Info("🔍 Debug mode: Streaming MkDocs logs (Press Ctrl+C to exit)")
		log.Info("─────────────────────────────────────────────────────────────")
		err = dockerClient.StreamLogs()
		if err != nil {
			logger.Error("Error streaming logs", zap.Error(err))
			log.Info("")
			log.Infof("❌ Error streaming logs: %v", err)
			return 1
		}
	}

	logger.Info("DocsServe command completed successfully")
	return 0
}

func handleDocsStop(workspaceRoot string, logger *logging.Logger) int {
	logger.Info("Stopping MkDocs documentation server")

	// Get Docker client
	dockerClient, err := docsInternal.NewClient(logger)
	if err != nil {
		logger.Error("Failed to initialize Docker client", zap.Error(err))
		log.Infof("❌ Failed to initialize: %v", err)
		return 1
	}
	defer dockerClient.Close()

	err = dockerClient.StopContainer()
	if err != nil {
		// Check if error is "container not found" - treat as success (idempotent operation)
		if strings.Contains(err.Error(), "no container found") {
			logger.Info("Container already stopped (not found)")
			log.Info("✅ MkDocs documentation server stopped")
			return 0
		}
		logger.Error("Failed to stop container", zap.Error(err))
		log.Infof("❌ Failed to stop container: %v", err)
		return 1
	}

	logger.Info("MkDocs server stopped successfully")
	log.Info("✅ MkDocs documentation server stopped")
	return 0
}
