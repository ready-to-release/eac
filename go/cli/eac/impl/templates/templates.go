package templates

import (
	"context"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/help"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
)

type templatesCommand struct{}

var _ core.SimpleCommandPort = (*templatesCommand)(nil)

func (c *templatesCommand) Name() string { return "templates" }

func (c *templatesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "templates",
		Short:         "Install project templates for documentation, AI prompts, and specifications",
		IsParent:      true,
		SubcommandGroups: []core.SubcommandGroup{
			{Name: "Template Types", Subcommands: []string{"install"}},
		},
		Examples: []string{
			"clie templates install docs",
			"clie templates install docs --destination ./custom-docs",
			"clie templates install ai --debug",
		},
	}
}

func (c *templatesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Templates()
}

var log = logging.C()


// printHelp prints the help for the templates command using registry metadata.
func printHelp() {
	cmd, _ := registry.Global().Get("templates")
	help.PrintHelp(os.Stdout, cmd, registry.Global())
}

// Templates command entry point.
func Templates() int {
	args := os.Args[2:] // Skip program name and "templates"

	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	if len(args) == 0 {
		printHelp()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printHelp()
		return 0
	case "install":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("unknown subcommand: %s\n", args[0])
		printHelp()
		return 1
	}
}

