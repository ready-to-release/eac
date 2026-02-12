package specs

import (
	"context"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/templates/install"
	"github.com/ready-to-release/eac/go/core/paths"
)

type templatesInstallSpecsCommand struct{}

var _ core.SimpleCommandPort = (*templatesInstallSpecsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&templatesInstallSpecsCommand{},
	}
}

func (c *templatesInstallSpecsCommand) Name() string { return "templates install specs" }

func (c *templatesInstallSpecsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates-install-specs",
		Short:         "Install specification templates without value replacements",
		Long:          "Install specification templates by copying files as-is (no variable substitution).\nTemplates preserve {{ .Variable }} placeholders for later customization.\n\nTemplate Source and Destination:\n  Source: templates/specs/ (fixed)\n  Destination: specs/risk-controls/ (fixed)\n\nUse Case:\n  Install specification templates once to your project, then customize them as needed.\n  This command copies files without replacing placeholders.\n\nExamples:\n  templates install specs\n  templates install specs --debug",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save detailed logs to out/commands.log"},
		},
	}
}

func (c *templatesInstallSpecsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return TemplatesInstallSpecs()
}

// TemplatesInstallSpecs installs specification templates.
func TemplatesInstallSpecs() int {
	return install.Run(install.Params{
		TemplateName: "Specification",
		TemplateDir: func(root string) string {
			return paths.TemplateSpecsPath(root)
		},
		ParseConfig: install.DebugOnlyConfig(func(workspaceRoot string) string {
			return paths.RiskControlsPath(workspaceRoot)
		}),
	})
}
