package show

import (
	"context"
	"fmt"
	"os"
	"sort"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

type showValidCommandsCommand struct{}

var _ core.SimpleCommandPort = (*showValidCommandsCommand)(nil)

func (c *showValidCommandsCommand) Name() string { return "show valid-commands" }

func (c *showValidCommandsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-valid-commands",
		Short:         "Show all valid commands in a table",
		Long:          "The show valid-commands command displays all registered commands in the EAC CLI.\nShows command names with their descriptions, sorted alphabetically.\n\nExpected Output:\n- Table with columns: Command, Description\n- Footer row showing total number of commands\n- Commands sorted alphabetically",
	}
}

func (c *showValidCommandsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowValidCommands()
}

func ShowValidCommands() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	all := registry.Global().All()

	// Extract and sort commands
	type cmdInfo struct {
		name  string
		short string
	}
	commands := make([]cmdInfo, 0, len(all))
	for _, cmd := range all {
		commands = append(commands, cmdInfo{
			name:  cmd.Name(),
			short: cmd.Metadata().Short,
		})
	}

	// Sort alphabetically
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].name < commands[j].name
	})

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Command", "Description")

	for _, cmd := range commands {
		tb.AddRow(cmd.name, cmd.short)
	}

	tb.WithFooter("Total Commands", len(commands))

	fmt.Println(tb.Build())
	return 0
}
