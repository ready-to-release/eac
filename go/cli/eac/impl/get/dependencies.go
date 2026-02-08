// Command: get dependencies
// Short: Get full dependency graph for all modules
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//	--as-plantuml: Output as PlantUML diagram
//	--as-mermaid: Output as Mermaid diagram
//
// Long:
// Long: Expected Output:
// Long: YAML dependency graph showing module relationships, including:
// Long:   - List of all module monikers
// Long:   - Dependency edges (module -> list of dependencies)
// Long: Alternative formats available: PlantUML diagram, Mermaid diagram.
package get

import (
	"fmt"
	"os"

	getInternal "github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(GetDependencies)
}

// dependenciesFlags defines valid flags for the get dependencies command

func GetDependencies() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Check for special diagram formats
	for _, arg := range os.Args {
		switch arg {
		case "--as-plantuml":
			return outputPlantUML(workspaceRoot)
		case "--as-mermaid":
			return outputMermaid(workspaceRoot)
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

// outputPlantUML generates PlantUML diagram format.
func outputPlantUML(workspaceRoot string) int {
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	diagram := repository.GetPlantUMLDiagram(graph)
	fmt.Println(diagram)
	return 0
}

// outputMermaid generates Mermaid diagram format.
func outputMermaid(workspaceRoot string) int {
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	diagram := repository.GetMermaidDiagram(graph)
	fmt.Println(diagram)
	return 0
}



