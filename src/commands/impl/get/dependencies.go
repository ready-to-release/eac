// Command: get dependencies
// Description: Get module dependency graph in structured format
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --as-plantuml: Output as PlantUML diagram
//   --as-mermaid: Output as Mermaid diagram
//   --as-execution-order: Output execution order only
package get

import (
	"os"
	"strings"

	getInternal "github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(GetDependencies)
}

func GetDependencies() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Check for special diagram formats
	for _, arg := range os.Args {
		if arg == "--as-plantuml" {
			return outputPlantUML(workspaceRoot)
		} else if arg == "--as-mermaid" {
			return outputMermaid(workspaceRoot)
		} else if arg == "--as-execution-order" {
			return outputExecutionOrder(workspaceRoot)
		}
	}

	// Use the shared get command helper for standard formats (YAML, JSON, TOML)
	return getInternal.ExecuteGetCommand(func() (interface{}, error) {
		graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
		if err != nil {
			return nil, err
		}
		return graph, nil
	})
}

// outputPlantUML generates PlantUML diagram format
func outputPlantUML(workspaceRoot string) int {
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	diagram := repository.GetPlantUMLDiagram(graph)
	log.Info(diagram)
	return 0
}

// outputMermaid generates Mermaid diagram format
func outputMermaid(workspaceRoot string) int {
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	diagram := repository.GetMermaidDiagram(graph)
	log.Info(diagram)
	return 0
}

// outputExecutionOrder generates execution order only
func outputExecutionOrder(workspaceRoot string) int {
	plan, err := repository.CalculateExecutionOrder(nil, workspaceRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Check for nested format flags
	hasYAML := false
	hasJSON := false
	for _, arg := range os.Args {
		if arg == "--as-yaml" {
			hasYAML = true
		} else if arg == "--as-json" {
			hasJSON = true
		}
	}

	if hasJSON {
		// Output as JSON (handled by get command helper)
		return getInternal.ExecuteGetCommand(func() (interface{}, error) {
			return plan, nil
		})
	} else if hasYAML || !hasJSON {
		// Default: output as YAML
		return getInternal.ExecuteGetCommand(func() (interface{}, error) {
			return plan, nil
		})
	}

	return 0
}

// Helper to check if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val || strings.Contains(item, val) {
			return true
		}
	}
	return false
}
