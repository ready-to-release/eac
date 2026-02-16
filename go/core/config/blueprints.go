package config

import (
	"fmt"
)

// BlueprintsConfig holds all blueprints loaded from defaults.
// Component kinds define tool bindings, default file patterns, and resources.
// Artifact matrices define reusable build configurations.
type BlueprintsConfig struct {
	// ComponentKinds maps component kind name to its definition (tool bindings, default patterns, resources).
	ComponentKinds map[string]*ComponentType `yaml:"component-kinds"`
	// ArtifactMatrices maps matrix name to its definition
	ArtifactMatrices map[string]*ArtifactMatrix `yaml:"artifact-matrices"`
}

// BlueprintsFileName is the config file for blueprints.
const BlueprintsFileName = "blueprints.yml"

// LoadBlueprintsDefaults loads blueprints from contract defaults.
func LoadBlueprintsDefaults(repoRoot string) (*BlueprintsConfig, error) {
	cfg, err := cloneBlueprintsDefaults()
	if err != nil {
		return nil, fmt.Errorf("loading blueprints defaults: %w", err)
	}
	return cfg, nil
}

// MergeBlueprintsConfig merges override into base. Override keys replace base keys.
// Either argument may be nil.
func MergeBlueprintsConfig(base, override *BlueprintsConfig) *BlueprintsConfig {
	if base == nil && override == nil {
		return &BlueprintsConfig{}
	}
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := &BlueprintsConfig{
		ComponentKinds:   make(map[string]*ComponentType),
		ArtifactMatrices: make(map[string]*ArtifactMatrix),
	}

	// Copy base component kinds, then override
	for k, v := range base.ComponentKinds {
		result.ComponentKinds[k] = v
	}
	for k, v := range override.ComponentKinds {
		result.ComponentKinds[k] = v
	}

	// Copy base artifact matrices, then override
	for k, v := range base.ArtifactMatrices {
		result.ArtifactMatrices[k] = v
	}
	for k, v := range override.ArtifactMatrices {
		result.ArtifactMatrices[k] = v
	}

	return result
}
