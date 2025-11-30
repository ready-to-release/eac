package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/src/cli/internal/cache"
	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/github"
	"github.com/ready-to-release/eac/src/cli/internal/terminal"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ContainerMode defines how the container should be configured
type ContainerMode int

const (
	ModeRun ContainerMode = iota
	ModeInteractive
)

// ExtensionConfig holds the configuration for an extension
type ExtensionConfig struct {
	Name               string
	Image              string
	ImagePullPolicy    string
	LoadLocal          bool
	AutoRemoveChildren bool
	Env                []conf.EnvVar
}

// ContainerHost manages Docker container operations for extensions
type ContainerHost struct {
	client  DockerClient
	ctx     context.Context
	rootDir string
}

// NewContainerHost creates a new ContainerHost instance
func NewContainerHost() (*ContainerHost, error) {
	ctx := context.Background()

	// Configure Docker client options
	clientOpts := []client.Opt{client.FromEnv}

	// Force Docker API version negotiation for compatibility
	// This prevents "client version X is too new" errors
	clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())

	// Override Docker host if R2R_DOCKER_HOST is set
	if dockerHost := os.Getenv("R2R_DOCKER_HOST"); dockerHost != "" {
		clientOpts = append(clientOpts, client.WithHost(dockerHost))
		log.Debug().Str("docker_host", dockerHost).Msg("Using custom Docker host from R2R_DOCKER_HOST")
	}

	cli, err := NewRealDockerClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}

	// Verify Docker daemon is accessible
	_, pingErr := cli.Ping(ctx)
	if pingErr != nil {
		cli.Close()
		// Check for common Docker service not running errors
		errStr := pingErr.Error()
		if strings.Contains(errStr, "docker_engine") ||
		   strings.Contains(errStr, "cannot connect to the Docker daemon") ||
		   strings.Contains(errStr, "Is the docker daemon running") ||
		   strings.Contains(errStr, "system cannot find the file specified") {
			return nil, fmt.Errorf("Docker service is not running. Please start Docker Desktop or the Docker daemon and try again")
		}
		return nil, fmt.Errorf("cannot connect to Docker daemon: %w", pingErr)
	}

	rootDir, err := conf.FindRepositoryRoot()
	if err != nil {
		return nil, fmt.Errorf("error finding root directory: %w", err)
	}

	return &ContainerHost{
		client:  cli,
		ctx:     ctx,
		rootDir: rootDir,
	}, nil
}

// ValidateExtensions checks if extensions are configured
func (ch *ContainerHost) ValidateExtensions() error {
	if len(conf.Global.Extensions) == 0 {
		return fmt.Errorf("config file does not contain any extensions. Please run 'r2r init' to initialize the configuration")
	}
	return nil
}

// FindExtension locates an extension by name in the configuration
func (ch *ContainerHost) FindExtension(name string) (*ExtensionConfig, error) {
	for _, ext := range conf.Global.Extensions {
		if ext.Name == name {
			// Apply default ImagePullPolicy if not specified
			imagePullPolicy := ext.ImagePullPolicy
			if imagePullPolicy == "" {
				// Default to AutoDetect
				imagePullPolicy = "AutoDetect"
			}

			// Note: Version extraction from image tag is not currently used
			// but kept for potential future metadata operations

			config := &ExtensionConfig{
				Name:               ext.Name,
				Image:              ext.Image,
				ImagePullPolicy:    imagePullPolicy,
				LoadLocal:          ext.LoadLocal || conf.Global.LoadLocal, // Use extension-level or global LoadLocal flag
				AutoRemoveChildren: ext.AutoRemoveChildren,
				Env:                ext.Env,
			}


			return config, nil
		}
	}
	return nil, fmt.Errorf("extension '%s' not found in config", name)
}

// BuildEnvironmentVars creates the environment variable list for a container
func (ch *ContainerHost) BuildEnvironmentVars(ext *ExtensionConfig) []string {
	envVars := []string{
		"R2R_CONTAINER_REPOROOT=" + "/var/task",
		"R2R_HOST_REPOROOT=" + ch.rootDir,
	}

	// Add terminal dimensions
	// First try to get from environment (if already set)
	cols := os.Getenv("COLUMNS")
	lines := os.Getenv("LINES")

	// Always try to detect terminal size for better accuracy
	if width, height, err := terminal.GetSize(); err == nil && width > 0 && height > 0 {
		// Successfully detected terminal size
		cols = strconv.Itoa(width)
		lines = strconv.Itoa(height)
		log.Debug().Int("detected_width", width).Int("detected_height", height).Msg("Terminal size detected")
		envVars = append(envVars, "COLUMNS="+cols, "LINES="+lines, "R2R_TERMINAL_DETECTION=auto")
	} else {
		// Failed to detect, use environment or defaults
		if cols == "" {
			cols = "80"
		}
		if lines == "" {
			lines = "24"
		}
		log.Debug().Str("cols", cols).Str("lines", lines).Msg("Using default terminal size")
		envVars = append(envVars, "COLUMNS="+cols, "LINES="+lines, "R2R_TERMINAL_DETECTION=default")
	}

	// 1. CI Environment Detection & Defaults
	if ch.detectCIEnvironment() {
		envVars = append(envVars, ch.getCIDefaults()...)
	} else {
		// 2. Inherit Current Shell Settings (when not in CI)
		envVars = append(envVars, ch.getShellColorSettings()...)
	}

	// 3. Add global environment variables from config
	if conf.Global.Environment != nil {
		for _, env := range conf.Global.Environment.Global {
			envVars = append(envVars, env.Name+"="+env.Value)
		}

		// Add secrets from config (these get values from host environment)
		for _, secret := range conf.Global.Environment.Secrets {
			if value := os.Getenv(secret.Env); value != "" {
				envVars = append(envVars, secret.Name+"="+value)
			}
		}
	}

	// 4. Always ensure GITHUB_USERNAME and GITHUB_TOKEN are available if set in host environment
	// This is critical for extensions that need to access GitHub Container Registry
	if githubUsername := os.Getenv("GITHUB_USERNAME"); githubUsername != "" {
		envVars = append(envVars, "GITHUB_USERNAME="+githubUsername)
	}
	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		envVars = append(envVars, "GITHUB_TOKEN="+githubToken)
	}

	// 5. Add extension-specific env vars (these can override defaults)
	for _, env := range ext.Env {
		if env.Value != "" {
			// Static value
			envVars = append(envVars, env.Name+"="+env.Value)
		} else {
			// Passthrough from host (value omitted)
			if hostValue := os.Getenv(env.Name); hostValue != "" {
				envVars = append(envVars, env.Name+"="+hostValue)
				log.Debug().Str("env", env.Name).Msg("Passing through environment variable from host")
			} else if env.Required {
				log.Error().Str("env", env.Name).Msg("Required environment variable not set on host")
			}
		}
	}

	return envVars
}

// detectCIEnvironment checks multiple CI indicators beyond just CI=true
func (ch *ContainerHost) detectCIEnvironment() bool {
	ciIndicators := []string{
		"CI", "CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS", "AZUREDEVOPS_URL", "GITLAB_CI",
		"AZURE_HTTP_USER_AGENT", "TF_BUILD", "BUILDKITE",
		"CIRCLECI", "TRAVIS", "DRONE", "SEMAPHORE",
		"APPVEYOR", "CODEBUILD_BUILD_ID", "TEAMCITY_VERSION",
	}

	for _, indicator := range ciIndicators {
		if value := os.Getenv(indicator); value != "" && value != "false" && value != "0" {
			return true
		}
	}
	return false
}

// getCIDefaults returns CI-appropriate environment settings
func (ch *ContainerHost) getCIDefaults() []string {
	return []string{
		"NO_COLOR=1",    // Disable colors in CI
		"TERM=dumb",     // Simple terminal for CI
		"FORCE_COLOR=0", // Force disable color
		"CI=true",       // Indicate CI environment
	}
}

// getShellColorSettings inherits current shell color capabilities
func (ch *ContainerHost) getShellColorSettings() []string {
	envVars := []string{}

	// Inherit color support from current shell
	colorEnvVars := []string{
		"TERM", "COLORTERM", "CLICOLOR", "CLICOLOR_FORCE",
		"NO_COLOR", "FORCE_COLOR", "COLOR",
	}

	for _, envVar := range colorEnvVars {
		if value := os.Getenv(envVar); value != "" {
			envVars = append(envVars, envVar+"="+value)
		}
	}

	// Apply sensible defaults if no color settings detected
	if len(envVars) == 0 {
		envVars = append(envVars, ch.getDefaultColorSettings()...)
	}

	return envVars
}

// getDefaultColorSettings provides fallback color settings for environments without explicit settings
func (ch *ContainerHost) getDefaultColorSettings() []string {
	// Detect terminal capabilities
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color" // Sensible default for modern terminals
	}

	return []string{
		"TERM=" + term,
		"COLORTERM=truecolor", // Modern terminal default supporting full color
		// Don't set NO_COLOR or FORCE_COLOR - let programs decide based on their logic
	}
}

// InspectImage inspects a Docker image and returns the inspection result
func (ch *ContainerHost) InspectImage(image string) (*image.InspectResponse, error) {
	imageInspect, err := ch.client.ImageInspect(ch.ctx, image)
	if err != nil {
		return nil, fmt.Errorf("error inspecting image: %w", err)
	}
	return &imageInspect, nil
}

// GetImageDigest returns the digest of a Docker image for cache invalidation
// Returns the image ID if no repo digests are available (local images)
func (ch *ContainerHost) GetImageDigest(imageName string) (string, error) {
	inspect, err := ch.InspectImage(imageName)
	if err != nil {
		return "", err
	}

	// Prefer RepoDigests (sha256:...) for pulled images
	if len(inspect.RepoDigests) > 0 {
		return inspect.RepoDigests[0], nil
	}

	// Fall back to image ID for local images
	return inspect.ID, nil
}

// CreateContainerConfig creates a container configuration based on mode and extension
func (ch *ContainerHost) CreateContainerConfig(ext *ExtensionConfig, mode ContainerMode, args []string, imageInspect *image.InspectResponse) *container.Config {
	envVars := ch.BuildEnvironmentVars(ext)

	config := &container.Config{
		Image: ext.Image,
		Env:   envVars,
	}

	switch mode {
	case ModeInteractive:
		config.Tty = true
		config.OpenStdin = true
		// If no entrypoint, use shell; if entrypoint exists, let it run
		if len(imageInspect.Config.Entrypoint) == 0 {
			config.Cmd = []string{"/bin/sh"}
		}
	case ModeRun:
		// Only enable TTY for truly interactive sessions
		// When running commands, we don't want TTY to avoid cursor position queries
		if len(args) == 0 {
			// No args means interactive mode
			log.Debug().Msg("ModeRun: No args detected, enabling TTY for interactive session")
			config.Tty = true
			config.OpenStdin = true
		} else {
			// Args present means command mode - enable TTY for proper terminal width detection
			// ANSI escape sequences will be filtered out by the CLI
			log.Debug().Int("args_count", len(args)).Strs("args", args).Msg("ModeRun: Args present, enabling TTY for terminal width detection")
			config.Tty = true
			config.OpenStdin = false  // Disable stdin for command mode to avoid TTY corruption
		}
		config.Cmd = args
	}

	// Only set WorkingDir if container does NOT have an entrypoint defined
	if len(imageInspect.Config.Entrypoint) == 0 {
		workdir := "/var/task"
		log.Debug().Str("workdir", workdir).Msg("No entrypoint found in extension container, setting workingdir")
		config.WorkingDir = workdir
	} else {
		log.Debug().Msg("Found entrypoint in extension container, not setting workingdir")
	}

	return config
}

// CreateHostConfig creates the host configuration with volume mounts
// volumeRequests are optional cache volumes requested by the extension via metadata
func (ch *ContainerHost) CreateHostConfig(ext *ExtensionConfig, volumeRequests []cache.VolumeRequest) *container.HostConfig {
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: ch.rootDir,
			Target: "/var/task",
		},
	}

	// Add Docker service mount based on platform
	dockerMount := ch.getDockerServiceMount()
	if dockerMount != nil {
		mounts = append(mounts, *dockerMount)
	}

	// Add cache volumes requested by the extension
	for _, vol := range volumeRequests {
		if vol.Type == "cache" {
			// Named volume: {extension-name}-{volume-name}
			volumeName := fmt.Sprintf("%s-%s", ext.Name, vol.Name)
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeVolume,
				Source: volumeName,
				Target: vol.Target,
			})
			log.Debug().
				Str("extension", ext.Name).
				Str("volume", volumeName).
				Str("target", vol.Target).
				Msg("Adding cache volume mount")
		}
		// Future: handle "bind" type for bind mounts if needed
	}

	return &container.HostConfig{
		AutoRemove: true,
		Mounts:     mounts,
	}
}

// getDockerServiceMount returns the appropriate Docker service mount for the current platform
func (ch *ContainerHost) getDockerServiceMount() *mount.Mount {
	// For all platforms (including WSL2/Windows), use the Unix socket path
	// Docker Desktop on Windows exposes the socket at this path in WSL2
	return &mount.Mount{
		Type:   mount.TypeBind,
		Source: "/var/run/docker.sock",
		Target: "/var/run/docker.sock",
	}
}

// CreateContainer creates a new Docker container with the specified configuration
func (ch *ContainerHost) CreateContainer(containerConfig *container.Config, hostConfig *container.HostConfig) (string, error) {
	resp, err := ch.client.ContainerCreate(ch.ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("error creating container: %w", err)
	}

	// TTY resize will be done after container starts (in StartContainer)

	return resp.ID, nil
}

// StartContainer starts a Docker container by ID
func (ch *ContainerHost) StartContainer(containerID string) error {
	if err := ch.client.ContainerStart(ch.ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("error starting container: %w", err)
	}

	// After starting, resize the TTY if needed
	// Check if container has TTY enabled
	inspect, err := ch.client.ContainerInspect(ch.ctx, containerID)
	if err == nil && inspect.Config.Tty {
		if width, height, err := terminal.GetSize(); err == nil && width > 0 && height > 0 {
			log.Debug().Int("terminal_width", width).Int("terminal_height", height).Msg("Resizing container TTY after start")
			resizeOptions := container.ResizeOptions{
				Height: uint(height),
				Width:  uint(width),
			}
			if err := ch.client.ContainerResize(ch.ctx, containerID, resizeOptions); err != nil {
				log.Debug().Err(err).Msg("Failed to resize container TTY after start")
			} else {
				log.Debug().Msg("Successfully resized container TTY after start")
			}
		}
	}

	return nil
}

// AttachToContainer attaches to a container for I/O operations
func (ch *ContainerHost) AttachToContainer(containerID string) (types.HijackedResponse, error) {
	// Inspect container to determine if stdin should be attached
	inspect, err := ch.client.ContainerInspect(ch.ctx, containerID)
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("error inspecting container: %w", err)
	}

	// Only attach stdin if the container has OpenStdin enabled
	attachStdin := inspect.Config.OpenStdin

	log.Debug().
		Bool("attach_stdin", attachStdin).
		Bool("container_open_stdin", inspect.Config.OpenStdin).
		Str("container_id", containerID).
		Msg("Attaching to container with appropriate stdin setting")

	attachResp, err := ch.client.ContainerAttach(ch.ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  attachStdin,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("error attaching to container: %w", err)
	}
	return attachResp, nil
}

// WaitForContainer waits for a container to finish execution
func (ch *ContainerHost) WaitForContainer(containerID string) (<-chan container.WaitResponse, <-chan error) {
	return ch.client.ContainerWait(ch.ctx, containerID, container.WaitConditionNotRunning)
}

// StopContainer stops a running container
func (ch *ContainerHost) StopContainer(containerID string) error {
	return ch.client.ContainerStop(ch.ctx, containerID, container.StopOptions{})
}

// StopContainerWithContext stops a running container with a specific context for timeout control
func (ch *ContainerHost) StopContainerWithContext(ctx context.Context, containerID string) error {
	// Docker will send SIGTERM first, then SIGKILL after the timeout
	// The default timeout is 10 seconds, but we're using the context to control it
	return ch.client.ContainerStop(ctx, containerID, container.StopOptions{})
}

// GetRootDir returns the root directory path
func (ch *ContainerHost) GetRootDir() string {
	return ch.rootDir
}

// CreateGitHubAuthConfig creates authentication configuration for GitHub Container Registry
// Returns both the registry.AuthConfig and base64-encoded auth string for Docker API calls
func CreateGitHubAuthConfig() (*registry.AuthConfig, string, error) {
	// Try multiple authentication sources in order of preference

	// 1. First try environment variables (highest priority for CI/CD)
	username := os.Getenv("GITHUB_USERNAME")
	password := os.Getenv("GITHUB_TOKEN")

	// 2. If no token in env, try GitHub CLI authentication
	if password == "" {
		log.Debug().Msg("No GITHUB_TOKEN found, trying GitHub CLI authentication")
		auth, err := github.GetCLIAuth()
		if err == nil && auth.Token != "" {
			password = auth.Token
			if username == "" && auth.Username != "" {
				username = auth.Username
			}
			log.Info().Msg("Using GitHub CLI authentication for ghcr.io")
		} else if err != nil {
			log.Debug().Err(err).Msg("GitHub CLI authentication not available")
		}
	}

	// If username is not set but we have a token, use a default username for GitHub Container Registry
	if username == "" && password != "" {
		username = "github-actions" // Generic username that works with personal access tokens and GITHUB_TOKEN
	}

	if password == "" {
		return nil, "", fmt.Errorf("authentication required: GITHUB_TOKEN environment variable must be set or GitHub CLI must be authenticated (run 'gh auth login')")
	}

	// Create authentication config
	authConfig := &registry.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: "ghcr.io",
	}

	// Encode authentication for Docker API
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return nil, "", fmt.Errorf("error encoding auth config: %w", err)
	}
	authStr := base64.StdEncoding.EncodeToString(encodedJSON)

	return authConfig, authStr, nil
}

// EnsureImageExists checks if an image exists locally and pulls it based on the pull policy
func (ch *ContainerHost) EnsureImageExists(imageName string, pullPolicy string, loadLocal bool) error {
	// Apply default if not specified
	if pullPolicy == "" {
		pullPolicy = "AutoDetect"
	}

	// Handle "AutoDetect" policy - choose based on image tag and local availability
	if pullPolicy == "AutoDetect" {
		// First check if image exists locally
		localImageInfo, err := ch.client.ImageInspect(ch.ctx, imageName)
		hasLocalImage := err == nil

		if hasLocalImage {
			log.Debug().
				Str("image", imageName).
				Int("repoDigests", len(localImageInfo.RepoDigests)).
				Str("id", localImageInfo.ID).
				Msg("Local image found")
		}

		// Extract tag from image name (format: registry/repo:tag)
		tagIndex := strings.LastIndex(imageName, ":")
		tag := ""
		if tagIndex > 0 && tagIndex < len(imageName)-1 {
			tag = imageName[tagIndex+1:]
		}

		// For development: When loadLocal is true, always prefer local image
		if hasLocalImage && loadLocal {
			// loadLocal explicitly requests using the local image (e.g., for development builds)
			fmt.Printf("🏠 Using local development image: %s\n", imageName)
			log.Info().
				Str("image", imageName).
				Bool("hasRepoDigests", len(localImageInfo.RepoDigests) > 0).
				Msg("Using local development image (loadLocal: true)")

			// Cache the local digest for future checks
			ch.cacheImageDigest(imageName, tag, localImageInfo)
			return nil
		} else if hasLocalImage && (tag == "latest" || tag == "main" || tag == "master") {
			// For dynamic tags with local image (not loadLocal mode), check cache TTL
			// This prevents hitting the registry on every command
			registryCache, _ := cache.LoadRegistryCache(ch.rootDir)
			cacheTTL := 300 // default 5 minutes
			if conf.Global.Registry != nil && conf.Global.Registry.CacheTTL > 0 {
				cacheTTL = conf.Global.Registry.CacheTTL
			}

			if registryCache != nil && !registryCache.IsExpired(cacheTTL) {
				// Cache is still valid, use local image
				log.Info().
					Str("image", imageName).
					Int("cacheTTL", cacheTTL).
					Msg("Using cached image (cache TTL not expired)")
				return nil
			}

			// Cache expired, pull for updates
			pullPolicy = "Always"
			log.Debug().Str("image", imageName).Str("tag", tag).Msg("Auto-detected pull policy: Always (dynamic tag, cache expired)")
		} else if hasLocalImage && tag != "" && tag != "latest" && tag != "main" && tag != "master" {
			// For specific version tags, check if it's a local build first (only if loadLocal is true)
			if loadLocal && len(localImageInfo.RepoDigests) == 0 {
				// Local build with version tag
				fmt.Printf("🏠 Using local development image: %s\n", imageName)
				log.Info().Str("image", imageName).Msg("Using local development image (AutoDetect: versioned local build)")
				return nil
			}
			// For remote images with version tags, use local if present
			// Version tags are immutable by convention, so we can cache aggressively
			log.Info().Str("image", imageName).Msg("Using cached image (AutoDetect: version tag)")
			return nil
		} else if tag == "latest" || tag == "main" || tag == "master" || tag == "" {
			// For dynamic tags without recent local image, always pull
			pullPolicy = "Always"
			log.Debug().Str("image", imageName).Str("tag", tag).Msg("Auto-detected pull policy: Always (dynamic tag)")
		} else {
			// For specific version tags, use IfNotPresent for aggressive caching
			// This includes: v1.0.0, 1.2.3, dev-59-abc123, release-2.0, etc.
			// Version tags are immutable by convention
			pullPolicy = "IfNotPresent"
			log.Debug().Str("image", imageName).Str("tag", tag).Msg("Auto-detected pull policy: IfNotPresent (version tag - cached aggressively)")
		}
	}

	// Handle "Never" policy - only use local image
	if pullPolicy == "Never" {
		_, err := ch.client.ImageInspect(ch.ctx, imageName)
		if err != nil {
			return fmt.Errorf("image pull policy is 'Never' but image '%s' not found locally", imageName)
		}
		log.Info().Str("image", imageName).Msg("Using local image (pull policy: Never)")
		return nil
	}

	// Handle "IfNotPresent" policy - check locally first
	if pullPolicy == "IfNotPresent" {
		_, err := ch.client.ImageInspect(ch.ctx, imageName)
		if err == nil {
			// Image exists locally, no need to pull
			log.Info().Str("image", imageName).Msg("Image already exists locally")
			return nil
		}
	}

	// For "Always" policy or when image not found with "IfNotPresent"
	log.Info().Str("image", imageName).Str("pullPolicy", pullPolicy).Msg("Pulling image from registry")

	// Get GitHub authentication using centralized function
	authConfig, authStr, err := CreateGitHubAuthConfig()
	if err != nil {
		return fmt.Errorf("error creating auth config: %w", err)
	}

	// Check if Docker daemon is running before attempting login
	_, pingErr := ch.client.Ping(ch.ctx)
	if pingErr != nil {
		// Check for common Docker service not running errors
		errStr := pingErr.Error()
		if strings.Contains(errStr, "docker_engine") ||
		   strings.Contains(errStr, "cannot connect to the Docker daemon") ||
		   strings.Contains(errStr, "Is the docker daemon running") ||
		   strings.Contains(errStr, "system cannot find the file specified") {
			return fmt.Errorf("Docker service is not running. Please start Docker Desktop or the Docker daemon and try again")
		}
		return fmt.Errorf("cannot connect to Docker: %w", pingErr)
	}

	// Log in to registry
	loginResp, err := ch.client.RegistryLogin(ch.ctx, *authConfig)
	if err != nil {
		// Check if this is a Docker service issue
		errStr := err.Error()
		if strings.Contains(errStr, "docker_engine") ||
		   strings.Contains(errStr, "cannot connect to the Docker daemon") ||
		   strings.Contains(errStr, "system cannot find the file specified") {
			return fmt.Errorf("Docker service is not running. Please start Docker Desktop or the Docker daemon and try again")
		}
		return fmt.Errorf("error logging in to registry: %w", err)
	}
	log.Info().Str("status", loginResp.Status).Msg("Successfully logged in to registry")

	// Pull image with user feedback
	fmt.Printf("🔍 Contacting registry for %s...\n", imageName)
	reader, err := ch.client.ImagePull(ch.ctx, imageName, image.PullOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return fmt.Errorf("error pulling image: %w", err)
	}
	defer reader.Close()

	// Display progress to user
	if err := DisplayDockerProgress(reader); err != nil {
		return fmt.Errorf("error during image pull: %w", err)
	}

	log.Info().Str("image", imageName).Msg("Successfully pulled image")

	// Cache the pulled image digest
	tag := ch.extractTag(imageName)
	if pulledImageInfo, err := ch.client.ImageInspect(ch.ctx, imageName); err == nil {
		ch.cacheImageDigest(imageName, tag, pulledImageInfo)
	}

	return nil
}

// ExecuteMetadataCommand executes the "extension-meta" command in an extension container
// and returns the raw YAML output string or an error.
// Note: Callers must ensure the image exists before calling this function.
func (ch *ContainerHost) ExecuteMetadataCommand(ext *ExtensionConfig) (string, error) {
	// Inspect image to get configuration
	imageInspect, err := ch.InspectImage(ext.Image)
	if err != nil {
		return "", fmt.Errorf("error inspecting image: %w", err)
	}

	// Create container configuration for metadata command
	containerConfig := ch.CreateContainerConfig(ext, ModeRun, []string{"extension-meta"}, imageInspect)

	// Override TTY and stdin settings for metadata retrieval
	containerConfig.Tty = false
	containerConfig.OpenStdin = false

	// No cache volumes for metadata retrieval (avoid chicken-and-egg problem)
	hostConfig := ch.CreateHostConfig(ext, nil)

	// Create container
	containerID, err := ch.CreateContainer(containerConfig, hostConfig)
	if err != nil {
		return "", fmt.Errorf("error creating container: %w", err)
	}

	// Attach to container to capture output
	attachResp, err := ch.client.ContainerAttach(ch.ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("error attaching to container: %w", err)
	}
	defer attachResp.Close()

	// Start container
	if err := ch.StartContainer(containerID); err != nil {
		return "", fmt.Errorf("error starting container: %w", err)
	}

	// Create timeout context (60 seconds)
	timeoutCtx, cancel := context.WithTimeout(ch.ctx, 60*time.Second)
	defer cancel()

	// Wait for container to finish with timeout
	statusCh, errCh := ch.client.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)

	// Capture output - use stdcopy to demultiplex since TTY=false
	outputChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		// Since TTY is disabled, Docker multiplexes stdout/stderr with headers
		// We need to use stdcopy.StdCopy to demultiplex
		var stdout, stderr bytes.Buffer
		_, err := stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
		if err != nil {
			errorChan <- fmt.Errorf("error reading output: %w", err)
			return
		}
		// Return stdout content (stderr is typically empty for metadata command)
		outputChan <- stdout.String()
	}()

	// Wait for container completion or timeout
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		// Check exit code
		if status.StatusCode != 0 {
			// Try to get error output
			select {
			case output := <-outputChan:
				return "", fmt.Errorf("extension-meta command failed with exit code %d: %s", status.StatusCode, output)
			case <-time.After(1 * time.Second):
				return "", fmt.Errorf("extension-meta command failed with exit code %d", status.StatusCode)
			}
		}
	case <-timeoutCtx.Done():
		// Timeout occurred, try to stop the container
		_ = ch.StopContainer(containerID)
		return "", fmt.Errorf("extension-meta command timed out after 60 seconds")
	}

	// Get the output
	select {
	case output := <-outputChan:
		return output, nil
	case err := <-errorChan:
		return "", err
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("timeout reading command output")
	}
}

// GetExtensionMetadata fetches and caches extension metadata
// Returns cached metadata if valid, otherwise fetches fresh metadata
func (ch *ContainerHost) GetExtensionMetadata(ext *ExtensionConfig) (*cache.ExtensionMeta, error) {
	// Get current image digest for cache validation
	imageDigest, err := ch.GetImageDigest(ext.Image)
	if err != nil {
		log.Debug().Err(err).Str("image", ext.Image).Msg("Failed to get image digest, will fetch fresh metadata")
		imageDigest = "" // Continue without digest, metadata will be fetched
	}

	// Try to load cached metadata
	cachedMeta, err := cache.LoadMetadataCache(ch.rootDir, ext.Name)
	if err != nil {
		log.Warn().Err(err).Str("extension", ext.Name).Msg("Error loading metadata cache")
	}

	// Check if cache is valid
	if cache.IsMetadataCacheValid(cachedMeta, imageDigest) {
		log.Debug().
			Str("extension", ext.Name).
			Str("digest", imageDigest).
			Msg("Using cached extension metadata")
		return cachedMeta.Metadata, nil
	}

	// Fetch fresh metadata
	log.Debug().Str("extension", ext.Name).Msg("Fetching fresh extension metadata")
	yamlOutput, err := ch.ExecuteMetadataCommand(ext)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch extension metadata: %w", err)
	}

	// Parse YAML metadata
	meta, err := parseExtensionMetadata(yamlOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extension metadata: %w", err)
	}

	// Save to cache
	newCache := cache.NewMetadataCache(ext.Name, imageDigest, meta)
	if err := cache.SaveMetadataCache(ch.rootDir, newCache); err != nil {
		log.Warn().Err(err).Str("extension", ext.Name).Msg("Failed to save metadata cache")
		// Continue even if cache save fails
	}

	return meta, nil
}

// parseExtensionMetadata parses YAML extension metadata into ExtensionMeta struct
func parseExtensionMetadata(yamlData string) (*cache.ExtensionMeta, error) {
	// Import yaml package is needed - using json as intermediate for simplicity
	// since the YAML structure matches our JSON struct tags
	var meta cache.ExtensionMeta

	// Use gopkg.in/yaml.v3 to parse
	if err := yaml.Unmarshal([]byte(yamlData), &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata YAML: %w", err)
	}

	return &meta, nil
}

// MergeMetadataEnv merges environment variables from extension metadata into config
// Metadata env vars are added first, then config env vars override them
func MergeMetadataEnv(ext *ExtensionConfig, meta *cache.ExtensionMeta) {
	if meta == nil || len(meta.Env) == 0 {
		return
	}

	// Build a map of existing config env vars for deduplication
	existingEnv := make(map[string]bool)
	for _, env := range ext.Env {
		existingEnv[env.Name] = true
	}

	// Add metadata env vars that aren't already in config
	for _, metaEnv := range meta.Env {
		if !existingEnv[metaEnv.Name] {
			ext.Env = append(ext.Env, conf.EnvVar{
				Name:     metaEnv.Name,
				Value:    metaEnv.Value,
				Required: metaEnv.Required,
			})
			log.Debug().
				Str("env", metaEnv.Name).
				Bool("required", metaEnv.Required).
				Msg("Added environment variable from extension metadata")
		}
	}
}

// Close closes the Docker client connection
// GetContainerSnapshot returns a snapshot of currently running containers
func (ch *ContainerHost) GetContainerSnapshot() (map[string]string, error) {
	containers, err := ch.client.ContainerList(ch.ctx, container.ListOptions{
		All: false, // Only running containers
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	snapshot := make(map[string]string)
	for _, cont := range containers {
		// Use container ID as key, image as value for identification
		snapshot[cont.ID] = cont.Image
	}

	return snapshot, nil
}

// WarnAboutNewContainers compares before/after snapshots and warns about new containers
// If autoRemove is true, it will stop and remove the containers instead of just warning
// expectedHostImages is a list of images that are expected to be created by the extension (e.g., serve commands)
func (ch *ContainerHost) WarnAboutNewContainers(beforeSnapshot, afterSnapshot map[string]string, extensionImage string, autoRemove bool, expectedHostImages []string) {
	for containerID, image := range afterSnapshot {
		if _, existed := beforeSnapshot[containerID]; !existed {
			// Skip our own main container
			if image == extensionImage {
				continue
			}

			// Skip expected host containers (e.g., from serve commands)
			if isExpectedHostImage(image, expectedHostImages) {
				log.Debug().
					Str("container_id", containerID[:12]).
					Str("image", image).
					Msg("Skipping expected host container")
				continue
			}

			if autoRemove {
				log.Info().
					Str("container_id", containerID[:12]).
					Str("image", image).
					Str("extension", extensionImage).
					Msg("Auto-removing detected child container: " + image)

				// Stop and remove the container
				if err := ch.client.ContainerStop(ch.ctx, containerID, container.StopOptions{}); err != nil {
					log.Warn().
						Str("container_id", containerID[:12]).
						Str("error", err.Error()).
						Msg("Failed to stop child container")
				}

				if err := ch.client.ContainerRemove(ch.ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
					log.Warn().
						Str("container_id", containerID[:12]).
						Str("error", err.Error()).
						Msg("Failed to remove child container")
				} else {
					log.Info().
						Str("container_id", containerID[:12]).
						Msg("Successfully removed child container")
				}
			} else {
				log.Warn().
					Str("container_id", containerID[:12]).
					Str("image", image).
					Str("extension", extensionImage).
					Msg("New container appeared during run: " + image + ". This could be an indication of missing internal cleanup of docker-in-docker for extension " + extensionImage)
			}
		}
	}
}

func (ch *ContainerHost) Close() error {
	return ch.client.Close()
}

// isExpectedHostImage checks if an image is in the list of expected host images
func isExpectedHostImage(image string, expectedImages []string) bool {
	for _, expected := range expectedImages {
		if image == expected {
			return true
		}
	}
	return false
}

// extractTag extracts the tag from an image name (format: registry/repo:tag)
func (ch *ContainerHost) extractTag(imageName string) string {
	tagIndex := strings.LastIndex(imageName, ":")
	if tagIndex > 0 && tagIndex < len(imageName)-1 {
		return imageName[tagIndex+1:]
	}
	return "latest" // Default tag
}

// extractExtensionName extracts the extension name from an image name
// For ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:0.0.2 -> "pwsh"
func (ch *ContainerHost) extractExtensionName(imageName string) string {
	// Remove tag first
	imageWithoutTag := imageName
	if idx := strings.LastIndex(imageName, ":"); idx > 0 {
		imageWithoutTag = imageName[:idx]
	}

	// Extract last path component as extension name
	parts := strings.Split(imageWithoutTag, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// cacheImageDigest saves the image digest to the registry cache for future lookups
func (ch *ContainerHost) cacheImageDigest(imageName, tag string, imageInfo image.InspectResponse) {
	// Get or create the digest to cache
	var digest string
	if len(imageInfo.RepoDigests) > 0 {
		digest = imageInfo.RepoDigests[0]
	} else {
		digest = imageInfo.ID
	}

	// Extract extension name from image
	extensionName := ch.extractExtensionName(imageName)

	// Load registry cache
	registryCache, err := cache.LoadRegistryCache(ch.rootDir)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load registry cache for digest update")
		return
	}

	// Update the digest in cache
	registryCache.SetImageDigest(extensionName, tag, digest)

	// Save cache
	if err := registryCache.SaveRegistryCache(ch.rootDir); err != nil {
		log.Warn().Err(err).Msg("Failed to save registry cache after digest update")
	} else {
		log.Debug().
			Str("extension", extensionName).
			Str("tag", tag).
			Str("digest", digest).
			Msg("Cached image digest")
	}
}

// ContainerCleanupOptions configures container cleanup behavior
type ContainerCleanupOptions struct {
	OnlyExtensions bool          // Only clean extension containers
	IncludeRunning bool          // Also remove running containers (default: false)
	OlderThan      time.Duration // Only remove containers older than this duration
	DryRun         bool          // Show what would be removed without removing
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	ContainersRemoved int
	SpaceReclaimed    int64
	Errors            []error
}

// CleanupContainers removes containers based on the provided options
func (ch *ContainerHost) CleanupContainers(opts ContainerCleanupOptions) (*CleanupResult, error) {
	result := &CleanupResult{}

	// List all containers (including stopped)
	containers, err := ch.client.ContainerList(ch.ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	log.Debug().Int("total_containers", len(containers)).Msg("Scanning containers for cleanup")

	for _, ctr := range containers {
		// Skip if container is running and we don't want to include running containers
		if ctr.State == "running" && !opts.IncludeRunning {
			continue
		}

		// Filter by extension containers if requested
		if opts.OnlyExtensions {
			// Check if container name or labels indicate it's an extension container
			isExtension := false
			for _, name := range ctr.Names {
				if strings.Contains(name, "r2r-") || strings.Contains(name, "extension-") {
					isExtension = true
					break
				}
			}
			if !isExtension {
				continue
			}
		}

		// Filter by age if specified
		if opts.OlderThan > 0 {
			createdTime := time.Unix(ctr.Created, 0)
			age := time.Since(createdTime)
			if age < opts.OlderThan {
				continue
			}
		}

		// Container matches criteria, remove it
		containerName := ctr.Names[0]
		if len(containerName) > 0 && containerName[0] == '/' {
			containerName = containerName[1:] // Remove leading slash
		}

		if opts.DryRun {
			log.Info().
				Str("container", containerName).
				Str("state", ctr.State).
				Str("image", ctr.Image).
				Msg("[DRY RUN] Would remove container")
			result.ContainersRemoved++
		} else {
			log.Info().
				Str("container", containerName).
				Str("state", ctr.State).
				Str("image", ctr.Image).
				Msg("Removing container")

			// Get container size before removal
			var containerSize int64
			inspectData, inspectErr := ch.client.ContainerInspect(ch.ctx, ctr.ID)
			if inspectErr == nil && inspectData.SizeRw != nil {
				containerSize = *inspectData.SizeRw
			}

			// Remove the container
			removeOpts := container.RemoveOptions{
				Force:         opts.IncludeRunning, // Force remove if including running containers
				RemoveVolumes: true,                // Also remove associated volumes
			}

			if err := ch.client.ContainerRemove(ch.ctx, ctr.ID, removeOpts); err != nil {
				log.Warn().
					Err(err).
					Str("container", containerName).
					Msg("Failed to remove container")
				result.Errors = append(result.Errors, fmt.Errorf("failed to remove %s: %w", containerName, err))
			} else {
				result.ContainersRemoved++
				result.SpaceReclaimed += containerSize // Only count space for successful removals
			}
		}
	}

	log.Info().
		Int("removed", result.ContainersRemoved).
		Int64("space_reclaimed_bytes", result.SpaceReclaimed).
		Int("errors", len(result.Errors)).
		Bool("dry_run", opts.DryRun).
		Msg("Container cleanup completed")

	return result, nil
}

// CleanupStoppedContainers is a convenience wrapper for cleaning up all stopped containers
func (ch *ContainerHost) CleanupStoppedContainers(dryRun bool) (*CleanupResult, error) {
	return ch.CleanupContainers(ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		DryRun:         dryRun,
	})
}

// CleanupExtensionContainers is a convenience wrapper for cleaning up extension containers
func (ch *ContainerHost) CleanupExtensionContainers(includeRunning, dryRun bool) (*CleanupResult, error) {
	return ch.CleanupContainers(ContainerCleanupOptions{
		OnlyExtensions: true,
		IncludeRunning: includeRunning,
		DryRun:         dryRun,
	})
}

// CleanupOldContainers removes containers older than the specified duration
func (ch *ContainerHost) CleanupOldContainers(olderThan time.Duration, dryRun bool) (*CleanupResult, error) {
	return ch.CleanupContainers(ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		OlderThan:      olderThan,
		DryRun:         dryRun,
	})
}
