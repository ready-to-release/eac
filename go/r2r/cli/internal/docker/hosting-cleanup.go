package docker

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
)

// ContainerCleanupOptions configures container cleanup behavior.
type ContainerCleanupOptions struct {
	OnlyExtensions bool          // Only clean extension containers
	IncludeRunning bool          // Also remove running containers (default: false)
	OlderThan      time.Duration // Only remove containers older than this duration
	DryRun         bool          // Show what would be removed without removing
}

// CleanupResult represents the result of a cleanup operation.
type CleanupResult struct {
	ContainersRemoved int
	SpaceReclaimed    int64
	Errors            []error
}

// CleanupContainers removes containers based on the provided options.
func (ch *ContainerHost) CleanupContainers(opts ContainerCleanupOptions) (*CleanupResult, error) {
	result := &CleanupResult{}

	// List all containers (including stopped)
	containers, err := ch.client.ContainerList(ch.ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	logging.Debugf("Scanning containers for cleanup: total_containers=%d", len(containers))

	for _, ctr := range containers {
		// Skip if container is running and we don't want to include running containers
		if ctr.State == "running" && !opts.IncludeRunning {
			continue
		}

		// Filter by extension containers if requested
		if opts.OnlyExtensions {
			// Check if container name or labels indicate it's an extension container
			isExtension := false
			for _, name := range ctr.Names {
				if strings.Contains(name, "r2r-") || strings.Contains(name, "extension-") {
					isExtension = true
					break
				}
			}
			if !isExtension {
				continue
			}
		}

		// Filter by age if specified
		if opts.OlderThan > 0 {
			createdTime := time.Unix(ctr.Created, 0)
			age := time.Since(createdTime)
			if age < opts.OlderThan {
				continue
			}
		}

		// Container matches criteria, remove it
		containerName := ctr.Names[0]
		if containerName != "" && containerName[0] == '/' {
			containerName = containerName[1:] // Remove leading slash
		}

		if opts.DryRun {
			logging.Infof("[DRY RUN] Would remove container: container=%s state=%s image=%s", containerName, ctr.State, ctr.Image)
			result.ContainersRemoved++
		} else {
			logging.Infof("Removing container: container=%s state=%s image=%s", containerName, ctr.State, ctr.Image)

			// Get container size before removal
			var containerSize int64
			inspectData, inspectErr := ch.client.ContainerInspect(ch.ctx, ctr.ID)
			if inspectErr == nil && inspectData.SizeRw != nil {
				containerSize = *inspectData.SizeRw
			}

			// Remove the container
			removeOpts := container.RemoveOptions{
				Force:         opts.IncludeRunning, // Force remove if including running containers
				RemoveVolumes: true,                // Also remove associated volumes
			}

			if err := ch.client.ContainerRemove(ch.ctx, ctr.ID, removeOpts); err != nil {
				logging.Warnf("Failed to remove container: container=%s err=%v", containerName, err)
				result.Errors = append(result.Errors, fmt.Errorf("failed to remove %s: %w", containerName, err))
			} else {
				result.ContainersRemoved++
				result.SpaceReclaimed += containerSize // Only count space for successful removals
			}
		}
	}

	logging.Infof("Container cleanup completed: removed=%d space_reclaimed_bytes=%d errors=%d dry_run=%v", result.ContainersRemoved, result.SpaceReclaimed, len(result.Errors), opts.DryRun)

	return result, nil
}

// CleanupStoppedContainers is a convenience wrapper for cleaning up all stopped containers.
func (ch *ContainerHost) CleanupStoppedContainers(dryRun bool) (*CleanupResult, error) {
	return ch.CleanupContainers(ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		DryRun:         dryRun,
	})
}

// CleanupExtensionContainers is a convenience wrapper for cleaning up extension containers.
func (ch *ContainerHost) CleanupExtensionContainers(includeRunning, dryRun bool) (*CleanupResult, error) {
	return ch.CleanupContainers(ContainerCleanupOptions{
		OnlyExtensions: true,
		IncludeRunning: includeRunning,
		DryRun:         dryRun,
	})
}

// CleanupOldContainers removes containers older than the specified duration.
func (ch *ContainerHost) CleanupOldContainers(olderThan time.Duration, dryRun bool) (*CleanupResult, error) {
	return ch.CleanupContainers(ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		OlderThan:      olderThan,
		DryRun:         dryRun,
	})
}
