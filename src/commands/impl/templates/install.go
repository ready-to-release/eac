// Command: templates install
// Short: Install templates without value replacements
// Long: Install templates by copying files as-is (no variable substitution).
// Long:
// Long: This command copies template files to your local project without rendering them.
// Long: All {{ .Variable }} placeholders remain unchanged for later customization.
// Long:
// Long: Available template types:
// Long:   reports  - Report templates (outputs to .r2r/templates/reports)
// Long:
// Long: How it works:
// Long:   1. Fetches template files (from local directory or GitHub)
// Long:   2. Copies files as-is to destination directory
// Long:   3. Preserves all {{ .Variable }} placeholders unchanged
// Long:
// Long: Use Case:
// Long:   Install templates once to your project, then customize them as needed.
// Long:   Unlike "apply", this command does NOT replace placeholders.
// Long:
// Long: Examples:
// Long:   templates install reports
// Long:   templates install reports --source ./my-templates
// Long:
// Long: Use "help templates install <template-type>" for detailed information.
package templates

import (
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
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
		log.Errorf("Error: template name required")
		log.Errorf("Usage: go run . templates install <template-name> [flags...]")
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
	log.Errorf("Error: unknown template: %s", templateName)
	showAvailableTemplates("install")
	return 1
}
