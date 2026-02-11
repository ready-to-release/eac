package get

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getComponentsCommand struct{}

var _ core.SimpleCommandPort = (*getComponentsCommand)(nil)

func (c *getComponentsCommand) Name() string { return "get components" }

func (c *getComponentsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-components",
		Short:         "List all components with dependencies and phase information",
		Long:          "Returns structured data about all components in the repository.\nEach component includes its type, root path, phase support (build, lint, test, scan),\nand bidirectional dependencies (depends_on, depended_by).\n\nFilter Examples:\n  get components --module eac    # Components in specific module\n  get components --type go                # Only Go components\n  get components --buildable              # Only buildable components\n  get components --lintable --scannable   # Lintable AND scannable",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", DefaultValue: "", Usage: "Filter to specific module"},
			{Name: "type", Type: "string", DefaultValue: "", Usage: "Filter by component type (go, typescript, book, etc.)"},
			{Name: "buildable", Type: "bool", DefaultValue: "false", Usage: "Only components with build phase"},
			{Name: "lintable", Type: "bool", DefaultValue: "false", Usage: "Only components with lint phase"},
			{Name: "testable", Type: "bool", DefaultValue: "false", Usage: "Only components with test phase"},
			{Name: "scannable", Type: "bool", DefaultValue: "false", Usage: "Only components with scan phase"},
		},
	}
}

func (c *getComponentsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetComponents()
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
