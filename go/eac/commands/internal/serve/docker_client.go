package serve

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/go/eac/core/environments"
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

	// ContainerAttach attaches to a container for streaming stdout/stderr
	ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error)

	// Close closes the Docker client connection
	Close() error
}

// RealDockerClient wraps the official Docker client to implement DockerClient interface.
type RealDockerClient struct {
	*client.Client
}

// NewDockerClient creates a Docker client based on environment.
// Returns a mock client if R2R_MOCK_STRUCTURIZR=true, otherwise returns a real client.
func NewDockerClient() (DockerClient, error) {
	if shouldUseMockDockerClient() {
		return newMockDockerClientForTests(), nil
	}
	return newRealDockerClient()
}

// shouldUseMockDockerClient checks if mock client should be used.
// Returns true when R2R_MOCK_STRUCTURIZR environment variable is set to "true".
// This is set by the BDD test infrastructure to prevent real Docker operations.
func shouldUseMockDockerClient() bool {
	return os.Getenv(environments.EnvR2RMockStructurizr) == "true"
}

// newRealDockerClient creates a real Docker client (extracted from NewDockerClient).
func newRealDockerClient() (DockerClient, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &RealDockerClient{Client: cli}, nil
}

// newMockDockerClientForTests creates a mock client that persists state to disk.
// This allows state to persist across separate process invocations (e.g., BDD tests
// that run commands as subprocesses), simulating how a real Docker daemon maintains state.
func newMockDockerClientForTests() DockerClient {
	client := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	// Load persisted state from disk if it exists
	client.loadState()
	return client
}

// ResetMockDockerClient resets the mock state by deleting the persisted state file.
// This should be called between test scenarios to prevent state leakage.
func ResetMockDockerClient() {
	deleteStateFile()
}

// Ensure RealDockerClient implements DockerClient interface.
var _ DockerClient = (*RealDockerClient)(nil)
