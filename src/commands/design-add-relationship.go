// Command: design add relationship
// Description: Add a relationship between containers in a module's architecture
// Usage: design add relationship <module> <source> <destination> --desc <description> [--tech <technology>]
package main

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/design/workspace"
)

func init() {
	Register("design add relationship", DesignAddRelationship)
}

// DesignAddRelationship adds a relationship to an existing workspace
func DesignAddRelationship() int {
	args := os.Args[4:] // Skip "go", "run", ".", "design", "add", "relationship"

	if len(args) < 3 {
		printDesignAddRelationshipUsage()
		return 1
	}

	module := args[0]
	source := args[1]
	destination := args[2]
	var description, tech string

	// Parse flags
	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "--description", "--desc", "-d":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "Error: --description requires a value\n")
				return 1
			}
		case "--tech", "-t":
			if i+1 < len(args) {
				tech = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "Error: --tech requires a value\n")
				return 1
			}
		case "--help", "-h":
			printDesignAddRelationshipUsage()
			return 0
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n\n", args[i])
			printDesignAddRelationshipUsage()
			return 1
		}
	}

	// Validate required flags
	if description == "" {
		fmt.Fprintf(os.Stderr, "Error: --description is required\n\n")
		printDesignAddRelationshipUsage()
		return 1
	}

	// Add relationship
	result, err := workspace.AddRelationship(module, source, destination, description, tech)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println(result)
	return 0
}

func printDesignAddRelationshipUsage() {
	fmt.Println("Add a relationship between containers in a module's architecture")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . design add relationship <module> <source> <destination> --desc <description> [--tech <technology>]")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("  <module>              Module moniker (e.g., src-cli, src-commands)")
	fmt.Println("  <source>              Source container ID")
	fmt.Println("  <destination>         Destination container ID")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --description, -d     Relationship description (required)")
	fmt.Println("  --tech, -t            Technology/protocol (optional, e.g., \"HTTP\", \"gRPC\", \"JSON\")")
	fmt.Println("  --help, -h            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run . design add relationship src-cli command_parser executor --desc \"sends parsed commands to\"")
	fmt.Println("  go run . design add relationship src-cli api_client github_api -d \"makes requests to\" -t \"HTTPS/REST\"")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  Modifies: specs/<module>/design/workspace.dsl")
}
