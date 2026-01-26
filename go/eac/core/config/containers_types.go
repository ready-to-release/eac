// containers_types.go defines container tool configuration types.
// These types support version-pinned container execution with mode-aware execution
// (build from Dockerfile locally, pull from registry in container mode).
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ContainersConfig holds all container tool configurations.
// The map keys are container names (e.g., "drawio-cli", "mermaid-cli").
// The special key "base_images" is reserved for base image version pins.
type ContainersConfig struct {
	// BaseImages maps base image names to pinned version tags.
	// These are used as ARG values when building Dockerfiles.
	// Example: {"python": "3.12.1", "node": "25.0.0"}
	BaseImages map[string]string `yaml:"base_images,omitempty"`

	// Tools maps container tool names to their configurations.
	// This uses inline YAML unmarshaling to allow tools at the same level as base_images.
	Tools map[string]*ContainerToolConfig `yaml:"-"`
}

// ContainerToolConfig defines a container-based tool.
type ContainerToolConfig struct {
	// Dockerfile is the path to Dockerfile relative to repo root.
	// If empty, the container can only be pulled from registry.
	Dockerfile string `yaml:"dockerfile,omitempty"`

	// Image is the GHCR image reference for container mode.
	// Example: "ghcr.io/ready-to-release/drawio-cli"
	Image string `yaml:"image,omitempty"`

	// Tag is the image tag for GHCR.
	// Default: "latest"
	Tag string `yaml:"tag,omitempty"`

	// Workdir is the working directory inside the container.
	// Default: "/workspace"
	Workdir string `yaml:"workdir,omitempty"`

	// Command is the default command/entrypoint.
	Command string `yaml:"command,omitempty"`

	// LocalBinding indicates whether to prefer build mode in local development.
	// When true (default), the container is built from Dockerfile in local mode.
	// When false, the container is always pulled from registry.
	LocalBinding *bool `yaml:"local_binding,omitempty"`

	// Description is a human-readable description of this container tool.
	Description string `yaml:"description,omitempty"`
}

// FullImage returns the complete image reference (image:tag).
// Returns empty string if Image is not set.
func (c *ContainerToolConfig) FullImage() string {
	if c == nil || c.Image == "" {
		return ""
	}
	tag := c.Tag
	if tag == "" {
		tag = "latest"
	}
	return c.Image + ":" + tag
}

// GetWorkdir returns workdir with default of "/workspace".
func (c *ContainerToolConfig) GetWorkdir() string {
	if c == nil || c.Workdir == "" {
		return "/workspace"
	}
	return c.Workdir
}

// GetLocalBinding returns the local_binding value with default of true.
func (c *ContainerToolConfig) GetLocalBinding() bool {
	if c == nil || c.LocalBinding == nil {
		return true
	}
	return *c.LocalBinding
}

// ShouldBuildLocally returns true if we should build from Dockerfile.
// Returns false if:
// - No dockerfile is configured (cannot build locally)
// - Running in container mode (R2R_DOCKER_MODE=true)
// - local_binding is explicitly set to false.
func (c *ContainerToolConfig) ShouldBuildLocally() bool {
	if c == nil {
		return false
	}

	// No dockerfile = cannot build locally
	if c.Dockerfile == "" {
		return false
	}

	// In container mode (r2r/eac), always pull from registry
	if IsContainerMode() {
		return false
	}

	// Local mode: use local_binding preference
	return c.GetLocalBinding()
}

// IsContainerMode returns true if running in container mode.
// This is detected by R2R_DOCKER_MODE=true environment variable.
func IsContainerMode() bool {
	return os.Getenv("R2R_DOCKER_MODE") == "true"
}

// UnmarshalYAML implements custom YAML unmarshaling for ContainersConfig.
// The containers section has "base_images" as a special key, and all other keys
// are container tool configurations.
func (c *ContainersConfig) UnmarshalYAML(value *yaml.Node) error {
	// Initialize maps
	c.BaseImages = make(map[string]string)
	c.Tools = make(map[string]*ContainerToolConfig)

	// The node should be a mapping
	if value.Kind != yaml.MappingNode {
		return nil
	}

	// Process key-value pairs
	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valueNode := value.Content[i+1]
		key := keyNode.Value

		if key == "base_images" {
			// Parse base_images as map[string]string
			if err := valueNode.Decode(&c.BaseImages); err != nil {
				return err
			}
		} else {
			// Parse as ContainerToolConfig
			var tool ContainerToolConfig
			if err := valueNode.Decode(&tool); err != nil {
				return err
			}
			c.Tools[key] = &tool
		}
	}

	return nil
}
