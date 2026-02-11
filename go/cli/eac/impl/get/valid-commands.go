package get

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	gethelper "github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

type getValidCommandsCommand struct{}

var _ core.SimpleCommandPort = (*getValidCommandsCommand)(nil)

func (c *getValidCommandsCommand) Name() string { return "get valid-commands" }

func (c *getValidCommandsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-valid-commands",
		Short:         "Get all valid commands in structured format",
		Long:          "Expected Output:\nYAML list of all valid commands with descriptions, sorted alphabetically.\nEach entry contains the command name and its description.",
	}
}

func (c *getValidCommandsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetValidCommands()
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
		reg := registry.Global()

		// Use a map to deduplicate commands
		commandMap := make(map[string]string) // command -> description

		// Add all explicitly registered commands
		for _, cmd := range reg.All() {
			commandMap[cmd.Name()] = cmd.Metadata().Short
		}

		// Extract parent commands from subcommand prefixes
		// e.g., "pipeline await-ci" implies "pipeline" is a parent command
		for _, cmdName := range reg.Names() {
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
