// Command: help
// Short: Display help information for commands
// Long: The help command provides comprehensive documentation for all available commands.
// Long: When called without arguments, it lists all commands with their short descriptions.
// Long: When called with a command name, it displays detailed help including description, flags, and usage.
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed information including all subcommands and advanced options
// HasSideEffects: false
package help

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Help)
}

// Help displays help information for commands
func Help() int {
	args := os.Args[2:] // Skip program name and "help"

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

	// Build full command name from parts (e.g., "specs create")
	commandName := strings.Join(commandParts, " ")

	// If command name is provided, show detailed help for that command
	if commandName != "" {
		return showCommandHelp(commandName, verbose)
	}

	// Otherwise, show list of all commands
	return showAllCommands(verbose)
}

// showAllCommands lists all available commands grouped by category
func showAllCommands(verbose bool) int {
	commands := registry.GetCommands()
	commandRegistry := registry.GetCommandRegistry()

	if len(commands) == 0 {
		fmt.Println("No commands available.")
		return 0
	}

	// Group commands by first word (category)
	categories := make(map[string][]string)
	var categoryOrder []string

	for cmdName := range commands {
		parts := strings.SplitN(cmdName, " ", 2)
		category := parts[0]

		if _, exists := categories[category]; !exists {
			categoryOrder = append(categoryOrder, category)
			categories[category] = []string{}
		}

		categories[category] = append(categories[category], cmdName)
	}

	// Sort categories and commands within each category
	sort.Strings(categoryOrder)
	for _, cmds := range categories {
		sort.Strings(cmds)
	}

	// Display header
	fmt.Println("Available commands:")
	fmt.Println()

	// Display commands by category
	for _, category := range categoryOrder {
		cmds := categories[category]

		// Skip if this is just a standalone command (no subcommands)
		if len(cmds) == 1 && cmds[0] == category {
			continue
		}

		fmt.Printf("%s:\n", category)
		for _, cmdName := range cmds {
			canonicalName := registry.GetCanonicalName(cmdName)
			reg := commandRegistry[canonicalName]

			// Get short description
			desc := ""
			if reg != nil {
				if reg.Short != "" {
					desc = reg.Short
				} else if reg.Description != "" {
					desc = reg.Description
				}
			}

			// Format: "  command-name    description"
			padding := strings.Repeat(" ", max(2, 30-len(cmdName)))
			fmt.Printf("  %s%s%s\n", cmdName, padding, desc)
		}
		fmt.Println()
	}

	// Show standalone commands (no category)
	standaloneCommands := []string{}
	for _, category := range categoryOrder {
		cmds := categories[category]
		if len(cmds) == 1 && cmds[0] == category {
			standaloneCommands = append(standaloneCommands, category)
		}
	}

	if len(standaloneCommands) > 0 {
		fmt.Println("Other commands:")
		for _, cmdName := range standaloneCommands {
			canonicalName := registry.GetCanonicalName(cmdName)
			reg := commandRegistry[canonicalName]

			desc := ""
			if reg != nil {
				if reg.Short != "" {
					desc = reg.Short
				} else if reg.Description != "" {
					desc = reg.Description
				}
			}

			padding := strings.Repeat(" ", max(2, 30-len(cmdName)))
			fmt.Printf("  %s%s%s\n", cmdName, padding, desc)
		}
		fmt.Println()
	}

	fmt.Println("Use 'help <command>' for detailed information about a specific command.")
	return 0
}

// showCommandHelp displays detailed help for a specific command
func showCommandHelp(commandName string, verbose bool) int {
	commandRegistry := registry.GetCommandRegistry()
	canonicalName := registry.GetCanonicalName(commandName)

	reg := commandRegistry[canonicalName]
	if reg == nil {
		fmt.Fprintf(os.Stderr, "Error: Command '%s' not found.\n", commandName)
		fmt.Fprintf(os.Stderr, "\nUse 'help' to see all available commands.\n")
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
			canonicalName := registry.GetCanonicalName(subcmd)
			subReg := commandRegistry[canonicalName]

			desc := ""
			if subReg != nil {
				if subReg.Short != "" {
					desc = subReg.Short
				} else if subReg.Description != "" {
					desc = subReg.Description
				}
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

	// Display additional info
	if verbose {
		fmt.Printf("ADDITIONAL INFORMATION\n")
		fmt.Printf("    Canonical name: %s\n", reg.CanonicalName)
		fmt.Printf("    Has side effects: %t\n", reg.HasSideEffects)
		if reg.HasSideEffects {
			fmt.Printf("    (This command modifies repository files)\n")
		}
		fmt.Println()
	}

	return 0
}

// buildSynopsis builds a synopsis line for a command
func buildSynopsis(reg *registry.CommandRegistration) string {
	parts := []string{reg.ActualCommand}

	// Add flags
	if len(reg.Flags) > 0 {
		parts = append(parts, "[flags]")
	}

	// Add arguments placeholder (could be enhanced based on command specifics)
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
		// Wrap usage text at reasonable width
		wrapped := wrapText(flag.Usage, 76) // 80 - 8 chars indent
		for _, line := range wrapped {
			fmt.Printf("        %s\n", line)
		}
	}

	// Display completion values if available
	if len(flag.Completion) > 0 && flag.Type != "bool" {
		fmt.Printf("        Valid values: %s\n", strings.Join(flag.Completion, ", "))
	}

	fmt.Println()
}

// wrapText wraps text to specified width preserving words
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
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
