// Command: get valid-commands
// Short: Get all valid commands in structured format
// Long:
// Long: Expected Output:
// Long: YAML list of all valid commands with descriptions, sorted alphabetically.
// Long: Each entry contains the command name and its description.
package get

import (
	"fmt"
	"os"
	"sort"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	gethelper "github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetValidCommands)
}

// validCommandsFlags defines valid flags for the get valid-commands command
var validCommandsFlags = []flags.FlagDefinition{
	{Name: "--as-yaml", HasValue: false, ValueType: "bool"},
	{Name: "--as-json", HasValue: false, ValueType: "bool"},
	{Name: "--as-toml", HasValue: false, ValueType: "bool"},
}

// CommandInfo represents a command with its description
type CommandInfo struct {
	Command     string `json:"command" yaml:"command"`
	Description string `json:"description" yaml:"description"`
}

func GetValidCommands() int {
	// Validate flags before parsing
	if err := flags.ValidateFlags(os.Args[3:], validCommandsFlags); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

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
