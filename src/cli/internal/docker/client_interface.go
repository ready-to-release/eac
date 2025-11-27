package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// DockerClient defines the interface for Docker operations used by the CLI module.
// This interface allows for mocking Docker operations in unit tests.
type DockerClient interface {
	// Ping pings the Docker daemon to verify connectivity
	Ping(ctx context.Context) (types.Ping, error)

	// ImageInspect inspects a Docker image and returns detailed information
	ImageInspect(ctx context.Context, imageID string) (image.InspectResponse, error)

	// ImageList lists Docker images
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)

	// ImagePull pulls a Docker image from a registry
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)

	// RegistryLogin authenticates with a Docker registry
	RegistryLogin(ctx context.Context, auth registry.AuthConfig) (registry.AuthenticateOKBody, error)

	// ContainerCreate creates a new container
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)

	// ContainerStart starts a container
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error

	// ContainerInspect inspects a container and returns detailed information
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)

	// ContainerResize resizes the TTY of a container
	ContainerResize(ctx context.Context, containerID string, options container.ResizeOptions) error

	// ContainerAttach attaches to a container for I/O operations
	ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error)

	// ContainerWait waits for a container to finish execution
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)

	// ContainerStop stops a running container
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error

	// ContainerRemove removes a container
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error

	// ContainerList lists containers
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)

	// Close closes the Docker client connection
	Close() error
}
