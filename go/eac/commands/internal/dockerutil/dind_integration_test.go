//go:build L2 && (integration || dind)

package dockerutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDinDIntegration_PortAllocation tests port allocation in Docker-in-Docker environment
func TestDinDIntegration_PortAllocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DinD integration test in short mode")
	}

	ctx := context.Background()
	cli, err := createRealDockerClientDirect()
	require.NoError(t, err, "Docker must be available for DinD integration tests")
	defer cli.Close()

	// Verify Docker is accessible
	_, err = cli.Ping(ctx)
	if err != nil {
		t.Skipf("Docker daemon not accessible: %v", err)
	}

	// Test port allocation and reservation
	port1, err := FindAndReservePort()
	require.NoError(t, err)
	assert.True(t, port1 >= getPortRangeStart() && port1 <= getPortRangeEnd())

	port2, err := FindAndReservePort()
	require.NoError(t, err)
	assert.NotEqual(t, port1, port2, "Allocated ports should be different")

	// Clean up
	ReleasePort(port1)
	ReleasePort(port2)
}

// TestDinDIntegration_ImagePull tests pulling an image in DinD environment
func TestDinDIntegration_ImagePull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DinD integration test in short mode")
	}

	ctx := context.Background()
	cli, err := createRealDockerClientDirect()
	require.NoError(t, err)
	defer cli.Close()

	// Verify Docker is accessible
	_, err = cli.Ping(ctx)
	if err != nil {
		t.Skipf("Docker daemon not accessible: %v", err)
	}

	// Use a small, well-known image for testing
	testImage := "alpine:latest"

	// Ensure image doesn't exist locally (clean state)
	cli.ImageRemove(ctx, testImage, image.RemoveOptions{Force: true})

	// Pull the image
	reader, err := cli.ImagePull(ctx, testImage, image.PullOptions{})
	require.NoError(t, err)
	defer reader.Close()

	// Consume the pull output
	_, err = copyOutput(reader)
	require.NoError(t, err)

	// Verify image exists
	images, err := cli.ImageList(ctx, image.ListOptions{})
	require.NoError(t, err)

	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == testImage {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Image should be present after pull")

	// Clean up
	cli.ImageRemove(ctx, testImage, image.RemoveOptions{Force: true})
}

// TestDinDIntegration_ContainerLifecycle tests full container lifecycle in DinD
func TestDinDIntegration_ContainerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DinD integration test in short mode")
	}

	ctx := context.Background()
	cli, err := createRealDockerClientDirect()
	require.NoError(t, err)
	defer cli.Close()

	// Verify Docker is accessible
	_, err = cli.Ping(ctx)
	if err != nil {
		t.Skipf("Docker daemon not accessible: %v", err)
	}

	// Use alpine for quick container tests
	testImage := "alpine:latest"

	// Ensure image exists
	reader, err := cli.ImagePull(ctx, testImage, image.PullOptions{})
	if err == nil {
		copyOutput(reader)
		reader.Close()
	}

	// Create container
	containerName := fmt.Sprintf("dind-test-%d", time.Now().Unix())
	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: testImage,
			Cmd:   []string{"sleep", "10"},
		},
		&container.HostConfig{
			AutoRemove: false,
		},
		nil,
		nil,
		containerName,
	)
	require.NoError(t, err)
	containerID := resp.ID

	// Ensure cleanup
	defer func() {
		cli.ContainerStop(ctx, containerID, container.StopOptions{})
		cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	}()

	// Start container
	err = cli.ContainerStart(ctx, containerID, container.StartOptions{})
	require.NoError(t, err)

	// Verify container is running
	inspect, err := cli.ContainerInspect(ctx, containerID)
	require.NoError(t, err)
	assert.True(t, inspect.State.Running)

	// Stop container
	timeout := 5
	err = cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
	require.NoError(t, err)

	// Verify container is stopped
	inspect, err = cli.ContainerInspect(ctx, containerID)
	require.NoError(t, err)
	assert.False(t, inspect.State.Running)

	// Remove container
	err = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
	require.NoError(t, err)
}

// TestDinDIntegration_ConcurrentPortAllocation tests concurrent port allocation
func TestDinDIntegration_ConcurrentPortAllocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DinD integration test in short mode")
	}

	ctx := context.Background()
	cli, err := createRealDockerClientDirect()
	require.NoError(t, err)
	defer cli.Close()

	// Verify Docker is accessible
	_, err = cli.Ping(ctx)
	if err != nil {
		t.Skipf("Docker daemon not accessible: %v", err)
	}

	// Allocate multiple ports concurrently
	numPorts := 10
	ports := make(chan int, numPorts)
	errors := make(chan error, numPorts)

	for i := 0; i < numPorts; i++ {
		go func() {
			port, err := FindAndReservePort()
			if err != nil {
				errors <- err
				return
			}
			ports <- port
		}()
	}

	// Collect results
	allocatedPorts := make([]int, 0, numPorts)
	for i := 0; i < numPorts; i++ {
		select {
		case port := <-ports:
			allocatedPorts = append(allocatedPorts, port)
		case err := <-errors:
			t.Fatalf("Port allocation failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Port allocation timed out")
		}
	}

	// Verify all ports are unique
	portSet := make(map[int]bool)
	for _, port := range allocatedPorts {
		assert.False(t, portSet[port], "Port %d was allocated twice", port)
		portSet[port] = true
	}

	// Clean up
	for _, port := range allocatedPorts {
		ReleasePort(port)
	}
}

// TestDinDIntegration_ParallelImageOperations tests parallel image operations
func TestDinDIntegration_ParallelImageOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DinD integration test in short mode")
	}

	ctx := context.Background()

	// Create DockerClient using the interface
	dockerCli, err := NewDockerClient()
	require.NoError(t, err)
	defer dockerCli.Close()

	// Test image (using alpine for quick tests)
	testImage := "alpine:latest"

	// Clean up image first
	rawCli, err := createRealDockerClientDirect()
	require.NoError(t, err)
	defer rawCli.Close()

	rawCli.ImageRemove(ctx, testImage, image.RemoveOptions{Force: true})

	// Create image operations
	operations := []ImageOperation{
		{
			Config: &ServeConfig{
				Image: testImage,
			},
			Index: 0,
		},
	}

	// Execute parallel image ensure
	results := ParallelImageEnsure(ctx, dockerCli, operations, &ParallelImageEnsureOptions{
		MaxConcurrency: 3,
	})

	// Verify results
	require.Len(t, results, 1)
	assert.NoError(t, results[0].Error, "Image pull failed")
	assert.True(t, results[0].Success)

	// Verify image exists
	images, err := dockerCli.ImageList(ctx, image.ListOptions{})
	require.NoError(t, err)

	found := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == testImage {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Image %s should be present after pull", testImage)

	// Clean up
	rawCli.ImageRemove(ctx, testImage, image.RemoveOptions{Force: true})
}

// Helper to create a real Docker client directly (not wrapped in interface)
func createRealDockerClientDirect() (*client.Client, error) {
	// Check if running in CI or test environment
	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		return nil, fmt.Errorf("Docker tests disabled via SKIP_DOCKER_TESTS")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return cli, nil
}

// Helper to consume output from image pull
func copyOutput(reader io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	for {
		n, err := reader.Read(buf)
		total += int64(n)
		if err == io.EOF {
			return total, nil
		}
		if err != nil {
			return total, err
		}
	}
}
