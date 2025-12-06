// Command: get build-times
// Description: Get build timing information from build logs
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
package get

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	get "github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetBuildTimes)
}

// BuildTiming represents timing data for a single module build
type BuildTiming struct {
	Module   string  `json:"module" yaml:"module"`
	Duration float64 `json:"duration_seconds" yaml:"duration_seconds"`
	Status   string  `json:"status" yaml:"status"` // PASS or FAIL
	Type     string  `json:"type" yaml:"type"`     // Module type (e.g., go-library, go-cli)
}

// BuildTimingSummary represents complete build timing analysis
type BuildTimingSummary struct {
	TotalBuilds    int                     `json:"total_builds" yaml:"total_builds"`
	PassedBuilds   int                     `json:"passed_builds" yaml:"passed_builds"`
	FailedBuilds   int                     `json:"failed_builds" yaml:"failed_builds"`
	TotalDuration  float64                 `json:"total_duration_seconds" yaml:"total_duration_seconds"`
	AvgDuration    float64                 `json:"avg_duration_seconds" yaml:"avg_duration_seconds"`
	BuildOutputDir string                  `json:"build_output_dir" yaml:"build_output_dir"`
	Timings        []BuildTiming           `json:"timings" yaml:"timings"`
	ByType         map[string]TypeSummary  `json:"by_type" yaml:"by_type"`
}

// TypeSummary represents aggregated timing data by module type
type TypeSummary struct {
	Type          string        `json:"type" yaml:"type"`
	TotalBuilds   int           `json:"total_builds" yaml:"total_builds"`
	PassedBuilds  int           `json:"passed_builds" yaml:"passed_builds"`
	FailedBuilds  int           `json:"failed_builds" yaml:"failed_builds"`
	TotalDuration float64       `json:"total_duration_seconds" yaml:"total_duration_seconds"`
	AvgDuration   float64       `json:"avg_duration_seconds" yaml:"avg_duration_seconds"`
	Modules       []BuildTiming `json:"modules" yaml:"modules"`
}

func GetBuildTimes() int {
	return GetBuildTimesFiltered(nil, "")
}

// GetBuildTimesFiltered gets build timings with optional module filtering
// If buildOutputDir is empty, defaults to out/build
func GetBuildTimesFiltered(moduleFilter []string, buildOutputDir string) int {
	return get.ExecuteGetCommand(func() (interface{}, error) {
		// Get repository root if needed
		if buildOutputDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current directory: %w", err)
			}

			repoRoot, err := findRepoRoot(cwd)
			if err != nil {
				return nil, fmt.Errorf("failed to find repository root: %w", err)
			}

			buildOutputDir = filepath.Join(repoRoot, "out", "build")
		}

		// Check if build output directory exists
		if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("build output directory not found: %s (run build first)", buildOutputDir)
		}

		// Parse orchestrator log
		timings, err := ParseBuildLog(buildOutputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to parse build logs: %w", err)
		}

		// Filter by modules if specified
		if len(moduleFilter) > 0 {
			timings = filterBuildTimingsByModules(timings, moduleFilter)
		}

		if len(timings) == 0 {
			return nil, fmt.Errorf("no build timing data found in %s (run build first)", buildOutputDir)
		}

		// Build summary
		summary := BuildBuildSummary(timings, buildOutputDir)

		return summary, nil
	})
}

// ParseBuildLog parses the orchestrator.log file to extract build timings
func ParseBuildLog(buildDir string) ([]BuildTiming, error) {
	logPath := filepath.Join(buildDir, "orchestrator.log")

	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open orchestrator.log: %w", err)
	}
	defer file.Close()

	var timings []BuildTiming
	scanner := bufio.NewScanner(file)

	// Regex to match build result lines
	// Format: "✅   1.1s eac-core                                       go-library"
	// Format: "❌   2.3s module-name                                   module-type"
	resultRe := regexp.MustCompile(`^([✅❌])\s+([0-9.]+)s\s+(\S+)\s+(\S+)\s*$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Try to match build result line
		if matches := resultRe.FindStringSubmatch(line); matches != nil {
			statusIcon := matches[1]
			durationStr := matches[2]
			module := matches[3]
			moduleType := matches[4]

			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				continue
			}

			status := "PASS"
			if statusIcon == "❌" {
				status = "FAIL"
			}

			timings = append(timings, BuildTiming{
				Module:   module,
				Duration: duration,
				Status:   status,
				Type:     moduleType,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return timings, nil
}

// filterBuildTimingsByModules filters timings to only include specified modules
func filterBuildTimingsByModules(timings []BuildTiming, modules []string) []BuildTiming {
	// Create a set of modules for O(1) lookup
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}

	filtered := []BuildTiming{}
	for _, timing := range timings {
		if moduleSet[timing.Module] {
			filtered = append(filtered, timing)
		}
	}

	return filtered
}

// BuildBuildSummary aggregates timing data and builds summary
func BuildBuildSummary(timings []BuildTiming, buildOutputDir string) *BuildTimingSummary {
	summary := &BuildTimingSummary{
		BuildOutputDir: buildOutputDir,
		Timings:        timings,
		ByType:         make(map[string]TypeSummary),
	}

	// Aggregate by type
	typeData := make(map[string]*TypeSummary)

	for _, timing := range timings {
		// Update overall stats
		summary.TotalBuilds++
		summary.TotalDuration += timing.Duration
		if timing.Status == "PASS" {
			summary.PassedBuilds++
		} else {
			summary.FailedBuilds++
		}

		// Update type stats
		if _, exists := typeData[timing.Type]; !exists {
			typeData[timing.Type] = &TypeSummary{
				Type:    timing.Type,
				Modules: []BuildTiming{},
			}
		}

		ts := typeData[timing.Type]
		ts.TotalBuilds++
		ts.TotalDuration += timing.Duration
		ts.Modules = append(ts.Modules, timing)
		if timing.Status == "PASS" {
			ts.PassedBuilds++
		} else {
			ts.FailedBuilds++
		}
	}

	// Calculate averages
	if summary.TotalBuilds > 0 {
		summary.AvgDuration = summary.TotalDuration / float64(summary.TotalBuilds)
	}

	for typeName, data := range typeData {
		if data.TotalBuilds > 0 {
			data.AvgDuration = data.TotalDuration / float64(data.TotalBuilds)
		}
		summary.ByType[typeName] = *data
	}

	return summary
}
