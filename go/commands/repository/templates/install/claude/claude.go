package claude

import (
	"context"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/templates/install"
	"github.com/ready-to-release/eac/go/core/paths"
)

type templatesInstallClaudeCommand struct{}

var _ core.SimpleCommandPort = (*templatesInstallClaudeCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&templatesInstallClaudeCommand{},
	}
}

func (c *templatesInstallClaudeCommand) Name() string { return "templates install claude" }

func (c *templatesInstallClaudeCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates-install-claude",
		Short:         "Install Claude Code configuration templates without value replacements",
		Long:          "Install Claude Code templates by copying files as-is (no variable substitution).\nTemplates preserve workflow configurations for Claude Code integration.\n\nTemplate Source and Destination:\n  Source: templates/claude/ (fixed)\n  Destination: .claude/ (fixed)\n\nInstalled Files:\n  agents/architect.md, agents/debugger.md, agents/test-engineer.md\n  commands/plan.md, commands/implement.md, commands/test.md, commands/review.md\n  skills/feature-workflow.md, skills/refactor-safe.md\n  setup/mcp-setup.md, setup/.mcp.json.template\n\nUse Case:\n  Install Claude Code workflow templates that demonstrate MCP command usage.\n  Templates are language-agnostic and showcase auto-discovery workflows.\n\nExamples:\n  templates install claude\n  templates install claude --debug",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save detailed logs to out/commands.log"},
		},
	}
}

func (c *templatesInstallClaudeCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return TemplatesInstallClaude()
}

// TemplatesInstallClaude installs Claude Code configuration templates.
func TemplatesInstallClaude() int {
	return install.Run(install.Params{
		TemplateName: "Claude Code",
		TemplateDir: func(root string) string {
			return paths.TemplatePath(root, "claude")
		},
		ParseConfig: install.DebugOnlyConfig(func(workspaceRoot string) string {
			return filepath.Join(workspaceRoot, ".claude")
		}),
	})
}
