package custom

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func init() {
	// Register for all commands (wildcard)
	Register("summary", RenderSummary, []string{"*"})
}

// RenderSummary produces a human-readable summary of module contracts
// Input: YAML bytes representing module contracts
// Output: Text summary with counts and statistics
func RenderSummary(yamlBytes []byte) (string, error) {
	// Parse YAML to extract module information
	var modules []map[string]interface{}
	if err := yaml.Unmarshal(yamlBytes, &modules); err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Collect statistics
	totalModules := len(modules)
	modulesByType := make(map[string]int)
	modulesByParent := make(map[string]int)
	totalDependencies := 0

	for _, module := range modules {
		// Count by type
		if moduleType, ok := module["type"].(string); ok {
			modulesByType[moduleType]++
		}

		// Count by parent
		if parent, ok := module["parent"].(string); ok {
			modulesByParent[parent]++
		}

		// Count dependencies
		if deps, ok := module["depends_on"].([]interface{}); ok {
			totalDependencies += len(deps)
		}
	}

	// Build summary output
	var sb strings.Builder

	sb.WriteString("=== Module Contracts Summary ===\n")
	sb.WriteString(fmt.Sprintf("\nTotal Modules: %d\n", totalModules))
	sb.WriteString(fmt.Sprintf("Total Dependencies: %d\n", totalDependencies))

	// Module types breakdown
	sb.WriteString("\n--- Modules by Type ---\n")
	for moduleType, count := range modulesByType {
		if moduleType == "" {
			moduleType = "(no type)"
		}
		sb.WriteString(fmt.Sprintf("  %-25s %d\n", moduleType, count))
	}

	// Parent hierarchy breakdown
	sb.WriteString("\n--- Modules by Parent ---\n")
	for parent, count := range modulesByParent {
		if parent == "" || parent == "." {
			parent = "(root)"
		}
		sb.WriteString(fmt.Sprintf("  %-25s %d\n", parent, count))
	}

	// List all module monikers
	sb.WriteString("\n--- Module Monikers ---\n")
	for _, module := range modules {
		if moniker, ok := module["moniker"].(string); ok {
			name := ""
			if n, ok := module["name"].(string); ok {
				name = n
			}
			sb.WriteString(fmt.Sprintf("  • %s", moniker))
			if name != "" && name != moniker {
				sb.WriteString(fmt.Sprintf(" (%s)", name))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}
