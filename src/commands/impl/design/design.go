// Command: design
// Description: Create, validate, and view architecture diagrams using Structurizr DSL
// HasSideEffects: false
package design

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Design)
}

// Design command entry point
func Design() int {
	args := os.Args[2:] // Skip "go" and "run" and "."

	if len(args) == 0 {
		printDesignUsage()
		return 1
	}

	// Check for subcommands
	switch args[0] {
	case "create":
		// Handled by separate registration in create/create.go
		return 0
	case "validate":
		// Handled by separate registration in validate.go
		return 0
	case "serve":
		// Handled by separate registration in serve.go
		return 0
	case "--help", "-h":
		printDesignUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printDesignUsage()
		return 1
	}
}

func printDesignUsage() {
	fmt.Println("Architecture documentation using Structurizr DSL and C4 model diagrams")
	fmt.Println()
	fmt.Println("Structurizr Lite is a web-based viewer for architecture diagrams defined in DSL.")
	fmt.Println("DSL files define System Context, Container, and Component views.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  r2r design <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create <module>    Generate workspace.dsl for a module")
	fmt.Println("  validate <module>  Validate workspace.dsl syntax")
	fmt.Println("  validate --all     Validate all workspace files")
	fmt.Println("  serve <module>     View diagrams in browser (http://localhost:8080)")
	fmt.Println()
	fmt.Println("Module Examples:")
	fmt.Println("  src-cli              → specs/src-cli/design/workspace.dsl")
	fmt.Println("  contracts            → specs/contracts/design/workspace.dsl")
	fmt.Println("  docs                 → specs/docs/design/workspace.dsl")
	fmt.Println("  specs/src-cli/design → specs/src-cli/design/workspace.dsl (auto-cleaned)")
	fmt.Println()
	fmt.Println("File Locations:")
	fmt.Println("  Source:    specs/<module>/design/workspace.dsl   (tracked in git)")
	fmt.Println("  Generated: specs/<module>/design/workspace.json  (ignored by git)")
	fmt.Println("  Generated: specs/<module>/design/.structurizr/   (ignored by git)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate architecture for a module")
	fmt.Println("  r2r design create src-cli")
	fmt.Println()
	fmt.Println("  # Generate with debug output")
	fmt.Println("  r2r design create contracts --debug")
	fmt.Println()
	fmt.Println("  # Validate workspace syntax")
	fmt.Println("  r2r design validate src-cli")
	fmt.Println()
	fmt.Println("  # View diagrams in browser")
	fmt.Println("  r2r design serve src-cli")
	fmt.Println()
	fmt.Println("Flags (create):")
	fmt.Println("  --debug, -d    Save AI prompts and responses to out/ for debugging")
	fmt.Println("  --force, -f    Overwrite existing workspace.dsl file")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - Docker (for validate and serve commands)")
	fmt.Println("  - Anthropic API key (for create command)")
	fmt.Println()
	fmt.Println("Help:")
	fmt.Println("  r2r design <subcommand> --help")
}
