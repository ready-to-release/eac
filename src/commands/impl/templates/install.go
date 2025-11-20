// Command: templates install
// Short: Install templates to local directory
// Long: Install templates to local directory for customization.
// Long:
// Long: This command copies built-in template files to your local project directory
// Long: where they can be customized. Templates are installed to the templates/
// Long: directory by default.
// Long:
// Long: Once installed locally, you can modify templates to match your project's
// Long: specific needs while keeping the originals as a reference.
// Long:
// Long: Example:
// Long:   templates install docs
// Long:   templates install specs
// HasSideEffects: true
package templates

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(TemplatesInstall)
}

// TemplatesInstall is the base handler for templates install commands
// It handles unknown template names and shows helpful error messages
func TemplatesInstall() int {
	// Check if a template name was provided
	// Args: [binary, "templates", "install", <template-name>, ...flags]
	if len(os.Args) < 4 {
		// No template name provided - this shouldn't happen since main.go
		// will show subcommand help, but handle it gracefully
		fmt.Fprintf(os.Stderr, "Error: template name required\n")
		fmt.Fprintf(os.Stderr, "Usage: go run . templates install <template-name> [flags...]\n")
		showAvailableTemplates("install")
		return 1
	}

	templateName := os.Args[3]

	// Check if this template exists by looking for the registered command
	specificCommand := "templates install " + templateName
	commands := registry.GetCommands()

	if _, exists := commands[specificCommand]; exists {
		// Template command exists - this shouldn't happen since main.go uses
		// longest match first, but if it does, delegate to it
		return commands[specificCommand]()
	}

	// Template doesn't exist - show helpful error
	fmt.Fprintf(os.Stderr, "Error: unknown template: %s\n", templateName)
	showAvailableTemplates("install")
	return 1
}
