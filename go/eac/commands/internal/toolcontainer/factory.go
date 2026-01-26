package toolcontainer

import (
	"context"
	"fmt"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// New creates a new Runner for the specified container tool.
// It automatically selects the appropriate mode based on the tool configuration
// and runtime environment:
// - If running in container mode (R2R_DOCKER_MODE=true), uses ContainerRunner
// - If tool has local_binding=true and a dockerfile, uses BuildRunner
// - Otherwise uses ContainerRunner
func New(ctx context.Context, toolName string, repoConfig *config.RepositoryConfig, workspaceRoot string) (Runner, error) {
	toolCfg := repoConfig.GetContainerConfig(toolName)
	if toolCfg == nil {
		return nil, fmt.Errorf("container %q not configured in repository.yml containers section", toolName)
	}

	if toolCfg.ShouldBuildLocally() {
		// Local mode: build from Dockerfile
		return NewBuildRunner(ctx, toolName, toolCfg, repoConfig.Containers.BaseImages, workspaceRoot)
	}

	// Container mode: pull from registry
	return NewContainerRunner(ctx, toolName, toolCfg, workspaceRoot)
}

// NewWithMode creates a runner with a specific mode, ignoring the tool configuration.
// This is useful for testing or when you need to override the default behavior.
func NewWithMode(ctx context.Context, toolName string, mode Mode, repoConfig *config.RepositoryConfig, workspaceRoot string) (Runner, error) {
	toolCfg := repoConfig.GetContainerConfig(toolName)
	if toolCfg == nil {
		return nil, fmt.Errorf("container %q not configured in repository.yml containers section", toolName)
	}

	switch mode {
	case ModeLocal:
		if toolCfg.Dockerfile == "" {
			return nil, fmt.Errorf("container %q has no dockerfile configured, cannot use local mode", toolName)
		}
		return NewBuildRunner(ctx, toolName, toolCfg, repoConfig.Containers.BaseImages, workspaceRoot)
	case ModeContainer:
		if toolCfg.Image == "" {
			return nil, fmt.Errorf("container %q has no image configured, cannot use container mode", toolName)
		}
		return NewContainerRunner(ctx, toolName, toolCfg, workspaceRoot)
	default:
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}
}
