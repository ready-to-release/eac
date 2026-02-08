// Command: serve design
// Short: View architecture diagrams in browser using Structurizr Lite (Docker)
// Long: Launches a Docker container running Structurizr Lite web viewer and opens your default browser.
// Long: The viewer provides interactive C4 model diagrams (system context, containers, components) defined
// Long: in workspace.dsl. When you run this command, it generates workspace.json and .structurizr/ files
// Long: The viewer runs on a dynamically allocated port (9000-9999) and updates automatically when you edit the DSL file.
// Long:
// Long: Expected Output:
// Long:   - Docker container running Structurizr Lite on port 9000-9999
// Long:   - Browser opens automatically to http://localhost:<port>
// Long:   - Live updates when workspace.dsl file changes
// Long:   - Generated workspace.json and .structurizr/ directory (git-ignored)
// Usage: serve design <module>
package design

import (
	"fmt"
	"os"
	"strings"

	designInternal "github.com/ready-to-release/eac/go/cli/eac/impl/design/helper"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

// commandFlags defines valid flags for the serve design command

var log = logging.C()

func init() {
	registry.Register(ServeDesign)
}

// ServeDesign starts Structurizr Lite viewer for a module.
func ServeDesign() int {
	args := os.Args[3:] // Skip program, "serve", and "design"

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	if len(args) == 0 {
		log.Info("❌ Error: module name required")
		log.Info("")
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
		if arg == "--force" {
			autoStop = true
			continue
		}
		// First non-flag argument is the module name
		if !strings.HasPrefix(arg, "-") && module == "" {
			module = arg
		}
	}

	if module == "" {
		log.Info("❌ Error: module name required")
		log.Info("")
		printServeUsage()
		return 1
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("❌ Failed to find repository root: %v", err)
		return 1
	}

	// Validate module name for security
	if err := designInternal.ValidateModuleName(module); err != nil {
		log.Errorf("❌ Invalid module name: %v", err)
		return 1
	}

	// Load module contracts and validate moniker exists (same as build command)
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		log.Errorf("❌ Failed to load module contracts: %v", err)
		return 1
	}

	mod, exists := moduleReport.Registry.Get(module)
	if !exists {
		log.Errorf("❌ Module not found: %s\n\nAvailable modules:\n%s",
			module, formatModuleList(moduleReport))
		return 1
	}

	// Use validated moniker
	module = mod.Moniker

	// Check if workspace exists
	workspacePath := paths.WorkspaceDSLPath(repoRoot, module)
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		log.Infof("❌ Workspace not found: %s", workspacePath)
		log.Info("")
		log.Info("💡 Create one first with:")
		log.Infof("   clie design create %s", module)
		return 1
	}

	// Start Structurizr Lite
	log.Infof("🚀 Starting Structurizr Lite for module: %s", module)
	log.Infof("📁 Workspace: %s", workspacePath)
	log.Info("")

	if err := designInternal.StartStructurizrLite(module, autoStop); err != nil {
		log.Errorf("❌ Failed to start Structurizr Lite: %v", err)
		return 1
	}

	return 0
}

// formatModuleList returns a formatted list of available modules.
func formatModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		// Get first package root for display
		var displayRoot string
		for _, root := range mod.GetComponentRoots() {
			displayRoot = root
			break
		}
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, displayRoot))
	}
	return sb.String()
}

func printServeUsage() {
	log.Info("Start Structurizr Lite viewer to view architecture diagrams in browser")
	log.Info("")
	log.Info("Launches a Docker container running Structurizr Lite and opens your browser.")
	log.Info("The viewer displays interactive C4 model diagrams defined in workspace.dsl.")
	log.Info("")
	log.Info("Usage:")
	log.Info("  clie design serve <module>")
	log.Info("")
	log.Info("Examples:")
	log.Info("  clie design serve clie-cli")
	log.Info("  clie design serve eac-cli")
	log.Info("")
	log.Info("Module Locations:")
	log.Info("  clie-cli        → specs/clie-cli/.design/workspace.dsl")
	log.Info("  eac-cli   → specs/eac-cli/.design/workspace.dsl")
	log.Info("")
	log.Info("What It Does:")
	log.Info("  1. Reads specs/<module>/.design/workspace.dsl")
	log.Info("  2. Generates workspace.json and .structurizr/ (ignored by git)")
	log.Info("  3. Allocates an available port in the 9000-9999 range")
	log.Info("  4. Starts Docker container with Structurizr Lite")
	log.Info("  5. Opens browser at the allocated port")
	log.Info("")
	log.Info("Multi-Instance Support:")
	log.Info("  Each module gets its own container with a unique port.")
	log.Info("  You can run multiple viewers simultaneously for different modules.")
	log.Info("")
	log.Info("Requirements:")
	log.Info("  - Docker must be running")
	log.Info("")
	log.Info("Note:")
	log.Info("  Accepts module name with or without 'specs/' prefix and '/design' suffix.")
	log.Info("  Generated files are automatically ignored by git.")
}
