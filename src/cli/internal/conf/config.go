package conf

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/cli/internal/cache"
	"github.com/ready-to-release/eac/src/cli/internal/github"
	"github.com/ready-to-release/eac/src/cli/internal/session"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type EnvVar struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
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
func validateConfig(cfg *Config) error {
	validationErrors := &ValidationError{}
	extensionNames := make(map[string]bool)

	// Docker image reference regex pattern
	imagePattern := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*:[a-zA-Z0-9._-]+$|^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)

	for i, ext := range cfg.Extensions {
		extContext := fmt.Sprintf("extension[%d]", i)
		if ext.Name != "" {
			extContext = fmt.Sprintf("extension %q", ext.Name)
		}

		// Required field validation
		if ext.Name == "" {
			validationErrors.Add(fmt.Sprintf("%s: name is required", extContext))
		}
		if ext.Image == "" {
			validationErrors.Add(fmt.Sprintf("%s: image is required", extContext))
		}

		// Unique extension names
		if ext.Name != "" {
			if extensionNames[ext.Name] {
				validationErrors.Add(fmt.Sprintf("%s: duplicate extension name %q", extContext, ext.Name))
			}
			extensionNames[ext.Name] = true
		}

		// Docker image reference validation
		if ext.Image != "" && !imagePattern.MatchString(ext.Image) {
			validationErrors.Add(fmt.Sprintf("%s: invalid Docker image reference %q", extContext, ext.Image))
		}

		// ImagePullPolicy validation
		if ext.ImagePullPolicy != "" {
			validPolicies := []string{"Always", "IfNotPresent", "Never", "AutoDetect"}
			valid := false
			for _, policy := range validPolicies {
				if ext.ImagePullPolicy == policy {
					valid = true
					break
				}
			}
			if !valid {
				validationErrors.Add(fmt.Sprintf("%s: invalid imagePullPolicy %q, must be one of: %s",
					extContext, ext.ImagePullPolicy, strings.Join(validPolicies, ", ")))
			}
		}

		// URL validation
		if ext.RepoURL != "" {
			if parsedURL, err := url.Parse(ext.RepoURL); err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
				validationErrors.Add(fmt.Sprintf("%s: invalid repo_url %q: must be a valid URL with scheme and host", extContext, ext.RepoURL))
			}
		}
		if ext.DocsURL != "" {
			if parsedURL, err := url.Parse(ext.DocsURL); err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
				validationErrors.Add(fmt.Sprintf("%s: invalid docs_url %q: must be a valid URL with scheme and host", extContext, ext.DocsURL))
			}
		}

		// Environment variable validation
		envNames := make(map[string]bool)
		for j, envVar := range ext.Env {
			envContext := fmt.Sprintf("%s.env[%d]", extContext, j)

			if envVar.Name == "" {
				validationErrors.Add(fmt.Sprintf("%s: name is required", envContext))
			} else {
				// Check for duplicate env var names within extension
				if envNames[envVar.Name] {
					validationErrors.Add(fmt.Sprintf("%s: duplicate environment variable name %q", envContext, envVar.Name))
				}
				envNames[envVar.Name] = true

				// Validate environment variable name format
				if !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(envVar.Name) {
					validationErrors.Add(fmt.Sprintf("%s: invalid environment variable name %q, must be uppercase alphanumeric with underscores", envContext, envVar.Name))
				}
			}
		}

		// Resource limit validation
		if ext.MemoryLimit != "" {
			if err := validateMemoryLimit(ext.MemoryLimit); err != nil {
				validationErrors.Add(fmt.Sprintf("%s: %v", extContext, err))
			}
		}
		if ext.CPULimit != "" {
			if err := validateCPULimit(ext.CPULimit); err != nil {
				validationErrors.Add(fmt.Sprintf("%s: %v", extContext, err))
			}
		}

		// Volume mount validation
		for j, volume := range ext.Volumes {
			volumeContext := fmt.Sprintf("%s.volumes[%d]", extContext, j)
			if volume.Host == "" {
				validationErrors.Add(fmt.Sprintf("%s: host path is required", volumeContext))
			}
			if volume.Container == "" {
				validationErrors.Add(fmt.Sprintf("%s: container path is required", volumeContext))
			}
		}

		// Port mapping validation
		for j, port := range ext.Ports {
			portContext := fmt.Sprintf("%s.ports[%d]", extContext, j)
			if port.Host < 1 || port.Host > 65535 {
				validationErrors.Add(fmt.Sprintf("%s: host port must be between 1-65535, got %d", portContext, port.Host))
			}
			if port.Container < 1 || port.Container > 65535 {
				validationErrors.Add(fmt.Sprintf("%s: container port must be between 1-65535, got %d", portContext, port.Container))
			}
		}

		// Network mode validation
		if ext.NetworkMode != "" {
			validNetworkModes := []string{"bridge", "host", "none"}
			valid := false
			for _, mode := range validNetworkModes {
				if ext.NetworkMode == mode {
					valid = true
					break
				}
			}
			if !valid {
				validationErrors.Add(fmt.Sprintf("%s: invalid network_mode %q, must be one of: %s",
					extContext, ext.NetworkMode, strings.Join(validNetworkModes, ", ")))
			}
		}
	}

	// Registry configuration validation
	if cfg.Registry != nil {
		if cfg.Registry.Default != "" {
			// Validate hostname format
			if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*[a-zA-Z0-9]$`).MatchString(cfg.Registry.Default) {
				validationErrors.Add(fmt.Sprintf("registry.default: invalid hostname %q", cfg.Registry.Default))
			}
		}
		if cfg.Registry.Timeout < 0 {
			validationErrors.Add("registry.timeout: must be non-negative")
		}
		if cfg.Registry.RetryAttempts < 0 {
			validationErrors.Add("registry.retry_attempts: must be non-negative")
		}
		if cfg.Registry.Authentication != nil {
			auth := cfg.Registry.Authentication
			if auth.UsernameEnv != "" && !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(auth.UsernameEnv) {
				validationErrors.Add(fmt.Sprintf("registry.authentication.username_env: invalid environment variable name %q", auth.UsernameEnv))
			}
			if auth.TokenEnv != "" && !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(auth.TokenEnv) {
				validationErrors.Add(fmt.Sprintf("registry.authentication.token_env: invalid environment variable name %q", auth.TokenEnv))
			}
		}
	}

	// Environment configuration validation
	if cfg.Environment != nil {
		// Validate global environment variables
		globalVarNames := make(map[string]bool)
		for i, envVar := range cfg.Environment.Global {
			envContext := fmt.Sprintf("environment.global[%d]", i)
			if envVar.Name == "" {
				validationErrors.Add(fmt.Sprintf("%s: name is required", envContext))
			} else {
				if globalVarNames[envVar.Name] {
					validationErrors.Add(fmt.Sprintf("%s: duplicate environment variable name %q", envContext, envVar.Name))
				}
				globalVarNames[envVar.Name] = true
				if !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(envVar.Name) {
					validationErrors.Add(fmt.Sprintf("%s: invalid environment variable name %q", envContext, envVar.Name))
				}
			}
		}

		// Validate secret environment variables
		secretVarNames := make(map[string]bool)
		for i, secretVar := range cfg.Environment.Secrets {
			secretContext := fmt.Sprintf("environment.secrets[%d]", i)
			if secretVar.Name == "" {
				validationErrors.Add(fmt.Sprintf("%s: name is required", secretContext))
			} else {
				if secretVarNames[secretVar.Name] {
					validationErrors.Add(fmt.Sprintf("%s: duplicate secret variable name %q", secretContext, secretVar.Name))
				}
				secretVarNames[secretVar.Name] = true
				if !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(secretVar.Name) {
					validationErrors.Add(fmt.Sprintf("%s: invalid environment variable name %q", secretContext, secretVar.Name))
				}
			}
			if secretVar.Env == "" {
				validationErrors.Add(fmt.Sprintf("%s: env is required", secretContext))
			} else {
				if !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(secretVar.Env) {
					validationErrors.Add(fmt.Sprintf("%s: invalid host environment variable name %q", secretContext, secretVar.Env))
				}
			}
		}
	}

	// Defaults configuration validation
	if cfg.Defaults != nil {
		if cfg.Defaults.Registry != "" {
			// Validate registry prefix format
			if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-/]*[a-zA-Z0-9]$`).MatchString(cfg.Defaults.Registry) {
				validationErrors.Add(fmt.Sprintf("defaults.registry: invalid registry prefix %q", cfg.Defaults.Registry))
			}
		}
		if cfg.Defaults.PullPolicy != "" {
			validPolicies := []string{"Always", "IfNotPresent", "Never"}
			valid := false
			for _, policy := range validPolicies {
				if cfg.Defaults.PullPolicy == policy {
					valid = true
					break
				}
			}
			if !valid {
				validationErrors.Add(fmt.Sprintf("defaults.pull_policy: invalid policy %q, must be one of: %s",
					cfg.Defaults.PullPolicy, strings.Join(validPolicies, ", ")))
			}
		}
		if cfg.Defaults.Timeout < 0 {
			validationErrors.Add("defaults.timeout: must be non-negative")
		}
		if cfg.Defaults.MemoryLimit != "" {
			if err := validateMemoryLimit(cfg.Defaults.MemoryLimit); err != nil {
				validationErrors.Add(fmt.Sprintf("defaults.memory_limit: %v", err))
			}
		}
		if cfg.Defaults.CPULimit != "" {
			if err := validateCPULimit(cfg.Defaults.CPULimit); err != nil {
				validationErrors.Add(fmt.Sprintf("defaults.cpu_limit: %v", err))
			}
		}

		// Validate default environment variables
		defaultVarNames := make(map[string]bool)
		for i, envVar := range cfg.Defaults.Environment {
			envContext := fmt.Sprintf("defaults.environment[%d]", i)
			if envVar.Name == "" {
				validationErrors.Add(fmt.Sprintf("%s: name is required", envContext))
			} else {
				if defaultVarNames[envVar.Name] {
					validationErrors.Add(fmt.Sprintf("%s: duplicate environment variable name %q", envContext, envVar.Name))
				}
				defaultVarNames[envVar.Name] = true
				if !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(envVar.Name) {
					validationErrors.Add(fmt.Sprintf("%s: invalid environment variable name %q", envContext, envVar.Name))
				}
			}
		}
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}
	return nil
}

// LoadConfig takes a named config file and loads it using viper
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
	log.Debug().Str("file", configFile).Msg("Merging override configuration (validation skipped for partial config)")

	// Merge the override config into the Global config
	mergeConfigs(&Global, &overrideConfig)

	// Re-validate the merged configuration
	if err := validateConfig(&Global); err != nil {
		// Don't attribute the error to the override file - it's the merged config that failed
		log.Error().Str("file", configFile).Err(err).Msg("Merged configuration validation failed")
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
				log.Debug().Str("extension", overrideExt.Name).Msg("Merging override extension with existing")
				mergeExtension(existingExt, &overrideExt)
			} else {
				// Add new extension
				log.Debug().Str("extension", overrideExt.Name).Msg("Adding new extension from override")
				base.Extensions = append(base.Extensions, overrideExt)
			}
		}
	}
}

// mergeExtension merges override extension fields into the base extension
// Only non-empty/non-zero override fields replace base fields
func mergeExtension(base *Extension, override *Extension) {
	log.Debug().
		Str("base_name", base.Name).
		Str("base_image", base.Image).
		Str("override_name", override.Name).
		Str("override_image", override.Image).
		Bool("override_load_local", override.LoadLocal).
		Msg("Merging extension details")

	// Only override non-empty string fields
	if override.Image != "" {
		log.Debug().Str("old", base.Image).Str("new", override.Image).Msg("Overriding image")
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
func detectCIEnvironment() bool {
	// Common CI environment variables
	ciVars := []string{
		"CI",                     // Generic CI indicator (GitHub Actions, GitLab CI, CircleCI, etc.)
		"GITHUB_ACTIONS",         // GitHub Actions
		"GITLAB_CI",              // GitLab CI
		"JENKINS_URL",            // Jenkins
		"TEAMCITY_VERSION",       // TeamCity
		"BUILDKITE",              // Buildkite
		"DRONE",                  // Drone
		"TRAVIS",                 // Travis CI
		"CIRCLECI",               // CircleCI
		"APPVEYOR",               // AppVeyor
		"CODEBUILD_BUILD_ID",     // AWS CodeBuild
		"BITBUCKET_BUILD_NUMBER", // Bitbucket Pipelines
		"AZURE_PIPELINES",        // Azure DevOps
		"BUILD_ID",               // Various CI systems
		"CONTINUOUS_INTEGRATION", // Generic
	}

	for _, v := range ciVars {
		if val := os.Getenv(v); val != "" && val != "false" && val != "0" {
			return true
		}
	}

	return false
}

// ValidatePinnedExtensions validates that extensions use pinned tags in CI environments
// Returns an error if running in CI and extensions have unpinned tags
func ValidatePinnedExtensions(cfg *Config, isCI bool) ([]string, error) {
	// Load or create cache
	registryCache, _ := cache.Load()
	if registryCache == nil {
		// Create new empty cache if none exists
		registryCache, _ = cache.Load() // This returns an empty cache when file doesn't exist
	}

	// Determine cache TTL
	cacheTTL := 300 // default 5 minutes
	if cfg.Registry != nil && cfg.Registry.CacheTTL > 0 {
		cacheTTL = cfg.Registry.CacheTTL
	}

	// Check if we need to refresh cache
	needsRefresh := registryCache.IsExpired(cacheTTL)

	// Collect unpinned extensions for error reporting (CI) and warning suppression check
	var unpinnedExtensions []string

	for _, ext := range cfg.Extensions {
		if hasLatestTag(ext.Image) {
			// Extract the base image without tag
			baseImage := ext.Image
			if idx := strings.LastIndex(baseImage, ":"); idx > 0 {
				baseImage = baseImage[:idx]
			}

			// Extract extension name
			var extensionName string
			if strings.Contains(baseImage, "/extensions/") {
				parts := strings.Split(baseImage, "/")
				if len(parts) > 0 {
					extensionName = parts[len(parts)-1]
				}
			}

			// Get pinned version from cache or fetch if needed
			var pinnedVersion string

			if extensionName != "" {
				// Check cache first
				if !needsRefresh {
					if extCache, ok := registryCache.GetExtension(extensionName); ok {
						pinnedVersion = extCache.LatestSHA
					}
				}

				// If not in cache or cache expired, fetch from GHCR
				if pinnedVersion == "" {
					log.Debug().Str("extension", extensionName).Str("baseImage", baseImage).Msg("Fetching latest tags from GHCR")
					pinnedVersion = fetchAndCacheExtensionTags(baseImage, extensionName, registryCache)
					log.Debug().Str("extension", extensionName).Str("pinnedVersion", pinnedVersion).Msg("Fetched pin version")
				}
			}

			// Fallback if we couldn't get a version
			if pinnedVersion == "" {
				pinnedVersion = "sha-<unavailable>"
			}

			// Collect unpinned extensions
			unpinnedExtensions = append(unpinnedExtensions,
				fmt.Sprintf("'%s' must be pinned, latest is: %s", ext.Name, pinnedVersion))
		}
	}

	// Save updated cache
	log.Debug().
		Int("extensionsInCache", len(registryCache.Extensions)).
		Msg("Saving registry cache")
	if err := registryCache.Save(); err != nil {
		log.Error().Err(err).Msg("Failed to save registry cache")
	} else {
		log.Debug().Msg("Registry cache saved successfully")
	}

	// In CI, return error if any extensions are unpinned
	if isCI && len(unpinnedExtensions) > 0 {
		// Skip fatal error if we're in a test environment
		if os.Getenv("R2R_TESTING") == "true" {
			log.Error().Msgf("Extensions MUST be pinned in CI:\n  - %s", strings.Join(unpinnedExtensions, "\n  - "))
			return unpinnedExtensions, fmt.Errorf("extensions MUST be pinned in CI:\n  - %s", strings.Join(unpinnedExtensions, "\n  - "))
		}
		return unpinnedExtensions, fmt.Errorf("extensions MUST be pinned in CI:\n  - %s", strings.Join(unpinnedExtensions, "\n  - "))
	}

	return unpinnedExtensions, nil
}

// checkLatestTags checks for usage of "latest" Docker image tags and logs warnings
func checkLatestTags(cfg *Config) {
	// Check if running in CI environment
	isCI := detectCIEnvironment()

	// Determine if we should suppress warnings (but still update cache)
	suppressWarnings := false
	var sessionID string

	// Check warning suppression for non-CI environments
	if !isCI {
		// Check explicit suppression
		if os.Getenv("R2R_SKIP_PIN_WARNING") == "true" {
			suppressWarnings = true
		} else {
			// Get session identifier (parent process ID works across platforms)
			sessionID = session.GetIdentifier()

			// Check if warnings were shown recently (within last hour) for this session
			warningFile := filepath.Join(os.TempDir(), fmt.Sprintf("nncli-%s-warning-disabled", sessionID))
			if stat, err := os.Stat(warningFile); err == nil {
				// If file exists and was modified within the last hour, suppress warnings
				if time.Since(stat.ModTime()) < time.Hour {
					suppressWarnings = true
				}
			}
		}
	}
	// In CI, we NEVER suppress - always check and fail on unpinned extensions

	// Validate pinned extensions
	unpinnedExtensions, err := ValidatePinnedExtensions(cfg, isCI)

	// Handle validation result
	if err != nil {
		// In CI with unpinned extensions - fatal error (unless in test)
		if os.Getenv("R2R_TESTING") != "true" {
			log.Fatal().Err(err).Msg("")
		}
		return
	}

	// Show warnings in non-CI if not suppressed
	if !isCI && !suppressWarnings && len(unpinnedExtensions) > 0 {
		for _, msg := range unpinnedExtensions {
			log.Warn().Msg(msg)
		}

		// Create/touch warning file to indicate warnings were shown for this session
		if sessionID == "" {
			sessionID = session.GetIdentifier()
		}
		warningFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("nncli-%s-warning-disabled", sessionID))
		if file, err := os.Create(warningFilePath); err == nil {
			file.Close()
			log.Debug().Str("file", warningFilePath).Str("session", sessionID).Msg("Created warning flag file")
		}

		// Also set environment variable for current process tree
		os.Setenv("R2R_SKIP_PIN_WARNING", "true")
	}
}

// fetchAndCacheExtensionTags fetches tags from GHCR and updates cache
func fetchAndCacheExtensionTags(baseImage, extensionName string, registryCache *cache.RegistryCache) string {
	log.Debug().Str("baseImage", baseImage).Str("extensionName", extensionName).Msg("fetchAndCacheExtensionTags called")
	client, err := github.NewRegistryClient()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create registry client")
		return ""
	}

	// Get the latest stable tag
	latestTag, err := client.GetLatestStableTag(baseImage)
	if err != nil {
		log.Debug().Err(err).Str("baseImage", baseImage).Msg("Failed to get latest stable tag, trying any tag")
		// Try any tag
		latestTag, err = client.GetLatestTag(baseImage)
		if err != nil {
			log.Debug().Err(err).Str("baseImage", baseImage).Msg("Failed to get any tag")
			return ""
		}
	}
	log.Debug().Str("latestTag", latestTag).Msg("Got latest tag")

	// Get all tags for caching
	allTags, _ := client.ListTags(baseImage)

	// Update cache
	registryCache.SetExtension(extensionName, latestTag, allTags)
	log.Debug().
		Str("extensionName", extensionName).
		Str("latestTag", latestTag).
		Int("tagsCount", len(allTags)).
		Int("cacheExtensions", len(registryCache.Extensions)).
		Msg("Updated cache with extension data")

	return latestTag
}

// getActualImageVersion tries to suggest a proper version tag
func getActualImageVersion(image string) string {
	// Extract base image name without tag
	baseImage := image
	if idx := strings.LastIndex(image, ":"); idx > 0 {
		baseImage = baseImage[:idx]
	}

	// Extract extension name from image path (e.g., "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh" -> "pwsh")
	var extensionName string
	if strings.Contains(baseImage, "/extensions/") {
		parts := strings.Split(baseImage, "/")
		if len(parts) > 0 {
			extensionName = parts[len(parts)-1]
		}
	}

	// Determine cache TTL (default 300 seconds = 5 minutes)
	cacheTTL := 300
	if Global.Registry != nil && Global.Registry.CacheTTL > 0 {
		cacheTTL = Global.Registry.CacheTTL
	}

	// Try to use cached data first
	registryCache, _ := cache.Load()
	if registryCache != nil && extensionName != "" {
		if !registryCache.IsExpired(cacheTTL) {
			if latestSHA, ok := registryCache.GetLatestSHA(extensionName); ok {
				cachePath := cache.GetCachePath()
				log.Debug().
					Str("extension", extensionName).
					Str("cached_sha", latestSHA).
					Str("cache_path", cachePath).
					Int("cache_age_seconds", int(time.Since(registryCache.UpdatedAt).Seconds())).
					Msg("Using cached SHA tag from registry cache")
				return latestSHA
			}
		}
	}

	// Cache miss or expired, try to query GitHub registry
	client, err := github.NewRegistryClient()
	if err == nil {
		log.Debug().
			Str("image", baseImage).
			Msg("Cache miss or expired, querying GitHub Container Registry")

		// Try to get the latest stable tag from the registry
		latestTag, err := client.GetLatestStableTag(baseImage)
		if err == nil && latestTag != "" {
			// Update cache if we have an extension name
			if extensionName != "" && registryCache != nil {
				// Also get all tags for caching
				allTags, _ := client.ListTags(baseImage)
				registryCache.SetExtension(extensionName, latestTag, allTags)
				registryCache.Save()
			}

			log.Debug().
				Str("image", baseImage).
				Str("latest_tag", latestTag).
				Msg("Found latest stable tag from GitHub registry")
			return latestTag
		}

		// Fall back to any available tag
		anyTag, err := client.GetLatestTag(baseImage)
		if err == nil && anyTag != "" {
			// Update cache if we have an extension name
			if extensionName != "" && registryCache != nil {
				allTags, _ := client.ListTags(baseImage)
				registryCache.SetExtension(extensionName, anyTag, allTags)
				registryCache.Save()
			}

			log.Debug().
				Str("image", baseImage).
				Str("tag", anyTag).
				Msg("Found available tag from GitHub registry")
			return anyTag
		}
	}

	// Fall back to checking local Docker images if registry query fails
	cmd := exec.Command("docker", "images", baseImage, "--format", "{{.Tag}}")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\n")

		// Look for run-XXX tags first (stable releases)
		runTagPattern := regexp.MustCompile(`^run-\d+$`)
		var bestRunTag string
		var bestRunNum int

		for _, line := range lines {
			tag := strings.TrimSpace(line)
			if runTagPattern.MatchString(tag) {
				var num int
				fmt.Sscanf(tag, "run-%d", &num)
				if num > bestRunNum {
					bestRunNum = num
					bestRunTag = tag
				}
			}
		}

		if bestRunTag != "" {
			return bestRunTag
		}

		// Look for stable semantic version tags (e.g., 1.0.0, v1.2.3)
		for _, line := range lines {
			tag := strings.TrimSpace(line)
			if tag != "" && (regexp.MustCompile(`^v?\d+\.\d+\.\d+$`).MatchString(tag)) {
				return tag
			}
		}

		// Look for sha- tags (commit-based, stable)
		for _, line := range lines {
			tag := strings.TrimSpace(line)
			if tag != "" && strings.HasPrefix(tag, "sha-") {
				return tag
			}
		}
	}

	// For extension images, provide the specific GitHub packages URL if we couldn't get a tag
	if extensionName != "" {
		return fmt.Sprintf("run-XXX  # Check for stable versions at https://github.com/ready-to-release/eac/src/cli/pkgs/container/r2r-cli%%2Fextensions%%2F%s", extensionName)
	}

	// This shouldn't happen for our extension images, but handle gracefully
	return "run-XXX  # Check your registry for available stable version tags"
}

// hasLatestTag checks if a Docker image reference uses an unpinned tag (latest, main, master)
func hasLatestTag(image string) bool {
	if image == "" {
		return false
	}

	// Check for explicit unpinned tags
	if strings.HasSuffix(image, ":latest") || strings.HasSuffix(image, ":main") || strings.HasSuffix(image, ":master") {
		return true
	}

	// Check for implicit latest (no tag specified)
	// Docker images have format: [registry/]namespace/name[:tag]
	// If there's no colon after the last slash, it's implicit latest
	lastSlash := strings.LastIndex(image, "/")
	colonIndex := strings.LastIndex(image, ":")

	// If no colon, or colon appears before the last slash (part of registry URL), it's implicit latest
	if colonIndex == -1 || (lastSlash != -1 && colonIndex < lastSlash) {
		return true
	}

	return false
}

// validateMemoryLimit validates Docker memory limit format (e.g., "512MB", "1GB")
func validateMemoryLimit(limit string) error {
	if limit == "" {
		return nil // Optional field
	}

	// Match Docker memory limit format: number followed by unit (b, k, m, g)
	// Case insensitive, supports: b, k, m, g (bytes, kilobytes, megabytes, gigabytes)
	memoryPattern := regexp.MustCompile(`^(\d+(\.\d+)?)\s*([bBkKmMgG][bB]?)$`)
	if !memoryPattern.MatchString(limit) {
		return fmt.Errorf("invalid memory limit format %q: must be a number followed by unit (B, KB, MB, GB)", limit)
	}

	// Extract value and unit
	matches := memoryPattern.FindStringSubmatch(limit)
	value := matches[1]

	// Ensure value is positive
	if val, err := strconv.ParseFloat(value, 64); err != nil || val <= 0 {
		return fmt.Errorf("memory limit must be a positive number, got %q", value)
	}

	return nil
}

// validateCPULimit validates Docker CPU limit format (e.g., "0.5", "2")
func validateCPULimit(limit string) error {
	if limit == "" {
		return nil // Optional field
	}

	// Parse as float to ensure it's a valid decimal number
	val, err := strconv.ParseFloat(limit, 64)
	if err != nil {
		return fmt.Errorf("invalid CPU limit format %q: must be a decimal number", limit)
	}

	// Ensure value is positive
	if val <= 0 {
		return fmt.Errorf("CPU limit must be a positive number, got %v", val)
	}

	return nil
}
