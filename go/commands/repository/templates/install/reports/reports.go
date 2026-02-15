package reports

import (
	"context"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/templates/install"
	"github.com/ready-to-release/eac/go/core/paths"
)

type templatesInstallReportsCommand struct{}

var _ core.SimpleCommandPort = (*templatesInstallReportsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&templatesInstallReportsCommand{},
	}
}

func (c *templatesInstallReportsCommand) Name() string { return "templates install reports" }

func (c *templatesInstallReportsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates-install-reports",
		Short:         "Install report templates without value replacements",
		Long:          "Install report templates by copying files as-is (no variable substitution).\nTemplates preserve {{ .Variable }} placeholders for later customization.\n\nTemplate Source and Destination:\n  Source: templates/reports/ (fixed)\n  Destination: .clie/templates/reports/ (fixed)\n\nUse Case:\n  Install templates once to your project, then customize them as needed.\n  This command copies files without replacing placeholders.\n\nExamples:\n  templates install reports\n  templates install reports --debug",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save detailed logs to out/commands.log"},
		},
	}
}

func (c *templatesInstallReportsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return TemplatesInstallReports()
}

// TemplatesInstallReports installs report templates.
func TemplatesInstallReports() int {
	return install.Run(install.Params{
		TemplateName: "Report",
		TemplateDir: func(root string) string {
			return paths.TemplateReportsPath(root)
		},
		ParseConfig: install.DebugOnlyConfig(func(workspaceRoot string) string {
			return filepath.Join(workspaceRoot, paths.CLIEDir, "templates", "reports")
		}),
	})
}
