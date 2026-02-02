// Command: get components
// Short: List all components with dependencies and phase information
// Long: Returns structured data about all components in the repository.
// Long: Each component includes its type, root path, phase support (build, lint, test, scan),
// Long: and bidirectional dependencies (depends_on, depended_by).
// Long:
// Long: Filter Examples:
// Long:   get components --module eac-cli    # Components in specific module
// Long:   get components --type go                # Only Go components
// Long:   get components --buildable              # Only buildable components
// Long:   get components --lintable --scannable   # Lintable AND scannable
// Flag.module: type=string, default="", usage=Filter to specific module
// Flag.type: type=string, default="", usage=Filter by component type (go, typescript, book, etc.)
// Flag.buildable: type=bool, default=false, usage=Only components with build phase
// Flag.lintable: type=bool, default=false, usage=Only components with lint phase
// Flag.testable: type=bool, default=false, usage=Only components with test phase
// Flag.scannable: type=bool, default=false, usage=Only components with scan phase
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(GetComponents)
}

// componentFilters holds the parsed filter flags.
type componentFilters struct {
	Module    string
	Type      string
	Buildable bool
	Lintable  bool
	Testable  bool
	Scannable bool
}

// parseComponentFilters extracts filter flags from command arguments.
func parseComponentFilters(args []string) componentFilters {
	filters := componentFilters{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--module":
			if i+1 < len(args) {
				filters.Module = args[i+1]
				i++
			}
		case "--type":
			if i+1 < len(args) {
				filters.Type = args[i+1]
				i++
			}
		case "--buildable":
			filters.Buildable = true
		case "--lintable":
			filters.Lintable = true
		case "--testable":
			filters.Testable = true
		case "--scannable":
			filters.Scannable = true
		}
	}
	return filters
}

// toReportFilters converts componentFilters to reports.ComponentFilters.
func (f componentFilters) toReportFilters() reports.ComponentFilters {
	return reports.ComponentFilters{
		Module:    f.Module,
		Type:      f.Type,
		Buildable: f.Buildable,
		Lintable:  f.Lintable,
		Testable:  f.Testable,
		Scannable: f.Scannable,
	}
}

// GetComponents returns all components with their phase and dependency information.
func GetComponents() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse filter flags
	filters := parseComponentFilters(os.Args[1:])

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		report, err := reports.GetComponents(workspaceRoot)
		if err != nil {
			return nil, err
		}

		// Apply filters
		filtered := reports.FilterComponents(report.Components, filters.toReportFilters())
		return filtered, nil
	})
}
