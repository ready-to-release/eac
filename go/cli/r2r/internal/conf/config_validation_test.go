//go:build L2
// +build L2

package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateConfig_ValidExtension verifies valid extension passes
func TestValidateConfig_ValidExtension(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "test/image:v1.0.0",
			},
		},
	}

	err := validateConfig(cfg)
	assert.NoError(t, err)
}

// TestValidateConfig_MissingExtensionName verifies missing name fails
func TestValidateConfig_MissingExtensionName(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "",
				Image: "test/image:v1",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

// TestValidateConfig_MissingExtensionImage verifies missing image fails
func TestValidateConfig_MissingExtensionImage(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image is required")
}

// TestValidateConfig_DuplicateExtensionNames verifies duplicate names fail
func TestValidateConfig_DuplicateExtensionNames(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{Name: "test-ext", Image: "image1:v1"},
			{Name: "test-ext", Image: "image2:v1"},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate extension name")
}

// TestValidateConfig_InvalidImageReference verifies invalid image format fails
func TestValidateConfig_InvalidImageReference(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{Name: "test-ext", Image: "invalid image reference!"},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Docker image reference")
}

// TestValidateConfig_InvalidPullPolicy verifies invalid pull policy fails
func TestValidateConfig_InvalidPullPolicy(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:            "test-ext",
				Image:           "test/image:v1",
				ImagePullPolicy: "InvalidPolicy",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid imagePullPolicy")
}

// TestValidateConfig_ValidPullPolicies verifies all valid pull policies pass
func TestValidateConfig_ValidPullPolicies(t *testing.T) {
	validPolicies := []string{"Always", "IfNotPresent", "Never", "AutoDetect"}

	for _, policy := range validPolicies {
		cfg := &Config{
			Extensions: []Extension{
				{
					Name:            "test-ext",
					Image:           "test/image:v1",
					ImagePullPolicy: policy,
				},
			},
		}

		err := validateConfig(cfg)
		assert.NoError(t, err, "Policy %s should be valid", policy)
	}
}

// TestValidateConfig_InvalidRepoURL verifies invalid repo URL fails
func TestValidateConfig_InvalidRepoURL(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:    "test-ext",
				Image:   "test/image:v1",
				RepoURL: "not-a-valid-url",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid repo_url")
}

// TestValidateConfig_InvalidDocsURL verifies invalid docs URL fails
func TestValidateConfig_InvalidDocsURL(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:    "test-ext",
				Image:   "test/image:v1",
				DocsURL: "invalid-url",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid docs_url")
}

// TestValidateConfig_ValidURLs verifies valid URLs pass
func TestValidateConfig_ValidURLs(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:    "test-ext",
				Image:   "test/image:v1",
				RepoURL: "https://github.com/org/repo",
				DocsURL: "https://docs.example.com",
			},
		},
	}

	err := validateConfig(cfg)
	assert.NoError(t, err)
}

// TestValidateConfig_MissingEnvVarName verifies missing env var name fails
func TestValidateConfig_MissingEnvVarName(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "test/image:v1",
				Env: []EnvVar{
					{Name: "", Value: "value"},
				},
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

// TestValidateConfig_DuplicateEnvVarNames verifies duplicate env var names fail
func TestValidateConfig_DuplicateEnvVarNames(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "test/image:v1",
				Env: []EnvVar{
					{Name: "VAR1", Value: "value1"},
					{Name: "VAR1", Value: "value2"},
				},
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate environment variable")
}

// TestValidateConfig_InvalidEnvVarName verifies invalid env var name format fails
func TestValidateConfig_InvalidEnvVarName(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "test/image:v1",
				Env: []EnvVar{
					{Name: "lowercase_var", Value: "value"},
				},
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid environment variable name")
}

// TestValidateConfig_ValidEnvVarNames verifies valid env var names pass
func TestValidateConfig_ValidEnvVarNames(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "test/image:v1",
				Env: []EnvVar{
					{Name: "VAR1", Value: "value1"},
					{Name: "VAR_2", Value: "value2"},
					{Name: "VAR_NAME_123", Value: "value3"},
				},
			},
		},
	}

	err := validateConfig(cfg)
	assert.NoError(t, err)
}

// TestValidateConfig_InvalidMemoryLimit verifies invalid memory limit fails
func TestValidateConfig_InvalidMemoryLimit(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:        "test-ext",
				Image:       "test/image:v1",
				MemoryLimit: "invalid",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid memory limit")
}

// TestValidateConfig_ValidMemoryLimits verifies valid memory limits pass
func TestValidateConfig_ValidMemoryLimits(t *testing.T) {
	validLimits := []string{"512m", "1g", "2G", "256MB", "1024kb"}

	for _, limit := range validLimits {
		cfg := &Config{
			Extensions: []Extension{
				{
					Name:        "test-ext",
					Image:       "test/image:v1",
					MemoryLimit: limit,
				},
			},
		}

		err := validateConfig(cfg)
		assert.NoError(t, err, "Memory limit %s should be valid", limit)
	}
}

// TestValidateConfig_InvalidCPULimit verifies invalid CPU limit fails
func TestValidateConfig_InvalidCPULimit(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:     "test-ext",
				Image:    "test/image:v1",
				CPULimit: "invalid",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CPU limit")
}

// TestValidateConfig_ValidCPULimits verifies valid CPU limits pass
func TestValidateConfig_ValidCPULimits(t *testing.T) {
	validLimits := []string{"0.5", "1", "2", "1.5"}

	for _, limit := range validLimits {
		cfg := &Config{
			Extensions: []Extension{
				{
					Name:     "test-ext",
					Image:    "test/image:v1",
					CPULimit: limit,
				},
			},
		}

		err := validateConfig(cfg)
		assert.NoError(t, err, "CPU limit %s should be valid", limit)
	}
}

// TestValidateConfig_MissingVolumePaths verifies missing volume paths fail
func TestValidateConfig_MissingVolumePaths(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "test/image:v1",
				Volumes: []VolumeMount{
					{Host: "", Container: "/data"},
				},
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host path is required")
}

// TestValidateConfig_InvalidPortRanges verifies invalid port ranges fail
func TestValidateConfig_InvalidPortRanges(t *testing.T) {
	tests := []struct {
		name          string
		host          int
		container     int
		expectedError string
	}{
		{"host port too low", 0, 80, "host port must be between 1-65535"},
		{"host port too high", 70000, 80, "host port must be between 1-65535"},
		{"container port too low", 8080, 0, "container port must be between 1-65535"},
		{"container port too high", 8080, 70000, "container port must be between 1-65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Extensions: []Extension{
					{
						Name:  "test-ext",
						Image: "test/image:v1",
						Ports: []PortMapping{
							{Host: tt.host, Container: tt.container},
						},
					},
				},
			}

			err := validateConfig(cfg)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestValidateConfig_ValidPorts verifies valid port ranges pass
func TestValidateConfig_ValidPorts(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:  "test-ext",
				Image: "test/image:v1",
				Ports: []PortMapping{
					{Host: 8080, Container: 80},
					{Host: 443, Container: 443},
					{Host: 3000, Container: 3000},
				},
			},
		},
	}

	err := validateConfig(cfg)
	assert.NoError(t, err)
}

// TestValidateConfig_InvalidNetworkMode verifies invalid network mode fails
func TestValidateConfig_InvalidNetworkMode(t *testing.T) {
	cfg := &Config{
		Extensions: []Extension{
			{
				Name:        "test-ext",
				Image:       "test/image:v1",
				NetworkMode: "invalid",
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid network_mode")
}

// TestValidateConfig_ValidNetworkModes verifies all valid network modes pass
func TestValidateConfig_ValidNetworkModes(t *testing.T) {
	validModes := []string{"bridge", "host", "none"}

	for _, mode := range validModes {
		cfg := &Config{
			Extensions: []Extension{
				{
					Name:        "test-ext",
					Image:       "test/image:v1",
					NetworkMode: mode,
				},
			},
		}

		err := validateConfig(cfg)
		assert.NoError(t, err, "Network mode %s should be valid", mode)
	}
}

// TestValidateConfig_RegistryNegativeTimeout verifies negative timeout fails
func TestValidateConfig_RegistryNegativeTimeout(t *testing.T) {
	cfg := &Config{
		Registry: &Registry{
			Default: "ghcr.io",
			Timeout: -1,
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout: must be non-negative")
}

// TestValidateConfig_RegistryNegativeRetryAttempts verifies negative retries fail
func TestValidateConfig_RegistryNegativeRetryAttempts(t *testing.T) {
	cfg := &Config{
		Registry: &Registry{
			Default:       "ghcr.io",
			RetryAttempts: -1,
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry_attempts: must be non-negative")
}

// TestValidateConfig_InvalidRegistryHostname verifies invalid registry hostname fails
func TestValidateConfig_InvalidRegistryHostname(t *testing.T) {
	cfg := &Config{
		Registry: &Registry{
			Default: "invalid!hostname",
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hostname")
}

// TestValidateConfig_ValidRegistryAuth verifies valid registry auth passes
func TestValidateConfig_ValidRegistryAuth(t *testing.T) {
	cfg := &Config{
		Registry: &Registry{
			Default: "ghcr.io",
			Authentication: &RegistryAuth{
				UsernameEnv: "GITHUB_USERNAME",
				TokenEnv:    "GITHUB_TOKEN",
			},
		},
	}

	err := validateConfig(cfg)
	assert.NoError(t, err)
}

// TestValidateConfig_GlobalEnvVarsDuplicates verifies duplicate global env vars fail
func TestValidateConfig_GlobalEnvVarsDuplicates(t *testing.T) {
	cfg := &Config{
		Environment: &Environment{
			Global: []EnvVar{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR1", Value: "value2"},
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate environment variable")
}

// TestValidateConfig_SecretsMissingEnv verifies secret missing env fails
func TestValidateConfig_SecretsMissingEnv(t *testing.T) {
	cfg := &Config{
		Environment: &Environment{
			Secrets: []SecretVar{
				{Name: "MY_SECRET", Env: ""},
			},
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "env is required")
}

// TestValidateConfig_DefaultsInvalidPullPolicy verifies invalid defaults pull policy fails
func TestValidateConfig_DefaultsInvalidPullPolicy(t *testing.T) {
	cfg := &Config{
		Defaults: &Defaults{
			PullPolicy: "Invalid",
		},
	}

	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid policy")
}

// TestValidateConfig_DefaultsValidPullPolicies verifies all valid default pull policies pass
func TestValidateConfig_DefaultsValidPullPolicies(t *testing.T) {
	validPolicies := []string{"Always", "IfNotPresent", "Never"}

	for _, policy := range validPolicies {
		cfg := &Config{
			Defaults: &Defaults{
				PullPolicy: policy,
			},
		}

		err := validateConfig(cfg)
		assert.NoError(t, err, "Policy %s should be valid for defaults", policy)
	}
}

// TestValidateConfig_ComplexValidConfig verifies complex valid config passes
func TestValidateConfig_ComplexValidConfig(t *testing.T) {
	cfg := &Config{
		Registry: &Registry{
			Default:       "ghcr.io",
			Timeout:       30,
			RetryAttempts: 3,
		},
		Defaults: &Defaults{
			PullPolicy:  "IfNotPresent",
			Timeout:     120,
			MemoryLimit: "512m",
			CPULimit:    "1.0",
		},
		Environment: &Environment{
			Global: []EnvVar{
				{Name: "GLOBAL_VAR", Value: "value"},
			},
			Secrets: []SecretVar{
				{Name: "MY_SECRET", Env: "SECRET_TOKEN"},
			},
		},
		Extensions: []Extension{
			{
				Name:            "ext1",
				Image:           "ghcr.io/org/ext1:v1.0.0",
				ImagePullPolicy: "Always",
				RepoURL:         "https://github.com/org/ext1",
				DocsURL:         "https://docs.example.com/ext1",
				Env: []EnvVar{
					{Name: "EXT_VAR", Value: "ext_value"},
				},
				Volumes: []VolumeMount{
					{Host: "/host/path", Container: "/container/path"},
				},
				Ports: []PortMapping{
					{Host: 8080, Container: 80},
				},
				NetworkMode: "bridge",
				MemoryLimit: "1g",
				CPULimit:    "2.0",
			},
		},
	}

	err := validateConfig(cfg)
	assert.NoError(t, err)
}
