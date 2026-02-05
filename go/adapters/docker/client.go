package docker

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/go/core/environments"
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
// Returns true when R2R_MOCK_DOCKER or R2R_MOCK_STRUCTURIZR (legacy) environment variable is set to "true".
// This is set by the BDD test infrastructure to prevent real Docker operations.
func shouldUseMockDockerClient() bool {
	return os.Getenv(environments.EnvR2RMockDocker) == "true" ||
		os.Getenv(environments.EnvR2RMockStructurizr) == "true" // Legacy support
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
	mockClient := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	// Load persisted state from disk if it exists
	mockClient.loadState()
	return mockClient
}

// ResetMockDockerClient resets the mock state by deleting the persisted state file.
// This should be called between test scenarios to prevent state leakage.
func ResetMockDockerClient() {
	deleteStateFile()
}

// Ensure RealDockerClient implements DockerClient interface.
var _ DockerClient = (*RealDockerClient)(nil)

// IsDockerAvailable checks if Docker is available and running.
// This is a convenience wrapper around util.IsDockerAvailable.
func IsDockerAvailable() bool {
	// Quick check: try to create a client and ping Docker
	cli, err := newRealDockerClient()
	if err != nil {
		return false
	}
	defer cli.Close()

	// Try to ping Docker daemon
	_, err = cli.(*RealDockerClient).Ping(context.Background())
	return err == nil
}

// IsWindowsDockerHost checks if the Docker daemon is running on a Windows host.
// This is needed to enable polling for file watching since filesystem events
// don't propagate through Docker mounted volumes on Windows.
// Works correctly in DinD environments by querying the Docker daemon directly.
func IsWindowsDockerHost() bool {
	cli, err := newRealDockerClient()
	if err != nil {
		return false
	}
	defer cli.Close()

	info, err := cli.(*RealDockerClient).Info(context.Background())
	if err != nil {
		return false
	}

	// Check kernel version for WSL2 (Windows Subsystem for Linux)
	// WSL2 kernels contain "microsoft" in the version string
	kernelVersion := strings.ToLower(info.KernelVersion)
	if strings.Contains(kernelVersion, "microsoft") {
		return true
	}

	// Also check OperatingSystem for "Docker Desktop" which runs on Windows/macOS
	operatingSystem := strings.ToLower(info.OperatingSystem)
	return strings.Contains(operatingSystem, "docker desktop")
}
