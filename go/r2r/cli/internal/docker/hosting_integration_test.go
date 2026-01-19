//go:build L2
// +build L2

package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
)

// TestContainerHost_L2_Integration tests basic container host functionality
func TestContainerHost_L2_Integration(t *testing.T) {
	// Skip if Docker is not available
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skip("Docker client not available:", err)
	}
	defer cli.Close()

	// Check if Docker daemon is responsive
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		t.Skip("Docker daemon not responsive:", err)
	}

	// Test basic container host operations
	t.Run("container_host_creation", func(t *testing.T) {
		host, err := NewContainerHost()
		if err != nil {
			t.Fatalf("Failed to create container host: %v", err)
		}
		defer host.Close()

		// Basic validation that host was created successfully
		if host == nil {
			t.Fatal("Container host should not be nil")
		}
	})

}
