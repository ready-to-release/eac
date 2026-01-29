// Command: templates
// Short: Install project templates for documentation, AI prompts, and specifications
// IsParent: true
// Group.Template Types: install
// Example: r2r templates install docs
// Example: r2r templates install docs --destination ./custom-docs
// Example: r2r templates install ai --debug
package templates

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/help"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Templates)
}

// printHelp prints the help for the templates command using registry metadata.
func printHelp() {
	reg := registry.GetCommand("templates")
	help.PrintHelp(os.Stdout, reg, registry.GetCommandRegistry())
}

// Templates command entry point.
func Templates() int {
	args := os.Args[2:] // Skip program name and "templates"

	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	if len(args) == 0 {
		printHelp()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printHelp()
		return 0
	case "install":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("unknown subcommand: %s\n", args[0])
		printHelp()
		return 1
	}
}

