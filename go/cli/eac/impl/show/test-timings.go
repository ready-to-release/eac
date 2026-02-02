// Command: show test-timings
// Short: Display test timing metrics in tables
// Long: The show test-timings command displays test timing analysis parsed from test logs.
// Long: Shows overall statistics, breakdown by module, and the slowest individual test scenarios.
// Long:
// Long: Expected Output:
// Long: - Table with overall metrics: total tests, passed/failed counts, test duration, average duration
// Long: - Summary by module table: modules with test counts, pass/fail, total and average times
// Long: - Top N slowest tests table showing duration, status, module name, and scenario
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get"
	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
)

func init() {
	registry.Register(ShowTestTimings)
}

func ShowTestTimings() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return ShowTestTimingsForModules(nil, 20, "", 0)
}

// ShowTestTimingsForModules displays test timings with optional module filtering
// If modules is nil, shows all timings. topN controls how many slowest tests to show.
// If testOutputDir is empty, defaults to out/test.
// If wallClockSeconds > 0, displays overhead/setup time.
func ShowTestTimingsForModules(modules []string, topN int, testOutputDir string, wallClockSeconds float64) int {
	// Get repository root if needed
	if testOutputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Errorf("failed to get current directory: %v", err)
			return 1
		}

		repoRoot, err := testdata.FindRepoRoot(cwd)
		if err != nil {
			log.Errorf("failed to find repository root: %v", err)
			return 1
		}

		// Load config to get test output directory
		cfg, err := config.Load(config.LoadOptions{RepoRoot: repoRoot})
		if err != nil {
			log.Errorf("failed to load config: %v", err)
			return 1
		}

		testOutputDir = filepath.Join(repoRoot, cfg.Repository.Paths.Out.Test)
	}
	if _, err := os.Stat(testOutputDir); os.IsNotExist(err) {
		log.Errorf("test output directory not found: %s (run tests first)", testOutputDir)
		return 1
	}

	// Parse test logs using the get command logic
	timings, err := get.ParseTestLogs(testOutputDir)
	if err != nil {
		log.Errorf("failed to parse test logs: %v", err)
		return 1
	}

	// Filter by modules if specified
	if len(modules) > 0 {
		timings = filterTimingsByModules(timings, modules)
	}

	if len(timings) == 0 {
		if len(modules) > 0 {
			log.Errorf("no test timing data found for modules: %v", modules)
		} else {
			log.Errorf("no test timing data found in %s (run tests first)", testOutputDir)
		}
		return 1
	}

	// Build summary
	summary := get.BuildSummary(timings, testOutputDir)

	// Display overall summary
	displayOverallSummary(summary, wallClockSeconds)

	// Display module summary table
	displayModuleSummary(summary)

	// Display slowest tests
	displaySlowestTests(summary, topN)

	return 0
}

// filterTimingsByModules filters timings to only include specified modules.
func filterTimingsByModules(timings []get.TestTiming, modules []string) []get.TestTiming {
	// Create a set of modules for O(1) lookup
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}

	filtered := []get.TestTiming{}
	for _, timing := range timings {
		if moduleSet[timing.Module] {
			filtered = append(filtered, timing)
		}
	}

	return filtered
}

// displayOverallSummary shows high-level statistics.
func displayOverallSummary(summary *get.TestTimingSummary, wallClockSeconds float64) {
	log.Info("# Test Timing Analysis\n")
	log.Infof("**Test Output Directory**: `%s`\n", summary.TestOutputDir)
	log.Info("")

	// Build summary table
	tb := render.NewTableBuilder().
		WithHeaders("Metric", "Value")

	tb.AddRow("Total Tests", fmt.Sprintf("%d", summary.TotalTests))
	tb.AddRow("Passed Tests", fmt.Sprintf("%d", summary.PassedTests))
	tb.AddRow("Failed Tests", fmt.Sprintf("%d", summary.FailedTests))
	tb.AddRow("Test Duration", fmt.Sprintf("%.2fs", summary.TotalDuration))

	// Add overhead/setup time if wall-clock time is provided
	if wallClockSeconds > 0 {
		overhead := wallClockSeconds - summary.TotalDuration
		if overhead < 0 {
			overhead = 0 // Parallel execution can make this negative
		}
		tb.AddRow("Setup/Overhead", fmt.Sprintf("%.2fs", overhead))
		tb.AddRow("Wall-Clock Time", fmt.Sprintf("%.2fs", wallClockSeconds))
	}

	tb.AddRow("Average Duration", fmt.Sprintf("%.2fs", summary.AvgDuration))

	log.Info(tb.Build())
	log.Info("")
}

// displayModuleSummary shows timing breakdown by module.
func displayModuleSummary(summary *get.TestTimingSummary) {
	log.Info("## Summary by Module\n")

	// Convert map to slice for sorting
	type moduleStat struct {
		Module        string
		TotalTests    int
		PassedTests   int
		FailedTests   int
		TotalDuration float64
		AvgDuration   float64
	}

	var stats []moduleStat
	for _, mod := range summary.ByModule {
		stats = append(stats, moduleStat{
			Module:        mod.Module,
			TotalTests:    mod.TotalTests,
			PassedTests:   mod.PassedTests,
			FailedTests:   mod.FailedTests,
			TotalDuration: mod.TotalDuration,
			AvgDuration:   mod.AvgDuration,
		})
	}

	// Sort by total duration (slowest first)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TotalDuration > stats[j].TotalDuration
	})

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Module", "Tests", "Passed", "Failed", "Total (s)", "Avg (s)")

	for _, stat := range stats {
		tb.AddRow(
			stat.Module,
			fmt.Sprintf("%d", stat.TotalTests),
			fmt.Sprintf("%d", stat.PassedTests),
			fmt.Sprintf("%d", stat.FailedTests),
			fmt.Sprintf("%.2f", stat.TotalDuration),
			fmt.Sprintf("%.2f", stat.AvgDuration),
		)
	}

	log.Info(tb.Build())
	log.Info("")
}

// displaySlowestTests shows the top N slowest individual test scenarios.
func displaySlowestTests(summary *get.TestTimingSummary, topN int) {
	log.Infof("## Top %d Slowest Tests\n", topN)

	// Sort timings by duration (slowest first)
	sortedTimings := make([]get.TestTiming, len(summary.Timings))
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
		WithHeaders("#", "Duration (s)", "Status", "Module", "Scenario")

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
			timing.Scenario,
		)
	}

	log.Info(tb.Build())
	log.Info("")
}
