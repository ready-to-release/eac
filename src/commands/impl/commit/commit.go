// Command: commit
// Description: Commit utilities for generating messages and managing commits
// HasSideEffects: false
package commit

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Commit)
}

// Commit command entry point
func Commit() int {
	args := os.Args[2:] // Skip program name and "commit"

	if len(args) == 0 {
		printCommitUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printCommitUsage()
		return 0
	case "message":
		// Handled by separate registration in message.go
		return 0
	case "reset":
		// Handled by separate registration in reset.go
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printCommitUsage()
		return 1
	}
}

func printCommitUsage() {
	fmt.Println("Commit utilities for generating messages and managing commits")
	fmt.Println()
	fmt.Println("Usage: r2r commit <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  message    Generate AI-powered commit message from staged changes")
	fmt.Println("  reset      Soft reset the latest commit, preserving changes")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate commit message for staged changes")
	fmt.Println("  r2r commit message")
	fmt.Println()
	fmt.Println("  # Generate with debug output")
	fmt.Println("  r2r commit message --debug")
	fmt.Println()
	fmt.Println("  # Reset the latest commit (keeps changes staged)")
	fmt.Println("  r2r commit reset")
	fmt.Println()
	fmt.Println("Use 'r2r commit <subcommand> --help' for more information about a command.")
}
