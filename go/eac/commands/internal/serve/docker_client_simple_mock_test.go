package serve

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	nat "github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleMockDockerClient_ImageOperations(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Initially no images
	images, err := mock.ImageList(ctx, image.ListOptions{})
	require.NoError(t, err)
	assert.Empty(t, images)

	// Pull an image
	reader, err := mock.ImagePull(ctx, "test:latest", image.PullOptions{})
	require.NoError(t, err)
	require.NotNil(t, reader)
	reader.Close()

	// Now image should exist
	images, err = mock.ImageList(ctx, image.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, images, 1)
	assert.Equal(t, "test:latest", images[0].RepoTags[0])
}

func TestSimpleMockDockerClient_ContainerLifecycle(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     map[string]bool{"test:latest": true},
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Create container
	resp, err := mock.ContainerCreate(ctx,
		&container.Config{Image: "test:latest"},
		&container.HostConfig{},
		nil, nil, "test-container")
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	// Start container
	err = mock.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.NoError(t, err)

	// List running containers
	containers, err := mock.ContainerList(ctx, container.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, containers, 1)
	assert.Equal(t, "running", containers[0].State)

	// Stop container
	err = mock.ContainerStop(ctx, resp.ID, container.StopOptions{})
	require.NoError(t, err)

	// Remove container
	err = mock.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
	require.NoError(t, err)

	// Should be gone
	containers, err = mock.ContainerList(ctx, container.ListOptions{All: true})
	require.NoError(t, err)
	assert.Empty(t, containers)
}

func TestSimpleMockDockerClient_PortTracking(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     map[string]bool{"test:latest": true},
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Create container with port binding
	resp, err := mock.ContainerCreate(ctx,
		&container.Config{Image: "test:latest"},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostPort: "9001"}},
			},
		},
		nil, nil, "test-container")
	require.NoError(t, err)

	// Start container
	mock.ContainerStart(ctx, resp.ID, container.StartOptions{})

	// Verify port is tracked
	containers, err := mock.ContainerList(ctx, container.ListOptions{})
	require.NoError(t, err)
	require.Len(t, containers, 1)
	require.Len(t, containers[0].Ports, 1)
	assert.Equal(t, uint16(9001), containers[0].Ports[0].PublicPort)
}

func TestSimpleMockDockerClient_Wait(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	waitCh, errCh := mock.ContainerWait(ctx, "test-id", container.WaitConditionNotRunning)

	// Should receive success immediately
	resp, ok := <-waitCh
	require.True(t, ok, "wait channel should have a response")
	assert.Equal(t, int64(0), resp.StatusCode)

	// Check for errors
	err, ok := <-errCh
	assert.False(t, ok, "error channel should be closed")
	assert.Nil(t, err, "should not have an error")

	// Channels should be closed after reading
	_, waitOpen := <-waitCh
	assert.False(t, waitOpen, "wait channel should be closed")
}

func TestSimpleMockDockerClient_Close(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}

	// Close should not error
	err := mock.Close()
	assert.NoError(t, err)
}

func TestSimpleMockDockerClient_ConcurrentAccess(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Launch 100 concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Mix of operations
			name := fmt.Sprintf("container-%d", idx)
			resp, err := mock.ContainerCreate(ctx, &container.Config{}, &container.HostConfig{}, nil, nil, name)
			require.NoError(t, err)

			err = mock.ContainerStart(ctx, resp.ID, container.StartOptions{})
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all containers created
	containers, err := mock.ContainerList(ctx, container.ListOptions{All: true})
	require.NoError(t, err)
	assert.Len(t, containers, 100)
}

func TestSimpleMockDockerClient_NilMapGuard(t *testing.T) {
	// Instantiate without initializing maps (defensive guard test)
	mock := &SimpleMockDockerClient{}
	ctx := context.Background()

	// Operations should not panic due to nil maps
	_, err := mock.ImageList(ctx, image.ListOptions{})
	assert.NoError(t, err)

	_, err = mock.ImagePull(ctx, "test:latest", image.PullOptions{})
	assert.NoError(t, err)

	resp, err := mock.ContainerCreate(ctx, &container.Config{}, &container.HostConfig{}, nil, nil, "test")
	assert.NoError(t, err)

	err = mock.ContainerStart(ctx, resp.ID, container.StartOptions{})
	assert.NoError(t, err)
}

func TestSimpleMockDockerClient_InvalidPort(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Create container with invalid port binding
	_, err := mock.ContainerCreate(ctx,
		&container.Config{Image: "test:latest"},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"8080/tcp": []nat.PortBinding{{HostPort: "invalid"}},
			},
		},
		nil, nil, "test-container")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port format")
}

func TestSimpleMockDockerClient_ContainerNotFound(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Try to start non-existent container
	err := mock.ContainerStart(ctx, "nonexistent-id", container.StartOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")

	// Try to stop non-existent container
	err = mock.ContainerStop(ctx, "nonexistent-id", container.StopOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")

	// Try to remove non-existent container
	err = mock.ContainerRemove(ctx, "nonexistent-id", container.RemoveOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")
}

func TestSimpleMockDockerClient_AlreadyRunning(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Create and start container
	resp, err := mock.ContainerCreate(ctx, &container.Config{}, &container.HostConfig{}, nil, nil, "test")
	require.NoError(t, err)

	err = mock.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.NoError(t, err)

	// Try to start again
	err = mock.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestSimpleMockDockerClient_AlreadyStopped(t *testing.T) {
	mock := &SimpleMockDockerClient{
		images:     make(map[string]bool),
		containers: make(map[string]*mockContainerState),
	}
	ctx := context.Background()

	// Create, start, and stop container
	resp, err := mock.ContainerCreate(ctx, &container.Config{}, &container.HostConfig{}, nil, nil, "test")
	require.NoError(t, err)

	err = mock.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.NoError(t, err)

	err = mock.ContainerStop(ctx, resp.ID, container.StopOptions{})
	require.NoError(t, err)

	// Try to stop again
	err = mock.ContainerStop(ctx, resp.ID, container.StopOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already stopped")
}
