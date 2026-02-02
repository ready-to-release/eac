package conf

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

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

// LoadConfig takes a named config file and loads it using viper.
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

// validateCPULimit validates Docker CPU limit format (e.g., "0.5", "2").
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
