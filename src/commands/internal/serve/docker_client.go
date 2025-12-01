package serve

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// DockerClient defines the interface for Docker operations used by the serve package.
// This interface allows for mocking Docker operations in unit tests.
type DockerClient interface {
	// ImageList lists Docker images
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)

	// ImageBuild builds a Docker image from a tar archive
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)

	// ImagePull pulls a Docker image from a registry
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)

	// ContainerCreate creates a new container
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)

	// ContainerStart starts a container
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error

	// ContainerList lists containers
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)

	// ContainerRemove removes a container
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error

	// ContainerStop stops a running container
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error

	// ContainerWait waits for a container to finish and returns its status
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)

	// ContainerLogs gets the logs from a container
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)

	// Close closes the Docker client connection
	Close() error
}

// RealDockerClient wraps the official Docker client to implement DockerClient interface
type RealDockerClient struct {
	*client.Client
}

// NewDockerClient creates a new Docker client with default options
func NewDockerClient() (DockerClient, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	return &RealDockerClient{Client: cli}, nil
}

// Ensure RealDockerClient implements DockerClient interface
var _ DockerClient = (*RealDockerClient)(nil)
