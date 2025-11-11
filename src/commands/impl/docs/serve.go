// Command: docs serve
// Description: Start or stop MkDocs server
package docs

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ready-to-release/eac/src/commands/impl/docs/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register("docs serve", DocsServe)
}

// DocsServe starts or stops MkDocs server
func DocsServe() int {
	args := os.Args[3:] // Skip "go", "run", ".", "docs", and "serve"

	var noBrowser bool
	var port int = 8000
	var stop bool
	var debug bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--no-browser":
			noBrowser = true
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

		if !noBrowser {
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
	if !noBrowser {
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
