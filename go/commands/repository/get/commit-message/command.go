package commitmessage

import (
	"context"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

type getCommitMessageCommand struct{}

var _ core.SimpleCommandPort = (*getCommitMessageCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&getCommitMessageCommand{},
	}
}

func (c *getCommitMessageCommand) Name() string { return "get commit-message" }

func (c *getCommitMessageCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-commit-message",
		Short:         "Generate AI-powered commit messages from staged changes",
		Long: "The get commit-message command uses AI to analyze your staged git changes and generate a structured,\nconventional commit message that follows project standards and includes module-specific details.\nThe generated message includes a top-level summary and per-module sections describing changes.\nAll output is validated against the commit message contract to ensure consistency and quality.\nBy default, the command outputs the commit message to stdout. Use --debug to save intermediate outputs.\nUse --commit to automatically create a git commit with the generated message.",
		Notes: "Expected Output:\n- Structured conventional commit message to stdout\n- Top-level summary and per-module sections\n- Validated against commit message contract\n- Debug outputs in out/ if --debug enabled",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Enable debug mode to save intermediate outputs (context, prompts, AI responses) to the 'out' directory for troubleshooting and analysis"},
			{Name: "commit", Shorthand: "c", Type: "bool", DefaultValue: "false", Usage: "Automatically create git commit with generated message"},
		},
	}
}

func (c *getCommitMessageCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetCommitMessage()
}
