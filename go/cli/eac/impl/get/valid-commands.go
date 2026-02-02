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
	"strings"

	gethelper "github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func init() {
	registry.Register(GetValidCommands)
}

// validCommandsFlags defines valid flags for the get valid-commands command

// CommandInfo represents a command with its description.
type CommandInfo struct {
	Command     string `json:"command" yaml:"command"`
	Description string `json:"description" yaml:"description"`
}

func GetValidCommands() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	return gethelper.ExecuteGetCommand(func() (interface{}, error) {
		reg := registry.GetCommandRegistry()

		// Use a map to deduplicate commands
		commandMap := make(map[string]string) // command -> description

		// Add all explicitly registered commands
		for _, cmd := range reg {
			commandMap[cmd.ActualCommand] = cmd.Short
		}

		// Extract parent commands from subcommand prefixes
		// e.g., "pipeline await-ci" implies "pipeline" is a parent command
		for cmdName := range reg {
			parts := strings.Fields(cmdName)
			// Build all parent prefixes
			for i := 1; i < len(parts); i++ {
				parent := strings.Join(parts[:i], " ")
				if _, exists := commandMap[parent]; !exists {
					// Add parent command with generic description
					commandMap[parent] = fmt.Sprintf("Parent command for %s subcommands", parent)
				}
			}
		}

		// Convert map to slice
		commands := make([]CommandInfo, 0, len(commandMap))
		for cmd, desc := range commandMap {
			commands = append(commands, CommandInfo{
				Command:     cmd,
				Description: desc,
			})
		}

		// Sort alphabetically by command name
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Command < commands[j].Command
		})

		return commands, nil
	})
}
