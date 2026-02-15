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

// gourceOptions holds all parsed command-line options for the gource command.
type gourceOptions struct {
	noBrowser    bool
	port         int
	stop         bool
	title        string
	resolution   string
	fileIdleTime int
	output       string
	format       string
	duration     int
	slow         float64
	turbo        bool
	debug        bool
}

// newGourceOptions returns gourceOptions with default values applied.
func newGourceOptions() *gourceOptions {
	return &gourceOptions{
		resolution:   "1920x1080",
		fileIdleTime: 1,
		format:       "mp4",
		duration:     60,
		slow:         1.0,
	}
}

// parseGourceArgs parses command-line arguments into gourceOptions.
// Returns nil and an exit code if parsing fails or --help was requested.
func parseGourceArgs(args []string) (*gourceOptions, int) {
	opts := newGourceOptions()

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--no-browser":
			opts.noBrowser = true
		case "--stop":
			opts.stop = true
		case "--turbo":
			opts.turbo = true
		case "--debug":
			opts.debug = true
		case "--port", "-p":
			if i+1 < len(args) {
				i++
				p, err := strconv.Atoi(args[i])
				if err != nil {
					log.Errorf("Error: invalid port number: %s", args[i])
					return nil, 1
				}
				opts.port = p
			} else {
				log.Errorf("Error: --port requires a value")
				return nil, 1
			}
		case "--title", "-t":
			if i+1 < len(args) {
				i++
				opts.title = args[i]
			} else {
				log.Errorf("Error: --title requires a value")
				return nil, 1
			}
		case "--resolution", "-r":
			if i+1 < len(args) {
				i++
				opts.resolution = args[i]
			} else {
				log.Errorf("Error: --resolution requires a value")
				return nil, 1
			}
		case "--file-idle-time", "-i":
			if i+1 < len(args) {
				i++
				t, err := strconv.Atoi(args[i])
				if err != nil {
					log.Errorf("Error: invalid file-idle-time: %s", args[i])
					return nil, 1
				}
				opts.fileIdleTime = t
			} else {
				log.Errorf("Error: --file-idle-time requires a value")
				return nil, 1
			}
		case "--output", "-o":
			if i+1 < len(args) {
				i++
				opts.output = args[i]
			} else {
				log.Errorf("Error: --output requires a value")
				return nil, 1
			}
		case "--format", "-f":
			if i+1 < len(args) {
				i++
				opts.format = strings.ToLower(args[i])
				if opts.format != "mp4" && opts.format != "webm" {
					log.Errorf("Error: --format must be 'mp4' or 'webm'")
					return nil, 1
				}
			} else {
				log.Errorf("Error: --format requires a value")
				return nil, 1
			}
		case "--duration", "-d":
			if i+1 < len(args) {
				i++
				d, err := strconv.Atoi(args[i])
				if err != nil || d <= 0 {
					log.Errorf("Error: --duration must be a positive integer (seconds)")
					return nil, 1
				}
				opts.duration = d
			} else {
				log.Errorf("Error: --duration requires a value")
				return nil, 1
			}
		case "--slow", "-s":
			if i+1 < len(args) {
				i++
				s, err := strconv.ParseFloat(args[i], 64)
				if err != nil || s <= 0 {
					log.Errorf("Error: --slow must be a positive number")
					return nil, 1
				}
				opts.slow = s
			} else {
				log.Errorf("Error: --slow requires a value")
				return nil, 1
			}
		case "--help", "-h":
			printUsage()
			return nil, 0
		default:
			if strings.HasPrefix(arg, "-") {
				log.Errorf("Error: unknown flag: %s", arg)
				return nil, 1
			}
		}
	}

	return opts, 0
}

// ServeGource starts a Gource visualization server for the repository.
func ServeGource() int {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "serve", and "gource"
	opts, exitCode := parseGourceArgs(args)
	if opts == nil {
		return exitCode
	}

	// Initialize logger
	if err := logging.ConfigureLoggingSimple(workspaceRoot, "commands", nil, opts.debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	containerName := "cli-gource"

	// Handle --stop flag
	if opts.stop {
		return handleStop(containerName)
	}

	ctx := context.Background()

	// Check if already running
	if code, handled := handleAlreadyRunning(ctx, containerName, opts.noBrowser); handled {
		return code
	}

	// Set default title to repository name
	if opts.title == "" {
		opts.title = filepath.Base(workspaceRoot)
	}

	// File output mode - render to file instead of streaming
	if opts.output != "" {
		return handleFileOutput(ctx, workspaceRoot, opts)
	}

	return startServeContainer(ctx, workspaceRoot, containerName, opts)
}

// handleAlreadyRunning checks whether the gource container is already running.
// Returns the exit code and true if the situation was handled (running or error),
// or 0 and false if the caller should proceed with starting a new container.
func handleAlreadyRunning(ctx context.Context, containerName string, noBrowser bool) (int, bool) {
	result, running, err := docker.IsServing(ctx, containerName)
	if err != nil {
		log.Errorf("Failed to check container status: %v", err)
		return 1, true
	}

	if running && result != nil {
		log.Info("Gource visualization is already running")
		log.Infof("URL: %s", result.URL)
		if !noBrowser {
			_, _ = docker.OpenBrowserWithFallback(result.URL)
		}
		return 0, true
	}

	return 0, false
}

// buildGourceEnvVars constructs the environment variable list for a gource container.
func buildGourceEnvVars(opts *gourceOptions) []string {
	return []string{
		fmt.Sprintf("GOURCE_TITLE=%s", opts.title),
		fmt.Sprintf("GOURCE_RESOLUTION=%s", opts.resolution),
		fmt.Sprintf("GOURCE_FILE_IDLE_TIME=%d", opts.fileIdleTime),
		fmt.Sprintf("GOURCE_DURATION=%d", opts.duration),
		fmt.Sprintf("GOURCE_SLOW=%g", opts.slow),
	}
}

// startServeContainer builds and starts the gource streaming container.
func startServeContainer(ctx context.Context, workspaceRoot, containerName string, opts *gourceOptions) int {
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
		EnvVars:       buildGourceEnvVars(opts),
		RestartPolicy: "no", // Stop when visualization ends
		PreferredPort: opts.port,
		Memory:        environments.GetContainerMemoryBytes(),
		CPUs:          float64(runtime.NumCPU()) / 2,
	}

	log.Info("Starting Gource visualization...")
	log.Infof("Repository: %s", workspaceRoot)
	log.Infof("Resolution: %s", opts.resolution)

	result, err := docker.StartServe(ctx, serveConfig)
	if err != nil {
		log.Errorf("Failed to start container: %v", err)
		return 1
	}

	log.Info("")
	log.Info("Gource visualization is running")
	log.Infof("URL: %s", result.URL)

	if !opts.noBrowser {
		_, _ = docker.OpenBrowserWithFallback(result.URL)
	}

	if !opts.debug {
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
func handleFileOutput(ctx context.Context, workspaceRoot string, opts *gourceOptions) int {
	log.Info("Rendering Gource visualization to file...")
	log.Infof("Repository: %s", workspaceRoot)
	log.Infof("Resolution: %s", opts.resolution)
	log.Infof("Format: %s", opts.format)
	log.Infof("Target duration: %d seconds (x%.1f slow = %d seconds)", opts.duration, opts.slow, int(float64(opts.duration)*opts.slow))
	if opts.turbo {
		log.Warn("Turbo mode: using 80% of system resources, might crash, dont use in CI")
	}

	outputPath, outputDir, err := resolveOutputPath(workspaceRoot, opts.output, opts.format)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	if err := buildGourceImage(ctx, workspaceRoot); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	return runRenderContainer(ctx, workspaceRoot, outputPath, outputDir, opts)
}

// resolveOutputPath resolves the output file path and ensures the directory exists.
// Returns the resolved output path, the output directory, and any error.
func resolveOutputPath(workspaceRoot, output, format string) (string, string, error) {
	var outputPath string
	if filepath.IsAbs(output) {
		outputPath = output
	} else {
		outputPath = filepath.Join(workspaceRoot, output)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Ensure output has correct extension
	ext := "." + format
	if !strings.HasSuffix(strings.ToLower(outputPath), ext) {
		outputPath += ext
	}

	return outputPath, outputDir, nil
}

// buildGourceImage checks image staleness and rebuilds the Docker image with no-cache.
func buildGourceImage(ctx context.Context, workspaceRoot string) error {
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

	log.Info("Building Docker image (no-cache)...")
	dockerfilePath := filepath.Join(workspaceRoot, "containers/gource/Dockerfile")
	contextPath := filepath.Join(workspaceRoot, "containers/gource")
	buildCmd := fmt.Sprintf("docker build --no-cache -t cli-gource:latest -f %s %s", dockerfilePath, contextPath)
	log.Infof("Running: %s", buildCmd)

	cmd := exec.CommandContext(ctx, "docker", "build", "--no-cache", "-t", "cli-gource:latest", "-f", dockerfilePath, contextPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %v", err)
	}

	return nil
}

// calculateRenderResources returns the memory and CPU allocation for the render container.
// Turbo mode uses 80% of system resources; normal mode uses 50%.
func calculateRenderResources(turbo bool) (memoryBytes int64, cpuCount float64) {
	if turbo {
		memoryBytes = int64(float64(environments.GetSystemMemoryBytes()) * 0.8)
		cpuCount = float64(runtime.NumCPU()) * 0.8
	} else {
		memoryBytes = environments.GetContainerMemoryBytes() // 50% of system
		cpuCount = float64(runtime.NumCPU()) / 2
	}
	return memoryBytes, cpuCount
}

// buildFileOutputEnvVars constructs environment variables for the file-output render container.
func buildFileOutputEnvVars(opts *gourceOptions, outputFilename string) []string {
	envVars := buildGourceEnvVars(opts)
	envVars = append(envVars,
		"GOURCE_OUTPUT_MODE=file",
		fmt.Sprintf("GOURCE_OUTPUT_FORMAT=%s", opts.format),
		fmt.Sprintf("GOURCE_OUTPUT_FILENAME=%s", outputFilename),
	)
	return envVars
}

// runRenderContainer configures and runs the Docker container for file rendering.
func runRenderContainer(ctx context.Context, workspaceRoot, outputPath, outputDir string, opts *gourceOptions) int {
	memoryBytes, cpuCount := calculateRenderResources(opts.turbo)
	outputFilename := filepath.Base(outputPath)

	runConfig := &docker.RunConfig{
		Image:   "cli-gource:latest",
		EnvVars: buildFileOutputEnvVars(opts, outputFilename),
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
