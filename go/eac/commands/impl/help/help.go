// Command: help
// Short: Display help information for commands
// Long: The help command provides comprehensive documentation for all available commands.
// Long: When called without arguments, it lists all commands with their short descriptions.
// Long: When called with a command name, it displays detailed help including description, flags, and usage.
// Long:
// Long: Expected Output:
// Long:   - NAME, SYNOPSIS, DESCRIPTION, COMMANDS, FLAGS sections
// Long:   - Command usage examples
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed information including all subcommands and advanced options
package help

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// commandFlags defines valid flags for the help command

var log = logging.C()

func init() {
	registry.Register(Help)
}

// Help displays help information for commands
func Help() int {
	args := os.Args[2:] // Skip program name and "help"

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
		log.Info("No commands available.")
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
	log.Info("Available commands:")
	log.Info("")

	// Display commands by category
	for _, category := range categoryOrder {
		cmds := categories[category]

		// Skip if this is just a standalone command (no subcommands)
		if len(cmds) == 1 && cmds[0] == category {
			continue
		}

		log.Infof("%s:", category)
		for _, cmdName := range cmds {
			reg := commandRegistry[cmdName]

			// Get short description
			desc := ""
			if reg != nil && reg.Short != "" {
				desc = reg.Short
			}

			// Format: "  command-name    description"
			padding := strings.Repeat(" ", max(2, 30-len(cmdName)))
			log.Infof("  %s%s%s", cmdName, padding, desc)
		}
		log.Info("")
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
		log.Info("Other commands:")
		for _, cmdName := range standaloneCommands {
			reg := commandRegistry[cmdName]

			desc := ""
			if reg != nil && reg.Short != "" {
				desc = reg.Short
			}

			padding := strings.Repeat(" ", max(2, 30-len(cmdName)))
			log.Infof("  %s%s%s", cmdName, padding, desc)
		}
		log.Info("")
	}

	log.Info("Use 'help <command>' for detailed information about a specific command.")
	return 0
}

// showCommandHelp displays detailed help for a specific command
func showCommandHelp(commandName string, verbose bool) int {
	commandRegistry := registry.GetCommandRegistry()

	reg := commandRegistry[commandName]

	// Check for subcommands even if the parent command doesn't exist
	subcommands := getSubcommands(commandName)

	if reg == nil {
		// If command not found, check if it has subcommands
		if len(subcommands) > 0 {
			// Display help for command category with subcommands
			log.Info("NAME")
			log.Infof("    %s - Command category\n", commandName)

			log.Info("DESCRIPTION")
			log.Infof("    The '%s' command category contains the following subcommands:\n", commandName)

			log.Info("COMMANDS")
			for _, subcmd := range subcommands {
				subReg := commandRegistry[subcmd]

				desc := ""
				if subReg != nil && subReg.Short != "" {
					desc = subReg.Short
				}

				// Extract just the subcommand part (e.g., "risk-profile" from "create risk-profile")
				subPart := strings.TrimPrefix(subcmd, commandName+" ")

				// Format with padding
				padding := strings.Repeat(" ", max(2, 30-len(subPart)))
				log.Infof("    %s%s%s", subPart, padding, desc)
			}
			log.Info("")
			log.Infof("Use 'help %s <command>' for detailed information about a specific command.", commandName)
			return 0
		}

		// No command and no subcommands found
		log.Errorf("Error: Command '%s' not found.", commandName)
		log.Error("\nUse 'help' to see all available commands.")
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

	// Display additional info
	if verbose {
		log.Info("ADDITIONAL INFORMATION")
		log.Infof("    Canonical name: %s", reg.CanonicalName)
		log.Info("")
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

	// Add arguments from command metadata
	if reg.Args != "" {
		parts = append(parts, "<"+reg.Args+">")
	}

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
	log.Infof("    %s %s (%s)", flagName, typeInfo, reqInfo)

	// Display usage with indentation
	if flag.Usage != "" {
		// Wrap usage text at reasonable width
		wrapped := wrapText(flag.Usage, 76) // 80 - 8 chars indent
		for _, line := range wrapped {
			log.Infof("        %s", line)
		}
	}

	// Display completion values if available
	if len(flag.Completion) > 0 && flag.Type != "bool" {
		log.Infof("        Valid values: %s", strings.Join(flag.Completion, ", "))
	}

	log.Info("")
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
