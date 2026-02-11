package gource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type serveGourceCommand struct{}

var _ core.SimpleCommandPort = (*serveGourceCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&serveGourceCommand{},
	}
}

func (c *serveGourceCommand) Name() string { return "serve gource" }

func (c *serveGourceCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "serve-gource",
		Short:         "Visualize repository git history using Gource in a web browser",
		Long:          "Launches a Docker container with live Gource visualization.\nOpens browser to view animated git history with files and contributors.\nUse --output to render to a video file instead of streaming.",
		Flags: []core.FlagSpec{
			{Name: "no-browser", Type: "bool", DefaultValue: "false", Usage: "Don't open browser"},
			{Name: "port", Shorthand: "p", Type: "int", DefaultValue: "0", Usage: "Port number (auto 9000-9999)"},
			{Name: "stop", Type: "bool", DefaultValue: "false", Usage: "Stop the running server"},
			{Name: "title", Shorthand: "t", Type: "string", Usage: "Custom title"},
			{Name: "resolution", Shorthand: "r", Type: "string", DefaultValue: "960x540", Usage: "Video resolution"},
			{Name: "file-idle-time", Shorthand: "i", Type: "int", DefaultValue: "1", Usage: "Seconds files remain visible (0=forever)"},
			{Name: "output", Shorthand: "o", Type: "string", Usage: "Output video file path (renders to file instead of streaming)"},
			{Name: "format", Shorthand: "f", Type: "string", DefaultValue: "mp4", Usage: "Video format: mp4 or webm"},
			{Name: "duration", Shorthand: "d", Type: "int", DefaultValue: "60", Usage: "Target video duration in seconds"},
			{Name: "slow", Shorthand: "s", Type: "float64", DefaultValue: "1.0", Usage: "Time dilation multiplier (2.0 = 2x slower, doubles video length)"},
			{Name: "turbo", Type: "bool", DefaultValue: "false", Usage: "Use 80% of CPU and RAM for faster rendering"},
			{Name: "debug", Type: "bool", DefaultValue: "false", Usage: "Enable debug logging"},
		},
	}
}

func (c *serveGourceCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ServeGource()
}

var log = logging.C()
// ServeGource starts a Gource visualization server for the repository.
func ServeGource() int {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "serve", and "gource"

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	var noBrowser bool
	var port int
	var stop bool
	var title string
	var resolution string = "1920x1080"
	var fileIdleTime int = 1
	var output string
	var format string = "mp4"
	var duration int = 60
	var slow float64 = 1.0
	var turbo bool
	var debug bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--no-browser":
			noBrowser = true
		case "--stop":
			stop = true
		case "--turbo":
			turbo = true
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
		case "--title", "-t":
			if i+1 < len(args) {
				i++
				title = args[i]
			} else {
				log.Errorf("Error: --title requires a value")
				return 1
			}
		case "--resolution", "-r":
			if i+1 < len(args) {
				i++
				resolution = args[i]
			} else {
				log.Errorf("Error: --resolution requires a value")
				return 1
			}
		case "--file-idle-time", "-i":
			if i+1 < len(args) {
				i++
				t, err := strconv.Atoi(args[i])
				if err != nil {
					log.Errorf("Error: invalid file-idle-time: %s", args[i])
					return 1
				}
				fileIdleTime = t
			} else {
				log.Errorf("Error: --file-idle-time requires a value")
				return 1
			}
		case "--output", "-o":
			if i+1 < len(args) {
				i++
				output = args[i]
			} else {
				log.Errorf("Error: --output requires a value")
				return 1
			}
		case "--format", "-f":
			if i+1 < len(args) {
				i++
				format = strings.ToLower(args[i])
				if format != "mp4" && format != "webm" {
					log.Errorf("Error: --format must be 'mp4' or 'webm'")
					return 1
				}
			} else {
				log.Errorf("Error: --format requires a value")
				return 1
			}
		case "--duration", "-d":
			if i+1 < len(args) {
				i++
				d, err := strconv.Atoi(args[i])
				if err != nil || d <= 0 {
					log.Errorf("Error: --duration must be a positive integer (seconds)")
					return 1
				}
				duration = d
			} else {
				log.Errorf("Error: --duration requires a value")
				return 1
			}
		case "--slow", "-s":
			if i+1 < len(args) {
				i++
				s, err := strconv.ParseFloat(args[i], 64)
				if err != nil || s <= 0 {
					log.Errorf("Error: --slow must be a positive number")
					return 1
				}
				slow = s
			} else {
				log.Errorf("Error: --slow requires a value")
				return 1
			}
		case "--help", "-h":
			printUsage()
			return 0
		default:
			if strings.HasPrefix(arg, "-") {
				log.Errorf("Error: unknown flag: %s", arg)
				return 1
			}
		}
	}

	// Initialize logger
	if err := logging.ConfigureLoggingSimple(workspaceRoot, "commands", nil, debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	containerName := "cli-gource"

	// Handle --stop flag
	if stop {
		return handleStop(containerName)
	}

	ctx := context.Background()

	// Check if already running
	result, running, err := docker.IsServing(ctx, containerName)
	if err != nil {
		log.Errorf("Failed to check container status: %v", err)
		return 1
	}

	if running && result != nil {
		log.Info("Gource visualization is already running")
		log.Infof("URL: %s", result.URL)
		if !noBrowser {
			_, _ = docker.OpenBrowserWithFallback(result.URL)
		}
		return 0
	}

	// Set default title to repository name
	if title == "" {
		title = filepath.Base(workspaceRoot)
	}

	// File output mode - render to file instead of streaming
	if output != "" {
		return handleFileOutput(ctx, workspaceRoot, title, resolution, fileIdleTime, output, format, duration, slow, turbo)
	}

	// Build serve config
	serveConfig := &docker.ServeConfig{
		Name:  containerName,
		Image: "cli-gource:latest",
		BuildInfo: &docker.BuildInfo{
			Dockerfile:  filepath.Join(workspaceRoot, "containers/gource/Dockerfile"),
			ContextPath: filepath.Join(workspaceRoot, "containers/gource"),
		},
		ContentPath:   workspaceRoot,
		ContainerPath: "/visualization/repo",
		ContainerPort: 80,
		EnvVars: []string{
			fmt.Sprintf("GOURCE_TITLE=%s", title),
			fmt.Sprintf("GOURCE_RESOLUTION=%s", resolution),
			fmt.Sprintf("GOURCE_FILE_IDLE_TIME=%d", fileIdleTime),
			fmt.Sprintf("GOURCE_DURATION=%d", duration),
			fmt.Sprintf("GOURCE_SLOW=%g", slow),
		},
		RestartPolicy: "no", // Stop when visualization ends
		PreferredPort: port,
		Memory:        environments.GetContainerMemoryBytes(),
		CPUs:          float64(runtime.NumCPU()) / 2,
	}

	// Start container
	log.Info("Starting Gource visualization...")
	log.Infof("Repository: %s", workspaceRoot)
	log.Infof("Resolution: %s", resolution)

	result, err = docker.StartServe(ctx, serveConfig)
	if err != nil {
		log.Errorf("Failed to start container: %v", err)
		return 1
	}

	log.Info("")
	log.Info("Gource visualization is running")
	log.Infof("URL: %s", result.URL)

	if !noBrowser {
		_, _ = docker.OpenBrowserWithFallback(result.URL)
	}

	if !debug {
		log.Info("")
		log.Info("Stop with: eac serve gource --stop")
	}

	log.Debugf("Serve gource completed: url=%s", result.URL)
	return 0
}

// handleStop stops the running gource server.
func handleStop(containerName string) int {
	ctx := context.Background()

	if err := docker.StopServe(ctx, containerName); err != nil {
		if strings.Contains(err.Error(), "no container found") {
			log.Info("Gource visualization stopped")
			return 0
		}
		log.Errorf("Failed to stop container: %v", err)
		return 1
	}

	log.Info("Gource visualization stopped")
	return 0
}

// handleFileOutput renders the visualization to a video file.
func handleFileOutput(ctx context.Context, workspaceRoot, title, resolution string, fileIdleTime int, output, format string, duration int, slow float64, turbo bool) int {
	log.Info("Rendering Gource visualization to file...")
	log.Infof("Repository: %s", workspaceRoot)
	log.Infof("Resolution: %s", resolution)
	log.Infof("Format: %s", format)
	log.Infof("Target duration: %d seconds (x%.1f slow = %d seconds)", duration, slow, int(float64(duration)*slow))
	if turbo {
		log.Warn("Turbo mode: using 80% of system resources, might crash, dont use in CI")
	}

	// Resolve output path relative to workspace root
	var outputPath string
	if filepath.IsAbs(output) {
		outputPath = output
	} else {
		outputPath = filepath.Join(workspaceRoot, output)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Errorf("Error: failed to create output directory: %v", err)
		return 1
	}

	// Ensure output has correct extension
	ext := "." + format
	if !strings.HasSuffix(strings.ToLower(outputPath), ext) {
		outputPath += ext
	}

	// Check image staleness
	serveConfig := &docker.ServeConfig{
		Image: "cli-gource:latest",
		BuildInfo: &docker.BuildInfo{
			Dockerfile:  filepath.Join(workspaceRoot, "containers/gource/Dockerfile"),
			ContextPath: filepath.Join(workspaceRoot, "containers/gource"),
		},
	}

	stale, reason, err := docker.CheckImageStale(ctx, serveConfig)
	if err != nil {
		log.Warnf("Could not check image staleness: %v", err)
	}
	if stale {
		log.Infof("Rebuilding image: %s", reason)
	}

	// Build image - always rebuild for file output to ensure latest entrypoint
	log.Info("Building Docker image (no-cache)...")
	{
		dockerfilePath := filepath.Join(workspaceRoot, "containers/gource/Dockerfile")
		contextPath := filepath.Join(workspaceRoot, "containers/gource")
		buildCmd := fmt.Sprintf("docker build --no-cache -t cli-gource:latest -f %s %s", dockerfilePath, contextPath)
		log.Infof("Running: %s", buildCmd)

		// Use exec to run docker build with --no-cache to ensure fresh entrypoint
		cmd := exec.CommandContext(ctx, "docker", "build", "--no-cache", "-t", "cli-gource:latest", "-f", dockerfilePath, contextPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("Error: failed to build image: %v", err)
			return 1
		}
	}

	// Calculate resources - turbo mode uses 80%, normal uses 50%
	var memoryBytes int64
	var cpuCount float64
	if turbo {
		memoryBytes = int64(float64(environments.GetSystemMemoryBytes()) * 0.8)
		cpuCount = float64(runtime.NumCPU()) * 0.8
	} else {
		memoryBytes = environments.GetContainerMemoryBytes() // 50% of system
		cpuCount = float64(runtime.NumCPU()) / 2
	}

	// Run container using adapters/docker
	outputFilename := filepath.Base(outputPath)
	runConfig := &docker.RunConfig{
		Image: "cli-gource:latest",
		EnvVars: []string{
			fmt.Sprintf("GOURCE_TITLE=%s", title),
			fmt.Sprintf("GOURCE_RESOLUTION=%s", resolution),
			fmt.Sprintf("GOURCE_FILE_IDLE_TIME=%d", fileIdleTime),
			fmt.Sprintf("GOURCE_DURATION=%d", duration),
			fmt.Sprintf("GOURCE_SLOW=%g", slow),
			"GOURCE_OUTPUT_MODE=file",
			fmt.Sprintf("GOURCE_OUTPUT_FORMAT=%s", format),
			fmt.Sprintf("GOURCE_OUTPUT_FILENAME=%s", outputFilename),
		},
		Mounts: []docker.MountConfig{
			{Source: workspaceRoot, Target: "/visualization/repo", ReadOnly: true},
			{Source: outputDir, Target: "/visualization/output", ReadOnly: false},
		},
		WorkingDir:    "/visualization/repo",
		Memory:        memoryBytes,
		CPUs:          cpuCount,
		ContainerName: "cli-gource-render",
		StreamLogs:    true,
	}

	log.Info("Starting render container...")
	log.Info("Rendering in progress (this may take a while)...")

	result, err := docker.RunContainer(ctx, runConfig)
	if err != nil {
		log.Errorf("Error running container: %v", err)
		return 1
	}

	if result.ExitCode != 0 {
		log.Errorf("Container exited with code %d", result.ExitCode)
		return 1
	}

	log.Info("")
	log.Infof("Video saved to: %s", outputPath)
	return 0
}

func printUsage() {
	log.Info("Visualize repository git history using Gource in a web browser")
	log.Info("")
	log.Info("Launches a Docker container with live Gource visualization.")
	log.Info("Opens browser to view animated git history with files and contributors.")
	log.Info("Use --output to render to a video file instead of streaming.")
	log.Info("")
	log.Info("Usage:")
	log.Info("  eac serve gource [flags]")
	log.Info("")
	log.Info("Examples:")
	log.Info("  eac serve gource")
	log.Info("  eac serve gource --title \"My Project\"")
	log.Info("  eac serve gource --resolution 1920x1080")
	log.Info("  eac serve gource --port 8888")
	log.Info("  eac serve gource --stop")
	log.Info("")
	log.Info("  # Render to video file for GitHub Pages")
	log.Info("  eac serve gource --output docs/assets/history.mp4")
	log.Info("  eac serve gource --output video.webm --format webm")
	log.Info("")
	log.Info("Flags:")
	log.Info("  -t, --title string        Custom title for the visualization")
	log.Info("  -r, --resolution string   Video resolution (default: 1920x1080)")
	log.Info("  -i, --file-idle-time int  Seconds files stay visible (default: 1, 0=forever)")
	log.Info("  -d, --duration int        Target video duration in seconds (default: 60)")
	log.Info("  -s, --slow float          Time dilation multiplier (2.0 = 2x slower, doubles length)")
	log.Info("  -p, --port int            Port number (auto 9000-9999 if not specified)")
	log.Info("  -o, --output string       Output video file path (renders to file instead of streaming)")
	log.Info("  -f, --format string       Video format: mp4 or webm (default: mp4)")
	log.Info("      --turbo               Use 80% of CPU and RAM for faster rendering")
	log.Info("      --no-browser          Don't open browser after starting")
	log.Info("      --stop                Stop the running server")
	log.Info("      --debug               Enable debug logging")
	log.Info("")
	log.Info("Requirements:")
	log.Info("  - Docker must be running")
	log.Info("  - Repository must be a git repository")
}
