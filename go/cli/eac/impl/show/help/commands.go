package help

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/logging"
)

type showHelpCommand struct{}

var _ core.SimpleCommandPort = (*showHelpCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&showHelpCommand{},
	}
}

func (c *showHelpCommand) Name() string { return "show help" }

func (c *showHelpCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-help",
		Short:         "Display help information for commands",
		Long:          "The show help command provides documentation for all available commands.\nWhen called without arguments, it lists all commands.\nWhen called with a command name, it displays detailed help including description, flags, and usage examples.",
		Flags: []core.FlagSpec{
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show detailed information including all subcommands and advanced options"},
		},
	}
}

func (c *showHelpCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowHelp()
}

var log = logging.C()

func ShowHelp() int {
	args := os.Args[3:] // Skip program name, "show", and "help"

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	verbose := false
	var commandParts []string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		} else if !strings.HasPrefix(arg, "-") {
			commandParts = append(commandParts, arg)
		}
	}

	// Build full command name from parts (e.g., "show modules")
	commandName := strings.Join(commandParts, " ")

	// If command name is provided, show detailed help for that command
	if commandName != "" {
		return showCommandHelp(commandName, verbose)
	}

	// Otherwise, show list of all commands
	return showAllCommands(verbose)
}

// showAllCommands lists all available commands.
func showAllCommands(verbose bool) int {
	reg := registry.Global()
	names := reg.Names()

	// Simple alphabetical sort (Names() already returns sorted, but ensure)
	sort.Strings(names)

	// Render as compact list
	result := render.RenderCompactList("Available Commands", names)
	log.Info(result)

	return 0
}

// showParentHelp displays help for a command prefix that has subcommands but no direct registration.
func showParentHelp(parentName string, subcommands []string, verbose bool) int {
	reg := registry.Global()

	// Display NAME section
	log.Info("NAME")
	log.Infof("    %s - Parent command with subcommands\n", parentName)

	// Display SYNOPSIS section
	log.Info("SYNOPSIS")
	log.Infof("    eac %s <subcommand> [arguments]\n", parentName)

	// Display COMMANDS section (subcommands)
	log.Info("COMMANDS")
	for _, subcmd := range subcommands {
		cmdPort, ok := reg.Get(subcmd)

		desc := ""
		if ok {
			desc = cmdPort.Metadata().Short
		}

		// Show full command name (e.g., "create spec") so users know the full command
		// Format with padding
		padding := strings.Repeat(" ", max(2, 24-len(subcmd)))
		log.Infof("    %s%s%s", subcmd, padding, desc)
	}
	log.Info("")

	// Display EXAMPLES section
	log.Info("EXAMPLES")
	if len(subcommands) > 0 {
		// Show first subcommand as example
		log.Infof("    eac %s", subcommands[0])
	}
	log.Info("")

	return 0
}

// showCommandHelp displays detailed help for a specific command.
func showCommandHelp(commandName string, verbose bool) int {
	reg := registry.Global()

	cmdPort, found := reg.Get(commandName)
	if !found {
		// Check if this is a parent prefix with subcommands
		subcommands := getSubcommands(commandName)
		if len(subcommands) > 0 {
			return showParentHelp(commandName, subcommands, verbose)
		}

		log.Errorf("Error: unknown command '%s'", commandName)
		log.Error("\nUse 'show help' to see all available commands.")
		return 1
	}

	meta := cmdPort.Metadata()

	// Display NAME section
	log.Info("NAME")
	log.Infof("    %s - %s\n", cmdPort.Name(), meta.Short)

	// Display SYNOPSIS section
	log.Info("SYNOPSIS")
	synopsis := buildSynopsis(cmdPort)
	log.Infof("    %s\n", synopsis)

	// Display DESCRIPTION section
	if meta.Long != "" {
		log.Info("DESCRIPTION")
		// Wrap long description with indentation
		lines := strings.Split(meta.Long, "\n")
		for _, line := range lines {
			if line == "" {
				log.Info("")
			} else {
				log.Infof("    %s", line)
			}
		}
		log.Info("")
	}

	// Display COMMANDS section (subcommands)
	subcommands := getSubcommands(commandName)
	if len(subcommands) > 0 {
		log.Info("COMMANDS")
		for _, subcmd := range subcommands {
			subcmdPort, ok := reg.Get(subcmd)

			desc := ""
			if ok {
				desc = subcmdPort.Metadata().Short
			}

			// Extract just the subcommand part (e.g., "create" from "work create")
			subPart := strings.TrimPrefix(subcmd, commandName+" ")

			// Format with padding
			padding := strings.Repeat(" ", max(2, 20-len(subPart)))
			log.Infof("    %s%s%s", subPart, padding, desc)
		}
		log.Info("")
	}

	// Display FLAGS section
	if len(meta.Flags) > 0 {
		log.Info("FLAGS")
		for _, flag := range meta.Flags {
			displayFlag(flag)
		}
		log.Info("")
	}

	// Display EXAMPLES section
	log.Info("EXAMPLES")
	log.Infof("    eac %s", cmdPort.Name())
	if len(meta.Flags) > 0 {
		// Show example with a flag if available
		for _, flag := range meta.Flags {
			if flag.Type == "bool" {
				log.Infof("    eac %s --%s", cmdPort.Name(), flag.Name)
				break
			}
		}
	}
	log.Info("")

	// Display additional info
	if verbose {
		log.Info("ADDITIONAL INFORMATION")
		log.Infof("    Canonical name: %s", meta.CanonicalName)
		log.Info("")
	}

	return 0
}

// buildSynopsis builds a synopsis line for a command.
func buildSynopsis(cmd core.CommandPort) string {
	meta := cmd.Metadata()
	parts := []string{"eac", cmd.Name()}

	// Add flags
	if len(meta.Flags) > 0 {
		parts = append(parts, "[flags]")
	}

	// Add arguments placeholder
	parts = append(parts, "[arguments]")

	return strings.Join(parts, " ")
}

// displayFlag formats and displays a single flag.
func displayFlag(flag core.FlagSpec) {
	// Build flag name with shorthand
	flagName := "--" + flag.Name
	if flag.Shorthand != "" {
		flagName = "-" + flag.Shorthand + ", " + flagName
	}

	// Build type and requirement info
	typeInfo := ""
	if flag.Type != "" && flag.Type != "bool" {
		typeInfo = fmt.Sprintf("<%s>", flag.Type)
	}

	requirements := []string{}
	if flag.Required {
		requirements = append(requirements, "required")
	} else {
		requirements = append(requirements, "optional")
	}

	if flag.DefaultValue != "" {
		requirements = append(requirements, fmt.Sprintf("default: %s", flag.DefaultValue))
	}

	reqInfo := strings.Join(requirements, ", ")

	// Display flag header
	log.Infof("    %s %s (%s)", flagName, typeInfo, reqInfo)

	// Display usage with indentation
	if flag.Usage != "" {
		log.Infof("        %s", flag.Usage)
	}

	log.Info("")
}

// getSubcommands returns all subcommands for a given parent command as sorted name strings.
func getSubcommands(parentCommand string) []string {
	subcmds := registry.Global().Subcommands(parentCommand)

	// Extract names and sort
	names := make([]string, 0, len(subcmds))
	for _, cmd := range subcmds {
		names = append(names, cmd.Name())
	}
	sort.Strings(names)

	return names
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
