// Command: templates apply
// Short: Apply templates with value replacements
// Long: Apply templates by processing files and replacing {{ .Variable }} placeholders with values.
// Long:
// Long: This command renders template files by substituting variables with actual values.
// Long: Use this when you want to generate documentation, reports, or other files from templates.
// Long:
// Long: Available template types:
// Long:   docs     - Documentation templates (outputs to .docs/reference)
// Long:
// Long: How it works:
// Long:   1. Fetches template files (from local directory or GitHub)
// Long:   2. Reads replacement values from JSON file (optional)
// Long:   3. Renders templates by replacing {{ .Variable }} with values
// Long:   4. Writes rendered files to destination directory
// Long:
// Long: Examples:
// Long:   templates apply docs
// Long:   templates apply docs --source ./my-templates
// Long:   templates apply docs --input-json values.json
// Long:
// Long: Use "help templates apply <template-type>" for detailed information.
// HasSideEffects: true
package templates

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(TemplatesApply)
}

// TemplatesApply is the base handler for templates apply commands
// It handles unknown template names and shows helpful error messages
func TemplatesApply() int {
	// Check if a template name was provided
	// Args: [binary, "templates", "apply", <template-name>, ...flags]
	if len(os.Args) < 4 {
		// No template name provided - this shouldn't happen since main.go
		// will show subcommand help, but handle it gracefully
		fmt.Fprintf(os.Stderr, "Error: template name required\n")
		fmt.Fprintf(os.Stderr, "Usage: go run . templates apply <template-name> [flags...]\n")
		showAvailableTemplates("apply")
		return 1
	}

	templateName := os.Args[3]

	// Check if this template exists by looking for the registered command
	specificCommand := "templates apply " + templateName
	commands := registry.GetCommands()

	if _, exists := commands[specificCommand]; exists {
		// Template command exists - this shouldn't happen since main.go uses
		// longest match first, but if it does, delegate to it
		return commands[specificCommand]()
	}

	// Template doesn't exist - show helpful error
	fmt.Fprintf(os.Stderr, "Error: unknown template: %s\n", templateName)
	showAvailableTemplates("apply")
	return 1
}

// showAvailableTemplates lists all available templates for a given category
func showAvailableTemplates(category string) {
	templates := getAvailableTemplates(category)

	if len(templates) == 0 {
		fmt.Fprintf(os.Stderr, "No templates available for category: %s\n", category)
		return
	}

	fmt.Fprintf(os.Stderr, "Available templates: %s\n", strings.Join(templates, ", "))
}

// getAvailableTemplates scans the registry for templates in a given category
func getAvailableTemplates(category string) []string {
	prefix := "templates " + category + " "
	commands := registry.GetCommands()

	templateMap := make(map[string]bool)
	for cmdName := range commands {
		if strings.HasPrefix(cmdName, prefix) {
			// Extract template name (everything after "templates <category> ")
			templateName := strings.TrimPrefix(cmdName, prefix)
			// Only take the first word (in case there are nested subcommands)
			parts := strings.Fields(templateName)
			if len(parts) > 0 {
				templateMap[parts[0]] = true
			}
		}
	}

	// Convert map to sorted slice
	templates := make([]string, 0, len(templateMap))
	for name := range templateMap {
		templates = append(templates, name)
	}
	sort.Strings(templates)

	return templates
}
