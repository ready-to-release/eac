// Command: show valid-commands
// Short: Show all valid commands in a table
// Long: The show valid-commands command displays all registered commands in the EAC CLI.
// Long: Shows command names with their descriptions, sorted alphabetically.
// Long:
// Long: Expected Output:
// Long: - Table with columns: Command, Description
// Long: - Footer row showing total number of commands
// Long: - Commands sorted alphabetically
package show

import (
	"fmt"
	"os"
	"sort"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func init() {
	registry.Register(ShowValidCommands)
}

func ShowValidCommands() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	reg := registry.GetCommandRegistry()

	// Extract and sort commands
	type cmdInfo struct {
		name  string
		short string
	}
	commands := make([]cmdInfo, 0, len(reg))
	for _, cmd := range reg {
		commands = append(commands, cmdInfo{
			name:  cmd.ActualCommand,
			short: cmd.Short,
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
