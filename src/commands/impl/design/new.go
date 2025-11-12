// Command: design new
// Description: Create a new architecture workspace for a module
// Usage: design new <module> --name <name> --description <description> [--force]
package design

import (
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/impl/design/internal/workspace"
)

func init() {
	registry.Register(DesignNew)
}

// DesignNew creates a new Structurizr workspace for a module
func DesignNew() int {
	args := os.Args[3:] // Skip "go", "run", ".", "design", "new"

	if len(args) == 0 {
		printDesignNewUsage()
		return 1
	}

	module := args[0]
	var name, description string
	var force bool

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name", "-n":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "Error: --name requires a value\n")
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
		case "--force", "-f":
			force = true
		case "--help", "-h":
			printDesignNewUsage()
			return 0
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n\n", args[i])
			printDesignNewUsage()
			return 1
		}
	}

	// Validate required flags
	if name == "" {
		fmt.Fprintf(os.Stderr, "Error: --name is required\n\n")
		printDesignNewUsage()
		return 1
	}

	if description == "" {
		fmt.Fprintf(os.Stderr, "Error: --description is required\n\n")
		printDesignNewUsage()
		return 1
	}

	// Create workspace
	result, err := workspace.Create(module, name, description, force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println(result)
	return 0
}

func printDesignNewUsage() {
	fmt.Println("Create a new architecture workspace for a module")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . design new <module> --name <name> --description <description> [--force]")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("  <module>              Module moniker (e.g., src-cli, src-commands)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --name, -n            Workspace name (required)")
	fmt.Println("  --description, -d     Workspace description (required)")
	fmt.Println("  --force, -f           Overwrite existing workspace")
	fmt.Println("  --help, -h            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run . design new src-cli --name \"CLI Architecture\" --description \"CLI system\"")
	fmt.Println("  go run . design new src-cli -n \"CLI\" -d \"CLI system\" --force")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  Creates: specs/<module>/design/workspace.dsl")
}
