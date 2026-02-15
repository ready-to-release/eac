package ai

import (
	"context"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/templates/install"
	"github.com/ready-to-release/eac/go/core/paths"
)

type templatesInstallAICommand struct{}

var _ core.SimpleCommandPort = (*templatesInstallAICommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&templatesInstallAICommand{},
	}
}

func (c *templatesInstallAICommand) Name() string { return "templates install ai" }

func (c *templatesInstallAICommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates-install-ai",
		Short:         "Install AI prompt templates without value replacements",
		Long:          "Install AI prompt templates by copying files as-is (no variable substitution).\nTemplates preserve {{ .Variable }} placeholders for later customization.\n\nTemplate Source and Destination:\n  Source: templates/ai/ (fixed)\n  Destination: .eac/templates/ai/ (fixed)\n\nUse Case:\n  Install AI prompt templates once to your project, then customize them as needed.\n  This command copies files without replacing placeholders.\n\nExamples:\n  templates install ai\n  templates install ai --debug",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save detailed logs to out/commands.log"},
		},
	}
}

func (c *templatesInstallAICommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return TemplatesInstallAI()
}

// TemplatesInstallAI installs AI prompt templates.
func TemplatesInstallAI() int {
	return install.Run(install.Params{
		TemplateName: "AI",
		TemplateDir: func(root string) string {
			return paths.TemplatePath(root, "ai")
		},
		ParseConfig: install.DebugOnlyConfig(func(workspaceRoot string) string {
			return filepath.Join(workspaceRoot, paths.EACDir, "templates", "ai")
		}),
	})
}
