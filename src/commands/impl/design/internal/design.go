package design

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	designNoOpen bool
	designPort   int
	designList   bool
	designStop   bool
)

var designCmd = &cobra.Command{
	Use:   "design [module]",
	Short: "View architecture documentation using Structurizr",
	Long: `View architecture documentation using Structurizr Lite.

Starts a Docker container running Structurizr Lite to display C4 architecture
diagrams for the specified module. The container will continue running until
you explicitly stop it with 'r2r design --stop'.`,
	Example: `  # View CLI architecture
  r2r design src-cli

  # List available modules
  r2r design --list

  # View without opening browser
  r2r design src-cli --no-open

  # Use custom port
  r2r design src-cli --port 8082

  # Stop running container
  r2r design --stop`,
	Args: cobra.MaximumNArgs(1),
	Run:  runDesignCommand,
}

// NewCommand returns the design command
func NewCommand() *cobra.Command {
	designCmd.Flags().BoolVar(&designNoOpen, "no-open", false, "Don't open browser automatically")
	designCmd.Flags().IntVarP(&designPort, "port", "p", 8081, "Port for Structurizr Lite")
	designCmd.Flags().BoolVarP(&designList, "list", "l", false, "List available modules with documentation")
	designCmd.Flags().BoolVar(&designStop, "stop", false, "Stop running Structurizr container")
	return designCmd
}

func runDesignCommand(cmd *cobra.Command, args []string) {
	// Handle --list flag
	if designList {
		handleListModules()
		return
	}

	// Handle --stop flag
	if designStop {
		handleStopContainer(args)
		return
	}

	// Require module argument if not listing or stopping
	if len(args) == 0 {
		fmt.Println("❌ Error: module name required")
		fmt.Println("\nUsage:")
		fmt.Println("  r2r design <module>")
		fmt.Println("\nRun 'r2r design --list' to see available modules")
		os.Exit(1)
	}

	module := args[0]

	// Create Structurizr client
	client, err := NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Validate module
	err = client.ValidateModule(module)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("\n💡 Tip: Run 'r2r design --list' to see available modules")
		os.Exit(1)
	}

	// Check if container already running
	running, info, err := client.IsRunning(module)
	if err != nil {
		fmt.Printf("❌ Failed to check container status: %v\n", err)
		os.Exit(1)
	}

	if running && info != nil {
		fmt.Printf("ℹ️  Structurizr is already running for module: %s\n", module)
		fmt.Printf("📊 Architecture documentation: %s\n", info.URL)

		// Open browser if requested
		if !designNoOpen {
			err = client.OpenBrowser(info.URL)
			if err != nil {
				fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
				fmt.Printf("📖 Please open manually: %s\n", info.URL)
			}
		}

		return
	}

	// Start container
	fmt.Printf("🚀 Starting Structurizr Lite for module: %s\n", module)

	info, err = client.StartContainer(module, designPort)
	if err != nil {
		// Check if this is a health check timeout warning
		if info != nil {
			fmt.Printf("⚠️  %v\n", err)
			fmt.Printf("📖 Try accessing manually: %s\n", info.URL)
		} else {
			fmt.Printf("❌ Failed to start container: %v\n", err)
			os.Exit(1)
		}
	}

	// Get module details for display
	moduleInfo, err := client.GetModuleInfo(module)
	if err != nil {
		// Ignore errors getting module info - it's not critical
	}

	// Display success message
	fmt.Printf("\n✅ Structurizr Lite is running for module: %s\n", module)
	fmt.Printf("📊 Architecture documentation: %s\n", info.URL)

	// Show available content if we have module info
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

		// Warnings for missing content
		if !moduleInfo.HasDocs {
			fmt.Println("\n⚠️  Module has no documentation sections")
		}
		if !moduleInfo.HasDecisions {
			fmt.Println("⚠️  Module has no architecture decisions")
		}
	}

	// Open browser
	if !designNoOpen {
		err = client.OpenBrowser(info.URL)
		if err != nil {
			fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
			fmt.Printf("📖 Please open manually: %s\n", info.URL)
		}
	}

	// Show tips
	fmt.Println("\n💡 Tips:")
	fmt.Println("  • Container will keep running until stopped")
	fmt.Printf("  • Stop with: r2r design --stop\n")
	fmt.Printf("  • Or: docker stop %s\n", info.Name)
	fmt.Printf("  • View logs: docker logs %s\n", info.Name)
}

func handleListModules() {
	client, err := NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	modules, err := client.ListModules()
	if err != nil {
		fmt.Printf("❌ Failed to list modules: %v\n", err)
		os.Exit(1)
	}

	if len(modules) == 0 {
		fmt.Println("ℹ️  No modules with architecture documentation found")
		fmt.Println("\nExpected location: docs/reference/design/<module>/workspace.dsl")
		return
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Println("Available modules with architecture documentation:")
	fmt.Println()
	fmt.Fprintln(w, "MODULE\tSTATUS\tVIEWS\tDOCS\tDECISIONS\tPATH")
	fmt.Fprintln(w, "──────\t──────\t─────\t────\t─────────\t────")

	for _, module := range modules {
		docsCount := "-"
		if module.HasDocs {
			docsCount = fmt.Sprintf("%d", module.DocCount)
		}

		decisionsCount := "-"
		if module.HasDecisions {
			decisionsCount = fmt.Sprintf("%d", module.DecisionCount)
		}

		viewsCount := fmt.Sprintf("%d", module.ViewCount)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			module.Name,
			module.GetStatus(),
			viewsCount,
			docsCount,
			decisionsCount,
			module.Path,
		)
	}

	w.Flush()

	fmt.Println("\n💡 To view documentation:")
	fmt.Println("  r2r design <module>")
}

func handleStopContainer(args []string) {
	client, err := NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// If module specified, stop that specific container
	if len(args) > 0 {
		module := args[0]
		err = client.StopContainer(module)
		if err != nil {
			fmt.Printf("❌ Failed to stop container: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Structurizr container stopped for module: %s\n", module)
		return
	}

	// Otherwise, try to find and stop any running Structurizr containers
	modules, err := client.ListModules()
	if err != nil {
		fmt.Printf("❌ Failed to list modules: %v\n", err)
		os.Exit(1)
	}

	stoppedAny := false
	for _, module := range modules {
		running, _, err := client.IsRunning(module.Name)
		if err != nil {
			continue
		}

		if running {
			err = client.StopContainer(module.Name)
			if err != nil {
				fmt.Printf("⚠️  Failed to stop container for %s: %v\n", module.Name, err)
			} else {
				fmt.Printf("✅ Stopped container for module: %s\n", module.Name)
				stoppedAny = true
			}
		}
	}

	if !stoppedAny {
		fmt.Println("ℹ️  No running Structurizr containers found")
	}
}
