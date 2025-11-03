package docker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/ready-to-release/eac/src/cli/internal/logger"
)

// IsRunningInContainer detects if we're running inside a Docker container
func IsRunningInContainer() bool {
	// Check for .dockerenv file (standard Docker indicator)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for Docker-specific cgroup entries
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		if len(data) > 0 {
			content := string(data)
			// Look for Docker or containerd indicators in cgroup
			if contains(content, "docker") || contains(content, "containerd") {
				return true
			}
		}
	}

	// Check for container-specific environment variables
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	return false
}

// CleanupChildContainers stops all containers that were started from within this container
// This is useful for Docker-in-Docker scenarios where the parent container starts child containers
func (ch *ContainerHost) CleanupChildContainers() error {
	ctx := context.Background()
	log := logger.Get()

	// Get our own container ID if we're running in a container
	containerID := os.Getenv("HOSTNAME") // In Docker, HOSTNAME is usually the container ID
	if containerID == "" {
		// Try to read from cgroup
		if data, err := os.ReadFile("/proc/self/cgroup"); err == nil {
			// Parse container ID from cgroup (format varies by Docker version)
			content := string(data)
			if id := extractContainerID(content); id != "" {
				containerID = id
			}
		}
	}

	log.WithField("parent_container", containerID).Debug().Msg("Checking for child containers to clean up")

	// List all running containers
	containers, err := ch.client.ContainerList(ctx, container.ListOptions{
		All: false, // Only running containers
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Track containers we need to stop
	var containersToStop []string

	for _, cont := range containers {
		// Skip our own container
		if cont.ID == containerID || cont.ID[:12] == containerID {
			continue
		}

		// Check if this container was started by r2r-cli (look for specific labels or naming patterns)
		// Containers started by Show-Documentation use pattern "mkdocs-show-*"
		for _, name := range cont.Names {
			if contains(name, "mkdocs-show-") || contains(name, "r2r-cli-") {
				containersToStop = append(containersToStop, cont.ID)
				log.WithFields(map[string]interface{}{
					"container_id":   cont.ID[:12],
					"container_name": name,
				}).Info().Msg("Found child container to clean up")
				break
			}
		}

		// Also check labels for r2r-cli managed containers
		if _, ok := cont.Labels["r2r-cli"]; ok {
			if !containsString(containersToStop, cont.ID) {
				containersToStop = append(containersToStop, cont.ID)
				log.WithField("container_id", cont.ID[:12]).Info().Msg("Found labeled container to clean up")
			}
		}
	}

	if len(containersToStop) == 0 {
		log.Debug().Msg("No child containers to clean up")
		return nil
	}

	log.WithField("count", len(containersToStop)).Info().Msg("Cleaning up child containers")

	// Stop all child containers gracefully
	stopTimeout := 5 * time.Second
	for _, id := range containersToStop {
		log.WithField("container_id", id[:12]).Debug().Msg("Stopping container")

		stopCtx, cancel := context.WithTimeout(ctx, stopTimeout)
		err := ch.client.ContainerStop(stopCtx, id, container.StopOptions{
			Timeout: &[]int{5}[0], // 5 second timeout
		})
		cancel()

		if err != nil {
			log.WithFields(map[string]interface{}{
				"container_id": id[:12],
				"error":        err.Error(),
			}).Warn().Msg("Failed to stop container gracefully")

			// Try force removal
			removeErr := ch.client.ContainerRemove(ctx, id, container.RemoveOptions{
				Force: true,
			})
			if removeErr != nil {
				log.WithFields(map[string]interface{}{
					"container_id": id[:12],
					"error":        removeErr.Error(),
				}).Error().Msg("Failed to force remove container")
			}
		} else {
			log.WithField("container_id", id[:12]).Info().Msg("Container stopped successfully")
		}
	}

	return nil
}

// CleanupOrphanedContainers removes containers that match r2r-cli patterns but are no longer needed
func (ch *ContainerHost) CleanupOrphanedContainers() error {
	ctx := context.Background()
	log := logger.Get()

	// Create filters for r2r-cli managed containers
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "r2r-cli")

	// List all containers (including stopped ones)
	containers, err := ch.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, cont := range containers {
		// Remove stopped containers
		if cont.State == "exited" || cont.State == "dead" {
			log.WithFields(map[string]interface{}{
				"container_id": cont.ID[:12],
				"state":        cont.State,
			}).Debug().Msg("Removing orphaned container")

			err := ch.client.ContainerRemove(ctx, cont.ID, container.RemoveOptions{
				Force: true,
			})
			if err != nil {
				log.WithFields(map[string]interface{}{
					"container_id": cont.ID[:12],
					"error":        err.Error(),
				}).Warn().Msg("Failed to remove orphaned container")
			}
		}
	}

	return nil
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// extractContainerID attempts to extract container ID from cgroup content
func extractContainerID(cgroupContent string) string {
	// Look for patterns like: /docker/<container-id> or /containerd/<container-id>
	lines := splitLines(cgroupContent)
	for _, line := range lines {
		if idx := lastIndex(line, "/docker/"); idx >= 0 {
			id := line[idx+8:] // Skip "/docker/"
			if len(id) >= 12 {
				return id[:12] // Return first 12 chars of container ID
			}
		}
		if idx := lastIndex(line, "/containerd/"); idx >= 0 {
			id := line[idx+12:] // Skip "/containerd/"
			if len(id) >= 12 {
				return id[:12]
			}
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func lastIndex(s, substr string) int {
	if len(substr) == 0 || len(substr) > len(s) {
		return -1
	}
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
