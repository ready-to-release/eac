// Command: show help
// Short: Display help information for commands
// Long: The show help command provides documentation for all available commands.
// Long: When called without arguments, it lists all commands.
// Long: When called with a command name, it displays detailed help including description, flags, and usage examples.
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed information including all subcommands and advanced options
package list

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// commandFlags defines valid flags for the list commands command

var log = logging.C()

func init() {
	registry.Register(ShowHelp)
}

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
	// Get sorted command names
	var names []string
	commands := registry.GetCommands()
	for name := range commands {
		names = append(names, name)
	}

	// Simple alphabetical sort
	sort.Strings(names)

	// Render as compact list
	result := render.RenderCompactList("Available Commands", names)
	log.Info(result)

	return 0
}

// showParentHelp displays help for a command prefix that has subcommands but no direct registration.
func showParentHelp(parentName string, subcommands []string, verbose bool) int {
	commandRegistry := registry.GetCommandRegistry()

	// Display NAME section
	log.Info("NAME")
	log.Infof("    %s - Parent command with subcommands\n", parentName)

	// Display SYNOPSIS section
	log.Info("SYNOPSIS")
	log.Infof("    eac %s <subcommand> [arguments]\n", parentName)

	// Display COMMANDS section (subcommands)
	log.Info("COMMANDS")
	for _, subcmd := range subcommands {
		subReg := commandRegistry[subcmd]

		desc := ""
		if subReg != nil && subReg.Short != "" {
			desc = subReg.Short
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
	commandRegistry := registry.GetCommandRegistry()

	reg := commandRegistry[commandName]
	if reg == nil {
		// Check if this is a parent prefix with subcommands
		subcommands := getSubcommands(commandName)
		if len(subcommands) > 0 {
			return showParentHelp(commandName, subcommands, verbose)
		}

		log.Errorf("Error: unknown command '%s'", commandName)
		log.Error("\nUse 'show help' to see all available commands.")
		return 1
	}

	// Display NAME section
	log.Info("NAME")
	log.Infof("    %s - %s\n", reg.ActualCommand, reg.Short)

	// Display SYNOPSIS section
	log.Info("SYNOPSIS")
	synopsis := buildSynopsis(reg)
	log.Infof("    %s\n", synopsis)

	// Display DESCRIPTION section
	if reg.Long != "" {
		log.Info("DESCRIPTION")
		// Wrap long description with indentation
		lines := strings.Split(reg.Long, "\n")
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
			subReg := commandRegistry[subcmd]

			desc := ""
			if subReg != nil && subReg.Short != "" {
				desc = subReg.Short
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
	if len(reg.Flags) > 0 {
		log.Info("FLAGS")
		for _, flag := range reg.Flags {
			displayFlag(flag)
		}
		log.Info("")
	}

	// Display EXAMPLES section
	log.Info("EXAMPLES")
	log.Infof("    eac %s", reg.ActualCommand)
	if len(reg.Flags) > 0 {
		// Show example with a flag if available
		for _, flag := range reg.Flags {
			if flag.Type == "bool" {
				log.Infof("    eac %s --%s", reg.ActualCommand, flag.Name)
				break
			}
		}
	}
	log.Info("")

	// Display additional info
	if verbose {
		log.Info("ADDITIONAL INFORMATION")
		log.Infof("    Canonical name: %s", reg.CanonicalName)
		log.Info("")
	}

	return 0
}

// buildSynopsis builds a synopsis line for a command.
func buildSynopsis(reg *registry.CommandRegistration) string {
	parts := []string{"eac", reg.ActualCommand}

	// Add flags
	if len(reg.Flags) > 0 {
		parts = append(parts, "[flags]")
	}

	// Add arguments placeholder
	parts = append(parts, "[arguments]")

	return strings.Join(parts, " ")
}

// displayFlag formats and displays a single flag.
func displayFlag(flag registry.FlagMetadata) {
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

	// Display completion values if available
	if len(flag.Completion) > 0 && flag.Type != "bool" {
		log.Infof("        Valid values: %s", strings.Join(flag.Completion, ", "))
	}

	log.Info("")
}

// getSubcommands returns all subcommands for a given parent command.
func getSubcommands(parentCommand string) []string {
	commands := registry.GetCommands()
	var subcommands []string

	prefix := parentCommand + " "
	for cmdName := range commands {
		// Check if this is a direct subcommand (no further nesting)
		if strings.HasPrefix(cmdName, prefix) {
			// Extract the part after the prefix
			remainder := strings.TrimPrefix(cmdName, prefix)
			// Only include if it's a direct child (no more spaces)
			if !strings.Contains(remainder, " ") {
				subcommands = append(subcommands, cmdName)
			}
		}
	}

	// Sort alphabetically
	sort.Strings(subcommands)
	return subcommands
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
