// Package specs contains helper functions for docs tests.
package specs

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

// waitForContainerReady polls container status until it's healthy or timeout.
// Uses exponential backoff starting at 100ms, capping at 500ms.
func waitForContainerReady(ctx context.Context, cli *client.Client, containerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	interval := 100 * time.Millisecond

	for time.Now().Before(deadline) {
		// Inspect container to check state
		inspect, err := cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}

		// Check if container is running
		if inspect.State.Running {
			// Container is running - consider it ready
			return nil
		}

		// Sleep before retry
		time.Sleep(interval)

		// Exponential backoff (cap at 500ms)
		interval = interval * 2
		if interval > 500*time.Millisecond {
			interval = 500 * time.Millisecond
		}
	}

	return fmt.Errorf("container did not become ready within %v", timeout)
}
