//nolint:errcheck // mock type assertions are intentional - panic on incorrect test setup is expected
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
	"github.com/stretchr/testify/mock"
)

// MockDockerClient is a mock implementation of DockerClient for testing.
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) Ping(ctx context.Context) (types.Ping, error) {
	args := m.Called(ctx)
	return args.Get(0).(types.Ping), args.Error(1)
}

func (m *MockDockerClient) ImageInspect(ctx context.Context, imageID string) (image.InspectResponse, error) {
	args := m.Called(ctx, imageID)
	if inspect := args.Get(0); inspect != nil {
		return inspect.(image.InspectResponse), args.Error(1)
	}
	return image.InspectResponse{}, args.Error(1)
}

func (m *MockDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	args := m.Called(ctx, options)
	if images := args.Get(0); images != nil {
		return images.([]image.Summary), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, refStr, options)
	if rc := args.Get(0); rc != nil {
		return rc.(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDockerClient) RegistryLogin(ctx context.Context, auth registry.AuthConfig) (registry.AuthenticateOKBody, error) {
	args := m.Called(ctx, auth)
	return args.Get(0).(registry.AuthenticateOKBody), args.Error(1)
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	args := m.Called(ctx, containerID)
	if inspect := args.Get(0); inspect != nil {
		return inspect.(container.InspectResponse), args.Error(1)
	}
	return container.InspectResponse{}, args.Error(1)
}

func (m *MockDockerClient) ContainerResize(ctx context.Context, containerID string, options container.ResizeOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error) {
	args := m.Called(ctx, containerID, options)
	if resp := args.Get(0); resp != nil {
		return resp.(types.HijackedResponse), args.Error(1)
	}
	return types.HijackedResponse{}, args.Error(1)
}

func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	args := m.Called(ctx, containerID, condition)
	return args.Get(0).(<-chan container.WaitResponse), args.Get(1).(<-chan error)
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	args := m.Called(ctx, options)
	if containers := args.Get(0); containers != nil {
		return containers.([]container.Summary), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDockerClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockDockerClient implements DockerClient.
var _ DockerClient = (*MockDockerClient)(nil)
