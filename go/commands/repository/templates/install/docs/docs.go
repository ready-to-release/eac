package docs

import (
	"context"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/templates/install"
	"github.com/ready-to-release/eac/go/core/paths"
)

type templatesInstallDocsCommand struct{}

var _ core.SimpleCommandPort = (*templatesInstallDocsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&templatesInstallDocsCommand{},
	}
}

func (c *templatesInstallDocsCommand) Name() string { return "templates install docs" }

func (c *templatesInstallDocsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates-install-docs",
		Short:         "Install documentation templates without value replacements",
		Long:          "Install documentation templates by copying files as-is (no variable substitution).\nTemplates preserve {{ .Variable }} placeholders for later customization.\n\nTemplate Source:\n  Always uses local templates/docs/ directory from the repository.\n\nUse Case:\n  Install templates once to your project, then customize them as needed.\n  This command copies files without replacing placeholders.\n\nExamples:\n  templates install docs\n  templates install docs --destination ./custom-docs\n  templates install docs --destination docs/api --debug",
		Flags: []core.FlagSpec{
			{Name: "destination", Type: "string", Usage: "Output directory (default: docs/reference/)"},
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save detailed logs to out/commands.log"},
		},
	}
}

func (c *templatesInstallDocsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return TemplatesInstallDocs()
}

// TemplatesInstallDocs installs documentation templates.
func TemplatesInstallDocs() int {
	return install.Run(install.Params{
		TemplateName: "Documentation",
		TemplateDir: func(root string) string {
			return paths.TemplatePath(root, "docs")
		},
		ParseConfig: parseDocsConfig,
	})
}

// parseDocsConfig parses docs-specific configuration including --destination flag.
func parseDocsConfig(workspaceRoot string, args []string) (*install.Config, error) {
	destination := filepath.Join(paths.DocsDir, "reference")
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debug = true
		case "--destination":
			if i+1 < len(args) {
				destination = args[i+1]
				i++
			}
		}
	}

	// Resolve destination path
	if !filepath.IsAbs(destination) {
		destination = filepath.Join(workspaceRoot, destination)
	}

	return &install.Config{
		Destination:   destination,
		WorkspaceRoot: workspaceRoot,
		Debug:         debug,
	}, nil
}
