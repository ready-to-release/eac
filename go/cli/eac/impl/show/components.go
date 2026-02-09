// Command: show components
// Short: Display all components in a human-readable markdown report
// Long: Shows components grouped by module with phase support and dependency information.
// Long: The report includes:
// Long:   - Summary statistics (total, buildable, lintable, testable, scannable)
// Long:   - Components grouped by module with phase icons
// Long:   - Dependency graph showing what each component depends on and what depends on it
// Long:
// Long: Filter Examples:
// Long:   show components                         # All components
// Long:   show components --module eac   # Components in specific module
// Long:   show components --type go               # Only Go components
// Long:   show components --buildable             # Only buildable components
// Flag.module: type=string, default="", usage=Filter to specific module
// Flag.type: type=string, default="", usage=Filter by component type (go, typescript, book, etc.)
// Flag.buildable: type=bool, default=false, usage=Only components with build phase
// Flag.lintable: type=bool, default=false, usage=Only components with lint phase
// Flag.testable: type=bool, default=false, usage=Only components with test phase
// Flag.scannable: type=bool, default=false, usage=Only components with scan phase
package show

import (
	"fmt"
	"os"
	"sort"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

// componentShowFilters holds the parsed filter flags.
type componentShowFilters struct {
	Module    string
	Type      string
	Buildable bool
	Lintable  bool
	Testable  bool
	Scannable bool
}

// parseComponentShowFilters extracts filter flags from command arguments.
func parseComponentShowFilters(args []string) componentShowFilters {
	filters := componentShowFilters{}
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

// toReportFilters converts componentShowFilters to reports.ComponentFilters.
func (f componentShowFilters) toReportFilters() reports.ComponentFilters {
	return reports.ComponentFilters{
		Module:    f.Module,
		Type:      f.Type,
		Buildable: f.Buildable,
		Lintable:  f.Lintable,
		Testable:  f.Testable,
		Scannable: f.Scannable,
	}
}

// ShowComponents displays all components in markdown format.
func ShowComponents() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse filter flags
	filters := parseComponentShowFilters(os.Args[1:])

	// Get component report
	report, err := reports.GetComponents(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Apply filters
	filtered := reports.FilterComponents(report.Components, filters.toReportFilters())

	if len(filtered) == 0 {
		fmt.Println("No components found matching the specified filters.")
		return 0
	}

	// Load config for DisplayOrder
	cfg, cfgErr := config.Load(config.DefaultLoadOptions())
	if cfgErr != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", cfgErr)
		return 1
	}

	// Print report
	fmt.Println("# Component Report")
	fmt.Println("")

	// Summary section
	printComponentSummary(filtered)

	// Components by type
	printComponentsByType(filtered)

	// Components by module (using DisplayOrder)
	printComponentsByModule(filtered, cfg.Repository.DisplayOrder)

	// Dependency graph
	printDependencyGraph(filtered)

	return 0
}

// printComponentSummary prints the summary section.
func printComponentSummary(components []*reports.ComponentInfo) {
	fmt.Println("## Summary")
	fmt.Println("")

	// Count phases
	var buildable, lintable, testable, scannable int
	modules := make(map[string]bool)
	types := make(map[string]bool)

	for _, comp := range components {
		modules[comp.Moniker] = true
		types[comp.Type] = true
		if comp.Phases != nil {
			if comp.Phases.Build != nil && comp.Phases.Build.Enabled {
				buildable++
			}
			if comp.Phases.Lint != nil && comp.Phases.Lint.Enabled {
				lintable++
			}
			if comp.Phases.Test != nil && comp.Phases.Test.Enabled {
				testable++
			}
			if comp.Phases.Scan != nil && comp.Phases.Scan.Enabled {
				scannable++
			}
		}
	}

	summaryTable := render.NewTableBuilder().
		WithHeaders("Metric", "Value").
		AddRow("Total Components", len(components)).
		AddRow("Modules", len(modules)).
		AddRow("Component Types", len(types)).
		AddRow("Buildable", buildable).
		AddRow("Lintable", lintable).
		AddRow("Testable", testable).
		AddRow("Scannable", scannable).
		Build()

	fmt.Println(summaryTable)
	fmt.Println("")
}

// printComponentsByType prints components grouped by type.
func printComponentsByType(components []*reports.ComponentInfo) {
	fmt.Println("## Components by Type")
	fmt.Println("")

	// Group by type
	byType := make(map[string][]*reports.ComponentInfo)
	for _, comp := range components {
		byType[comp.Type] = append(byType[comp.Type], comp)
	}

	// Sort types
	var types []string
	for t := range byType {
		types = append(types, t)
	}
	sort.Strings(types)

	tb := render.NewTableBuilder().
		WithHeaders("Type", "Count", "Modules")

	for _, t := range types {
		comps := byType[t]
		modules := make(map[string]bool)
		for _, c := range comps {
			modules[c.Moniker] = true
		}
		var moduleList []string
		for m := range modules {
			moduleList = append(moduleList, m)
		}
		sort.Strings(moduleList)

		// Truncate module list if too long
		display := ""
		if len(moduleList) <= 3 {
			for i, m := range moduleList {
				if i > 0 {
					display += ", "
				}
				display += m
			}
		} else {
			display = fmt.Sprintf("%s, %s, ... (%d more)",
				moduleList[0], moduleList[1], len(moduleList)-2)
		}

		tb.AddRow(t, len(comps), display)
	}

	fmt.Println(tb.Build())
	fmt.Println("")
}

// printComponentsByModule prints components grouped by module in display order.
func printComponentsByModule(components []*reports.ComponentInfo, displayOrder *config.DisplayOrder) {
	fmt.Println("## Components by Module")
	fmt.Println("")

	// Group by module
	byModule := make(map[string][]*reports.ComponentInfo)
	for _, comp := range components {
		byModule[comp.Moniker] = append(byModule[comp.Moniker], comp)
	}

	// Iterate in display order
	for _, module := range displayOrder.Modules {
		comps, ok := byModule[module]
		if !ok || len(comps) == 0 {
			continue
		}

		fmt.Printf("### %s\n\n", module)

		tb := render.NewTableBuilder().
			WithHeaders("Component", "Type", "Root", "Build", "Lint", "Test", "Scan")

		for _, comp := range comps {
			tb.AddRow(
				comp.Component,
				comp.Type,
				Code(comp.Root),
				phaseIcon(comp.Phases != nil && comp.Phases.Build != nil && comp.Phases.Build.Enabled),
				phaseIcon(comp.Phases != nil && comp.Phases.Lint != nil && comp.Phases.Lint.Enabled),
				phaseIcon(comp.Phases != nil && comp.Phases.Test != nil && comp.Phases.Test.Enabled),
				phaseIcon(comp.Phases != nil && comp.Phases.Scan != nil && comp.Phases.Scan.Enabled),
			)
		}

		fmt.Println(tb.Build())
		fmt.Println("")
	}
}

// printDependencyGraph prints the dependency graph section.
func printDependencyGraph(components []*reports.ComponentInfo) {
	// Check if any dependencies exist
	hasDeps := false
	hasReverseDeps := false
	for _, comp := range components {
		if comp.Dependencies != nil {
			if len(comp.Dependencies.DependsOn) > 0 {
				hasDeps = true
			}
			if len(comp.Dependencies.DependedBy) > 0 {
				hasReverseDeps = true
			}
		}
	}

	if !hasDeps && !hasReverseDeps {
		return
	}

	fmt.Println("## Dependency Graph")
	fmt.Println("")

	if hasDeps {
		fmt.Println("### Dependencies (what each component needs)")
		fmt.Println("")

		tb := render.NewTableBuilder().
			WithHeaders("Component", "Depends On")

		for _, comp := range components {
			if comp.Dependencies == nil || len(comp.Dependencies.DependsOn) == 0 {
				continue
			}
			key := fmt.Sprintf("%s:%s", comp.Moniker, comp.Component)
			deps := render.FormatListForColumn(comp.Dependencies.DependsOn, 3)
			tb.AddRow(key, deps)
		}

		fmt.Println(tb.Build())
		fmt.Println("")
	}

	if hasReverseDeps {
		fmt.Println("### Dependents (what depends on each component)")
		fmt.Println("")

		tb := render.NewTableBuilder().
			WithHeaders("Component", "Used By")

		for _, comp := range components {
			if comp.Dependencies == nil || len(comp.Dependencies.DependedBy) == 0 {
				continue
			}
			key := fmt.Sprintf("%s:%s", comp.Moniker, comp.Component)
			deps := render.FormatListForColumn(comp.Dependencies.DependedBy, 3)
			tb.AddRow(key, deps)
		}

		fmt.Println(tb.Build())
		fmt.Println("")
	}
}

// phaseIcon returns a checkmark or dash based on phase enablement.
func phaseIcon(enabled bool) string {
	if enabled {
		return Emoji("success")
	}
	return "-"
}
