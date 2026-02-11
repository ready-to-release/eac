package templates

import (
	"context"
	"os"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

type templatesInstallCommand struct{}

var _ core.SimpleCommandPort = (*templatesInstallCommand)(nil)

func (c *templatesInstallCommand) Name() string { return "templates install" }

func (c *templatesInstallCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates-install",
		Short:         "Install templates without value replacements",
		Long:          "Install templates by copying files as-is (no variable substitution).\n\nThis command copies template files to your local project without rendering them.\nAll {{ .Variable }} placeholders remain unchanged for later customization.\n\nAvailable template types:\n  docs     - Documentation templates (outputs to docs/reference/)\n  ai       - AI prompt templates (outputs to .eac/templates/ai/)\n  reports  - Report templates (outputs to .eac/templates/reports/)\n  specs    - Specification templates (outputs to specs/risk-controls/)\n  claude   - Claude Code templates (outputs to .claude/)\n\nHow it works:\n  1. Uses local template files from repository\n  2. Copies files as-is to destination directory\n  3. Preserves all {{ .Variable }} placeholders unchanged\n\nExpected Output:\n  - Template files copied as-is to destination directory\n  - All {{ .Variable }} placeholders preserved (not replaced)\n\nUse Case:\n  Install templates once to your project, then customize them as needed.\n\nExamples:\n  templates install docs\n  templates install ai\n  templates install reports\n  templates install specs\n  templates install claude\n  templates install docs --destination ./custom-docs\n\nUse \"help templates install <template-type>\" for detailed information.",
	}
}

func (c *templatesInstallCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return TemplatesInstall()
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

	if cmd, exists := registry.Global().Get(specificCommand); exists {
		// Template command exists - this shouldn't happen since main.go uses
		// longest match first, but if it does, delegate to it
		if exec, ok := cmd.(core.SimpleCommandPort); ok {
			return exec.Execute(context.Background(), nil)
		}
		return 0
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
	names := registry.Global().Names()

	templateMap := make(map[string]bool)
	for _, cmdName := range names {
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
