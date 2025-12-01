// Command: work
// Description: Workspace management for parallel development using git worktrees
package work

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
)

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
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printWorkUsage()
		return 1
	}
}

func printWorkUsage() {
	fmt.Println("Workspace management for parallel development using git worktrees")
	fmt.Println()
	fmt.Println("Usage: r2r work <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Workspace Lifecycle:")
	fmt.Println("  create <branch>           Create new workspace for parallel development")
	fmt.Println("  list                      List all workspaces and their status")
	fmt.Println("  remove [branch]           Remove workspace and optionally delete branches")
	fmt.Println()
	fmt.Println("Development Workflow:")
	fmt.Println("  commit                    Commit changes with AI-generated messages")
	fmt.Println("  pull                      Sync workspace with latest main via rebase")
	fmt.Println()
	fmt.Println("Completion:")
	fmt.Println("  merge                     Merge workspace to main (squash by default)")
	fmt.Println("  pr                        Create pull request with AI-generated description")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create workspace")
	fmt.Println("  r2r work create feature/authentication")
	fmt.Println()
	fmt.Println("  # Make changes and commit")
	fmt.Println("  r2r work commit --all")
	fmt.Println()
	fmt.Println("  # Sync with main")
	fmt.Println("  r2r work pull")
	fmt.Println()
	fmt.Println("  # Merge back to main")
	fmt.Println("  r2r work merge")
	fmt.Println()
	fmt.Println("  # Or create PR for review")
	fmt.Println("  r2r work pr")
	fmt.Println()
	fmt.Println("  # List all workspaces")
	fmt.Println("  r2r work list")
	fmt.Println()
	fmt.Println("Use 'r2r work <subcommand> --help' for more information about a command.")
}
