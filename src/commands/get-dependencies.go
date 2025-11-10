// Command: get dependencies
// Description: Get module dependency graph in structured format
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --as-plantuml: Output as PlantUML diagram
//   --as-mermaid: Output as Mermaid diagram
//   --as-execution-order: Output execution order only
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/get"
	"github.com/ready-to-release/eac/src/internal/repository"
)

func init() {
	Register("get dependencies", GetDependencies)
}

func GetDependencies() int {
	// Check for special diagram formats
	for _, arg := range os.Args {
		if arg == "--as-plantuml" {
			return outputPlantUML()
		} else if arg == "--as-mermaid" {
			return outputMermaid()
		} else if arg == "--as-execution-order" {
			return outputExecutionOrder()
		}
	}

	// Use the shared get command helper for standard formats (YAML, JSON, TOML)
	return get.ExecuteGetCommand(func() (interface{}, error) {
		graph, err := repository.GetModuleDependencyGraph("../..", "0.1.0")
		if err != nil {
			return nil, err
		}
		return graph, nil
	})
}

// outputPlantUML generates PlantUML diagram format
func outputPlantUML() int {
	graph, err := repository.GetModuleDependencyGraph("../..", "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	diagram := repository.GetPlantUMLDiagram(graph)
	fmt.Print(diagram)
	return 0
}

// outputMermaid generates Mermaid diagram format
func outputMermaid() int {
	graph, err := repository.GetModuleDependencyGraph("../..", "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	diagram := repository.GetMermaidDiagram(graph)
	fmt.Print(diagram)
	return 0
}

// outputExecutionOrder generates execution order only
func outputExecutionOrder() int {
	plan, err := repository.CalculateExecutionOrder(nil, "../..", "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
		return get.ExecuteGetCommand(func() (interface{}, error) {
			return plan, nil
		})
	} else if hasYAML || !hasJSON {
		// Default: output as YAML
		return get.ExecuteGetCommand(func() (interface{}, error) {
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
