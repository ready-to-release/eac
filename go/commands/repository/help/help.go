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
	"github.com/ready-to-release/eac/go/core/logging"
)

type helpCommand struct{}

var _ core.SimpleCommandPort = (*helpCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&helpCommand{},
	}
}

func (c *helpCommand) Name() string { return "help" }

func (c *helpCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "help",
		Short:         "Display help information for commands",
		Long: "The help command provides comprehensive documentation for all available commands.\nWhen called without arguments, it lists all commands with their short descriptions.\nWhen called with a command name, it displays detailed help including description, flags, and usage.",
		Notes: "Expected Output:\n  - NAME, SYNOPSIS, DESCRIPTION, COMMANDS, FLAGS sections\n  - Command usage examples",
		Flags: []core.FlagSpec{
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show detailed information including all subcommands and advanced options"},
		},
	}
}

func (c *helpCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Help()
}

var log = logging.C()

// Help displays help information for commands.
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

// showAllCommands lists all available commands grouped by category.
func showAllCommands(verbose bool) int {
	reg := registry.Global()
	allCmds := reg.All()

	if len(allCmds) == 0 {
		log.Info("No commands available.")
		return 0
	}

	// Group commands by first word (category)
	categories := make(map[string][]string)
	var categoryOrder []string

	for _, cmd := range allCmds {
		cmdName := cmd.Name()
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
			// Look up the command to get its short description
			cmdPort, ok := reg.Get(cmdName)
			desc := ""
			if ok {
				desc = cmdPort.Metadata().Short
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
			cmdPort, ok := reg.Get(cmdName)
			desc := ""
			if ok {
				desc = cmdPort.Metadata().Short
			}

			padding := strings.Repeat(" ", max(2, 30-len(cmdName)))
			log.Infof("  %s%s%s", cmdName, padding, desc)
		}
		log.Info("")
	}

	log.Info("Use 'help <command>' for detailed information about a specific command.")
	return 0
}

// showCommandHelp displays detailed help for a specific command.
func showCommandHelp(commandName string, verbose bool) int {
	reg := registry.Global()

	cmdPort, found := reg.Get(commandName)

	// Check for subcommands even if the parent command doesn't exist
	subEntries := getSubcommandEntries(commandName)

	if !found {
		// If command not found, check if it has subcommands
		if len(subEntries) > 0 {
			// Display help for command category with subcommands
			log.Info("NAME")
			log.Infof("    %s - Command category\n", commandName)

			log.Info("DESCRIPTION")
			log.Infof("    The '%s' command category contains the following subcommands:\n", commandName)

			log.Info("COMMANDS")
			for _, entry := range subEntries {
				desc := entry.Cmd.Metadata().Short
				if entry.Key != entry.Cmd.Name() {
					desc = desc + " (-> " + entry.Cmd.Name() + ")"
				}

				// Extract just the subcommand part (e.g., "risk-profile" from "create risk-profile")
				subPart := strings.TrimPrefix(entry.Key, commandName+" ")

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
	if len(subEntries) > 0 {
		log.Info("COMMANDS")
		for _, entry := range subEntries {
			desc := entry.Cmd.Metadata().Short
			if entry.Key != entry.Cmd.Name() {
				desc = desc + " (-> " + entry.Cmd.Name() + ")"
			}

			// Extract just the subcommand part (e.g., "create" from "work create")
			subPart := strings.TrimPrefix(entry.Key, commandName+" ")

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
	parts := []string{cmd.Name()}

	// Add flags
	if len(meta.Flags) > 0 {
		parts = append(parts, "[flags]")
	}

	// Add arguments from command metadata
	if meta.Args != "" {
		parts = append(parts, "<"+meta.Args+">")
	}

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
		// Wrap usage text at reasonable width
		wrapped := wrapText(flag.Usage, 76) // 80 - 8 chars indent
		for _, line := range wrapped {
			log.Infof("        %s", line)
		}
	}

	log.Info("")
}

// wrapText wraps text to specified width preserving words.
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

// getSubcommandEntries returns all direct subcommand entries for a given parent command.
func getSubcommandEntries(parentCommand string) []core.SubcommandEntry {
	return registry.Global().SubcommandEntries(parentCommand)
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
