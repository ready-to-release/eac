// Command: docs
// Description: Serve project documentation using MkDocs
// Usage: docs serve [--no-auto-open-link] [--port <port>] [--stop]
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ready-to-release/eac/src/commands/docs"
)

func init() {
	Register("docs", Docs)
	Register("docs serve", DocsServe)
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

// DocsServe starts or stops MkDocs server
func DocsServe() int {
	args := os.Args[3:] // Skip "go", "run", ".", "docs", and "serve"

	var noAutoOpenLink bool
	var port int = 8000
	var stop bool
	var debug bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--no-auto-open-link":
			noAutoOpenLink = true
		case "--stop":
			stop = true
		case "--debug":
			debug = true
		case "--port", "-p":
			if i+1 < len(args) {
				i++
				p, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid port number: %s\n", args[i])
					return 1
				}
				port = p
			} else {
				fmt.Fprintf(os.Stderr, "Error: --port requires a value\n")
				return 1
			}
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			return 1
		}
	}

	// Handle --stop flag
	if stop {
		return handleDocsStop()
	}

	// Create client
	client, err := docs.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	// Check if already running
	running, info, err := client.IsRunning()
	if err != nil {
		fmt.Printf("❌ Failed to check container status: %v\n", err)
		return 1
	}

	if running && info != nil {
		fmt.Printf("ℹ️  MkDocs is already running\n")
		fmt.Printf("📚 Documentation: %s\n", info.URL)

		if !noAutoOpenLink {
			err = client.OpenBrowser(info.URL)
			if err != nil {
				fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
				fmt.Printf("📖 Please open manually: %s\n", info.URL)
			}
		}
		return 0
	}

	// Start container
	fmt.Printf("🚀 Starting MkDocs documentation server\n")

	info, err = client.StartContainer(port)
	if err != nil {
		if info != nil {
			fmt.Printf("⚠️  %v\n", err)
			fmt.Printf("📖 Try accessing manually: %s\n", info.URL)
		} else {
			fmt.Printf("❌ Failed to start container: %v\n", err)
			return 1
		}
	}

	// Display success
	fmt.Printf("\n✅ MkDocs documentation server is running\n")
	fmt.Printf("📚 Documentation: %s\n", info.URL)

	// Open browser
	if !noAutoOpenLink {
		err = client.OpenBrowser(info.URL)
		if err != nil {
			fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
			fmt.Printf("📖 Please open manually: %s\n", info.URL)
		}
	}

	// Show tips
	if !debug {
		fmt.Println("\n💡 Tips:")
		fmt.Println("  • Container will keep running until stopped")
		fmt.Printf("  • Stop with: go run . docs serve --stop\n")
		fmt.Printf("  • Or: docker stop %s\n", info.Name)
		fmt.Printf("  • View logs: docker logs %s\n", info.Name)
	}

	// Stream logs if debug mode
	if debug {
		fmt.Println("\n🔍 Debug mode: Streaming MkDocs logs (Press Ctrl+C to exit)")
		fmt.Println("─────────────────────────────────────────────────────────────")
		err = client.StreamLogs()
		if err != nil {
			fmt.Printf("\n❌ Error streaming logs: %v\n", err)
			return 1
		}
	}

	return 0
}

func handleDocsStop() int {
	client, err := docs.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	err = client.StopContainer()
	if err != nil {
		fmt.Printf("❌ Failed to stop container: %v\n", err)
		return 1
	}
	fmt.Printf("✅ MkDocs documentation server stopped\n")
	return 0
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
