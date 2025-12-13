// Command: get valid-commands
// Short: Get all valid commands in structured format
// Long:
// Long: Expected Output:
// Long: YAML list of all valid commands with descriptions, sorted alphabetically.
// Long: Each entry contains the command name and its description.
package get

import (
	"sort"

	gethelper "github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetValidCommands)
}

// CommandInfo represents a command with its description
type CommandInfo struct {
	Command     string `json:"command" yaml:"command"`
	Description string `json:"description" yaml:"description"`
}

func GetValidCommands() int {
	return gethelper.ExecuteGetCommand(func() (interface{}, error) {
		reg := registry.GetCommandRegistry()

		// Extract and sort commands
		commands := make([]CommandInfo, 0, len(reg))
		for _, cmd := range reg {
			commands = append(commands, CommandInfo{
				Command:     cmd.ActualCommand,
				Description: cmd.Short,
			})
		}

		// Sort alphabetically by command name
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Command < commands[j].Command
		})

		return commands, nil
	})
}
