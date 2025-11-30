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

	docsInternal "github.com/ready-to-release/eac/src/commands/impl/docs/helper"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
	"go.uber.org/zap"
)

func init() {
	registry.Register(ServeDocs)
}

// printHelp displays help information for the serve docs command
func printHelp() {
	fmt.Println("NAME")
	fmt.Println("    serve docs - Start or stop MkDocs server")
	fmt.Println()
	fmt.Println("SYNOPSIS")
	fmt.Println("    eac serve docs [flags]")
	fmt.Println()
	fmt.Println("DESCRIPTION")
	fmt.Println("    The serve docs command manages the MkDocs documentation server using Docker.")
	fmt.Println("    It can start a server on a specified port, stop a running server, and open")
	fmt.Println("    the documentation in your browser.")
	fmt.Println()
	fmt.Println("FLAGS")
	fmt.Println("    --no-browser     Don't open browser after starting server")
	fmt.Println("    -p, --port       Port number for MkDocs server (default: auto-allocated 9000-9999)")
	fmt.Println("    --stop           Stop the running MkDocs server")
	fmt.Println("    --debug          Enable debug mode with log streaming")
	fmt.Println("    -h, --help       Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println("    eac serve docs                  # Start server with auto-allocated port")
	fmt.Println("    eac serve docs --port 9001      # Start server on specific port")
	fmt.Println("    eac serve docs --no-browser     # Start without opening browser")
	fmt.Println("    eac serve docs --stop           # Stop the running server")
	fmt.Println()
}

// writeDebugFile writes content to a debug file when debug mode is enabled.
// Files are written to out/logs/docs/<filename> in the workspace root.
func writeDebugFile(workspaceRoot string, logger *logging.Logger, filename string, content string) {
	if !logger.IsDebugMode() {
		return
	}

	debugDir := filepath.Join(workspaceRoot, "out", "logs", "docs")
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
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip "go", "run", ".", "docs", and "serve"

	var noBrowser bool
	var port int = 0 // 0 means auto-allocate from 9000-9999 range
	var stop bool
	var debug bool

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
		case "--port", "-p":
			if i+1 < len(args) {
				i++
				p, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid port number: %s\n", args[i])
					return 1
				}
				port = p
			} else {
				fmt.Fprintf(os.Stderr, "Error: --port requires a value\n")
				return 1
			}
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
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
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		return 1
	}
	defer logger.Sync()

	logger.Debug("DocsServe command started",
		zap.Bool("noBrowser", noBrowser),
		zap.Int("port", port),
		zap.Bool("stop", stop),
		zap.Bool("debug", debug))

	// Handle --stop flag
	if stop {
		return handleDocsStop(workspaceRoot, logger)
	}

	// Get Docker client
	dockerClient, err := docsInternal.NewClient(logger)
	if err != nil {
		logger.Error("Failed to initialize Docker client", zap.Error(err))
		fmt.Printf("❌ Failed to initialize: %v\n", err)
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
		fmt.Printf("❌ Failed to check container status: %v\n", err)
		return 1
	}

	if running && info != nil {
		// If user requested a specific port and it's different from running container, error
		if port != 0 && info.Port != port {
			logger.Error("MkDocs container already running on different port",
				zap.Int("runningPort", info.Port),
				zap.Int("requestedPort", port))
			fmt.Printf("❌ MkDocs is already running on port %d\n", info.Port)
			fmt.Printf("📚 Running at: %s\n", info.URL)
			fmt.Printf("\n💡 To use a different port:\n")
			fmt.Printf("  1. Stop the running container: go run . docs serve --stop\n")
			fmt.Printf("  2. Start with new port: go run . docs serve --port %d\n", port)
			return 1
		}

		// Container is running on the expected port (or port was auto-allocated)
		logger.Info("MkDocs container already running",
			zap.String("url", info.URL),
			zap.Int("port", info.Port))
		fmt.Printf("ℹ️  MkDocs is already running\n")
		fmt.Printf("📚 Documentation: %s\n", info.URL)

		if !noBrowser {
			opened, err := dockerClient.OpenBrowserWithFallback(info.URL)
			if err != nil {
				logger.Warn("Failed to open browser", zap.Error(err))
				fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
				fmt.Printf("📖 Please open manually: %s\n", info.URL)
			} else if !opened {
				logger.Debug("Browser opening skipped (DinD mode or no display)")
				fmt.Printf("📖 Open in your browser: %s\n", info.URL)
			}
		}
		return 0
	}

	// Start container
	logger.Info("Starting MkDocs documentation server", zap.Int("port", port))
	fmt.Printf("🚀 Starting MkDocs documentation server\n")

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
			fmt.Printf("⚠️  %v\n", err)
			fmt.Printf("📖 Try accessing manually: %s\n", info.URL)
		} else {
			logger.Error("Failed to start container", zap.Error(err))
			fmt.Printf("❌ Failed to start container: %v\n", err)
			return 1
		}
	}

	// Display success
	logger.Info("MkDocs server started successfully",
		zap.String("url", info.URL),
		zap.Int("port", info.Port),
		zap.String("containerName", info.Name))
	fmt.Printf("\n✅ MkDocs documentation server is running\n")
	fmt.Printf("📚 Documentation: %s\n", info.URL)

	// Open browser (skipped in DinD mode)
	if !noBrowser {
		opened, err := dockerClient.OpenBrowserWithFallback(info.URL)
		if err != nil {
			logger.Warn("Failed to open browser", zap.Error(err))
			fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
			fmt.Printf("📖 Please open manually: %s\n", info.URL)
		} else if !opened {
			logger.Debug("Browser opening skipped (DinD mode or no display)")
			fmt.Printf("📖 Open in your browser: %s\n", info.URL)
		} else {
			logger.Debug("Browser opened successfully")
		}
	}

	// Show tips
	if !debug {
		fmt.Println("\n💡 Tips:")
		fmt.Println("  • Container will keep running until stopped")
		fmt.Printf("  • Stop with: go run . docs serve --stop\n")
		fmt.Printf("  • Or: docker stop %s\n", info.Name)
		fmt.Printf("  • View logs: docker logs %s\n", info.Name)
	}

	// Stream logs if debug mode
	if debug {
		logger.Info("Starting log streaming (debug mode)")
		fmt.Println("\n🔍 Debug mode: Streaming MkDocs logs (Press Ctrl+C to exit)")
		fmt.Println("─────────────────────────────────────────────────────────────")
		err = dockerClient.StreamLogs()
		if err != nil {
			logger.Error("Error streaming logs", zap.Error(err))
			fmt.Printf("\n❌ Error streaming logs: %v\n", err)
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
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer dockerClient.Close()

	err = dockerClient.StopContainer()
	if err != nil {
		// Check if error is "container not found" - treat as success (idempotent operation)
		if strings.Contains(err.Error(), "no container found") {
			logger.Info("Container already stopped (not found)")
			fmt.Printf("✅ MkDocs documentation server stopped\n")
			return 0
		}
		logger.Error("Failed to stop container", zap.Error(err))
		fmt.Printf("❌ Failed to stop container: %v\n", err)
		return 1
	}

	logger.Info("MkDocs server stopped successfully")
	fmt.Printf("✅ MkDocs documentation server stopped\n")
	return 0
}
