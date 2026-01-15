package conf

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
	"github.com/spf13/viper"
)

type EnvVar struct {
	Name     string `mapstructure:"name"`
	Value    string `mapstructure:"value"`    // If empty, pass through value from host env var with same name
	Required bool   `mapstructure:"required"` // If true, fail at runtime when passthrough env var is not set on host
}

type SecretVar struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type RegistryAuth struct {
	Required    bool   `mapstructure:"required"`
	UsernameEnv string `mapstructure:"username_env"`
	TokenEnv    string `mapstructure:"token_env"`
}

type Registry struct {
	Default        string        `mapstructure:"default"`
	Authentication *RegistryAuth `mapstructure:"authentication,omitempty"`
	Timeout        int           `mapstructure:"timeout"`
	RetryAttempts  int           `mapstructure:"retry_attempts"`
	CacheTTL       int           `mapstructure:"ghcr_cache_seconds"` // Default 300 (5 minutes)
}

type Environment struct {
	Global  []EnvVar    `mapstructure:"global,omitempty"`
	Secrets []SecretVar `mapstructure:"secrets,omitempty"`
}

type Defaults struct {
	Registry    string   `mapstructure:"registry"`
	PullPolicy  string   `mapstructure:"pull_policy"`
	RemoveAfter bool     `mapstructure:"remove_after"`
	Timeout     int      `mapstructure:"timeout"`
	MemoryLimit string   `mapstructure:"memory_limit"`
	CPULimit    string   `mapstructure:"cpu_limit"`
	Environment []EnvVar `mapstructure:"environment,omitempty"`
}

type VolumeMount struct {
	Host      string `mapstructure:"host"`
	Container string `mapstructure:"container"`
	Readonly  bool   `mapstructure:"readonly"`
}

type PortMapping struct {
	Host      int `mapstructure:"host"`
	Container int `mapstructure:"container"`
}

type Extension struct {
	Name                  string        `mapstructure:"name,omitempty"`
	Description           string        `mapstructure:"description,omitempty"`
	Version               string        `mapstructure:"version,omitempty"`
	Image                 string        `mapstructure:"image,omitempty"`
	ImagePullPolicy       string        `mapstructure:"image_pull_policy,omitempty"`
	LoadLocal             bool          `mapstructure:"load_local"`
	AutoRemoveChildren    bool          `mapstructure:"auto_remove_children"`
	RepoURL               string        `mapstructure:"repo_url,omitempty"`
	DocsURL               string        `mapstructure:"docs_url,omitempty"`
	Env                   []EnvVar      `mapstructure:"env,omitempty"`
	Volumes               []VolumeMount `mapstructure:"volumes,omitempty"`
	Ports                 []PortMapping `mapstructure:"ports,omitempty"`
	WorkingDir            string        `mapstructure:"working_dir,omitempty"`
	Entrypoint            []string      `mapstructure:"entrypoint,omitempty"`
	Command               []string      `mapstructure:"command,omitempty"`
	Privileged            bool          `mapstructure:"privileged"`
	NetworkMode           string        `mapstructure:"network_mode,omitempty"`
	MetadataSchemaVersion string        `mapstructure:"metadata_schema_version,omitempty"`
	MemoryLimit           string        `mapstructure:"memory_limit,omitempty"`
	CPULimit              string        `mapstructure:"cpu_limit,omitempty"`
}

type Config struct {
	Registry    *Registry    `mapstructure:"registry,omitempty"`
	Defaults    *Defaults    `mapstructure:"defaults,omitempty"`
	Environment *Environment `mapstructure:"environment,omitempty"`
	Extensions  []Extension  `mapstructure:"extensions,omitempty"`
	LoadLocal   bool         `mapstructure:"load_local"` // Global flag to use local development images
}

func (c *Config) GetExtensions() []Extension {
	// get the extensions from the config file
	return []Extension{}
}

var Global Config

// RootDir stores the repository root directory path (set during InitConfig)
var RootDir string

// configLoaded tracks whether the configuration has been loaded
var configLoaded bool

// ResetConfigLoaded resets the configLoaded flag (for testing)
func ResetConfigLoaded() {
	configLoaded = false
}

// ValidationError aggregates multiple validation errors
type ValidationError struct {
	Errors []string
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("configuration validation failed:\n  - %s", strings.Join(ve.Errors, "\n  - "))
}

func (ve *ValidationError) Add(err string) {
	ve.Errors = append(ve.Errors, err)
}

func (ve *ValidationError) HasErrors() bool {
	return len(ve.Errors) > 0
}

// validateConfig performs comprehensive validation of configuration
func LoadConfig(configFile string) error {
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return WrapConfigError(err, configFile)
	}

	if err := viper.Unmarshal(&Global); err != nil {
		return NewYAMLUnmarshalError(configFile, err)
	}

	if err := validateConfig(&Global); err != nil {
		return NewValidationError(configFile, err)
	}

	// Check for "latest" tags and log warnings only if not already loaded
	if !configLoaded {
		checkLatestTags(&Global)
		configLoaded = true
	}

	return nil
}

// MergeConfigFile merges an override configuration file into the existing Global config
func MergeConfigFile(configFile string) error {
	// Create a new viper instance for the override file
	overrideViper := viper.New()
	overrideViper.SetConfigFile(configFile)

	if err := overrideViper.ReadInConfig(); err != nil {
		return WrapConfigError(err, configFile)
	}

	// Create a temporary config to hold the override values
	var overrideConfig Config
	if err := overrideViper.Unmarshal(&overrideConfig); err != nil {
		return NewYAMLUnmarshalError(configFile, err)
	}

	// Don't validate the override config - it's allowed to have partial definitions
	// The validation will happen after merging with the base config
	logging.Debugf("Merging override configuration (validation skipped for partial config): file=%s", configFile)

	// Merge the override config into the Global config
	mergeConfigs(&Global, &overrideConfig)

	// Re-validate the merged configuration
	if err := validateConfig(&Global); err != nil {
		// Don't attribute the error to the override file - it's the merged config that failed
		logging.Errorf("Merged configuration validation failed: file=%s err=%v", configFile, err)
		return fmt.Errorf("merged configuration is invalid after applying %s: %w", configFile, err)
	}

	// Don't check for latest tags again - already done in LoadConfig
	// checkLatestTags(&Global)

	return nil
}

// mergeConfigs merges the override config into the base config
// Override config values take precedence over base config values
func mergeConfigs(base *Config, override *Config) {
	// Merge Registry settings
	if override.Registry != nil {
		if base.Registry == nil {
			base.Registry = override.Registry
		} else {
			if override.Registry.Default != "" {
				base.Registry.Default = override.Registry.Default
			}
			if override.Registry.Timeout != 0 {
				base.Registry.Timeout = override.Registry.Timeout
			}
			if override.Registry.RetryAttempts != 0 {
				base.Registry.RetryAttempts = override.Registry.RetryAttempts
			}
			if override.Registry.Authentication != nil {
				base.Registry.Authentication = override.Registry.Authentication
			}
		}
	}

	// Merge Defaults settings
	if override.Defaults != nil {
		if base.Defaults == nil {
			base.Defaults = override.Defaults
		} else {
			if override.Defaults.Registry != "" {
				base.Defaults.Registry = override.Defaults.Registry
			}
			if override.Defaults.PullPolicy != "" {
				base.Defaults.PullPolicy = override.Defaults.PullPolicy
			}
			if override.Defaults.RemoveAfter {
				base.Defaults.RemoveAfter = override.Defaults.RemoveAfter
			}
			if override.Defaults.Timeout != 0 {
				base.Defaults.Timeout = override.Defaults.Timeout
			}
			if override.Defaults.MemoryLimit != "" {
				base.Defaults.MemoryLimit = override.Defaults.MemoryLimit
			}
			if override.Defaults.CPULimit != "" {
				base.Defaults.CPULimit = override.Defaults.CPULimit
			}
			if len(override.Defaults.Environment) > 0 {
				base.Defaults.Environment = mergeEnvVars(base.Defaults.Environment, override.Defaults.Environment)
			}
		}
	}

	// Merge Environment settings
	if override.Environment != nil {
		if base.Environment == nil {
			base.Environment = override.Environment
		} else {
			if len(override.Environment.Global) > 0 {
				base.Environment.Global = mergeEnvVars(base.Environment.Global, override.Environment.Global)
			}
			if len(override.Environment.Secrets) > 0 {
				base.Environment.Secrets = mergeSecretVars(base.Environment.Secrets, override.Environment.Secrets)
			}
		}
	}

	// Merge Extensions - this is the most important part for the integration tests
	// Override extensions completely replace base extensions with the same name
	if len(override.Extensions) > 0 {
		// Create a map of base extensions for efficient lookup
		baseExtMap := make(map[string]*Extension)
		for i := range base.Extensions {
			baseExtMap[base.Extensions[i].Name] = &base.Extensions[i]
		}

		// Process override extensions
		for _, overrideExt := range override.Extensions {
			if existingExt, exists := baseExtMap[overrideExt.Name]; exists {
				// Merge the override fields into the existing extension
				logging.Debugf("Merging override extension with existing: extension=%s", overrideExt.Name)
				mergeExtension(existingExt, &overrideExt)
			} else {
				// Add new extension
				logging.Debugf("Adding new extension from override: extension=%s", overrideExt.Name)
				base.Extensions = append(base.Extensions, overrideExt)
			}
		}
	}
}

// mergeExtension merges override extension fields into the base extension
// Only non-empty/non-zero override fields replace base fields
func mergeExtension(base *Extension, override *Extension) {
	logging.Debugf("Merging extension details: base_name=%s base_image=%s override_name=%s override_image=%s override_load_local=%v",
		base.Name, base.Image, override.Name, override.Image, override.LoadLocal)

	// Only override non-empty string fields
	if override.Image != "" {
		logging.Debugf("Overriding image: old=%s new=%s", base.Image, override.Image)
		base.Image = override.Image
	}
	if override.Description != "" {
		base.Description = override.Description
	}
	if override.ImagePullPolicy != "" {
		base.ImagePullPolicy = override.ImagePullPolicy
	}

	// Override boolean fields only if explicitly set to true in override
	// This allows the override to set LoadLocal to true without forcing it to false
	if override.LoadLocal {
		base.LoadLocal = override.LoadLocal
	}

	// Merge environment variables
	if len(override.Env) > 0 {
		base.Env = mergeEnvVars(base.Env, override.Env)
	}

	// Override resource limits if specified
	if override.MemoryLimit != "" {
		base.MemoryLimit = override.MemoryLimit
	}
	if override.CPULimit != "" {
		base.CPULimit = override.CPULimit
	}

	// Override volumes if specified
	if len(override.Volumes) > 0 {
		base.Volumes = override.Volumes
	}

	// Override ports if specified
	if len(override.Ports) > 0 {
		base.Ports = override.Ports
	}

	// Override network mode if specified
	if override.NetworkMode != "" {
		base.NetworkMode = override.NetworkMode
	}

	// Override other fields if specified
	if len(override.Command) > 0 {
		base.Command = override.Command
	}
	if len(override.Entrypoint) > 0 {
		base.Entrypoint = override.Entrypoint
	}
	if override.WorkingDir != "" {
		base.WorkingDir = override.WorkingDir
	}
	if override.Privileged {
		base.Privileged = override.Privileged
	}
}

// mergeEnvVars merges environment variables, with override taking precedence
func mergeEnvVars(base []EnvVar, override []EnvVar) []EnvVar {
	envMap := make(map[string]string)

	// Add base environment variables
	for _, env := range base {
		envMap[env.Name] = env.Value
	}

	// Override with new values
	for _, env := range override {
		envMap[env.Name] = env.Value
	}

	// Convert back to slice
	result := make([]EnvVar, 0, len(envMap))
	for name, value := range envMap {
		result = append(result, EnvVar{Name: name, Value: value})
	}

	return result
}

// mergeSecretVars merges secret variables, with override taking precedence
func mergeSecretVars(base []SecretVar, override []SecretVar) []SecretVar {
	secretMap := make(map[string]string)

	// Add base secret variables
	for _, secret := range base {
		secretMap[secret.Name] = secret.Env
	}

	// Override with new values
	for _, secret := range override {
		secretMap[secret.Name] = secret.Env
	}

	// Convert back to slice
	result := make([]SecretVar, 0, len(secretMap))
	for name, env := range secretMap {
		result = append(result, SecretVar{Name: name, Env: env})
	}

	return result
}

// detectCIEnvironment checks if the CLI is running in a CI/CD environment
