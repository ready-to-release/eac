package serve

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestStartServe_Success tests successful container creation and startup
func TestStartServe_Success(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	config := &ServeConfig{
		Name:          "test-serve",
		Image:         "test:latest",
		ContentPath:   "/test/content",
		ContainerPath: "/data",
		ContainerPort: 8000,
		PreferredPort: 0, // Auto-allocate
	}

	// Mock: Check container doesn't exist
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{}, nil).Once()

	// Mock: Image exists
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{
		{RepoTags: []string{"test:latest"}},
	}, nil)

	// Mock: No existing container to remove
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{}, nil).Once()

	// Mock: Container creation
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(name string) bool {
		return strings.HasPrefix(name, "test-serve-")
	})).Return(container.CreateResponse{ID: "container123"}, nil)

	// Mock: Container start
	mockClient.On("ContainerStart", ctx, "container123", container.StartOptions{}).Return(nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient for testing
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	result, err := StartServe(ctx, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "container123", result.ContainerID)
	assert.True(t, strings.HasPrefix(result.ContainerName, "test-serve-"))
	assert.Greater(t, result.HostPort, 0)
	assert.Contains(t, result.URL, "http://localhost:")

	mockClient.AssertExpectations(t)

	// Clean up reserved port
	ReleasePort(result.HostPort)
}

// TestStartServe_ContainerAlreadyRunning tests that StartServe returns existing container
func TestStartServe_ContainerAlreadyRunning(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	config := &ServeConfig{
		Name:          "test-serve",
		Image:         "test:latest",
		ContentPath:   "/test/content",
		ContainerPath: "/data",
		ContainerPort: 8000,
		PreferredPort: 9000,
	}

	// Mock: Container already running
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:    "existing123",
			Names: []string{"/test-serve-9000"},
			State: "running",
			Ports: []types.Port{{PublicPort: 9000}},
		},
	}, nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	result, err := StartServe(ctx, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "existing123", result.ContainerID)
	assert.Equal(t, "test-serve-9000", result.ContainerName)
	assert.Equal(t, 9000, result.HostPort)

	mockClient.AssertExpectations(t)
}

// TestStartServe_ImagePullRequired tests image pulling when image doesn't exist
func TestStartServe_ImagePullRequired(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	config := &ServeConfig{
		Name:          "test-serve",
		Image:         "new-image:latest",
		ContentPath:   "/test/content",
		ContainerPath: "/data",
		ContainerPort: 8000,
		PreferredPort: 0,
	}

	// Mock: Container doesn't exist
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{}, nil)

	// Mock: Image doesn't exist
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)

	// Mock: Image pull
	mockClient.On("ImagePull", ctx, "new-image:latest", image.PullOptions{}).
		Return(io.NopCloser(strings.NewReader("{}")), nil)

	// Mock: Container creation
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "container456"}, nil)

	// Mock: Container start
	mockClient.On("ContainerStart", ctx, "container456", container.StartOptions{}).Return(nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	result, err := StartServe(ctx, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "container456", result.ContainerID)

	mockClient.AssertExpectations(t)

	// Clean up
	ReleasePort(result.HostPort)
}

// TestStartServe_PortAllocationFailure tests error when no ports available
func TestStartServe_PortAllocationFailure(t *testing.T) {
	ctx := context.Background()

	// Reserve all ports in range
	for port := PortRangeStart; port <= PortRangeEnd; port++ {
		ReservePort(port)
	}
	defer func() {
		// Clean up
		for port := PortRangeStart; port <= PortRangeEnd; port++ {
			ReleasePort(port)
		}
	}()

	config := &ServeConfig{
		Name:          "test-serve",
		Image:         "test:latest",
		ContentPath:   "/test/content",
		ContainerPath: "/data",
		ContainerPort: 8000,
		PreferredPort: 0,
	}

	// Execute
	result, err := StartServe(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to allocate port")
}

// TestStopServe_Success tests successful container stop
func TestStopServe_Success(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: Container exists
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:    "container123",
			Names: []string{"/test-serve-8000"},
		},
	}, nil)

	// Mock: Stop container
	mockClient.On("ContainerStop", ctx, "container123", mock.Anything).Return(nil)

	// Mock: Remove container
	mockClient.On("ContainerRemove", ctx, "container123", container.RemoveOptions{}).Return(nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	err := StopServe(ctx, "test-serve")

	// Assert
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestStopServe_ContainerNotFound tests error when container doesn't exist
func TestStopServe_ContainerNotFound(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: No containers
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{}, nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	err := StopServe(ctx, "nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no container found matching")
	mockClient.AssertExpectations(t)
}

// TestStopServe_StopError tests error during container stop
func TestStopServe_StopError(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: Container exists
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:    "container123",
			Names: []string{"/test-serve-8000"},
		},
	}, nil)

	// Mock: Stop fails
	mockClient.On("ContainerStop", ctx, "container123", mock.Anything).
		Return(errors.New("stop failed"))

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	err := StopServe(ctx, "test-serve")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stop container")
	mockClient.AssertExpectations(t)
}

// TestIsServing_ContainerRunning tests detecting running container
func TestIsServing_ContainerRunning(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: Running container
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == false // Only running containers
	})).Return([]types.Container{
		{
			ID:    "container123",
			Names: []string{"/test-serve-8000"},
			Ports: []types.Port{{PublicPort: 8000}},
		},
	}, nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	result, isRunning, err := IsServing(ctx, "test-serve")

	// Assert
	assert.NoError(t, err)
	assert.True(t, isRunning)
	assert.NotNil(t, result)
	assert.Equal(t, "container123", result.ContainerID)
	assert.Equal(t, "test-serve-8000", result.ContainerName)
	assert.Equal(t, 8000, result.HostPort)
	mockClient.AssertExpectations(t)
}

// TestIsServing_ContainerNotRunning tests when no container is running
func TestIsServing_ContainerNotRunning(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: No running containers
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == false
	})).Return([]types.Container{}, nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	result, isRunning, err := IsServing(ctx, "test-serve")

	// Assert
	assert.NoError(t, err)
	assert.False(t, isRunning)
	assert.Nil(t, result)
	mockClient.AssertExpectations(t)
}

// TestListServing_MultipleContainers tests listing multiple running containers
func TestListServing_MultipleContainers(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: Multiple running containers
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == false
	})).Return([]types.Container{
		{
			ID:    "container1",
			Names: []string{"/test-serve-8000"},
			Ports: []types.Port{{PublicPort: 8000}},
		},
		{
			ID:    "container2",
			Names: []string{"/test-serve-8001"},
			Ports: []types.Port{{PublicPort: 8001}},
		},
	}, nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	results, err := ListServing(ctx, "test-serve")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "container1", results[0].ContainerID)
	assert.Equal(t, "container2", results[1].ContainerID)
	mockClient.AssertExpectations(t)
}

// TestListServing_EmptyList tests when no containers are running
func TestListServing_EmptyList(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: No containers
	mockClient.On("ContainerList", ctx, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == false
	})).Return([]types.Container{}, nil)

	// Mock: Close
	mockClient.On("Close").Return(nil)

	// Override createDockerClient
	originalCreate := createDockerClient
	createDockerClient = func() (DockerClient, error) {
		return mockClient, nil
	}
	defer func() { createDockerClient = originalCreate }()

	// Execute
	results, err := ListServing(ctx, "test-serve")

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, results)
	mockClient.AssertExpectations(t)
}
