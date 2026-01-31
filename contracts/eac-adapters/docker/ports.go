// Package docker defines the interface contract for Docker container operations.
//
// This package provides interfaces for:
//   - Docker client operations (images, containers)
//   - Container serving lifecycle management
//   - Port allocation and reservation
//
// Implementations are provided by the github.com/ready-to-release/eac/go/eac/adapters/docker module.
package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Client defines the interface for Docker operations.
// This interface allows for mocking Docker operations in unit tests.
type Client interface {
	// ImageList lists Docker images.
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)

	// ImageBuild builds a Docker image from a tar archive.
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)

	// ImagePull pulls a Docker image from a registry.
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)

	// ContainerCreate creates a new container.
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)

	// ContainerStart starts a container.
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error

	// ContainerList lists containers.
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)

	// ContainerRemove removes a container.
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error

	// ContainerStop stops a running container.
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error

	// ContainerWait waits for a container to finish and returns its status.
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)

	// ContainerLogs gets the logs from a container.
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)

	// ContainerAttach attaches to a container for streaming stdout/stderr.
	ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error)

	// Close closes the Docker client connection.
	Close() error
}

// ServeConfig holds configuration for serving a website in a container.
type ServeConfig struct {
	// Name is the base container name (e.g., "cli-mkdocs", "structurizr-lite-mymodule").
	// The actual container name will include the port for multi-instance support.
	Name string

	// Image is the Docker image to use.
	Image string

	// BuildInfo contains information for building the image locally.
	// If nil, the image is expected to exist or be pulled from a registry.
	BuildInfo *BuildInfo

	// ContentPath is the local path to the content to serve.
	// In DinD mode, this will be translated to the host path automatically.
	ContentPath string

	// ContainerPath is where to mount the content inside the container.
	ContainerPath string

	// ContainerPort is the port the service listens on inside the container.
	ContainerPort int

	// Command is the command to run (optional, uses image default if nil).
	Command []string

	// EnvVars are additional environment variables to set.
	EnvVars []string

	// RestartPolicy sets the container restart policy (default: "unless-stopped").
	RestartPolicy string

	// PreferredPort is the preferred host port (0 = auto-allocate).
	PreferredPort int

	// Memory is the memory limit in bytes (0 = no limit).
	Memory int64

	// CPUs is the number of CPUs to allocate (0 = no limit).
	CPUs float64
}

// BuildInfo holds information for building a local Docker image.
type BuildInfo struct {
	// Dockerfile is the path to the Dockerfile.
	Dockerfile string
	// ContextPath is the build context path.
	ContextPath string
}

// ServeResult holds the result of starting a serve container.
type ServeResult struct {
	// ContainerID is the Docker container ID.
	ContainerID string
	// ContainerName is the full container name (includes port).
	ContainerName string
	// HostPort is the port on the host where the service is accessible.
	HostPort int
	// URL is the full URL to access the service.
	URL string
}

// ServeManager manages container serving lifecycle.
type ServeManager interface {
	// StartServe starts a container serving web content.
	StartServe(ctx context.Context, config *ServeConfig) (*ServeResult, error)

	// StopServe stops a serving container by name or name prefix.
	StopServe(ctx context.Context, namePattern string) error

	// IsServing checks if a container with the given name pattern is currently serving.
	IsServing(ctx context.Context, namePattern string) (*ServeResult, bool, error)

	// ListServing returns all running serve containers matching the name pattern.
	ListServing(ctx context.Context, namePattern string) ([]*ServeResult, error)

	// CheckImageStale checks if the image for a config needs to be rebuilt.
	CheckImageStale(ctx context.Context, config *ServeConfig) (bool, string, error)
}

// PortManager handles port allocation and reservation.
type PortManager interface {
	// FindAvailablePort finds an unused port in the configured port range.
	FindAvailablePort() (int, error)

	// IsPortAvailable checks if a specific port is available for binding.
	IsPortAvailable(port int) bool

	// FindAndReservePortOrUse returns the preferred port if available and reserves it,
	// otherwise finds and reserves a random available port.
	FindAndReservePortOrUse(preferredPort int) (int, error)

	// ReleasePort releases a previously reserved port.
	ReleasePort(port int)
}

// ImageOperation defines an image that needs to be ensured (built or pulled).
type ImageOperation struct {
	// Config is the serve configuration containing image details.
	Config *ServeConfig
	// Index is the operation index for tracking (optional).
	Index int
}

// ImageResult holds the result of an image operation.
type ImageResult struct {
	// Image is the image name/tag.
	Image string
	// Index is the operation index.
	Index int
	// Error is any error that occurred.
	Error error
	// Success indicates if the operation succeeded.
	Success bool
}

// ParallelImageEnsureOptions configures parallel image operations.
type ParallelImageEnsureOptions struct {
	// MaxConcurrency limits how many images can be processed simultaneously.
	// Default is 3 if not specified.
	MaxConcurrency int
}
