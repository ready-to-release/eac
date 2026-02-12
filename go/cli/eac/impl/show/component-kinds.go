package show

import (
	"context"
	"fmt"
	"os"
	"sort"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showComponentTypesCommand struct{}

var _ core.SimpleCommandPort = (*showComponentTypesCommand)(nil)

func (c *showComponentTypesCommand) Name() string { return "show component-kinds" }

func (c *showComponentTypesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-component-kinds",
		Short:         "Display all component kinds grouped by count",
		Long:          "The show component-kinds command displays all component kinds used in the repository,\ngrouped by count to show which kinds are most commonly used.\nThis helps understand the technology mix and module distribution in the repository.\nThe output is formatted as a Markdown table with a footer showing total kinds.",
	}
}

func (c *showComponentTypesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowComponentTypes()
}

func ShowComponentTypes() int {
	return ExecuteShowCommand(showComponentTypesImpl)
}

func showComponentTypesImpl() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Generate module contracts report
	report, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Count by component type across all modules
	// Use GetComponentType to resolve actual types, not just names
	typeCounts := make(map[string]int)
	for _, mod := range report.Modules {
		for _, compName := range mod.GetEnabledComponents() {
			// Resolve actual type (may differ from name for named components like book types)
			compType := mod.Components.GetComponentType(compName)
			typeCounts[compType]++
		}
	}

	// Sort types alphabetically
	types := make([]string, 0, len(typeCounts))
	for t := range typeCounts {
		types = append(types, t)
	}
	sort.Strings(types)

	// Build markdown table
	tb := render.NewTableBuilder()
	tb.WithHeaders("Component Type", "Count")

	for _, t := range types {
		tb.AddRow(t, fmt.Sprintf("%d", typeCounts[t]))
	}

	// Add footer with total
	tb.WithFooter("Total Types", fmt.Sprintf("%d", len(types)))

	fmt.Println(tb.Build())
	return 0
}
