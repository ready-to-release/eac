// Command: work
// Description: Workspace management for parallel development using git worktrees
package work

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Work)
}

// Work command entry point
func Work() int {
	args := os.Args[2:] // Skip program name and "work"

	if len(args) == 0 {
		printWorkUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printWorkUsage()
		return 0
	case "create", "list", "commit", "pull", "merge", "pr", "remove":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("Error: unknown subcommand: %s", args[0])
		log.Info("")
		printWorkUsage()
		return 1
	}
}

func printWorkUsage() {
	log.Info("Workspace management for parallel development using git worktrees")
	log.Info("")
	log.Info("Usage: r2r work <subcommand> [args...]")
	log.Info("")
	log.Info("Workspace Lifecycle:")
	log.Info("  create <branch>           Create new workspace for parallel development")
	log.Info("  list                      List all workspaces and their status")
	log.Info("  remove [branch]           Remove workspace and optionally delete branches")
	log.Info("")
	log.Info("Development Workflow:")
	log.Info("  commit                    Commit changes with AI-generated messages")
	log.Info("  pull                      Sync workspace with latest main via rebase")
	log.Info("")
	log.Info("Completion:")
	log.Info("  merge                     Merge workspace to main (squash by default)")
	log.Info("  pr                        Create pull request with AI-generated description")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Create workspace")
	log.Info("  r2r work create feature/authentication")
	log.Info("")
	log.Info("  # Make changes and commit")
	log.Info("  r2r work commit --all")
	log.Info("")
	log.Info("  # Sync with main")
	log.Info("  r2r work pull")
	log.Info("")
	log.Info("  # Merge back to main")
	log.Info("  r2r work merge")
	log.Info("")
	log.Info("  # Or create PR for review")
	log.Info("  r2r work pr")
	log.Info("")
	log.Info("  # List all workspaces")
	log.Info("  r2r work list")
	log.Info("")
	log.Info("Use 'r2r work <subcommand> --help' for more information about a command.")
}
