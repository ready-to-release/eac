// Command: design
// Description: Architecture documentation tools using Structurizr - Main command
package design

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register("design", Design)
}

// Design command entry point
func Design() int {
	args := os.Args[2:] // Skip "go" and "run" and "."

	if len(args) == 0 {
		printDesignUsage()
		return 1
	}

	// Check for subcommands
	switch args[0] {
	case "new", "add", "export":
		// Handled by separate registrations in respective files
		return 0
	case "serve":
		// Handled by separate registration
		return 0
	case "list":
		// Handled by separate registration
		return 0
	case "validate":
		// Handled by separate registration
		return 0
	case "--help", "-h":
		printDesignUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printDesignUsage()
		return 1
	}
}

func printDesignUsage() {
	fmt.Println("Architecture documentation tools using Structurizr")
	fmt.Println()
	fmt.Println("Usage: go run . design <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Authoring Subcommands:")
	fmt.Println("  new <module>              Create new architecture workspace")
	fmt.Println("  add container             Add container to workspace")
	fmt.Println("  add relationship          Add relationship between containers")
	fmt.Println("  export <module>           Export workspace DSL content")
	fmt.Println()
	fmt.Println("Viewing Subcommands:")
	fmt.Println("  serve <module>            Start Structurizr Lite viewer")
	fmt.Println("  list                      List available modules with documentation")
	fmt.Println()
	fmt.Println("Validation Subcommands:")
	fmt.Println("  validate <module>         Validate workspace file for one module")
	fmt.Println("  validate --all            Validate workspace files for all modules")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create workspace")
	fmt.Println("  go run . design new src-cli --name \"CLI\" --description \"CLI Architecture\"")
	fmt.Println()
	fmt.Println("  # Add containers")
	fmt.Println("  go run . design add container src-cli parser --tech \"Go\" --desc \"Parses commands\"")
	fmt.Println()
	fmt.Println("  # Add relationships")
	fmt.Println("  go run . design add relationship src-cli parser executor --desc \"sends to\"")
	fmt.Println()
	fmt.Println("  # Export workspace")
	fmt.Println("  go run . design export src-cli")
	fmt.Println()
	fmt.Println("  # View in browser")
	fmt.Println("  go run . design serve src-cli")
	fmt.Println()
	fmt.Println("  # List modules")
	fmt.Println("  go run . design list")
	fmt.Println()
	fmt.Println("  # Validate workspace")
	fmt.Println("  go run . design validate src-cli")
	fmt.Println("  go run . design validate --all")
	fmt.Println()
	fmt.Println("For detailed help on a subcommand:")
	fmt.Println("  go run . design <subcommand> --help")
}
