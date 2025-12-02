// Package environments provides environment contract management for test execution contexts
package environments

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/repository"
	"gopkg.in/yaml.v3"
)

// Metadata holds contract version and scope information
type Metadata struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Scope       string `yaml:"scope"`
}

// Environment represents a test execution environment
type Environment struct {
	Moniker     string   `yaml:"moniker"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Level       string   `yaml:"level"`       // L2, L3, L4
	Type        string   `yaml:"type"`        // docker, docker-compose, plte, production
	EnvTags     []string `yaml:"env_tags"`    // Environment classification tags (e.g., "local", "k8s", "staging")
	SystemDeps  []string `yaml:"system_deps"` // Required system dependencies (@deps:docker, @deps:kubectl, etc.)
}

// GetTestTag returns the test tag for this environment (@env:<moniker>)
func (e *Environment) GetTestTag() string {
	return fmt.Sprintf("@env:%s", e.Moniker)
}

// EnvironmentContract represents the complete environment system contract
type EnvironmentContract struct {
	Metadata     Metadata      `yaml:"metadata"`
	Environments []Environment `yaml:"environments"`
}

// LoadEnvironmentContract reads and parses the environment contract from the repository.
// It reads directly from .r2r/eac/environments.yml
func LoadEnvironmentContract() (*EnvironmentContract, error) {
	eacRoot, err := repository.GetRepoEACConfigRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to get EAC config root: %w", err)
	}

	envPath := filepath.Join(eacRoot, "environments.yml")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment contract from %s: %w", envPath, err)
	}

	var contract EnvironmentContract
	if err := yaml.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("failed to parse environment contract: %w", err)
	}

	return &contract, nil
}

// GetEnvironment returns a specific environment by moniker
func (c *EnvironmentContract) GetEnvironment(moniker string) (*Environment, error) {
	for _, env := range c.Environments {
		if env.Moniker == moniker {
			return &env, nil
		}
	}
	return nil, fmt.Errorf("environment not found: %s", moniker)
}

// GetEnvironmentsByLevel returns all environments for a specific level
func (c *EnvironmentContract) GetEnvironmentsByLevel(level string) []Environment {
	var envs []Environment
	for _, env := range c.Environments {
		if env.Level == level {
			envs = append(envs, env)
		}
	}
	return envs
}

// GetEnvironmentsByType returns all environments of a specific type
func (c *EnvironmentContract) GetEnvironmentsByType(envType string) []Environment {
	var envs []Environment
	for _, env := range c.Environments {
		if env.Type == envType {
			envs = append(envs, env)
		}
	}
	return envs
}

// ValidateContract validates the environment contract structure and data
func (c *EnvironmentContract) ValidateContract() error {
	if c.Metadata.Version == "" {
		return fmt.Errorf("contract metadata missing version")
	}

	if len(c.Environments) == 0 {
		return fmt.Errorf("contract contains no environments")
	}

	// Check for duplicate monikers
	seen := make(map[string]bool)
	for _, env := range c.Environments {
		if env.Moniker == "" {
			return fmt.Errorf("environment missing moniker")
		}
		if seen[env.Moniker] {
			return fmt.Errorf("duplicate environment moniker: %s", env.Moniker)
		}
		seen[env.Moniker] = true

		// Validate required fields
		if env.Name == "" {
			return fmt.Errorf("environment %s missing name", env.Moniker)
		}
		if env.Level == "" {
			return fmt.Errorf("environment %s missing level", env.Moniker)
		}
		if env.Type == "" {
			return fmt.Errorf("environment %s missing type", env.Moniker)
		}

		// Validate level values
		validLevels := map[string]bool{"L0": true, "L1": true, "L2": true, "L3": true, "L4": true}
		if !validLevels[env.Level] {
			return fmt.Errorf("environment %s has invalid level: %s (must be L0, L1, L2, L3, or L4)", env.Moniker, env.Level)
		}
	}

	return nil
}

// GetAllMonikers returns a list of all environment monikers
func (c *EnvironmentContract) GetAllMonikers() []string {
	monikers := make([]string, len(c.Environments))
	for i, env := range c.Environments {
		monikers[i] = env.Moniker
	}
	return monikers
}
