package render

import (
	"fmt"

	"github.com/ready-to-release/eac/src/commands/render/custom"
	"gopkg.in/yaml.v3"
)

// RenderAsCustom converts a Go struct to a custom format using a registered renderer
// The struct is first marshaled to YAML, then passed to the custom renderer
// commandName should be in kebab-case format (e.g., "get-files")
//
// Example:
//
//	output, err := RenderAsCustom(myStruct, "table", "get-files")
func RenderAsCustom(v interface{}, rendererName string, commandName string) (string, error) {
	// First marshal to YAML (single source of truth)
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Get the custom renderer (with command filtering, expects kebab-case)
	renderer, err := custom.Get(rendererName, commandName)
	if err != nil {
		return "", err
	}

	// Apply the custom renderer
	result, err := renderer(yamlBytes)
	if err != nil {
		return "", fmt.Errorf("custom renderer %q failed: %w", rendererName, err)
	}

	return result, nil
}

// ListCustomRenderers returns custom renderer names that support the given command
// commandName should be in kebab-case format (e.g., "get-files")
// If commandName is empty, returns all renderers
func ListCustomRenderers(commandName string) []string {
	return custom.List(commandName)
}
