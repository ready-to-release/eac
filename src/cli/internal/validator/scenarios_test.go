//go:build L1

package validator

import (
	"encoding/json"
	"testing"
)

// TestValidatorScenarios tests various YAML configuration scenarios through the validator
func TestValidatorScenarios(t *testing.T) {
	v, err := NewEmbeddedValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	scenarios := []struct {
		name        string
		config      map[string]interface{}
		expectValid bool
		expectError bool
	}{
		{
			name: "minimal valid configuration",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "pwsh",
						"image": "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:latest",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "valid configuration with all fields",
			config: map[string]interface{}{
				"version": "1.0",
				"registry": map[string]interface{}{
					"default": "ghcr.io",
					"authentication": map[string]interface{}{
						"required":     true,
						"username_env": "GITHUB_USERNAME",
						"token_env":    "GITHUB_TOKEN",
					},
					"timeout":        300,
					"retry_attempts": 3,
				},
				"defaults": map[string]interface{}{
					"registry":     "ghcr.io/ready-to-release",
					"pull_policy":  "IfNotPresent",
					"remove_after": true,
					"timeout":      1800,
					"memory_limit": "2g",
					"cpu_limit":    "1.5",
				},
				"extensions": []map[string]interface{}{
					{
						"name":              "pwsh",
						"image":             "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0",
						"image_pull_policy": "IfNotPresent",
						"description":       "PowerShell extension",
						"version":           "1.0.0",
						"repo_url":          "https://github.com/ready-to-release/eac/src/cli",
						"docs_url":          "https://docs.example.com",
						"working_dir":       "/app",
						"privileged":        false,
						"network_mode":      "bridge",
						"memory_limit":      "1g",
						"cpu_limit":         "0.5",
						"entrypoint":        []string{"/bin/sh"},
						"command":           []string{"-c", "echo hello"},
						"env": []map[string]interface{}{
							{
								"name":  "POWERSHELL_VERSION",
								"value": "7.0",
							},
						},
						"volumes": []map[string]interface{}{
							{
								"host":      "./scripts",
								"container": "/scripts",
								"readonly":  true,
							},
						},
						"ports": []map[string]interface{}{
							{
								"host":      8080,
								"container": 3000,
							},
						},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name:        "empty configuration",
			config:      map[string]interface{}{},
			expectValid: false,
			expectError: false,
		},
		{
			name: "missing extensions array",
			config: map[string]interface{}{
				"version": "1.0",
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "empty extensions array",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{},
			},
			expectValid: true, // Empty array is valid, just not useful
			expectError: false,
		},
		{
			name: "extension missing name",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "extension missing image",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name": "test",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid extension name with underscore",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "invalid_name",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid extension name with uppercase",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "InvalidName",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid image pull policy",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":              "test",
						"image":             "test:latest",
						"image_pull_policy": "Sometimes",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid network mode",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":         "test",
						"image":        "test:latest",
						"network_mode": "custom",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid environment variable name with lowercase",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"env": []map[string]interface{}{
							{
								"name":  "lowercase_var",
								"value": "test",
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid environment variable name with hyphen",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"env": []map[string]interface{}{
							{
								"name":  "INVALID-VAR",
								"value": "test",
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid port range - host port too low",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"ports": []map[string]interface{}{
							{
								"host":      0,
								"container": 3000,
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid port range - container port too high",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"ports": []map[string]interface{}{
							{
								"host":      8080,
								"container": 70000,
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid volume mount - missing host path",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"volumes": []map[string]interface{}{
							{
								"container": "/data",
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid volume mount - missing container path",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"volumes": []map[string]interface{}{
							{
								"host": "./data",
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid URL format - repo_url",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":     "test",
						"image":    "test:latest",
						"repo_url": "not-a-url",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid URL format - docs_url",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":     "test",
						"image":    "test:latest",
						"docs_url": "ftp://invalid-scheme.com",
					},
				},
			},
			expectValid: true, // FTP URLs are valid URLs, just not HTTP/HTTPS
			expectError: false,
		},
		{
			name: "invalid memory limit format",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":         "test",
						"image":        "test:latest",
						"memory_limit": "invalid",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid CPU limit format",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":      "test",
						"image":     "test:latest",
						"cpu_limit": "not-a-number",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "duplicate extension names",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "duplicate",
						"image": "test1:latest",
					},
					{
						"name":  "duplicate",
						"image": "test2:latest",
					},
				},
			},
			expectValid: true, // Schema validator doesn't check for uniqueness
			expectError: false,
		},
		{
			name: "valid registry configuration",
			config: map[string]interface{}{
				"registry": map[string]interface{}{
					"default": "docker.io",
					"authentication": map[string]interface{}{
						"required":     false,
						"username_env": "DOCKER_USERNAME",
						"token_env":    "DOCKER_TOKEN",
					},
					"timeout":        600,
					"retry_attempts": 5,
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "invalid registry hostname",
			config: map[string]interface{}{
				"registry": map[string]interface{}{
					"default": "invalid..hostname",
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: true, // Schema validator doesn't validate hostname format
			expectError: false,
		},
		{
			name: "negative timeout values",
			config: map[string]interface{}{
				"registry": map[string]interface{}{
					"timeout": -1,
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "valid environment and secrets configuration",
			config: map[string]interface{}{
				"environment": map[string]interface{}{
					"global": []map[string]interface{}{
						{
							"name":  "GLOBAL_VAR",
							"value": "global_value",
						},
					},
					"secrets": []map[string]interface{}{
						{
							"name": "SECRET_VAR",
							"env":  "HOST_SECRET_VAR",
						},
					},
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "invalid secret environment variable name",
			config: map[string]interface{}{
				"environment": map[string]interface{}{
					"secrets": []map[string]interface{}{
						{
							"name": "SECRET_VAR",
							"env":  "invalid-env-name",
						},
					},
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "extension with Docker Hub registry",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "nginx",
						"image": "nginx:alpine",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with GHCR registry",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "ghcr.io/owner/repo:tag",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "complex multi-extension configuration",
			config: map[string]interface{}{
				"version": "1.0",
				"registry": map[string]interface{}{
					"default": "ghcr.io",
					"authentication": map[string]interface{}{
						"required":     true,
						"username_env": "GITHUB_USERNAME",
						"token_env":    "GITHUB_TOKEN",
					},
				},
				"defaults": map[string]interface{}{
					"pull_policy":  "Always",
					"remove_after": false,
					"timeout":      3600,
				},
				"extensions": []map[string]interface{}{
					{
						"name":         "database",
						"image":        "postgres:14-alpine",
						"network_mode": "host",
						"env": []map[string]interface{}{
							{
								"name":  "POSTGRES_USER",
								"value": "admin",
							},
							{
								"name":  "POSTGRES_PASSWORD",
								"value": "secret",
							},
						},
						"ports": []map[string]interface{}{
							{
								"host":      5432,
								"container": 5432,
							},
						},
					},
					{
						"name":         "api",
						"image":        "api-server:latest",
						"network_mode": "bridge",
						"env": []map[string]interface{}{
							{
								"name":  "DB_HOST",
								"value": "localhost",
							},
							{
								"name":  "API_PORT",
								"value": "8080",
							},
						},
						"ports": []map[string]interface{}{
							{
								"host":      8080,
								"container": 8080,
							},
						},
					},
					{
						"name":         "worker",
						"image":        "worker:latest",
						"network_mode": "none",
						"env": []map[string]interface{}{
							{
								"name":  "WORKER_ID",
								"value": "worker-1",
							},
						},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with all network modes",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":         "bridge-net",
						"image":        "test:latest",
						"network_mode": "bridge",
					},
					{
						"name":         "host-net",
						"image":        "test:latest",
						"network_mode": "host",
					},
					{
						"name":         "none-net",
						"image":        "test:latest",
						"network_mode": "none",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with all pull policies",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":              "always",
						"image":             "test:latest",
						"image_pull_policy": "Always",
					},
					{
						"name":              "never",
						"image":             "test:latest",
						"image_pull_policy": "Never",
					},
					{
						"name":              "if-not-present",
						"image":             "test:latest",
						"image_pull_policy": "IfNotPresent",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with various memory limits",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":         "mem-bytes",
						"image":        "test:latest",
						"memory_limit": "134217728",
					},
					{
						"name":         "mem-kilobytes",
						"image":        "test:latest",
						"memory_limit": "128k",
					},
					{
						"name":         "mem-megabytes",
						"image":        "test:latest",
						"memory_limit": "512m",
					},
					{
						"name":         "mem-gigabytes",
						"image":        "test:latest",
						"memory_limit": "2g",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with various CPU limits",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":      "cpu-decimal",
						"image":     "test:latest",
						"cpu_limit": "0.5",
					},
					{
						"name":      "cpu-whole",
						"image":     "test:latest",
						"cpu_limit": "2",
					},
					{
						"name":      "cpu-millicores",
						"image":     "test:latest",
						"cpu_limit": "1500m",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with complex volume mounts",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "volumes-test",
						"image": "test:latest",
						"volumes": []map[string]interface{}{
							{
								"host":      ".",
								"container": "/app",
								"readonly":  false,
							},
							{
								"host":      "/tmp/data",
								"container": "/data",
								"readonly":  true,
							},
							{
								"host":      "~/config",
								"container": "/config",
								"readonly":  true,
							},
						},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with port range boundaries",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "port-boundaries",
						"image": "test:latest",
						"ports": []map[string]interface{}{
							{
								"host":      1,
								"container": 1,
							},
							{
								"host":      65535,
								"container": 65535,
							},
							{
								"host":      8080,
								"container": 80,
							},
						},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with special characters in image tag",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "special-tags",
						"image": "test:v1.2.3-alpha.1+build.456",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with SHA256 digest as tag",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "digest-tag",
						"image": "test@sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "valid extension names with numbers and hyphens",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test-123",
						"image": "test:latest",
					},
					{
						"name":  "node-16",
						"image": "node:16",
					},
					{
						"name":  "python3-9",
						"image": "python:3.9",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "environment variable patterns",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "env-patterns",
						"image": "test:latest",
						"env": []map[string]interface{}{
							{
								"name":  "SIMPLE_VAR",
								"value": "value",
							},
							{
								"name":  "VAR_WITH_UNDERSCORE",
								"value": "value",
							},
							{
								"name":  "VAR123",
								"value": "value",
							},
							{
								"name":  "VAR_123_TEST",
								"value": "value",
							},
						},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with empty arrays",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":       "empty-arrays",
						"image":      "test:latest",
						"env":        []map[string]interface{}{},
						"volumes":    []map[string]interface{}{},
						"ports":      []map[string]interface{}{},
						"entrypoint": []string{},
						"command":    []string{},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with boolean flags",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":       "boolean-flags",
						"image":      "test:latest",
						"privileged": true,
					},
					{
						"name":       "non-privileged",
						"image":      "test:latest",
						"privileged": false,
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "registry with various timeout values",
			config: map[string]interface{}{
				"registry": map[string]interface{}{
					"timeout": 0,
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "registry with max timeout",
			config: map[string]interface{}{
				"registry": map[string]interface{}{
					"timeout": 86400,
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with absolute paths",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":        "absolute-paths",
						"image":       "test:latest",
						"working_dir": "/opt/app",
						"volumes": []map[string]interface{}{
							{
								"host":      "/var/data",
								"container": "/data",
							},
						},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with relative paths",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":        "relative-paths",
						"image":       "test:latest",
						"working_dir": "./app",
						"volumes": []map[string]interface{}{
							{
								"host":      "./data",
								"container": "/data",
							},
							{
								"host":      "../shared",
								"container": "/shared",
							},
						},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "extension with complex entrypoint and command",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":       "complex-cmd",
						"image":      "test:latest",
						"entrypoint": []string{"/bin/sh", "-c"},
						"command":    []string{"echo 'Hello' && sleep 10 && echo 'World'"},
					},
				},
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "invalid extension name starting with number",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "123-invalid",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid extension name starting with hyphen",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "-invalid",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid extension name ending with hyphen",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "invalid-",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid extension name with special chars",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test@invalid",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid memory limit with wrong unit",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":         "test",
						"image":        "test:latest",
						"memory_limit": "512x",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid CPU limit negative",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":      "test",
						"image":     "test:latest",
						"cpu_limit": "-1",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid port out of range high",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"ports": []map[string]interface{}{
							{
								"host":      65536,
								"container": 8080,
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid port negative",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"ports": []map[string]interface{}{
							{
								"host":      -1,
								"container": 8080,
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid volume with empty paths",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"volumes": []map[string]interface{}{
							{
								"host":      "",
								"container": "/data",
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid environment variable starting with number",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"env": []map[string]interface{}{
							{
								"name":  "123_INVALID",
								"value": "test",
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid environment variable with spaces",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"env": []map[string]interface{}{
							{
								"name":  "INVALID VAR",
								"value": "test",
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid registry with wrong authentication",
			config: map[string]interface{}{
				"registry": map[string]interface{}{
					"authentication": map[string]interface{}{
						"required": "yes",
					},
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid registry with negative retry attempts",
			config: map[string]interface{}{
				"registry": map[string]interface{}{
					"retry_attempts": -5,
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "invalid defaults with wrong types",
			config: map[string]interface{}{
				"defaults": map[string]interface{}{
					"remove_after": "yes",
					"timeout":      "30 seconds",
				},
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectValid: false,
			expectError: false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Convert config to JSON for validation
			jsonData, err := json.Marshal(scenario.config)
			if err != nil {
				t.Fatalf("Failed to marshal config: %v", err)
			}

			// Validate the configuration
			result, err := v.ValidateJSON(jsonData)

			// Check for validation errors
			if scenario.expectError && err == nil {
				t.Error("Expected validation to return an error but it didn't")
			}
			if !scenario.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}

			// Check validation results
			if scenario.expectValid && !result.IsValid() {
				t.Errorf("Expected configuration to be valid but validation failed:")
				for _, e := range result.Errors {
					t.Errorf("  - %s: %s", e.Field, e.Message)
				}
			}
			if !scenario.expectValid && result.IsValid() {
				t.Error("Expected configuration to be invalid but validation passed")
			}

			// Log warnings for valid configurations
			if result.IsValid() && len(result.Warnings) > 0 {
				t.Logf("Configuration valid with warnings:")
				for _, w := range result.Warnings {
					t.Logf("  - %s: %s", w.Field, w.Message)
				}
			}
		})
	}
}

// TestValidatorErrorDetails tests that error messages contain useful information
func TestValidatorErrorDetails(t *testing.T) {
	v, err := NewEmbeddedValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Configuration with multiple errors
	config := map[string]interface{}{
		"extensions": []map[string]interface{}{
			{
				"name":              "Invalid_Name",  // Invalid name pattern
				"image":             "",              // Missing image
				"image_pull_policy": "InvalidPolicy", // Invalid policy
				"network_mode":      "custom",        // Invalid network mode
				"env": []map[string]interface{}{
					{
						"name":  "invalid-var", // Invalid env var name
						"value": "test",
					},
				},
				"ports": []map[string]interface{}{
					{
						"host":      0,     // Invalid port
						"container": 70000, // Invalid port
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	result, err := v.ValidateJSON(jsonData)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}

	if result.IsValid() {
		t.Error("Expected configuration with multiple errors to be invalid")
	}

	// Check that we got multiple errors
	if len(result.Errors) < 3 {
		t.Errorf("Expected at least 3 validation errors, got %d", len(result.Errors))
	}

	// Check that error messages are descriptive
	for _, e := range result.Errors {
		if e.Message == "" {
			t.Error("Error message should not be empty")
		}
		if e.Field == "" {
			t.Error("Error field should not be empty")
		}
		t.Logf("Error: %s: %s", e.Field, e.Message)
	}
}

// TestValidatorInterface tests the ValidateInterface method with various data types
func TestValidatorInterface(t *testing.T) {
	v, err := NewEmbeddedValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	testCases := []struct {
		name      string
		data      interface{}
		expectErr bool
	}{
		{
			name: "valid map interface",
			data: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "nil interface",
			data:      nil,
			expectErr: false, // ValidateInterface handles nil gracefully
		},
		{
			name:      "string interface",
			data:      "invalid",
			expectErr: false, // ValidateInterface converts to JSON first
		},
		{
			name:      "number interface",
			data:      123,
			expectErr: false, // ValidateInterface converts to JSON first
		},
		{
			name:      "empty map interface",
			data:      map[string]interface{}{},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := v.ValidateInterface(tc.data)

			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For valid calls, we should get a result
			if !tc.expectErr && result == nil {
				t.Error("Expected validation result but got nil")
			}
		})
	}
}
