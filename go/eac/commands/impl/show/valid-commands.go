// Command: show valid-commands
// Short: Show all valid commands in a table
package show

import (
	"sort"

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ShowValidCommands)
}

func ShowValidCommands() int {
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

	log.Info(tb.Build())
	return 0
}
