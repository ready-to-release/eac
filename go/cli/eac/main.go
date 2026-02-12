// Main dispatcher for go/cli/eac
//
// Usage: eac <command> [subcommand] [args...]
//
// Commands registered via CommandPort implementations.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/clibase/executor"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

// initialWorkingDir stores the working directory when the program started.
var initialWorkingDir string

// nl returns the platform-appropriate line ending.
func nl() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func main() {
	defer recoverPanic()
	initWorkingDir()
	bootstrapProviders()
	os.Exit(dispatch())
}

// recoverPanic is a deferred panic handler that prints a full stack trace.
func recoverPanic() {
	if r := recover(); r != nil {
		newline := nl()
		fmt.Fprintf(os.Stderr, "%s=== PANIC: Unhandled Exception ===%s", newline, newline)
		fmt.Fprintf(os.Stderr, "Error: %v%s%s", r, newline, newline)
		fmt.Fprintf(os.Stderr, "Stack Trace:%s", newline)

		buf := make([]byte, 4096)
		for {
			n := runtime.Stack(buf, false)
			if n < len(buf) {
				stackStr := strings.ReplaceAll(string(buf[:n]), "\n", newline)
				fmt.Fprintf(os.Stderr, "%s%s", stackStr, newline)
				break
			}
			buf = make([]byte, len(buf)*2)
		}

		fmt.Fprintf(os.Stderr, "%s=== End Stack Trace ===%s", newline, newline)
		os.Exit(2)
	}
}

// initWorkingDir resolves the initial working directory from the CLI wrapper
// environment variable or the current process working directory.
func initWorkingDir() {
	initialWorkingDir = os.Getenv(environments.EnvCLIEPWD)
	if initialWorkingDir == "" {
		var err error
		initialWorkingDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not determine working directory: %v\n", err)
			os.Exit(1)
		}
	}
}

// bootstrapProviders wires dependency providers before any command runs.
func bootstrapProviders() {
	// Git remote provider: lets config package resolve remote URLs via go-git.
	config.SetGitRemoteProvider(func(repoRoot, remoteName string) (string, error) {
		repo, err := git.NewManager(nil).Open(repoRoot)
		if err != nil {
			return "", err
		}
		return repo.RemoteURL(remoteName)
	})

	// Initialize tool system so simple commands (work, pipeline, release) that
	// bypass cmdframework's phaseInitDeferred still get a working ToolSystem.
	// Framework commands re-initialize in phaseInitDeferred; this is safe to call twice.
	if repoRoot, err := repository.GetRepositoryRoot(""); err == nil {
		configRoot := filepath.Join(repoRoot, ".eac")
		_ = tool.InitializeGlobalBridges(repoRoot, configRoot)
	}

	// Report dependencies: routes gh commands through the tool registry.
	reports.InitDeps(reports.DefaultReportDeps(gh.New(tool.GlobalToolSystem(), initialWorkingDir)))
}

// dispatch resolves the command from os.Args and executes it.
// Returns the process exit code.
func dispatch() int {
	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	reg := buildCommandRegistry()
	exec := executor.New(reg)

	// Resolve command using longest-match
	cmdName, found := resolveCommand(os.Args[1:], reg)

	if found {
		if hasHelpFlag(os.Args[1:]) {
			cmd, _ := reg.Get(cmdName)
			printCommandHelp(cmd, reg)
			return 0
		}
		return exec.Execute(context.Background(), cmdName, os.Args[1:])
	}

	// Command not found - check if it's a parent command with subcommands
	prefix := resolvePrefix(os.Args[1:])
	subs := reg.Subcommands(prefix)

	if len(subs) > 0 {
		cmd, hasCmd := reg.Get(prefix)
		if hasCmd {
			if _, isSimple := cmd.(core.SimpleCommandPort); isSimple {
				return exec.Execute(context.Background(), prefix, os.Args[1:])
			}
		}
		printSubcommandHelp(prefix, subs)
		return 0
	}

	fmt.Fprintf(os.Stderr, "Error: Command not found: %s\n\n", prefix)
	printUsage()
	return 1
}

// resolveCommand finds the longest matching command name from args.
// Returns the command name and whether it was found.
func resolveCommand(args []string, reg core.CommandRegistryPort) (string, bool) {
	for argCount := len(args); argCount >= 1; argCount-- {
		testPath := strings.Join(args[:argCount], " ")
		if _, ok := reg.Get(testPath); ok {
			return testPath, true
		}
	}
	return "", false
}

// resolvePrefix finds the longest non-flag prefix from args.
func resolvePrefix(args []string) string {
	var parts []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		parts = append(parts, arg)
	}
	return strings.Join(parts, " ")
}

// hasHelpFlag checks if --help or -h is present in args.
func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
		// Stop at -- separator
		if arg == "--" {
			return false
		}
	}
	return false
}

// printCommandHelp prints help information for a command using port metadata.
func printCommandHelp(cmd core.CommandPort, reg core.CommandRegistryPort) {
	if cmd == nil {
		fmt.Println("No help available.")
		return
	}

	meta := cmd.Metadata()

	// NAME
	fmt.Println("NAME")
	fmt.Printf("    %s - %s\n\n", cmd.Name(), meta.Short)

	// SYNOPSIS
	fmt.Println("SYNOPSIS")
	synopsis := cmd.Name()
	if len(meta.Flags) > 0 {
		synopsis += " [flags]"
	}
	if meta.Args != "" {
		synopsis += " <" + meta.Args + ">"
	}
	fmt.Printf("    %s\n\n", synopsis)

	// DESCRIPTION
	if meta.Long != "" {
		fmt.Println("DESCRIPTION")
		for _, line := range strings.Split(meta.Long, "\n") {
			if line == "" {
				fmt.Println()
			} else {
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println()
	}

	// COMMANDS (subcommands)
	subs := reg.Subcommands(cmd.Name())
	if len(subs) > 0 {
		fmt.Println("COMMANDS")
		for _, sub := range subs {
			subMeta := sub.Metadata()
			subPart := strings.TrimPrefix(sub.Name(), cmd.Name()+" ")
			padding := strings.Repeat(" ", max(2, 24-len(subPart)))
			fmt.Printf("    %s%s%s\n", subPart, padding, subMeta.Short)
		}
		fmt.Println()
	}

	// FLAGS
	if len(meta.Flags) > 0 {
		fmt.Println("FLAGS")
		for _, flag := range meta.Flags {
			flagName := "--" + flag.Name
			if flag.Shorthand != "" {
				flagName = "-" + flag.Shorthand + ", " + flagName
			}

			typeInfo := ""
			if flag.Type != "" && flag.Type != "bool" {
				typeInfo = fmt.Sprintf(" <%s>", flag.Type)
			}

			req := "optional"
			if flag.Required {
				req = "required"
			}
			if flag.DefaultValue != "" {
				req += ", default: " + flag.DefaultValue
			}

			fmt.Printf("    %s%s (%s)\n", flagName, typeInfo, req)
			if flag.Usage != "" {
				fmt.Printf("        %s\n", flag.Usage)
			}
			fmt.Println()
		}
	}

	// EXAMPLES
	if len(meta.Examples) > 0 {
		fmt.Println("EXAMPLES")
		for _, ex := range meta.Examples {
			fmt.Printf("    %s\n", ex)
		}
		fmt.Println()
	}
}

// printSubcommandHelp prints help for a parent command showing its subcommands.
func printSubcommandHelp(prefix string, subs []core.CommandPort) {
	fmt.Printf("Usage: eac %s <subcommand> [args...]\n\n", prefix)
	fmt.Printf("Available subcommands for '%s':\n", prefix)

	// Sort by name
	names := make([]string, len(subs))
	nameMap := make(map[string]core.CommandPort, len(subs))
	for i, sub := range subs {
		names[i] = sub.Name()
		nameMap[sub.Name()] = sub
	}
	slices.Sort(names)

	for _, name := range names {
		sub := nameMap[name]
		subPart := strings.TrimPrefix(name, prefix+" ")
		desc := sub.Metadata().Short
		padding := strings.Repeat(" ", max(2, 24-len(subPart)))
		fmt.Printf("  %s%s%s\n", subPart, padding, desc)
	}

	fmt.Printf("\nUse 'eac help %s <command>' for detailed information about a specific command.\n", prefix)
}

func printUsage() {
	fmt.Println("Usage: eac <command> [subcommand] [args...]")
	fmt.Println()
	fmt.Println("Use 'eac help' for a list of available commands.")
}
