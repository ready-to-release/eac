//go:build L2
// +build L2

package docker

import (
	"context"
	"os"
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

	// Test with a real extension (simple test without building images)
	t.Run("real_extension_availability", func(t *testing.T) {
		// Check if we have GitHub auth for pulling images
		if os.Getenv("GITHUB_TOKEN") == "" {
			t.Skip("GitHub authentication not available")
		}

		host, err := NewContainerHost()
		if err != nil {
			t.Skip("Failed to create container host:", err)
		}
		defer host.Close()

		ext := &ExtensionConfig{
			Name:            "pwsh",
			Image:           "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:0.0.2",
			ImagePullPolicy: "IfNotPresent",
		}

		// Test metadata command (expected to fail since pwsh doesn't support it yet)
		_, err = host.ExecuteMetadataCommand(ext)
		if err == nil {
			t.Log("Extension unexpectedly supports metadata command")
		} else {
			t.Logf("Extension correctly does not support metadata command: %v", err)
		}
	})
}
