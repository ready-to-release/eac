// Command: templates
// Description: Manage project templates for documentation and specifications
// HasSideEffects: true
package templates

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register(Templates)
}

// Templates command entry point
func Templates() int {
	args := os.Args[2:] // Skip program name and "templates"

	if len(args) == 0 {
		printTemplatesUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printTemplatesUsage()
		return 0
	case "apply", "install", "list":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printTemplatesUsage()
		return 1
	}
}

func printTemplatesUsage() {
	fmt.Println("Manage project templates for documentation and specifications")
	fmt.Println()
	fmt.Println("Usage: r2r templates <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  apply                     Apply templates with variable substitution")
	fmt.Println("  install                   Install template files to local directory")
	fmt.Println("  list                      List all placeholder variables in templates")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List all template variables")
	fmt.Println("  r2r templates list")
	fmt.Println()
	fmt.Println("  # Install specification templates")
	fmt.Println("  r2r templates install specs")
	fmt.Println()
	fmt.Println("  # Apply documentation templates")
	fmt.Println("  r2r templates apply docs")
	fmt.Println()
	fmt.Println("Use 'r2r templates <subcommand> --help' for more information about a command.")
}
