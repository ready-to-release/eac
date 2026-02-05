// Main dispatcher for go/cli/eac
//
// Usage: go run . <command> [subcommand] [args...]
//
// Commands auto-discovered via file scanning.
// Convention:
//
//	File: show-modules.go → Command: "show modules" → Function: ShowModules()
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/help"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/environments"
)

// InitialWorkingDir stores the working directory when the program started.
var InitialWorkingDir string

// nl returns the platform-appropriate line ending.
func nl() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func main() {
	// Global panic handler with full stack trace
	defer func() {
		if r := recover(); r != nil {
			newline := nl()
			fmt.Fprintf(os.Stderr, "%s=== PANIC: Unhandled Exception ===%s", newline, newline)
			fmt.Fprintf(os.Stderr, "Error: %v%s%s", r, newline, newline)
			fmt.Fprintf(os.Stderr, "Stack Trace:%s", newline)

			// Print full stack trace
			buf := make([]byte, 4096)
			for {
				n := runtime.Stack(buf, false)
				if n < len(buf) {
					// Replace \n with platform-appropriate line ending in stack trace
					stackStr := strings.ReplaceAll(string(buf[:n]), "\n", newline)
					fmt.Fprintf(os.Stderr, "%s%s", stackStr, newline)
					break
				}
				buf = make([]byte, len(buf)*2)
			}

			fmt.Fprintf(os.Stderr, "%s=== End Stack Trace ===%s", newline, newline)
			os.Exit(2)
		}
	}()

	// Check if we have an original PWD from the CLI wrapper
	// If not, use current directory
	InitialWorkingDir = os.Getenv(environments.EnvR2RPWD)
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

	var cmdFunc registry.CommandFunc
	var exists bool

	// Try longest match first for nested commands
	commands := registry.GetCommands()
	for argCount := len(os.Args) - 1; argCount >= 1; argCount-- {
		testPath := strings.Join(os.Args[1:argCount+1], " ")
		if fn, found := commands[testPath]; found {
			cmdFunc = fn
			exists = true
			break
		}
	}

	if exists {
		// Check if --help or -h is in arguments after the command
		// Only intercept help if the command has complete registry metadata
		// (parent commands like get, show, work handle their own help)
		for _, arg := range os.Args[1:] {
			if arg == "--help" || arg == "-h" {
				// Check if this command has complete help metadata
				cmdRegistry := registry.GetCommandRegistry()
				var cmdName string
				for argCount := len(os.Args) - 1; argCount >= 1; argCount-- {
					testPath := strings.Join(os.Args[1:argCount+1], " ")
					if strings.HasPrefix(os.Args[argCount], "-") {
						continue
					}
					if _, found := cmdRegistry[testPath]; found {
						cmdName = testPath
						break
					}
				}
				if cmdName != "" {
					if reg := cmdRegistry[cmdName]; reg != nil && reg.Short != "" {
						// Has complete metadata, print help from registry
						printCommandHelp(cmdFunc)
						os.Exit(0)
					}
				}
				// Let the command handle its own help
				break
			}
		}
	}

	if !exists {
		// Check if this is a parent command (has subcommands)
		// We need to find where the command args end and flags/args begin
		cmdArgCount := 1
		for i := 1; i < len(os.Args); i++ {
			if strings.HasPrefix(os.Args[i], "-") {
				// This is a flag, so previous args were the command
				break
			}
			cmdArgCount = i
		}

		prefix := strings.Join(os.Args[1:cmdArgCount+1], " ")
		subcommands := getSubcommands(prefix)

		if len(subcommands) > 0 {
			// Handle default behaviors for specific verbs
			switch prefix {
			case "build":
				// Default: build (all modules)
				if fn, found := commands["build"]; found {
					exitCode := fn()
					os.Exit(exitCode)
				}
			case "test":
				// Default: test (all modules)
				if fn, found := commands["test"]; found {
					exitCode := fn()
					os.Exit(exitCode)
				}
			}

			printSubcommandHelp(prefix, subcommands)
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "Error: Command not found: %s\n\n", prefix)
		printUsage()
		os.Exit(1)
	}

	exitCode := cmdFunc()
	os.Exit(exitCode)
}

// getSubcommands returns all commands that start with the given prefix.
func getSubcommands(prefix string) []string {
	var subcommands []string
	searchPrefix := prefix
	if prefix != "" {
		searchPrefix = prefix + " "
	}

	commands := registry.GetCommands()
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

// printSubcommandHelp prints help for a parent command.
func printSubcommandHelp(prefix string, subcommands []string) {
	// Check if we have registry metadata for this parent command
	cmdRegistry := registry.GetCommandRegistry()
	if reg := cmdRegistry[prefix]; reg != nil && reg.IsParent && len(reg.Subcommands) > 0 {
		// Use registry-based help
		help.PrintHelp(os.Stdout, reg, cmdRegistry)
		return
	}

	// Fallback to simple list (for commands not yet migrated)
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
	commands := registry.GetCommands()
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

// printCommandHelp prints help information for a command using its registration metadata.
func printCommandHelp(_ registry.CommandFunc) {
	// Find command by reconstructing the command name from args
	cmdRegistry := registry.GetCommandRegistry()
	var cmdName string
	for argCount := len(os.Args) - 1; argCount >= 1; argCount-- {
		testPath := strings.Join(os.Args[1:argCount+1], " ")
		// Skip if this looks like a flag
		if strings.HasPrefix(os.Args[argCount], "-") {
			continue
		}
		if _, found := cmdRegistry[testPath]; found {
			cmdName = testPath
			break
		}
	}

	if cmdName == "" {
		fmt.Println("No help available.")
		return
	}

	reg := cmdRegistry[cmdName]
	if reg == nil {
		fmt.Println("No help available.")
		return
	}

	// Use unified help printer
	help.PrintHelp(os.Stdout, reg, cmdRegistry)
}
