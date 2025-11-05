package custom

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func init() {
	// Register ONLY for get-modules command
	Register("count", RenderCount, []string{"get-modules"})
}

// RenderCount produces a simple count of modules
// This renderer is ONLY available for the "get modules" command
// Input: YAML bytes representing module contracts
// Output: Simple count message
func RenderCount(yamlBytes []byte) (string, error) {
	// Parse YAML to count modules
	var modules []map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &modules); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	return fmt.Sprintf("Total modules: %d\n", len(modules)), nil
}
