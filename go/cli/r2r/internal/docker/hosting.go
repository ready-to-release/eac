package docker

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/conf"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/logging"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/envconsts"
)

// ContainerMode defines how the container should be configured.
type ContainerMode int

const (
	ModeRun ContainerMode = iota
	ModeInteractive
)

// ExtensionConfig holds the configuration for an extension.
type ExtensionConfig struct {
	Name               string
	Image              string
	ImagePullPolicy    string
	LoadLocal          bool
	AutoRemoveChildren bool
	Env                []conf.EnvVar
}

// ContainerHost manages Docker container operations for extensions.
type ContainerHost struct {
	client     DockerClient
	ctx        context.Context
	rootDir    string
	clientOpts []client.Opt   // stored for lazy initialization
	pingOnce   sync.Once      // ensures Ping only happens once
	pingErr    error          // stores Ping error if any
}

// NewContainerHost creates a new ContainerHost instance.
// Docker connectivity is verified lazily on first use for faster startup (~800ms savings).
func NewContainerHost() (*ContainerHost, error) {
	ctx := context.Background()

	// Configure Docker client options
	clientOpts := []client.Opt{client.FromEnv}

	// Force Docker API version negotiation for compatibility
	// This prevents "client version X is too new" errors
	clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())

	// Override Docker host if R2R_DOCKER_HOST is set
	if dockerHost := os.Getenv(envconsts.EnvR2RDockerHost); dockerHost != "" {
		clientOpts = append(clientOpts, client.WithHost(dockerHost))
		logging.Debugf("Using custom Docker host from R2R_DOCKER_HOST: docker_host=%s", dockerHost)
	}

	cli, err := NewRealDockerClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}

	// Use cached RootDir from config (Phase 4 optimization: ~50ms savings)
	// Falls back to FindRepositoryRoot() if not cached
	rootDir := conf.RootDir
	if rootDir == "" {
		rootDir, err = conf.FindRepositoryRoot()
		if err != nil {
			cli.Close()
			return nil, fmt.Errorf("error finding root directory: %w", err)
		}
	}

	return &ContainerHost{
		client:     cli,
		ctx:        ctx,
		rootDir:    rootDir,
		clientOpts: clientOpts,
	}, nil
}

// EnsureConnected verifies Docker daemon connectivity (lazy Ping).
// This should be called before operations that require Docker communication.
// Uses sync.Once to ensure the check only happens once per ContainerHost instance.
func (ch *ContainerHost) EnsureConnected() error {
	ch.pingOnce.Do(func() {
		_, pingErr := ch.client.Ping(ch.ctx)
		if pingErr != nil {
			// Check for common Docker service not running errors
			errStr := pingErr.Error()
			if strings.Contains(errStr, "docker_engine") ||
				strings.Contains(errStr, "cannot connect to the Docker daemon") ||
				strings.Contains(errStr, "Is the docker daemon running") ||
				strings.Contains(errStr, "system cannot find the file specified") {
				ch.pingErr = fmt.Errorf("docker service is not running: please start Docker Desktop or the Docker daemon and try again")
			} else {
				ch.pingErr = fmt.Errorf("cannot connect to Docker daemon: %w", pingErr)
			}
		}
	})
	return ch.pingErr
}

// ValidateExtensions checks if extensions are configured.
func (ch *ContainerHost) ValidateExtensions() error {
	if len(conf.Global.Extensions) == 0 {
		return fmt.Errorf("config file does not contain any extensions. Please run 'r2r init' to initialize the configuration")
	}
	return nil
}

// FindExtension locates an extension by name in the configuration.
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
