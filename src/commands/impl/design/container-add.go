// Command: design add container
// Description: Add a container to a module's architecture
// Usage: design add container <module> <name> --tech <technology> --desc <description>
package design

import (
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/impl/design/internal/workspace"
)

func init() {
	registry.Register("design add container", DesignAddContainer)
}

// DesignAddContainer adds a container to an existing workspace
func DesignAddContainer() int {
	args := os.Args[4:] // Skip "go", "run", ".", "design", "add", "container"

	if len(args) < 2 {
		printDesignAddContainerUsage()
		return 1
	}

	module := args[0]
	containerName := args[1]
	var tech, description string

	// Parse flags
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--tech", "-t":
			if i+1 < len(args) {
				tech = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "Error: --tech requires a value\n")
				return 1
			}
		case "--description", "--desc", "-d":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "Error: --description requires a value\n")
				return 1
			}
		case "--help", "-h":
			printDesignAddContainerUsage()
			return 0
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n\n", args[i])
			printDesignAddContainerUsage()
			return 1
		}
	}

	// Validate required flags
	if tech == "" {
		fmt.Fprintf(os.Stderr, "Error: --tech is required\n\n")
		printDesignAddContainerUsage()
		return 1
	}

	if description == "" {
		fmt.Fprintf(os.Stderr, "Error: --description is required\n\n")
		printDesignAddContainerUsage()
		return 1
	}

	// Add container
	result, err := workspace.AddContainer(module, containerName, tech, description)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println(result)
	return 0
}

func printDesignAddContainerUsage() {
	fmt.Println("Add a container to a module's architecture")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . design add container <module> <name> --tech <technology> --desc <description>")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("  <module>              Module moniker (e.g., src-cli, src-commands)")
	fmt.Println("  <name>                Container name (will be sanitized to ID)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --tech, -t            Technology/platform (required, e.g., \"Go\", \"React\", \"PostgreSQL\")")
	fmt.Println("  --description, -d     Container purpose and responsibilities (required)")
	fmt.Println("  --help, -h            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run . design add container src-cli command_parser --tech \"Go\" --desc \"Parses CLI commands\"")
	fmt.Println("  go run . design add container src-cli executor -t \"Go\" -d \"Executes validated commands\"")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  Modifies: specs/<module>/design/workspace.dsl")
}
