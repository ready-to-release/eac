package config

import (
	"fmt"
)

// EnvironmentsConfig represents the environments.yml configuration.
// Contains environment definitions and optional provider-specific sections.
type EnvironmentsConfig struct {
	Environments []Environment `yaml:"environments"`
	Azure        *AzureConfig  `yaml:"azure,omitempty"`
	AWS          *AWSConfig    `yaml:"aws,omitempty"`
	GCP          *GCPConfig    `yaml:"gcp,omitempty"`
}

// Environment represents an execution environment.
// Serves both as test isolation context AND deployment target.
type Environment struct {
	Moniker          string            `yaml:"moniker"`
	Name             string            `yaml:"name"`
	Description      string            `yaml:"description"`
	Level            string            `yaml:"level"`                        // L0, L1, L2, L3, L4
	Type             string            `yaml:"type"`                         // unit, docker, docker-compose, plte, production
	EnvironmentCode  string            `yaml:"environment_code,omitempty"`   // Short code for naming conventions (e.g., "d", "da", "p")
	Subscription     string            `yaml:"subscription,omitempty"`       // Name reference to provider subscription/account
	ApprovalRequired bool              `yaml:"approval_required,omitempty"`  // Gates production deployments
	Definition       map[string]string `yaml:"definition,omitempty"`         // Flexible per-environment metadata
	SystemDeps       []string          `yaml:"system_deps"`                  // Required system dependencies (docker, go, az, etc.)
}

// GetTestTag returns the test tag for this environment (@env:<moniker>).
func (e *Environment) GetTestTag() string {
	return fmt.Sprintf("@env:%s", e.Moniker)
}

// GetEnvironment returns an environment by moniker.
func (c *EnvironmentsConfig) GetEnvironment(moniker string) (*Environment, bool) {
	for i := range c.Environments {
		if c.Environments[i].Moniker == moniker {
			return &c.Environments[i], true
		}
	}
	return nil, false
}

// GetEnvironmentsByLevel returns all environments for a specific level.
func (c *EnvironmentsConfig) GetEnvironmentsByLevel(level string) []Environment {
	var result []Environment
	for _, env := range c.Environments {
		if env.Level == level {
			result = append(result, env)
		}
	}
	return result
}

// GetEnvironmentsByType returns all environments of a specific type.
func (c *EnvironmentsConfig) GetEnvironmentsByType(envType string) []Environment {
	var result []Environment
	for _, env := range c.Environments {
		if env.Type == envType {
			result = append(result, env)
		}
	}
	return result
}

// AllMonikers returns a list of all environment monikers.
func (c *EnvironmentsConfig) AllMonikers() []string {
	monikers := make([]string, len(c.Environments))
	for i, env := range c.Environments {
		monikers[i] = env.Moniker
	}
	return monikers
}

// IsDeployable returns true if this environment has deployment metadata.
func (e *Environment) IsDeployable() bool {
	return e.Subscription != "" || len(e.Definition) > 0
}

// GetDefinition returns a definition value by key, or empty string.
func (e *Environment) GetDefinition(key string) string {
	if e.Definition == nil {
		return ""
	}
	return e.Definition[key]
}

// ResolveSubscriptionID looks up the actual subscription ID from the azure config.
// The environment's Subscription field is a name reference, not an ID.
func (cfg *EnvironmentsConfig) ResolveSubscriptionID(env *Environment) string {
	if env.Subscription == "" || cfg.Azure == nil {
		return ""
	}
	for _, sub := range cfg.Azure.Subscriptions {
		if sub.Name == env.Subscription {
			return sub.ID
		}
	}
	return ""
}

// ResolveTenantID looks up the tenant ID for an environment's subscription.
func (cfg *EnvironmentsConfig) ResolveTenantID(env *Environment) string {
	if env.Subscription == "" || cfg.Azure == nil {
		return ""
	}
	for _, sub := range cfg.Azure.Subscriptions {
		if sub.Name == env.Subscription {
			// Find the tenant by name
			for _, t := range cfg.Azure.Tenants {
				if t.Name == sub.Tenant {
					return t.ID
				}
			}
		}
	}
	return ""
}

// AzureConfig holds normalized Azure infrastructure definitions.
type AzureConfig struct {
	Tenants       []AzureTenant       `yaml:"tenants"`
	Subscriptions []AzureSubscription `yaml:"subscriptions"`
}

// AzureTenant defines an Azure AD tenant.
type AzureTenant struct {
	Name   string `yaml:"name"`
	ID     string `yaml:"id"`
	Domain string `yaml:"domain,omitempty"`
}

// AzureSubscription defines an Azure subscription.
type AzureSubscription struct {
	Name         string   `yaml:"name"`
	ID           string   `yaml:"id"`
	Tenant       string   `yaml:"tenant"`       // references AzureTenant.Name
	Environments []string `yaml:"environments"` // which environment monikers use this
}

// AWSConfig holds normalized AWS infrastructure definitions.
type AWSConfig struct {
	Accounts []AWSAccount `yaml:"accounts"`
}

// AWSAccount defines an AWS account.
type AWSAccount struct {
	Name         string   `yaml:"name"`
	ID           string   `yaml:"id"`
	Region       string   `yaml:"region,omitempty"`
	Environments []string `yaml:"environments"`
}

// GCPConfig holds normalized GCP infrastructure definitions.
type GCPConfig struct {
	Projects []GCPProject `yaml:"projects"`
}

// GCPProject defines a GCP project.
type GCPProject struct {
	Name         string   `yaml:"name"`
	ID           string   `yaml:"id"`
	Region       string   `yaml:"region,omitempty"`
	Environments []string `yaml:"environments"`
}

// Validate validates the environment configuration.
func (c *EnvironmentsConfig) Validate() error {
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
