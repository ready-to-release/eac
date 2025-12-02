// Command: templates
// Description: Manage project templates for documentation and specifications
package templates

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

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
		log.Errorf("unknown subcommand: %s\n", args[0])
		printTemplatesUsage()
		return 1
	}
}

func printTemplatesUsage() {
	log.Info("Manage project templates for documentation and specifications")
	log.Info("")
	log.Info("Usage: r2r templates <subcommand> [args...]")
	log.Info("")
	log.Info("Subcommands:")
	log.Info("  apply                     Apply templates with variable substitution")
	log.Info("  install                   Install template files to local directory")
	log.Info("  list                      List all placeholder variables in templates")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # List all template variables")
	log.Info("  r2r templates list")
	log.Info("")
	log.Info("  # Install specification templates")
	log.Info("  r2r templates install specs")
	log.Info("")
	log.Info("  # Apply documentation templates")
	log.Info("  r2r templates apply docs")
	log.Info("")
	log.Info("Use 'r2r templates <subcommand> --help' for more information about a command.")
}
