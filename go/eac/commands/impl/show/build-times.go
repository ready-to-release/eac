// Command: show build-times
// Description: Show build timing analysis in a human-readable format
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ShowBuildTimes)
}

func ShowBuildTimes() int {
	return ShowBuildTimesForModules(nil, 20, "")
}

// ShowBuildTimesForModules displays build timings with optional module filtering
// If modules is nil, shows all timings. topN controls how many slowest builds to show.
// If buildOutputDir is empty, defaults to out/build.
func ShowBuildTimesForModules(modules []string, topN int, buildOutputDir string) int {
	// Get repository root if needed
	if buildOutputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Errorf("failed to get current directory: %v", err)
			return 1
		}

		repoRoot, err := findRepoRoot(cwd)
		if err != nil {
			log.Errorf("failed to find repository root: %v", err)
			return 1
		}

		buildOutputDir = filepath.Join(repoRoot, "out", "build")
	}
	if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
		log.Errorf("build output directory not found: %s (run build first)", buildOutputDir)
		return 1
	}

	// Parse build logs using the get command logic
	timings, err := get.ParseBuildLog(buildOutputDir)
	if err != nil {
		log.Errorf("failed to parse build logs: %v", err)
		return 1
	}

	// Filter by modules if specified
	if len(modules) > 0 {
		timings = filterBuildTimingsByModules(timings, modules)
	}

	if len(timings) == 0 {
		if len(modules) > 0 {
			log.Errorf("no build timing data found for modules: %v", modules)
		} else {
			log.Errorf("no build timing data found in %s (run build first)", buildOutputDir)
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

// filterBuildTimingsByModules filters timings to only include specified modules
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

// displayBuildOverallSummary shows high-level statistics
func displayBuildOverallSummary(summary *get.BuildTimingSummary) {
	log.Info("# Build Timing Analysis\n")
	log.Infof("**Build Output Directory**: `%s`\n", summary.BuildOutputDir)
	log.Info("")

	// Build summary table
	tb := render.NewTableBuilder().
		WithHeaders("Metric", "Value")

	tb.AddRow("Total Builds", fmt.Sprintf("%d", summary.TotalBuilds))
	tb.AddRow("Passed Builds", fmt.Sprintf("%d", summary.PassedBuilds))
	tb.AddRow("Failed Builds", fmt.Sprintf("%d", summary.FailedBuilds))
	tb.AddRow("Total Duration", fmt.Sprintf("%.2fs", summary.TotalDuration))
	tb.AddRow("Average Duration", fmt.Sprintf("%.2fs", summary.AvgDuration))

	log.Info(tb.Build())
	log.Info("")
}

// displayBuildTypeSummary shows timing breakdown by module type
func displayBuildTypeSummary(summary *get.BuildTimingSummary) {
	log.Info("## Summary by Type\n")

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

	log.Info(tb.Build())
	log.Info("")
}

// displaySlowestBuilds shows the top N slowest individual module builds
func displaySlowestBuilds(summary *get.BuildTimingSummary, topN int) {
	log.Infof("## Top %d Slowest Builds\n", topN)

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
		status := "✅"
		if timing.Status == "FAIL" {
			status = "❌"
		}

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%.2f", timing.Duration),
			status,
			timing.Module,
			timing.Type,
		)
	}

	log.Info(tb.Build())
	log.Info("")
}
