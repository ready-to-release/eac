// Command: docs
// Description: Project documentation tools using MkDocs - Main command
package docs

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register("docs", Docs)
}

// Docs command entry point
func Docs() int {
	args := os.Args[2:] // Skip "go" and "run" and "."

	if len(args) == 0 {
		printDocsUsage()
		return 1
	}

	// Check for subcommands
	switch args[0] {
	case "serve":
		// Handled by separate registration
		return 0
	case "--help", "-h":
		printDocsUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printDocsUsage()
		return 1
	}
}

func printDocsUsage() {
	fmt.Println("Usage: go run . docs <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  serve  Start or stop MkDocs documentation server")
	fmt.Println()
	fmt.Println("Serve options:")
	fmt.Println("  --no-auto-open-link    Don't open browser automatically")
	fmt.Println("  --port, -p <port>      Port for MkDocs server (default: 8000)")
	fmt.Println("  --debug                Stream container logs to stdout")
	fmt.Println("  --stop                 Stop the running container")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run . docs serve")
	fmt.Println("  go run . docs serve --no-auto-open-link")
	fmt.Println("  go run . docs serve --port 8001")
	fmt.Println("  go run . docs serve --debug")
	fmt.Println("  go run . docs serve --stop")
}
