// Command: design serve
// Description: Start or stop Structurizr server for a module
package design

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ready-to-release/eac/src/commands/impl/design/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register("design serve", DesignServe)
}

// DesignServe starts or stops Structurizr server for a module
func DesignServe() int {
	args := os.Args[3:] // Skip "go", "run", ".", and "design" and "serve"

	var module string
	var noBrowser bool
	var port int = 8081
	var stop bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--no-browser":
			noBrowser = true
		case "--stop":
			stop = true
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
			if arg[0] != '-' {
				module = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				return 1
			}
		}
	}

	// Handle --stop flag
	if stop {
		return handleStop(module)
	}

	// Require module argument for starting server
	if module == "" {
		fmt.Println("❌ Error: module name required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run . design serve <module>")
		fmt.Println("\nRun 'go run . design list' to see available modules")
		return 1
	}

	// Create client
	client, err := design.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	// Validate module
	err = client.ValidateModule(module)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("\n💡 Tip: Run 'go run . design list' to see available modules")
		return 1
	}

	// Check if already running
	running, info, err := client.IsRunning(module)
	if err != nil {
		fmt.Printf("❌ Failed to check container status: %v\n", err)
		return 1
	}

	if running && info != nil {
		fmt.Printf("ℹ️  Structurizr is already running for module: %s\n", module)
		fmt.Printf("📊 Architecture documentation: %s\n", info.URL)

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
	fmt.Printf("🚀 Starting Structurizr Lite for module: %s\n", module)

	info, err = client.StartContainer(module, port)
	if err != nil {
		if info != nil {
			fmt.Printf("⚠️  %v\n", err)
			fmt.Printf("📖 Try accessing manually: %s\n", info.URL)
		} else {
			fmt.Printf("❌ Failed to start container: %v\n", err)
			return 1
		}
	}

	// Get module details
	moduleInfo, err := client.GetModuleInfo(module)
	if err != nil {
		// Ignore errors - not critical
	}

	// Display success
	fmt.Printf("\n✅ Structurizr Lite is running for module: %s\n", module)
	fmt.Printf("📊 Architecture documentation: %s\n", info.URL)

	if moduleInfo != nil {
		if moduleInfo.ViewCount > 0 {
			fmt.Printf("\n📈 Available views: %d\n", moduleInfo.ViewCount)
		}
		if moduleInfo.HasDocs {
			fmt.Printf("📚 Documentation sections: %d\n", moduleInfo.DocCount)
		}
		if moduleInfo.HasDecisions {
			fmt.Printf("🎯 Architecture Decisions: %d ADRs\n", moduleInfo.DecisionCount)
		}

		if !moduleInfo.HasDocs {
			fmt.Println("\n⚠️  Module has no documentation sections")
		}
		if !moduleInfo.HasDecisions {
			fmt.Println("⚠️  Module has no architecture decisions")
		}
	}

	// Open browser
	if !noBrowser {
		err = client.OpenBrowser(info.URL)
		if err != nil {
			fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
			fmt.Printf("📖 Please open manually: %s\n", info.URL)
		}
	}

	// Show tips
	fmt.Println("\n💡 Tips:")
	fmt.Println("  • Container will keep running until stopped")
	fmt.Printf("  • Stop with: go run . design serve cli --stop\n")
	fmt.Printf("  • Or: docker stop %s\n", info.Name)
	fmt.Printf("  • View logs: docker logs %s\n", info.Name)

	return 0
}

func handleStop(module string) int {
	client, err := design.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	// If module specified, stop that container
	if module != "" {
		err = client.StopContainer(module)
		if err != nil {
			fmt.Printf("❌ Failed to stop container: %v\n", err)
			return 1
		}
		fmt.Printf("✅ Structurizr container stopped for module: %s\n", module)
		return 0
	}

	// Otherwise, stop all running containers
	modules, err := client.ListModules()
	if err != nil {
		fmt.Printf("❌ Failed to list modules: %v\n", err)
		return 1
	}

	stoppedAny := false
	for _, mod := range modules {
		running, _, err := client.IsRunning(mod.Name)
		if err != nil {
			continue
		}

		if running {
			err = client.StopContainer(mod.Name)
			if err != nil {
				fmt.Printf("⚠️  Failed to stop container for %s: %v\n", mod.Name, err)
			} else {
				fmt.Printf("✅ Stopped container for module: %s\n", mod.Name)
				stoppedAny = true
			}
		}
	}

	if !stoppedAny {
		fmt.Println("ℹ️  No running Structurizr containers found")
	}

	return 0
}
