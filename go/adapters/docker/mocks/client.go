//nolint:errcheck // Mock implementations use type assertions that are controlled in tests
package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
)

// MockDockerClient is a mock implementation of DockerClient for testing.
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	args := m.Called(ctx, options)
	if images := args.Get(0); images != nil {
		return images.([]image.Summary), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	args := m.Called(ctx, buildContext, options)
	return args.Get(0).(types.ImageBuildResponse), args.Error(1)
}

func (m *MockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, refStr, options)
	if rc := args.Get(0); rc != nil {
		return rc.(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDockerClient) ImageInspectWithRaw(ctx context.Context, imageID string) (image.InspectResponse, []byte, error) {
	args := m.Called(ctx, imageID)
	return args.Get(0).(image.InspectResponse), args.Get(1).([]byte), args.Error(2)
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	args := m.Called(ctx, options)
	if containers := args.Get(0); containers != nil {
		return containers.([]types.Container), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	args := m.Called(ctx, containerID, condition)
	if waitChan := args.Get(0); waitChan != nil {
		return waitChan.(<-chan container.WaitResponse), args.Get(1).(<-chan error)
	}
	return nil, nil
}

func (m *MockDockerClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, containerID, options)
	if rc := args.Get(0); rc != nil {
		return rc.(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDockerClient) ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error) {
	args := m.Called(ctx, containerID, options)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (m *MockDockerClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
