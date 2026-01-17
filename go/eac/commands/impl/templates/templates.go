// Command: templates
package templates

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Templates)
}

// commandFlags defines valid flags for the templates command

// Templates command entry point.
func Templates() int {
	args := os.Args[2:] // Skip program name and "templates"

	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	if len(args) == 0 {
		printTemplatesUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printTemplatesUsage()
		return 0
	case "install":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("unknown subcommand: %s\n", args[0])
		printTemplatesUsage()
		return 1
	}
}

func printTemplatesUsage() {
	log.Info("Install project templates for documentation, AI prompts, and specifications")
	log.Info("")
	log.Info("Usage: r2r templates install <template-type> [flags...]")
	log.Info("")
	log.Info("Template Types:")
	log.Info("  docs      Install documentation templates to docs/reference/")
	log.Info("  ai        Install AI prompt templates to .r2r/eac/templates/ai/")
	log.Info("  reports   Install report templates to .r2r/templates/reports/")
	log.Info("  specs     Install specification templates to specs/risk-controls/")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Install documentation templates")
	log.Info("  r2r templates install docs")
	log.Info("")
	log.Info("  # Install docs to custom location")
	log.Info("  r2r templates install docs --destination ./custom-docs")
	log.Info("")
	log.Info("  # Install AI templates with debug logging")
	log.Info("  r2r templates install ai --debug")
	log.Info("")
	log.Info("Use 'r2r templates install <template-type> --help' for more information.")
}
