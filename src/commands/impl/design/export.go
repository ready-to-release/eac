// Command: design export
// Description: Export workspace DSL content
// Usage: design export <module> [--output <file>]
package design

import (
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/impl/design/internal/workspace"
)

func init() {
	registry.Register("design export", DesignExport)
}

// DesignExport exports the workspace DSL content
func DesignExport() int {
	args := os.Args[3:] // Skip "go", "run", ".", "design", "export"

	if len(args) == 0 {
		printDesignExportUsage()
		return 1
	}

	module := args[0]
	var outputFile string

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "Error: --output requires a value\n")
				return 1
			}
		case "--help", "-h":
			printDesignExportUsage()
			return 0
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n\n", args[i])
			printDesignExportUsage()
			return 1
		}
	}

	// Export workspace
	content, size, err := workspace.Export(module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Write to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write to file: %v\n", err)
			return 1
		}
		fmt.Printf("Exported workspace to %s (%d bytes)\n", outputFile, size)
	} else {
		fmt.Print(content)
	}

	return 0
}

func printDesignExportUsage() {
	fmt.Println("Export workspace DSL content")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . design export <module> [--output <file>]")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("  <module>              Module moniker (e.g., src-cli, src-commands)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --output, -o          Output file path (default: stdout)")
	fmt.Println("  --help, -h            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run . design export src-cli")
	fmt.Println("  go run . design export src-cli --output architecture.dsl")
	fmt.Println("  go run . design export src-cli | grep \"container\"")
	fmt.Println()
	fmt.Println("Input:")
	fmt.Println("  Reads from: specs/<module>/design/workspace.dsl")
}
