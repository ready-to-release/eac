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

	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register(ShowHelp)
}

func ShowHelp() int {
	args := os.Args[3:] // Skip program name, "show", and "help"

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

// showAllCommands lists all available commands
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
	fmt.Println(result)

	return 0
}

// showParentHelp displays help for a command prefix that has subcommands but no direct registration
func showParentHelp(parentName string, subcommands []string, verbose bool) int {
	commandRegistry := registry.GetCommandRegistry()

	// Display NAME section
	fmt.Printf("NAME\n")
	fmt.Printf("    %s - Parent command with subcommands\n\n", parentName)

	// Display SYNOPSIS section
	fmt.Printf("SYNOPSIS\n")
	fmt.Printf("    eac %s <subcommand> [arguments]\n\n", parentName)

	// Display COMMANDS section (subcommands)
	fmt.Printf("COMMANDS\n")
	for _, subcmd := range subcommands {
		subReg := commandRegistry[subcmd]

		desc := ""
		if subReg != nil && subReg.Short != "" {
			desc = subReg.Short
		}

		// Show full command name (e.g., "create spec") so users know the full command
		// Format with padding
		padding := strings.Repeat(" ", max(2, 24-len(subcmd)))
		fmt.Printf("    %s%s%s\n", subcmd, padding, desc)
	}
	fmt.Println()

	// Display EXAMPLES section
	fmt.Printf("EXAMPLES\n")
	if len(subcommands) > 0 {
		// Show first subcommand as example
		fmt.Printf("    eac %s\n", subcommands[0])
	}
	fmt.Println()

	return 0
}

// showCommandHelp displays detailed help for a specific command
func showCommandHelp(commandName string, verbose bool) int {
	commandRegistry := registry.GetCommandRegistry()

	reg := commandRegistry[commandName]
	if reg == nil {
		// Check if this is a parent prefix with subcommands
		subcommands := getSubcommands(commandName)
		if len(subcommands) > 0 {
			return showParentHelp(commandName, subcommands, verbose)
		}

		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", commandName)
		fmt.Fprintf(os.Stderr, "\nUse 'show help' to see all available commands.\n")
		return 1
	}

	// Display NAME section
	fmt.Printf("NAME\n")
	fmt.Printf("    %s - %s\n\n", reg.ActualCommand, reg.Short)

	// Display SYNOPSIS section
	fmt.Printf("SYNOPSIS\n")
	synopsis := buildSynopsis(reg)
	fmt.Printf("    %s\n\n", synopsis)

	// Display DESCRIPTION section
	if reg.Long != "" {
		fmt.Printf("DESCRIPTION\n")
		// Wrap long description with indentation
		lines := strings.Split(reg.Long, "\n")
		for _, line := range lines {
			if line == "" {
				fmt.Println()
			} else {
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println()
	}

	// Display COMMANDS section (subcommands)
	subcommands := getSubcommands(commandName)
	if len(subcommands) > 0 {
		fmt.Printf("COMMANDS\n")
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
			fmt.Printf("    %s%s%s\n", subPart, padding, desc)
		}
		fmt.Println()
	}

	// Display FLAGS section
	if len(reg.Flags) > 0 {
		fmt.Printf("FLAGS\n")
		for _, flag := range reg.Flags {
			displayFlag(flag)
		}
		fmt.Println()
	}

	// Display EXAMPLES section
	fmt.Printf("EXAMPLES\n")
	fmt.Printf("    eac %s\n", reg.ActualCommand)
	if len(reg.Flags) > 0 {
		// Show example with a flag if available
		for _, flag := range reg.Flags {
			if flag.Type == "bool" {
				fmt.Printf("    eac %s --%s\n", reg.ActualCommand, flag.Name)
				break
			}
		}
	}
	fmt.Println()

	// Display additional info
	if verbose {
		fmt.Printf("ADDITIONAL INFORMATION\n")
		fmt.Printf("    Canonical name: %s\n", reg.CanonicalName)
		fmt.Println()
	}

	return 0
}

// buildSynopsis builds a synopsis line for a command
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

// displayFlag formats and displays a single flag
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
	fmt.Printf("    %s %s (%s)\n", flagName, typeInfo, reqInfo)

	// Display usage with indentation
	if flag.Usage != "" {
		fmt.Printf("        %s\n", flag.Usage)
	}

	// Display completion values if available
	if len(flag.Completion) > 0 && flag.Type != "bool" {
		fmt.Printf("        Valid values: %s\n", strings.Join(flag.Completion, ", "))
	}

	fmt.Println()
}

// getSubcommands returns all subcommands for a given parent command
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

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
