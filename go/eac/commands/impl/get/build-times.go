// Command: get build-times
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
//
// Long:
// Long: Expected Output:
// Long: YAML with build timing metrics parsed from per-module build.manifest.json files:
// Long:   - Per-module timing data with duration in seconds
// Long:   - Aggregated statistics by module type
// Long:   - Overall summary with total builds and average duration
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain/reports"
)

func init() {
	registry.Register(GetBuildTimes)
}

// BuildTiming represents timing data for a single module build.
type BuildTiming struct {
	Module   string  `json:"module" yaml:"module"`
	Duration float64 `json:"duration_seconds" yaml:"duration_seconds"`
	Status   string  `json:"status" yaml:"status"` // PASS or FAIL
	Type     string  `json:"type" yaml:"type"`     // Module type (e.g., go, container, typescript, static)
}

// BuildTimingSummary represents complete build timing analysis.
type BuildTimingSummary struct {
	TotalBuilds    int                    `json:"total_builds" yaml:"total_builds"`
	PassedBuilds   int                    `json:"passed_builds" yaml:"passed_builds"`
	FailedBuilds   int                    `json:"failed_builds" yaml:"failed_builds"`
	TotalDuration  float64                `json:"total_duration_seconds" yaml:"total_duration_seconds"`
	AvgDuration    float64                `json:"avg_duration_seconds" yaml:"avg_duration_seconds"`
	BuildOutputDir string                 `json:"build_output_dir" yaml:"build_output_dir"`
	Timings        []BuildTiming          `json:"timings" yaml:"timings"`
	ByType         map[string]TypeSummary `json:"by_type" yaml:"by_type"`
}

// TypeSummary represents aggregated timing data by module type.
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
// If buildOutputDir is empty, defaults to out/build.
func GetBuildTimesFiltered(moduleFilter []string, buildOutputDir string) int {
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		// Get repository root if needed
		var repoRoot string
		if buildOutputDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current directory: %w", err)
			}

			repoRoot, err = testdata.FindRepoRoot(cwd)
			if err != nil {
				return nil, fmt.Errorf("failed to find repository root: %w", err)
			}

			// Load config for path resolution
			cfg, err := config.Load(config.LoadOptions{RepoRoot: repoRoot})
			if err != nil {
				return nil, fmt.Errorf("failed to load config: %w", err)
			}

			buildOutputDir = filepath.Join(repoRoot, cfg.Repository.Paths.Out.Build)
		} else {
			// If buildOutputDir is provided, derive repo root from it
			var err error
			repoRoot, err = testdata.FindRepoRoot(buildOutputDir)
			if err != nil {
				return nil, fmt.Errorf("failed to find repository root: %w", err)
			}
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

		// Populate module types from contracts
		timings, err = populateModuleTypes(timings, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to populate module types: %w", err)
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

// ParseBuildLog reads build timing data from per-module manifest files.
// It scans all subdirectories in the build output dir for build.manifest.json files.
func ParseBuildLog(buildDir string) ([]BuildTiming, error) {
	var timings []BuildTiming

	// Scan build directory for module subdirectories
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read build directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moduleName := entry.Name()
		manifestPath := filepath.Join(buildDir, moduleName, "build.manifest.json")

		// Check if manifest exists
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		// Read manifest file
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		// Parse manifest JSON - we only need duration_seconds and moniker
		var manifest struct {
			Moniker         string  `json:"moniker"`
			Type            string  `json:"type"`
			DurationSeconds float64 `json:"duration_seconds"`
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		// Use directory name as moniker if not in manifest (backward compat)
		moniker := manifest.Moniker
		if moniker == "" {
			moniker = moduleName
		}

		timings = append(timings, BuildTiming{
			Module:   moniker,
			Duration: manifest.DurationSeconds,
			Status:   "PASS", // If manifest exists, build succeeded
			Type:     manifest.Type,
		})
	}

	return timings, nil
}

// filterBuildTimingsByModules filters timings to only include specified modules.
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

// populateModuleTypes loads module contracts and populates the Type field for each timing.
func populateModuleTypes(timings []BuildTiming, repoRoot string) ([]BuildTiming, error) {
	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		return timings, fmt.Errorf("failed to load module contracts: %w", err)
	}

	// Populate package types
	for i := range timings {
		if module, exists := moduleReport.Registry.Get(timings[i].Module); exists {
			timings[i].Type = module.GetComponentTypesDisplay()
		} else {
			// Module not found in contracts, use a placeholder
			timings[i].Type = "unknown"
		}
	}

	return timings, nil
}

// BuildBuildSummary aggregates timing data and builds summary.
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
