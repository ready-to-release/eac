package show

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get"
	"github.com/ready-to-release/eac/go/commands/base"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/reports"
)

type showBuildTimesCommand struct{}

var _ core.SimpleCommandPort = (*showBuildTimesCommand)(nil)

func (c *showBuildTimesCommand) Name() string { return "show build-times" }

func (c *showBuildTimesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-build-times",
		Short:         "Display build timing metrics in tables",
		Long:          "The show build-times command displays build timing analysis parsed from build logs.\nShows overall statistics, breakdown by module type, and the slowest individual builds.\n\nExpected Output:\n- Table with overall metrics: total builds, passed/failed counts, total/average duration\n- Summary by type table: module types with build counts, pass/fail, total and average times\n- Top N slowest builds table showing duration, status, module name, and type",
	}
}

func (c *showBuildTimesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowBuildTimes()
}

func ShowBuildTimes() int {
	return ShowBuildTimesForModules(nil, 20, "")
}

// ShowBuildTimesForModules displays build timings with optional module filtering
// If modules is nil, shows all timings. topN controls how many slowest builds to show.
// If buildOutputDir is empty, defaults to out/build.
func ShowBuildTimesForModules(modules []string, topN int, buildOutputDir string) int {
	// Get repository root
	var repoRoot string
	if buildOutputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get current directory: %v\n", err)
			return 1
		}

		repoRoot, err = base.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
			return 1
		}

		// Load config to get build output directory
		cfg, err := config.Load(config.LoadOptions{RepoRoot: repoRoot})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
			return 1
		}

		buildOutputDir = filepath.Join(repoRoot, cfg.Repository.Paths.Out.Build)
	} else {
		// Derive repo root from buildOutputDir
		var err error
		repoRoot, err = base.FindRepoRoot(buildOutputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
			return 1
		}
	}

	if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: build output directory not found: %s (run build first)\n", buildOutputDir)
		return 1
	}

	// Parse build logs using the get command logic
	timings, err := get.ParseBuildLog(buildOutputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse build logs: %v\n", err)
		return 1
	}

	// Populate module types from contracts
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	for i := range timings {
		if module, exists := moduleReport.Registry.Get(timings[i].Module); exists {
			timings[i].Type = module.GetComponentTypesDisplay()
		} else {
			timings[i].Type = "unknown"
		}
	}

	// Filter by modules if specified
	if len(modules) > 0 {
		timings = filterBuildTimingsByModules(timings, modules)
	}

	if len(timings) == 0 {
		if len(modules) > 0 {
			fmt.Fprintf(os.Stderr, "Error: no build timing data found for modules: %v\n", modules)
		} else {
			fmt.Fprintf(os.Stderr, "Error: no build timing data found in %s (run build first)\n", buildOutputDir)
		}
		return 1
	}

	// Build summary
	summary := get.BuildBuildSummary(timings, buildOutputDir)

	// Display overall summary
	displayBuildOverallSummary(summary)

	// Display type summary table
	displayBuildTypeSummary(summary)

	// Display slowest builds
	displaySlowestBuilds(summary, topN)

	return 0
}

// filterBuildTimingsByModules filters timings to only include specified modules.
func filterBuildTimingsByModules(timings []get.BuildTiming, modules []string) []get.BuildTiming {
	// Create a set of modules for O(1) lookup
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}

	filtered := []get.BuildTiming{}
	for _, timing := range timings {
		if moduleSet[timing.Module] {
			filtered = append(filtered, timing)
		}
	}

	return filtered
}

// displayBuildOverallSummary shows high-level statistics.
func displayBuildOverallSummary(summary *get.BuildTimingSummary) {
	fmt.Println("# Build Timing Analysis")
	fmt.Println("")
	fmt.Printf("**Build Output Directory**: `%s`\n", summary.BuildOutputDir)
	fmt.Println("")

	// Build summary table
	tb := render.NewTableBuilder().
		WithHeaders("Metric", "Value")

	tb.AddRow("Total Builds", fmt.Sprintf("%d", summary.TotalBuilds))
	tb.AddRow("Passed Builds", fmt.Sprintf("%d", summary.PassedBuilds))
	tb.AddRow("Failed Builds", fmt.Sprintf("%d", summary.FailedBuilds))
	tb.AddRow("Total Duration", fmt.Sprintf("%.2fs", summary.TotalDuration))
	tb.AddRow("Average Duration", fmt.Sprintf("%.2fs", summary.AvgDuration))

	fmt.Println(tb.Build())
	fmt.Println("")
}

// displayBuildTypeSummary shows timing breakdown by module type.
func displayBuildTypeSummary(summary *get.BuildTimingSummary) {
	fmt.Println("## Summary by Type")
	fmt.Println("")

	// Convert map to slice for sorting
	type typeStat struct {
		Type          string
		TotalBuilds   int
		PassedBuilds  int
		FailedBuilds  int
		TotalDuration float64
		AvgDuration   float64
	}

	var stats []typeStat
	for _, ts := range summary.ByType {
		stats = append(stats, typeStat{
			Type:          ts.Type,
			TotalBuilds:   ts.TotalBuilds,
			PassedBuilds:  ts.PassedBuilds,
			FailedBuilds:  ts.FailedBuilds,
			TotalDuration: ts.TotalDuration,
			AvgDuration:   ts.AvgDuration,
		})
	}

	// Sort by total duration (slowest first)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TotalDuration > stats[j].TotalDuration
	})

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Type", "Builds", "Passed", "Failed", "Total (s)", "Avg (s)")

	for _, stat := range stats {
		tb.AddRow(
			stat.Type,
			fmt.Sprintf("%d", stat.TotalBuilds),
			fmt.Sprintf("%d", stat.PassedBuilds),
			fmt.Sprintf("%d", stat.FailedBuilds),
			fmt.Sprintf("%.2f", stat.TotalDuration),
			fmt.Sprintf("%.2f", stat.AvgDuration),
		)
	}

	fmt.Println(tb.Build())
	fmt.Println("")
}

// displaySlowestBuilds shows the top N slowest individual module builds.
func displaySlowestBuilds(summary *get.BuildTimingSummary, topN int) {
	fmt.Printf("## Top %d Slowest Builds\n", topN)
	fmt.Println("")

	// Sort timings by duration (slowest first)
	sortedTimings := make([]get.BuildTiming, len(summary.Timings))
	copy(sortedTimings, summary.Timings)

	sort.Slice(sortedTimings, func(i, j int) bool {
		return sortedTimings[i].Duration > sortedTimings[j].Duration
	})

	// Limit to topN
	if len(sortedTimings) > topN {
		sortedTimings = sortedTimings[:topN]
	}

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("#", "Duration (s)", "Status", "Module", "Type")

	for i, timing := range sortedTimings {
		status := "✅ "
		if timing.Status == "FAIL" {
			status = "❌ "
		}

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%.2f", timing.Duration),
			status,
			timing.Module,
			timing.Type,
		)
	}

	fmt.Println(tb.Build())
	fmt.Println("")
}
