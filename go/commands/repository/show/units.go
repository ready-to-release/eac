package show

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showUnitsCommand struct{}

var _ core.SimpleCommandPort = (*showUnitsCommand)(nil)

func (c *showUnitsCommand) Name() string { return "show units" }

func (c *showUnitsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-units",
		Short:         "Display units of work in a human-readable markdown report",
		Long:          "Shows units for a framework (build|test|lint|scan) with cache status visualization.\nThe report includes:\n  - Summary statistics (total, cached, stale, new, container, host)\n  - Cache status overview with percentages\n  - Units grouped by module with status icons\n  - Stale units section for quick identification\n\nFilter Examples:\n  show units build                        # All build units\n  show units test --module eac   # Test units for specific module\n  show units lint --stale                 # Only stale lint units\n  show units scan --container             # Only container-based scan units",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", DefaultValue: "", Usage: "Filter to specific module"},
			{Name: "component", Type: "string", DefaultValue: "", Usage: "Filter to specific component"},
			{Name: "cached", Type: "bool", DefaultValue: "false", Usage: "Only show cached (up-to-date) units"},
			{Name: "stale", Type: "bool", DefaultValue: "false", Usage: "Only show stale units"},
			{Name: "container", Type: "bool", DefaultValue: "false", Usage: "Only show container-based units"},
			{Name: "host", Type: "bool", DefaultValue: "false", Usage: "Only show host-installed units"},
		},
	}
}

func (c *showUnitsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowUnits()
}

// unitShowFilters holds the parsed filter flags for show units.
type unitShowFilters struct {
	Module    string
	Component string
	Cached    bool
	Stale     bool
	Container bool
	Host      bool
}

// parseUnitShowFilters extracts filter flags from command arguments.
func parseUnitShowFilters(args []string) unitShowFilters {
	filters := unitShowFilters{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--module":
			if i+1 < len(args) {
				filters.Module = args[i+1]
				i++
			}
		case "--component":
			if i+1 < len(args) {
				filters.Component = args[i+1]
				i++
			}
		case "--cached":
			filters.Cached = true
		case "--stale":
			filters.Stale = true
		case "--container":
			filters.Container = true
		case "--host":
			filters.Host = true
		}
	}
	return filters
}

// toReportFilters converts unitShowFilters to reports.UnitFilters.
func (f unitShowFilters) toReportFilters() reports.UnitFilters {
	return reports.UnitFilters{
		Module:    f.Module,
		Component: f.Component,
		Cached:    f.Cached,
		Stale:     f.Stale,
		Container: f.Container,
		Host:      f.Host,
	}
}

// ShowUnits displays units of work in markdown format.
func ShowUnits() int {
	// Parse framework argument (should be after "show units")
	args := os.Args[3:] // Skip program name, "show", "units"
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: show units <build|test|lint|scan> [flags]\n")
		return 1
	}

	frameworkStr := args[0]
	if frameworkStr[0] == '-' {
		fmt.Fprintf(os.Stderr, "Usage: show units <build|test|lint|scan> [flags]\n")
		return 1
	}

	var framework reports.Framework
	switch frameworkStr {
	case "build":
		framework = reports.FrameworkBuild
	case "test":
		framework = reports.FrameworkTest
	case "lint":
		framework = reports.FrameworkLint
	case "scan":
		framework = reports.FrameworkScan
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid framework '%s'. Use: build, test, lint, or scan\n", frameworkStr)
		return 1
	}

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

	// Parse filter flags (from remaining args)
	filters := parseUnitShowFilters(args[1:])

	// Get unit report
	report, err := reports.GetUnits(workspaceRoot, framework)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Apply filters
	filtered := reports.FilterUnits(report.Units, filters.toReportFilters())

	if len(filtered) == 0 {
		fmt.Println("No units found matching the specified filters.")
		return 0
	}

	// Print report
	fmt.Printf("# %s Units Report\n", capitalize(frameworkStr))
	fmt.Println("")

	// Summary section
	printUnitSummary(filtered)

	// Cache status overview
	printCacheStatusOverview(filtered)

	// Units by module
	printUnitsByModule(filtered)

	// Stale units section
	printStaleUnits(filtered)

	// Dependency chain
	printUnitDependencies(filtered)

	return 0
}

// printUnitSummary prints the summary section.
func printUnitSummary(units []*reports.UnitInfo) {
	fmt.Println("## Summary")
	fmt.Println("")

	// Count by status
	var cached, stale, new_, container, host int
	for _, unit := range units {
		if unit.CacheStatus != nil {
			switch unit.CacheStatus.State {
			case "up_to_date":
				cached++
			case "stale":
				stale++
			case "new":
				new_++
			}
		}
		if unit.Container {
			container++
		}
		if unit.HostInstalled {
			host++
		}
	}

	tb := render.NewTableBuilder().
		WithHeaders("Metric", "Value").
		AddRow("Total Units", len(units)).
		AddRow("Cached (up-to-date)", cached).
		AddRow("Stale (needs rebuild)", stale).
		AddRow("New (never run)", new_).
		AddRow("Container-based", container).
		AddRow("Host-installed", host).
		Build()

	fmt.Println(tb)
	fmt.Println("")
}

// printCacheStatusOverview prints the cache status breakdown.
func printCacheStatusOverview(units []*reports.UnitInfo) {
	fmt.Println("## Cache Status Overview")
	fmt.Println("")

	var cached, stale, new_ int
	for _, unit := range units {
		if unit.CacheStatus != nil {
			switch unit.CacheStatus.State {
			case "up_to_date":
				cached++
			case "stale":
				stale++
			case "new":
				new_++
			}
		}
	}

	total := len(units)
	tb := render.NewTableBuilder().
		WithHeaders("Status", "Count", "Percentage").
		AddRow(Emoji("success")+" Cached", cached, formatPercent(cached, total)).
		AddRow(Emoji("failure")+" Stale", stale, formatPercent(stale, total)).
		AddRow(Emoji("info")+" New", new_, formatPercent(new_, total)).
		Build()

	fmt.Println(tb)
	fmt.Println("")
}

// printUnitsByModule prints units grouped by module.
func printUnitsByModule(units []*reports.UnitInfo) {
	fmt.Println("## Units by Module")
	fmt.Println("")

	// Group by module
	byModule := make(map[string][]*reports.UnitInfo)
	for _, unit := range units {
		byModule[unit.Module] = append(byModule[unit.Module], unit)
	}

	// Sort modules
	var modules []string
	for m := range byModule {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	for _, module := range modules {
		moduleUnits := byModule[module]
		fmt.Printf("### %s (%d units)\n\n", module, len(moduleUnits))

		tb := render.NewTableBuilder().
			WithHeaders("Status", "Component", "Tool", "Weight", "Type", "Last Run", "Reason")

		for _, unit := range moduleUnits {
			status := cacheStatusIcon(unit.CacheStatus)
			runType := "host"
			if unit.Container {
				runType = "container"
			}
			lastRun := formatLastRun(unit.CacheStatus)
			reason := formatReason(unit.CacheStatus)

			tb.AddRow(status, unit.Component, unit.Tool, unit.Weight, runType, lastRun, reason)
		}

		fmt.Println(tb.Build())
		fmt.Println("")
	}
}

// printStaleUnits prints a dedicated section for stale units.
func printStaleUnits(units []*reports.UnitInfo) {
	// Get stale units
	staleUnits := reports.GetStaleUnits(units)
	if len(staleUnits) == 0 {
		return
	}

	fmt.Println("## Stale Units (need execution)")
	fmt.Println("")
	fmt.Println("These units will be rebuilt on next run:")
	fmt.Println("")

	tb := render.NewTableBuilder().
		WithHeaders("Module", "Component", "Reason")

	for _, unit := range staleUnits {
		reason := "-"
		if unit.CacheStatus != nil && unit.CacheStatus.Reason != "" {
			reason = unit.CacheStatus.Reason
		}
		tb.AddRow(unit.Module, unit.Component, reason)
	}

	fmt.Println(tb.Build())
	fmt.Println("")
}

// printUnitDependencies prints intra-module dependencies if any exist.
func printUnitDependencies(units []*reports.UnitInfo) {
	// Check if any dependencies exist
	hasDeps := false
	for _, unit := range units {
		if len(unit.Dependencies) > 0 {
			hasDeps = true
			break
		}
	}

	if !hasDeps {
		return
	}

	fmt.Println("## Dependency Chain")
	fmt.Println("")
	fmt.Println("### Intra-module Dependencies")
	fmt.Println("")

	tb := render.NewTableBuilder().
		WithHeaders("Unit", "Depends On")

	for _, unit := range units {
		if len(unit.Dependencies) == 0 {
			continue
		}
		key := fmt.Sprintf("%s:%s", unit.Module, unit.Component)
		deps := ""
		for i, d := range unit.Dependencies {
			if i > 0 {
				deps += ", "
			}
			deps += d
		}
		tb.AddRow(key, deps)
	}

	fmt.Println(tb.Build())
	fmt.Println("")
}

// cacheStatusIcon returns the appropriate icon for cache status.
func cacheStatusIcon(status *reports.CacheStatus) string {
	if status == nil {
		return Emoji("info")
	}
	switch status.State {
	case "up_to_date":
		return Emoji("success")
	case "stale":
		return Emoji("failure")
	case "new":
		return Emoji("info")
	default:
		return "-"
	}
}

// formatLastRun formats the last execution time as a human-readable relative time.
func formatLastRun(status *reports.CacheStatus) string {
	if status == nil || status.ExecutedAt.IsZero() {
		return "never"
	}

	duration := time.Since(status.ExecutedAt)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	default:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

// formatReason returns the reason for stale status or "-".
func formatReason(status *reports.CacheStatus) string {
	if status == nil || status.Reason == "" || status.State == "up_to_date" {
		return "-"
	}
	return status.Reason
}

// formatPercent formats a count as a percentage.
func formatPercent(count, total int) string {
	if total == 0 {
		return "0%"
	}
	pct := (count * 100) / total
	return fmt.Sprintf("%d%%", pct)
}

// capitalize returns the string with the first letter capitalized.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-'a'+'A') + s[1:]
}
