package config

import (
	"fmt"
)

// EnvironmentsConfig represents the environments.yml configuration
type EnvironmentsConfig struct {
	Metadata     Metadata      `yaml:"metadata"`
	Environments []Environment `yaml:"environments"`
}

// Environment represents a test execution environment
type Environment struct {
	Moniker     string   `yaml:"moniker"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Level       string   `yaml:"level"`       // L0, L1, L2, L3, L4
	Type        string   `yaml:"type"`        // unit, docker, docker-compose, plte, production
	EnvTags     []string `yaml:"env_tags"`    // Environment classification tags
	SystemDeps  []string `yaml:"system_deps"` // Required system dependencies (@deps:docker, etc.)
}

// GetTestTag returns the test tag for this environment (@env:<moniker>)
func (e *Environment) GetTestTag() string {
	return fmt.Sprintf("@env:%s", e.Moniker)
}

// GetEnvironment returns an environment by moniker
func (c *EnvironmentsConfig) GetEnvironment(moniker string) (*Environment, bool) {
	for i := range c.Environments {
		if c.Environments[i].Moniker == moniker {
			return &c.Environments[i], true
		}
	}
	return nil, false
}

// GetEnvironmentsByLevel returns all environments for a specific level
func (c *EnvironmentsConfig) GetEnvironmentsByLevel(level string) []Environment {
	var result []Environment
	for _, env := range c.Environments {
		if env.Level == level {
			result = append(result, env)
		}
	}
	return result
}

// GetEnvironmentsByType returns all environments of a specific type
func (c *EnvironmentsConfig) GetEnvironmentsByType(envType string) []Environment {
	var result []Environment
	for _, env := range c.Environments {
		if env.Type == envType {
			result = append(result, env)
		}
	}
	return result
}

// AllMonikers returns a list of all environment monikers
func (c *EnvironmentsConfig) AllMonikers() []string {
	monikers := make([]string, len(c.Environments))
	for i, env := range c.Environments {
		monikers[i] = env.Moniker
	}
	return monikers
}

// Validate validates the environment configuration
func (c *EnvironmentsConfig) Validate() error {
	if c.Metadata.Version == "" {
		return fmt.Errorf("metadata missing version")
	}

	if len(c.Environments) == 0 {
		return fmt.Errorf("no environments defined")
	}

	seen := make(map[string]bool)
	validLevels := map[string]bool{"L0": true, "L1": true, "L2": true, "L3": true, "L4": true}

	for _, env := range c.Environments {
		if env.Moniker == "" {
			return fmt.Errorf("environment missing moniker")
		}
		if seen[env.Moniker] {
			return fmt.Errorf("duplicate environment moniker: %s", env.Moniker)
		}
		seen[env.Moniker] = true

		if env.Name == "" {
			return fmt.Errorf("environment %s missing name", env.Moniker)
		}
		if env.Level == "" {
			return fmt.Errorf("environment %s missing level", env.Moniker)
		}
		if !validLevels[env.Level] {
			return fmt.Errorf("environment %s has invalid level: %s", env.Moniker, env.Level)
		}
		if env.Type == "" {
			return fmt.Errorf("environment %s missing type", env.Moniker)
		}
	}

	return nil
}

