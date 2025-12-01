// Command: commit
// Description: Commit utilities for generating messages and managing commits
package commit

import (
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
)

var log = logging.C()

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
		log.Errorf("Error: unknown subcommand: %s\n", args[0])
		printCommitUsage()
		return 1
	}
}

func printCommitUsage() {
	log.Info("Commit utilities for generating messages and managing commits")
	log.Info("")
	log.Info("Usage: r2r commit <subcommand> [args...]")
	log.Info("")
	log.Info("Subcommands:")
	log.Info("  message    Generate AI-powered commit message from staged changes")
	log.Info("  reset      Soft reset the latest commit, preserving changes")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Generate commit message for staged changes")
	log.Info("  r2r commit message")
	log.Info("")
	log.Info("  # Generate with debug output")
	log.Info("  r2r commit message --debug")
	log.Info("")
	log.Info("  # Reset the latest commit (keeps changes staged)")
	log.Info("  r2r commit reset")
	log.Info("")
	log.Info("Use 'r2r commit <subcommand> --help' for more information about a command.")
}
