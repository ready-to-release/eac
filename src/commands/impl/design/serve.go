// Command: design serve
// Description: View architecture diagrams in browser using Structurizr Lite (Docker)
// Short: View architecture diagrams in browser using Structurizr Lite (Docker)
// Long: Launches a Docker container running Structurizr Lite web viewer and opens your default browser.
// Long: The viewer provides interactive C4 model diagrams (system context, containers, components) defined
// Long: in workspace.dsl. When you run this command, it generates workspace.json and .structurizr/ files
// Long: The viewer runs on a dynamically allocated port (9000-9999) and updates automatically when you edit the DSL file.
// Usage: design serve <module>
// HasSideEffects: true
package design

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	design "github.com/ready-to-release/eac/src/commands/impl/design/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(DesignServe)
}

// DesignServe starts Structurizr Lite viewer for a module
func DesignServe() int {
	args := os.Args[3:] // Skip "go", "run", ".", "design", and "serve"

	if len(args) == 0 {
		fmt.Println("❌ Error: module name required")
		fmt.Println()
		printServeUsage()
		return 1
	}

	// Parse flags and module name
	var module string
	var autoStop bool

	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printServeUsage()
			return 0
		}
		if arg == "--force" || arg == "-f" {
			autoStop = true
			continue
		}
		// First non-flag argument is the module name
		if !strings.HasPrefix(arg, "-") && module == "" {
			module = arg
		}
	}

	if module == "" {
		fmt.Println("❌ Error: module name required")
		fmt.Println()
		printServeUsage()
		return 1
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Printf("❌ Failed to find repository root: %v\n", err)
		return 1
	}

	// Validate module name for security
	if err := design.ValidateModuleName(module); err != nil {
		fmt.Printf("❌ Invalid module name: %v\n", err)
		return 1
	}

	// Load module contracts and validate moniker exists (same as build command)
	moduleReport, err := reports.GetModuleContracts(repoRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load module contracts: %v\n", err)
		return 1
	}

	mod, exists := moduleReport.Registry.Get(module)
	if !exists {
		fmt.Fprintf(os.Stderr, "❌ Module not found: %s\n\nAvailable modules:\n%s\n",
			module, formatModuleList(moduleReport))
		return 1
	}

	// Use validated moniker
	module = mod.Moniker

	// Check if workspace exists
	workspacePath := filepath.Join(repoRoot, "specs", module, ".design", "workspace.dsl")
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		fmt.Printf("❌ Workspace not found: %s\n", workspacePath)
		fmt.Printf("\n💡 Create one first with:\n")
		fmt.Printf("   r2r design create %s\n", module)
		return 1
	}

	// Start Structurizr Lite
	fmt.Printf("🚀 Starting Structurizr Lite for module: %s\n", module)
	fmt.Printf("📁 Workspace: %s\n", workspacePath)
	fmt.Println()

	if err := design.StartStructurizrLite(module, autoStop); err != nil {
		fmt.Printf("❌ Failed to start Structurizr Lite: %v\n", err)
		return 1
	}

	return 0
}

// formatModuleList returns a formatted list of available modules
func formatModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, mod.Source.Root))
	}
	return sb.String()
}

func printServeUsage() {
	fmt.Println("Start Structurizr Lite viewer to view architecture diagrams in browser")
	fmt.Println()
	fmt.Println("Launches a Docker container running Structurizr Lite and opens your browser.")
	fmt.Println("The viewer displays interactive C4 model diagrams defined in workspace.dsl.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  r2r design serve <module>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  r2r design serve src-cli")
	fmt.Println("  r2r design serve src-commands")
	fmt.Println()
	fmt.Println("Module Locations:")
	fmt.Println("  src-cli        → specs/src-cli/.design/workspace.dsl")
	fmt.Println("  src-commands   → specs/src-commands/.design/workspace.dsl")
	fmt.Println()
	fmt.Println("What It Does:")
	fmt.Println("  1. Reads specs/<module>/.design/workspace.dsl")
	fmt.Println("  2. Generates workspace.json and .structurizr/ (ignored by git)")
	fmt.Println("  3. Allocates an available port in the 9000-9999 range")
	fmt.Println("  4. Starts Docker container with Structurizr Lite")
	fmt.Println("  5. Opens browser at the allocated port")
	fmt.Println()
	fmt.Println("Multi-Instance Support:")
	fmt.Println("  Each module gets its own container with a unique port.")
	fmt.Println("  You can run multiple viewers simultaneously for different modules.")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - Docker must be running")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  Accepts module name with or without 'specs/' prefix and '/design' suffix.")
	fmt.Println("  Generated files are automatically ignored by git.")
}
