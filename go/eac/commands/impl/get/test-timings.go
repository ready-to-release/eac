// Command: get test-timings
// Description: Get test timing information from test logs
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
package get

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetTestTimings)
}

// TestTiming represents timing data for a single test scenario
type TestTiming struct {
	Module   string  `json:"module" yaml:"module"`
	Package  string  `json:"package" yaml:"package"`
	Scenario string  `json:"scenario" yaml:"scenario"`
	Duration float64 `json:"duration_seconds" yaml:"duration_seconds"`
	Status   string  `json:"status" yaml:"status"` // PASS or FAIL
}

// ModuleSummary represents aggregated timing data for a module
type ModuleSummary struct {
	Module        string        `json:"module" yaml:"module"`
	TotalTests    int           `json:"total_tests" yaml:"total_tests"`
	PassedTests   int           `json:"passed_tests" yaml:"passed_tests"`
	FailedTests   int           `json:"failed_tests" yaml:"failed_tests"`
	TotalDuration float64       `json:"total_duration_seconds" yaml:"total_duration_seconds"`
	AvgDuration   float64       `json:"avg_duration_seconds" yaml:"avg_duration_seconds"`
	Scenarios     []TestTiming  `json:"scenarios" yaml:"scenarios"`
}

// TestTimingSummary represents complete timing analysis
type TestTimingSummary struct {
	TotalTests     int                       `json:"total_tests" yaml:"total_tests"`
	PassedTests    int                       `json:"passed_tests" yaml:"passed_tests"`
	FailedTests    int                       `json:"failed_tests" yaml:"failed_tests"`
	TotalDuration  float64                   `json:"total_duration_seconds" yaml:"total_duration_seconds"`
	AvgDuration    float64                   `json:"avg_duration_seconds" yaml:"avg_duration_seconds"`
	TestOutputDir  string                    `json:"test_output_dir" yaml:"test_output_dir"`
	Timings        []TestTiming              `json:"timings" yaml:"timings"`
	ByModule       map[string]ModuleSummary  `json:"by_module" yaml:"by_module"`
}

func GetTestTimings() int {
	return GetTestTimingsFiltered(nil)
}

// GetTestTimingsFiltered gets test timings with optional module filtering
func GetTestTimingsFiltered(moduleFilter []string) int {
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		// Get repository root
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		repoRoot, err := findRepoRoot(cwd)
		if err != nil {
			return nil, fmt.Errorf("failed to find repository root: %w", err)
		}

		// Find test output directory
		testOutputDir := filepath.Join(repoRoot, "out", "test")
		if _, err := os.Stat(testOutputDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("test output directory not found: %s (run tests first)", testOutputDir)
		}

		// Parse all test logs
		timings, err := ParseTestLogs(testOutputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to parse test logs: %w", err)
		}

		// Filter by modules if specified
		if len(moduleFilter) > 0 {
			timings = filterTimingsByModules(timings, moduleFilter)
		}

		if len(timings) == 0 {
			return nil, fmt.Errorf("no test timing data found in %s (run tests first)", testOutputDir)
		}

		// Build summary
		summary := BuildSummary(timings, testOutputDir)

		return summary, nil
	})
}

// ParseTestLogs recursively finds and parses all test.log files
func ParseTestLogs(testDir string) ([]TestTiming, error) {
	var allTimings []TestTiming

	err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process test.log files
		if !info.IsDir() && info.Name() == "test.log" {
			timings, err := parseTestLog(path)
			if err != nil {
				// Log warning but continue processing other files
				fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", path, err)
				return nil
			}
			allTimings = append(allTimings, timings...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return allTimings, nil
}

// TestEvent represents a JSON test event from go test
type TestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
}

// parseTestLog parses a single test.log file using JSON test events
func parseTestLog(logPath string) ([]TestTiming, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Extract module and package from path
	// Path format: out/test/<suite>/<module>/<package>/test.log
	relPath, err := filepath.Rel(filepath.Join(filepath.Dir(logPath), "..", "..", ".."), logPath)
	if err != nil {
		relPath = logPath
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	var module, pkg string
	if len(parts) >= 4 {
		module = parts[1]
		pkg = parts[2]
	}

	var timings []TestTiming
	scanner := bufio.NewScanner(file)
	decoder := json.NewDecoder(file)

	// Try JSON parsing first
	file.Seek(0, 0) // Reset to beginning
	for {
		var event TestEvent
		if err := decoder.Decode(&event); err != nil {
			break
		}

		// Only process "pass" or "fail" events with Test field (include zero elapsed time)
		if (event.Action == "pass" || event.Action == "fail") && event.Test != "" {
			// Extract scenario name from test name (after last /)
			scenario := event.Test
			if idx := strings.LastIndex(event.Test, "/"); idx != -1 {
				scenario = event.Test[idx+1:]
			}

			status := "PASS"
			if event.Action == "fail" {
				status = "FAIL"
			}

			timings = append(timings, TestTiming{
				Module:   module,
				Package:  pkg,
				Scenario: scenario,
				Duration: event.Elapsed,
				Status:   status,
			})
		}
	}

	// If JSON parsing found tests, return them
	if len(timings) > 0 {
		return timings, nil
	}

	// Fallback to regex parsing for non-JSON logs
	file.Seek(0, 0) // Reset to beginning
	scanner = bufio.NewScanner(file)
	scenarioRe := regexp.MustCompile(`^\s+---\s+(PASS|FAIL):\s+\S+/(.+?)\s+\(([0-9.]+)s\)`)

	for scanner.Scan() {
		line := scanner.Text()

		if matches := scenarioRe.FindStringSubmatch(line); matches != nil {
			status := matches[1]
			scenario := matches[2]
			durationStr := matches[3]

			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				continue
			}

			timings = append(timings, TestTiming{
				Module:   module,
				Package:  pkg,
				Scenario: scenario,
				Duration: duration,
				Status:   status,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return timings, nil
}

// filterTimingsByModules filters timings to only include specified modules
func filterTimingsByModules(timings []TestTiming, modules []string) []TestTiming {
	// Create a set of modules for O(1) lookup
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}

	filtered := []TestTiming{}
	for _, timing := range timings {
		if moduleSet[timing.Module] {
			filtered = append(filtered, timing)
		}
	}

	return filtered
}

// BuildSummary aggregates timing data and builds summary
func BuildSummary(timings []TestTiming, testOutputDir string) *TestTimingSummary {
	summary := &TestTimingSummary{
		TestOutputDir: testOutputDir,
		Timings:       timings,
		ByModule:      make(map[string]ModuleSummary),
	}

	// Aggregate by module
	moduleData := make(map[string]*ModuleSummary)

	for _, timing := range timings {
		// Update overall stats
		summary.TotalTests++
		summary.TotalDuration += timing.Duration
		if timing.Status == "PASS" {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}

		// Update module stats
		if _, exists := moduleData[timing.Module]; !exists {
			moduleData[timing.Module] = &ModuleSummary{
				Module:    timing.Module,
				Scenarios: []TestTiming{},
			}
		}

		mod := moduleData[timing.Module]
		mod.TotalTests++
		mod.TotalDuration += timing.Duration
		mod.Scenarios = append(mod.Scenarios, timing)
		if timing.Status == "PASS" {
			mod.PassedTests++
		} else {
			mod.FailedTests++
		}
	}

	// Calculate averages
	if summary.TotalTests > 0 {
		summary.AvgDuration = summary.TotalDuration / float64(summary.TotalTests)
	}

	for module, data := range moduleData {
		if data.TotalTests > 0 {
			data.AvgDuration = data.TotalDuration / float64(data.TotalTests)
		}
		summary.ByModule[module] = *data
	}

	return summary
}
