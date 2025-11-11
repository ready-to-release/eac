// Main dispatcher for src/commands
//
// Usage: go run . <command> [subcommand] [args...]
//
// Commands auto-discovered via file scanning.
// Convention:
//   File: show-modules.go → Command: "show modules" → Function: ShowModules()
package main

import (
	"fmt"
	"os"
	"strings"
)

// CommandFunc is the signature for all command functions
type CommandFunc func() int

// CommandRegistration holds command metadata
type CommandRegistration struct {
	Func          CommandFunc
	DisplayName   string // "get files" (with spaces)
	CanonicalName string // "get-files" (kebab-case)
}

// commands maps command names to their implementation functions
var commands = map[string]CommandFunc{}

// commandRegistry maps canonical kebab-case names to registrations
var commandRegistry = map[string]*CommandRegistration{}

// InitialWorkingDir stores the working directory when the program started
var InitialWorkingDir string

// Register allows command files to register themselves
// commandName should be in display format with spaces (e.g., "get files")
func Register(commandName string, fn CommandFunc) {
	// Store in original commands map for backward compatibility
	commands[commandName] = fn

	// Derive canonical kebab-case name
	canonicalName := strings.ReplaceAll(commandName, " ", "-")

	// Store in registry with both forms
	commandRegistry[canonicalName] = &CommandRegistration{
		Func:          fn,
		DisplayName:   commandName,
		CanonicalName: canonicalName,
	}
}

// GetCanonicalName returns the kebab-case canonical name for a command
// Input can be either "get files" (with spaces) or "get-files" (kebab-case)
func GetCanonicalName(commandName string) string {
	return strings.ReplaceAll(commandName, " ", "-")
}

// GetCommandByCanonical retrieves a command registration by its canonical name
func GetCommandByCanonical(canonicalName string) *CommandRegistration {
	return commandRegistry[canonicalName]
}


func main() {
	// Check if we have an original PWD from the CLI wrapper
	// If not, use current directory
	InitialWorkingDir = os.Getenv("CLI_ORIGINAL_PWD")
	if InitialWorkingDir == "" {
		var err error
		InitialWorkingDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not determine working directory: %v\n", err)
			os.Exit(1)
		}
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var cmdFunc CommandFunc
	var exists bool

	// Try longest match first for nested commands
	for argCount := len(os.Args) - 1; argCount >= 1; argCount-- {
		testPath := strings.Join(os.Args[1:argCount+1], " ")
		if fn, found := commands[testPath]; found {
			cmdFunc = fn
			exists = true
			break
		}
	}

	if !exists {
		// Check if this is a parent command (has subcommands)
		prefix := strings.Join(os.Args[1:], " ")
		subcommands := getSubcommands(prefix)

		if len(subcommands) > 0 {
			printSubcommandHelp(prefix, subcommands)
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "Error: Command not found: %s\n\n", prefix)
		printUsage()
		os.Exit(1)
	}

	os.Exit(cmdFunc())
}

// getSubcommands returns all commands that start with the given prefix
func getSubcommands(prefix string) []string {
	var subcommands []string
	searchPrefix := prefix
	if prefix != "" {
		searchPrefix = prefix + " "
	}

	for cmdName := range commands {
		if strings.HasPrefix(cmdName, searchPrefix) && cmdName != prefix {
			// Extract just the next part after the prefix
			remainder := strings.TrimPrefix(cmdName, searchPrefix)
			parts := strings.Fields(remainder)
			if len(parts) > 0 {
				subcommand := parts[0]
				// Only add unique subcommands
				found := false
				for _, existing := range subcommands {
					if existing == subcommand {
						found = true
						break
					}
				}
				if !found {
					subcommands = append(subcommands, subcommand)
				}
			}
		}
	}

	// Sort for consistent output
	for i := 0; i < len(subcommands); i++ {
		for j := i + 1; j < len(subcommands); j++ {
			if subcommands[i] > subcommands[j] {
				subcommands[i], subcommands[j] = subcommands[j], subcommands[i]
			}
		}
	}

	return subcommands
}

// printSubcommandHelp prints help for a parent command
func printSubcommandHelp(prefix string, subcommands []string) {
	if prefix == "" {
		fmt.Println("Usage: go run . <command> [subcommand] [args...]")
		fmt.Println("")
		fmt.Println("Available commands:")
	} else {
		fmt.Printf("Usage: go run . %s <subcommand>\n", prefix)
		fmt.Println("")
		fmt.Printf("Available subcommands for '%s':\n", prefix)
	}

	for _, sub := range subcommands {
		fmt.Printf("  %s\n", sub)
	}
}

func printUsage() {
	fmt.Println("Usage: go run . <command> [subcommand] [args...]")
	fmt.Println("")
	fmt.Println("Available commands:")

	var names []string
	for name := range commands {
		names = append(names, name)
	}

	// Sort
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}

	for _, name := range names {
		fmt.Printf("  %s\n", name)
	}
}
