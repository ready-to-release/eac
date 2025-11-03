//go:build L0

package validator

import (
	"encoding/json"
	"testing"
)

// TestEmbeddedValidatorCreation tests that the embedded validator can be created
func TestEmbeddedValidatorCreation(t *testing.T) {
	v, err := NewEmbeddedValidator()
	if err != nil {
		t.Fatalf("Failed to create embedded validator: %v", err)
	}
	if v == nil {
		t.Fatal("Validator should not be nil")
	}
}

// TestValidateValidConfig tests validation of a valid configuration
func TestValidateValidConfig(t *testing.T) {
	v, err := NewEmbeddedValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Valid minimal configuration
	config := map[string]interface{}{
		"extensions": []map[string]interface{}{
			{
				"name":  "test-extension",
				"image": "test:latest",
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

	if !result.IsValid() {
		t.Errorf("Valid config should pass validation. Errors: %v", result.Errors)
	}
}

// TestValidateInvalidConfig tests validation of invalid configurations
func TestValidateInvalidConfig(t *testing.T) {
	v, err := NewEmbeddedValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	testCases := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
	}{
		{
			name:        "missing extensions",
			config:      map[string]interface{}{},
			expectError: true,
		},
		{
			name: "missing extension name",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"image": "test:latest",
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing extension image",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name": "test",
					},
				},
			},
			expectError: true,
		},
		{
			name: "invalid extension name pattern",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "Invalid_Name",
						"image": "test:latest",
					},
				},
			},
			expectError: true,
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
			expectError: true,
		},
		{
			name: "invalid environment variable name",
			config: map[string]interface{}{
				"extensions": []map[string]interface{}{
					{
						"name":  "test",
						"image": "test:latest",
						"env": []map[string]interface{}{
							{
								"name":  "invalid-var",
								"value": "test",
							},
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.config)
			if err != nil {
				t.Fatalf("Failed to marshal config: %v", err)
			}

			result, err := v.ValidateJSON(jsonData)
			if err != nil {
				t.Fatalf("Validation error: %v", err)
			}

			if tc.expectError && result.IsValid() {
				t.Error("Expected validation to fail but it passed")
			}
			if !tc.expectError && !result.IsValid() {
				t.Errorf("Expected validation to pass but it failed: %v", result.Errors)
			}
		})
	}
}

// TestValidateComplexConfig tests validation of a complex configuration
func TestValidateComplexConfig(t *testing.T) {
	v, err := NewEmbeddedValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Complex but valid configuration
	config := map[string]interface{}{
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
			"environment": []map[string]interface{}{
				{
					"name":  "DEFAULT_VAR",
					"value": "default",
				},
			},
		},
		"environment": map[string]interface{}{
			"global": []map[string]interface{}{
				{
					"name":  "GLOBAL_VAR",
					"value": "global",
				},
			},
			"secrets": []map[string]interface{}{
				{
					"name": "SECRET_VAR",
					"env":  "HOST_SECRET",
				},
			},
		},
		"extensions": []map[string]interface{}{
			{
				"name":                    "test-extension",
				"description":             "A test extension",
				"version":                 "1.0.0",
				"image":                   "test:1.0.0",
				"image_pull_policy":       "IfNotPresent",
				"repo_url":                "https://github.com/test/repo",
				"docs_url":                "https://docs.test.com",
				"metadata_schema_version": "1.0",
				"env": []map[string]interface{}{
					{
						"name":  "TEST_VAR",
						"value": "test_value",
					},
				},
				"volumes": []map[string]interface{}{
					{
						"host":      "./data",
						"container": "/data",
						"readonly":  true,
					},
				},
				"ports": []map[string]interface{}{
					{
						"host":      8080,
						"container": 3000,
					},
				},
				"working_dir":  "/app",
				"entrypoint":   []string{"/bin/sh"},
				"command":      []string{"--help"},
				"privileged":   true,
				"network_mode": "bridge",
				"memory_limit": "1g",
				"cpu_limit":    "1.0",
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

	if !result.IsValid() {
		t.Errorf("Complex valid config should pass validation. Errors:")
		for _, e := range result.Errors {
			t.Errorf("  - %s: %s", e.Field, e.Message)
		}
	}
}

// TestGetEmbeddedSchemaVersion tests schema version retrieval
func TestGetEmbeddedSchemaVersion(t *testing.T) {
	version := GetEmbeddedSchemaVersion()
	if version == "" || version == "unknown" {
		t.Error("Should return a valid schema version")
	}
	// We expect v1.0 for our current schema
	if version != "v1.0" {
		t.Errorf("Expected schema version v1.0, got %s", version)
	}
}

// TestGetEmbeddedSchema tests that we can retrieve the raw schema
func TestGetEmbeddedSchema(t *testing.T) {
	schema := GetEmbeddedSchema()
	if schema == "" {
		t.Fatal("Embedded schema should not be empty")
	}

	// Try to parse it as JSON to verify it's valid
	var schemaObj map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
		t.Fatalf("Embedded schema is not valid JSON: %v", err)
	}

	// Check for expected schema properties
	if _, ok := schemaObj["$schema"]; !ok {
		t.Error("Schema should have $schema property")
	}
	if _, ok := schemaObj["properties"]; !ok {
		t.Error("Schema should have properties")
	}
}
