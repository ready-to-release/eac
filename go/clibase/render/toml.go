package render

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// RenderAsTOML converts a Go struct to TOML using direct TOML marshaling.
// If the struct has toml tags they are used; otherwise go-toml uses field names.
//
// Example:
//
//	type Person struct {
//	    Name  string `toml:"name" yaml:"name"`
//	    Age   int    `toml:"age"  yaml:"age"`
//	}
//	person := Person{Name: "Alice", Age: 30}
//	tomlStr, err := RenderAsTOML(person)
func RenderAsTOML(v interface{}) (string, error) {
	// Direct TOML marshaling (avoids YAML intermediate)
	tomlBytes, err := toml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to TOML: %w", err)
	}
	return string(tomlBytes), nil
}

// RenderAsTOMLViaYAML converts a Go struct to TOML via YAML intermediate.
// Use this only when the struct exclusively uses yaml tags.
func RenderAsTOMLViaYAML(v interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	var intermediate interface{}
	if err := yaml.Unmarshal(yamlBytes, &intermediate); err != nil {
		return "", fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	tomlBytes, err := toml.Marshal(intermediate)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to TOML: %w", err)
	}

	return string(tomlBytes), nil
}

// RenderAsTOMLOrPanic is a convenience wrapper that panics on error
// Useful for cases where marshaling is guaranteed to succeed.
func RenderAsTOMLOrPanic(v interface{}) string {
	result, err := RenderAsTOML(v)
	if err != nil {
		panic(err)
	}
	return result
}
