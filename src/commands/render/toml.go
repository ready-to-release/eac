package render

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// RenderAsTOML converts a Go struct to TOML by first marshaling to YAML then to TOML
// This ensures YAML is the single source of truth for serialization
// Order is preserved where possible (TOML spec requires tables to be at end)
//
// Example:
//
//	type Person struct {
//	    Name  string `yaml:"name"`
//	    Age   int    `yaml:"age"`
//	}
//	person := Person{Name: "Alice", Age: 30}
//	tomlStr, err := RenderAsTOML(person)
func RenderAsTOML(v interface{}) (string, error) {
	// First marshal to YAML to ensure consistent serialization
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Unmarshal YAML to generic interface
	var intermediate interface{}
	if err := yaml.Unmarshal(yamlBytes, &intermediate); err != nil {
		return "", fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Marshal to TOML
	tomlBytes, err := toml.Marshal(intermediate)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to TOML: %w", err)
	}

	return string(tomlBytes), nil
}

// RenderAsTOMLOrPanic is a convenience wrapper that panics on error
// Useful for cases where marshaling is guaranteed to succeed
func RenderAsTOMLOrPanic(v interface{}) string {
	result, err := RenderAsTOML(v)
	if err != nil {
		panic(err)
	}
	return result
}
