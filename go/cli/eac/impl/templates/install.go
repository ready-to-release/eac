// Command: templates install
// Short: Install templates without value replacements
// Long: Install templates by copying files as-is (no variable substitution).
// Long:
// Long: This command copies template files to your local project without rendering them.
// Long: All {{ .Variable }} placeholders remain unchanged for later customization.
// Long:
// Long: Available template types:
// Long:   docs     - Documentation templates (outputs to docs/reference/)
// Long:   ai       - AI prompt templates (outputs to .eac/templates/ai/)
// Long:   reports  - Report templates (outputs to .eac/templates/reports/)
// Long:   specs    - Specification templates (outputs to specs/risk-controls/)
// Long:   claude   - Claude Code templates (outputs to .claude/)
// Long:
// Long: How it works:
// Long:   1. Uses local template files from repository
// Long:   2. Copies files as-is to destination directory
// Long:   3. Preserves all {{ .Variable }} placeholders unchanged
// Long:
// Long: Expected Output:
// Long:   - Template files copied as-is to destination directory
// Long:   - All {{ .Variable }} placeholders preserved (not replaced)
// Long:
// Long: Use Case:
// Long:   Install templates once to your project, then customize them as needed.
// Long:
// Long: Examples:
// Long:   templates install docs
// Long:   templates install ai
// Long:   templates install reports
// Long:   templates install specs
// Long:   templates install claude
// Long:   templates install docs --destination ./custom-docs
// Long:
// Long: Use "help templates install <template-type>" for detailed information.
package templates

import (
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

// commandFlags defines valid flags for the templates install command.
func init() {
	registry.Register(TemplatesInstall)
}

// TemplatesInstall is the base handler for templates install commands
// It handles unknown template names and shows helpful error messages.
func TemplatesInstall() int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

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

// showAvailableTemplates lists all available templates for a given category.
func showAvailableTemplates(category string) {
	templates := getAvailableTemplates(category)

	if len(templates) == 0 {
		log.Errorf("No templates available for category: %s", category)
		return
	}

	log.Info("Available templates:")
	log.Info("  docs     - Documentation templates")
	log.Info("  ai       - AI prompt templates")
	log.Info("  reports  - Report templates")
	log.Info("  specs    - Specification templates")
	log.Info("  claude   - Claude Code workflow templates")
}

// getAvailableTemplates scans the registry for templates in a given category.
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
